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

// ValidatorData is an auto generated low-level Go binding around an user-defined struct.
type ValidatorData struct {
	Address common.Address
	TokenID *big.Int
}

// ValidatorPoolMetaData contains all meta data concerning the ValidatorPool contract.
var ValidatorPoolMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"MaintenanceScheduled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"validatorStakingTokenID\",\"type\":\"uint256\"}],\"name\":\"ValidatorJoined\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"publicStakingTokenID\",\"type\":\"uint256\"}],\"name\":\"ValidatorLeft\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"ValidatorMajorSlashed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"publicStakingTokenID\",\"type\":\"uint256\"}],\"name\":\"ValidatorMinorSlashed\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"CLAIM_PERIOD\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_INTERVAL_WITHOUT_SNAPSHOTS\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"POSITION_LOCK_PERIOD\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimExitingNFTPosition\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"collectProfits\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"payoutEth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"payoutToken\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"completeETHDKG\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDisputerReward\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"validator_\",\"type\":\"address\"}],\"name\":\"getLocation\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"validators_\",\"type\":\"address[]\"}],\"name\":\"getLocations\",\"outputs\":[{\"internalType\":\"string[]\",\"name\":\"\",\"type\":\"string[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMaxNumValidators\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getStakeAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index_\",\"type\":\"uint256\"}],\"name\":\"getValidator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index_\",\"type\":\"uint256\"}],\"name\":\"getValidatorData\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"_address\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_tokenID\",\"type\":\"uint256\"}],\"internalType\":\"structValidatorData\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getValidatorsAddresses\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getValidatorsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"stakeAmount_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxNumValidators_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"disputerReward_\",\"type\":\"uint256\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initializeETHDKG\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account_\",\"type\":\"address\"}],\"name\":\"isAccusable\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isConsensusRunning\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account_\",\"type\":\"address\"}],\"name\":\"isInExitingQueue\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isMaintenanceScheduled\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account_\",\"type\":\"address\"}],\"name\":\"isValidator\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"dishonestValidator_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"disputer_\",\"type\":\"address\"}],\"name\":\"majorSlash\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"dishonestValidator_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"disputer_\",\"type\":\"address\"}],\"name\":\"minorSlash\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"onERC721Received\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pauseConsensus\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"aliceNetHeight_\",\"type\":\"uint256\"}],\"name\":\"pauseConsensusOnArbitraryHeight\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"validators_\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"stakerTokenIDs_\",\"type\":\"uint256[]\"}],\"name\":\"registerValidators\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"scheduleMaintenance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"disputerReward_\",\"type\":\"uint256\"}],\"name\":\"setDisputerReward\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"ip_\",\"type\":\"string\"}],\"name\":\"setLocation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"maxNumValidators_\",\"type\":\"uint256\"}],\"name\":\"setMaxNumValidators\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"stakeAmount_\",\"type\":\"uint256\"}],\"name\":\"setStakeAmount\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"}],\"name\":\"skimExcessEth\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"excess\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"}],\"name\":\"skimExcessToken\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"excess\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account_\",\"type\":\"address\"}],\"name\":\"tryGetTokenID\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unregisterAllValidators\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"validators_\",\"type\":\"address[]\"}],\"name\":\"unregisterValidators\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// ValidatorPoolABI is the input ABI used to generate the binding from.
// Deprecated: Use ValidatorPoolMetaData.ABI instead.
var ValidatorPoolABI = ValidatorPoolMetaData.ABI

// ValidatorPool is an auto generated Go binding around an Ethereum contract.
type ValidatorPool struct {
	ValidatorPoolCaller     // Read-only binding to the contract
	ValidatorPoolTransactor // Write-only binding to the contract
	ValidatorPoolFilterer   // Log filterer for contract events
}

// ValidatorPoolCaller is an auto generated read-only Go binding around an Ethereum contract.
type ValidatorPoolCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorPoolTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ValidatorPoolTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorPoolFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ValidatorPoolFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorPoolSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ValidatorPoolSession struct {
	Contract     *ValidatorPool    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ValidatorPoolCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ValidatorPoolCallerSession struct {
	Contract *ValidatorPoolCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// ValidatorPoolTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ValidatorPoolTransactorSession struct {
	Contract     *ValidatorPoolTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// ValidatorPoolRaw is an auto generated low-level Go binding around an Ethereum contract.
type ValidatorPoolRaw struct {
	Contract *ValidatorPool // Generic contract binding to access the raw methods on
}

// ValidatorPoolCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ValidatorPoolCallerRaw struct {
	Contract *ValidatorPoolCaller // Generic read-only contract binding to access the raw methods on
}

// ValidatorPoolTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ValidatorPoolTransactorRaw struct {
	Contract *ValidatorPoolTransactor // Generic write-only contract binding to access the raw methods on
}

// NewValidatorPool creates a new instance of ValidatorPool, bound to a specific deployed contract.
func NewValidatorPool(address common.Address, backend bind.ContractBackend) (*ValidatorPool, error) {
	contract, err := bindValidatorPool(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ValidatorPool{ValidatorPoolCaller: ValidatorPoolCaller{contract: contract}, ValidatorPoolTransactor: ValidatorPoolTransactor{contract: contract}, ValidatorPoolFilterer: ValidatorPoolFilterer{contract: contract}}, nil
}

// NewValidatorPoolCaller creates a new read-only instance of ValidatorPool, bound to a specific deployed contract.
func NewValidatorPoolCaller(address common.Address, caller bind.ContractCaller) (*ValidatorPoolCaller, error) {
	contract, err := bindValidatorPool(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolCaller{contract: contract}, nil
}

// NewValidatorPoolTransactor creates a new write-only instance of ValidatorPool, bound to a specific deployed contract.
func NewValidatorPoolTransactor(address common.Address, transactor bind.ContractTransactor) (*ValidatorPoolTransactor, error) {
	contract, err := bindValidatorPool(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolTransactor{contract: contract}, nil
}

// NewValidatorPoolFilterer creates a new log filterer instance of ValidatorPool, bound to a specific deployed contract.
func NewValidatorPoolFilterer(address common.Address, filterer bind.ContractFilterer) (*ValidatorPoolFilterer, error) {
	contract, err := bindValidatorPool(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolFilterer{contract: contract}, nil
}

// bindValidatorPool binds a generic wrapper to an already deployed contract.
func bindValidatorPool(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ValidatorPoolABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ValidatorPool *ValidatorPoolRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ValidatorPool.Contract.ValidatorPoolCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ValidatorPool *ValidatorPoolRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPool.Contract.ValidatorPoolTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ValidatorPool *ValidatorPoolRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ValidatorPool.Contract.ValidatorPoolTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ValidatorPool *ValidatorPoolCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ValidatorPool.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ValidatorPool *ValidatorPoolTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPool.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ValidatorPool *ValidatorPoolTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ValidatorPool.Contract.contract.Transact(opts, method, params...)
}

// CLAIMPERIOD is a free data retrieval call binding the contract method 0x21241dfe.
//
// Solidity: function CLAIM_PERIOD() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCaller) CLAIMPERIOD(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "CLAIM_PERIOD")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CLAIMPERIOD is a free data retrieval call binding the contract method 0x21241dfe.
//
// Solidity: function CLAIM_PERIOD() view returns(uint256)
func (_ValidatorPool *ValidatorPoolSession) CLAIMPERIOD() (*big.Int, error) {
	return _ValidatorPool.Contract.CLAIMPERIOD(&_ValidatorPool.CallOpts)
}

// CLAIMPERIOD is a free data retrieval call binding the contract method 0x21241dfe.
//
// Solidity: function CLAIM_PERIOD() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCallerSession) CLAIMPERIOD() (*big.Int, error) {
	return _ValidatorPool.Contract.CLAIMPERIOD(&_ValidatorPool.CallOpts)
}

// MAXINTERVALWITHOUTSNAPSHOTS is a free data retrieval call binding the contract method 0x61aee135.
//
// Solidity: function MAX_INTERVAL_WITHOUT_SNAPSHOTS() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCaller) MAXINTERVALWITHOUTSNAPSHOTS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "MAX_INTERVAL_WITHOUT_SNAPSHOTS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXINTERVALWITHOUTSNAPSHOTS is a free data retrieval call binding the contract method 0x61aee135.
//
// Solidity: function MAX_INTERVAL_WITHOUT_SNAPSHOTS() view returns(uint256)
func (_ValidatorPool *ValidatorPoolSession) MAXINTERVALWITHOUTSNAPSHOTS() (*big.Int, error) {
	return _ValidatorPool.Contract.MAXINTERVALWITHOUTSNAPSHOTS(&_ValidatorPool.CallOpts)
}

// MAXINTERVALWITHOUTSNAPSHOTS is a free data retrieval call binding the contract method 0x61aee135.
//
// Solidity: function MAX_INTERVAL_WITHOUT_SNAPSHOTS() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCallerSession) MAXINTERVALWITHOUTSNAPSHOTS() (*big.Int, error) {
	return _ValidatorPool.Contract.MAXINTERVALWITHOUTSNAPSHOTS(&_ValidatorPool.CallOpts)
}

// POSITIONLOCKPERIOD is a free data retrieval call binding the contract method 0x9c87e3ed.
//
// Solidity: function POSITION_LOCK_PERIOD() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCaller) POSITIONLOCKPERIOD(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "POSITION_LOCK_PERIOD")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// POSITIONLOCKPERIOD is a free data retrieval call binding the contract method 0x9c87e3ed.
//
// Solidity: function POSITION_LOCK_PERIOD() view returns(uint256)
func (_ValidatorPool *ValidatorPoolSession) POSITIONLOCKPERIOD() (*big.Int, error) {
	return _ValidatorPool.Contract.POSITIONLOCKPERIOD(&_ValidatorPool.CallOpts)
}

// POSITIONLOCKPERIOD is a free data retrieval call binding the contract method 0x9c87e3ed.
//
// Solidity: function POSITION_LOCK_PERIOD() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCallerSession) POSITIONLOCKPERIOD() (*big.Int, error) {
	return _ValidatorPool.Contract.POSITIONLOCKPERIOD(&_ValidatorPool.CallOpts)
}

// GetDisputerReward is a free data retrieval call binding the contract method 0x9ccdf830.
//
// Solidity: function getDisputerReward() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCaller) GetDisputerReward(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "getDisputerReward")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetDisputerReward is a free data retrieval call binding the contract method 0x9ccdf830.
//
// Solidity: function getDisputerReward() view returns(uint256)
func (_ValidatorPool *ValidatorPoolSession) GetDisputerReward() (*big.Int, error) {
	return _ValidatorPool.Contract.GetDisputerReward(&_ValidatorPool.CallOpts)
}

// GetDisputerReward is a free data retrieval call binding the contract method 0x9ccdf830.
//
// Solidity: function getDisputerReward() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCallerSession) GetDisputerReward() (*big.Int, error) {
	return _ValidatorPool.Contract.GetDisputerReward(&_ValidatorPool.CallOpts)
}

