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

// ALCBMetaData contains all meta data concerning the ALCB contract.
var ALCBMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"centralBridgeRouterAddress_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"accountType\",\"type\":\"uint8\"}],\"name\":\"AccountTypeNotSupported\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"supply\",\"type\":\"uint256\"}],\"name\":\"BurnAmountExceedsSupply\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CannotSetRouterToZeroAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CannotTransferToZeroAddress\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"toAddress\",\"type\":\"address\"}],\"name\":\"ContractsDisallowedDeposits\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DepositAmountZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DepositBurnFail\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"EthTransferFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"InsufficientEth\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"}],\"name\":\"InsufficientFee\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"contractBalance\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"poolBalance\",\"type\":\"uint256\"}],\"name\":\"InvalidBalance\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"InvalidBurnAmount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"depositID\",\"type\":\"uint256\"}],\"name\":\"InvalidDepositId\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"MinimumBurnNotMet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"MinimumMintNotMet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimumValue\",\"type\":\"uint256\"}],\"name\":\"MinimumValueNotMet\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MutexLocked\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyDistribution\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyFactory\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositID\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"accountType\",\"type\":\"uint8\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DepositReceived\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minEth_\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minEth_\",\"type\":\"uint256\"}],\"name\":\"burnTo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"accountType_\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"routerVersion_\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"data_\",\"type\":\"bytes\"}],\"name\":\"depositTokensOnBridges\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"numBTK_\",\"type\":\"uint256\"}],\"name\":\"destroyALCBs\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"distribute\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCentralBridgeRouterAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"depositID\",\"type\":\"uint256\"}],\"name\":\"getDeposit\",\"outputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"accountType\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"internalType\":\"structDeposit\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDepositID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"poolBalance_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalSupply_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numBTK_\",\"type\":\"uint256\"}],\"name\":\"getEthFromALCBsBurn\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"totalSupply_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numBTK_\",\"type\":\"uint256\"}],\"name\":\"getEthToMintALCBs\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"numBTK_\",\"type\":\"uint256\"}],\"name\":\"getLatestEthFromALCBsBurn\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"numBTK_\",\"type\":\"uint256\"}],\"name\":\"getLatestEthToMintALCBs\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"numEth_\",\"type\":\"uint256\"}],\"name\":\"getLatestMintedALCBsFromEth\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMarketSpread\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"poolBalance_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numEth_\",\"type\":\"uint256\"}],\"name\":\"getMintedALCBsFromEth\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPoolBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTotalALCBsDeposited\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getYield\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"minBTK_\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numBTK\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"accountType_\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"minBTK_\",\"type\":\"uint256\"}],\"name\":\"mintDeposit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"minBTK_\",\"type\":\"uint256\"}],\"name\":\"mintTo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numBTK\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"accountType_\",\"type\":\"uint8\"},{\"internalType\":\"bool\",\"name\":\"allowed_\",\"type\":\"bool\"}],\"name\":\"setAccountType\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"accountType_\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"}],\"name\":\"virtualMintDeposit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ALCBABI is the input ABI used to generate the binding from.
// Deprecated: Use ALCBMetaData.ABI instead.
var ALCBABI = ALCBMetaData.ABI

// ALCB is an auto generated Go binding around an Ethereum contract.
type ALCB struct {
	ALCBCaller     // Read-only binding to the contract
	ALCBTransactor // Write-only binding to the contract
	ALCBFilterer   // Log filterer for contract events
}

