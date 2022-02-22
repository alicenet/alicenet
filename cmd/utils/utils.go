package utils

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/monitor"
	"github.com/MadBase/MadNet/config"
	"github.com/MadBase/MadNet/logging"
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

// ApproveTokensCommand is the command that approves the transfer of tokens to the staking contract
var ApproveTokensCommand = cobra.Command{
	Use:   "approvetokens",
	Short: "",
	Long:  "",
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

// RegisterCommand is the command the stakes tokens
var RegisterCommand = cobra.Command{
	Use:   "register",
	Short: "",
	Long:  "",
	Run:   utilsNode}

// TransferTokensCommand is the command that should follow ApproveTokens and does the actual transfer
var TransferTokensCommand = cobra.Command{
	Use:   "transfertokens",
	Short: "",
	Long:  "",
	Run:   utilsNode}

// UnregisterCommand is the command the requests the caller is removed from validator pool
var UnregisterCommand = cobra.Command{
	Use:   "unregister",
	Short: "Removes the default account from the validator pool",
	Long:  "",
	Run:   utilsNode}

//
var UpdateValueCommand = cobra.Command{
	Use:   "updatevalue",
	Short: "Dynamically updates a value",
	Long:  "",
	Run:   utilsNode}

// DepositCommand is the command that triggers a token deposit for the caller
var DepositCommand = cobra.Command{
	Use:   "deposit",
	Short: "Creates a token deposit into the sidechain",
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
		config.Configuration.Ethereum.FinalityDelay)

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
	stakingTokenBalance, err := c.StakingToken().BalanceOf(callOpts, acct.Address)
	if err != nil {
		logger.Warnf("Failed to check staking token (%v) balance account %v: %v", c.StakingTokenAddress().Hex(), acct.Address.Hex(), err)
		return
	}

	utilityTokenBalance, err := c.UtilityToken().BalanceOf(callOpts, acct.Address)
	if err != nil {
		logger.Warnf("Failed to check utility token (%v) balance account %v: %v", c.UtilityTokenAddress().Hex(), acct.Address.Hex(), err)
		return
	}

	logger.Infof("ValidatorPool() address is %v", c.ValidatorPoolAddress().Hex())
	isValidator, err := c.ValidatorPool().IsValidator(callOpts, acct.Address)
	if err != nil {
		logger.Warnf("Failed checking whether %v is a validator: %v", acct.Address.Hex(), err)
		return
	}

	rewardBalance, err := c.Staking().BalanceReward(callOpts)
	if err != nil {
		logger.Warnf("Failed to check balance: %v", err)
	}

	unlockedRewardBalance, err := c.Staking().BalanceUnlockedReward(callOpts)
	if err != nil {
		logger.Warnf("Failed to check balance: %v", err)
	}

	stakeBalance, err := c.Staking().BalanceStake(callOpts)
	if err != nil {
		logger.Warnf("Failed to check balance: %v", err)
	}

	unlockedBalance, err := c.Staking().BalanceUnlocked(callOpts)
	if err != nil {
		logger.Warnf("Failed to check balance: %v", err)
	}

	epoch, err := c.Validators().Epoch(callOpts)
	if err != nil {
		logger.Warnf("Failed to find current epoch: %v", err)
	}

	logger.Info(strings.Repeat("-", 80))
	logger.Infof("          Crypto contract: %v", c.CryptoAddress().Hex())
	logger.Infof("         Deposit contract: %v", c.DepositAddress().Hex())
	logger.Infof("          EthDKG contract: %v", c.EthdkgAddress().Hex())
	logger.Infof("*       Registry contract: %v", c.RegistryAddress().Hex())
	logger.Infof("    StakingToken contract: %v", c.StakingTokenAddress().Hex())
	logger.Infof("        Governor contract: %v", c.GovernorAddress().Hex())
	logger.Infof("      Validators contract: %v", c.ValidatorsAddress().Hex())
	logger.Infof("  ValidatorsPool contract: %v", c.ValidatorPoolAddress().Hex())
	logger.Info(strings.Repeat("-", 80))
	logger.Infof(" Default Account: %v", acct.Address.Hex())
	logger.Infof("              Public key: 0x%x", crypto.FromECDSAPub(&keys.PrivateKey.PublicKey))
	logger.Infof("             Wei balance: %v", weiBalance)
	logger.Infof("   Staking token balance: %v", stakingTokenBalance)
	logger.Infof("   Utility token balance: %v", utilityTokenBalance)
	logger.Infof("            Is Validator: %v", isValidator)
	logger.Infof("           Stake balance: %v", stakeBalance)
	logger.Infof("  Unlocked Stake balance: %v", unlockedBalance)
	logger.Infof("   Locked Reward balance: %v", rewardBalance)
	logger.Infof(" Unlocked Reward balance: %v", unlockedRewardBalance)
	logger.Infof("                   Epoch: %v", epoch)
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
	case "approvetokens":
		exitCode = approvetokens(logger, eth, cmd, args)
	case "ethdkg":
		exitCode = ethdkg(logger, eth, cmd, args)
	case "register":
		exitCode = register(logger, eth, cmd, args)
	case "sendwei":
		exitCode = sendwei(logger, eth, cmd, args)
	case "unregister":
		exitCode = unregister(logger, eth, cmd, args)
	case "utils":
		exitCode = 0
	case "updatevalue":
		exitCode = updatevalue(logger, eth, cmd, args)
	case "transfertokens":
		exitCode = transfertokens(logger, eth, cmd, args)
	case "deposit":
		exitCode = deposittokens(logger, eth, cmd, args)
	default:
		logger.Errorf("Could not find handler for %v", cmd.Use)
		exitCode = 1
	}

	os.Exit(exitCode)
}

