package logs

import (
	"common/config"
	"github.com/charmbracelet/log"
	"os"
	"time"
)

var logger *log.Logger

func InitLog(appName string) {
	logger = log.New(os.Stderr)
	if config.Conf.Log.Level == "DEBUG" {
		logger.SetLevel(log.DebugLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}
	logger.SetPrefix(appName)
	logger.SetReportTimestamp(true)
	logger.SetTimeFormat(time.DateTime)
}

func Fatal(format string, v ...any) {
	if len(v) == 0 {
		logger.Fatal(format)
	} else {
		logger.Fatalf(format)
	}
}

func Info(format string, v ...any) {
	if len(v) == 0 {
		logger.Info(format)
	} else {
		logger.Infof(format)
	}
}

func Debug(format string, v ...any) {
	if len(v) == 0 {
		logger.Debug(format)
	} else {
		logger.Debugf(format)
	}
}

func Warn(format string, v ...any) {
	if len(v) == 0 {
		logger.Warn(format)
	} else {
		logger.Warnf(format)
	}
}

func Error(format string, v ...any) {
	if len(v) == 0 {
		logger.Error(format)
	} else {
		logger.Errorf(format)
	}
}
