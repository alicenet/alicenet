package ethereum

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
)

// Ethereum specific errors
var (
	ErrAccountNotFound           = errors.New("could not find specified account")
	ErrKeysNotFound              = errors.New("account either not found or not unlocked")
	ErrPassCodeNotFound          = errors.New("could not find specified passCode")
	ErrInvalidTxMaxGasFeeAllowed = fmt.Errorf("txMaxGasFeeAllowedInGwei should be greater than %v Gwei", constants.EthereumMinGasFeeAllowedInGwei)
)

var ETH_MAX_PRIORITY_FEE_PER_GAS_NOT_FOUND string = "Method eth_maxPriorityFeePerGas not found"

var _ Network = &Details{}

// Network contains state information about a connection to the Ethereum node
type Network interface {
	Close()
	IsAccessible() bool
	EndpointInSync(ctx context.Context) (bool, uint32, error)
	GetPeerCount(ctx context.Context) (uint64, error)
	GetChainID() *big.Int
	GetTxNotFoundMaxBlocks() uint64
	GetTxMaxStaleBlocks() uint64
	GetTransactionByHash(ctx context.Context, txHash common.Hash) (tx *types.Transaction, isPending bool, err error)
	GetTransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	GetHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	GetBlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	GetBlockBaseFeeAndSuggestedGasTip(ctx context.Context) (*big.Int, *big.Int, error)
	GetCallOpts(context.Context, accounts.Account) (*bind.CallOpts, error)
	GetCallOptsLatestBlock(ctx context.Context, account accounts.Account) *bind.CallOpts
	GetTransactionOpts(context.Context, accounts.Account) (*bind.TransactOpts, error)
	GetAccount(common.Address) (accounts.Account, error)
	GetAccountKeys(addr common.Address) (*keystore.Key, error)
	GetBalance(common.Address) (*big.Int, error)
	GetCurrentHeight(context.Context) (uint64, error)
	GetFinalizedHeight(context.Context) (uint64, error)
	GetEndpoint() string
	GetDefaultAccount() accounts.Account
	GetKnownAccounts() []accounts.Account
	GetTimeoutContext() (context.Context, context.CancelFunc)
	GetEvents(ctx context.Context, firstBlock uint64, lastBlock uint64, addresses []common.Address) ([]types.Log, error)
	GetFinalityDelay() uint64
	GetTxMaxGasFeeAllowed() *big.Int
	Contracts() Contracts
	GetPendingNonce(ctx context.Context, account common.Address) (uint64, error)
	SignTransaction(tx types.TxData, signerAddress common.Address) (*types.Transaction, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	ExtractTransactionSender(tx *types.Transaction) (common.Address, error)
	RetryTransaction(ctx context.Context, tx *types.Transaction, baseFee *big.Int, gasTipCap *big.Int) (*types.Transaction, error)
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

type Details struct {
	logger               *logrus.Logger
	endpoint             string
	keystore             *keystore.KeyStore
	finalityDelay        uint64
	accounts             map[common.Address]accounts.Account
	accountIndex         map[common.Address]int
	defaultAccount       accounts.Account
	keys                 map[common.Address]*keystore.Key
	passCodes            map[common.Address]string
	contracts            Contracts
	rpcClient            *rpc.Client
	client               *ethclient.Client
	chainID              *big.Int
	txMaxGasFeeAllowed   *big.Int
	endpointMinimumPeers uint32
}

// NewEndpoint creates a new Ethereum abstraction
func NewEndpoint(
	endpoint string,
	pathKeystore string,
	pathPassCodes string,
	defaultAccount string,
	finalityDelay uint64,
	txMaxGasFeeAllowedInGwei uint64,
	endpointMinimumPeers uint32) (*Details, error) {

	logger := logging.GetLogger("ethereum")

	if txMaxGasFeeAllowedInGwei < constants.EthereumMinGasFeeAllowedInGwei {
		return nil, ErrInvalidTxMaxGasFeeAllowed
	}

	txMaxGasFeeAllowedInWei := new(big.Int).Mul(new(big.Int).SetUint64(txMaxGasFeeAllowedInGwei), new(big.Int).SetUint64(1_000_000_000))

	eth := &Details{
		endpoint:           endpoint,
		logger:             logger,
		accounts:           make(map[common.Address]accounts.Account),
		keys:               make(map[common.Address]*keystore.Key),
		passCodes:          make(map[common.Address]string),
		finalityDelay:      finalityDelay,
		txMaxGasFeeAllowed: txMaxGasFeeAllowedInWei,
	}

	eth.contracts = NewContractDetails(eth)

	// Load accounts + passCodes
	eth.loadAccounts(pathKeystore)
	err := eth.loadPassCodes(pathPassCodes)
	if err != nil {
		return nil, fmt.Errorf("Error in NewEthereumEndpoint at eth.LoadPassCodes: %v", err)
	}

	// Designate accounts
	var acct accounts.Account
	acct, err = eth.GetAccount(common.HexToAddress(defaultAccount))
	if err != nil {
		return nil, fmt.Errorf("Can't find user to set as default %v: %v", defaultAccount, err)
	}
	eth.setDefaultAccount(acct)
	if err := eth.unlockAccount(acct); err != nil {
		return nil, fmt.Errorf("Could not unlock account: %v", err)
	}

	// Low level rpc client
	ctx, cancel := context.WithTimeout(context.Background(), constants.MonitorTimeout)
	defer cancel()
	rpcClient, rpcErr := rpc.DialContext(ctx, endpoint)
	if rpcErr != nil {
		return nil, fmt.Errorf("Error in NewEndpoint at rpc.DialContext: %v", rpcErr)
	}
	eth.rpcClient = rpcClient
	ethClient := ethclient.NewClient(rpcClient)
	eth.client = ethClient
	// instantiate but don't initiate the new transaction with default finality Delay.
	eth.chainID, err = ethClient.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error in NewEndpoint at ethClient.ChainID: %v", err)
	}

	logger.Debug("Completed initialization")

	return eth, nil
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
func (eth *Details) loadPassCodes(filePath string) error {
	logger := eth.logger

	logger.Infof("loadPassCodes(\"%v\")...", filePath)
	passcodes := make(map[common.Address]string)

	file, err := os.Open(filePath)
	if err != nil {
		logger.Errorf("Failed to open passCode file \"%v\": %s", filePath, err)
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

	eth.passCodes = passcodes

	return nil
}

// UnlockAccount unlocks the previously loaded account using the previously loaded passCodes
func (eth *Details) unlockAccount(acct accounts.Account) error {

	eth.logger.Infof("Unlocking account address:%v", acct.Address.String())

	passCode, passCodeFound := eth.passCodes[acct.Address]
	if !passCodeFound {
		return ErrPassCodeNotFound
	}

	err := eth.keystore.Unlock(acct, passCode)
	if err != nil {
		return err
	}

	// Open the account key file
	keyJSON, err := ioutil.ReadFile(acct.URL.Path)
	if err != nil {
		return err
	}

	// Get the private key
	key, err := keystore.DecryptKey(keyJSON, passCode)
	if err != nil {
		return err
	}

	eth.keys[acct.Address] = key

	return nil
}

// setDefaultAccount designates the account to be used by default
func (eth *Details) setDefaultAccount(acct accounts.Account) {
	eth.defaultAccount = acct
}

// Bump the tip cap for retries
func (eth *Details) bumpTipCap(gasTipCap *big.Int) *big.Int {
	// calculate percentage% increase in GasTipCap
	gasTipCapPercent := new(big.Int).Mul(gasTipCap, big.NewInt(int64(constants.EthereumTipCapPercentageBump)))
	gasTipCapPercent = new(big.Int).Div(gasTipCapPercent, big.NewInt(100))
	resultTipCap := new(big.Int).Add(gasTipCap, gasTipCapPercent)
	return resultTipCap
}

// GetSyncProgress returns a flag if we are syncing, a pointer to a struct if we are, or an error
func (eth *Details) getSyncProgress() (bool, *ethereum.SyncProgress, error) {

	ctx, ctxCancel := eth.GetTimeoutContext()
	defer ctxCancel()
	progress, err := eth.client.SyncProgress(ctx)
	if err != nil {
		return false, nil, err
	}

	if progress == nil {
		return false, nil, nil
	}

	return true, progress, nil
}

//ChainID returns the ID used to build ethereum client
func (eth *Details) GetChainID() *big.Int {
	return eth.chainID
}

// Get finality delay
func (eth *Details) GetFinalityDelay() uint64 {
	return eth.finalityDelay
}

// close the ethereum client
func (eth *Details) Close() {
	eth.client.Close()
}

// Return the contract interface for the smart contract details
func (eth *Details) Contracts() Contracts {
	return eth.contracts
}

// wrapper around ethclient.TransactionByHash
func (eth *Details) GetTransactionByHash(ctx context.Context, txHash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	return eth.client.TransactionByHash(ctx, txHash)
}

// wrapper around ethclient.TransactionReceipt
func (eth *Details) GetTransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return eth.client.TransactionReceipt(ctx, txHash)
}

// wrapper around ethclient.HeaderByNumber
func (eth *Details) GetHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return eth.client.HeaderByNumber(ctx, number)
}

// wrapper around ethclient.BlockByNumber
func (eth *Details) GetBlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return eth.client.BlockByNumber(ctx, number)
}

