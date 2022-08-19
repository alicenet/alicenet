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

// Deposit is an auto generated low-level Go binding around an user-defined struct.
type Deposit struct {
	AccountType uint8
	Account     common.Address
	Value       *big.Int
}

// BTokenMetaData contains all meta data concerning the BToken contract.
var BTokenMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"supply\",\"type\":\"uint256\"}],\"name\":\"BurnAmountExceedsSupply\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"toAddress\",\"type\":\"address\"}],\"name\":\"ContractsDisallowedDeposits\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DepositAmountZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DepositBurnFail\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"name\":\"InexistentRouterContract\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"InsufficientEth\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"}],\"name\":\"InsufficientFee\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"contractBalance\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"poolBalance\",\"type\":\"uint256\"}],\"name\":\"InvalidBalance\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"InvalidBurnAmount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"depositID\",\"type\":\"uint256\"}],\"name\":\"InvalidDepositId\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"MinimumBurnNotMet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"MinimumMintNotMet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimumValue\",\"type\":\"uint256\"}],\"name\":\"MinimumValueNotMet\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MutexLocked\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyDistribution\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyFactory\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositID\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"accountType\",\"type\":\"uint8\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DepositReceived\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minEth_\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minEth_\",\"type\":\"uint256\"}],\"name\":\"burnTo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"accountType_\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"bridgeVersion\",\"type\":\"uint16\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"depositTokensOnBridges\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"numBTK_\",\"type\":\"uint256\"}],\"name\":\"destroyBTokens\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"distribute\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"depositID\",\"type\":\"uint256\"}],\"name\":\"getDeposit\",\"outputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"accountType\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"internalType\":\"structDeposit\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDepositID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"poolBalance_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalSupply_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numBTK_\",\"type\":\"uint256\"}],\"name\":\"getEthFromBTokensBurn\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"totalSupply_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numBTK_\",\"type\":\"uint256\"}],\"name\":\"getEthToMintBTokens\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"numBTK_\",\"type\":\"uint256\"}],\"name\":\"getLatestEthFromBTokensBurn\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"numBTK_\",\"type\":\"uint256\"}],\"name\":\"getLatestEthToMintBTokens\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth_\",\"type\":\"uint256\"}],\"name\":\"getLatestMintedBTokensFromEth\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMarketSpread\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"poolBalance_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numEth_\",\"type\":\"uint256\"}],\"name\":\"getMintedBTokensFromEth\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPoolBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTotalBTokensDeposited\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getYield\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"minBTK_\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numBTK\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"accountType_\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"minBTK_\",\"type\":\"uint256\"}],\"name\":\"mintDeposit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"minBTK_\",\"type\":\"uint256\"}],\"name\":\"mintTo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numBTK\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"accountType_\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"}],\"name\":\"virtualMintDeposit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// BTokenABI is the input ABI used to generate the binding from.
// Deprecated: Use BTokenMetaData.ABI instead.
var BTokenABI = BTokenMetaData.ABI

// BToken is an auto generated Go binding around an Ethereum contract.
type BToken struct {
	BTokenCaller     // Read-only binding to the contract
	BTokenTransactor // Write-only binding to the contract
	BTokenFilterer   // Log filterer for contract events
}

// BTokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type BTokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BTokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BTokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BTokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BTokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BTokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BTokenSession struct {
	Contract     *BToken           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BTokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BTokenCallerSession struct {
	Contract *BTokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// BTokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BTokenTransactorSession struct {
	Contract     *BTokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BTokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type BTokenRaw struct {
	Contract *BToken // Generic contract binding to access the raw methods on
}

// BTokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BTokenCallerRaw struct {
	Contract *BTokenCaller // Generic read-only contract binding to access the raw methods on
}

// BTokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BTokenTransactorRaw struct {
	Contract *BTokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBToken creates a new instance of BToken, bound to a specific deployed contract.
func NewBToken(address common.Address, backend bind.ContractBackend) (*BToken, error) {
	contract, err := bindBToken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BToken{BTokenCaller: BTokenCaller{contract: contract}, BTokenTransactor: BTokenTransactor{contract: contract}, BTokenFilterer: BTokenFilterer{contract: contract}}, nil
}

// NewBTokenCaller creates a new read-only instance of BToken, bound to a specific deployed contract.
func NewBTokenCaller(address common.Address, caller bind.ContractCaller) (*BTokenCaller, error) {
	contract, err := bindBToken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BTokenCaller{contract: contract}, nil
}

// NewBTokenTransactor creates a new write-only instance of BToken, bound to a specific deployed contract.
func NewBTokenTransactor(address common.Address, transactor bind.ContractTransactor) (*BTokenTransactor, error) {
	contract, err := bindBToken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BTokenTransactor{contract: contract}, nil
}

// NewBTokenFilterer creates a new log filterer instance of BToken, bound to a specific deployed contract.
func NewBTokenFilterer(address common.Address, filterer bind.ContractFilterer) (*BTokenFilterer, error) {
	contract, err := bindBToken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BTokenFilterer{contract: contract}, nil
}

