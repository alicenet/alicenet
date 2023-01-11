package evm

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/layer1"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/ethereum/go-ethereum/core/types"
	eCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
)

// Ethereum specific errors.
var (
	ErrAccountNotFound  = errors.New("could not find specified account")
	ErrKeysNotFound     = errors.New("account either not found or not unlocked")
	ErrPassCodeNotFound = errors.New("could not find specified passCode")
)

type ErrTxTooExpensive struct {
	message string
}

func (e *ErrTxTooExpensive) Error() string {
	return e.message
}

var ETH_MAX_PRIORITY_FEE_PER_GAS_NOT_FOUND string = "Method eth_maxPriorityFeePerGas not found"

var _ layer1.Client = &Client{}

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

type accountInfo struct {
	account  accounts.Account
	index    int
	key      *keystore.Key
	passCode string
}

type Client struct {
	logger               *logrus.Logger
	endpoint             string
	keystore             *keystore.KeyStore
	finalityDelay        uint64
	accounts             map[common.Address]accountInfo
	defaultAccount       accounts.Account
	rpcClient            *rpc.Client
	internalClient       *ethclient.Client
	chainID              *big.Int
	txMaxGasFeeAllowed   *big.Int
	endpointMinimumPeers uint64
}

// NewClient creates a new Ethereum abstraction.
func NewClient(
	logger *logrus.Logger,
	endpoint string,
	pathKeystore string,
	pathPassCodes string,
	defaultAccount string,
	unlockAdditionalAccounts bool,
	finalityDelay uint64,
	txMaxGasFeeAllowedInGwei uint64,
	endpointMinimumPeers uint64,
) (*Client, error) {

	if txMaxGasFeeAllowedInGwei < constants.EthereumMinGasFeeAllowedInGwei {
		return nil, fmt.Errorf(
			"txMaxGasFeeAllowedInGwei should be greater than %v Gwei",
			constants.EthereumMinGasFeeAllowedInGwei,
		)
	}

	txMaxGasFeeAllowedInWei := new(
		big.Int,
	).Mul(new(big.Int).SetUint64(txMaxGasFeeAllowedInGwei), new(big.Int).SetUint64(1_000_000_000))

	cl := &Client{
		endpoint:           endpoint,
		logger:             logger,
		accounts:           make(map[common.Address]accountInfo),
		finalityDelay:      finalityDelay,
		txMaxGasFeeAllowed: txMaxGasFeeAllowedInWei,
	}

	// Load accounts + passCodes
	cl.loadAccounts(pathKeystore)
	err := cl.loadPassCodes(pathPassCodes)
	if err != nil {
		return nil, fmt.Errorf("Error in NewEthereumEndpoint at cl.LoadPassCodes: %v", err)
	}

	// Designate accounts
	var acct accounts.Account
	acct, err = cl.GetAccount(common.HexToAddress(defaultAccount))
	if err != nil {
		return nil, fmt.Errorf("Can't find user to set as default %v: %v", defaultAccount, err)
	}
	cl.setDefaultAccount(acct)
	if err := cl.unlockAccount(acct); err != nil {
		return nil, fmt.Errorf("Could not unlock account: %v", err)
	}

	// if this flag is set, unlock all known accounts (in the pathKeystore). This
	// should allow performing transactions with accounts that are not the default
	// account.
	if unlockAdditionalAccounts {
		accountList := cl.GetKnownAccounts()
		for _, acct := range accountList {
			if err := cl.unlockAccount(acct); err != nil {
				return nil, fmt.Errorf("Could not unlock additional account: %v", err)
			}
		}
	}

	// Low level rpc client
	ctx, cancel := context.WithTimeout(context.Background(), constants.MonitorTimeout)
	defer cancel()
	rpcClient, rpcErr := rpc.DialContext(ctx, endpoint)
	if rpcErr != nil {
		return nil, fmt.Errorf("Error in NewEndpoint at rpc.DialContext: %v", rpcErr)
	}
	cl.rpcClient = rpcClient
	ethClient := ethclient.NewClient(rpcClient)
	cl.internalClient = ethClient
	cl.chainID, err = ethClient.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error in NewEndpoint at ethClient.ChainID: %v", err)
	}

	logger.Debug("Completed initialization")

	return cl, nil
}

