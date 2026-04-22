package escape

import (
	"errors"
	"fmt"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

// errNotAConflict is returned by extractConflict when the error chain does
// not carry a typed Conflict model. Callers must treat this as a real error
// and never fall back to using an empty instance id.
var errNotAConflict = errors.New("error is not a conflict")

func extractConflict(err error) (string, error) {
	if err == nil {
		return "", errNotAConflict
	}
	var oapiErr *v3.GenericOpenAPIError
	if errors.As(err, &oapiErr) {
		if conflict, ok := oapiErr.Model().(v3.Conflict); ok {
			id := conflict.GetInstanceId()
			if id == "" {
				return "", fmt.Errorf("conflict response missing instanceId: %w", err)
			}
			return id, nil
		}
	}
	return "", errNotAConflict
}
