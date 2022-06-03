package ethereum

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

	"github.com/MadBase/MadNet/bridge/bindings"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
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

var _ Network = &Details{}

// Network contains state information about a connection to the Ethereum node
type Network interface {

	// Extensions for use with simulator
	ChainID() *big.Int
	Close() error
	Commit()

	IsAccessible() bool

	GetCallOpts(context.Context, accounts.Account) (*bind.CallOpts, error)
	GetCallOptsLatestBlock(ctx context.Context, account accounts.Account) *bind.CallOpts
	GetTransactionOpts(context.Context, accounts.Account) (*bind.TransactOpts, error)

	UnlockAccount(accounts.Account) error
	UnlockAccountWithPasscode(accounts.Account, string) error

	TransferEther(common.Address, common.Address, *big.Int) (*types.Transaction, error)

	GetAccount(common.Address) (accounts.Account, error)
	GetAccountKeys(addr common.Address) (*keystore.Key, error)
	GetBalance(common.Address) (*big.Int, error)
	GetClient() Client
	GetCoinbaseAddress() common.Address
	GetCurrentHeight(context.Context) (uint64, error)
	GetDefaultAccount() accounts.Account
	GetEndpoint() string
	GetEvents(ctx context.Context, firstBlock uint64, lastBlock uint64, addresses []common.Address) ([]types.Log, error)
	GetFinalizedHeight(context.Context) (uint64, error)
	GetKnownAccounts() []accounts.Account
	GetPeerCount(context.Context) (uint64, error)
	GetSnapshot() ([]byte, error)
	GetSyncProgress() (bool, *ethereum.SyncProgress, error)
	GetTimeoutContext() (context.Context, context.CancelFunc)
	GetValidators(context.Context) ([]common.Address, error)
	SetFinalityDelay(uint64)
	GetFinalityDelay() uint64
	RetryCount() int
	RetryDelay() time.Duration

	Timeout() time.Duration

	GetTxFeePercentageToIncrease() int
	GetTxMaxGasFeeAllowedInGwei() uint64

	Contracts() Contracts
}

type Client interface {

	// geth.ChainReader
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	BlockNumber(ctx context.Context) (uint64, error)
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error)
	TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error)
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)

	// geth.TransactionReader
	TransactionByHash(ctx context.Context, txHash common.Hash) (tx *types.Transaction, isPending bool, err error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)

	// geth.ChainStateReader
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)

	CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)

	// -- bind.ContractTransactor
	PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error

	// -- bind.ContractFilterer
	FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error)
	SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
}

// Contracts contains bindings to smart contract system
type Contracts interface {
	LookupContracts(ctx context.Context, registryAddress common.Address) error

	Ethdkg() bindings.IETHDKG
	EthdkgAddress() common.Address
	AToken() bindings.IAToken
	ATokenAddress() common.Address
	BToken() bindings.IBToken
	BTokenAddress() common.Address
	PublicStaking() bindings.IPublicStaking
	PublicStakingAddress() common.Address
	ValidatorStaking() bindings.IValidatorStaking
	ValidatorStakingAddress() common.Address
	ContractFactory() bindings.IAliceNetFactory
	ContractFactoryAddress() common.Address
	SnapshotsAddress() common.Address
	Snapshots() bindings.ISnapshots
	ValidatorPool() bindings.IValidatorPool
	ValidatorPoolAddress() common.Address
	Governance() bindings.IGovernance
	GovernanceAddress() common.Address
}

type Details struct {
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
	contracts                 Contracts
	client                    Client
	close                     func() error
	commit                    func()
	chainID                   *big.Int
	syncing                   func(ctx context.Context) (*ethereum.SyncProgress, error)
	peerCount                 func(ctx context.Context) (uint64, error)
	txFeePercentageToIncrease int
	txMaxGasFeeAllowedInGwei  uint64
}

//NewSimulator returns a simulator for testing
func NewSimulator(
	privateKeys []*ecdsa.PrivateKey,
	retryCount int,
	retryDelay time.Duration,
	timeout time.Duration,
	finalityDelay int,
	wei *big.Int,
	txFeePercentageToIncrease int,
	txMaxGasFeeAllowedInGwei uint64,
) (*Details, error) {
	logger := logging.GetLogger("ethereum")

	if len(privateKeys) < 1 {
		return nil, errors.New("at least 1 private key")
	}

	pathKeystore, err := ioutil.TempDir("", "simulator-keystore-")
	if err != nil {
		return nil, err
	}

	eth := &Details{}
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
	eth.txFeePercentageToIncrease = txFeePercentageToIncrease
	eth.txMaxGasFeeAllowedInGwei = txMaxGasFeeAllowedInGwei
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
		msg := &JsonRPCMessage{
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
			panic(fmt.Errorf("error committing evm_mine on rpc: %v", err))
		}
	}

	return eth, nil
}

