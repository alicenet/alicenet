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

// AccusationInvalidTxConsumptionMetaData contains all meta data concerning the AccusationInvalidTxConsumption contract.
var AccusationInvalidTxConsumptionMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"BEUint16OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"BEUint16OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"BooleanOffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"BooleanOffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"Bytes32OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"Bytes32OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"BytesOffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"BytesOffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"bClaimsChainId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pClaimsChainId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"snapshotsChainId\",\"type\":\"uint256\"}],\"name\":\"ChainIdDoesNotMatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ChainIdZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"dataOffset\",\"type\":\"uint256\"}],\"name\":\"DataOffsetOverflow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DataOffsetOverflow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DefaultLeafNotFoundInKeyPath\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"EllipticCurvePairingFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"bClaimsHeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pClaimsHeight\",\"type\":\"uint256\"}],\"name\":\"HeightDeltaShouldBeOne\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"HeightZero\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InclusionZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"bytesLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"requiredBytesLength\",\"type\":\"uint256\"}],\"name\":\"InsufficientBytes\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"keyHeight\",\"type\":\"uint256\"}],\"name\":\"InvalidKeyHeight\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidNonInclusionMerkleProof\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proofHeight\",\"type\":\"uint256\"}],\"name\":\"InvalidProofHeight\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proofSize\",\"type\":\"uint256\"}],\"name\":\"InvalidProofMinimumSize\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proofSize\",\"type\":\"uint256\"}],\"name\":\"InvalidProofSize\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"signatureVersion\",\"type\":\"uint8\"}],\"name\":\"InvalidSignatureVersion\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"LEUint16OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"LEUint16OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"LEUint256OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"LEUint256OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"proofOfInclusionTxHashKey\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"proofAgainstStateRootKey\",\"type\":\"bytes32\"}],\"name\":\"MerkleProofKeyDoesNotMatchConsumedDepositKey\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"utxoId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"proofAgainstStateRootKey\",\"type\":\"bytes32\"}],\"name\":\"MerkleProofKeyDoesNotMatchUTXOIDBeingSpent\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NoTransactionInAccusedProposal\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"dataOffset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"NotEnoughBytes\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"srcLength\",\"type\":\"uint256\"}],\"name\":\"OffsetOutOfBounds\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"}],\"name\":\"OffsetParameterOverflow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyETHDKG\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyFactory\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlySnapshots\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"}],\"name\":\"OnlyValidatorPool\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProofDoesNotMatchTrieRoot\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProvidedLeafNotFoundInKeyPath\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"RoundZero\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"signatureLength\",\"type\":\"uint256\"}],\"name\":\"SignatureLengthMustBe65Bytes\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SignatureVerificationFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"}],\"name\":\"SignerNotValidValidator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"dataSectionSize\",\"type\":\"uint16\"}],\"name\":\"SizeThresholdExceeded\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"proofAgainstStateRootKey\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"proofOfInclusionTxHashKey\",\"type\":\"bytes32\"}],\"name\":\"UTXODoesnotMatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PRE_SALT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"pClaims_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"pClaimsSig_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"bClaims_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"bClaimsSigGroup_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"txInPreImage_\",\"type\":\"bytes\"},{\"internalType\":\"bytes[3]\",\"name\":\"proofs_\",\"type\":\"bytes[3]\"}],\"name\":\"accuseInvalidTransactionConsumption\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
}

// AccusationInvalidTxConsumptionABI is the input ABI used to generate the binding from.
// Deprecated: Use AccusationInvalidTxConsumptionMetaData.ABI instead.
var AccusationInvalidTxConsumptionABI = AccusationInvalidTxConsumptionMetaData.ABI

// AccusationInvalidTxConsumption is an auto generated Go binding around an Ethereum contract.
type AccusationInvalidTxConsumption struct {
	AccusationInvalidTxConsumptionCaller     // Read-only binding to the contract
	AccusationInvalidTxConsumptionTransactor // Write-only binding to the contract
	AccusationInvalidTxConsumptionFilterer   // Log filterer for contract events
}

