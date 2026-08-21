package resultstore

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendEventCreatesCompactUTF8JSONL(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "events.jsonl")
	input := []byte(`{
  "time": "2026-08-20T18:00:00+08:00",
  "skill_version": "0.2.0-dev.3",
  "schema_version": 1,
  "run_id": "run-1",
  "mode": "development",
  "event": "step_observed",
  "message": "页面可见"
}`)

	result, err := Append(path, input, KindEvents)
	if err != nil {
		t.Fatal(err)
	}
	if result.LineNumber != 1 {
		t.Fatalf("line number = %d", result.LineNumber)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatal("unexpected UTF-8 BOM")
	}
	if !bytes.HasSuffix(data, []byte{'\n'}) {
		t.Fatal("missing final newline")
	}
	if bytes.Count(data, []byte{'\n'}) != 1 {
		t.Fatalf("expected one physical line, got %q", data)
	}
	if !bytes.Contains(data, []byte("页面可见")) {
		t.Fatalf("UTF-8 text missing: %q", data)
	}
	if validation := ValidateFile(path, KindEvents); !validation.OK {
		t.Fatalf("validation failed: %v", validation.Errors)
	}
}

func TestInitCreatesEmptyValidFileAndIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	first, err := Init(path, KindEvents)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Created {
		t.Fatal("expected file to be created")
	}
	if validation := ValidateFile(path, KindEvents); !validation.OK || validation.LineCount != 0 {
		t.Fatalf("unexpected validation: %+v", validation)
	}

	second, err := Init(path, KindEvents)
	if err != nil {
		t.Fatal(err)
	}
	if second.Created {
		t.Fatal("existing valid file should not be recreated")
	}
}

func TestAppendRejectsMultipleJSONValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	_, err := Append(path, []byte(`{"event":"one"}{"event":"two"}`), KindGeneric)
	if err == nil || !strings.Contains(err.Error(), "invalid_json_input") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("file should not be created, stat error: %v", statErr)
	}
}

func TestEventValidationRejectsDeprecatedSelfCheckEvent(t *testing.T) {
	data := []byte(`{"time":"2026-08-21T12:00:00+08:00","skill_version":"0.2.0-dev.5","schema_version":2,"run_id":"run-1","mode":"development","event":"self_check"}` + "\n")
	validation := Validate(data, KindEvents)
	if validation.OK {
		t.Fatal("expected deprecated event to be rejected")
	}
	joined := strings.Join(validation.Errors, ";")
	if !strings.Contains(joined, "deprecated_event_self_check_use_self_check_finished") {
		t.Fatalf("unexpected errors: %v", validation.Errors)
	}
}

func TestEventValidationAcceptsLegacySelfCheckEvent(t *testing.T) {
	data := []byte(`{"time":"2026-08-20T12:00:00+08:00","skill_version":"0.2.0-dev.4","schema_version":1,"run_id":"run-1","mode":"development","event":"self_check"}` + "\n")
	validation := Validate(data, KindEvents)
	if !validation.OK {
		t.Fatalf("legacy schema should remain valid: %v", validation.Errors)
	}
}

func TestEventValidationAcceptsSelfCheckFinished(t *testing.T) {
	data := []byte(`{"time":"2026-08-21T12:00:00+08:00","skill_version":"0.2.0-dev.5","schema_version":2,"run_id":"run-1","mode":"development","event":"self_check_finished"}` + "\n")
	validation := Validate(data, KindEvents)
	if !validation.OK {
		t.Fatalf("unexpected errors: %v", validation.Errors)
	}
}

func TestAppendRejectsDuplicateExecutionBeforeWrite(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "executions.jsonl")
	first := []byte(`{"executionId":"exec-1","runId":"run-1","caseId":"TC-1","attempt":1,"status":"passed"}`)
	if _, err := Append(path, first, KindExecutions); err != nil {
		t.Fatal(err)
	}

	duplicate := []byte(`{"executionId":"exec-1","runId":"run-1","caseId":"TC-1","attempt":2,"status":"passed"}`)
	_, err := Append(path, duplicate, KindExecutions)
	if err == nil || !strings.Contains(err.Error(), "duplicate_execution_id") {
		t.Fatalf("unexpected error: %v", err)
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if bytes.Count(data, []byte{'\n'}) != 1 {
		t.Fatalf("duplicate was appended: %q", data)
	}
}

func TestAppendRejectsDuplicateCaseAttemptBeforeWrite(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "executions.jsonl")
	first := []byte(`{"executionId":"exec-1","runId":"run-1","caseId":"TC-1","attempt":1,"status":"passed"}`)
	if _, err := Append(path, first, KindExecutions); err != nil {
		t.Fatal(err)
	}

	duplicate := []byte(`{"executionId":"exec-2","runId":"run-1","caseId":"TC-1","attempt":1,"status":"passed"}`)
	_, err := Append(path, duplicate, KindExecutions)
	if err == nil || !strings.Contains(err.Error(), "duplicate_case_attempt") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecutionValidationAcceptsLegacySnakeCaseIDs(t *testing.T) {
	data := []byte("{\"execution_id\":\"exec-1\",\"run_id\":\"run-1\",\"case_id\":\"TC-1\",\"attempt\":1,\"status\":\"passed\"}\n")
	validation := Validate(data, KindExecutions)
	if !validation.OK {
		t.Fatalf("unexpected errors: %v", validation.Errors)
	}
}

func TestAppendRejectsCorruptExistingFile(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "events.jsonl")
	if err := os.WriteFile(path, []byte("{\"event\":\"one\"}{\"event\":\"two\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Append(path, []byte(`{"event":"three"}`), KindGeneric)
	if err == nil || !strings.Contains(err.Error(), "existing_jsonl_invalid") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsFragmentsAndMissingNewline(t *testing.T) {
	validation := Validate([]byte("{\"ok\":true}\nfragment"), KindGeneric)
	if validation.OK {
		t.Fatal("expected invalid JSONL")
	}
	joined := strings.Join(validation.Errors, ";")
	if !strings.Contains(joined, "missing_final_newline") || !strings.Contains(joined, "line_2: invalid_json") {
		t.Fatalf("unexpected errors: %v", validation.Errors)
	}
}
