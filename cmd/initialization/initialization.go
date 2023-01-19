package initialization

import (
	"fmt"
	"github.com/alicenet/alicenet/cmd/ethkey"
	"os"

	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/logging"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Command is the cobra.Command specifically for initializing the alicenet client.
var Command = cobra.Command{
	Use:   "init",
	Short: "Initialize the files/folders required for running the alicenet client",
	Long:  "Initialize the files/folders required for running the alicenet client",
	Run:   initializeFilesAndFolders,
}

const (
	passcodesFile         = "/passcodes.txt"
	mainNetFactoryAddress = "0x758a3B3D8958d3794F2Def31e943Cdc449bB2FB9"
	mainNetStartingBlock  = 15540020
)

func initializeFilesAndFolders(cmd *cobra.Command, args []string) {
	logger := logging.GetLogger("init").WithField("Component", cmd.Use)

	rootPath := config.Configuration.Initialization.Path
	network := config.Configuration.Initialization.Network

	if rootPath == "" {
		logger.Info("No path specified - defaulting to home directory")
		rootPath = os.Getenv("HOME")
	} else if rootPath == "./" || rootPath == "." {
		var err error
		rootPath, err = os.Getwd()
		if err != nil {
			logger.WithError(err).Fatalf("Could not determine absolute path for: %v", rootPath)
		}
	}

	if network == "" {
		logger.Info("No environment specified - defaulting to mainnet")
		network = "mainnet"
	}

	var chainId int
	var startingBlock uint64
	var factoryAddress string
	switch network {
	case "testnet":
		chainId = 42
		startingBlock = 0
		factoryAddress = "<0xFACTORY_ETHEREUM_ADDRESS>"
	case "mainnet":
		chainId = 21
		startingBlock = mainNetStartingBlock
		factoryAddress = mainNetFactoryAddress
	default:
		logger.Fatal("Invalid network specified - must be either testnet or mainnet")
	}

	logger.Info("Initializing AliceNet configuration files and folders...")

	// alicenet related paths and files
	alicenetPath := path.Join(rootPath, ".alicenet")
	envPath := path.Join(alicenetPath, network)
	stateDBPath := path.Join(envPath, "stateDB")
	transactionDBPath := path.Join(envPath, "transactionDB")
	monitorDBPath := path.Join(envPath, "monitorDB")
	keystoresPath := path.Join(envPath, "keystores")
	keysPath := path.Join(keystoresPath, "keys")
	configPath := path.Join(envPath, "config.toml")

	paths := []string{envPath, stateDBPath, transactionDBPath, monitorDBPath, keystoresPath, keysPath}

	// loop through the paths checking to see if they exist and exit if any of them do
	for _, path := range paths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			logger.WithError(err).Error("Path already exists: ", path)
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
		err := os.MkdirAll(path, 0o700)
		if err != nil {
			logger.WithError(err).Error("Failed to create path: ", path)
			removePath(logger, alicenetPath)
		}
	}

	// create the keyfile if cancelling flag not specified
	var err error
	defaultAccount := "<0xETHEREUM_ADDRESS>"
	generatePrivateKey := true
	if !config.Configuration.Initialization.GenerateKeys {
		generatePrivateKey, err = ethkey.ReadYesOrNoAnswer("Do you wish to create your address and private key? Yes/no: ")
		if err != nil {
			logger.Fatalf(err.Error())
		}
	}

	if generatePrivateKey {
		keyJSON, key, passphrase, err := ethkey.GenerateKeyFile(config.Configuration.Initialization.GenerateKeys, logger)
		if err != nil {
			logger.Fatalf(err.Error())
		}

		defaultAccount = key.Address.Hex()
		keyFilePath := keysPath + "/" + defaultAccount
		if err := os.WriteFile(keyFilePath, keyJSON, 0600); err != nil {
			logger.Fatalf("Failed to write keyfile to %s: %v", keyFilePath, err)
		}

		fmt.Printf("The following Ethereum address was generated and saved as your default account: %s.\n", key.Address.Hex())
		fmt.Printf("Your private key is stored in: %s. Please maintain this file secure in order to protect your assets.\n", keyFilePath)

		savePasscodesFile, err := ethkey.ReadYesOrNoAnswer("Do you wish to store the password in a file? Yes/no: ")
		if err != nil {
			logger.Fatalf(err.Error())
		}

		if savePasscodesFile {
			passphraseData := []byte(key.Address.Hex() + "=" + passphrase)
			passphraseFilePath := keystoresPath + passcodesFile
			if err := os.WriteFile(passphraseFilePath, passphraseData, 0600); err != nil {
				logger.Fatalf("Failed to write passphrase to %s: %v", passphraseFilePath, err)
			}
			fmt.Printf("The password that was used to generate the private key is stored in: %s. Please maintain this file secure in order to protect your assets.\n", passphraseFilePath)
		}
	} else {
		fmt.Printf("In order to configure your node properly, please save your private key to the following path %s using your address as file name.\n", keysPath)
	}

	transportPrivateKey := "<16_BYTES_TRANSPORT_PRIVATE_KEY>"
	tpk, err := ethkey.GenerateRandomString(24)
	if err != nil {
		logger.Fatalf("Failed to generate Transport.PrivateKey with error %v", err)
	}
	transportPrivateKey = tpk

	validatorSymmetricKey := "<SOME_SUPER_FANCY_SECRET_THAT_WILL_BE_HASHED>"
	vspk, err := ethkey.GenerateRandomString(32)
	if err != nil {
		logger.Fatalf("Failed to generate Validator.SymmetricKey with error %v", err)
	}
	validatorSymmetricKey = vspk

	ethereumEndpointURL := config.Configuration.Ethereum.Endpoint
	if ethereumEndpointURL == "" {
		saveEthereumEndpoint, err := ethkey.ReadYesOrNoAnswer("Do you wish to enter Ethereum endpoint? Yes/no: ")
		if err != nil {
			logger.Fatalf(err.Error())
		}

		if saveEthereumEndpoint {
			ee, err := ethkey.ReadInput("Please enter Ethereum endpoint: ")
			if err != nil {
				logger.Fatalf(err.Error())
			}
			ethereumEndpointURL = ee
		} else {
			ethereumEndpointURL = "<ETHEREUM_ENDPOINT_URL>"
			fmt.Println(fmt.Sprintf("In order to configure your node properly, please save the Ethereum endpoint to the following file %s.", configPath))
		}
	}

	configObj := &config.RootConfiguration{
		Logging: config.LoggingConfig{
			Consensus: "info",
		},
		Chain: config.ChainConfig{
			ID:                    chainId,
			StateDbPath:           stateDBPath,
			StateDbInMemory:       false,
			TransactionDbPath:     transactionDBPath,
			TransactionDbInMemory: false,
			MonitorDbPath:         monitorDBPath,
			MonitorDbInMemory:     false,
		},
		Transport: config.TransportConfig{
			UPnP:                       false,
			PrivateKey:                 transportPrivateKey,
			BootNodeAddresses:          "<BOOTNODE_ADDRESS>",
			OriginLimit:                3,
			LocalStateListeningAddress: "0.0.0.0:8883",
			P2PListeningAddress:        "0.0.0.0:4342",
			PeerLimitMax:               24,
			PeerLimitMin:               3,
		},
		Ethereum: config.EthereumConfig{
			Endpoint:                 ethereumEndpointURL,
			EndpointMinimumPeers:     1,
			DefaultAccount:           defaultAccount,
			Keystore:                 keysPath,
			PassCodes:                path.Join(keystoresPath, "/passcodes.txt"),
			FactoryAddress:           factoryAddress,
			StartingBlock:            startingBlock,
			ProcessingBlockBatchSize: 1_000,
			TxMaxGasFeeAllowedInGwei: 500,
			TxMetricsDisplay:         false,
		},
		Utils: config.UtilsConfig{
			Status: true,
		},
		Validator: config.ValidatorConfig{
			SymmetricKey: validatorSymmetricKey,
		},
	}

	configBytes, err := config.CreateTOML(configObj)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal config")
		removePath(logger, alicenetPath)
	}

	// create config.toml file as text file in environment folder
	err = os.WriteFile(configPath, configBytes, 0600)
	if err != nil {
		logger.WithError(err).Error("Error creating config.toml file")
		removePath(logger, alicenetPath)
	}

	logger.Info("Created config file")
}

// removePath removes the root folder and its sub-folders then exit
func removePath(logger *logrus.Entry, alicenetPath string) {
	err := os.RemoveAll(alicenetPath)
	if err != nil {
		logger.WithError(err).Error("Failed to remove path: ", alicenetPath)
		logger.Fatal("Please remove the root folder and its sub-folders manually")
	}
	logger.Fatal("Exiting")
}
