package blockchain

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/logging"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Ethereum specific errors
var (
	ErrAccountNotFound  = errors.New("could not find specified account")
	ErrKeysNotFound     = errors.New("account either not found or not unlocked")
	ErrPasscodeNotFound = errors.New("could not find specified passcode")
)

var ETH_MAX_PRIORITY_FEE_PER_GAS_NOT_FOUND string = "Method eth_maxPriorityFeePerGas not found"

type EthereumDetails struct {
	logger                    *logrus.Logger
	endpoint                  string
	keystore                  *keystore.KeyStore
	finalityDelay             uint64
	accounts                  map[common.Address]accounts.Account
	accountIndex              map[common.Address]int
	coinbase                  common.Address
	defaultAccount            accounts.Account
	keys                      map[common.Address]*keystore.Key
	passcodes                 map[common.Address]string
	timeout                   time.Duration
	retryCount                int
	retryDelay                time.Duration
	contracts                 interfaces.Contracts
	client                    interfaces.GethClient
	close                     func() error
	commit                    func()
	chainID                   *big.Int
	syncing                   func(ctx context.Context) (*ethereum.SyncProgress, error)
	peerCount                 func(ctx context.Context) (uint64, error)
	queue                     interfaces.TxnQueue
	selectors                 interfaces.SelectorMap
	txFeePercentageToIncrease int
	txMaxFeeThresholdInGwei   uint64
	txCheckFrequency          time.Duration
	txTimeoutForReplacement   time.Duration
}

//NewEthereumSimulator returns a simulator for testing
func NewEthereumSimulator(
	privateKeys []*ecdsa.PrivateKey,
	retryCount int,
	retryDelay time.Duration,
	timeout time.Duration,
	finalityDelay int,
	wei *big.Int,
	txFeePercentageToIncrease int,
	txMaxFeeThresholdInGwei uint64,
	txCheckFrequency time.Duration,
	txTimeoutForReplacement time.Duration) (*EthereumDetails, error) {
	logger := logging.GetLogger("ethereum")

	if len(privateKeys) < 1 {
		return nil, errors.New("at least 1 private key")
	}

	pathKeystore, err := ioutil.TempDir("", "simulator-keystore-")
	if err != nil {
		return nil, err
	}

	eth := &EthereumDetails{}
	eth.accounts = make(map[common.Address]accounts.Account)
	eth.accountIndex = make(map[common.Address]int)
	eth.contracts = &ContractDetails{eth: eth}
	eth.finalityDelay = uint64(finalityDelay)
	eth.keystore = keystore.NewKeyStore(pathKeystore, keystore.StandardScryptN, keystore.StandardScryptP)
	eth.keys = make(map[common.Address]*keystore.Key)
	eth.logger = logger
	eth.passcodes = make(map[common.Address]string)
	eth.retryCount = retryCount
	eth.retryDelay = retryDelay
	eth.timeout = timeout
	eth.selectors = NewKnownSelectors()
	eth.txFeePercentageToIncrease = txFeePercentageToIncrease
	eth.txMaxFeeThresholdInGwei = txMaxFeeThresholdInGwei
	eth.txCheckFrequency = txCheckFrequency
	eth.txTimeoutForReplacement = txTimeoutForReplacement

	for idx, privateKey := range privateKeys {
		account, err := eth.keystore.ImportECDSA(privateKey, "abc123")
		if err != nil {
			return nil, err
		}

		eth.accounts[account.Address] = account
		eth.accountIndex[account.Address] = idx
		eth.passcodes[account.Address] = "abc123"

		logger.Debugf("Account address:%v", account.Address.String())

		keyID, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}

		eth.keys[account.Address] = &keystore.Key{Address: account.Address, PrivateKey: privateKey, Id: keyID}

		if idx == 0 {
			eth.defaultAccount = account
		}
	}

	genAlloc := make(core.GenesisAlloc)
	for address := range eth.accounts {
		genAlloc[address] = core.GenesisAccount{Balance: wei}
	}

	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		return nil, err
	}
	eth.client = client
	eth.queue = NewTxnQueue(client, eth.selectors, timeout)
	eth.queue.StartLoop()

	eth.chainID = big.NewInt(1337)
	eth.peerCount = func(context.Context) (uint64, error) {
		return 0, nil
	}
	eth.syncing = func(ctx context.Context) (*ethereum.SyncProgress, error) {
		return nil, nil
	}

	eth.SetClose(func() error {
		os.RemoveAll(pathKeystore)
		client.Close()
		return nil
	})

	eth.commit = func() {
		c := http.Client{}
		msg := &JsonrpcMessage{
			Version: "2.0",
			ID:      []byte("1"),
			Method:  "evm_mine",
			Params:  make([]byte, 0),
		}

		if msg.Params, err = json.Marshal(make([]string, 0)); err != nil {
			panic(err)
		}

		var buff bytes.Buffer
		err := json.NewEncoder(&buff).Encode(msg)
		if err != nil {
			log.Fatal(err)
		}

		retryCount := 5
		var worked bool
		for i := 0; i < retryCount; i++ {
			reader := bytes.NewReader(buff.Bytes())
			resp, err := c.Post(
				"http://127.0.0.1:8545",
				"application/json",
				reader,
			)

			if err != nil {
				log.Printf("error calling evm_mine rpc: %v", err)
				<-time.After(5 * time.Second)
				continue
			}

			_, err = io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("error reading response from evm_mine rpc: %v", err)
				<-time.After(5 * time.Second)
				continue
			}

			worked = true
			break
		}

		if !worked {
			panic(fmt.Errorf("error commiting evm_mine on rpc: %v", err))
		}
	}

	return eth, nil
}

