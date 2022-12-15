package root

import (
	"github.com/alicenet/alicenet/config"
	"github.com/spf13/cobra"
	"time"
)

var Cmd = &cobra.Command{
	Use:   "alicenet",
	Short: "Short description of alicenet",
	Long:  "This is a not so long description for alicenet",
	//PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
	//	return nil
}

func init() {
	Cmd.PersistentFlags().StringVar(&config.Configuration.ConfigurationFileName, "config", "", "config file")
	// logging levels (&config.Configuration.ConfigurationFileName, "config", "c", "", "Name of config file")
	Cmd.PersistentFlags().StringVar(&config.Configuration.LoggingLevels, "logging", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.AliceNet, "loglevel.alicenet", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Consensus, "loglevel.consensus", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Transport, "loglevel.transport", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.App, "loglevel.app", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Db, "loglevel.db", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Gossipbus, "loglevel.gossipbus", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Badger, "loglevel.badger", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.PeerMan, "loglevel.peerman", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.LocalRPC, "loglevel.localRPC", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Dman, "loglevel.dman", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Yamux, "loglevel.yamux", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Ethereum, "loglevel.ethereum", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Main, "loglevel.main", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Deploy, "loglevel.deploy", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Utils, "loglevel.utils", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Monitor, "loglevel.monitor", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Dkg, "loglevel.dkg", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Services, "loglevel.services", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Settings, "loglevel.settings", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Validator, "loglevel.validator", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.MuxHandler, "loglevel.muxHandler", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Bootnode, "loglevel.bootnode", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.P2pmux, "loglevel.p2pmux", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Status, "loglevel.status", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Logging.Test, "loglevel.test", "", "")
	//chain
	Cmd.PersistentFlags().IntVar(&config.Configuration.Chain.ID, "chain.id", 0, "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Chain.StateDbPath, "chain.stateDb", "", "")
	Cmd.PersistentFlags().BoolVar(&config.Configuration.Chain.StateDbInMemory, "chain.stateDbInMemory", false, "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Chain.TransactionDbPath, "chain.transactionDb", "", "")
	Cmd.PersistentFlags().BoolVar(&config.Configuration.Chain.TransactionDbInMemory, "chain.transactionDbInMemory", false, "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Chain.MonitorDbPath, "chain.monitorDb", "", "")
	//ethereum
	Cmd.PersistentFlags().StringVar(&config.Configuration.Ethereum.Endpoint, "ethereum.endpoint", "", "")
	Cmd.PersistentFlags().Uint64Var(&config.Configuration.Ethereum.EndpointMinimumPeers, "ethereum.endpointMinimumPeers", 0, "Minimum peers required")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Ethereum.Keystore, "ethereum.keystore", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Ethereum.DefaultAccount, "ethereum.defaultAccount", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Ethereum.PassCodes, "ethereum.passCodes", "", "PassCodes for keystore")
	Cmd.PersistentFlags().Uint64Var(&config.Configuration.Ethereum.StartingBlock, "ethereum.startingBlock", 0, "he first block we care about")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Ethereum.FactoryAddress, "ethereum.factoryAddress", "", "")
	Cmd.PersistentFlags().Uint64Var(&config.Configuration.Ethereum.TxMaxGasFeeAllowedInGwei, "ethereum.txMaxGasFeeAllowedInGwei", 0, "")
	Cmd.PersistentFlags().BoolVar(&config.Configuration.Ethereum.TxMetricsDisplay, "ethereum.txMetricsDisplay", false, "")
	Cmd.PersistentFlags().Uint64Var(&config.Configuration.Ethereum.ProcessingBlockBatchSize, "ethereum.processingBlockBatchSize", 0, "")
	// transport
	Cmd.PersistentFlags().IntVar(&config.Configuration.Transport.PeerLimitMin, "transport.peerLimitMin", 0, "")
	Cmd.PersistentFlags().IntVar(&config.Configuration.Transport.PeerLimitMax, "transport.peerLimitMax", 0, "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Transport.PrivateKey, "transport.privateKey", "", "")
	Cmd.PersistentFlags().IntVar(&config.Configuration.Transport.OriginLimit, "transport.originLimit", 0, "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Transport.Whitelist, "transport.whitelist", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Transport.BootNodeAddresses, "transport.bootnodeAddresses", "", "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Transport.P2PListeningAddress, "transport.p2pListeningAddress", "", "")
	Cmd.PersistentFlags().BoolVar(&config.Configuration.Transport.UPnP, "transport.upnp", false, "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Transport.LocalStateListeningAddress, "transport.localStateListeningAddress", "", "")
	Cmd.PersistentFlags().DurationVar(&config.Configuration.Transport.Timeout, "transport.timeout", 1*time.Second, "")
	Cmd.PersistentFlags().BoolVar(&config.Configuration.Transport.FirewallMode, "transport.firewallMode", false, "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Transport.FirewallHost, "transport.firewallHost", "", "")
	// firewall
	Cmd.PersistentFlags().BoolVar(&config.Configuration.Firewalld.Enabled, "firewalld.enabled", false, "")
	Cmd.PersistentFlags().StringVar(&config.Configuration.Firewalld.SocketFile, "firewalld.socketFile", "", "")
}
