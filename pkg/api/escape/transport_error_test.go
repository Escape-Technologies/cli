package escape

import (
	"context"
	"net"
	"testing"
)

// closedPortResourceID is only used to build a request the transport then
// refuses. The generated client returns a nil *http.Response for that failure.
const closedPortResourceID = "00000000-0000-0000-0000-000000000001"

// pointClientAtClosedPort aims the v3 client at a port that accepts nothing,
// so Execute fails before any HTTP response exists.
func pointClientAtClosedPort(t *testing.T) {
	t.Helper()

	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	t.Setenv("ESCAPE_API_URL", "http://"+address)
	t.Setenv("ESCAPE_API_KEY", "00000000-0000-0000-0000-000000000000")
}

// TestTransportFailureReturnsErrorInsteadOfPanicking covers wrappers that
// used to read StatusCode on the nil response a refused connection produces.
func TestTransportFailureReturnsErrorInsteadOfPanicking(t *testing.T) {
	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "CancelScan",
			call: func() error {
				return CancelScan(context.Background(), closedPortResourceID)
			},
		},
		{
			name: "UpdateAsset",
			call: func() error {
				_, err := UpdateAsset(context.Background(), closedPortResourceID, UpdateAssetInput{})
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pointClientAtClosedPort(t)

			var callErr error
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						t.Fatalf("panic: %v", recovered)
					}
				}()

				callErr = test.call()
			}()

			if callErr == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