// GetLocation is a free data retrieval call binding the contract method 0xd9e0dc59.
//
// Solidity: function getLocation(address validator_) view returns(string)
func (_ValidatorPool *ValidatorPoolCaller) GetLocation(opts *bind.CallOpts, validator_ common.Address) (string, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "getLocation", validator_)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetLocation is a free data retrieval call binding the contract method 0xd9e0dc59.
//
// Solidity: function getLocation(address validator_) view returns(string)
func (_ValidatorPool *ValidatorPoolSession) GetLocation(validator_ common.Address) (string, error) {
	return _ValidatorPool.Contract.GetLocation(&_ValidatorPool.CallOpts, validator_)
}

// GetLocation is a free data retrieval call binding the contract method 0xd9e0dc59.
//
// Solidity: function getLocation(address validator_) view returns(string)
func (_ValidatorPool *ValidatorPoolCallerSession) GetLocation(validator_ common.Address) (string, error) {
	return _ValidatorPool.Contract.GetLocation(&_ValidatorPool.CallOpts, validator_)
}

// GetLocations is a free data retrieval call binding the contract method 0x76207f9c.
//
// Solidity: function getLocations(address[] validators_) view returns(string[])
func (_ValidatorPool *ValidatorPoolCaller) GetLocations(opts *bind.CallOpts, validators_ []common.Address) ([]string, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "getLocations", validators_)

	if err != nil {
		return *new([]string), err
	}

	out0 := *abi.ConvertType(out[0], new([]string)).(*[]string)

	return out0, err

}

// GetLocations is a free data retrieval call binding the contract method 0x76207f9c.
//
// Solidity: function getLocations(address[] validators_) view returns(string[])
func (_ValidatorPool *ValidatorPoolSession) GetLocations(validators_ []common.Address) ([]string, error) {
	return _ValidatorPool.Contract.GetLocations(&_ValidatorPool.CallOpts, validators_)
}

// GetLocations is a free data retrieval call binding the contract method 0x76207f9c.
//
// Solidity: function getLocations(address[] validators_) view returns(string[])
func (_ValidatorPool *ValidatorPoolCallerSession) GetLocations(validators_ []common.Address) ([]string, error) {
	return _ValidatorPool.Contract.GetLocations(&_ValidatorPool.CallOpts, validators_)
}

// GetMaxNumValidators is a free data retrieval call binding the contract method 0xd2992f54.
//
// Solidity: function getMaxNumValidators() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCaller) GetMaxNumValidators(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "getMaxNumValidators")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMaxNumValidators is a free data retrieval call binding the contract method 0xd2992f54.
//
// Solidity: function getMaxNumValidators() view returns(uint256)
func (_ValidatorPool *ValidatorPoolSession) GetMaxNumValidators() (*big.Int, error) {
	return _ValidatorPool.Contract.GetMaxNumValidators(&_ValidatorPool.CallOpts)
}

// GetMaxNumValidators is a free data retrieval call binding the contract method 0xd2992f54.
//
// Solidity: function getMaxNumValidators() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCallerSession) GetMaxNumValidators() (*big.Int, error) {
	return _ValidatorPool.Contract.GetMaxNumValidators(&_ValidatorPool.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ValidatorPool *ValidatorPoolCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ValidatorPool *ValidatorPoolSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ValidatorPool.Contract.GetMetamorphicContractAddress(&_ValidatorPool.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ValidatorPool *ValidatorPoolCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ValidatorPool.Contract.GetMetamorphicContractAddress(&_ValidatorPool.CallOpts, _salt, _factory)
}

// GetStakeAmount is a free data retrieval call binding the contract method 0x722580b6.
//
// Solidity: function getStakeAmount() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCaller) GetStakeAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "getStakeAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetStakeAmount is a free data retrieval call binding the contract method 0x722580b6.
//
// Solidity: function getStakeAmount() view returns(uint256)
func (_ValidatorPool *ValidatorPoolSession) GetStakeAmount() (*big.Int, error) {
	return _ValidatorPool.Contract.GetStakeAmount(&_ValidatorPool.CallOpts)
}

// GetStakeAmount is a free data retrieval call binding the contract method 0x722580b6.
//
// Solidity: function getStakeAmount() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCallerSession) GetStakeAmount() (*big.Int, error) {
	return _ValidatorPool.Contract.GetStakeAmount(&_ValidatorPool.CallOpts)
}

// GetValidator is a free data retrieval call binding the contract method 0xb5d89627.
//
// Solidity: function getValidator(uint256 index_) view returns(address)
func (_ValidatorPool *ValidatorPoolCaller) GetValidator(opts *bind.CallOpts, index_ *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "getValidator", index_)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetValidator is a free data retrieval call binding the contract method 0xb5d89627.
//
// Solidity: function getValidator(uint256 index_) view returns(address)
func (_ValidatorPool *ValidatorPoolSession) GetValidator(index_ *big.Int) (common.Address, error) {
	return _ValidatorPool.Contract.GetValidator(&_ValidatorPool.CallOpts, index_)
}

