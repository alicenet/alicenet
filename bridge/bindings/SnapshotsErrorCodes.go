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

// SnapshotsErrorCodesMetaData contains all meta data concerning the SnapshotsErrorCodes contract.
var SnapshotsErrorCodesMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"SNAPSHOT_CALLER_NOT_ETHDKG_PARTICIPANT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SNAPSHOT_CONSENSUS_RUNNING\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SNAPSHOT_INCORRECT_BLOCK_HEIGHT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SNAPSHOT_INCORRECT_CHAIN_ID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SNAPSHOT_MIGRATION_INPUT_DATA_MISMATCH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SNAPSHOT_MIGRATION_NOT_ALLOWED\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SNAPSHOT_MIN_BLOCKS_INTERVAL_NOT_PASSED\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SNAPSHOT_ONLY_VALIDATORS_ALLOWED\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SNAPSHOT_SIGNATURE_VERIFICATION_FAILED\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SNAPSHOT_WRONG_MASTER_PUBLIC_KEY\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// SnapshotsErrorCodesABI is the input ABI used to generate the binding from.
// Deprecated: Use SnapshotsErrorCodesMetaData.ABI instead.
var SnapshotsErrorCodesABI = SnapshotsErrorCodesMetaData.ABI

// SnapshotsErrorCodes is an auto generated Go binding around an Ethereum contract.
type SnapshotsErrorCodes struct {
	SnapshotsErrorCodesCaller     // Read-only binding to the contract
	SnapshotsErrorCodesTransactor // Write-only binding to the contract
	SnapshotsErrorCodesFilterer   // Log filterer for contract events
}

// SnapshotsErrorCodesCaller is an auto generated read-only Go binding around an Ethereum contract.
type SnapshotsErrorCodesCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SnapshotsErrorCodesTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SnapshotsErrorCodesTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SnapshotsErrorCodesFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SnapshotsErrorCodesFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SnapshotsErrorCodesSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SnapshotsErrorCodesSession struct {
	Contract     *SnapshotsErrorCodes // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// SnapshotsErrorCodesCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SnapshotsErrorCodesCallerSession struct {
	Contract *SnapshotsErrorCodesCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// SnapshotsErrorCodesTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SnapshotsErrorCodesTransactorSession struct {
	Contract     *SnapshotsErrorCodesTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// SnapshotsErrorCodesRaw is an auto generated low-level Go binding around an Ethereum contract.
type SnapshotsErrorCodesRaw struct {
	Contract *SnapshotsErrorCodes // Generic contract binding to access the raw methods on
}

// SnapshotsErrorCodesCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SnapshotsErrorCodesCallerRaw struct {
	Contract *SnapshotsErrorCodesCaller // Generic read-only contract binding to access the raw methods on
}

// SnapshotsErrorCodesTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SnapshotsErrorCodesTransactorRaw struct {
	Contract *SnapshotsErrorCodesTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSnapshotsErrorCodes creates a new instance of SnapshotsErrorCodes, bound to a specific deployed contract.
