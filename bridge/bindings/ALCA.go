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

// ALCAMetaData contains all meta data concerning the ALCA contract.
var ALCAMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"legacyToken_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"InvalidAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidConversionAmount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyALCABurner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyALCAMinter\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyFactory\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"convert\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"externalBurn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"externalMint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"finishEarlyStage\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLegacyTokenAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"migrate\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"migrateTo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ALCAABI is the input ABI used to generate the binding from.
// Deprecated: Use ALCAMetaData.ABI instead.
var ALCAABI = ALCAMetaData.ABI

// ALCA is an auto generated Go binding around an Ethereum contract.
type ALCA struct {
	ALCACaller     // Read-only binding to the contract
	ALCATransactor // Write-only binding to the contract
	ALCAFilterer   // Log filterer for contract events
}

// ALCACaller is an auto generated read-only Go binding around an Ethereum contract.
type ALCACaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ALCATransactor is an auto generated write-only Go binding around an Ethereum contract.
type ALCATransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ALCAFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ALCAFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ALCASession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ALCASession struct {
	Contract     *ALCA             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ALCACallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ALCACallerSession struct {
	Contract *ALCACaller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ALCATransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ALCATransactorSession struct {
	Contract     *ALCATransactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ALCARaw is an auto generated low-level Go binding around an Ethereum contract.
type ALCARaw struct {
	Contract *ALCA // Generic contract binding to access the raw methods on
}

// ALCACallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ALCACallerRaw struct {
	Contract *ALCACaller // Generic read-only contract binding to access the raw methods on
}

// ALCATransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ALCATransactorRaw struct {
	Contract *ALCATransactor // Generic write-only contract binding to access the raw methods on
}

// NewALCA creates a new instance of ALCA, bound to a specific deployed contract.
func NewALCA(address common.Address, backend bind.ContractBackend) (*ALCA, error) {
	contract, err := bindALCA(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ALCA{ALCACaller: ALCACaller{contract: contract}, ALCATransactor: ALCATransactor{contract: contract}, ALCAFilterer: ALCAFilterer{contract: contract}}, nil
}

// NewALCACaller creates a new read-only instance of ALCA, bound to a specific deployed contract.
func NewALCACaller(address common.Address, caller bind.ContractCaller) (*ALCACaller, error) {
	contract, err := bindALCA(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ALCACaller{contract: contract}, nil
}

// NewALCATransactor creates a new write-only instance of ALCA, bound to a specific deployed contract.
func NewALCATransactor(address common.Address, transactor bind.ContractTransactor) (*ALCATransactor, error) {
	contract, err := bindALCA(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ALCATransactor{contract: contract}, nil
}

// NewALCAFilterer creates a new log filterer instance of ALCA, bound to a specific deployed contract.
func NewALCAFilterer(address common.Address, filterer bind.ContractFilterer) (*ALCAFilterer, error) {
	contract, err := bindALCA(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ALCAFilterer{contract: contract}, nil
}

// bindALCA binds a generic wrapper to an already deployed contract.
func bindALCA(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ALCAABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ALCA *ALCARaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ALCA.Contract.ALCACaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ALCA *ALCARaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ALCA.Contract.ALCATransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ALCA *ALCARaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ALCA.Contract.ALCATransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ALCA *ALCACallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ALCA.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ALCA *ALCATransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ALCA.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ALCA *ALCATransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ALCA.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_ALCA *ALCACaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ALCA.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_ALCA *ALCASession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _ALCA.Contract.Allowance(&_ALCA.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_ALCA *ALCACallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _ALCA.Contract.Allowance(&_ALCA.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_ALCA *ALCACaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ALCA.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_ALCA *ALCASession) BalanceOf(account common.Address) (*big.Int, error) {
	return _ALCA.Contract.BalanceOf(&_ALCA.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_ALCA *ALCACallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _ALCA.Contract.BalanceOf(&_ALCA.CallOpts, account)
}

// Convert is a free data retrieval call binding the contract method 0xa3908e1b.
//
// Solidity: function convert(uint256 amount) view returns(uint256)
func (_ALCA *ALCACaller) Convert(opts *bind.CallOpts, amount *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ALCA.contract.Call(opts, &out, "convert", amount)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Convert is a free data retrieval call binding the contract method 0xa3908e1b.
//
// Solidity: function convert(uint256 amount) view returns(uint256)
func (_ALCA *ALCASession) Convert(amount *big.Int) (*big.Int, error) {
	return _ALCA.Contract.Convert(&_ALCA.CallOpts, amount)
}

// Convert is a free data retrieval call binding the contract method 0xa3908e1b.
//
// Solidity: function convert(uint256 amount) view returns(uint256)
func (_ALCA *ALCACallerSession) Convert(amount *big.Int) (*big.Int, error) {
	return _ALCA.Contract.Convert(&_ALCA.CallOpts, amount)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ALCA *ALCACaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ALCA.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ALCA *ALCASession) Decimals() (uint8, error) {
	return _ALCA.Contract.Decimals(&_ALCA.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ALCA *ALCACallerSession) Decimals() (uint8, error) {
	return _ALCA.Contract.Decimals(&_ALCA.CallOpts)
}

// GetLegacyTokenAddress is a free data retrieval call binding the contract method 0x035c7099.
//
// Solidity: function getLegacyTokenAddress() view returns(address)
func (_ALCA *ALCACaller) GetLegacyTokenAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ALCA.contract.Call(opts, &out, "getLegacyTokenAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetLegacyTokenAddress is a free data retrieval call binding the contract method 0x035c7099.
//
// Solidity: function getLegacyTokenAddress() view returns(address)
func (_ALCA *ALCASession) GetLegacyTokenAddress() (common.Address, error) {
	return _ALCA.Contract.GetLegacyTokenAddress(&_ALCA.CallOpts)
}

// GetLegacyTokenAddress is a free data retrieval call binding the contract method 0x035c7099.
//
// Solidity: function getLegacyTokenAddress() view returns(address)
func (_ALCA *ALCACallerSession) GetLegacyTokenAddress() (common.Address, error) {
	return _ALCA.Contract.GetLegacyTokenAddress(&_ALCA.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ALCA *ALCACaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _ALCA.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ALCA *ALCASession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ALCA.Contract.GetMetamorphicContractAddress(&_ALCA.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ALCA *ALCACallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ALCA.Contract.GetMetamorphicContractAddress(&_ALCA.CallOpts, _salt, _factory)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ALCA *ALCACaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _ALCA.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ALCA *ALCASession) Name() (string, error) {
	return _ALCA.Contract.Name(&_ALCA.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ALCA *ALCACallerSession) Name() (string, error) {
	return _ALCA.Contract.Name(&_ALCA.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ALCA *ALCACaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _ALCA.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ALCA *ALCASession) Symbol() (string, error) {
	return _ALCA.Contract.Symbol(&_ALCA.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ALCA *ALCACallerSession) Symbol() (string, error) {
	return _ALCA.Contract.Symbol(&_ALCA.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ALCA *ALCACaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ALCA.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ALCA *ALCASession) TotalSupply() (*big.Int, error) {
	return _ALCA.Contract.TotalSupply(&_ALCA.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ALCA *ALCACallerSession) TotalSupply() (*big.Int, error) {
	return _ALCA.Contract.TotalSupply(&_ALCA.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ALCA *ALCATransactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ALCA *ALCASession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.Approve(&_ALCA.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ALCA *ALCATransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.Approve(&_ALCA.TransactOpts, spender, amount)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_ALCA *ALCATransactor) DecreaseAllowance(opts *bind.TransactOpts, spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _ALCA.contract.Transact(opts, "decreaseAllowance", spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_ALCA *ALCASession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.DecreaseAllowance(&_ALCA.TransactOpts, spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_ALCA *ALCATransactorSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.DecreaseAllowance(&_ALCA.TransactOpts, spender, subtractedValue)
}

// ExternalBurn is a paid mutator transaction binding the contract method 0x6f6ebec8.
//
// Solidity: function externalBurn(address from, uint256 amount) returns()
func (_ALCA *ALCATransactor) ExternalBurn(opts *bind.TransactOpts, from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.contract.Transact(opts, "externalBurn", from, amount)
}

// ExternalBurn is a paid mutator transaction binding the contract method 0x6f6ebec8.
//
// Solidity: function externalBurn(address from, uint256 amount) returns()
func (_ALCA *ALCASession) ExternalBurn(from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.ExternalBurn(&_ALCA.TransactOpts, from, amount)
}

// ExternalBurn is a paid mutator transaction binding the contract method 0x6f6ebec8.
//
// Solidity: function externalBurn(address from, uint256 amount) returns()
func (_ALCA *ALCATransactorSession) ExternalBurn(from common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.ExternalBurn(&_ALCA.TransactOpts, from, amount)
}

// ExternalMint is a paid mutator transaction binding the contract method 0x99f98898.
//
// Solidity: function externalMint(address to, uint256 amount) returns()
func (_ALCA *ALCATransactor) ExternalMint(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.contract.Transact(opts, "externalMint", to, amount)
}

// ExternalMint is a paid mutator transaction binding the contract method 0x99f98898.
//
// Solidity: function externalMint(address to, uint256 amount) returns()
func (_ALCA *ALCASession) ExternalMint(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.ExternalMint(&_ALCA.TransactOpts, to, amount)
}

// ExternalMint is a paid mutator transaction binding the contract method 0x99f98898.
//
// Solidity: function externalMint(address to, uint256 amount) returns()
func (_ALCA *ALCATransactorSession) ExternalMint(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.ExternalMint(&_ALCA.TransactOpts, to, amount)
}

// FinishEarlyStage is a paid mutator transaction binding the contract method 0xae424781.
//
// Solidity: function finishEarlyStage() returns()
func (_ALCA *ALCATransactor) FinishEarlyStage(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ALCA.contract.Transact(opts, "finishEarlyStage")
}

// FinishEarlyStage is a paid mutator transaction binding the contract method 0xae424781.
//
// Solidity: function finishEarlyStage() returns()
func (_ALCA *ALCASession) FinishEarlyStage() (*types.Transaction, error) {
	return _ALCA.Contract.FinishEarlyStage(&_ALCA.TransactOpts)
}

// FinishEarlyStage is a paid mutator transaction binding the contract method 0xae424781.
//
// Solidity: function finishEarlyStage() returns()
func (_ALCA *ALCATransactorSession) FinishEarlyStage() (*types.Transaction, error) {
	return _ALCA.Contract.FinishEarlyStage(&_ALCA.TransactOpts)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_ALCA *ALCATransactor) IncreaseAllowance(opts *bind.TransactOpts, spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _ALCA.contract.Transact(opts, "increaseAllowance", spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_ALCA *ALCASession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.IncreaseAllowance(&_ALCA.TransactOpts, spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_ALCA *ALCATransactorSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.IncreaseAllowance(&_ALCA.TransactOpts, spender, addedValue)
}

// Migrate is a paid mutator transaction binding the contract method 0x454b0608.
//
// Solidity: function migrate(uint256 amount) returns(uint256)
func (_ALCA *ALCATransactor) Migrate(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.contract.Transact(opts, "migrate", amount)
}

// Migrate is a paid mutator transaction binding the contract method 0x454b0608.
//
// Solidity: function migrate(uint256 amount) returns(uint256)
func (_ALCA *ALCASession) Migrate(amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.Migrate(&_ALCA.TransactOpts, amount)
}

// Migrate is a paid mutator transaction binding the contract method 0x454b0608.
//
// Solidity: function migrate(uint256 amount) returns(uint256)
func (_ALCA *ALCATransactorSession) Migrate(amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.Migrate(&_ALCA.TransactOpts, amount)
}

// MigrateTo is a paid mutator transaction binding the contract method 0x0d213d31.
//
// Solidity: function migrateTo(address to, uint256 amount) returns(uint256)
func (_ALCA *ALCATransactor) MigrateTo(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.contract.Transact(opts, "migrateTo", to, amount)
}

// MigrateTo is a paid mutator transaction binding the contract method 0x0d213d31.
//
// Solidity: function migrateTo(address to, uint256 amount) returns(uint256)
func (_ALCA *ALCASession) MigrateTo(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.MigrateTo(&_ALCA.TransactOpts, to, amount)
}

// MigrateTo is a paid mutator transaction binding the contract method 0x0d213d31.
//
// Solidity: function migrateTo(address to, uint256 amount) returns(uint256)
func (_ALCA *ALCATransactorSession) MigrateTo(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.MigrateTo(&_ALCA.TransactOpts, to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_ALCA *ALCATransactor) Transfer(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.contract.Transact(opts, "transfer", to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_ALCA *ALCASession) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.Transfer(&_ALCA.TransactOpts, to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_ALCA *ALCATransactorSession) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.Transfer(&_ALCA.TransactOpts, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_ALCA *ALCATransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.contract.Transact(opts, "transferFrom", from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_ALCA *ALCASession) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.TransferFrom(&_ALCA.TransactOpts, from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_ALCA *ALCATransactorSession) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCA.Contract.TransferFrom(&_ALCA.TransactOpts, from, to, amount)
}

// ALCAApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the ALCA contract.
type ALCAApprovalIterator struct {
	Event *ALCAApproval // Event containing the contract specifics and raw log

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
func (it *ALCAApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ALCAApproval)
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
		it.Event = new(ALCAApproval)
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
func (it *ALCAApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ALCAApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ALCAApproval represents a Approval event raised by the ALCA contract.
type ALCAApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ALCA *ALCAFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*ALCAApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ALCA.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &ALCAApprovalIterator{contract: _ALCA.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ALCA *ALCAFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ALCAApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ALCA.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ALCAApproval)
				if err := _ALCA.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_ALCA *ALCAFilterer) ParseApproval(log types.Log) (*ALCAApproval, error) {
	event := new(ALCAApproval)
	if err := _ALCA.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ALCATransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the ALCA contract.
type ALCATransferIterator struct {
	Event *ALCATransfer // Event containing the contract specifics and raw log

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
func (it *ALCATransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ALCATransfer)
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
		it.Event = new(ALCATransfer)
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
func (it *ALCATransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ALCATransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ALCATransfer represents a Transfer event raised by the ALCA contract.
type ALCATransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ALCA *ALCAFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*ALCATransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ALCA.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &ALCATransferIterator{contract: _ALCA.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ALCA *ALCAFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ALCATransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ALCA.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ALCATransfer)
				if err := _ALCA.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_ALCA *ALCAFilterer) ParseTransfer(log types.Log) (*ALCATransfer, error) {
	event := new(ALCATransfer)
	if err := _ALCA.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
