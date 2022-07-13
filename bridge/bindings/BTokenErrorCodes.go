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

// BTokenErrorCodesMetaData contains all meta data concerning the BTokenErrorCodes contract.
var BTokenErrorCodesMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"BTOKEN_BURN_AMOUNT_EXCEEDS_SUPPLY\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BTOKEN_CONTRACTS_DISALLOWED_DEPOSITS\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BTOKEN_DEPOSIT_AMOUNT_ZERO\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BTOKEN_DEPOSIT_BURN_FAIL\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BTOKEN_INVALID_BALANCE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BTOKEN_INVALID_BURN_AMOUNT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BTOKEN_INVALID_DEPOSIT_ID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BTOKEN_MARKET_SPREAD_TOO_LOW\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BTOKEN_MINIMUM_BURN_NOT_MET\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BTOKEN_MINIMUM_MINT_NOT_MET\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BTOKEN_MINT_INSUFFICIENT_ETH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BTOKEN_SPLIT_VALUE_SUM_ERROR\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// BTokenErrorCodesABI is the input ABI used to generate the binding from.
// Deprecated: Use BTokenErrorCodesMetaData.ABI instead.
var BTokenErrorCodesABI = BTokenErrorCodesMetaData.ABI

// BTokenErrorCodes is an auto generated Go binding around an Ethereum contract.
type BTokenErrorCodes struct {
	BTokenErrorCodesCaller     // Read-only binding to the contract
	BTokenErrorCodesTransactor // Write-only binding to the contract
	BTokenErrorCodesFilterer   // Log filterer for contract events
}

// BTokenErrorCodesCaller is an auto generated read-only Go binding around an Ethereum contract.
type BTokenErrorCodesCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BTokenErrorCodesTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BTokenErrorCodesTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BTokenErrorCodesFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BTokenErrorCodesFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BTokenErrorCodesSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BTokenErrorCodesSession struct {
	Contract     *BTokenErrorCodes // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BTokenErrorCodesCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BTokenErrorCodesCallerSession struct {
	Contract *BTokenErrorCodesCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// BTokenErrorCodesTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BTokenErrorCodesTransactorSession struct {
	Contract     *BTokenErrorCodesTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// BTokenErrorCodesRaw is an auto generated low-level Go binding around an Ethereum contract.
type BTokenErrorCodesRaw struct {
	Contract *BTokenErrorCodes // Generic contract binding to access the raw methods on
}

// BTokenErrorCodesCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BTokenErrorCodesCallerRaw struct {
	Contract *BTokenErrorCodesCaller // Generic read-only contract binding to access the raw methods on
}

// BTokenErrorCodesTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BTokenErrorCodesTransactorRaw struct {
	Contract *BTokenErrorCodesTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBTokenErrorCodes creates a new instance of BTokenErrorCodes, bound to a specific deployed contract.
