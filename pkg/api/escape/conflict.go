package escape

import (
	"errors"

	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
)

func extractConflict(err error) (string, error) {
	if err == nil {
		return "", nil
	}
	if oapiErr, ok := err.(*v2.GenericOpenAPIError); ok {
		if conflict, ok := oapiErr.Model().(v2.CreateLocation409Response); ok {
			return conflict.InstanceId, nil
		}
	}
	s, newErr := extractConflict(errors.Unwrap(err))
	if s != "" {
		return s, nil
	}
	return "", newErr
}