// AccusationInvalidTxConsumptionCaller is an auto generated read-only Go binding around an Ethereum contract.
type AccusationInvalidTxConsumptionCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AccusationInvalidTxConsumptionTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AccusationInvalidTxConsumptionTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AccusationInvalidTxConsumptionFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AccusationInvalidTxConsumptionFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AccusationInvalidTxConsumptionSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AccusationInvalidTxConsumptionSession struct {
	Contract     *AccusationInvalidTxConsumption // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                   // Call options to use throughout this session
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// AccusationInvalidTxConsumptionCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AccusationInvalidTxConsumptionCallerSession struct {
	Contract *AccusationInvalidTxConsumptionCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                         // Call options to use throughout this session
}

// AccusationInvalidTxConsumptionTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AccusationInvalidTxConsumptionTransactorSession struct {
	Contract     *AccusationInvalidTxConsumptionTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                         // Transaction auth options to use throughout this session
}

// AccusationInvalidTxConsumptionRaw is an auto generated low-level Go binding around an Ethereum contract.
type AccusationInvalidTxConsumptionRaw struct {
	Contract *AccusationInvalidTxConsumption // Generic contract binding to access the raw methods on
}

// AccusationInvalidTxConsumptionCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AccusationInvalidTxConsumptionCallerRaw struct {
	Contract *AccusationInvalidTxConsumptionCaller // Generic read-only contract binding to access the raw methods on
}

// AccusationInvalidTxConsumptionTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AccusationInvalidTxConsumptionTransactorRaw struct {
	Contract *AccusationInvalidTxConsumptionTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAccusationInvalidTxConsumption creates a new instance of AccusationInvalidTxConsumption, bound to a specific deployed contract.
