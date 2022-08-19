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

// ValidatorStakingMetaData contains all meta data concerning the ValidatorStaking contract.
var ValidatorStakingMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"magic\",\"type\":\"uint256\"}],\"name\":\"BadMagic\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"balance\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"reserve\",\"type\":\"uint256\"}],\"name\":\"BalanceLessThanReserve\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"CallerNotTokenOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CircuitBreakerOpened\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FreeAfterTimeNotReached\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"InvalidTokenId\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"LockDurationGreaterThanGovernanceLock\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"LockDurationGreaterThanMintLock\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"LockDurationWithdrawTimeNotReached\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MintAmountExceedsMaximumSupply\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MintAmountZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyAToken\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyFactory\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyGovernance\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyStakingPositionDescriptor\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyValidatorPool\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"slush\",\"type\":\"uint256\"}],\"name\":\"SlushTooLarge\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"approved\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenID_\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"payoutEth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"payoutAToken\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenID_\",\"type\":\"uint256\"}],\"name\":\"burnTo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"payoutEth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"payoutAToken\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"circuitBreakerState\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenID_\",\"type\":\"uint256\"}],\"name\":\"collectEth\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"payout\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenID_\",\"type\":\"uint256\"}],\"name\":\"collectEthTo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"payout\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenID_\",\"type\":\"uint256\"}],\"name\":\"collectToken\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"payout\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenID_\",\"type\":\"uint256\"}],\"name\":\"collectTokenTo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"payout\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"magic_\",\"type\":\"uint8\"}],\"name\":\"depositEth\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"magic_\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"}],\"name\":\"depositToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenID_\",\"type\":\"uint256\"}],\"name\":\"estimateEthCollection\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"payout\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"estimateExcessEth\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"excess\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"estimateExcessToken\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"excess\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenID_\",\"type\":\"uint256\"}],\"name\":\"estimateTokenCollection\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"payout\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAccumulatorScaleFactor\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getApproved\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEthAccumulator\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"accumulator\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"slush\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMaxGovernanceLock\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMaxMintLock\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenID_\",\"type\":\"uint256\"}],\"name\":\"getPosition\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"shares\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"freeAfter\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"withdrawFreeAfter\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"accumulatorEth\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"accumulatorToken\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTokenAccumulator\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"accumulator\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"slush\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTotalReserveAToken\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTotalReserveEth\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getTotalShares\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenID_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockDuration_\",\"type\":\"uint256\"}],\"name\":\"lockOwnPosition\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"caller_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenID_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockDuration_\",\"type\":\"uint256\"}],\"name\":\"lockPosition\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenID_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockDuration_\",\"type\":\"uint256\"}],\"name\":\"lockWithdraw\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenID\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockDuration_\",\"type\":\"uint256\"}],\"name\":\"mintTo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenID\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"}],\"name\":\"skimExcessEth\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"excess\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to_\",\"type\":\"address\"}],\"name\":\"skimExcessToken\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"excess\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenID_\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tripCB\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ValidatorStakingABI is the input ABI used to generate the binding from.
// Deprecated: Use ValidatorStakingMetaData.ABI instead.
var ValidatorStakingABI = ValidatorStakingMetaData.ABI

// ValidatorStaking is an auto generated Go binding around an Ethereum contract.
type ValidatorStaking struct {
	ValidatorStakingCaller     // Read-only binding to the contract
	ValidatorStakingTransactor // Write-only binding to the contract
	ValidatorStakingFilterer   // Log filterer for contract events
}

// ValidatorStakingCaller is an auto generated read-only Go binding around an Ethereum contract.
type ValidatorStakingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorStakingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ValidatorStakingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorStakingFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ValidatorStakingFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ValidatorStakingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ValidatorStakingSession struct {
	Contract     *ValidatorStaking // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ValidatorStakingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ValidatorStakingCallerSession struct {
	Contract *ValidatorStakingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// ValidatorStakingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ValidatorStakingTransactorSession struct {
	Contract     *ValidatorStakingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// ValidatorStakingRaw is an auto generated low-level Go binding around an Ethereum contract.
type ValidatorStakingRaw struct {
	Contract *ValidatorStaking // Generic contract binding to access the raw methods on
}

// ValidatorStakingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ValidatorStakingCallerRaw struct {
	Contract *ValidatorStakingCaller // Generic read-only contract binding to access the raw methods on
}

// ValidatorStakingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ValidatorStakingTransactorRaw struct {
	Contract *ValidatorStakingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewValidatorStaking creates a new instance of ValidatorStaking, bound to a specific deployed contract.
