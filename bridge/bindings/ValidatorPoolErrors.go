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

// ValidatorPoolErrorsMetaData contains all meta data concerning the ValidatorPoolErrors contract.
var ValidatorPoolErrorsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"AddressAlreadyValidator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"AddressNotAccusable\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"AddressNotValidator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"CallerNotValidator\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ConsensusRunning\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ETHDKGRoundRunning\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"EthBalanceChangedDuringOperation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"stakeAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimumRequiredAmount\",\"type\":\"uint256\"}],\"name\":\"InsufficientFundsInStakePosition\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"InvalidIndex\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"availableValidators\",\"type\":\"uint256\"}],\"name\":\"LengthGreaterThanAvailableValidators\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"currentBlockNumber\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"targetBlockNumber\",\"type\":\"uint256\"}],\"name\":\"MinimumBlockIntervalNotMet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"requiredSlots\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"availableSlots\",\"type\":\"uint256\"}],\"name\":\"NotEnoughValidatorSlotsAvailable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OnlyStakingContractsAllowed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PayoutTooLow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProfitsOnlyClaimableWhileConsensusRunning\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorsLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stakerTokenIDsLength\",\"type\":\"uint256\"}],\"name\":\"RegistrationParameterLengthMismatch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"SenderNotInExitingQueue\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"positionId\",\"type\":\"uint256\"}],\"name\":\"SenderShouldOwnPosition\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TokenBalanceChangedDuringOperation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WaitingPeriodNotMet\",\"type\":\"error\"}]",
}

// ValidatorPoolErrorsABI is the input ABI used to generate the binding from.
// Deprecated: Use ValidatorPoolErrorsMetaData.ABI instead.
var ValidatorPoolErrorsABI = ValidatorPoolErrorsMetaData.ABI

// ValidatorPoolErrors is an auto generated Go binding around an Ethereum contract.
type ValidatorPoolErrors struct {
	ValidatorPoolErrorsCaller     // Read-only binding to the contract
	ValidatorPoolErrorsTransactor // Write-only binding to the contract
	ValidatorPoolErrorsFilterer   // Log filterer for contract events
}

// ValidatorPoolErrorsCaller is an auto generated read-only Go binding around an Ethereum contract.
type ValidatorPoolErrorsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorPoolErrorsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ValidatorPoolErrorsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorPoolErrorsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ValidatorPoolErrorsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorPoolErrorsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ValidatorPoolErrorsSession struct {
	Contract     *ValidatorPoolErrors // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// ValidatorPoolErrorsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ValidatorPoolErrorsCallerSession struct {
	Contract *ValidatorPoolErrorsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// ValidatorPoolErrorsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ValidatorPoolErrorsTransactorSession struct {
	Contract     *ValidatorPoolErrorsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// ValidatorPoolErrorsRaw is an auto generated low-level Go binding around an Ethereum contract.
type ValidatorPoolErrorsRaw struct {
	Contract *ValidatorPoolErrors // Generic contract binding to access the raw methods on
}

// ValidatorPoolErrorsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ValidatorPoolErrorsCallerRaw struct {
	Contract *ValidatorPoolErrorsCaller // Generic read-only contract binding to access the raw methods on
}

// ValidatorPoolErrorsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ValidatorPoolErrorsTransactorRaw struct {
	Contract *ValidatorPoolErrorsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewValidatorPoolErrors creates a new instance of ValidatorPoolErrors, bound to a specific deployed contract.
func NewValidatorPoolErrors(address common.Address, backend bind.ContractBackend) (*ValidatorPoolErrors, error) {
	contract, err := bindValidatorPoolErrors(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolErrors{ValidatorPoolErrorsCaller: ValidatorPoolErrorsCaller{contract: contract}, ValidatorPoolErrorsTransactor: ValidatorPoolErrorsTransactor{contract: contract}, ValidatorPoolErrorsFilterer: ValidatorPoolErrorsFilterer{contract: contract}}, nil
}

// NewValidatorPoolErrorsCaller creates a new read-only instance of ValidatorPoolErrors, bound to a specific deployed contract.
func NewValidatorPoolErrorsCaller(address common.Address, caller bind.ContractCaller) (*ValidatorPoolErrorsCaller, error) {
	contract, err := bindValidatorPoolErrors(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolErrorsCaller{contract: contract}, nil
}

// NewValidatorPoolErrorsTransactor creates a new write-only instance of ValidatorPoolErrors, bound to a specific deployed contract.
func NewValidatorPoolErrorsTransactor(address common.Address, transactor bind.ContractTransactor) (*ValidatorPoolErrorsTransactor, error) {
	contract, err := bindValidatorPoolErrors(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolErrorsTransactor{contract: contract}, nil
}

// NewValidatorPoolErrorsFilterer creates a new log filterer instance of ValidatorPoolErrors, bound to a specific deployed contract.
func NewValidatorPoolErrorsFilterer(address common.Address, filterer bind.ContractFilterer) (*ValidatorPoolErrorsFilterer, error) {
	contract, err := bindValidatorPoolErrors(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolErrorsFilterer{contract: contract}, nil
}

// bindValidatorPoolErrors binds a generic wrapper to an already deployed contract.
func bindValidatorPoolErrors(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ValidatorPoolErrorsABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ValidatorPoolErrors *ValidatorPoolErrorsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ValidatorPoolErrors.Contract.ValidatorPoolErrorsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ValidatorPoolErrors *ValidatorPoolErrorsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPoolErrors.Contract.ValidatorPoolErrorsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ValidatorPoolErrors *ValidatorPoolErrorsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ValidatorPoolErrors.Contract.ValidatorPoolErrorsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ValidatorPoolErrors *ValidatorPoolErrorsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ValidatorPoolErrors.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ValidatorPoolErrors *ValidatorPoolErrorsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPoolErrors.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ValidatorPoolErrors *ValidatorPoolErrorsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ValidatorPoolErrors.Contract.contract.Transact(opts, method, params...)
}