// wrapper around ethclient.PendingNonceAt
func (eth *Details) GetPendingNonce(ctx context.Context, account common.Address) (uint64, error) {
	return eth.client.PendingNonceAt(ctx, account)
}

// wrapper around ethclient.SendTransaction
func (eth *Details) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return eth.client.SendTransaction(ctx, tx)
}

// How many blocks we should wait for removing a tx in case we don't find it in
// the layer1 chain
func (eth *Details) GetTxNotFoundMaxBlocks() uint64 {
	return constants.EthereumTxNotFoundMaxBlocks
}

// Number of blocks to wait for a tx in the memory pool w/o returning to the
// caller asking for retry
func (eth *Details) GetTxMaxStaleBlocks() uint64 {
	return constants.EthereumTxMaxStaleBlocks
}

func (eth *Details) GetPeerCount(ctx context.Context) (uint64, error) {
	// Let's see how many peers our endpoint has
	var peerCountString string
	if err := eth.rpcClient.CallContext(ctx, &peerCountString, "net_peerCount"); err != nil {
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

// GetAccount returns the account specified
func (eth *Details) GetAccount(addr common.Address) (accounts.Account, error) {
	acct, accountFound := eth.accounts[addr]
	if !accountFound {
		return acct, ErrAccountNotFound
	}

	return acct, nil
}

// Get the private key for an account
func (eth *Details) GetAccountKeys(addr common.Address) (*keystore.Key, error) {
	if key, ok := eth.keys[addr]; ok {
		return key, nil
	}
	return nil, ErrKeysNotFound
}

// GetDefaultAccount returns the default account
func (eth *Details) GetDefaultAccount() accounts.Account {
	return eth.defaultAccount
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

// Get ip address where are connected to the ethereum node
func (eth *Details) GetEndpoint() string {
	return eth.endpoint
}

// Get all ethereum accounts that we have access in the keystore.
func (eth *Details) GetKnownAccounts() []accounts.Account {
	accounts := make([]accounts.Account, len(eth.accounts))

	for address, accountIndex := range eth.accountIndex {
		accounts[accountIndex] = eth.accounts[address]
	}

	return accounts
}

// Get timeout context
func (eth *Details) GetTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), constants.MonitorTimeout)
}

