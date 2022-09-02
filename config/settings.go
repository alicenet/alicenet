package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type InitConfig struct {
	Path    string
	Network string
}

type BootnodeConfig struct {
	Name             string
	ListeningAddress string
	CacheSize        int
}

type EthKeyConfig struct {
	PasswordFile    string
	Json            bool
	PrivateKey      string
	LightKDF        bool
	Private         bool
	NewPasswordFile string
}

type ChainConfig struct {
	ID                    int    `toml:"id" comment:"AliceNet ChainID that corresponds to the AliceNet Network that you are trying to connect."`
	StateDbPath           string `toml:"stateDB" comment:"Path to the location where to save the state database. The state database is responsible for storing the AliceNet blockchain data (blocks, validator's data). Store this database in a safe location. If this database is deleted, the node will re-sync with its peers from scratch. DON'T DELETE THIS DATABASE IF YOU ARE RUNNING A VALIDATOR NODE!!! If this database is deleted and you are running a validator node, the validator's data will be permanently deleted and the node will not be able to proceed with its work as a validator, even after a re-sync. Therefore, you may be susceptible to a slashing event"`
	StateDbInMemory       bool   `toml:"stateDBInMemory"`
	TransactionDbPath     string `toml:"transactionDB" comment:"Path to the location where to save the transaction database. The transaction database is responsible for storing the AliceNet blockchain data (transactions). Store this database in a safe location. If this database is deleted, the node will re-sync all transactions with its peers."`
	TransactionDbInMemory bool   `toml:"transactionDBInMemory"`
	MonitorDbPath         string `toml:"monitorDB" comment:"Path to the location where to save the monitor database. The monitor database is responsible for storing the (events, tasks, receipts) coming from layer 1 blockchains that AliceNet is anchored with. Store this database in a safe location. If this database is deleted, the node will replay all events by querying the layer1 chains using the information provided below."`
	MonitorDbInMemory     bool   `toml:"monitorDBInMemory" comment:"Flags to save any of the databases above only on memory. USE ONLY RECOMMENDED TO SET TRUE FOR TESTING PURPOSES."`
}

type EthereumConfig struct {
	DefaultAccount           string `toml:"defaultAccount" comment:"Ethereum address that will be used to sign transactions and connect to the AliceNet services on ethereum."`
	Endpoint                 string `toml:"endpoint" comment:"Ethereum endpoint url to the ethereum chain where the AliceNet network infra-structure that you are trying to connect lives. Ethereum mainnet for AliceNet mainnet and Goerli for Alicenet Testnet. Infura and Alchemy services can be used, but if you are running your own validator node, we recommend to use a more secure node."`
	EndpointMinimumPeers     uint64 `toml:"endpointMinimumPeers" comment:"Minimum number of peers connected to your ethereum node that you wish to reach before trying to process ethereum blocks to retrieve the AliceNet events."`
	Keystore                 string `toml:"keystore" comment:"Path to the encrypted private key used on the address above."`
	PassCodes                string `toml:"passCodes" comment:"Path to the file containing the password to unlock the account private key."`
	FactoryAddress           string `toml:"factoryAddress" comment:"Ethereum address of the AliceNet factory of smart contracts. The factory is responsible for registering and deploying every contract used by the AliceNet infra-structure."`
	StartingBlock            uint64 `toml:"startingBlock" comment:"The ethereum block where the AliceNet contract factory was deployed. This block is used as starting block to retrieve all events (e.g snapshots, deposits) necessary to initialize and validate your AliceNet node."`
	TxMaxGasFeeAllowedInGwei uint64 `toml:"txMaxGasFeeAllowedInGwei" comment:"The maximum gas price that you are willing to pay (in GWEI) for a transaction done by your node. If you are validator, putting this value too low can result in your node failing to fulfill the validators duty, hence, being passive for a slashing."`
	TxMetricsDisplay         bool   `toml:"txMetricsDisplay" comment:"Flag to decide if the ethereum transactions information will be shown on the logs."`
	ProcessingBlockBatchSize uint64 `toml:"processingBlockBatchSize" comment:"Batch size of blocks that will be downloaded and processed from your endpoint address. Every ethereum block starting from the startingBlock until the latest ethereum block will be downloaded and all events (e.g snapshots, deposits) coming from AliceNet smart contracts will be processed in a chronological order. If this value is too large, your endpoint may end up being overloaded with API requests."`
}

type TransportConfig struct {
	Size                       int           `toml:",omitempty"`
	Timeout                    time.Duration `toml:",omitempty"`
	OriginLimit                int           `toml:"originLimit" comment:"Maximum number of peers that we can retrieve from the bootnode?"`
	PeerLimitMin               int           `toml:"peerLimitMin" comment:"Minimum number of peers that you wish to be connect with, before trying to attempt to download blockchain data and participate in consensus."`
	PeerLimitMax               int           `toml:"peerLimitMax" comment:"Maximum number of peers that you wish to be connected with. Upper bound to limit bandwidth shared with the peers."`
	FirewallMode               bool          `toml:",omitempty"`
	FirewallHost               string        `toml:",omitempty"`
	Whitelist                  string        `toml:",omitempty"`
	PrivateKey                 string        `toml:"privateKey" comment:"16 Byte private key that is used to encrypt and decrypt information shared with peers. Generate this with a safe random generator."`
	BootNodeAddresses          string        `toml:"bootNodeAddresses" comment:"Address to a bootnode running on the desired AliceNet network that you are trying to connect with. A bootnode is a software client responsible for sharing information about alicenet peers. Your node will connect to a bootnode to retrieve an initial list of peers to try to connect with."`
	P2PListeningAddress        string        `toml:"p2pListeningAddress" comment:"Address and port where you node will be listening for requests coming from other peers. The address should be publicly reachable."`
	LocalStateListeningAddress string        `toml:"localStateListeningAddress" comment:"Address and port where your node will be listening for rpc requests."`
	UPnP                       bool          `toml:"upnp" comment:"IF UPNP should be used to discover opened ports to connect with the peers."`
}