// LoadAccounts Scans the directory specified and loads all the accounts found.
func (cl *Client) loadAccounts(directoryPath string) {
	logger := cl.logger

	logger.Infof("LoadAccounts(\"%v\")...", directoryPath)
	ks := keystore.NewKeyStore(directoryPath, keystore.StandardScryptN, keystore.StandardScryptP)
	accts := make(map[common.Address]accountInfo, 10)

	var index int
	for _, wallet := range ks.Wallets() {
		for _, account := range wallet.Accounts() {
			logger.Infof("... found account %v", account.Address.Hex())
			accts[account.Address] = accountInfo{account: account, index: index}
			index++
		}
	}

	cl.accounts = accts
	cl.keystore = ks
}

// LoadPasscodes loads the specified passcode file.
func (cl *Client) loadPassCodes(filePath string) error {
	logger := cl.logger

	logger.Infof("loadPassCodes(\"%v\")...", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		logger.Errorf("Failed to open passCode file \"%v\": %s", filePath, err)
	}
	defer file.Close()

	passCodes := make(map[string]string)
	if file != nil {
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, "#") {
				components := strings.Split(line, "=")
				if len(components) == 2 {
					keyHex := strings.TrimPrefix(
						strings.ToLower(strings.TrimSpace(components[0])),
						"0x",
					)
					passCodes[keyHex] = strings.TrimSpace(components[1])
				}
			}
		}
	}

	for addr, accountInfo := range cl.accounts {
		passCode, present := passCodes[strings.TrimPrefix(strings.ToLower(addr.Hex()), "0x")]
		if !present {
			passCode, err = prompt.Stdin.PromptPassword(
				fmt.Sprintf("Please provide the passCode for this address %s: ", addr.Hex()),
			)
			if err != nil {
				return err
			}
		}
		accountInfo.passCode = passCode
		cl.accounts[addr] = accountInfo
	}

	return nil
}

// UnlockAccount unlocks the previously loaded account using the previously loaded passCodes.
func (cl *Client) unlockAccount(acct accounts.Account) error {
	cl.logger.Infof("Unlocking account address:%v", acct.Address.String())

	accountInfo, accountNotFound := cl.accounts[acct.Address]
	if !accountNotFound || accountInfo.passCode == "" {
		return ErrPassCodeNotFound
	}
	passCode := accountInfo.passCode

	err := cl.keystore.Unlock(acct, passCode)
	if err != nil {
		return err
	}

	// Open the account key file
	keyJSON, err := os.ReadFile(acct.URL.Path)
	if err != nil {
		return err
	}

	// Get the private key
	key, err := keystore.DecryptKey(keyJSON, passCode)
	if err != nil {
		return err
	}

	accountInfo.key = key
	cl.accounts[acct.Address] = accountInfo

	return nil
}

// setDefaultAccount designates the account to be used by default.
func (cl *Client) setDefaultAccount(acct accounts.Account) {
	cl.defaultAccount = acct
}

// bump the tip cap for retries.
func (cl *Client) bumpTipCap(gasTipCap *big.Int) *big.Int {
	// calculate percentage% increase in GasTipCap
	gasTipCapPercent := new(
		big.Int,
	).Mul(gasTipCap, big.NewInt(int64(constants.EthereumTipCapPercentageBump)))
	gasTipCapPercent = new(big.Int).Div(gasTipCapPercent, big.NewInt(100))
	resultTipCap := new(big.Int).Add(gasTipCap, gasTipCapPercent)
	return resultTipCap
}

// getSyncProgress returns a flag if we are syncing, a pointer to a struct if we are, or an error.
func (cl *Client) getSyncProgress() (bool, *ethereum.SyncProgress, error) {
	ctx, ctxCancel := cl.GetTimeoutContext()
	defer ctxCancel()
	progress, err := cl.internalClient.SyncProgress(ctx)
	if err != nil {
		return false, nil, err
	}

	if progress == nil {
		return false, nil, nil
	}

	return true, progress, nil
}

