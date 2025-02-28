package log

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

type Hook func(logrus.Level, string)

var (
	log        *logrus.Logger
	hooks      map[int32]Hook
	currHookID atomic.Int32
)

func init() {
	log = logrus.New()
	log.SetOutput(os.Stderr)
	log.SetLevel(logrus.WarnLevel)
	hooks = make(map[int32]Hook)
}

func SetLevel(level logrus.Level) {
	log.SetLevel(level)
}

func doLog(level logrus.Level, format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	log.Logf(level, format, args...)
	for _, hook := range hooks {
		hook(level, line)
	}
}

func Trace(format string, args ...interface{}) { doLog(logrus.TraceLevel, format, args...) }
func Debug(format string, args ...interface{}) { doLog(logrus.DebugLevel, format, args...) }
func Info(format string, args ...interface{})  { doLog(logrus.InfoLevel, format, args...) }
func Warn(format string, args ...interface{})  { doLog(logrus.WarnLevel, format, args...) }
func Error(format string, args ...interface{}) { doLog(logrus.ErrorLevel, format, args...) }

// AddHook adds a hook to the logger.
// The hook will be called with the level and message of the log entry.
// This function returns a unique ID for the hook, which can be used to remove the hook later.
func AddHook(hook Hook) int32 {
	id := currHookID.Add(1)
	hooks[id] = hook
	return id
}

// RemoveHook removes a hook from the logger.
// The hook will no longer be called with the level and message of the log entry.
func RemoveHook(id int32) {
	delete(hooks, id)
}
