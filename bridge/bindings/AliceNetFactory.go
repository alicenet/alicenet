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

// AliceNetFactoryMetaData contains all meta data concerning the AliceNetFactory contract.
var AliceNetFactoryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"selfAddr_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"name\":\"Deployed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"name\":\"DeployedProxy\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"name\":\"DeployedRaw\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"name\":\"DeployedStatic\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"name\":\"DeployedTemplate\",\"type\":\"event\"},{\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value_\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"cdata_\",\"type\":\"bytes\"}],\"name\":\"callAny\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"contracts\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"contracts_\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target_\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"cdata_\",\"type\":\"bytes\"}],\"name\":\"delegateCallAny\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"delegator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"delegator_\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"deployCode_\",\"type\":\"bytes\"}],\"name\":\"deployCreate\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value_\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"salt_\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"deployCode_\",\"type\":\"bytes\"}],\"name\":\"deployCreate2\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"salt_\",\"type\":\"bytes32\"}],\"name\":\"deployProxy\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"salt_\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"initCallData_\",\"type\":\"bytes\"}],\"name\":\"deployStatic\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"deployCode_\",\"type\":\"bytes\"}],\"name\":\"deployTemplate\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getImplementation\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNumContracts\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"contract_\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"initCallData_\",\"type\":\"bytes\"}],\"name\":\"initializeContract\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"salt_\",\"type\":\"bytes32\"}],\"name\":\"lookup\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes[]\",\"name\":\"cdata_\",\"type\":\"bytes[]\"}],\"name\":\"multiCall\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"owner_\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newDelegator_\",\"type\":\"address\"}],\"name\":\"setDelegator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementationAddress_\",\"type\":\"address\"}],\"name\":\"setImplementation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner_\",\"type\":\"address\"}],\"name\":\"setOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"salt_\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"newImpl_\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"initCallData_\",\"type\":\"bytes\"}],\"name\":\"upgradeProxy\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// AliceNetFactoryABI is the input ABI used to generate the binding from.
// Deprecated: Use AliceNetFactoryMetaData.ABI instead.
var AliceNetFactoryABI = AliceNetFactoryMetaData.ABI

// AliceNetFactory is an auto generated Go binding around an Ethereum contract.
type AliceNetFactory struct {
	AliceNetFactoryCaller     // Read-only binding to the contract
	AliceNetFactoryTransactor // Write-only binding to the contract
	AliceNetFactoryFilterer   // Log filterer for contract events
}