// bindBToken binds a generic wrapper to an already deployed contract.
func bindBToken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BTokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BToken *BTokenRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BToken.Contract.BTokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BToken *BTokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BToken.Contract.BTokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BToken *BTokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BToken.Contract.BTokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BToken *BTokenCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BToken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BToken *BTokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BToken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BToken *BTokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BToken.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_BToken *BTokenCaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_BToken *BTokenSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _BToken.Contract.Allowance(&_BToken.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_BToken *BTokenCallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _BToken.Contract.Allowance(&_BToken.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_BToken *BTokenCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_BToken *BTokenSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _BToken.Contract.BalanceOf(&_BToken.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_BToken *BTokenCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _BToken.Contract.BalanceOf(&_BToken.CallOpts, account)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_BToken *BTokenCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_BToken *BTokenSession) Decimals() (uint8, error) {
	return _BToken.Contract.Decimals(&_BToken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_BToken *BTokenCallerSession) Decimals() (uint8, error) {
	return _BToken.Contract.Decimals(&_BToken.CallOpts)
}

// GetDeposit is a free data retrieval call binding the contract method 0x9f9fb968.
//
// Solidity: function getDeposit(uint256 depositID) view returns((uint8,address,uint256))
func (_BToken *BTokenCaller) GetDeposit(opts *bind.CallOpts, depositID *big.Int) (Deposit, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "getDeposit", depositID)

	if err != nil {
		return *new(Deposit), err
	}

	out0 := *abi.ConvertType(out[0], new(Deposit)).(*Deposit)

	return out0, err

}

// GetDeposit is a free data retrieval call binding the contract method 0x9f9fb968.
//
// Solidity: function getDeposit(uint256 depositID) view returns((uint8,address,uint256))
func (_BToken *BTokenSession) GetDeposit(depositID *big.Int) (Deposit, error) {
	return _BToken.Contract.GetDeposit(&_BToken.CallOpts, depositID)
}

// GetDeposit is a free data retrieval call binding the contract method 0x9f9fb968.
//
// Solidity: function getDeposit(uint256 depositID) view returns((uint8,address,uint256))
func (_BToken *BTokenCallerSession) GetDeposit(depositID *big.Int) (Deposit, error) {
	return _BToken.Contract.GetDeposit(&_BToken.CallOpts, depositID)
}

// GetDepositID is a free data retrieval call binding the contract method 0x94f70c8a.
//
// Solidity: function getDepositID() view returns(uint256)
func (_BToken *BTokenCaller) GetDepositID(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "getDepositID")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetDepositID is a free data retrieval call binding the contract method 0x94f70c8a.
//
// Solidity: function getDepositID() view returns(uint256)
func (_BToken *BTokenSession) GetDepositID() (*big.Int, error) {
	return _BToken.Contract.GetDepositID(&_BToken.CallOpts)
}

// GetDepositID is a free data retrieval call binding the contract method 0x94f70c8a.
//
// Solidity: function getDepositID() view returns(uint256)
func (_BToken *BTokenCallerSession) GetDepositID() (*big.Int, error) {
	return _BToken.Contract.GetDepositID(&_BToken.CallOpts)
}

// GetEthFromBTokensBurn is a free data retrieval call binding the contract method 0xc3559a18.
//
// Solidity: function getEthFromBTokensBurn(uint256 poolBalance_, uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
func (_BToken *BTokenCaller) GetEthFromBTokensBurn(opts *bind.CallOpts, poolBalance_ *big.Int, totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "getEthFromBTokensBurn", poolBalance_, totalSupply_, numBTK_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetEthFromBTokensBurn is a free data retrieval call binding the contract method 0xc3559a18.
//
// Solidity: function getEthFromBTokensBurn(uint256 poolBalance_, uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
func (_BToken *BTokenSession) GetEthFromBTokensBurn(poolBalance_ *big.Int, totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error) {
	return _BToken.Contract.GetEthFromBTokensBurn(&_BToken.CallOpts, poolBalance_, totalSupply_, numBTK_)
}

// GetEthFromBTokensBurn is a free data retrieval call binding the contract method 0xc3559a18.
//
// Solidity: function getEthFromBTokensBurn(uint256 poolBalance_, uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
func (_BToken *BTokenCallerSession) GetEthFromBTokensBurn(poolBalance_ *big.Int, totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error) {
	return _BToken.Contract.GetEthFromBTokensBurn(&_BToken.CallOpts, poolBalance_, totalSupply_, numBTK_)
}

// GetEthToMintBTokens is a free data retrieval call binding the contract method 0x06f1ad83.
//
// Solidity: function getEthToMintBTokens(uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
func (_BToken *BTokenCaller) GetEthToMintBTokens(opts *bind.CallOpts, totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "getEthToMintBTokens", totalSupply_, numBTK_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetEthToMintBTokens is a free data retrieval call binding the contract method 0x06f1ad83.
//
// Solidity: function getEthToMintBTokens(uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
func (_BToken *BTokenSession) GetEthToMintBTokens(totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error) {
	return _BToken.Contract.GetEthToMintBTokens(&_BToken.CallOpts, totalSupply_, numBTK_)
}

// GetEthToMintBTokens is a free data retrieval call binding the contract method 0x06f1ad83.
//
// Solidity: function getEthToMintBTokens(uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
func (_BToken *BTokenCallerSession) GetEthToMintBTokens(totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error) {
	return _BToken.Contract.GetEthToMintBTokens(&_BToken.CallOpts, totalSupply_, numBTK_)
}

// GetLatestEthFromBTokensBurn is a free data retrieval call binding the contract method 0xc425d3a0.
//
// Solidity: function getLatestEthFromBTokensBurn(uint256 numBTK_) view returns(uint256 numEth)
func (_BToken *BTokenCaller) GetLatestEthFromBTokensBurn(opts *bind.CallOpts, numBTK_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "getLatestEthFromBTokensBurn", numBTK_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLatestEthFromBTokensBurn is a free data retrieval call binding the contract method 0xc425d3a0.
//
// Solidity: function getLatestEthFromBTokensBurn(uint256 numBTK_) view returns(uint256 numEth)
func (_BToken *BTokenSession) GetLatestEthFromBTokensBurn(numBTK_ *big.Int) (*big.Int, error) {
	return _BToken.Contract.GetLatestEthFromBTokensBurn(&_BToken.CallOpts, numBTK_)
}

// GetLatestEthFromBTokensBurn is a free data retrieval call binding the contract method 0xc425d3a0.
//
// Solidity: function getLatestEthFromBTokensBurn(uint256 numBTK_) view returns(uint256 numEth)
func (_BToken *BTokenCallerSession) GetLatestEthFromBTokensBurn(numBTK_ *big.Int) (*big.Int, error) {
	return _BToken.Contract.GetLatestEthFromBTokensBurn(&_BToken.CallOpts, numBTK_)
}

// GetLatestEthToMintBTokens is a free data retrieval call binding the contract method 0xbcd6abb4.
//
// Solidity: function getLatestEthToMintBTokens(uint256 numBTK_) view returns(uint256 numEth)
func (_BToken *BTokenCaller) GetLatestEthToMintBTokens(opts *bind.CallOpts, numBTK_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "getLatestEthToMintBTokens", numBTK_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLatestEthToMintBTokens is a free data retrieval call binding the contract method 0xbcd6abb4.
//
// Solidity: function getLatestEthToMintBTokens(uint256 numBTK_) view returns(uint256 numEth)
func (_BToken *BTokenSession) GetLatestEthToMintBTokens(numBTK_ *big.Int) (*big.Int, error) {
	return _BToken.Contract.GetLatestEthToMintBTokens(&_BToken.CallOpts, numBTK_)
}

// GetLatestEthToMintBTokens is a free data retrieval call binding the contract method 0xbcd6abb4.
//
// Solidity: function getLatestEthToMintBTokens(uint256 numBTK_) view returns(uint256 numEth)
func (_BToken *BTokenCallerSession) GetLatestEthToMintBTokens(numBTK_ *big.Int) (*big.Int, error) {
	return _BToken.Contract.GetLatestEthToMintBTokens(&_BToken.CallOpts, numBTK_)
}

// GetLatestMintedBTokensFromEth is a free data retrieval call binding the contract method 0x5878677c.
//
// Solidity: function getLatestMintedBTokensFromEth(uint256 numEth_) view returns(uint256)
func (_BToken *BTokenCaller) GetLatestMintedBTokensFromEth(opts *bind.CallOpts, numEth_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "getLatestMintedBTokensFromEth", numEth_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLatestMintedBTokensFromEth is a free data retrieval call binding the contract method 0x5878677c.
//
// Solidity: function getLatestMintedBTokensFromEth(uint256 numEth_) view returns(uint256)
func (_BToken *BTokenSession) GetLatestMintedBTokensFromEth(numEth_ *big.Int) (*big.Int, error) {
	return _BToken.Contract.GetLatestMintedBTokensFromEth(&_BToken.CallOpts, numEth_)
}

// GetLatestMintedBTokensFromEth is a free data retrieval call binding the contract method 0x5878677c.
//
// Solidity: function getLatestMintedBTokensFromEth(uint256 numEth_) view returns(uint256)
func (_BToken *BTokenCallerSession) GetLatestMintedBTokensFromEth(numEth_ *big.Int) (*big.Int, error) {
	return _BToken.Contract.GetLatestMintedBTokensFromEth(&_BToken.CallOpts, numEth_)
}

// GetMarketSpread is a free data retrieval call binding the contract method 0x086cfefd.
//
// Solidity: function getMarketSpread() pure returns(uint256)
func (_BToken *BTokenCaller) GetMarketSpread(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "getMarketSpread")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMarketSpread is a free data retrieval call binding the contract method 0x086cfefd.
//
// Solidity: function getMarketSpread() pure returns(uint256)
func (_BToken *BTokenSession) GetMarketSpread() (*big.Int, error) {
	return _BToken.Contract.GetMarketSpread(&_BToken.CallOpts)
}

// GetMarketSpread is a free data retrieval call binding the contract method 0x086cfefd.
//
// Solidity: function getMarketSpread() pure returns(uint256)
func (_BToken *BTokenCallerSession) GetMarketSpread() (*big.Int, error) {
	return _BToken.Contract.GetMarketSpread(&_BToken.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_BToken *BTokenCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_BToken *BTokenSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _BToken.Contract.GetMetamorphicContractAddress(&_BToken.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_BToken *BTokenCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _BToken.Contract.GetMetamorphicContractAddress(&_BToken.CallOpts, _salt, _factory)
}

// GetMintedBTokensFromEth is a free data retrieval call binding the contract method 0x823fa03a.
//
// Solidity: function getMintedBTokensFromEth(uint256 poolBalance_, uint256 numEth_) pure returns(uint256)
func (_BToken *BTokenCaller) GetMintedBTokensFromEth(opts *bind.CallOpts, poolBalance_ *big.Int, numEth_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "getMintedBTokensFromEth", poolBalance_, numEth_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMintedBTokensFromEth is a free data retrieval call binding the contract method 0x823fa03a.
//
// Solidity: function getMintedBTokensFromEth(uint256 poolBalance_, uint256 numEth_) pure returns(uint256)
func (_BToken *BTokenSession) GetMintedBTokensFromEth(poolBalance_ *big.Int, numEth_ *big.Int) (*big.Int, error) {
	return _BToken.Contract.GetMintedBTokensFromEth(&_BToken.CallOpts, poolBalance_, numEth_)
}

// GetMintedBTokensFromEth is a free data retrieval call binding the contract method 0x823fa03a.
//
// Solidity: function getMintedBTokensFromEth(uint256 poolBalance_, uint256 numEth_) pure returns(uint256)
func (_BToken *BTokenCallerSession) GetMintedBTokensFromEth(poolBalance_ *big.Int, numEth_ *big.Int) (*big.Int, error) {
	return _BToken.Contract.GetMintedBTokensFromEth(&_BToken.CallOpts, poolBalance_, numEth_)
}

// GetPoolBalance is a free data retrieval call binding the contract method 0xabd70aa2.
//
// Solidity: function getPoolBalance() view returns(uint256)
func (_BToken *BTokenCaller) GetPoolBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "getPoolBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPoolBalance is a free data retrieval call binding the contract method 0xabd70aa2.
//
// Solidity: function getPoolBalance() view returns(uint256)
func (_BToken *BTokenSession) GetPoolBalance() (*big.Int, error) {
	return _BToken.Contract.GetPoolBalance(&_BToken.CallOpts)
}

// GetPoolBalance is a free data retrieval call binding the contract method 0xabd70aa2.
//
// Solidity: function getPoolBalance() view returns(uint256)
func (_BToken *BTokenCallerSession) GetPoolBalance() (*big.Int, error) {
	return _BToken.Contract.GetPoolBalance(&_BToken.CallOpts)
}

// GetTotalBTokensDeposited is a free data retrieval call binding the contract method 0x5ecef3af.
//
// Solidity: function getTotalBTokensDeposited() view returns(uint256)
func (_BToken *BTokenCaller) GetTotalBTokensDeposited(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "getTotalBTokensDeposited")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalBTokensDeposited is a free data retrieval call binding the contract method 0x5ecef3af.
//
// Solidity: function getTotalBTokensDeposited() view returns(uint256)
func (_BToken *BTokenSession) GetTotalBTokensDeposited() (*big.Int, error) {
	return _BToken.Contract.GetTotalBTokensDeposited(&_BToken.CallOpts)
}

// GetTotalBTokensDeposited is a free data retrieval call binding the contract method 0x5ecef3af.
//
// Solidity: function getTotalBTokensDeposited() view returns(uint256)
func (_BToken *BTokenCallerSession) GetTotalBTokensDeposited() (*big.Int, error) {
	return _BToken.Contract.GetTotalBTokensDeposited(&_BToken.CallOpts)
}

// GetYield is a free data retrieval call binding the contract method 0x7c262871.
//
// Solidity: function getYield() view returns(uint256)
func (_BToken *BTokenCaller) GetYield(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "getYield")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetYield is a free data retrieval call binding the contract method 0x7c262871.
//
// Solidity: function getYield() view returns(uint256)
func (_BToken *BTokenSession) GetYield() (*big.Int, error) {
	return _BToken.Contract.GetYield(&_BToken.CallOpts)
}

// GetYield is a free data retrieval call binding the contract method 0x7c262871.
//
// Solidity: function getYield() view returns(uint256)
func (_BToken *BTokenCallerSession) GetYield() (*big.Int, error) {
	return _BToken.Contract.GetYield(&_BToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_BToken *BTokenCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_BToken *BTokenSession) Name() (string, error) {
	return _BToken.Contract.Name(&_BToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_BToken *BTokenCallerSession) Name() (string, error) {
	return _BToken.Contract.Name(&_BToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_BToken *BTokenCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_BToken *BTokenSession) Symbol() (string, error) {
	return _BToken.Contract.Symbol(&_BToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_BToken *BTokenCallerSession) Symbol() (string, error) {
	return _BToken.Contract.Symbol(&_BToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_BToken *BTokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BToken.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_BToken *BTokenSession) TotalSupply() (*big.Int, error) {
	return _BToken.Contract.TotalSupply(&_BToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_BToken *BTokenCallerSession) TotalSupply() (*big.Int, error) {
	return _BToken.Contract.TotalSupply(&_BToken.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_BToken *BTokenTransactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_BToken *BTokenSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.Approve(&_BToken.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_BToken *BTokenTransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.Approve(&_BToken.TransactOpts, spender, amount)
}

// Burn is a paid mutator transaction binding the contract method 0xb390c0ab.
//
// Solidity: function burn(uint256 amount_, uint256 minEth_) returns(uint256 numEth)
func (_BToken *BTokenTransactor) Burn(opts *bind.TransactOpts, amount_ *big.Int, minEth_ *big.Int) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "burn", amount_, minEth_)
}

// Burn is a paid mutator transaction binding the contract method 0xb390c0ab.
//
// Solidity: function burn(uint256 amount_, uint256 minEth_) returns(uint256 numEth)
func (_BToken *BTokenSession) Burn(amount_ *big.Int, minEth_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.Burn(&_BToken.TransactOpts, amount_, minEth_)
}

// Burn is a paid mutator transaction binding the contract method 0xb390c0ab.
//
// Solidity: function burn(uint256 amount_, uint256 minEth_) returns(uint256 numEth)
func (_BToken *BTokenTransactorSession) Burn(amount_ *big.Int, minEth_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.Burn(&_BToken.TransactOpts, amount_, minEth_)
}

// BurnTo is a paid mutator transaction binding the contract method 0x9b057203.
//
// Solidity: function burnTo(address to_, uint256 amount_, uint256 minEth_) returns(uint256 numEth)
func (_BToken *BTokenTransactor) BurnTo(opts *bind.TransactOpts, to_ common.Address, amount_ *big.Int, minEth_ *big.Int) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "burnTo", to_, amount_, minEth_)
}

// BurnTo is a paid mutator transaction binding the contract method 0x9b057203.
//
// Solidity: function burnTo(address to_, uint256 amount_, uint256 minEth_) returns(uint256 numEth)
func (_BToken *BTokenSession) BurnTo(to_ common.Address, amount_ *big.Int, minEth_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.BurnTo(&_BToken.TransactOpts, to_, amount_, minEth_)
}

// BurnTo is a paid mutator transaction binding the contract method 0x9b057203.
//
// Solidity: function burnTo(address to_, uint256 amount_, uint256 minEth_) returns(uint256 numEth)
func (_BToken *BTokenTransactorSession) BurnTo(to_ common.Address, amount_ *big.Int, minEth_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.BurnTo(&_BToken.TransactOpts, to_, amount_, minEth_)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_BToken *BTokenTransactor) DecreaseAllowance(opts *bind.TransactOpts, spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "decreaseAllowance", spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_BToken *BTokenSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.DecreaseAllowance(&_BToken.TransactOpts, spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_BToken *BTokenTransactorSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.DecreaseAllowance(&_BToken.TransactOpts, spender, subtractedValue)
}

// Deposit is a paid mutator transaction binding the contract method 0x00838172.
//
// Solidity: function deposit(uint8 accountType_, address to_, uint256 amount_) returns(uint256)
func (_BToken *BTokenTransactor) Deposit(opts *bind.TransactOpts, accountType_ uint8, to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "deposit", accountType_, to_, amount_)
}

// Deposit is a paid mutator transaction binding the contract method 0x00838172.
//
// Solidity: function deposit(uint8 accountType_, address to_, uint256 amount_) returns(uint256)
func (_BToken *BTokenSession) Deposit(accountType_ uint8, to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.Deposit(&_BToken.TransactOpts, accountType_, to_, amount_)
}

// Deposit is a paid mutator transaction binding the contract method 0x00838172.
//
// Solidity: function deposit(uint8 accountType_, address to_, uint256 amount_) returns(uint256)
func (_BToken *BTokenTransactorSession) Deposit(accountType_ uint8, to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.Deposit(&_BToken.TransactOpts, accountType_, to_, amount_)
}

// DepositTokensOnBridges is a paid mutator transaction binding the contract method 0xddeae0b3.
//
// Solidity: function depositTokensOnBridges(uint16 bridgeVersion, bytes data) payable returns()
func (_BToken *BTokenTransactor) DepositTokensOnBridges(opts *bind.TransactOpts, bridgeVersion uint16, data []byte) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "depositTokensOnBridges", bridgeVersion, data)
}

// DepositTokensOnBridges is a paid mutator transaction binding the contract method 0xddeae0b3.
//
// Solidity: function depositTokensOnBridges(uint16 bridgeVersion, bytes data) payable returns()
func (_BToken *BTokenSession) DepositTokensOnBridges(bridgeVersion uint16, data []byte) (*types.Transaction, error) {
	return _BToken.Contract.DepositTokensOnBridges(&_BToken.TransactOpts, bridgeVersion, data)
}

// DepositTokensOnBridges is a paid mutator transaction binding the contract method 0xddeae0b3.
//
// Solidity: function depositTokensOnBridges(uint16 bridgeVersion, bytes data) payable returns()
func (_BToken *BTokenTransactorSession) DepositTokensOnBridges(bridgeVersion uint16, data []byte) (*types.Transaction, error) {
	return _BToken.Contract.DepositTokensOnBridges(&_BToken.TransactOpts, bridgeVersion, data)
}

// DestroyBTokens is a paid mutator transaction binding the contract method 0x2dc6b024.
//
// Solidity: function destroyBTokens(uint256 numBTK_) returns(bool)
func (_BToken *BTokenTransactor) DestroyBTokens(opts *bind.TransactOpts, numBTK_ *big.Int) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "destroyBTokens", numBTK_)
}

// DestroyBTokens is a paid mutator transaction binding the contract method 0x2dc6b024.
//
// Solidity: function destroyBTokens(uint256 numBTK_) returns(bool)
func (_BToken *BTokenSession) DestroyBTokens(numBTK_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.DestroyBTokens(&_BToken.TransactOpts, numBTK_)
}

// DestroyBTokens is a paid mutator transaction binding the contract method 0x2dc6b024.
//
// Solidity: function destroyBTokens(uint256 numBTK_) returns(bool)
func (_BToken *BTokenTransactorSession) DestroyBTokens(numBTK_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.DestroyBTokens(&_BToken.TransactOpts, numBTK_)
}

// Distribute is a paid mutator transaction binding the contract method 0xe4fc6b6d.
//
// Solidity: function distribute() returns(bool)
func (_BToken *BTokenTransactor) Distribute(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "distribute")
}

// Distribute is a paid mutator transaction binding the contract method 0xe4fc6b6d.
//
// Solidity: function distribute() returns(bool)
func (_BToken *BTokenSession) Distribute() (*types.Transaction, error) {
	return _BToken.Contract.Distribute(&_BToken.TransactOpts)
}

// Distribute is a paid mutator transaction binding the contract method 0xe4fc6b6d.
//
// Solidity: function distribute() returns(bool)
func (_BToken *BTokenTransactorSession) Distribute() (*types.Transaction, error) {
	return _BToken.Contract.Distribute(&_BToken.TransactOpts)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_BToken *BTokenTransactor) IncreaseAllowance(opts *bind.TransactOpts, spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "increaseAllowance", spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_BToken *BTokenSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.IncreaseAllowance(&_BToken.TransactOpts, spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_BToken *BTokenTransactorSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.IncreaseAllowance(&_BToken.TransactOpts, spender, addedValue)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_BToken *BTokenTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_BToken *BTokenSession) Initialize() (*types.Transaction, error) {
	return _BToken.Contract.Initialize(&_BToken.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_BToken *BTokenTransactorSession) Initialize() (*types.Transaction, error) {
	return _BToken.Contract.Initialize(&_BToken.TransactOpts)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 minBTK_) payable returns(uint256 numBTK)
func (_BToken *BTokenTransactor) Mint(opts *bind.TransactOpts, minBTK_ *big.Int) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "mint", minBTK_)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 minBTK_) payable returns(uint256 numBTK)
func (_BToken *BTokenSession) Mint(minBTK_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.Mint(&_BToken.TransactOpts, minBTK_)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 minBTK_) payable returns(uint256 numBTK)
func (_BToken *BTokenTransactorSession) Mint(minBTK_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.Mint(&_BToken.TransactOpts, minBTK_)
}

// MintDeposit is a paid mutator transaction binding the contract method 0x4f232628.
//
// Solidity: function mintDeposit(uint8 accountType_, address to_, uint256 minBTK_) payable returns(uint256)
func (_BToken *BTokenTransactor) MintDeposit(opts *bind.TransactOpts, accountType_ uint8, to_ common.Address, minBTK_ *big.Int) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "mintDeposit", accountType_, to_, minBTK_)
}

// MintDeposit is a paid mutator transaction binding the contract method 0x4f232628.
//
// Solidity: function mintDeposit(uint8 accountType_, address to_, uint256 minBTK_) payable returns(uint256)
func (_BToken *BTokenSession) MintDeposit(accountType_ uint8, to_ common.Address, minBTK_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.MintDeposit(&_BToken.TransactOpts, accountType_, to_, minBTK_)
}

// MintDeposit is a paid mutator transaction binding the contract method 0x4f232628.
//
// Solidity: function mintDeposit(uint8 accountType_, address to_, uint256 minBTK_) payable returns(uint256)
func (_BToken *BTokenTransactorSession) MintDeposit(accountType_ uint8, to_ common.Address, minBTK_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.MintDeposit(&_BToken.TransactOpts, accountType_, to_, minBTK_)
}

// MintTo is a paid mutator transaction binding the contract method 0x449a52f8.
//
// Solidity: function mintTo(address to_, uint256 minBTK_) payable returns(uint256 numBTK)
func (_BToken *BTokenTransactor) MintTo(opts *bind.TransactOpts, to_ common.Address, minBTK_ *big.Int) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "mintTo", to_, minBTK_)
}

// MintTo is a paid mutator transaction binding the contract method 0x449a52f8.
//
// Solidity: function mintTo(address to_, uint256 minBTK_) payable returns(uint256 numBTK)
func (_BToken *BTokenSession) MintTo(to_ common.Address, minBTK_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.MintTo(&_BToken.TransactOpts, to_, minBTK_)
}

// MintTo is a paid mutator transaction binding the contract method 0x449a52f8.
//
// Solidity: function mintTo(address to_, uint256 minBTK_) payable returns(uint256 numBTK)
func (_BToken *BTokenTransactorSession) MintTo(to_ common.Address, minBTK_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.MintTo(&_BToken.TransactOpts, to_, minBTK_)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_BToken *BTokenTransactor) Transfer(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "transfer", to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_BToken *BTokenSession) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.Transfer(&_BToken.TransactOpts, to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_BToken *BTokenTransactorSession) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.Transfer(&_BToken.TransactOpts, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_BToken *BTokenTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "transferFrom", from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_BToken *BTokenSession) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.TransferFrom(&_BToken.TransactOpts, from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_BToken *BTokenTransactorSession) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.TransferFrom(&_BToken.TransactOpts, from, to, amount)
}

// VirtualMintDeposit is a paid mutator transaction binding the contract method 0x92178278.
//
// Solidity: function virtualMintDeposit(uint8 accountType_, address to_, uint256 amount_) returns(uint256)
func (_BToken *BTokenTransactor) VirtualMintDeposit(opts *bind.TransactOpts, accountType_ uint8, to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _BToken.contract.Transact(opts, "virtualMintDeposit", accountType_, to_, amount_)
}

// VirtualMintDeposit is a paid mutator transaction binding the contract method 0x92178278.
//
// Solidity: function virtualMintDeposit(uint8 accountType_, address to_, uint256 amount_) returns(uint256)
func (_BToken *BTokenSession) VirtualMintDeposit(accountType_ uint8, to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.VirtualMintDeposit(&_BToken.TransactOpts, accountType_, to_, amount_)
}

// VirtualMintDeposit is a paid mutator transaction binding the contract method 0x92178278.
//
// Solidity: function virtualMintDeposit(uint8 accountType_, address to_, uint256 amount_) returns(uint256)
func (_BToken *BTokenTransactorSession) VirtualMintDeposit(accountType_ uint8, to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _BToken.Contract.VirtualMintDeposit(&_BToken.TransactOpts, accountType_, to_, amount_)
}

// BTokenApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the BToken contract.
type BTokenApprovalIterator struct {
	Event *BTokenApproval // Event containing the contract specifics and raw log

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
func (it *BTokenApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BTokenApproval)
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
		it.Event = new(BTokenApproval)
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
func (it *BTokenApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BTokenApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BTokenApproval represents a Approval event raised by the BToken contract.
type BTokenApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_BToken *BTokenFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*BTokenApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _BToken.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &BTokenApprovalIterator{contract: _BToken.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_BToken *BTokenFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *BTokenApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _BToken.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BTokenApproval)
				if err := _BToken.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_BToken *BTokenFilterer) ParseApproval(log types.Log) (*BTokenApproval, error) {
	event := new(BTokenApproval)
	if err := _BToken.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BTokenDepositReceivedIterator is returned from FilterDepositReceived and is used to iterate over the raw logs and unpacked data for DepositReceived events raised by the BToken contract.
type BTokenDepositReceivedIterator struct {
	Event *BTokenDepositReceived // Event containing the contract specifics and raw log

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
func (it *BTokenDepositReceivedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BTokenDepositReceived)
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
		it.Event = new(BTokenDepositReceived)
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
func (it *BTokenDepositReceivedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BTokenDepositReceivedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BTokenDepositReceived represents a DepositReceived event raised by the BToken contract.
type BTokenDepositReceived struct {
	DepositID   *big.Int
	AccountType uint8
	Depositor   common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterDepositReceived is a free log retrieval operation binding the contract event 0x9d291e7244fa9bf5a85ec47d5a52012ccc92231b41c94308155ad2702eef9d4d.
//
// Solidity: event DepositReceived(uint256 indexed depositID, uint8 indexed accountType, address indexed depositor, uint256 amount)
func (_BToken *BTokenFilterer) FilterDepositReceived(opts *bind.FilterOpts, depositID []*big.Int, accountType []uint8, depositor []common.Address) (*BTokenDepositReceivedIterator, error) {

	var depositIDRule []interface{}
	for _, depositIDItem := range depositID {
		depositIDRule = append(depositIDRule, depositIDItem)
	}
	var accountTypeRule []interface{}
	for _, accountTypeItem := range accountType {
		accountTypeRule = append(accountTypeRule, accountTypeItem)
	}
	var depositorRule []interface{}
	for _, depositorItem := range depositor {
		depositorRule = append(depositorRule, depositorItem)
	}

	logs, sub, err := _BToken.contract.FilterLogs(opts, "DepositReceived", depositIDRule, accountTypeRule, depositorRule)
	if err != nil {
		return nil, err
	}
	return &BTokenDepositReceivedIterator{contract: _BToken.contract, event: "DepositReceived", logs: logs, sub: sub}, nil
}

// WatchDepositReceived is a free log subscription operation binding the contract event 0x9d291e7244fa9bf5a85ec47d5a52012ccc92231b41c94308155ad2702eef9d4d.
//
// Solidity: event DepositReceived(uint256 indexed depositID, uint8 indexed accountType, address indexed depositor, uint256 amount)
func (_BToken *BTokenFilterer) WatchDepositReceived(opts *bind.WatchOpts, sink chan<- *BTokenDepositReceived, depositID []*big.Int, accountType []uint8, depositor []common.Address) (event.Subscription, error) {

	var depositIDRule []interface{}
	for _, depositIDItem := range depositID {
		depositIDRule = append(depositIDRule, depositIDItem)
	}
	var accountTypeRule []interface{}
	for _, accountTypeItem := range accountType {
		accountTypeRule = append(accountTypeRule, accountTypeItem)
	}
	var depositorRule []interface{}
	for _, depositorItem := range depositor {
		depositorRule = append(depositorRule, depositorItem)
	}

	logs, sub, err := _BToken.contract.WatchLogs(opts, "DepositReceived", depositIDRule, accountTypeRule, depositorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BTokenDepositReceived)
				if err := _BToken.contract.UnpackLog(event, "DepositReceived", log); err != nil {
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

// ParseDepositReceived is a log parse operation binding the contract event 0x9d291e7244fa9bf5a85ec47d5a52012ccc92231b41c94308155ad2702eef9d4d.
//
// Solidity: event DepositReceived(uint256 indexed depositID, uint8 indexed accountType, address indexed depositor, uint256 amount)
func (_BToken *BTokenFilterer) ParseDepositReceived(log types.Log) (*BTokenDepositReceived, error) {
	event := new(BTokenDepositReceived)
	if err := _BToken.contract.UnpackLog(event, "DepositReceived", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BTokenInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the BToken contract.
type BTokenInitializedIterator struct {
	Event *BTokenInitialized // Event containing the contract specifics and raw log

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
func (it *BTokenInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BTokenInitialized)
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
		it.Event = new(BTokenInitialized)
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
func (it *BTokenInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BTokenInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BTokenInitialized represents a Initialized event raised by the BToken contract.
type BTokenInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_BToken *BTokenFilterer) FilterInitialized(opts *bind.FilterOpts) (*BTokenInitializedIterator, error) {

	logs, sub, err := _BToken.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &BTokenInitializedIterator{contract: _BToken.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_BToken *BTokenFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *BTokenInitialized) (event.Subscription, error) {

	logs, sub, err := _BToken.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BTokenInitialized)
				if err := _BToken.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_BToken *BTokenFilterer) ParseInitialized(log types.Log) (*BTokenInitialized, error) {
	event := new(BTokenInitialized)
	if err := _BToken.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BTokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the BToken contract.
type BTokenTransferIterator struct {
	Event *BTokenTransfer // Event containing the contract specifics and raw log

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
func (it *BTokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BTokenTransfer)
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
		it.Event = new(BTokenTransfer)
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
func (it *BTokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BTokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BTokenTransfer represents a Transfer event raised by the BToken contract.
type BTokenTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_BToken *BTokenFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*BTokenTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _BToken.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &BTokenTransferIterator{contract: _BToken.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_BToken *BTokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *BTokenTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _BToken.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BTokenTransfer)
				if err := _BToken.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_BToken *BTokenFilterer) ParseTransfer(log types.Log) (*BTokenTransfer, error) {
	event := new(BTokenTransfer)
	if err := _BToken.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
