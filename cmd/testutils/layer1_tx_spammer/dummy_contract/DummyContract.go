// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package dummy_contract

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

// DummyContractMetaData contains all meta data concerning the DummyContract contract.
var DummyContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_DISTRIBUTED_SHARES_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_HAS_COMMITMENTS\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_NOT_PARTICIPATING_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_NOT_VALIDATOR\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_PARTICIPATING_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ACCUSED_SUBMITTED_SHARES_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_DISPUTER_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_DISPUTER_NOT_PARTICIPATING_IN_ROUND\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_INVALID_KEY_OR_PROOF\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_MIN_VALIDATORS_NOT_MET\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_DISPUTE_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_POST_GPKJ_SUBMISSION_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_POST_KEYSHARE_SUBMISSION_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_POST_REGISTRATION_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_NOT_IN_POST_SHARED_DISTRIBUTION_PHASE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_ONLY_VALIDATORS_ALLOWED\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_SHARES_AND_COMMITMENTS_MISMATCH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ETHDKG_VARIABLE_CANNOT_BE_SET_WHILE_RUNNING\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"dummy\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_salt\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"_factory\",\"type\":\"address\"}],\"name\":\"getMetamorphicContractAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6101406040523480156200001257600080fd5b50338073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff1681525050506200008d7f41546f6b656e000000000000000000000000000000000000000000000000000060001b62000081620002a460201b60201c565b620002ae60201b60201c565b73ffffffffffffffffffffffffffffffffffffffff1660a08173ffffffffffffffffffffffffffffffffffffffff1681525050620001047f42546f6b656e000000000000000000000000000000000000000000000000000060001b620000f8620002a460201b60201c565b620002ae60201b60201c565b73ffffffffffffffffffffffffffffffffffffffff1660c08173ffffffffffffffffffffffffffffffffffffffff16815250506200017b7f5075626c69635374616b696e670000000000000000000000000000000000000060001b6200016f620002a460201b60201c565b620002ae60201b60201c565b73ffffffffffffffffffffffffffffffffffffffff1660e08173ffffffffffffffffffffffffffffffffffffffff1681525050620001f27f56616c696461746f72506f6f6c0000000000000000000000000000000000000060001b620001e6620002a460201b60201c565b620002ae60201b60201c565b73ffffffffffffffffffffffffffffffffffffffff166101008173ffffffffffffffffffffffffffffffffffffffff16815250506200026a7f455448444b47000000000000000000000000000000000000000000000000000060001b6200025e620002a460201b60201c565b620002ae60201b60201c565b73ffffffffffffffffffffffffffffffffffffffff166101208173ffffffffffffffffffffffffffffffffffffffff16815250506200046e565b6000608051905090565b6000807f1c0bf703a3415cada9785e89e9d70314c3111ae7d8e04f33bb42eb1d264088be60001b9050828482604051602001620002ee939291906200041e565b6040516020818303038152906040528051906020012060001c91505092915050565b600081905092915050565b7fff00000000000000000000000000000000000000000000000000000000000000600082015250565b60006200035360018362000310565b915062000360826200031b565b600182019050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600062000398826200036b565b9050919050565b60008160601b9050919050565b6000620003b9826200039f565b9050919050565b6000620003cd82620003ac565b9050919050565b620003e9620003e3826200038b565b620003c0565b82525050565b6000819050919050565b6000819050919050565b620004186200041282620003ef565b620003f9565b82525050565b60006200042b8262000344565b9150620004398286620003d4565b6014820191506200044b828562000403565b6020820191506200045d828462000403565b602082019150819050949350505050565b60805160a05160c05160e0516101005161012051610d92620004b26000396000505060005050600050506000505060006108a10152600061046e0152610d926000f3fe608060405234801561001057600080fd5b50600436106101425760003560e01c80637e9f3983116100b8578063b23b83581161007c578063b23b835814610323578063d65915e214610341578063e0c1afd81461035f578063e11879cc1461037d578063f5f46e731461039b578063fa0f33b9146103b957610142565b80637e9f39831461027b57806383c069e4146102995780638653a465146102b75780638e25d1e1146102e7578063a852713f1461030557610142565b806355b83c561161010a57806355b83c56146101c757806360987646146101e55780636d429ef2146102035780637385db5d14610221578063763df93d1461023f57806379ec82961461025d57610142565b80632838edae1461014757806332e43a111461016557806340c10f191461016f5780634c4cfd751461018b5780634cd291bf146101a9575b600080fd5b61014f6103d7565b60405161015c91906108de565b60405180910390f35b61016d6103fb565b005b61018960048036038101906101849190610992565b61046c565b005b6101936105b5565b6040516101a091906108de565b60405180910390f35b6101b16105d9565b6040516101be91906108de565b60405180910390f35b6101cf6105fd565b6040516101dc91906108de565b60405180910390f35b6101ed610621565b6040516101fa91906108de565b60405180910390f35b61020b610645565b60405161021891906108de565b60405180910390f35b610229610669565b60405161023691906108de565b60405180910390f35b61024761068d565b60405161025491906108de565b60405180910390f35b6102656106b1565b60405161027291906108de565b60405180910390f35b6102836106d5565b60405161029091906108de565b60405180910390f35b6102a16106f9565b6040516102ae91906108de565b60405180910390f35b6102d160048036038101906102cc91906109fe565b61071d565b6040516102de9190610a4d565b60405180910390f35b6102ef61077d565b6040516102fc91906108de565b60405180910390f35b61030d6107a1565b60405161031a91906108de565b60405180910390f35b61032b6107c5565b60405161033891906108de565b60405180910390f35b6103496107e9565b60405161035691906108de565b60405180910390f35b61036761080d565b60405161037491906108de565b60405180910390f35b610385610831565b60405161039291906108de565b60405180910390f35b6103a3610855565b6040516103b091906108de565b60405180910390f35b6103c1610879565b6040516103ce91906108de565b60405180910390f35b7f313038000000000000000000000000000000000000000000000000000000000081565b600073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461046a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161046190610b37565b60405180910390fd5b565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16147f32303030000000000000000000000000000000000000000000000000000000006040516020016104ec9190610b78565b6040516020818303038152906040529061053c576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105339190610c1b565b60405180910390fd5b5061054561089d565b73ffffffffffffffffffffffffffffffffffffffff166399f9889883836040518363ffffffff1660e01b815260040161057f929190610c4c565b600060405180830381600087803b15801561059957600080fd5b505af11580156105ad573d6000803e3d6000fd5b505050505050565b7f313130000000000000000000000000000000000000000000000000000000000081565b7f313039000000000000000000000000000000000000000000000000000000000081565b7f313035000000000000000000000000000000000000000000000000000000000081565b7f313136000000000000000000000000000000000000000000000000000000000081565b7f313138000000000000000000000000000000000000000000000000000000000081565b7f313033000000000000000000000000000000000000000000000000000000000081565b7f313131000000000000000000000000000000000000000000000000000000000081565b7f313031000000000000000000000000000000000000000000000000000000000081565b7f313032000000000000000000000000000000000000000000000000000000000081565b7f313030000000000000000000000000000000000000000000000000000000000081565b6000807f1c0bf703a3415cada9785e89e9d70314c3111ae7d8e04f33bb42eb1d264088be60001b905082848260405160200161075b93929190610d14565b6040516020818303038152906040528051906020012060001c91505092915050565b7f313036000000000000000000000000000000000000000000000000000000000081565b7f313135000000000000000000000000000000000000000000000000000000000081565b7f313137000000000000000000000000000000000000000000000000000000000081565b7f313133000000000000000000000000000000000000000000000000000000000081565b7f313134000000000000000000000000000000000000000000000000000000000081565b7f313037000000000000000000000000000000000000000000000000000000000081565b7f313034000000000000000000000000000000000000000000000000000000000081565b7f313132000000000000000000000000000000000000000000000000000000000081565b60007f0000000000000000000000000000000000000000000000000000000000000000905090565b6000819050919050565b6108d8816108c5565b82525050565b60006020820190506108f360008301846108cf565b92915050565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610929826108fe565b9050919050565b6109398161091e565b811461094457600080fd5b50565b60008135905061095681610930565b92915050565b6000819050919050565b61096f8161095c565b811461097a57600080fd5b50565b60008135905061098c81610966565b92915050565b600080604083850312156109a9576109a86108f9565b5b60006109b785828601610947565b92505060206109c88582860161097d565b9150509250929050565b6109db816108c5565b81146109e657600080fd5b50565b6000813590506109f8816109d2565b92915050565b60008060408385031215610a1557610a146108f9565b5b6000610a23858286016109e9565b9250506020610a3485828601610947565b9150509250929050565b610a478161091e565b82525050565b6000602082019050610a626000830184610a3e565b92915050565b600082825260208201905092915050565b7f6c6f6e6720737472696e672065787065637465642066726f6d2061646472657360008201527f73203020746f2061646472657373206c6f707320746f2061646472657373206c60208201527f6f707320746f2061646472657373206c6f707320746f2061646472657373206c60408201527f6f707320746f2061646472657373206c6f707300000000000000000000000000606082015250565b6000610b21607383610a68565b9150610b2c82610a79565b608082019050919050565b60006020820190508181036000830152610b5081610b14565b9050919050565b6000819050919050565b610b72610b6d826108c5565b610b57565b82525050565b6000610b848284610b61565b60208201915081905092915050565b600081519050919050565b60005b83811015610bbc578082015181840152602081019050610ba1565b83811115610bcb576000848401525b50505050565b6000601f19601f8301169050919050565b6000610bed82610b93565b610bf78185610a68565b9350610c07818560208601610b9e565b610c1081610bd1565b840191505092915050565b60006020820190508181036000830152610c358184610be2565b905092915050565b610c468161095c565b82525050565b6000604082019050610c616000830185610a3e565b610c6e6020830184610c3d565b9392505050565b600081905092915050565b7fff00000000000000000000000000000000000000000000000000000000000000600082015250565b6000610cb6600183610c75565b9150610cc182610c80565b600182019050919050565b60008160601b9050919050565b6000610ce482610ccc565b9050919050565b6000610cf682610cd9565b9050919050565b610d0e610d098261091e565b610ceb565b82525050565b6000610d1f82610ca9565b9150610d2b8286610cfd565b601482019150610d3b8285610b61565b602082019150610d4b8284610b61565b60208201915081905094935050505056fea26469706673582212204850147ff482ac0bcfc413ce7f77e50405bb287d65ff895e51c6bc428a8ca70c64736f6c634300080d0033",
}

