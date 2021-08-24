package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/MadBase/MadNet/cmd/bootnode"
	"github.com/MadBase/MadNet/cmd/deploy"
	"github.com/MadBase/MadNet/cmd/utils"
	"github.com/MadBase/MadNet/cmd/validator"
	"github.com/MadBase/MadNet/config"
	"github.com/MadBase/MadNet/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type option struct {
	name  string
	short string
	usage string
	value interface{}
}

// Runner wraps a cobra command's Run() and sets up loggers first
// It assumes 'logging' flag uses the format "pkg1=debug,pkg2=error"
func runner(commandRun func(*cobra.Command, []string)) func(*cobra.Command, []string) {
	logger := logging.GetLogger("main")
	return func(a *cobra.Command, b []string) {
		loggingLevels := config.Configuration.LoggingLevels
		if len(loggingLevels) == 0 {
			commandRun(a, b)
		} else {
			loggers := strings.Split(loggingLevels, ",")
			for _, levelSetting := range loggers {
				settingComponent := strings.Split(levelSetting, "=")
				if len(settingComponent) != 2 {
					logger.Fatalf("Malformed log level setting %q", levelSetting)
				}

				loggerName := settingComponent[0]
				loggerLevel := strings.ToLower(settingComponent[1])

				lgr := logging.GetLogger(loggerName)

				switch loggerLevel {
				case "debug":
					lgr.SetLevel(logrus.DebugLevel)
				case "info":
					lgr.SetLevel(logrus.InfoLevel)
				case "warn":
					lgr.SetLevel(logrus.WarnLevel)
				case "error":
					lgr.SetLevel(logrus.ErrorLevel)
				}
			}
			commandRun(a, b)
		}
	}
}

