package bilogs

import (
	"context"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var globalBuffer *LogBuffer

func init() {
	globalBuffer = NewLogBuffer(1000)
	log.AddHook(sendLogBuilder(globalBuffer))
}

func sendLogBuilder(buffer *LogBuffer) func(logrus.Level, string) {
	return func(_ logrus.Level, line string) {
		buffer.AddLog(line)
	}
}

func StartMonitoring(ctx context.Context, client *ssh.Client) {
	ch, err := openEscapeChannel(ctx, client)
	if err != nil {
		return
	}
	
	go healthTicker(ctx, ch)
	go bufferLogProcessor(ctx, ch, globalBuffer)
}








