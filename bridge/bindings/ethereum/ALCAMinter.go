// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ethereum

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

// ALCAMinterMetaData contains all meta data concerning the ALCAMinter contract.
var ALCAMinterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyALCA\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyFactory\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ALCAMinterABI is the input ABI used to generate the binding from.
// Deprecated: Use ALCAMinterMetaData.ABI instead.
var ALCAMinterABI = ALCAMinterMetaData.ABI

// ALCAMinter is an auto generated Go binding around an Ethereum contract.
type ALCAMinter struct {
	ALCAMinterCaller     // Read-only binding to the contract
	ALCAMinterTransactor // Write-only binding to the contract
	ALCAMinterFilterer   // Log filterer for contract events
}

// ALCAMinterCaller is an auto generated read-only Go binding around an Ethereum contract.
type ALCAMinterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ALCAMinterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ALCAMinterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ALCAMinterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ALCAMinterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ALCAMinterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ALCAMinterSession struct {
	Contract     *ALCAMinter       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ALCAMinterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ALCAMinterCallerSession struct {
	Contract *ALCAMinterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// ALCAMinterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ALCAMinterTransactorSession struct {
	Contract     *ALCAMinterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// ALCAMinterRaw is an auto generated low-level Go binding around an Ethereum contract.
type ALCAMinterRaw struct {
	Contract *ALCAMinter // Generic contract binding to access the raw methods on
}

// ALCAMinterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ALCAMinterCallerRaw struct {
	Contract *ALCAMinterCaller // Generic read-only contract binding to access the raw methods on
}

// ALCAMinterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ALCAMinterTransactorRaw struct {
	Contract *ALCAMinterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewALCAMinter creates a new instance of ALCAMinter, bound to a specific deployed contract.
func NewALCAMinter(address common.Address, backend bind.ContractBackend) (*ALCAMinter, error) {
	contract, err := bindALCAMinter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ALCAMinter{ALCAMinterCaller: ALCAMinterCaller{contract: contract}, ALCAMinterTransactor: ALCAMinterTransactor{contract: contract}, ALCAMinterFilterer: ALCAMinterFilterer{contract: contract}}, nil
}

// NewALCAMinterCaller creates a new read-only instance of ALCAMinter, bound to a specific deployed contract.
func NewALCAMinterCaller(address common.Address, caller bind.ContractCaller) (*ALCAMinterCaller, error) {
	contract, err := bindALCAMinter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ALCAMinterCaller{contract: contract}, nil
}

// NewALCAMinterTransactor creates a new write-only instance of ALCAMinter, bound to a specific deployed contract.
func NewALCAMinterTransactor(address common.Address, transactor bind.ContractTransactor) (*ALCAMinterTransactor, error) {
	contract, err := bindALCAMinter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ALCAMinterTransactor{contract: contract}, nil
}

// NewALCAMinterFilterer creates a new log filterer instance of ALCAMinter, bound to a specific deployed contract.
func NewALCAMinterFilterer(address common.Address, filterer bind.ContractFilterer) (*ALCAMinterFilterer, error) {
	contract, err := bindALCAMinter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ALCAMinterFilterer{contract: contract}, nil
}

// bindALCAMinter binds a generic wrapper to an already deployed contract.
func bindALCAMinter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ALCAMinterABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ALCAMinter *ALCAMinterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ALCAMinter.Contract.ALCAMinterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ALCAMinter *ALCAMinterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ALCAMinter.Contract.ALCAMinterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ALCAMinter *ALCAMinterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ALCAMinter.Contract.ALCAMinterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ALCAMinter *ALCAMinterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ALCAMinter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ALCAMinter *ALCAMinterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ALCAMinter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ALCAMinter *ALCAMinterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ALCAMinter.Contract.contract.Transact(opts, method, params...)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ALCAMinter *ALCAMinterCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _ALCAMinter.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ALCAMinter *ALCAMinterSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ALCAMinter.Contract.GetMetamorphicContractAddress(&_ALCAMinter.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ALCAMinter *ALCAMinterCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ALCAMinter.Contract.GetMetamorphicContractAddress(&_ALCAMinter.CallOpts, _salt, _factory)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to_, uint256 amount_) returns()
func (_ALCAMinter *ALCAMinterTransactor) Mint(opts *bind.TransactOpts, to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _ALCAMinter.contract.Transact(opts, "mint", to_, amount_)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to_, uint256 amount_) returns()
func (_ALCAMinter *ALCAMinterSession) Mint(to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _ALCAMinter.Contract.Mint(&_ALCAMinter.TransactOpts, to_, amount_)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to_, uint256 amount_) returns()
func (_ALCAMinter *ALCAMinterTransactorSession) Mint(to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _ALCAMinter.Contract.Mint(&_ALCAMinter.TransactOpts, to_, amount_)
}
