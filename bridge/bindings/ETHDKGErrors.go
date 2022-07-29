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

// ETHDKGErrorsMetaData contains all meta data concerning the ETHDKGErrors contract.
var ETHDKGErrorsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"accused\",\"type\":\"address\"}],\"name\":\"AccusedDidNotDistributeSharesInRound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"accused\",\"type\":\"address\"}],\"name\":\"AccusedDidNotParticipateInGPKJSubmission\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"accused\",\"type\":\"address\"}],\"name\":\"AccusedDidNotSubmitGPKJInRound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"accused\",\"type\":\"address\"}],\"name\":\"AccusedDistributedGPKJ\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"accused\",\"type\":\"address\"}],\"name\":\"AccusedDistributedSharesInRound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"accused\",\"type\":\"address\"}],\"name\":\"AccusedHasCommitments\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"accused\",\"type\":\"address\"}],\"name\":\"AccusedNotParticipatingInRound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"accused\",\"type\":\"address\"}],\"name\":\"AccusedNotValidator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"accused\",\"type\":\"address\"}],\"name\":\"AccusedParticipatingInRound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"accused\",\"type\":\"address\"}],\"name\":\"AccusedSubmittedSharesInRound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorsLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"encryptedSharesHashLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"commitmentsLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numParticipants\",\"type\":\"uint256\"}],\"name\":\"ArgumentsLengthDoesNotEqualNumberOfParticipants\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CommitmentNotOnCurve\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CommitmentZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"disputer\",\"type\":\"address\"}],\"name\":\"DisputerDidNotDistributeSharesInRound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"disputer\",\"type\":\"address\"}],\"name\":\"DisputerDidNotSubmitGPKJInRound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"disputer\",\"type\":\"address\"}],\"name\":\"DisputerNotParticipatingInRound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DistributedShareHashZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumPhase\",\"name\":\"currentPhase\",\"type\":\"uint8\"}],\"name\":\"ETHDKGNotInDisputePhase\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumPhase\",\"name\":\"currentPhase\",\"type\":\"uint8\"}],\"name\":\"ETHDKGNotInGPKJSubmissionPhase\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumPhase\",\"name\":\"currentPhase\",\"type\":\"uint8\"}],\"name\":\"ETHDKGNotInKeyshareSubmissionPhase\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumPhase\",\"name\":\"currentPhase\",\"type\":\"uint8\"}],\"name\":\"ETHDKGNotInMasterPublicKeySubmissionPhase\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumPhase\",\"name\":\"currentPhase\",\"type\":\"uint8\"}],\"name\":\"ETHDKGNotInPostGPKJDisputePhase\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumPhase\",\"name\":\"currentPhase\",\"type\":\"uint8\"}],\"name\":\"ETHDKGNotInPostGPKJSubmissionPhase\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumPhase\",\"name\":\"currentPhase\",\"type\":\"uint8\"}],\"name\":\"ETHDKGNotInPostKeyshareSubmissionPhase\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumPhase\",\"name\":\"currentPhase\",\"type\":\"uint8\"}],\"name\":\"ETHDKGNotInPostRegistrationAccusationPhase\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumPhase\",\"name\":\"currentPhase\",\"type\":\"uint8\"}],\"name\":\"ETHDKGNotInRegistrationPhase\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumPhase\",\"name\":\"currentPhase\",\"type\":\"uint8\"}],\"name\":\"ETHDKGNotInSharedDistributionPhase\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ETHDKGRequisitesIncomplete\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"GPKJZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"commitmentsLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expectedCommitmentsLength\",\"type\":\"uint256\"}],\"name\":\"InvalidCommitments\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"commitmentsLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expectedCommitmentsLength\",\"type\":\"uint256\"}],\"name\":\"InvalidCommitmentsAmount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"sharesLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expectedSharesLength\",\"type\":\"uint256\"}],\"name\":\"InvalidEncryptedSharesAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidKeyOrProof\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidKeyshareG1\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidKeyshareG2\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"participantNonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"InvalidNonce\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"}],\"name\":\"InvalidOrDuplicatedParticipant\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"expectedHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"actualHash\",\"type\":\"bytes32\"}],\"name\":\"InvalidSharesOrCommitments\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MasterPublicKeyPairingCheckFailure\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorsAccountsLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorIndexesLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorSharesLength\",\"type\":\"uint256\"}],\"name\":\"MigrationInputDataMismatch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"MigrationRequiresZeroNonce\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"currentValidatorsLength\",\"type\":\"uint256\"}],\"name\":\"MinimumValidatorsNotMet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"enumPhase\",\"name\":\"currentPhase\",\"type\":\"uint8\"}],\"name\":\"NotInPostSharedDistributionPhase\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"OnlyValidatorsAllowed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"}],\"name\":\"ParticipantDistributedSharesInRound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"}],\"name\":\"ParticipantParticipatingInRound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"}],\"name\":\"ParticipantSubmittedGPKJInRound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"}],\"name\":\"ParticipantSubmittedKeysharesInRound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PublicKeyNotOnCurve\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PublicKeyZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"expected\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"actual\",\"type\":\"bytes32\"}],\"name\":\"SharesAndCommitmentsMismatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"VariableNotSettableWhileETHDKGRunning\",\"type\":\"error\"}]",
}

