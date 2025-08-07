package escape

import (
	"errors"

	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
)

func extractConflict(err error) (string, error) {
	if err == nil {
		return "", nil
	}
	if oapiErr, ok := err.(*v3.GenericOpenAPIError); ok {
		if conflict, ok := oapiErr.Model().(v3.IgnoreScan409Response); ok {
			return conflict.GetInstanceId(), nil
		}
	}
	s, newErr := extractConflict(errors.Unwrap(err))
	if s != "" {
		return s, nil
	}
	return "", newErr
}
