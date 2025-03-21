package private

import (
	"context"

	"golang.org/x/crypto/ssh"
)

func StartMonitoring(ctx context.Context, client *ssh.Client) {
	ch, err := openEscapeChannel(ctx, client)
	if err != nil {
		return
	}
	
	go healthTicker(ctx, ch)
}








