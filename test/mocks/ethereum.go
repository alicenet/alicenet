package mocks

import (
	"context"
	"github.com/MadBase/MadNet/blockchain/ethereum"
	bind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	types "github.com/ethereum/go-ethereum/core/types"
)

type EthereumMock struct {
	*MockNetwork
	ContractsMock        *MockContracts
	ETHDKGMock           *MockIETHDKG
	GovernanceMock       *MockIGovernance
	ATokenMock           *MockIAToken
	BTokenMock           *MockIBToken
	PublicStakingMock    *MockIPublicStaking
	SnapshotsMock        *MockISnapshots
	ValidatorPoolMock    *MockIValidatorPool
	ValidatorStakingMock *MockIValidatorStaking
}

var _ ethereum.Network = (*MockNetwork)(nil)

func NewMockEthereum() *EthereumMock {
	eth := NewMockNetwork()
	var bh uint64 = 0
	eth.GetCurrentHeightFunc.SetDefaultHook(func(context.Context) (uint64, error) { bh++; return bh, nil })
	eth.GetFinalityDelayFunc.SetDefaultReturn(6)
	eth.GetTransactionOptsFunc.SetDefaultReturn(&bind.TransactOpts{}, nil)

	contracts := NewMockContracts()
	eth.ContractsFunc.SetDefaultReturn(contracts)

	ethdkg := NewMockIETHDKG()
	contracts.EthdkgFunc.SetDefaultReturn(ethdkg)

	governance := NewMockIGovernance()
	contracts.GovernanceFunc.SetDefaultReturn(governance)

	atoken := NewMockIAToken()
	contracts.ATokenFunc.SetDefaultReturn(atoken)

	btoken := NewMockIBToken()
	contracts.BTokenFunc.SetDefaultReturn(btoken)

	publicstaking := NewMockIPublicStaking()
	contracts.PublicStakingFunc.SetDefaultReturn(publicstaking)

	snapshots := NewMockLinkedSnapshots()
	contracts.SnapshotsFunc.SetDefaultReturn(snapshots)

	validatorpool := NewMockIValidatorPool()
	contracts.ValidatorPoolFunc.SetDefaultReturn(validatorpool)

	validatorstaking := NewMockIValidatorStaking()
	contracts.ValidatorStakingFunc.SetDefaultReturn(validatorstaking)

	return &EthereumMock{
		MockNetwork:   eth,
		ContractsMock: contracts,

		ETHDKGMock:           ethdkg,
		GovernanceMock:       governance,
		ATokenMock:           atoken,
		BTokenMock:           btoken,
		PublicStakingMock:    publicstaking,
		SnapshotsMock:        snapshots,
		ValidatorPoolMock:    validatorpool,
		ValidatorStakingMock: validatorstaking,
	}
}

func NewMockLinkedSnapshots() *MockISnapshots {
	m := NewMockISnapshots()
	m.SnapshotFunc.SetDefaultHook(func(*bind.TransactOpts, []byte, []byte) (*types.Transaction, error) { return NewMockSnapshotTx(), nil })
	return m
}

func NewMockLinkedTransactionWatcher() *MockIWatcher {
	transaction := NewMockIWatcher()
	transaction.WaitFunc.SetDefaultReturn(&types.Receipt{Status: 1}, nil)
	return transaction
}

// //NewSimulator returns a simulator for testing
// func NewSimulator(
// 	privateKeys []*ecdsa.PrivateKey,
// 	finalityDelay int,
// 	wei *big.Int,
// 	txMaxGasFeeAllowedInGwei uint64,
// ) (*Details, error) {
// 	logger := logging.GetLogger("ethereum")

// 	if len(privateKeys) < 1 {
// 		return nil, errors.New("at least 1 private key")
// 	}

// 	pathKeyStore, err := ioutil.TempDir("", "simulator-keystore-")
// 	if err != nil {
// 		return nil, err
// 	}