// EndpointInSync Checks if our endpoint is good to use
// -- This function is different. Because we need to be aware of errors, State is always updated
func (eth *Details) EndpointInSync(ctx context.Context) (bool, uint32, error) {
	// Default to assuming everything is awful
	inSync := false
	peerCount := uint32(0)

	// Check if the endpoint is itself still syncing
	syncing, progress, err := eth.getSyncProgress()
	if err != nil {
		return inSync, peerCount, fmt.Errorf("Could not check if Ethereum endpoint it still syncing: %v", err)
	}

	if syncing && progress != nil {
		eth.logger.Debugf("Ethereum endpoint syncing... at block %v of %v.", progress.CurrentBlock, progress.HighestBlock)
	}

	peerCount64, err := eth.GetPeerCount(ctx)
	if err != nil {
		return inSync, peerCount, err
	}
	peerCount = uint32(peerCount64)

	if !syncing && peerCount >= uint32(eth.endpointMinimumPeers) {
		inSync = true
	}

	return inSync, peerCount, err
}

// Get ethereum events from a block range
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

// Get the max gas fee allowed to perform transactions in WEI
func (eth *Details) GetTxMaxGasFeeAllowed() *big.Int {
	return eth.txMaxGasFeeAllowed
}

// Get the base fee and suggestedGasTip for the latest ethereum block
func (eth *Details) GetBlockBaseFeeAndSuggestedGasTip(ctx context.Context) (*big.Int, *big.Int, error) {
	block, err := eth.client.BlockByNumber(ctx, nil)
	if err != nil && block == nil {
		return nil, nil, fmt.Errorf("could not get block number: %w", err)
	}

	eth.logger.Infof("Previous BaseFee:%v GasUsed:%v GasLimit:%v",
		block.BaseFee().String(),
		block.GasUsed(),
		block.GasLimit())

	baseFee := block.BaseFee()
	tipCap, err := eth.client.SuggestGasTipCap(ctx)
	if err != nil {
		if err.Error() == ETH_MAX_PRIORITY_FEE_PER_GAS_NOT_FOUND {
			tipCap = big.NewInt(1_000_000_000)
		} else {
			return nil, nil, fmt.Errorf("could not get suggested gas tip cap: %w", err)
		}
	}
	return baseFee, tipCap, nil
}

