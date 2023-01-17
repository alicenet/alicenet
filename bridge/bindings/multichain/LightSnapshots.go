// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package multichain

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

// LightSnapshotsMetaData contains all meta data concerning the LightSnapshots contract.
var LightSnapshotsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"chainID_\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"epochLength_\",\"type\":\"uint32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"Bytes32OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"Bytes32OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ChainIdZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"dataOffset\",\"type\":\"uint256\"}],\"name\":\"DataOffsetOverflow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"HeightZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"LEUint16OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"LEUint16OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"dataOffset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"NotEnoughBytes\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyFactory\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyValidatorPool\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"dataSectionSize\",\"type\":\"uint16\"}],\"name\":\"SizeThresholdExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"onlyAdminAllowed\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"height\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"isSafeToProceedConsensus\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"uint256[4]\",\"name\":\"masterPublicKey\",\"type\":\"uint256[4]\"},{\"indexed\":false,\"internalType\":\"uint256[2]\",\"name\":\"signature\",\"type\":\"uint256[2]\"},{\"components\":[{\"internalType\":\"uint32\",\"name\":\"chainId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"height\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"txCount\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"prevBlock\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"txRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"headerRoot\",\"type\":\"bytes32\"}],\"indexed\":false,\"internalType\":\"structBClaimsParserLibrary.BClaims\",\"name\":\"bClaims\",\"type\":\"tuple\"}],\"name\":\"SnapshotTaken\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"groupSignature_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"bClaims_\",\"type\":\"bytes\"}],\"name\":\"checkBClaimsSignature\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAliceNetHeightFromLatestSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch_\",\"type\":\"uint256\"}],\"name\":\"getAliceNetHeightFromSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getBlockClaimsFromLatestSnapshot\",\"outputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"chainId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"height\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"txCount\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"prevBlock\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"txRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"headerRoot\",\"type\":\"bytes32\"}],\"internalType\":\"structBClaimsParserLibrary.BClaims\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch_\",\"type\":\"uint256\"}],\"name\":\"getBlockClaimsFromSnapshot\",\"outputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"chainId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"height\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"txCount\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"prevBlock\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"txRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"headerRoot\",\"type\":\"bytes32\"}],\"internalType\":\"structBClaimsParserLibrary.BClaims\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getChainId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getChainIdFromLatestSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch_\",\"type\":\"uint256\"}],\"name\":\"getChainIdFromSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCommittedHeightFromLatestSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch_\",\"type\":\"uint256\"}],\"name\":\"getCommittedHeightFromSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"height\",\"type\":\"uint256\"}],\"name\":\"getEpochFromHeight\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEpochLength\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLatestSnapshot\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"committedAt\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint32\",\"name\":\"chainId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"height\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"txCount\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"prevBlock\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"txRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"headerRoot\",\"type\":\"bytes32\"}],\"internalType\":\"structBClaimsParserLibrary.BClaims\",\"name\":\"blockClaims\",\"type\":\"tuple\"}],\"internalType\":\"structSnapshot\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMinimumIntervalBetweenSnapshots\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch_\",\"type\":\"uint256\"}],\"name\":\"getSnapshot\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"committedAt\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint32\",\"name\":\"chainId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"height\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"txCount\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"prevBlock\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"txRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"headerRoot\",\"type\":\"bytes32\"}],\"internalType\":\"structBClaimsParserLibrary.BClaims\",\"name\":\"blockClaims\",\"type\":\"tuple\"}],\"internalType\":\"structSnapshot\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSnapshotDesperationDelay\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSnapshotDesperationFactor\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"desperationDelay_\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"desperationFactor_\",\"type\":\"uint32\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isMock\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"lastSnapshotCommittedAt\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"groupSignatureHash\",\"type\":\"bytes32\"}],\"name\":\"isValidatorElectedToPerformSnapshot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"numValidators\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"myIdx\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"blocksSinceDesperation\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"blsig\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"desperationFactor\",\"type\":\"uint256\"}],\"name\":\"mayValidatorSnapshot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes[]\",\"name\":\"groupSignature_\",\"type\":\"bytes[]\"},{\"internalType\":\"bytes[]\",\"name\":\"bClaims_\",\"type\":\"bytes[]\"}],\"name\":\"migrateSnapshots\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"height_\",\"type\":\"uint256\"}],\"name\":\"setCommittedHeightFromLatestSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"epochLength_\",\"type\":\"uint32\"}],\"name\":\"setEpochLength\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"minimumIntervalBetweenSnapshots_\",\"type\":\"uint32\"}],\"name\":\"setMinimumIntervalBetweenSnapshots\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"desperationDelay_\",\"type\":\"uint32\"}],\"name\":\"setSnapshotDesperationDelay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"desperationFactor_\",\"type\":\"uint32\"}],\"name\":\"setSnapshotDesperationFactor\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"groupSignature_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"bClaims_\",\"type\":\"bytes\"}],\"name\":\"snapshot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"groupSignature_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"bClaims_\",\"type\":\"bytes\"}],\"name\":\"snapshotWithValidData\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// LightSnapshotsABI is the input ABI used to generate the binding from.
// Deprecated: Use LightSnapshotsMetaData.ABI instead.
var LightSnapshotsABI = LightSnapshotsMetaData.ABI

// LightSnapshots is an auto generated Go binding around an Ethereum contract.
type LightSnapshots struct {
	LightSnapshotsCaller     // Read-only binding to the contract
	LightSnapshotsTransactor // Write-only binding to the contract
	LightSnapshotsFilterer   // Log filterer for contract events
}

// LightSnapshotsCaller is an auto generated read-only Go binding around an Ethereum contract.
type LightSnapshotsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LightSnapshotsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type LightSnapshotsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LightSnapshotsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type LightSnapshotsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LightSnapshotsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type LightSnapshotsSession struct {
	Contract     *LightSnapshots   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// LightSnapshotsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type LightSnapshotsCallerSession struct {
	Contract *LightSnapshotsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// LightSnapshotsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type LightSnapshotsTransactorSession struct {
	Contract     *LightSnapshotsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// LightSnapshotsRaw is an auto generated low-level Go binding around an Ethereum contract.
type LightSnapshotsRaw struct {
	Contract *LightSnapshots // Generic contract binding to access the raw methods on
}

// LightSnapshotsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type LightSnapshotsCallerRaw struct {
	Contract *LightSnapshotsCaller // Generic read-only contract binding to access the raw methods on
}

// LightSnapshotsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type LightSnapshotsTransactorRaw struct {
	Contract *LightSnapshotsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewLightSnapshots creates a new instance of LightSnapshots, bound to a specific deployed contract.
func NewLightSnapshots(address common.Address, backend bind.ContractBackend) (*LightSnapshots, error) {
	contract, err := bindLightSnapshots(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &LightSnapshots{LightSnapshotsCaller: LightSnapshotsCaller{contract: contract}, LightSnapshotsTransactor: LightSnapshotsTransactor{contract: contract}, LightSnapshotsFilterer: LightSnapshotsFilterer{contract: contract}}, nil
}

// NewLightSnapshotsCaller creates a new read-only instance of LightSnapshots, bound to a specific deployed contract.
func NewLightSnapshotsCaller(address common.Address, caller bind.ContractCaller) (*LightSnapshotsCaller, error) {
	contract, err := bindLightSnapshots(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &LightSnapshotsCaller{contract: contract}, nil
}

// NewLightSnapshotsTransactor creates a new write-only instance of LightSnapshots, bound to a specific deployed contract.
func NewLightSnapshotsTransactor(address common.Address, transactor bind.ContractTransactor) (*LightSnapshotsTransactor, error) {
	contract, err := bindLightSnapshots(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &LightSnapshotsTransactor{contract: contract}, nil
}

// NewLightSnapshotsFilterer creates a new log filterer instance of LightSnapshots, bound to a specific deployed contract.
func NewLightSnapshotsFilterer(address common.Address, filterer bind.ContractFilterer) (*LightSnapshotsFilterer, error) {
	contract, err := bindLightSnapshots(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &LightSnapshotsFilterer{contract: contract}, nil
}

// bindLightSnapshots binds a generic wrapper to an already deployed contract.
func bindLightSnapshots(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(LightSnapshotsABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LightSnapshots *LightSnapshotsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _LightSnapshots.Contract.LightSnapshotsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LightSnapshots *LightSnapshotsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LightSnapshots.Contract.LightSnapshotsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LightSnapshots *LightSnapshotsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LightSnapshots.Contract.LightSnapshotsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LightSnapshots *LightSnapshotsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _LightSnapshots.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LightSnapshots *LightSnapshotsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LightSnapshots.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LightSnapshots *LightSnapshotsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LightSnapshots.Contract.contract.Transact(opts, method, params...)
}

// CheckBClaimsSignature is a free data retrieval call binding the contract method 0x0204dffd.
//
// Solidity: function checkBClaimsSignature(bytes groupSignature_, bytes bClaims_) pure returns(bool)
func (_LightSnapshots *LightSnapshotsCaller) CheckBClaimsSignature(opts *bind.CallOpts, groupSignature_ []byte, bClaims_ []byte) (bool, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "checkBClaimsSignature", groupSignature_, bClaims_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CheckBClaimsSignature is a free data retrieval call binding the contract method 0x0204dffd.
//
// Solidity: function checkBClaimsSignature(bytes groupSignature_, bytes bClaims_) pure returns(bool)
func (_LightSnapshots *LightSnapshotsSession) CheckBClaimsSignature(groupSignature_ []byte, bClaims_ []byte) (bool, error) {
	return _LightSnapshots.Contract.CheckBClaimsSignature(&_LightSnapshots.CallOpts, groupSignature_, bClaims_)
}

// CheckBClaimsSignature is a free data retrieval call binding the contract method 0x0204dffd.
//
// Solidity: function checkBClaimsSignature(bytes groupSignature_, bytes bClaims_) pure returns(bool)
func (_LightSnapshots *LightSnapshotsCallerSession) CheckBClaimsSignature(groupSignature_ []byte, bClaims_ []byte) (bool, error) {
	return _LightSnapshots.Contract.CheckBClaimsSignature(&_LightSnapshots.CallOpts, groupSignature_, bClaims_)
}

// GetAliceNetHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0xff07fc0e.
//
// Solidity: function getAliceNetHeightFromLatestSnapshot() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCaller) GetAliceNetHeightFromLatestSnapshot(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getAliceNetHeightFromLatestSnapshot")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAliceNetHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0xff07fc0e.
//
// Solidity: function getAliceNetHeightFromLatestSnapshot() view returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) GetAliceNetHeightFromLatestSnapshot() (*big.Int, error) {
	return _LightSnapshots.Contract.GetAliceNetHeightFromLatestSnapshot(&_LightSnapshots.CallOpts)
}

// GetAliceNetHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0xff07fc0e.
//
// Solidity: function getAliceNetHeightFromLatestSnapshot() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCallerSession) GetAliceNetHeightFromLatestSnapshot() (*big.Int, error) {
	return _LightSnapshots.Contract.GetAliceNetHeightFromLatestSnapshot(&_LightSnapshots.CallOpts)
}

// GetAliceNetHeightFromSnapshot is a free data retrieval call binding the contract method 0xc5e8fde1.
//
// Solidity: function getAliceNetHeightFromSnapshot(uint256 epoch_) view returns(uint256)
func (_LightSnapshots *LightSnapshotsCaller) GetAliceNetHeightFromSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getAliceNetHeightFromSnapshot", epoch_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAliceNetHeightFromSnapshot is a free data retrieval call binding the contract method 0xc5e8fde1.
//
// Solidity: function getAliceNetHeightFromSnapshot(uint256 epoch_) view returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) GetAliceNetHeightFromSnapshot(epoch_ *big.Int) (*big.Int, error) {
	return _LightSnapshots.Contract.GetAliceNetHeightFromSnapshot(&_LightSnapshots.CallOpts, epoch_)
}

// GetAliceNetHeightFromSnapshot is a free data retrieval call binding the contract method 0xc5e8fde1.
//
// Solidity: function getAliceNetHeightFromSnapshot(uint256 epoch_) view returns(uint256)
func (_LightSnapshots *LightSnapshotsCallerSession) GetAliceNetHeightFromSnapshot(epoch_ *big.Int) (*big.Int, error) {
	return _LightSnapshots.Contract.GetAliceNetHeightFromSnapshot(&_LightSnapshots.CallOpts, epoch_)
}

// GetBlockClaimsFromLatestSnapshot is a free data retrieval call binding the contract method 0xc2ea6603.
//
// Solidity: function getBlockClaimsFromLatestSnapshot() view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
func (_LightSnapshots *LightSnapshotsCaller) GetBlockClaimsFromLatestSnapshot(opts *bind.CallOpts) (BClaimsParserLibraryBClaims, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getBlockClaimsFromLatestSnapshot")

	if err != nil {
		return *new(BClaimsParserLibraryBClaims), err
	}

	out0 := *abi.ConvertType(out[0], new(BClaimsParserLibraryBClaims)).(*BClaimsParserLibraryBClaims)

	return out0, err

}

// GetBlockClaimsFromLatestSnapshot is a free data retrieval call binding the contract method 0xc2ea6603.
//
// Solidity: function getBlockClaimsFromLatestSnapshot() view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
func (_LightSnapshots *LightSnapshotsSession) GetBlockClaimsFromLatestSnapshot() (BClaimsParserLibraryBClaims, error) {
	return _LightSnapshots.Contract.GetBlockClaimsFromLatestSnapshot(&_LightSnapshots.CallOpts)
}

// GetBlockClaimsFromLatestSnapshot is a free data retrieval call binding the contract method 0xc2ea6603.
//
// Solidity: function getBlockClaimsFromLatestSnapshot() view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
func (_LightSnapshots *LightSnapshotsCallerSession) GetBlockClaimsFromLatestSnapshot() (BClaimsParserLibraryBClaims, error) {
	return _LightSnapshots.Contract.GetBlockClaimsFromLatestSnapshot(&_LightSnapshots.CallOpts)
}

// GetBlockClaimsFromSnapshot is a free data retrieval call binding the contract method 0x45dfc599.
//
// Solidity: function getBlockClaimsFromSnapshot(uint256 epoch_) view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
func (_LightSnapshots *LightSnapshotsCaller) GetBlockClaimsFromSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (BClaimsParserLibraryBClaims, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getBlockClaimsFromSnapshot", epoch_)

	if err != nil {
		return *new(BClaimsParserLibraryBClaims), err
	}

	out0 := *abi.ConvertType(out[0], new(BClaimsParserLibraryBClaims)).(*BClaimsParserLibraryBClaims)

	return out0, err

}

// GetBlockClaimsFromSnapshot is a free data retrieval call binding the contract method 0x45dfc599.
//
// Solidity: function getBlockClaimsFromSnapshot(uint256 epoch_) view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
func (_LightSnapshots *LightSnapshotsSession) GetBlockClaimsFromSnapshot(epoch_ *big.Int) (BClaimsParserLibraryBClaims, error) {
	return _LightSnapshots.Contract.GetBlockClaimsFromSnapshot(&_LightSnapshots.CallOpts, epoch_)
}

// GetBlockClaimsFromSnapshot is a free data retrieval call binding the contract method 0x45dfc599.
//
// Solidity: function getBlockClaimsFromSnapshot(uint256 epoch_) view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
func (_LightSnapshots *LightSnapshotsCallerSession) GetBlockClaimsFromSnapshot(epoch_ *big.Int) (BClaimsParserLibraryBClaims, error) {
	return _LightSnapshots.Contract.GetBlockClaimsFromSnapshot(&_LightSnapshots.CallOpts, epoch_)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCaller) GetChainId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getChainId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() view returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) GetChainId() (*big.Int, error) {
	return _LightSnapshots.Contract.GetChainId(&_LightSnapshots.CallOpts)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCallerSession) GetChainId() (*big.Int, error) {
	return _LightSnapshots.Contract.GetChainId(&_LightSnapshots.CallOpts)
}

// GetChainIdFromLatestSnapshot is a free data retrieval call binding the contract method 0xd9c11657.
//
// Solidity: function getChainIdFromLatestSnapshot() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCaller) GetChainIdFromLatestSnapshot(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getChainIdFromLatestSnapshot")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChainIdFromLatestSnapshot is a free data retrieval call binding the contract method 0xd9c11657.
//
// Solidity: function getChainIdFromLatestSnapshot() view returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) GetChainIdFromLatestSnapshot() (*big.Int, error) {
	return _LightSnapshots.Contract.GetChainIdFromLatestSnapshot(&_LightSnapshots.CallOpts)
}

// GetChainIdFromLatestSnapshot is a free data retrieval call binding the contract method 0xd9c11657.
//
// Solidity: function getChainIdFromLatestSnapshot() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCallerSession) GetChainIdFromLatestSnapshot() (*big.Int, error) {
	return _LightSnapshots.Contract.GetChainIdFromLatestSnapshot(&_LightSnapshots.CallOpts)
}