func NewValidatorStaking(address common.Address, backend bind.ContractBackend) (*ValidatorStaking, error) {
	contract, err := bindValidatorStaking(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ValidatorStaking{ValidatorStakingCaller: ValidatorStakingCaller{contract: contract}, ValidatorStakingTransactor: ValidatorStakingTransactor{contract: contract}, ValidatorStakingFilterer: ValidatorStakingFilterer{contract: contract}}, nil
}

// NewValidatorStakingCaller creates a new read-only instance of ValidatorStaking, bound to a specific deployed contract.
func NewValidatorStakingCaller(address common.Address, caller bind.ContractCaller) (*ValidatorStakingCaller, error) {
	contract, err := bindValidatorStaking(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ValidatorStakingCaller{contract: contract}, nil
}

// NewValidatorStakingTransactor creates a new write-only instance of ValidatorStaking, bound to a specific deployed contract.
func NewValidatorStakingTransactor(address common.Address, transactor bind.ContractTransactor) (*ValidatorStakingTransactor, error) {
	contract, err := bindValidatorStaking(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ValidatorStakingTransactor{contract: contract}, nil
}

// NewValidatorStakingFilterer creates a new log filterer instance of ValidatorStaking, bound to a specific deployed contract.
func NewValidatorStakingFilterer(address common.Address, filterer bind.ContractFilterer) (*ValidatorStakingFilterer, error) {
	contract, err := bindValidatorStaking(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ValidatorStakingFilterer{contract: contract}, nil
}

// bindValidatorStaking binds a generic wrapper to an already deployed contract.
func bindValidatorStaking(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ValidatorStakingABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ValidatorStaking *ValidatorStakingRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ValidatorStaking.Contract.ValidatorStakingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ValidatorStaking *ValidatorStakingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.ValidatorStakingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ValidatorStaking *ValidatorStakingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.ValidatorStakingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ValidatorStaking *ValidatorStakingCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ValidatorStaking.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ValidatorStaking *ValidatorStakingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ValidatorStaking *ValidatorStakingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_ValidatorStaking *ValidatorStakingCaller) BalanceOf(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "balanceOf", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_ValidatorStaking *ValidatorStakingSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _ValidatorStaking.Contract.BalanceOf(&_ValidatorStaking.CallOpts, owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address owner) view returns(uint256)
func (_ValidatorStaking *ValidatorStakingCallerSession) BalanceOf(owner common.Address) (*big.Int, error) {
	return _ValidatorStaking.Contract.BalanceOf(&_ValidatorStaking.CallOpts, owner)
}

// CircuitBreakerState is a free data retrieval call binding the contract method 0x89465c62.
//
// Solidity: function circuitBreakerState() view returns(bool)
func (_ValidatorStaking *ValidatorStakingCaller) CircuitBreakerState(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "circuitBreakerState")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CircuitBreakerState is a free data retrieval call binding the contract method 0x89465c62.
//
// Solidity: function circuitBreakerState() view returns(bool)
func (_ValidatorStaking *ValidatorStakingSession) CircuitBreakerState() (bool, error) {
	return _ValidatorStaking.Contract.CircuitBreakerState(&_ValidatorStaking.CallOpts)
}

// CircuitBreakerState is a free data retrieval call binding the contract method 0x89465c62.
//
// Solidity: function circuitBreakerState() view returns(bool)
func (_ValidatorStaking *ValidatorStakingCallerSession) CircuitBreakerState() (bool, error) {
	return _ValidatorStaking.Contract.CircuitBreakerState(&_ValidatorStaking.CallOpts)
}

// EstimateEthCollection is a free data retrieval call binding the contract method 0x20ea0d48.
//
// Solidity: function estimateEthCollection(uint256 tokenID_) view returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingCaller) EstimateEthCollection(opts *bind.CallOpts, tokenID_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "estimateEthCollection", tokenID_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// EstimateEthCollection is a free data retrieval call binding the contract method 0x20ea0d48.
//
// Solidity: function estimateEthCollection(uint256 tokenID_) view returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingSession) EstimateEthCollection(tokenID_ *big.Int) (*big.Int, error) {
	return _ValidatorStaking.Contract.EstimateEthCollection(&_ValidatorStaking.CallOpts, tokenID_)
}

// EstimateEthCollection is a free data retrieval call binding the contract method 0x20ea0d48.
//
// Solidity: function estimateEthCollection(uint256 tokenID_) view returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingCallerSession) EstimateEthCollection(tokenID_ *big.Int) (*big.Int, error) {
	return _ValidatorStaking.Contract.EstimateEthCollection(&_ValidatorStaking.CallOpts, tokenID_)
}

// EstimateExcessEth is a free data retrieval call binding the contract method 0x905953ed.
//
// Solidity: function estimateExcessEth() view returns(uint256 excess)
func (_ValidatorStaking *ValidatorStakingCaller) EstimateExcessEth(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "estimateExcessEth")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// EstimateExcessEth is a free data retrieval call binding the contract method 0x905953ed.
//
// Solidity: function estimateExcessEth() view returns(uint256 excess)
func (_ValidatorStaking *ValidatorStakingSession) EstimateExcessEth() (*big.Int, error) {
	return _ValidatorStaking.Contract.EstimateExcessEth(&_ValidatorStaking.CallOpts)
}

// EstimateExcessEth is a free data retrieval call binding the contract method 0x905953ed.
//
// Solidity: function estimateExcessEth() view returns(uint256 excess)
func (_ValidatorStaking *ValidatorStakingCallerSession) EstimateExcessEth() (*big.Int, error) {
	return _ValidatorStaking.Contract.EstimateExcessEth(&_ValidatorStaking.CallOpts)
}

// EstimateExcessToken is a free data retrieval call binding the contract method 0x3eed3eff.
//
// Solidity: function estimateExcessToken() view returns(uint256 excess)
func (_ValidatorStaking *ValidatorStakingCaller) EstimateExcessToken(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "estimateExcessToken")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// EstimateExcessToken is a free data retrieval call binding the contract method 0x3eed3eff.
//
// Solidity: function estimateExcessToken() view returns(uint256 excess)
func (_ValidatorStaking *ValidatorStakingSession) EstimateExcessToken() (*big.Int, error) {
	return _ValidatorStaking.Contract.EstimateExcessToken(&_ValidatorStaking.CallOpts)
}

// EstimateExcessToken is a free data retrieval call binding the contract method 0x3eed3eff.
//
// Solidity: function estimateExcessToken() view returns(uint256 excess)
func (_ValidatorStaking *ValidatorStakingCallerSession) EstimateExcessToken() (*big.Int, error) {
	return _ValidatorStaking.Contract.EstimateExcessToken(&_ValidatorStaking.CallOpts)
}

// EstimateTokenCollection is a free data retrieval call binding the contract method 0x93c5748d.
//
// Solidity: function estimateTokenCollection(uint256 tokenID_) view returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingCaller) EstimateTokenCollection(opts *bind.CallOpts, tokenID_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "estimateTokenCollection", tokenID_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// EstimateTokenCollection is a free data retrieval call binding the contract method 0x93c5748d.
//
// Solidity: function estimateTokenCollection(uint256 tokenID_) view returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingSession) EstimateTokenCollection(tokenID_ *big.Int) (*big.Int, error) {
	return _ValidatorStaking.Contract.EstimateTokenCollection(&_ValidatorStaking.CallOpts, tokenID_)
}

// EstimateTokenCollection is a free data retrieval call binding the contract method 0x93c5748d.
//
// Solidity: function estimateTokenCollection(uint256 tokenID_) view returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingCallerSession) EstimateTokenCollection(tokenID_ *big.Int) (*big.Int, error) {
	return _ValidatorStaking.Contract.EstimateTokenCollection(&_ValidatorStaking.CallOpts, tokenID_)
}

// GetAccumulatorScaleFactor is a free data retrieval call binding the contract method 0x99785132.
//
// Solidity: function getAccumulatorScaleFactor() pure returns(uint256)
func (_ValidatorStaking *ValidatorStakingCaller) GetAccumulatorScaleFactor(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "getAccumulatorScaleFactor")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAccumulatorScaleFactor is a free data retrieval call binding the contract method 0x99785132.
//
// Solidity: function getAccumulatorScaleFactor() pure returns(uint256)
func (_ValidatorStaking *ValidatorStakingSession) GetAccumulatorScaleFactor() (*big.Int, error) {
	return _ValidatorStaking.Contract.GetAccumulatorScaleFactor(&_ValidatorStaking.CallOpts)
}

// GetAccumulatorScaleFactor is a free data retrieval call binding the contract method 0x99785132.
//
// Solidity: function getAccumulatorScaleFactor() pure returns(uint256)
func (_ValidatorStaking *ValidatorStakingCallerSession) GetAccumulatorScaleFactor() (*big.Int, error) {
	return _ValidatorStaking.Contract.GetAccumulatorScaleFactor(&_ValidatorStaking.CallOpts)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_ValidatorStaking *ValidatorStakingCaller) GetApproved(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "getApproved", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_ValidatorStaking *ValidatorStakingSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _ValidatorStaking.Contract.GetApproved(&_ValidatorStaking.CallOpts, tokenId)
}

// GetApproved is a free data retrieval call binding the contract method 0x081812fc.
//
// Solidity: function getApproved(uint256 tokenId) view returns(address)
func (_ValidatorStaking *ValidatorStakingCallerSession) GetApproved(tokenId *big.Int) (common.Address, error) {
	return _ValidatorStaking.Contract.GetApproved(&_ValidatorStaking.CallOpts, tokenId)
}

// GetEthAccumulator is a free data retrieval call binding the contract method 0x548652d2.
//
// Solidity: function getEthAccumulator() view returns(uint256 accumulator, uint256 slush)
func (_ValidatorStaking *ValidatorStakingCaller) GetEthAccumulator(opts *bind.CallOpts) (struct {
	Accumulator *big.Int
	Slush       *big.Int
}, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "getEthAccumulator")

	outstruct := new(struct {
		Accumulator *big.Int
		Slush       *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Accumulator = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Slush = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetEthAccumulator is a free data retrieval call binding the contract method 0x548652d2.
//
// Solidity: function getEthAccumulator() view returns(uint256 accumulator, uint256 slush)
func (_ValidatorStaking *ValidatorStakingSession) GetEthAccumulator() (struct {
	Accumulator *big.Int
	Slush       *big.Int
}, error) {
	return _ValidatorStaking.Contract.GetEthAccumulator(&_ValidatorStaking.CallOpts)
}

// GetEthAccumulator is a free data retrieval call binding the contract method 0x548652d2.
//
// Solidity: function getEthAccumulator() view returns(uint256 accumulator, uint256 slush)
func (_ValidatorStaking *ValidatorStakingCallerSession) GetEthAccumulator() (struct {
	Accumulator *big.Int
	Slush       *big.Int
}, error) {
	return _ValidatorStaking.Contract.GetEthAccumulator(&_ValidatorStaking.CallOpts)
}

// GetMaxGovernanceLock is a free data retrieval call binding the contract method 0xf44d258b.
//
// Solidity: function getMaxGovernanceLock() pure returns(uint256)
func (_ValidatorStaking *ValidatorStakingCaller) GetMaxGovernanceLock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "getMaxGovernanceLock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMaxGovernanceLock is a free data retrieval call binding the contract method 0xf44d258b.
//
// Solidity: function getMaxGovernanceLock() pure returns(uint256)
func (_ValidatorStaking *ValidatorStakingSession) GetMaxGovernanceLock() (*big.Int, error) {
	return _ValidatorStaking.Contract.GetMaxGovernanceLock(&_ValidatorStaking.CallOpts)
}

// GetMaxGovernanceLock is a free data retrieval call binding the contract method 0xf44d258b.
//
// Solidity: function getMaxGovernanceLock() pure returns(uint256)
func (_ValidatorStaking *ValidatorStakingCallerSession) GetMaxGovernanceLock() (*big.Int, error) {
	return _ValidatorStaking.Contract.GetMaxGovernanceLock(&_ValidatorStaking.CallOpts)
}

// GetMaxMintLock is a free data retrieval call binding the contract method 0x090f70f0.
//
// Solidity: function getMaxMintLock() pure returns(uint256)
func (_ValidatorStaking *ValidatorStakingCaller) GetMaxMintLock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "getMaxMintLock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMaxMintLock is a free data retrieval call binding the contract method 0x090f70f0.
//
// Solidity: function getMaxMintLock() pure returns(uint256)
func (_ValidatorStaking *ValidatorStakingSession) GetMaxMintLock() (*big.Int, error) {
	return _ValidatorStaking.Contract.GetMaxMintLock(&_ValidatorStaking.CallOpts)
}

// GetMaxMintLock is a free data retrieval call binding the contract method 0x090f70f0.
//
// Solidity: function getMaxMintLock() pure returns(uint256)
func (_ValidatorStaking *ValidatorStakingCallerSession) GetMaxMintLock() (*big.Int, error) {
	return _ValidatorStaking.Contract.GetMaxMintLock(&_ValidatorStaking.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ValidatorStaking *ValidatorStakingCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ValidatorStaking *ValidatorStakingSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ValidatorStaking.Contract.GetMetamorphicContractAddress(&_ValidatorStaking.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_ValidatorStaking *ValidatorStakingCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _ValidatorStaking.Contract.GetMetamorphicContractAddress(&_ValidatorStaking.CallOpts, _salt, _factory)
}

// GetPosition is a free data retrieval call binding the contract method 0xeb02c301.
//
// Solidity: function getPosition(uint256 tokenID_) view returns(uint256 shares, uint256 freeAfter, uint256 withdrawFreeAfter, uint256 accumulatorEth, uint256 accumulatorToken)
func (_ValidatorStaking *ValidatorStakingCaller) GetPosition(opts *bind.CallOpts, tokenID_ *big.Int) (struct {
	Shares            *big.Int
	FreeAfter         *big.Int
	WithdrawFreeAfter *big.Int
	AccumulatorEth    *big.Int
	AccumulatorToken  *big.Int
}, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "getPosition", tokenID_)

	outstruct := new(struct {
		Shares            *big.Int
		FreeAfter         *big.Int
		WithdrawFreeAfter *big.Int
		AccumulatorEth    *big.Int
		AccumulatorToken  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Shares = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.FreeAfter = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.WithdrawFreeAfter = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.AccumulatorEth = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.AccumulatorToken = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetPosition is a free data retrieval call binding the contract method 0xeb02c301.
//
// Solidity: function getPosition(uint256 tokenID_) view returns(uint256 shares, uint256 freeAfter, uint256 withdrawFreeAfter, uint256 accumulatorEth, uint256 accumulatorToken)
func (_ValidatorStaking *ValidatorStakingSession) GetPosition(tokenID_ *big.Int) (struct {
	Shares            *big.Int
	FreeAfter         *big.Int
	WithdrawFreeAfter *big.Int
	AccumulatorEth    *big.Int
	AccumulatorToken  *big.Int
}, error) {
	return _ValidatorStaking.Contract.GetPosition(&_ValidatorStaking.CallOpts, tokenID_)
}

// GetPosition is a free data retrieval call binding the contract method 0xeb02c301.
//
// Solidity: function getPosition(uint256 tokenID_) view returns(uint256 shares, uint256 freeAfter, uint256 withdrawFreeAfter, uint256 accumulatorEth, uint256 accumulatorToken)
func (_ValidatorStaking *ValidatorStakingCallerSession) GetPosition(tokenID_ *big.Int) (struct {
	Shares            *big.Int
	FreeAfter         *big.Int
	WithdrawFreeAfter *big.Int
	AccumulatorEth    *big.Int
	AccumulatorToken  *big.Int
}, error) {
	return _ValidatorStaking.Contract.GetPosition(&_ValidatorStaking.CallOpts, tokenID_)
}

// GetTokenAccumulator is a free data retrieval call binding the contract method 0xc47c6e14.
//
// Solidity: function getTokenAccumulator() view returns(uint256 accumulator, uint256 slush)
func (_ValidatorStaking *ValidatorStakingCaller) GetTokenAccumulator(opts *bind.CallOpts) (struct {
	Accumulator *big.Int
	Slush       *big.Int
}, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "getTokenAccumulator")

	outstruct := new(struct {
		Accumulator *big.Int
		Slush       *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Accumulator = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Slush = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetTokenAccumulator is a free data retrieval call binding the contract method 0xc47c6e14.
//
// Solidity: function getTokenAccumulator() view returns(uint256 accumulator, uint256 slush)
func (_ValidatorStaking *ValidatorStakingSession) GetTokenAccumulator() (struct {
	Accumulator *big.Int
	Slush       *big.Int
}, error) {
	return _ValidatorStaking.Contract.GetTokenAccumulator(&_ValidatorStaking.CallOpts)
}

// GetTokenAccumulator is a free data retrieval call binding the contract method 0xc47c6e14.
//
// Solidity: function getTokenAccumulator() view returns(uint256 accumulator, uint256 slush)
func (_ValidatorStaking *ValidatorStakingCallerSession) GetTokenAccumulator() (struct {
	Accumulator *big.Int
	Slush       *big.Int
}, error) {
	return _ValidatorStaking.Contract.GetTokenAccumulator(&_ValidatorStaking.CallOpts)
}

// GetTotalReserveAToken is a free data retrieval call binding the contract method 0x856de8d2.
//
// Solidity: function getTotalReserveAToken() view returns(uint256)
func (_ValidatorStaking *ValidatorStakingCaller) GetTotalReserveAToken(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "getTotalReserveAToken")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalReserveAToken is a free data retrieval call binding the contract method 0x856de8d2.
//
// Solidity: function getTotalReserveAToken() view returns(uint256)
func (_ValidatorStaking *ValidatorStakingSession) GetTotalReserveAToken() (*big.Int, error) {
	return _ValidatorStaking.Contract.GetTotalReserveAToken(&_ValidatorStaking.CallOpts)
}

// GetTotalReserveAToken is a free data retrieval call binding the contract method 0x856de8d2.
//
// Solidity: function getTotalReserveAToken() view returns(uint256)
func (_ValidatorStaking *ValidatorStakingCallerSession) GetTotalReserveAToken() (*big.Int, error) {
	return _ValidatorStaking.Contract.GetTotalReserveAToken(&_ValidatorStaking.CallOpts)
}

// GetTotalReserveEth is a free data retrieval call binding the contract method 0x19b8be2f.
//
// Solidity: function getTotalReserveEth() view returns(uint256)
func (_ValidatorStaking *ValidatorStakingCaller) GetTotalReserveEth(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "getTotalReserveEth")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalReserveEth is a free data retrieval call binding the contract method 0x19b8be2f.
//
// Solidity: function getTotalReserveEth() view returns(uint256)
func (_ValidatorStaking *ValidatorStakingSession) GetTotalReserveEth() (*big.Int, error) {
	return _ValidatorStaking.Contract.GetTotalReserveEth(&_ValidatorStaking.CallOpts)
}

// GetTotalReserveEth is a free data retrieval call binding the contract method 0x19b8be2f.
//
// Solidity: function getTotalReserveEth() view returns(uint256)
func (_ValidatorStaking *ValidatorStakingCallerSession) GetTotalReserveEth() (*big.Int, error) {
	return _ValidatorStaking.Contract.GetTotalReserveEth(&_ValidatorStaking.CallOpts)
}

// GetTotalShares is a free data retrieval call binding the contract method 0xd5002f2e.
//
// Solidity: function getTotalShares() view returns(uint256)
func (_ValidatorStaking *ValidatorStakingCaller) GetTotalShares(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "getTotalShares")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalShares is a free data retrieval call binding the contract method 0xd5002f2e.
//
// Solidity: function getTotalShares() view returns(uint256)
func (_ValidatorStaking *ValidatorStakingSession) GetTotalShares() (*big.Int, error) {
	return _ValidatorStaking.Contract.GetTotalShares(&_ValidatorStaking.CallOpts)
}

// GetTotalShares is a free data retrieval call binding the contract method 0xd5002f2e.
//
// Solidity: function getTotalShares() view returns(uint256)
func (_ValidatorStaking *ValidatorStakingCallerSession) GetTotalShares() (*big.Int, error) {
	return _ValidatorStaking.Contract.GetTotalShares(&_ValidatorStaking.CallOpts)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_ValidatorStaking *ValidatorStakingCaller) IsApprovedForAll(opts *bind.CallOpts, owner common.Address, operator common.Address) (bool, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "isApprovedForAll", owner, operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_ValidatorStaking *ValidatorStakingSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _ValidatorStaking.Contract.IsApprovedForAll(&_ValidatorStaking.CallOpts, owner, operator)
}

// IsApprovedForAll is a free data retrieval call binding the contract method 0xe985e9c5.
//
// Solidity: function isApprovedForAll(address owner, address operator) view returns(bool)
func (_ValidatorStaking *ValidatorStakingCallerSession) IsApprovedForAll(owner common.Address, operator common.Address) (bool, error) {
	return _ValidatorStaking.Contract.IsApprovedForAll(&_ValidatorStaking.CallOpts, owner, operator)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ValidatorStaking *ValidatorStakingCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ValidatorStaking *ValidatorStakingSession) Name() (string, error) {
	return _ValidatorStaking.Contract.Name(&_ValidatorStaking.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_ValidatorStaking *ValidatorStakingCallerSession) Name() (string, error) {
	return _ValidatorStaking.Contract.Name(&_ValidatorStaking.CallOpts)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_ValidatorStaking *ValidatorStakingCaller) OwnerOf(opts *bind.CallOpts, tokenId *big.Int) (common.Address, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "ownerOf", tokenId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_ValidatorStaking *ValidatorStakingSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _ValidatorStaking.Contract.OwnerOf(&_ValidatorStaking.CallOpts, tokenId)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 tokenId) view returns(address)
func (_ValidatorStaking *ValidatorStakingCallerSession) OwnerOf(tokenId *big.Int) (common.Address, error) {
	return _ValidatorStaking.Contract.OwnerOf(&_ValidatorStaking.CallOpts, tokenId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_ValidatorStaking *ValidatorStakingCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_ValidatorStaking *ValidatorStakingSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _ValidatorStaking.Contract.SupportsInterface(&_ValidatorStaking.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_ValidatorStaking *ValidatorStakingCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _ValidatorStaking.Contract.SupportsInterface(&_ValidatorStaking.CallOpts, interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ValidatorStaking *ValidatorStakingCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ValidatorStaking *ValidatorStakingSession) Symbol() (string, error) {
	return _ValidatorStaking.Contract.Symbol(&_ValidatorStaking.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_ValidatorStaking *ValidatorStakingCallerSession) Symbol() (string, error) {
	return _ValidatorStaking.Contract.Symbol(&_ValidatorStaking.CallOpts)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenID_) view returns(string)
func (_ValidatorStaking *ValidatorStakingCaller) TokenURI(opts *bind.CallOpts, tokenID_ *big.Int) (string, error) {
	var out []interface{}
	err := _ValidatorStaking.contract.Call(opts, &out, "tokenURI", tokenID_)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenID_) view returns(string)
func (_ValidatorStaking *ValidatorStakingSession) TokenURI(tokenID_ *big.Int) (string, error) {
	return _ValidatorStaking.Contract.TokenURI(&_ValidatorStaking.CallOpts, tokenID_)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 tokenID_) view returns(string)
func (_ValidatorStaking *ValidatorStakingCallerSession) TokenURI(tokenID_ *big.Int) (string, error) {
	return _ValidatorStaking.Contract.TokenURI(&_ValidatorStaking.CallOpts, tokenID_)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_ValidatorStaking *ValidatorStakingTransactor) Approve(opts *bind.TransactOpts, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "approve", to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_ValidatorStaking *ValidatorStakingSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.Approve(&_ValidatorStaking.TransactOpts, to, tokenId)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address to, uint256 tokenId) returns()
func (_ValidatorStaking *ValidatorStakingTransactorSession) Approve(to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.Approve(&_ValidatorStaking.TransactOpts, to, tokenId)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 tokenID_) returns(uint256 payoutEth, uint256 payoutAToken)
func (_ValidatorStaking *ValidatorStakingTransactor) Burn(opts *bind.TransactOpts, tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "burn", tokenID_)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 tokenID_) returns(uint256 payoutEth, uint256 payoutAToken)
func (_ValidatorStaking *ValidatorStakingSession) Burn(tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.Burn(&_ValidatorStaking.TransactOpts, tokenID_)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 tokenID_) returns(uint256 payoutEth, uint256 payoutAToken)
func (_ValidatorStaking *ValidatorStakingTransactorSession) Burn(tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.Burn(&_ValidatorStaking.TransactOpts, tokenID_)
}

// BurnTo is a paid mutator transaction binding the contract method 0xea785a5e.
//
// Solidity: function burnTo(address to_, uint256 tokenID_) returns(uint256 payoutEth, uint256 payoutAToken)
func (_ValidatorStaking *ValidatorStakingTransactor) BurnTo(opts *bind.TransactOpts, to_ common.Address, tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "burnTo", to_, tokenID_)
}

// BurnTo is a paid mutator transaction binding the contract method 0xea785a5e.
//
// Solidity: function burnTo(address to_, uint256 tokenID_) returns(uint256 payoutEth, uint256 payoutAToken)
func (_ValidatorStaking *ValidatorStakingSession) BurnTo(to_ common.Address, tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.BurnTo(&_ValidatorStaking.TransactOpts, to_, tokenID_)
}

// BurnTo is a paid mutator transaction binding the contract method 0xea785a5e.
//
// Solidity: function burnTo(address to_, uint256 tokenID_) returns(uint256 payoutEth, uint256 payoutAToken)
func (_ValidatorStaking *ValidatorStakingTransactorSession) BurnTo(to_ common.Address, tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.BurnTo(&_ValidatorStaking.TransactOpts, to_, tokenID_)
}

// CollectEth is a paid mutator transaction binding the contract method 0x2a0d8bd1.
//
// Solidity: function collectEth(uint256 tokenID_) returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingTransactor) CollectEth(opts *bind.TransactOpts, tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "collectEth", tokenID_)
}

// CollectEth is a paid mutator transaction binding the contract method 0x2a0d8bd1.
//
// Solidity: function collectEth(uint256 tokenID_) returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingSession) CollectEth(tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.CollectEth(&_ValidatorStaking.TransactOpts, tokenID_)
}

// CollectEth is a paid mutator transaction binding the contract method 0x2a0d8bd1.
//
// Solidity: function collectEth(uint256 tokenID_) returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingTransactorSession) CollectEth(tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.CollectEth(&_ValidatorStaking.TransactOpts, tokenID_)
}

// CollectEthTo is a paid mutator transaction binding the contract method 0xbe444379.
//
// Solidity: function collectEthTo(address to_, uint256 tokenID_) returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingTransactor) CollectEthTo(opts *bind.TransactOpts, to_ common.Address, tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "collectEthTo", to_, tokenID_)
}

// CollectEthTo is a paid mutator transaction binding the contract method 0xbe444379.
//
// Solidity: function collectEthTo(address to_, uint256 tokenID_) returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingSession) CollectEthTo(to_ common.Address, tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.CollectEthTo(&_ValidatorStaking.TransactOpts, to_, tokenID_)
}

// CollectEthTo is a paid mutator transaction binding the contract method 0xbe444379.
//
// Solidity: function collectEthTo(address to_, uint256 tokenID_) returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingTransactorSession) CollectEthTo(to_ common.Address, tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.CollectEthTo(&_ValidatorStaking.TransactOpts, to_, tokenID_)
}

// CollectToken is a paid mutator transaction binding the contract method 0xe35c3e28.
//
// Solidity: function collectToken(uint256 tokenID_) returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingTransactor) CollectToken(opts *bind.TransactOpts, tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "collectToken", tokenID_)
}

// CollectToken is a paid mutator transaction binding the contract method 0xe35c3e28.
//
// Solidity: function collectToken(uint256 tokenID_) returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingSession) CollectToken(tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.CollectToken(&_ValidatorStaking.TransactOpts, tokenID_)
}

// CollectToken is a paid mutator transaction binding the contract method 0xe35c3e28.
//
// Solidity: function collectToken(uint256 tokenID_) returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingTransactorSession) CollectToken(tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.CollectToken(&_ValidatorStaking.TransactOpts, tokenID_)
}

// CollectTokenTo is a paid mutator transaction binding the contract method 0x8853b950.
//
// Solidity: function collectTokenTo(address to_, uint256 tokenID_) returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingTransactor) CollectTokenTo(opts *bind.TransactOpts, to_ common.Address, tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "collectTokenTo", to_, tokenID_)
}

// CollectTokenTo is a paid mutator transaction binding the contract method 0x8853b950.
//
// Solidity: function collectTokenTo(address to_, uint256 tokenID_) returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingSession) CollectTokenTo(to_ common.Address, tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.CollectTokenTo(&_ValidatorStaking.TransactOpts, to_, tokenID_)
}

// CollectTokenTo is a paid mutator transaction binding the contract method 0x8853b950.
//
// Solidity: function collectTokenTo(address to_, uint256 tokenID_) returns(uint256 payout)
func (_ValidatorStaking *ValidatorStakingTransactorSession) CollectTokenTo(to_ common.Address, tokenID_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.CollectTokenTo(&_ValidatorStaking.TransactOpts, to_, tokenID_)
}

// DepositEth is a paid mutator transaction binding the contract method 0x99a89ecc.
//
// Solidity: function depositEth(uint8 magic_) payable returns()
func (_ValidatorStaking *ValidatorStakingTransactor) DepositEth(opts *bind.TransactOpts, magic_ uint8) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "depositEth", magic_)
}