// Get the private key for an account.
func (cl *Client) getAccountKeys(addr common.Address) (*keystore.Key, error) {
	accountInfo, ok := cl.accounts[addr]
	if !ok || accountInfo.key == nil {
		return nil, ErrKeysNotFound
	}
	return accountInfo.key, nil
}

// ChainID returns the ID used to build ethereum client.
func (cl *Client) GetChainID() *big.Int {
	return cl.chainID
}

// Get finality delay.
func (cl *Client) GetFinalityDelay() uint64 {
	return cl.finalityDelay
}

// close the ethereum client.
func (cl *Client) Close() {
	cl.internalClient.Close()
}

// wrapper around ethclient.TransactionByHash.
func (cl *Client) GetTransactionByHash(
	ctx context.Context,
	txHash common.Hash,
) (tx *types.Transaction, isPending bool, err error) {
	return cl.internalClient.TransactionByHash(ctx, txHash)
}

// wrapper around ethclient.TransactionReceipt.
func (cl *Client) GetTransactionReceipt(
	ctx context.Context,
	txHash common.Hash,
) (*types.Receipt, error) {
	return cl.internalClient.TransactionReceipt(ctx, txHash)
}

func (cl *Client) GetInternalClient() *ethclient.Client {
	return cl.internalClient
}

func (cl *Client) GetLogger() *logrus.Logger {
	return cl.logger
}

// wrapper around ethclient.HeaderByNumber.
func (cl *Client) GetHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return cl.internalClient.HeaderByNumber(ctx, number)
}

// wrapper around ethclient.BlockByNumber.
func (cl *Client) GetBlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return cl.internalClient.BlockByNumber(ctx, number)
}

// wrapper around ethclient.PendingNonceAt.
func (cl *Client) GetPendingNonce(ctx context.Context, account common.Address) (uint64, error) {
	return cl.internalClient.PendingNonceAt(ctx, account)
}

// wrapper around ethclient.SendTransaction.
func (cl *Client) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return cl.internalClient.SendTransaction(ctx, tx)
}

// How many blocks we should wait for removing a tx in case we don't find it in
// the layer1 chain.
func (cl *Client) GetTxNotFoundMaxBlocks() uint64 {
	return constants.EthereumTxNotFoundMaxBlocks
}

// Number of blocks to wait for a tx in the memory pool w/o returning to the
// caller asking for retry.
func (cl *Client) GetTxMaxStaleBlocks() uint64 {
	return constants.EthereumTxMaxStaleBlocks
}

func (cl *Client) GetPeerCount(ctx context.Context) (uint64, error) {
	// Let's see how many peers our endpoint has
	var peerCountString string
	if err := cl.rpcClient.CallContext(ctx, &peerCountString, "net_peerCount"); err != nil {
		cl.logger.Warnf("could not get peerCount: %v", err)
		return 0, err
	}

	var peerCount uint64
	_, err := fmt.Sscanf(peerCountString, "0x%x", &peerCount)
	if err != nil {
		cl.logger.Warnf("could not parse peerCount: %v", err)
		return 0, err
	}
	return peerCount, nil
}

// IsAccessible checks against endpoint to confirm server responds.
func (cl *Client) IsAccessible() bool {
	ctx, cancel := cl.GetTimeoutContext()
	defer cancel()
	block, err := cl.internalClient.BlockByNumber(ctx, nil)
	if err == nil && block != nil {
		return true
	}
	cl.logger.WithError(err).Warning("IsEthereumAccessible()...false")
	return false
}

// GetAccount returns the account specified.
func (cl *Client) GetAccount(addr common.Address) (accounts.Account, error) {
	accountInfo, accountFound := cl.accounts[addr]
	if !accountFound {
		return accounts.Account{}, ErrAccountNotFound
	}

	return accountInfo.account, nil
}

// GetDefaultAccount returns the default account.
func (cl *Client) GetDefaultAccount() accounts.Account {
	return cl.defaultAccount
}

