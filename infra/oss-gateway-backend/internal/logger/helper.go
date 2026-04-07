package logger

import (
	"os"
	"oss-gateway/internal/config"

	"github.com/sirupsen/logrus"
)

func NewLogger(app *config.AppConfig) *logrus.Entry {
	log := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(app.LogConfig.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// Set log format
	if app.LogConfig.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// Set output
	log.SetOutput(os.Stdout)

	return log.WithFields(logrus.Fields{
		"service": app.CommonConfig.Name,
	})
}
