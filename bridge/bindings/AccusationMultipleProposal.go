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

// AccusationMultipleProposalMetaData contains all meta data concerning the AccusationMultipleProposal contract.
var AccusationMultipleProposalMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"Bytes32OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"Bytes32OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ChainIdZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"dataOffset\",\"type\":\"uint256\"}],\"name\":\"DataOffsetOverflow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"HeightZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"LEUint16OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"LEUint16OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"LEUint256OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"LEUint256OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"dataOffset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"NotEnoughBytes\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"dataSectionSize\",\"type\":\"uint16\"}],\"name\":\"SizeThresholdExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PRE_SALT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"signature0_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"pClaims0_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature1_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"pClaims1_\",\"type\":\"bytes\"}],\"name\":\"accuseMultipleProposal\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id_\",\"type\":\"bytes32\"}],\"name\":\"isAccused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// AccusationMultipleProposalABI is the input ABI used to generate the binding from.
// Deprecated: Use AccusationMultipleProposalMetaData.ABI instead.
var AccusationMultipleProposalABI = AccusationMultipleProposalMetaData.ABI

// AccusationMultipleProposal is an auto generated Go binding around an Ethereum contract.
type AccusationMultipleProposal struct {
	AccusationMultipleProposalCaller     // Read-only binding to the contract
	AccusationMultipleProposalTransactor // Write-only binding to the contract
	AccusationMultipleProposalFilterer   // Log filterer for contract events
}

// AccusationMultipleProposalCaller is an auto generated read-only Go binding around an Ethereum contract.
type AccusationMultipleProposalCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AccusationMultipleProposalTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AccusationMultipleProposalTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AccusationMultipleProposalFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AccusationMultipleProposalFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AccusationMultipleProposalSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AccusationMultipleProposalSession struct {
	Contract     *AccusationMultipleProposal // Generic contract binding to set the session for
	CallOpts     bind.CallOpts               // Call options to use throughout this session
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// AccusationMultipleProposalCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AccusationMultipleProposalCallerSession struct {
	Contract *AccusationMultipleProposalCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                     // Call options to use throughout this session
}

// AccusationMultipleProposalTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AccusationMultipleProposalTransactorSession struct {
	Contract     *AccusationMultipleProposalTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                     // Transaction auth options to use throughout this session
}

// AccusationMultipleProposalRaw is an auto generated low-level Go binding around an Ethereum contract.
type AccusationMultipleProposalRaw struct {
	Contract *AccusationMultipleProposal // Generic contract binding to access the raw methods on
}

// AccusationMultipleProposalCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AccusationMultipleProposalCallerRaw struct {
	Contract *AccusationMultipleProposalCaller // Generic read-only contract binding to access the raw methods on
}

// AccusationMultipleProposalTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AccusationMultipleProposalTransactorRaw struct {
	Contract *AccusationMultipleProposalTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAccusationMultipleProposal creates a new instance of AccusationMultipleProposal, bound to a specific deployed contract.
