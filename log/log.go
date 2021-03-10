package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	DebugLevel = logrus.DebugLevel
	InfoLevel  = logrus.InfoLevel
	WarnLevel  = logrus.WarnLevel
	ErrorLevel = logrus.ErrorLevel
)

func init() {
	format := new(logrus.TextFormatter)
	format.FullTimestamp = true
	format.TimestampFormat = "2006-01-02 15:04:05"
	logrus.SetFormatter(format)
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

func SetLevel(level logrus.Level) {
	logrus.SetLevel(level)
}

func IsDebug() bool {
	if logrus.GetLevel() == logrus.DebugLevel {
		return true
	}
	return false
}

func Print(level logrus.Level, msg string) {
	switch level {
	case logrus.DebugLevel:
		logrus.Debug(msg)
	case logrus.InfoLevel:
		logrus.Info(msg)
	case logrus.WarnLevel:
		logrus.Warn(msg)
	case logrus.ErrorLevel:
		logrus.Error(msg)
	}
}