func (eth *EthereumDetails) SetContractFactory(ctx context.Context, contractFactoryAddress common.Address) error {
	c := &ContractDetails{eth: eth}
	err := c.LookupContracts(ctx, contractFactoryAddress)
	if err != nil {
		return err
	}

	eth.contracts = c
	return nil
}

// A value of this type can a JSON-RPC request, notification, successful response or
// error response. Which one it is depends on the fields.
type JsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *jsonError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (err *jsonError) Error() string {
	if err.Message == "" {
		return fmt.Sprintf("json-rpc error %d", err.Code)
	}
	return err.Message
}

// NewEthereumEndpoint creates a new Ethereum abstraction
func NewEthereumEndpoint(
	endpoint string,
	pathKeystore string,
	pathPasscodes string,
	defaultAccount string,
	timeout time.Duration,
	retryCount int,
	retryDelay time.Duration,
	finalityDelay int,
	txFeePercentageToIncrease int,
	txMaxFeeThresholdInGwei uint64,
	txCheckFrequency time.Duration,
	txTimeoutForReplacement time.Duration) (*EthereumDetails, error) {

	logger := logging.GetLogger("ethereum")

	eth := &EthereumDetails{
		endpoint:                  endpoint,
		logger:                    logger,
		accounts:                  make(map[common.Address]accounts.Account),
		keys:                      make(map[common.Address]*keystore.Key),
		passcodes:                 make(map[common.Address]string),
		finalityDelay:             uint64(finalityDelay),
		timeout:                   timeout,
		retryCount:                retryCount,
		retryDelay:                retryDelay,
		selectors:                 NewKnownSelectors(),
		txFeePercentageToIncrease: txFeePercentageToIncrease,
		txMaxFeeThresholdInGwei:   txMaxFeeThresholdInGwei,
		txCheckFrequency:          txCheckFrequency,
		txTimeoutForReplacement:   txTimeoutForReplacement}

	eth.contracts = &ContractDetails{eth: eth}

	// Load accounts + passcodes
	eth.loadAccounts(pathKeystore)
	err := eth.loadPasscodes(pathPasscodes)
	if err != nil {
		logger.Errorf("Error in NewEthereumEndpoint at eth.LoadPasscodes: %v", err)
		return nil, err
	}

	// Designate accounts
	var acct accounts.Account
	acct, err = eth.GetAccount(common.HexToAddress(defaultAccount))
	if err != nil {
		logger.Errorf("Can't find user to set as default %v: %v", defaultAccount, err)
		return nil, err
	}
	eth.SetDefaultAccount(acct)

	// Low level rpc client
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	rpcClient, rpcErr := rpc.DialContext(ctx, endpoint)
	if rpcErr != nil {
		logger.Errorf("Error in NewEthereumEndpoint at rpc.DialContext: %v", rpcErr)
		return nil, rpcErr
	}
	ethClient := ethclient.NewClient(rpcClient)
	eth.client = ethClient
	eth.queue = NewTxnQueue(ethClient, eth.selectors, timeout)
	eth.queue.StartLoop()
	eth.chainID, err = ethClient.ChainID(ctx)
	if err != nil {
		logger.Errorf("Error in NewEthereumEndpoint at ethClient.ChainID: %v", err)
		return nil, err
	}

	eth.peerCount = func(ctx context.Context) (uint64, error) {
		return eth.getPeerCount(ctx, rpcClient)
	}
	eth.syncing = ethClient.SyncProgress

	// Find coinbase
	if e := rpcClient.CallContext(ctx, &eth.coinbase, "eth_coinbase"); e != nil {
		logger.Warnf("Failed to determine coinbase: %v", e)
	} else {
		logger.Infof("Coinbase: %v", eth.coinbase.Hex())
	}

	logger.Debug("Completed initialization")
	eth.close = func() error { return nil }
	eth.commit = func() {}

	return eth, nil
}

