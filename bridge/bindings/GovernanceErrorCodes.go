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

// GovernanceErrorCodesMetaData contains all meta data concerning the GovernanceErrorCodes contract.
var GovernanceErrorCodesMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"GOVERNANCE_ONLY_FACTORY_ALLOWED\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// GovernanceErrorCodesABI is the input ABI used to generate the binding from.
// Deprecated: Use GovernanceErrorCodesMetaData.ABI instead.
var GovernanceErrorCodesABI = GovernanceErrorCodesMetaData.ABI

// GovernanceErrorCodes is an auto generated Go binding around an Ethereum contract.
type GovernanceErrorCodes struct {
	GovernanceErrorCodesCaller     // Read-only binding to the contract
	GovernanceErrorCodesTransactor // Write-only binding to the contract
	GovernanceErrorCodesFilterer   // Log filterer for contract events
}

// GovernanceErrorCodesCaller is an auto generated read-only Go binding around an Ethereum contract.
type GovernanceErrorCodesCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovernanceErrorCodesTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GovernanceErrorCodesTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovernanceErrorCodesFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GovernanceErrorCodesFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovernanceErrorCodesSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GovernanceErrorCodesSession struct {
	Contract     *GovernanceErrorCodes // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// GovernanceErrorCodesCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GovernanceErrorCodesCallerSession struct {
	Contract *GovernanceErrorCodesCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// GovernanceErrorCodesTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GovernanceErrorCodesTransactorSession struct {
	Contract     *GovernanceErrorCodesTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// GovernanceErrorCodesRaw is an auto generated low-level Go binding around an Ethereum contract.
type GovernanceErrorCodesRaw struct {
	Contract *GovernanceErrorCodes // Generic contract binding to access the raw methods on
}

// GovernanceErrorCodesCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GovernanceErrorCodesCallerRaw struct {
	Contract *GovernanceErrorCodesCaller // Generic read-only contract binding to access the raw methods on
}

// GovernanceErrorCodesTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GovernanceErrorCodesTransactorRaw struct {
	Contract *GovernanceErrorCodesTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGovernanceErrorCodes creates a new instance of GovernanceErrorCodes, bound to a specific deployed contract.
func NewGovernanceErrorCodes(address common.Address, backend bind.ContractBackend) (*GovernanceErrorCodes, error) {
	contract, err := bindGovernanceErrorCodes(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GovernanceErrorCodes{GovernanceErrorCodesCaller: GovernanceErrorCodesCaller{contract: contract}, GovernanceErrorCodesTransactor: GovernanceErrorCodesTransactor{contract: contract}, GovernanceErrorCodesFilterer: GovernanceErrorCodesFilterer{contract: contract}}, nil
}

// NewGovernanceErrorCodesCaller creates a new read-only instance of GovernanceErrorCodes, bound to a specific deployed contract.
func NewGovernanceErrorCodesCaller(address common.Address, caller bind.ContractCaller) (*GovernanceErrorCodesCaller, error) {
	contract, err := bindGovernanceErrorCodes(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GovernanceErrorCodesCaller{contract: contract}, nil
}

// NewGovernanceErrorCodesTransactor creates a new write-only instance of GovernanceErrorCodes, bound to a specific deployed contract.
func NewGovernanceErrorCodesTransactor(address common.Address, transactor bind.ContractTransactor) (*GovernanceErrorCodesTransactor, error) {
	contract, err := bindGovernanceErrorCodes(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GovernanceErrorCodesTransactor{contract: contract}, nil
}

// NewGovernanceErrorCodesFilterer creates a new log filterer instance of GovernanceErrorCodes, bound to a specific deployed contract.
func NewGovernanceErrorCodesFilterer(address common.Address, filterer bind.ContractFilterer) (*GovernanceErrorCodesFilterer, error) {
	contract, err := bindGovernanceErrorCodes(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GovernanceErrorCodesFilterer{contract: contract}, nil
}

// bindGovernanceErrorCodes binds a generic wrapper to an already deployed contract.
func bindGovernanceErrorCodes(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(GovernanceErrorCodesABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovernanceErrorCodes *GovernanceErrorCodesRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovernanceErrorCodes.Contract.GovernanceErrorCodesCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovernanceErrorCodes *GovernanceErrorCodesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovernanceErrorCodes.Contract.GovernanceErrorCodesTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovernanceErrorCodes *GovernanceErrorCodesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovernanceErrorCodes.Contract.GovernanceErrorCodesTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovernanceErrorCodes *GovernanceErrorCodesCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovernanceErrorCodes.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovernanceErrorCodes *GovernanceErrorCodesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovernanceErrorCodes.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovernanceErrorCodes *GovernanceErrorCodesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovernanceErrorCodes.Contract.contract.Transact(opts, method, params...)
}

// GOVERNANCEONLYFACTORYALLOWED is a free data retrieval call binding the contract method 0x4ce9164d.
//
// Solidity: function GOVERNANCE_ONLY_FACTORY_ALLOWED() view returns(bytes32)
func (_GovernanceErrorCodes *GovernanceErrorCodesCaller) GOVERNANCEONLYFACTORYALLOWED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _GovernanceErrorCodes.contract.Call(opts, &out, "GOVERNANCE_ONLY_FACTORY_ALLOWED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GOVERNANCEONLYFACTORYALLOWED is a free data retrieval call binding the contract method 0x4ce9164d.
//
// Solidity: function GOVERNANCE_ONLY_FACTORY_ALLOWED() view returns(bytes32)
func (_GovernanceErrorCodes *GovernanceErrorCodesSession) GOVERNANCEONLYFACTORYALLOWED() ([32]byte, error) {
	return _GovernanceErrorCodes.Contract.GOVERNANCEONLYFACTORYALLOWED(&_GovernanceErrorCodes.CallOpts)
}

// GOVERNANCEONLYFACTORYALLOWED is a free data retrieval call binding the contract method 0x4ce9164d.
//
// Solidity: function GOVERNANCE_ONLY_FACTORY_ALLOWED() view returns(bytes32)
func (_GovernanceErrorCodes *GovernanceErrorCodesCallerSession) GOVERNANCEONLYFACTORYALLOWED() ([32]byte, error) {
	return _GovernanceErrorCodes.Contract.GOVERNANCEONLYFACTORYALLOWED(&_GovernanceErrorCodes.CallOpts)
}