// ETHDKGErrorsABI is the input ABI used to generate the binding from.
// Deprecated: Use ETHDKGErrorsMetaData.ABI instead.
var ETHDKGErrorsABI = ETHDKGErrorsMetaData.ABI

// ETHDKGErrors is an auto generated Go binding around an Ethereum contract.
type ETHDKGErrors struct {
	ETHDKGErrorsCaller     // Read-only binding to the contract
	ETHDKGErrorsTransactor // Write-only binding to the contract
	ETHDKGErrorsFilterer   // Log filterer for contract events
}

// ETHDKGErrorsCaller is an auto generated read-only Go binding around an Ethereum contract.
type ETHDKGErrorsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ETHDKGErrorsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ETHDKGErrorsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ETHDKGErrorsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ETHDKGErrorsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ETHDKGErrorsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ETHDKGErrorsSession struct {
	Contract     *ETHDKGErrors     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ETHDKGErrorsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ETHDKGErrorsCallerSession struct {
	Contract *ETHDKGErrorsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// ETHDKGErrorsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ETHDKGErrorsTransactorSession struct {
	Contract     *ETHDKGErrorsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// ETHDKGErrorsRaw is an auto generated low-level Go binding around an Ethereum contract.
type ETHDKGErrorsRaw struct {
	Contract *ETHDKGErrors // Generic contract binding to access the raw methods on
}

// ETHDKGErrorsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ETHDKGErrorsCallerRaw struct {
	Contract *ETHDKGErrorsCaller // Generic read-only contract binding to access the raw methods on
}

// ETHDKGErrorsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ETHDKGErrorsTransactorRaw struct {
	Contract *ETHDKGErrorsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewETHDKGErrors creates a new instance of ETHDKGErrors, bound to a specific deployed contract.
func NewETHDKGErrors(address common.Address, backend bind.ContractBackend) (*ETHDKGErrors, error) {
	contract, err := bindETHDKGErrors(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ETHDKGErrors{ETHDKGErrorsCaller: ETHDKGErrorsCaller{contract: contract}, ETHDKGErrorsTransactor: ETHDKGErrorsTransactor{contract: contract}, ETHDKGErrorsFilterer: ETHDKGErrorsFilterer{contract: contract}}, nil
}

// NewETHDKGErrorsCaller creates a new read-only instance of ETHDKGErrors, bound to a specific deployed contract.
func NewETHDKGErrorsCaller(address common.Address, caller bind.ContractCaller) (*ETHDKGErrorsCaller, error) {
	contract, err := bindETHDKGErrors(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ETHDKGErrorsCaller{contract: contract}, nil
}

// NewETHDKGErrorsTransactor creates a new write-only instance of ETHDKGErrors, bound to a specific deployed contract.
func NewETHDKGErrorsTransactor(address common.Address, transactor bind.ContractTransactor) (*ETHDKGErrorsTransactor, error) {
	contract, err := bindETHDKGErrors(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ETHDKGErrorsTransactor{contract: contract}, nil
}

// NewETHDKGErrorsFilterer creates a new log filterer instance of ETHDKGErrors, bound to a specific deployed contract.
func NewETHDKGErrorsFilterer(address common.Address, filterer bind.ContractFilterer) (*ETHDKGErrorsFilterer, error) {
	contract, err := bindETHDKGErrors(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ETHDKGErrorsFilterer{contract: contract}, nil
}

// bindETHDKGErrors binds a generic wrapper to an already deployed contract.
func bindETHDKGErrors(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ETHDKGErrorsABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ETHDKGErrors *ETHDKGErrorsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ETHDKGErrors.Contract.ETHDKGErrorsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ETHDKGErrors *ETHDKGErrorsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ETHDKGErrors.Contract.ETHDKGErrorsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ETHDKGErrors *ETHDKGErrorsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ETHDKGErrors.Contract.ETHDKGErrorsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ETHDKGErrors *ETHDKGErrorsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ETHDKGErrors.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ETHDKGErrors *ETHDKGErrorsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ETHDKGErrors.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ETHDKGErrors *ETHDKGErrorsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ETHDKGErrors.Contract.contract.Transact(opts, method, params...)
}