// GetBalance returns the ETHER balance of account specified.
func (cl *Client) GetBalance(addr common.Address) (*big.Int, error) {
	ctx, cancel := cl.GetTimeoutContext()
	defer cancel()
	balance, err := cl.internalClient.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

// Get ip address where are connected to the ethereum node.
func (cl *Client) GetEndpoint() string {
	return cl.endpoint
}

// Get all ethereum accounts that we have access in the keystore.
func (cl *Client) GetKnownAccounts() []accounts.Account {
	accounts := make([]accounts.Account, len(cl.accounts))

	for _, accountInfo := range cl.accounts {
		accounts[accountInfo.index] = accountInfo.account
	}

	return accounts
}

// Get timeout context.
func (cl *Client) GetTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), constants.MonitorTimeout)
}

// EndpointInSync Checks if our endpoint is good to use
// -- This function is different. Because we need to be aware of errors, State is always updated.
func (cl *Client) EndpointInSync(ctx context.Context) (bool, uint32, error) {
	// Default to assuming everything is awful
	inSync := false
	peerCount := uint32(0)

	// Check if the endpoint is itself still syncing
	syncing, progress, err := cl.getSyncProgress()
	if err != nil {
		return inSync, peerCount, fmt.Errorf(
			"Could not check if Ethereum endpoint it still syncing: %v",
			err,
		)
	}

	if syncing && progress != nil {
		cl.logger.Debugf(
			"Ethereum endpoint syncing... at block %v of %v.",
			progress.CurrentBlock,
			progress.HighestBlock,
		)
	}

	peerCount64, err := cl.GetPeerCount(ctx)
	if err != nil {
		return inSync, peerCount, err
	}
	peerCount = uint32(peerCount64)

	if !syncing && peerCount >= uint32(cl.endpointMinimumPeers) {
		inSync = true
	}

	return inSync, peerCount, err
}

// Get ethereum events from a block range.
func (cl *Client) GetEvents(
	ctx context.Context,
	firstBlock, lastBlock uint64,
	addresses []common.Address,
) ([]types.Log, error) {
	logger := cl.logger

	logger.Tracef(
		"...GetEvents(firstBlock:%v,lastBlock:%v,addresses:%x)",
		firstBlock,
		lastBlock,
		addresses,
	)

	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(firstBlock),
		ToBlock:   new(big.Int).SetUint64(lastBlock),
		Addresses: addresses,
	}

	logs, err := cl.internalClient.FilterLogs(ctx, query)
	if err != nil {
		logger.Errorf("Could not filter logs: %v", err)
		return nil, err
	}

	for idx, log := range logs {
		logger.Tracef("Log[%v] Block[%v]:%v", idx, log.BlockNumber, log)
		for idx, hash := range log.Topics {
			logger.Tracef("Hash[%v]:%x", idx, hash)
		}
	}

	return logs, nil
}

// Get the max gas fee allowed to perform transactions in WEI.
func (cl *Client) GetTxMaxGasFeeAllowed() *big.Int {
	return cl.txMaxGasFeeAllowed
}

// Get the base fee and suggestedGasTip for the latest ethereum block.
func (cl *Client) GetBlockBaseFeeAndSuggestedGasTip(
	ctx context.Context,
) (*big.Int, *big.Int, error) {
	block, err := cl.internalClient.BlockByNumber(ctx, nil)
	if err != nil && block == nil {
		return nil, nil, fmt.Errorf("could not get block number: %w", err)
	}

	cl.logger.Infof("Previous BaseFee:%v GasUsed:%v GasLimit:%v",
		block.BaseFee().String(),
		block.GasUsed(),
		block.GasLimit())

	baseFee := block.BaseFee()
	tipCap, err := cl.internalClient.SuggestGasTipCap(ctx)
	if err != nil {
		if err.Error() == ETH_MAX_PRIORITY_FEE_PER_GAS_NOT_FOUND {
			tipCap = big.NewInt(1_000_000_000)
		} else {
			return nil, nil, fmt.Errorf("could not get suggested gas tip cap: %w", err)
		}
	}
	return baseFee, tipCap, nil
}

