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

// ValidatorPoolErrorCodesMetaData contains all meta data concerning the ValidatorPoolErrorCodes contract.
var ValidatorPoolErrorCodesMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"VALIDATORPOOL_ADDRESS_ALREADY_VALIDATOR\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_ADDRESS_NOT_ACCUSABLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_ADDRESS_NOT_VALIDATOR\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_CALLER_NOT_VALIDATOR\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_CONSENSUS_RUNNING\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_DISHONEST_VALIDATOR_NOT_ACCUSABLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_ETHDKG_ROUND_RUNNING\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_ETH_BALANCE_CHANGED\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_FACTORY_SHOULD_OWN_POSITION\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_INSUFFICIENT_FUNDS_IN_STAKE_POSITION\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_INVALID_INDEX\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_MAX_VALIDATORS_MET\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_MINIMUM_STAKE_NOT_MET\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_MIN_BLOCK_INTERVAL_NOT_MET\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_ONLY_CONTRACTS_ALLOWED\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_PAYOUT_TOO_LOW\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_PROFITS_ONLY_CLAIMABLE_DURING_CONSENSUS\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_REGISTRATION_PARAMETER_LENGTH_MISMATCH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_SENDER_NOT_IN_EXITING_QUEUE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_TOKEN_BALANCE_CHANGED\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_VALIDATORS_GREATER_THAN_AVAILABLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VALIDATORPOOL_WAITING_PERIOD_NOT_MET\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ValidatorPoolErrorCodesABI is the input ABI used to generate the binding from.
// Deprecated: Use ValidatorPoolErrorCodesMetaData.ABI instead.
var ValidatorPoolErrorCodesABI = ValidatorPoolErrorCodesMetaData.ABI

// ValidatorPoolErrorCodes is an auto generated Go binding around an Ethereum contract.
type ValidatorPoolErrorCodes struct {
	ValidatorPoolErrorCodesCaller     // Read-only binding to the contract
	ValidatorPoolErrorCodesTransactor // Write-only binding to the contract
	ValidatorPoolErrorCodesFilterer   // Log filterer for contract events
}

// ValidatorPoolErrorCodesCaller is an auto generated read-only Go binding around an Ethereum contract.
type ValidatorPoolErrorCodesCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorPoolErrorCodesTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ValidatorPoolErrorCodesTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorPoolErrorCodesFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ValidatorPoolErrorCodesFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorPoolErrorCodesSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ValidatorPoolErrorCodesSession struct {
	Contract     *ValidatorPoolErrorCodes // Generic contract binding to set the session for
	CallOpts     bind.CallOpts            // Call options to use throughout this session
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// ValidatorPoolErrorCodesCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ValidatorPoolErrorCodesCallerSession struct {
	Contract *ValidatorPoolErrorCodesCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                  // Call options to use throughout this session
}

// ValidatorPoolErrorCodesTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ValidatorPoolErrorCodesTransactorSession struct {
	Contract     *ValidatorPoolErrorCodesTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                  // Transaction auth options to use throughout this session
}

// ValidatorPoolErrorCodesRaw is an auto generated low-level Go binding around an Ethereum contract.
type ValidatorPoolErrorCodesRaw struct {
	Contract *ValidatorPoolErrorCodes // Generic contract binding to access the raw methods on
}

// ValidatorPoolErrorCodesCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ValidatorPoolErrorCodesCallerRaw struct {
	Contract *ValidatorPoolErrorCodesCaller // Generic read-only contract binding to access the raw methods on
}

// ValidatorPoolErrorCodesTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ValidatorPoolErrorCodesTransactorRaw struct {
	Contract *ValidatorPoolErrorCodesTransactor // Generic write-only contract binding to access the raw methods on
}

