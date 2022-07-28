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

// BTokenErrorsMetaData contains all meta data concerning the BTokenErrors contract.
var BTokenErrorsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"supply\",\"type\":\"uint256\"}],\"name\":\"BurnAmountExceedsSupply\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"toAddress\",\"type\":\"address\"}],\"name\":\"ContractsDisallowedDeposits\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DepositAmountZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DepositBurnFail\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"name\":\"InexistentRouterContract\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"InsufficientEth\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"}],\"name\":\"InsufficientFee\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"contractBalance\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"poolBalance\",\"type\":\"uint256\"}],\"name\":\"InvalidBalance\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"InvalidBurnAmount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"depositID\",\"type\":\"uint256\"}],\"name\":\"InvalidDepositId\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"MinimumBurnNotMet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"MinimumMintNotMet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimumValue\",\"type\":\"uint256\"}],\"name\":\"MinimumValueNotMet\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SplitValueSumError\",\"type\":\"error\"}]",
}

// BTokenErrorsABI is the input ABI used to generate the binding from.
// Deprecated: Use BTokenErrorsMetaData.ABI instead.
var BTokenErrorsABI = BTokenErrorsMetaData.ABI

// BTokenErrors is an auto generated Go binding around an Ethereum contract.
type BTokenErrors struct {
	BTokenErrorsCaller     // Read-only binding to the contract
	BTokenErrorsTransactor // Write-only binding to the contract
	BTokenErrorsFilterer   // Log filterer for contract events
}

// BTokenErrorsCaller is an auto generated read-only Go binding around an Ethereum contract.
type BTokenErrorsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BTokenErrorsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BTokenErrorsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BTokenErrorsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BTokenErrorsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BTokenErrorsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BTokenErrorsSession struct {
	Contract     *BTokenErrors     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BTokenErrorsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BTokenErrorsCallerSession struct {
	Contract *BTokenErrorsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// BTokenErrorsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BTokenErrorsTransactorSession struct {
	Contract     *BTokenErrorsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// BTokenErrorsRaw is an auto generated low-level Go binding around an Ethereum contract.
type BTokenErrorsRaw struct {
	Contract *BTokenErrors // Generic contract binding to access the raw methods on
}

// BTokenErrorsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BTokenErrorsCallerRaw struct {
	Contract *BTokenErrorsCaller // Generic read-only contract binding to access the raw methods on
}

// BTokenErrorsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BTokenErrorsTransactorRaw struct {
	Contract *BTokenErrorsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBTokenErrors creates a new instance of BTokenErrors, bound to a specific deployed contract.
func NewBTokenErrors(address common.Address, backend bind.ContractBackend) (*BTokenErrors, error) {
	contract, err := bindBTokenErrors(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BTokenErrors{BTokenErrorsCaller: BTokenErrorsCaller{contract: contract}, BTokenErrorsTransactor: BTokenErrorsTransactor{contract: contract}, BTokenErrorsFilterer: BTokenErrorsFilterer{contract: contract}}, nil
}

// NewBTokenErrorsCaller creates a new read-only instance of BTokenErrors, bound to a specific deployed contract.
func NewBTokenErrorsCaller(address common.Address, caller bind.ContractCaller) (*BTokenErrorsCaller, error) {
	contract, err := bindBTokenErrors(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BTokenErrorsCaller{contract: contract}, nil
}

// NewBTokenErrorsTransactor creates a new write-only instance of BTokenErrors, bound to a specific deployed contract.
func NewBTokenErrorsTransactor(address common.Address, transactor bind.ContractTransactor) (*BTokenErrorsTransactor, error) {
	contract, err := bindBTokenErrors(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BTokenErrorsTransactor{contract: contract}, nil
}

// NewBTokenErrorsFilterer creates a new log filterer instance of BTokenErrors, bound to a specific deployed contract.
func NewBTokenErrorsFilterer(address common.Address, filterer bind.ContractFilterer) (*BTokenErrorsFilterer, error) {
	contract, err := bindBTokenErrors(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BTokenErrorsFilterer{contract: contract}, nil
}

// bindBTokenErrors binds a generic wrapper to an already deployed contract.
func bindBTokenErrors(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BTokenErrorsABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BTokenErrors *BTokenErrorsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BTokenErrors.Contract.BTokenErrorsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BTokenErrors *BTokenErrorsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BTokenErrors.Contract.BTokenErrorsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BTokenErrors *BTokenErrorsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BTokenErrors.Contract.BTokenErrorsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BTokenErrors *BTokenErrorsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BTokenErrors.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BTokenErrors *BTokenErrorsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BTokenErrors.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BTokenErrors *BTokenErrorsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BTokenErrors.Contract.contract.Transact(opts, method, params...)
}