func NewAccusationInvalidTxConsumption(address common.Address, backend bind.ContractBackend) (*AccusationInvalidTxConsumption, error) {
	contract, err := bindAccusationInvalidTxConsumption(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AccusationInvalidTxConsumption{AccusationInvalidTxConsumptionCaller: AccusationInvalidTxConsumptionCaller{contract: contract}, AccusationInvalidTxConsumptionTransactor: AccusationInvalidTxConsumptionTransactor{contract: contract}, AccusationInvalidTxConsumptionFilterer: AccusationInvalidTxConsumptionFilterer{contract: contract}}, nil
}

// NewAccusationInvalidTxConsumptionCaller creates a new read-only instance of AccusationInvalidTxConsumption, bound to a specific deployed contract.
func NewAccusationInvalidTxConsumptionCaller(address common.Address, caller bind.ContractCaller) (*AccusationInvalidTxConsumptionCaller, error) {
	contract, err := bindAccusationInvalidTxConsumption(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AccusationInvalidTxConsumptionCaller{contract: contract}, nil
}

// NewAccusationInvalidTxConsumptionTransactor creates a new write-only instance of AccusationInvalidTxConsumption, bound to a specific deployed contract.
func NewAccusationInvalidTxConsumptionTransactor(address common.Address, transactor bind.ContractTransactor) (*AccusationInvalidTxConsumptionTransactor, error) {
	contract, err := bindAccusationInvalidTxConsumption(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AccusationInvalidTxConsumptionTransactor{contract: contract}, nil
}

// NewAccusationInvalidTxConsumptionFilterer creates a new log filterer instance of AccusationInvalidTxConsumption, bound to a specific deployed contract.
func NewAccusationInvalidTxConsumptionFilterer(address common.Address, filterer bind.ContractFilterer) (*AccusationInvalidTxConsumptionFilterer, error) {
	contract, err := bindAccusationInvalidTxConsumption(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AccusationInvalidTxConsumptionFilterer{contract: contract}, nil
}

// bindAccusationInvalidTxConsumption binds a generic wrapper to an already deployed contract.
func bindAccusationInvalidTxConsumption(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AccusationInvalidTxConsumptionABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AccusationInvalidTxConsumption.Contract.AccusationInvalidTxConsumptionCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AccusationInvalidTxConsumption.Contract.AccusationInvalidTxConsumptionTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AccusationInvalidTxConsumption.Contract.AccusationInvalidTxConsumptionTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AccusationInvalidTxConsumption.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AccusationInvalidTxConsumption.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AccusationInvalidTxConsumption.Contract.contract.Transact(opts, method, params...)
}

// PRESALT is a free data retrieval call binding the contract method 0x9abd83dc.
//
// Solidity: function PRE_SALT() view returns(bytes32)
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionCaller) PRESALT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AccusationInvalidTxConsumption.contract.Call(opts, &out, "PRE_SALT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PRESALT is a free data retrieval call binding the contract method 0x9abd83dc.
//
// Solidity: function PRE_SALT() view returns(bytes32)
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionSession) PRESALT() ([32]byte, error) {
	return _AccusationInvalidTxConsumption.Contract.PRESALT(&_AccusationInvalidTxConsumption.CallOpts)
}

// PRESALT is a free data retrieval call binding the contract method 0x9abd83dc.
//
// Solidity: function PRE_SALT() view returns(bytes32)
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionCallerSession) PRESALT() ([32]byte, error) {
	return _AccusationInvalidTxConsumption.Contract.PRESALT(&_AccusationInvalidTxConsumption.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _AccusationInvalidTxConsumption.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _AccusationInvalidTxConsumption.Contract.GetMetamorphicContractAddress(&_AccusationInvalidTxConsumption.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _AccusationInvalidTxConsumption.Contract.GetMetamorphicContractAddress(&_AccusationInvalidTxConsumption.CallOpts, _salt, _factory)
}

// AccuseInvalidTransactionConsumption is a paid mutator transaction binding the contract method 0x6ae40457.
//
// Solidity: function accuseInvalidTransactionConsumption(bytes pClaims_, bytes pClaimsSig_, bytes bClaims_, bytes bClaimsSigGroup_, bytes txInPreImage_, bytes[3] proofs_) returns(address)
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionTransactor) AccuseInvalidTransactionConsumption(opts *bind.TransactOpts, pClaims_ []byte, pClaimsSig_ []byte, bClaims_ []byte, bClaimsSigGroup_ []byte, txInPreImage_ []byte, proofs_ [3][]byte) (*types.Transaction, error) {
	return _AccusationInvalidTxConsumption.contract.Transact(opts, "accuseInvalidTransactionConsumption", pClaims_, pClaimsSig_, bClaims_, bClaimsSigGroup_, txInPreImage_, proofs_)
}

// AccuseInvalidTransactionConsumption is a paid mutator transaction binding the contract method 0x6ae40457.
//
// Solidity: function accuseInvalidTransactionConsumption(bytes pClaims_, bytes pClaimsSig_, bytes bClaims_, bytes bClaimsSigGroup_, bytes txInPreImage_, bytes[3] proofs_) returns(address)
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionSession) AccuseInvalidTransactionConsumption(pClaims_ []byte, pClaimsSig_ []byte, bClaims_ []byte, bClaimsSigGroup_ []byte, txInPreImage_ []byte, proofs_ [3][]byte) (*types.Transaction, error) {
	return _AccusationInvalidTxConsumption.Contract.AccuseInvalidTransactionConsumption(&_AccusationInvalidTxConsumption.TransactOpts, pClaims_, pClaimsSig_, bClaims_, bClaimsSigGroup_, txInPreImage_, proofs_)
}

// AccuseInvalidTransactionConsumption is a paid mutator transaction binding the contract method 0x6ae40457.
//
// Solidity: function accuseInvalidTransactionConsumption(bytes pClaims_, bytes pClaimsSig_, bytes bClaims_, bytes bClaimsSigGroup_, bytes txInPreImage_, bytes[3] proofs_) returns(address)
func (_AccusationInvalidTxConsumption *AccusationInvalidTxConsumptionTransactorSession) AccuseInvalidTransactionConsumption(pClaims_ []byte, pClaimsSig_ []byte, bClaims_ []byte, bClaimsSigGroup_ []byte, txInPreImage_ []byte, proofs_ [3][]byte) (*types.Transaction, error) {
	return _AccusationInvalidTxConsumption.Contract.AccuseInvalidTransactionConsumption(&_AccusationInvalidTxConsumption.TransactOpts, pClaims_, pClaimsSig_, bClaims_, bClaimsSigGroup_, txInPreImage_, proofs_)
}
