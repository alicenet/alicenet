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

// Participant is an auto generated low-level Go binding around an user-defined struct.
type Participant struct {
	PublicKey                   [2]*big.Int
	Nonce                       uint64
	Index                       uint64
	Phase                       uint8
	DistributedSharesHash       [32]byte
	CommitmentsFirstCoefficient [2]*big.Int
	KeyShares                   [2]*big.Int
	Gpkj                        [4]*big.Int
}

// ETHDKGMetaData contains all meta data concerning the ETHDKG contract.
var ETHDKGMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorsAccountsLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorIndexesLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorSharesLength\",\"type\":\"uint256\"}],\"name\":\"MigrationInputDataMismatch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"MigrationRequiresZeroNonce\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"currentValidatorsLength\",\"type\":\"uint256\"}],\"name\":\"MinimumValidatorsNotMet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyETHDKGAccusations\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyETHDKGPhases\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyFactory\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlySnapshots\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyValidatorPool\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"OnlyValidatorsAllowed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"ParticipantNotFoundInLastRound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"VariableNotSettableWhileETHDKGRunning\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256[2]\",\"name\":\"publicKey\",\"type\":\"uint256[2]\"}],\"name\":\"AddressRegistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"GPKJSubmissionComplete\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"KeyShareSubmissionComplete\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256[2]\",\"name\":\"keyShareG1\",\"type\":\"uint256[2]\"},{\"indexed\":false,\"internalType\":\"uint256[2]\",\"name\":\"keyShareG1CorrectnessProof\",\"type\":\"uint256[2]\"},{\"indexed\":false,\"internalType\":\"uint256[4]\",\"name\":\"keyShareG2\",\"type\":\"uint256[4]\"}],\"name\":\"KeyShareSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256[4]\",\"name\":\"mpk\",\"type\":\"uint256[4]\"}],\"name\":\"MPKSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"RegistrationComplete\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"startBlock\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"numberValidators\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"phaseLength\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"confirmationLength\",\"type\":\"uint256\"}],\"name\":\"RegistrationOpened\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"ShareDistributionComplete\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"encryptedShares\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"uint256[2][]\",\"name\":\"commitments\",\"type\":\"uint256[2][]\"}],\"name\":\"SharesDistributed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"share0\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"share1\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"share2\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"share3\",\"type\":\"uint256\"}],\"name\":\"ValidatorMemberAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"validatorCount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"ethHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"aliceNetHeight\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"groupKey0\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"groupKey1\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"groupKey2\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"groupKey3\",\"type\":\"uint256\"}],\"name\":\"ValidatorSetCompleted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"dishonestAddresses\",\"type\":\"address[]\"}],\"name\":\"accuseParticipantDidNotDistributeShares\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"dishonestAddresses\",\"type\":\"address[]\"}],\"name\":\"accuseParticipantDidNotSubmitGPKJ\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"dishonestAddresses\",\"type\":\"address[]\"}],\"name\":\"accuseParticipantDidNotSubmitKeyShares\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"dishonestAddress\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"encryptedShares\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[2][]\",\"name\":\"commitments\",\"type\":\"uint256[2][]\"},{\"internalType\":\"uint256[2]\",\"name\":\"sharedKey\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"sharedKeyCorrectnessProof\",\"type\":\"uint256[2]\"}],\"name\":\"accuseParticipantDistributedBadShares\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"dishonestAddresses\",\"type\":\"address[]\"}],\"name\":\"accuseParticipantNotRegistered\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"validators\",\"type\":\"address[]\"},{\"internalType\":\"bytes32[]\",\"name\":\"encryptedSharesHash\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256[2][][]\",\"name\":\"commitments\",\"type\":\"uint256[2][][]\"},{\"internalType\":\"address\",\"name\":\"dishonestAddress\",\"type\":\"address\"}],\"name\":\"accuseParticipantSubmittedBadGPKJ\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"complete\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"encryptedShares\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[2][]\",\"name\":\"commitments\",\"type\":\"uint256[2][]\"}],\"name\":\"distributeShares\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getBadParticipants\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getConfirmationLength\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getETHDKGPhase\",\"outputs\":[{\"internalType\":\"enumPhase\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"}],\"name\":\"getLastRoundParticipantIndex\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMasterPublicKey\",\"outputs\":[{\"internalType\":\"uint256[4]\",\"name\":\"\",\"type\":\"uint256[4]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMasterPublicKeyHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMinValidators\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNumParticipants\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"}],\"name\":\"getParticipantInternalState\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"publicKey\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"index\",\"type\":\"uint64\"},{\"internalType\":\"enumPhase\",\"name\":\"phase\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"distributedSharesHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[2]\",\"name\":\"commitmentsFirstCoefficient\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"keyShares\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[4]\",\"name\":\"gpkj\",\"type\":\"uint256[4]\"}],\"internalType\":\"structParticipant\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"participantAddresses\",\"type\":\"address[]\"}],\"name\":\"getParticipantsInternalState\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"publicKey\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"index\",\"type\":\"uint64\"},{\"internalType\":\"enumPhase\",\"name\":\"phase\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"distributedSharesHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[2]\",\"name\":\"commitmentsFirstCoefficient\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"keyShares\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[4]\",\"name\":\"gpkj\",\"type\":\"uint256[4]\"}],\"internalType\":\"structParticipant[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPhaseLength\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPhaseStartBlock\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"phaseLength_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"confirmationLength_\",\"type\":\"uint256\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initializeETHDKG\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isETHDKGCompleted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isETHDKGHalted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isETHDKGRunning\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isMasterPublicKeySet\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"validatorsAccounts_\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"validatorIndexes_\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[4][]\",\"name\":\"validatorShares_\",\"type\":\"uint256[4][]\"},{\"internalType\":\"uint8\",\"name\":\"validatorCount_\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"epoch_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"sideChainHeight_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"ethHeight_\",\"type\":\"uint256\"},{\"internalType\":\"uint256[4]\",\"name\":\"masterPublicKey_\",\"type\":\"uint256[4]\"}],\"name\":\"migrateValidators\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[2]\",\"name\":\"publicKey\",\"type\":\"uint256[2]\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"confirmationLength_\",\"type\":\"uint16\"}],\"name\":\"setConfirmationLength\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"aliceNetHeight\",\"type\":\"uint256\"}],\"name\":\"setCustomAliceNetHeight\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"phaseLength_\",\"type\":\"uint16\"}],\"name\":\"setPhaseLength\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[4]\",\"name\":\"gpkj\",\"type\":\"uint256[4]\"}],\"name\":\"submitGPKJ\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[2]\",\"name\":\"keyShareG1\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"keyShareG1CorrectnessProof\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[4]\",\"name\":\"keyShareG2\",\"type\":\"uint256[4]\"}],\"name\":\"submitKeyShare\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[4]\",\"name\":\"masterPublicKey_\",\"type\":\"uint256[4]\"}],\"name\":\"submitMasterPublicKey\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ETHDKGABI is the input ABI used to generate the binding from.
// Deprecated: Use ETHDKGMetaData.ABI instead.
var ETHDKGABI = ETHDKGMetaData.ABI

// ETHDKG is an auto generated Go binding around an Ethereum contract.
type ETHDKG struct {
	ETHDKGCaller     // Read-only binding to the contract
	ETHDKGTransactor // Write-only binding to the contract
	ETHDKGFilterer   // Log filterer for contract events
}