func (eth *EthereumDetails) GetFinalityDelay() uint64 {
	return eth.finalityDelay
}

func (eth *EthereumDetails) KnownSelectors() interfaces.SelectorMap {
	return eth.selectors
}

func (eth *EthereumDetails) Close() error {
	eth.queue.Close()
	return eth.close()
}

func (eth *EthereumDetails) SetClose(fn func() error) {
	if eth.close != nil {
		var prevClose (func() error) = eth.close
		eth.close = func() error {
			err := prevClose()
			if err != nil {
				return err
			}

			return fn()
		}
	} else {
		eth.close = fn
	}
}

func (eth *EthereumDetails) Commit() {
	eth.commit()
}

func (eth *EthereumDetails) Contracts() interfaces.Contracts {
	return eth.contracts
}

func (eth *EthereumDetails) GetPeerCount(ctx context.Context) (uint64, error) {
	return eth.peerCount(ctx)
}

func (eth *EthereumDetails) Queue() interfaces.TxnQueue {
	return eth.queue
}

func (eth *EthereumDetails) getPeerCount(ctx context.Context, rpcClient *rpc.Client) (uint64, error) {
	// Let's see how many peers our endpoint has
	var peerCountString string
	if err := rpcClient.CallContext(ctx, &peerCountString, "net_peerCount"); err != nil {
		eth.logger.Warnf("could not get peerCount: %v", err)
		return 0, err
	}

	var peerCount uint64
	_, err := fmt.Sscanf(peerCountString, "0x%x", &peerCount)
	if err != nil {
		eth.logger.Warnf("could not parse peerCount: %v", err)
		return 0, err
	}
	return peerCount, nil
}

//IsEthereumAccessible checks against endpoint to confirm server responds
func (eth *EthereumDetails) IsEthereumAccessible() bool {
	ctx, cancel := eth.GetTimeoutContext()
	defer cancel()
	block, err := eth.client.BlockByNumber(ctx, nil)
	if err == nil && block != nil {
		return true
	}

	eth.logger.Debug("IsEthereumAccessible()...false")
	return false
}

//ChainID returns the ID used to build ethereum client
func (eth *EthereumDetails) ChainID() *big.Int {
	return eth.chainID
}

//LoadAccounts Scans the directory specified and loads all the accounts found
func (eth *EthereumDetails) loadAccounts(directoryPath string) {
	logger := eth.logger

	logger.Infof("LoadAccounts(\"%v\")...", directoryPath)
	ks := keystore.NewKeyStore(directoryPath, keystore.StandardScryptN, keystore.StandardScryptP)
	accts := make(map[common.Address]accounts.Account, 10)
	acctIndex := make(map[common.Address]int, 10)

	var index int
	for _, wallet := range ks.Wallets() {
		for _, account := range wallet.Accounts() {
			logger.Infof("... found account %v", account.Address.Hex())
			accts[account.Address] = account
			acctIndex[account.Address] = index
			index++
		}
	}

	eth.accounts = accts
	eth.accountIndex = acctIndex
	eth.keystore = ks
}