func register(logger *logrus.Entry, eth interfaces.Ethereum, cmd *cobra.Command, args []string) int {

	// More ethereum setup
	acct := eth.GetDefaultAccount()
	eth.GetCoinbaseAddress()

	if acct.Address.String() == eth.GetCoinbaseAddress().String() {
		logger.Infof("Skipping validator registration for admin acount: %v", acct.Address.String())
		return 0
	}

	c := eth.Contracts()
	ctx := context.Background()

	txnOpts, err := eth.GetTransactionOpts(ctx, acct)
	if err != nil {
		logger.Errorf("Can not build transaction options: %v", err)
	}

	// Contract orchestration
	// Approve tokens for staking
	var maxRetries int = 10
	for nRetries := 0; nRetries < maxRetries; nRetries++ {
		txn, err := c.StakingToken().Approve(txnOpts, c.ValidatorsAddress(), big.NewInt(1_000_000))
		if err != nil {
			logger.Errorf("StakingToken.Approve() failed: %v", err)
			return 1
		}
		rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
		if err != nil {
			logger.Errorf("StakingToken.Approve() failed: %v", err)
			return 1
		}
		if rcpt != nil && rcpt.Status != 1 {
			logger.Errorf("StakingToken.Approve() failed")
		} else {
			break
		}
	}

	// Lock tokens as stake
	for nRetries := 0; nRetries < maxRetries; nRetries++ {
		txn, err := c.Staking().LockStake(txnOpts, big.NewInt(1_000_000))
		if err != nil {
			logger.Errorf("Staking.LockStake() failed: %v", err)
			return 1
		}
		rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
		if err != nil {
			logger.Errorf("Staking.LockStake() failed: %v", err)
			return 1
		}
		if rcpt != nil && rcpt.Status != 1 {
			logger.Errorf("Staking.LockStake() failed")
		} else {
			break
		}
	}

	// Actually join validator pool
	for nRetries := 0; nRetries < maxRetries; nRetries++ {
		txn, err := c.ValidatorPool().AddValidator(txnOpts, acct.Address)
		if err != nil {
			logger.Errorf("Could not add %v as validator: %v", acct.Address.Hex(), err)
		}
		rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
		if err != nil {
			logger.Errorf("Could not add %v as validator: %v", acct.Address.Hex(), err)
		}
		if rcpt != nil && rcpt.Status != 1 {
			logger.Errorf("Validators.AddValidator() failed")
		} else {
			break
		}
	}

	logger.Infof("Registered the address %x", acct.Address.Hex())
	return 0
}