// DepositEth is a paid mutator transaction binding the contract method 0x99a89ecc.
//
// Solidity: function depositEth(uint8 magic_) payable returns()
func (_ValidatorStaking *ValidatorStakingSession) DepositEth(magic_ uint8) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.DepositEth(&_ValidatorStaking.TransactOpts, magic_)
}

// DepositEth is a paid mutator transaction binding the contract method 0x99a89ecc.
//
// Solidity: function depositEth(uint8 magic_) payable returns()
func (_ValidatorStaking *ValidatorStakingTransactorSession) DepositEth(magic_ uint8) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.DepositEth(&_ValidatorStaking.TransactOpts, magic_)
}

// DepositToken is a paid mutator transaction binding the contract method 0x8191f5e5.
//
// Solidity: function depositToken(uint8 magic_, uint256 amount_) returns()
func (_ValidatorStaking *ValidatorStakingTransactor) DepositToken(opts *bind.TransactOpts, magic_ uint8, amount_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "depositToken", magic_, amount_)
}

// DepositToken is a paid mutator transaction binding the contract method 0x8191f5e5.
//
// Solidity: function depositToken(uint8 magic_, uint256 amount_) returns()
func (_ValidatorStaking *ValidatorStakingSession) DepositToken(magic_ uint8, amount_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.DepositToken(&_ValidatorStaking.TransactOpts, magic_, amount_)
}