// GetValidator is a free data retrieval call binding the contract method 0xb5d89627.
//
// Solidity: function getValidator(uint256 index_) view returns(address)
func (_ValidatorPool *ValidatorPoolCallerSession) GetValidator(index_ *big.Int) (common.Address, error) {
	return _ValidatorPool.Contract.GetValidator(&_ValidatorPool.CallOpts, index_)
}

// GetValidatorData is a free data retrieval call binding the contract method 0xc0951451.
//
// Solidity: function getValidatorData(uint256 index_) view returns((address,uint256))
func (_ValidatorPool *ValidatorPoolCaller) GetValidatorData(opts *bind.CallOpts, index_ *big.Int) (ValidatorData, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "getValidatorData", index_)

	if err != nil {
		return *new(ValidatorData), err
	}

	out0 := *abi.ConvertType(out[0], new(ValidatorData)).(*ValidatorData)

	return out0, err

}

// GetValidatorData is a free data retrieval call binding the contract method 0xc0951451.
//
// Solidity: function getValidatorData(uint256 index_) view returns((address,uint256))
func (_ValidatorPool *ValidatorPoolSession) GetValidatorData(index_ *big.Int) (ValidatorData, error) {
	return _ValidatorPool.Contract.GetValidatorData(&_ValidatorPool.CallOpts, index_)
}

// GetValidatorData is a free data retrieval call binding the contract method 0xc0951451.
//
// Solidity: function getValidatorData(uint256 index_) view returns((address,uint256))
func (_ValidatorPool *ValidatorPoolCallerSession) GetValidatorData(index_ *big.Int) (ValidatorData, error) {
	return _ValidatorPool.Contract.GetValidatorData(&_ValidatorPool.CallOpts, index_)
}

// GetValidatorsAddresses is a free data retrieval call binding the contract method 0x9c7d8961.
//
// Solidity: function getValidatorsAddresses() view returns(address[])
func (_ValidatorPool *ValidatorPoolCaller) GetValidatorsAddresses(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "getValidatorsAddresses")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetValidatorsAddresses is a free data retrieval call binding the contract method 0x9c7d8961.
//
// Solidity: function getValidatorsAddresses() view returns(address[])
func (_ValidatorPool *ValidatorPoolSession) GetValidatorsAddresses() ([]common.Address, error) {
	return _ValidatorPool.Contract.GetValidatorsAddresses(&_ValidatorPool.CallOpts)
}

// GetValidatorsAddresses is a free data retrieval call binding the contract method 0x9c7d8961.
//
// Solidity: function getValidatorsAddresses() view returns(address[])
func (_ValidatorPool *ValidatorPoolCallerSession) GetValidatorsAddresses() ([]common.Address, error) {
	return _ValidatorPool.Contract.GetValidatorsAddresses(&_ValidatorPool.CallOpts)
}

// GetValidatorsCount is a free data retrieval call binding the contract method 0x27498240.
//
// Solidity: function getValidatorsCount() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCaller) GetValidatorsCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "getValidatorsCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetValidatorsCount is a free data retrieval call binding the contract method 0x27498240.
//
// Solidity: function getValidatorsCount() view returns(uint256)
func (_ValidatorPool *ValidatorPoolSession) GetValidatorsCount() (*big.Int, error) {
	return _ValidatorPool.Contract.GetValidatorsCount(&_ValidatorPool.CallOpts)
}

// GetValidatorsCount is a free data retrieval call binding the contract method 0x27498240.
//
// Solidity: function getValidatorsCount() view returns(uint256)
func (_ValidatorPool *ValidatorPoolCallerSession) GetValidatorsCount() (*big.Int, error) {
	return _ValidatorPool.Contract.GetValidatorsCount(&_ValidatorPool.CallOpts)
}

