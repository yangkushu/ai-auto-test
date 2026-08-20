package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yangkushu/ai-auto-test/internal/resultstore"
)

var version = "dev"

type commandOutput struct {
	OK         bool     `json:"ok"`
	Action     string   `json:"action"`
	Version    string   `json:"version"`
	Path       string   `json:"path,omitempty"`
	Kind       string   `json:"kind,omitempty"`
	LineCount  *int     `json:"lineCount,omitempty"`
	LineNumber *int     `json:"lineNumber,omitempty"`
	Created    *bool    `json:"created,omitempty"`
	Errors     []string `json:"errors,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fail("usage", "expected init-jsonl, append-jsonl, validate-jsonl, or version")
	}

	switch os.Args[1] {
	case "init-jsonl":
		initJSONL(os.Args[2:])
	case "append-jsonl":
		appendJSONL(os.Args[2:])
	case "validate-jsonl":
		validateJSONL(os.Args[2:])
	case "version":
		writeOutput(commandOutput{OK: true, Action: "version", Version: version})
	default:
		fail("usage", "unknown command: "+os.Args[1])
	}
}

func initJSONL(args []string) {
	flags := flag.NewFlagSet("init-jsonl", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	path := flags.String("file", "", "JSONL file")
	kindValue := flags.String("kind", "generic", "generic, events, or executions")
	if err := flags.Parse(args); err != nil {
		fail("init-jsonl", err.Error())
	}
	if strings.TrimSpace(*path) == "" {
		fail("init-jsonl", "file_required")
	}
	kind, err := resultstore.ParseKind(*kindValue)
	if err != nil {
		fail("init-jsonl", err.Error())
	}
	result, err := resultstore.Init(*path, kind)
	if err != nil {
		fail("init-jsonl", err.Error())
	}
	created := result.Created
	writeOutput(commandOutput{
		OK:      true,
		Action:  "init-jsonl",
		Version: version,
		Path:    *path,
		Kind:    string(kind),
		Created: &created,
	})
}

func appendJSONL(args []string) {
	flags := flag.NewFlagSet("append-jsonl", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	path := flags.String("file", "", "JSONL file")
	kindValue := flags.String("kind", "generic", "generic, events, or executions")
	jsonValue := flags.String("json", "", "one JSON object; stdin is used when omitted")
	jsonFile := flags.String("json-file", "", "file containing one JSON object")
	if err := flags.Parse(args); err != nil {
		fail("append-jsonl", err.Error())
	}
	if strings.TrimSpace(*path) == "" {
		fail("append-jsonl", "file_required")
	}
	kind, err := resultstore.ParseKind(*kindValue)
	if err != nil {
		fail("append-jsonl", err.Error())
	}

	if strings.TrimSpace(*jsonValue) != "" && strings.TrimSpace(*jsonFile) != "" {
		fail("append-jsonl", "json_and_json_file_are_mutually_exclusive")
	}

	input := []byte(*jsonValue)
	if strings.TrimSpace(*jsonFile) != "" {
		input, err = os.ReadFile(*jsonFile)
		if err != nil {
			fail("append-jsonl", "json_file_read_failed: "+err.Error())
		}
	} else if strings.TrimSpace(*jsonValue) == "" {
		input, err = io.ReadAll(io.LimitReader(os.Stdin, 8<<20+1))
		if err != nil {
			fail("append-jsonl", "stdin_read_failed: "+err.Error())
		}
	}
	result, err := resultstore.Append(*path, input, kind)
	if err != nil {
		fail("append-jsonl", err.Error())
	}
	lineNumber := result.LineNumber
	writeOutput(commandOutput{
		OK:         true,
		Action:     "append-jsonl",
		Version:    version,
		Path:       *path,
		Kind:       string(kind),
		LineNumber: &lineNumber,
	})
}

func validateJSONL(args []string) {
	flags := flag.NewFlagSet("validate-jsonl", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	path := flags.String("file", "", "JSONL file")
	kindValue := flags.String("kind", "generic", "generic, events, or executions")
	if err := flags.Parse(args); err != nil {
		fail("validate-jsonl", err.Error())
	}
	if strings.TrimSpace(*path) == "" {
		fail("validate-jsonl", "file_required")
	}
	kind, err := resultstore.ParseKind(*kindValue)
	if err != nil {
		fail("validate-jsonl", err.Error())
	}
	result := resultstore.ValidateFile(*path, kind)
	lineCount := result.LineCount
	output := commandOutput{
		OK:        result.OK,
		Action:    "validate-jsonl",
		Version:   version,
		Path:      *path,
		Kind:      string(kind),
		LineCount: &lineCount,
		Errors:    result.Errors,
	}
	writeOutput(output)
	if !result.OK {
		os.Exit(1)
	}
}

func fail(action, message string) {
	writeOutput(commandOutput{
		OK:      false,
		Action:  action,
		Version: version,
		Errors:  []string{message},
	})
	os.Exit(1)
}

func writeOutput(output commandOutput) {
	encoded, err := json.Marshal(output)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println(string(encoded))
}
