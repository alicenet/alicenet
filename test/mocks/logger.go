package mocks

import (
	"io"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

func NewMockLogger() *logrus.Logger {
	logger := logrus.New()

	testDebug, _ := strconv.ParseBool(os.Getenv("TEST_DEBUG"))
	if !testDebug {
		logger.SetOutput(io.Discard)
	}

	return logger
}