// DummyContractABI is the input ABI used to generate the binding from.
// Deprecated: Use DummyContractMetaData.ABI instead.
var DummyContractABI = DummyContractMetaData.ABI

// DummyContractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use DummyContractMetaData.Bin instead.
var DummyContractBin = DummyContractMetaData.Bin

// DeployDummyContract deploys a new Ethereum contract, binding an instance of DummyContract to it.
func DeployDummyContract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *DummyContract, error) {
	parsed, err := DummyContractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(DummyContractBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &DummyContract{DummyContractCaller: DummyContractCaller{contract: contract}, DummyContractTransactor: DummyContractTransactor{contract: contract}, DummyContractFilterer: DummyContractFilterer{contract: contract}}, nil
}

// DummyContract is an auto generated Go binding around an Ethereum contract.
type DummyContract struct {
	DummyContractCaller     // Read-only binding to the contract
	DummyContractTransactor // Write-only binding to the contract
	DummyContractFilterer   // Log filterer for contract events
}

// DummyContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type DummyContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DummyContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DummyContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DummyContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DummyContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DummyContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DummyContractSession struct {
	Contract     *DummyContract    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DummyContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DummyContractCallerSession struct {
	Contract *DummyContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// DummyContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DummyContractTransactorSession struct {
	Contract     *DummyContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// DummyContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type DummyContractRaw struct {
	Contract *DummyContract // Generic contract binding to access the raw methods on
}

// DummyContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DummyContractCallerRaw struct {
	Contract *DummyContractCaller // Generic read-only contract binding to access the raw methods on
}

// DummyContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DummyContractTransactorRaw struct {
	Contract *DummyContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDummyContract creates a new instance of DummyContract, bound to a specific deployed contract.
func NewDummyContract(address common.Address, backend bind.ContractBackend) (*DummyContract, error) {
	contract, err := bindDummyContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &DummyContract{DummyContractCaller: DummyContractCaller{contract: contract}, DummyContractTransactor: DummyContractTransactor{contract: contract}, DummyContractFilterer: DummyContractFilterer{contract: contract}}, nil
}

// NewDummyContractCaller creates a new read-only instance of DummyContract, bound to a specific deployed contract.
func NewDummyContractCaller(address common.Address, caller bind.ContractCaller) (*DummyContractCaller, error) {
	contract, err := bindDummyContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DummyContractCaller{contract: contract}, nil
}

// NewDummyContractTransactor creates a new write-only instance of DummyContract, bound to a specific deployed contract.
func NewDummyContractTransactor(address common.Address, transactor bind.ContractTransactor) (*DummyContractTransactor, error) {
	contract, err := bindDummyContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DummyContractTransactor{contract: contract}, nil
}

// NewDummyContractFilterer creates a new log filterer instance of DummyContract, bound to a specific deployed contract.
func NewDummyContractFilterer(address common.Address, filterer bind.ContractFilterer) (*DummyContractFilterer, error) {
	contract, err := bindDummyContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DummyContractFilterer{contract: contract}, nil
}

