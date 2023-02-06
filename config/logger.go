package config

import (
	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"
)

func setLogger(name string, level string) {
	lgr := logging.GetLogger(name)
	switch level {
	case "trace":
		lgr.SetLevel(logrus.TraceLevel)
	case "debug":
		lgr.SetLevel(logrus.DebugLevel)
	case "info":
		lgr.SetLevel(logrus.InfoLevel)
	case "warn":
		lgr.SetLevel(logrus.WarnLevel)
	case "error":
		lgr.SetLevel(logrus.ErrorLevel)
	case "fatal":
		lgr.SetLevel(logrus.FatalLevel)
	case "panic":
		lgr.SetLevel(logrus.PanicLevel)
	default:
		lgr.SetLevel(logrus.InfoLevel)
	}
}