func NewSnapshotsErrorCodes(address common.Address, backend bind.ContractBackend) (*SnapshotsErrorCodes, error) {
	contract, err := bindSnapshotsErrorCodes(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SnapshotsErrorCodes{SnapshotsErrorCodesCaller: SnapshotsErrorCodesCaller{contract: contract}, SnapshotsErrorCodesTransactor: SnapshotsErrorCodesTransactor{contract: contract}, SnapshotsErrorCodesFilterer: SnapshotsErrorCodesFilterer{contract: contract}}, nil
}

// NewSnapshotsErrorCodesCaller creates a new read-only instance of SnapshotsErrorCodes, bound to a specific deployed contract.
func NewSnapshotsErrorCodesCaller(address common.Address, caller bind.ContractCaller) (*SnapshotsErrorCodesCaller, error) {
	contract, err := bindSnapshotsErrorCodes(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SnapshotsErrorCodesCaller{contract: contract}, nil
}

// NewSnapshotsErrorCodesTransactor creates a new write-only instance of SnapshotsErrorCodes, bound to a specific deployed contract.
func NewSnapshotsErrorCodesTransactor(address common.Address, transactor bind.ContractTransactor) (*SnapshotsErrorCodesTransactor, error) {
	contract, err := bindSnapshotsErrorCodes(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SnapshotsErrorCodesTransactor{contract: contract}, nil
}

// NewSnapshotsErrorCodesFilterer creates a new log filterer instance of SnapshotsErrorCodes, bound to a specific deployed contract.
func NewSnapshotsErrorCodesFilterer(address common.Address, filterer bind.ContractFilterer) (*SnapshotsErrorCodesFilterer, error) {
	contract, err := bindSnapshotsErrorCodes(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SnapshotsErrorCodesFilterer{contract: contract}, nil
}

// bindSnapshotsErrorCodes binds a generic wrapper to an already deployed contract.
func bindSnapshotsErrorCodes(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SnapshotsErrorCodesABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SnapshotsErrorCodes *SnapshotsErrorCodesRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SnapshotsErrorCodes.Contract.SnapshotsErrorCodesCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SnapshotsErrorCodes *SnapshotsErrorCodesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SnapshotsErrorCodes.Contract.SnapshotsErrorCodesTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SnapshotsErrorCodes *SnapshotsErrorCodesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SnapshotsErrorCodes.Contract.SnapshotsErrorCodesTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SnapshotsErrorCodes.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SnapshotsErrorCodes *SnapshotsErrorCodesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SnapshotsErrorCodes.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SnapshotsErrorCodes *SnapshotsErrorCodesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SnapshotsErrorCodes.Contract.contract.Transact(opts, method, params...)
}

// SNAPSHOTCALLERNOTETHDKGPARTICIPANT is a free data retrieval call binding the contract method 0xadcb3a44.
//
// Solidity: function SNAPSHOT_CALLER_NOT_ETHDKG_PARTICIPANT() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCaller) SNAPSHOTCALLERNOTETHDKGPARTICIPANT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SnapshotsErrorCodes.contract.Call(opts, &out, "SNAPSHOT_CALLER_NOT_ETHDKG_PARTICIPANT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SNAPSHOTCALLERNOTETHDKGPARTICIPANT is a free data retrieval call binding the contract method 0xadcb3a44.
//
// Solidity: function SNAPSHOT_CALLER_NOT_ETHDKG_PARTICIPANT() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesSession) SNAPSHOTCALLERNOTETHDKGPARTICIPANT() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTCALLERNOTETHDKGPARTICIPANT(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTCALLERNOTETHDKGPARTICIPANT is a free data retrieval call binding the contract method 0xadcb3a44.
//
// Solidity: function SNAPSHOT_CALLER_NOT_ETHDKG_PARTICIPANT() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCallerSession) SNAPSHOTCALLERNOTETHDKGPARTICIPANT() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTCALLERNOTETHDKGPARTICIPANT(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTCONSENSUSRUNNING is a free data retrieval call binding the contract method 0x8d17de64.
//
// Solidity: function SNAPSHOT_CONSENSUS_RUNNING() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCaller) SNAPSHOTCONSENSUSRUNNING(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SnapshotsErrorCodes.contract.Call(opts, &out, "SNAPSHOT_CONSENSUS_RUNNING")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SNAPSHOTCONSENSUSRUNNING is a free data retrieval call binding the contract method 0x8d17de64.
//
// Solidity: function SNAPSHOT_CONSENSUS_RUNNING() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesSession) SNAPSHOTCONSENSUSRUNNING() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTCONSENSUSRUNNING(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTCONSENSUSRUNNING is a free data retrieval call binding the contract method 0x8d17de64.
//
// Solidity: function SNAPSHOT_CONSENSUS_RUNNING() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCallerSession) SNAPSHOTCONSENSUSRUNNING() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTCONSENSUSRUNNING(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTINCORRECTBLOCKHEIGHT is a free data retrieval call binding the contract method 0x0c24555d.
//
// Solidity: function SNAPSHOT_INCORRECT_BLOCK_HEIGHT() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCaller) SNAPSHOTINCORRECTBLOCKHEIGHT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SnapshotsErrorCodes.contract.Call(opts, &out, "SNAPSHOT_INCORRECT_BLOCK_HEIGHT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SNAPSHOTINCORRECTBLOCKHEIGHT is a free data retrieval call binding the contract method 0x0c24555d.
//
// Solidity: function SNAPSHOT_INCORRECT_BLOCK_HEIGHT() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesSession) SNAPSHOTINCORRECTBLOCKHEIGHT() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTINCORRECTBLOCKHEIGHT(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTINCORRECTBLOCKHEIGHT is a free data retrieval call binding the contract method 0x0c24555d.
//
// Solidity: function SNAPSHOT_INCORRECT_BLOCK_HEIGHT() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCallerSession) SNAPSHOTINCORRECTBLOCKHEIGHT() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTINCORRECTBLOCKHEIGHT(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTINCORRECTCHAINID is a free data retrieval call binding the contract method 0x3f772c6f.
//
// Solidity: function SNAPSHOT_INCORRECT_CHAIN_ID() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCaller) SNAPSHOTINCORRECTCHAINID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SnapshotsErrorCodes.contract.Call(opts, &out, "SNAPSHOT_INCORRECT_CHAIN_ID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SNAPSHOTINCORRECTCHAINID is a free data retrieval call binding the contract method 0x3f772c6f.
//
// Solidity: function SNAPSHOT_INCORRECT_CHAIN_ID() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesSession) SNAPSHOTINCORRECTCHAINID() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTINCORRECTCHAINID(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTINCORRECTCHAINID is a free data retrieval call binding the contract method 0x3f772c6f.
//
// Solidity: function SNAPSHOT_INCORRECT_CHAIN_ID() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCallerSession) SNAPSHOTINCORRECTCHAINID() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTINCORRECTCHAINID(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTMIGRATIONINPUTDATAMISMATCH is a free data retrieval call binding the contract method 0x9854bdc5.
//
// Solidity: function SNAPSHOT_MIGRATION_INPUT_DATA_MISMATCH() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCaller) SNAPSHOTMIGRATIONINPUTDATAMISMATCH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SnapshotsErrorCodes.contract.Call(opts, &out, "SNAPSHOT_MIGRATION_INPUT_DATA_MISMATCH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SNAPSHOTMIGRATIONINPUTDATAMISMATCH is a free data retrieval call binding the contract method 0x9854bdc5.
//
// Solidity: function SNAPSHOT_MIGRATION_INPUT_DATA_MISMATCH() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesSession) SNAPSHOTMIGRATIONINPUTDATAMISMATCH() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTMIGRATIONINPUTDATAMISMATCH(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTMIGRATIONINPUTDATAMISMATCH is a free data retrieval call binding the contract method 0x9854bdc5.
//
// Solidity: function SNAPSHOT_MIGRATION_INPUT_DATA_MISMATCH() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCallerSession) SNAPSHOTMIGRATIONINPUTDATAMISMATCH() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTMIGRATIONINPUTDATAMISMATCH(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTMIGRATIONNOTALLOWED is a free data retrieval call binding the contract method 0x4f2adaee.
//
// Solidity: function SNAPSHOT_MIGRATION_NOT_ALLOWED() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCaller) SNAPSHOTMIGRATIONNOTALLOWED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SnapshotsErrorCodes.contract.Call(opts, &out, "SNAPSHOT_MIGRATION_NOT_ALLOWED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SNAPSHOTMIGRATIONNOTALLOWED is a free data retrieval call binding the contract method 0x4f2adaee.
//
// Solidity: function SNAPSHOT_MIGRATION_NOT_ALLOWED() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesSession) SNAPSHOTMIGRATIONNOTALLOWED() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTMIGRATIONNOTALLOWED(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTMIGRATIONNOTALLOWED is a free data retrieval call binding the contract method 0x4f2adaee.
//
// Solidity: function SNAPSHOT_MIGRATION_NOT_ALLOWED() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCallerSession) SNAPSHOTMIGRATIONNOTALLOWED() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTMIGRATIONNOTALLOWED(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTMINBLOCKSINTERVALNOTPASSED is a free data retrieval call binding the contract method 0x4a1ec2ee.
//
// Solidity: function SNAPSHOT_MIN_BLOCKS_INTERVAL_NOT_PASSED() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCaller) SNAPSHOTMINBLOCKSINTERVALNOTPASSED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SnapshotsErrorCodes.contract.Call(opts, &out, "SNAPSHOT_MIN_BLOCKS_INTERVAL_NOT_PASSED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SNAPSHOTMINBLOCKSINTERVALNOTPASSED is a free data retrieval call binding the contract method 0x4a1ec2ee.
//
// Solidity: function SNAPSHOT_MIN_BLOCKS_INTERVAL_NOT_PASSED() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesSession) SNAPSHOTMINBLOCKSINTERVALNOTPASSED() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTMINBLOCKSINTERVALNOTPASSED(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTMINBLOCKSINTERVALNOTPASSED is a free data retrieval call binding the contract method 0x4a1ec2ee.
//
// Solidity: function SNAPSHOT_MIN_BLOCKS_INTERVAL_NOT_PASSED() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCallerSession) SNAPSHOTMINBLOCKSINTERVALNOTPASSED() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTMINBLOCKSINTERVALNOTPASSED(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTONLYVALIDATORSALLOWED is a free data retrieval call binding the contract method 0x83d995fa.
//
// Solidity: function SNAPSHOT_ONLY_VALIDATORS_ALLOWED() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCaller) SNAPSHOTONLYVALIDATORSALLOWED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SnapshotsErrorCodes.contract.Call(opts, &out, "SNAPSHOT_ONLY_VALIDATORS_ALLOWED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SNAPSHOTONLYVALIDATORSALLOWED is a free data retrieval call binding the contract method 0x83d995fa.
//
// Solidity: function SNAPSHOT_ONLY_VALIDATORS_ALLOWED() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesSession) SNAPSHOTONLYVALIDATORSALLOWED() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTONLYVALIDATORSALLOWED(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTONLYVALIDATORSALLOWED is a free data retrieval call binding the contract method 0x83d995fa.
//
// Solidity: function SNAPSHOT_ONLY_VALIDATORS_ALLOWED() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCallerSession) SNAPSHOTONLYVALIDATORSALLOWED() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTONLYVALIDATORSALLOWED(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTSIGNATUREVERIFICATIONFAILED is a free data retrieval call binding the contract method 0x1d449ed1.
//
// Solidity: function SNAPSHOT_SIGNATURE_VERIFICATION_FAILED() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCaller) SNAPSHOTSIGNATUREVERIFICATIONFAILED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SnapshotsErrorCodes.contract.Call(opts, &out, "SNAPSHOT_SIGNATURE_VERIFICATION_FAILED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SNAPSHOTSIGNATUREVERIFICATIONFAILED is a free data retrieval call binding the contract method 0x1d449ed1.
//
// Solidity: function SNAPSHOT_SIGNATURE_VERIFICATION_FAILED() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesSession) SNAPSHOTSIGNATUREVERIFICATIONFAILED() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTSIGNATUREVERIFICATIONFAILED(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTSIGNATUREVERIFICATIONFAILED is a free data retrieval call binding the contract method 0x1d449ed1.
//
// Solidity: function SNAPSHOT_SIGNATURE_VERIFICATION_FAILED() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCallerSession) SNAPSHOTSIGNATUREVERIFICATIONFAILED() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTSIGNATUREVERIFICATIONFAILED(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTWRONGMASTERPUBLICKEY is a free data retrieval call binding the contract method 0x85c9dba1.
//
// Solidity: function SNAPSHOT_WRONG_MASTER_PUBLIC_KEY() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCaller) SNAPSHOTWRONGMASTERPUBLICKEY(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SnapshotsErrorCodes.contract.Call(opts, &out, "SNAPSHOT_WRONG_MASTER_PUBLIC_KEY")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SNAPSHOTWRONGMASTERPUBLICKEY is a free data retrieval call binding the contract method 0x85c9dba1.
//
// Solidity: function SNAPSHOT_WRONG_MASTER_PUBLIC_KEY() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesSession) SNAPSHOTWRONGMASTERPUBLICKEY() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTWRONGMASTERPUBLICKEY(&_SnapshotsErrorCodes.CallOpts)
}

// SNAPSHOTWRONGMASTERPUBLICKEY is a free data retrieval call binding the contract method 0x85c9dba1.
//
// Solidity: function SNAPSHOT_WRONG_MASTER_PUBLIC_KEY() view returns(bytes32)
func (_SnapshotsErrorCodes *SnapshotsErrorCodesCallerSession) SNAPSHOTWRONGMASTERPUBLICKEY() ([32]byte, error) {
	return _SnapshotsErrorCodes.Contract.SNAPSHOTWRONGMASTERPUBLICKEY(&_SnapshotsErrorCodes.CallOpts)
}
