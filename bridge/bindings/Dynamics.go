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

// CanonicalVersion is an auto generated low-level Go binding around an user-defined struct.
type CanonicalVersion struct {
	Major          uint32
	Minor          uint32
	Patch          uint32
	ExecutionEpoch uint32
	BinaryHash     [32]byte
}

// Configuration is an auto generated low-level Go binding around an user-defined struct.
type Configuration struct {
	MinEpochsBetweenUpdates *big.Int
	MaxEpochsBetweenUpdates *big.Int
}

// DynamicValues is an auto generated low-level Go binding around an user-defined struct.
type DynamicValues struct {
	EncoderVersion          uint8
	ProposalTimeout         *big.Int
	PreVoteTimeout          uint32
	PreCommitTimeout        uint32
	MaxBlockSize            uint32
	DataStoreFee            uint64
	ValueStoreFee           uint64
	MinScaledTransactionFee *big.Int
}

// DynamicsMetaData contains all meta data concerning the Dynamics contract.
var DynamicsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"}],\"name\":\"DynamicValueNotFound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"ExistentNodeAtPosition\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"InexistentNodeAtPosition\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sentHash\",\"type\":\"bytes32\"}],\"name\":\"InvalidAliceNetNodeHash\",\"type\":\"error\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"major\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"minor\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"patch\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"executionEpoch\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"binaryHash\",\"type\":\"bytes32\"}],\"internalType\":\"structCanonicalVersion\",\"name\":\"newVersion\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint32\",\"name\":\"major\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"minor\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"patch\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"executionEpoch\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"binaryHash\",\"type\":\"bytes32\"}],\"internalType\":\"structCanonicalVersion\",\"name\":\"current\",\"type\":\"tuple\"}],\"name\":\"InvalidAliceNetNodeVersion\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidData\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"codeSize\",\"type\":\"uint256\"}],\"name\":\"InvalidExtCodeSize\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"head\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"tail\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"InvalidNodeId\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"scheduledDate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minScheduledDate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxScheduledDate\",\"type\":\"uint256\"}],\"name\":\"InvalidScheduledDate\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyFactory\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlySnapshots\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"name\":\"DeployedStorageContract\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"rawDynamicValues\",\"type\":\"bytes\"}],\"name\":\"DynamicValueChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"major\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"minor\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"patch\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"executionEpoch\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"binaryHash\",\"type\":\"bytes32\"}],\"indexed\":false,\"internalType\":\"structCanonicalVersion\",\"name\":\"version\",\"type\":\"tuple\"}],\"name\":\"NewAliceNetNodeVersionAvailable\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"relativeExecutionEpoch\",\"type\":\"uint32\"},{\"components\":[{\"internalType\":\"enumVersion\",\"name\":\"encoderVersion\",\"type\":\"uint8\"},{\"internalType\":\"uint24\",\"name\":\"proposalTimeout\",\"type\":\"uint24\"},{\"internalType\":\"uint32\",\"name\":\"preVoteTimeout\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"preCommitTimeout\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"maxBlockSize\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"dataStoreFee\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"valueStoreFee\",\"type\":\"uint64\"},{\"internalType\":\"uint128\",\"name\":\"minScaledTransactionFee\",\"type\":\"uint128\"}],\"internalType\":\"structDynamicValues\",\"name\":\"newValue\",\"type\":\"tuple\"}],\"name\":\"changeDynamicValues\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"decodeDynamicValues\",\"outputs\":[{\"components\":[{\"internalType\":\"enumVersion\",\"name\":\"encoderVersion\",\"type\":\"uint8\"},{\"internalType\":\"uint24\",\"name\":\"proposalTimeout\",\"type\":\"uint24\"},{\"internalType\":\"uint32\",\"name\":\"preVoteTimeout\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"preCommitTimeout\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"maxBlockSize\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"dataStoreFee\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"valueStoreFee\",\"type\":\"uint64\"},{\"internalType\":\"uint128\",\"name\":\"minScaledTransactionFee\",\"type\":\"uint128\"}],\"internalType\":\"structDynamicValues\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"deployStorage\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"contractAddr\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"enumVersion\",\"name\":\"encoderVersion\",\"type\":\"uint8\"},{\"internalType\":\"uint24\",\"name\":\"proposalTimeout\",\"type\":\"uint24\"},{\"internalType\":\"uint32\",\"name\":\"preVoteTimeout\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"preCommitTimeout\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"maxBlockSize\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"dataStoreFee\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"valueStoreFee\",\"type\":\"uint64\"},{\"internalType\":\"uint128\",\"name\":\"minScaledTransactionFee\",\"type\":\"uint128\"}],\"internalType\":\"structDynamicValues\",\"name\":\"value\",\"type\":\"tuple\"}],\"name\":\"encodeDynamicValues\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getConfiguration\",\"outputs\":[{\"components\":[{\"internalType\":\"uint128\",\"name\":\"minEpochsBetweenUpdates\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"maxEpochsBetweenUpdates\",\"type\":\"uint128\"}],\"internalType\":\"structConfiguration\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEncodingVersion\",\"outputs\":[{\"internalType\":\"enumVersion\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLatestAliceNetVersion\",\"outputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"major\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"minor\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"patch\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"executionEpoch\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"binaryHash\",\"type\":\"bytes32\"}],\"internalType\":\"structCanonicalVersion\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLatestDynamicValues\",\"outputs\":[{\"components\":[{\"internalType\":\"enumVersion\",\"name\":\"encoderVersion\",\"type\":\"uint8\"},{\"internalType\":\"uint24\",\"name\":\"proposalTimeout\",\"type\":\"uint24\"},{\"internalType\":\"uint32\",\"name\":\"preVoteTimeout\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"preCommitTimeout\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"maxBlockSize\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"dataStoreFee\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"valueStoreFee\",\"type\":\"uint64\"},{\"internalType\":\"uint128\",\"name\":\"minScaledTransactionFee\",\"type\":\"uint128\"}],\"internalType\":\"structDynamicValues\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"}],\"name\":\"getPreviousDynamicValues\",\"outputs\":[{\"components\":[{\"internalType\":\"enumVersion\",\"name\":\"encoderVersion\",\"type\":\"uint8\"},{\"internalType\":\"uint24\",\"name\":\"proposalTimeout\",\"type\":\"uint24\"},{\"internalType\":\"uint32\",\"name\":\"preVoteTimeout\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"preCommitTimeout\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"maxBlockSize\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"dataStoreFee\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"valueStoreFee\",\"type\":\"uint64\"},{\"internalType\":\"uint128\",\"name\":\"minScaledTransactionFee\",\"type\":\"uint128\"}],\"internalType\":\"structDynamicValues\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint128\",\"name\":\"minEpochsBetweenUpdates\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"maxEpochsBetweenUpdates\",\"type\":\"uint128\"}],\"internalType\":\"structConfiguration\",\"name\":\"newConfig\",\"type\":\"tuple\"}],\"name\":\"setConfiguration\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"relativeUpdateEpoch\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"majorVersion\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"minorVersion\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"patch\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"binaryHash\",\"type\":\"bytes32\"}],\"name\":\"updateAliceNetNodeVersion\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"currentEpoch\",\"type\":\"uint32\"}],\"name\":\"updateHead\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// DynamicsABI is the input ABI used to generate the binding from.
// Deprecated: Use DynamicsMetaData.ABI instead.
var DynamicsABI = DynamicsMetaData.ABI

