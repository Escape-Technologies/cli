package escape

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

type parsable interface {
	UnmarshalJSON(data []byte) (err error)
}

func parseJSONOrYAML[T parsable](body []byte, v T) (T, error) {
	finalBody, err := yamlToJSON(body)
	if err != nil {
		finalBody = body
	}
	err = v.UnmarshalJSON(finalBody)
	if err != nil {
		return v, fmt.Errorf("file is neither json nor yaml: %w", err)
	}
	return v, nil
}

func yamlToJSON(body []byte) ([]byte, error) {
	var v map[string]any
	err := yaml.Unmarshal(body, &v)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal yaml: %w", err)
	}
	jsonBody, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal yaml to json: %w", err)
	}
	return jsonBody, nil
}

func isJSONOrYAML(body []byte) bool {
	if json.Valid(body) {
		return true
	}
	return yaml.Unmarshal(body, &map[string]any{}) == nil
}