// DepositToken is a paid mutator transaction binding the contract method 0x8191f5e5.
//
// Solidity: function depositToken(uint8 magic_, uint256 amount_) returns()
func (_ValidatorStaking *ValidatorStakingTransactorSession) DepositToken(magic_ uint8, amount_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.DepositToken(&_ValidatorStaking.TransactOpts, magic_, amount_)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_ValidatorStaking *ValidatorStakingTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_ValidatorStaking *ValidatorStakingSession) Initialize() (*types.Transaction, error) {
	return _ValidatorStaking.Contract.Initialize(&_ValidatorStaking.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_ValidatorStaking *ValidatorStakingTransactorSession) Initialize() (*types.Transaction, error) {
	return _ValidatorStaking.Contract.Initialize(&_ValidatorStaking.TransactOpts)
}

// LockOwnPosition is a paid mutator transaction binding the contract method 0xe42a673c.
//
// Solidity: function lockOwnPosition(uint256 tokenID_, uint256 lockDuration_) returns(uint256)
func (_ValidatorStaking *ValidatorStakingTransactor) LockOwnPosition(opts *bind.TransactOpts, tokenID_ *big.Int, lockDuration_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "lockOwnPosition", tokenID_, lockDuration_)
}

// LockOwnPosition is a paid mutator transaction binding the contract method 0xe42a673c.
//
// Solidity: function lockOwnPosition(uint256 tokenID_, uint256 lockDuration_) returns(uint256)
func (_ValidatorStaking *ValidatorStakingSession) LockOwnPosition(tokenID_ *big.Int, lockDuration_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.LockOwnPosition(&_ValidatorStaking.TransactOpts, tokenID_, lockDuration_)
}