// AliceNetFactoryCaller is an auto generated read-only Go binding around an Ethereum contract.
type AliceNetFactoryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AliceNetFactoryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AliceNetFactoryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AliceNetFactoryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AliceNetFactoryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AliceNetFactorySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AliceNetFactorySession struct {
	Contract     *AliceNetFactory  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AliceNetFactoryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AliceNetFactoryCallerSession struct {
	Contract *AliceNetFactoryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// AliceNetFactoryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AliceNetFactoryTransactorSession struct {
	Contract     *AliceNetFactoryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// AliceNetFactoryRaw is an auto generated low-level Go binding around an Ethereum contract.
type AliceNetFactoryRaw struct {
	Contract *AliceNetFactory // Generic contract binding to access the raw methods on
}

// AliceNetFactoryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AliceNetFactoryCallerRaw struct {
	Contract *AliceNetFactoryCaller // Generic read-only contract binding to access the raw methods on
}

// AliceNetFactoryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AliceNetFactoryTransactorRaw struct {
	Contract *AliceNetFactoryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAliceNetFactory creates a new instance of AliceNetFactory, bound to a specific deployed contract.
func NewAliceNetFactory(address common.Address, backend bind.ContractBackend) (*AliceNetFactory, error) {
	contract, err := bindAliceNetFactory(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AliceNetFactory{AliceNetFactoryCaller: AliceNetFactoryCaller{contract: contract}, AliceNetFactoryTransactor: AliceNetFactoryTransactor{contract: contract}, AliceNetFactoryFilterer: AliceNetFactoryFilterer{contract: contract}}, nil
}

// NewAliceNetFactoryCaller creates a new read-only instance of AliceNetFactory, bound to a specific deployed contract.
func NewAliceNetFactoryCaller(address common.Address, caller bind.ContractCaller) (*AliceNetFactoryCaller, error) {
	contract, err := bindAliceNetFactory(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AliceNetFactoryCaller{contract: contract}, nil
}

// NewAliceNetFactoryTransactor creates a new write-only instance of AliceNetFactory, bound to a specific deployed contract.
func NewAliceNetFactoryTransactor(address common.Address, transactor bind.ContractTransactor) (*AliceNetFactoryTransactor, error) {
	contract, err := bindAliceNetFactory(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AliceNetFactoryTransactor{contract: contract}, nil
}

// NewAliceNetFactoryFilterer creates a new log filterer instance of AliceNetFactory, bound to a specific deployed contract.
func NewAliceNetFactoryFilterer(address common.Address, filterer bind.ContractFilterer) (*AliceNetFactoryFilterer, error) {
	contract, err := bindAliceNetFactory(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AliceNetFactoryFilterer{contract: contract}, nil
}

// bindAliceNetFactory binds a generic wrapper to an already deployed contract.
func bindAliceNetFactory(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AliceNetFactoryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AliceNetFactory *AliceNetFactoryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AliceNetFactory.Contract.AliceNetFactoryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AliceNetFactory *AliceNetFactoryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.AliceNetFactoryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AliceNetFactory *AliceNetFactoryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.AliceNetFactoryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AliceNetFactory *AliceNetFactoryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AliceNetFactory.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AliceNetFactory *AliceNetFactoryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AliceNetFactory *AliceNetFactoryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.contract.Transact(opts, method, params...)
}

// Contracts is a free data retrieval call binding the contract method 0x6c0f79b6.
//
// Solidity: function contracts() view returns(bytes32[] contracts_)
func (_AliceNetFactory *AliceNetFactoryCaller) Contracts(opts *bind.CallOpts) ([][32]byte, error) {
	var out []interface{}
	err := _AliceNetFactory.contract.Call(opts, &out, "contracts")

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// Contracts is a free data retrieval call binding the contract method 0x6c0f79b6.
//
// Solidity: function contracts() view returns(bytes32[] contracts_)
func (_AliceNetFactory *AliceNetFactorySession) Contracts() ([][32]byte, error) {
	return _AliceNetFactory.Contract.Contracts(&_AliceNetFactory.CallOpts)
}

// Contracts is a free data retrieval call binding the contract method 0x6c0f79b6.
//
// Solidity: function contracts() view returns(bytes32[] contracts_)
func (_AliceNetFactory *AliceNetFactoryCallerSession) Contracts() ([][32]byte, error) {
	return _AliceNetFactory.Contract.Contracts(&_AliceNetFactory.CallOpts)
}

// Delegator is a free data retrieval call binding the contract method 0xce9b7930.
//
// Solidity: function delegator() view returns(address delegator_)
func (_AliceNetFactory *AliceNetFactoryCaller) Delegator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AliceNetFactory.contract.Call(opts, &out, "delegator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Delegator is a free data retrieval call binding the contract method 0xce9b7930.
//
// Solidity: function delegator() view returns(address delegator_)
func (_AliceNetFactory *AliceNetFactorySession) Delegator() (common.Address, error) {
	return _AliceNetFactory.Contract.Delegator(&_AliceNetFactory.CallOpts)
}

// Delegator is a free data retrieval call binding the contract method 0xce9b7930.
//
// Solidity: function delegator() view returns(address delegator_)
func (_AliceNetFactory *AliceNetFactoryCallerSession) Delegator() (common.Address, error) {
	return _AliceNetFactory.Contract.Delegator(&_AliceNetFactory.CallOpts)
}

// GetImplementation is a free data retrieval call binding the contract method 0xaaf10f42.
//
// Solidity: function getImplementation() view returns(address)
func (_AliceNetFactory *AliceNetFactoryCaller) GetImplementation(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AliceNetFactory.contract.Call(opts, &out, "getImplementation")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetImplementation is a free data retrieval call binding the contract method 0xaaf10f42.
//
// Solidity: function getImplementation() view returns(address)
func (_AliceNetFactory *AliceNetFactorySession) GetImplementation() (common.Address, error) {
	return _AliceNetFactory.Contract.GetImplementation(&_AliceNetFactory.CallOpts)
}

// GetImplementation is a free data retrieval call binding the contract method 0xaaf10f42.
//
// Solidity: function getImplementation() view returns(address)
func (_AliceNetFactory *AliceNetFactoryCallerSession) GetImplementation() (common.Address, error) {
	return _AliceNetFactory.Contract.GetImplementation(&_AliceNetFactory.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_AliceNetFactory *AliceNetFactoryCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _AliceNetFactory.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_AliceNetFactory *AliceNetFactorySession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _AliceNetFactory.Contract.GetMetamorphicContractAddress(&_AliceNetFactory.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_AliceNetFactory *AliceNetFactoryCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _AliceNetFactory.Contract.GetMetamorphicContractAddress(&_AliceNetFactory.CallOpts, _salt, _factory)
}

// GetNumContracts is a free data retrieval call binding the contract method 0xcfe10b30.
//
// Solidity: function getNumContracts() view returns(uint256)
func (_AliceNetFactory *AliceNetFactoryCaller) GetNumContracts(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AliceNetFactory.contract.Call(opts, &out, "getNumContracts")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNumContracts is a free data retrieval call binding the contract method 0xcfe10b30.
//
// Solidity: function getNumContracts() view returns(uint256)
func (_AliceNetFactory *AliceNetFactorySession) GetNumContracts() (*big.Int, error) {
	return _AliceNetFactory.Contract.GetNumContracts(&_AliceNetFactory.CallOpts)
}

// GetNumContracts is a free data retrieval call binding the contract method 0xcfe10b30.
//
// Solidity: function getNumContracts() view returns(uint256)
func (_AliceNetFactory *AliceNetFactoryCallerSession) GetNumContracts() (*big.Int, error) {
	return _AliceNetFactory.Contract.GetNumContracts(&_AliceNetFactory.CallOpts)
}

// Lookup is a free data retrieval call binding the contract method 0xf39ec1f7.
//
// Solidity: function lookup(bytes32 salt_) view returns(address addr)
func (_AliceNetFactory *AliceNetFactoryCaller) Lookup(opts *bind.CallOpts, salt_ [32]byte) (common.Address, error) {
	var out []interface{}
	err := _AliceNetFactory.contract.Call(opts, &out, "lookup", salt_)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Lookup is a free data retrieval call binding the contract method 0xf39ec1f7.
//
// Solidity: function lookup(bytes32 salt_) view returns(address addr)
func (_AliceNetFactory *AliceNetFactorySession) Lookup(salt_ [32]byte) (common.Address, error) {
	return _AliceNetFactory.Contract.Lookup(&_AliceNetFactory.CallOpts, salt_)
}

// Lookup is a free data retrieval call binding the contract method 0xf39ec1f7.
//
// Solidity: function lookup(bytes32 salt_) view returns(address addr)
func (_AliceNetFactory *AliceNetFactoryCallerSession) Lookup(salt_ [32]byte) (common.Address, error) {
	return _AliceNetFactory.Contract.Lookup(&_AliceNetFactory.CallOpts, salt_)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address owner_)
func (_AliceNetFactory *AliceNetFactoryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AliceNetFactory.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address owner_)
func (_AliceNetFactory *AliceNetFactorySession) Owner() (common.Address, error) {
	return _AliceNetFactory.Contract.Owner(&_AliceNetFactory.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address owner_)
func (_AliceNetFactory *AliceNetFactoryCallerSession) Owner() (common.Address, error) {
	return _AliceNetFactory.Contract.Owner(&_AliceNetFactory.CallOpts)
}

// CallAny is a paid mutator transaction binding the contract method 0x12e6bf6a.
//
// Solidity: function callAny(address target_, uint256 value_, bytes cdata_) payable returns()
func (_AliceNetFactory *AliceNetFactoryTransactor) CallAny(opts *bind.TransactOpts, target_ common.Address, value_ *big.Int, cdata_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.contract.Transact(opts, "callAny", target_, value_, cdata_)
}

// CallAny is a paid mutator transaction binding the contract method 0x12e6bf6a.
//
// Solidity: function callAny(address target_, uint256 value_, bytes cdata_) payable returns()
func (_AliceNetFactory *AliceNetFactorySession) CallAny(target_ common.Address, value_ *big.Int, cdata_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.CallAny(&_AliceNetFactory.TransactOpts, target_, value_, cdata_)
}

// CallAny is a paid mutator transaction binding the contract method 0x12e6bf6a.
//
// Solidity: function callAny(address target_, uint256 value_, bytes cdata_) payable returns()
func (_AliceNetFactory *AliceNetFactoryTransactorSession) CallAny(target_ common.Address, value_ *big.Int, cdata_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.CallAny(&_AliceNetFactory.TransactOpts, target_, value_, cdata_)
}

// DelegateCallAny is a paid mutator transaction binding the contract method 0x4713ee7a.
//
// Solidity: function delegateCallAny(address target_, bytes cdata_) payable returns()
func (_AliceNetFactory *AliceNetFactoryTransactor) DelegateCallAny(opts *bind.TransactOpts, target_ common.Address, cdata_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.contract.Transact(opts, "delegateCallAny", target_, cdata_)
}

// DelegateCallAny is a paid mutator transaction binding the contract method 0x4713ee7a.
//
// Solidity: function delegateCallAny(address target_, bytes cdata_) payable returns()
func (_AliceNetFactory *AliceNetFactorySession) DelegateCallAny(target_ common.Address, cdata_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.DelegateCallAny(&_AliceNetFactory.TransactOpts, target_, cdata_)
}

// DelegateCallAny is a paid mutator transaction binding the contract method 0x4713ee7a.
//
// Solidity: function delegateCallAny(address target_, bytes cdata_) payable returns()
func (_AliceNetFactory *AliceNetFactoryTransactorSession) DelegateCallAny(target_ common.Address, cdata_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.DelegateCallAny(&_AliceNetFactory.TransactOpts, target_, cdata_)
}

// DeployCreate is a paid mutator transaction binding the contract method 0x27fe1822.
//
// Solidity: function deployCreate(bytes deployCode_) returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryTransactor) DeployCreate(opts *bind.TransactOpts, deployCode_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.contract.Transact(opts, "deployCreate", deployCode_)
}

// DeployCreate is a paid mutator transaction binding the contract method 0x27fe1822.
//
// Solidity: function deployCreate(bytes deployCode_) returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactorySession) DeployCreate(deployCode_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.DeployCreate(&_AliceNetFactory.TransactOpts, deployCode_)
}

// DeployCreate is a paid mutator transaction binding the contract method 0x27fe1822.
//
// Solidity: function deployCreate(bytes deployCode_) returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryTransactorSession) DeployCreate(deployCode_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.DeployCreate(&_AliceNetFactory.TransactOpts, deployCode_)
}

// DeployCreate2 is a paid mutator transaction binding the contract method 0x56f2a761.
//
// Solidity: function deployCreate2(uint256 value_, bytes32 salt_, bytes deployCode_) payable returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryTransactor) DeployCreate2(opts *bind.TransactOpts, value_ *big.Int, salt_ [32]byte, deployCode_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.contract.Transact(opts, "deployCreate2", value_, salt_, deployCode_)
}

// DeployCreate2 is a paid mutator transaction binding the contract method 0x56f2a761.
//
// Solidity: function deployCreate2(uint256 value_, bytes32 salt_, bytes deployCode_) payable returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactorySession) DeployCreate2(value_ *big.Int, salt_ [32]byte, deployCode_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.DeployCreate2(&_AliceNetFactory.TransactOpts, value_, salt_, deployCode_)
}

// DeployCreate2 is a paid mutator transaction binding the contract method 0x56f2a761.
//
// Solidity: function deployCreate2(uint256 value_, bytes32 salt_, bytes deployCode_) payable returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryTransactorSession) DeployCreate2(value_ *big.Int, salt_ [32]byte, deployCode_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.DeployCreate2(&_AliceNetFactory.TransactOpts, value_, salt_, deployCode_)
}

// DeployProxy is a paid mutator transaction binding the contract method 0x39cab472.
//
// Solidity: function deployProxy(bytes32 salt_) returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryTransactor) DeployProxy(opts *bind.TransactOpts, salt_ [32]byte) (*types.Transaction, error) {
	return _AliceNetFactory.contract.Transact(opts, "deployProxy", salt_)
}

// DeployProxy is a paid mutator transaction binding the contract method 0x39cab472.
//
// Solidity: function deployProxy(bytes32 salt_) returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactorySession) DeployProxy(salt_ [32]byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.DeployProxy(&_AliceNetFactory.TransactOpts, salt_)
}

// DeployProxy is a paid mutator transaction binding the contract method 0x39cab472.
//
// Solidity: function deployProxy(bytes32 salt_) returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryTransactorSession) DeployProxy(salt_ [32]byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.DeployProxy(&_AliceNetFactory.TransactOpts, salt_)
}

// DeployStatic is a paid mutator transaction binding the contract method 0xfa481da5.
//
// Solidity: function deployStatic(bytes32 salt_, bytes initCallData_) returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryTransactor) DeployStatic(opts *bind.TransactOpts, salt_ [32]byte, initCallData_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.contract.Transact(opts, "deployStatic", salt_, initCallData_)
}

// DeployStatic is a paid mutator transaction binding the contract method 0xfa481da5.
//
// Solidity: function deployStatic(bytes32 salt_, bytes initCallData_) returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactorySession) DeployStatic(salt_ [32]byte, initCallData_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.DeployStatic(&_AliceNetFactory.TransactOpts, salt_, initCallData_)
}

// DeployStatic is a paid mutator transaction binding the contract method 0xfa481da5.
//
// Solidity: function deployStatic(bytes32 salt_, bytes initCallData_) returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryTransactorSession) DeployStatic(salt_ [32]byte, initCallData_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.DeployStatic(&_AliceNetFactory.TransactOpts, salt_, initCallData_)
}

// DeployTemplate is a paid mutator transaction binding the contract method 0x17cff2c5.
//
// Solidity: function deployTemplate(bytes deployCode_) returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryTransactor) DeployTemplate(opts *bind.TransactOpts, deployCode_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.contract.Transact(opts, "deployTemplate", deployCode_)
}