// ETHDKGCaller is an auto generated read-only Go binding around an Ethereum contract.
type ETHDKGCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ETHDKGTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ETHDKGTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ETHDKGFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ETHDKGFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ETHDKGSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ETHDKGSession struct {
	Contract     *ETHDKG           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ETHDKGCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ETHDKGCallerSession struct {
	Contract *ETHDKGCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ETHDKGTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ETHDKGTransactorSession struct {
	Contract     *ETHDKGTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ETHDKGRaw is an auto generated low-level Go binding around an Ethereum contract.
type ETHDKGRaw struct {
	Contract *ETHDKG // Generic contract binding to access the raw methods on
}

// ETHDKGCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ETHDKGCallerRaw struct {
	Contract *ETHDKGCaller // Generic read-only contract binding to access the raw methods on
}

// ETHDKGTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ETHDKGTransactorRaw struct {
	Contract *ETHDKGTransactor // Generic write-only contract binding to access the raw methods on
}

// NewETHDKG creates a new instance of ETHDKG, bound to a specific deployed contract.
func NewETHDKG(address common.Address, backend bind.ContractBackend) (*ETHDKG, error) {
	contract, err := bindETHDKG(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ETHDKG{ETHDKGCaller: ETHDKGCaller{contract: contract}, ETHDKGTransactor: ETHDKGTransactor{contract: contract}, ETHDKGFilterer: ETHDKGFilterer{contract: contract}}, nil
}

// NewETHDKGCaller creates a new read-only instance of ETHDKG, bound to a specific deployed contract.
func NewETHDKGCaller(address common.Address, caller bind.ContractCaller) (*ETHDKGCaller, error) {
	contract, err := bindETHDKG(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ETHDKGCaller{contract: contract}, nil
}

// NewETHDKGTransactor creates a new write-only instance of ETHDKG, bound to a specific deployed contract.
func NewETHDKGTransactor(address common.Address, transactor bind.ContractTransactor) (*ETHDKGTransactor, error) {
	contract, err := bindETHDKG(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ETHDKGTransactor{contract: contract}, nil
}

// NewETHDKGFilterer creates a new log filterer instance of ETHDKG, bound to a specific deployed contract.
func NewETHDKGFilterer(address common.Address, filterer bind.ContractFilterer) (*ETHDKGFilterer, error) {
	contract, err := bindETHDKG(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ETHDKGFilterer{contract: contract}, nil
}

// bindETHDKG binds a generic wrapper to an already deployed contract.
func bindETHDKG(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ETHDKGABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ETHDKG *ETHDKGRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ETHDKG.Contract.ETHDKGCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ETHDKG *ETHDKGRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ETHDKG.Contract.ETHDKGTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ETHDKG *ETHDKGRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ETHDKG.Contract.ETHDKGTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ETHDKG *ETHDKGCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ETHDKG.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ETHDKG *ETHDKGTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ETHDKG.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ETHDKG *ETHDKGTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ETHDKG.Contract.contract.Transact(opts, method, params...)
}

// GetBadParticipants is a free data retrieval call binding the contract method 0x32d4d570.
//
// Solidity: function getBadParticipants() view returns(uint256)
func (_ETHDKG *ETHDKGCaller) GetBadParticipants(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getBadParticipants")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetBadParticipants is a free data retrieval call binding the contract method 0x32d4d570.
//
// Solidity: function getBadParticipants() view returns(uint256)
func (_ETHDKG *ETHDKGSession) GetBadParticipants() (*big.Int, error) {
	return _ETHDKG.Contract.GetBadParticipants(&_ETHDKG.CallOpts)
}

// GetBadParticipants is a free data retrieval call binding the contract method 0x32d4d570.
//
// Solidity: function getBadParticipants() view returns(uint256)
func (_ETHDKG *ETHDKGCallerSession) GetBadParticipants() (*big.Int, error) {
	return _ETHDKG.Contract.GetBadParticipants(&_ETHDKG.CallOpts)
}

// GetConfirmationLength is a free data retrieval call binding the contract method 0x8c848d32.
//
// Solidity: function getConfirmationLength() view returns(uint256)
func (_ETHDKG *ETHDKGCaller) GetConfirmationLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getConfirmationLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetConfirmationLength is a free data retrieval call binding the contract method 0x8c848d32.
//
// Solidity: function getConfirmationLength() view returns(uint256)
func (_ETHDKG *ETHDKGSession) GetConfirmationLength() (*big.Int, error) {
	return _ETHDKG.Contract.GetConfirmationLength(&_ETHDKG.CallOpts)
}

// GetConfirmationLength is a free data retrieval call binding the contract method 0x8c848d32.
//
// Solidity: function getConfirmationLength() view returns(uint256)
func (_ETHDKG *ETHDKGCallerSession) GetConfirmationLength() (*big.Int, error) {
	return _ETHDKG.Contract.GetConfirmationLength(&_ETHDKG.CallOpts)
}

// GetETHDKGPhase is a free data retrieval call binding the contract method 0x2958e81c.
//
// Solidity: function getETHDKGPhase() view returns(uint8)
func (_ETHDKG *ETHDKGCaller) GetETHDKGPhase(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getETHDKGPhase")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetETHDKGPhase is a free data retrieval call binding the contract method 0x2958e81c.
//
// Solidity: function getETHDKGPhase() view returns(uint8)
func (_ETHDKG *ETHDKGSession) GetETHDKGPhase() (uint8, error) {
	return _ETHDKG.Contract.GetETHDKGPhase(&_ETHDKG.CallOpts)
}

// GetETHDKGPhase is a free data retrieval call binding the contract method 0x2958e81c.
//
// Solidity: function getETHDKGPhase() view returns(uint8)
func (_ETHDKG *ETHDKGCallerSession) GetETHDKGPhase() (uint8, error) {
	return _ETHDKG.Contract.GetETHDKGPhase(&_ETHDKG.CallOpts)
}

// GetLastRoundParticipantIndex is a free data retrieval call binding the contract method 0x775694e1.
//
// Solidity: function getLastRoundParticipantIndex(address participant) view returns(uint256)
func (_ETHDKG *ETHDKGCaller) GetLastRoundParticipantIndex(opts *bind.CallOpts, participant common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getLastRoundParticipantIndex", participant)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLastRoundParticipantIndex is a free data retrieval call binding the contract method 0x775694e1.
//
// Solidity: function getLastRoundParticipantIndex(address participant) view returns(uint256)
func (_ETHDKG *ETHDKGSession) GetLastRoundParticipantIndex(participant common.Address) (*big.Int, error) {
	return _ETHDKG.Contract.GetLastRoundParticipantIndex(&_ETHDKG.CallOpts, participant)
}

// GetLastRoundParticipantIndex is a free data retrieval call binding the contract method 0x775694e1.
//
// Solidity: function getLastRoundParticipantIndex(address participant) view returns(uint256)
func (_ETHDKG *ETHDKGCallerSession) GetLastRoundParticipantIndex(participant common.Address) (*big.Int, error) {
	return _ETHDKG.Contract.GetLastRoundParticipantIndex(&_ETHDKG.CallOpts, participant)
}

// GetMasterPublicKey is a free data retrieval call binding the contract method 0xe146372a.
//
// Solidity: function getMasterPublicKey() view returns(uint256[4])
func (_ETHDKG *ETHDKGCaller) GetMasterPublicKey(opts *bind.CallOpts) ([4]*big.Int, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getMasterPublicKey")

	if err != nil {
		return *new([4]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([4]*big.Int)).(*[4]*big.Int)

	return out0, err

}

// GetMasterPublicKey is a free data retrieval call binding the contract method 0xe146372a.
//
// Solidity: function getMasterPublicKey() view returns(uint256[4])
func (_ETHDKG *ETHDKGSession) GetMasterPublicKey() ([4]*big.Int, error) {
	return _ETHDKG.Contract.GetMasterPublicKey(&_ETHDKG.CallOpts)
}

// GetMasterPublicKey is a free data retrieval call binding the contract method 0xe146372a.
//
// Solidity: function getMasterPublicKey() view returns(uint256[4])
func (_ETHDKG *ETHDKGCallerSession) GetMasterPublicKey() ([4]*big.Int, error) {
	return _ETHDKG.Contract.GetMasterPublicKey(&_ETHDKG.CallOpts)
}

// GetMasterPublicKeyHash is a free data retrieval call binding the contract method 0x1c67d576.
//
// Solidity: function getMasterPublicKeyHash() view returns(bytes32)
func (_ETHDKG *ETHDKGCaller) GetMasterPublicKeyHash(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getMasterPublicKeyHash")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetMasterPublicKeyHash is a free data retrieval call binding the contract method 0x1c67d576.
//
// Solidity: function getMasterPublicKeyHash() view returns(bytes32)
func (_ETHDKG *ETHDKGSession) GetMasterPublicKeyHash() ([32]byte, error) {
	return _ETHDKG.Contract.GetMasterPublicKeyHash(&_ETHDKG.CallOpts)
}

// GetMasterPublicKeyHash is a free data retrieval call binding the contract method 0x1c67d576.
//
// Solidity: function getMasterPublicKeyHash() view returns(bytes32)
func (_ETHDKG *ETHDKGCallerSession) GetMasterPublicKeyHash() ([32]byte, error) {
	return _ETHDKG.Contract.GetMasterPublicKeyHash(&_ETHDKG.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ETHDKG *ETHDKGCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ETHDKG *ETHDKGSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ETHDKG.Contract.GetMetamorphicContractAddress(&_ETHDKG.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ETHDKG *ETHDKGCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ETHDKG.Contract.GetMetamorphicContractAddress(&_ETHDKG.CallOpts, _salt, _factory)
}

// GetMinValidators is a free data retrieval call binding the contract method 0xecbadb36.
//
// Solidity: function getMinValidators() pure returns(uint256)
func (_ETHDKG *ETHDKGCaller) GetMinValidators(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getMinValidators")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMinValidators is a free data retrieval call binding the contract method 0xecbadb36.
//
// Solidity: function getMinValidators() pure returns(uint256)
func (_ETHDKG *ETHDKGSession) GetMinValidators() (*big.Int, error) {
	return _ETHDKG.Contract.GetMinValidators(&_ETHDKG.CallOpts)
}

// GetMinValidators is a free data retrieval call binding the contract method 0xecbadb36.
//
// Solidity: function getMinValidators() pure returns(uint256)
func (_ETHDKG *ETHDKGCallerSession) GetMinValidators() (*big.Int, error) {
	return _ETHDKG.Contract.GetMinValidators(&_ETHDKG.CallOpts)
}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_ETHDKG *ETHDKGCaller) GetNonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getNonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_ETHDKG *ETHDKGSession) GetNonce() (*big.Int, error) {
	return _ETHDKG.Contract.GetNonce(&_ETHDKG.CallOpts)
}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_ETHDKG *ETHDKGCallerSession) GetNonce() (*big.Int, error) {
	return _ETHDKG.Contract.GetNonce(&_ETHDKG.CallOpts)
}

// GetNumParticipants is a free data retrieval call binding the contract method 0xfd478ca9.
//
// Solidity: function getNumParticipants() view returns(uint256)
func (_ETHDKG *ETHDKGCaller) GetNumParticipants(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getNumParticipants")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNumParticipants is a free data retrieval call binding the contract method 0xfd478ca9.
//
// Solidity: function getNumParticipants() view returns(uint256)
func (_ETHDKG *ETHDKGSession) GetNumParticipants() (*big.Int, error) {
	return _ETHDKG.Contract.GetNumParticipants(&_ETHDKG.CallOpts)
}

// GetNumParticipants is a free data retrieval call binding the contract method 0xfd478ca9.
//
// Solidity: function getNumParticipants() view returns(uint256)
func (_ETHDKG *ETHDKGCallerSession) GetNumParticipants() (*big.Int, error) {
	return _ETHDKG.Contract.GetNumParticipants(&_ETHDKG.CallOpts)
}

// GetParticipantInternalState is a free data retrieval call binding the contract method 0xbf7786b6.
//
// Solidity: function getParticipantInternalState(address participant) view returns((uint256[2],uint64,uint64,uint8,bytes32,uint256[2],uint256[2],uint256[4]))
func (_ETHDKG *ETHDKGCaller) GetParticipantInternalState(opts *bind.CallOpts, participant common.Address) (Participant, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getParticipantInternalState", participant)

	if err != nil {
		return *new(Participant), err
	}

	out0 := *abi.ConvertType(out[0], new(Participant)).(*Participant)

	return out0, err

}

// GetParticipantInternalState is a free data retrieval call binding the contract method 0xbf7786b6.
//
// Solidity: function getParticipantInternalState(address participant) view returns((uint256[2],uint64,uint64,uint8,bytes32,uint256[2],uint256[2],uint256[4]))
func (_ETHDKG *ETHDKGSession) GetParticipantInternalState(participant common.Address) (Participant, error) {
	return _ETHDKG.Contract.GetParticipantInternalState(&_ETHDKG.CallOpts, participant)
}

// GetParticipantInternalState is a free data retrieval call binding the contract method 0xbf7786b6.
//
// Solidity: function getParticipantInternalState(address participant) view returns((uint256[2],uint64,uint64,uint8,bytes32,uint256[2],uint256[2],uint256[4]))
func (_ETHDKG *ETHDKGCallerSession) GetParticipantInternalState(participant common.Address) (Participant, error) {
	return _ETHDKG.Contract.GetParticipantInternalState(&_ETHDKG.CallOpts, participant)
}

// GetParticipantsInternalState is a free data retrieval call binding the contract method 0xc016baee.
//
// Solidity: function getParticipantsInternalState(address[] participantAddresses) view returns((uint256[2],uint64,uint64,uint8,bytes32,uint256[2],uint256[2],uint256[4])[])
func (_ETHDKG *ETHDKGCaller) GetParticipantsInternalState(opts *bind.CallOpts, participantAddresses []common.Address) ([]Participant, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getParticipantsInternalState", participantAddresses)

	if err != nil {
		return *new([]Participant), err
	}

	out0 := *abi.ConvertType(out[0], new([]Participant)).(*[]Participant)

	return out0, err

}

// GetParticipantsInternalState is a free data retrieval call binding the contract method 0xc016baee.
//
// Solidity: function getParticipantsInternalState(address[] participantAddresses) view returns((uint256[2],uint64,uint64,uint8,bytes32,uint256[2],uint256[2],uint256[4])[])
func (_ETHDKG *ETHDKGSession) GetParticipantsInternalState(participantAddresses []common.Address) ([]Participant, error) {
	return _ETHDKG.Contract.GetParticipantsInternalState(&_ETHDKG.CallOpts, participantAddresses)
}

// GetParticipantsInternalState is a free data retrieval call binding the contract method 0xc016baee.
//
// Solidity: function getParticipantsInternalState(address[] participantAddresses) view returns((uint256[2],uint64,uint64,uint8,bytes32,uint256[2],uint256[2],uint256[4])[])
func (_ETHDKG *ETHDKGCallerSession) GetParticipantsInternalState(participantAddresses []common.Address) ([]Participant, error) {
	return _ETHDKG.Contract.GetParticipantsInternalState(&_ETHDKG.CallOpts, participantAddresses)
}

// GetPhaseLength is a free data retrieval call binding the contract method 0x106da57d.
//
// Solidity: function getPhaseLength() view returns(uint256)
func (_ETHDKG *ETHDKGCaller) GetPhaseLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getPhaseLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPhaseLength is a free data retrieval call binding the contract method 0x106da57d.
//
// Solidity: function getPhaseLength() view returns(uint256)
func (_ETHDKG *ETHDKGSession) GetPhaseLength() (*big.Int, error) {
	return _ETHDKG.Contract.GetPhaseLength(&_ETHDKG.CallOpts)
}

// GetPhaseLength is a free data retrieval call binding the contract method 0x106da57d.
//
// Solidity: function getPhaseLength() view returns(uint256)
func (_ETHDKG *ETHDKGCallerSession) GetPhaseLength() (*big.Int, error) {
	return _ETHDKG.Contract.GetPhaseLength(&_ETHDKG.CallOpts)
}

// GetPhaseStartBlock is a free data retrieval call binding the contract method 0xa2bc9c78.
//
// Solidity: function getPhaseStartBlock() view returns(uint256)
func (_ETHDKG *ETHDKGCaller) GetPhaseStartBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "getPhaseStartBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPhaseStartBlock is a free data retrieval call binding the contract method 0xa2bc9c78.
//
// Solidity: function getPhaseStartBlock() view returns(uint256)
func (_ETHDKG *ETHDKGSession) GetPhaseStartBlock() (*big.Int, error) {
	return _ETHDKG.Contract.GetPhaseStartBlock(&_ETHDKG.CallOpts)
}

// GetPhaseStartBlock is a free data retrieval call binding the contract method 0xa2bc9c78.
//
// Solidity: function getPhaseStartBlock() view returns(uint256)
func (_ETHDKG *ETHDKGCallerSession) GetPhaseStartBlock() (*big.Int, error) {
	return _ETHDKG.Contract.GetPhaseStartBlock(&_ETHDKG.CallOpts)
}

// IsETHDKGCompleted is a free data retrieval call binding the contract method 0x2b7c6724.
//
// Solidity: function isETHDKGCompleted() view returns(bool)
func (_ETHDKG *ETHDKGCaller) IsETHDKGCompleted(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "isETHDKGCompleted")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsETHDKGCompleted is a free data retrieval call binding the contract method 0x2b7c6724.
//
// Solidity: function isETHDKGCompleted() view returns(bool)
func (_ETHDKG *ETHDKGSession) IsETHDKGCompleted() (bool, error) {
	return _ETHDKG.Contract.IsETHDKGCompleted(&_ETHDKG.CallOpts)
}

// IsETHDKGCompleted is a free data retrieval call binding the contract method 0x2b7c6724.
//
// Solidity: function isETHDKGCompleted() view returns(bool)
func (_ETHDKG *ETHDKGCallerSession) IsETHDKGCompleted() (bool, error) {
	return _ETHDKG.Contract.IsETHDKGCompleted(&_ETHDKG.CallOpts)
}

// IsETHDKGHalted is a free data retrieval call binding the contract method 0x43ced534.
//
// Solidity: function isETHDKGHalted() view returns(bool)
func (_ETHDKG *ETHDKGCaller) IsETHDKGHalted(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "isETHDKGHalted")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsETHDKGHalted is a free data retrieval call binding the contract method 0x43ced534.
//
// Solidity: function isETHDKGHalted() view returns(bool)
func (_ETHDKG *ETHDKGSession) IsETHDKGHalted() (bool, error) {
	return _ETHDKG.Contract.IsETHDKGHalted(&_ETHDKG.CallOpts)
}

// IsETHDKGHalted is a free data retrieval call binding the contract method 0x43ced534.
//
// Solidity: function isETHDKGHalted() view returns(bool)
func (_ETHDKG *ETHDKGCallerSession) IsETHDKGHalted() (bool, error) {
	return _ETHDKG.Contract.IsETHDKGHalted(&_ETHDKG.CallOpts)
}

// IsETHDKGRunning is a free data retrieval call binding the contract method 0x747b217c.
//
// Solidity: function isETHDKGRunning() view returns(bool)
func (_ETHDKG *ETHDKGCaller) IsETHDKGRunning(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "isETHDKGRunning")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsETHDKGRunning is a free data retrieval call binding the contract method 0x747b217c.
//
// Solidity: function isETHDKGRunning() view returns(bool)
func (_ETHDKG *ETHDKGSession) IsETHDKGRunning() (bool, error) {
	return _ETHDKG.Contract.IsETHDKGRunning(&_ETHDKG.CallOpts)
}

// IsETHDKGRunning is a free data retrieval call binding the contract method 0x747b217c.
//
// Solidity: function isETHDKGRunning() view returns(bool)
func (_ETHDKG *ETHDKGCallerSession) IsETHDKGRunning() (bool, error) {
	return _ETHDKG.Contract.IsETHDKGRunning(&_ETHDKG.CallOpts)
}

// IsMasterPublicKeySet is a free data retrieval call binding the contract method 0x08efcf16.
//
// Solidity: function isMasterPublicKeySet() view returns(bool)
func (_ETHDKG *ETHDKGCaller) IsMasterPublicKeySet(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ETHDKG.contract.Call(opts, &out, "isMasterPublicKeySet")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsMasterPublicKeySet is a free data retrieval call binding the contract method 0x08efcf16.
//
// Solidity: function isMasterPublicKeySet() view returns(bool)
func (_ETHDKG *ETHDKGSession) IsMasterPublicKeySet() (bool, error) {
	return _ETHDKG.Contract.IsMasterPublicKeySet(&_ETHDKG.CallOpts)
}

// IsMasterPublicKeySet is a free data retrieval call binding the contract method 0x08efcf16.
//
// Solidity: function isMasterPublicKeySet() view returns(bool)
func (_ETHDKG *ETHDKGCallerSession) IsMasterPublicKeySet() (bool, error) {
	return _ETHDKG.Contract.IsMasterPublicKeySet(&_ETHDKG.CallOpts)
}

// AccuseParticipantDidNotDistributeShares is a paid mutator transaction binding the contract method 0xdae681bc.
//
// Solidity: function accuseParticipantDidNotDistributeShares(address[] dishonestAddresses) returns()
func (_ETHDKG *ETHDKGTransactor) AccuseParticipantDidNotDistributeShares(opts *bind.TransactOpts, dishonestAddresses []common.Address) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "accuseParticipantDidNotDistributeShares", dishonestAddresses)
}

// AccuseParticipantDidNotDistributeShares is a paid mutator transaction binding the contract method 0xdae681bc.
//
// Solidity: function accuseParticipantDidNotDistributeShares(address[] dishonestAddresses) returns()
func (_ETHDKG *ETHDKGSession) AccuseParticipantDidNotDistributeShares(dishonestAddresses []common.Address) (*types.Transaction, error) {
	return _ETHDKG.Contract.AccuseParticipantDidNotDistributeShares(&_ETHDKG.TransactOpts, dishonestAddresses)
}

// AccuseParticipantDidNotDistributeShares is a paid mutator transaction binding the contract method 0xdae681bc.
//
// Solidity: function accuseParticipantDidNotDistributeShares(address[] dishonestAddresses) returns()
func (_ETHDKG *ETHDKGTransactorSession) AccuseParticipantDidNotDistributeShares(dishonestAddresses []common.Address) (*types.Transaction, error) {
	return _ETHDKG.Contract.AccuseParticipantDidNotDistributeShares(&_ETHDKG.TransactOpts, dishonestAddresses)
}

// AccuseParticipantDidNotSubmitGPKJ is a paid mutator transaction binding the contract method 0x7df24ee9.
//
// Solidity: function accuseParticipantDidNotSubmitGPKJ(address[] dishonestAddresses) returns()
func (_ETHDKG *ETHDKGTransactor) AccuseParticipantDidNotSubmitGPKJ(opts *bind.TransactOpts, dishonestAddresses []common.Address) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "accuseParticipantDidNotSubmitGPKJ", dishonestAddresses)
}

// AccuseParticipantDidNotSubmitGPKJ is a paid mutator transaction binding the contract method 0x7df24ee9.
//
// Solidity: function accuseParticipantDidNotSubmitGPKJ(address[] dishonestAddresses) returns()
func (_ETHDKG *ETHDKGSession) AccuseParticipantDidNotSubmitGPKJ(dishonestAddresses []common.Address) (*types.Transaction, error) {
	return _ETHDKG.Contract.AccuseParticipantDidNotSubmitGPKJ(&_ETHDKG.TransactOpts, dishonestAddresses)
}

// AccuseParticipantDidNotSubmitGPKJ is a paid mutator transaction binding the contract method 0x7df24ee9.
//
// Solidity: function accuseParticipantDidNotSubmitGPKJ(address[] dishonestAddresses) returns()
func (_ETHDKG *ETHDKGTransactorSession) AccuseParticipantDidNotSubmitGPKJ(dishonestAddresses []common.Address) (*types.Transaction, error) {
	return _ETHDKG.Contract.AccuseParticipantDidNotSubmitGPKJ(&_ETHDKG.TransactOpts, dishonestAddresses)
}

// AccuseParticipantDidNotSubmitKeyShares is a paid mutator transaction binding the contract method 0x043a6f12.
//
// Solidity: function accuseParticipantDidNotSubmitKeyShares(address[] dishonestAddresses) returns()
func (_ETHDKG *ETHDKGTransactor) AccuseParticipantDidNotSubmitKeyShares(opts *bind.TransactOpts, dishonestAddresses []common.Address) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "accuseParticipantDidNotSubmitKeyShares", dishonestAddresses)
}

// AccuseParticipantDidNotSubmitKeyShares is a paid mutator transaction binding the contract method 0x043a6f12.
//
// Solidity: function accuseParticipantDidNotSubmitKeyShares(address[] dishonestAddresses) returns()
func (_ETHDKG *ETHDKGSession) AccuseParticipantDidNotSubmitKeyShares(dishonestAddresses []common.Address) (*types.Transaction, error) {
	return _ETHDKG.Contract.AccuseParticipantDidNotSubmitKeyShares(&_ETHDKG.TransactOpts, dishonestAddresses)
}

// AccuseParticipantDidNotSubmitKeyShares is a paid mutator transaction binding the contract method 0x043a6f12.
//
// Solidity: function accuseParticipantDidNotSubmitKeyShares(address[] dishonestAddresses) returns()
func (_ETHDKG *ETHDKGTransactorSession) AccuseParticipantDidNotSubmitKeyShares(dishonestAddresses []common.Address) (*types.Transaction, error) {
	return _ETHDKG.Contract.AccuseParticipantDidNotSubmitKeyShares(&_ETHDKG.TransactOpts, dishonestAddresses)
}

// AccuseParticipantDistributedBadShares is a paid mutator transaction binding the contract method 0xedbe7bf7.
//
// Solidity: function accuseParticipantDistributedBadShares(address dishonestAddress, uint256[] encryptedShares, uint256[2][] commitments, uint256[2] sharedKey, uint256[2] sharedKeyCorrectnessProof) returns()
func (_ETHDKG *ETHDKGTransactor) AccuseParticipantDistributedBadShares(opts *bind.TransactOpts, dishonestAddress common.Address, encryptedShares []*big.Int, commitments [][2]*big.Int, sharedKey [2]*big.Int, sharedKeyCorrectnessProof [2]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "accuseParticipantDistributedBadShares", dishonestAddress, encryptedShares, commitments, sharedKey, sharedKeyCorrectnessProof)
}

// AccuseParticipantDistributedBadShares is a paid mutator transaction binding the contract method 0xedbe7bf7.
//
// Solidity: function accuseParticipantDistributedBadShares(address dishonestAddress, uint256[] encryptedShares, uint256[2][] commitments, uint256[2] sharedKey, uint256[2] sharedKeyCorrectnessProof) returns()
func (_ETHDKG *ETHDKGSession) AccuseParticipantDistributedBadShares(dishonestAddress common.Address, encryptedShares []*big.Int, commitments [][2]*big.Int, sharedKey [2]*big.Int, sharedKeyCorrectnessProof [2]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.AccuseParticipantDistributedBadShares(&_ETHDKG.TransactOpts, dishonestAddress, encryptedShares, commitments, sharedKey, sharedKeyCorrectnessProof)
}

// AccuseParticipantDistributedBadShares is a paid mutator transaction binding the contract method 0xedbe7bf7.
//
// Solidity: function accuseParticipantDistributedBadShares(address dishonestAddress, uint256[] encryptedShares, uint256[2][] commitments, uint256[2] sharedKey, uint256[2] sharedKeyCorrectnessProof) returns()
func (_ETHDKG *ETHDKGTransactorSession) AccuseParticipantDistributedBadShares(dishonestAddress common.Address, encryptedShares []*big.Int, commitments [][2]*big.Int, sharedKey [2]*big.Int, sharedKeyCorrectnessProof [2]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.AccuseParticipantDistributedBadShares(&_ETHDKG.TransactOpts, dishonestAddress, encryptedShares, commitments, sharedKey, sharedKeyCorrectnessProof)
}

// AccuseParticipantNotRegistered is a paid mutator transaction binding the contract method 0xf72c45b6.
//
// Solidity: function accuseParticipantNotRegistered(address[] dishonestAddresses) returns()
func (_ETHDKG *ETHDKGTransactor) AccuseParticipantNotRegistered(opts *bind.TransactOpts, dishonestAddresses []common.Address) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "accuseParticipantNotRegistered", dishonestAddresses)
}

// AccuseParticipantNotRegistered is a paid mutator transaction binding the contract method 0xf72c45b6.
//
// Solidity: function accuseParticipantNotRegistered(address[] dishonestAddresses) returns()
func (_ETHDKG *ETHDKGSession) AccuseParticipantNotRegistered(dishonestAddresses []common.Address) (*types.Transaction, error) {
	return _ETHDKG.Contract.AccuseParticipantNotRegistered(&_ETHDKG.TransactOpts, dishonestAddresses)
}

// AccuseParticipantNotRegistered is a paid mutator transaction binding the contract method 0xf72c45b6.
//
// Solidity: function accuseParticipantNotRegistered(address[] dishonestAddresses) returns()
func (_ETHDKG *ETHDKGTransactorSession) AccuseParticipantNotRegistered(dishonestAddresses []common.Address) (*types.Transaction, error) {
	return _ETHDKG.Contract.AccuseParticipantNotRegistered(&_ETHDKG.TransactOpts, dishonestAddresses)
}

// AccuseParticipantSubmittedBadGPKJ is a paid mutator transaction binding the contract method 0x80001264.
//
// Solidity: function accuseParticipantSubmittedBadGPKJ(address[] validators, bytes32[] encryptedSharesHash, uint256[2][][] commitments, address dishonestAddress) returns()
func (_ETHDKG *ETHDKGTransactor) AccuseParticipantSubmittedBadGPKJ(opts *bind.TransactOpts, validators []common.Address, encryptedSharesHash [][32]byte, commitments [][][2]*big.Int, dishonestAddress common.Address) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "accuseParticipantSubmittedBadGPKJ", validators, encryptedSharesHash, commitments, dishonestAddress)
}

// AccuseParticipantSubmittedBadGPKJ is a paid mutator transaction binding the contract method 0x80001264.
//
// Solidity: function accuseParticipantSubmittedBadGPKJ(address[] validators, bytes32[] encryptedSharesHash, uint256[2][][] commitments, address dishonestAddress) returns()
func (_ETHDKG *ETHDKGSession) AccuseParticipantSubmittedBadGPKJ(validators []common.Address, encryptedSharesHash [][32]byte, commitments [][][2]*big.Int, dishonestAddress common.Address) (*types.Transaction, error) {
	return _ETHDKG.Contract.AccuseParticipantSubmittedBadGPKJ(&_ETHDKG.TransactOpts, validators, encryptedSharesHash, commitments, dishonestAddress)
}

// AccuseParticipantSubmittedBadGPKJ is a paid mutator transaction binding the contract method 0x80001264.
//
// Solidity: function accuseParticipantSubmittedBadGPKJ(address[] validators, bytes32[] encryptedSharesHash, uint256[2][][] commitments, address dishonestAddress) returns()
func (_ETHDKG *ETHDKGTransactorSession) AccuseParticipantSubmittedBadGPKJ(validators []common.Address, encryptedSharesHash [][32]byte, commitments [][][2]*big.Int, dishonestAddress common.Address) (*types.Transaction, error) {
	return _ETHDKG.Contract.AccuseParticipantSubmittedBadGPKJ(&_ETHDKG.TransactOpts, validators, encryptedSharesHash, commitments, dishonestAddress)
}

// Complete is a paid mutator transaction binding the contract method 0x522e1177.
//
// Solidity: function complete() returns()
func (_ETHDKG *ETHDKGTransactor) Complete(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "complete")
}