// LockOwnPosition is a paid mutator transaction binding the contract method 0xe42a673c.
//
// Solidity: function lockOwnPosition(uint256 tokenID_, uint256 lockDuration_) returns(uint256)
func (_ValidatorStaking *ValidatorStakingTransactorSession) LockOwnPosition(tokenID_ *big.Int, lockDuration_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.LockOwnPosition(&_ValidatorStaking.TransactOpts, tokenID_, lockDuration_)
}

// LockPosition is a paid mutator transaction binding the contract method 0x0cc65dfb.
//
// Solidity: function lockPosition(address caller_, uint256 tokenID_, uint256 lockDuration_) returns(uint256)
func (_ValidatorStaking *ValidatorStakingTransactor) LockPosition(opts *bind.TransactOpts, caller_ common.Address, tokenID_ *big.Int, lockDuration_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "lockPosition", caller_, tokenID_, lockDuration_)
}

// LockPosition is a paid mutator transaction binding the contract method 0x0cc65dfb.
//
// Solidity: function lockPosition(address caller_, uint256 tokenID_, uint256 lockDuration_) returns(uint256)
func (_ValidatorStaking *ValidatorStakingSession) LockPosition(caller_ common.Address, tokenID_ *big.Int, lockDuration_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.LockPosition(&_ValidatorStaking.TransactOpts, caller_, tokenID_, lockDuration_)
}