func NewBTokenErrorCodes(address common.Address, backend bind.ContractBackend) (*BTokenErrorCodes, error) {
	contract, err := bindBTokenErrorCodes(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BTokenErrorCodes{BTokenErrorCodesCaller: BTokenErrorCodesCaller{contract: contract}, BTokenErrorCodesTransactor: BTokenErrorCodesTransactor{contract: contract}, BTokenErrorCodesFilterer: BTokenErrorCodesFilterer{contract: contract}}, nil
}

// NewBTokenErrorCodesCaller creates a new read-only instance of BTokenErrorCodes, bound to a specific deployed contract.
func NewBTokenErrorCodesCaller(address common.Address, caller bind.ContractCaller) (*BTokenErrorCodesCaller, error) {
	contract, err := bindBTokenErrorCodes(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BTokenErrorCodesCaller{contract: contract}, nil
}

// NewBTokenErrorCodesTransactor creates a new write-only instance of BTokenErrorCodes, bound to a specific deployed contract.
func NewBTokenErrorCodesTransactor(address common.Address, transactor bind.ContractTransactor) (*BTokenErrorCodesTransactor, error) {
	contract, err := bindBTokenErrorCodes(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BTokenErrorCodesTransactor{contract: contract}, nil
}

// NewBTokenErrorCodesFilterer creates a new log filterer instance of BTokenErrorCodes, bound to a specific deployed contract.
func NewBTokenErrorCodesFilterer(address common.Address, filterer bind.ContractFilterer) (*BTokenErrorCodesFilterer, error) {
	contract, err := bindBTokenErrorCodes(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BTokenErrorCodesFilterer{contract: contract}, nil
}

// bindBTokenErrorCodes binds a generic wrapper to an already deployed contract.
func bindBTokenErrorCodes(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BTokenErrorCodesABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BTokenErrorCodes *BTokenErrorCodesRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BTokenErrorCodes.Contract.BTokenErrorCodesCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BTokenErrorCodes *BTokenErrorCodesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BTokenErrorCodes.Contract.BTokenErrorCodesTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BTokenErrorCodes *BTokenErrorCodesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BTokenErrorCodes.Contract.BTokenErrorCodesTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BTokenErrorCodes *BTokenErrorCodesCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BTokenErrorCodes.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BTokenErrorCodes *BTokenErrorCodesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BTokenErrorCodes.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BTokenErrorCodes *BTokenErrorCodesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BTokenErrorCodes.Contract.contract.Transact(opts, method, params...)
}

// BTOKENBURNAMOUNTEXCEEDSSUPPLY is a free data retrieval call binding the contract method 0x1e13bfbb.
//
// Solidity: function BTOKEN_BURN_AMOUNT_EXCEEDS_SUPPLY() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCaller) BTOKENBURNAMOUNTEXCEEDSSUPPLY(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BTokenErrorCodes.contract.Call(opts, &out, "BTOKEN_BURN_AMOUNT_EXCEEDS_SUPPLY")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BTOKENBURNAMOUNTEXCEEDSSUPPLY is a free data retrieval call binding the contract method 0x1e13bfbb.
//
// Solidity: function BTOKEN_BURN_AMOUNT_EXCEEDS_SUPPLY() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesSession) BTOKENBURNAMOUNTEXCEEDSSUPPLY() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENBURNAMOUNTEXCEEDSSUPPLY(&_BTokenErrorCodes.CallOpts)
}

// BTOKENBURNAMOUNTEXCEEDSSUPPLY is a free data retrieval call binding the contract method 0x1e13bfbb.
//
// Solidity: function BTOKEN_BURN_AMOUNT_EXCEEDS_SUPPLY() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCallerSession) BTOKENBURNAMOUNTEXCEEDSSUPPLY() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENBURNAMOUNTEXCEEDSSUPPLY(&_BTokenErrorCodes.CallOpts)
}

// BTOKENCONTRACTSDISALLOWEDDEPOSITS is a free data retrieval call binding the contract method 0x0e19d024.
//
// Solidity: function BTOKEN_CONTRACTS_DISALLOWED_DEPOSITS() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCaller) BTOKENCONTRACTSDISALLOWEDDEPOSITS(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BTokenErrorCodes.contract.Call(opts, &out, "BTOKEN_CONTRACTS_DISALLOWED_DEPOSITS")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BTOKENCONTRACTSDISALLOWEDDEPOSITS is a free data retrieval call binding the contract method 0x0e19d024.
//
// Solidity: function BTOKEN_CONTRACTS_DISALLOWED_DEPOSITS() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesSession) BTOKENCONTRACTSDISALLOWEDDEPOSITS() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENCONTRACTSDISALLOWEDDEPOSITS(&_BTokenErrorCodes.CallOpts)
}

// BTOKENCONTRACTSDISALLOWEDDEPOSITS is a free data retrieval call binding the contract method 0x0e19d024.
//
// Solidity: function BTOKEN_CONTRACTS_DISALLOWED_DEPOSITS() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCallerSession) BTOKENCONTRACTSDISALLOWEDDEPOSITS() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENCONTRACTSDISALLOWEDDEPOSITS(&_BTokenErrorCodes.CallOpts)
}

// BTOKENDEPOSITAMOUNTZERO is a free data retrieval call binding the contract method 0xb949d485.
//
// Solidity: function BTOKEN_DEPOSIT_AMOUNT_ZERO() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCaller) BTOKENDEPOSITAMOUNTZERO(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BTokenErrorCodes.contract.Call(opts, &out, "BTOKEN_DEPOSIT_AMOUNT_ZERO")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BTOKENDEPOSITAMOUNTZERO is a free data retrieval call binding the contract method 0xb949d485.
//
// Solidity: function BTOKEN_DEPOSIT_AMOUNT_ZERO() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesSession) BTOKENDEPOSITAMOUNTZERO() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENDEPOSITAMOUNTZERO(&_BTokenErrorCodes.CallOpts)
}

// BTOKENDEPOSITAMOUNTZERO is a free data retrieval call binding the contract method 0xb949d485.
//
// Solidity: function BTOKEN_DEPOSIT_AMOUNT_ZERO() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCallerSession) BTOKENDEPOSITAMOUNTZERO() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENDEPOSITAMOUNTZERO(&_BTokenErrorCodes.CallOpts)
}

// BTOKENDEPOSITBURNFAIL is a free data retrieval call binding the contract method 0xfe4a969a.
//
// Solidity: function BTOKEN_DEPOSIT_BURN_FAIL() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCaller) BTOKENDEPOSITBURNFAIL(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BTokenErrorCodes.contract.Call(opts, &out, "BTOKEN_DEPOSIT_BURN_FAIL")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BTOKENDEPOSITBURNFAIL is a free data retrieval call binding the contract method 0xfe4a969a.
//
// Solidity: function BTOKEN_DEPOSIT_BURN_FAIL() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesSession) BTOKENDEPOSITBURNFAIL() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENDEPOSITBURNFAIL(&_BTokenErrorCodes.CallOpts)
}

