package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/alicenet/alicenet/cmd/bootnode"
	"github.com/alicenet/alicenet/cmd/firewalld"
	"github.com/alicenet/alicenet/cmd/utils"
	"github.com/alicenet/alicenet/cmd/validator"
	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/logging"
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
			logger.Infof("Setting log level for '%v' to '%v'", logName, logLevel)
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
	case "debug":
		lgr.SetLevel(logrus.DebugLevel)
	case "info":
		lgr.SetLevel(logrus.InfoLevel)
	case "warn":
		lgr.SetLevel(logrus.WarnLevel)
	case "error":
		lgr.SetLevel(logrus.ErrorLevel)
	default:
		lgr.SetLevel(logrus.InfoLevel)
	}
}

func main() {
	logger := logging.GetLogger("main")
	logger.SetLevel(logrus.InfoLevel)

	// Root for all commands
	rootCommand := cobra.Command{
		Use:   "alicenet",
		Short: "Short description of alicenet",
		Long:  "This is a not so long description for alicenet"}

	// All the configuration options available. Used for command line and config file.
	options := map[*cobra.Command][]*option{
		&rootCommand: {
			{"config", "c", "Name of config file", &config.Configuration.ConfigurationFileName},
			{"logging", "", "", &config.Configuration.LoggingLevels},
			{"loglevel.alicenet", "", "", &config.Configuration.Logging.AliceNet},
			{"loglevel.consensus", "", "", &config.Configuration.Logging.Consensus},
			{"loglevel.transport", "", "", &config.Configuration.Logging.Transport},
			{"loglevel.app", "", "", &config.Configuration.Logging.App},
			{"loglevel.db", "", "", &config.Configuration.Logging.Db},
			{"loglevel.gossipbus", "", "", &config.Configuration.Logging.Gossipbus},
			{"loglevel.badger", "", "", &config.Configuration.Logging.Badger},
			{"loglevel.peerMan", "", "", &config.Configuration.Logging.PeerMan},
			{"loglevel.localRPC", "", "", &config.Configuration.Logging.LocalRPC},
			{"loglevel.dman", "", "", &config.Configuration.Logging.Dman},
			{"loglevel.peer", "", "", &config.Configuration.Logging.Peer},
			{"loglevel.yamux", "", "", &config.Configuration.Logging.Yamux},
			{"loglevel.ethereum", "", "", &config.Configuration.Logging.Ethereum},
			{"loglevel.main", "", "", &config.Configuration.Logging.Main},
			{"loglevel.deploy", "", "", &config.Configuration.Logging.Deploy},
			{"loglevel.utils", "", "", &config.Configuration.Logging.Utils},
			{"loglevel.monitor", "", "", &config.Configuration.Logging.Monitor},
			{"loglevel.dkg", "", "", &config.Configuration.Logging.Dkg},
			{"loglevel.services", "", "", &config.Configuration.Logging.Services},
			{"loglevel.settings", "", "", &config.Configuration.Logging.Settings},
			{"loglevel.validator", "", "", &config.Configuration.Logging.Validator},
			{"loglevel.muxHandler", "", "", &config.Configuration.Logging.MuxHandler},
			{"loglevel.bootnode", "", "", &config.Configuration.Logging.Bootnode},
			{"loglevel.p2pmux", "", "", &config.Configuration.Logging.P2pmux},
			{"loglevel.status", "", "", &config.Configuration.Logging.Status},
			{"loglevel.test", "", "", &config.Configuration.Logging.Test},
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
			{"ethereum.txFeePercentageToIncrease", "", "", &config.Configuration.Ethereum.TxFeePercentageToIncrease},
			{"ethereum.txMaxFeeThresholdInGwei", "", "", &config.Configuration.Ethereum.TxMaxFeeThresholdInGwei},
			{"ethereum.txCheckFrequency", "", "", &config.Configuration.Ethereum.TxCheckFrequency},
			{"ethereum.txTimeoutForReplacement", "", "", &config.Configuration.Ethereum.TxTimeoutForReplacement},
			{"monitor.batchSize", "", "", &config.Configuration.Monitor.BatchSize},
			{"monitor.interval", "", "", &config.Configuration.Monitor.Interval},
			{"monitor.timeout", "", "", &config.Configuration.Monitor.Timeout},
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
			{"transport.firewallHost", "", "", &config.Configuration.Transport.FirewallHost},
			{"firewalld.enabled", "", "", &config.Configuration.Firewalld.Enabled},
			{"firewalld.socketFile", "", "", &config.Configuration.Firewalld.SocketFile},
		},

		&utils.Command: {
			{"utils.status", "", "", &config.Configuration.Utils.Status}},

		&utils.EthdkgCommand:  {},
		&utils.SendWeiCommand: {},

		&bootnode.Command: {
			{"bootnode.listeningAddress", "", "", &config.Configuration.BootNode.ListeningAddress},
			{"bootnode.cacheSize", "", "", &config.Configuration.BootNode.CacheSize}},

		&firewalld.Command: {},

		&validator.Command: {
			{"validator.rewardAccount", "", "", &config.Configuration.Validator.RewardAccount},
			{"validator.rewardCurveSpec", "", "", &config.Configuration.Validator.RewardCurveSpec}},

		// &deploy.Command: {
		// 	{"deploy.migrations", "", "", &config.Configuration.Deploy.Migrations},
		// 	{"deploy.testMigrations", "", "", &config.Configuration.Deploy.TestMigrations}},
	}

	// Establish command hierarchy
	hierarchy := map[*cobra.Command]*cobra.Command{
		&firewalld.Command: &rootCommand,
		&bootnode.Command:  &rootCommand,
		&validator.Command: &rootCommand,
		// &deploy.Command:              &rootCommand,
		&utils.Command:        &rootCommand,
		&utils.EthdkgCommand:  &utils.Command,
		&utils.SendWeiCommand: &utils.Command,
	}

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
				} else if uint64Ptr, ok := o.value.(*uint64); ok {
					cFlags.Uint64VarP(uint64Ptr, o.name, o.short, 0, o.usage)
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
				err := flag.Value.Set(viper.GetString(flag.Name))
				if err != nil {
					logger.Warnf("Setting flag %q failed:%q", flag.Name, err)
				}
			})
		}

		logger.Debugf("onInitialize() -- Configuration:%v", config.Configuration)
	})

	// Really start application here
	err := rootCommand.Execute()
	if err != nil {
		logger.Fatalf("Execute() failed:%q", err)
	}
	logger.Debugf("main() -- Configuration:%q", config.Configuration.Ethereum)
}