// LockPosition is a paid mutator transaction binding the contract method 0x0cc65dfb.
//
// Solidity: function lockPosition(address caller_, uint256 tokenID_, uint256 lockDuration_) returns(uint256)
func (_ValidatorStaking *ValidatorStakingTransactorSession) LockPosition(caller_ common.Address, tokenID_ *big.Int, lockDuration_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.LockPosition(&_ValidatorStaking.TransactOpts, caller_, tokenID_, lockDuration_)
}

// LockWithdraw is a paid mutator transaction binding the contract method 0x0e4eb15b.
//
// Solidity: function lockWithdraw(uint256 tokenID_, uint256 lockDuration_) returns(uint256)
func (_ValidatorStaking *ValidatorStakingTransactor) LockWithdraw(opts *bind.TransactOpts, tokenID_ *big.Int, lockDuration_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "lockWithdraw", tokenID_, lockDuration_)
}

// LockWithdraw is a paid mutator transaction binding the contract method 0x0e4eb15b.
//
// Solidity: function lockWithdraw(uint256 tokenID_, uint256 lockDuration_) returns(uint256)
func (_ValidatorStaking *ValidatorStakingSession) LockWithdraw(tokenID_ *big.Int, lockDuration_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.LockWithdraw(&_ValidatorStaking.TransactOpts, tokenID_, lockDuration_)
}

