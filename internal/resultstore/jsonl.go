package resultstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const maxInputSize = 8 << 20

type Kind string

const (
	KindGeneric    Kind = "generic"
	KindEvents     Kind = "events"
	KindExecutions Kind = "executions"
)

type Validation struct {
	OK        bool     `json:"ok"`
	LineCount int      `json:"lineCount"`
	Errors    []string `json:"errors,omitempty"`
}

type AppendResult struct {
	LineNumber int `json:"lineNumber"`
}

type InitResult struct {
	Created bool `json:"created"`
}

func ParseKind(value string) (Kind, error) {
	kind := Kind(value)
	switch kind {
	case KindGeneric, KindEvents, KindExecutions:
		return kind, nil
	default:
		return "", fmt.Errorf("unsupported kind %q", value)
	}
}

func ValidateFile(path string, kind Kind) Validation {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Validation{Errors: []string{"file_not_found"}}
		}
		return Validation{Errors: []string{"read_failed: " + err.Error()}}
	}
	return Validate(data, kind)
}

func Init(path string, kind Kind) (InitResult, error) {
	directory := filepath.Dir(path)
	info, err := os.Stat(directory)
	if err != nil || !info.IsDir() {
		return InitResult{}, errors.New("parent_directory_not_found")
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err == nil {
		if syncErr := file.Sync(); syncErr != nil {
			file.Close()
			return InitResult{}, fmt.Errorf("sync_failed: %w", syncErr)
		}
		if closeErr := file.Close(); closeErr != nil {
			return InitResult{}, fmt.Errorf("close_failed: %w", closeErr)
		}
		return InitResult{Created: true}, nil
	}
	if !errors.Is(err, os.ErrExist) {
		return InitResult{}, fmt.Errorf("create_failed: %w", err)
	}

	validation := ValidateFile(path, kind)
	if !validation.OK {
		return InitResult{}, fmt.Errorf("existing_jsonl_invalid: %s", strings.Join(validation.Errors, "; "))
	}
	return InitResult{Created: false}, nil
}

func Validate(data []byte, kind Kind) Validation {
	result := Validation{}
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		result.Errors = append(result.Errors, "utf8_bom_not_allowed")
	}
	if !utf8.Valid(data) {
		result.Errors = append(result.Errors, "invalid_utf8")
		return result
	}
	if len(data) > 0 && data[len(data)-1] != '\n' {
		result.Errors = append(result.Errors, "missing_final_newline")
	}

	lines := bytes.Split(data, []byte{'\n'})
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	result.LineCount = len(lines)

	executionIDs := make(map[string]int)
	caseAttempts := make(map[string]int)
	for index, rawLine := range lines {
		lineNumber := index + 1
		line := bytes.TrimSuffix(rawLine, []byte{'\r'})
		if len(bytes.TrimSpace(line)) == 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("line_%d: empty_line", lineNumber))
			continue
		}

		var object map[string]json.RawMessage
		if err := json.Unmarshal(line, &object); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("line_%d: invalid_json: %v", lineNumber, err))
			continue
		}
		if object == nil {
			result.Errors = append(result.Errors, fmt.Sprintf("line_%d: json_object_required", lineNumber))
			continue
		}

		switch kind {
		case KindEvents:
			validateEvent(lineNumber, object, &result)
		case KindExecutions:
			validateExecution(lineNumber, object, executionIDs, caseAttempts, &result)
		}
	}

	result.OK = len(result.Errors) == 0
	return result
}

func Append(path string, input []byte, kind Kind) (AppendResult, error) {
	if len(input) == 0 {
		return AppendResult{}, errors.New("json_required")
	}
	if len(input) > maxInputSize {
		return AppendResult{}, fmt.Errorf("json_too_large: max=%d", maxInputSize)
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal(input, &object); err != nil {
		return AppendResult{}, fmt.Errorf("invalid_json_input: %w", err)
	}
	if object == nil {
		return AppendResult{}, errors.New("json_object_required")
	}
	line, err := json.Marshal(object)
	if err != nil {
		return AppendResult{}, fmt.Errorf("compact_json_failed: %w", err)
	}

	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return AppendResult{}, fmt.Errorf("read_existing_failed: %w", err)
	}
	if errors.Is(err, os.ErrNotExist) {
		existing = nil
	}

	if len(existing) > 0 {
		validation := Validate(existing, kind)
		if !validation.OK {
			return AppendResult{}, fmt.Errorf("existing_jsonl_invalid: %s", strings.Join(validation.Errors, "; "))
		}
	}

	candidate := make([]byte, 0, len(existing)+len(line)+1)
	candidate = append(candidate, existing...)
	candidate = append(candidate, line...)
	candidate = append(candidate, '\n')
	validation := Validate(candidate, kind)
	if !validation.OK {
		return AppendResult{}, fmt.Errorf("candidate_jsonl_invalid: %s", strings.Join(validation.Errors, "; "))
	}

	directory := filepath.Dir(path)
	info, err := os.Stat(directory)
	if err != nil || !info.IsDir() {
		return AppendResult{}, errors.New("parent_directory_not_found")
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return AppendResult{}, fmt.Errorf("open_failed: %w", err)
	}
	defer file.Close()

	written, err := file.Write(append(line, '\n'))
	if err != nil {
		return AppendResult{}, fmt.Errorf("append_failed: %w", err)
	}
	if written != len(line)+1 {
		return AppendResult{}, fmt.Errorf("short_write: expected=%d actual=%d", len(line)+1, written)
	}
	if err := file.Sync(); err != nil {
		return AppendResult{}, fmt.Errorf("sync_failed: %w", err)
	}

	after := ValidateFile(path, kind)
	if !after.OK {
		return AppendResult{}, fmt.Errorf("post_append_validation_failed: %s", strings.Join(after.Errors, "; "))
	}
	if after.LineCount != validation.LineCount {
		return AppendResult{}, fmt.Errorf("line_count_mismatch: expected=%d actual=%d", validation.LineCount, after.LineCount)
	}

	return AppendResult{LineNumber: after.LineCount}, nil
}