// LoadPasscodes loads the specified passcode file
func (eth *EthereumDetails) loadPasscodes(filePath string) error {
	logger := eth.logger

	logger.Infof("loadPasscodes(\"%v\")...", filePath)
	passcodes := make(map[common.Address]string)

	file, err := os.Open(filePath)
	if err != nil {
		logger.Errorf("Failed to open passcode file \"%v\": %s", filePath, err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") {
			components := strings.Split(line, "=")
			if len(components) == 2 {
				address := strings.TrimSpace(components[0])
				passcode := strings.TrimSpace(components[1])

				passcodes[common.HexToAddress(address)] = passcode
			}
		}
	}

	eth.passcodes = passcodes

	return nil
}

func (eth *EthereumDetails) UnlockAccountWithPasscode(acct accounts.Account, passcode string) error {
	return eth.keystore.Unlock(acct, passcode)
}

// UnlockAccount unlocks the previously loaded account using the previously loaded passcodes
func (eth *EthereumDetails) UnlockAccount(acct accounts.Account) error {

	eth.logger.Infof("Unlocking account address:%v", acct.Address.String())

	passcode, passcodeFound := eth.passcodes[acct.Address]
	if !passcodeFound {
		return ErrPasscodeNotFound
	}

	err := eth.keystore.Unlock(acct, passcode)
	if err != nil {
		return err
	}

	// Open the account key file
	keyJSON, err := ioutil.ReadFile(acct.URL.Path)
	if err != nil {
		return err
	}

	// Get the private key
	key, err := keystore.DecryptKey(keyJSON, passcode)
	if err != nil {
		return err
	}

	eth.keys[acct.Address] = key

	return nil
}

// GetGethClient returns an amalgamated geth client interface
func (eth *EthereumDetails) GetGethClient() interfaces.GethClient {
	return eth.client
}

// GetAccount returns the account specified
func (eth *EthereumDetails) GetAccount(addr common.Address) (accounts.Account, error) {
	acct, accountFound := eth.accounts[addr]
	if !accountFound {
		return acct, ErrAccountNotFound
	}

	return acct, nil
}

func (eth *EthereumDetails) GetAccountKeys(addr common.Address) (*keystore.Key, error) {
	if key, ok := eth.keys[addr]; ok {
		return key, nil
	}
	return nil, ErrKeysNotFound
}

// SetDefaultAccount designates the account to be used by default
func (eth *EthereumDetails) SetDefaultAccount(acct accounts.Account) {
	eth.defaultAccount = acct
}

// GetDefaultAccount returns the default account
func (eth *EthereumDetails) GetDefaultAccount() accounts.Account {
	return eth.defaultAccount
}

// GetCoinbaseAddress returns the account to use for contract deploys
func (eth *EthereumDetails) GetCoinbaseAddress() common.Address {
	return eth.coinbase
}

// GetBalance returns the ETHER balance of account specified
func (eth *EthereumDetails) GetBalance(addr common.Address) (*big.Int, error) {
	ctx, cancel := eth.GetTimeoutContext()
	defer cancel()
	balance, err := eth.client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (eth *EthereumDetails) GetEndpoint() string {
	return eth.endpoint
}

func (eth *EthereumDetails) GetKnownAccounts() []accounts.Account {
	accounts := make([]accounts.Account, len(eth.accounts))

	for address, accountIndex := range eth.accountIndex {
		accounts[accountIndex] = eth.accounts[address]
	}

	return accounts
}

func (eth *EthereumDetails) GetTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), eth.timeout)
}

// GetSyncProgress returns a flag if we are syncing, a pointer to a struct if we are, or an error
func (eth *EthereumDetails) GetSyncProgress() (bool, *ethereum.SyncProgress, error) {

	ctx, ctxCancel := eth.GetTimeoutContext()
	progress, err := eth.syncing(ctx)
	defer ctxCancel()

	if err == nil && progress == nil {
		return false, nil, nil
	}

	if err == nil && progress != nil {
		return true, progress, nil
	}

	return false, nil, err
}