// IsAccusable is a free data retrieval call binding the contract method 0x20c2856d.
//
// Solidity: function isAccusable(address account_) view returns(bool)
func (_ValidatorPool *ValidatorPoolCaller) IsAccusable(opts *bind.CallOpts, account_ common.Address) (bool, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "isAccusable", account_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAccusable is a free data retrieval call binding the contract method 0x20c2856d.
//
// Solidity: function isAccusable(address account_) view returns(bool)
func (_ValidatorPool *ValidatorPoolSession) IsAccusable(account_ common.Address) (bool, error) {
	return _ValidatorPool.Contract.IsAccusable(&_ValidatorPool.CallOpts, account_)
}

// IsAccusable is a free data retrieval call binding the contract method 0x20c2856d.
//
// Solidity: function isAccusable(address account_) view returns(bool)
func (_ValidatorPool *ValidatorPoolCallerSession) IsAccusable(account_ common.Address) (bool, error) {
	return _ValidatorPool.Contract.IsAccusable(&_ValidatorPool.CallOpts, account_)
}

// IsConsensusRunning is a free data retrieval call binding the contract method 0xc8d1a5e4.
//
// Solidity: function isConsensusRunning() view returns(bool)
func (_ValidatorPool *ValidatorPoolCaller) IsConsensusRunning(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "isConsensusRunning")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsConsensusRunning is a free data retrieval call binding the contract method 0xc8d1a5e4.
//
// Solidity: function isConsensusRunning() view returns(bool)
func (_ValidatorPool *ValidatorPoolSession) IsConsensusRunning() (bool, error) {
	return _ValidatorPool.Contract.IsConsensusRunning(&_ValidatorPool.CallOpts)
}

// IsConsensusRunning is a free data retrieval call binding the contract method 0xc8d1a5e4.
//
// Solidity: function isConsensusRunning() view returns(bool)
func (_ValidatorPool *ValidatorPoolCallerSession) IsConsensusRunning() (bool, error) {
	return _ValidatorPool.Contract.IsConsensusRunning(&_ValidatorPool.CallOpts)
}

// IsInExitingQueue is a free data retrieval call binding the contract method 0xe4ad75f1.
//
// Solidity: function isInExitingQueue(address account_) view returns(bool)
func (_ValidatorPool *ValidatorPoolCaller) IsInExitingQueue(opts *bind.CallOpts, account_ common.Address) (bool, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "isInExitingQueue", account_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsInExitingQueue is a free data retrieval call binding the contract method 0xe4ad75f1.
//
// Solidity: function isInExitingQueue(address account_) view returns(bool)
func (_ValidatorPool *ValidatorPoolSession) IsInExitingQueue(account_ common.Address) (bool, error) {
	return _ValidatorPool.Contract.IsInExitingQueue(&_ValidatorPool.CallOpts, account_)
}

// IsInExitingQueue is a free data retrieval call binding the contract method 0xe4ad75f1.
//
// Solidity: function isInExitingQueue(address account_) view returns(bool)
func (_ValidatorPool *ValidatorPoolCallerSession) IsInExitingQueue(account_ common.Address) (bool, error) {
	return _ValidatorPool.Contract.IsInExitingQueue(&_ValidatorPool.CallOpts, account_)
}

// IsMaintenanceScheduled is a free data retrieval call binding the contract method 0x1885570f.
//
// Solidity: function isMaintenanceScheduled() view returns(bool)
func (_ValidatorPool *ValidatorPoolCaller) IsMaintenanceScheduled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "isMaintenanceScheduled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsMaintenanceScheduled is a free data retrieval call binding the contract method 0x1885570f.
//
// Solidity: function isMaintenanceScheduled() view returns(bool)
func (_ValidatorPool *ValidatorPoolSession) IsMaintenanceScheduled() (bool, error) {
	return _ValidatorPool.Contract.IsMaintenanceScheduled(&_ValidatorPool.CallOpts)
}

// IsMaintenanceScheduled is a free data retrieval call binding the contract method 0x1885570f.
//
// Solidity: function isMaintenanceScheduled() view returns(bool)
func (_ValidatorPool *ValidatorPoolCallerSession) IsMaintenanceScheduled() (bool, error) {
	return _ValidatorPool.Contract.IsMaintenanceScheduled(&_ValidatorPool.CallOpts)
}

// IsValidator is a free data retrieval call binding the contract method 0xfacd743b.
//
// Solidity: function isValidator(address account_) view returns(bool)
func (_ValidatorPool *ValidatorPoolCaller) IsValidator(opts *bind.CallOpts, account_ common.Address) (bool, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "isValidator", account_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidator is a free data retrieval call binding the contract method 0xfacd743b.
//
// Solidity: function isValidator(address account_) view returns(bool)
func (_ValidatorPool *ValidatorPoolSession) IsValidator(account_ common.Address) (bool, error) {
	return _ValidatorPool.Contract.IsValidator(&_ValidatorPool.CallOpts, account_)
}

// IsValidator is a free data retrieval call binding the contract method 0xfacd743b.
//
// Solidity: function isValidator(address account_) view returns(bool)
func (_ValidatorPool *ValidatorPoolCallerSession) IsValidator(account_ common.Address) (bool, error) {
	return _ValidatorPool.Contract.IsValidator(&_ValidatorPool.CallOpts, account_)
}

// TryGetTokenID is a free data retrieval call binding the contract method 0xee9e49bd.
//
// Solidity: function tryGetTokenID(address account_) view returns(bool, address, uint256)
func (_ValidatorPool *ValidatorPoolCaller) TryGetTokenID(opts *bind.CallOpts, account_ common.Address) (bool, common.Address, *big.Int, error) {
	var out []interface{}
	err := _ValidatorPool.contract.Call(opts, &out, "tryGetTokenID", account_)

	if err != nil {
		return *new(bool), *new(common.Address), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)
	out1 := *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	out2 := *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return out0, out1, out2, err

}

// TryGetTokenID is a free data retrieval call binding the contract method 0xee9e49bd.
//
// Solidity: function tryGetTokenID(address account_) view returns(bool, address, uint256)
func (_ValidatorPool *ValidatorPoolSession) TryGetTokenID(account_ common.Address) (bool, common.Address, *big.Int, error) {
	return _ValidatorPool.Contract.TryGetTokenID(&_ValidatorPool.CallOpts, account_)
}

// TryGetTokenID is a free data retrieval call binding the contract method 0xee9e49bd.
//
// Solidity: function tryGetTokenID(address account_) view returns(bool, address, uint256)
func (_ValidatorPool *ValidatorPoolCallerSession) TryGetTokenID(account_ common.Address) (bool, common.Address, *big.Int, error) {
	return _ValidatorPool.Contract.TryGetTokenID(&_ValidatorPool.CallOpts, account_)
}

// ClaimExitingNFTPosition is a paid mutator transaction binding the contract method 0x769cc695.
//
// Solidity: function claimExitingNFTPosition() returns(uint256)
func (_ValidatorPool *ValidatorPoolTransactor) ClaimExitingNFTPosition(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "claimExitingNFTPosition")
}

// ClaimExitingNFTPosition is a paid mutator transaction binding the contract method 0x769cc695.
//
// Solidity: function claimExitingNFTPosition() returns(uint256)
func (_ValidatorPool *ValidatorPoolSession) ClaimExitingNFTPosition() (*types.Transaction, error) {
	return _ValidatorPool.Contract.ClaimExitingNFTPosition(&_ValidatorPool.TransactOpts)
}

// ClaimExitingNFTPosition is a paid mutator transaction binding the contract method 0x769cc695.
//
// Solidity: function claimExitingNFTPosition() returns(uint256)
func (_ValidatorPool *ValidatorPoolTransactorSession) ClaimExitingNFTPosition() (*types.Transaction, error) {
	return _ValidatorPool.Contract.ClaimExitingNFTPosition(&_ValidatorPool.TransactOpts)
}

// CollectProfits is a paid mutator transaction binding the contract method 0xc958e0d6.
//
// Solidity: function collectProfits() returns(uint256 payoutEth, uint256 payoutToken)
func (_ValidatorPool *ValidatorPoolTransactor) CollectProfits(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "collectProfits")
}

// CollectProfits is a paid mutator transaction binding the contract method 0xc958e0d6.
//
// Solidity: function collectProfits() returns(uint256 payoutEth, uint256 payoutToken)
func (_ValidatorPool *ValidatorPoolSession) CollectProfits() (*types.Transaction, error) {
	return _ValidatorPool.Contract.CollectProfits(&_ValidatorPool.TransactOpts)
}

// CollectProfits is a paid mutator transaction binding the contract method 0xc958e0d6.
//
// Solidity: function collectProfits() returns(uint256 payoutEth, uint256 payoutToken)
func (_ValidatorPool *ValidatorPoolTransactorSession) CollectProfits() (*types.Transaction, error) {
	return _ValidatorPool.Contract.CollectProfits(&_ValidatorPool.TransactOpts)
}

// CompleteETHDKG is a paid mutator transaction binding the contract method 0x8f579924.
//
// Solidity: function completeETHDKG() returns()
func (_ValidatorPool *ValidatorPoolTransactor) CompleteETHDKG(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "completeETHDKG")
}

// CompleteETHDKG is a paid mutator transaction binding the contract method 0x8f579924.
//
// Solidity: function completeETHDKG() returns()
func (_ValidatorPool *ValidatorPoolSession) CompleteETHDKG() (*types.Transaction, error) {
	return _ValidatorPool.Contract.CompleteETHDKG(&_ValidatorPool.TransactOpts)
}

// CompleteETHDKG is a paid mutator transaction binding the contract method 0x8f579924.
//
// Solidity: function completeETHDKG() returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) CompleteETHDKG() (*types.Transaction, error) {
	return _ValidatorPool.Contract.CompleteETHDKG(&_ValidatorPool.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x80d85911.
//
// Solidity: function initialize(uint256 stakeAmount_, uint256 maxNumValidators_, uint256 disputerReward_) returns()
func (_ValidatorPool *ValidatorPoolTransactor) Initialize(opts *bind.TransactOpts, stakeAmount_ *big.Int, maxNumValidators_ *big.Int, disputerReward_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "initialize", stakeAmount_, maxNumValidators_, disputerReward_)
}

// Initialize is a paid mutator transaction binding the contract method 0x80d85911.
//
// Solidity: function initialize(uint256 stakeAmount_, uint256 maxNumValidators_, uint256 disputerReward_) returns()
func (_ValidatorPool *ValidatorPoolSession) Initialize(stakeAmount_ *big.Int, maxNumValidators_ *big.Int, disputerReward_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.Contract.Initialize(&_ValidatorPool.TransactOpts, stakeAmount_, maxNumValidators_, disputerReward_)
}

// Initialize is a paid mutator transaction binding the contract method 0x80d85911.
//
// Solidity: function initialize(uint256 stakeAmount_, uint256 maxNumValidators_, uint256 disputerReward_) returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) Initialize(stakeAmount_ *big.Int, maxNumValidators_ *big.Int, disputerReward_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.Contract.Initialize(&_ValidatorPool.TransactOpts, stakeAmount_, maxNumValidators_, disputerReward_)
}

// InitializeETHDKG is a paid mutator transaction binding the contract method 0x57b51c9c.
//
// Solidity: function initializeETHDKG() returns()
func (_ValidatorPool *ValidatorPoolTransactor) InitializeETHDKG(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "initializeETHDKG")
}

// InitializeETHDKG is a paid mutator transaction binding the contract method 0x57b51c9c.
//
// Solidity: function initializeETHDKG() returns()
func (_ValidatorPool *ValidatorPoolSession) InitializeETHDKG() (*types.Transaction, error) {
	return _ValidatorPool.Contract.InitializeETHDKG(&_ValidatorPool.TransactOpts)
}

