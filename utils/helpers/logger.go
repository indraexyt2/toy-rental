package helpers

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Logger = logrus.New()

func SetupLogger(isProduction bool) {
	if isProduction {
		Logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
	}

	Logger.SetOutput(os.Stdout)

	if isProduction {
		Logger.SetLevel(logrus.InfoLevel)
	} else {
		Logger.SetLevel(logrus.DebugLevel)
	}

	Logger.Info("Setup logger with logrus")
}
