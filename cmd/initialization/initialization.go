package initialization

import (
	"github.com/alicenet/alicenet/cmd/ethkey"
	"os"

	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/logging"
	toml "github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

// Command is the cobra.Command specifically for initializing the alicenet client.
var Command = cobra.Command{
	Use:   "init",
	Short: "Initialize the files/folders required for running the alicenet client",
	Long:  "Initialize the files/folders required for running the alicenet client",
	Run:   initialiseFilesAndFolders,
}

const passcodesFile = "/passcodes.txt"

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

	var chainId int
	switch network {
	case "testnet":
		chainId = 42
	case "mainnet":
		chainId = 21
	default:
		logger.Fatal("Invalid network specified - must be either testnet or mainnet")
	}

	logger.Info("Initializing AliceNet configuration files and folders...")

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
			// remove the root folder and its subfolders then exit
			err := os.RemoveAll(alicenetPath)
			if err != nil {
				logger.WithError(err).Error("Failed to remove path: ", alicenetPath)
				logger.Fatal("Please remove the root folder and its subfolders manually")
			}
			logger.Fatal("Exiting")
		}
	}

	// create the keyfile if cancelling flag not specified
	defaultAccount := "<0xETHEREUM_ADDRESS>"
	if !config.Configuration.Initialization.DontGenerateEthkey {
		keyJSON, key, passphrase, err := ethkey.GenerateKeyFile(logger)
		if err != nil {
			logger.Fatalf(err.Error())
		}

		defaultAccount = key.Address.Hex()
		keyFilePath := keysPath + "/" + defaultAccount
		if err := os.WriteFile(keyFilePath, keyJSON, 0600); err != nil {
			logger.Fatalf("Failed to write keyfile to %s: %v", keyFilePath, err)
		}

		passphraseData := []byte(key.Address.Hex() + "=" + passphrase)
		passphraseFilePath := keystoresPath + passcodesFile
		if err := os.WriteFile(passphraseFilePath, passphraseData, 0600); err != nil {
			logger.Fatalf("Failed to write passphrase to %s: %v", passphraseFilePath, err)
		}
	}

	config := &config.RootSerializableConfiguration{
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
			PrivateKey:                 "<16_BYTES_TRANSPORT_PRIVATE_KEY>",
			BootNodeAddresses:          "<BOOTNODE_ADDRESS>",
			OriginLimit:                50,
			LocalStateListeningAddress: "0.0.0.0:8883",
			P2PListeningAddress:        "0.0.0.0:4342",
			PeerLimitMax:               24,
			PeerLimitMin:               3,
		},
		Ethereum: config.EthereumConfig{
			Endpoint:                 "<ETHEREUM_ENDPOINT_URL>",
			EndpointMinimumPeers:     1,
			DefaultAccount:           defaultAccount,
			Keystore:                 keystoresPath,
			PassCodes:                keystoresPath + passcodesFile,
			FactoryAddress:           "<0xFACTORY_ETHEREUM_ADDRESS>",
			StartingBlock:            0,
			ProcessingBlockBatchSize: 1_000,
			TxMaxGasFeeAllowedInGwei: 500,
			TxMetricsDisplay:         false,
		},
		Utils: config.UtilsConfig{
			Status: true,
		},
		Validator: config.ValidatorConfig{
			RewardCurveSpec: 1,
			RewardAccount:   "0x<ALICENET_ADDRESS>",
			SymmetricKey:    "<SOME_SUPER_FANCY_SECRET_THAT_WILL_BE_HASHED>",
		},
	}
	b, err := toml.Marshal(config)
	if err != nil {
		logger.WithError(err).Fatal("Failed to marshal config")
	}

	// create config.toml file as text file in environment folder
	configFile, err := os.Create(configPath)
	if err != nil {
		logger.Error("Error creating config.toml file: ", err)
		os.Exit(1)
	}

	// write config.toml file
	_, err = configFile.Write(b)
	if err != nil {
		logger.WithError(err).Fatal("Failed to write to file")
	}
	err = configFile.Close()
	if err != nil {
		logger.WithError(err).Fatal("Failed to close the file")
	}

	logger.Info("created config file")

	os.Exit(0)
}
