package util

import (
	"github.com/sirupsen/logrus"
	"os"
)

// InitLogger initializes the global logrus logger based on the development mode flag.
func InitLogger(isDevMode bool) {
	// Set the output to standard out.
	logrus.SetOutput(os.Stdout)

	if isDevMode {
		// In dev mode, show all logs including debug messages.
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
		logrus.Debug("Logger initialized in DEBUG mode.")
	} else {
		// In production mode, only show info level and above.
		logrus.SetLevel(logrus.InfoLevel)
		// Use a simpler formatter for production.
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
}