// 	eth := &Details{}
// 	eth.accounts = make(map[common.Address]accounts.Account)
// 	eth.accountIndex = make(map[common.Address]int)
// 	eth.contracts = &ContractDetails{eth: eth}
// 	eth.finalityDelay = uint64(finalityDelay)
// 	eth.keystore = keystore.NewKeyStore(pathKeyStore, keystore.StandardScryptN, keystore.StandardScryptP)
// 	eth.keys = make(map[common.Address]*keystore.Key)
// 	eth.logger = logger
// 	eth.passCodes = make(map[common.Address]string)
// 	eth.txMaxGasFeeAllowedInGwei = txMaxGasFeeAllowedInGwei
// 	for idx, privateKey := range privateKeys {
// 		account, err := eth.keystore.ImportECDSA(privateKey, "abc123")
// 		if err != nil {
// 			return nil, err
// 		}

// 		eth.accounts[account.Address] = account
// 		eth.accountIndex[account.Address] = idx
// 		eth.passCodes[account.Address] = "abc123"

// 		logger.Debugf("Account address:%v", account.Address.String())

// 		keyID, err := uuid.NewRandom()
// 		if err != nil {
// 			return nil, err
// 		}

// 		eth.keys[account.Address] = &keystore.Key{Address: account.Address, PrivateKey: privateKey, Id: keyID}

// 		if idx == 0 {
// 			eth.defaultAccount = account
// 		}
// 	}

// 	genAlloc := make(core.GenesisAlloc)
// 	for address := range eth.accounts {
// 		genAlloc[address] = core.GenesisAccount{Balance: wei}
// 	}

// 	client, err := ethclient.Dial("http://127.0.0.1:8545")
// 	if err != nil {
// 		return nil, err
// 	}
// 	eth.client = client

// 	eth.chainID = big.NewInt(1337)
// 	eth.peerCount = func(context.Context) (uint64, error) {
// 		return 0, nil
// 	}
// 	eth.syncing = func(ctx context.Context) (*ethereum.SyncProgress, error) {
// 		return nil, nil
// 	}

// 	// eth.commit = func() {
// 	// 	c := http.Client{}
// 	// 	msg := &JsonRPCMessage{
// 	// 		Version: "2.0",
// 	// 		ID:      []byte("1"),
// 	// 		Method:  "evm_mine",
// 	// 		Params:  make([]byte, 0),
// 	// 	}

// 	// 	if msg.Params, err = json.Marshal(make([]string, 0)); err != nil {
// 	// 		panic(err)
// 	// 	}

// 	// 	var buff bytes.Buffer
// 	// 	err := json.NewEncoder(&buff).Encode(msg)
// 	// 	if err != nil {
// 	// 		log.Fatal(err)
// 	// 	}

// 	// 	retryCount := 5
// 	// 	var worked bool
// 	// 	for i := 0; i < retryCount; i++ {
// 	// 		reader := bytes.NewReader(buff.Bytes())
// 	// 		resp, err := c.Post(
// 	// 			"http://127.0.0.1:8545",
// 	// 			"application/json",
// 	// 			reader,
// 	// 		)

// 	// 		if err != nil {
// 	// 			log.Printf("error calling evm_mine rpc: %v", err)
// 	// 			<-time.After(5 * time.Second)
// 	// 			continue
// 	// 		}

// 	// 		_, err = io.ReadAll(resp.Body)
// 	// 		if err != nil {
// 	// 			log.Printf("error reading response from evm_mine rpc: %v", err)
// 	// 			<-time.After(5 * time.Second)
// 	// 			continue
// 	// 		}

// 	// 		worked = true
// 	// 		break
// 	// 	}

// 	// 	if !worked {
// 	// 		panic(fmt.Errorf("error committing evm_mine on rpc: %v", err))
// 	// 	}
// 	// }

// 	return eth, nil
// }
