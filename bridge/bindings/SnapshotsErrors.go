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

// SnapshotsErrorsMetaData contains all meta data concerning the SnapshotsErrors contract.
var SnapshotsErrorsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"ConsensusNotRunning\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"EpochMustBeNonZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blockHeight\",\"type\":\"uint256\"}],\"name\":\"InvalidBlockHeight\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"}],\"name\":\"InvalidChainId\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"calculatedMasterKeyHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"expectedMasterKeyHash\",\"type\":\"bytes32\"}],\"name\":\"InvalidMasterPublicKey\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newBlockHeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"oldBlockHeight\",\"type\":\"uint256\"}],\"name\":\"InvalidRingBufferBlockHeight\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"groupSignatureLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bClaimsLength\",\"type\":\"uint256\"}],\"name\":\"MigrationInputDataMismatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MigrationNotAllowedAtCurrentEpoch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"currentBlocksInterval\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimumBlocksInterval\",\"type\":\"uint256\"}],\"name\":\"MinimumBlocksIntervalNotPassed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"OnlyValidatorsAllowed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SignatureVerificationFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"}],\"name\":\"SnapshotsNotInBuffer\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"groupSignatureHash\",\"type\":\"bytes32\"}],\"name\":\"ValidatorNotElected\",\"type\":\"error\"}]",
}

// SnapshotsErrorsABI is the input ABI used to generate the binding from.
// Deprecated: Use SnapshotsErrorsMetaData.ABI instead.
var SnapshotsErrorsABI = SnapshotsErrorsMetaData.ABI

// SnapshotsErrors is an auto generated Go binding around an Ethereum contract.
type SnapshotsErrors struct {
	SnapshotsErrorsCaller     // Read-only binding to the contract
	SnapshotsErrorsTransactor // Write-only binding to the contract
	SnapshotsErrorsFilterer   // Log filterer for contract events
}

// SnapshotsErrorsCaller is an auto generated read-only Go binding around an Ethereum contract.
type SnapshotsErrorsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SnapshotsErrorsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SnapshotsErrorsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SnapshotsErrorsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SnapshotsErrorsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SnapshotsErrorsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SnapshotsErrorsSession struct {
	Contract     *SnapshotsErrors  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SnapshotsErrorsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SnapshotsErrorsCallerSession struct {
	Contract *SnapshotsErrorsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// SnapshotsErrorsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SnapshotsErrorsTransactorSession struct {
	Contract     *SnapshotsErrorsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// SnapshotsErrorsRaw is an auto generated low-level Go binding around an Ethereum contract.
type SnapshotsErrorsRaw struct {
	Contract *SnapshotsErrors // Generic contract binding to access the raw methods on
}

// SnapshotsErrorsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SnapshotsErrorsCallerRaw struct {
	Contract *SnapshotsErrorsCaller // Generic read-only contract binding to access the raw methods on
}

// SnapshotsErrorsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SnapshotsErrorsTransactorRaw struct {
	Contract *SnapshotsErrorsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSnapshotsErrors creates a new instance of SnapshotsErrors, bound to a specific deployed contract.
func NewSnapshotsErrors(address common.Address, backend bind.ContractBackend) (*SnapshotsErrors, error) {
	contract, err := bindSnapshotsErrors(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SnapshotsErrors{SnapshotsErrorsCaller: SnapshotsErrorsCaller{contract: contract}, SnapshotsErrorsTransactor: SnapshotsErrorsTransactor{contract: contract}, SnapshotsErrorsFilterer: SnapshotsErrorsFilterer{contract: contract}}, nil
}

// NewSnapshotsErrorsCaller creates a new read-only instance of SnapshotsErrors, bound to a specific deployed contract.
func NewSnapshotsErrorsCaller(address common.Address, caller bind.ContractCaller) (*SnapshotsErrorsCaller, error) {
	contract, err := bindSnapshotsErrors(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SnapshotsErrorsCaller{contract: contract}, nil
}

// NewSnapshotsErrorsTransactor creates a new write-only instance of SnapshotsErrors, bound to a specific deployed contract.
func NewSnapshotsErrorsTransactor(address common.Address, transactor bind.ContractTransactor) (*SnapshotsErrorsTransactor, error) {
	contract, err := bindSnapshotsErrors(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SnapshotsErrorsTransactor{contract: contract}, nil
}

// NewSnapshotsErrorsFilterer creates a new log filterer instance of SnapshotsErrors, bound to a specific deployed contract.
func NewSnapshotsErrorsFilterer(address common.Address, filterer bind.ContractFilterer) (*SnapshotsErrorsFilterer, error) {
	contract, err := bindSnapshotsErrors(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SnapshotsErrorsFilterer{contract: contract}, nil
}

// bindSnapshotsErrors binds a generic wrapper to an already deployed contract.
func bindSnapshotsErrors(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SnapshotsErrorsABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SnapshotsErrors *SnapshotsErrorsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SnapshotsErrors.Contract.SnapshotsErrorsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SnapshotsErrors *SnapshotsErrorsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SnapshotsErrors.Contract.SnapshotsErrorsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SnapshotsErrors *SnapshotsErrorsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SnapshotsErrors.Contract.SnapshotsErrorsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SnapshotsErrors *SnapshotsErrorsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SnapshotsErrors.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SnapshotsErrors *SnapshotsErrorsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SnapshotsErrors.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SnapshotsErrors *SnapshotsErrorsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SnapshotsErrors.Contract.contract.Transact(opts, method, params...)
}
