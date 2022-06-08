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

// ATokenMetaData contains all meta data concerning the AToken contract.
var ATokenMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"legacyToken_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"allowMigration\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"externalBurn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"externalMint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLegacyTokenAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"migrate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ATokenABI is the input ABI used to generate the binding from.
// Deprecated: Use ATokenMetaData.ABI instead.
var ATokenABI = ATokenMetaData.ABI

// AToken is an auto generated Go binding around an Ethereum contract.
type AToken struct {
	ATokenCaller     // Read-only binding to the contract
	ATokenTransactor // Write-only binding to the contract
	ATokenFilterer   // Log filterer for contract events
}

// ATokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type ATokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ATokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ATokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ATokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ATokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ATokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ATokenSession struct {
	Contract     *AToken           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ATokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ATokenCallerSession struct {
	Contract *ATokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ATokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ATokenTransactorSession struct {
	Contract     *ATokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ATokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type ATokenRaw struct {
	Contract *AToken // Generic contract binding to access the raw methods on
}

// ATokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ATokenCallerRaw struct {
	Contract *ATokenCaller // Generic read-only contract binding to access the raw methods on
}

// ATokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ATokenTransactorRaw struct {
	Contract *ATokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAToken creates a new instance of AToken, bound to a specific deployed contract.
func NewAToken(address common.Address, backend bind.ContractBackend) (*AToken, error) {
	contract, err := bindAToken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AToken{ATokenCaller: ATokenCaller{contract: contract}, ATokenTransactor: ATokenTransactor{contract: contract}, ATokenFilterer: ATokenFilterer{contract: contract}}, nil
}

// NewATokenCaller creates a new read-only instance of AToken, bound to a specific deployed contract.
func NewATokenCaller(address common.Address, caller bind.ContractCaller) (*ATokenCaller, error) {
	contract, err := bindAToken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ATokenCaller{contract: contract}, nil
}

// NewATokenTransactor creates a new write-only instance of AToken, bound to a specific deployed contract.
func NewATokenTransactor(address common.Address, transactor bind.ContractTransactor) (*ATokenTransactor, error) {
	contract, err := bindAToken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ATokenTransactor{contract: contract}, nil
}

// NewATokenFilterer creates a new log filterer instance of AToken, bound to a specific deployed contract.
func NewATokenFilterer(address common.Address, filterer bind.ContractFilterer) (*ATokenFilterer, error) {
	contract, err := bindAToken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ATokenFilterer{contract: contract}, nil
}

// bindAToken binds a generic wrapper to an already deployed contract.
func bindAToken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ATokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AToken *ATokenRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AToken.Contract.ATokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AToken *ATokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AToken.Contract.ATokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AToken *ATokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AToken.Contract.ATokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AToken *ATokenCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AToken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AToken *ATokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AToken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AToken *ATokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AToken.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_AToken *ATokenCaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AToken.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_AToken *ATokenSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _AToken.Contract.Allowance(&_AToken.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_AToken *ATokenCallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _AToken.Contract.Allowance(&_AToken.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_AToken *ATokenCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AToken.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_AToken *ATokenSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _AToken.Contract.BalanceOf(&_AToken.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_AToken *ATokenCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _AToken.Contract.BalanceOf(&_AToken.CallOpts, account)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_AToken *ATokenCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _AToken.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_AToken *ATokenSession) Decimals() (uint8, error) {
	return _AToken.Contract.Decimals(&_AToken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_AToken *ATokenCallerSession) Decimals() (uint8, error) {
	return _AToken.Contract.Decimals(&_AToken.CallOpts)
}

// GetLegacyTokenAddress is a free data retrieval call binding the contract method 0x035c7099.
//
// Solidity: function getLegacyTokenAddress() view returns(address)
func (_AToken *ATokenCaller) GetLegacyTokenAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AToken.contract.Call(opts, &out, "getLegacyTokenAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetLegacyTokenAddress is a free data retrieval call binding the contract method 0x035c7099.
//
// Solidity: function getLegacyTokenAddress() view returns(address)
func (_AToken *ATokenSession) GetLegacyTokenAddress() (common.Address, error) {
	return _AToken.Contract.GetLegacyTokenAddress(&_AToken.CallOpts)
}

// GetLegacyTokenAddress is a free data retrieval call binding the contract method 0x035c7099.
//
// Solidity: function getLegacyTokenAddress() view returns(address)
func (_AToken *ATokenCallerSession) GetLegacyTokenAddress() (common.Address, error) {
	return _AToken.Contract.GetLegacyTokenAddress(&_AToken.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_AToken *ATokenCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _AToken.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_AToken *ATokenSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _AToken.Contract.GetMetamorphicContractAddress(&_AToken.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_AToken *ATokenCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _AToken.Contract.GetMetamorphicContractAddress(&_AToken.CallOpts, _salt, _factory)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AToken *ATokenCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AToken.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AToken *ATokenSession) Name() (string, error) {
	return _AToken.Contract.Name(&_AToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_AToken *ATokenCallerSession) Name() (string, error) {
	return _AToken.Contract.Name(&_AToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AToken *ATokenCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AToken.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AToken *ATokenSession) Symbol() (string, error) {
	return _AToken.Contract.Symbol(&_AToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_AToken *ATokenCallerSession) Symbol() (string, error) {
	return _AToken.Contract.Symbol(&_AToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_AToken *ATokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AToken.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_AToken *ATokenSession) TotalSupply() (*big.Int, error) {
	return _AToken.Contract.TotalSupply(&_AToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_AToken *ATokenCallerSession) TotalSupply() (*big.Int, error) {
	return _AToken.Contract.TotalSupply(&_AToken.CallOpts)
}

// AllowMigration is a paid mutator transaction binding the contract method 0x35d69059.
//
// Solidity: function allowMigration() returns()
func (_AToken *ATokenTransactor) AllowMigration(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AToken.contract.Transact(opts, "allowMigration")
}

// AllowMigration is a paid mutator transaction binding the contract method 0x35d69059.
//
// Solidity: function allowMigration() returns()
func (_AToken *ATokenSession) AllowMigration() (*types.Transaction, error) {
	return _AToken.Contract.AllowMigration(&_AToken.TransactOpts)
}

// AllowMigration is a paid mutator transaction binding the contract method 0x35d69059.
//
// Solidity: function allowMigration() returns()
func (_AToken *ATokenTransactorSession) AllowMigration() (*types.Transaction, error) {
	return _AToken.Contract.AllowMigration(&_AToken.TransactOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_AToken *ATokenTransactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_AToken *ATokenSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.Approve(&_AToken.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_AToken *ATokenTransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.Approve(&_AToken.TransactOpts, spender, amount)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_AToken *ATokenTransactor) DecreaseAllowance(opts *bind.TransactOpts, spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _AToken.contract.Transact(opts, "decreaseAllowance", spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_AToken *ATokenSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.DecreaseAllowance(&_AToken.TransactOpts, spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_AToken *ATokenTransactorSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.DecreaseAllowance(&_AToken.TransactOpts, spender, subtractedValue)
}

// ExternalBurn is a paid mutator transaction binding the contract method 0x6f6ebec8.
//
// Solidity: function externalBurn(address from, uint256 amount) returns()
func (_AToken *ATokenTransactor) ExternalBurn(opts *bind.TransactOpts, from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.contract.Transact(opts, "externalBurn", from, amount)
}

// ExternalBurn is a paid mutator transaction binding the contract method 0x6f6ebec8.
//
// Solidity: function externalBurn(address from, uint256 amount) returns()
func (_AToken *ATokenSession) ExternalBurn(from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.ExternalBurn(&_AToken.TransactOpts, from, amount)
}

// ExternalBurn is a paid mutator transaction binding the contract method 0x6f6ebec8.
//
// Solidity: function externalBurn(address from, uint256 amount) returns()
func (_AToken *ATokenTransactorSession) ExternalBurn(from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.ExternalBurn(&_AToken.TransactOpts, from, amount)
}

// ExternalMint is a paid mutator transaction binding the contract method 0x99f98898.
//
// Solidity: function externalMint(address to, uint256 amount) returns()
func (_AToken *ATokenTransactor) ExternalMint(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.contract.Transact(opts, "externalMint", to, amount)
}

// ExternalMint is a paid mutator transaction binding the contract method 0x99f98898.
//
// Solidity: function externalMint(address to, uint256 amount) returns()
func (_AToken *ATokenSession) ExternalMint(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.ExternalMint(&_AToken.TransactOpts, to, amount)
}

// ExternalMint is a paid mutator transaction binding the contract method 0x99f98898.
//
// Solidity: function externalMint(address to, uint256 amount) returns()
func (_AToken *ATokenTransactorSession) ExternalMint(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.ExternalMint(&_AToken.TransactOpts, to, amount)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_AToken *ATokenTransactor) IncreaseAllowance(opts *bind.TransactOpts, spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _AToken.contract.Transact(opts, "increaseAllowance", spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_AToken *ATokenSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.IncreaseAllowance(&_AToken.TransactOpts, spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_AToken *ATokenTransactorSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.IncreaseAllowance(&_AToken.TransactOpts, spender, addedValue)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_AToken *ATokenTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AToken.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_AToken *ATokenSession) Initialize() (*types.Transaction, error) {
	return _AToken.Contract.Initialize(&_AToken.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_AToken *ATokenTransactorSession) Initialize() (*types.Transaction, error) {
	return _AToken.Contract.Initialize(&_AToken.TransactOpts)
}

// Migrate is a paid mutator transaction binding the contract method 0x454b0608.
//
// Solidity: function migrate(uint256 amount) returns()
func (_AToken *ATokenTransactor) Migrate(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _AToken.contract.Transact(opts, "migrate", amount)
}

// Migrate is a paid mutator transaction binding the contract method 0x454b0608.
//
// Solidity: function migrate(uint256 amount) returns()
func (_AToken *ATokenSession) Migrate(amount *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.Migrate(&_AToken.TransactOpts, amount)
}

// Migrate is a paid mutator transaction binding the contract method 0x454b0608.
//
// Solidity: function migrate(uint256 amount) returns()
func (_AToken *ATokenTransactorSession) Migrate(amount *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.Migrate(&_AToken.TransactOpts, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_AToken *ATokenTransactor) Transfer(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.contract.Transact(opts, "transfer", to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_AToken *ATokenSession) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.Transfer(&_AToken.TransactOpts, to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_AToken *ATokenTransactorSession) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.Transfer(&_AToken.TransactOpts, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_AToken *ATokenTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.contract.Transact(opts, "transferFrom", from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_AToken *ATokenSession) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.TransferFrom(&_AToken.TransactOpts, from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_AToken *ATokenTransactorSession) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AToken.Contract.TransferFrom(&_AToken.TransactOpts, from, to, amount)
}

// ATokenApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the AToken contract.
type ATokenApprovalIterator struct {
	Event *ATokenApproval // Event containing the contract specifics and raw log

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
func (it *ATokenApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ATokenApproval)
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
		it.Event = new(ATokenApproval)
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
func (it *ATokenApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ATokenApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ATokenApproval represents a Approval event raised by the AToken contract.
type ATokenApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_AToken *ATokenFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*ATokenApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _AToken.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &ATokenApprovalIterator{contract: _AToken.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_AToken *ATokenFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ATokenApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _AToken.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ATokenApproval)
				if err := _AToken.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_AToken *ATokenFilterer) ParseApproval(log types.Log) (*ATokenApproval, error) {
	event := new(ATokenApproval)
	if err := _AToken.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ATokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the AToken contract.
type ATokenTransferIterator struct {
	Event *ATokenTransfer // Event containing the contract specifics and raw log

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
func (it *ATokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ATokenTransfer)
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
		it.Event = new(ATokenTransfer)
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
func (it *ATokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ATokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ATokenTransfer represents a Transfer event raised by the AToken contract.
type ATokenTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_AToken *ATokenFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*ATokenTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _AToken.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &ATokenTransferIterator{contract: _AToken.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_AToken *ATokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ATokenTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _AToken.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ATokenTransfer)
				if err := _AToken.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_AToken *ATokenFilterer) ParseTransfer(log types.Log) (*ATokenTransfer, error) {
	event := new(ATokenTransfer)
	if err := _AToken.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