// bindDummyContract binds a generic wrapper to an already deployed contract.
func bindDummyContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(DummyContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DummyContract *DummyContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DummyContract.Contract.DummyContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DummyContract *DummyContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DummyContract.Contract.DummyContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DummyContract *DummyContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DummyContract.Contract.DummyContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DummyContract *DummyContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DummyContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DummyContract *DummyContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DummyContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DummyContract *DummyContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DummyContract.Contract.contract.Transact(opts, method, params...)
}

// ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND is a free data retrieval call binding the contract method 0xfa0f33b9.
//
// Solidity: function ETHDKG_ACCUSED_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_ACCUSED_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND is a free data retrieval call binding the contract method 0xfa0f33b9.
//
// Solidity: function ETHDKG_ACCUSED_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND(&_DummyContract.CallOpts)
}

// ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND is a free data retrieval call binding the contract method 0xfa0f33b9.
//
// Solidity: function ETHDKG_ACCUSED_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDDIDNOTDISTRIBUTESHARESINROUND(&_DummyContract.CallOpts)
}

// ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND is a free data retrieval call binding the contract method 0x2838edae.
//
// Solidity: function ETHDKG_ACCUSED_DISTRIBUTED_SHARES_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_ACCUSED_DISTRIBUTED_SHARES_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND is a free data retrieval call binding the contract method 0x2838edae.
//
// Solidity: function ETHDKG_ACCUSED_DISTRIBUTED_SHARES_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND(&_DummyContract.CallOpts)
}

// ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND is a free data retrieval call binding the contract method 0x2838edae.
//
// Solidity: function ETHDKG_ACCUSED_DISTRIBUTED_SHARES_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDDISTRIBUTEDSHARESINROUND(&_DummyContract.CallOpts)
}

// ETHDKGACCUSEDHASCOMMITMENTS is a free data retrieval call binding the contract method 0x4cd291bf.
//
// Solidity: function ETHDKG_ACCUSED_HAS_COMMITMENTS() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGACCUSEDHASCOMMITMENTS(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_ACCUSED_HAS_COMMITMENTS")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDHASCOMMITMENTS is a free data retrieval call binding the contract method 0x4cd291bf.
//
// Solidity: function ETHDKG_ACCUSED_HAS_COMMITMENTS() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGACCUSEDHASCOMMITMENTS() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDHASCOMMITMENTS(&_DummyContract.CallOpts)
}

// ETHDKGACCUSEDHASCOMMITMENTS is a free data retrieval call binding the contract method 0x4cd291bf.
//
// Solidity: function ETHDKG_ACCUSED_HAS_COMMITMENTS() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGACCUSEDHASCOMMITMENTS() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDHASCOMMITMENTS(&_DummyContract.CallOpts)
}

// ETHDKGACCUSEDNOTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0xe11879cc.
//
// Solidity: function ETHDKG_ACCUSED_NOT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGACCUSEDNOTPARTICIPATINGINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_ACCUSED_NOT_PARTICIPATING_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDNOTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0xe11879cc.
//
// Solidity: function ETHDKG_ACCUSED_NOT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGACCUSEDNOTPARTICIPATINGINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDNOTPARTICIPATINGINROUND(&_DummyContract.CallOpts)
}

// ETHDKGACCUSEDNOTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0xe11879cc.
//
// Solidity: function ETHDKG_ACCUSED_NOT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGACCUSEDNOTPARTICIPATINGINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDNOTPARTICIPATINGINROUND(&_DummyContract.CallOpts)
}

// ETHDKGACCUSEDNOTVALIDATOR is a free data retrieval call binding the contract method 0xf5f46e73.
//
// Solidity: function ETHDKG_ACCUSED_NOT_VALIDATOR() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGACCUSEDNOTVALIDATOR(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_ACCUSED_NOT_VALIDATOR")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDNOTVALIDATOR is a free data retrieval call binding the contract method 0xf5f46e73.
//
// Solidity: function ETHDKG_ACCUSED_NOT_VALIDATOR() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGACCUSEDNOTVALIDATOR() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDNOTVALIDATOR(&_DummyContract.CallOpts)
}

// ETHDKGACCUSEDNOTVALIDATOR is a free data retrieval call binding the contract method 0xf5f46e73.
//
// Solidity: function ETHDKG_ACCUSED_NOT_VALIDATOR() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGACCUSEDNOTVALIDATOR() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDNOTVALIDATOR(&_DummyContract.CallOpts)
}

// ETHDKGACCUSEDPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x55b83c56.
//
// Solidity: function ETHDKG_ACCUSED_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGACCUSEDPARTICIPATINGINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_ACCUSED_PARTICIPATING_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x55b83c56.
//
// Solidity: function ETHDKG_ACCUSED_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGACCUSEDPARTICIPATINGINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDPARTICIPATINGINROUND(&_DummyContract.CallOpts)
}

// ETHDKGACCUSEDPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x55b83c56.
//
// Solidity: function ETHDKG_ACCUSED_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGACCUSEDPARTICIPATINGINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDPARTICIPATINGINROUND(&_DummyContract.CallOpts)
}

// ETHDKGACCUSEDSUBMITTEDSHARESINROUND is a free data retrieval call binding the contract method 0xb23b8358.
//
// Solidity: function ETHDKG_ACCUSED_SUBMITTED_SHARES_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGACCUSEDSUBMITTEDSHARESINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_ACCUSED_SUBMITTED_SHARES_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGACCUSEDSUBMITTEDSHARESINROUND is a free data retrieval call binding the contract method 0xb23b8358.
//
// Solidity: function ETHDKG_ACCUSED_SUBMITTED_SHARES_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGACCUSEDSUBMITTEDSHARESINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDSUBMITTEDSHARESINROUND(&_DummyContract.CallOpts)
}

// ETHDKGACCUSEDSUBMITTEDSHARESINROUND is a free data retrieval call binding the contract method 0xb23b8358.
//
// Solidity: function ETHDKG_ACCUSED_SUBMITTED_SHARES_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGACCUSEDSUBMITTEDSHARESINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGACCUSEDSUBMITTEDSHARESINROUND(&_DummyContract.CallOpts)
}

// ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND is a free data retrieval call binding the contract method 0xd65915e2.
//
// Solidity: function ETHDKG_DISPUTER_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_DISPUTER_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND is a free data retrieval call binding the contract method 0xd65915e2.
//
// Solidity: function ETHDKG_DISPUTER_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND(&_DummyContract.CallOpts)
}

// ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND is a free data retrieval call binding the contract method 0xd65915e2.
//
// Solidity: function ETHDKG_DISPUTER_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGDISPUTERDIDNOTDISTRIBUTESHARESINROUND(&_DummyContract.CallOpts)
}

// ETHDKGDISPUTERNOTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x763df93d.
//
// Solidity: function ETHDKG_DISPUTER_NOT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGDISPUTERNOTPARTICIPATINGINROUND(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_DISPUTER_NOT_PARTICIPATING_IN_ROUND")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGDISPUTERNOTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x763df93d.
//
// Solidity: function ETHDKG_DISPUTER_NOT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGDISPUTERNOTPARTICIPATINGINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGDISPUTERNOTPARTICIPATINGINROUND(&_DummyContract.CallOpts)
}

// ETHDKGDISPUTERNOTPARTICIPATINGINROUND is a free data retrieval call binding the contract method 0x763df93d.
//
// Solidity: function ETHDKG_DISPUTER_NOT_PARTICIPATING_IN_ROUND() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGDISPUTERNOTPARTICIPATINGINROUND() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGDISPUTERNOTPARTICIPATINGINROUND(&_DummyContract.CallOpts)
}

// ETHDKGINVALIDKEYORPROOF is a free data retrieval call binding the contract method 0xa852713f.
//
// Solidity: function ETHDKG_INVALID_KEY_OR_PROOF() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGINVALIDKEYORPROOF(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_INVALID_KEY_OR_PROOF")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGINVALIDKEYORPROOF is a free data retrieval call binding the contract method 0xa852713f.
//
// Solidity: function ETHDKG_INVALID_KEY_OR_PROOF() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGINVALIDKEYORPROOF() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGINVALIDKEYORPROOF(&_DummyContract.CallOpts)
}

// ETHDKGINVALIDKEYORPROOF is a free data retrieval call binding the contract method 0xa852713f.
//
// Solidity: function ETHDKG_INVALID_KEY_OR_PROOF() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGINVALIDKEYORPROOF() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGINVALIDKEYORPROOF(&_DummyContract.CallOpts)
}

// ETHDKGMINVALIDATORSNOTMET is a free data retrieval call binding the contract method 0x7e9f3983.
//
// Solidity: function ETHDKG_MIN_VALIDATORS_NOT_MET() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGMINVALIDATORSNOTMET(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_MIN_VALIDATORS_NOT_MET")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGMINVALIDATORSNOTMET is a free data retrieval call binding the contract method 0x7e9f3983.
//
// Solidity: function ETHDKG_MIN_VALIDATORS_NOT_MET() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGMINVALIDATORSNOTMET() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGMINVALIDATORSNOTMET(&_DummyContract.CallOpts)
}

// ETHDKGMINVALIDATORSNOTMET is a free data retrieval call binding the contract method 0x7e9f3983.
//
// Solidity: function ETHDKG_MIN_VALIDATORS_NOT_MET() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGMINVALIDATORSNOTMET() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGMINVALIDATORSNOTMET(&_DummyContract.CallOpts)
}

// ETHDKGNOTINDISPUTEPHASE is a free data retrieval call binding the contract method 0x4c4cfd75.
//
// Solidity: function ETHDKG_NOT_IN_DISPUTE_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGNOTINDISPUTEPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_NOT_IN_DISPUTE_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINDISPUTEPHASE is a free data retrieval call binding the contract method 0x4c4cfd75.
//
// Solidity: function ETHDKG_NOT_IN_DISPUTE_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGNOTINDISPUTEPHASE() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGNOTINDISPUTEPHASE(&_DummyContract.CallOpts)
}

// ETHDKGNOTINDISPUTEPHASE is a free data retrieval call binding the contract method 0x4c4cfd75.
//
// Solidity: function ETHDKG_NOT_IN_DISPUTE_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGNOTINDISPUTEPHASE() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGNOTINDISPUTEPHASE(&_DummyContract.CallOpts)
}

// ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE is a free data retrieval call binding the contract method 0x6d429ef2.
//
// Solidity: function ETHDKG_NOT_IN_POST_GPKJ_SUBMISSION_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_NOT_IN_POST_GPKJ_SUBMISSION_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE is a free data retrieval call binding the contract method 0x6d429ef2.
//
// Solidity: function ETHDKG_NOT_IN_POST_GPKJ_SUBMISSION_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE(&_DummyContract.CallOpts)
}

// ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE is a free data retrieval call binding the contract method 0x6d429ef2.
//
// Solidity: function ETHDKG_NOT_IN_POST_GPKJ_SUBMISSION_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGNOTINPOSTGPKJSUBMISSIONPHASE(&_DummyContract.CallOpts)
}

// ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE is a free data retrieval call binding the contract method 0x60987646.
//
// Solidity: function ETHDKG_NOT_IN_POST_KEYSHARE_SUBMISSION_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_NOT_IN_POST_KEYSHARE_SUBMISSION_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE is a free data retrieval call binding the contract method 0x60987646.
//
// Solidity: function ETHDKG_NOT_IN_POST_KEYSHARE_SUBMISSION_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE(&_DummyContract.CallOpts)
}

// ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE is a free data retrieval call binding the contract method 0x60987646.
//
// Solidity: function ETHDKG_NOT_IN_POST_KEYSHARE_SUBMISSION_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGNOTINPOSTKEYSHARESUBMISSIONPHASE(&_DummyContract.CallOpts)
}

// ETHDKGNOTINPOSTREGISTRATIONPHASE is a free data retrieval call binding the contract method 0x7385db5d.
//
// Solidity: function ETHDKG_NOT_IN_POST_REGISTRATION_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGNOTINPOSTREGISTRATIONPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_NOT_IN_POST_REGISTRATION_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINPOSTREGISTRATIONPHASE is a free data retrieval call binding the contract method 0x7385db5d.
//
// Solidity: function ETHDKG_NOT_IN_POST_REGISTRATION_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGNOTINPOSTREGISTRATIONPHASE() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGNOTINPOSTREGISTRATIONPHASE(&_DummyContract.CallOpts)
}

// ETHDKGNOTINPOSTREGISTRATIONPHASE is a free data retrieval call binding the contract method 0x7385db5d.
//
// Solidity: function ETHDKG_NOT_IN_POST_REGISTRATION_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGNOTINPOSTREGISTRATIONPHASE() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGNOTINPOSTREGISTRATIONPHASE(&_DummyContract.CallOpts)
}

// ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE is a free data retrieval call binding the contract method 0x8e25d1e1.
//
// Solidity: function ETHDKG_NOT_IN_POST_SHARED_DISTRIBUTION_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_NOT_IN_POST_SHARED_DISTRIBUTION_PHASE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE is a free data retrieval call binding the contract method 0x8e25d1e1.
//
// Solidity: function ETHDKG_NOT_IN_POST_SHARED_DISTRIBUTION_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE(&_DummyContract.CallOpts)
}

// ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE is a free data retrieval call binding the contract method 0x8e25d1e1.
//
// Solidity: function ETHDKG_NOT_IN_POST_SHARED_DISTRIBUTION_PHASE() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGNOTINPOSTSHAREDDISTRIBUTIONPHASE(&_DummyContract.CallOpts)
}

// ETHDKGONLYVALIDATORSALLOWED is a free data retrieval call binding the contract method 0x83c069e4.
//
// Solidity: function ETHDKG_ONLY_VALIDATORS_ALLOWED() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGONLYVALIDATORSALLOWED(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_ONLY_VALIDATORS_ALLOWED")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGONLYVALIDATORSALLOWED is a free data retrieval call binding the contract method 0x83c069e4.
//
// Solidity: function ETHDKG_ONLY_VALIDATORS_ALLOWED() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGONLYVALIDATORSALLOWED() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGONLYVALIDATORSALLOWED(&_DummyContract.CallOpts)
}

// ETHDKGONLYVALIDATORSALLOWED is a free data retrieval call binding the contract method 0x83c069e4.
//
// Solidity: function ETHDKG_ONLY_VALIDATORS_ALLOWED() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGONLYVALIDATORSALLOWED() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGONLYVALIDATORSALLOWED(&_DummyContract.CallOpts)
}

// ETHDKGSHARESANDCOMMITMENTSMISMATCH is a free data retrieval call binding the contract method 0xe0c1afd8.
//
// Solidity: function ETHDKG_SHARES_AND_COMMITMENTS_MISMATCH() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGSHARESANDCOMMITMENTSMISMATCH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_SHARES_AND_COMMITMENTS_MISMATCH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGSHARESANDCOMMITMENTSMISMATCH is a free data retrieval call binding the contract method 0xe0c1afd8.
//
// Solidity: function ETHDKG_SHARES_AND_COMMITMENTS_MISMATCH() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGSHARESANDCOMMITMENTSMISMATCH() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGSHARESANDCOMMITMENTSMISMATCH(&_DummyContract.CallOpts)
}

