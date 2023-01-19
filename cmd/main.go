package main

import (
	"bytes"
	"github.com/alicenet/alicenet/cmd/bootnode"
	"github.com/alicenet/alicenet/cmd/ethkey"
	"github.com/alicenet/alicenet/cmd/node"
	"github.com/alicenet/alicenet/cmd/root"
	"github.com/alicenet/alicenet/cmd/utils"
	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Variables set by goreleaser process: https://goreleaser.com/cookbooks/using-main.version.
var (
	// Version from git tag.
	version               = "dev"
	defaultConfigLocation = "/.alicenet/mainnet/config.toml"
)

// Runner wraps a cobra command's Run() and sets up loggers first.
func runner(commandRun func(*cobra.Command, []string)) func(*cobra.Command, []string) {
	logger := logging.GetLogger("main")
	return func(a *cobra.Command, b []string) {
		loggingLevels := config.Configuration.Logging
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
		if len(config.Configuration.LoggingLevels) > 0 {
			loggers := strings.Split(config.Configuration.LoggingLevels, ",")
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

func main() {
	logger := logging.GetLogger("main")
	logger.SetLevel(logrus.InfoLevel)

	config.Configuration.Version = version

	// Root for all commands
	rootCmd := root.Cmd
	utilsCmd := utils.Command
	nodeCmd := node.Command
	bootNodeCmd := bootnode.Command
	ethkeyCmd := ethkey.Generate
	rootCmd.AddCommand(
		utilsCmd,
		nodeCmd,
		bootNodeCmd,
		ethkeyCmd,
	)
	options := [5]*cobra.Command{rootCmd, utilsCmd, nodeCmd, bootNodeCmd, ethkeyCmd}
	// If none command and option are present, the `node` command with the default --config option will be executed.
	setDefaultCommandIfNonePresent(node.Command, logger)

	// This has to be registered prior to root command execute. Cobra executes this first thing when executing.
	cobra.OnInitialize(func() {
		// Read the config file
		logger.Infof("configuration file set is %s", config.Configuration.ConfigurationFileName)
		file, err := os.Open(config.Configuration.ConfigurationFileName)
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

		/* The logic here feels backwards to me but it isn't.
		Command line flags aren't set till this func returns.
		So we set flags from config file here and when func returns the command line will overwrite.
		*/
		for _, cmd := range options {
			// Find all the flags

			if cmd.Run != nil {
				cmd.Run = runner(cmd.Run)
			}

			cmd.Flags().VisitAll(func(flag *pflag.Flag) {

				// -help defined by pflag internals and will not parse correctly.
				if flag.Name == "help" {
					return
				}
				value := viper.Get(flag.Name)
				var err error
				if value != nil {
					if flag.Value.Type() == "string" {
						err = flag.Value.Set(value.(string))
					} else if flag.Value.Type() == "bool" {
						err = flag.Value.Set(strconv.FormatBool(value.(bool)))
					} else if flag.Value.Type() == "int" {
						err = flag.Value.Set(strconv.FormatInt(value.(int64), 10))
					} else if flag.Value.Type() == "uint64" {
						err = flag.Value.Set(strconv.FormatUint(value.(uint64), 10))
					} else if flag.Value.Type() == "duration" {
						duration := value.(time.Duration)
						s := duration.String()
						if strings.HasSuffix(s, "m0s") {
							s = s[:len(s)-2]
						}
						if strings.HasSuffix(s, "h0m") {
							s = s[:len(s)-2]
						}
						err = flag.Value.Set(s)
					}
				}

				if err != nil {
					logger.Warnf("Setting flag %q failed:%q", flag.Name, err)
				}
			})
		}

		logger.Debugf("onInitialize() -- Configuration:%v", config.Configuration)
	})

	// Really start application here
	err := rootCmd.Execute()
	if err != nil {
		logger.Fatalf("Execute() failed:%q", err)
	}
	logger.Debugf("main() -- Configuration:%v", config.Configuration.Ethereum)
}

// setDefaultCommandIfNonePresent to be able to run a node if none command is present.
func setDefaultCommandIfNonePresent(defaultCommand *cobra.Command, logger *logrus.Logger) {
	if len(os.Args) != 1 {
		return
	}

	// Adding the `node` command to args.
	os.Args = append([]string{os.Args[0], defaultCommand.Use})

	// Setting te default --config location if it is not present in command options.
	if config.Configuration.ConfigurationFileName == "" {
		homeDirectory, err := os.UserHomeDir()
		if err != nil {
			logger.Fatalf("failed to obtain user's home directory with error: %v", err)
		}

		config.Configuration.ConfigurationFileName = homeDirectory + defaultConfigLocation
	}
}
