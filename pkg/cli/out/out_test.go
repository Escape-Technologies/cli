package out

import (
	"fmt"
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
		t.Run(c, func(t *testing.T) {
			print(outputPretty, nil, c)
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
			t.Run(txt, func(t *testing.T) {
				print(o, c, txt)
			})
		}
	}
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
