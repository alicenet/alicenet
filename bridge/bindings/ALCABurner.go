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

// ALCABurnerMetaData contains all meta data concerning the ALCABurner contract.
var ALCABurnerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyALCA\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyFactory\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
}

// ALCABurnerABI is the input ABI used to generate the binding from.
// Deprecated: Use ALCABurnerMetaData.ABI instead.
var ALCABurnerABI = ALCABurnerMetaData.ABI

// ALCABurner is an auto generated Go binding around an Ethereum contract.
type ALCABurner struct {
	ALCABurnerCaller     // Read-only binding to the contract
	ALCABurnerTransactor // Write-only binding to the contract
	ALCABurnerFilterer   // Log filterer for contract events
}

// ALCABurnerCaller is an auto generated read-only Go binding around an Ethereum contract.
type ALCABurnerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ALCABurnerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ALCABurnerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ALCABurnerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ALCABurnerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ALCABurnerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ALCABurnerSession struct {
	Contract     *ALCABurner       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ALCABurnerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ALCABurnerCallerSession struct {
	Contract *ALCABurnerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// ALCABurnerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ALCABurnerTransactorSession struct {
	Contract     *ALCABurnerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// ALCABurnerRaw is an auto generated low-level Go binding around an Ethereum contract.
type ALCABurnerRaw struct {
	Contract *ALCABurner // Generic contract binding to access the raw methods on
}

// ALCABurnerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ALCABurnerCallerRaw struct {
	Contract *ALCABurnerCaller // Generic read-only contract binding to access the raw methods on
}

// ALCABurnerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ALCABurnerTransactorRaw struct {
	Contract *ALCABurnerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewALCABurner creates a new instance of ALCABurner, bound to a specific deployed contract.
func NewALCABurner(address common.Address, backend bind.ContractBackend) (*ALCABurner, error) {
	contract, err := bindALCABurner(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ALCABurner{ALCABurnerCaller: ALCABurnerCaller{contract: contract}, ALCABurnerTransactor: ALCABurnerTransactor{contract: contract}, ALCABurnerFilterer: ALCABurnerFilterer{contract: contract}}, nil
}

// NewALCABurnerCaller creates a new read-only instance of ALCABurner, bound to a specific deployed contract.
func NewALCABurnerCaller(address common.Address, caller bind.ContractCaller) (*ALCABurnerCaller, error) {
	contract, err := bindALCABurner(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ALCABurnerCaller{contract: contract}, nil
}

// NewALCABurnerTransactor creates a new write-only instance of ALCABurner, bound to a specific deployed contract.
func NewALCABurnerTransactor(address common.Address, transactor bind.ContractTransactor) (*ALCABurnerTransactor, error) {
	contract, err := bindALCABurner(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ALCABurnerTransactor{contract: contract}, nil
}

// NewALCABurnerFilterer creates a new log filterer instance of ALCABurner, bound to a specific deployed contract.
func NewALCABurnerFilterer(address common.Address, filterer bind.ContractFilterer) (*ALCABurnerFilterer, error) {
	contract, err := bindALCABurner(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ALCABurnerFilterer{contract: contract}, nil
}

// bindALCABurner binds a generic wrapper to an already deployed contract.
func bindALCABurner(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ALCABurnerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ALCABurner *ALCABurnerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ALCABurner.Contract.ALCABurnerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ALCABurner *ALCABurnerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ALCABurner.Contract.ALCABurnerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ALCABurner *ALCABurnerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ALCABurner.Contract.ALCABurnerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ALCABurner *ALCABurnerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ALCABurner.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ALCABurner *ALCABurnerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ALCABurner.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ALCABurner *ALCABurnerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ALCABurner.Contract.contract.Transact(opts, method, params...)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ALCABurner *ALCABurnerCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _ALCABurner.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ALCABurner *ALCABurnerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ALCABurner.Contract.GetMetamorphicContractAddress(&_ALCABurner.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ALCABurner *ALCABurnerCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ALCABurner.Contract.GetMetamorphicContractAddress(&_ALCABurner.CallOpts, _salt, _factory)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address from_, uint256 amount_) returns()
func (_ALCABurner *ALCABurnerTransactor) Burn(opts *bind.TransactOpts, from_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _ALCABurner.contract.Transact(opts, "burn", from_, amount_)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address from_, uint256 amount_) returns()
func (_ALCABurner *ALCABurnerSession) Burn(from_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _ALCABurner.Contract.Burn(&_ALCABurner.TransactOpts, from_, amount_)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address from_, uint256 amount_) returns()
func (_ALCABurner *ALCABurnerTransactorSession) Burn(from_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _ALCABurner.Contract.Burn(&_ALCABurner.TransactOpts, from_, amount_)
}
