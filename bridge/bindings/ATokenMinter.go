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

// ATokenMinterMetaData contains all meta data concerning the ATokenMinter contract.
var ATokenMinterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ATokenMinterABI is the input ABI used to generate the binding from.
// Deprecated: Use ATokenMinterMetaData.ABI instead.
var ATokenMinterABI = ATokenMinterMetaData.ABI

// ATokenMinter is an auto generated Go binding around an Ethereum contract.
type ATokenMinter struct {
	ATokenMinterCaller     // Read-only binding to the contract
	ATokenMinterTransactor // Write-only binding to the contract
	ATokenMinterFilterer   // Log filterer for contract events
}

// ATokenMinterCaller is an auto generated read-only Go binding around an Ethereum contract.
type ATokenMinterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ATokenMinterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ATokenMinterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ATokenMinterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ATokenMinterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ATokenMinterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ATokenMinterSession struct {
	Contract     *ATokenMinter     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ATokenMinterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ATokenMinterCallerSession struct {
	Contract *ATokenMinterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// ATokenMinterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ATokenMinterTransactorSession struct {
	Contract     *ATokenMinterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// ATokenMinterRaw is an auto generated low-level Go binding around an Ethereum contract.
type ATokenMinterRaw struct {
	Contract *ATokenMinter // Generic contract binding to access the raw methods on
}

// ATokenMinterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ATokenMinterCallerRaw struct {
	Contract *ATokenMinterCaller // Generic read-only contract binding to access the raw methods on
}

// ATokenMinterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ATokenMinterTransactorRaw struct {
	Contract *ATokenMinterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewATokenMinter creates a new instance of ATokenMinter, bound to a specific deployed contract.
func NewATokenMinter(address common.Address, backend bind.ContractBackend) (*ATokenMinter, error) {
	contract, err := bindATokenMinter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ATokenMinter{ATokenMinterCaller: ATokenMinterCaller{contract: contract}, ATokenMinterTransactor: ATokenMinterTransactor{contract: contract}, ATokenMinterFilterer: ATokenMinterFilterer{contract: contract}}, nil
}

// NewATokenMinterCaller creates a new read-only instance of ATokenMinter, bound to a specific deployed contract.
func NewATokenMinterCaller(address common.Address, caller bind.ContractCaller) (*ATokenMinterCaller, error) {
	contract, err := bindATokenMinter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ATokenMinterCaller{contract: contract}, nil
}

// NewATokenMinterTransactor creates a new write-only instance of ATokenMinter, bound to a specific deployed contract.
func NewATokenMinterTransactor(address common.Address, transactor bind.ContractTransactor) (*ATokenMinterTransactor, error) {
	contract, err := bindATokenMinter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ATokenMinterTransactor{contract: contract}, nil
}

// NewATokenMinterFilterer creates a new log filterer instance of ATokenMinter, bound to a specific deployed contract.
func NewATokenMinterFilterer(address common.Address, filterer bind.ContractFilterer) (*ATokenMinterFilterer, error) {
	contract, err := bindATokenMinter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ATokenMinterFilterer{contract: contract}, nil
}

// bindATokenMinter binds a generic wrapper to an already deployed contract.
func bindATokenMinter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ATokenMinterABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ATokenMinter *ATokenMinterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ATokenMinter.Contract.ATokenMinterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ATokenMinter *ATokenMinterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ATokenMinter.Contract.ATokenMinterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ATokenMinter *ATokenMinterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ATokenMinter.Contract.ATokenMinterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ATokenMinter *ATokenMinterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ATokenMinter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ATokenMinter *ATokenMinterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ATokenMinter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ATokenMinter *ATokenMinterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ATokenMinter.Contract.contract.Transact(opts, method, params...)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ATokenMinter *ATokenMinterCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _ATokenMinter.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ATokenMinter *ATokenMinterSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ATokenMinter.Contract.GetMetamorphicContractAddress(&_ATokenMinter.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ATokenMinter *ATokenMinterCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ATokenMinter.Contract.GetMetamorphicContractAddress(&_ATokenMinter.CallOpts, _salt, _factory)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 amount) returns()
func (_ATokenMinter *ATokenMinterTransactor) Mint(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ATokenMinter.contract.Transact(opts, "mint", to, amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 amount) returns()
func (_ATokenMinter *ATokenMinterSession) Mint(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ATokenMinter.Contract.Mint(&_ATokenMinter.TransactOpts, to, amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 amount) returns()
func (_ATokenMinter *ATokenMinterTransactorSession) Mint(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ATokenMinter.Contract.Mint(&_ATokenMinter.TransactOpts, to, amount)
}