// DeployTemplate is a paid mutator transaction binding the contract method 0x17cff2c5.
//
// Solidity: function deployTemplate(bytes deployCode_) returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactorySession) DeployTemplate(deployCode_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.DeployTemplate(&_AliceNetFactory.TransactOpts, deployCode_)
}

// DeployTemplate is a paid mutator transaction binding the contract method 0x17cff2c5.
//
// Solidity: function deployTemplate(bytes deployCode_) returns(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryTransactorSession) DeployTemplate(deployCode_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.DeployTemplate(&_AliceNetFactory.TransactOpts, deployCode_)
}

// InitializeContract is a paid mutator transaction binding the contract method 0xe1d7a8e4.
//
// Solidity: function initializeContract(address contract_, bytes initCallData_) returns()
func (_AliceNetFactory *AliceNetFactoryTransactor) InitializeContract(opts *bind.TransactOpts, contract_ common.Address, initCallData_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.contract.Transact(opts, "initializeContract", contract_, initCallData_)
}

// InitializeContract is a paid mutator transaction binding the contract method 0xe1d7a8e4.
//
// Solidity: function initializeContract(address contract_, bytes initCallData_) returns()
func (_AliceNetFactory *AliceNetFactorySession) InitializeContract(contract_ common.Address, initCallData_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.InitializeContract(&_AliceNetFactory.TransactOpts, contract_, initCallData_)
}

