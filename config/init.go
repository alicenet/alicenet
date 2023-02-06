package config

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
)

// setDefaultCommandIfNonePresent to be able to run a node if none command is present.
func setDefaultCommandIfNonePresent(defaultCommand *cobra.Command, logger *logrus.Logger) {
	if len(os.Args) != 1 {
		return
	}

	// Adding the `node` command to args.
	os.Args = append([]string{os.Args[0], defaultCommand.Use})

	// Setting te default --config location if it is not present in command options.
	if Configuration.ConfigurationFileName == "" {
		homeDirectory, err := os.UserHomeDir()
		if err != nil {
			logger.Fatalf("failed to obtain user's home directory with error: %v", err)
		}

		Configuration.ConfigurationFileName = homeDirectory + defaultConfigLocation
	}
}

func initialiseDataFromConfigFile(logger *logrus.Logger, options []*cobra.Command, heirachy map[*cobra.Command]*cobra.Command) {

	// Convert option abstraction into concrete settings for Cobra and Viper
	for _, c := range options {
		//cFlags := c.PersistentFlags() // just a convenience thing

		if c.Run != nil {
			c.Run = runner(c.Run)
		}

		if parentCommand, present := heirachy[c]; present {
			//cFlags := c.Flags()
			parentCommand.AddCommand(c)
		}
	}

	// This has to be registered prior to root command execute. Cobra executes this first thing when executing.
	cobra.OnInitialize(func() {
		// Read the config file
		logger.Infof("configuration file set is %s", Configuration.ConfigurationFileName)
		file, err := os.Open(Configuration.ConfigurationFileName)
		if err == nil {
			bs, err := ioutil.ReadAll(file)
			if err == nil {
				reader := bytes.NewReader(bs)
				viper.SetConfigType("toml") // TODO: Set config type based on file extension. Viper supports more than toml.
				err := viper.ReadConfig(reader)
				if err != nil {
					logger.Warnf("Reading config failed:%q", err)
				}
			} else {
				logger.Warnf("Reading file failed:%q", err)
			}
		} else {
			logger.Debugf("Opening file failed: %q", err)
		}
		// Overwrites config value(s) if user passes it as cmd line arg.
		overWriteUponUserCall(logger, options)

	})
}