// Complete is a paid mutator transaction binding the contract method 0x522e1177.
//
// Solidity: function complete() returns()
func (_ETHDKG *ETHDKGSession) Complete() (*types.Transaction, error) {
	return _ETHDKG.Contract.Complete(&_ETHDKG.TransactOpts)
}

// Complete is a paid mutator transaction binding the contract method 0x522e1177.
//
// Solidity: function complete() returns()
func (_ETHDKG *ETHDKGTransactorSession) Complete() (*types.Transaction, error) {
	return _ETHDKG.Contract.Complete(&_ETHDKG.TransactOpts)
}

// DistributeShares is a paid mutator transaction binding the contract method 0x80b97e01.
//
// Solidity: function distributeShares(uint256[] encryptedShares, uint256[2][] commitments) returns()
func (_ETHDKG *ETHDKGTransactor) DistributeShares(opts *bind.TransactOpts, encryptedShares []*big.Int, commitments [][2]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "distributeShares", encryptedShares, commitments)
}

// DistributeShares is a paid mutator transaction binding the contract method 0x80b97e01.
//
// Solidity: function distributeShares(uint256[] encryptedShares, uint256[2][] commitments) returns()
func (_ETHDKG *ETHDKGSession) DistributeShares(encryptedShares []*big.Int, commitments [][2]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.DistributeShares(&_ETHDKG.TransactOpts, encryptedShares, commitments)
}

