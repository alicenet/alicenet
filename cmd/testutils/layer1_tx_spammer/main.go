package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alicenet/alicenet/cmd/testutils/layer1_tx_spammer/dummy_contract"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/evm"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

const (
	numTestAccounts    int           = 10
	spammerTxTime      time.Duration = 2 * time.Second
	networkTimeout     time.Duration = 10 * time.Second
	maxExecutionBlocks uint64        = 8
)

func setupClient(
	endpoint string,
	keyStorePath string,
	passCodePath string,
	defaultAccount string,
	finalityDelay uint64,
) (layer1.Client, *ethclient.Client) {
	eth, err := evm.NewClient(
		logging.GetLogger("ethereum"),
		endpoint,
		keyStorePath,
		passCodePath,
		defaultAccount,
		true,
		finalityDelay,
		500,
		0,
	)
	if err != nil {
		panic(fmt.Errorf("failed to create ethereum client: %v", err))
	}

	internalClient, err := ethclient.Dial(endpoint)
	if err != nil {
		panic(fmt.Errorf("failed to create internal client: %v", err))
	}

	return eth, internalClient
}

func parseKeyStoreAndPassCode(
	tempDir string,
	logger *logrus.Entry,
	keyStorePath string,
	passCodePath string,
	defaultAccountStr string,
) (string, string, string, bool) {
	var defaultAccount common.Address
	if defaultAccountStr != "" {
		defaultAccountBytes, err := hex.DecodeString(defaultAccountStr)
		if err != nil {
			panic(fmt.Errorf("failed to decode default account string: %v", err))
		}
		defaultAccount = common.BytesToAddress(defaultAccountBytes)
	}
	isUsingTestAccounts := false
	if keyStorePath == "" || passCodePath == "" {
		if defaultAccountStr != "" {
			logger.Warn(
				"Ignoring default account sent because keyStorePath or passCodePath were not set!",
			)
		}
		logger.Warn(
			"Keystore or passCodePath not specified, using test accounts instead. If don't want this behavior specify all 2 parameters!",
		)

		var accounts []accounts.Account
		keyStorePath, passCodePath, accounts = tests.CreateAccounts(tempDir, numTestAccounts)
		var accountsAddresses []common.Address
		for _, account := range accounts {
			accountsAddresses = append(accountsAddresses, account.Address)
		}
		logger.Infof("Created the following accounts: %v", accountsAddresses)
		defaultAccount = accounts[0].Address
		isUsingTestAccounts = true
	}

	return keyStorePath, passCodePath, defaultAccount.Hex(), isUsingTestAccounts
}

func initDatabase(ctx context.Context, path string, inMemory bool) *badger.DB {
	db, err := utils.OpenBadger(ctx.Done(), path, inMemory)
	if err != nil {
		panic(err)
	}
	return db
}

func getCustomTransactionOptions(
	ctx context.Context,
	eth layer1.Client,
	fromAccount accounts.Account,
) (*bind.TransactOpts, error) {
	txnOpts, err := eth.GetTransactionOpts(ctx, fromAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %v", err)
	}
	// extract the baseFee
	baseFeeWithOutMultiplier := new(
		big.Int,
	).Div(txnOpts.GasFeeCap, big.NewInt(constants.EthereumBaseFeeMultiplier))
	baseFee := new(big.Int).Sub(baseFeeWithOutMultiplier, txnOpts.GasTipCap)
	// multiply the GasTip by 10
	txnOpts.GasTipCap = new(big.Int).Mul(txnOpts.GasTipCap, big.NewInt(10))
	totalFee := new(big.Int).Add(baseFee, txnOpts.GasTipCap)
	// multiply the GasFeeCap by 10
	txnOpts.GasFeeCap = new(big.Int).Mul(totalFee, big.NewInt(10))
	return txnOpts, nil
}

func deployDummyContract(
	internalClient *ethclient.Client,
	watcher transaction.Watcher,
	txnOpts *bind.TransactOpts,
) error {
	networkCtx, cf := context.WithTimeout(context.Background(), networkTimeout)
	defer cf()
	_, txn, _, err := dummy_contract.DeployDummyContract(txnOpts, internalClient)
	if err != nil {
		return fmt.Errorf("failed to deploy dummy contract: %v", err)
	}
	_, err = watcher.Subscribe(networkCtx, txn, nil)
	if err != nil {
		return fmt.Errorf("failed to deploy dummy contract, err: %v", err)
	}
	return nil
}

func sendEther(
	eth layer1.Client,
	logger *logrus.Entry,
	watcher transaction.Watcher,
	fromAddress common.Address,
) error {
	txn, err := eth.TransferNativeToken(
		fromAddress,
		eth.GetDefaultAccount().Address,
		big.NewInt(100),
	)
	if err != nil {
		return fmt.Errorf("failed to send ether from %v", fromAddress.Hex())
	}
	networkCtx, cf := context.WithTimeout(context.Background(), networkTimeout)
	defer cf()
	_, err = watcher.Subscribe(networkCtx, txn, nil)
	if err != nil {
		return fmt.Errorf(
			"failed to subscribe txn for sendEther FromAccount: %v err: %v",
			fromAddress.Hex(),
			err,
		)
	}
	return nil
}

type WorkerStatus int

const (
	Resting WorkerStatus = iota
	Executing
)

func (status WorkerStatus) String() string {
	return [...]string{
		"Resting",
		"Executing",
	}[status]
}

type WorkScheduler struct {
	CurrentHeight uint64
	Status        WorkerStatus
}

