package config

import (
	"github.com/alicenet/alicenet/logging"
	"github.com/spf13/cobra"
	"reflect"
	"strings"
)

// Runner wraps a cobra command's Run() and sets up loggers first.
func runner(commandRun func(*cobra.Command, []string)) func(*cobra.Command, []string) {
	logger := logging.GetLogger("main")
	return func(a *cobra.Command, b []string) {
		loggingLevels := Configuration.Logging
		llr := reflect.ValueOf(loggingLevels)
		for i := 0; i < llr.NumField(); i++ {
			logName := strings.ToLower(llr.Type().Field(i).Name)
			logLevel := strings.ToLower(llr.Field(i).String())
			if logLevel == "" {
				logLevel = "info"
			}
			if logLevel != "info" {
				logger.Infof("Setting log level for '%v' to '%v'", logName, logLevel)
			}
			setLogger(logName, logLevel)
		}
		// backwards compatibility
		if len(Configuration.LoggingLevels) > 0 {
			loggers := strings.Split(Configuration.LoggingLevels, ",")
			for _, levelSetting := range loggers {
				settingComponent := strings.Split(levelSetting, "=")
				if len(settingComponent) != 2 {
					logger.Fatalf("Malformed log level setting %q", levelSetting)
				}
				logger.Infof("Overwriting log level for '%v' to '%v'", settingComponent[0], settingComponent[1])
				setLogger(settingComponent[0], settingComponent[1])
			}
		}
		commandRun(a, b)
	}
}