// InitializeContract is a paid mutator transaction binding the contract method 0xe1d7a8e4.
//
// Solidity: function initializeContract(address contract_, bytes initCallData_) returns()
func (_AliceNetFactory *AliceNetFactoryTransactorSession) InitializeContract(contract_ common.Address, initCallData_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.InitializeContract(&_AliceNetFactory.TransactOpts, contract_, initCallData_)
}

// MultiCall is a paid mutator transaction binding the contract method 0x348a0cdc.
//
// Solidity: function multiCall(bytes[] cdata_) returns()
func (_AliceNetFactory *AliceNetFactoryTransactor) MultiCall(opts *bind.TransactOpts, cdata_ [][]byte) (*types.Transaction, error) {
	return _AliceNetFactory.contract.Transact(opts, "multiCall", cdata_)
}

// MultiCall is a paid mutator transaction binding the contract method 0x348a0cdc.
//
// Solidity: function multiCall(bytes[] cdata_) returns()
func (_AliceNetFactory *AliceNetFactorySession) MultiCall(cdata_ [][]byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.MultiCall(&_AliceNetFactory.TransactOpts, cdata_)
}

// MultiCall is a paid mutator transaction binding the contract method 0x348a0cdc.
//
// Solidity: function multiCall(bytes[] cdata_) returns()
func (_AliceNetFactory *AliceNetFactoryTransactorSession) MultiCall(cdata_ [][]byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.MultiCall(&_AliceNetFactory.TransactOpts, cdata_)
}