// ALCBCaller is an auto generated read-only Go binding around an Ethereum contract.
type ALCBCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ALCBTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ALCBTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ALCBFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ALCBFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ALCBSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ALCBSession struct {
	Contract     *ALCB             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ALCBCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ALCBCallerSession struct {
	Contract *ALCBCaller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ALCBTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ALCBTransactorSession struct {
	Contract     *ALCBTransactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ALCBRaw is an auto generated low-level Go binding around an Ethereum contract.
type ALCBRaw struct {
	Contract *ALCB // Generic contract binding to access the raw methods on
}

// ALCBCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ALCBCallerRaw struct {
	Contract *ALCBCaller // Generic read-only contract binding to access the raw methods on
}

// ALCBTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ALCBTransactorRaw struct {
	Contract *ALCBTransactor // Generic write-only contract binding to access the raw methods on
}

// NewALCB creates a new instance of ALCB, bound to a specific deployed contract.
func NewALCB(address common.Address, backend bind.ContractBackend) (*ALCB, error) {
	contract, err := bindALCB(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ALCB{ALCBCaller: ALCBCaller{contract: contract}, ALCBTransactor: ALCBTransactor{contract: contract}, ALCBFilterer: ALCBFilterer{contract: contract}}, nil
}

// NewALCBCaller creates a new read-only instance of ALCB, bound to a specific deployed contract.
func NewALCBCaller(address common.Address, caller bind.ContractCaller) (*ALCBCaller, error) {
	contract, err := bindALCB(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ALCBCaller{contract: contract}, nil
}

// NewALCBTransactor creates a new write-only instance of ALCB, bound to a specific deployed contract.
func NewALCBTransactor(address common.Address, transactor bind.ContractTransactor) (*ALCBTransactor, error) {
	contract, err := bindALCB(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ALCBTransactor{contract: contract}, nil
}

// NewALCBFilterer creates a new log filterer instance of ALCB, bound to a specific deployed contract.
func NewALCBFilterer(address common.Address, filterer bind.ContractFilterer) (*ALCBFilterer, error) {
	contract, err := bindALCB(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ALCBFilterer{contract: contract}, nil
}

// bindALCB binds a generic wrapper to an already deployed contract.
func bindALCB(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ALCBABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ALCB *ALCBRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ALCB.Contract.ALCBCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ALCB *ALCBRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ALCB.Contract.ALCBTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ALCB *ALCBRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ALCB.Contract.ALCBTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ALCB *ALCBCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ALCB.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ALCB *ALCBTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ALCB.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ALCB *ALCBTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ALCB.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_ALCB *ALCBCaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_ALCB *ALCBSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _ALCB.Contract.Allowance(&_ALCB.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_ALCB *ALCBCallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _ALCB.Contract.Allowance(&_ALCB.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_ALCB *ALCBCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_ALCB *ALCBSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _ALCB.Contract.BalanceOf(&_ALCB.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_ALCB *ALCBCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _ALCB.Contract.BalanceOf(&_ALCB.CallOpts, account)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ALCB *ALCBCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ALCB *ALCBSession) Decimals() (uint8, error) {
	return _ALCB.Contract.Decimals(&_ALCB.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_ALCB *ALCBCallerSession) Decimals() (uint8, error) {
	return _ALCB.Contract.Decimals(&_ALCB.CallOpts)
}

// GetCentralBridgeRouterAddress is a free data retrieval call binding the contract method 0xff32fefc.
//
// Solidity: function getCentralBridgeRouterAddress() view returns(address)
func (_ALCB *ALCBCaller) GetCentralBridgeRouterAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getCentralBridgeRouterAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetCentralBridgeRouterAddress is a free data retrieval call binding the contract method 0xff32fefc.
//
// Solidity: function getCentralBridgeRouterAddress() view returns(address)
func (_ALCB *ALCBSession) GetCentralBridgeRouterAddress() (common.Address, error) {
	return _ALCB.Contract.GetCentralBridgeRouterAddress(&_ALCB.CallOpts)
}

// GetCentralBridgeRouterAddress is a free data retrieval call binding the contract method 0xff32fefc.
//
// Solidity: function getCentralBridgeRouterAddress() view returns(address)
func (_ALCB *ALCBCallerSession) GetCentralBridgeRouterAddress() (common.Address, error) {
	return _ALCB.Contract.GetCentralBridgeRouterAddress(&_ALCB.CallOpts)
}

// GetDeposit is a free data retrieval call binding the contract method 0x9f9fb968.
//
// Solidity: function getDeposit(uint256 depositID) view returns((uint8,address,uint256))
func (_ALCB *ALCBCaller) GetDeposit(opts *bind.CallOpts, depositID *big.Int) (Deposit, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getDeposit", depositID)

	if err != nil {
		return *new(Deposit), err
	}

	out0 := *abi.ConvertType(out[0], new(Deposit)).(*Deposit)

	return out0, err

}

// GetDeposit is a free data retrieval call binding the contract method 0x9f9fb968.
//
// Solidity: function getDeposit(uint256 depositID) view returns((uint8,address,uint256))
func (_ALCB *ALCBSession) GetDeposit(depositID *big.Int) (Deposit, error) {
	return _ALCB.Contract.GetDeposit(&_ALCB.CallOpts, depositID)
}

// GetDeposit is a free data retrieval call binding the contract method 0x9f9fb968.
//
// Solidity: function getDeposit(uint256 depositID) view returns((uint8,address,uint256))
func (_ALCB *ALCBCallerSession) GetDeposit(depositID *big.Int) (Deposit, error) {
	return _ALCB.Contract.GetDeposit(&_ALCB.CallOpts, depositID)
}

// GetDepositID is a free data retrieval call binding the contract method 0x94f70c8a.
//
// Solidity: function getDepositID() view returns(uint256)
func (_ALCB *ALCBCaller) GetDepositID(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getDepositID")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetDepositID is a free data retrieval call binding the contract method 0x94f70c8a.
//
// Solidity: function getDepositID() view returns(uint256)
func (_ALCB *ALCBSession) GetDepositID() (*big.Int, error) {
	return _ALCB.Contract.GetDepositID(&_ALCB.CallOpts)
}

// GetDepositID is a free data retrieval call binding the contract method 0x94f70c8a.
//
// Solidity: function getDepositID() view returns(uint256)
func (_ALCB *ALCBCallerSession) GetDepositID() (*big.Int, error) {
	return _ALCB.Contract.GetDepositID(&_ALCB.CallOpts)
}

// GetEthFromALCBsBurn is a free data retrieval call binding the contract method 0x0b6774a1.
//
// Solidity: function getEthFromALCBsBurn(uint256 poolBalance_, uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
func (_ALCB *ALCBCaller) GetEthFromALCBsBurn(opts *bind.CallOpts, poolBalance_ *big.Int, totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getEthFromALCBsBurn", poolBalance_, totalSupply_, numBTK_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetEthFromALCBsBurn is a free data retrieval call binding the contract method 0x0b6774a1.
//
// Solidity: function getEthFromALCBsBurn(uint256 poolBalance_, uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
func (_ALCB *ALCBSession) GetEthFromALCBsBurn(poolBalance_ *big.Int, totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error) {
	return _ALCB.Contract.GetEthFromALCBsBurn(&_ALCB.CallOpts, poolBalance_, totalSupply_, numBTK_)
}

// GetEthFromALCBsBurn is a free data retrieval call binding the contract method 0x0b6774a1.
//
// Solidity: function getEthFromALCBsBurn(uint256 poolBalance_, uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
func (_ALCB *ALCBCallerSession) GetEthFromALCBsBurn(poolBalance_ *big.Int, totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error) {
	return _ALCB.Contract.GetEthFromALCBsBurn(&_ALCB.CallOpts, poolBalance_, totalSupply_, numBTK_)
}

// GetEthToMintALCBs is a free data retrieval call binding the contract method 0x0619c2f3.
//
// Solidity: function getEthToMintALCBs(uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
func (_ALCB *ALCBCaller) GetEthToMintALCBs(opts *bind.CallOpts, totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getEthToMintALCBs", totalSupply_, numBTK_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetEthToMintALCBs is a free data retrieval call binding the contract method 0x0619c2f3.
//
// Solidity: function getEthToMintALCBs(uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
func (_ALCB *ALCBSession) GetEthToMintALCBs(totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error) {
	return _ALCB.Contract.GetEthToMintALCBs(&_ALCB.CallOpts, totalSupply_, numBTK_)
}

// GetEthToMintALCBs is a free data retrieval call binding the contract method 0x0619c2f3.
//
// Solidity: function getEthToMintALCBs(uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
func (_ALCB *ALCBCallerSession) GetEthToMintALCBs(totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error) {
	return _ALCB.Contract.GetEthToMintALCBs(&_ALCB.CallOpts, totalSupply_, numBTK_)
}

// GetLatestEthFromALCBsBurn is a free data retrieval call binding the contract method 0xff326a2c.
//
// Solidity: function getLatestEthFromALCBsBurn(uint256 numBTK_) view returns(uint256 numEth)
func (_ALCB *ALCBCaller) GetLatestEthFromALCBsBurn(opts *bind.CallOpts, numBTK_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getLatestEthFromALCBsBurn", numBTK_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLatestEthFromALCBsBurn is a free data retrieval call binding the contract method 0xff326a2c.
//
// Solidity: function getLatestEthFromALCBsBurn(uint256 numBTK_) view returns(uint256 numEth)
func (_ALCB *ALCBSession) GetLatestEthFromALCBsBurn(numBTK_ *big.Int) (*big.Int, error) {
	return _ALCB.Contract.GetLatestEthFromALCBsBurn(&_ALCB.CallOpts, numBTK_)
}

// GetLatestEthFromALCBsBurn is a free data retrieval call binding the contract method 0xff326a2c.
//
// Solidity: function getLatestEthFromALCBsBurn(uint256 numBTK_) view returns(uint256 numEth)
func (_ALCB *ALCBCallerSession) GetLatestEthFromALCBsBurn(numBTK_ *big.Int) (*big.Int, error) {
	return _ALCB.Contract.GetLatestEthFromALCBsBurn(&_ALCB.CallOpts, numBTK_)
}

// GetLatestEthToMintALCBs is a free data retrieval call binding the contract method 0x777acd84.
//
// Solidity: function getLatestEthToMintALCBs(uint256 numBTK_) view returns(uint256 numEth)
func (_ALCB *ALCBCaller) GetLatestEthToMintALCBs(opts *bind.CallOpts, numBTK_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getLatestEthToMintALCBs", numBTK_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLatestEthToMintALCBs is a free data retrieval call binding the contract method 0x777acd84.
//
// Solidity: function getLatestEthToMintALCBs(uint256 numBTK_) view returns(uint256 numEth)
func (_ALCB *ALCBSession) GetLatestEthToMintALCBs(numBTK_ *big.Int) (*big.Int, error) {
	return _ALCB.Contract.GetLatestEthToMintALCBs(&_ALCB.CallOpts, numBTK_)
}

// GetLatestEthToMintALCBs is a free data retrieval call binding the contract method 0x777acd84.
//
// Solidity: function getLatestEthToMintALCBs(uint256 numBTK_) view returns(uint256 numEth)
func (_ALCB *ALCBCallerSession) GetLatestEthToMintALCBs(numBTK_ *big.Int) (*big.Int, error) {
	return _ALCB.Contract.GetLatestEthToMintALCBs(&_ALCB.CallOpts, numBTK_)
}

// GetLatestMintedALCBsFromEth is a free data retrieval call binding the contract method 0x71497ca6.
//
// Solidity: function getLatestMintedALCBsFromEth(uint256 numEth_) view returns(uint256)
func (_ALCB *ALCBCaller) GetLatestMintedALCBsFromEth(opts *bind.CallOpts, numEth_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getLatestMintedALCBsFromEth", numEth_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLatestMintedALCBsFromEth is a free data retrieval call binding the contract method 0x71497ca6.
//
// Solidity: function getLatestMintedALCBsFromEth(uint256 numEth_) view returns(uint256)
func (_ALCB *ALCBSession) GetLatestMintedALCBsFromEth(numEth_ *big.Int) (*big.Int, error) {
	return _ALCB.Contract.GetLatestMintedALCBsFromEth(&_ALCB.CallOpts, numEth_)
}

// GetLatestMintedALCBsFromEth is a free data retrieval call binding the contract method 0x71497ca6.
//
// Solidity: function getLatestMintedALCBsFromEth(uint256 numEth_) view returns(uint256)
func (_ALCB *ALCBCallerSession) GetLatestMintedALCBsFromEth(numEth_ *big.Int) (*big.Int, error) {
	return _ALCB.Contract.GetLatestMintedALCBsFromEth(&_ALCB.CallOpts, numEth_)
}

// GetMarketSpread is a free data retrieval call binding the contract method 0x086cfefd.
//
// Solidity: function getMarketSpread() pure returns(uint256)
func (_ALCB *ALCBCaller) GetMarketSpread(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getMarketSpread")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMarketSpread is a free data retrieval call binding the contract method 0x086cfefd.
//
// Solidity: function getMarketSpread() pure returns(uint256)
func (_ALCB *ALCBSession) GetMarketSpread() (*big.Int, error) {
	return _ALCB.Contract.GetMarketSpread(&_ALCB.CallOpts)
}

// GetMarketSpread is a free data retrieval call binding the contract method 0x086cfefd.
//
// Solidity: function getMarketSpread() pure returns(uint256)
func (_ALCB *ALCBCallerSession) GetMarketSpread() (*big.Int, error) {
	return _ALCB.Contract.GetMarketSpread(&_ALCB.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ALCB *ALCBCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ALCB *ALCBSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ALCB.Contract.GetMetamorphicContractAddress(&_ALCB.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ALCB *ALCBCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ALCB.Contract.GetMetamorphicContractAddress(&_ALCB.CallOpts, _salt, _factory)
}

// GetMintedALCBsFromEth is a free data retrieval call binding the contract method 0x6e1d3f22.
//
// Solidity: function getMintedALCBsFromEth(uint256 poolBalance_, uint256 numEth_) pure returns(uint256)
func (_ALCB *ALCBCaller) GetMintedALCBsFromEth(opts *bind.CallOpts, poolBalance_ *big.Int, numEth_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getMintedALCBsFromEth", poolBalance_, numEth_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMintedALCBsFromEth is a free data retrieval call binding the contract method 0x6e1d3f22.
//
// Solidity: function getMintedALCBsFromEth(uint256 poolBalance_, uint256 numEth_) pure returns(uint256)
func (_ALCB *ALCBSession) GetMintedALCBsFromEth(poolBalance_ *big.Int, numEth_ *big.Int) (*big.Int, error) {
	return _ALCB.Contract.GetMintedALCBsFromEth(&_ALCB.CallOpts, poolBalance_, numEth_)
}

// GetMintedALCBsFromEth is a free data retrieval call binding the contract method 0x6e1d3f22.
//
// Solidity: function getMintedALCBsFromEth(uint256 poolBalance_, uint256 numEth_) pure returns(uint256)
func (_ALCB *ALCBCallerSession) GetMintedALCBsFromEth(poolBalance_ *big.Int, numEth_ *big.Int) (*big.Int, error) {
	return _ALCB.Contract.GetMintedALCBsFromEth(&_ALCB.CallOpts, poolBalance_, numEth_)
}

// GetPoolBalance is a free data retrieval call binding the contract method 0xabd70aa2.
//
// Solidity: function getPoolBalance() view returns(uint256)
func (_ALCB *ALCBCaller) GetPoolBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getPoolBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPoolBalance is a free data retrieval call binding the contract method 0xabd70aa2.
//
// Solidity: function getPoolBalance() view returns(uint256)
func (_ALCB *ALCBSession) GetPoolBalance() (*big.Int, error) {
	return _ALCB.Contract.GetPoolBalance(&_ALCB.CallOpts)
}

// GetPoolBalance is a free data retrieval call binding the contract method 0xabd70aa2.
//
// Solidity: function getPoolBalance() view returns(uint256)
func (_ALCB *ALCBCallerSession) GetPoolBalance() (*big.Int, error) {
	return _ALCB.Contract.GetPoolBalance(&_ALCB.CallOpts)
}

// GetTotalALCBsDeposited is a free data retrieval call binding the contract method 0x90813858.
//
// Solidity: function getTotalALCBsDeposited() view returns(uint256)
func (_ALCB *ALCBCaller) GetTotalALCBsDeposited(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getTotalALCBsDeposited")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalALCBsDeposited is a free data retrieval call binding the contract method 0x90813858.
//
// Solidity: function getTotalALCBsDeposited() view returns(uint256)
func (_ALCB *ALCBSession) GetTotalALCBsDeposited() (*big.Int, error) {
	return _ALCB.Contract.GetTotalALCBsDeposited(&_ALCB.CallOpts)
}

// GetTotalALCBsDeposited is a free data retrieval call binding the contract method 0x90813858.
//
// Solidity: function getTotalALCBsDeposited() view returns(uint256)
func (_ALCB *ALCBCallerSession) GetTotalALCBsDeposited() (*big.Int, error) {
	return _ALCB.Contract.GetTotalALCBsDeposited(&_ALCB.CallOpts)
}

// GetYield is a free data retrieval call binding the contract method 0x7c262871.
//
// Solidity: function getYield() view returns(uint256)
func (_ALCB *ALCBCaller) GetYield(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "getYield")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetYield is a free data retrieval call binding the contract method 0x7c262871.
//
// Solidity: function getYield() view returns(uint256)
func (_ALCB *ALCBSession) GetYield() (*big.Int, error) {
	return _ALCB.Contract.GetYield(&_ALCB.CallOpts)
}

// GetYield is a free data retrieval call binding the contract method 0x7c262871.
//
// Solidity: function getYield() view returns(uint256)
func (_ALCB *ALCBCallerSession) GetYield() (*big.Int, error) {
	return _ALCB.Contract.GetYield(&_ALCB.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ALCB *ALCBCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ALCB *ALCBSession) Name() (string, error) {
	return _ALCB.Contract.Name(&_ALCB.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ALCB *ALCBCallerSession) Name() (string, error) {
	return _ALCB.Contract.Name(&_ALCB.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ALCB *ALCBCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ALCB *ALCBSession) Symbol() (string, error) {
	return _ALCB.Contract.Symbol(&_ALCB.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ALCB *ALCBCallerSession) Symbol() (string, error) {
	return _ALCB.Contract.Symbol(&_ALCB.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ALCB *ALCBCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ALCB.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ALCB *ALCBSession) TotalSupply() (*big.Int, error) {
	return _ALCB.Contract.TotalSupply(&_ALCB.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_ALCB *ALCBCallerSession) TotalSupply() (*big.Int, error) {
	return _ALCB.Contract.TotalSupply(&_ALCB.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ALCB *ALCBTransactor) Approve(opts *bind.TransactOpts, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "approve", spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ALCB *ALCBSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.Approve(&_ALCB.TransactOpts, spender, amount)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 amount) returns(bool)
func (_ALCB *ALCBTransactorSession) Approve(spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.Approve(&_ALCB.TransactOpts, spender, amount)
}

// Burn is a paid mutator transaction binding the contract method 0xb390c0ab.
//
// Solidity: function burn(uint256 amount_, uint256 minEth_) returns(uint256 numEth)
func (_ALCB *ALCBTransactor) Burn(opts *bind.TransactOpts, amount_ *big.Int, minEth_ *big.Int) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "burn", amount_, minEth_)
}

// Burn is a paid mutator transaction binding the contract method 0xb390c0ab.
//
// Solidity: function burn(uint256 amount_, uint256 minEth_) returns(uint256 numEth)
func (_ALCB *ALCBSession) Burn(amount_ *big.Int, minEth_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.Burn(&_ALCB.TransactOpts, amount_, minEth_)
}

// Burn is a paid mutator transaction binding the contract method 0xb390c0ab.
//
// Solidity: function burn(uint256 amount_, uint256 minEth_) returns(uint256 numEth)
func (_ALCB *ALCBTransactorSession) Burn(amount_ *big.Int, minEth_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.Burn(&_ALCB.TransactOpts, amount_, minEth_)
}

// BurnTo is a paid mutator transaction binding the contract method 0x9b057203.
//
// Solidity: function burnTo(address to_, uint256 amount_, uint256 minEth_) returns(uint256 numEth)
func (_ALCB *ALCBTransactor) BurnTo(opts *bind.TransactOpts, to_ common.Address, amount_ *big.Int, minEth_ *big.Int) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "burnTo", to_, amount_, minEth_)
}

// BurnTo is a paid mutator transaction binding the contract method 0x9b057203.
//
// Solidity: function burnTo(address to_, uint256 amount_, uint256 minEth_) returns(uint256 numEth)
func (_ALCB *ALCBSession) BurnTo(to_ common.Address, amount_ *big.Int, minEth_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.BurnTo(&_ALCB.TransactOpts, to_, amount_, minEth_)
}

// BurnTo is a paid mutator transaction binding the contract method 0x9b057203.
//
// Solidity: function burnTo(address to_, uint256 amount_, uint256 minEth_) returns(uint256 numEth)
func (_ALCB *ALCBTransactorSession) BurnTo(to_ common.Address, amount_ *big.Int, minEth_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.BurnTo(&_ALCB.TransactOpts, to_, amount_, minEth_)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_ALCB *ALCBTransactor) DecreaseAllowance(opts *bind.TransactOpts, spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "decreaseAllowance", spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_ALCB *ALCBSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.DecreaseAllowance(&_ALCB.TransactOpts, spender, subtractedValue)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 subtractedValue) returns(bool)
func (_ALCB *ALCBTransactorSession) DecreaseAllowance(spender common.Address, subtractedValue *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.DecreaseAllowance(&_ALCB.TransactOpts, spender, subtractedValue)
}

// Deposit is a paid mutator transaction binding the contract method 0x00838172.
//
// Solidity: function deposit(uint8 accountType_, address to_, uint256 amount_) returns(uint256)
func (_ALCB *ALCBTransactor) Deposit(opts *bind.TransactOpts, accountType_ uint8, to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "deposit", accountType_, to_, amount_)
}

// Deposit is a paid mutator transaction binding the contract method 0x00838172.
//
// Solidity: function deposit(uint8 accountType_, address to_, uint256 amount_) returns(uint256)
func (_ALCB *ALCBSession) Deposit(accountType_ uint8, to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.Deposit(&_ALCB.TransactOpts, accountType_, to_, amount_)
}

// Deposit is a paid mutator transaction binding the contract method 0x00838172.
//
// Solidity: function deposit(uint8 accountType_, address to_, uint256 amount_) returns(uint256)
func (_ALCB *ALCBTransactorSession) Deposit(accountType_ uint8, to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.Deposit(&_ALCB.TransactOpts, accountType_, to_, amount_)
}

// DepositTokensOnBridges is a paid mutator transaction binding the contract method 0xc4602af3.
//
// Solidity: function depositTokensOnBridges(uint8 routerVersion_, bytes data_) payable returns()
func (_ALCB *ALCBTransactor) DepositTokensOnBridges(opts *bind.TransactOpts, routerVersion_ uint8, data_ []byte) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "depositTokensOnBridges", routerVersion_, data_)
}

// DepositTokensOnBridges is a paid mutator transaction binding the contract method 0xc4602af3.
//
// Solidity: function depositTokensOnBridges(uint8 routerVersion_, bytes data_) payable returns()
func (_ALCB *ALCBSession) DepositTokensOnBridges(routerVersion_ uint8, data_ []byte) (*types.Transaction, error) {
	return _ALCB.Contract.DepositTokensOnBridges(&_ALCB.TransactOpts, routerVersion_, data_)
}

// DepositTokensOnBridges is a paid mutator transaction binding the contract method 0xc4602af3.
//
// Solidity: function depositTokensOnBridges(uint8 routerVersion_, bytes data_) payable returns()
func (_ALCB *ALCBTransactorSession) DepositTokensOnBridges(routerVersion_ uint8, data_ []byte) (*types.Transaction, error) {
	return _ALCB.Contract.DepositTokensOnBridges(&_ALCB.TransactOpts, routerVersion_, data_)
}

// DestroyALCBs is a paid mutator transaction binding the contract method 0x24663df8.
//
// Solidity: function destroyALCBs(uint256 numBTK_) returns(bool)
func (_ALCB *ALCBTransactor) DestroyALCBs(opts *bind.TransactOpts, numBTK_ *big.Int) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "destroyALCBs", numBTK_)
}

// DestroyALCBs is a paid mutator transaction binding the contract method 0x24663df8.
//
// Solidity: function destroyALCBs(uint256 numBTK_) returns(bool)
func (_ALCB *ALCBSession) DestroyALCBs(numBTK_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.DestroyALCBs(&_ALCB.TransactOpts, numBTK_)
}

// DestroyALCBs is a paid mutator transaction binding the contract method 0x24663df8.
//
// Solidity: function destroyALCBs(uint256 numBTK_) returns(bool)
func (_ALCB *ALCBTransactorSession) DestroyALCBs(numBTK_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.DestroyALCBs(&_ALCB.TransactOpts, numBTK_)
}

// Distribute is a paid mutator transaction binding the contract method 0xe4fc6b6d.
//
// Solidity: function distribute() returns(bool)
func (_ALCB *ALCBTransactor) Distribute(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "distribute")
}

// Distribute is a paid mutator transaction binding the contract method 0xe4fc6b6d.
//
// Solidity: function distribute() returns(bool)
func (_ALCB *ALCBSession) Distribute() (*types.Transaction, error) {
	return _ALCB.Contract.Distribute(&_ALCB.TransactOpts)
}

// Distribute is a paid mutator transaction binding the contract method 0xe4fc6b6d.
//
// Solidity: function distribute() returns(bool)
func (_ALCB *ALCBTransactorSession) Distribute() (*types.Transaction, error) {
	return _ALCB.Contract.Distribute(&_ALCB.TransactOpts)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_ALCB *ALCBTransactor) IncreaseAllowance(opts *bind.TransactOpts, spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "increaseAllowance", spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_ALCB *ALCBSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.IncreaseAllowance(&_ALCB.TransactOpts, spender, addedValue)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 addedValue) returns(bool)
func (_ALCB *ALCBTransactorSession) IncreaseAllowance(spender common.Address, addedValue *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.IncreaseAllowance(&_ALCB.TransactOpts, spender, addedValue)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 minBTK_) payable returns(uint256 numBTK)
func (_ALCB *ALCBTransactor) Mint(opts *bind.TransactOpts, minBTK_ *big.Int) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "mint", minBTK_)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 minBTK_) payable returns(uint256 numBTK)
func (_ALCB *ALCBSession) Mint(minBTK_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.Mint(&_ALCB.TransactOpts, minBTK_)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 minBTK_) payable returns(uint256 numBTK)
func (_ALCB *ALCBTransactorSession) Mint(minBTK_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.Mint(&_ALCB.TransactOpts, minBTK_)
}

// MintDeposit is a paid mutator transaction binding the contract method 0x4f232628.
//
// Solidity: function mintDeposit(uint8 accountType_, address to_, uint256 minBTK_) payable returns(uint256)
func (_ALCB *ALCBTransactor) MintDeposit(opts *bind.TransactOpts, accountType_ uint8, to_ common.Address, minBTK_ *big.Int) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "mintDeposit", accountType_, to_, minBTK_)
}

// MintDeposit is a paid mutator transaction binding the contract method 0x4f232628.
//
// Solidity: function mintDeposit(uint8 accountType_, address to_, uint256 minBTK_) payable returns(uint256)
func (_ALCB *ALCBSession) MintDeposit(accountType_ uint8, to_ common.Address, minBTK_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.MintDeposit(&_ALCB.TransactOpts, accountType_, to_, minBTK_)
}

// MintDeposit is a paid mutator transaction binding the contract method 0x4f232628.
//
// Solidity: function mintDeposit(uint8 accountType_, address to_, uint256 minBTK_) payable returns(uint256)
func (_ALCB *ALCBTransactorSession) MintDeposit(accountType_ uint8, to_ common.Address, minBTK_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.MintDeposit(&_ALCB.TransactOpts, accountType_, to_, minBTK_)
}

// MintTo is a paid mutator transaction binding the contract method 0x449a52f8.
//
// Solidity: function mintTo(address to_, uint256 minBTK_) payable returns(uint256 numBTK)
func (_ALCB *ALCBTransactor) MintTo(opts *bind.TransactOpts, to_ common.Address, minBTK_ *big.Int) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "mintTo", to_, minBTK_)
}

// MintTo is a paid mutator transaction binding the contract method 0x449a52f8.
//
// Solidity: function mintTo(address to_, uint256 minBTK_) payable returns(uint256 numBTK)
func (_ALCB *ALCBSession) MintTo(to_ common.Address, minBTK_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.MintTo(&_ALCB.TransactOpts, to_, minBTK_)
}

// MintTo is a paid mutator transaction binding the contract method 0x449a52f8.
//
// Solidity: function mintTo(address to_, uint256 minBTK_) payable returns(uint256 numBTK)
func (_ALCB *ALCBTransactorSession) MintTo(to_ common.Address, minBTK_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.MintTo(&_ALCB.TransactOpts, to_, minBTK_)
}

// SetAccountType is a paid mutator transaction binding the contract method 0x14c8d876.
//
// Solidity: function setAccountType(uint8 accountType_, bool allowed_) returns()
func (_ALCB *ALCBTransactor) SetAccountType(opts *bind.TransactOpts, accountType_ uint8, allowed_ bool) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "setAccountType", accountType_, allowed_)
}

// SetAccountType is a paid mutator transaction binding the contract method 0x14c8d876.
//
// Solidity: function setAccountType(uint8 accountType_, bool allowed_) returns()
func (_ALCB *ALCBSession) SetAccountType(accountType_ uint8, allowed_ bool) (*types.Transaction, error) {
	return _ALCB.Contract.SetAccountType(&_ALCB.TransactOpts, accountType_, allowed_)
}

// SetAccountType is a paid mutator transaction binding the contract method 0x14c8d876.
//
// Solidity: function setAccountType(uint8 accountType_, bool allowed_) returns()
func (_ALCB *ALCBTransactorSession) SetAccountType(accountType_ uint8, allowed_ bool) (*types.Transaction, error) {
	return _ALCB.Contract.SetAccountType(&_ALCB.TransactOpts, accountType_, allowed_)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_ALCB *ALCBTransactor) Transfer(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "transfer", to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_ALCB *ALCBSession) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.Transfer(&_ALCB.TransactOpts, to, amount)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 amount) returns(bool)
func (_ALCB *ALCBTransactorSession) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.Transfer(&_ALCB.TransactOpts, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_ALCB *ALCBTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "transferFrom", from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_ALCB *ALCBSession) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.TransferFrom(&_ALCB.TransactOpts, from, to, amount)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 amount) returns(bool)
func (_ALCB *ALCBTransactorSession) TransferFrom(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.TransferFrom(&_ALCB.TransactOpts, from, to, amount)
}

// VirtualMintDeposit is a paid mutator transaction binding the contract method 0x92178278.
//
// Solidity: function virtualMintDeposit(uint8 accountType_, address to_, uint256 amount_) returns(uint256)
func (_ALCB *ALCBTransactor) VirtualMintDeposit(opts *bind.TransactOpts, accountType_ uint8, to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _ALCB.contract.Transact(opts, "virtualMintDeposit", accountType_, to_, amount_)
}

// VirtualMintDeposit is a paid mutator transaction binding the contract method 0x92178278.
//
// Solidity: function virtualMintDeposit(uint8 accountType_, address to_, uint256 amount_) returns(uint256)
func (_ALCB *ALCBSession) VirtualMintDeposit(accountType_ uint8, to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.VirtualMintDeposit(&_ALCB.TransactOpts, accountType_, to_, amount_)
}

// VirtualMintDeposit is a paid mutator transaction binding the contract method 0x92178278.
//
// Solidity: function virtualMintDeposit(uint8 accountType_, address to_, uint256 amount_) returns(uint256)
func (_ALCB *ALCBTransactorSession) VirtualMintDeposit(accountType_ uint8, to_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _ALCB.Contract.VirtualMintDeposit(&_ALCB.TransactOpts, accountType_, to_, amount_)
}

// ALCBApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the ALCB contract.
type ALCBApprovalIterator struct {
	Event *ALCBApproval // Event containing the contract specifics and raw log

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
func (it *ALCBApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ALCBApproval)
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
		it.Event = new(ALCBApproval)
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
func (it *ALCBApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ALCBApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ALCBApproval represents a Approval event raised by the ALCB contract.
type ALCBApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ALCB *ALCBFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*ALCBApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ALCB.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &ALCBApprovalIterator{contract: _ALCB.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_ALCB *ALCBFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ALCBApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _ALCB.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ALCBApproval)
				if err := _ALCB.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_ALCB *ALCBFilterer) ParseApproval(log types.Log) (*ALCBApproval, error) {
	event := new(ALCBApproval)
	if err := _ALCB.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ALCBDepositReceivedIterator is returned from FilterDepositReceived and is used to iterate over the raw logs and unpacked data for DepositReceived events raised by the ALCB contract.
type ALCBDepositReceivedIterator struct {
	Event *ALCBDepositReceived // Event containing the contract specifics and raw log

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
func (it *ALCBDepositReceivedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ALCBDepositReceived)
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
		it.Event = new(ALCBDepositReceived)
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
func (it *ALCBDepositReceivedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ALCBDepositReceivedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ALCBDepositReceived represents a DepositReceived event raised by the ALCB contract.
type ALCBDepositReceived struct {
	DepositID   *big.Int
	AccountType uint8
	Depositor   common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterDepositReceived is a free log retrieval operation binding the contract event 0x9d291e7244fa9bf5a85ec47d5a52012ccc92231b41c94308155ad2702eef9d4d.
//
// Solidity: event DepositReceived(uint256 indexed depositID, uint8 indexed accountType, address indexed depositor, uint256 amount)
func (_ALCB *ALCBFilterer) FilterDepositReceived(opts *bind.FilterOpts, depositID []*big.Int, accountType []uint8, depositor []common.Address) (*ALCBDepositReceivedIterator, error) {

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

	logs, sub, err := _ALCB.contract.FilterLogs(opts, "DepositReceived", depositIDRule, accountTypeRule, depositorRule)
	if err != nil {
		return nil, err
	}
	return &ALCBDepositReceivedIterator{contract: _ALCB.contract, event: "DepositReceived", logs: logs, sub: sub}, nil
}

// WatchDepositReceived is a free log subscription operation binding the contract event 0x9d291e7244fa9bf5a85ec47d5a52012ccc92231b41c94308155ad2702eef9d4d.
//
// Solidity: event DepositReceived(uint256 indexed depositID, uint8 indexed accountType, address indexed depositor, uint256 amount)
func (_ALCB *ALCBFilterer) WatchDepositReceived(opts *bind.WatchOpts, sink chan<- *ALCBDepositReceived, depositID []*big.Int, accountType []uint8, depositor []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _ALCB.contract.WatchLogs(opts, "DepositReceived", depositIDRule, accountTypeRule, depositorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ALCBDepositReceived)
				if err := _ALCB.contract.UnpackLog(event, "DepositReceived", log); err != nil {
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
func (_ALCB *ALCBFilterer) ParseDepositReceived(log types.Log) (*ALCBDepositReceived, error) {
	event := new(ALCBDepositReceived)
	if err := _ALCB.contract.UnpackLog(event, "DepositReceived", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ALCBTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the ALCB contract.
type ALCBTransferIterator struct {
	Event *ALCBTransfer // Event containing the contract specifics and raw log

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
func (it *ALCBTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ALCBTransfer)
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
		it.Event = new(ALCBTransfer)
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
func (it *ALCBTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ALCBTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ALCBTransfer represents a Transfer event raised by the ALCB contract.
type ALCBTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ALCB *ALCBFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*ALCBTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ALCB.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &ALCBTransferIterator{contract: _ALCB.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_ALCB *ALCBFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ALCBTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ALCB.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ALCBTransfer)
				if err := _ALCB.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_ALCB *ALCBFilterer) ParseTransfer(log types.Log) (*ALCBTransfer, error) {
	event := new(ALCBTransfer)
	if err := _ALCB.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