type DeployConfig struct {
	Migrations     bool `toml:",omitempty"`
	TestMigrations bool `toml:",omitempty"`
}

type UtilsConfig struct {
	Status bool `toml:"status" comment:"Flag to decide if the status will be shown on the logs. Maybe be a little noisy."`
}

type ValidatorConfig struct {
	Repl            bool   `toml:",omitempty"`
	RewardAccount   string `toml:"rewardAccount" comment:"Address of the AliceNet account used to do transactions in the AliceNet network."`
	RewardCurveSpec int    `toml:"rewardCurveSpec" comment:"Type of elliptic curve used to generate the AliceNet address. 1: secp256k1 (same as ethereum), 2: BN128"`
	SymmetricKey    string `toml:"symmetricKey"`
}

type LoggingConfig struct {
	AliceNet   string `toml:",omitempty"`
	Consensus  string `toml:"consensus"`
	Transport  string `toml:",omitempty"`
	App        string `toml:",omitempty"`
	Db         string `toml:",omitempty"`
	Gossipbus  string `toml:",omitempty"`
	Badger     string `toml:",omitempty"`
	PeerMan    string `toml:",omitempty"`
	LocalRPC   string `toml:",omitempty"`
	Dman       string `toml:",omitempty"`
	Peer       string `toml:",omitempty"`
	Yamux      string `toml:",omitempty"`
	Ethereum   string `toml:",omitempty"`
	Main       string `toml:",omitempty"`
	Deploy     string `toml:",omitempty"`
	Utils      string `toml:",omitempty"`
	Monitor    string `toml:",omitempty"`
	Dkg        string `toml:",omitempty"`
	Services   string `toml:",omitempty"`
	Settings   string `toml:",omitempty"`
	Validator  string `toml:",omitempty"`
	MuxHandler string `toml:",omitempty"`
	Bootnode   string `toml:",omitempty"`
	P2pmux     string `toml:",omitempty"`
	Status     string `toml:",omitempty"`
	Test       string `toml:",omitempty"`
}

type FirewalldConfig struct {
	Enabled    bool   `toml:",omitempty"`
	SocketFile string `toml:",omitempty"`
}
type RootSerializableConfiguration struct {
	Logging   LoggingConfig   `toml:"loglevel"`
	Chain     ChainConfig     `toml:"chain"`
	Transport TransportConfig `toml:"transport"`
	Ethereum  EthereumConfig  `toml:"ethereum"`
	Utils     UtilsConfig     `toml:"utils"`
	Validator ValidatorConfig `toml:"validator" comment:"OPTIONAL: Only necessary if you plan to run a validator node."`
}
type RootConfiguration struct {
	ConfigurationFileName string
	LoggingLevels         string // backwards compatibility
	Deploy                DeployConfig
	Firewalld             FirewalldConfig
	BootNode              BootnodeConfig
	EthKey                EthKeyConfig
	Version               string
	Initialization        InitConfig
	Logging               LoggingConfig
	Chain                 ChainConfig
	Transport             TransportConfig
	Ethereum              EthereumConfig
	Utils                 UtilsConfig
	Validator             ValidatorConfig
}

// Configuration contains all active settings.
var Configuration RootConfiguration

type s struct {
	v interface{}
}

var flagMap map[s]*pflag.Flag

// SetBinding registers a particular Flag as tied to a particular pointer.
func SetBinding(ptr interface{}, f *pflag.Flag) {
	logger := logging.GetLogger("settings")
	logger.SetLevel(logrus.WarnLevel)
	if flagMap == nil {
		flagMap = make(map[s]*pflag.Flag)
	}
	logger.Debugf("registering %q of type %q to %p", f.Name, f.Value.Type(), ptr)
	flagMap[s{ptr}] = f
}

// SetValue takes a ptr and updates the value of the flag that's pointing to it.
func SetValue(ptr, value interface{}) {
	logger := logging.GetLogger("settings")
	f, ok := flagMap[s{ptr}]
	if !ok {
		logger.Warnf("Could not find binding for %q", ptr)
	} else {
		logger.Debugf("Setting value of %q (%q) to %v", ptr, f.Name, value)

		viper.Set(f.Name, value) // Apparently the bindings don't work both directions

		val := fmt.Sprint(value)
		err := f.Value.Set(val) // This is for cobra but not sure if it matters, but don't want to risk it
		if err != nil {
			logger.Warnf("Failed to set value of flag %v to %v", f.Name, val)
		}

		logger.Infof("retrieved value is %v", f.Value.String())
	}
}

func (t TransportConfig) BootNodes() []string {
	bootNodeAddresses := strings.Split(t.BootNodeAddresses, ",")
	for idx := range bootNodeAddresses {
		bootNodeAddresses[idx] = strings.TrimSpace(bootNodeAddresses[idx])
	}
	return bootNodeAddresses
}