// LockWithdraw is a paid mutator transaction binding the contract method 0x0e4eb15b.
//
// Solidity: function lockWithdraw(uint256 tokenID_, uint256 lockDuration_) returns(uint256)
func (_ValidatorStaking *ValidatorStakingTransactorSession) LockWithdraw(tokenID_ *big.Int, lockDuration_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.LockWithdraw(&_ValidatorStaking.TransactOpts, tokenID_, lockDuration_)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 amount_) returns(uint256 tokenID)
func (_ValidatorStaking *ValidatorStakingTransactor) Mint(opts *bind.TransactOpts, amount_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "mint", amount_)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 amount_) returns(uint256 tokenID)
func (_ValidatorStaking *ValidatorStakingSession) Mint(amount_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.Mint(&_ValidatorStaking.TransactOpts, amount_)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 amount_) returns(uint256 tokenID)
func (_ValidatorStaking *ValidatorStakingTransactorSession) Mint(amount_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.Mint(&_ValidatorStaking.TransactOpts, amount_)
}

// MintTo is a paid mutator transaction binding the contract method 0x2baf2acb.
//
// Solidity: function mintTo(address to_, uint256 amount_, uint256 lockDuration_) returns(uint256 tokenID)
func (_ValidatorStaking *ValidatorStakingTransactor) MintTo(opts *bind.TransactOpts, to_ common.Address, amount_ *big.Int, lockDuration_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "mintTo", to_, amount_, lockDuration_)
}

// MintTo is a paid mutator transaction binding the contract method 0x2baf2acb.
//
// Solidity: function mintTo(address to_, uint256 amount_, uint256 lockDuration_) returns(uint256 tokenID)
func (_ValidatorStaking *ValidatorStakingSession) MintTo(to_ common.Address, amount_ *big.Int, lockDuration_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.MintTo(&_ValidatorStaking.TransactOpts, to_, amount_, lockDuration_)
}

// MintTo is a paid mutator transaction binding the contract method 0x2baf2acb.
//
// Solidity: function mintTo(address to_, uint256 amount_, uint256 lockDuration_) returns(uint256 tokenID)
func (_ValidatorStaking *ValidatorStakingTransactorSession) MintTo(to_ common.Address, amount_ *big.Int, lockDuration_ *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.MintTo(&_ValidatorStaking.TransactOpts, to_, amount_, lockDuration_)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_ValidatorStaking *ValidatorStakingTransactor) SafeTransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "safeTransferFrom", from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_ValidatorStaking *ValidatorStakingSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.SafeTransferFrom(&_ValidatorStaking.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom is a paid mutator transaction binding the contract method 0x42842e0e.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId) returns()
func (_ValidatorStaking *ValidatorStakingTransactorSession) SafeTransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.SafeTransferFrom(&_ValidatorStaking.TransactOpts, from, to, tokenId)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_ValidatorStaking *ValidatorStakingTransactor) SafeTransferFrom0(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "safeTransferFrom0", from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_ValidatorStaking *ValidatorStakingSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.SafeTransferFrom0(&_ValidatorStaking.TransactOpts, from, to, tokenId, data)
}

// SafeTransferFrom0 is a paid mutator transaction binding the contract method 0xb88d4fde.
//
// Solidity: function safeTransferFrom(address from, address to, uint256 tokenId, bytes data) returns()
func (_ValidatorStaking *ValidatorStakingTransactorSession) SafeTransferFrom0(from common.Address, to common.Address, tokenId *big.Int, data []byte) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.SafeTransferFrom0(&_ValidatorStaking.TransactOpts, from, to, tokenId, data)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_ValidatorStaking *ValidatorStakingTransactor) SetApprovalForAll(opts *bind.TransactOpts, operator common.Address, approved bool) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "setApprovalForAll", operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_ValidatorStaking *ValidatorStakingSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.SetApprovalForAll(&_ValidatorStaking.TransactOpts, operator, approved)
}

// SetApprovalForAll is a paid mutator transaction binding the contract method 0xa22cb465.
//
// Solidity: function setApprovalForAll(address operator, bool approved) returns()
func (_ValidatorStaking *ValidatorStakingTransactorSession) SetApprovalForAll(operator common.Address, approved bool) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.SetApprovalForAll(&_ValidatorStaking.TransactOpts, operator, approved)
}

// SkimExcessEth is a paid mutator transaction binding the contract method 0x971b505b.
//
// Solidity: function skimExcessEth(address to_) returns(uint256 excess)
func (_ValidatorStaking *ValidatorStakingTransactor) SkimExcessEth(opts *bind.TransactOpts, to_ common.Address) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "skimExcessEth", to_)
}

// SkimExcessEth is a paid mutator transaction binding the contract method 0x971b505b.
//
// Solidity: function skimExcessEth(address to_) returns(uint256 excess)
func (_ValidatorStaking *ValidatorStakingSession) SkimExcessEth(to_ common.Address) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.SkimExcessEth(&_ValidatorStaking.TransactOpts, to_)
}

// SkimExcessEth is a paid mutator transaction binding the contract method 0x971b505b.
//
// Solidity: function skimExcessEth(address to_) returns(uint256 excess)
func (_ValidatorStaking *ValidatorStakingTransactorSession) SkimExcessEth(to_ common.Address) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.SkimExcessEth(&_ValidatorStaking.TransactOpts, to_)
}

// SkimExcessToken is a paid mutator transaction binding the contract method 0x7aa507fb.
//
// Solidity: function skimExcessToken(address to_) returns(uint256 excess)
func (_ValidatorStaking *ValidatorStakingTransactor) SkimExcessToken(opts *bind.TransactOpts, to_ common.Address) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "skimExcessToken", to_)
}

// SkimExcessToken is a paid mutator transaction binding the contract method 0x7aa507fb.
//
// Solidity: function skimExcessToken(address to_) returns(uint256 excess)
func (_ValidatorStaking *ValidatorStakingSession) SkimExcessToken(to_ common.Address) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.SkimExcessToken(&_ValidatorStaking.TransactOpts, to_)
}

