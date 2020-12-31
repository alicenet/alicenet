// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package main

import (
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
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// MerkleProofABI is the input ABI used to generate the binding from.
const MerkleProofABI = "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"self\",\"type\":\"uint256\"},{\"internalType\":\"uint16\",\"name\":\"index\",\"type\":\"uint16\"}],\"name\":\"bitSet\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"proof\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"root\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"key\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bitset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"height\",\"type\":\"uint256\"}],\"name\":\"checkProof\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]"

// MerkleProofBin is the compiled bytecode used for deploying new contracts.
var MerkleProofBin = "0x608060405234801561001057600080fd5b50610264806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c8063920006181461003b578063d31a6e8f14610076575b600080fd5b6100626004803603604081101561005157600080fd5b508035906020013561ffff16610133565b604080519115158252519081900360200190f35b610062600480360360c081101561008c57600080fd5b8101906020810181356401000000008111156100a757600080fd5b8201836020820111156100b957600080fd5b803590602001918460018302840111640100000000831117156100db57600080fd5b91908080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152509295505082359350505060208101359060408101359060608101359060800135610144565b60ff0361ffff161c60019081161490565b600080857fbc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a82805b8681101561021d57610181888260ff03610133565b1561019757602082019150818c0151945061019b565b8294505b6101a8898260ff03610133565b156101e35784846040516020018083815260200182815260200192505050604051602081830303815290604052805190602001209350610215565b838560405160200180838152602001828152602001925050506040516020818303038152906040528051906020012093505b60010161016c565b50505090961497965050505050505056fea264697066735822122015b53d3b8377591f3f686e82858a1663fac8efc9267c93aac38e39dc21b6d54864736f6c63430006060033"

// DeployMerkleProof deploys a new Ethereum contract, binding an instance of MerkleProof to it.
func DeployMerkleProof(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *MerkleProof, error) {
	parsed, err := abi.JSON(strings.NewReader(MerkleProofABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(MerkleProofBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &MerkleProof{MerkleProofCaller: MerkleProofCaller{contract: contract}, MerkleProofTransactor: MerkleProofTransactor{contract: contract}, MerkleProofFilterer: MerkleProofFilterer{contract: contract}}, nil
}

// MerkleProof is an auto generated Go binding around an Ethereum contract.
type MerkleProof struct {
	MerkleProofCaller     // Read-only binding to the contract
	MerkleProofTransactor // Write-only binding to the contract
	MerkleProofFilterer   // Log filterer for contract events
}

// MerkleProofCaller is an auto generated read-only Go binding around an Ethereum contract.
type MerkleProofCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MerkleProofTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MerkleProofTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MerkleProofFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MerkleProofFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MerkleProofSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MerkleProofSession struct {
	Contract     *MerkleProof      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MerkleProofCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MerkleProofCallerSession struct {
	Contract *MerkleProofCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// MerkleProofTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MerkleProofTransactorSession struct {
	Contract     *MerkleProofTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// MerkleProofRaw is an auto generated low-level Go binding around an Ethereum contract.
type MerkleProofRaw struct {
	Contract *MerkleProof // Generic contract binding to access the raw methods on
}

// MerkleProofCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MerkleProofCallerRaw struct {
	Contract *MerkleProofCaller // Generic read-only contract binding to access the raw methods on
}

// MerkleProofTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MerkleProofTransactorRaw struct {
	Contract *MerkleProofTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMerkleProof creates a new instance of MerkleProof, bound to a specific deployed contract.
func NewMerkleProof(address common.Address, backend bind.ContractBackend) (*MerkleProof, error) {
	contract, err := bindMerkleProof(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MerkleProof{MerkleProofCaller: MerkleProofCaller{contract: contract}, MerkleProofTransactor: MerkleProofTransactor{contract: contract}, MerkleProofFilterer: MerkleProofFilterer{contract: contract}}, nil
}

// NewMerkleProofCaller creates a new read-only instance of MerkleProof, bound to a specific deployed contract.
func NewMerkleProofCaller(address common.Address, caller bind.ContractCaller) (*MerkleProofCaller, error) {
	contract, err := bindMerkleProof(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MerkleProofCaller{contract: contract}, nil
}

// NewMerkleProofTransactor creates a new write-only instance of MerkleProof, bound to a specific deployed contract.
func NewMerkleProofTransactor(address common.Address, transactor bind.ContractTransactor) (*MerkleProofTransactor, error) {
	contract, err := bindMerkleProof(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MerkleProofTransactor{contract: contract}, nil
}

// NewMerkleProofFilterer creates a new log filterer instance of MerkleProof, bound to a specific deployed contract.
func NewMerkleProofFilterer(address common.Address, filterer bind.ContractFilterer) (*MerkleProofFilterer, error) {
	contract, err := bindMerkleProof(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MerkleProofFilterer{contract: contract}, nil
}

// bindMerkleProof binds a generic wrapper to an already deployed contract.
func bindMerkleProof(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MerkleProofABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MerkleProof *MerkleProofRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MerkleProof.Contract.MerkleProofCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MerkleProof *MerkleProofRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MerkleProof.Contract.MerkleProofTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MerkleProof *MerkleProofRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MerkleProof.Contract.MerkleProofTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MerkleProof *MerkleProofCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _MerkleProof.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MerkleProof *MerkleProofTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MerkleProof.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MerkleProof *MerkleProofTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MerkleProof.Contract.contract.Transact(opts, method, params...)
}

// BitSet is a free data retrieval call binding the contract method 0x92000618.
//
// Solidity: function bitSet(uint256 self, uint16 index) constant returns(bool)
func (_MerkleProof *MerkleProofCaller) BitSet(opts *bind.CallOpts, self *big.Int, index uint16) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _MerkleProof.contract.Call(opts, out, "bitSet", self, index)
	return *ret0, err
}

// BitSet is a free data retrieval call binding the contract method 0x92000618.
//
// Solidity: function bitSet(uint256 self, uint16 index) constant returns(bool)
func (_MerkleProof *MerkleProofSession) BitSet(self *big.Int, index uint16) (bool, error) {
	return _MerkleProof.Contract.BitSet(&_MerkleProof.CallOpts, self, index)
}

// BitSet is a free data retrieval call binding the contract method 0x92000618.
//
// Solidity: function bitSet(uint256 self, uint16 index) constant returns(bool)
func (_MerkleProof *MerkleProofCallerSession) BitSet(self *big.Int, index uint16) (bool, error) {
	return _MerkleProof.Contract.BitSet(&_MerkleProof.CallOpts, self, index)
}

// CheckProof is a free data retrieval call binding the contract method 0xd31a6e8f.
//
// Solidity: function checkProof(bytes proof, bytes32 root, bytes32 hash, uint256 key, uint256 bitset, uint256 height) constant returns(bool)
func (_MerkleProof *MerkleProofCaller) CheckProof(opts *bind.CallOpts, proof []byte, root [32]byte, hash [32]byte, key *big.Int, bitset *big.Int, height *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _MerkleProof.contract.Call(opts, out, "checkProof", proof, root, hash, key, bitset, height)
	return *ret0, err
}

// CheckProof is a free data retrieval call binding the contract method 0xd31a6e8f.
//
// Solidity: function checkProof(bytes proof, bytes32 root, bytes32 hash, uint256 key, uint256 bitset, uint256 height) constant returns(bool)
func (_MerkleProof *MerkleProofSession) CheckProof(proof []byte, root [32]byte, hash [32]byte, key *big.Int, bitset *big.Int, height *big.Int) (bool, error) {
	return _MerkleProof.Contract.CheckProof(&_MerkleProof.CallOpts, proof, root, hash, key, bitset, height)
}

// CheckProof is a free data retrieval call binding the contract method 0xd31a6e8f.
//
// Solidity: function checkProof(bytes proof, bytes32 root, bytes32 hash, uint256 key, uint256 bitset, uint256 height) constant returns(bool)
func (_MerkleProof *MerkleProofCallerSession) CheckProof(proof []byte, root [32]byte, hash [32]byte, key *big.Int, bitset *big.Int, height *big.Int) (bool, error) {
	return _MerkleProof.Contract.CheckProof(&_MerkleProof.CallOpts, proof, root, hash, key, bitset, height)
}