// ETHDKGSHARESANDCOMMITMENTSMISMATCH is a free data retrieval call binding the contract method 0xe0c1afd8.
//
// Solidity: function ETHDKG_SHARES_AND_COMMITMENTS_MISMATCH() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGSHARESANDCOMMITMENTSMISMATCH() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGSHARESANDCOMMITMENTSMISMATCH(&_DummyContract.CallOpts)
}

// ETHDKGVARIABLECANNOTBESETWHILERUNNING is a free data retrieval call binding the contract method 0x79ec8296.
//
// Solidity: function ETHDKG_VARIABLE_CANNOT_BE_SET_WHILE_RUNNING() view returns(bytes32)
func (_DummyContract *DummyContractCaller) ETHDKGVARIABLECANNOTBESETWHILERUNNING(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "ETHDKG_VARIABLE_CANNOT_BE_SET_WHILE_RUNNING")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ETHDKGVARIABLECANNOTBESETWHILERUNNING is a free data retrieval call binding the contract method 0x79ec8296.
//
// Solidity: function ETHDKG_VARIABLE_CANNOT_BE_SET_WHILE_RUNNING() view returns(bytes32)
func (_DummyContract *DummyContractSession) ETHDKGVARIABLECANNOTBESETWHILERUNNING() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGVARIABLECANNOTBESETWHILERUNNING(&_DummyContract.CallOpts)
}

// ETHDKGVARIABLECANNOTBESETWHILERUNNING is a free data retrieval call binding the contract method 0x79ec8296.
//
// Solidity: function ETHDKG_VARIABLE_CANNOT_BE_SET_WHILE_RUNNING() view returns(bytes32)
func (_DummyContract *DummyContractCallerSession) ETHDKGVARIABLECANNOTBESETWHILERUNNING() ([32]byte, error) {
	return _DummyContract.Contract.ETHDKGVARIABLECANNOTBESETWHILERUNNING(&_DummyContract.CallOpts)
}

// Dummy is a free data retrieval call binding the contract method 0x32e43a11.
//
// Solidity: function dummy() view returns()
func (_DummyContract *DummyContractCaller) Dummy(opts *bind.CallOpts) error {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "dummy")

	if err != nil {
		return err
	}

	return err

}

// Dummy is a free data retrieval call binding the contract method 0x32e43a11.
//
// Solidity: function dummy() view returns()
func (_DummyContract *DummyContractSession) Dummy() error {
	return _DummyContract.Contract.Dummy(&_DummyContract.CallOpts)
}

// Dummy is a free data retrieval call binding the contract method 0x32e43a11.
//
// Solidity: function dummy() view returns()
func (_DummyContract *DummyContractCallerSession) Dummy() error {
	return _DummyContract.Contract.Dummy(&_DummyContract.CallOpts)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_DummyContract *DummyContractCaller) GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error) {
	var out []interface{}
	err := _DummyContract.contract.Call(opts, &out, "getMetamorphicContractAddress", _salt, _factory)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_DummyContract *DummyContractSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _DummyContract.Contract.GetMetamorphicContractAddress(&_DummyContract.CallOpts, _salt, _factory)
}

// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
//
// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
func (_DummyContract *DummyContractCallerSession) GetMetamorphicContractAddress(_salt [32]byte, _factory common.Address) (common.Address, error) {
	return _DummyContract.Contract.GetMetamorphicContractAddress(&_DummyContract.CallOpts, _salt, _factory)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 amount) returns()
func (_DummyContract *DummyContractTransactor) Mint(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _DummyContract.contract.Transact(opts, "mint", to, amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 amount) returns()
func (_DummyContract *DummyContractSession) Mint(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _DummyContract.Contract.Mint(&_DummyContract.TransactOpts, to, amount)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 amount) returns()
func (_DummyContract *DummyContractTransactorSession) Mint(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _DummyContract.Contract.Mint(&_DummyContract.TransactOpts, to, amount)
}