// DistributeShares is a paid mutator transaction binding the contract method 0x80b97e01.
//
// Solidity: function distributeShares(uint256[] encryptedShares, uint256[2][] commitments) returns()
func (_ETHDKG *ETHDKGTransactorSession) DistributeShares(encryptedShares []*big.Int, commitments [][2]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.DistributeShares(&_ETHDKG.TransactOpts, encryptedShares, commitments)
}

// Initialize is a paid mutator transaction binding the contract method 0xe4a30116.
//
// Solidity: function initialize(uint256 phaseLength_, uint256 confirmationLength_) returns()
func (_ETHDKG *ETHDKGTransactor) Initialize(opts *bind.TransactOpts, phaseLength_ *big.Int, confirmationLength_ *big.Int) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "initialize", phaseLength_, confirmationLength_)
}

// Initialize is a paid mutator transaction binding the contract method 0xe4a30116.
//
// Solidity: function initialize(uint256 phaseLength_, uint256 confirmationLength_) returns()
func (_ETHDKG *ETHDKGSession) Initialize(phaseLength_ *big.Int, confirmationLength_ *big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.Initialize(&_ETHDKG.TransactOpts, phaseLength_, confirmationLength_)
}

// Initialize is a paid mutator transaction binding the contract method 0xe4a30116.
//
// Solidity: function initialize(uint256 phaseLength_, uint256 confirmationLength_) returns()
func (_ETHDKG *ETHDKGTransactorSession) Initialize(phaseLength_ *big.Int, confirmationLength_ *big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.Initialize(&_ETHDKG.TransactOpts, phaseLength_, confirmationLength_)
}

// InitializeETHDKG is a paid mutator transaction binding the contract method 0x57b51c9c.
//
// Solidity: function initializeETHDKG() returns()
func (_ETHDKG *ETHDKGTransactor) InitializeETHDKG(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "initializeETHDKG")
}

// InitializeETHDKG is a paid mutator transaction binding the contract method 0x57b51c9c.
//
// Solidity: function initializeETHDKG() returns()
func (_ETHDKG *ETHDKGSession) InitializeETHDKG() (*types.Transaction, error) {
	return _ETHDKG.Contract.InitializeETHDKG(&_ETHDKG.TransactOpts)
}

// InitializeETHDKG is a paid mutator transaction binding the contract method 0x57b51c9c.
//
// Solidity: function initializeETHDKG() returns()
func (_ETHDKG *ETHDKGTransactorSession) InitializeETHDKG() (*types.Transaction, error) {
	return _ETHDKG.Contract.InitializeETHDKG(&_ETHDKG.TransactOpts)
}

// MigrateValidators is a paid mutator transaction binding the contract method 0x4890465a.
//
// Solidity: function migrateValidators(address[] validatorsAccounts_, uint256[] validatorIndexes_, uint256[4][] validatorShares_, uint8 validatorCount_, uint256 epoch_, uint256 sideChainHeight_, uint256 ethHeight_, uint256[4] masterPublicKey_) returns()
func (_ETHDKG *ETHDKGTransactor) MigrateValidators(opts *bind.TransactOpts, validatorsAccounts_ []common.Address, validatorIndexes_ []*big.Int, validatorShares_ [][4]*big.Int, validatorCount_ uint8, epoch_ *big.Int, sideChainHeight_ *big.Int, ethHeight_ *big.Int, masterPublicKey_ [4]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "migrateValidators", validatorsAccounts_, validatorIndexes_, validatorShares_, validatorCount_, epoch_, sideChainHeight_, ethHeight_, masterPublicKey_)
}