// BTOKENDEPOSITBURNFAIL is a free data retrieval call binding the contract method 0xfe4a969a.
//
// Solidity: function BTOKEN_DEPOSIT_BURN_FAIL() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCallerSession) BTOKENDEPOSITBURNFAIL() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENDEPOSITBURNFAIL(&_BTokenErrorCodes.CallOpts)
}

// BTOKENINVALIDBALANCE is a free data retrieval call binding the contract method 0xa3d600f1.
//
// Solidity: function BTOKEN_INVALID_BALANCE() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCaller) BTOKENINVALIDBALANCE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BTokenErrorCodes.contract.Call(opts, &out, "BTOKEN_INVALID_BALANCE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BTOKENINVALIDBALANCE is a free data retrieval call binding the contract method 0xa3d600f1.
//
// Solidity: function BTOKEN_INVALID_BALANCE() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesSession) BTOKENINVALIDBALANCE() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENINVALIDBALANCE(&_BTokenErrorCodes.CallOpts)
}

// BTOKENINVALIDBALANCE is a free data retrieval call binding the contract method 0xa3d600f1.
//
// Solidity: function BTOKEN_INVALID_BALANCE() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCallerSession) BTOKENINVALIDBALANCE() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENINVALIDBALANCE(&_BTokenErrorCodes.CallOpts)
}

// BTOKENINVALIDBURNAMOUNT is a free data retrieval call binding the contract method 0xf87e114e.
//
// Solidity: function BTOKEN_INVALID_BURN_AMOUNT() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCaller) BTOKENINVALIDBURNAMOUNT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BTokenErrorCodes.contract.Call(opts, &out, "BTOKEN_INVALID_BURN_AMOUNT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BTOKENINVALIDBURNAMOUNT is a free data retrieval call binding the contract method 0xf87e114e.
//
// Solidity: function BTOKEN_INVALID_BURN_AMOUNT() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesSession) BTOKENINVALIDBURNAMOUNT() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENINVALIDBURNAMOUNT(&_BTokenErrorCodes.CallOpts)
}

// BTOKENINVALIDBURNAMOUNT is a free data retrieval call binding the contract method 0xf87e114e.
//
// Solidity: function BTOKEN_INVALID_BURN_AMOUNT() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCallerSession) BTOKENINVALIDBURNAMOUNT() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENINVALIDBURNAMOUNT(&_BTokenErrorCodes.CallOpts)
}

// BTOKENINVALIDDEPOSITID is a free data retrieval call binding the contract method 0x3dd42816.
//
// Solidity: function BTOKEN_INVALID_DEPOSIT_ID() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCaller) BTOKENINVALIDDEPOSITID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BTokenErrorCodes.contract.Call(opts, &out, "BTOKEN_INVALID_DEPOSIT_ID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BTOKENINVALIDDEPOSITID is a free data retrieval call binding the contract method 0x3dd42816.
//
// Solidity: function BTOKEN_INVALID_DEPOSIT_ID() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesSession) BTOKENINVALIDDEPOSITID() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENINVALIDDEPOSITID(&_BTokenErrorCodes.CallOpts)
}

// BTOKENINVALIDDEPOSITID is a free data retrieval call binding the contract method 0x3dd42816.
//
// Solidity: function BTOKEN_INVALID_DEPOSIT_ID() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCallerSession) BTOKENINVALIDDEPOSITID() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENINVALIDDEPOSITID(&_BTokenErrorCodes.CallOpts)
}

// BTOKENMARKETSPREADTOOLOW is a free data retrieval call binding the contract method 0x17be6132.
//
// Solidity: function BTOKEN_MARKET_SPREAD_TOO_LOW() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCaller) BTOKENMARKETSPREADTOOLOW(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BTokenErrorCodes.contract.Call(opts, &out, "BTOKEN_MARKET_SPREAD_TOO_LOW")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BTOKENMARKETSPREADTOOLOW is a free data retrieval call binding the contract method 0x17be6132.
//
// Solidity: function BTOKEN_MARKET_SPREAD_TOO_LOW() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesSession) BTOKENMARKETSPREADTOOLOW() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENMARKETSPREADTOOLOW(&_BTokenErrorCodes.CallOpts)
}

// BTOKENMARKETSPREADTOOLOW is a free data retrieval call binding the contract method 0x17be6132.
//
// Solidity: function BTOKEN_MARKET_SPREAD_TOO_LOW() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCallerSession) BTOKENMARKETSPREADTOOLOW() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENMARKETSPREADTOOLOW(&_BTokenErrorCodes.CallOpts)
}

