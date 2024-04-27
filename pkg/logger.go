package pkg

import (
	"github.com/pnsafonov/pind/pkg/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
)

var (
	LogTextFormatter = &log.TextFormatter{
		DisableQuote: true, // don't print \n char, print linebreak
	}
)

func initLogger(configLog *config.Log, isService bool) {
	wr := io.MultiWriter()
	count := 0

	// rotator logger
	if configLog.RotatorEnabled {
		configRotator := configLog.Rotator

		// create dir for logs
		// logrus can do it
		//dir := filepath.Dir(configRotator.Filename)
		//if !os_utils.Exists(dir) {
		//	err := os.MkdirAll(dir, 0744)
		//	if err != nil {
		//		log.Errorf("initLogger, os.MkdirAll err = %v, dir = %s", err, dir)
		//	}
		//}

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

	if isService {
		LogTextFormatter.DisableColors = true
	}

	log.SetFormatter(LogTextFormatter)
	log.SetOutput(wr)
	log.SetLevel(configLog.Level)

	return
}

func InitConsoleLogger() {
	log.SetFormatter(LogTextFormatter)
	log.SetOutput(os.Stderr)
}