// GetChainIdFromSnapshot is a free data retrieval call binding the contract method 0x19f74669.
//
// Solidity: function getChainIdFromSnapshot(uint256 epoch_) view returns(uint256)
func (_LightSnapshots *LightSnapshotsCaller) GetChainIdFromSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getChainIdFromSnapshot", epoch_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChainIdFromSnapshot is a free data retrieval call binding the contract method 0x19f74669.
//
// Solidity: function getChainIdFromSnapshot(uint256 epoch_) view returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) GetChainIdFromSnapshot(epoch_ *big.Int) (*big.Int, error) {
	return _LightSnapshots.Contract.GetChainIdFromSnapshot(&_LightSnapshots.CallOpts, epoch_)
}

// GetChainIdFromSnapshot is a free data retrieval call binding the contract method 0x19f74669.
//
// Solidity: function getChainIdFromSnapshot(uint256 epoch_) view returns(uint256)
func (_LightSnapshots *LightSnapshotsCallerSession) GetChainIdFromSnapshot(epoch_ *big.Int) (*big.Int, error) {
	return _LightSnapshots.Contract.GetChainIdFromSnapshot(&_LightSnapshots.CallOpts, epoch_)
}

// GetCommittedHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0x026c2b7e.
//
// Solidity: function getCommittedHeightFromLatestSnapshot() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCaller) GetCommittedHeightFromLatestSnapshot(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getCommittedHeightFromLatestSnapshot")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCommittedHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0x026c2b7e.
//
// Solidity: function getCommittedHeightFromLatestSnapshot() view returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) GetCommittedHeightFromLatestSnapshot() (*big.Int, error) {
	return _LightSnapshots.Contract.GetCommittedHeightFromLatestSnapshot(&_LightSnapshots.CallOpts)
}

// GetCommittedHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0x026c2b7e.
//
// Solidity: function getCommittedHeightFromLatestSnapshot() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCallerSession) GetCommittedHeightFromLatestSnapshot() (*big.Int, error) {
	return _LightSnapshots.Contract.GetCommittedHeightFromLatestSnapshot(&_LightSnapshots.CallOpts)
}

// GetCommittedHeightFromSnapshot is a free data retrieval call binding the contract method 0xe18c697a.
//
// Solidity: function getCommittedHeightFromSnapshot(uint256 epoch_) view returns(uint256)
func (_LightSnapshots *LightSnapshotsCaller) GetCommittedHeightFromSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getCommittedHeightFromSnapshot", epoch_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCommittedHeightFromSnapshot is a free data retrieval call binding the contract method 0xe18c697a.
//
// Solidity: function getCommittedHeightFromSnapshot(uint256 epoch_) view returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) GetCommittedHeightFromSnapshot(epoch_ *big.Int) (*big.Int, error) {
	return _LightSnapshots.Contract.GetCommittedHeightFromSnapshot(&_LightSnapshots.CallOpts, epoch_)
}

// GetCommittedHeightFromSnapshot is a free data retrieval call binding the contract method 0xe18c697a.
//
// Solidity: function getCommittedHeightFromSnapshot(uint256 epoch_) view returns(uint256)
func (_LightSnapshots *LightSnapshotsCallerSession) GetCommittedHeightFromSnapshot(epoch_ *big.Int) (*big.Int, error) {
	return _LightSnapshots.Contract.GetCommittedHeightFromSnapshot(&_LightSnapshots.CallOpts, epoch_)
}

// GetEpoch is a free data retrieval call binding the contract method 0x757991a8.
//
// Solidity: function getEpoch() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCaller) GetEpoch(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getEpoch")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetEpoch is a free data retrieval call binding the contract method 0x757991a8.
//
// Solidity: function getEpoch() view returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) GetEpoch() (*big.Int, error) {
	return _LightSnapshots.Contract.GetEpoch(&_LightSnapshots.CallOpts)
}

// GetEpoch is a free data retrieval call binding the contract method 0x757991a8.
//
// Solidity: function getEpoch() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCallerSession) GetEpoch() (*big.Int, error) {
	return _LightSnapshots.Contract.GetEpoch(&_LightSnapshots.CallOpts)
}

// GetEpochFromHeight is a free data retrieval call binding the contract method 0x2eee30ce.
//
// Solidity: function getEpochFromHeight(uint256 height) view returns(uint256)
func (_LightSnapshots *LightSnapshotsCaller) GetEpochFromHeight(opts *bind.CallOpts, height *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getEpochFromHeight", height)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetEpochFromHeight is a free data retrieval call binding the contract method 0x2eee30ce.
//
// Solidity: function getEpochFromHeight(uint256 height) view returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) GetEpochFromHeight(height *big.Int) (*big.Int, error) {
	return _LightSnapshots.Contract.GetEpochFromHeight(&_LightSnapshots.CallOpts, height)
}

// GetEpochFromHeight is a free data retrieval call binding the contract method 0x2eee30ce.
//
// Solidity: function getEpochFromHeight(uint256 height) view returns(uint256)
func (_LightSnapshots *LightSnapshotsCallerSession) GetEpochFromHeight(height *big.Int) (*big.Int, error) {
	return _LightSnapshots.Contract.GetEpochFromHeight(&_LightSnapshots.CallOpts, height)
}

// GetEpochLength is a free data retrieval call binding the contract method 0xcfe8a73b.
//
// Solidity: function getEpochLength() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCaller) GetEpochLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getEpochLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetEpochLength is a free data retrieval call binding the contract method 0xcfe8a73b.
//
// Solidity: function getEpochLength() view returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) GetEpochLength() (*big.Int, error) {
	return _LightSnapshots.Contract.GetEpochLength(&_LightSnapshots.CallOpts)
}

// GetEpochLength is a free data retrieval call binding the contract method 0xcfe8a73b.
//
// Solidity: function getEpochLength() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCallerSession) GetEpochLength() (*big.Int, error) {
	return _LightSnapshots.Contract.GetEpochLength(&_LightSnapshots.CallOpts)
}

// GetLatestSnapshot is a free data retrieval call binding the contract method 0xd518f243.
//
// Solidity: function getLatestSnapshot() view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
func (_LightSnapshots *LightSnapshotsCaller) GetLatestSnapshot(opts *bind.CallOpts) (Snapshot, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getLatestSnapshot")

	if err != nil {
		return *new(Snapshot), err
	}

	out0 := *abi.ConvertType(out[0], new(Snapshot)).(*Snapshot)

	return out0, err

}

// GetLatestSnapshot is a free data retrieval call binding the contract method 0xd518f243.
//
// Solidity: function getLatestSnapshot() view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
func (_LightSnapshots *LightSnapshotsSession) GetLatestSnapshot() (Snapshot, error) {
	return _LightSnapshots.Contract.GetLatestSnapshot(&_LightSnapshots.CallOpts)
}

// GetLatestSnapshot is a free data retrieval call binding the contract method 0xd518f243.
//
// Solidity: function getLatestSnapshot() view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
func (_LightSnapshots *LightSnapshotsCallerSession) GetLatestSnapshot() (Snapshot, error) {
	return _LightSnapshots.Contract.GetLatestSnapshot(&_LightSnapshots.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_LightSnapshots *LightSnapshotsCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_LightSnapshots *LightSnapshotsSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _LightSnapshots.Contract.GetMetamorphicContractAddress(&_LightSnapshots.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_LightSnapshots *LightSnapshotsCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _LightSnapshots.Contract.GetMetamorphicContractAddress(&_LightSnapshots.CallOpts, _salt, _factory)
}

// GetMinimumIntervalBetweenSnapshots is a free data retrieval call binding the contract method 0x42438d7b.
//
// Solidity: function getMinimumIntervalBetweenSnapshots() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCaller) GetMinimumIntervalBetweenSnapshots(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getMinimumIntervalBetweenSnapshots")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMinimumIntervalBetweenSnapshots is a free data retrieval call binding the contract method 0x42438d7b.
//
// Solidity: function getMinimumIntervalBetweenSnapshots() view returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) GetMinimumIntervalBetweenSnapshots() (*big.Int, error) {
	return _LightSnapshots.Contract.GetMinimumIntervalBetweenSnapshots(&_LightSnapshots.CallOpts)
}

// GetMinimumIntervalBetweenSnapshots is a free data retrieval call binding the contract method 0x42438d7b.
//
// Solidity: function getMinimumIntervalBetweenSnapshots() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCallerSession) GetMinimumIntervalBetweenSnapshots() (*big.Int, error) {
	return _LightSnapshots.Contract.GetMinimumIntervalBetweenSnapshots(&_LightSnapshots.CallOpts)
}

