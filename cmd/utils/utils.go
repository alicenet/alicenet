package utils

import (
	"context"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/handlers"
	"github.com/alicenet/alicenet/logging"
)

// Command is the cobra.Command specifically for running as an edge node, i.e. not a validator or relay.
var Command = cobra.Command{
	Use:   "utils",
	Short: "A collection of tools for node administration",
	Long:  "utils is a misc. collection of tools. Ranges from initial config to automating Ethereum setup",
	Run:   utilsNode,
}

// SendWeiCommand is the command that sends wei from one account to another.
var SendWeiCommand = cobra.Command{
	Use:   "sendwei",
	Short: "",
	Long:  "",
	Run:   utilsNode,
}

func setupEthereum(logger *logrus.Entry) (layer1.Client, layer1.AllSmartContracts, error) {
	logger.Info("Connecting to Ethereum endpoint ...")
	eth, err := ethereum.NewClient(
		config.Configuration.Ethereum.Endpoint,
		config.Configuration.Ethereum.Keystore,
		config.Configuration.Ethereum.PassCodes,
		config.Configuration.Ethereum.DefaultAccount,
		true,
		constants.EthereumFinalityDelay,
		config.Configuration.Ethereum.TxMaxGasFeeAllowedInGwei,
		config.Configuration.Ethereum.EndpointMinimumPeers,
	)
	if err != nil {
		return nil, nil, err
	}

	factoryAddress := common.HexToAddress(config.Configuration.Ethereum.FactoryAddress)

	// Initialize and find all the contracts
	contractsHandler := handlers.NewAllSmartContractsHandle(eth, factoryAddress)

	return eth, contractsHandler, err
}

// LogStatus sends simple info about our Ethereum setup to the logger.
func LogStatus(logger *logrus.Entry, eth layer1.Client, contracts layer1.AllSmartContracts) {
	acct := eth.GetDefaultAccount()

	weiBalance, err := eth.GetBalance(acct.Address)
	if err != nil {
		logger.Warnf("Failed to check ETHER balance account %v: %v", acct.Address.Hex(), err)
		return
	}

	c := contracts.EthereumContracts()
	callOpts, err := eth.GetCallOpts(context.Background(), acct)
	if err != nil {
		logger.Warnf("Failed to get call options: %v", err)
		return
	}

	isValidator, err := c.ValidatorPool().IsValidator(callOpts, acct.Address)
	if err != nil {
		logger.Warnf("Failed checking whether %v is a validator: %v", acct.Address.Hex(), err)
		return
	}

	logger.Info(strings.Repeat("-", 80))
	logger.Info("        	ETHEREUM CONTRACTS")
	logger.Info(strings.Repeat("-", 80))
	logger.Infof("          EthDKG contract: %v", c.EthdkgAddress().Hex())
	logger.Infof(" ContractFactory contract: %v", c.ContractFactoryAddress().Hex())
	logger.Infof("  ValidatorsPool contract: %v", c.ValidatorPoolAddress().Hex())
	logger.Info(strings.Repeat("-", 80))
	logger.Infof(" Default Account: %v", acct.Address.Hex())
	logger.Infof("             Wei balance: %v", weiBalance)
	logger.Infof("            Is Validator: %v", isValidator)
	logger.Info(strings.Repeat("-", 80))
}

func utilsNode(cmd *cobra.Command, args []string) {
	logger := logging.GetLogger("utils").WithField("Component", cmd.Use)

	// Utils wide setup
	eth, contracts, err := setupEthereum(logger)
	if err != nil {
		logger.Errorf("Could not connect to Ethereum: %v", err)
		return
	}

	if config.Configuration.Utils.Status {
		LogStatus(logger, eth, contracts)
	}

	// Route command
	var exitCode int
	switch cmd.Use {
	case "sendwei":
		exitCode = sendWei(logger, eth, cmd, args)
	case "utils":
		exitCode = 0
	default:
		logger.Errorf("Could not find handler for %v", cmd.Use)
		exitCode = 1
	}

	os.Exit(exitCode)
}

func sendWei(logger *logrus.Entry, eth layer1.Client, cmd *cobra.Command, args []string) int {
	if len(args) < 2 {
		logger.Errorf("Arguments must include: amount, who\nwho can be a space delimited list of addresses")
		return 1
	}

	wei, ok := new(big.Int).SetString(args[0], 10)
	if !ok {
		logger.Errorf("Could not parse wei amount (base 10).")
		return 1
	}

	from := eth.GetDefaultAccount()
	for idx := 1; idx < len(args); idx++ {
		_, err := ethereum.TransferEther(eth, logger, from.Address, common.HexToAddress(args[idx]), wei)
		if err != nil {
			logger.Errorf("Transfer failed: %v", err)
			return 1
		}
	}

	return 0
}