func unregister(logger *logrus.Entry, eth interfaces.Ethereum, cmd *cobra.Command, args []string) int {

	// More ethereum setup
	acct := eth.GetDefaultAccount()
	c := eth.Contracts()

	txnOpts, err := eth.GetTransactionOpts(context.Background(), acct)
	if err != nil {
		logger.Errorf("Can not build transaction options: %v", err)
	}

	// Contract orchestration
	_, err = c.ValidatorPool().RemoveValidator(txnOpts, acct.Address)
	if err != nil {
		logger.Errorf("Account %v could not leave validators pool: %v", acct.Address.Hex(), err)
		return 1
	}
	logger.Infof("Unregistered the address %x", acct.Address.Hex())

	return 0
}

func approvetokens(logger *logrus.Entry, eth interfaces.Ethereum, cmd *cobra.Command, args []string) int {

	// Arguments are 1) who is being approved, and 2) amount being approved
	if len(args) != 2 {
		logger.Errorf("Arguments should be: who, amount")
		return 1
	}

	toAddressString := args[0]
	toAddress := common.HexToAddress(toAddressString)

	amountStr := args[1]
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		logger.Errorf("Could not parse approval amount (base 10).")
		return 1
	}

	// More ethereum setup
	acct := eth.GetDefaultAccount()
	c := eth.Contracts()

	txnOpts, err := eth.GetTransactionOpts(context.Background(), acct)
	if err != nil {
		logger.Errorf("Can not build transaction options: %v", err)
		return 1
	}

	action := func() bool {

		txn, err := c.StakingToken().Approve(txnOpts, toAddress, amount)
		if err != nil && err.Error() == "replacement transaction underpriced" {
			return true
		}

		if err != nil {
			logger.Errorf("failed to approve sending %v to %v: %v", amount, toAddress.Hex(), err)
			return false
		}

		rcpt, err := eth.Queue().QueueAndWait(context.Background(), txn)
		if err != nil {
			logger.Infof("waiting for receipt failed: %v", err)
			return true
		}

		if rcpt != nil && rcpt.Status != 1 {
			logger.Infof("approval receipt success")
			return true
		}

		return false
	}

	for try := true; try; try = action() {
		logger.Infof("retrying...")
		time.Sleep(time.Second)
	}
	logger.Infof("Approved amount %d to address %x", amount, toAddress)

	return 0
}

func updatevalue(logger *logrus.Entry, eth interfaces.Ethereum, cmd *cobra.Command, args []string) int {

	acct := eth.GetDefaultAccount()

	txnOpts, err := eth.GetTransactionOpts(context.Background(), acct)
	if err != nil {
		logger.Errorf("txnopts failed: %v", err)
		return 1
	}

	epoch, valid := new(big.Int).SetString(args[0], 10)
	if !valid {
		logger.Errorf("Could not set epoch using: %v", args[0])
		return 1
	}

	key, valid := new(big.Int).SetString(args[1], 10)
	if !valid {
		logger.Errorf("Could not determine key using: %v", args[1])
		return 1
	}

	value := blockchain.StringToBytes32(args[2])

	logger = logger.WithFields(logrus.Fields{
		"Epoch": epoch,
		"Key":   key,
		"Value": value,
	})

	txn, err := eth.Contracts().Governor().UpdateValue(txnOpts, epoch, key, value)
	if err != nil {
		logger.Errorf("Could not call updateValue: %v", err)
		return 1
	}

	rcpt, err := eth.Queue().QueueAndWait(context.Background(), txn)
	if err != nil {
		logger.Errorf("Could not retrieve receipt: %v", err)
		return 1
	}

	if rcpt.Status != 1 {
		logger.Error("Transaction failed")
		return 1
	}

	logger.Info("Value updated")

	return 0
}

