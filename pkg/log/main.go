package log

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

type Hook struct {
	Name     string
	Callback func(logrus.Level, string)
}

var (
	log        *logrus.Logger
	hooks      map[string]Hook
	globalBuffer *logBuffer
)

func init() {
	log = logrus.New()
	log.SetOutput(os.Stderr)
	log.SetLevel(logrus.WarnLevel)
	hooks = make(map[string]Hook)
	globalBuffer = newLogBuffer(1000)
}

func SetLevel(level logrus.Level) {
	log.SetLevel(level)
}

func doLog(level logrus.Level, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	log.Logf(level, format, args...)
	globalBuffer.AddLog(level, line)
	for _, hook := range hooks {
		hook.Callback(level, line)
	}
}

func Trace(format string, args ...interface{}) { doLog(logrus.TraceLevel, format, args...) }
func Debug(format string, args ...interface{}) { doLog(logrus.DebugLevel, format, args...) }
func Info(format string, args ...interface{})  { doLog(logrus.InfoLevel, format, args...) }
func Warn(format string, args ...interface{})  { doLog(logrus.WarnLevel, format, args...) }
func Error(format string, args ...interface{}) { doLog(logrus.ErrorLevel, format, args...) }


// AddHook adds a named hook to the logger.
// The hook will be called with the level and message of the log entry.
// This function returns a unique ID for the hook, which can be used to remove the hook later.
func AddHook(name string, callback func(logrus.Level, string)) {
	globalBuffer.AddHook(name, callback)
}

// RemoveHook removes a hook from the logger.
// The hook will no longer be called with the level and message of the log entry.
func RemoveHook(name string) {
	globalBuffer.RemoveHook(name)
}