// Get transaction options in order to do a transaction
func (eth *Details) GetTransactionOpts(ctx context.Context, account accounts.Account) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyStoreTransactorWithChainID(eth.keystore, account, eth.chainID)
	if err != nil {
		return nil, fmt.Errorf("could not create transactor for %v: %v", account.Address.Hex(), err)
	}

	subCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	baseFee, tipCap, err := eth.GetBlockBaseFeeAndSuggestedGasTip(subCtx)
	if err != nil {
		return nil, err
	}

	feeCap, err := ComputeGasFeeCap(eth, baseFee, tipCap)
	if err != nil {
		return nil, err
	}

	eth.logger.WithFields(logrus.Fields{
		"MinerTip":          tipCap,
		"MaximumGasAllowed": eth.txMaxGasFeeAllowed.String(),
	}).Infof("Creating TX with MaximumGasPrice: %v WEI", feeCap)

	opts.Context = ctx
	opts.GasFeeCap = feeCap
	opts.GasTipCap = tipCap
	return opts, nil
}

// Safe function to call the smart contract state at the finalized block (block
// that we judge safe).
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

// Extracts the sender of a transaction
func (eth *Details) ExtractTransactionSender(tx *types.Transaction) (common.Address, error) {
	fromAddr, err := types.NewLondonSigner(eth.chainID).Sender(tx)
	if err != nil {
		return common.Address{}, err
	}
	return fromAddr, nil
}

// Retry a transaction done by the ethereum default account
func (eth *Details) RetryTransaction(ctx context.Context, tx *types.Transaction, baseFee *big.Int, gasTipCap *big.Int) (*types.Transaction, error) {
	fromAddr, err := eth.ExtractTransactionSender(tx)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(fromAddr[:], eth.defaultAccount.Address[:]) {
		return nil, fmt.Errorf("cannot retry transaction: %v from %v, expected signed tx from default account %v", tx.Hash().Hex(), fromAddr, eth.defaultAccount)
	}

	var newTipCap *big.Int
	if tx.GasTipCap() == nil || gasTipCap.Cmp(tx.GasTipCap()) >= 0 {
		newTipCap = gasTipCap
	} else {
		newTipCap = tx.GasTipCap()
	}

	// Increasing tip cap to replace old tx and make the tx more likely to be chosen
	// by a layer1 miner
	increasedTipCap := eth.bumpTipCap(newTipCap)
	gasFeeCap, err := ComputeGasFeeCap(eth, baseFee, increasedTipCap)
	if err != nil {
		return nil, err
	}

	txRough := &types.DynamicFeeTx{
		ChainID:   tx.ChainId(),
		Nonce:     tx.Nonce(),
		GasTipCap: increasedTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       tx.Gas(),
		To:        tx.To(),
		Value:     tx.Value(),
		Data:      tx.Data(),
	}

	signedTx, err := eth.SignTransaction(txRough, fromAddr)
	if err != nil {
		return nil, err
	}

	err = eth.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("sending error: %v", err)
	}

	return signedTx, err
}

