// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// BClaimsParserLibraryBClaims is an auto generated low-level Go binding around an user-defined struct.
type BClaimsParserLibraryBClaims struct {
	ChainId    uint32
	Height     uint32
	TxCount    uint32
	PrevBlock  [32]byte
	TxRoot     [32]byte
	StateRoot  [32]byte
	HeaderRoot [32]byte
}

// Snapshot is an auto generated low-level Go binding around an user-defined struct.
type Snapshot struct {
	CommittedAt *big.Int
	BlockClaims BClaimsParserLibraryBClaims
}

// SnapshotsMetaData contains all meta data concerning the Snapshots contract.
var SnapshotsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"chainID_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"epochLength_\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"Bytes32OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"Bytes32OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ChainIdZero\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ConsensusNotRunning\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"dataOffset\",\"type\":\"uint256\"}],\"name\":\"DataOffsetOverflow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DataOffsetOverflow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"EpochMustBeNonZero\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"HeightZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"bytesLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"requiredBytesLength\",\"type\":\"uint256\"}],\"name\":\"InsufficientBytes\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blockHeight\",\"type\":\"uint256\"}],\"name\":\"InvalidBlockHeight\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"}],\"name\":\"InvalidChainId\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"calculatedMasterKeyHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"expectedMasterKeyHash\",\"type\":\"bytes32\"}],\"name\":\"InvalidMasterPublicKey\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newBlockHeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"oldBlockHeight\",\"type\":\"uint256\"}],\"name\":\"InvalidRingBufferBlockHeight\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"LEUint16OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"LEUint16OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"LEUint256OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"LEUint256OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"groupSignatureLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bClaimsLength\",\"type\":\"uint256\"}],\"name\":\"MigrationInputDataMismatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MigrationNotAllowedAtCurrentEpoch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"currentBlocksInterval\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimumBlocksInterval\",\"type\":\"uint256\"}],\"name\":\"MinimumBlocksIntervalNotPassed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"dataOffset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"NotEnoughBytes\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyDynamics\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyETHDKG\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyFactory\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyValidatorPool\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"OnlyValidatorsAllowed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SignatureVerificationFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"dataSectionSize\",\"type\":\"uint16\"}],\"name\":\"SizeThresholdExceeded\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"}],\"name\":\"SnapshotsNotInBuffer\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"groupSignatureHash\",\"type\":\"bytes32\"}],\"name\":\"ValidatorNotElected\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"height\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isSafeToProceedConsensus\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"uint256[4]\",\"name\":\"masterPublicKey\",\"type\":\"uint256[4]\"},{\"indexed\":false,\"internalType\":\"uint256[2]\",\"name\":\"signature\",\"type\":\"uint256[2]\"},{\"components\":[{\"internalType\":\"uint32\",\"name\":\"chainId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"height\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"txCount\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"prevBlock\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"txRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"headerRoot\",\"type\":\"bytes32\"}],\"indexed\":false,\"internalType\":\"structBClaimsParserLibrary.BClaims\",\"name\":\"bClaims\",\"type\":\"tuple\"}],\"name\":\"SnapshotTaken\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"groupSignature_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"bClaims_\",\"type\":\"bytes\"}],\"name\":\"checkBClaimsSignature\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAliceNetHeightFromLatestSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch_\",\"type\":\"uint256\"}],\"name\":\"getAliceNetHeightFromSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getBlockClaimsFromLatestSnapshot\",\"outputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"chainId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"height\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"txCount\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"prevBlock\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"txRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"headerRoot\",\"type\":\"bytes32\"}],\"internalType\":\"structBClaimsParserLibrary.BClaims\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch_\",\"type\":\"uint256\"}],\"name\":\"getBlockClaimsFromSnapshot\",\"outputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"chainId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"height\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"txCount\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"prevBlock\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"txRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"headerRoot\",\"type\":\"bytes32\"}],\"internalType\":\"structBClaimsParserLibrary.BClaims\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getChainId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getChainIdFromLatestSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch_\",\"type\":\"uint256\"}],\"name\":\"getChainIdFromSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCommittedHeightFromLatestSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch_\",\"type\":\"uint256\"}],\"name\":\"getCommittedHeightFromSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"height\",\"type\":\"uint256\"}],\"name\":\"getEpochFromHeight\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEpochLength\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLatestSnapshot\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"committedAt\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint32\",\"name\":\"chainId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"height\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"txCount\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"prevBlock\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"txRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"headerRoot\",\"type\":\"bytes32\"}],\"internalType\":\"structBClaimsParserLibrary.BClaims\",\"name\":\"blockClaims\",\"type\":\"tuple\"}],\"internalType\":\"structSnapshot\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMinimumIntervalBetweenSnapshots\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch_\",\"type\":\"uint256\"}],\"name\":\"getSnapshot\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"committedAt\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint32\",\"name\":\"chainId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"height\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"txCount\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"prevBlock\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"txRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"headerRoot\",\"type\":\"bytes32\"}],\"internalType\":\"structBClaimsParserLibrary.BClaims\",\"name\":\"blockClaims\",\"type\":\"tuple\"}],\"internalType\":\"structSnapshot\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSnapshotDesperationDelay\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSnapshotDesperationFactor\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"desperationDelay_\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"desperationFactor_\",\"type\":\"uint32\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"lastSnapshotCommittedAt\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"groupSignatureHash\",\"type\":\"bytes32\"}],\"name\":\"isValidatorElectedToPerformSnapshot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"numValidators\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"myIdx\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"blocksSinceDesperation\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"randomSeed\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"desperationFactor\",\"type\":\"uint256\"}],\"name\":\"mayValidatorSnapshot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes[]\",\"name\":\"groupSignature_\",\"type\":\"bytes[]\"},{\"internalType\":\"bytes[]\",\"name\":\"bClaims_\",\"type\":\"bytes[]\"}],\"name\":\"migrateSnapshots\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"minimumIntervalBetweenSnapshots_\",\"type\":\"uint32\"}],\"name\":\"setMinimumIntervalBetweenSnapshots\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"desperationDelay_\",\"type\":\"uint32\"}],\"name\":\"setSnapshotDesperationDelay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"desperationFactor_\",\"type\":\"uint32\"}],\"name\":\"setSnapshotDesperationFactor\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"groupSignature_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"bClaims_\",\"type\":\"bytes\"}],\"name\":\"snapshot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// SnapshotsABI is the input ABI used to generate the binding from.
// Deprecated: Use SnapshotsMetaData.ABI instead.
var SnapshotsABI = SnapshotsMetaData.ABI