// SetDelegator is a paid mutator transaction binding the contract method 0x83cd9cc3.
//
// Solidity: function setDelegator(address newDelegator_) returns()
func (_AliceNetFactory *AliceNetFactoryTransactor) SetDelegator(opts *bind.TransactOpts, newDelegator_ common.Address) (*types.Transaction, error) {
	return _AliceNetFactory.contract.Transact(opts, "setDelegator", newDelegator_)
}

// SetDelegator is a paid mutator transaction binding the contract method 0x83cd9cc3.
//
// Solidity: function setDelegator(address newDelegator_) returns()
func (_AliceNetFactory *AliceNetFactorySession) SetDelegator(newDelegator_ common.Address) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.SetDelegator(&_AliceNetFactory.TransactOpts, newDelegator_)
}

// SetDelegator is a paid mutator transaction binding the contract method 0x83cd9cc3.
//
// Solidity: function setDelegator(address newDelegator_) returns()
func (_AliceNetFactory *AliceNetFactoryTransactorSession) SetDelegator(newDelegator_ common.Address) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.SetDelegator(&_AliceNetFactory.TransactOpts, newDelegator_)
}

// SetImplementation is a paid mutator transaction binding the contract method 0xd784d426.
//
// Solidity: function setImplementation(address newImplementationAddress_) returns()
func (_AliceNetFactory *AliceNetFactoryTransactor) SetImplementation(opts *bind.TransactOpts, newImplementationAddress_ common.Address) (*types.Transaction, error) {
	return _AliceNetFactory.contract.Transact(opts, "setImplementation", newImplementationAddress_)
}

// SetImplementation is a paid mutator transaction binding the contract method 0xd784d426.
//
// Solidity: function setImplementation(address newImplementationAddress_) returns()
func (_AliceNetFactory *AliceNetFactorySession) SetImplementation(newImplementationAddress_ common.Address) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.SetImplementation(&_AliceNetFactory.TransactOpts, newImplementationAddress_)
}

// SetImplementation is a paid mutator transaction binding the contract method 0xd784d426.
//
// Solidity: function setImplementation(address newImplementationAddress_) returns()
func (_AliceNetFactory *AliceNetFactoryTransactorSession) SetImplementation(newImplementationAddress_ common.Address) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.SetImplementation(&_AliceNetFactory.TransactOpts, newImplementationAddress_)
}

// SetOwner is a paid mutator transaction binding the contract method 0x13af4035.
//
// Solidity: function setOwner(address newOwner_) returns()
func (_AliceNetFactory *AliceNetFactoryTransactor) SetOwner(opts *bind.TransactOpts, newOwner_ common.Address) (*types.Transaction, error) {
	return _AliceNetFactory.contract.Transact(opts, "setOwner", newOwner_)
}

// SetOwner is a paid mutator transaction binding the contract method 0x13af4035.
//
// Solidity: function setOwner(address newOwner_) returns()
func (_AliceNetFactory *AliceNetFactorySession) SetOwner(newOwner_ common.Address) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.SetOwner(&_AliceNetFactory.TransactOpts, newOwner_)
}

// SetOwner is a paid mutator transaction binding the contract method 0x13af4035.
//
// Solidity: function setOwner(address newOwner_) returns()
func (_AliceNetFactory *AliceNetFactoryTransactorSession) SetOwner(newOwner_ common.Address) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.SetOwner(&_AliceNetFactory.TransactOpts, newOwner_)
}