// GetSnapshot is a free data retrieval call binding the contract method 0x76f10ad0.
//
// Solidity: function getSnapshot(uint256 epoch_) view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
func (_LightSnapshots *LightSnapshotsCaller) GetSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (Snapshot, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getSnapshot", epoch_)

	if err != nil {
		return *new(Snapshot), err
	}

	out0 := *abi.ConvertType(out[0], new(Snapshot)).(*Snapshot)

	return out0, err

}

// GetSnapshot is a free data retrieval call binding the contract method 0x76f10ad0.
//
// Solidity: function getSnapshot(uint256 epoch_) view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
func (_LightSnapshots *LightSnapshotsSession) GetSnapshot(epoch_ *big.Int) (Snapshot, error) {
	return _LightSnapshots.Contract.GetSnapshot(&_LightSnapshots.CallOpts, epoch_)
}

// GetSnapshot is a free data retrieval call binding the contract method 0x76f10ad0.
//
// Solidity: function getSnapshot(uint256 epoch_) view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
func (_LightSnapshots *LightSnapshotsCallerSession) GetSnapshot(epoch_ *big.Int) (Snapshot, error) {
	return _LightSnapshots.Contract.GetSnapshot(&_LightSnapshots.CallOpts, epoch_)
}

// GetSnapshotDesperationDelay is a free data retrieval call binding the contract method 0xd17fcc56.
//
// Solidity: function getSnapshotDesperationDelay() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCaller) GetSnapshotDesperationDelay(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getSnapshotDesperationDelay")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSnapshotDesperationDelay is a free data retrieval call binding the contract method 0xd17fcc56.
//
// Solidity: function getSnapshotDesperationDelay() view returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) GetSnapshotDesperationDelay() (*big.Int, error) {
	return _LightSnapshots.Contract.GetSnapshotDesperationDelay(&_LightSnapshots.CallOpts)
}

// GetSnapshotDesperationDelay is a free data retrieval call binding the contract method 0xd17fcc56.
//
// Solidity: function getSnapshotDesperationDelay() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCallerSession) GetSnapshotDesperationDelay() (*big.Int, error) {
	return _LightSnapshots.Contract.GetSnapshotDesperationDelay(&_LightSnapshots.CallOpts)
}

// GetSnapshotDesperationFactor is a free data retrieval call binding the contract method 0x7cc4cce6.
//
// Solidity: function getSnapshotDesperationFactor() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCaller) GetSnapshotDesperationFactor(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "getSnapshotDesperationFactor")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSnapshotDesperationFactor is a free data retrieval call binding the contract method 0x7cc4cce6.
//
// Solidity: function getSnapshotDesperationFactor() view returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) GetSnapshotDesperationFactor() (*big.Int, error) {
	return _LightSnapshots.Contract.GetSnapshotDesperationFactor(&_LightSnapshots.CallOpts)
}

// GetSnapshotDesperationFactor is a free data retrieval call binding the contract method 0x7cc4cce6.
//
// Solidity: function getSnapshotDesperationFactor() view returns(uint256)
func (_LightSnapshots *LightSnapshotsCallerSession) GetSnapshotDesperationFactor() (*big.Int, error) {
	return _LightSnapshots.Contract.GetSnapshotDesperationFactor(&_LightSnapshots.CallOpts)
}

// IsMock is a free data retrieval call binding the contract method 0x28ccaa29.
//
// Solidity: function isMock() pure returns(bool)
func (_LightSnapshots *LightSnapshotsCaller) IsMock(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "isMock")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsMock is a free data retrieval call binding the contract method 0x28ccaa29.
//
// Solidity: function isMock() pure returns(bool)
func (_LightSnapshots *LightSnapshotsSession) IsMock() (bool, error) {
	return _LightSnapshots.Contract.IsMock(&_LightSnapshots.CallOpts)
}

// IsMock is a free data retrieval call binding the contract method 0x28ccaa29.
//
// Solidity: function isMock() pure returns(bool)
func (_LightSnapshots *LightSnapshotsCallerSession) IsMock() (bool, error) {
	return _LightSnapshots.Contract.IsMock(&_LightSnapshots.CallOpts)
}

// IsValidatorElectedToPerformSnapshot is a free data retrieval call binding the contract method 0xc0e83e81.
//
// Solidity: function isValidatorElectedToPerformSnapshot(address validator, uint256 lastSnapshotCommittedAt, bytes32 groupSignatureHash) pure returns(bool)
func (_LightSnapshots *LightSnapshotsCaller) IsValidatorElectedToPerformSnapshot(opts *bind.CallOpts, validator common.Address, lastSnapshotCommittedAt *big.Int, groupSignatureHash [32]byte) (bool, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "isValidatorElectedToPerformSnapshot", validator, lastSnapshotCommittedAt, groupSignatureHash)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidatorElectedToPerformSnapshot is a free data retrieval call binding the contract method 0xc0e83e81.
//
// Solidity: function isValidatorElectedToPerformSnapshot(address validator, uint256 lastSnapshotCommittedAt, bytes32 groupSignatureHash) pure returns(bool)
func (_LightSnapshots *LightSnapshotsSession) IsValidatorElectedToPerformSnapshot(validator common.Address, lastSnapshotCommittedAt *big.Int, groupSignatureHash [32]byte) (bool, error) {
	return _LightSnapshots.Contract.IsValidatorElectedToPerformSnapshot(&_LightSnapshots.CallOpts, validator, lastSnapshotCommittedAt, groupSignatureHash)
}

// IsValidatorElectedToPerformSnapshot is a free data retrieval call binding the contract method 0xc0e83e81.
//
// Solidity: function isValidatorElectedToPerformSnapshot(address validator, uint256 lastSnapshotCommittedAt, bytes32 groupSignatureHash) pure returns(bool)
func (_LightSnapshots *LightSnapshotsCallerSession) IsValidatorElectedToPerformSnapshot(validator common.Address, lastSnapshotCommittedAt *big.Int, groupSignatureHash [32]byte) (bool, error) {
	return _LightSnapshots.Contract.IsValidatorElectedToPerformSnapshot(&_LightSnapshots.CallOpts, validator, lastSnapshotCommittedAt, groupSignatureHash)
}

