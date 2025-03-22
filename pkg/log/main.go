package log

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	log          *logrus.Logger
	globalBuffer *logBuffer
)

func init() {
	log = logrus.New()
	log.SetOutput(os.Stderr)
	log.SetLevel(logrus.WarnLevel)
	globalBuffer = newLogBuffer(1000)
}

func SetLevel(level logrus.Level) {
	log.SetLevel(level)
}

func doLog(level logrus.Level, format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	log.Logf(level, format, args...)
	globalBuffer.Ingest(LogItem{Level: level, Message: line})
}

func Trace(format string, args ...any) { doLog(logrus.TraceLevel, format, args...) }
func Debug(format string, args ...any) { doLog(logrus.DebugLevel, format, args...) }
func Info(format string, args ...any)  { doLog(logrus.InfoLevel, format, args...) }
func Warn(format string, args ...any)  { doLog(logrus.WarnLevel, format, args...) }
func Error(format string, args ...any) { doLog(logrus.ErrorLevel, format, args...) }

// AddHook adds a named hook to the logger.
// The hook will be called with the level and message of the log entry.
// This function returns a unique ID for the hook, which can be used to remove the hook later.
func AddHook(name string, callback func(LogItem)) {
	globalBuffer.AddHook(name, callback)
}

// RemoveHook removes a hook from the logger.
// The hook will no longer be called with the level and message of the log entry.
func RemoveHook(name string) {
	globalBuffer.RemoveHook(name)
}
