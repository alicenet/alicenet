package logging

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var loggers = map[string]*logrus.Logger{}
var loggersMutex = &sync.Mutex{}

//LogFormatter applies consistent formatting to every message
type LogFormatter struct {
	frm func(*logrus.Entry) ([]byte, error)
}

//Format satisfies logrus' Format interface while staying flexible
func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if f.frm != nil {
		return f.frm(entry)
	}

	genericFormatter := logrus.TextFormatter{PadLevelText: true, TimestampFormat: "1-2|15:04:05.000", FullTimestamp: true}

	return genericFormatter.Format(entry)
}

//LogWriter struct used to provide an io.Writer
type LogWriter struct {
	logger *logrus.Logger
	level  logrus.Level
}

func (logWriter *LogWriter) Write(p []byte) (n int, err error) {
	logWriter.logger.Log(logWriter.level, strings.TrimRight(string(p), "\n"))
	return len(p), nil
}

//GetLogger either returns an existing logger for package specified or creates a new one
func GetLogger(pkgName string) *logrus.Logger {
	loggersMutex.Lock()
	logger, exists := loggers[pkgName]
	loggersMutex.Unlock()

	if !exists {
		formatter := &LogFormatter{frm: func(entry *logrus.Entry) ([]byte, error) {
			defaultFormat, e := (&logrus.TextFormatter{PadLevelText: true, TimestampFormat: "1-2|15:04:05.000", FullTimestamp: true}).Format(entry)

			pkg := fmt.Sprintf("%-10s ", pkgName)

			line := bytes.Join([][]byte{
				[]byte(pkg), defaultFormat},
				[]byte(" "))

			return line, e
		}}

		logger = logrus.New()
		logger.SetFormatter(formatter)

		logger.SetLevel(logrus.DebugLevel)

		loggersMutex.Lock()
		loggers[pkgName] = logger
		loggersMutex.Unlock()
	}

	return logger
}

//GetLogWriter returns an io.Writer that maps to the named logger at the specified level
func GetLogWriter(pkgName string, level logrus.Level) *LogWriter {
	logger := GetLogger(pkgName)

	return &LogWriter{logger, level}
}

// GetKnownLoggers returns all loggers currently configured
func GetKnownLoggers() []*logrus.Logger {
	ret := make([]*logrus.Logger, 0)
	for _, logger := range loggers {
		ret = append(ret, logger)
	}

	return ret
}
