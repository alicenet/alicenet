package logging

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetLoggerNew(t *testing.T) {
	logNumber := len(GetKnownLoggers())

	_ = GetLogger("TestGetLoggerNew")

	assert.Equal(t, logNumber+1, len(GetKnownLoggers()), "There should be 1 more logger.")
}

func TestGetLoggerDefaultLevel(t *testing.T) {
	logger := GetLogger("TestGetLoggerDefaultLevel")
	assert.Equal(t, logrus.DebugLevel, logger.GetLevel(), "Default level is DEBUG")
}

func TestGetLoggerPreserveLevel(t *testing.T) {
	const loggerName = "TestGetLoggerPreserveLevel"

	logger := GetLogger(loggerName)
	assert.Equal(t, logrus.DebugLevel, logger.GetLevel(), "Default level is DEBUG")
	logger.SetLevel(logrus.DebugLevel)

	newLogger := GetLogger(loggerName)
	assert.Equal(t, logrus.DebugLevel, newLogger.GetLevel(), "Level was set to DEBUG")
}

func TestGetLogWriter(t *testing.T) {
	const loggerName = "TestGetLogWriter"
	logger := GetLogger(loggerName)
	logger.SetLevel(logrus.ErrorLevel)

	out := GetLogWriter(loggerName, logrus.WarnLevel)

	// TODO: Probably have to capture stdout to really automate this test
	fmt.Fprintf(out, "Should not be visible even with a test -v ./...")
}
