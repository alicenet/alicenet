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

// ETHDKGErrorCodesMetaData contains all meta data concerning the ETHDKGErrorCodes contract.
var ETHDKGErrorCodesMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_DID_NOT_PARTICIPATE_IN_GPKJ_SUBMISSION\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_DID_NOT_SUBMIT_GPKJ_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_DISTRIBUTED_GPKJ\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_DISTRIBUTED_SHARES_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_HAS_COMMITMENTS\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_NOT_PARTICIPATING_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_NOT_VALIDATOR\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_PARTICIPATING_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_SUBMITTED_SHARES_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_COMMITMENT_NOT_ON_CURVE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_COMMITMENT_ZERO\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_DISPUTER_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_DISPUTER_DID_NOT_SUBMIT_GPKJ_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_DISPUTER_NOT_PARTICIPATING_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_DISTRIBUTED_SHARE_HASH_ZERO\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_GPKJ_ZERO\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_INVALID_ARGS\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_INVALID_COMMITMENTS\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_INVALID_KEYSHARE_G1\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_INVALID_KEYSHARE_G2\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_INVALID_KEY_OR_PROOF\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_INVALID_NONCE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_INVALID_NUM_COMMITMENTS\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_INVALID_NUM_ENCRYPTED_SHARES\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_INVALID_OR_DUPLICATED_PARTICIPANT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_INVALID_SHARES_OR_COMMITMENTS\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_KEYSHARE_PHASE_INVALID_NONCE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_MASTER_PUBLIC_KEY_PAIRING_CHECK_FAILURE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_MIGRATION_INPUT_DATA_MISMATCH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_MIGRATION_INVALID_NONCE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_MIN_VALIDATORS_NOT_MET\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_DISPUTE_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_GPKJ_SUBMISSION_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_KEYSHARE_SUBMISSION_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_MASTER_PUBLIC_KEY_SUBMISSION_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_POST_GPKJ_DISPUTE_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_POST_GPKJ_SUBMISSION_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_POST_KEYSHARE_SUBMISSION_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_POST_REGISTRATION_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_POST_SHARED_DISTRIBUTION_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_REGISTRATION_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_SHARED_DISTRIBUTION_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ONLY_VALIDATORS_ALLOWED\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_PARTICIPANT_DISTRIBUTED_SHARES_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_PARTICIPANT_PARTICIPATING_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_PARTICIPANT_SUBMITTED_GPKJ_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_PARTICIPANT_SUBMITTED_KEYSHARES_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_PUBLIC_KEY_NOT_ON_CURVE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_PUBLIC_KEY_ZERO\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_REQUISITES_INCOMPLETE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_SHARES_AND_COMMITMENTS_MISMATCH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_VARIABLE_CANNOT_BE_SET_WHILE_RUNNING\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ETHDKGErrorCodesABI is the input ABI used to generate the binding from.
// Deprecated: Use ETHDKGErrorCodesMetaData.ABI instead.
var ETHDKGErrorCodesABI = ETHDKGErrorCodesMetaData.ABI

// ETHDKGErrorCodes is an auto generated Go binding around an Ethereum contract.
type ETHDKGErrorCodes struct {
	ETHDKGErrorCodesCaller     // Read-only binding to the contract
	ETHDKGErrorCodesTransactor // Write-only binding to the contract
	ETHDKGErrorCodesFilterer   // Log filterer for contract events
}

// ETHDKGErrorCodesCaller is an auto generated read-only Go binding around an Ethereum contract.
type ETHDKGErrorCodesCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ETHDKGErrorCodesTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ETHDKGErrorCodesTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ETHDKGErrorCodesFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ETHDKGErrorCodesFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ETHDKGErrorCodesSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ETHDKGErrorCodesSession struct {
	Contract     *ETHDKGErrorCodes // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ETHDKGErrorCodesCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ETHDKGErrorCodesCallerSession struct {
	Contract *ETHDKGErrorCodesCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// ETHDKGErrorCodesTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ETHDKGErrorCodesTransactorSession struct {
	Contract     *ETHDKGErrorCodesTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// ETHDKGErrorCodesRaw is an auto generated low-level Go binding around an Ethereum contract.
type ETHDKGErrorCodesRaw struct {
	Contract *ETHDKGErrorCodes // Generic contract binding to access the raw methods on
}

// ETHDKGErrorCodesCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ETHDKGErrorCodesCallerRaw struct {
	Contract *ETHDKGErrorCodesCaller // Generic read-only contract binding to access the raw methods on
}

// ETHDKGErrorCodesTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ETHDKGErrorCodesTransactorRaw struct {
	Contract *ETHDKGErrorCodesTransactor // Generic write-only contract binding to access the raw methods on
}

