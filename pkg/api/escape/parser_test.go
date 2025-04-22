package escape

import (
	"testing"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
)

func TestParseJSONOrYAMLIntegration(t *testing.T) {
	t.Parallel()
	tests := []string{
		`{"name": "test", "data":{"kind": "KUBERNETES", "parameters": {}}}`,
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			_, err := ParseJSONOrYAML([]byte(test), &v2.UpdateIntegrationRequest{})
			if err != nil {
				t.Errorf("ParseJSONOrYAML(%s) = %s", test, err.Error())
			}
		})
	}
}

func TestIsJSONOrYAML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		body     string
		expected bool
	}{
		{`{"name": "test", "age": 10}`, true},
		{`name: test
age: 10`, true},
		{`---
name: test
age: 10`, true},
	}

	for _, test := range tests {
		t.Run(test.body, func(t *testing.T) {
			result := isJSONOrYAML([]byte(test.body))
			if result != test.expected {
				t.Errorf("isJSONOrYAML(%s) = %t; expected %t", test.body, result, test.expected)
			}
		})
	}
}