func (eth *Details) SignTransaction(tx types.TxData, signerAddress common.Address) (*types.Transaction, error) {
	signer := types.NewLondonSigner(eth.chainID)
	userKey, err := eth.GetAccountKeys(signerAddress)
	if err != nil {
		eth.logger.Errorf("getting account keys error:%v", err)
		return nil, err
	}
	signedTx, err := types.SignNewTx(userKey.PrivateKey, signer, tx)
	if err != nil {
		eth.logger.Errorf("signing error:%v", err)
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

////////////////////////////////////////////////////////////////

// Get the current validators
func GetValidators(eth Network, logger *logrus.Logger, ctx context.Context) ([]common.Address, error) {
	c := eth.Contracts()
	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	if err != nil {
		return nil, err
	}
	validatorAddresses, err := c.ValidatorPool().GetValidatorsAddresses(callOpts)
	if err != nil {
		logger.Warnf("Could not call contract:%v", err)
		return nil, err
	}

	return validatorAddresses, nil
}

// TransferEther transfer's ether from one account to another, assumes from is unlocked
func TransferEther(eth Network, logger *logrus.Entry, from common.Address, to common.Address, wei *big.Int) (*types.Transaction, error) {
	ctx, cancel := eth.GetTimeoutContext()
	defer cancel()

	nonce, err := eth.GetPendingNonce(ctx, from)
	if err != nil {
		return nil, err
	}

	gasLimit := uint64(21000)
	baseFee, tipCap, err := eth.GetBlockBaseFeeAndSuggestedGasTip(ctx)
	if err != nil {
		return nil, err
	}
	feeCap, err := ComputeGasFeeCap(eth, baseFee, tipCap)
	if err != nil {
		return nil, err
	}
	chainID := eth.GetChainID()

	txRough := &types.DynamicFeeTx{
		ChainID:   chainID,
		To:        &to,
		GasFeeCap: feeCap,
		GasTipCap: tipCap,
		Gas:       gasLimit,
		Nonce:     nonce,
		Value:     wei,
	}

	logger.Debugf(
		"TransferEther => chainID:%v from:%v nonce:%v, to:%v, wei:%v, gasLimit:%v, gasPrice:%v",
		chainID,
		from.Hex(),
		nonce,
		to.Hex(),
		wei,
		gasLimit,
		feeCap,
	)
	signedTx, err := eth.SignTransaction(txRough, from)
	if err != nil {
		logger.Errorf("signing transaction failed: %v", err)
		return nil, err
	}
	err = eth.SendTransaction(ctx, signedTx)
	if err != nil {
		logger.Errorf("sending error: %v", err)
		return nil, err
	}
	return signedTx, nil
}

// Function to compute the gas fee that will be valid for the next 8 full blocks before we are priced out
func ComputeGasFeeCap(eth Network, baseFee *big.Int, tipCap *big.Int) (*big.Int, error) {
	baseFeeMultiplied := new(big.Int).Mul(big.NewInt(constants.EthereumBaseFeeMultiplier), baseFee)
	feeCap := new(big.Int).Add(baseFeeMultiplied, tipCap)
	if feeCap.Cmp(eth.GetTxMaxGasFeeAllowed()) > 0 {
		return nil, fmt.Errorf("max tx fee computed: %v is greater than limit: %v", feeCap.String(), eth.GetTxMaxGasFeeAllowed().String())
	}
	return feeCap, nil
}
