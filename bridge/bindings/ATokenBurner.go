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

// ATokenBurnerMetaData contains all meta data concerning the ATokenBurner contract.
var ATokenBurnerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
}

// ATokenBurnerABI is the input ABI used to generate the binding from.
// Deprecated: Use ATokenBurnerMetaData.ABI instead.
var ATokenBurnerABI = ATokenBurnerMetaData.ABI

// ATokenBurner is an auto generated Go binding around an Ethereum contract.
type ATokenBurner struct {
	ATokenBurnerCaller     // Read-only binding to the contract
	ATokenBurnerTransactor // Write-only binding to the contract
	ATokenBurnerFilterer   // Log filterer for contract events
}

// ATokenBurnerCaller is an auto generated read-only Go binding around an Ethereum contract.
type ATokenBurnerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ATokenBurnerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ATokenBurnerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ATokenBurnerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ATokenBurnerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ATokenBurnerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ATokenBurnerSession struct {
	Contract     *ATokenBurner     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ATokenBurnerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ATokenBurnerCallerSession struct {
	Contract *ATokenBurnerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// ATokenBurnerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ATokenBurnerTransactorSession struct {
	Contract     *ATokenBurnerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// ATokenBurnerRaw is an auto generated low-level Go binding around an Ethereum contract.
type ATokenBurnerRaw struct {
	Contract *ATokenBurner // Generic contract binding to access the raw methods on
}

// ATokenBurnerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ATokenBurnerCallerRaw struct {
	Contract *ATokenBurnerCaller // Generic read-only contract binding to access the raw methods on
}

// ATokenBurnerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ATokenBurnerTransactorRaw struct {
	Contract *ATokenBurnerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewATokenBurner creates a new instance of ATokenBurner, bound to a specific deployed contract.
func NewATokenBurner(address common.Address, backend bind.ContractBackend) (*ATokenBurner, error) {
	contract, err := bindATokenBurner(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ATokenBurner{ATokenBurnerCaller: ATokenBurnerCaller{contract: contract}, ATokenBurnerTransactor: ATokenBurnerTransactor{contract: contract}, ATokenBurnerFilterer: ATokenBurnerFilterer{contract: contract}}, nil
}

// NewATokenBurnerCaller creates a new read-only instance of ATokenBurner, bound to a specific deployed contract.
func NewATokenBurnerCaller(address common.Address, caller bind.ContractCaller) (*ATokenBurnerCaller, error) {
	contract, err := bindATokenBurner(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ATokenBurnerCaller{contract: contract}, nil
}

// NewATokenBurnerTransactor creates a new write-only instance of ATokenBurner, bound to a specific deployed contract.
func NewATokenBurnerTransactor(address common.Address, transactor bind.ContractTransactor) (*ATokenBurnerTransactor, error) {
	contract, err := bindATokenBurner(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ATokenBurnerTransactor{contract: contract}, nil
}

// NewATokenBurnerFilterer creates a new log filterer instance of ATokenBurner, bound to a specific deployed contract.
func NewATokenBurnerFilterer(address common.Address, filterer bind.ContractFilterer) (*ATokenBurnerFilterer, error) {
	contract, err := bindATokenBurner(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ATokenBurnerFilterer{contract: contract}, nil
}

// bindATokenBurner binds a generic wrapper to an already deployed contract.
func bindATokenBurner(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ATokenBurnerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ATokenBurner *ATokenBurnerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ATokenBurner.Contract.ATokenBurnerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ATokenBurner *ATokenBurnerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ATokenBurner.Contract.ATokenBurnerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ATokenBurner *ATokenBurnerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ATokenBurner.Contract.ATokenBurnerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ATokenBurner *ATokenBurnerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ATokenBurner.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ATokenBurner *ATokenBurnerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ATokenBurner.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ATokenBurner *ATokenBurnerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ATokenBurner.Contract.contract.Transact(opts, method, params...)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ATokenBurner *ATokenBurnerCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _ATokenBurner.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ATokenBurner *ATokenBurnerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ATokenBurner.Contract.GetMetamorphicContractAddress(&_ATokenBurner.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ATokenBurner *ATokenBurnerCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ATokenBurner.Contract.GetMetamorphicContractAddress(&_ATokenBurner.CallOpts, _salt, _factory)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address to, uint256 amount) returns()
func (_ATokenBurner *ATokenBurnerTransactor) Burn(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ATokenBurner.contract.Transact(opts, "burn", to, amount)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address to, uint256 amount) returns()
func (_ATokenBurner *ATokenBurnerSession) Burn(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ATokenBurner.Contract.Burn(&_ATokenBurner.TransactOpts, to, amount)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address to, uint256 amount) returns()
func (_ATokenBurner *ATokenBurnerTransactorSession) Burn(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ATokenBurner.Contract.Burn(&_ATokenBurner.TransactOpts, to, amount)
}
