package initialization

import (
	"os"

	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/logging"
	toml "github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

// Command is the cobra.Command specifically for initialising the alicenet client.
var Command = cobra.Command{
	Use:   "init",
	Short: "Initialize the files/folders required for running the alicenet client",
	Long:  "Initialize the files/folders required for running the alicenet client",
	Run:   initialiseFilesAndFolders,
}

func initialiseFilesAndFolders(cmd *cobra.Command, args []string) {
	logger := logging.GetLogger("init").WithField("Component", cmd.Use)

	path := config.Configuration.Initialization.Path
	network := config.Configuration.Initialization.Network

	if path == "" {
		logger.Info("No path specified - defaulting to home directory")
		path = os.Getenv("HOME")
	}

	if network == "" {
		logger.Info("No environment specified - defaulting to mainnet")
		network = "mainnet"
	}

	logger.Info("Initialising AliceNet configuration files and folders...")

	// alicenet related paths and files
	alicenetPath := path + "/alicenet"
	envPath := alicenetPath + "/" + network
	stateDBPath := envPath + "/stateDB"
	transactionDBPath := envPath + "/transactionDB"
	monitorDBPath := envPath + "/monitorDB"
	keystoresPath := envPath + "/keystores"
	keysPath := keystoresPath + "/keys"
	configPath := envPath + "/config.toml"

	paths := []string{alicenetPath, envPath, stateDBPath, transactionDBPath, monitorDBPath, keystoresPath, keysPath}

	// loop through the paths checking to see if they exist and exit if any of them do
	for _, path := range paths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			logger.WithError(err).Fatal("Path already exists", path)
			logger.Fatal("Remove all existing paths and try again")
		}
	}

	// check for the existence of the config file and exit if it exists
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		logger.WithError(err).Error("Config file already exists: ", configPath)
		logger.Fatal("Remove the existing config file and try again")
	}

	// create the paths
	for _, path := range paths {
		err := os.MkdirAll(path, 0o755)
		if err != nil {
			logger.WithError(err).Fatal("Failed to create path: ", path)
			// remove the root folder and its subfolders then exit
			err := os.RemoveAll(alicenetPath)
			if err != nil {
				logger.WithError(err).Fatal("Failed to remove path: ", alicenetPath)
				logger.Fatal("Please remove the root folder and its subfolders manually")
			}
			logger.Fatal("Exiting")
		}
	}

	type LogLevel struct {
		Admin      string `toml:"admin" comment:"Controls the logging level of the alicenet sub-services. To get the complete list of supported services check ./constants/shared.go"`
		Blockchain string `toml:"blockchain"`
		Consensus  string `toml:"consensus"`
	}
	type Chain struct {
		ChainID               string `toml:"id" comment:"AliceNet ChainID that corresponds to the AliceNet Network that you are trying to connect."`
		StateDB               string `toml:"stateDB" comment:"Path to the location where to save the state database. The state database is responsible for storing the AliceNet blockchain data (blocks, validator's data). Store this database in a safe location. If this database is deleted, the node will re-sync with its peers from scratch. DON'T DELETE THIS DATABASE IF YOU ARE RUNNING A VALIDATOR NODE!!! If this database is deleted and you are running a validator node, the validator's data will be permanently deleted and the node will not be able to proceed with its work as a validator, even after a re-sync. Therefore, you may be susceptible to a slashing event"`
		TransactionDB         string `toml:"transactionDB" comment:"Path to the location where to save the transaction database. The transaction database is responsible for storing the AliceNet blockchain data (transactions). Store this database in a safe location. If this database is deleted, the node will re-sync all transactions with its peers."`
		MonitorDB             string `toml:"monitorDB" comment:"Path to the location where to save the monitor database. The monitor database is responsible for storing the (events, tasks, receipts) coming from layer 1 blockchains that AliceNet is anchored with. Store this database in a safe location. If this database is deleted, the node will replay all events by querying the layer1 chains using the information provided below."`
		MonitorDBInMemory     bool   `toml:"monitorDBInMemory" comment:"Flags to save any of the databases above only on memory. USE ONLY RECOMMENDED TO SET TRUE FOR TESTING PURPOSES."`
		StateDBInMemory       bool   `toml:"stateDBInMemory"`
		TransactionDBInMemory bool   `toml:"transactionDBInMemory"`
	}
	type Transport struct {
		Upnp                       bool   `toml:"upnp" comment:"IF UPNP should be used to discover opened ports to connect with the peers."`
		PrivateKey                 string `toml:"privateKey" comment:"16 Byte private key that is used to encrypt and decrypt information shared with peers. Generate this with a safe random generator."`
		BootNodeAddresses          string `toml:"bootNodeAddresses" comment:"Address to a bootnode running on the desired AliceNet network that you are trying to connect with. A bootnode is a software client responsible for sharing information about alicenet peers. Your node will connect to a bootnode to retrieve an initial list of peers to try to connect with."`
		OriginLimit                int    `toml:"originLimit" comment:"Maximum number of peers that we can retrieve from the bootnode?"`
		LocalStateListeningAddress string `toml:"localStateListeningAddress" comment:"Address and port where your node will be listening for rpc requests."`
		P2pListeningAddress        string `toml:"p2pListeningAddress" comment:"Address and port where you node will be listening for requests coming from other peers. The address should be publicly reachable."`
		PeerLimitMax               int    `toml:"peerLimitMax" comment:"Maximum number of peers that you wish to be connected with. Upper bound to limit bandwidth shared with the peers."`
		PeerLimitMin               int    `toml:"peerLimitMin" comment:"Minimum number of peers that you wish to be connect with, before trying to attempt to download blockchain data and participate in consensus."`
	}
	type Ethereum struct {
		Endpoint                 string `toml:"endpoint" comment:"Ethereum endpoint url to the ethereum chain where the AliceNet network infra-structure that you are trying to connect lives. Ethereum mainnet for AliceNet mainnet and Goerli for Alicenet Testnet. Infura and Alchemy services can be used, but if you are running your own validator node, we recommend to use a more secure node."`
		EndpointMinimumPeers     int    `toml:"endpointMinimumPeers" comment:"Minimum number of peers connected to your ethereum node that you wish to reach before trying to process ethereum blocks to retrieve the AliceNet events."`
		DefaultAccount           string `toml:"defaultAccount" comment:"Ethereum address that will be used to sign transactions and connect to the AliceNet services on ethereum."`
		Keystore                 string `toml:"keystore" comment:"Path to the encrypted private key used on the address above."`
		PassCodes                string `toml:"passCodes" comment:"Path to the file containing the password to unlock the account private key."`
		FactoryAddress           string `toml:"factoryAddress" comment:"Ethereum address of the AliceNet factory of smart contracts. The factory is responsible for registering and deploying every contract used by the AliceNet infra-structure."`
		StartingBlock            string `toml:"startingBlock" comment:"The ethereum block where the AliceNet contract factory was deployed. This block is used as starting block to retrieve all events (e.g snapshots, deposits) necessary to initialize and validate your AliceNet node."`
		ProcessingBlockBatchSize int    `toml:"processingBlockBatchSize" comment:"Batch size of blocks that will be downloaded and processed from your endpoint address. Every ethereum block starting from the startingBlock until the latest ethereum block will be downloaded and all events (e.g snapshots, deposits) coming from AliceNet smart contracts will be processed in a chronological order. If this value is too large, your endpoint may end up being overloaded with API requests."`
		TxMaxGasFeeAllowedInGwei int    `toml:"txMaxGasFeeAllowedInGwei" comment:"The maximum gas price that you are willing to pay (in GWEI) for a transaction done by your node. If you are validator, putting this value too low can result in your node failing to fulfill the validators duty, hence, being passive for a slashing."`
		TxMetricsDisplay         bool   `toml:"txMetricsDisplay" comment:"Flag to decide if the ethereum transactions information will be shown on the logs."`
	}
	type Utils struct {
		Status bool `toml:"status" comment:"Flag to decide if the status will be shown on the logs. Maybe be a little noisy."`
	}
	type Validator struct {
		RewardCurveSpec int    `toml:"rewardCurveSpec" comment:"Type of elliptic curve used to generate the AliceNet address. 1: secp256k1 (same as ethereum), 2: BN128"`
		RewardAccount   string `toml:"rewardAccount" comment:"Address of the AliceNet account used to do transactions in the AliceNet network."`
		SymmetricKey    string `toml:"symmetricKey"`
	}
	type Config struct {
		LogLevel  LogLevel  `toml:"loglevel"`
		Chain     Chain     `toml:"chain"`
		Transport Transport `toml:"transport"`
		Ethereum  Ethereum  `toml:"ethereum"`
		Utils     Utils     `toml:"utils"`
		Validator Validator `toml:"validator" comment:"OPTIONAL: Only necessary if you plan to run a validator node."`
	}

	var chainId string
	switch network {
	case "testnet":
		chainId = "42"
	case "mainnet":
		chainId = "21"
	default:
		chainId = "21"
	}

	config := Config{
		LogLevel{
			Admin:      "info",
			Blockchain: "info",
			Consensus:  "info",
		},
		Chain{
			ChainID:               chainId,
			StateDB:               stateDBPath,
			TransactionDB:         transactionDBPath,
			MonitorDB:             monitorDBPath,
			MonitorDBInMemory:     false,
			StateDBInMemory:       false,
			TransactionDBInMemory: false,
		},
		Transport{
			Upnp:                       false,
			PrivateKey:                 "<16_BYTES_TRANSPORT_PRIVATE_KEY>",
			BootNodeAddresses:          "<BOOTNODE_ADDRESS>",
			OriginLimit:                50,
			LocalStateListeningAddress: "0.0.0.0:8883",
			P2pListeningAddress:        "0.0.0.0:4342",
			PeerLimitMax:               24,
			PeerLimitMin:               3,
		},
		Ethereum{
			Endpoint:                 "<ETHEREUM_ENDPOINT_URL>",
			EndpointMinimumPeers:     1,
			DefaultAccount:           "<0xETHEREUM_ADDRESS>",
			Keystore:                 keystoresPath,
			PassCodes:                keystoresPath + "/passcodes.txt",
			FactoryAddress:           "<0xFACTORY_ETHEREUM_ADDRESS>",
			StartingBlock:            "<StartingBlock>",
			ProcessingBlockBatchSize: 1_000,
			TxMaxGasFeeAllowedInGwei: 500,
			TxMetricsDisplay:         false,
		},
		Utils{
			Status: true,
		},
		Validator{
			RewardCurveSpec: 1,
			RewardAccount:   "0x<ALICENET_ADDRESS>",
			SymmetricKey:    "<SOME_SUPER_FANCY_SECRET_THAT_WILL_BE_HASHED>",
		},
	}
	b, err := toml.Marshal(config)
	if err != nil {
		logger.Fatal(err)
	}

	// create config.toml file as text file in environment folder
	configFile, err := os.Create(configPath)
	if err != nil {
		logger.Error("Error creating config.toml file: ", err)
		os.Exit(1)
	}

	// write config.toml file
	configFile.Write(b)
	configFile.Close()

	logger.Info("created config file")

	os.Exit(0)
}