func NewAccusationMultipleProposal(address common.Address, backend bind.ContractBackend) (*AccusationMultipleProposal, error) {
	contract, err := bindAccusationMultipleProposal(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AccusationMultipleProposal{AccusationMultipleProposalCaller: AccusationMultipleProposalCaller{contract: contract}, AccusationMultipleProposalTransactor: AccusationMultipleProposalTransactor{contract: contract}, AccusationMultipleProposalFilterer: AccusationMultipleProposalFilterer{contract: contract}}, nil
}

// NewAccusationMultipleProposalCaller creates a new read-only instance of AccusationMultipleProposal, bound to a specific deployed contract.
func NewAccusationMultipleProposalCaller(address common.Address, caller bind.ContractCaller) (*AccusationMultipleProposalCaller, error) {
	contract, err := bindAccusationMultipleProposal(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AccusationMultipleProposalCaller{contract: contract}, nil
}

// NewAccusationMultipleProposalTransactor creates a new write-only instance of AccusationMultipleProposal, bound to a specific deployed contract.
func NewAccusationMultipleProposalTransactor(address common.Address, transactor bind.ContractTransactor) (*AccusationMultipleProposalTransactor, error) {
	contract, err := bindAccusationMultipleProposal(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AccusationMultipleProposalTransactor{contract: contract}, nil
}

// NewAccusationMultipleProposalFilterer creates a new log filterer instance of AccusationMultipleProposal, bound to a specific deployed contract.
func NewAccusationMultipleProposalFilterer(address common.Address, filterer bind.ContractFilterer) (*AccusationMultipleProposalFilterer, error) {
	contract, err := bindAccusationMultipleProposal(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AccusationMultipleProposalFilterer{contract: contract}, nil
}

// bindAccusationMultipleProposal binds a generic wrapper to an already deployed contract.
func bindAccusationMultipleProposal(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AccusationMultipleProposalABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AccusationMultipleProposal *AccusationMultipleProposalRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AccusationMultipleProposal.Contract.AccusationMultipleProposalCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AccusationMultipleProposal *AccusationMultipleProposalRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AccusationMultipleProposal.Contract.AccusationMultipleProposalTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AccusationMultipleProposal *AccusationMultipleProposalRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AccusationMultipleProposal.Contract.AccusationMultipleProposalTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AccusationMultipleProposal *AccusationMultipleProposalCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AccusationMultipleProposal.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AccusationMultipleProposal *AccusationMultipleProposalTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AccusationMultipleProposal.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AccusationMultipleProposal *AccusationMultipleProposalTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AccusationMultipleProposal.Contract.contract.Transact(opts, method, params...)
}

// PRESALT is a free data retrieval call binding the contract method 0x9abd83dc.
//
// Solidity: function PRE_SALT() view returns(bytes32)
func (_AccusationMultipleProposal *AccusationMultipleProposalCaller) PRESALT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AccusationMultipleProposal.contract.Call(opts, &out, "PRE_SALT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PRESALT is a free data retrieval call binding the contract method 0x9abd83dc.
//
// Solidity: function PRE_SALT() view returns(bytes32)
func (_AccusationMultipleProposal *AccusationMultipleProposalSession) PRESALT() ([32]byte, error) {
	return _AccusationMultipleProposal.Contract.PRESALT(&_AccusationMultipleProposal.CallOpts)
}

// PRESALT is a free data retrieval call binding the contract method 0x9abd83dc.
//
// Solidity: function PRE_SALT() view returns(bytes32)
func (_AccusationMultipleProposal *AccusationMultipleProposalCallerSession) PRESALT() ([32]byte, error) {
	return _AccusationMultipleProposal.Contract.PRESALT(&_AccusationMultipleProposal.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_AccusationMultipleProposal *AccusationMultipleProposalCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _AccusationMultipleProposal.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_AccusationMultipleProposal *AccusationMultipleProposalSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _AccusationMultipleProposal.Contract.GetMetamorphicContractAddress(&_AccusationMultipleProposal.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_AccusationMultipleProposal *AccusationMultipleProposalCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _AccusationMultipleProposal.Contract.GetMetamorphicContractAddress(&_AccusationMultipleProposal.CallOpts, _salt, _factory)
}

// IsAccused is a free data retrieval call binding the contract method 0x5e773967.
//
// Solidity: function isAccused(bytes32 id_) view returns(bool)
func (_AccusationMultipleProposal *AccusationMultipleProposalCaller) IsAccused(opts *bind.CallOpts, id_ [32]byte) (bool, error) {
	var out []interface{}
	err := _AccusationMultipleProposal.contract.Call(opts, &out, "isAccused", id_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAccused is a free data retrieval call binding the contract method 0x5e773967.
//
// Solidity: function isAccused(bytes32 id_) view returns(bool)
func (_AccusationMultipleProposal *AccusationMultipleProposalSession) IsAccused(id_ [32]byte) (bool, error) {
	return _AccusationMultipleProposal.Contract.IsAccused(&_AccusationMultipleProposal.CallOpts, id_)
}

// IsAccused is a free data retrieval call binding the contract method 0x5e773967.
//
// Solidity: function isAccused(bytes32 id_) view returns(bool)
func (_AccusationMultipleProposal *AccusationMultipleProposalCallerSession) IsAccused(id_ [32]byte) (bool, error) {
	return _AccusationMultipleProposal.Contract.IsAccused(&_AccusationMultipleProposal.CallOpts, id_)
}

// AccuseMultipleProposal is a paid mutator transaction binding the contract method 0xdfd94cf9.
//
// Solidity: function accuseMultipleProposal(bytes signature0_, bytes pClaims0_, bytes signature1_, bytes pClaims1_) returns(address)
func (_AccusationMultipleProposal *AccusationMultipleProposalTransactor) AccuseMultipleProposal(opts *bind.TransactOpts, signature0_ []byte, pClaims0_ []byte, signature1_ []byte, pClaims1_ []byte) (*types.Transaction, error) {
	return _AccusationMultipleProposal.contract.Transact(opts, "accuseMultipleProposal", signature0_, pClaims0_, signature1_, pClaims1_)
}

// AccuseMultipleProposal is a paid mutator transaction binding the contract method 0xdfd94cf9.
//
// Solidity: function accuseMultipleProposal(bytes signature0_, bytes pClaims0_, bytes signature1_, bytes pClaims1_) returns(address)
func (_AccusationMultipleProposal *AccusationMultipleProposalSession) AccuseMultipleProposal(signature0_ []byte, pClaims0_ []byte, signature1_ []byte, pClaims1_ []byte) (*types.Transaction, error) {
	return _AccusationMultipleProposal.Contract.AccuseMultipleProposal(&_AccusationMultipleProposal.TransactOpts, signature0_, pClaims0_, signature1_, pClaims1_)
}

// AccuseMultipleProposal is a paid mutator transaction binding the contract method 0xdfd94cf9.
//
// Solidity: function accuseMultipleProposal(bytes signature0_, bytes pClaims0_, bytes signature1_, bytes pClaims1_) returns(address)
func (_AccusationMultipleProposal *AccusationMultipleProposalTransactorSession) AccuseMultipleProposal(signature0_ []byte, pClaims0_ []byte, signature1_ []byte, pClaims1_ []byte) (*types.Transaction, error) {
	return _AccusationMultipleProposal.Contract.AccuseMultipleProposal(&_AccusationMultipleProposal.TransactOpts, signature0_, pClaims0_, signature1_, pClaims1_)
}