func deposittokens(logger *logrus.Entry, eth interfaces.Ethereum, cmd *cobra.Command, args []string) int {
	// More ethereum setup
	acct := eth.GetDefaultAccount()
	c := eth.Contracts()
	amount := big.NewInt(10000)
	ctx := context.Background()
	txnOpts, err := eth.GetTransactionOpts(ctx, acct)
	if err != nil {
		logger.Errorf("txnopts failed: %v", err)
		return 1
	}
	// approve
	txn, err := c.UtilityToken().Approve(txnOpts, c.DepositAddress(), amount)
	if err != nil {
		logger.Errorf("approval failed: %v", err)
		return 1
	}
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	if err != nil {
		logger.Errorf("approval receipt failed: %v", err)
		return 1
	}
	logger.Infof("approval receipt status: %v", rcpt.Status)
	// deposit
	txn, err = c.Deposit().Deposit(txnOpts, amount)
	if err != nil {
		logger.Errorf("deposit failed: %v", err)
		return 1
	}
	rcpt, err = eth.Queue().QueueAndWait(ctx, txn)
	if err != nil {
		logger.Errorf("deposit receipt failed: %v", err)
		return 1
	}
	logger.Infof("deposit receipt status: %v", rcpt.Status)
	return 0
}

func transfertokens(logger *logrus.Entry, eth interfaces.Ethereum, cmd *cobra.Command, args []string) int {

	// Arguments are 1) src of tokens, and 2) amount to transfer
	if len(args) != 2 {
		logger.Errorf("Arguments should be: who, amount")
		return 1
	}

	fromAddressString := args[0]
	fromAddress := common.HexToAddress(fromAddressString)

	amountStr := args[1]
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		logger.Errorf("Could not parse approval amount (base 10).")
		return 1
	}

	// More ethereum setup
	acct := eth.GetDefaultAccount()
	c := eth.Contracts()

	txnOpts, err := eth.GetTransactionOpts(context.Background(), acct)
	if err != nil {
		logger.Errorf("Can not build transaction options: %v", err)
	}

	// Contract orchestration
	_, err = c.StakingToken().TransferFrom(txnOpts, fromAddress, acct.Address, amount)
	if err != nil {
		logger.Errorf("Could not transfer %v tokens from %v to %v: %v", amount, fromAddressString, acct.Address.Hex(), err)
		return 1
	}
	logger.Infof("Transfered %v tokens from %v to %v", amount, fromAddressString, acct.Address.Hex())

	return 0
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

	// More ethereum setup
	acct := eth.GetDefaultAccount()
	c := eth.Contracts()

	ctx := context.Background()

	txnOpts, err := eth.GetTransactionOpts(ctx, acct)
	if err != nil {
		logger.Errorf("Can not build transaction options: %v", err)
		return 1
	}

	//
	txn, err := c.ValidatorPool().InitializeETHDKG(txnOpts)
	if err != nil {
		logger.Errorf("Could not initialize ethdkg: %v", err)
		return 1
	}

	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	if err != nil {
		logger.Error("Failed looking for transaction events.")
		return 1
	}

	logs := rcpt.Logs

	logger.Infof("Found %v log events after initializing ethdkg", len(logs))

	ethDkgEvents := monitor.GetETHDKGEvents()
	regOpenedEvent, ok := ethDkgEvents["RegistrationOpened"]
	if !ok {
		panic(fmt.Errorf("could not find event named RegistrationOpened"))
	}

	for _, log := range logs {
		if log.Topics[0].Hex() == regOpenedEvent.ID.Hex() {
			event, err := c.Ethdkg().ParseRegistrationOpened(*log)
			logger.Infof("ETHDKG registration is now open...\nDkgStarts:%v\nNonce:%v",
				event.StartBlock,
				event.Nonce)
			if err != nil {
				logger.Warnf("Could not parse RegistrationOpen event: %v", err)
			}
		}
	}

	return 0
}
