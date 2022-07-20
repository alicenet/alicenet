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

// MultipleProposalAccusationMetaData contains all meta data concerning the MultipleProposalAccusation contract.
var MultipleProposalAccusationMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"signature0_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"pClaims0_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature1_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"pClaims1_\",\"type\":\"bytes\"}],\"name\":\"AccuseMultipleProposal\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id_\",\"type\":\"bytes32\"}],\"name\":\"isAccused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// MultipleProposalAccusationABI is the input ABI used to generate the binding from.
// Deprecated: Use MultipleProposalAccusationMetaData.ABI instead.
var MultipleProposalAccusationABI = MultipleProposalAccusationMetaData.ABI

// MultipleProposalAccusation is an auto generated Go binding around an Ethereum contract.
type MultipleProposalAccusation struct {
	MultipleProposalAccusationCaller     // Read-only binding to the contract
	MultipleProposalAccusationTransactor // Write-only binding to the contract
	MultipleProposalAccusationFilterer   // Log filterer for contract events
}

// MultipleProposalAccusationCaller is an auto generated read-only Go binding around an Ethereum contract.
type MultipleProposalAccusationCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultipleProposalAccusationTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MultipleProposalAccusationTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultipleProposalAccusationFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MultipleProposalAccusationFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MultipleProposalAccusationSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MultipleProposalAccusationSession struct {
	Contract     *MultipleProposalAccusation // Generic contract binding to set the session for
	CallOpts     bind.CallOpts               // Call options to use throughout this session
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// MultipleProposalAccusationCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MultipleProposalAccusationCallerSession struct {
	Contract *MultipleProposalAccusationCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                     // Call options to use throughout this session
}

// MultipleProposalAccusationTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MultipleProposalAccusationTransactorSession struct {
	Contract     *MultipleProposalAccusationTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                     // Transaction auth options to use throughout this session
}

// MultipleProposalAccusationRaw is an auto generated low-level Go binding around an Ethereum contract.
type MultipleProposalAccusationRaw struct {
	Contract *MultipleProposalAccusation // Generic contract binding to access the raw methods on
}

// MultipleProposalAccusationCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MultipleProposalAccusationCallerRaw struct {
	Contract *MultipleProposalAccusationCaller // Generic read-only contract binding to access the raw methods on
}

// MultipleProposalAccusationTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MultipleProposalAccusationTransactorRaw struct {
	Contract *MultipleProposalAccusationTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMultipleProposalAccusation creates a new instance of MultipleProposalAccusation, bound to a specific deployed contract.