func (eth *EthereumDetails) GetEvents(ctx context.Context, firstBlock uint64, lastBlock uint64, addresses []common.Address) ([]types.Log, error) {

	logger := eth.logger

	logger.Debugf("...GetEvents(firstBlock:%v,lastBlock:%v,addresses:%x)", firstBlock, lastBlock, addresses)

	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(firstBlock),
		ToBlock:   new(big.Int).SetUint64(lastBlock),
		Addresses: addresses}

	logs, err := eth.client.FilterLogs(ctx, query)
	if err != nil {
		logger.Errorf("Could not filter logs: %v", err)
		return nil, err
	}

	for idx, log := range logs {
		logger.Debugf("Log[%v] Block[%v]:%v", idx, log.BlockNumber, log)
		for idx, hash := range log.Topics {
			logger.Debugf("Hash[%v]:%x", idx, hash)
		}
	}

	return logs, nil
}

func (eth *EthereumDetails) RetryCount() int {
	return eth.retryCount
}

func (eth *EthereumDetails) RetryDelay() time.Duration {
	return eth.retryDelay
}

func (eth *EthereumDetails) Timeout() time.Duration {
	return eth.timeout
}

func (eth *EthereumDetails) GetTxFeePercentageToIncrease() int {
	return eth.txFeePercentageToIncrease
}

func (eth *EthereumDetails) GetTxMaxFeeThresholdInGwei() uint64 {
	return eth.txMaxFeeThresholdInGwei
}

func (eth *EthereumDetails) GetTxCheckFrequency() time.Duration {
	return eth.txCheckFrequency
}

func (eth *EthereumDetails) GetTxTimeoutForReplacement() time.Duration {
	return eth.txTimeoutForReplacement
}

func (eth *EthereumDetails) GetTransactionOpts(ctx context.Context, account accounts.Account) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyStoreTransactorWithChainID(eth.keystore, account, eth.chainID)
	if err != nil {
		eth.logger.Errorf("could not create transactor for %v: %v", account.Address.Hex(), err)
		return nil, err
	}
	subCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	block, err := eth.client.BlockByNumber(subCtx, nil)
	if err != nil && block == nil {
		return nil, fmt.Errorf("could not get block number: %w", err)
	}

	eth.logger.Infof("Previous BaseFee:%v GasUsed:%v GasLimit:%v",
		block.BaseFee().String(),
		block.GasUsed(),
		block.GasLimit())

	baseFee := block.BaseFee()

	bmi64 := int64(2)
	bm := new(big.Int).SetInt64(bmi64)
	bf := new(big.Int).Set(baseFee)
	baseFee2x := new(big.Int).Mul(bm, bf)

	tipCap, err := eth.client.SuggestGasTipCap(subCtx)
	if err != nil {
		if err.Error() == ETH_MAX_PRIORITY_FEE_PER_GAS_NOT_FOUND {
			tipCap = big.NewInt(1)
		} else {
			return nil, fmt.Errorf("could not get suggested gas tip cap: %w", err)
		}
	}
	feeCap := new(big.Int).Add(baseFee2x, new(big.Int).Set(tipCap))

	opts.Context = ctx
	opts.GasFeeCap = new(big.Int).Set(feeCap)
	opts.GasTipCap = new(big.Int).Set(tipCap)
	return opts, nil
}

func (eth *EthereumDetails) GetCallOpts(ctx context.Context, account accounts.Account) *bind.CallOpts {

	return &bind.CallOpts{
		BlockNumber: nil,
		Context:     ctx,
		Pending:     false,
		From:        account.Address}
}

