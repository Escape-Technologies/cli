package private

import (
	"context"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func healthTicker(ctx context.Context, client *ssh.Client) {
	ch, _, err := client.OpenChannel("escape_health_channel", nil)
	if err != nil {
		log.Error("failed to open health channel: %v", err)
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		<-ctx.Done()
		ticker.Stop()
	}()
	for range ticker.C {
		ch.SendRequest("health", false, nil)
	}
}

func sendLogBuilder(ch ssh.Channel) func(logrus.Level, string) {
	return func(_ logrus.Level, line string) {
		ch.Write([]byte(line))
	}
}

func startMonitoring(ctx context.Context, client *ssh.Client) {
	healthTicker(ctx, client)

	// // ch, reqs, err := client.OpenChannel("log", nil)
	// ch, _, err := client.OpenChannel("log", nil)

	// if err != nil {
	// 	log.Error("failed to open channel: %v", err)
	// 	return
	// }
	// hookID := log.AddHook(sendLogBuilder(ch))
	// defer log.RemoveHook(hookID)
	// log.Info("Log channel opened")
	// ch.SendRequest("shell", true, nil)
	// // go logRequests(reqs)
	// go func() {
	// 	<-ctx.Done()
	// 	ch.Close()
	// }()
}