// Dynamics is an auto generated Go binding around an Ethereum contract.
type Dynamics struct {
	DynamicsCaller     // Read-only binding to the contract
	DynamicsTransactor // Write-only binding to the contract
	DynamicsFilterer   // Log filterer for contract events
}

// DynamicsCaller is an auto generated read-only Go binding around an Ethereum contract.
type DynamicsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DynamicsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DynamicsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DynamicsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DynamicsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DynamicsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DynamicsSession struct {
	Contract     *Dynamics         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DynamicsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DynamicsCallerSession struct {
	Contract *DynamicsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// DynamicsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DynamicsTransactorSession struct {
	Contract     *DynamicsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// DynamicsRaw is an auto generated low-level Go binding around an Ethereum contract.
type DynamicsRaw struct {
	Contract *Dynamics // Generic contract binding to access the raw methods on
}

// DynamicsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DynamicsCallerRaw struct {
	Contract *DynamicsCaller // Generic read-only contract binding to access the raw methods on
}

// DynamicsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DynamicsTransactorRaw struct {
	Contract *DynamicsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDynamics creates a new instance of Dynamics, bound to a specific deployed contract.
func NewDynamics(address common.Address, backend bind.ContractBackend) (*Dynamics, error) {
	contract, err := bindDynamics(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Dynamics{DynamicsCaller: DynamicsCaller{contract: contract}, DynamicsTransactor: DynamicsTransactor{contract: contract}, DynamicsFilterer: DynamicsFilterer{contract: contract}}, nil
}

// NewDynamicsCaller creates a new read-only instance of Dynamics, bound to a specific deployed contract.
func NewDynamicsCaller(address common.Address, caller bind.ContractCaller) (*DynamicsCaller, error) {
	contract, err := bindDynamics(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DynamicsCaller{contract: contract}, nil
}

// NewDynamicsTransactor creates a new write-only instance of Dynamics, bound to a specific deployed contract.
func NewDynamicsTransactor(address common.Address, transactor bind.ContractTransactor) (*DynamicsTransactor, error) {
	contract, err := bindDynamics(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DynamicsTransactor{contract: contract}, nil
}

// NewDynamicsFilterer creates a new log filterer instance of Dynamics, bound to a specific deployed contract.
func NewDynamicsFilterer(address common.Address, filterer bind.ContractFilterer) (*DynamicsFilterer, error) {
	contract, err := bindDynamics(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DynamicsFilterer{contract: contract}, nil
}

// bindDynamics binds a generic wrapper to an already deployed contract.
func bindDynamics(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(DynamicsABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Dynamics *DynamicsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Dynamics.Contract.DynamicsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Dynamics *DynamicsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Dynamics.Contract.DynamicsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Dynamics *DynamicsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Dynamics.Contract.DynamicsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Dynamics *DynamicsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Dynamics.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Dynamics *DynamicsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Dynamics.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Dynamics *DynamicsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Dynamics.Contract.contract.Transact(opts, method, params...)
}

// DecodeDynamicValues is a free data retrieval call binding the contract method 0x5234bca0.
//
// Solidity: function decodeDynamicValues(address addr) view returns((uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128))
func (_Dynamics *DynamicsCaller) DecodeDynamicValues(opts *bind.CallOpts, addr common.Address) (DynamicValues, error) {
	var out []interface{}
	err := _Dynamics.contract.Call(opts, &out, "decodeDynamicValues", addr)

	if err != nil {
		return *new(DynamicValues), err
	}

	out0 := *abi.ConvertType(out[0], new(DynamicValues)).(*DynamicValues)

	return out0, err

}

// DecodeDynamicValues is a free data retrieval call binding the contract method 0x5234bca0.
//
// Solidity: function decodeDynamicValues(address addr) view returns((uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128))
func (_Dynamics *DynamicsSession) DecodeDynamicValues(addr common.Address) (DynamicValues, error) {
	return _Dynamics.Contract.DecodeDynamicValues(&_Dynamics.CallOpts, addr)
}

// DecodeDynamicValues is a free data retrieval call binding the contract method 0x5234bca0.
//
// Solidity: function decodeDynamicValues(address addr) view returns((uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128))
func (_Dynamics *DynamicsCallerSession) DecodeDynamicValues(addr common.Address) (DynamicValues, error) {
	return _Dynamics.Contract.DecodeDynamicValues(&_Dynamics.CallOpts, addr)
}

// EncodeDynamicValues is a free data retrieval call binding the contract method 0x972e374c.
//
// Solidity: function encodeDynamicValues((uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128) value) pure returns(bytes)
func (_Dynamics *DynamicsCaller) EncodeDynamicValues(opts *bind.CallOpts, value DynamicValues) ([]byte, error) {
	var out []interface{}
	err := _Dynamics.contract.Call(opts, &out, "encodeDynamicValues", value)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// EncodeDynamicValues is a free data retrieval call binding the contract method 0x972e374c.
//
// Solidity: function encodeDynamicValues((uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128) value) pure returns(bytes)
func (_Dynamics *DynamicsSession) EncodeDynamicValues(value DynamicValues) ([]byte, error) {
	return _Dynamics.Contract.EncodeDynamicValues(&_Dynamics.CallOpts, value)
}

// EncodeDynamicValues is a free data retrieval call binding the contract method 0x972e374c.
//
// Solidity: function encodeDynamicValues((uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128) value) pure returns(bytes)
func (_Dynamics *DynamicsCallerSession) EncodeDynamicValues(value DynamicValues) ([]byte, error) {
	return _Dynamics.Contract.EncodeDynamicValues(&_Dynamics.CallOpts, value)
}

// GetConfiguration is a free data retrieval call binding the contract method 0x6bd50cef.
//
// Solidity: function getConfiguration() view returns((uint128,uint128))
func (_Dynamics *DynamicsCaller) GetConfiguration(opts *bind.CallOpts) (Configuration, error) {
	var out []interface{}
	err := _Dynamics.contract.Call(opts, &out, "getConfiguration")

	if err != nil {
		return *new(Configuration), err
	}

	out0 := *abi.ConvertType(out[0], new(Configuration)).(*Configuration)

	return out0, err

}

// GetConfiguration is a free data retrieval call binding the contract method 0x6bd50cef.
//
// Solidity: function getConfiguration() view returns((uint128,uint128))
func (_Dynamics *DynamicsSession) GetConfiguration() (Configuration, error) {
	return _Dynamics.Contract.GetConfiguration(&_Dynamics.CallOpts)
}

// GetConfiguration is a free data retrieval call binding the contract method 0x6bd50cef.
//
// Solidity: function getConfiguration() view returns((uint128,uint128))
func (_Dynamics *DynamicsCallerSession) GetConfiguration() (Configuration, error) {
	return _Dynamics.Contract.GetConfiguration(&_Dynamics.CallOpts)
}

// GetEncodingVersion is a free data retrieval call binding the contract method 0x49f13f68.
//
// Solidity: function getEncodingVersion() pure returns(uint8)
func (_Dynamics *DynamicsCaller) GetEncodingVersion(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Dynamics.contract.Call(opts, &out, "getEncodingVersion")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetEncodingVersion is a free data retrieval call binding the contract method 0x49f13f68.
//
// Solidity: function getEncodingVersion() pure returns(uint8)
func (_Dynamics *DynamicsSession) GetEncodingVersion() (uint8, error) {
	return _Dynamics.Contract.GetEncodingVersion(&_Dynamics.CallOpts)
}

// GetEncodingVersion is a free data retrieval call binding the contract method 0x49f13f68.
//
// Solidity: function getEncodingVersion() pure returns(uint8)
func (_Dynamics *DynamicsCallerSession) GetEncodingVersion() (uint8, error) {
	return _Dynamics.Contract.GetEncodingVersion(&_Dynamics.CallOpts)
}

// GetLatestAliceNetVersion is a free data retrieval call binding the contract method 0xd85200cc.
//
// Solidity: function getLatestAliceNetVersion() view returns((uint32,uint32,uint32,uint32,bytes32))
func (_Dynamics *DynamicsCaller) GetLatestAliceNetVersion(opts *bind.CallOpts) (CanonicalVersion, error) {
	var out []interface{}
	err := _Dynamics.contract.Call(opts, &out, "getLatestAliceNetVersion")

	if err != nil {
		return *new(CanonicalVersion), err
	}

	out0 := *abi.ConvertType(out[0], new(CanonicalVersion)).(*CanonicalVersion)

	return out0, err

}

// GetLatestAliceNetVersion is a free data retrieval call binding the contract method 0xd85200cc.
//
// Solidity: function getLatestAliceNetVersion() view returns((uint32,uint32,uint32,uint32,bytes32))
func (_Dynamics *DynamicsSession) GetLatestAliceNetVersion() (CanonicalVersion, error) {
	return _Dynamics.Contract.GetLatestAliceNetVersion(&_Dynamics.CallOpts)
}

// GetLatestAliceNetVersion is a free data retrieval call binding the contract method 0xd85200cc.
//
// Solidity: function getLatestAliceNetVersion() view returns((uint32,uint32,uint32,uint32,bytes32))
func (_Dynamics *DynamicsCallerSession) GetLatestAliceNetVersion() (CanonicalVersion, error) {
	return _Dynamics.Contract.GetLatestAliceNetVersion(&_Dynamics.CallOpts)
}

// GetLatestDynamicValues is a free data retrieval call binding the contract method 0xf53efe02.
//
// Solidity: function getLatestDynamicValues() view returns((uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128))
func (_Dynamics *DynamicsCaller) GetLatestDynamicValues(opts *bind.CallOpts) (DynamicValues, error) {
	var out []interface{}
	err := _Dynamics.contract.Call(opts, &out, "getLatestDynamicValues")

	if err != nil {
		return *new(DynamicValues), err
	}

	out0 := *abi.ConvertType(out[0], new(DynamicValues)).(*DynamicValues)

	return out0, err

}

// GetLatestDynamicValues is a free data retrieval call binding the contract method 0xf53efe02.
//
// Solidity: function getLatestDynamicValues() view returns((uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128))
func (_Dynamics *DynamicsSession) GetLatestDynamicValues() (DynamicValues, error) {
	return _Dynamics.Contract.GetLatestDynamicValues(&_Dynamics.CallOpts)
}

// GetLatestDynamicValues is a free data retrieval call binding the contract method 0xf53efe02.
//
// Solidity: function getLatestDynamicValues() view returns((uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128))
func (_Dynamics *DynamicsCallerSession) GetLatestDynamicValues() (DynamicValues, error) {
	return _Dynamics.Contract.GetLatestDynamicValues(&_Dynamics.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_Dynamics *DynamicsCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _Dynamics.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_Dynamics *DynamicsSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _Dynamics.Contract.GetMetamorphicContractAddress(&_Dynamics.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_Dynamics *DynamicsCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _Dynamics.Contract.GetMetamorphicContractAddress(&_Dynamics.CallOpts, _salt, _factory)
}

// GetPreviousDynamicValues is a free data retrieval call binding the contract method 0x72015859.
//
// Solidity: function getPreviousDynamicValues(uint256 epoch) view returns((uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128))
func (_Dynamics *DynamicsCaller) GetPreviousDynamicValues(opts *bind.CallOpts, epoch *big.Int) (DynamicValues, error) {
	var out []interface{}
	err := _Dynamics.contract.Call(opts, &out, "getPreviousDynamicValues", epoch)

	if err != nil {
		return *new(DynamicValues), err
	}

	out0 := *abi.ConvertType(out[0], new(DynamicValues)).(*DynamicValues)

	return out0, err

}

// GetPreviousDynamicValues is a free data retrieval call binding the contract method 0x72015859.
//
// Solidity: function getPreviousDynamicValues(uint256 epoch) view returns((uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128))
func (_Dynamics *DynamicsSession) GetPreviousDynamicValues(epoch *big.Int) (DynamicValues, error) {
	return _Dynamics.Contract.GetPreviousDynamicValues(&_Dynamics.CallOpts, epoch)
}

// GetPreviousDynamicValues is a free data retrieval call binding the contract method 0x72015859.
//
// Solidity: function getPreviousDynamicValues(uint256 epoch) view returns((uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128))
func (_Dynamics *DynamicsCallerSession) GetPreviousDynamicValues(epoch *big.Int) (DynamicValues, error) {
	return _Dynamics.Contract.GetPreviousDynamicValues(&_Dynamics.CallOpts, epoch)
}

// ChangeDynamicValues is a paid mutator transaction binding the contract method 0xe84f2fd0.
//
// Solidity: function changeDynamicValues(uint32 relativeExecutionEpoch, (uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128) newValue) returns()
func (_Dynamics *DynamicsTransactor) ChangeDynamicValues(opts *bind.TransactOpts, relativeExecutionEpoch uint32, newValue DynamicValues) (*types.Transaction, error) {
	return _Dynamics.contract.Transact(opts, "changeDynamicValues", relativeExecutionEpoch, newValue)
}

// ChangeDynamicValues is a paid mutator transaction binding the contract method 0xe84f2fd0.
//
// Solidity: function changeDynamicValues(uint32 relativeExecutionEpoch, (uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128) newValue) returns()
func (_Dynamics *DynamicsSession) ChangeDynamicValues(relativeExecutionEpoch uint32, newValue DynamicValues) (*types.Transaction, error) {
	return _Dynamics.Contract.ChangeDynamicValues(&_Dynamics.TransactOpts, relativeExecutionEpoch, newValue)
}

// ChangeDynamicValues is a paid mutator transaction binding the contract method 0xe84f2fd0.
//
// Solidity: function changeDynamicValues(uint32 relativeExecutionEpoch, (uint8,uint24,uint32,uint32,uint32,uint64,uint64,uint128) newValue) returns()
func (_Dynamics *DynamicsTransactorSession) ChangeDynamicValues(relativeExecutionEpoch uint32, newValue DynamicValues) (*types.Transaction, error) {
	return _Dynamics.Contract.ChangeDynamicValues(&_Dynamics.TransactOpts, relativeExecutionEpoch, newValue)
}

// DeployStorage is a paid mutator transaction binding the contract method 0xa130fd2e.
//
// Solidity: function deployStorage(bytes data) returns(address contractAddr)
func (_Dynamics *DynamicsTransactor) DeployStorage(opts *bind.TransactOpts, data []byte) (*types.Transaction, error) {
	return _Dynamics.contract.Transact(opts, "deployStorage", data)
}

// DeployStorage is a paid mutator transaction binding the contract method 0xa130fd2e.
//
// Solidity: function deployStorage(bytes data) returns(address contractAddr)
func (_Dynamics *DynamicsSession) DeployStorage(data []byte) (*types.Transaction, error) {
	return _Dynamics.Contract.DeployStorage(&_Dynamics.TransactOpts, data)
}

// DeployStorage is a paid mutator transaction binding the contract method 0xa130fd2e.
//
// Solidity: function deployStorage(bytes data) returns(address contractAddr)
func (_Dynamics *DynamicsTransactorSession) DeployStorage(data []byte) (*types.Transaction, error) {
	return _Dynamics.Contract.DeployStorage(&_Dynamics.TransactOpts, data)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_Dynamics *DynamicsTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Dynamics.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_Dynamics *DynamicsSession) Initialize() (*types.Transaction, error) {
	return _Dynamics.Contract.Initialize(&_Dynamics.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_Dynamics *DynamicsTransactorSession) Initialize() (*types.Transaction, error) {
	return _Dynamics.Contract.Initialize(&_Dynamics.TransactOpts)
}

// SetConfiguration is a paid mutator transaction binding the contract method 0x402292df.
//
// Solidity: function setConfiguration((uint128,uint128) newConfig) returns()
func (_Dynamics *DynamicsTransactor) SetConfiguration(opts *bind.TransactOpts, newConfig Configuration) (*types.Transaction, error) {
	return _Dynamics.contract.Transact(opts, "setConfiguration", newConfig)
}

// SetConfiguration is a paid mutator transaction binding the contract method 0x402292df.
//
// Solidity: function setConfiguration((uint128,uint128) newConfig) returns()
func (_Dynamics *DynamicsSession) SetConfiguration(newConfig Configuration) (*types.Transaction, error) {
	return _Dynamics.Contract.SetConfiguration(&_Dynamics.TransactOpts, newConfig)
}

// SetConfiguration is a paid mutator transaction binding the contract method 0x402292df.
//
// Solidity: function setConfiguration((uint128,uint128) newConfig) returns()
func (_Dynamics *DynamicsTransactorSession) SetConfiguration(newConfig Configuration) (*types.Transaction, error) {
	return _Dynamics.Contract.SetConfiguration(&_Dynamics.TransactOpts, newConfig)
}

// UpdateAliceNetNodeVersion is a paid mutator transaction binding the contract method 0xab8a8411.
//
// Solidity: function updateAliceNetNodeVersion(uint32 relativeUpdateEpoch, uint32 majorVersion, uint32 minorVersion, uint32 patch, bytes32 binaryHash) returns()
func (_Dynamics *DynamicsTransactor) UpdateAliceNetNodeVersion(opts *bind.TransactOpts, relativeUpdateEpoch uint32, majorVersion uint32, minorVersion uint32, patch uint32, binaryHash [32]byte) (*types.Transaction, error) {
	return _Dynamics.contract.Transact(opts, "updateAliceNetNodeVersion", relativeUpdateEpoch, majorVersion, minorVersion, patch, binaryHash)
}

// UpdateAliceNetNodeVersion is a paid mutator transaction binding the contract method 0xab8a8411.
//
// Solidity: function updateAliceNetNodeVersion(uint32 relativeUpdateEpoch, uint32 majorVersion, uint32 minorVersion, uint32 patch, bytes32 binaryHash) returns()
func (_Dynamics *DynamicsSession) UpdateAliceNetNodeVersion(relativeUpdateEpoch uint32, majorVersion uint32, minorVersion uint32, patch uint32, binaryHash [32]byte) (*types.Transaction, error) {
	return _Dynamics.Contract.UpdateAliceNetNodeVersion(&_Dynamics.TransactOpts, relativeUpdateEpoch, majorVersion, minorVersion, patch, binaryHash)
}

// UpdateAliceNetNodeVersion is a paid mutator transaction binding the contract method 0xab8a8411.
//
// Solidity: function updateAliceNetNodeVersion(uint32 relativeUpdateEpoch, uint32 majorVersion, uint32 minorVersion, uint32 patch, bytes32 binaryHash) returns()
func (_Dynamics *DynamicsTransactorSession) UpdateAliceNetNodeVersion(relativeUpdateEpoch uint32, majorVersion uint32, minorVersion uint32, patch uint32, binaryHash [32]byte) (*types.Transaction, error) {
	return _Dynamics.Contract.UpdateAliceNetNodeVersion(&_Dynamics.TransactOpts, relativeUpdateEpoch, majorVersion, minorVersion, patch, binaryHash)
}

// UpdateHead is a paid mutator transaction binding the contract method 0x3b4ef26a.
//
// Solidity: function updateHead(uint32 currentEpoch) returns()
func (_Dynamics *DynamicsTransactor) UpdateHead(opts *bind.TransactOpts, currentEpoch uint32) (*types.Transaction, error) {
	return _Dynamics.contract.Transact(opts, "updateHead", currentEpoch)
}

// UpdateHead is a paid mutator transaction binding the contract method 0x3b4ef26a.
//
// Solidity: function updateHead(uint32 currentEpoch) returns()
func (_Dynamics *DynamicsSession) UpdateHead(currentEpoch uint32) (*types.Transaction, error) {
	return _Dynamics.Contract.UpdateHead(&_Dynamics.TransactOpts, currentEpoch)
}

// UpdateHead is a paid mutator transaction binding the contract method 0x3b4ef26a.
//
// Solidity: function updateHead(uint32 currentEpoch) returns()
func (_Dynamics *DynamicsTransactorSession) UpdateHead(currentEpoch uint32) (*types.Transaction, error) {
	return _Dynamics.Contract.UpdateHead(&_Dynamics.TransactOpts, currentEpoch)
}

// DynamicsDeployedStorageContractIterator is returned from FilterDeployedStorageContract and is used to iterate over the raw logs and unpacked data for DeployedStorageContract events raised by the Dynamics contract.
type DynamicsDeployedStorageContractIterator struct {
	Event *DynamicsDeployedStorageContract // Event containing the contract specifics and raw log

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
func (it *DynamicsDeployedStorageContractIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DynamicsDeployedStorageContract)
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
		it.Event = new(DynamicsDeployedStorageContract)
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
func (it *DynamicsDeployedStorageContractIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DynamicsDeployedStorageContractIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DynamicsDeployedStorageContract represents a DeployedStorageContract event raised by the Dynamics contract.
type DynamicsDeployedStorageContract struct {
	ContractAddr common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterDeployedStorageContract is a free log retrieval operation binding the contract event 0xc958f5befa7c4f4b38447cf2e058acacc3a6ea235df7bfea1e658fce46ed6e18.
//
// Solidity: event DeployedStorageContract(address contractAddr)
func (_Dynamics *DynamicsFilterer) FilterDeployedStorageContract(opts *bind.FilterOpts) (*DynamicsDeployedStorageContractIterator, error) {

	logs, sub, err := _Dynamics.contract.FilterLogs(opts, "DeployedStorageContract")
	if err != nil {
		return nil, err
	}
	return &DynamicsDeployedStorageContractIterator{contract: _Dynamics.contract, event: "DeployedStorageContract", logs: logs, sub: sub}, nil
}

// WatchDeployedStorageContract is a free log subscription operation binding the contract event 0xc958f5befa7c4f4b38447cf2e058acacc3a6ea235df7bfea1e658fce46ed6e18.
//
// Solidity: event DeployedStorageContract(address contractAddr)
func (_Dynamics *DynamicsFilterer) WatchDeployedStorageContract(opts *bind.WatchOpts, sink chan<- *DynamicsDeployedStorageContract) (event.Subscription, error) {

	logs, sub, err := _Dynamics.contract.WatchLogs(opts, "DeployedStorageContract")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DynamicsDeployedStorageContract)
				if err := _Dynamics.contract.UnpackLog(event, "DeployedStorageContract", log); err != nil {
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

// ParseDeployedStorageContract is a log parse operation binding the contract event 0xc958f5befa7c4f4b38447cf2e058acacc3a6ea235df7bfea1e658fce46ed6e18.
//
// Solidity: event DeployedStorageContract(address contractAddr)
func (_Dynamics *DynamicsFilterer) ParseDeployedStorageContract(log types.Log) (*DynamicsDeployedStorageContract, error) {
	event := new(DynamicsDeployedStorageContract)
	if err := _Dynamics.contract.UnpackLog(event, "DeployedStorageContract", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DynamicsDynamicValueChangedIterator is returned from FilterDynamicValueChanged and is used to iterate over the raw logs and unpacked data for DynamicValueChanged events raised by the Dynamics contract.
type DynamicsDynamicValueChangedIterator struct {
	Event *DynamicsDynamicValueChanged // Event containing the contract specifics and raw log

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
func (it *DynamicsDynamicValueChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DynamicsDynamicValueChanged)
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
		it.Event = new(DynamicsDynamicValueChanged)
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
func (it *DynamicsDynamicValueChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DynamicsDynamicValueChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DynamicsDynamicValueChanged represents a DynamicValueChanged event raised by the Dynamics contract.
type DynamicsDynamicValueChanged struct {
	Epoch            *big.Int
	RawDynamicValues []byte
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterDynamicValueChanged is a free log retrieval operation binding the contract event 0x75892f5f7bbecdec9b5e7d1b8dd98acb161abe69a3922b3abfbefaefb0490326.
//
// Solidity: event DynamicValueChanged(uint256 epoch, bytes rawDynamicValues)
func (_Dynamics *DynamicsFilterer) FilterDynamicValueChanged(opts *bind.FilterOpts) (*DynamicsDynamicValueChangedIterator, error) {

	logs, sub, err := _Dynamics.contract.FilterLogs(opts, "DynamicValueChanged")
	if err != nil {
		return nil, err
	}
	return &DynamicsDynamicValueChangedIterator{contract: _Dynamics.contract, event: "DynamicValueChanged", logs: logs, sub: sub}, nil
}

// WatchDynamicValueChanged is a free log subscription operation binding the contract event 0x75892f5f7bbecdec9b5e7d1b8dd98acb161abe69a3922b3abfbefaefb0490326.
//
// Solidity: event DynamicValueChanged(uint256 epoch, bytes rawDynamicValues)
func (_Dynamics *DynamicsFilterer) WatchDynamicValueChanged(opts *bind.WatchOpts, sink chan<- *DynamicsDynamicValueChanged) (event.Subscription, error) {

	logs, sub, err := _Dynamics.contract.WatchLogs(opts, "DynamicValueChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DynamicsDynamicValueChanged)
				if err := _Dynamics.contract.UnpackLog(event, "DynamicValueChanged", log); err != nil {
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

// ParseDynamicValueChanged is a log parse operation binding the contract event 0x75892f5f7bbecdec9b5e7d1b8dd98acb161abe69a3922b3abfbefaefb0490326.
//
// Solidity: event DynamicValueChanged(uint256 epoch, bytes rawDynamicValues)
func (_Dynamics *DynamicsFilterer) ParseDynamicValueChanged(log types.Log) (*DynamicsDynamicValueChanged, error) {
	event := new(DynamicsDynamicValueChanged)
	if err := _Dynamics.contract.UnpackLog(event, "DynamicValueChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DynamicsInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Dynamics contract.
type DynamicsInitializedIterator struct {
	Event *DynamicsInitialized // Event containing the contract specifics and raw log

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
func (it *DynamicsInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DynamicsInitialized)
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
		it.Event = new(DynamicsInitialized)
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
func (it *DynamicsInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DynamicsInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DynamicsInitialized represents a Initialized event raised by the Dynamics contract.
type DynamicsInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Dynamics *DynamicsFilterer) FilterInitialized(opts *bind.FilterOpts) (*DynamicsInitializedIterator, error) {

	logs, sub, err := _Dynamics.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &DynamicsInitializedIterator{contract: _Dynamics.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Dynamics *DynamicsFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *DynamicsInitialized) (event.Subscription, error) {

	logs, sub, err := _Dynamics.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DynamicsInitialized)
				if err := _Dynamics.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_Dynamics *DynamicsFilterer) ParseInitialized(log types.Log) (*DynamicsInitialized, error) {
	event := new(DynamicsInitialized)
	if err := _Dynamics.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DynamicsNewAliceNetNodeVersionAvailableIterator is returned from FilterNewAliceNetNodeVersionAvailable and is used to iterate over the raw logs and unpacked data for NewAliceNetNodeVersionAvailable events raised by the Dynamics contract.
type DynamicsNewAliceNetNodeVersionAvailableIterator struct {
	Event *DynamicsNewAliceNetNodeVersionAvailable // Event containing the contract specifics and raw log

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
func (it *DynamicsNewAliceNetNodeVersionAvailableIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DynamicsNewAliceNetNodeVersionAvailable)
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
		it.Event = new(DynamicsNewAliceNetNodeVersionAvailable)
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
func (it *DynamicsNewAliceNetNodeVersionAvailableIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DynamicsNewAliceNetNodeVersionAvailableIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DynamicsNewAliceNetNodeVersionAvailable represents a NewAliceNetNodeVersionAvailable event raised by the Dynamics contract.
type DynamicsNewAliceNetNodeVersionAvailable struct {
	Version CanonicalVersion
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNewAliceNetNodeVersionAvailable is a free log retrieval operation binding the contract event 0xd05fe7fc00936621efe67e776e805ddf5a4e50d53b63179d2c05523244ec95ff.
//
// Solidity: event NewAliceNetNodeVersionAvailable((uint32,uint32,uint32,uint32,bytes32) version)
func (_Dynamics *DynamicsFilterer) FilterNewAliceNetNodeVersionAvailable(opts *bind.FilterOpts) (*DynamicsNewAliceNetNodeVersionAvailableIterator, error) {

	logs, sub, err := _Dynamics.contract.FilterLogs(opts, "NewAliceNetNodeVersionAvailable")
	if err != nil {
		return nil, err
	}
	return &DynamicsNewAliceNetNodeVersionAvailableIterator{contract: _Dynamics.contract, event: "NewAliceNetNodeVersionAvailable", logs: logs, sub: sub}, nil
}

// WatchNewAliceNetNodeVersionAvailable is a free log subscription operation binding the contract event 0xd05fe7fc00936621efe67e776e805ddf5a4e50d53b63179d2c05523244ec95ff.
//
// Solidity: event NewAliceNetNodeVersionAvailable((uint32,uint32,uint32,uint32,bytes32) version)
func (_Dynamics *DynamicsFilterer) WatchNewAliceNetNodeVersionAvailable(opts *bind.WatchOpts, sink chan<- *DynamicsNewAliceNetNodeVersionAvailable) (event.Subscription, error) {

	logs, sub, err := _Dynamics.contract.WatchLogs(opts, "NewAliceNetNodeVersionAvailable")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DynamicsNewAliceNetNodeVersionAvailable)
				if err := _Dynamics.contract.UnpackLog(event, "NewAliceNetNodeVersionAvailable", log); err != nil {
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

// ParseNewAliceNetNodeVersionAvailable is a log parse operation binding the contract event 0xd05fe7fc00936621efe67e776e805ddf5a4e50d53b63179d2c05523244ec95ff.
//
// Solidity: event NewAliceNetNodeVersionAvailable((uint32,uint32,uint32,uint32,bytes32) version)
func (_Dynamics *DynamicsFilterer) ParseNewAliceNetNodeVersionAvailable(log types.Log) (*DynamicsNewAliceNetNodeVersionAvailable, error) {
	event := new(DynamicsNewAliceNetNodeVersionAvailable)
	if err := _Dynamics.contract.UnpackLog(event, "NewAliceNetNodeVersionAvailable", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