func worker(
	mainCtx context.Context,
	eth layer1.Client,
	internalClient *ethclient.Client,
	watcher transaction.Watcher,
	account accounts.Account,
	maxRestBlocks uint64,
) {
	logger := logging.GetLogger("test").WithFields(logrus.Fields{
		"component": "worker",
		"account":   account.Address.Hex(),
	})
	workScheduler := &WorkScheduler{Status: Executing}
	blockCounter := uint64(0)
	logEntry := logger.WithField("height", workScheduler.CurrentHeight)
	for {
		select {
		case <-mainCtx.Done():
			return
		case <-time.After(spammerTxTime):
		}
		logEntry.Debug("sending ether around")
		err := sendEther(eth, logger, watcher, account.Address)
		if err != nil {
			logger.Error(err)
		}
		networkCtx, cf := context.WithTimeout(context.Background(), networkTimeout)
		defer cf()
		height, err := eth.GetCurrentHeight(networkCtx)
		if err != nil {
			logger.Error(err)
		}
		if height > workScheduler.CurrentHeight {
			logEntry.Debug("got a new block")
			workScheduler.CurrentHeight = height
			blockCounter++
		}
		if workScheduler.Status == Resting {
			if blockCounter >= maxRestBlocks {
				logEntry.Debug("worker will execute in the next block")
				workScheduler.Status = Executing
				blockCounter = 0
			}
		} else {
			if blockCounter > maxExecutionBlocks {
				logEntry.Debug("worker will rest in the next block")
				workScheduler.Status = Resting
				blockCounter = 0
			}
		}
		if workScheduler.Status == Executing {
			txnOpts, err := getCustomTransactionOptions(networkCtx, eth, account)
			if err != nil {
				logger.Error(err)
			} else {
				logEntry.Debug("deploying dummy heavy contract")
				err := deployDummyContract(internalClient, watcher, txnOpts)
				if err != nil {
					logger.Error(err)
				}
			}
		}
	}
}

func main() {
	keyStorePathPtr := flag.String(
		"keyStorePath",
		"",
		"Path to folder with the encrypted private keys. If not provided, test accounts will be used.",
	)
	passCodePathPtr := flag.String(
		"passCodePath",
		"",
		"Path to the file containing the password to decrypt the private key. If not provided, test accounts will be used.",
	)
	defaultAccountPtr := flag.String(
		"defaultAccount",
		"",
		"Account inside the key store that will be used as fallback. If not provided, the first test account will be used.",
	)
	endPointPtr := flag.String(
		"endPoint",
		"http://127.0.0.1:8545",
		"Endpoint to connect with the layer 1 server. If not provided, defaults to 127.0.0.1:8545 ",
	)
	finalityDelayPtr := flag.Int64(
		"finalityDelay",
		12,
		"Number of blocks to wait to consider a transaction final",
	)
	saveStatePtr := flag.Bool(
		"saveState",
		false,
		"If the tracked transaction should be saved to database. The db will be saved on the local folder",
	)
	maxRestBlocksPtr := flag.Int64(
		"maxRestBlocks",
		4,
		"Number of blocks that the workers will not send heavy transactions to let the baseFee decrease. If set to 0, heavy transactions will always be sent.",
	)
	flag.Parse()

	if !strings.Contains(*endPointPtr, "https://") && !strings.Contains(*endPointPtr, "http://") {
		panic("Incorrect endpoint. Endpoint should start with 'http://' or 'https://'")
	}

	logger := logging.GetLogger("test").WithFields(logrus.Fields{
		"keyStorePath":   *keyStorePathPtr,
		"passCodePath":   *passCodePathPtr,
		"defaultAccount": *defaultAccountPtr,
		"finalityDelay":  *finalityDelayPtr,
		"endPoint":       *endPointPtr,
		"saveState":      *saveStatePtr,
		"maxRestBlocks":  *maxRestBlocksPtr,
	})

	logger.Logger.SetLevel(logrus.DebugLevel)

	logger.Info("Starting the spammer...")

	tempDir, err := os.MkdirTemp("", "spammerdir")
	if err != nil {
		panic(fmt.Errorf("failed to create tmp dir: %v", err))
	}
	defer os.RemoveAll(tempDir)

	keyStorePath, passCodePath, defaultAccount, isUsingTestAccounts := parseKeyStoreAndPassCode(
		tempDir,
		logger,
		*keyStorePathPtr,
		*passCodePathPtr,
		*defaultAccountPtr,
	)
	eth, internalClient := setupClient(
		*endPointPtr,
		keyStorePath,
		passCodePath,
		defaultAccount,
		uint64(*finalityDelayPtr),
	)
	defer eth.Close()

	mainCtx, cf := context.WithCancel(context.Background())
	defer cf()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		logger.Warnf("Received shutdown signal %s", sig)
		cf()
	}()

	var recoverDB *db.Database
	if *saveStatePtr {
		logger.Info("Saving state to database")
		rawMonitorDb := initDatabase(mainCtx, "./spammer-database", false)
		defer rawMonitorDb.Close()
		recoverDB = &db.Database{}
		recoverDB.Init(rawMonitorDb)
	} else {
		// db that only exists on memory
		recoverDB = mocks.NewTestDB()
		defer recoverDB.DB().Close()
	}

	watcher := transaction.WatcherFromNetwork(eth, recoverDB, false, constants.TxPollingTime)
	defer watcher.Close()
	if isUsingTestAccounts {
		// this function will block for finality delay blocks
		if err := tests.FundAccounts(eth, watcher, logger); err != nil {
			panic("Unable to fund test accounts")
		}
		tests.SetNextBlockBaseFee(eth.GetEndpoint(), 100_000_000_000)
	}

	// spawn num of Workers
	for _, account := range eth.GetKnownAccounts() {
		go worker(mainCtx, eth, internalClient, watcher, account, uint64(*maxRestBlocksPtr))
	}

	logger.Info("Exiting ...")
}