// SkimExcessToken is a paid mutator transaction binding the contract method 0x7aa507fb.
//
// Solidity: function skimExcessToken(address to_) returns(uint256 excess)
func (_ValidatorStaking *ValidatorStakingTransactorSession) SkimExcessToken(to_ common.Address) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.SkimExcessToken(&_ValidatorStaking.TransactOpts, to_)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_ValidatorStaking *ValidatorStakingTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_ValidatorStaking *ValidatorStakingSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.TransferFrom(&_ValidatorStaking.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_ValidatorStaking *ValidatorStakingTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _ValidatorStaking.Contract.TransferFrom(&_ValidatorStaking.TransactOpts, from, to, tokenId)
}

// TripCB is a paid mutator transaction binding the contract method 0xadfdc03f.
//
// Solidity: function tripCB() returns()
func (_ValidatorStaking *ValidatorStakingTransactor) TripCB(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ValidatorStaking.contract.Transact(opts, "tripCB")
}

// TripCB is a paid mutator transaction binding the contract method 0xadfdc03f.
//
// Solidity: function tripCB() returns()
func (_ValidatorStaking *ValidatorStakingSession) TripCB() (*types.Transaction, error) {
	return _ValidatorStaking.Contract.TripCB(&_ValidatorStaking.TransactOpts)
}

// TripCB is a paid mutator transaction binding the contract method 0xadfdc03f.
//
// Solidity: function tripCB() returns()
func (_ValidatorStaking *ValidatorStakingTransactorSession) TripCB() (*types.Transaction, error) {
	return _ValidatorStaking.Contract.TripCB(&_ValidatorStaking.TransactOpts)
}

// ValidatorStakingApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the ValidatorStaking contract.
type ValidatorStakingApprovalIterator struct {
	Event *ValidatorStakingApproval // Event containing the contract specifics and raw log

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
func (it *ValidatorStakingApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorStakingApproval)
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
		it.Event = new(ValidatorStakingApproval)
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
func (it *ValidatorStakingApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorStakingApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorStakingApproval represents a Approval event raised by the ValidatorStaking contract.
type ValidatorStakingApproval struct {
	Owner    common.Address
	Approved common.Address
	TokenId  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_ValidatorStaking *ValidatorStakingFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*ValidatorStakingApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _ValidatorStaking.contract.FilterLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorStakingApprovalIterator{contract: _ValidatorStaking.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_ValidatorStaking *ValidatorStakingFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *ValidatorStakingApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var approvedRule []interface{}
	for _, approvedItem := range approved {
		approvedRule = append(approvedRule, approvedItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _ValidatorStaking.contract.WatchLogs(opts, "Approval", ownerRule, approvedRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorStakingApproval)
				if err := _ValidatorStaking.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
func (_ValidatorStaking *ValidatorStakingFilterer) ParseApproval(log types.Log) (*ValidatorStakingApproval, error) {
	event := new(ValidatorStakingApproval)
	if err := _ValidatorStaking.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorStakingApprovalForAllIterator is returned from FilterApprovalForAll and is used to iterate over the raw logs and unpacked data for ApprovalForAll events raised by the ValidatorStaking contract.
type ValidatorStakingApprovalForAllIterator struct {
	Event *ValidatorStakingApprovalForAll // Event containing the contract specifics and raw log

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
func (it *ValidatorStakingApprovalForAllIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorStakingApprovalForAll)
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
		it.Event = new(ValidatorStakingApprovalForAll)
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
func (it *ValidatorStakingApprovalForAllIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorStakingApprovalForAllIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorStakingApprovalForAll represents a ApprovalForAll event raised by the ValidatorStaking contract.
type ValidatorStakingApprovalForAll struct {
	Owner    common.Address
	Operator common.Address
	Approved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_ValidatorStaking *ValidatorStakingFilterer) FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*ValidatorStakingApprovalForAllIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ValidatorStaking.contract.FilterLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorStakingApprovalForAllIterator{contract: _ValidatorStaking.contract, event: "ApprovalForAll", logs: logs, sub: sub}, nil
}

// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_ValidatorStaking *ValidatorStakingFilterer) WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *ValidatorStakingApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}

	logs, sub, err := _ValidatorStaking.contract.WatchLogs(opts, "ApprovalForAll", ownerRule, operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorStakingApprovalForAll)
				if err := _ValidatorStaking.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
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

// ParseApprovalForAll is a log parse operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
//
// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
func (_ValidatorStaking *ValidatorStakingFilterer) ParseApprovalForAll(log types.Log) (*ValidatorStakingApprovalForAll, error) {
	event := new(ValidatorStakingApprovalForAll)
	if err := _ValidatorStaking.contract.UnpackLog(event, "ApprovalForAll", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorStakingInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the ValidatorStaking contract.
type ValidatorStakingInitializedIterator struct {
	Event *ValidatorStakingInitialized // Event containing the contract specifics and raw log

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
func (it *ValidatorStakingInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorStakingInitialized)
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
		it.Event = new(ValidatorStakingInitialized)
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
func (it *ValidatorStakingInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorStakingInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorStakingInitialized represents a Initialized event raised by the ValidatorStaking contract.
type ValidatorStakingInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ValidatorStaking *ValidatorStakingFilterer) FilterInitialized(opts *bind.FilterOpts) (*ValidatorStakingInitializedIterator, error) {

	logs, sub, err := _ValidatorStaking.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ValidatorStakingInitializedIterator{contract: _ValidatorStaking.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ValidatorStaking *ValidatorStakingFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ValidatorStakingInitialized) (event.Subscription, error) {

	logs, sub, err := _ValidatorStaking.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorStakingInitialized)
				if err := _ValidatorStaking.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_ValidatorStaking *ValidatorStakingFilterer) ParseInitialized(log types.Log) (*ValidatorStakingInitialized, error) {
	event := new(ValidatorStakingInitialized)
	if err := _ValidatorStaking.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ValidatorStakingTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the ValidatorStaking contract.
type ValidatorStakingTransferIterator struct {
	Event *ValidatorStakingTransfer // Event containing the contract specifics and raw log

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
func (it *ValidatorStakingTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ValidatorStakingTransfer)
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
		it.Event = new(ValidatorStakingTransfer)
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
func (it *ValidatorStakingTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ValidatorStakingTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ValidatorStakingTransfer represents a Transfer event raised by the ValidatorStaking contract.
type ValidatorStakingTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_ValidatorStaking *ValidatorStakingFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*ValidatorStakingTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _ValidatorStaking.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &ValidatorStakingTransferIterator{contract: _ValidatorStaking.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_ValidatorStaking *ValidatorStakingFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *ValidatorStakingTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _ValidatorStaking.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ValidatorStakingTransfer)
				if err := _ValidatorStaking.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_ValidatorStaking *ValidatorStakingFilterer) ParseTransfer(log types.Log) (*ValidatorStakingTransfer, error) {
	event := new(ValidatorStakingTransfer)
	if err := _ValidatorStaking.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
