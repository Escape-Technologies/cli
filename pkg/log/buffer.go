package log

import (
	"github.com/sirupsen/logrus"
)

type logBuffer struct {
	logs chan struct {
		level logrus.Level
		log   string
	}
	hooks      map[string]func(logrus.Level, string)
}

func newLogBuffer(bufferSize int) *logBuffer {
	return &logBuffer{
		logs: make(chan struct {
			level logrus.Level
			log   string
		}, bufferSize),
		hooks: make(map[string]func(logrus.Level, string)),
	}
}

func (b *logBuffer) GetLog() string {
	select {
	case log := <-b.logs:
		return log
	default:
		return ""
	}
}

func (b *logBuffer) GetLogs() []string {
	logs := make([]string, 0, len(b.logs))
	for {
		log := b.GetLog()
		if log == "" {
			break
		}
		logs = append(logs, log)
	}
	return logs
}

func (b *logBuffer) AddLog(level logrus.Level, log string) {
	for _, hook := range b.hooks {
		hook(level, log)
	}
	select {
	case b.logs <- struct {
		level logrus.Level
		log   string
	}{level: level, log: log}:
	default:
	}
}


func (b *logBuffer) Close() {
	close(b.logs)
}


func (b *logBuffer) AddHook(name string, callback func(logrus.Level, string)) {
	b.hooks[name] = callback
	for log := range b.logs {
		callback(log.level, log.log)
	}
}

func (b *logBuffer) RemoveHook(name string) {
	delete(b.hooks, name)
}