// Get transaction options in order to do a transaction.
func (cl *Client) GetTransactionOpts(
	ctx context.Context,
	account accounts.Account,
) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyStoreTransactorWithChainID(cl.keystore, account, cl.chainID)
	if err != nil {
		return nil, fmt.Errorf("could not create transactor for %v: %v", account.Address.Hex(), err)
	}

	subCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	baseFee, tipCap, err := cl.GetBlockBaseFeeAndSuggestedGasTip(subCtx)
	if err != nil {
		return nil, err
	}

	feeCap, err := ComputeGasFeeCap(cl, baseFee, tipCap)
	if err != nil {
		return nil, err
	}

	cl.logger.WithFields(logrus.Fields{
		"MinerTip":          tipCap,
		"MaximumGasAllowed": cl.txMaxGasFeeAllowed.String(),
	}).Infof("Creating TX with MaximumGasPrice: %v WEI", feeCap)

	opts.Context = ctx
	opts.GasFeeCap = feeCap
	opts.GasTipCap = tipCap
	return opts, nil
}

// Safe function to call the smart contract state at the finalized block (block
// that we judge safe).
func (cl *Client) GetCallOpts(
	ctx context.Context,
	account accounts.Account,
) (*bind.CallOpts, error) {
	finalizedHeightU64, err := cl.GetFinalizedHeight(ctx)
	if err != nil {
		return nil, err
	}
	finalizedHeight := new(big.Int).SetUint64(finalizedHeightU64)
	return &bind.CallOpts{
		BlockNumber: finalizedHeight,
		Context:     ctx,
		Pending:     false,
		From:        account.Address,
	}, nil
}

// Function to call the smart contract state at the latest block seen by the
// ethereum node. USE THIS FUNCTION CAREFULLY, THE LATEST BLOCK IS SUSCEPTIBLE
// TO CHAIN RE-ORGS AND SHOULD NEVER BE USED AS TEST IF SOMETHING WAS COMMITTED
// TO OUR CONTRACTS. IDEALLY, ONLY USE THIS FUNCTION FOR MONITORING FUNCTIONS
// THAT DEPENDS ON THE LATEST BLOCK. Otherwise use GetCallOpts!!!
func (cl *Client) GetCallOptsLatestBlock(
	ctx context.Context,
	account accounts.Account,
) *bind.CallOpts {
	return &bind.CallOpts{
		BlockNumber: nil,
		Context:     ctx,
		Pending:     false,
		From:        account.Address,
	}
}

// Extracts the sender of a transaction.
func (cl *Client) ExtractTransactionSender(tx *types.Transaction) (common.Address, error) {
	fromAddr, err := types.NewLondonSigner(cl.chainID).Sender(tx)
	if err != nil {
		return common.Address{}, err
	}
	return fromAddr, nil
}

// Retry a transaction done by any of the known ethereum accounts.
func (cl *Client) RetryTransaction(
	ctx context.Context,
	tx *types.Transaction,
	baseFee, gasTipCap *big.Int,
) (*types.Transaction, error) {
	fromAddr, err := cl.ExtractTransactionSender(tx)
	if err != nil {
		return nil, err
	}

	var newTipCap *big.Int

	if tx.GasTipCap() == nil || gasTipCap.Cmp(tx.GasTipCap()) >= 0 {
		newTipCap = gasTipCap
	} else {
		newTipCap = tx.GasTipCap()
	}

	// Increasing tip cap to replace old tx and make the tx more likely to be chosen
	// by a layer1 miner
	increasedTipCap := cl.bumpTipCap(newTipCap)

	// In case we reach the new tip cap, we start to use the the suggested tip
	// again. If we reached this point, our previous transaction may be already get
	// pruned from the node and our tx will likely fail. So, we try to restart the
	// gas tip and see if a transaction will be accepted in an attempt to pay less.
	maxTipCap := new(big.Int).Mul(gasTipCap, big.NewInt(constants.EthereumMaxGasTipMultiplier))
	if increasedTipCap.Cmp(maxTipCap) > 0 {
		increasedTipCap = gasTipCap
	}

	gasFeeCap, err := ComputeGasFeeCap(cl, baseFee, increasedTipCap)
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

	signedTx, err := cl.SignTransaction(txRough, fromAddr)
	if err != nil {
		return nil, err
	}

	err = cl.internalClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("sending tx error: %v", err)
	}

	return signedTx, err
}

