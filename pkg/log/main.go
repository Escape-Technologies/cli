package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	log *logrus.Logger
)

func init() {
	log = logrus.New()
	log.SetOutput(os.Stderr)
	log.SetLevel(logrus.WarnLevel)
}

func SetLevel(level logrus.Level) {
	log.SetLevel(level)
}

func Trace(format string, args ...interface{}) { log.Tracef(format, args...) }
func Debug(format string, args ...interface{}) { log.Debugf(format, args...) }
func Info(format string, args ...interface{})  { log.Infof(format, args...) }
func Warn(format string, args ...interface{})  { log.Warnf(format, args...) }
func Error(format string, args ...interface{}) { log.Errorf(format, args...) }