func (eth *Details) SetContractFactory(ctx context.Context, contractFactoryAddress common.Address) error {
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
type JsonRPCMessage struct {
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

// NewEndpoint creates a new Ethereum abstraction
func NewEndpoint(
	endpoint string,
	pathKeystore string,
	pathPasscodes string,
	defaultAccount string,
	timeout time.Duration,
	retryCount int,
	retryDelay time.Duration,
	txFeePercentageToIncrease int,
	txMaxGasFeeAllowedInGwei uint64) (*Details, error) {

	logger := logging.GetLogger("ethereum")

	eth := &Details{
		endpoint:                  endpoint,
		logger:                    logger,
		accounts:                  make(map[common.Address]accounts.Account),
		keys:                      make(map[common.Address]*keystore.Key),
		passcodes:                 make(map[common.Address]string),
		finalityDelay:             constants.DefaultFinalityDelay,
		timeout:                   timeout,
		retryCount:                retryCount,
		retryDelay:                retryDelay,
		txFeePercentageToIncrease: txFeePercentageToIncrease,
		txMaxGasFeeAllowedInGwei:  txMaxGasFeeAllowedInGwei}

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
	// instantiate but don't initiate the new transaction with default finality Delay.
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

func (eth *Details) SetFinalityDelay(newFinalityDelay uint64) {
	eth.finalityDelay = newFinalityDelay
}

func (eth *Details) GetFinalityDelay() uint64 {
	return eth.finalityDelay
}
func (eth *Details) Close() error {
	return eth.close()
}

func (eth *Details) SetClose(fn func() error) {
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

func (eth *Details) Commit() {
	eth.commit()
}

func (eth *Details) Contracts() Contracts {
	return eth.contracts
}

func (eth *Details) GetPeerCount(ctx context.Context) (uint64, error) {
	return eth.peerCount(ctx)
}

func (eth *Details) getPeerCount(ctx context.Context, rpcClient *rpc.Client) (uint64, error) {
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

//IsAccessible checks against endpoint to confirm server responds
func (eth *Details) IsAccessible() bool {
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
func (eth *Details) ChainID() *big.Int {
	return eth.chainID
}

//LoadAccounts Scans the directory specified and loads all the accounts found
func (eth *Details) loadAccounts(directoryPath string) {
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
func (eth *Details) loadPasscodes(filePath string) error {
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

func (eth *Details) UnlockAccountWithPasscode(acct accounts.Account, passcode string) error {
	return eth.keystore.Unlock(acct, passcode)
}

// UnlockAccount unlocks the previously loaded account using the previously loaded passcodes
func (eth *Details) UnlockAccount(acct accounts.Account) error {

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

// GetClient returns an amalgamated geth client interface
func (eth *Details) GetClient() Client {
	return eth.client
}

// GetAccount returns the account specified
func (eth *Details) GetAccount(addr common.Address) (accounts.Account, error) {
	acct, accountFound := eth.accounts[addr]
	if !accountFound {
		return acct, ErrAccountNotFound
	}

	return acct, nil
}

func (eth *Details) GetAccountKeys(addr common.Address) (*keystore.Key, error) {
	if key, ok := eth.keys[addr]; ok {
		return key, nil
	}
	return nil, ErrKeysNotFound
}

// SetDefaultAccount designates the account to be used by default
func (eth *Details) SetDefaultAccount(acct accounts.Account) {
	eth.defaultAccount = acct
}

// GetDefaultAccount returns the default account
func (eth *Details) GetDefaultAccount() accounts.Account {
	return eth.defaultAccount
}

// GetCoinbaseAddress returns the account to use for contract deploys
func (eth *Details) GetCoinbaseAddress() common.Address {
	return eth.coinbase
}

// GetBalance returns the ETHER balance of account specified
func (eth *Details) GetBalance(addr common.Address) (*big.Int, error) {
	ctx, cancel := eth.GetTimeoutContext()
	defer cancel()
	balance, err := eth.client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (eth *Details) GetEndpoint() string {
	return eth.endpoint
}

func (eth *Details) GetKnownAccounts() []accounts.Account {
	accounts := make([]accounts.Account, len(eth.accounts))

	for address, accountIndex := range eth.accountIndex {
		accounts[accountIndex] = eth.accounts[address]
	}

	return accounts
}

func (eth *Details) GetTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), eth.timeout)
}

// GetSyncProgress returns a flag if we are syncing, a pointer to a struct if we are, or an error
func (eth *Details) GetSyncProgress() (bool, *ethereum.SyncProgress, error) {

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

func (eth *Details) GetEvents(ctx context.Context, firstBlock uint64, lastBlock uint64, addresses []common.Address) ([]types.Log, error) {

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

func (eth *Details) RetryCount() int {
	return eth.retryCount
}

func (eth *Details) RetryDelay() time.Duration {
	return eth.retryDelay
}

func (eth *Details) Timeout() time.Duration {
	return eth.timeout
}

func (eth *Details) GetTxFeePercentageToIncrease() int {
	return eth.txFeePercentageToIncrease
}

func (eth *Details) GetTxMaxGasFeeAllowedInGwei() uint64 {
	return eth.txMaxGasFeeAllowedInGwei
}

func (eth *Details) GetTransactionOpts(ctx context.Context, account accounts.Account) (*bind.TransactOpts, error) {
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

	// This should give us 8 full blocks before we are priced out
	bmi64 := int64(2)
	bm := new(big.Int).SetInt64(bmi64)
	bf := new(big.Int).Set(baseFee)
	baseFee2x := new(big.Int).Mul(bm, bf)

	tipCap, err := eth.client.SuggestGasTipCap(subCtx)
	if err != nil {
		if err.Error() == ETH_MAX_PRIORITY_FEE_PER_GAS_NOT_FOUND {
			tipCap = big.NewInt(1_000_000_000)
		} else {
			return nil, fmt.Errorf("could not get suggested gas tip cap: %w", err)
		}
	}
	feeCap := new(big.Int).Add(baseFee2x, new(big.Int).Set(tipCap))

	txMaxGasFeeAllowedInGwei := new(big.Int).SetUint64(eth.GetTxMaxGasFeeAllowedInGwei())
	// make sure that the max fee that we are going to pay on this tx doesn't pass the limit that we set on config
	txMaxFeeThresholdInWei := new(big.Int).Mul(txMaxGasFeeAllowedInGwei, new(big.Int).SetUint64(1_000_000_000))
	if feeCap.Cmp(txMaxFeeThresholdInWei) > 0 {
		return nil, fmt.Errorf("max tx fee computed: %v is greater than limit set on config: %v", feeCap.String(), txMaxFeeThresholdInWei.String())
	}

	eth.logger.WithFields(logrus.Fields{
		"MinerTip":          tipCap,
		"MaximumGasAllowed": txMaxFeeThresholdInWei.String(),
	}).Infof("Creating TX with MaximumGasPrice: %v WEI", feeCap)
	opts.Context = ctx
	opts.GasFeeCap = new(big.Int).Set(feeCap)
	opts.GasTipCap = new(big.Int).Set(tipCap)
	return opts, nil
}

func (eth *Details) GetCallOpts(ctx context.Context, account accounts.Account) (*bind.CallOpts, error) {

	finalizedHeightU64, err := eth.GetFinalizedHeight(ctx)
	if err != nil {
		return nil, err
	}
	finalizedHeight := new(big.Int).SetUint64(finalizedHeightU64)
	return &bind.CallOpts{
		BlockNumber: finalizedHeight,
		Context:     ctx,
		Pending:     false,
		From:        account.Address}, nil
}

// Function to call the smart contract state at the latest block seen by the
// ethereum node. USE THIS FUNCTION CAREFULLY, THE LATEST BLOCK IS SUSCEPTIBLE
// TO CHAIN RE-ORGS AND SHOULD NEVER BE USED AS TEST IF SOMETHING WAS COMMITTED
// TO OUR CONTRACTS. IDEALLY, ONLY USE THIS FUNCTION FOR MONITORING FUNCTIONS
// THAT DEPENDS ON THE LATEST BLOCK. Otherwise use GetCallOpts!!!
func (eth *Details) GetCallOptsLatestBlock(ctx context.Context, account accounts.Account) *bind.CallOpts {
	return &bind.CallOpts{
		BlockNumber: nil,
		Context:     ctx,
		Pending:     false,
		From:        account.Address}
}

// TransferEther transfer's ether from one account to another, assumes from is unlocked
func (eth *Details) TransferEther(from common.Address, to common.Address, wei *big.Int) (*types.Transaction, error) {

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
func (eth *Details) GetCurrentHeight(ctx context.Context) (uint64, error) {
	return eth.client.BlockNumber(ctx)
}

// GetFinalizedHeight gets the height of the endpoints chain at which is is considered finalized
func (eth *Details) GetFinalizedHeight(ctx context.Context) (uint64, error) {
	height, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return height, err
	}

	if eth.finalityDelay >= height {
		return 0, nil
	}
	return height - eth.finalityDelay, nil

}

func (eth *Details) GetSnapshot() ([]byte, error) {
	return nil, nil
}

func (eth *Details) GetValidators(ctx context.Context) ([]common.Address, error) {
	c := eth.contracts
	callOpts, err := eth.GetCallOpts(ctx, eth.defaultAccount)
	if err != nil {
		return nil, err
	}
	validatorAddresses, err := c.ValidatorPool().GetValidatorsAddresses(callOpts)
	if err != nil {
		eth.logger.Warnf("Could not call contract:%v", err)
		return nil, err
	}

	return validatorAddresses, nil
}

func (eth *Details) Clone(defaultAccount accounts.Account) *Details {
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
