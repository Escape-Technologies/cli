package bilogs

import (
	"context"

	"github.com/Escape-Technologies/cli/pkg/log"
	"golang.org/x/crypto/ssh"
)

func bufferLogProcessor(ctx context.Context, ch ssh.Channel, buffer *LogBuffer) {
	for {
		select {
		case logMsg, ok := <-buffer.GetLogs():
			if !ok {
				return
			}
			err := logSender(ch, logMsg)
			if err != nil {
				log.Error("failed to send log: %v", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}


func logSender(ch ssh.Channel, logMsg string) error {
	_, err := ch.SendRequest("log", false, []byte(logMsg))
	if err != nil {
		log.Error("failed to send log request: %v", err)
		return err
	}
	return nil
}