// BTOKENMINIMUMBURNNOTMET is a free data retrieval call binding the contract method 0x42e745e4.
//
// Solidity: function BTOKEN_MINIMUM_BURN_NOT_MET() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCaller) BTOKENMINIMUMBURNNOTMET(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BTokenErrorCodes.contract.Call(opts, &out, "BTOKEN_MINIMUM_BURN_NOT_MET")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BTOKENMINIMUMBURNNOTMET is a free data retrieval call binding the contract method 0x42e745e4.
//
// Solidity: function BTOKEN_MINIMUM_BURN_NOT_MET() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesSession) BTOKENMINIMUMBURNNOTMET() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENMINIMUMBURNNOTMET(&_BTokenErrorCodes.CallOpts)
}

// BTOKENMINIMUMBURNNOTMET is a free data retrieval call binding the contract method 0x42e745e4.
//
// Solidity: function BTOKEN_MINIMUM_BURN_NOT_MET() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCallerSession) BTOKENMINIMUMBURNNOTMET() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENMINIMUMBURNNOTMET(&_BTokenErrorCodes.CallOpts)
}

// BTOKENMINIMUMMINTNOTMET is a free data retrieval call binding the contract method 0xfc45f4cf.
//
// Solidity: function BTOKEN_MINIMUM_MINT_NOT_MET() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCaller) BTOKENMINIMUMMINTNOTMET(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BTokenErrorCodes.contract.Call(opts, &out, "BTOKEN_MINIMUM_MINT_NOT_MET")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BTOKENMINIMUMMINTNOTMET is a free data retrieval call binding the contract method 0xfc45f4cf.
//
// Solidity: function BTOKEN_MINIMUM_MINT_NOT_MET() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesSession) BTOKENMINIMUMMINTNOTMET() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENMINIMUMMINTNOTMET(&_BTokenErrorCodes.CallOpts)
}

// BTOKENMINIMUMMINTNOTMET is a free data retrieval call binding the contract method 0xfc45f4cf.
//
// Solidity: function BTOKEN_MINIMUM_MINT_NOT_MET() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCallerSession) BTOKENMINIMUMMINTNOTMET() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENMINIMUMMINTNOTMET(&_BTokenErrorCodes.CallOpts)
}

// BTOKENMINTINSUFFICIENTETH is a free data retrieval call binding the contract method 0x2810d142.
//
// Solidity: function BTOKEN_MINT_INSUFFICIENT_ETH() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCaller) BTOKENMINTINSUFFICIENTETH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BTokenErrorCodes.contract.Call(opts, &out, "BTOKEN_MINT_INSUFFICIENT_ETH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BTOKENMINTINSUFFICIENTETH is a free data retrieval call binding the contract method 0x2810d142.
//
// Solidity: function BTOKEN_MINT_INSUFFICIENT_ETH() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesSession) BTOKENMINTINSUFFICIENTETH() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENMINTINSUFFICIENTETH(&_BTokenErrorCodes.CallOpts)
}

// BTOKENMINTINSUFFICIENTETH is a free data retrieval call binding the contract method 0x2810d142.
//
// Solidity: function BTOKEN_MINT_INSUFFICIENT_ETH() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCallerSession) BTOKENMINTINSUFFICIENTETH() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENMINTINSUFFICIENTETH(&_BTokenErrorCodes.CallOpts)
}

// BTOKENSPLITVALUESUMERROR is a free data retrieval call binding the contract method 0xc7f53c4f.
//
// Solidity: function BTOKEN_SPLIT_VALUE_SUM_ERROR() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCaller) BTOKENSPLITVALUESUMERROR(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BTokenErrorCodes.contract.Call(opts, &out, "BTOKEN_SPLIT_VALUE_SUM_ERROR")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BTOKENSPLITVALUESUMERROR is a free data retrieval call binding the contract method 0xc7f53c4f.
//
// Solidity: function BTOKEN_SPLIT_VALUE_SUM_ERROR() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesSession) BTOKENSPLITVALUESUMERROR() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENSPLITVALUESUMERROR(&_BTokenErrorCodes.CallOpts)
}

// BTOKENSPLITVALUESUMERROR is a free data retrieval call binding the contract method 0xc7f53c4f.
//
// Solidity: function BTOKEN_SPLIT_VALUE_SUM_ERROR() view returns(bytes32)
func (_BTokenErrorCodes *BTokenErrorCodesCallerSession) BTOKENSPLITVALUESUMERROR() ([32]byte, error) {
	return _BTokenErrorCodes.Contract.BTOKENSPLITVALUESUMERROR(&_BTokenErrorCodes.CallOpts)
}
