package out

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintShouldNotPanicWithString(t *testing.T) {
	t.Parallel()

	cases := []string{
		"",
		"Hello, world!",
		// Other weird characters
		"😀😃😄😁",
		"\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f",
		// Chinese characters
		"你好",
		// Japanese characters
		"こんにちは",
		// Korean characters
		"안녕하세요",
		// Russian characters
		"Привет",
	}

	for _, c := range cases {
		t.Run(c, func(_ *testing.T) {
			pprint(outputPretty, nil, c)
		})
	}
}

func TestPrintShouldNotPanicWithObject(t *testing.T) {
	t.Parallel()

	cases := []any{
		"",
		nil,
		[]string{"a", "b", "c"},
		map[string]string{"a": "b", "c": "d"},
		struct {
			A string
			B int
		}{A: "a", B: 1},
	}

	for i, c := range cases {
		for _, o := range []outputT{outputPretty, outputJSON, outputYAML} {
			txt := fmt.Sprintf("case %d %s", i, o)
			t.Run(txt, func(_ *testing.T) {
				pprint(o, c, txt)
			})
		}
	}
}

func TestLogJSONKeepsStdoutToOneDocument(t *testing.T) {
	if err := SetOutput("json"); err != nil {
		t.Fatalf("set output: %v", err)
	}

	t.Cleanup(func() { _ = SetOutput("pretty") })

	stdout, stderr := captureOutAndErr(t, func() {
		Print(map[string]any{"ids": []string{"abc"}}, "Updated issue abc")
		Log("Updated issue abc")
	})

	if strings.TrimSpace(stderr) != "Updated issue abc" {
		t.Fatalf("stderr = %q", stderr)
	}

	if strings.Contains(stdout, "msg") {
		t.Fatalf("log document leaked onto stdout: %s", stdout)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout is not one JSON document: %v\n%s", err, stdout)
	}

	ids, ok := payload["ids"].([]any)
	if !ok || len(ids) != 1 || ids[0] != "abc" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestLogJSONPrintsTheMessageWhenNothingElseDoes(t *testing.T) {
	if err := SetOutput("json"); err != nil {
		t.Fatalf("set output: %v", err)
	}

	t.Cleanup(func() { _ = SetOutput("pretty") })

	stdout, stderr := captureOutAndErr(t, func() {
		Log("Tag deleted")
	})
	if stderr != "" {
		t.Fatalf("stderr = %q", stderr)
	}

	var payload Message
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stdout is not one JSON document: %v\n%s", err, stdout)
	}

	if payload.Msg != "Tag deleted" {
		t.Fatalf("msg = %q", payload.Msg)
	}
}

func TestLogPrettyPrintsTheMessage(t *testing.T) {
	if err := SetOutput("pretty"); err != nil {
		t.Fatalf("set output: %v", err)
	}

	stdout, stderr := captureOutAndErr(t, func() {
		Log("Location deleted")
	})
	if stderr != "" {
		t.Fatalf("stderr = %q", stderr)
	}

	if strings.TrimSpace(stdout) != "Location deleted" {
		t.Fatalf("stdout = %q", stdout)
	}
}

func captureOutAndErr(t *testing.T, fn func()) (string, string) {
	t.Helper()

	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}

	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	origOut := os.Stdout
	origErr := os.Stderr
	os.Stdout = outW
	os.Stderr = errW
	defer func() {
		os.Stdout = origOut
		os.Stderr = origErr
	}()

	fn()
	if err := outW.Close(); err != nil {
		t.Fatalf("close stdout: %v", err)
	}

	if err := errW.Close(); err != nil {
		t.Fatalf("close stderr: %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	if _, err := io.Copy(&outBuf, outR); err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	if _, err := io.Copy(&errBuf, errR); err != nil {
		t.Fatalf("read stderr: %v", err)
	}

	return outBuf.String(), errBuf.String()
}

func TestGetOutput(t *testing.T) {
	t.Parallel()

	cases := map[string]outputT{
		"":       outputPretty,
		"pretty": outputPretty,
		"json":   outputJSON,
		"yaml":   outputYAML,
		"yml":    outputYAML,
	}

	for in, exp := range cases {
		t.Run(in, func(t *testing.T) {
			res := getOutput(in)
			if res == nil {
				t.Errorf("expected %s, got nil", exp)
				return
			}

			if *res != exp {
				t.Errorf("expected %s, got %s", exp, *res)
			}
		})
	}
}