func NewMultipleProposalAccusation(address common.Address, backend bind.ContractBackend) (*MultipleProposalAccusation, error) {
	contract, err := bindMultipleProposalAccusation(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MultipleProposalAccusation{MultipleProposalAccusationCaller: MultipleProposalAccusationCaller{contract: contract}, MultipleProposalAccusationTransactor: MultipleProposalAccusationTransactor{contract: contract}, MultipleProposalAccusationFilterer: MultipleProposalAccusationFilterer{contract: contract}}, nil
}

// NewMultipleProposalAccusationCaller creates a new read-only instance of MultipleProposalAccusation, bound to a specific deployed contract.
func NewMultipleProposalAccusationCaller(address common.Address, caller bind.ContractCaller) (*MultipleProposalAccusationCaller, error) {
	contract, err := bindMultipleProposalAccusation(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MultipleProposalAccusationCaller{contract: contract}, nil
}

// NewMultipleProposalAccusationTransactor creates a new write-only instance of MultipleProposalAccusation, bound to a specific deployed contract.
func NewMultipleProposalAccusationTransactor(address common.Address, transactor bind.ContractTransactor) (*MultipleProposalAccusationTransactor, error) {
	contract, err := bindMultipleProposalAccusation(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MultipleProposalAccusationTransactor{contract: contract}, nil
}

// NewMultipleProposalAccusationFilterer creates a new log filterer instance of MultipleProposalAccusation, bound to a specific deployed contract.
func NewMultipleProposalAccusationFilterer(address common.Address, filterer bind.ContractFilterer) (*MultipleProposalAccusationFilterer, error) {
	contract, err := bindMultipleProposalAccusation(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MultipleProposalAccusationFilterer{contract: contract}, nil
}

// bindMultipleProposalAccusation binds a generic wrapper to an already deployed contract.
func bindMultipleProposalAccusation(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MultipleProposalAccusationABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultipleProposalAccusation *MultipleProposalAccusationRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MultipleProposalAccusation.Contract.MultipleProposalAccusationCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultipleProposalAccusation *MultipleProposalAccusationRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultipleProposalAccusation.Contract.MultipleProposalAccusationTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultipleProposalAccusation *MultipleProposalAccusationRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultipleProposalAccusation.Contract.MultipleProposalAccusationTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MultipleProposalAccusation *MultipleProposalAccusationCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MultipleProposalAccusation.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MultipleProposalAccusation *MultipleProposalAccusationTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MultipleProposalAccusation.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MultipleProposalAccusation *MultipleProposalAccusationTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MultipleProposalAccusation.Contract.contract.Transact(opts, method, params...)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_MultipleProposalAccusation *MultipleProposalAccusationCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _MultipleProposalAccusation.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_MultipleProposalAccusation *MultipleProposalAccusationSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _MultipleProposalAccusation.Contract.GetMetamorphicContractAddress(&_MultipleProposalAccusation.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_MultipleProposalAccusation *MultipleProposalAccusationCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _MultipleProposalAccusation.Contract.GetMetamorphicContractAddress(&_MultipleProposalAccusation.CallOpts, _salt, _factory)
}

// IsAccused is a free data retrieval call binding the contract method 0x5e773967.
//
// Solidity: function isAccused(bytes32 id_) view returns(bool)
func (_MultipleProposalAccusation *MultipleProposalAccusationCaller) IsAccused(opts *bind.CallOpts, id_ [32]byte) (bool, error) {
	var out []interface{}
	err := _MultipleProposalAccusation.contract.Call(opts, &out, "isAccused", id_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAccused is a free data retrieval call binding the contract method 0x5e773967.
//
// Solidity: function isAccused(bytes32 id_) view returns(bool)
func (_MultipleProposalAccusation *MultipleProposalAccusationSession) IsAccused(id_ [32]byte) (bool, error) {
	return _MultipleProposalAccusation.Contract.IsAccused(&_MultipleProposalAccusation.CallOpts, id_)
}

// IsAccused is a free data retrieval call binding the contract method 0x5e773967.
//
// Solidity: function isAccused(bytes32 id_) view returns(bool)
func (_MultipleProposalAccusation *MultipleProposalAccusationCallerSession) IsAccused(id_ [32]byte) (bool, error) {
	return _MultipleProposalAccusation.Contract.IsAccused(&_MultipleProposalAccusation.CallOpts, id_)
}

// AccuseMultipleProposal is a paid mutator transaction binding the contract method 0x7f321b2d.
//
// Solidity: function AccuseMultipleProposal(bytes signature0_, bytes pClaims0_, bytes signature1_, bytes pClaims1_) returns(address)
func (_MultipleProposalAccusation *MultipleProposalAccusationTransactor) AccuseMultipleProposal(opts *bind.TransactOpts, signature0_ []byte, pClaims0_ []byte, signature1_ []byte, pClaims1_ []byte) (*types.Transaction, error) {
	return _MultipleProposalAccusation.contract.Transact(opts, "AccuseMultipleProposal", signature0_, pClaims0_, signature1_, pClaims1_)
}

// AccuseMultipleProposal is a paid mutator transaction binding the contract method 0x7f321b2d.
//
// Solidity: function AccuseMultipleProposal(bytes signature0_, bytes pClaims0_, bytes signature1_, bytes pClaims1_) returns(address)
func (_MultipleProposalAccusation *MultipleProposalAccusationSession) AccuseMultipleProposal(signature0_ []byte, pClaims0_ []byte, signature1_ []byte, pClaims1_ []byte) (*types.Transaction, error) {
	return _MultipleProposalAccusation.Contract.AccuseMultipleProposal(&_MultipleProposalAccusation.TransactOpts, signature0_, pClaims0_, signature1_, pClaims1_)
}

// AccuseMultipleProposal is a paid mutator transaction binding the contract method 0x7f321b2d.
//
// Solidity: function AccuseMultipleProposal(bytes signature0_, bytes pClaims0_, bytes signature1_, bytes pClaims1_) returns(address)
func (_MultipleProposalAccusation *MultipleProposalAccusationTransactorSession) AccuseMultipleProposal(signature0_ []byte, pClaims0_ []byte, signature1_ []byte, pClaims1_ []byte) (*types.Transaction, error) {
	return _MultipleProposalAccusation.Contract.AccuseMultipleProposal(&_MultipleProposalAccusation.TransactOpts, signature0_, pClaims0_, signature1_, pClaims1_)
}