// MigrateValidators is a paid mutator transaction binding the contract method 0x4890465a.
//
// Solidity: function migrateValidators(address[] validatorsAccounts_, uint256[] validatorIndexes_, uint256[4][] validatorShares_, uint8 validatorCount_, uint256 epoch_, uint256 sideChainHeight_, uint256 ethHeight_, uint256[4] masterPublicKey_) returns()
func (_ETHDKG *ETHDKGSession) MigrateValidators(validatorsAccounts_ []common.Address, validatorIndexes_ []*big.Int, validatorShares_ [][4]*big.Int, validatorCount_ uint8, epoch_ *big.Int, sideChainHeight_ *big.Int, ethHeight_ *big.Int, masterPublicKey_ [4]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.MigrateValidators(&_ETHDKG.TransactOpts, validatorsAccounts_, validatorIndexes_, validatorShares_, validatorCount_, epoch_, sideChainHeight_, ethHeight_, masterPublicKey_)
}

// MigrateValidators is a paid mutator transaction binding the contract method 0x4890465a.
//
// Solidity: function migrateValidators(address[] validatorsAccounts_, uint256[] validatorIndexes_, uint256[4][] validatorShares_, uint8 validatorCount_, uint256 epoch_, uint256 sideChainHeight_, uint256 ethHeight_, uint256[4] masterPublicKey_) returns()
func (_ETHDKG *ETHDKGTransactorSession) MigrateValidators(validatorsAccounts_ []common.Address, validatorIndexes_ []*big.Int, validatorShares_ [][4]*big.Int, validatorCount_ uint8, epoch_ *big.Int, sideChainHeight_ *big.Int, ethHeight_ *big.Int, masterPublicKey_ [4]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.MigrateValidators(&_ETHDKG.TransactOpts, validatorsAccounts_, validatorIndexes_, validatorShares_, validatorCount_, epoch_, sideChainHeight_, ethHeight_, masterPublicKey_)
}

// Register is a paid mutator transaction binding the contract method 0x3442af5c.
//
// Solidity: function register(uint256[2] publicKey) returns()
func (_ETHDKG *ETHDKGTransactor) Register(opts *bind.TransactOpts, publicKey [2]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "register", publicKey)
}

// Register is a paid mutator transaction binding the contract method 0x3442af5c.
//
// Solidity: function register(uint256[2] publicKey) returns()
func (_ETHDKG *ETHDKGSession) Register(publicKey [2]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.Register(&_ETHDKG.TransactOpts, publicKey)
}

// Register is a paid mutator transaction binding the contract method 0x3442af5c.
//
// Solidity: function register(uint256[2] publicKey) returns()
func (_ETHDKG *ETHDKGTransactorSession) Register(publicKey [2]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.Register(&_ETHDKG.TransactOpts, publicKey)
}

// SetConfirmationLength is a paid mutator transaction binding the contract method 0xff3e5e45.
//
// Solidity: function setConfirmationLength(uint16 confirmationLength_) returns()
func (_ETHDKG *ETHDKGTransactor) SetConfirmationLength(opts *bind.TransactOpts, confirmationLength_ uint16) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "setConfirmationLength", confirmationLength_)
}

// SetConfirmationLength is a paid mutator transaction binding the contract method 0xff3e5e45.
//
// Solidity: function setConfirmationLength(uint16 confirmationLength_) returns()
func (_ETHDKG *ETHDKGSession) SetConfirmationLength(confirmationLength_ uint16) (*types.Transaction, error) {
	return _ETHDKG.Contract.SetConfirmationLength(&_ETHDKG.TransactOpts, confirmationLength_)
}

// SetConfirmationLength is a paid mutator transaction binding the contract method 0xff3e5e45.
//
// Solidity: function setConfirmationLength(uint16 confirmationLength_) returns()
func (_ETHDKG *ETHDKGTransactorSession) SetConfirmationLength(confirmationLength_ uint16) (*types.Transaction, error) {
	return _ETHDKG.Contract.SetConfirmationLength(&_ETHDKG.TransactOpts, confirmationLength_)
}

// SetCustomAliceNetHeight is a paid mutator transaction binding the contract method 0xdf8d157b.
//
// Solidity: function setCustomAliceNetHeight(uint256 aliceNetHeight) returns()
func (_ETHDKG *ETHDKGTransactor) SetCustomAliceNetHeight(opts *bind.TransactOpts, aliceNetHeight *big.Int) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "setCustomAliceNetHeight", aliceNetHeight)
}

// SetCustomAliceNetHeight is a paid mutator transaction binding the contract method 0xdf8d157b.
//
// Solidity: function setCustomAliceNetHeight(uint256 aliceNetHeight) returns()
func (_ETHDKG *ETHDKGSession) SetCustomAliceNetHeight(aliceNetHeight *big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.SetCustomAliceNetHeight(&_ETHDKG.TransactOpts, aliceNetHeight)
}

// SetCustomAliceNetHeight is a paid mutator transaction binding the contract method 0xdf8d157b.
//
// Solidity: function setCustomAliceNetHeight(uint256 aliceNetHeight) returns()
func (_ETHDKG *ETHDKGTransactorSession) SetCustomAliceNetHeight(aliceNetHeight *big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.SetCustomAliceNetHeight(&_ETHDKG.TransactOpts, aliceNetHeight)
}

// SetPhaseLength is a paid mutator transaction binding the contract method 0x8a3c24cc.
//
// Solidity: function setPhaseLength(uint16 phaseLength_) returns()
func (_ETHDKG *ETHDKGTransactor) SetPhaseLength(opts *bind.TransactOpts, phaseLength_ uint16) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "setPhaseLength", phaseLength_)
}

// SetPhaseLength is a paid mutator transaction binding the contract method 0x8a3c24cc.
//
// Solidity: function setPhaseLength(uint16 phaseLength_) returns()
func (_ETHDKG *ETHDKGSession) SetPhaseLength(phaseLength_ uint16) (*types.Transaction, error) {
	return _ETHDKG.Contract.SetPhaseLength(&_ETHDKG.TransactOpts, phaseLength_)
}

// SetPhaseLength is a paid mutator transaction binding the contract method 0x8a3c24cc.
//
// Solidity: function setPhaseLength(uint16 phaseLength_) returns()
func (_ETHDKG *ETHDKGTransactorSession) SetPhaseLength(phaseLength_ uint16) (*types.Transaction, error) {
	return _ETHDKG.Contract.SetPhaseLength(&_ETHDKG.TransactOpts, phaseLength_)
}

// SubmitGPKJ is a paid mutator transaction binding the contract method 0x101f49c1.
//
// Solidity: function submitGPKJ(uint256[4] gpkj) returns()
func (_ETHDKG *ETHDKGTransactor) SubmitGPKJ(opts *bind.TransactOpts, gpkj [4]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "submitGPKJ", gpkj)
}

// SubmitGPKJ is a paid mutator transaction binding the contract method 0x101f49c1.
//
// Solidity: function submitGPKJ(uint256[4] gpkj) returns()
func (_ETHDKG *ETHDKGSession) SubmitGPKJ(gpkj [4]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.SubmitGPKJ(&_ETHDKG.TransactOpts, gpkj)
}

// SubmitGPKJ is a paid mutator transaction binding the contract method 0x101f49c1.
//
// Solidity: function submitGPKJ(uint256[4] gpkj) returns()
func (_ETHDKG *ETHDKGTransactorSession) SubmitGPKJ(gpkj [4]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.SubmitGPKJ(&_ETHDKG.TransactOpts, gpkj)
}

// SubmitKeyShare is a paid mutator transaction binding the contract method 0x62a6523e.
//
// Solidity: function submitKeyShare(uint256[2] keyShareG1, uint256[2] keyShareG1CorrectnessProof, uint256[4] keyShareG2) returns()
func (_ETHDKG *ETHDKGTransactor) SubmitKeyShare(opts *bind.TransactOpts, keyShareG1 [2]*big.Int, keyShareG1CorrectnessProof [2]*big.Int, keyShareG2 [4]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "submitKeyShare", keyShareG1, keyShareG1CorrectnessProof, keyShareG2)
}

// SubmitKeyShare is a paid mutator transaction binding the contract method 0x62a6523e.
//
// Solidity: function submitKeyShare(uint256[2] keyShareG1, uint256[2] keyShareG1CorrectnessProof, uint256[4] keyShareG2) returns()
func (_ETHDKG *ETHDKGSession) SubmitKeyShare(keyShareG1 [2]*big.Int, keyShareG1CorrectnessProof [2]*big.Int, keyShareG2 [4]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.SubmitKeyShare(&_ETHDKG.TransactOpts, keyShareG1, keyShareG1CorrectnessProof, keyShareG2)
}

// SubmitKeyShare is a paid mutator transaction binding the contract method 0x62a6523e.
//
// Solidity: function submitKeyShare(uint256[2] keyShareG1, uint256[2] keyShareG1CorrectnessProof, uint256[4] keyShareG2) returns()
func (_ETHDKG *ETHDKGTransactorSession) SubmitKeyShare(keyShareG1 [2]*big.Int, keyShareG1CorrectnessProof [2]*big.Int, keyShareG2 [4]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.SubmitKeyShare(&_ETHDKG.TransactOpts, keyShareG1, keyShareG1CorrectnessProof, keyShareG2)
}

// SubmitMasterPublicKey is a paid mutator transaction binding the contract method 0xe8323224.
//
// Solidity: function submitMasterPublicKey(uint256[4] masterPublicKey_) returns()
func (_ETHDKG *ETHDKGTransactor) SubmitMasterPublicKey(opts *bind.TransactOpts, masterPublicKey_ [4]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.contract.Transact(opts, "submitMasterPublicKey", masterPublicKey_)
}

// SubmitMasterPublicKey is a paid mutator transaction binding the contract method 0xe8323224.
//
// Solidity: function submitMasterPublicKey(uint256[4] masterPublicKey_) returns()
func (_ETHDKG *ETHDKGSession) SubmitMasterPublicKey(masterPublicKey_ [4]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.SubmitMasterPublicKey(&_ETHDKG.TransactOpts, masterPublicKey_)
}

// SubmitMasterPublicKey is a paid mutator transaction binding the contract method 0xe8323224.
//
// Solidity: function submitMasterPublicKey(uint256[4] masterPublicKey_) returns()
func (_ETHDKG *ETHDKGTransactorSession) SubmitMasterPublicKey(masterPublicKey_ [4]*big.Int) (*types.Transaction, error) {
	return _ETHDKG.Contract.SubmitMasterPublicKey(&_ETHDKG.TransactOpts, masterPublicKey_)
}