// TransferEther transfer's ether from one account to another, assumes from is unlocked
func (eth *EthereumDetails) TransferEther(from common.Address, to common.Address, wei *big.Int) (*types.Transaction, error) {

	ctx, cancel := eth.GetTimeoutContext()
	defer cancel()

	nonce, err := eth.client.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, err
	}

	gasPrice, err := eth.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	block, err := eth.client.BlockByNumber(ctx, nil)
	if err != nil && block == nil {
		return nil, fmt.Errorf("could not get block number: %w", err)
	}

	eth.logger.Infof("Previous BaseFee:%v GasUsed:%v GasLimit:%v",
		block.BaseFee().String(),
		block.GasUsed(),
		block.GasLimit())

	gasLimit := uint64(21000)
	eth.logger.Infof("gasLimit:%v SuggestGasPrice(): %v", gasLimit, gasPrice.String())

	baseFee := block.BaseFee()

	bmi64 := int64(100)
	bm := new(big.Int).SetInt64(bmi64)
	bf := new(big.Int).Set(baseFee)
	baseFee2x := new(big.Int).Mul(bm, bf)
	tipCap, err := eth.client.SuggestGasTipCap(ctx)
	if err != nil {
		if err.Error() == ETH_MAX_PRIORITY_FEE_PER_GAS_NOT_FOUND {
			tipCap = big.NewInt(1)
		} else {
			return nil, fmt.Errorf("could not get suggested gas tip cap: %w", err)
		}
	}
	feeCap := new(big.Int).Add(baseFee2x, new(big.Int).Set(tipCap))

	txRough := &types.DynamicFeeTx{}
	txRough.ChainID = eth.chainID
	txRough.To = &to
	txRough.GasFeeCap = new(big.Int).Set(feeCap)
	txRough.GasTipCap = new(big.Int).Set(tipCap)
	txRough.Gas = gasLimit
	txRough.Nonce = nonce
	txRough.Value = wei

	eth.logger.Debugf("TransferEther => chainID:%v from:%v nonce:%v, to:%v, wei:%v, gasLimit:%v, gasPrice:%v",
		eth.chainID, from.Hex(), nonce, to.Hex(), wei, gasLimit, gasPrice)

	signer := types.NewLondonSigner(eth.chainID)

	signedTx, err := types.SignNewTx(eth.keys[from].PrivateKey, signer, txRough)
	if err != nil {
		eth.logger.Errorf("signing error:%v", err)
		return nil, err
	}
	err = eth.client.SendTransaction(ctx, signedTx)
	if err != nil {
		eth.logger.Errorf("sending error:%v", err)
		return nil, err
	}

	return signedTx, nil
}

// GetCurrentHeight gets the height of the endpoints chain
func (eth *EthereumDetails) GetCurrentHeight(ctx context.Context) (uint64, error) {
	header, err := eth.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}

	return header.Number.Uint64(), nil
}

// GetFinalizedHeight gets the height of the endpoints chain at which is is considered finalized
func (eth *EthereumDetails) GetFinalizedHeight(ctx context.Context) (uint64, error) {
	height, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return height, err
	}

	if eth.finalityDelay >= height {
		return 0, nil
	}
	return height - eth.finalityDelay, nil

}

func (eth *EthereumDetails) GetSnapshot() ([]byte, error) {
	return nil, nil
}

func (eth *EthereumDetails) GetValidators(ctx context.Context) ([]common.Address, error) {
	c := eth.contracts
	validatorAddresses, err := c.ValidatorPool().GetValidatorsAddresses(eth.GetCallOpts(ctx, eth.defaultAccount))
	if err != nil {
		eth.logger.Warnf("Could not call contract:%v", err)
		return nil, err
	}

	return validatorAddresses, nil
}

func (eth *EthereumDetails) Clone(defaultAccount accounts.Account) *EthereumDetails {
	nEth := *eth

	nEth.defaultAccount = defaultAccount

	return &nEth
}

// StringToBytes32 is useful for convert a Go string into a bytes32 useful calling Solidity
func StringToBytes32(str string) (b [32]byte) {
	copy(b[:], []byte(str)[0:32])
	return
}

func logAndEat(logger *logrus.Logger, err error) {
	if err != nil {
		logger.Error(err)
	}
}

/*
type Updater struct {
	err     error
	Logger  *logrus.Logger
	Updater *bindings.DiamondUpdateFacet
	TxnOpts *bind.TransactOpts
}

//
func (u *Updater) Add(signature string, facet common.Address) *types.Transaction {
	if u.err != nil {
		return nil
	}

	selector := CalculateSelector(signature)
	if u.Logger != nil {
		u.Logger.Infof("Registering %v as %x with %v", signature, selector, facet.Hex())
	}

	txn, err := u.Updater.AddFacet(u.TxnOpts, selector, facet)
	u.err = err
	return txn
}
*/