// InitializeETHDKG is a paid mutator transaction binding the contract method 0x57b51c9c.
//
// Solidity: function initializeETHDKG() returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) InitializeETHDKG() (*types.Transaction, error) {
	return _ValidatorPool.Contract.InitializeETHDKG(&_ValidatorPool.TransactOpts)
}

// MajorSlash is a paid mutator transaction binding the contract method 0x048d56c7.
//
// Solidity: function majorSlash(address dishonestValidator_, address disputer_) returns()
func (_ValidatorPool *ValidatorPoolTransactor) MajorSlash(opts *bind.TransactOpts, dishonestValidator_ common.Address, disputer_ common.Address) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "majorSlash", dishonestValidator_, disputer_)
}

// MajorSlash is a paid mutator transaction binding the contract method 0x048d56c7.
//
// Solidity: function majorSlash(address dishonestValidator_, address disputer_) returns()
func (_ValidatorPool *ValidatorPoolSession) MajorSlash(dishonestValidator_ common.Address, disputer_ common.Address) (*types.Transaction, error) {
	return _ValidatorPool.Contract.MajorSlash(&_ValidatorPool.TransactOpts, dishonestValidator_, disputer_)
}

// MajorSlash is a paid mutator transaction binding the contract method 0x048d56c7.
//
// Solidity: function majorSlash(address dishonestValidator_, address disputer_) returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) MajorSlash(dishonestValidator_ common.Address, disputer_ common.Address) (*types.Transaction, error) {
	return _ValidatorPool.Contract.MajorSlash(&_ValidatorPool.TransactOpts, dishonestValidator_, disputer_)
}

// MinorSlash is a paid mutator transaction binding the contract method 0x64c0461c.
//
// Solidity: function minorSlash(address dishonestValidator_, address disputer_) returns()
func (_ValidatorPool *ValidatorPoolTransactor) MinorSlash(opts *bind.TransactOpts, dishonestValidator_ common.Address, disputer_ common.Address) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "minorSlash", dishonestValidator_, disputer_)
}

// MinorSlash is a paid mutator transaction binding the contract method 0x64c0461c.
//
// Solidity: function minorSlash(address dishonestValidator_, address disputer_) returns()
func (_ValidatorPool *ValidatorPoolSession) MinorSlash(dishonestValidator_ common.Address, disputer_ common.Address) (*types.Transaction, error) {
	return _ValidatorPool.Contract.MinorSlash(&_ValidatorPool.TransactOpts, dishonestValidator_, disputer_)
}

// MinorSlash is a paid mutator transaction binding the contract method 0x64c0461c.
//
// Solidity: function minorSlash(address dishonestValidator_, address disputer_) returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) MinorSlash(dishonestValidator_ common.Address, disputer_ common.Address) (*types.Transaction, error) {
	return _ValidatorPool.Contract.MinorSlash(&_ValidatorPool.TransactOpts, dishonestValidator_, disputer_)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address , address , uint256 , bytes ) returns(bytes4)
func (_ValidatorPool *ValidatorPoolTransactor) OnERC721Received(opts *bind.TransactOpts, arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "onERC721Received", arg0, arg1, arg2, arg3)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address , address , uint256 , bytes ) returns(bytes4)
func (_ValidatorPool *ValidatorPoolSession) OnERC721Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _ValidatorPool.Contract.OnERC721Received(&_ValidatorPool.TransactOpts, arg0, arg1, arg2, arg3)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address , address , uint256 , bytes ) returns(bytes4)
func (_ValidatorPool *ValidatorPoolTransactorSession) OnERC721Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _ValidatorPool.Contract.OnERC721Received(&_ValidatorPool.TransactOpts, arg0, arg1, arg2, arg3)
}

// PauseConsensus is a paid mutator transaction binding the contract method 0x1e5975f4.
//
// Solidity: function pauseConsensus() returns()
func (_ValidatorPool *ValidatorPoolTransactor) PauseConsensus(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "pauseConsensus")
}

// PauseConsensus is a paid mutator transaction binding the contract method 0x1e5975f4.
//
// Solidity: function pauseConsensus() returns()
func (_ValidatorPool *ValidatorPoolSession) PauseConsensus() (*types.Transaction, error) {
	return _ValidatorPool.Contract.PauseConsensus(&_ValidatorPool.TransactOpts)
}

// PauseConsensus is a paid mutator transaction binding the contract method 0x1e5975f4.
//
// Solidity: function pauseConsensus() returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) PauseConsensus() (*types.Transaction, error) {
	return _ValidatorPool.Contract.PauseConsensus(&_ValidatorPool.TransactOpts)
}

// PauseConsensusOnArbitraryHeight is a paid mutator transaction binding the contract method 0xbc33bb01.
//
// Solidity: function pauseConsensusOnArbitraryHeight(uint256 aliceNetHeight_) returns()
func (_ValidatorPool *ValidatorPoolTransactor) PauseConsensusOnArbitraryHeight(opts *bind.TransactOpts, aliceNetHeight_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "pauseConsensusOnArbitraryHeight", aliceNetHeight_)
}

// PauseConsensusOnArbitraryHeight is a paid mutator transaction binding the contract method 0xbc33bb01.
//
// Solidity: function pauseConsensusOnArbitraryHeight(uint256 aliceNetHeight_) returns()
func (_ValidatorPool *ValidatorPoolSession) PauseConsensusOnArbitraryHeight(aliceNetHeight_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.Contract.PauseConsensusOnArbitraryHeight(&_ValidatorPool.TransactOpts, aliceNetHeight_)
}

// PauseConsensusOnArbitraryHeight is a paid mutator transaction binding the contract method 0xbc33bb01.
//
// Solidity: function pauseConsensusOnArbitraryHeight(uint256 aliceNetHeight_) returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) PauseConsensusOnArbitraryHeight(aliceNetHeight_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.Contract.PauseConsensusOnArbitraryHeight(&_ValidatorPool.TransactOpts, aliceNetHeight_)
}

// RegisterValidators is a paid mutator transaction binding the contract method 0x65bd91af.
//
// Solidity: function registerValidators(address[] validators_, uint256[] stakerTokenIDs_) returns()
func (_ValidatorPool *ValidatorPoolTransactor) RegisterValidators(opts *bind.TransactOpts, validators_ []common.Address, stakerTokenIDs_ []*big.Int) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "registerValidators", validators_, stakerTokenIDs_)
}

// RegisterValidators is a paid mutator transaction binding the contract method 0x65bd91af.
//
// Solidity: function registerValidators(address[] validators_, uint256[] stakerTokenIDs_) returns()
func (_ValidatorPool *ValidatorPoolSession) RegisterValidators(validators_ []common.Address, stakerTokenIDs_ []*big.Int) (*types.Transaction, error) {
	return _ValidatorPool.Contract.RegisterValidators(&_ValidatorPool.TransactOpts, validators_, stakerTokenIDs_)
}

// RegisterValidators is a paid mutator transaction binding the contract method 0x65bd91af.
//
// Solidity: function registerValidators(address[] validators_, uint256[] stakerTokenIDs_) returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) RegisterValidators(validators_ []common.Address, stakerTokenIDs_ []*big.Int) (*types.Transaction, error) {
	return _ValidatorPool.Contract.RegisterValidators(&_ValidatorPool.TransactOpts, validators_, stakerTokenIDs_)
}

// ScheduleMaintenance is a paid mutator transaction binding the contract method 0x2380db1a.
//
// Solidity: function scheduleMaintenance() returns()
func (_ValidatorPool *ValidatorPoolTransactor) ScheduleMaintenance(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "scheduleMaintenance")
}

// ScheduleMaintenance is a paid mutator transaction binding the contract method 0x2380db1a.
//
// Solidity: function scheduleMaintenance() returns()
func (_ValidatorPool *ValidatorPoolSession) ScheduleMaintenance() (*types.Transaction, error) {
	return _ValidatorPool.Contract.ScheduleMaintenance(&_ValidatorPool.TransactOpts)
}