// ETHDKGAddressRegisteredIterator is returned from FilterAddressRegistered and is used to iterate over the raw logs and unpacked data for AddressRegistered events raised by the ETHDKG contract.
type ETHDKGAddressRegisteredIterator struct {
	Event *ETHDKGAddressRegistered // Event containing the contract specifics and raw log

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
func (it *ETHDKGAddressRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHDKGAddressRegistered)
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
		it.Event = new(ETHDKGAddressRegistered)
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
func (it *ETHDKGAddressRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHDKGAddressRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHDKGAddressRegistered represents a AddressRegistered event raised by the ETHDKG contract.
type ETHDKGAddressRegistered struct {
	Account   common.Address
	Index     *big.Int
	Nonce     *big.Int
	PublicKey [2]*big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterAddressRegistered is a free log retrieval operation binding the contract event 0x7f1304057ec61140fbf2f5f236790f34fcafe123d3eb0d298d92317c97da500d.
//
// Solidity: event AddressRegistered(address account, uint256 index, uint256 nonce, uint256[2] publicKey)
func (_ETHDKG *ETHDKGFilterer) FilterAddressRegistered(opts *bind.FilterOpts) (*ETHDKGAddressRegisteredIterator, error) {

	logs, sub, err := _ETHDKG.contract.FilterLogs(opts, "AddressRegistered")
	if err != nil {
		return nil, err
	}
	return &ETHDKGAddressRegisteredIterator{contract: _ETHDKG.contract, event: "AddressRegistered", logs: logs, sub: sub}, nil
}

// WatchAddressRegistered is a free log subscription operation binding the contract event 0x7f1304057ec61140fbf2f5f236790f34fcafe123d3eb0d298d92317c97da500d.
//
// Solidity: event AddressRegistered(address account, uint256 index, uint256 nonce, uint256[2] publicKey)
func (_ETHDKG *ETHDKGFilterer) WatchAddressRegistered(opts *bind.WatchOpts, sink chan<- *ETHDKGAddressRegistered) (event.Subscription, error) {

	logs, sub, err := _ETHDKG.contract.WatchLogs(opts, "AddressRegistered")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHDKGAddressRegistered)
				if err := _ETHDKG.contract.UnpackLog(event, "AddressRegistered", log); err != nil {
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

// ParseAddressRegistered is a log parse operation binding the contract event 0x7f1304057ec61140fbf2f5f236790f34fcafe123d3eb0d298d92317c97da500d.
//
// Solidity: event AddressRegistered(address account, uint256 index, uint256 nonce, uint256[2] publicKey)
func (_ETHDKG *ETHDKGFilterer) ParseAddressRegistered(log types.Log) (*ETHDKGAddressRegistered, error) {
	event := new(ETHDKGAddressRegistered)
	if err := _ETHDKG.contract.UnpackLog(event, "AddressRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ETHDKGGPKJSubmissionCompleteIterator is returned from FilterGPKJSubmissionComplete and is used to iterate over the raw logs and unpacked data for GPKJSubmissionComplete events raised by the ETHDKG contract.
type ETHDKGGPKJSubmissionCompleteIterator struct {
	Event *ETHDKGGPKJSubmissionComplete // Event containing the contract specifics and raw log

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
func (it *ETHDKGGPKJSubmissionCompleteIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHDKGGPKJSubmissionComplete)
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
		it.Event = new(ETHDKGGPKJSubmissionComplete)
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
func (it *ETHDKGGPKJSubmissionCompleteIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHDKGGPKJSubmissionCompleteIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHDKGGPKJSubmissionComplete represents a GPKJSubmissionComplete event raised by the ETHDKG contract.
type ETHDKGGPKJSubmissionComplete struct {
	BlockNumber *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterGPKJSubmissionComplete is a free log retrieval operation binding the contract event 0x87bfe600b78cad9f7cf68c99eb582c1748f636b3269842b37d5873b0e069f628.
//
// Solidity: event GPKJSubmissionComplete(uint256 blockNumber)
func (_ETHDKG *ETHDKGFilterer) FilterGPKJSubmissionComplete(opts *bind.FilterOpts) (*ETHDKGGPKJSubmissionCompleteIterator, error) {

	logs, sub, err := _ETHDKG.contract.FilterLogs(opts, "GPKJSubmissionComplete")
	if err != nil {
		return nil, err
	}
	return &ETHDKGGPKJSubmissionCompleteIterator{contract: _ETHDKG.contract, event: "GPKJSubmissionComplete", logs: logs, sub: sub}, nil
}

// WatchGPKJSubmissionComplete is a free log subscription operation binding the contract event 0x87bfe600b78cad9f7cf68c99eb582c1748f636b3269842b37d5873b0e069f628.
//
// Solidity: event GPKJSubmissionComplete(uint256 blockNumber)
func (_ETHDKG *ETHDKGFilterer) WatchGPKJSubmissionComplete(opts *bind.WatchOpts, sink chan<- *ETHDKGGPKJSubmissionComplete) (event.Subscription, error) {

	logs, sub, err := _ETHDKG.contract.WatchLogs(opts, "GPKJSubmissionComplete")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHDKGGPKJSubmissionComplete)
				if err := _ETHDKG.contract.UnpackLog(event, "GPKJSubmissionComplete", log); err != nil {
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

// ParseGPKJSubmissionComplete is a log parse operation binding the contract event 0x87bfe600b78cad9f7cf68c99eb582c1748f636b3269842b37d5873b0e069f628.
//
// Solidity: event GPKJSubmissionComplete(uint256 blockNumber)
func (_ETHDKG *ETHDKGFilterer) ParseGPKJSubmissionComplete(log types.Log) (*ETHDKGGPKJSubmissionComplete, error) {
	event := new(ETHDKGGPKJSubmissionComplete)
	if err := _ETHDKG.contract.UnpackLog(event, "GPKJSubmissionComplete", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ETHDKGInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the ETHDKG contract.
type ETHDKGInitializedIterator struct {
	Event *ETHDKGInitialized // Event containing the contract specifics and raw log

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
func (it *ETHDKGInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHDKGInitialized)
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
		it.Event = new(ETHDKGInitialized)
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
func (it *ETHDKGInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHDKGInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHDKGInitialized represents a Initialized event raised by the ETHDKG contract.
type ETHDKGInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ETHDKG *ETHDKGFilterer) FilterInitialized(opts *bind.FilterOpts) (*ETHDKGInitializedIterator, error) {

	logs, sub, err := _ETHDKG.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ETHDKGInitializedIterator{contract: _ETHDKG.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ETHDKG *ETHDKGFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ETHDKGInitialized) (event.Subscription, error) {

	logs, sub, err := _ETHDKG.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHDKGInitialized)
				if err := _ETHDKG.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_ETHDKG *ETHDKGFilterer) ParseInitialized(log types.Log) (*ETHDKGInitialized, error) {
	event := new(ETHDKGInitialized)
	if err := _ETHDKG.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ETHDKGKeyShareSubmissionCompleteIterator is returned from FilterKeyShareSubmissionComplete and is used to iterate over the raw logs and unpacked data for KeyShareSubmissionComplete events raised by the ETHDKG contract.
type ETHDKGKeyShareSubmissionCompleteIterator struct {
	Event *ETHDKGKeyShareSubmissionComplete // Event containing the contract specifics and raw log

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
func (it *ETHDKGKeyShareSubmissionCompleteIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHDKGKeyShareSubmissionComplete)
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
		it.Event = new(ETHDKGKeyShareSubmissionComplete)
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
func (it *ETHDKGKeyShareSubmissionCompleteIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHDKGKeyShareSubmissionCompleteIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHDKGKeyShareSubmissionComplete represents a KeyShareSubmissionComplete event raised by the ETHDKG contract.
type ETHDKGKeyShareSubmissionComplete struct {
	BlockNumber *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterKeyShareSubmissionComplete is a free log retrieval operation binding the contract event 0x522cec98f6caa194456c44afa9e8cef9ac63eecb0be60e20d180ce19cfb0ef59.
//
// Solidity: event KeyShareSubmissionComplete(uint256 blockNumber)
func (_ETHDKG *ETHDKGFilterer) FilterKeyShareSubmissionComplete(opts *bind.FilterOpts) (*ETHDKGKeyShareSubmissionCompleteIterator, error) {

	logs, sub, err := _ETHDKG.contract.FilterLogs(opts, "KeyShareSubmissionComplete")
	if err != nil {
		return nil, err
	}
	return &ETHDKGKeyShareSubmissionCompleteIterator{contract: _ETHDKG.contract, event: "KeyShareSubmissionComplete", logs: logs, sub: sub}, nil
}

// WatchKeyShareSubmissionComplete is a free log subscription operation binding the contract event 0x522cec98f6caa194456c44afa9e8cef9ac63eecb0be60e20d180ce19cfb0ef59.
//
// Solidity: event KeyShareSubmissionComplete(uint256 blockNumber)
func (_ETHDKG *ETHDKGFilterer) WatchKeyShareSubmissionComplete(opts *bind.WatchOpts, sink chan<- *ETHDKGKeyShareSubmissionComplete) (event.Subscription, error) {

	logs, sub, err := _ETHDKG.contract.WatchLogs(opts, "KeyShareSubmissionComplete")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHDKGKeyShareSubmissionComplete)
				if err := _ETHDKG.contract.UnpackLog(event, "KeyShareSubmissionComplete", log); err != nil {
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

// ParseKeyShareSubmissionComplete is a log parse operation binding the contract event 0x522cec98f6caa194456c44afa9e8cef9ac63eecb0be60e20d180ce19cfb0ef59.
//
// Solidity: event KeyShareSubmissionComplete(uint256 blockNumber)
func (_ETHDKG *ETHDKGFilterer) ParseKeyShareSubmissionComplete(log types.Log) (*ETHDKGKeyShareSubmissionComplete, error) {
	event := new(ETHDKGKeyShareSubmissionComplete)
	if err := _ETHDKG.contract.UnpackLog(event, "KeyShareSubmissionComplete", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ETHDKGKeyShareSubmittedIterator is returned from FilterKeyShareSubmitted and is used to iterate over the raw logs and unpacked data for KeyShareSubmitted events raised by the ETHDKG contract.
type ETHDKGKeyShareSubmittedIterator struct {
	Event *ETHDKGKeyShareSubmitted // Event containing the contract specifics and raw log

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
func (it *ETHDKGKeyShareSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHDKGKeyShareSubmitted)
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
		it.Event = new(ETHDKGKeyShareSubmitted)
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
func (it *ETHDKGKeyShareSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHDKGKeyShareSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHDKGKeyShareSubmitted represents a KeyShareSubmitted event raised by the ETHDKG contract.
type ETHDKGKeyShareSubmitted struct {
	Account                    common.Address
	Index                      *big.Int
	Nonce                      *big.Int
	KeyShareG1                 [2]*big.Int
	KeyShareG1CorrectnessProof [2]*big.Int
	KeyShareG2                 [4]*big.Int
	Raw                        types.Log // Blockchain specific contextual infos
}

// FilterKeyShareSubmitted is a free log retrieval operation binding the contract event 0x6162e2d11398e4063e4c8565dafc4fb6755bbead93747ea836a5ef73a594aaf7.
//
// Solidity: event KeyShareSubmitted(address account, uint256 index, uint256 nonce, uint256[2] keyShareG1, uint256[2] keyShareG1CorrectnessProof, uint256[4] keyShareG2)
func (_ETHDKG *ETHDKGFilterer) FilterKeyShareSubmitted(opts *bind.FilterOpts) (*ETHDKGKeyShareSubmittedIterator, error) {

	logs, sub, err := _ETHDKG.contract.FilterLogs(opts, "KeyShareSubmitted")
	if err != nil {
		return nil, err
	}
	return &ETHDKGKeyShareSubmittedIterator{contract: _ETHDKG.contract, event: "KeyShareSubmitted", logs: logs, sub: sub}, nil
}

// WatchKeyShareSubmitted is a free log subscription operation binding the contract event 0x6162e2d11398e4063e4c8565dafc4fb6755bbead93747ea836a5ef73a594aaf7.
//
// Solidity: event KeyShareSubmitted(address account, uint256 index, uint256 nonce, uint256[2] keyShareG1, uint256[2] keyShareG1CorrectnessProof, uint256[4] keyShareG2)
func (_ETHDKG *ETHDKGFilterer) WatchKeyShareSubmitted(opts *bind.WatchOpts, sink chan<- *ETHDKGKeyShareSubmitted) (event.Subscription, error) {

	logs, sub, err := _ETHDKG.contract.WatchLogs(opts, "KeyShareSubmitted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHDKGKeyShareSubmitted)
				if err := _ETHDKG.contract.UnpackLog(event, "KeyShareSubmitted", log); err != nil {
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

// ParseKeyShareSubmitted is a log parse operation binding the contract event 0x6162e2d11398e4063e4c8565dafc4fb6755bbead93747ea836a5ef73a594aaf7.
//
// Solidity: event KeyShareSubmitted(address account, uint256 index, uint256 nonce, uint256[2] keyShareG1, uint256[2] keyShareG1CorrectnessProof, uint256[4] keyShareG2)
func (_ETHDKG *ETHDKGFilterer) ParseKeyShareSubmitted(log types.Log) (*ETHDKGKeyShareSubmitted, error) {
	event := new(ETHDKGKeyShareSubmitted)
	if err := _ETHDKG.contract.UnpackLog(event, "KeyShareSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ETHDKGMPKSetIterator is returned from FilterMPKSet and is used to iterate over the raw logs and unpacked data for MPKSet events raised by the ETHDKG contract.
type ETHDKGMPKSetIterator struct {
	Event *ETHDKGMPKSet // Event containing the contract specifics and raw log

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
func (it *ETHDKGMPKSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHDKGMPKSet)
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
		it.Event = new(ETHDKGMPKSet)
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
func (it *ETHDKGMPKSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHDKGMPKSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHDKGMPKSet represents a MPKSet event raised by the ETHDKG contract.
type ETHDKGMPKSet struct {
	BlockNumber *big.Int
	Nonce       *big.Int
	Mpk         [4]*big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterMPKSet is a free log retrieval operation binding the contract event 0x71b1ebd27be320895a22125d6458e3363aefa6944a312ede4bf275867e6d5a71.
//
// Solidity: event MPKSet(uint256 blockNumber, uint256 nonce, uint256[4] mpk)
func (_ETHDKG *ETHDKGFilterer) FilterMPKSet(opts *bind.FilterOpts) (*ETHDKGMPKSetIterator, error) {

	logs, sub, err := _ETHDKG.contract.FilterLogs(opts, "MPKSet")
	if err != nil {
		return nil, err
	}
	return &ETHDKGMPKSetIterator{contract: _ETHDKG.contract, event: "MPKSet", logs: logs, sub: sub}, nil
}

// WatchMPKSet is a free log subscription operation binding the contract event 0x71b1ebd27be320895a22125d6458e3363aefa6944a312ede4bf275867e6d5a71.
//
// Solidity: event MPKSet(uint256 blockNumber, uint256 nonce, uint256[4] mpk)
func (_ETHDKG *ETHDKGFilterer) WatchMPKSet(opts *bind.WatchOpts, sink chan<- *ETHDKGMPKSet) (event.Subscription, error) {

	logs, sub, err := _ETHDKG.contract.WatchLogs(opts, "MPKSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHDKGMPKSet)
				if err := _ETHDKG.contract.UnpackLog(event, "MPKSet", log); err != nil {
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

// ParseMPKSet is a log parse operation binding the contract event 0x71b1ebd27be320895a22125d6458e3363aefa6944a312ede4bf275867e6d5a71.
//
// Solidity: event MPKSet(uint256 blockNumber, uint256 nonce, uint256[4] mpk)
func (_ETHDKG *ETHDKGFilterer) ParseMPKSet(log types.Log) (*ETHDKGMPKSet, error) {
	event := new(ETHDKGMPKSet)
	if err := _ETHDKG.contract.UnpackLog(event, "MPKSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ETHDKGRegistrationCompleteIterator is returned from FilterRegistrationComplete and is used to iterate over the raw logs and unpacked data for RegistrationComplete events raised by the ETHDKG contract.
type ETHDKGRegistrationCompleteIterator struct {
	Event *ETHDKGRegistrationComplete // Event containing the contract specifics and raw log

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
func (it *ETHDKGRegistrationCompleteIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHDKGRegistrationComplete)
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
		it.Event = new(ETHDKGRegistrationComplete)
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
func (it *ETHDKGRegistrationCompleteIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHDKGRegistrationCompleteIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHDKGRegistrationComplete represents a RegistrationComplete event raised by the ETHDKG contract.
type ETHDKGRegistrationComplete struct {
	BlockNumber *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterRegistrationComplete is a free log retrieval operation binding the contract event 0x833013b96b786b4eca83baac286920e5e53956c21ff3894f1d9f02e97d6ed764.
//
// Solidity: event RegistrationComplete(uint256 blockNumber)
func (_ETHDKG *ETHDKGFilterer) FilterRegistrationComplete(opts *bind.FilterOpts) (*ETHDKGRegistrationCompleteIterator, error) {

	logs, sub, err := _ETHDKG.contract.FilterLogs(opts, "RegistrationComplete")
	if err != nil {
		return nil, err
	}
	return &ETHDKGRegistrationCompleteIterator{contract: _ETHDKG.contract, event: "RegistrationComplete", logs: logs, sub: sub}, nil
}

// WatchRegistrationComplete is a free log subscription operation binding the contract event 0x833013b96b786b4eca83baac286920e5e53956c21ff3894f1d9f02e97d6ed764.
//
// Solidity: event RegistrationComplete(uint256 blockNumber)
func (_ETHDKG *ETHDKGFilterer) WatchRegistrationComplete(opts *bind.WatchOpts, sink chan<- *ETHDKGRegistrationComplete) (event.Subscription, error) {

	logs, sub, err := _ETHDKG.contract.WatchLogs(opts, "RegistrationComplete")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHDKGRegistrationComplete)
				if err := _ETHDKG.contract.UnpackLog(event, "RegistrationComplete", log); err != nil {
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

// ParseRegistrationComplete is a log parse operation binding the contract event 0x833013b96b786b4eca83baac286920e5e53956c21ff3894f1d9f02e97d6ed764.
//
// Solidity: event RegistrationComplete(uint256 blockNumber)
func (_ETHDKG *ETHDKGFilterer) ParseRegistrationComplete(log types.Log) (*ETHDKGRegistrationComplete, error) {
	event := new(ETHDKGRegistrationComplete)
	if err := _ETHDKG.contract.UnpackLog(event, "RegistrationComplete", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ETHDKGRegistrationOpenedIterator is returned from FilterRegistrationOpened and is used to iterate over the raw logs and unpacked data for RegistrationOpened events raised by the ETHDKG contract.
type ETHDKGRegistrationOpenedIterator struct {
	Event *ETHDKGRegistrationOpened // Event containing the contract specifics and raw log

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
func (it *ETHDKGRegistrationOpenedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHDKGRegistrationOpened)
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
		it.Event = new(ETHDKGRegistrationOpened)
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
func (it *ETHDKGRegistrationOpenedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHDKGRegistrationOpenedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHDKGRegistrationOpened represents a RegistrationOpened event raised by the ETHDKG contract.
type ETHDKGRegistrationOpened struct {
	StartBlock         *big.Int
	NumberValidators   *big.Int
	Nonce              *big.Int
	PhaseLength        *big.Int
	ConfirmationLength *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterRegistrationOpened is a free log retrieval operation binding the contract event 0xbda431b9b63510f1398bf33d700e013315bcba905507078a1780f13ea5b354b9.
//
// Solidity: event RegistrationOpened(uint256 startBlock, uint256 numberValidators, uint256 nonce, uint256 phaseLength, uint256 confirmationLength)
func (_ETHDKG *ETHDKGFilterer) FilterRegistrationOpened(opts *bind.FilterOpts) (*ETHDKGRegistrationOpenedIterator, error) {

	logs, sub, err := _ETHDKG.contract.FilterLogs(opts, "RegistrationOpened")
	if err != nil {
		return nil, err
	}
	return &ETHDKGRegistrationOpenedIterator{contract: _ETHDKG.contract, event: "RegistrationOpened", logs: logs, sub: sub}, nil
}

// WatchRegistrationOpened is a free log subscription operation binding the contract event 0xbda431b9b63510f1398bf33d700e013315bcba905507078a1780f13ea5b354b9.
//
// Solidity: event RegistrationOpened(uint256 startBlock, uint256 numberValidators, uint256 nonce, uint256 phaseLength, uint256 confirmationLength)
func (_ETHDKG *ETHDKGFilterer) WatchRegistrationOpened(opts *bind.WatchOpts, sink chan<- *ETHDKGRegistrationOpened) (event.Subscription, error) {

	logs, sub, err := _ETHDKG.contract.WatchLogs(opts, "RegistrationOpened")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHDKGRegistrationOpened)
				if err := _ETHDKG.contract.UnpackLog(event, "RegistrationOpened", log); err != nil {
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

// ParseRegistrationOpened is a log parse operation binding the contract event 0xbda431b9b63510f1398bf33d700e013315bcba905507078a1780f13ea5b354b9.
//
// Solidity: event RegistrationOpened(uint256 startBlock, uint256 numberValidators, uint256 nonce, uint256 phaseLength, uint256 confirmationLength)
func (_ETHDKG *ETHDKGFilterer) ParseRegistrationOpened(log types.Log) (*ETHDKGRegistrationOpened, error) {
	event := new(ETHDKGRegistrationOpened)
	if err := _ETHDKG.contract.UnpackLog(event, "RegistrationOpened", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ETHDKGShareDistributionCompleteIterator is returned from FilterShareDistributionComplete and is used to iterate over the raw logs and unpacked data for ShareDistributionComplete events raised by the ETHDKG contract.
type ETHDKGShareDistributionCompleteIterator struct {
	Event *ETHDKGShareDistributionComplete // Event containing the contract specifics and raw log

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
func (it *ETHDKGShareDistributionCompleteIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHDKGShareDistributionComplete)
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
		it.Event = new(ETHDKGShareDistributionComplete)
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
func (it *ETHDKGShareDistributionCompleteIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHDKGShareDistributionCompleteIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHDKGShareDistributionComplete represents a ShareDistributionComplete event raised by the ETHDKG contract.
type ETHDKGShareDistributionComplete struct {
	BlockNumber *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterShareDistributionComplete is a free log retrieval operation binding the contract event 0xbfe94ffef5ddde4d25ac7b652f3f67686ea63f9badbfe1f25451e26fc262d11c.
//
// Solidity: event ShareDistributionComplete(uint256 blockNumber)
func (_ETHDKG *ETHDKGFilterer) FilterShareDistributionComplete(opts *bind.FilterOpts) (*ETHDKGShareDistributionCompleteIterator, error) {

	logs, sub, err := _ETHDKG.contract.FilterLogs(opts, "ShareDistributionComplete")
	if err != nil {
		return nil, err
	}
	return &ETHDKGShareDistributionCompleteIterator{contract: _ETHDKG.contract, event: "ShareDistributionComplete", logs: logs, sub: sub}, nil
}

// WatchShareDistributionComplete is a free log subscription operation binding the contract event 0xbfe94ffef5ddde4d25ac7b652f3f67686ea63f9badbfe1f25451e26fc262d11c.
//
// Solidity: event ShareDistributionComplete(uint256 blockNumber)
func (_ETHDKG *ETHDKGFilterer) WatchShareDistributionComplete(opts *bind.WatchOpts, sink chan<- *ETHDKGShareDistributionComplete) (event.Subscription, error) {

	logs, sub, err := _ETHDKG.contract.WatchLogs(opts, "ShareDistributionComplete")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHDKGShareDistributionComplete)
				if err := _ETHDKG.contract.UnpackLog(event, "ShareDistributionComplete", log); err != nil {
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

// ParseShareDistributionComplete is a log parse operation binding the contract event 0xbfe94ffef5ddde4d25ac7b652f3f67686ea63f9badbfe1f25451e26fc262d11c.
//
// Solidity: event ShareDistributionComplete(uint256 blockNumber)
func (_ETHDKG *ETHDKGFilterer) ParseShareDistributionComplete(log types.Log) (*ETHDKGShareDistributionComplete, error) {
	event := new(ETHDKGShareDistributionComplete)
	if err := _ETHDKG.contract.UnpackLog(event, "ShareDistributionComplete", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ETHDKGSharesDistributedIterator is returned from FilterSharesDistributed and is used to iterate over the raw logs and unpacked data for SharesDistributed events raised by the ETHDKG contract.
type ETHDKGSharesDistributedIterator struct {
	Event *ETHDKGSharesDistributed // Event containing the contract specifics and raw log

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
func (it *ETHDKGSharesDistributedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHDKGSharesDistributed)
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
		it.Event = new(ETHDKGSharesDistributed)
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
func (it *ETHDKGSharesDistributedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHDKGSharesDistributedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHDKGSharesDistributed represents a SharesDistributed event raised by the ETHDKG contract.
type ETHDKGSharesDistributed struct {
	Account         common.Address
	Index           *big.Int
	Nonce           *big.Int
	EncryptedShares []*big.Int
	Commitments     [][2]*big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterSharesDistributed is a free log retrieval operation binding the contract event 0xf0c8b0ef2867c2b4639b404a0296b6bbf0bf97e20856af42144a5a6035c0d0d2.
//
// Solidity: event SharesDistributed(address account, uint256 index, uint256 nonce, uint256[] encryptedShares, uint256[2][] commitments)
func (_ETHDKG *ETHDKGFilterer) FilterSharesDistributed(opts *bind.FilterOpts) (*ETHDKGSharesDistributedIterator, error) {

	logs, sub, err := _ETHDKG.contract.FilterLogs(opts, "SharesDistributed")
	if err != nil {
		return nil, err
	}
	return &ETHDKGSharesDistributedIterator{contract: _ETHDKG.contract, event: "SharesDistributed", logs: logs, sub: sub}, nil
}

// WatchSharesDistributed is a free log subscription operation binding the contract event 0xf0c8b0ef2867c2b4639b404a0296b6bbf0bf97e20856af42144a5a6035c0d0d2.
//
// Solidity: event SharesDistributed(address account, uint256 index, uint256 nonce, uint256[] encryptedShares, uint256[2][] commitments)
func (_ETHDKG *ETHDKGFilterer) WatchSharesDistributed(opts *bind.WatchOpts, sink chan<- *ETHDKGSharesDistributed) (event.Subscription, error) {

	logs, sub, err := _ETHDKG.contract.WatchLogs(opts, "SharesDistributed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHDKGSharesDistributed)
				if err := _ETHDKG.contract.UnpackLog(event, "SharesDistributed", log); err != nil {
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

// ParseSharesDistributed is a log parse operation binding the contract event 0xf0c8b0ef2867c2b4639b404a0296b6bbf0bf97e20856af42144a5a6035c0d0d2.
//
// Solidity: event SharesDistributed(address account, uint256 index, uint256 nonce, uint256[] encryptedShares, uint256[2][] commitments)
func (_ETHDKG *ETHDKGFilterer) ParseSharesDistributed(log types.Log) (*ETHDKGSharesDistributed, error) {
	event := new(ETHDKGSharesDistributed)
	if err := _ETHDKG.contract.UnpackLog(event, "SharesDistributed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ETHDKGValidatorMemberAddedIterator is returned from FilterValidatorMemberAdded and is used to iterate over the raw logs and unpacked data for ValidatorMemberAdded events raised by the ETHDKG contract.
type ETHDKGValidatorMemberAddedIterator struct {
	Event *ETHDKGValidatorMemberAdded // Event containing the contract specifics and raw log

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
func (it *ETHDKGValidatorMemberAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHDKGValidatorMemberAdded)
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
		it.Event = new(ETHDKGValidatorMemberAdded)
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
func (it *ETHDKGValidatorMemberAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHDKGValidatorMemberAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHDKGValidatorMemberAdded represents a ValidatorMemberAdded event raised by the ETHDKG contract.
type ETHDKGValidatorMemberAdded struct {
	Account common.Address
	Index   *big.Int
	Nonce   *big.Int
	Epoch   *big.Int
	Share0  *big.Int
	Share1  *big.Int
	Share2  *big.Int
	Share3  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterValidatorMemberAdded is a free log retrieval operation binding the contract event 0x09b90b08bbc3dbe22e9d2a0bc9c2c7614c7511cd0ad72177727a1e762115bf06.
//
// Solidity: event ValidatorMemberAdded(address account, uint256 index, uint256 nonce, uint256 epoch, uint256 share0, uint256 share1, uint256 share2, uint256 share3)
func (_ETHDKG *ETHDKGFilterer) FilterValidatorMemberAdded(opts *bind.FilterOpts) (*ETHDKGValidatorMemberAddedIterator, error) {

	logs, sub, err := _ETHDKG.contract.FilterLogs(opts, "ValidatorMemberAdded")
	if err != nil {
		return nil, err
	}
	return &ETHDKGValidatorMemberAddedIterator{contract: _ETHDKG.contract, event: "ValidatorMemberAdded", logs: logs, sub: sub}, nil
}

// WatchValidatorMemberAdded is a free log subscription operation binding the contract event 0x09b90b08bbc3dbe22e9d2a0bc9c2c7614c7511cd0ad72177727a1e762115bf06.
//
// Solidity: event ValidatorMemberAdded(address account, uint256 index, uint256 nonce, uint256 epoch, uint256 share0, uint256 share1, uint256 share2, uint256 share3)
func (_ETHDKG *ETHDKGFilterer) WatchValidatorMemberAdded(opts *bind.WatchOpts, sink chan<- *ETHDKGValidatorMemberAdded) (event.Subscription, error) {

	logs, sub, err := _ETHDKG.contract.WatchLogs(opts, "ValidatorMemberAdded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHDKGValidatorMemberAdded)
				if err := _ETHDKG.contract.UnpackLog(event, "ValidatorMemberAdded", log); err != nil {
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

// ParseValidatorMemberAdded is a log parse operation binding the contract event 0x09b90b08bbc3dbe22e9d2a0bc9c2c7614c7511cd0ad72177727a1e762115bf06.
//
// Solidity: event ValidatorMemberAdded(address account, uint256 index, uint256 nonce, uint256 epoch, uint256 share0, uint256 share1, uint256 share2, uint256 share3)
func (_ETHDKG *ETHDKGFilterer) ParseValidatorMemberAdded(log types.Log) (*ETHDKGValidatorMemberAdded, error) {
	event := new(ETHDKGValidatorMemberAdded)
	if err := _ETHDKG.contract.UnpackLog(event, "ValidatorMemberAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ETHDKGValidatorSetCompletedIterator is returned from FilterValidatorSetCompleted and is used to iterate over the raw logs and unpacked data for ValidatorSetCompleted events raised by the ETHDKG contract.
type ETHDKGValidatorSetCompletedIterator struct {
	Event *ETHDKGValidatorSetCompleted // Event containing the contract specifics and raw log

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
func (it *ETHDKGValidatorSetCompletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ETHDKGValidatorSetCompleted)
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
		it.Event = new(ETHDKGValidatorSetCompleted)
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
func (it *ETHDKGValidatorSetCompletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ETHDKGValidatorSetCompletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ETHDKGValidatorSetCompleted represents a ValidatorSetCompleted event raised by the ETHDKG contract.
type ETHDKGValidatorSetCompleted struct {
	ValidatorCount *big.Int
	Nonce          *big.Int
	Epoch          *big.Int
	EthHeight      *big.Int
	AliceNetHeight *big.Int
	GroupKey0      *big.Int
	GroupKey1      *big.Int
	GroupKey2      *big.Int
	GroupKey3      *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterValidatorSetCompleted is a free log retrieval operation binding the contract event 0xd7237b781669fa700ecf77be6cd8fa0f4b98b1a24ac584a9b6b44c509216718a.
//
// Solidity: event ValidatorSetCompleted(uint256 validatorCount, uint256 nonce, uint256 epoch, uint256 ethHeight, uint256 aliceNetHeight, uint256 groupKey0, uint256 groupKey1, uint256 groupKey2, uint256 groupKey3)
func (_ETHDKG *ETHDKGFilterer) FilterValidatorSetCompleted(opts *bind.FilterOpts) (*ETHDKGValidatorSetCompletedIterator, error) {

	logs, sub, err := _ETHDKG.contract.FilterLogs(opts, "ValidatorSetCompleted")
	if err != nil {
		return nil, err
	}
	return &ETHDKGValidatorSetCompletedIterator{contract: _ETHDKG.contract, event: "ValidatorSetCompleted", logs: logs, sub: sub}, nil
}

// WatchValidatorSetCompleted is a free log subscription operation binding the contract event 0xd7237b781669fa700ecf77be6cd8fa0f4b98b1a24ac584a9b6b44c509216718a.
//
// Solidity: event ValidatorSetCompleted(uint256 validatorCount, uint256 nonce, uint256 epoch, uint256 ethHeight, uint256 aliceNetHeight, uint256 groupKey0, uint256 groupKey1, uint256 groupKey2, uint256 groupKey3)
func (_ETHDKG *ETHDKGFilterer) WatchValidatorSetCompleted(opts *bind.WatchOpts, sink chan<- *ETHDKGValidatorSetCompleted) (event.Subscription, error) {

	logs, sub, err := _ETHDKG.contract.WatchLogs(opts, "ValidatorSetCompleted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ETHDKGValidatorSetCompleted)
				if err := _ETHDKG.contract.UnpackLog(event, "ValidatorSetCompleted", log); err != nil {
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

// ParseValidatorSetCompleted is a log parse operation binding the contract event 0xd7237b781669fa700ecf77be6cd8fa0f4b98b1a24ac584a9b6b44c509216718a.
//
// Solidity: event ValidatorSetCompleted(uint256 validatorCount, uint256 nonce, uint256 epoch, uint256 ethHeight, uint256 aliceNetHeight, uint256 groupKey0, uint256 groupKey1, uint256 groupKey2, uint256 groupKey3)
func (_ETHDKG *ETHDKGFilterer) ParseValidatorSetCompleted(log types.Log) (*ETHDKGValidatorSetCompleted, error) {
	event := new(ETHDKGValidatorSetCompleted)
	if err := _ETHDKG.contract.UnpackLog(event, "ValidatorSetCompleted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