func validateEvent(lineNumber int, object map[string]json.RawMessage, result *Validation) {
	requireString(lineNumber, "time", object, result)
	requireString(lineNumber, "skill_version", object, result)
	requireNumber(lineNumber, "schema_version", object, result)
	requireString(lineNumber, "run_id", object, result)
	requireString(lineNumber, "mode", object, result)
	event := requireString(lineNumber, "event", object, result)
	if event == "self_check" && schemaVersionAtLeast(object, 2) {
		result.Errors = append(result.Errors, fmt.Sprintf("line_%d: deprecated_event_self_check_use_self_check_finished", lineNumber))
	}
}

// schemaVersionAtLeast preserves validation of historical result files. Schema 1
// used self_check; schema 2 makes self_check_finished the canonical event name.
func schemaVersionAtLeast(object map[string]json.RawMessage, minimum int) bool {
	raw, exists := object["schema_version"]
	if !exists {
		return false
	}

	var version int
	if err := json.Unmarshal(raw, &version); err != nil {
		return false
	}
	return version >= minimum
}

func validateExecution(lineNumber int, object map[string]json.RawMessage, executionIDs, caseAttempts map[string]int, result *Validation) {
	executionID := requireStringAlias(lineNumber, []string{"executionId", "execution_id"}, object, result)
	caseID := requireStringAlias(lineNumber, []string{"caseId", "case_id"}, object, result)
	attempt := requirePositiveInteger(lineNumber, "attempt", object, result)
	requireStringAlias(lineNumber, []string{"runId", "run_id"}, object, result)
	requireString(lineNumber, "status", object, result)
	stage := optionalStage(lineNumber, object, result)

	if executionID != "" {
		if previous, exists := executionIDs[executionID]; exists {
			result.Errors = append(result.Errors, fmt.Sprintf("line_%d: duplicate_execution_id_%q_first_seen_line_%d", lineNumber, executionID, previous))
		} else {
			executionIDs[executionID] = lineNumber
		}
	}
	if caseID != "" && attempt > 0 {
		key := fmt.Sprintf("%s#%s#%d", caseID, stage, attempt)
		if previous, exists := caseAttempts[key]; exists {
			result.Errors = append(result.Errors, fmt.Sprintf("line_%d: duplicate_case_attempt_%q_first_seen_line_%d", lineNumber, key, previous))
		} else {
			caseAttempts[key] = lineNumber
		}
	}
}

// optionalStage keeps Schema 1-4 execution histories valid. Schema 5 writers
// must provide fast or browser; legacy records use a separate uniqueness scope.
func optionalStage(lineNumber int, object map[string]json.RawMessage, result *Validation) string {
	raw, exists := object["stage"]
	if !exists {
		return "legacy"
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("line_%d: invalid_stage", lineNumber))
		return "invalid"
	}
	switch value {
	case "fast", "browser":
		return value
	default:
		result.Errors = append(result.Errors, fmt.Sprintf("line_%d: invalid_stage", lineNumber))
		return "invalid"
	}
}

func requireStringAlias(lineNumber int, fields []string, object map[string]json.RawMessage, result *Validation) string {
	for _, field := range fields {
		if _, exists := object[field]; exists {
			return requireString(lineNumber, field, object, result)
		}
	}
	result.Errors = append(result.Errors, fmt.Sprintf("line_%d: missing_%s", lineNumber, strings.Join(fields, "_or_")))
	return ""
}

func requireString(lineNumber int, field string, object map[string]json.RawMessage, result *Validation) string {
	raw, exists := object[field]
	if !exists {
		result.Errors = append(result.Errors, fmt.Sprintf("line_%d: missing_%s", lineNumber, field))
		return ""
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil || strings.TrimSpace(value) == "" {
		result.Errors = append(result.Errors, fmt.Sprintf("line_%d: invalid_%s", lineNumber, field))
		return ""
	}
	return value
}

func requireNumber(lineNumber int, field string, object map[string]json.RawMessage, result *Validation) {
	raw, exists := object[field]
	if !exists {
		result.Errors = append(result.Errors, fmt.Sprintf("line_%d: missing_%s", lineNumber, field))
		return
	}
	var value json.Number
	if err := json.Unmarshal(raw, &value); err != nil || value.String() == "" {
		result.Errors = append(result.Errors, fmt.Sprintf("line_%d: invalid_%s", lineNumber, field))
	}
}

func requirePositiveInteger(lineNumber int, field string, object map[string]json.RawMessage, result *Validation) int {
	raw, exists := object[field]
	if !exists {
		result.Errors = append(result.Errors, fmt.Sprintf("line_%d: missing_%s", lineNumber, field))
		return 0
	}
	var value int
	if err := json.Unmarshal(raw, &value); err != nil || value < 1 {
		result.Errors = append(result.Errors, fmt.Sprintf("line_%d: invalid_%s", lineNumber, field))
		return 0
	}
	return value
}
