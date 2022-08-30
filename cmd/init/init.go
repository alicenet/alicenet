package utils

import (
	"os"

	"github.com/alicenet/alicenet/logging"
	"github.com/spf13/cobra"
)

// Command is the cobra.Command specifically for initialising the alicenet client.
var Command = cobra.Command{
	Use:   "init",
	Short: "Initialise the files/folders required for running the alicenet client",
	Long:  "Initialise the files/folders required for running the alicenet client",
	Run:   initialiseFilesAndFolders,
}

func initialiseFilesAndFolders(cmd *cobra.Command, args []string) {
	logger := logging.GetLogger("init").WithField("Component", cmd.Use)

	var path string
	cmd.Flags().StringVarP(&path, "path", "p", "", "Path to save the files/folders")

	var environment string
	cmd.Flags().StringVarP(&environment, "env", "e", "", "Environment to use (testnet, mainnet)")

	if path == "" {
		logger.Info("No path specified - defaulting to home directory")
		path = os.Getenv("HOME")
	}

	if environment == "" {
		logger.Info("No environment specified - defaulting to mainnet")
		environment = "mainnet"
	}

	logger.Info("Initialising AliceNet configuration files and folders...")

	// create alicenet folder in path directory
	alicenetPath := path + "/alicenet"
	err := os.MkdirAll(alicenetPath, 0o755)
	if err != nil {
		logger.Error("Error creating alicenet folder: ", err)
		os.Exit(1)
	}

	// create subfolder based on environment
	envPath := alicenetPath + "/" + environment
	err = os.MkdirAll(envPath, 0o755)
	if err != nil {
		logger.Error("Error creating environment folder: ", err)
		os.Exit(1)
	}

	// create stateDB folder in environment folder
	stateDBPath := envPath + "/stateDB"
	err = os.MkdirAll(stateDBPath, 0o755)
	if err != nil {
		logger.Error("Error creating stateDB folder: ", err)
		os.Exit(1)
	}

	// create transactionDB folder in environment folder
	transactionDBPath := envPath + "/transactionDB"
	err = os.MkdirAll(transactionDBPath, 0o755)
	if err != nil {
		logger.Error("Error creating transactionDB folder: ", err)
		os.Exit(1)
	}

	// create monitorDB folder in environment folder
	monitorDBPath := envPath + "/monitorDB"
	err = os.MkdirAll(monitorDBPath, 0o755)
	if err != nil {
		logger.Error("Error creating monitorDB folder: ", err)
		os.Exit(1)
	}

	// create keystores folder in environment folder
	keystoresPath := envPath + "/keystores"
	err = os.MkdirAll(keystoresPath, 0o755)
	if err != nil {
		logger.Error("Error creating keystores folder: ", err)
		os.Exit(1)
	}

	// create keys folder in keystores folder
	keysPath := keystoresPath + "/keys"
	err = os.MkdirAll(keysPath, 0o755)
	if err != nil {
		logger.Error("Error creating keys folder: ", err)
		os.Exit(1)
	}

	// create config.toml file as text file in environment folder
	configPath := envPath + "/config.toml"
	configFile, err := os.Create(configPath)
	if err != nil {
		logger.Error("Error creating config.toml file: ", err)
		os.Exit(1)
	}

	var chainId string
	switch environment {
	case "testnet":
		chainId = "42"
	case "mainnet":
		chainId = "21"
	default:
		chainId = "21"
	}

	// write config.toml file
	configFile.WriteString("[loglevel]\n")
	configFile.WriteString("# Controls the logging level of the alicenet sub-services. To get the complete\n")
	configFile.WriteString("# list of supported services check ./constants/shared.go\n")
	configFile.WriteString(`admin = "info"` + "\n")
	configFile.WriteString(`blockchain = "info"` + "\n")
	configFile.WriteString(`consensus = "info"` + "\n")
	configFile.WriteString("[chain]\n")
	configFile.WriteString("# AliceNet ChainID that corresponds to the AliceNet Network that you are trying" + "\n")
	configFile.WriteString("# to connect." + "\n")
	configFile.WriteString(`id = "` + chainId + `"` + "\n")
	configFile.WriteString("# Path to the location where to save the state database. The state database is" + "\n")
	configFile.WriteString("# responsible for storing the AliceNet blockchain data (blocks, validator's" + "\n")
	configFile.WriteString("# data). Store this database in a safe location. If this database is deleted," + "\n")
	configFile.WriteString("# the node will re-sync with its peers from scratch. DON'T DELETE THIS DATABASE" + "\n")
	configFile.WriteString("# IF YOU ARE RUNNING A VALIDATOR NODE!!! If this database is deleted and you" + "\n")
	configFile.WriteString("# are running a validator node, the validator's data will be permanently" + "\n")
	configFile.WriteString("# deleted and the node will not be able to proceed with its work as a" + "\n")
	configFile.WriteString("# validator, even after a re-sync. Therefore, you may be susceptible to a" + "\n")
	configFile.WriteString("# slashing event." + "\n")
	configFile.WriteString(`stateDB = "` + stateDBPath + `"` + "\n")
	configFile.WriteString("# Path to the location where to save the transaction database. The transaction" + "\n")
	configFile.WriteString("# database is responsible for storing the AliceNet blockchain data" + "\n")
	configFile.WriteString("# (transactions). Store this database in a safe location. If this database is" + "\n")
	configFile.WriteString("# deleted, the node will re-sync all transactions with its peers." + "\n")
	configFile.WriteString(`transactionDB = "` + transactionDBPath + `"` + "\n")
	configFile.WriteString("# Path to the location where to save the monitor database. The monitor database" + "\n")
	configFile.WriteString("# is responsible for storing the (events, tasks, receipts) coming from layer 1" + "\n")
	configFile.WriteString("# blockchains that AliceNet is anchored with. Store this database in a safe" + "\n")
	configFile.WriteString("# location. If this database is deleted, the node will replay all events by" + "\n")
	configFile.WriteString("# querying the layer1 chains using the information provided below." + "\n")
	configFile.WriteString(`monitorDB = "` + monitorDBPath + `"` + "\n")
	configFile.WriteString("# Flags to save any of the databases above only on memory. USE ONLY RECOMMENDED" + "\n")
	configFile.WriteString("# TO SET TRUE FOR TESTING PURPOSES." + "\n")
	configFile.WriteString("monitorDBInMemory = false" + "\n")
	configFile.WriteString("stateDBInMemory = false" + "\n")
	configFile.WriteString("transactionDBInMemory = false" + "\n")
	configFile.WriteString("[transport]" + "\n")
	configFile.WriteString("# IF UPNP should be used to discover opened ports to connect with the peers." + "\n")
	configFile.WriteString("upnp = false" + "\n")
	configFile.WriteString("# 16 Byte private key that is used to encrypt and decrypt information shared" + "\n")
	configFile.WriteString("# with peers. Generate this with a safe random generator." + "\n")
	configFile.WriteString(`privateKey = "<16_BYTES_TRASNPORT_PRIVATE_KEY>"` + "\n")
	configFile.WriteString("# Address to a bootnode running on the desired AliceNet network that you are" + "\n")
	configFile.WriteString("# trying to connect with. A bootnode is a software client responsible for" + "\n")
	configFile.WriteString("# sharing information about alicenet peers. Your node will connect to a" + "\n")
	configFile.WriteString("# bootnode to retrieve an initial list of peers to try to connect with." + "\n")
	configFile.WriteString(`bootNodeAddresses = "<BOOTNODE_ADDRESS>"` + "\n")
	configFile.WriteString("# Maximum number of peers that we can retrieve from the bootnode?" + "\n")
	configFile.WriteString("originLimit = 50" + "\n")
	configFile.WriteString("# Address and port where your node will be listening for rpc requests." + "\n")
	configFile.WriteString(`localStateListeningAddress = "0.0.0.0:8883"` + "\n")
	configFile.WriteString("# Address and port where you node will be listening for requests coming from" + "\n")
	configFile.WriteString("# other peers. The address should be publicly reachable." + "\n")
	configFile.WriteString(`p2pListeningAddress = "0.0.0.0:4342"` + "\n")
	configFile.WriteString("# Maximum number of peers that you wish to be connected with. Upper bound to" + "\n")
	configFile.WriteString("# limit bandwidth shared with the peers." + "\n")
	configFile.WriteString("peerLimitMax = 24" + "\n")
	configFile.WriteString("# Minimum number of peers that you wish to be connect with, before trying to" + "\n")
	configFile.WriteString("# attempt to download blockchain data and participate in consensus." + "\n")
	configFile.WriteString("peerLimitMin = 3" + "\n")
	configFile.WriteString("[ethereum]" + "\n")
	configFile.WriteString("# Ethereum endpoint url to the ethereum chain where the AliceNet network" + "\n")
	configFile.WriteString("# infra-structure that you are trying to connect lives. Ethereum mainnet for" + "\n")
	configFile.WriteString("# AliceNet mainnet and Goerli for Alicenet Testnet. Infura and Alchemy services" + "\n")
	configFile.WriteString("# can be used, but if you are running your own validator node, we recommend to" + "\n")
	configFile.WriteString("# use a more secure node." + "\n")
	configFile.WriteString(`endpoint = "<ETHEREUM_ENDPOINT_URL>"` + "\n")
	configFile.WriteString("# Minimum number of peers connected to your ethereum node that you wish to" + "\n")
	configFile.WriteString("# reach before trying to process ethereum blocks to retrieve the AliceNet" + "\n")
	configFile.WriteString("# events." + "\n")
	configFile.WriteString("endpointMinimumPeers = 1" + "\n")
	configFile.WriteString("# Ethereum address that will be used to sign transactions and connect to the" + "\n")
	configFile.WriteString("# AliceNet services on ethereum." + "\n")
	configFile.WriteString(`defaultAccount = "<0xETHEREUM_ADDRESS>"` + "\n")
	configFile.WriteString("# Path to the encrypted private key used on the address above." + "\n")
	configFile.WriteString(`keystore = "` + keysPath + `"` + "\n")
	configFile.WriteString("# Path to the file containing the password to unlock the account private key." + "\n")
	configFile.WriteString(`passCodes = "` + keystoresPath + `/passcodes.txt"` + "\n")
	configFile.WriteString("# Ethereum address of the AliceNet factory of smart contracts. The factory is" + "\n")
	configFile.WriteString("# responsible for registering and deploying every contract used by the AliceNet" + "\n")
	configFile.WriteString("# infra-structure." + "\n")
	configFile.WriteString(`factoryAddress = "<0xFACTORY_ETHEREUM_ADDRESS>"` + "\n")
	configFile.WriteString("# The ethereum block where the AliceNet contract factory was deployed. This" + "\n")
	configFile.WriteString("# block is used as starting block to retrieve all events (e.g snapshots," + "\n")
	configFile.WriteString("# deposits) necessary to initialize and validate your AliceNet node." + "\n")
	configFile.WriteString("startingBlock = <StartingBlock>" + "\n")
	configFile.WriteString("# Batch size of blocks that will be downloaded and processed from your endpoint" + "\n")
	configFile.WriteString("# address. Every ethereum block starting from the `startingBlock` until the" + "\n")
	configFile.WriteString("# latest ethereum block will be downloaded and all events (e.g snapshots," + "\n")
	configFile.WriteString("# deposits) coming from AliceNet smart contracts will be processed in a" + "\n")
	configFile.WriteString("# chronological order. If this value is too large, your endpoint may end up" + "\n")
	configFile.WriteString("# being overloaded with API requests." + "\n")
	configFile.WriteString("processingBlockBatchSize = 1_000" + "\n")
	configFile.WriteString("# The maximum gas price that you are willing to pay (in GWEI) for a transaction" + "\n")
	configFile.WriteString("# done by your node. If you are validator, putting this value too low can" + "\n")
	configFile.WriteString("# result in your node failing to fulfill the validators duty, hence, being" + "\n")
	configFile.WriteString("# passive for a slashing." + "\n")
	configFile.WriteString("txMaxGasFeeAllowedInGwei = 500" + "\n")
	configFile.WriteString("# Flag to decide if the ethereum transactions information will be shown on the" + "\n")
	configFile.WriteString("# logs." + "\n")
	configFile.WriteString("txMetricsDisplay = false" + "\n")
	configFile.WriteString("[utils]" + "\n")
	configFile.WriteString("# Flag to decide if the status will be shown on the logs. Maybe be a little" + "\n")
	configFile.WriteString("# noisy." + "\n")
	configFile.WriteString("status = true" + "\n")
	configFile.WriteString("# OPTIONAL: Only necessary if you plan to run a validator node." + "\n")
	configFile.WriteString("[validator]" + "\n")
	configFile.WriteString("# Type of elliptic curve used to generate the AliceNet address. 1: secp256k1 (same" + "\n")
	configFile.WriteString("# as ethereum), 2: BN128" + "\n")
	configFile.WriteString("rewardCurveSpec = 1" + "\n")
	configFile.WriteString("# Address of the AliceNet account used to do transactions in the AliceNet" + "\n")
	configFile.WriteString("# network." + "\n")
	configFile.WriteString(`rewardAccount = "0x<ALICENET_ADDRESS>"` + "\n")
	configFile.WriteString(`symmetricKey = "<SOME_SUPER_FANCY_SECRET_THAT_WILL_BE_HASHED>\n"` + "\n")

	logger.Info("created config file")

	os.Exit(0)
}