// ScheduleMaintenance is a paid mutator transaction binding the contract method 0x2380db1a.
//
// Solidity: function scheduleMaintenance() returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) ScheduleMaintenance() (*types.Transaction, error) {
	return _ValidatorPool.Contract.ScheduleMaintenance(&_ValidatorPool.TransactOpts)
}

// SetDisputerReward is a paid mutator transaction binding the contract method 0x7d907284.
//
// Solidity: function setDisputerReward(uint256 disputerReward_) returns()
func (_ValidatorPool *ValidatorPoolTransactor) SetDisputerReward(opts *bind.TransactOpts, disputerReward_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "setDisputerReward", disputerReward_)
}

// SetDisputerReward is a paid mutator transaction binding the contract method 0x7d907284.
//
// Solidity: function setDisputerReward(uint256 disputerReward_) returns()
func (_ValidatorPool *ValidatorPoolSession) SetDisputerReward(disputerReward_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.Contract.SetDisputerReward(&_ValidatorPool.TransactOpts, disputerReward_)
}

// SetDisputerReward is a paid mutator transaction binding the contract method 0x7d907284.
//
// Solidity: function setDisputerReward(uint256 disputerReward_) returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) SetDisputerReward(disputerReward_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.Contract.SetDisputerReward(&_ValidatorPool.TransactOpts, disputerReward_)
}

// SetLocation is a paid mutator transaction binding the contract method 0x827bfbdf.
//
// Solidity: function setLocation(string ip_) returns()
func (_ValidatorPool *ValidatorPoolTransactor) SetLocation(opts *bind.TransactOpts, ip_ string) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "setLocation", ip_)
}

// SetLocation is a paid mutator transaction binding the contract method 0x827bfbdf.
//
// Solidity: function setLocation(string ip_) returns()
func (_ValidatorPool *ValidatorPoolSession) SetLocation(ip_ string) (*types.Transaction, error) {
	return _ValidatorPool.Contract.SetLocation(&_ValidatorPool.TransactOpts, ip_)
}

// SetLocation is a paid mutator transaction binding the contract method 0x827bfbdf.
//
// Solidity: function setLocation(string ip_) returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) SetLocation(ip_ string) (*types.Transaction, error) {
	return _ValidatorPool.Contract.SetLocation(&_ValidatorPool.TransactOpts, ip_)
}

// SetMaxNumValidators is a paid mutator transaction binding the contract method 0x6c0da0b4.
//
// Solidity: function setMaxNumValidators(uint256 maxNumValidators_) returns()
func (_ValidatorPool *ValidatorPoolTransactor) SetMaxNumValidators(opts *bind.TransactOpts, maxNumValidators_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "setMaxNumValidators", maxNumValidators_)
}

// SetMaxNumValidators is a paid mutator transaction binding the contract method 0x6c0da0b4.
//
// Solidity: function setMaxNumValidators(uint256 maxNumValidators_) returns()
func (_ValidatorPool *ValidatorPoolSession) SetMaxNumValidators(maxNumValidators_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.Contract.SetMaxNumValidators(&_ValidatorPool.TransactOpts, maxNumValidators_)
}

// SetMaxNumValidators is a paid mutator transaction binding the contract method 0x6c0da0b4.
//
// Solidity: function setMaxNumValidators(uint256 maxNumValidators_) returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) SetMaxNumValidators(maxNumValidators_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.Contract.SetMaxNumValidators(&_ValidatorPool.TransactOpts, maxNumValidators_)
}

// SetStakeAmount is a paid mutator transaction binding the contract method 0x43808c50.
//
// Solidity: function setStakeAmount(uint256 stakeAmount_) returns()
func (_ValidatorPool *ValidatorPoolTransactor) SetStakeAmount(opts *bind.TransactOpts, stakeAmount_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "setStakeAmount", stakeAmount_)
}

// SetStakeAmount is a paid mutator transaction binding the contract method 0x43808c50.
//
// Solidity: function setStakeAmount(uint256 stakeAmount_) returns()
func (_ValidatorPool *ValidatorPoolSession) SetStakeAmount(stakeAmount_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.Contract.SetStakeAmount(&_ValidatorPool.TransactOpts, stakeAmount_)
}

// SetStakeAmount is a paid mutator transaction binding the contract method 0x43808c50.
//
// Solidity: function setStakeAmount(uint256 stakeAmount_) returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) SetStakeAmount(stakeAmount_ *big.Int) (*types.Transaction, error) {
	return _ValidatorPool.Contract.SetStakeAmount(&_ValidatorPool.TransactOpts, stakeAmount_)
}

// SkimExcessEth is a paid mutator transaction binding the contract method 0x971b505b.
//
// Solidity: function skimExcessEth(address to_) returns(uint256 excess)
func (_ValidatorPool *ValidatorPoolTransactor) SkimExcessEth(opts *bind.TransactOpts, to_ common.Address) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "skimExcessEth", to_)
}

// SkimExcessEth is a paid mutator transaction binding the contract method 0x971b505b.
//
// Solidity: function skimExcessEth(address to_) returns(uint256 excess)
func (_ValidatorPool *ValidatorPoolSession) SkimExcessEth(to_ common.Address) (*types.Transaction, error) {
	return _ValidatorPool.Contract.SkimExcessEth(&_ValidatorPool.TransactOpts, to_)
}

// SkimExcessEth is a paid mutator transaction binding the contract method 0x971b505b.
//
// Solidity: function skimExcessEth(address to_) returns(uint256 excess)
func (_ValidatorPool *ValidatorPoolTransactorSession) SkimExcessEth(to_ common.Address) (*types.Transaction, error) {
	return _ValidatorPool.Contract.SkimExcessEth(&_ValidatorPool.TransactOpts, to_)
}

// SkimExcessToken is a paid mutator transaction binding the contract method 0x7aa507fb.
//
// Solidity: function skimExcessToken(address to_) returns(uint256 excess)
func (_ValidatorPool *ValidatorPoolTransactor) SkimExcessToken(opts *bind.TransactOpts, to_ common.Address) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "skimExcessToken", to_)
}

// SkimExcessToken is a paid mutator transaction binding the contract method 0x7aa507fb.
//
// Solidity: function skimExcessToken(address to_) returns(uint256 excess)
func (_ValidatorPool *ValidatorPoolSession) SkimExcessToken(to_ common.Address) (*types.Transaction, error) {
	return _ValidatorPool.Contract.SkimExcessToken(&_ValidatorPool.TransactOpts, to_)
}

// SkimExcessToken is a paid mutator transaction binding the contract method 0x7aa507fb.
//
// Solidity: function skimExcessToken(address to_) returns(uint256 excess)
func (_ValidatorPool *ValidatorPoolTransactorSession) SkimExcessToken(to_ common.Address) (*types.Transaction, error) {
	return _ValidatorPool.Contract.SkimExcessToken(&_ValidatorPool.TransactOpts, to_)
}

// UnregisterAllValidators is a paid mutator transaction binding the contract method 0xf6442e24.
//
// Solidity: function unregisterAllValidators() returns()
func (_ValidatorPool *ValidatorPoolTransactor) UnregisterAllValidators(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "unregisterAllValidators")
}

// UnregisterAllValidators is a paid mutator transaction binding the contract method 0xf6442e24.
//
// Solidity: function unregisterAllValidators() returns()
func (_ValidatorPool *ValidatorPoolSession) UnregisterAllValidators() (*types.Transaction, error) {
	return _ValidatorPool.Contract.UnregisterAllValidators(&_ValidatorPool.TransactOpts)
}

// UnregisterAllValidators is a paid mutator transaction binding the contract method 0xf6442e24.
//
// Solidity: function unregisterAllValidators() returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) UnregisterAllValidators() (*types.Transaction, error) {
	return _ValidatorPool.Contract.UnregisterAllValidators(&_ValidatorPool.TransactOpts)
}