// Snapshots is an auto generated Go binding around an Ethereum contract.
type Snapshots struct {
	SnapshotsCaller     // Read-only binding to the contract
	SnapshotsTransactor // Write-only binding to the contract
	SnapshotsFilterer   // Log filterer for contract events
}

// SnapshotsCaller is an auto generated read-only Go binding around an Ethereum contract.
type SnapshotsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SnapshotsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SnapshotsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SnapshotsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SnapshotsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SnapshotsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SnapshotsSession struct {
	Contract     *Snapshots        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SnapshotsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SnapshotsCallerSession struct {
	Contract *SnapshotsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// SnapshotsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SnapshotsTransactorSession struct {
	Contract     *SnapshotsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// SnapshotsRaw is an auto generated low-level Go binding around an Ethereum contract.
type SnapshotsRaw struct {
	Contract *Snapshots // Generic contract binding to access the raw methods on
}

// SnapshotsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SnapshotsCallerRaw struct {
	Contract *SnapshotsCaller // Generic read-only contract binding to access the raw methods on
}

// SnapshotsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SnapshotsTransactorRaw struct {
	Contract *SnapshotsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSnapshots creates a new instance of Snapshots, bound to a specific deployed contract.
func NewSnapshots(address common.Address, backend bind.ContractBackend) (*Snapshots, error) {
	contract, err := bindSnapshots(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Snapshots{SnapshotsCaller: SnapshotsCaller{contract: contract}, SnapshotsTransactor: SnapshotsTransactor{contract: contract}, SnapshotsFilterer: SnapshotsFilterer{contract: contract}}, nil
}

// NewSnapshotsCaller creates a new read-only instance of Snapshots, bound to a specific deployed contract.
func NewSnapshotsCaller(address common.Address, caller bind.ContractCaller) (*SnapshotsCaller, error) {
	contract, err := bindSnapshots(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SnapshotsCaller{contract: contract}, nil
}

// NewSnapshotsTransactor creates a new write-only instance of Snapshots, bound to a specific deployed contract.
func NewSnapshotsTransactor(address common.Address, transactor bind.ContractTransactor) (*SnapshotsTransactor, error) {
	contract, err := bindSnapshots(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SnapshotsTransactor{contract: contract}, nil
}

// NewSnapshotsFilterer creates a new log filterer instance of Snapshots, bound to a specific deployed contract.
func NewSnapshotsFilterer(address common.Address, filterer bind.ContractFilterer) (*SnapshotsFilterer, error) {
	contract, err := bindSnapshots(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SnapshotsFilterer{contract: contract}, nil
}

// bindSnapshots binds a generic wrapper to an already deployed contract.
func bindSnapshots(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SnapshotsABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Snapshots *SnapshotsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Snapshots.Contract.SnapshotsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Snapshots *SnapshotsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Snapshots.Contract.SnapshotsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Snapshots *SnapshotsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Snapshots.Contract.SnapshotsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Snapshots *SnapshotsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Snapshots.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Snapshots *SnapshotsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Snapshots.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Snapshots *SnapshotsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Snapshots.Contract.contract.Transact(opts, method, params...)
}

// CheckBClaimsSignature is a free data retrieval call binding the contract method 0x0204dffd.
//
// Solidity: function checkBClaimsSignature(bytes groupSignature_, bytes bClaims_) view returns(bool)
func (_Snapshots *SnapshotsCaller) CheckBClaimsSignature(opts *bind.CallOpts, groupSignature_ []byte, bClaims_ []byte) (bool, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "checkBClaimsSignature", groupSignature_, bClaims_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CheckBClaimsSignature is a free data retrieval call binding the contract method 0x0204dffd.
//
// Solidity: function checkBClaimsSignature(bytes groupSignature_, bytes bClaims_) view returns(bool)
func (_Snapshots *SnapshotsSession) CheckBClaimsSignature(groupSignature_ []byte, bClaims_ []byte) (bool, error) {
	return _Snapshots.Contract.CheckBClaimsSignature(&_Snapshots.CallOpts, groupSignature_, bClaims_)
}

// CheckBClaimsSignature is a free data retrieval call binding the contract method 0x0204dffd.
//
// Solidity: function checkBClaimsSignature(bytes groupSignature_, bytes bClaims_) view returns(bool)
func (_Snapshots *SnapshotsCallerSession) CheckBClaimsSignature(groupSignature_ []byte, bClaims_ []byte) (bool, error) {
	return _Snapshots.Contract.CheckBClaimsSignature(&_Snapshots.CallOpts, groupSignature_, bClaims_)
}

// GetAliceNetHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0xff07fc0e.
//
// Solidity: function getAliceNetHeightFromLatestSnapshot() view returns(uint256)
func (_Snapshots *SnapshotsCaller) GetAliceNetHeightFromLatestSnapshot(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getAliceNetHeightFromLatestSnapshot")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAliceNetHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0xff07fc0e.
//
// Solidity: function getAliceNetHeightFromLatestSnapshot() view returns(uint256)
func (_Snapshots *SnapshotsSession) GetAliceNetHeightFromLatestSnapshot() (*big.Int, error) {
	return _Snapshots.Contract.GetAliceNetHeightFromLatestSnapshot(&_Snapshots.CallOpts)
}

// GetAliceNetHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0xff07fc0e.
//
// Solidity: function getAliceNetHeightFromLatestSnapshot() view returns(uint256)
func (_Snapshots *SnapshotsCallerSession) GetAliceNetHeightFromLatestSnapshot() (*big.Int, error) {
	return _Snapshots.Contract.GetAliceNetHeightFromLatestSnapshot(&_Snapshots.CallOpts)
}

// GetAliceNetHeightFromSnapshot is a free data retrieval call binding the contract method 0xc5e8fde1.
//
// Solidity: function getAliceNetHeightFromSnapshot(uint256 epoch_) view returns(uint256)
func (_Snapshots *SnapshotsCaller) GetAliceNetHeightFromSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getAliceNetHeightFromSnapshot", epoch_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAliceNetHeightFromSnapshot is a free data retrieval call binding the contract method 0xc5e8fde1.
//
// Solidity: function getAliceNetHeightFromSnapshot(uint256 epoch_) view returns(uint256)
func (_Snapshots *SnapshotsSession) GetAliceNetHeightFromSnapshot(epoch_ *big.Int) (*big.Int, error) {
	return _Snapshots.Contract.GetAliceNetHeightFromSnapshot(&_Snapshots.CallOpts, epoch_)
}

// GetAliceNetHeightFromSnapshot is a free data retrieval call binding the contract method 0xc5e8fde1.
//
// Solidity: function getAliceNetHeightFromSnapshot(uint256 epoch_) view returns(uint256)
func (_Snapshots *SnapshotsCallerSession) GetAliceNetHeightFromSnapshot(epoch_ *big.Int) (*big.Int, error) {
	return _Snapshots.Contract.GetAliceNetHeightFromSnapshot(&_Snapshots.CallOpts, epoch_)
}

// GetBlockClaimsFromLatestSnapshot is a free data retrieval call binding the contract method 0xc2ea6603.
//
// Solidity: function getBlockClaimsFromLatestSnapshot() view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
func (_Snapshots *SnapshotsCaller) GetBlockClaimsFromLatestSnapshot(opts *bind.CallOpts) (BClaimsParserLibraryBClaims, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getBlockClaimsFromLatestSnapshot")

	if err != nil {
		return *new(BClaimsParserLibraryBClaims), err
	}

	out0 := *abi.ConvertType(out[0], new(BClaimsParserLibraryBClaims)).(*BClaimsParserLibraryBClaims)

	return out0, err

}

// GetBlockClaimsFromLatestSnapshot is a free data retrieval call binding the contract method 0xc2ea6603.
//
// Solidity: function getBlockClaimsFromLatestSnapshot() view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
func (_Snapshots *SnapshotsSession) GetBlockClaimsFromLatestSnapshot() (BClaimsParserLibraryBClaims, error) {
	return _Snapshots.Contract.GetBlockClaimsFromLatestSnapshot(&_Snapshots.CallOpts)
}

// GetBlockClaimsFromLatestSnapshot is a free data retrieval call binding the contract method 0xc2ea6603.
//
// Solidity: function getBlockClaimsFromLatestSnapshot() view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
func (_Snapshots *SnapshotsCallerSession) GetBlockClaimsFromLatestSnapshot() (BClaimsParserLibraryBClaims, error) {
	return _Snapshots.Contract.GetBlockClaimsFromLatestSnapshot(&_Snapshots.CallOpts)
}

// GetBlockClaimsFromSnapshot is a free data retrieval call binding the contract method 0x45dfc599.
//
// Solidity: function getBlockClaimsFromSnapshot(uint256 epoch_) view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
func (_Snapshots *SnapshotsCaller) GetBlockClaimsFromSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (BClaimsParserLibraryBClaims, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getBlockClaimsFromSnapshot", epoch_)

	if err != nil {
		return *new(BClaimsParserLibraryBClaims), err
	}

	out0 := *abi.ConvertType(out[0], new(BClaimsParserLibraryBClaims)).(*BClaimsParserLibraryBClaims)

	return out0, err

}

// GetBlockClaimsFromSnapshot is a free data retrieval call binding the contract method 0x45dfc599.
//
// Solidity: function getBlockClaimsFromSnapshot(uint256 epoch_) view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
func (_Snapshots *SnapshotsSession) GetBlockClaimsFromSnapshot(epoch_ *big.Int) (BClaimsParserLibraryBClaims, error) {
	return _Snapshots.Contract.GetBlockClaimsFromSnapshot(&_Snapshots.CallOpts, epoch_)
}

// GetBlockClaimsFromSnapshot is a free data retrieval call binding the contract method 0x45dfc599.
//
// Solidity: function getBlockClaimsFromSnapshot(uint256 epoch_) view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
func (_Snapshots *SnapshotsCallerSession) GetBlockClaimsFromSnapshot(epoch_ *big.Int) (BClaimsParserLibraryBClaims, error) {
	return _Snapshots.Contract.GetBlockClaimsFromSnapshot(&_Snapshots.CallOpts, epoch_)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() view returns(uint256)
func (_Snapshots *SnapshotsCaller) GetChainId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getChainId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() view returns(uint256)
func (_Snapshots *SnapshotsSession) GetChainId() (*big.Int, error) {
	return _Snapshots.Contract.GetChainId(&_Snapshots.CallOpts)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() view returns(uint256)
func (_Snapshots *SnapshotsCallerSession) GetChainId() (*big.Int, error) {
	return _Snapshots.Contract.GetChainId(&_Snapshots.CallOpts)
}

// GetChainIdFromLatestSnapshot is a free data retrieval call binding the contract method 0xd9c11657.
//
// Solidity: function getChainIdFromLatestSnapshot() view returns(uint256)
func (_Snapshots *SnapshotsCaller) GetChainIdFromLatestSnapshot(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getChainIdFromLatestSnapshot")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChainIdFromLatestSnapshot is a free data retrieval call binding the contract method 0xd9c11657.
//
// Solidity: function getChainIdFromLatestSnapshot() view returns(uint256)
func (_Snapshots *SnapshotsSession) GetChainIdFromLatestSnapshot() (*big.Int, error) {
	return _Snapshots.Contract.GetChainIdFromLatestSnapshot(&_Snapshots.CallOpts)
}

// GetChainIdFromLatestSnapshot is a free data retrieval call binding the contract method 0xd9c11657.
//
// Solidity: function getChainIdFromLatestSnapshot() view returns(uint256)
func (_Snapshots *SnapshotsCallerSession) GetChainIdFromLatestSnapshot() (*big.Int, error) {
	return _Snapshots.Contract.GetChainIdFromLatestSnapshot(&_Snapshots.CallOpts)
}

// GetChainIdFromSnapshot is a free data retrieval call binding the contract method 0x19f74669.
//
// Solidity: function getChainIdFromSnapshot(uint256 epoch_) view returns(uint256)
func (_Snapshots *SnapshotsCaller) GetChainIdFromSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getChainIdFromSnapshot", epoch_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChainIdFromSnapshot is a free data retrieval call binding the contract method 0x19f74669.
//
// Solidity: function getChainIdFromSnapshot(uint256 epoch_) view returns(uint256)
func (_Snapshots *SnapshotsSession) GetChainIdFromSnapshot(epoch_ *big.Int) (*big.Int, error) {
	return _Snapshots.Contract.GetChainIdFromSnapshot(&_Snapshots.CallOpts, epoch_)
}

// GetChainIdFromSnapshot is a free data retrieval call binding the contract method 0x19f74669.
//
// Solidity: function getChainIdFromSnapshot(uint256 epoch_) view returns(uint256)
func (_Snapshots *SnapshotsCallerSession) GetChainIdFromSnapshot(epoch_ *big.Int) (*big.Int, error) {
	return _Snapshots.Contract.GetChainIdFromSnapshot(&_Snapshots.CallOpts, epoch_)
}

// GetCommittedHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0x026c2b7e.
//
// Solidity: function getCommittedHeightFromLatestSnapshot() view returns(uint256)
func (_Snapshots *SnapshotsCaller) GetCommittedHeightFromLatestSnapshot(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getCommittedHeightFromLatestSnapshot")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCommittedHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0x026c2b7e.
//
// Solidity: function getCommittedHeightFromLatestSnapshot() view returns(uint256)
func (_Snapshots *SnapshotsSession) GetCommittedHeightFromLatestSnapshot() (*big.Int, error) {
	return _Snapshots.Contract.GetCommittedHeightFromLatestSnapshot(&_Snapshots.CallOpts)
}

// GetCommittedHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0x026c2b7e.
//
// Solidity: function getCommittedHeightFromLatestSnapshot() view returns(uint256)
func (_Snapshots *SnapshotsCallerSession) GetCommittedHeightFromLatestSnapshot() (*big.Int, error) {
	return _Snapshots.Contract.GetCommittedHeightFromLatestSnapshot(&_Snapshots.CallOpts)
}

// GetCommittedHeightFromSnapshot is a free data retrieval call binding the contract method 0xe18c697a.
//
// Solidity: function getCommittedHeightFromSnapshot(uint256 epoch_) view returns(uint256)
func (_Snapshots *SnapshotsCaller) GetCommittedHeightFromSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getCommittedHeightFromSnapshot", epoch_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCommittedHeightFromSnapshot is a free data retrieval call binding the contract method 0xe18c697a.
//
// Solidity: function getCommittedHeightFromSnapshot(uint256 epoch_) view returns(uint256)
func (_Snapshots *SnapshotsSession) GetCommittedHeightFromSnapshot(epoch_ *big.Int) (*big.Int, error) {
	return _Snapshots.Contract.GetCommittedHeightFromSnapshot(&_Snapshots.CallOpts, epoch_)
}

// GetCommittedHeightFromSnapshot is a free data retrieval call binding the contract method 0xe18c697a.
//
// Solidity: function getCommittedHeightFromSnapshot(uint256 epoch_) view returns(uint256)
func (_Snapshots *SnapshotsCallerSession) GetCommittedHeightFromSnapshot(epoch_ *big.Int) (*big.Int, error) {
	return _Snapshots.Contract.GetCommittedHeightFromSnapshot(&_Snapshots.CallOpts, epoch_)
}

// GetEpoch is a free data retrieval call binding the contract method 0x757991a8.
//
// Solidity: function getEpoch() view returns(uint256)
func (_Snapshots *SnapshotsCaller) GetEpoch(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getEpoch")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetEpoch is a free data retrieval call binding the contract method 0x757991a8.
//
// Solidity: function getEpoch() view returns(uint256)
func (_Snapshots *SnapshotsSession) GetEpoch() (*big.Int, error) {
	return _Snapshots.Contract.GetEpoch(&_Snapshots.CallOpts)
}

// GetEpoch is a free data retrieval call binding the contract method 0x757991a8.
//
// Solidity: function getEpoch() view returns(uint256)
func (_Snapshots *SnapshotsCallerSession) GetEpoch() (*big.Int, error) {
	return _Snapshots.Contract.GetEpoch(&_Snapshots.CallOpts)
}

// GetEpochFromHeight is a free data retrieval call binding the contract method 0x2eee30ce.
//
// Solidity: function getEpochFromHeight(uint256 height) view returns(uint256)
func (_Snapshots *SnapshotsCaller) GetEpochFromHeight(opts *bind.CallOpts, height *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getEpochFromHeight", height)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetEpochFromHeight is a free data retrieval call binding the contract method 0x2eee30ce.
//
// Solidity: function getEpochFromHeight(uint256 height) view returns(uint256)
func (_Snapshots *SnapshotsSession) GetEpochFromHeight(height *big.Int) (*big.Int, error) {
	return _Snapshots.Contract.GetEpochFromHeight(&_Snapshots.CallOpts, height)
}

// GetEpochFromHeight is a free data retrieval call binding the contract method 0x2eee30ce.
//
// Solidity: function getEpochFromHeight(uint256 height) view returns(uint256)
func (_Snapshots *SnapshotsCallerSession) GetEpochFromHeight(height *big.Int) (*big.Int, error) {
	return _Snapshots.Contract.GetEpochFromHeight(&_Snapshots.CallOpts, height)
}

// GetEpochLength is a free data retrieval call binding the contract method 0xcfe8a73b.
//
// Solidity: function getEpochLength() view returns(uint256)
func (_Snapshots *SnapshotsCaller) GetEpochLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getEpochLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetEpochLength is a free data retrieval call binding the contract method 0xcfe8a73b.
//
// Solidity: function getEpochLength() view returns(uint256)
func (_Snapshots *SnapshotsSession) GetEpochLength() (*big.Int, error) {
	return _Snapshots.Contract.GetEpochLength(&_Snapshots.CallOpts)
}

// GetEpochLength is a free data retrieval call binding the contract method 0xcfe8a73b.
//
// Solidity: function getEpochLength() view returns(uint256)
func (_Snapshots *SnapshotsCallerSession) GetEpochLength() (*big.Int, error) {
	return _Snapshots.Contract.GetEpochLength(&_Snapshots.CallOpts)
}

// GetLatestSnapshot is a free data retrieval call binding the contract method 0xd518f243.
//
// Solidity: function getLatestSnapshot() view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
func (_Snapshots *SnapshotsCaller) GetLatestSnapshot(opts *bind.CallOpts) (Snapshot, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getLatestSnapshot")

	if err != nil {
		return *new(Snapshot), err
	}

	out0 := *abi.ConvertType(out[0], new(Snapshot)).(*Snapshot)

	return out0, err

}

// GetLatestSnapshot is a free data retrieval call binding the contract method 0xd518f243.
//
// Solidity: function getLatestSnapshot() view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
func (_Snapshots *SnapshotsSession) GetLatestSnapshot() (Snapshot, error) {
	return _Snapshots.Contract.GetLatestSnapshot(&_Snapshots.CallOpts)
}

// GetLatestSnapshot is a free data retrieval call binding the contract method 0xd518f243.
//
// Solidity: function getLatestSnapshot() view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
func (_Snapshots *SnapshotsCallerSession) GetLatestSnapshot() (Snapshot, error) {
	return _Snapshots.Contract.GetLatestSnapshot(&_Snapshots.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_Snapshots *SnapshotsCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_Snapshots *SnapshotsSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _Snapshots.Contract.GetMetamorphicContractAddress(&_Snapshots.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_Snapshots *SnapshotsCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _Snapshots.Contract.GetMetamorphicContractAddress(&_Snapshots.CallOpts, _salt, _factory)
}

// GetMinimumIntervalBetweenSnapshots is a free data retrieval call binding the contract method 0x42438d7b.
//
// Solidity: function getMinimumIntervalBetweenSnapshots() view returns(uint256)
func (_Snapshots *SnapshotsCaller) GetMinimumIntervalBetweenSnapshots(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getMinimumIntervalBetweenSnapshots")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMinimumIntervalBetweenSnapshots is a free data retrieval call binding the contract method 0x42438d7b.
//
// Solidity: function getMinimumIntervalBetweenSnapshots() view returns(uint256)
func (_Snapshots *SnapshotsSession) GetMinimumIntervalBetweenSnapshots() (*big.Int, error) {
	return _Snapshots.Contract.GetMinimumIntervalBetweenSnapshots(&_Snapshots.CallOpts)
}

// GetMinimumIntervalBetweenSnapshots is a free data retrieval call binding the contract method 0x42438d7b.
//
// Solidity: function getMinimumIntervalBetweenSnapshots() view returns(uint256)
func (_Snapshots *SnapshotsCallerSession) GetMinimumIntervalBetweenSnapshots() (*big.Int, error) {
	return _Snapshots.Contract.GetMinimumIntervalBetweenSnapshots(&_Snapshots.CallOpts)
}

// GetSnapshot is a free data retrieval call binding the contract method 0x76f10ad0.
//
// Solidity: function getSnapshot(uint256 epoch_) view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
func (_Snapshots *SnapshotsCaller) GetSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (Snapshot, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getSnapshot", epoch_)

	if err != nil {
		return *new(Snapshot), err
	}

	out0 := *abi.ConvertType(out[0], new(Snapshot)).(*Snapshot)

	return out0, err

}

// GetSnapshot is a free data retrieval call binding the contract method 0x76f10ad0.
//
// Solidity: function getSnapshot(uint256 epoch_) view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
func (_Snapshots *SnapshotsSession) GetSnapshot(epoch_ *big.Int) (Snapshot, error) {
	return _Snapshots.Contract.GetSnapshot(&_Snapshots.CallOpts, epoch_)
}

// GetSnapshot is a free data retrieval call binding the contract method 0x76f10ad0.
//
// Solidity: function getSnapshot(uint256 epoch_) view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
func (_Snapshots *SnapshotsCallerSession) GetSnapshot(epoch_ *big.Int) (Snapshot, error) {
	return _Snapshots.Contract.GetSnapshot(&_Snapshots.CallOpts, epoch_)
}

// GetSnapshotDesperationDelay is a free data retrieval call binding the contract method 0xd17fcc56.
//
// Solidity: function getSnapshotDesperationDelay() view returns(uint256)
func (_Snapshots *SnapshotsCaller) GetSnapshotDesperationDelay(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getSnapshotDesperationDelay")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSnapshotDesperationDelay is a free data retrieval call binding the contract method 0xd17fcc56.
//
// Solidity: function getSnapshotDesperationDelay() view returns(uint256)
func (_Snapshots *SnapshotsSession) GetSnapshotDesperationDelay() (*big.Int, error) {
	return _Snapshots.Contract.GetSnapshotDesperationDelay(&_Snapshots.CallOpts)
}

// GetSnapshotDesperationDelay is a free data retrieval call binding the contract method 0xd17fcc56.
//
// Solidity: function getSnapshotDesperationDelay() view returns(uint256)
func (_Snapshots *SnapshotsCallerSession) GetSnapshotDesperationDelay() (*big.Int, error) {
	return _Snapshots.Contract.GetSnapshotDesperationDelay(&_Snapshots.CallOpts)
}

// GetSnapshotDesperationFactor is a free data retrieval call binding the contract method 0x7cc4cce6.
//
// Solidity: function getSnapshotDesperationFactor() view returns(uint256)
func (_Snapshots *SnapshotsCaller) GetSnapshotDesperationFactor(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "getSnapshotDesperationFactor")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSnapshotDesperationFactor is a free data retrieval call binding the contract method 0x7cc4cce6.
//
// Solidity: function getSnapshotDesperationFactor() view returns(uint256)
func (_Snapshots *SnapshotsSession) GetSnapshotDesperationFactor() (*big.Int, error) {
	return _Snapshots.Contract.GetSnapshotDesperationFactor(&_Snapshots.CallOpts)
}

// GetSnapshotDesperationFactor is a free data retrieval call binding the contract method 0x7cc4cce6.
//
// Solidity: function getSnapshotDesperationFactor() view returns(uint256)
func (_Snapshots *SnapshotsCallerSession) GetSnapshotDesperationFactor() (*big.Int, error) {
	return _Snapshots.Contract.GetSnapshotDesperationFactor(&_Snapshots.CallOpts)
}

// IsValidatorElectedToPerformSnapshot is a free data retrieval call binding the contract method 0xc0e83e81.
//
// Solidity: function isValidatorElectedToPerformSnapshot(address validator, uint256 lastSnapshotCommittedAt, bytes32 groupSignatureHash) view returns(bool)
func (_Snapshots *SnapshotsCaller) IsValidatorElectedToPerformSnapshot(opts *bind.CallOpts, validator common.Address, lastSnapshotCommittedAt *big.Int, groupSignatureHash [32]byte) (bool, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "isValidatorElectedToPerformSnapshot", validator, lastSnapshotCommittedAt, groupSignatureHash)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidatorElectedToPerformSnapshot is a free data retrieval call binding the contract method 0xc0e83e81.
//
// Solidity: function isValidatorElectedToPerformSnapshot(address validator, uint256 lastSnapshotCommittedAt, bytes32 groupSignatureHash) view returns(bool)
func (_Snapshots *SnapshotsSession) IsValidatorElectedToPerformSnapshot(validator common.Address, lastSnapshotCommittedAt *big.Int, groupSignatureHash [32]byte) (bool, error) {
	return _Snapshots.Contract.IsValidatorElectedToPerformSnapshot(&_Snapshots.CallOpts, validator, lastSnapshotCommittedAt, groupSignatureHash)
}

// IsValidatorElectedToPerformSnapshot is a free data retrieval call binding the contract method 0xc0e83e81.
//
// Solidity: function isValidatorElectedToPerformSnapshot(address validator, uint256 lastSnapshotCommittedAt, bytes32 groupSignatureHash) view returns(bool)
func (_Snapshots *SnapshotsCallerSession) IsValidatorElectedToPerformSnapshot(validator common.Address, lastSnapshotCommittedAt *big.Int, groupSignatureHash [32]byte) (bool, error) {
	return _Snapshots.Contract.IsValidatorElectedToPerformSnapshot(&_Snapshots.CallOpts, validator, lastSnapshotCommittedAt, groupSignatureHash)
}

// MayValidatorSnapshot is a free data retrieval call binding the contract method 0xf45fa246.
//
// Solidity: function mayValidatorSnapshot(uint256 numValidators, uint256 myIdx, uint256 blocksSinceDesperation, bytes32 randomSeed, uint256 desperationFactor) pure returns(bool)
func (_Snapshots *SnapshotsCaller) MayValidatorSnapshot(opts *bind.CallOpts, numValidators *big.Int, myIdx *big.Int, blocksSinceDesperation *big.Int, randomSeed [32]byte, desperationFactor *big.Int) (bool, error) {
	var out []interface{}
	err := _Snapshots.contract.Call(opts, &out, "mayValidatorSnapshot", numValidators, myIdx, blocksSinceDesperation, randomSeed, desperationFactor)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// MayValidatorSnapshot is a free data retrieval call binding the contract method 0xf45fa246.
//
// Solidity: function mayValidatorSnapshot(uint256 numValidators, uint256 myIdx, uint256 blocksSinceDesperation, bytes32 randomSeed, uint256 desperationFactor) pure returns(bool)
func (_Snapshots *SnapshotsSession) MayValidatorSnapshot(numValidators *big.Int, myIdx *big.Int, blocksSinceDesperation *big.Int, randomSeed [32]byte, desperationFactor *big.Int) (bool, error) {
	return _Snapshots.Contract.MayValidatorSnapshot(&_Snapshots.CallOpts, numValidators, myIdx, blocksSinceDesperation, randomSeed, desperationFactor)
}

// MayValidatorSnapshot is a free data retrieval call binding the contract method 0xf45fa246.
//
// Solidity: function mayValidatorSnapshot(uint256 numValidators, uint256 myIdx, uint256 blocksSinceDesperation, bytes32 randomSeed, uint256 desperationFactor) pure returns(bool)
func (_Snapshots *SnapshotsCallerSession) MayValidatorSnapshot(numValidators *big.Int, myIdx *big.Int, blocksSinceDesperation *big.Int, randomSeed [32]byte, desperationFactor *big.Int) (bool, error) {
	return _Snapshots.Contract.MayValidatorSnapshot(&_Snapshots.CallOpts, numValidators, myIdx, blocksSinceDesperation, randomSeed, desperationFactor)
}

// Initialize is a paid mutator transaction binding the contract method 0x3ecc0f5e.
//
// Solidity: function initialize(uint32 desperationDelay_, uint32 desperationFactor_) returns()
func (_Snapshots *SnapshotsTransactor) Initialize(opts *bind.TransactOpts, desperationDelay_ uint32, desperationFactor_ uint32) (*types.Transaction, error) {
	return _Snapshots.contract.Transact(opts, "initialize", desperationDelay_, desperationFactor_)
}

// Initialize is a paid mutator transaction binding the contract method 0x3ecc0f5e.
//
// Solidity: function initialize(uint32 desperationDelay_, uint32 desperationFactor_) returns()
func (_Snapshots *SnapshotsSession) Initialize(desperationDelay_ uint32, desperationFactor_ uint32) (*types.Transaction, error) {
	return _Snapshots.Contract.Initialize(&_Snapshots.TransactOpts, desperationDelay_, desperationFactor_)
}

// Initialize is a paid mutator transaction binding the contract method 0x3ecc0f5e.
//
// Solidity: function initialize(uint32 desperationDelay_, uint32 desperationFactor_) returns()
func (_Snapshots *SnapshotsTransactorSession) Initialize(desperationDelay_ uint32, desperationFactor_ uint32) (*types.Transaction, error) {
	return _Snapshots.Contract.Initialize(&_Snapshots.TransactOpts, desperationDelay_, desperationFactor_)
}

// MigrateSnapshots is a paid mutator transaction binding the contract method 0xae2728ea.
//
// Solidity: function migrateSnapshots(bytes[] groupSignature_, bytes[] bClaims_) returns(bool)
func (_Snapshots *SnapshotsTransactor) MigrateSnapshots(opts *bind.TransactOpts, groupSignature_ [][]byte, bClaims_ [][]byte) (*types.Transaction, error) {
	return _Snapshots.contract.Transact(opts, "migrateSnapshots", groupSignature_, bClaims_)
}

// MigrateSnapshots is a paid mutator transaction binding the contract method 0xae2728ea.
//
// Solidity: function migrateSnapshots(bytes[] groupSignature_, bytes[] bClaims_) returns(bool)
func (_Snapshots *SnapshotsSession) MigrateSnapshots(groupSignature_ [][]byte, bClaims_ [][]byte) (*types.Transaction, error) {
	return _Snapshots.Contract.MigrateSnapshots(&_Snapshots.TransactOpts, groupSignature_, bClaims_)
}

// MigrateSnapshots is a paid mutator transaction binding the contract method 0xae2728ea.
//
// Solidity: function migrateSnapshots(bytes[] groupSignature_, bytes[] bClaims_) returns(bool)
func (_Snapshots *SnapshotsTransactorSession) MigrateSnapshots(groupSignature_ [][]byte, bClaims_ [][]byte) (*types.Transaction, error) {
	return _Snapshots.Contract.MigrateSnapshots(&_Snapshots.TransactOpts, groupSignature_, bClaims_)
}

// SetMinimumIntervalBetweenSnapshots is a paid mutator transaction binding the contract method 0xeb7c7afe.
//
// Solidity: function setMinimumIntervalBetweenSnapshots(uint32 minimumIntervalBetweenSnapshots_) returns()
func (_Snapshots *SnapshotsTransactor) SetMinimumIntervalBetweenSnapshots(opts *bind.TransactOpts, minimumIntervalBetweenSnapshots_ uint32) (*types.Transaction, error) {
	return _Snapshots.contract.Transact(opts, "setMinimumIntervalBetweenSnapshots", minimumIntervalBetweenSnapshots_)
}

// SetMinimumIntervalBetweenSnapshots is a paid mutator transaction binding the contract method 0xeb7c7afe.
//
// Solidity: function setMinimumIntervalBetweenSnapshots(uint32 minimumIntervalBetweenSnapshots_) returns()
func (_Snapshots *SnapshotsSession) SetMinimumIntervalBetweenSnapshots(minimumIntervalBetweenSnapshots_ uint32) (*types.Transaction, error) {
	return _Snapshots.Contract.SetMinimumIntervalBetweenSnapshots(&_Snapshots.TransactOpts, minimumIntervalBetweenSnapshots_)
}

// SetMinimumIntervalBetweenSnapshots is a paid mutator transaction binding the contract method 0xeb7c7afe.
//
// Solidity: function setMinimumIntervalBetweenSnapshots(uint32 minimumIntervalBetweenSnapshots_) returns()
func (_Snapshots *SnapshotsTransactorSession) SetMinimumIntervalBetweenSnapshots(minimumIntervalBetweenSnapshots_ uint32) (*types.Transaction, error) {
	return _Snapshots.Contract.SetMinimumIntervalBetweenSnapshots(&_Snapshots.TransactOpts, minimumIntervalBetweenSnapshots_)
}

// SetSnapshotDesperationDelay is a paid mutator transaction binding the contract method 0xc2e8fef2.
//
// Solidity: function setSnapshotDesperationDelay(uint32 desperationDelay_) returns()
func (_Snapshots *SnapshotsTransactor) SetSnapshotDesperationDelay(opts *bind.TransactOpts, desperationDelay_ uint32) (*types.Transaction, error) {
	return _Snapshots.contract.Transact(opts, "setSnapshotDesperationDelay", desperationDelay_)
}

// SetSnapshotDesperationDelay is a paid mutator transaction binding the contract method 0xc2e8fef2.
//
// Solidity: function setSnapshotDesperationDelay(uint32 desperationDelay_) returns()
func (_Snapshots *SnapshotsSession) SetSnapshotDesperationDelay(desperationDelay_ uint32) (*types.Transaction, error) {
	return _Snapshots.Contract.SetSnapshotDesperationDelay(&_Snapshots.TransactOpts, desperationDelay_)
}

// SetSnapshotDesperationDelay is a paid mutator transaction binding the contract method 0xc2e8fef2.
//
// Solidity: function setSnapshotDesperationDelay(uint32 desperationDelay_) returns()
func (_Snapshots *SnapshotsTransactorSession) SetSnapshotDesperationDelay(desperationDelay_ uint32) (*types.Transaction, error) {
	return _Snapshots.Contract.SetSnapshotDesperationDelay(&_Snapshots.TransactOpts, desperationDelay_)
}

// SetSnapshotDesperationFactor is a paid mutator transaction binding the contract method 0x3fa7a1ad.
//
// Solidity: function setSnapshotDesperationFactor(uint32 desperationFactor_) returns()
func (_Snapshots *SnapshotsTransactor) SetSnapshotDesperationFactor(opts *bind.TransactOpts, desperationFactor_ uint32) (*types.Transaction, error) {
	return _Snapshots.contract.Transact(opts, "setSnapshotDesperationFactor", desperationFactor_)
}

// SetSnapshotDesperationFactor is a paid mutator transaction binding the contract method 0x3fa7a1ad.
//
// Solidity: function setSnapshotDesperationFactor(uint32 desperationFactor_) returns()
func (_Snapshots *SnapshotsSession) SetSnapshotDesperationFactor(desperationFactor_ uint32) (*types.Transaction, error) {
	return _Snapshots.Contract.SetSnapshotDesperationFactor(&_Snapshots.TransactOpts, desperationFactor_)
}

// SetSnapshotDesperationFactor is a paid mutator transaction binding the contract method 0x3fa7a1ad.
//
// Solidity: function setSnapshotDesperationFactor(uint32 desperationFactor_) returns()
func (_Snapshots *SnapshotsTransactorSession) SetSnapshotDesperationFactor(desperationFactor_ uint32) (*types.Transaction, error) {
	return _Snapshots.Contract.SetSnapshotDesperationFactor(&_Snapshots.TransactOpts, desperationFactor_)
}

// Snapshot is a paid mutator transaction binding the contract method 0x08ca1f25.
//
// Solidity: function snapshot(bytes groupSignature_, bytes bClaims_) returns(bool)
func (_Snapshots *SnapshotsTransactor) Snapshot(opts *bind.TransactOpts, groupSignature_ []byte, bClaims_ []byte) (*types.Transaction, error) {
	return _Snapshots.contract.Transact(opts, "snapshot", groupSignature_, bClaims_)
}

// Snapshot is a paid mutator transaction binding the contract method 0x08ca1f25.
//
// Solidity: function snapshot(bytes groupSignature_, bytes bClaims_) returns(bool)
func (_Snapshots *SnapshotsSession) Snapshot(groupSignature_ []byte, bClaims_ []byte) (*types.Transaction, error) {
	return _Snapshots.Contract.Snapshot(&_Snapshots.TransactOpts, groupSignature_, bClaims_)
}

// Snapshot is a paid mutator transaction binding the contract method 0x08ca1f25.
//
// Solidity: function snapshot(bytes groupSignature_, bytes bClaims_) returns(bool)
func (_Snapshots *SnapshotsTransactorSession) Snapshot(groupSignature_ []byte, bClaims_ []byte) (*types.Transaction, error) {
	return _Snapshots.Contract.Snapshot(&_Snapshots.TransactOpts, groupSignature_, bClaims_)
}

// SnapshotsInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Snapshots contract.
type SnapshotsInitializedIterator struct {
	Event *SnapshotsInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SnapshotsInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SnapshotsInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SnapshotsInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SnapshotsInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SnapshotsInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SnapshotsInitialized represents a Initialized event raised by the Snapshots contract.
type SnapshotsInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Snapshots *SnapshotsFilterer) FilterInitialized(opts *bind.FilterOpts) (*SnapshotsInitializedIterator, error) {

	logs, sub, err := _Snapshots.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &SnapshotsInitializedIterator{contract: _Snapshots.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Snapshots *SnapshotsFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *SnapshotsInitialized) (event.Subscription, error) {

	logs, sub, err := _Snapshots.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SnapshotsInitialized)
				if err := _Snapshots.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Snapshots *SnapshotsFilterer) ParseInitialized(log types.Log) (*SnapshotsInitialized, error) {
	event := new(SnapshotsInitialized)
	if err := _Snapshots.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SnapshotsSnapshotTakenIterator is returned from FilterSnapshotTaken and is used to iterate over the raw logs and unpacked data for SnapshotTaken events raised by the Snapshots contract.
type SnapshotsSnapshotTakenIterator struct {
	Event *SnapshotsSnapshotTaken // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SnapshotsSnapshotTakenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SnapshotsSnapshotTaken)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SnapshotsSnapshotTaken)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SnapshotsSnapshotTakenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SnapshotsSnapshotTakenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SnapshotsSnapshotTaken represents a SnapshotTaken event raised by the Snapshots contract.
type SnapshotsSnapshotTaken struct {
	ChainId                  *big.Int
	Epoch                    *big.Int
	Height                   *big.Int
	Validator                common.Address
	IsSafeToProceedConsensus bool
	MasterPublicKey          [4]*big.Int
	Signature                [2]*big.Int
	BClaims                  BClaimsParserLibraryBClaims
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterSnapshotTaken is a free log retrieval operation binding the contract event 0x709e2f13a448a6bdef51a1fcadf45303dee0b6bff5bdf628829f401b019b7e9b.
//
// Solidity: event SnapshotTaken(uint256 chainId, uint256 indexed epoch, uint256 height, address indexed validator, bool isSafeToProceedConsensus, uint256[4] masterPublicKey, uint256[2] signature, (uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32) bClaims)
func (_Snapshots *SnapshotsFilterer) FilterSnapshotTaken(opts *bind.FilterOpts, epoch []*big.Int, validator []common.Address) (*SnapshotsSnapshotTakenIterator, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _Snapshots.contract.FilterLogs(opts, "SnapshotTaken", epochRule, validatorRule)
	if err != nil {
		return nil, err
	}
	return &SnapshotsSnapshotTakenIterator{contract: _Snapshots.contract, event: "SnapshotTaken", logs: logs, sub: sub}, nil
}

// WatchSnapshotTaken is a free log subscription operation binding the contract event 0x709e2f13a448a6bdef51a1fcadf45303dee0b6bff5bdf628829f401b019b7e9b.
//
// Solidity: event SnapshotTaken(uint256 chainId, uint256 indexed epoch, uint256 height, address indexed validator, bool isSafeToProceedConsensus, uint256[4] masterPublicKey, uint256[2] signature, (uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32) bClaims)
func (_Snapshots *SnapshotsFilterer) WatchSnapshotTaken(opts *bind.WatchOpts, sink chan<- *SnapshotsSnapshotTaken, epoch []*big.Int, validator []common.Address) (event.Subscription, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _Snapshots.contract.WatchLogs(opts, "SnapshotTaken", epochRule, validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SnapshotsSnapshotTaken)
				if err := _Snapshots.contract.UnpackLog(event, "SnapshotTaken", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSnapshotTaken is a log parse operation binding the contract event 0x709e2f13a448a6bdef51a1fcadf45303dee0b6bff5bdf628829f401b019b7e9b.
//
// Solidity: event SnapshotTaken(uint256 chainId, uint256 indexed epoch, uint256 height, address indexed validator, bool isSafeToProceedConsensus, uint256[4] masterPublicKey, uint256[2] signature, (uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32) bClaims)
func (_Snapshots *SnapshotsFilterer) ParseSnapshotTaken(log types.Log) (*SnapshotsSnapshotTaken, error) {
	event := new(SnapshotsSnapshotTaken)
	if err := _Snapshots.contract.UnpackLog(event, "SnapshotTaken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
