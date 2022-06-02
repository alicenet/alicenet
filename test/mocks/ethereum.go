package mocks

import (
	"context"
	"math/big"
	"time"

	bind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	types "github.com/ethereum/go-ethereum/core/types"
)

type EthereumMock struct {
	*MockBaseEthereum
	GethClientMock         *MockGethClient
	TransactionWatcherMock *MockITransactionWatcher
	ContractsMock          *MockContracts

	ETHDKGMock           *MockIETHDKG
	GovernanceMock       *MockIGovernance
	ATokenMock           *MockIAToken
	BTokenMock           *MockIBToken
	PublicStakingMock    *MockIPublicStaking
	SnapshotsMock        *MockISnapshots
	ValidatorPoolMock    *MockIValidatorPool
	ValidatorStakingMock *MockIValidatorStaking
}

var _ ethereumInterfaces.IEthereum = (*MockBaseEthereum)(nil)

func NewMockEthereum() *EthereumMock {
	eth := NewMockBaseEthereum()
	var bh uint64 = 0
	eth.GetCurrentHeightFunc.SetDefaultHook(func(context.Context) (uint64, error) { bh++; return bh, nil })
	eth.GetFinalityDelayFunc.SetDefaultReturn(6)
	eth.GetTransactionOptsFunc.SetDefaultReturn(&bind.TransactOpts{}, nil)
	eth.RetryCountFunc.SetDefaultReturn(3)
	eth.RetryDelayFunc.SetDefaultReturn(time.Millisecond)
	eth.GetTxFeePercentageToIncreaseFunc.SetDefaultReturn(50)
	eth.GetTxMaxGasFeeAllowedInGweiFunc.SetDefaultReturn(500)

	geth := NewMockLinkedGethClient()
	eth.GetGethClientFunc.SetDefaultReturn(geth)

	txWatcher := NewMockLinkedTransactionWatcher()
	eth.TransactionWatcherFunc.SetDefaultReturn(txWatcher)

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
		MockBaseEthereum:       eth,
		GethClientMock:         geth,
		TransactionWatcherMock: txWatcher,
		ContractsMock:          contracts,

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

func NewMockLinkedTransactionWatcher() *MockITransactionWatcher {
	txWatcher := NewMockITransactionWatcher()
	txWatcher.WaitTransactionFunc.SetDefaultReturn(&types.Receipt{Status: 1}, nil)
	return txWatcher
}

func NewMockLinkedGethClient() *MockGethClient {
	geth := NewMockGethClient()
	geth.SuggestGasTipCapFunc.SetDefaultReturn(big.NewInt(15000), nil)
	geth.SuggestGasPriceFunc.SetDefaultReturn(big.NewInt(1000), nil)
	return geth
}