func main() {
	logger := logging.GetLogger("main")
	logger.SetLevel(logrus.InfoLevel)

	// Root for all commands
	rootCommand := cobra.Command{
		Use:   "madnet",
		Short: "Short description of madnet",
		Long:  "This is a not so long description for madnet"}

	// All the configuration options available. Used for command line and config file.
	options := map[*cobra.Command][]*option{
		&rootCommand: {
			{"config", "c", "Name of config file", &config.Configuration.ConfigurationFileName},
			{"logging", "", "", &config.Configuration.LoggingLevels},
			{"chain.id", "", "", &config.Configuration.Chain.ID},
			{"chain.stateDB", "", "", &config.Configuration.Chain.StateDbPath},
			{"chain.stateDBInMemory", "", "", &config.Configuration.Chain.StateDbInMemory},
			{"chain.transactionDB", "", "", &config.Configuration.Chain.TransactionDbPath},
			{"chain.transactionDBInMemory", "", "", &config.Configuration.Chain.TransactionDbInMemory},
			{"chain.monitorDB", "", "", &config.Configuration.Chain.MonitorDbPath},
			{"chain.monitorDBInMemory", "", "", &config.Configuration.Chain.MonitorDbInMemory},
			{"ethereum.endpoint", "", "", &config.Configuration.Ethereum.Endpoint},
			{"ethereum.endpointPeers", "", "Minimum peers required", &config.Configuration.Ethereum.EndpointMinimumPeers},
			{"ethereum.keystore", "", "", &config.Configuration.Ethereum.Keystore},
			{"ethereum.timeout", "", "", &config.Configuration.Ethereum.Timeout},
			{"ethereum.testEther", "", "", &config.Configuration.Ethereum.TestEther},
			{"ethereum.deployAccount", "", "", &config.Configuration.Ethereum.DeployAccount},
			{"ethereum.defaultAccount", "", "", &config.Configuration.Ethereum.DefaultAccount},
			{"ethereum.finalityDelay", "", "Number blocks before we consider a block final", &config.Configuration.Ethereum.FinalityDelay},
			{"ethereum.retryCount", "", "Number of times to retry an Ethereum operation", &config.Configuration.Ethereum.RetryCount},
			{"ethereum.retryDelay", "", "Delay between retry attempts", &config.Configuration.Ethereum.RetryDelay},
			{"ethereum.passcodes", "", "Passcodes for keystore", &config.Configuration.Ethereum.Passcodes},
			{"ethereum.startingBlock", "", "The first block we care about", &config.Configuration.Ethereum.StartingBlock},
			{"ethereum.registryAddress", "", "", &config.Configuration.Ethereum.RegistryAddress},
			{"monitor.batchSize", "", "", &config.Configuration.Monitor.BatchSize},
			{"monitor.interval", "", "", &config.Configuration.Monitor.Interval},
			{"transport.peerLimitMin", "", "", &config.Configuration.Transport.PeerLimitMin},
			{"transport.peerLimitMax", "", "", &config.Configuration.Transport.PeerLimitMax},
			{"transport.privateKey", "", "", &config.Configuration.Transport.PrivateKey},
			{"transport.originLimit", "", "", &config.Configuration.Transport.OriginLimit},
			{"transport.whitelist", "", "", &config.Configuration.Transport.Whitelist},
			{"transport.bootnodeAddresses", "", "", &config.Configuration.Transport.BootNodeAddresses},
			{"transport.p2pListeningAddress", "", "", &config.Configuration.Transport.P2PListeningAddress},
			{"transport.discoveryListeningAddress", "", "", &config.Configuration.Transport.DiscoveryListeningAddress},
			{"transport.upnp", "", "", &config.Configuration.Transport.UPnP},
			{"transport.localStateListeningAddress", "", "", &config.Configuration.Transport.LocalStateListeningAddress},
			{"transport.timeout", "", "", &config.Configuration.Transport.Timeout},
			{"transport.firewallMode", "", "", &config.Configuration.Transport.FirewallMode},
			{"transport.firewallHost", "", "", &config.Configuration.Transport.FirewallHost}},

		&utils.Command: {
			{"utils.status", "", "", &config.Configuration.Utils.Status}},

		&utils.ApproveTokensCommand:  {},
		&utils.EthdkgCommand:         {},
		&utils.RegisterCommand:       {},
		&utils.SendWeiCommand:        {},
		&utils.TransferTokensCommand: {},
		&utils.UnregisterCommand:     {},
		&utils.DepositCommand:        {},

		&bootnode.Command: {
			{"bootnode.listeningAddress", "", "", &config.Configuration.BootNode.ListeningAddress},
			{"bootnode.cacheSize", "", "", &config.Configuration.BootNode.CacheSize}},

		&validator.Command: {
			{"validator.rewardAccount", "", "", &config.Configuration.Validator.RewardAccount},
			{"validator.rewardCurveSpec", "", "", &config.Configuration.Validator.RewardCurveSpec}},

		&deploy.Command: {
			{"deploy.migrations", "", "", &config.Configuration.Deploy.Migrations},
			{"deploy.testMigrations", "", "", &config.Configuration.Deploy.TestMigrations}},
	}

	// Establish command hierarchy
	hierarchy := map[*cobra.Command]*cobra.Command{
		&bootnode.Command:            &rootCommand,
		&validator.Command:           &rootCommand,
		&deploy.Command:              &rootCommand,
		&utils.Command:               &rootCommand,
		&utils.ApproveTokensCommand:  &utils.Command,
		&utils.EthdkgCommand:         &utils.Command,
		&utils.RegisterCommand:       &utils.Command,
		&utils.SendWeiCommand:        &utils.Command,
		&utils.TransferTokensCommand: &utils.Command,
		&utils.UnregisterCommand:     &utils.Command,
		&utils.DepositCommand:        &utils.Command}

	// Convert option abstraction into concrete settings for Cobra and Viper
	for c := range options {

		cFlags := c.PersistentFlags() // just a convenience thing

		if c.Run != nil {
			c.Run = runner(c.Run)
		}

		if parentCommand, present := hierarchy[c]; present {
			cFlags = c.Flags()
			parentCommand.AddCommand(c)
		}

		var defaultStringArray []string
		for _, o := range options[c] {

			typeOfPtr := reflect.TypeOf(o.value)
			if typeOfPtr.Kind() != reflect.Ptr {
				logger.Fatalf("Option value for %v should be supplied as a pointer.", o.name)
			} else {
				// These cascading type asserts don't work in a switch statement
				if durPtr, ok := o.value.(*time.Duration); ok {
					cFlags.DurationVarP(durPtr, o.name, o.short, 1*time.Second, o.usage)
				} else if strPtr, ok := o.value.(*string); ok {
					cFlags.StringVarP(strPtr, o.name, o.short, "", o.usage)
				} else if strArrayPtr, ok := o.value.(*[]string); ok {
					cFlags.StringArrayVarP(strArrayPtr, o.name, o.short, defaultStringArray, o.usage)
				} else if intPtr, ok := o.value.(*int); ok {
					cFlags.IntVarP(intPtr, o.name, o.short, 0, o.usage)
				} else if boolPtr, ok := o.value.(*bool); ok {
					cFlags.BoolVarP(boolPtr, o.name, o.short, false, o.usage)
				} else {
					logger.Fatalf("Configuration structure has unknown type for %v.", o.name)
				}

				// Viper has to lookup the pflag Cobra created because Cobra can't
				f := cFlags.Lookup(o.name)
				config.SetBinding(o.value, f) // Register all the pointers to the flag using them
				if err := viper.BindPFlag(o.name, f); err != nil {
					logger.Fatalf("Could not bind to pflag: %v\n", o.name)
				} else {
					logger.Debugf("Binding of %q was successful\n", o.name)
				}
			}
		}
	}

	// This has to be registered prior to root command execute. Cobra executes this first thing when executing.
	cobra.OnInitialize(func() {

		// Read the config file
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
			logger.Warnf("Opening file failed:%q", err)
		}

		/* The logic here feels backwards to me but it isn't.
		Command line flags aren't set till this func returns.
		So we set flags from config file here and when func returns the command line will overwrite.
		*/
		for cmd := range options {
			// Find all the flags
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				flag.Value.Set(viper.GetString(flag.Name))
			})
		}

		logger.Debugf("onInitialize() -- Configuration:%v", config.Configuration)
	})

	// Really start application here
	rootCommand.Execute()
	logger.Debugf("main() -- Configuration:%q", config.Configuration.Ethereum)
}
