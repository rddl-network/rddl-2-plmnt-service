package service

import (
	"fmt"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
)

type AppLogger struct {
	logger log.Logger
}

func getLogger() AppLogger {
	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	// Set log level
	switch config.GetConfig().LogLevel {
	case "debug":
		logger = level.NewFilter(logger, level.AllowDebug())
	case "info":
		logger = level.NewFilter(logger, level.AllowInfo())
	case "warn":
		logger = level.NewFilter(logger, level.AllowWarn())
	default:
		logger = level.NewFilter(logger, level.AllowError())
	}

	return AppLogger{logger}
}

func (al AppLogger) Error(kevals ...interface{}) {
	err := level.Error(al.logger).Log(kevals...)
	if err != nil {
		fmt.Println(err)
	}
}

func (al AppLogger) Warn(kevals ...interface{}) {
	err := level.Warn(al.logger).Log(kevals...)
	if err != nil {
		fmt.Println(err)
	}
}

func (al AppLogger) Info(kevals ...interface{}) {
	err := level.Info(al.logger).Log(kevals...)
	if err != nil {
		fmt.Println(err)
	}
}

func (al AppLogger) Debug(kevals ...interface{}) {
	err := level.Debug(al.logger).Log(kevals...)
	if err != nil {
		fmt.Println(err)
	}
}