// NewValidatorPoolErrorCodes creates a new instance of ValidatorPoolErrorCodes, bound to a specific deployed contract.
func NewValidatorPoolErrorCodes(address common.Address, backend bind.ContractBackend) (*ValidatorPoolErrorCodes, error) {
	contract, err := bindValidatorPoolErrorCodes(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolErrorCodes{ValidatorPoolErrorCodesCaller: ValidatorPoolErrorCodesCaller{contract: contract}, ValidatorPoolErrorCodesTransactor: ValidatorPoolErrorCodesTransactor{contract: contract}, ValidatorPoolErrorCodesFilterer: ValidatorPoolErrorCodesFilterer{contract: contract}}, nil
}

// NewValidatorPoolErrorCodesCaller creates a new read-only instance of ValidatorPoolErrorCodes, bound to a specific deployed contract.
func NewValidatorPoolErrorCodesCaller(address common.Address, caller bind.ContractCaller) (*ValidatorPoolErrorCodesCaller, error) {
	contract, err := bindValidatorPoolErrorCodes(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolErrorCodesCaller{contract: contract}, nil
}

// NewValidatorPoolErrorCodesTransactor creates a new write-only instance of ValidatorPoolErrorCodes, bound to a specific deployed contract.
func NewValidatorPoolErrorCodesTransactor(address common.Address, transactor bind.ContractTransactor) (*ValidatorPoolErrorCodesTransactor, error) {
	contract, err := bindValidatorPoolErrorCodes(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolErrorCodesTransactor{contract: contract}, nil
}

// NewValidatorPoolErrorCodesFilterer creates a new log filterer instance of ValidatorPoolErrorCodes, bound to a specific deployed contract.
func NewValidatorPoolErrorCodesFilterer(address common.Address, filterer bind.ContractFilterer) (*ValidatorPoolErrorCodesFilterer, error) {
	contract, err := bindValidatorPoolErrorCodes(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolErrorCodesFilterer{contract: contract}, nil
}

// bindValidatorPoolErrorCodes binds a generic wrapper to an already deployed contract.
func bindValidatorPoolErrorCodes(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ValidatorPoolErrorCodesABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ValidatorPoolErrorCodes.Contract.ValidatorPoolErrorCodesCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPoolErrorCodes.Contract.ValidatorPoolErrorCodesTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ValidatorPoolErrorCodes.Contract.ValidatorPoolErrorCodesTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ValidatorPoolErrorCodes.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPoolErrorCodes.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ValidatorPoolErrorCodes.Contract.contract.Transact(opts, method, params...)
}

// VALIDATORPOOLADDRESSALREADYVALIDATOR is a free data retrieval call binding the contract method 0xb8e6dc4d.
//
// Solidity: function VALIDATORPOOL_ADDRESS_ALREADY_VALIDATOR() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLADDRESSALREADYVALIDATOR(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_ADDRESS_ALREADY_VALIDATOR")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLADDRESSALREADYVALIDATOR is a free data retrieval call binding the contract method 0xb8e6dc4d.
//
// Solidity: function VALIDATORPOOL_ADDRESS_ALREADY_VALIDATOR() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLADDRESSALREADYVALIDATOR() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLADDRESSALREADYVALIDATOR(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLADDRESSALREADYVALIDATOR is a free data retrieval call binding the contract method 0xb8e6dc4d.
//
// Solidity: function VALIDATORPOOL_ADDRESS_ALREADY_VALIDATOR() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLADDRESSALREADYVALIDATOR() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLADDRESSALREADYVALIDATOR(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLADDRESSNOTACCUSABLE is a free data retrieval call binding the contract method 0x0e144d69.
//
// Solidity: function VALIDATORPOOL_ADDRESS_NOT_ACCUSABLE() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLADDRESSNOTACCUSABLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_ADDRESS_NOT_ACCUSABLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLADDRESSNOTACCUSABLE is a free data retrieval call binding the contract method 0x0e144d69.
//
// Solidity: function VALIDATORPOOL_ADDRESS_NOT_ACCUSABLE() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLADDRESSNOTACCUSABLE() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLADDRESSNOTACCUSABLE(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLADDRESSNOTACCUSABLE is a free data retrieval call binding the contract method 0x0e144d69.
//
// Solidity: function VALIDATORPOOL_ADDRESS_NOT_ACCUSABLE() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLADDRESSNOTACCUSABLE() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLADDRESSNOTACCUSABLE(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLADDRESSNOTVALIDATOR is a free data retrieval call binding the contract method 0x4ce78d00.
//
// Solidity: function VALIDATORPOOL_ADDRESS_NOT_VALIDATOR() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLADDRESSNOTVALIDATOR(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_ADDRESS_NOT_VALIDATOR")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLADDRESSNOTVALIDATOR is a free data retrieval call binding the contract method 0x4ce78d00.
//
// Solidity: function VALIDATORPOOL_ADDRESS_NOT_VALIDATOR() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLADDRESSNOTVALIDATOR() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLADDRESSNOTVALIDATOR(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLADDRESSNOTVALIDATOR is a free data retrieval call binding the contract method 0x4ce78d00.
//
// Solidity: function VALIDATORPOOL_ADDRESS_NOT_VALIDATOR() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLADDRESSNOTVALIDATOR() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLADDRESSNOTVALIDATOR(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLCALLERNOTVALIDATOR is a free data retrieval call binding the contract method 0xdf6f3c26.
//
// Solidity: function VALIDATORPOOL_CALLER_NOT_VALIDATOR() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLCALLERNOTVALIDATOR(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_CALLER_NOT_VALIDATOR")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLCALLERNOTVALIDATOR is a free data retrieval call binding the contract method 0xdf6f3c26.
//
// Solidity: function VALIDATORPOOL_CALLER_NOT_VALIDATOR() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLCALLERNOTVALIDATOR() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLCALLERNOTVALIDATOR(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLCALLERNOTVALIDATOR is a free data retrieval call binding the contract method 0xdf6f3c26.
//
// Solidity: function VALIDATORPOOL_CALLER_NOT_VALIDATOR() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLCALLERNOTVALIDATOR() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLCALLERNOTVALIDATOR(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLCONSENSUSRUNNING is a free data retrieval call binding the contract method 0x8ee837c4.
//
// Solidity: function VALIDATORPOOL_CONSENSUS_RUNNING() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLCONSENSUSRUNNING(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_CONSENSUS_RUNNING")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLCONSENSUSRUNNING is a free data retrieval call binding the contract method 0x8ee837c4.
//
// Solidity: function VALIDATORPOOL_CONSENSUS_RUNNING() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLCONSENSUSRUNNING() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLCONSENSUSRUNNING(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLCONSENSUSRUNNING is a free data retrieval call binding the contract method 0x8ee837c4.
//
// Solidity: function VALIDATORPOOL_CONSENSUS_RUNNING() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLCONSENSUSRUNNING() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLCONSENSUSRUNNING(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLDISHONESTVALIDATORNOTACCUSABLE is a free data retrieval call binding the contract method 0x4c08af9b.
//
// Solidity: function VALIDATORPOOL_DISHONEST_VALIDATOR_NOT_ACCUSABLE() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLDISHONESTVALIDATORNOTACCUSABLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_DISHONEST_VALIDATOR_NOT_ACCUSABLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLDISHONESTVALIDATORNOTACCUSABLE is a free data retrieval call binding the contract method 0x4c08af9b.
//
// Solidity: function VALIDATORPOOL_DISHONEST_VALIDATOR_NOT_ACCUSABLE() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLDISHONESTVALIDATORNOTACCUSABLE() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLDISHONESTVALIDATORNOTACCUSABLE(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLDISHONESTVALIDATORNOTACCUSABLE is a free data retrieval call binding the contract method 0x4c08af9b.
//
// Solidity: function VALIDATORPOOL_DISHONEST_VALIDATOR_NOT_ACCUSABLE() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLDISHONESTVALIDATORNOTACCUSABLE() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLDISHONESTVALIDATORNOTACCUSABLE(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLETHDKGROUNDRUNNING is a free data retrieval call binding the contract method 0x9126af77.
//
// Solidity: function VALIDATORPOOL_ETHDKG_ROUND_RUNNING() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLETHDKGROUNDRUNNING(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_ETHDKG_ROUND_RUNNING")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLETHDKGROUNDRUNNING is a free data retrieval call binding the contract method 0x9126af77.
//
// Solidity: function VALIDATORPOOL_ETHDKG_ROUND_RUNNING() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLETHDKGROUNDRUNNING() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLETHDKGROUNDRUNNING(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLETHDKGROUNDRUNNING is a free data retrieval call binding the contract method 0x9126af77.
//
// Solidity: function VALIDATORPOOL_ETHDKG_ROUND_RUNNING() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLETHDKGROUNDRUNNING() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLETHDKGROUNDRUNNING(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLETHBALANCECHANGED is a free data retrieval call binding the contract method 0x826b711b.
//
// Solidity: function VALIDATORPOOL_ETH_BALANCE_CHANGED() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLETHBALANCECHANGED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_ETH_BALANCE_CHANGED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLETHBALANCECHANGED is a free data retrieval call binding the contract method 0x826b711b.
//
// Solidity: function VALIDATORPOOL_ETH_BALANCE_CHANGED() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLETHBALANCECHANGED() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLETHBALANCECHANGED(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLETHBALANCECHANGED is a free data retrieval call binding the contract method 0x826b711b.
//
// Solidity: function VALIDATORPOOL_ETH_BALANCE_CHANGED() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLETHBALANCECHANGED() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLETHBALANCECHANGED(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLFACTORYSHOULDOWNPOSITION is a free data retrieval call binding the contract method 0x527991a3.
//
// Solidity: function VALIDATORPOOL_FACTORY_SHOULD_OWN_POSITION() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLFACTORYSHOULDOWNPOSITION(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_FACTORY_SHOULD_OWN_POSITION")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLFACTORYSHOULDOWNPOSITION is a free data retrieval call binding the contract method 0x527991a3.
//
// Solidity: function VALIDATORPOOL_FACTORY_SHOULD_OWN_POSITION() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLFACTORYSHOULDOWNPOSITION() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLFACTORYSHOULDOWNPOSITION(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLFACTORYSHOULDOWNPOSITION is a free data retrieval call binding the contract method 0x527991a3.
//
// Solidity: function VALIDATORPOOL_FACTORY_SHOULD_OWN_POSITION() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLFACTORYSHOULDOWNPOSITION() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLFACTORYSHOULDOWNPOSITION(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLINSUFFICIENTFUNDSINSTAKEPOSITION is a free data retrieval call binding the contract method 0x2c20647a.
//
// Solidity: function VALIDATORPOOL_INSUFFICIENT_FUNDS_IN_STAKE_POSITION() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLINSUFFICIENTFUNDSINSTAKEPOSITION(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_INSUFFICIENT_FUNDS_IN_STAKE_POSITION")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLINSUFFICIENTFUNDSINSTAKEPOSITION is a free data retrieval call binding the contract method 0x2c20647a.
//
// Solidity: function VALIDATORPOOL_INSUFFICIENT_FUNDS_IN_STAKE_POSITION() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLINSUFFICIENTFUNDSINSTAKEPOSITION() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLINSUFFICIENTFUNDSINSTAKEPOSITION(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLINSUFFICIENTFUNDSINSTAKEPOSITION is a free data retrieval call binding the contract method 0x2c20647a.
//
// Solidity: function VALIDATORPOOL_INSUFFICIENT_FUNDS_IN_STAKE_POSITION() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLINSUFFICIENTFUNDSINSTAKEPOSITION() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLINSUFFICIENTFUNDSINSTAKEPOSITION(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLINVALIDINDEX is a free data retrieval call binding the contract method 0x38dce6e0.
//
// Solidity: function VALIDATORPOOL_INVALID_INDEX() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLINVALIDINDEX(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_INVALID_INDEX")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLINVALIDINDEX is a free data retrieval call binding the contract method 0x38dce6e0.
//
// Solidity: function VALIDATORPOOL_INVALID_INDEX() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLINVALIDINDEX() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLINVALIDINDEX(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLINVALIDINDEX is a free data retrieval call binding the contract method 0x38dce6e0.
//
// Solidity: function VALIDATORPOOL_INVALID_INDEX() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLINVALIDINDEX() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLINVALIDINDEX(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLMAXVALIDATORSMET is a free data retrieval call binding the contract method 0xfb5f6a21.
//
// Solidity: function VALIDATORPOOL_MAX_VALIDATORS_MET() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLMAXVALIDATORSMET(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_MAX_VALIDATORS_MET")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLMAXVALIDATORSMET is a free data retrieval call binding the contract method 0xfb5f6a21.
//
// Solidity: function VALIDATORPOOL_MAX_VALIDATORS_MET() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLMAXVALIDATORSMET() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLMAXVALIDATORSMET(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLMAXVALIDATORSMET is a free data retrieval call binding the contract method 0xfb5f6a21.
//
// Solidity: function VALIDATORPOOL_MAX_VALIDATORS_MET() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLMAXVALIDATORSMET() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLMAXVALIDATORSMET(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLMINIMUMSTAKENOTMET is a free data retrieval call binding the contract method 0x057f2b29.
//
// Solidity: function VALIDATORPOOL_MINIMUM_STAKE_NOT_MET() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLMINIMUMSTAKENOTMET(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_MINIMUM_STAKE_NOT_MET")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLMINIMUMSTAKENOTMET is a free data retrieval call binding the contract method 0x057f2b29.
//
// Solidity: function VALIDATORPOOL_MINIMUM_STAKE_NOT_MET() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLMINIMUMSTAKENOTMET() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLMINIMUMSTAKENOTMET(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLMINIMUMSTAKENOTMET is a free data retrieval call binding the contract method 0x057f2b29.
//
// Solidity: function VALIDATORPOOL_MINIMUM_STAKE_NOT_MET() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLMINIMUMSTAKENOTMET() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLMINIMUMSTAKENOTMET(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLMINBLOCKINTERVALNOTMET is a free data retrieval call binding the contract method 0x08abeefe.
//
// Solidity: function VALIDATORPOOL_MIN_BLOCK_INTERVAL_NOT_MET() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLMINBLOCKINTERVALNOTMET(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_MIN_BLOCK_INTERVAL_NOT_MET")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLMINBLOCKINTERVALNOTMET is a free data retrieval call binding the contract method 0x08abeefe.
//
// Solidity: function VALIDATORPOOL_MIN_BLOCK_INTERVAL_NOT_MET() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLMINBLOCKINTERVALNOTMET() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLMINBLOCKINTERVALNOTMET(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLMINBLOCKINTERVALNOTMET is a free data retrieval call binding the contract method 0x08abeefe.
//
// Solidity: function VALIDATORPOOL_MIN_BLOCK_INTERVAL_NOT_MET() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLMINBLOCKINTERVALNOTMET() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLMINBLOCKINTERVALNOTMET(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLONLYCONTRACTSALLOWED is a free data retrieval call binding the contract method 0x7527a98f.
//
// Solidity: function VALIDATORPOOL_ONLY_CONTRACTS_ALLOWED() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLONLYCONTRACTSALLOWED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_ONLY_CONTRACTS_ALLOWED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLONLYCONTRACTSALLOWED is a free data retrieval call binding the contract method 0x7527a98f.
//
// Solidity: function VALIDATORPOOL_ONLY_CONTRACTS_ALLOWED() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLONLYCONTRACTSALLOWED() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLONLYCONTRACTSALLOWED(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLONLYCONTRACTSALLOWED is a free data retrieval call binding the contract method 0x7527a98f.
//
// Solidity: function VALIDATORPOOL_ONLY_CONTRACTS_ALLOWED() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLONLYCONTRACTSALLOWED() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLONLYCONTRACTSALLOWED(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLPAYOUTTOOLOW is a free data retrieval call binding the contract method 0x0ce7d41e.
//
// Solidity: function VALIDATORPOOL_PAYOUT_TOO_LOW() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLPAYOUTTOOLOW(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_PAYOUT_TOO_LOW")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLPAYOUTTOOLOW is a free data retrieval call binding the contract method 0x0ce7d41e.
//
// Solidity: function VALIDATORPOOL_PAYOUT_TOO_LOW() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLPAYOUTTOOLOW() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLPAYOUTTOOLOW(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLPAYOUTTOOLOW is a free data retrieval call binding the contract method 0x0ce7d41e.
//
// Solidity: function VALIDATORPOOL_PAYOUT_TOO_LOW() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLPAYOUTTOOLOW() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLPAYOUTTOOLOW(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLPROFITSONLYCLAIMABLEDURINGCONSENSUS is a free data retrieval call binding the contract method 0x4c454a97.
//
// Solidity: function VALIDATORPOOL_PROFITS_ONLY_CLAIMABLE_DURING_CONSENSUS() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLPROFITSONLYCLAIMABLEDURINGCONSENSUS(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_PROFITS_ONLY_CLAIMABLE_DURING_CONSENSUS")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLPROFITSONLYCLAIMABLEDURINGCONSENSUS is a free data retrieval call binding the contract method 0x4c454a97.
//
// Solidity: function VALIDATORPOOL_PROFITS_ONLY_CLAIMABLE_DURING_CONSENSUS() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLPROFITSONLYCLAIMABLEDURINGCONSENSUS() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLPROFITSONLYCLAIMABLEDURINGCONSENSUS(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLPROFITSONLYCLAIMABLEDURINGCONSENSUS is a free data retrieval call binding the contract method 0x4c454a97.
//
// Solidity: function VALIDATORPOOL_PROFITS_ONLY_CLAIMABLE_DURING_CONSENSUS() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLPROFITSONLYCLAIMABLEDURINGCONSENSUS() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLPROFITSONLYCLAIMABLEDURINGCONSENSUS(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLREGISTRATIONPARAMETERLENGTHMISMATCH is a free data retrieval call binding the contract method 0xe5b50fe6.
//
// Solidity: function VALIDATORPOOL_REGISTRATION_PARAMETER_LENGTH_MISMATCH() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLREGISTRATIONPARAMETERLENGTHMISMATCH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_REGISTRATION_PARAMETER_LENGTH_MISMATCH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLREGISTRATIONPARAMETERLENGTHMISMATCH is a free data retrieval call binding the contract method 0xe5b50fe6.
//
// Solidity: function VALIDATORPOOL_REGISTRATION_PARAMETER_LENGTH_MISMATCH() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLREGISTRATIONPARAMETERLENGTHMISMATCH() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLREGISTRATIONPARAMETERLENGTHMISMATCH(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLREGISTRATIONPARAMETERLENGTHMISMATCH is a free data retrieval call binding the contract method 0xe5b50fe6.
//
// Solidity: function VALIDATORPOOL_REGISTRATION_PARAMETER_LENGTH_MISMATCH() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLREGISTRATIONPARAMETERLENGTHMISMATCH() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLREGISTRATIONPARAMETERLENGTHMISMATCH(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLSENDERNOTINEXITINGQUEUE is a free data retrieval call binding the contract method 0x55efc658.
//
// Solidity: function VALIDATORPOOL_SENDER_NOT_IN_EXITING_QUEUE() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLSENDERNOTINEXITINGQUEUE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_SENDER_NOT_IN_EXITING_QUEUE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLSENDERNOTINEXITINGQUEUE is a free data retrieval call binding the contract method 0x55efc658.
//
// Solidity: function VALIDATORPOOL_SENDER_NOT_IN_EXITING_QUEUE() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLSENDERNOTINEXITINGQUEUE() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLSENDERNOTINEXITINGQUEUE(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLSENDERNOTINEXITINGQUEUE is a free data retrieval call binding the contract method 0x55efc658.
//
// Solidity: function VALIDATORPOOL_SENDER_NOT_IN_EXITING_QUEUE() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLSENDERNOTINEXITINGQUEUE() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLSENDERNOTINEXITINGQUEUE(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLTOKENBALANCECHANGED is a free data retrieval call binding the contract method 0x329fa341.
//
// Solidity: function VALIDATORPOOL_TOKEN_BALANCE_CHANGED() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLTOKENBALANCECHANGED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_TOKEN_BALANCE_CHANGED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLTOKENBALANCECHANGED is a free data retrieval call binding the contract method 0x329fa341.
//
// Solidity: function VALIDATORPOOL_TOKEN_BALANCE_CHANGED() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLTOKENBALANCECHANGED() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLTOKENBALANCECHANGED(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLTOKENBALANCECHANGED is a free data retrieval call binding the contract method 0x329fa341.
//
// Solidity: function VALIDATORPOOL_TOKEN_BALANCE_CHANGED() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLTOKENBALANCECHANGED() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLTOKENBALANCECHANGED(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLVALIDATORSGREATERTHANAVAILABLE is a free data retrieval call binding the contract method 0x7cd62d16.
//
// Solidity: function VALIDATORPOOL_VALIDATORS_GREATER_THAN_AVAILABLE() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLVALIDATORSGREATERTHANAVAILABLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_VALIDATORS_GREATER_THAN_AVAILABLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLVALIDATORSGREATERTHANAVAILABLE is a free data retrieval call binding the contract method 0x7cd62d16.
//
// Solidity: function VALIDATORPOOL_VALIDATORS_GREATER_THAN_AVAILABLE() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLVALIDATORSGREATERTHANAVAILABLE() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLVALIDATORSGREATERTHANAVAILABLE(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLVALIDATORSGREATERTHANAVAILABLE is a free data retrieval call binding the contract method 0x7cd62d16.
//
// Solidity: function VALIDATORPOOL_VALIDATORS_GREATER_THAN_AVAILABLE() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLVALIDATORSGREATERTHANAVAILABLE() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLVALIDATORSGREATERTHANAVAILABLE(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLWAITINGPERIODNOTMET is a free data retrieval call binding the contract method 0x0e2d4d55.
//
// Solidity: function VALIDATORPOOL_WAITING_PERIOD_NOT_MET() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCaller) VALIDATORPOOLWAITINGPERIODNOTMET(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ValidatorPoolErrorCodes.contract.Call(opts, &out, "VALIDATORPOOL_WAITING_PERIOD_NOT_MET")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VALIDATORPOOLWAITINGPERIODNOTMET is a free data retrieval call binding the contract method 0x0e2d4d55.
//
// Solidity: function VALIDATORPOOL_WAITING_PERIOD_NOT_MET() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesSession) VALIDATORPOOLWAITINGPERIODNOTMET() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLWAITINGPERIODNOTMET(&_ValidatorPoolErrorCodes.CallOpts)
}

// VALIDATORPOOLWAITINGPERIODNOTMET is a free data retrieval call binding the contract method 0x0e2d4d55.
//
// Solidity: function VALIDATORPOOL_WAITING_PERIOD_NOT_MET() view returns(bytes32)
func (_ValidatorPoolErrorCodes *ValidatorPoolErrorCodesCallerSession) VALIDATORPOOLWAITINGPERIODNOTMET() ([32]byte, error) {
	return _ValidatorPoolErrorCodes.Contract.VALIDATORPOOLWAITINGPERIODNOTMET(&_ValidatorPoolErrorCodes.CallOpts)
}
