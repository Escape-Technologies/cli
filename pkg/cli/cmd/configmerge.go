package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// mergeProfileConfiguration deep-merges a partial profiles update-configuration
// body into the profile's current configuration and returns the JSON to send.
//
// The update API replaces the whole configuration, so the result is always a
// complete document: keys the patch omits stay as they are now. Objects merge
// recursively. Arrays, scalars, and JSON null replace the current value
// (arrays are not concatenated; null clears a field the caller named).
// Other top-level keys on the patch are preserved.
func mergeProfileConfiguration(current v3.GetProfile200ResponseConfiguration, patch []byte) ([]byte, error) {
	currentRaw, err := json.Marshal(current)
	if err != nil {
		return nil, fmt.Errorf("read current configuration: %w", err)
	}

	var base any
	if err := json.Unmarshal(currentRaw, &base); err != nil {
		return nil, fmt.Errorf("read current configuration: %w", err)
	}

	var patchDoc map[string]any
	if err := json.Unmarshal(patch, &patchDoc); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if patchDoc == nil {
		return nil, errors.New(`merge requires a "configuration" object`)
	}

	patchConfig, ok := patchDoc["configuration"]
	if !ok {
		return nil, errors.New(`merge requires a "configuration" object`)
	}

	if _, isObject := patchConfig.(map[string]any); !isObject {
		return nil, errors.New(`"configuration" must be a JSON object`)
	}

	patchDoc["configuration"] = mergeJSON(base, patchConfig)
	out, err := json.Marshal(patchDoc)
	if err != nil {
		return nil, fmt.Errorf("encode merged configuration: %w", err)
	}

	return out, nil
}

// mergeJSON deep-merges patch into base.
// Two objects merge key by key. Anything else, including arrays and null,
// replaces base. Lists such as users and hotstart URLs are replaced rather
// than concatenated so a patch does not duplicate entries.
func mergeJSON(base, patch any) any {
	baseObj, baseOK := base.(map[string]any)
	patchObj, patchOK := patch.(map[string]any)
	if !baseOK || !patchOK {
		return patch
	}

	out := make(map[string]any, len(baseObj)+len(patchObj))
	for key, value := range baseObj {
		out[key] = value
	}

	for key, value := range patchObj {
		if existing, ok := out[key]; ok {
			out[key] = mergeJSON(existing, value)
			continue
		}

		out[key] = value
	}

	return out
}