// MayValidatorSnapshot is a free data retrieval call binding the contract method 0xf45fa246.
//
// Solidity: function mayValidatorSnapshot(uint256 numValidators, uint256 myIdx, uint256 blocksSinceDesperation, bytes32 blsig, uint256 desperationFactor) pure returns(bool)
func (_LightSnapshots *LightSnapshotsCaller) MayValidatorSnapshot(opts *bind.CallOpts, numValidators *big.Int, myIdx *big.Int, blocksSinceDesperation *big.Int, blsig [32]byte, desperationFactor *big.Int) (bool, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "mayValidatorSnapshot", numValidators, myIdx, blocksSinceDesperation, blsig, desperationFactor)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// MayValidatorSnapshot is a free data retrieval call binding the contract method 0xf45fa246.
//
// Solidity: function mayValidatorSnapshot(uint256 numValidators, uint256 myIdx, uint256 blocksSinceDesperation, bytes32 blsig, uint256 desperationFactor) pure returns(bool)
func (_LightSnapshots *LightSnapshotsSession) MayValidatorSnapshot(numValidators *big.Int, myIdx *big.Int, blocksSinceDesperation *big.Int, blsig [32]byte, desperationFactor *big.Int) (bool, error) {
	return _LightSnapshots.Contract.MayValidatorSnapshot(&_LightSnapshots.CallOpts, numValidators, myIdx, blocksSinceDesperation, blsig, desperationFactor)
}

// MayValidatorSnapshot is a free data retrieval call binding the contract method 0xf45fa246.
//
// Solidity: function mayValidatorSnapshot(uint256 numValidators, uint256 myIdx, uint256 blocksSinceDesperation, bytes32 blsig, uint256 desperationFactor) pure returns(bool)
func (_LightSnapshots *LightSnapshotsCallerSession) MayValidatorSnapshot(numValidators *big.Int, myIdx *big.Int, blocksSinceDesperation *big.Int, blsig [32]byte, desperationFactor *big.Int) (bool, error) {
	return _LightSnapshots.Contract.MayValidatorSnapshot(&_LightSnapshots.CallOpts, numValidators, myIdx, blocksSinceDesperation, blsig, desperationFactor)
}