// UpgradeProxy is a paid mutator transaction binding the contract method 0x043c9414.
//
// Solidity: function upgradeProxy(bytes32 salt_, address newImpl_, bytes initCallData_) returns()
func (_AliceNetFactory *AliceNetFactoryTransactor) UpgradeProxy(opts *bind.TransactOpts, salt_ [32]byte, newImpl_ common.Address, initCallData_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.contract.Transact(opts, "upgradeProxy", salt_, newImpl_, initCallData_)
}

// UpgradeProxy is a paid mutator transaction binding the contract method 0x043c9414.
//
// Solidity: function upgradeProxy(bytes32 salt_, address newImpl_, bytes initCallData_) returns()
func (_AliceNetFactory *AliceNetFactorySession) UpgradeProxy(salt_ [32]byte, newImpl_ common.Address, initCallData_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.UpgradeProxy(&_AliceNetFactory.TransactOpts, salt_, newImpl_, initCallData_)
}

// UpgradeProxy is a paid mutator transaction binding the contract method 0x043c9414.
//
// Solidity: function upgradeProxy(bytes32 salt_, address newImpl_, bytes initCallData_) returns()
func (_AliceNetFactory *AliceNetFactoryTransactorSession) UpgradeProxy(salt_ [32]byte, newImpl_ common.Address, initCallData_ []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.UpgradeProxy(&_AliceNetFactory.TransactOpts, salt_, newImpl_, initCallData_)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_AliceNetFactory *AliceNetFactoryTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _AliceNetFactory.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_AliceNetFactory *AliceNetFactorySession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.Fallback(&_AliceNetFactory.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() returns()
func (_AliceNetFactory *AliceNetFactoryTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _AliceNetFactory.Contract.Fallback(&_AliceNetFactory.TransactOpts, calldata)
}

// AliceNetFactoryDeployedIterator is returned from FilterDeployed and is used to iterate over the raw logs and unpacked data for Deployed events raised by the AliceNetFactory contract.
type AliceNetFactoryDeployedIterator struct {
	Event *AliceNetFactoryDeployed // Event containing the contract specifics and raw log

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
func (it *AliceNetFactoryDeployedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AliceNetFactoryDeployed)
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
		it.Event = new(AliceNetFactoryDeployed)
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
func (it *AliceNetFactoryDeployedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AliceNetFactoryDeployedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AliceNetFactoryDeployed represents a Deployed event raised by the AliceNetFactory contract.
type AliceNetFactoryDeployed struct {
	Salt         [32]byte
	ContractAddr common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterDeployed is a free log retrieval operation binding the contract event 0xe491e278e37782abe0872fe7c7b549cd7b0713d0c5c1e84a81899a5fdf32087b.
//
// Solidity: event Deployed(bytes32 salt, address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) FilterDeployed(opts *bind.FilterOpts) (*AliceNetFactoryDeployedIterator, error) {

	logs, sub, err := _AliceNetFactory.contract.FilterLogs(opts, "Deployed")
	if err != nil {
		return nil, err
	}
	return &AliceNetFactoryDeployedIterator{contract: _AliceNetFactory.contract, event: "Deployed", logs: logs, sub: sub}, nil
}

// WatchDeployed is a free log subscription operation binding the contract event 0xe491e278e37782abe0872fe7c7b549cd7b0713d0c5c1e84a81899a5fdf32087b.
//
// Solidity: event Deployed(bytes32 salt, address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) WatchDeployed(opts *bind.WatchOpts, sink chan<- *AliceNetFactoryDeployed) (event.Subscription, error) {

	logs, sub, err := _AliceNetFactory.contract.WatchLogs(opts, "Deployed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AliceNetFactoryDeployed)
				if err := _AliceNetFactory.contract.UnpackLog(event, "Deployed", log); err != nil {
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

// ParseDeployed is a log parse operation binding the contract event 0xe491e278e37782abe0872fe7c7b549cd7b0713d0c5c1e84a81899a5fdf32087b.
//
// Solidity: event Deployed(bytes32 salt, address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) ParseDeployed(log types.Log) (*AliceNetFactoryDeployed, error) {
	event := new(AliceNetFactoryDeployed)
	if err := _AliceNetFactory.contract.UnpackLog(event, "Deployed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AliceNetFactoryDeployedProxyIterator is returned from FilterDeployedProxy and is used to iterate over the raw logs and unpacked data for DeployedProxy events raised by the AliceNetFactory contract.
type AliceNetFactoryDeployedProxyIterator struct {
	Event *AliceNetFactoryDeployedProxy // Event containing the contract specifics and raw log

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
func (it *AliceNetFactoryDeployedProxyIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AliceNetFactoryDeployedProxy)
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
		it.Event = new(AliceNetFactoryDeployedProxy)
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
func (it *AliceNetFactoryDeployedProxyIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AliceNetFactoryDeployedProxyIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AliceNetFactoryDeployedProxy represents a DeployedProxy event raised by the AliceNetFactory contract.
type AliceNetFactoryDeployedProxy struct {
	ContractAddr common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterDeployedProxy is a free log retrieval operation binding the contract event 0x06690e5b52be10a3d5820ec875c3dd3327f3077954a09f104201e40e5f7082c6.
//
// Solidity: event DeployedProxy(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) FilterDeployedProxy(opts *bind.FilterOpts) (*AliceNetFactoryDeployedProxyIterator, error) {

	logs, sub, err := _AliceNetFactory.contract.FilterLogs(opts, "DeployedProxy")
	if err != nil {
		return nil, err
	}
	return &AliceNetFactoryDeployedProxyIterator{contract: _AliceNetFactory.contract, event: "DeployedProxy", logs: logs, sub: sub}, nil
}

// WatchDeployedProxy is a free log subscription operation binding the contract event 0x06690e5b52be10a3d5820ec875c3dd3327f3077954a09f104201e40e5f7082c6.
//
// Solidity: event DeployedProxy(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) WatchDeployedProxy(opts *bind.WatchOpts, sink chan<- *AliceNetFactoryDeployedProxy) (event.Subscription, error) {

	logs, sub, err := _AliceNetFactory.contract.WatchLogs(opts, "DeployedProxy")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AliceNetFactoryDeployedProxy)
				if err := _AliceNetFactory.contract.UnpackLog(event, "DeployedProxy", log); err != nil {
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

// ParseDeployedProxy is a log parse operation binding the contract event 0x06690e5b52be10a3d5820ec875c3dd3327f3077954a09f104201e40e5f7082c6.
//
// Solidity: event DeployedProxy(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) ParseDeployedProxy(log types.Log) (*AliceNetFactoryDeployedProxy, error) {
	event := new(AliceNetFactoryDeployedProxy)
	if err := _AliceNetFactory.contract.UnpackLog(event, "DeployedProxy", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AliceNetFactoryDeployedRawIterator is returned from FilterDeployedRaw and is used to iterate over the raw logs and unpacked data for DeployedRaw events raised by the AliceNetFactory contract.
type AliceNetFactoryDeployedRawIterator struct {
	Event *AliceNetFactoryDeployedRaw // Event containing the contract specifics and raw log

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
func (it *AliceNetFactoryDeployedRawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AliceNetFactoryDeployedRaw)
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
		it.Event = new(AliceNetFactoryDeployedRaw)
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
func (it *AliceNetFactoryDeployedRawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AliceNetFactoryDeployedRawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AliceNetFactoryDeployedRaw represents a DeployedRaw event raised by the AliceNetFactory contract.
type AliceNetFactoryDeployedRaw struct {
	ContractAddr common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterDeployedRaw is a free log retrieval operation binding the contract event 0xd3acf0da590cfcd8f020afd7f40b7e6e4c8bd2fc9eb7aec9836837b667685b3a.
//
// Solidity: event DeployedRaw(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) FilterDeployedRaw(opts *bind.FilterOpts) (*AliceNetFactoryDeployedRawIterator, error) {

	logs, sub, err := _AliceNetFactory.contract.FilterLogs(opts, "DeployedRaw")
	if err != nil {
		return nil, err
	}
	return &AliceNetFactoryDeployedRawIterator{contract: _AliceNetFactory.contract, event: "DeployedRaw", logs: logs, sub: sub}, nil
}

// WatchDeployedRaw is a free log subscription operation binding the contract event 0xd3acf0da590cfcd8f020afd7f40b7e6e4c8bd2fc9eb7aec9836837b667685b3a.
//
// Solidity: event DeployedRaw(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) WatchDeployedRaw(opts *bind.WatchOpts, sink chan<- *AliceNetFactoryDeployedRaw) (event.Subscription, error) {

	logs, sub, err := _AliceNetFactory.contract.WatchLogs(opts, "DeployedRaw")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AliceNetFactoryDeployedRaw)
				if err := _AliceNetFactory.contract.UnpackLog(event, "DeployedRaw", log); err != nil {
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

// ParseDeployedRaw is a log parse operation binding the contract event 0xd3acf0da590cfcd8f020afd7f40b7e6e4c8bd2fc9eb7aec9836837b667685b3a.
//
// Solidity: event DeployedRaw(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) ParseDeployedRaw(log types.Log) (*AliceNetFactoryDeployedRaw, error) {
	event := new(AliceNetFactoryDeployedRaw)
	if err := _AliceNetFactory.contract.UnpackLog(event, "DeployedRaw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AliceNetFactoryDeployedStaticIterator is returned from FilterDeployedStatic and is used to iterate over the raw logs and unpacked data for DeployedStatic events raised by the AliceNetFactory contract.
type AliceNetFactoryDeployedStaticIterator struct {
	Event *AliceNetFactoryDeployedStatic // Event containing the contract specifics and raw log

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
func (it *AliceNetFactoryDeployedStaticIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AliceNetFactoryDeployedStatic)
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
		it.Event = new(AliceNetFactoryDeployedStatic)
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
func (it *AliceNetFactoryDeployedStaticIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AliceNetFactoryDeployedStaticIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AliceNetFactoryDeployedStatic represents a DeployedStatic event raised by the AliceNetFactory contract.
type AliceNetFactoryDeployedStatic struct {
	ContractAddr common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterDeployedStatic is a free log retrieval operation binding the contract event 0xe8b9cb7d60827a7d55e211f1382dd0f129adb541af9fe45a09ab4a18b76e7c65.
//
// Solidity: event DeployedStatic(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) FilterDeployedStatic(opts *bind.FilterOpts) (*AliceNetFactoryDeployedStaticIterator, error) {

	logs, sub, err := _AliceNetFactory.contract.FilterLogs(opts, "DeployedStatic")
	if err != nil {
		return nil, err
	}
	return &AliceNetFactoryDeployedStaticIterator{contract: _AliceNetFactory.contract, event: "DeployedStatic", logs: logs, sub: sub}, nil
}

// WatchDeployedStatic is a free log subscription operation binding the contract event 0xe8b9cb7d60827a7d55e211f1382dd0f129adb541af9fe45a09ab4a18b76e7c65.
//
// Solidity: event DeployedStatic(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) WatchDeployedStatic(opts *bind.WatchOpts, sink chan<- *AliceNetFactoryDeployedStatic) (event.Subscription, error) {

	logs, sub, err := _AliceNetFactory.contract.WatchLogs(opts, "DeployedStatic")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AliceNetFactoryDeployedStatic)
				if err := _AliceNetFactory.contract.UnpackLog(event, "DeployedStatic", log); err != nil {
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

// ParseDeployedStatic is a log parse operation binding the contract event 0xe8b9cb7d60827a7d55e211f1382dd0f129adb541af9fe45a09ab4a18b76e7c65.
//
// Solidity: event DeployedStatic(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) ParseDeployedStatic(log types.Log) (*AliceNetFactoryDeployedStatic, error) {
	event := new(AliceNetFactoryDeployedStatic)
	if err := _AliceNetFactory.contract.UnpackLog(event, "DeployedStatic", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AliceNetFactoryDeployedTemplateIterator is returned from FilterDeployedTemplate and is used to iterate over the raw logs and unpacked data for DeployedTemplate events raised by the AliceNetFactory contract.
type AliceNetFactoryDeployedTemplateIterator struct {
	Event *AliceNetFactoryDeployedTemplate // Event containing the contract specifics and raw log

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
func (it *AliceNetFactoryDeployedTemplateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AliceNetFactoryDeployedTemplate)
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
		it.Event = new(AliceNetFactoryDeployedTemplate)
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
func (it *AliceNetFactoryDeployedTemplateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AliceNetFactoryDeployedTemplateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AliceNetFactoryDeployedTemplate represents a DeployedTemplate event raised by the AliceNetFactory contract.
type AliceNetFactoryDeployedTemplate struct {
	ContractAddr common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterDeployedTemplate is a free log retrieval operation binding the contract event 0x6cd94ea1c5d9f99038bb4629d8a759399654d3861b73bf3a2b0cf484dae72138.
//
// Solidity: event DeployedTemplate(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) FilterDeployedTemplate(opts *bind.FilterOpts) (*AliceNetFactoryDeployedTemplateIterator, error) {

	logs, sub, err := _AliceNetFactory.contract.FilterLogs(opts, "DeployedTemplate")
	if err != nil {
		return nil, err
	}
	return &AliceNetFactoryDeployedTemplateIterator{contract: _AliceNetFactory.contract, event: "DeployedTemplate", logs: logs, sub: sub}, nil
}

// WatchDeployedTemplate is a free log subscription operation binding the contract event 0x6cd94ea1c5d9f99038bb4629d8a759399654d3861b73bf3a2b0cf484dae72138.
//
// Solidity: event DeployedTemplate(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) WatchDeployedTemplate(opts *bind.WatchOpts, sink chan<- *AliceNetFactoryDeployedTemplate) (event.Subscription, error) {

	logs, sub, err := _AliceNetFactory.contract.WatchLogs(opts, "DeployedTemplate")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AliceNetFactoryDeployedTemplate)
				if err := _AliceNetFactory.contract.UnpackLog(event, "DeployedTemplate", log); err != nil {
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

// ParseDeployedTemplate is a log parse operation binding the contract event 0x6cd94ea1c5d9f99038bb4629d8a759399654d3861b73bf3a2b0cf484dae72138.
//
// Solidity: event DeployedTemplate(address contractAddr)
func (_AliceNetFactory *AliceNetFactoryFilterer) ParseDeployedTemplate(log types.Log) (*AliceNetFactoryDeployedTemplate, error) {
	event := new(AliceNetFactoryDeployedTemplate)
	if err := _AliceNetFactory.contract.UnpackLog(event, "DeployedTemplate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