// UnregisterValidators is a paid mutator transaction binding the contract method 0xc6e86ad6.
//
// Solidity: function unregisterValidators(address[] validators_) returns()
func (_ValidatorPool *ValidatorPoolTransactor) UnregisterValidators(opts *bind.TransactOpts, validators_ []common.Address) (*types.Transaction, error) {
	return _ValidatorPool.contract.Transact(opts, "unregisterValidators", validators_)
}

// UnregisterValidators is a paid mutator transaction binding the contract method 0xc6e86ad6.
//
// Solidity: function unregisterValidators(address[] validators_) returns()
func (_ValidatorPool *ValidatorPoolSession) UnregisterValidators(validators_ []common.Address) (*types.Transaction, error) {
	return _ValidatorPool.Contract.UnregisterValidators(&_ValidatorPool.TransactOpts, validators_)
}

// UnregisterValidators is a paid mutator transaction binding the contract method 0xc6e86ad6.
//
// Solidity: function unregisterValidators(address[] validators_) returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) UnregisterValidators(validators_ []common.Address) (*types.Transaction, error) {
	return _ValidatorPool.Contract.UnregisterValidators(&_ValidatorPool.TransactOpts, validators_)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_ValidatorPool *ValidatorPoolTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorPool.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_ValidatorPool *ValidatorPoolSession) Receive() (*types.Transaction, error) {
	return _ValidatorPool.Contract.Receive(&_ValidatorPool.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_ValidatorPool *ValidatorPoolTransactorSession) Receive() (*types.Transaction, error) {
	return _ValidatorPool.Contract.Receive(&_ValidatorPool.TransactOpts)
}

// ValidatorPoolMaintenanceScheduledIterator is returned from FilterMaintenanceScheduled and is used to iterate over the raw logs and unpacked data for MaintenanceScheduled events raised by the ValidatorPool contract.
type ValidatorPoolMaintenanceScheduledIterator struct {
	Event *ValidatorPoolMaintenanceScheduled // Event containing the contract specifics and raw log

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
func (it *ValidatorPoolMaintenanceScheduledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorPoolMaintenanceScheduled)
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
		it.Event = new(ValidatorPoolMaintenanceScheduled)
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
func (it *ValidatorPoolMaintenanceScheduledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorPoolMaintenanceScheduledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorPoolMaintenanceScheduled represents a MaintenanceScheduled event raised by the ValidatorPool contract.
type ValidatorPoolMaintenanceScheduled struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterMaintenanceScheduled is a free log retrieval operation binding the contract event 0xc77f315ab4072b428052ff8f369916ce39f7fa7e925613f3e9b28fe383c565c8.
//
// Solidity: event MaintenanceScheduled()
func (_ValidatorPool *ValidatorPoolFilterer) FilterMaintenanceScheduled(opts *bind.FilterOpts) (*ValidatorPoolMaintenanceScheduledIterator, error) {

	logs, sub, err := _ValidatorPool.contract.FilterLogs(opts, "MaintenanceScheduled")
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolMaintenanceScheduledIterator{contract: _ValidatorPool.contract, event: "MaintenanceScheduled", logs: logs, sub: sub}, nil
}

// WatchMaintenanceScheduled is a free log subscription operation binding the contract event 0xc77f315ab4072b428052ff8f369916ce39f7fa7e925613f3e9b28fe383c565c8.
//
// Solidity: event MaintenanceScheduled()
func (_ValidatorPool *ValidatorPoolFilterer) WatchMaintenanceScheduled(opts *bind.WatchOpts, sink chan<- *ValidatorPoolMaintenanceScheduled) (event.Subscription, error) {

	logs, sub, err := _ValidatorPool.contract.WatchLogs(opts, "MaintenanceScheduled")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorPoolMaintenanceScheduled)
				if err := _ValidatorPool.contract.UnpackLog(event, "MaintenanceScheduled", log); err != nil {
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

// ParseMaintenanceScheduled is a log parse operation binding the contract event 0xc77f315ab4072b428052ff8f369916ce39f7fa7e925613f3e9b28fe383c565c8.
//
// Solidity: event MaintenanceScheduled()
func (_ValidatorPool *ValidatorPoolFilterer) ParseMaintenanceScheduled(log types.Log) (*ValidatorPoolMaintenanceScheduled, error) {
	event := new(ValidatorPoolMaintenanceScheduled)
	if err := _ValidatorPool.contract.UnpackLog(event, "MaintenanceScheduled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorPoolValidatorJoinedIterator is returned from FilterValidatorJoined and is used to iterate over the raw logs and unpacked data for ValidatorJoined events raised by the ValidatorPool contract.
type ValidatorPoolValidatorJoinedIterator struct {
	Event *ValidatorPoolValidatorJoined // Event containing the contract specifics and raw log

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
func (it *ValidatorPoolValidatorJoinedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorPoolValidatorJoined)
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
		it.Event = new(ValidatorPoolValidatorJoined)
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
func (it *ValidatorPoolValidatorJoinedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorPoolValidatorJoinedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorPoolValidatorJoined represents a ValidatorJoined event raised by the ValidatorPool contract.
type ValidatorPoolValidatorJoined struct {
	Account                 common.Address
	ValidatorStakingTokenID *big.Int
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterValidatorJoined is a free log retrieval operation binding the contract event 0xe30848520248cd6b60cf19fe62a302a47e2d2c1c147deea1188e471751557a52.
//
// Solidity: event ValidatorJoined(address indexed account, uint256 validatorStakingTokenID)
func (_ValidatorPool *ValidatorPoolFilterer) FilterValidatorJoined(opts *bind.FilterOpts, account []common.Address) (*ValidatorPoolValidatorJoinedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ValidatorPool.contract.FilterLogs(opts, "ValidatorJoined", accountRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolValidatorJoinedIterator{contract: _ValidatorPool.contract, event: "ValidatorJoined", logs: logs, sub: sub}, nil
}

// WatchValidatorJoined is a free log subscription operation binding the contract event 0xe30848520248cd6b60cf19fe62a302a47e2d2c1c147deea1188e471751557a52.
//
// Solidity: event ValidatorJoined(address indexed account, uint256 validatorStakingTokenID)
func (_ValidatorPool *ValidatorPoolFilterer) WatchValidatorJoined(opts *bind.WatchOpts, sink chan<- *ValidatorPoolValidatorJoined, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ValidatorPool.contract.WatchLogs(opts, "ValidatorJoined", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorPoolValidatorJoined)
				if err := _ValidatorPool.contract.UnpackLog(event, "ValidatorJoined", log); err != nil {
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

// ParseValidatorJoined is a log parse operation binding the contract event 0xe30848520248cd6b60cf19fe62a302a47e2d2c1c147deea1188e471751557a52.
//
// Solidity: event ValidatorJoined(address indexed account, uint256 validatorStakingTokenID)
func (_ValidatorPool *ValidatorPoolFilterer) ParseValidatorJoined(log types.Log) (*ValidatorPoolValidatorJoined, error) {
	event := new(ValidatorPoolValidatorJoined)
	if err := _ValidatorPool.contract.UnpackLog(event, "ValidatorJoined", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorPoolValidatorLeftIterator is returned from FilterValidatorLeft and is used to iterate over the raw logs and unpacked data for ValidatorLeft events raised by the ValidatorPool contract.
type ValidatorPoolValidatorLeftIterator struct {
	Event *ValidatorPoolValidatorLeft // Event containing the contract specifics and raw log

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
func (it *ValidatorPoolValidatorLeftIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorPoolValidatorLeft)
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
		it.Event = new(ValidatorPoolValidatorLeft)
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
func (it *ValidatorPoolValidatorLeftIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorPoolValidatorLeftIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorPoolValidatorLeft represents a ValidatorLeft event raised by the ValidatorPool contract.
type ValidatorPoolValidatorLeft struct {
	Account              common.Address
	PublicStakingTokenID *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterValidatorLeft is a free log retrieval operation binding the contract event 0x33ff7b2beda3cb99406d3401fd9e8d9001b93e74b845cf7346f6e7f70c703e73.
//
// Solidity: event ValidatorLeft(address indexed account, uint256 publicStakingTokenID)
func (_ValidatorPool *ValidatorPoolFilterer) FilterValidatorLeft(opts *bind.FilterOpts, account []common.Address) (*ValidatorPoolValidatorLeftIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ValidatorPool.contract.FilterLogs(opts, "ValidatorLeft", accountRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolValidatorLeftIterator{contract: _ValidatorPool.contract, event: "ValidatorLeft", logs: logs, sub: sub}, nil
}

// WatchValidatorLeft is a free log subscription operation binding the contract event 0x33ff7b2beda3cb99406d3401fd9e8d9001b93e74b845cf7346f6e7f70c703e73.
//
// Solidity: event ValidatorLeft(address indexed account, uint256 publicStakingTokenID)
func (_ValidatorPool *ValidatorPoolFilterer) WatchValidatorLeft(opts *bind.WatchOpts, sink chan<- *ValidatorPoolValidatorLeft, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ValidatorPool.contract.WatchLogs(opts, "ValidatorLeft", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorPoolValidatorLeft)
				if err := _ValidatorPool.contract.UnpackLog(event, "ValidatorLeft", log); err != nil {
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

// ParseValidatorLeft is a log parse operation binding the contract event 0x33ff7b2beda3cb99406d3401fd9e8d9001b93e74b845cf7346f6e7f70c703e73.
//
// Solidity: event ValidatorLeft(address indexed account, uint256 publicStakingTokenID)
func (_ValidatorPool *ValidatorPoolFilterer) ParseValidatorLeft(log types.Log) (*ValidatorPoolValidatorLeft, error) {
	event := new(ValidatorPoolValidatorLeft)
	if err := _ValidatorPool.contract.UnpackLog(event, "ValidatorLeft", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorPoolValidatorMajorSlashedIterator is returned from FilterValidatorMajorSlashed and is used to iterate over the raw logs and unpacked data for ValidatorMajorSlashed events raised by the ValidatorPool contract.
type ValidatorPoolValidatorMajorSlashedIterator struct {
	Event *ValidatorPoolValidatorMajorSlashed // Event containing the contract specifics and raw log

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
func (it *ValidatorPoolValidatorMajorSlashedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorPoolValidatorMajorSlashed)
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
		it.Event = new(ValidatorPoolValidatorMajorSlashed)
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
func (it *ValidatorPoolValidatorMajorSlashedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorPoolValidatorMajorSlashedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorPoolValidatorMajorSlashed represents a ValidatorMajorSlashed event raised by the ValidatorPool contract.
type ValidatorPoolValidatorMajorSlashed struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterValidatorMajorSlashed is a free log retrieval operation binding the contract event 0xee806478c61c75fc3ec50328b2af43290d1860ef40d5dfbba62ece0e1e3abe9e.
//
// Solidity: event ValidatorMajorSlashed(address indexed account)
func (_ValidatorPool *ValidatorPoolFilterer) FilterValidatorMajorSlashed(opts *bind.FilterOpts, account []common.Address) (*ValidatorPoolValidatorMajorSlashedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ValidatorPool.contract.FilterLogs(opts, "ValidatorMajorSlashed", accountRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolValidatorMajorSlashedIterator{contract: _ValidatorPool.contract, event: "ValidatorMajorSlashed", logs: logs, sub: sub}, nil
}

// WatchValidatorMajorSlashed is a free log subscription operation binding the contract event 0xee806478c61c75fc3ec50328b2af43290d1860ef40d5dfbba62ece0e1e3abe9e.
//
// Solidity: event ValidatorMajorSlashed(address indexed account)
func (_ValidatorPool *ValidatorPoolFilterer) WatchValidatorMajorSlashed(opts *bind.WatchOpts, sink chan<- *ValidatorPoolValidatorMajorSlashed, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ValidatorPool.contract.WatchLogs(opts, "ValidatorMajorSlashed", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorPoolValidatorMajorSlashed)
				if err := _ValidatorPool.contract.UnpackLog(event, "ValidatorMajorSlashed", log); err != nil {
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

// ParseValidatorMajorSlashed is a log parse operation binding the contract event 0xee806478c61c75fc3ec50328b2af43290d1860ef40d5dfbba62ece0e1e3abe9e.
//
// Solidity: event ValidatorMajorSlashed(address indexed account)
func (_ValidatorPool *ValidatorPoolFilterer) ParseValidatorMajorSlashed(log types.Log) (*ValidatorPoolValidatorMajorSlashed, error) {
	event := new(ValidatorPoolValidatorMajorSlashed)
	if err := _ValidatorPool.contract.UnpackLog(event, "ValidatorMajorSlashed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorPoolValidatorMinorSlashedIterator is returned from FilterValidatorMinorSlashed and is used to iterate over the raw logs and unpacked data for ValidatorMinorSlashed events raised by the ValidatorPool contract.
type ValidatorPoolValidatorMinorSlashedIterator struct {
	Event *ValidatorPoolValidatorMinorSlashed // Event containing the contract specifics and raw log

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
func (it *ValidatorPoolValidatorMinorSlashedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorPoolValidatorMinorSlashed)
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
		it.Event = new(ValidatorPoolValidatorMinorSlashed)
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
func (it *ValidatorPoolValidatorMinorSlashedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorPoolValidatorMinorSlashedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorPoolValidatorMinorSlashed represents a ValidatorMinorSlashed event raised by the ValidatorPool contract.
type ValidatorPoolValidatorMinorSlashed struct {
	Account              common.Address
	PublicStakingTokenID *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterValidatorMinorSlashed is a free log retrieval operation binding the contract event 0x23f67a6ac6d764dca01e28630334f5b636e2b1928c0a5d5b5428da3f69167208.
//
// Solidity: event ValidatorMinorSlashed(address indexed account, uint256 publicStakingTokenID)
func (_ValidatorPool *ValidatorPoolFilterer) FilterValidatorMinorSlashed(opts *bind.FilterOpts, account []common.Address) (*ValidatorPoolValidatorMinorSlashedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ValidatorPool.contract.FilterLogs(opts, "ValidatorMinorSlashed", accountRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorPoolValidatorMinorSlashedIterator{contract: _ValidatorPool.contract, event: "ValidatorMinorSlashed", logs: logs, sub: sub}, nil
}

// WatchValidatorMinorSlashed is a free log subscription operation binding the contract event 0x23f67a6ac6d764dca01e28630334f5b636e2b1928c0a5d5b5428da3f69167208.
//
// Solidity: event ValidatorMinorSlashed(address indexed account, uint256 publicStakingTokenID)
func (_ValidatorPool *ValidatorPoolFilterer) WatchValidatorMinorSlashed(opts *bind.WatchOpts, sink chan<- *ValidatorPoolValidatorMinorSlashed, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ValidatorPool.contract.WatchLogs(opts, "ValidatorMinorSlashed", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorPoolValidatorMinorSlashed)
				if err := _ValidatorPool.contract.UnpackLog(event, "ValidatorMinorSlashed", log); err != nil {
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

// ParseValidatorMinorSlashed is a log parse operation binding the contract event 0x23f67a6ac6d764dca01e28630334f5b636e2b1928c0a5d5b5428da3f69167208.
//
// Solidity: event ValidatorMinorSlashed(address indexed account, uint256 publicStakingTokenID)
func (_ValidatorPool *ValidatorPoolFilterer) ParseValidatorMinorSlashed(log types.Log) (*ValidatorPoolValidatorMinorSlashed, error) {
	event := new(ValidatorPoolValidatorMinorSlashed)
	if err := _ValidatorPool.contract.UnpackLog(event, "ValidatorMinorSlashed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