// MigrateSnapshots is a free data retrieval call binding the contract method 0xae2728ea.
//
// Solidity: function migrateSnapshots(bytes[] groupSignature_, bytes[] bClaims_) pure returns(bool)
func (_LightSnapshots *LightSnapshotsCaller) MigrateSnapshots(opts *bind.CallOpts, groupSignature_ [][]byte, bClaims_ [][]byte) (bool, error) {
	var out []interface{}
	err := _LightSnapshots.contract.Call(opts, &out, "migrateSnapshots", groupSignature_, bClaims_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// MigrateSnapshots is a free data retrieval call binding the contract method 0xae2728ea.
//
// Solidity: function migrateSnapshots(bytes[] groupSignature_, bytes[] bClaims_) pure returns(bool)
func (_LightSnapshots *LightSnapshotsSession) MigrateSnapshots(groupSignature_ [][]byte, bClaims_ [][]byte) (bool, error) {
	return _LightSnapshots.Contract.MigrateSnapshots(&_LightSnapshots.CallOpts, groupSignature_, bClaims_)
}

// MigrateSnapshots is a free data retrieval call binding the contract method 0xae2728ea.
//
// Solidity: function migrateSnapshots(bytes[] groupSignature_, bytes[] bClaims_) pure returns(bool)
func (_LightSnapshots *LightSnapshotsCallerSession) MigrateSnapshots(groupSignature_ [][]byte, bClaims_ [][]byte) (bool, error) {
	return _LightSnapshots.Contract.MigrateSnapshots(&_LightSnapshots.CallOpts, groupSignature_, bClaims_)
}

// Initialize is a paid mutator transaction binding the contract method 0x3ecc0f5e.
//
// Solidity: function initialize(uint32 desperationDelay_, uint32 desperationFactor_) returns()
func (_LightSnapshots *LightSnapshotsTransactor) Initialize(opts *bind.TransactOpts, desperationDelay_ uint32, desperationFactor_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.contract.Transact(opts, "initialize", desperationDelay_, desperationFactor_)
}

// Initialize is a paid mutator transaction binding the contract method 0x3ecc0f5e.
//
// Solidity: function initialize(uint32 desperationDelay_, uint32 desperationFactor_) returns()
func (_LightSnapshots *LightSnapshotsSession) Initialize(desperationDelay_ uint32, desperationFactor_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.Contract.Initialize(&_LightSnapshots.TransactOpts, desperationDelay_, desperationFactor_)
}

// Initialize is a paid mutator transaction binding the contract method 0x3ecc0f5e.
//
// Solidity: function initialize(uint32 desperationDelay_, uint32 desperationFactor_) returns()
func (_LightSnapshots *LightSnapshotsTransactorSession) Initialize(desperationDelay_ uint32, desperationFactor_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.Contract.Initialize(&_LightSnapshots.TransactOpts, desperationDelay_, desperationFactor_)
}

// SetCommittedHeightFromLatestSnapshot is a paid mutator transaction binding the contract method 0xff914b1e.
//
// Solidity: function setCommittedHeightFromLatestSnapshot(uint256 height_) returns(uint256)
func (_LightSnapshots *LightSnapshotsTransactor) SetCommittedHeightFromLatestSnapshot(opts *bind.TransactOpts, height_ *big.Int) (*types.Transaction, error) {
	return _LightSnapshots.contract.Transact(opts, "setCommittedHeightFromLatestSnapshot", height_)
}

// SetCommittedHeightFromLatestSnapshot is a paid mutator transaction binding the contract method 0xff914b1e.
//
// Solidity: function setCommittedHeightFromLatestSnapshot(uint256 height_) returns(uint256)
func (_LightSnapshots *LightSnapshotsSession) SetCommittedHeightFromLatestSnapshot(height_ *big.Int) (*types.Transaction, error) {
	return _LightSnapshots.Contract.SetCommittedHeightFromLatestSnapshot(&_LightSnapshots.TransactOpts, height_)
}

// SetCommittedHeightFromLatestSnapshot is a paid mutator transaction binding the contract method 0xff914b1e.
//
// Solidity: function setCommittedHeightFromLatestSnapshot(uint256 height_) returns(uint256)
func (_LightSnapshots *LightSnapshotsTransactorSession) SetCommittedHeightFromLatestSnapshot(height_ *big.Int) (*types.Transaction, error) {
	return _LightSnapshots.Contract.SetCommittedHeightFromLatestSnapshot(&_LightSnapshots.TransactOpts, height_)
}

// SetEpochLength is a paid mutator transaction binding the contract method 0xdeb1e56e.
//
// Solidity: function setEpochLength(uint32 epochLength_) returns()
func (_LightSnapshots *LightSnapshotsTransactor) SetEpochLength(opts *bind.TransactOpts, epochLength_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.contract.Transact(opts, "setEpochLength", epochLength_)
}

// SetEpochLength is a paid mutator transaction binding the contract method 0xdeb1e56e.
//
// Solidity: function setEpochLength(uint32 epochLength_) returns()
func (_LightSnapshots *LightSnapshotsSession) SetEpochLength(epochLength_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.Contract.SetEpochLength(&_LightSnapshots.TransactOpts, epochLength_)
}

// SetEpochLength is a paid mutator transaction binding the contract method 0xdeb1e56e.
//
// Solidity: function setEpochLength(uint32 epochLength_) returns()
func (_LightSnapshots *LightSnapshotsTransactorSession) SetEpochLength(epochLength_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.Contract.SetEpochLength(&_LightSnapshots.TransactOpts, epochLength_)
}

// SetMinimumIntervalBetweenSnapshots is a paid mutator transaction binding the contract method 0xeb7c7afe.
//
// Solidity: function setMinimumIntervalBetweenSnapshots(uint32 minimumIntervalBetweenSnapshots_) returns()
func (_LightSnapshots *LightSnapshotsTransactor) SetMinimumIntervalBetweenSnapshots(opts *bind.TransactOpts, minimumIntervalBetweenSnapshots_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.contract.Transact(opts, "setMinimumIntervalBetweenSnapshots", minimumIntervalBetweenSnapshots_)
}

// SetMinimumIntervalBetweenSnapshots is a paid mutator transaction binding the contract method 0xeb7c7afe.
//
// Solidity: function setMinimumIntervalBetweenSnapshots(uint32 minimumIntervalBetweenSnapshots_) returns()
func (_LightSnapshots *LightSnapshotsSession) SetMinimumIntervalBetweenSnapshots(minimumIntervalBetweenSnapshots_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.Contract.SetMinimumIntervalBetweenSnapshots(&_LightSnapshots.TransactOpts, minimumIntervalBetweenSnapshots_)
}

// SetMinimumIntervalBetweenSnapshots is a paid mutator transaction binding the contract method 0xeb7c7afe.
//
// Solidity: function setMinimumIntervalBetweenSnapshots(uint32 minimumIntervalBetweenSnapshots_) returns()
func (_LightSnapshots *LightSnapshotsTransactorSession) SetMinimumIntervalBetweenSnapshots(minimumIntervalBetweenSnapshots_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.Contract.SetMinimumIntervalBetweenSnapshots(&_LightSnapshots.TransactOpts, minimumIntervalBetweenSnapshots_)
}

// SetSnapshotDesperationDelay is a paid mutator transaction binding the contract method 0xc2e8fef2.
//
// Solidity: function setSnapshotDesperationDelay(uint32 desperationDelay_) returns()
func (_LightSnapshots *LightSnapshotsTransactor) SetSnapshotDesperationDelay(opts *bind.TransactOpts, desperationDelay_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.contract.Transact(opts, "setSnapshotDesperationDelay", desperationDelay_)
}

// SetSnapshotDesperationDelay is a paid mutator transaction binding the contract method 0xc2e8fef2.
//
// Solidity: function setSnapshotDesperationDelay(uint32 desperationDelay_) returns()
func (_LightSnapshots *LightSnapshotsSession) SetSnapshotDesperationDelay(desperationDelay_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.Contract.SetSnapshotDesperationDelay(&_LightSnapshots.TransactOpts, desperationDelay_)
}

// SetSnapshotDesperationDelay is a paid mutator transaction binding the contract method 0xc2e8fef2.
//
// Solidity: function setSnapshotDesperationDelay(uint32 desperationDelay_) returns()
func (_LightSnapshots *LightSnapshotsTransactorSession) SetSnapshotDesperationDelay(desperationDelay_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.Contract.SetSnapshotDesperationDelay(&_LightSnapshots.TransactOpts, desperationDelay_)
}

// SetSnapshotDesperationFactor is a paid mutator transaction binding the contract method 0x3fa7a1ad.
//
// Solidity: function setSnapshotDesperationFactor(uint32 desperationFactor_) returns()
func (_LightSnapshots *LightSnapshotsTransactor) SetSnapshotDesperationFactor(opts *bind.TransactOpts, desperationFactor_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.contract.Transact(opts, "setSnapshotDesperationFactor", desperationFactor_)
}

// SetSnapshotDesperationFactor is a paid mutator transaction binding the contract method 0x3fa7a1ad.
//
// Solidity: function setSnapshotDesperationFactor(uint32 desperationFactor_) returns()
func (_LightSnapshots *LightSnapshotsSession) SetSnapshotDesperationFactor(desperationFactor_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.Contract.SetSnapshotDesperationFactor(&_LightSnapshots.TransactOpts, desperationFactor_)
}

// SetSnapshotDesperationFactor is a paid mutator transaction binding the contract method 0x3fa7a1ad.
//
// Solidity: function setSnapshotDesperationFactor(uint32 desperationFactor_) returns()
func (_LightSnapshots *LightSnapshotsTransactorSession) SetSnapshotDesperationFactor(desperationFactor_ uint32) (*types.Transaction, error) {
	return _LightSnapshots.Contract.SetSnapshotDesperationFactor(&_LightSnapshots.TransactOpts, desperationFactor_)
}

// Snapshot is a paid mutator transaction binding the contract method 0x08ca1f25.
//
// Solidity: function snapshot(bytes groupSignature_, bytes bClaims_) returns(bool)
func (_LightSnapshots *LightSnapshotsTransactor) Snapshot(opts *bind.TransactOpts, groupSignature_ []byte, bClaims_ []byte) (*types.Transaction, error) {
	return _LightSnapshots.contract.Transact(opts, "snapshot", groupSignature_, bClaims_)
}

// Snapshot is a paid mutator transaction binding the contract method 0x08ca1f25.
//
// Solidity: function snapshot(bytes groupSignature_, bytes bClaims_) returns(bool)
func (_LightSnapshots *LightSnapshotsSession) Snapshot(groupSignature_ []byte, bClaims_ []byte) (*types.Transaction, error) {
	return _LightSnapshots.Contract.Snapshot(&_LightSnapshots.TransactOpts, groupSignature_, bClaims_)
}

// Snapshot is a paid mutator transaction binding the contract method 0x08ca1f25.
//
// Solidity: function snapshot(bytes groupSignature_, bytes bClaims_) returns(bool)
func (_LightSnapshots *LightSnapshotsTransactorSession) Snapshot(groupSignature_ []byte, bClaims_ []byte) (*types.Transaction, error) {
	return _LightSnapshots.Contract.Snapshot(&_LightSnapshots.TransactOpts, groupSignature_, bClaims_)
}

// SnapshotWithValidData is a paid mutator transaction binding the contract method 0x20d8a106.
//
// Solidity: function snapshotWithValidData(bytes groupSignature_, bytes bClaims_) returns(bool)
func (_LightSnapshots *LightSnapshotsTransactor) SnapshotWithValidData(opts *bind.TransactOpts, groupSignature_ []byte, bClaims_ []byte) (*types.Transaction, error) {
	return _LightSnapshots.contract.Transact(opts, "snapshotWithValidData", groupSignature_, bClaims_)
}

// SnapshotWithValidData is a paid mutator transaction binding the contract method 0x20d8a106.
//
// Solidity: function snapshotWithValidData(bytes groupSignature_, bytes bClaims_) returns(bool)
func (_LightSnapshots *LightSnapshotsSession) SnapshotWithValidData(groupSignature_ []byte, bClaims_ []byte) (*types.Transaction, error) {
	return _LightSnapshots.Contract.SnapshotWithValidData(&_LightSnapshots.TransactOpts, groupSignature_, bClaims_)
}

// SnapshotWithValidData is a paid mutator transaction binding the contract method 0x20d8a106.
//
// Solidity: function snapshotWithValidData(bytes groupSignature_, bytes bClaims_) returns(bool)
func (_LightSnapshots *LightSnapshotsTransactorSession) SnapshotWithValidData(groupSignature_ []byte, bClaims_ []byte) (*types.Transaction, error) {
	return _LightSnapshots.Contract.SnapshotWithValidData(&_LightSnapshots.TransactOpts, groupSignature_, bClaims_)
}

// LightSnapshotsInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the LightSnapshots contract.
type LightSnapshotsInitializedIterator struct {
	Event *LightSnapshotsInitialized // Event containing the contract specifics and raw log

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
func (it *LightSnapshotsInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LightSnapshotsInitialized)
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
		it.Event = new(LightSnapshotsInitialized)
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
func (it *LightSnapshotsInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LightSnapshotsInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LightSnapshotsInitialized represents a Initialized event raised by the LightSnapshots contract.
type LightSnapshotsInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_LightSnapshots *LightSnapshotsFilterer) FilterInitialized(opts *bind.FilterOpts) (*LightSnapshotsInitializedIterator, error) {

	logs, sub, err := _LightSnapshots.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &LightSnapshotsInitializedIterator{contract: _LightSnapshots.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_LightSnapshots *LightSnapshotsFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *LightSnapshotsInitialized) (event.Subscription, error) {

	logs, sub, err := _LightSnapshots.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LightSnapshotsInitialized)
				if err := _LightSnapshots.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_LightSnapshots *LightSnapshotsFilterer) ParseInitialized(log types.Log) (*LightSnapshotsInitialized, error) {
	event := new(LightSnapshotsInitialized)
	if err := _LightSnapshots.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LightSnapshotsSnapshotTakenIterator is returned from FilterSnapshotTaken and is used to iterate over the raw logs and unpacked data for SnapshotTaken events raised by the LightSnapshots contract.
type LightSnapshotsSnapshotTakenIterator struct {
	Event *LightSnapshotsSnapshotTaken // Event containing the contract specifics and raw log

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
func (it *LightSnapshotsSnapshotTakenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LightSnapshotsSnapshotTaken)
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
		it.Event = new(LightSnapshotsSnapshotTaken)
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
func (it *LightSnapshotsSnapshotTakenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LightSnapshotsSnapshotTakenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LightSnapshotsSnapshotTaken represents a SnapshotTaken event raised by the LightSnapshots contract.
type LightSnapshotsSnapshotTaken struct {
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
func (_LightSnapshots *LightSnapshotsFilterer) FilterSnapshotTaken(opts *bind.FilterOpts, epoch []*big.Int, validator []common.Address) (*LightSnapshotsSnapshotTakenIterator, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _LightSnapshots.contract.FilterLogs(opts, "SnapshotTaken", epochRule, validatorRule)
	if err != nil {
		return nil, err
	}
	return &LightSnapshotsSnapshotTakenIterator{contract: _LightSnapshots.contract, event: "SnapshotTaken", logs: logs, sub: sub}, nil
}

// WatchSnapshotTaken is a free log subscription operation binding the contract event 0x709e2f13a448a6bdef51a1fcadf45303dee0b6bff5bdf628829f401b019b7e9b.
//
// Solidity: event SnapshotTaken(uint256 chainId, uint256 indexed epoch, uint256 height, address indexed validator, bool isSafeToProceedConsensus, uint256[4] masterPublicKey, uint256[2] signature, (uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32) bClaims)
func (_LightSnapshots *LightSnapshotsFilterer) WatchSnapshotTaken(opts *bind.WatchOpts, sink chan<- *LightSnapshotsSnapshotTaken, epoch []*big.Int, validator []common.Address) (event.Subscription, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _LightSnapshots.contract.WatchLogs(opts, "SnapshotTaken", epochRule, validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LightSnapshotsSnapshotTaken)
				if err := _LightSnapshots.contract.UnpackLog(event, "SnapshotTaken", log); err != nil {
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
func (_LightSnapshots *LightSnapshotsFilterer) ParseSnapshotTaken(log types.Log) (*LightSnapshotsSnapshotTaken, error) {
	event := new(LightSnapshotsSnapshotTaken)
	if err := _LightSnapshots.contract.UnpackLog(event, "SnapshotTaken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
