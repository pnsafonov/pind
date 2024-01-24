package pkg

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"pind/pkg/config"
)

func initLogger(configLog *config.Log) {
	wr := io.MultiWriter()
	count := 0

	// rotator logger
	if configLog.RotatorEnabled {
		configRotator := configLog.Rotator

		rotator0 := &lumberjack.Logger{
			Filename:   configRotator.Filename,
			MaxSize:    configRotator.MaxSize,
			MaxBackups: configRotator.MaxBackups,
			MaxAge:     configRotator.MaxAge,
			Compress:   false,
			LocalTime:  configRotator.LocalTime,
		}

		wr = io.MultiWriter(wr, rotator0)
		count++
	}

	// if no logger set default to stderr
	if configLog.StdErrEnabled || count == 0 {
		wr = io.MultiWriter(wr, os.Stderr)
	}

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(wr)
	log.SetLevel(configLog.Level)

	return
}

func InitConsoleLogger() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stderr)
}
