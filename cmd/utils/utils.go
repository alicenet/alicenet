package utils

import (
	"context"
	"math/big"
	"os"
	"strings"

	"github.com/alicenet/alicenet/blockchain"
	"github.com/alicenet/alicenet/blockchain/dkg"
	"github.com/alicenet/alicenet/blockchain/dkg/dtest"
	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Command is the cobra.Command specifically for running as an edge node, i.e. not a validator or relay
var Command = cobra.Command{
	Use:   "utils",
	Short: "A collection of tools for node administration",
	Long:  "utils is a misc. collection of tools. Ranges from initial config to automating Ethereum setup",
	Run:   utilsNode}

// EthdkgCommand is the command that triggers a fresh start of the ETHDKG process
var EthdkgCommand = cobra.Command{
	Use:   "ethdkg",
	Short: "",
	Long:  "",
	Run:   utilsNode}

// SendWeiCommand is the command that sends wei from one account to another
var SendWeiCommand = cobra.Command{
	Use:   "sendwei",
	Short: "",
	Long:  "",
	Run:   utilsNode}

func setupEthereum(logger *logrus.Entry) (interfaces.Ethereum, error) {
	logger.Info("Connecting to Ethereum endpoint ...")
	eth, err := blockchain.NewEthereumEndpoint(
		config.Configuration.Ethereum.Endpoint,
		config.Configuration.Ethereum.Keystore,
		config.Configuration.Ethereum.Passcodes,
		config.Configuration.Ethereum.DefaultAccount,
		config.Configuration.Ethereum.Timeout,
		config.Configuration.Ethereum.RetryCount,
		config.Configuration.Ethereum.RetryDelay,
		config.Configuration.Ethereum.FinalityDelay,
		config.Configuration.Ethereum.TxFeePercentageToIncrease,
		config.Configuration.Ethereum.TxMaxFeeThresholdInGwei,
		config.Configuration.Ethereum.TxCheckFrequency,
		config.Configuration.Ethereum.TxTimeoutForReplacement)

	if err != nil {
		return nil, err
	}

	registryAddress := common.HexToAddress(config.Configuration.Ethereum.RegistryAddress)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = eth.Contracts().LookupContracts(ctx, registryAddress)

	return eth, err
}

// LogStatus sends simple info about our Ethereum setup to the logger
func LogStatus(logger *logrus.Entry, eth interfaces.Ethereum) {

	acct := eth.GetDefaultAccount()
	err := eth.UnlockAccount(acct)
	if err != nil {
		logger.Warnf("Failed to unlock account %v: %v", acct.Address.Hex(), err)
		return
	}

	keys, err := eth.GetAccountKeys(acct.Address)
	if err != nil {
		logger.Warnf("Failed to retrieve account %v keys: %v", acct.Address.Hex(), err)
		return
	}

	weiBalance, err := eth.GetBalance(acct.Address)
	if err != nil {
		logger.Warnf("Failed to check ETHER balance account %v: %v", acct.Address.Hex(), err)
		return
	}

	c := eth.Contracts()
	callOpts := eth.GetCallOpts(context.Background(), acct)

	logger.Infof("ValidatorPool() address is %v", c.ValidatorPoolAddress().Hex())
	isValidator, err := c.ValidatorPool().IsValidator(callOpts, acct.Address)
	if err != nil {
		logger.Warnf("Failed checking whether %v is a validator: %v", acct.Address.Hex(), err)
		return
	}

	logger.Info(strings.Repeat("-", 80))
	logger.Infof("          EthDKG contract: %v", c.EthdkgAddress().Hex())
	logger.Infof(" ContractFactory contract: %v", c.ContractFactoryAddress().Hex())
	logger.Infof("  ValidatorsPool contract: %v", c.ValidatorPoolAddress().Hex())
	logger.Info(strings.Repeat("-", 80))
	logger.Infof(" Default Account: %v", acct.Address.Hex())
	logger.Infof("              Public key: 0x%x", crypto.FromECDSAPub(&keys.PrivateKey.PublicKey))
	logger.Infof("             Wei balance: %v", weiBalance)
	logger.Infof("            Is Validator: %v", isValidator)
	logger.Info(strings.Repeat("-", 80))
}

func utilsNode(cmd *cobra.Command, args []string) {

	logger := logging.GetLogger("utils").WithField("Component", cmd.Use)

	// Utils wide setup
	eth, err := setupEthereum(logger)
	if err != nil {
		logger.Errorf("Could not connect to Ethereum: %v", err)
		return
	}

	if config.Configuration.Utils.Status {
		LogStatus(logger, eth)
	}

	// Unlock the default account setup
	acct := eth.GetDefaultAccount()
	err = eth.UnlockAccount(acct)
	if err != nil {
		logger.Errorf("Can not unlock account %v: %v", acct.Address.Hex(), err)
	}

	// Route command
	var exitCode int
	switch cmd.Use {
	case "ethdkg":
		exitCode = ethdkg(logger, eth, cmd, args)
	case "sendwei":
		exitCode = sendwei(logger, eth, cmd, args)
	case "utils":
		exitCode = 0
	default:
		logger.Errorf("Could not find handler for %v", cmd.Use)
		exitCode = 1
	}

	os.Exit(exitCode)
}

func sendwei(logger *logrus.Entry, eth interfaces.Ethereum, cmd *cobra.Command, args []string) int {

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
		_, err := eth.TransferEther(from.Address, common.HexToAddress(args[idx]), wei)
		if err != nil {
			logger.Errorf("Transfer failed: %v", err)
			return 1
		}
	}

	return 0
}

func ethdkg(logger *logrus.Entry, eth interfaces.Ethereum, cmd *cobra.Command, args []string) int {

	// Ethereum setup
	acct := eth.GetDefaultAccount()
	ctx := context.Background()

	txnOpts, err := eth.GetTransactionOpts(ctx, acct)
	if err != nil {
		logger.Errorf("Can not build transaction options: %v", err)
		return 1
	}

	_, rcpt, err := dkg.InitializeETHDKG(eth, txnOpts, ctx)
	if err != nil {
		logger.Errorf("could not initialize ETHDKG: %v", err)
		return 1
	}

	logs := rcpt.Logs

	logger.Infof("Found %v log events after initializing ethdkg", len(logs))

	_, err = dtest.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	if err != nil {
		logger.Errorf("could not get ETHDKG RegistrationOpened event: %v", err)
		return 1
	}

	return 0
}
