package logging_test

import (
	"fmt"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetValidLoggers(t *testing.T) {
	n := len(constants.ValidLoggers)
	assert.True(t, n > 0)

	for idx := 0; idx < n; idx++ {
		logger := logging.GetLogger(constants.ValidLoggers[idx])
		assert.NotNil(t, logger)
	}

}

func TestGetInvalidLogger(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r, "Should have panic'ed when requesting a bad logger")
	}()
	logger := logging.GetLogger("Foo")
	logger.Infof("should never happen")
}

func TestGetLoggerDefaultLevel(t *testing.T) {
	loggerName := constants.ValidLoggers[0]
	logger := logging.GetLogger(loggerName)
	assert.Equal(t, logrus.InfoLevel, logger.GetLevel(), "Default level is INFO")
}

func TestLog(t *testing.T) {
	defer func() {
		r := recover()
		assert.Nil(t, r, "No panics during/after logging")
	}()
	logger := logging.GetLogger(constants.ValidLoggers[0])
	logger.Infof("foo")
}

func TestGetLoggerPreserveLevel(t *testing.T) {
	loggerName := constants.ValidLoggers[0]

	logger := logging.GetLogger(loggerName)
	assert.Equal(t, logrus.InfoLevel, logger.GetLevel(), "Default should be INFO")

	logger.SetLevel(logrus.DebugLevel)

	newLogger := logging.GetLogger(loggerName)
	assert.Equal(t, logrus.DebugLevel, newLogger.GetLevel(), "Level was set to DEBUG")
}

func TestGetLogWriter(t *testing.T) {
	loggerName := constants.ValidLoggers[0]
	logger := logging.GetLogger(loggerName)
	logger.SetLevel(logrus.ErrorLevel)

	out := logging.GetLogWriter(loggerName, logrus.WarnLevel)

	// TODO: Probably have to capture stdout to really automate this test
	fmt.Fprintf(out, "Should not be visible even with a test -v ./...")
}

func TestGetKnownLoggers(t *testing.T) {
	known := logging.GetKnownLoggers()
	assert.Equal(t, len(constants.ValidLoggers), len(known))
}