// Sign an ethereum transaction.
func (cl *Client) SignTransaction(
	tx types.TxData,
	signerAddress common.Address,
) (*types.Transaction, error) {
	signer := types.NewLondonSigner(cl.chainID)
	userKey, err := cl.getAccountKeys(signerAddress)
	if err != nil {
		return nil, fmt.Errorf("getting account keys error:%v", err)
	}
	signedTx, err := types.SignNewTx(userKey.PrivateKey, signer, tx)
	if err != nil {
		return nil, fmt.Errorf("signing error:%v", err)
	}
	return signedTx, nil
}

// GetCurrentHeight gets the height of the endpoints chain.
func (cl *Client) GetCurrentHeight(ctx context.Context) (uint64, error) {
	return cl.internalClient.BlockNumber(ctx)
}

// GetFinalizedHeight gets the height of the endpoints chain at which is is considered finalized.
func (cl *Client) GetFinalizedHeight(ctx context.Context) (uint64, error) {
	height, err := cl.GetCurrentHeight(ctx)
	if err != nil {
		return height, err
	}

	if cl.finalityDelay >= height {
		return 0, nil
	}
	return height - cl.finalityDelay, nil
}

// create a new signer for ETH accounts.
func (cl *Client) CreateSecp256k1Signer() (*crypto.Secp256k1Signer, error) {
	secp256k1Signer := &crypto.Secp256k1Signer{}
	key, err := cl.getAccountKeys(cl.defaultAccount.Address)
	if err != nil {
		return nil, err
	}
	err = secp256k1Signer.SetPrivk(eCrypto.FromECDSA(key.PrivateKey))
	if err != nil {
		return nil, err
	}
	return secp256k1Signer, nil
}

// TransferEther transfer's ether/native token from one account to another, assumes from is unlocked.
func (cl *Client) TransferNativeToken(
	from, to common.Address,
	wei *big.Int,
) (*types.Transaction, error) {
	ctx, cancel := cl.GetTimeoutContext()
	defer cancel()

	nonce, err := cl.GetPendingNonce(ctx, from)
	if err != nil {
		return nil, err
	}

	gasLimit := uint64(21000)
	baseFee, tipCap, err := cl.GetBlockBaseFeeAndSuggestedGasTip(ctx)
	if err != nil {
		return nil, err
	}
	feeCap, err := ComputeGasFeeCap(cl, baseFee, tipCap)
	if err != nil {
		return nil, err
	}
	chainID := cl.GetChainID()

	txRough := &types.DynamicFeeTx{
		ChainID:   chainID,
		To:        &to,
		GasFeeCap: feeCap,
		GasTipCap: tipCap,
		Gas:       gasLimit,
		Nonce:     nonce,
		Value:     wei,
	}

	cl.logger.WithFields(logrus.Fields{
		"chainID":  chainID,
		"from":     from.Hex(),
		"nonce":    nonce,
		"to":       to.Hex(),
		"wei":      wei,
		"gasLimit": gasLimit,
		"gasPrice": feeCap,
	}).Debug("Transferring ether")
	signedTx, err := cl.SignTransaction(txRough, from)
	if err != nil {
		cl.logger.Errorf("signing transaction failed: %v", err)
		return nil, err
	}
	err = cl.SendTransaction(ctx, signedTx)
	if err != nil {
		cl.logger.Errorf("sending error: %v", err)
		return nil, err
	}
	return signedTx, nil
}

// Function to compute the gas fee that will be valid for the next 8 full blocks before we are priced out.
func ComputeGasFeeCap(cl layer1.Client, baseFee, tipCap *big.Int) (*big.Int, error) {
	baseFeeMultiplied := new(big.Int).Mul(big.NewInt(constants.EthereumBaseFeeMultiplier), baseFee)
	feeCap := new(big.Int).Add(baseFeeMultiplied, tipCap)
	if feeCap.Cmp(cl.GetTxMaxGasFeeAllowed()) > 0 {
		return nil, &ErrTxTooExpensive{
			fmt.Sprintf(
				"max tx fee computed: %v is greater than limit: %v",
				feeCap.String(),
				cl.GetTxMaxGasFeeAllowed().String(),
			),
		}
	}
	return feeCap, nil
}
