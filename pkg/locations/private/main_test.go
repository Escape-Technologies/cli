package private

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestIsDialTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "deadline exceeded",
			err:  context.DeadlineExceeded,
			want: true,
		},
		{
			name: "wrapped deadline exceeded",
			err:  fmt.Errorf("getConn: %w", context.DeadlineExceeded),
			want: true,
		},
		{
			name: "context deadline exceeded message",
			err:  errors.New("failed to get conn: context deadline exceeded"),
			want: true,
		},
		{
			name: "i/o timeout message",
			err:  errors.New("failed to dial target: i/o timeout"),
			want: true,
		},
		{
			name: "connection refused",
			err:  errors.New("connection refused"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isDialTimeout(tt.err); got != tt.want {
				t.Fatalf("isDialTimeout(%q) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