// NewETHDKGErrorCodes creates a new instance of ETHDKGErrorCodes, bound to a specific deployed contract.
func NewETHDKGErrorCodes(address common.Address, backend bind.ContractBackend) (*ETHDKGErrorCodes, error) {
	contract, err := bindETHDKGErrorCodes(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ETHDKGErrorCodes{ETHDKGErrorCodesCaller: ETHDKGErrorCodesCaller{contract: contract}, ETHDKGErrorCodesTransactor: ETHDKGErrorCodesTransactor{contract: contract}, ETHDKGErrorCodesFilterer: ETHDKGErrorCodesFilterer{contract: contract}}, nil
}

// NewETHDKGErrorCodesCaller creates a new read-only instance of ETHDKGErrorCodes, bound to a specific deployed contract.
func NewETHDKGErrorCodesCaller(address common.Address, caller bind.ContractCaller) (*ETHDKGErrorCodesCaller, error) {
	contract, err := bindETHDKGErrorCodes(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ETHDKGErrorCodesCaller{contract: contract}, nil
}

// NewETHDKGErrorCodesTransactor creates a new write-only instance of ETHDKGErrorCodes, bound to a specific deployed contract.
func NewETHDKGErrorCodesTransactor(address common.Address, transactor bind.ContractTransactor) (*ETHDKGErrorCodesTransactor, error) {
	contract, err := bindETHDKGErrorCodes(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ETHDKGErrorCodesTransactor{contract: contract}, nil
}

// NewETHDKGErrorCodesFilterer creates a new log filterer instance of ETHDKGErrorCodes, bound to a specific deployed contract.
func NewETHDKGErrorCodesFilterer(address common.Address, filterer bind.ContractFilterer) (*ETHDKGErrorCodesFilterer, error) {
	contract, err := bindETHDKGErrorCodes(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ETHDKGErrorCodesFilterer{contract: contract}, nil
}

// bindETHDKGErrorCodes binds a generic wrapper to an already deployed contract.
func bindETHDKGErrorCodes(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ETHDKGErrorCodesABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ETHDKGErrorCodes *ETHDKGErrorCodesRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ETHDKGErrorCodes.Contract.ETHDKGErrorCodesCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ETHDKGErrorCodes *ETHDKGErrorCodesRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGErrorCodesTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ETHDKGErrorCodes *ETHDKGErrorCodesRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGErrorCodesTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ETHDKGErrorCodes.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ETHDKGErrorCodes *ETHDKGErrorCodesTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ETHDKGErrorCodes.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ETHDKGErrorCodes *ETHDKGErrorCodesTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ETHDKGErrorCodes.Contract.contract.Transact(opts, method, params...)
}

// ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND is a free data retrieval call binding the contract method 0xfa0f33b9.
//
// Solidity: function ETHDKG_ACCUSED_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_ACCUSED_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND is a free data retrieval call binding the contract method 0xfa0f33b9.
//
// Solidity: function ETHDKG_ACCUSED_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND is a free data retrieval call binding the contract method 0xfa0f33b9.
//
// Solidity: function ETHDKG_ACCUSED_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDDIDNOTPARTICIPATEINGPKJSUBMISSION is a free data retrieval call binding the contract method 0x0348f5cc.
//
// Solidity: function ETHDKG_ACCUSED_DID_NOT_PARTICIPATE_IN_GPKJ_SUBMISSION() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGACCUSEDDIDNOTPARTICIPATEINGPKJSUBMISSION(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_ACCUSED_DID_NOT_PARTICIPATE_IN_GPKJ_SUBMISSION")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDDIDNOTPARTICIPATEINGPKJSUBMISSION is a free data retrieval call binding the contract method 0x0348f5cc.
//
// Solidity: function ETHDKG_ACCUSED_DID_NOT_PARTICIPATE_IN_GPKJ_SUBMISSION() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGACCUSEDDIDNOTPARTICIPATEINGPKJSUBMISSION() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDDIDNOTPARTICIPATEINGPKJSUBMISSION(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDDIDNOTPARTICIPATEINGPKJSUBMISSION is a free data retrieval call binding the contract method 0x0348f5cc.
//
// Solidity: function ETHDKG_ACCUSED_DID_NOT_PARTICIPATE_IN_GPKJ_SUBMISSION() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGACCUSEDDIDNOTPARTICIPATEINGPKJSUBMISSION() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDDIDNOTPARTICIPATEINGPKJSUBMISSION(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDDIDNOTSUBMITGPKJINROUND is a free data retrieval call binding the contract method 0xac4683be.
//
// Solidity: function ETHDKG_ACCUSED_DID_NOT_SUBMIT_GPKJ_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGACCUSEDDIDNOTSUBMITGPKJINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_ACCUSED_DID_NOT_SUBMIT_GPKJ_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDDIDNOTSUBMITGPKJINROUND is a free data retrieval call binding the contract method 0xac4683be.
//
// Solidity: function ETHDKG_ACCUSED_DID_NOT_SUBMIT_GPKJ_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGACCUSEDDIDNOTSUBMITGPKJINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDDIDNOTSUBMITGPKJINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDDIDNOTSUBMITGPKJINROUND is a free data retrieval call binding the contract method 0xac4683be.
//
// Solidity: function ETHDKG_ACCUSED_DID_NOT_SUBMIT_GPKJ_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGACCUSEDDIDNOTSUBMITGPKJINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDDIDNOTSUBMITGPKJINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDDISTRIBUTEDGPKJ is a free data retrieval call binding the contract method 0xaf1c8f58.
//
// Solidity: function ETHDKG_ACCUSED_DISTRIBUTED_GPKJ() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGACCUSEDDISTRIBUTEDGPKJ(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_ACCUSED_DISTRIBUTED_GPKJ")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDDISTRIBUTEDGPKJ is a free data retrieval call binding the contract method 0xaf1c8f58.
//
// Solidity: function ETHDKG_ACCUSED_DISTRIBUTED_GPKJ() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGACCUSEDDISTRIBUTEDGPKJ() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDDISTRIBUTEDGPKJ(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDDISTRIBUTEDGPKJ is a free data retrieval call binding the contract method 0xaf1c8f58.
//
// Solidity: function ETHDKG_ACCUSED_DISTRIBUTED_GPKJ() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGACCUSEDDISTRIBUTEDGPKJ() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDDISTRIBUTEDGPKJ(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND is a free data retrieval call binding the contract method 0x2838edae.
//
// Solidity: function ETHDKG_ACCUSED_DISTRIBUTED_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_ACCUSED_DISTRIBUTED_SHARES_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND is a free data retrieval call binding the contract method 0x2838edae.
//
// Solidity: function ETHDKG_ACCUSED_DISTRIBUTED_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND is a free data retrieval call binding the contract method 0x2838edae.
//
// Solidity: function ETHDKG_ACCUSED_DISTRIBUTED_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDHASCOMMITMENTS is a free data retrieval call binding the contract method 0x4cd291bf.
//
// Solidity: function ETHDKG_ACCUSED_HAS_COMMITMENTS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGACCUSEDHASCOMMITMENTS(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_ACCUSED_HAS_COMMITMENTS")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDHASCOMMITMENTS is a free data retrieval call binding the contract method 0x4cd291bf.
//
// Solidity: function ETHDKG_ACCUSED_HAS_COMMITMENTS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGACCUSEDHASCOMMITMENTS() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDHASCOMMITMENTS(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDHASCOMMITMENTS is a free data retrieval call binding the contract method 0x4cd291bf.
//
// Solidity: function ETHDKG_ACCUSED_HAS_COMMITMENTS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGACCUSEDHASCOMMITMENTS() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDHASCOMMITMENTS(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDNOTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0xe11879cc.
//
// Solidity: function ETHDKG_ACCUSED_NOT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGACCUSEDNOTPARTICIPATINGINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_ACCUSED_NOT_PARTICIPATING_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDNOTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0xe11879cc.
//
// Solidity: function ETHDKG_ACCUSED_NOT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGACCUSEDNOTPARTICIPATINGINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDNOTPARTICIPATINGINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDNOTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0xe11879cc.
//
// Solidity: function ETHDKG_ACCUSED_NOT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGACCUSEDNOTPARTICIPATINGINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDNOTPARTICIPATINGINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDNOTVALIDATOR is a free data retrieval call binding the contract method 0xf5f46e73.
//
// Solidity: function ETHDKG_ACCUSED_NOT_VALIDATOR() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGACCUSEDNOTVALIDATOR(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_ACCUSED_NOT_VALIDATOR")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDNOTVALIDATOR is a free data retrieval call binding the contract method 0xf5f46e73.
//
// Solidity: function ETHDKG_ACCUSED_NOT_VALIDATOR() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGACCUSEDNOTVALIDATOR() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDNOTVALIDATOR(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDNOTVALIDATOR is a free data retrieval call binding the contract method 0xf5f46e73.
//
// Solidity: function ETHDKG_ACCUSED_NOT_VALIDATOR() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGACCUSEDNOTVALIDATOR() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDNOTVALIDATOR(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x55b83c56.
//
// Solidity: function ETHDKG_ACCUSED_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGACCUSEDPARTICIPATINGINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_ACCUSED_PARTICIPATING_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x55b83c56.
//
// Solidity: function ETHDKG_ACCUSED_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGACCUSEDPARTICIPATINGINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDPARTICIPATINGINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x55b83c56.
//
// Solidity: function ETHDKG_ACCUSED_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGACCUSEDPARTICIPATINGINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDPARTICIPATINGINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDSUBMITTEDSHARESINROUND is a free data retrieval call binding the contract method 0xb23b8358.
//
// Solidity: function ETHDKG_ACCUSED_SUBMITTED_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGACCUSEDSUBMITTEDSHARESINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_ACCUSED_SUBMITTED_SHARES_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDSUBMITTEDSHARESINROUND is a free data retrieval call binding the contract method 0xb23b8358.
//
// Solidity: function ETHDKG_ACCUSED_SUBMITTED_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGACCUSEDSUBMITTEDSHARESINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDSUBMITTEDSHARESINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGACCUSEDSUBMITTEDSHARESINROUND is a free data retrieval call binding the contract method 0xb23b8358.
//
// Solidity: function ETHDKG_ACCUSED_SUBMITTED_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGACCUSEDSUBMITTEDSHARESINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGACCUSEDSUBMITTEDSHARESINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGCOMMITMENTNOTONCURVE is a free data retrieval call binding the contract method 0xe58f04ed.
//
// Solidity: function ETHDKG_COMMITMENT_NOT_ON_CURVE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGCOMMITMENTNOTONCURVE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_COMMITMENT_NOT_ON_CURVE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGCOMMITMENTNOTONCURVE is a free data retrieval call binding the contract method 0xe58f04ed.
//
// Solidity: function ETHDKG_COMMITMENT_NOT_ON_CURVE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGCOMMITMENTNOTONCURVE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGCOMMITMENTNOTONCURVE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGCOMMITMENTNOTONCURVE is a free data retrieval call binding the contract method 0xe58f04ed.
//
// Solidity: function ETHDKG_COMMITMENT_NOT_ON_CURVE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGCOMMITMENTNOTONCURVE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGCOMMITMENTNOTONCURVE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGCOMMITMENTZERO is a free data retrieval call binding the contract method 0x81687f80.
//
// Solidity: function ETHDKG_COMMITMENT_ZERO() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGCOMMITMENTZERO(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_COMMITMENT_ZERO")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGCOMMITMENTZERO is a free data retrieval call binding the contract method 0x81687f80.
//
// Solidity: function ETHDKG_COMMITMENT_ZERO() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGCOMMITMENTZERO() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGCOMMITMENTZERO(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGCOMMITMENTZERO is a free data retrieval call binding the contract method 0x81687f80.
//
// Solidity: function ETHDKG_COMMITMENT_ZERO() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGCOMMITMENTZERO() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGCOMMITMENTZERO(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND is a free data retrieval call binding the contract method 0xd65915e2.
//
// Solidity: function ETHDKG_DISPUTER_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_DISPUTER_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND is a free data retrieval call binding the contract method 0xd65915e2.
//
// Solidity: function ETHDKG_DISPUTER_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND is a free data retrieval call binding the contract method 0xd65915e2.
//
// Solidity: function ETHDKG_DISPUTER_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGDISPUTERDIDNOTSUBMITGPKJINROUND is a free data retrieval call binding the contract method 0x3b2b8245.
//
// Solidity: function ETHDKG_DISPUTER_DID_NOT_SUBMIT_GPKJ_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGDISPUTERDIDNOTSUBMITGPKJINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_DISPUTER_DID_NOT_SUBMIT_GPKJ_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGDISPUTERDIDNOTSUBMITGPKJINROUND is a free data retrieval call binding the contract method 0x3b2b8245.
//
// Solidity: function ETHDKG_DISPUTER_DID_NOT_SUBMIT_GPKJ_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGDISPUTERDIDNOTSUBMITGPKJINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGDISPUTERDIDNOTSUBMITGPKJINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGDISPUTERDIDNOTSUBMITGPKJINROUND is a free data retrieval call binding the contract method 0x3b2b8245.
//
// Solidity: function ETHDKG_DISPUTER_DID_NOT_SUBMIT_GPKJ_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGDISPUTERDIDNOTSUBMITGPKJINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGDISPUTERDIDNOTSUBMITGPKJINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGDISPUTERNOTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x763df93d.
//
// Solidity: function ETHDKG_DISPUTER_NOT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGDISPUTERNOTPARTICIPATINGINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_DISPUTER_NOT_PARTICIPATING_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGDISPUTERNOTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x763df93d.
//
// Solidity: function ETHDKG_DISPUTER_NOT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGDISPUTERNOTPARTICIPATINGINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGDISPUTERNOTPARTICIPATINGINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGDISPUTERNOTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x763df93d.
//
// Solidity: function ETHDKG_DISPUTER_NOT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGDISPUTERNOTPARTICIPATINGINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGDISPUTERNOTPARTICIPATINGINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGDISTRIBUTEDSHAREHASHZERO is a free data retrieval call binding the contract method 0xf54980c7.
//
// Solidity: function ETHDKG_DISTRIBUTED_SHARE_HASH_ZERO() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGDISTRIBUTEDSHAREHASHZERO(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_DISTRIBUTED_SHARE_HASH_ZERO")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGDISTRIBUTEDSHAREHASHZERO is a free data retrieval call binding the contract method 0xf54980c7.
//
// Solidity: function ETHDKG_DISTRIBUTED_SHARE_HASH_ZERO() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGDISTRIBUTEDSHAREHASHZERO() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGDISTRIBUTEDSHAREHASHZERO(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGDISTRIBUTEDSHAREHASHZERO is a free data retrieval call binding the contract method 0xf54980c7.
//
// Solidity: function ETHDKG_DISTRIBUTED_SHARE_HASH_ZERO() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGDISTRIBUTEDSHAREHASHZERO() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGDISTRIBUTEDSHAREHASHZERO(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGGPKJZERO is a free data retrieval call binding the contract method 0xc4e9cbe3.
//
// Solidity: function ETHDKG_GPKJ_ZERO() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGGPKJZERO(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_GPKJ_ZERO")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGGPKJZERO is a free data retrieval call binding the contract method 0xc4e9cbe3.
//
// Solidity: function ETHDKG_GPKJ_ZERO() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGGPKJZERO() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGGPKJZERO(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGGPKJZERO is a free data retrieval call binding the contract method 0xc4e9cbe3.
//
// Solidity: function ETHDKG_GPKJ_ZERO() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGGPKJZERO() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGGPKJZERO(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDARGS is a free data retrieval call binding the contract method 0x4d76291d.
//
// Solidity: function ETHDKG_INVALID_ARGS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGINVALIDARGS(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_INVALID_ARGS")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGINVALIDARGS is a free data retrieval call binding the contract method 0x4d76291d.
//
// Solidity: function ETHDKG_INVALID_ARGS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGINVALIDARGS() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDARGS(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDARGS is a free data retrieval call binding the contract method 0x4d76291d.
//
// Solidity: function ETHDKG_INVALID_ARGS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGINVALIDARGS() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDARGS(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDCOMMITMENTS is a free data retrieval call binding the contract method 0xf8fd7944.
//
// Solidity: function ETHDKG_INVALID_COMMITMENTS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGINVALIDCOMMITMENTS(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_INVALID_COMMITMENTS")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGINVALIDCOMMITMENTS is a free data retrieval call binding the contract method 0xf8fd7944.
//
// Solidity: function ETHDKG_INVALID_COMMITMENTS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGINVALIDCOMMITMENTS() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDCOMMITMENTS(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDCOMMITMENTS is a free data retrieval call binding the contract method 0xf8fd7944.
//
// Solidity: function ETHDKG_INVALID_COMMITMENTS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGINVALIDCOMMITMENTS() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDCOMMITMENTS(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDKEYSHAREG1 is a free data retrieval call binding the contract method 0xdd35d7da.
//
// Solidity: function ETHDKG_INVALID_KEYSHARE_G1() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGINVALIDKEYSHAREG1(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_INVALID_KEYSHARE_G1")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGINVALIDKEYSHAREG1 is a free data retrieval call binding the contract method 0xdd35d7da.
//
// Solidity: function ETHDKG_INVALID_KEYSHARE_G1() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGINVALIDKEYSHAREG1() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDKEYSHAREG1(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDKEYSHAREG1 is a free data retrieval call binding the contract method 0xdd35d7da.
//
// Solidity: function ETHDKG_INVALID_KEYSHARE_G1() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGINVALIDKEYSHAREG1() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDKEYSHAREG1(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDKEYSHAREG2 is a free data retrieval call binding the contract method 0x8468c092.
//
// Solidity: function ETHDKG_INVALID_KEYSHARE_G2() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGINVALIDKEYSHAREG2(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_INVALID_KEYSHARE_G2")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGINVALIDKEYSHAREG2 is a free data retrieval call binding the contract method 0x8468c092.
//
// Solidity: function ETHDKG_INVALID_KEYSHARE_G2() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGINVALIDKEYSHAREG2() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDKEYSHAREG2(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDKEYSHAREG2 is a free data retrieval call binding the contract method 0x8468c092.
//
// Solidity: function ETHDKG_INVALID_KEYSHARE_G2() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGINVALIDKEYSHAREG2() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDKEYSHAREG2(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDKEYORPROOF is a free data retrieval call binding the contract method 0xa852713f.
//
// Solidity: function ETHDKG_INVALID_KEY_OR_PROOF() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGINVALIDKEYORPROOF(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_INVALID_KEY_OR_PROOF")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGINVALIDKEYORPROOF is a free data retrieval call binding the contract method 0xa852713f.
//
// Solidity: function ETHDKG_INVALID_KEY_OR_PROOF() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGINVALIDKEYORPROOF() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDKEYORPROOF(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDKEYORPROOF is a free data retrieval call binding the contract method 0xa852713f.
//
// Solidity: function ETHDKG_INVALID_KEY_OR_PROOF() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGINVALIDKEYORPROOF() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDKEYORPROOF(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDNONCE is a free data retrieval call binding the contract method 0x3341bcdf.
//
// Solidity: function ETHDKG_INVALID_NONCE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGINVALIDNONCE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_INVALID_NONCE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGINVALIDNONCE is a free data retrieval call binding the contract method 0x3341bcdf.
//
// Solidity: function ETHDKG_INVALID_NONCE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGINVALIDNONCE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDNONCE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDNONCE is a free data retrieval call binding the contract method 0x3341bcdf.
//
// Solidity: function ETHDKG_INVALID_NONCE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGINVALIDNONCE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDNONCE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDNUMCOMMITMENTS is a free data retrieval call binding the contract method 0x0bca7264.
//
// Solidity: function ETHDKG_INVALID_NUM_COMMITMENTS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGINVALIDNUMCOMMITMENTS(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_INVALID_NUM_COMMITMENTS")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGINVALIDNUMCOMMITMENTS is a free data retrieval call binding the contract method 0x0bca7264.
//
// Solidity: function ETHDKG_INVALID_NUM_COMMITMENTS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGINVALIDNUMCOMMITMENTS() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDNUMCOMMITMENTS(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDNUMCOMMITMENTS is a free data retrieval call binding the contract method 0x0bca7264.
//
// Solidity: function ETHDKG_INVALID_NUM_COMMITMENTS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGINVALIDNUMCOMMITMENTS() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDNUMCOMMITMENTS(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDNUMENCRYPTEDSHARES is a free data retrieval call binding the contract method 0xe8dcd67a.
//
// Solidity: function ETHDKG_INVALID_NUM_ENCRYPTED_SHARES() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGINVALIDNUMENCRYPTEDSHARES(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_INVALID_NUM_ENCRYPTED_SHARES")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGINVALIDNUMENCRYPTEDSHARES is a free data retrieval call binding the contract method 0xe8dcd67a.
//
// Solidity: function ETHDKG_INVALID_NUM_ENCRYPTED_SHARES() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGINVALIDNUMENCRYPTEDSHARES() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDNUMENCRYPTEDSHARES(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDNUMENCRYPTEDSHARES is a free data retrieval call binding the contract method 0xe8dcd67a.
//
// Solidity: function ETHDKG_INVALID_NUM_ENCRYPTED_SHARES() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGINVALIDNUMENCRYPTEDSHARES() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDNUMENCRYPTEDSHARES(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDORDUPLICATEDPARTICIPANT is a free data retrieval call binding the contract method 0x2f5e98de.
//
// Solidity: function ETHDKG_INVALID_OR_DUPLICATED_PARTICIPANT() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGINVALIDORDUPLICATEDPARTICIPANT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_INVALID_OR_DUPLICATED_PARTICIPANT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGINVALIDORDUPLICATEDPARTICIPANT is a free data retrieval call binding the contract method 0x2f5e98de.
//
// Solidity: function ETHDKG_INVALID_OR_DUPLICATED_PARTICIPANT() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGINVALIDORDUPLICATEDPARTICIPANT() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDORDUPLICATEDPARTICIPANT(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDORDUPLICATEDPARTICIPANT is a free data retrieval call binding the contract method 0x2f5e98de.
//
// Solidity: function ETHDKG_INVALID_OR_DUPLICATED_PARTICIPANT() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGINVALIDORDUPLICATEDPARTICIPANT() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDORDUPLICATEDPARTICIPANT(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDSHARESORCOMMITMENTS is a free data retrieval call binding the contract method 0x6d1e4818.
//
// Solidity: function ETHDKG_INVALID_SHARES_OR_COMMITMENTS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGINVALIDSHARESORCOMMITMENTS(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_INVALID_SHARES_OR_COMMITMENTS")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGINVALIDSHARESORCOMMITMENTS is a free data retrieval call binding the contract method 0x6d1e4818.
//
// Solidity: function ETHDKG_INVALID_SHARES_OR_COMMITMENTS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGINVALIDSHARESORCOMMITMENTS() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDSHARESORCOMMITMENTS(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGINVALIDSHARESORCOMMITMENTS is a free data retrieval call binding the contract method 0x6d1e4818.
//
// Solidity: function ETHDKG_INVALID_SHARES_OR_COMMITMENTS() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGINVALIDSHARESORCOMMITMENTS() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGINVALIDSHARESORCOMMITMENTS(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGKEYSHAREPHASEINVALIDNONCE is a free data retrieval call binding the contract method 0xc169fbc4.
//
// Solidity: function ETHDKG_KEYSHARE_PHASE_INVALID_NONCE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGKEYSHAREPHASEINVALIDNONCE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_KEYSHARE_PHASE_INVALID_NONCE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGKEYSHAREPHASEINVALIDNONCE is a free data retrieval call binding the contract method 0xc169fbc4.
//
// Solidity: function ETHDKG_KEYSHARE_PHASE_INVALID_NONCE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGKEYSHAREPHASEINVALIDNONCE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGKEYSHAREPHASEINVALIDNONCE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGKEYSHAREPHASEINVALIDNONCE is a free data retrieval call binding the contract method 0xc169fbc4.
//
// Solidity: function ETHDKG_KEYSHARE_PHASE_INVALID_NONCE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGKEYSHAREPHASEINVALIDNONCE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGKEYSHAREPHASEINVALIDNONCE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGMASTERPUBLICKEYPAIRINGCHECKFAILURE is a free data retrieval call binding the contract method 0x2ef8cd6e.
//
// Solidity: function ETHDKG_MASTER_PUBLIC_KEY_PAIRING_CHECK_FAILURE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGMASTERPUBLICKEYPAIRINGCHECKFAILURE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_MASTER_PUBLIC_KEY_PAIRING_CHECK_FAILURE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGMASTERPUBLICKEYPAIRINGCHECKFAILURE is a free data retrieval call binding the contract method 0x2ef8cd6e.
//
// Solidity: function ETHDKG_MASTER_PUBLIC_KEY_PAIRING_CHECK_FAILURE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGMASTERPUBLICKEYPAIRINGCHECKFAILURE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGMASTERPUBLICKEYPAIRINGCHECKFAILURE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGMASTERPUBLICKEYPAIRINGCHECKFAILURE is a free data retrieval call binding the contract method 0x2ef8cd6e.
//
// Solidity: function ETHDKG_MASTER_PUBLIC_KEY_PAIRING_CHECK_FAILURE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGMASTERPUBLICKEYPAIRINGCHECKFAILURE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGMASTERPUBLICKEYPAIRINGCHECKFAILURE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGMIGRATIONINPUTDATAMISMATCH is a free data retrieval call binding the contract method 0x64a20638.
//
// Solidity: function ETHDKG_MIGRATION_INPUT_DATA_MISMATCH() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGMIGRATIONINPUTDATAMISMATCH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_MIGRATION_INPUT_DATA_MISMATCH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGMIGRATIONINPUTDATAMISMATCH is a free data retrieval call binding the contract method 0x64a20638.
//
// Solidity: function ETHDKG_MIGRATION_INPUT_DATA_MISMATCH() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGMIGRATIONINPUTDATAMISMATCH() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGMIGRATIONINPUTDATAMISMATCH(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGMIGRATIONINPUTDATAMISMATCH is a free data retrieval call binding the contract method 0x64a20638.
//
// Solidity: function ETHDKG_MIGRATION_INPUT_DATA_MISMATCH() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGMIGRATIONINPUTDATAMISMATCH() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGMIGRATIONINPUTDATAMISMATCH(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGMIGRATIONINVALIDNONCE is a free data retrieval call binding the contract method 0x06a621df.
//
// Solidity: function ETHDKG_MIGRATION_INVALID_NONCE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGMIGRATIONINVALIDNONCE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_MIGRATION_INVALID_NONCE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGMIGRATIONINVALIDNONCE is a free data retrieval call binding the contract method 0x06a621df.
//
// Solidity: function ETHDKG_MIGRATION_INVALID_NONCE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGMIGRATIONINVALIDNONCE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGMIGRATIONINVALIDNONCE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGMIGRATIONINVALIDNONCE is a free data retrieval call binding the contract method 0x06a621df.
//
// Solidity: function ETHDKG_MIGRATION_INVALID_NONCE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGMIGRATIONINVALIDNONCE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGMIGRATIONINVALIDNONCE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGMINVALIDATORSNOTMET is a free data retrieval call binding the contract method 0x7e9f3983.
//
// Solidity: function ETHDKG_MIN_VALIDATORS_NOT_MET() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGMINVALIDATORSNOTMET(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_MIN_VALIDATORS_NOT_MET")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGMINVALIDATORSNOTMET is a free data retrieval call binding the contract method 0x7e9f3983.
//
// Solidity: function ETHDKG_MIN_VALIDATORS_NOT_MET() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGMINVALIDATORSNOTMET() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGMINVALIDATORSNOTMET(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGMINVALIDATORSNOTMET is a free data retrieval call binding the contract method 0x7e9f3983.
//
// Solidity: function ETHDKG_MIN_VALIDATORS_NOT_MET() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGMINVALIDATORSNOTMET() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGMINVALIDATORSNOTMET(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINDISPUTEPHASE is a free data retrieval call binding the contract method 0x4c4cfd75.
//
// Solidity: function ETHDKG_NOT_IN_DISPUTE_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGNOTINDISPUTEPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_NOT_IN_DISPUTE_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINDISPUTEPHASE is a free data retrieval call binding the contract method 0x4c4cfd75.
//
// Solidity: function ETHDKG_NOT_IN_DISPUTE_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGNOTINDISPUTEPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINDISPUTEPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINDISPUTEPHASE is a free data retrieval call binding the contract method 0x4c4cfd75.
//
// Solidity: function ETHDKG_NOT_IN_DISPUTE_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGNOTINDISPUTEPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINDISPUTEPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINGPKJSUBMISSIONPHASE is a free data retrieval call binding the contract method 0x53187f6a.
//
// Solidity: function ETHDKG_NOT_IN_GPKJ_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGNOTINGPKJSUBMISSIONPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_NOT_IN_GPKJ_SUBMISSION_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINGPKJSUBMISSIONPHASE is a free data retrieval call binding the contract method 0x53187f6a.
//
// Solidity: function ETHDKG_NOT_IN_GPKJ_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGNOTINGPKJSUBMISSIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINGPKJSUBMISSIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINGPKJSUBMISSIONPHASE is a free data retrieval call binding the contract method 0x53187f6a.
//
// Solidity: function ETHDKG_NOT_IN_GPKJ_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGNOTINGPKJSUBMISSIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINGPKJSUBMISSIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINKEYSHARESUBMISSIONPHASE is a free data retrieval call binding the contract method 0x8789ca22.
//
// Solidity: function ETHDKG_NOT_IN_KEYSHARE_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGNOTINKEYSHARESUBMISSIONPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_NOT_IN_KEYSHARE_SUBMISSION_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINKEYSHARESUBMISSIONPHASE is a free data retrieval call binding the contract method 0x8789ca22.
//
// Solidity: function ETHDKG_NOT_IN_KEYSHARE_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGNOTINKEYSHARESUBMISSIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINKEYSHARESUBMISSIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINKEYSHARESUBMISSIONPHASE is a free data retrieval call binding the contract method 0x8789ca22.
//
// Solidity: function ETHDKG_NOT_IN_KEYSHARE_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGNOTINKEYSHARESUBMISSIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINKEYSHARESUBMISSIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINMASTERPUBLICKEYSUBMISSIONPHASE is a free data retrieval call binding the contract method 0x3072e363.
//
// Solidity: function ETHDKG_NOT_IN_MASTER_PUBLIC_KEY_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGNOTINMASTERPUBLICKEYSUBMISSIONPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_NOT_IN_MASTER_PUBLIC_KEY_SUBMISSION_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINMASTERPUBLICKEYSUBMISSIONPHASE is a free data retrieval call binding the contract method 0x3072e363.
//
// Solidity: function ETHDKG_NOT_IN_MASTER_PUBLIC_KEY_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGNOTINMASTERPUBLICKEYSUBMISSIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINMASTERPUBLICKEYSUBMISSIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINMASTERPUBLICKEYSUBMISSIONPHASE is a free data retrieval call binding the contract method 0x3072e363.
//
// Solidity: function ETHDKG_NOT_IN_MASTER_PUBLIC_KEY_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGNOTINMASTERPUBLICKEYSUBMISSIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINMASTERPUBLICKEYSUBMISSIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINPOSTGPKJDISPUTEPHASE is a free data retrieval call binding the contract method 0xc4423781.
//
// Solidity: function ETHDKG_NOT_IN_POST_GPKJ_DISPUTE_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGNOTINPOSTGPKJDISPUTEPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_NOT_IN_POST_GPKJ_DISPUTE_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINPOSTGPKJDISPUTEPHASE is a free data retrieval call binding the contract method 0xc4423781.
//
// Solidity: function ETHDKG_NOT_IN_POST_GPKJ_DISPUTE_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGNOTINPOSTGPKJDISPUTEPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINPOSTGPKJDISPUTEPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINPOSTGPKJDISPUTEPHASE is a free data retrieval call binding the contract method 0xc4423781.
//
// Solidity: function ETHDKG_NOT_IN_POST_GPKJ_DISPUTE_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGNOTINPOSTGPKJDISPUTEPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINPOSTGPKJDISPUTEPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE is a free data retrieval call binding the contract method 0x6d429ef2.
//
// Solidity: function ETHDKG_NOT_IN_POST_GPKJ_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_NOT_IN_POST_GPKJ_SUBMISSION_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE is a free data retrieval call binding the contract method 0x6d429ef2.
//
// Solidity: function ETHDKG_NOT_IN_POST_GPKJ_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE is a free data retrieval call binding the contract method 0x6d429ef2.
//
// Solidity: function ETHDKG_NOT_IN_POST_GPKJ_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE is a free data retrieval call binding the contract method 0x60987646.
//
// Solidity: function ETHDKG_NOT_IN_POST_KEYSHARE_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_NOT_IN_POST_KEYSHARE_SUBMISSION_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE is a free data retrieval call binding the contract method 0x60987646.
//
// Solidity: function ETHDKG_NOT_IN_POST_KEYSHARE_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE is a free data retrieval call binding the contract method 0x60987646.
//
// Solidity: function ETHDKG_NOT_IN_POST_KEYSHARE_SUBMISSION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINPOSTREGISTRATIONPHASE is a free data retrieval call binding the contract method 0x7385db5d.
//
// Solidity: function ETHDKG_NOT_IN_POST_REGISTRATION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGNOTINPOSTREGISTRATIONPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_NOT_IN_POST_REGISTRATION_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINPOSTREGISTRATIONPHASE is a free data retrieval call binding the contract method 0x7385db5d.
//
// Solidity: function ETHDKG_NOT_IN_POST_REGISTRATION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGNOTINPOSTREGISTRATIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINPOSTREGISTRATIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINPOSTREGISTRATIONPHASE is a free data retrieval call binding the contract method 0x7385db5d.
//
// Solidity: function ETHDKG_NOT_IN_POST_REGISTRATION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGNOTINPOSTREGISTRATIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINPOSTREGISTRATIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE is a free data retrieval call binding the contract method 0x8e25d1e1.
//
// Solidity: function ETHDKG_NOT_IN_POST_SHARED_DISTRIBUTION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_NOT_IN_POST_SHARED_DISTRIBUTION_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE is a free data retrieval call binding the contract method 0x8e25d1e1.
//
// Solidity: function ETHDKG_NOT_IN_POST_SHARED_DISTRIBUTION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE is a free data retrieval call binding the contract method 0x8e25d1e1.
//
// Solidity: function ETHDKG_NOT_IN_POST_SHARED_DISTRIBUTION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINREGISTRATIONPHASE is a free data retrieval call binding the contract method 0xf4edda56.
//
// Solidity: function ETHDKG_NOT_IN_REGISTRATION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGNOTINREGISTRATIONPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_NOT_IN_REGISTRATION_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINREGISTRATIONPHASE is a free data retrieval call binding the contract method 0xf4edda56.
//
// Solidity: function ETHDKG_NOT_IN_REGISTRATION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGNOTINREGISTRATIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINREGISTRATIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINREGISTRATIONPHASE is a free data retrieval call binding the contract method 0xf4edda56.
//
// Solidity: function ETHDKG_NOT_IN_REGISTRATION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGNOTINREGISTRATIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINREGISTRATIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINSHAREDDISTRIBUTIONPHASE is a free data retrieval call binding the contract method 0xba50ad20.
//
// Solidity: function ETHDKG_NOT_IN_SHARED_DISTRIBUTION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGNOTINSHAREDDISTRIBUTIONPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_NOT_IN_SHARED_DISTRIBUTION_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINSHAREDDISTRIBUTIONPHASE is a free data retrieval call binding the contract method 0xba50ad20.
//
// Solidity: function ETHDKG_NOT_IN_SHARED_DISTRIBUTION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGNOTINSHAREDDISTRIBUTIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINSHAREDDISTRIBUTIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGNOTINSHAREDDISTRIBUTIONPHASE is a free data retrieval call binding the contract method 0xba50ad20.
//
// Solidity: function ETHDKG_NOT_IN_SHARED_DISTRIBUTION_PHASE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGNOTINSHAREDDISTRIBUTIONPHASE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGNOTINSHAREDDISTRIBUTIONPHASE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGONLYVALIDATORSALLOWED is a free data retrieval call binding the contract method 0x83c069e4.
//
// Solidity: function ETHDKG_ONLY_VALIDATORS_ALLOWED() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGONLYVALIDATORSALLOWED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_ONLY_VALIDATORS_ALLOWED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGONLYVALIDATORSALLOWED is a free data retrieval call binding the contract method 0x83c069e4.
//
// Solidity: function ETHDKG_ONLY_VALIDATORS_ALLOWED() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGONLYVALIDATORSALLOWED() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGONLYVALIDATORSALLOWED(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGONLYVALIDATORSALLOWED is a free data retrieval call binding the contract method 0x83c069e4.
//
// Solidity: function ETHDKG_ONLY_VALIDATORS_ALLOWED() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGONLYVALIDATORSALLOWED() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGONLYVALIDATORSALLOWED(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGPARTICIPANTDISTRIBUTEDSHARESINROUND is a free data retrieval call binding the contract method 0xbdf1d0a8.
//
// Solidity: function ETHDKG_PARTICIPANT_DISTRIBUTED_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGPARTICIPANTDISTRIBUTEDSHARESINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_PARTICIPANT_DISTRIBUTED_SHARES_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGPARTICIPANTDISTRIBUTEDSHARESINROUND is a free data retrieval call binding the contract method 0xbdf1d0a8.
//
// Solidity: function ETHDKG_PARTICIPANT_DISTRIBUTED_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGPARTICIPANTDISTRIBUTEDSHARESINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGPARTICIPANTDISTRIBUTEDSHARESINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGPARTICIPANTDISTRIBUTEDSHARESINROUND is a free data retrieval call binding the contract method 0xbdf1d0a8.
//
// Solidity: function ETHDKG_PARTICIPANT_DISTRIBUTED_SHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGPARTICIPANTDISTRIBUTEDSHARESINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGPARTICIPANTDISTRIBUTEDSHARESINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGPARTICIPANTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x5bbeb11b.
//
// Solidity: function ETHDKG_PARTICIPANT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGPARTICIPANTPARTICIPATINGINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_PARTICIPANT_PARTICIPATING_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGPARTICIPANTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x5bbeb11b.
//
// Solidity: function ETHDKG_PARTICIPANT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGPARTICIPANTPARTICIPATINGINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGPARTICIPANTPARTICIPATINGINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGPARTICIPANTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x5bbeb11b.
//
// Solidity: function ETHDKG_PARTICIPANT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGPARTICIPANTPARTICIPATINGINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGPARTICIPANTPARTICIPATINGINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGPARTICIPANTSUBMITTEDGPKJINROUND is a free data retrieval call binding the contract method 0xc43f133a.
//
// Solidity: function ETHDKG_PARTICIPANT_SUBMITTED_GPKJ_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGPARTICIPANTSUBMITTEDGPKJINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_PARTICIPANT_SUBMITTED_GPKJ_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGPARTICIPANTSUBMITTEDGPKJINROUND is a free data retrieval call binding the contract method 0xc43f133a.
//
// Solidity: function ETHDKG_PARTICIPANT_SUBMITTED_GPKJ_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGPARTICIPANTSUBMITTEDGPKJINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGPARTICIPANTSUBMITTEDGPKJINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGPARTICIPANTSUBMITTEDGPKJINROUND is a free data retrieval call binding the contract method 0xc43f133a.
//
// Solidity: function ETHDKG_PARTICIPANT_SUBMITTED_GPKJ_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGPARTICIPANTSUBMITTEDGPKJINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGPARTICIPANTSUBMITTEDGPKJINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGPARTICIPANTSUBMITTEDKEYSHARESINROUND is a free data retrieval call binding the contract method 0x1723b084.
//
// Solidity: function ETHDKG_PARTICIPANT_SUBMITTED_KEYSHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGPARTICIPANTSUBMITTEDKEYSHARESINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_PARTICIPANT_SUBMITTED_KEYSHARES_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGPARTICIPANTSUBMITTEDKEYSHARESINROUND is a free data retrieval call binding the contract method 0x1723b084.
//
// Solidity: function ETHDKG_PARTICIPANT_SUBMITTED_KEYSHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGPARTICIPANTSUBMITTEDKEYSHARESINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGPARTICIPANTSUBMITTEDKEYSHARESINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGPARTICIPANTSUBMITTEDKEYSHARESINROUND is a free data retrieval call binding the contract method 0x1723b084.
//
// Solidity: function ETHDKG_PARTICIPANT_SUBMITTED_KEYSHARES_IN_ROUND() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGPARTICIPANTSUBMITTEDKEYSHARESINROUND() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGPARTICIPANTSUBMITTEDKEYSHARESINROUND(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGPUBLICKEYNOTONCURVE is a free data retrieval call binding the contract method 0x29896fd4.
//
// Solidity: function ETHDKG_PUBLIC_KEY_NOT_ON_CURVE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGPUBLICKEYNOTONCURVE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_PUBLIC_KEY_NOT_ON_CURVE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGPUBLICKEYNOTONCURVE is a free data retrieval call binding the contract method 0x29896fd4.
//
// Solidity: function ETHDKG_PUBLIC_KEY_NOT_ON_CURVE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGPUBLICKEYNOTONCURVE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGPUBLICKEYNOTONCURVE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGPUBLICKEYNOTONCURVE is a free data retrieval call binding the contract method 0x29896fd4.
//
// Solidity: function ETHDKG_PUBLIC_KEY_NOT_ON_CURVE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGPUBLICKEYNOTONCURVE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGPUBLICKEYNOTONCURVE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGPUBLICKEYZERO is a free data retrieval call binding the contract method 0x54ad808e.
//
// Solidity: function ETHDKG_PUBLIC_KEY_ZERO() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGPUBLICKEYZERO(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_PUBLIC_KEY_ZERO")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGPUBLICKEYZERO is a free data retrieval call binding the contract method 0x54ad808e.
//
// Solidity: function ETHDKG_PUBLIC_KEY_ZERO() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGPUBLICKEYZERO() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGPUBLICKEYZERO(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGPUBLICKEYZERO is a free data retrieval call binding the contract method 0x54ad808e.
//
// Solidity: function ETHDKG_PUBLIC_KEY_ZERO() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGPUBLICKEYZERO() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGPUBLICKEYZERO(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGREQUISITESINCOMPLETE is a free data retrieval call binding the contract method 0x23277f87.
//
// Solidity: function ETHDKG_REQUISITES_INCOMPLETE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGREQUISITESINCOMPLETE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_REQUISITES_INCOMPLETE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGREQUISITESINCOMPLETE is a free data retrieval call binding the contract method 0x23277f87.
//
// Solidity: function ETHDKG_REQUISITES_INCOMPLETE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGREQUISITESINCOMPLETE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGREQUISITESINCOMPLETE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGREQUISITESINCOMPLETE is a free data retrieval call binding the contract method 0x23277f87.
//
// Solidity: function ETHDKG_REQUISITES_INCOMPLETE() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGREQUISITESINCOMPLETE() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGREQUISITESINCOMPLETE(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGSHARESANDCOMMITMENTSMISMATCH is a free data retrieval call binding the contract method 0xe0c1afd8.
//
// Solidity: function ETHDKG_SHARES_AND_COMMITMENTS_MISMATCH() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGSHARESANDCOMMITMENTSMISMATCH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_SHARES_AND_COMMITMENTS_MISMATCH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGSHARESANDCOMMITMENTSMISMATCH is a free data retrieval call binding the contract method 0xe0c1afd8.
//
// Solidity: function ETHDKG_SHARES_AND_COMMITMENTS_MISMATCH() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGSHARESANDCOMMITMENTSMISMATCH() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGSHARESANDCOMMITMENTSMISMATCH(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGSHARESANDCOMMITMENTSMISMATCH is a free data retrieval call binding the contract method 0xe0c1afd8.
//
// Solidity: function ETHDKG_SHARES_AND_COMMITMENTS_MISMATCH() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGSHARESANDCOMMITMENTSMISMATCH() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGSHARESANDCOMMITMENTSMISMATCH(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGVARIABLECANNOTBESETWHILERUNNING is a free data retrieval call binding the contract method 0x79ec8296.
//
// Solidity: function ETHDKG_VARIABLE_CANNOT_BE_SET_WHILE_RUNNING() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCaller) ETHDKGVARIABLECANNOTBESETWHILERUNNING(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ETHDKGErrorCodes.contract.Call(opts, &out, "ETHDKG_VARIABLE_CANNOT_BE_SET_WHILE_RUNNING")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGVARIABLECANNOTBESETWHILERUNNING is a free data retrieval call binding the contract method 0x79ec8296.
//
// Solidity: function ETHDKG_VARIABLE_CANNOT_BE_SET_WHILE_RUNNING() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesSession) ETHDKGVARIABLECANNOTBESETWHILERUNNING() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGVARIABLECANNOTBESETWHILERUNNING(&_ETHDKGErrorCodes.CallOpts)
}

// ETHDKGVARIABLECANNOTBESETWHILERUNNING is a free data retrieval call binding the contract method 0x79ec8296.
//
// Solidity: function ETHDKG_VARIABLE_CANNOT_BE_SET_WHILE_RUNNING() view returns(bytes32)
func (_ETHDKGErrorCodes *ETHDKGErrorCodesCallerSession) ETHDKGVARIABLECANNOTBESETWHILERUNNING() ([32]byte, error) {
	return _ETHDKGErrorCodes.Contract.ETHDKGVARIABLECANNOTBESETWHILERUNNING(&_ETHDKGErrorCodes.CallOpts)
}
