package utils

import "github.com/sirupsen/logrus"

// LogAndEat - utils function to log an error.
func LogAndEat(logger *logrus.Logger, err error) {
	if err != nil {
		logger.Error(err)
	}
}
