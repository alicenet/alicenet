// Generated by ifacemaker. DO NOT EDIT.

package bindings

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// IValidatorPoolTransactor ...
type IValidatorPoolTransactor interface {
	// ClaimExitingNFTPosition is a paid mutator transaction binding the contract method 0x769cc695.
	//
	// Solidity: function claimExitingNFTPosition() returns(uint256)
	ClaimExitingNFTPosition(opts *bind.TransactOpts) (*types.Transaction, error)
	// CollectProfits is a paid mutator transaction binding the contract method 0xc958e0d6.
	//
	// Solidity: function collectProfits() returns(uint256 payoutEth, uint256 payoutToken)
	CollectProfits(opts *bind.TransactOpts) (*types.Transaction, error)
	// CompleteETHDKG is a paid mutator transaction binding the contract method 0x8f579924.
	//
	// Solidity: function completeETHDKG() returns()
	CompleteETHDKG(opts *bind.TransactOpts) (*types.Transaction, error)
	// Initialize is a paid mutator transaction binding the contract method 0x60a2da44.
	//
	// Solidity: function initialize(uint256 stakeAmount_, uint256 maxNumValidators_, uint256 disputerReward_, uint256 maxIntervalWithoutSnapshots_) returns()
	Initialize(opts *bind.TransactOpts, stakeAmount_ *big.Int, maxNumValidators_ *big.Int, disputerReward_ *big.Int, maxIntervalWithoutSnapshots_ *big.Int) (*types.Transaction, error)
	// InitializeETHDKG is a paid mutator transaction binding the contract method 0x57b51c9c.
	//
	// Solidity: function initializeETHDKG() returns()
	InitializeETHDKG(opts *bind.TransactOpts) (*types.Transaction, error)
	// MajorSlash is a paid mutator transaction binding the contract method 0x048d56c7.
	//
	// Solidity: function majorSlash(address dishonestValidator_, address disputer_) returns()
	MajorSlash(opts *bind.TransactOpts, dishonestValidator_ common.Address, disputer_ common.Address) (*types.Transaction, error)
	// MinorSlash is a paid mutator transaction binding the contract method 0x64c0461c.
	//
	// Solidity: function minorSlash(address dishonestValidator_, address disputer_) returns()
	MinorSlash(opts *bind.TransactOpts, dishonestValidator_ common.Address, disputer_ common.Address) (*types.Transaction, error)
	// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
	//
	// Solidity: function onERC721Received(address , address , uint256 , bytes ) returns(bytes4)
	OnERC721Received(opts *bind.TransactOpts, arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 []byte) (*types.Transaction, error)
	// PauseConsensus is a paid mutator transaction binding the contract method 0x1e5975f4.
	//
	// Solidity: function pauseConsensus() returns()
	PauseConsensus(opts *bind.TransactOpts) (*types.Transaction, error)
	// PauseConsensusOnArbitraryHeight is a paid mutator transaction binding the contract method 0xbc33bb01.
	//
	// Solidity: function pauseConsensusOnArbitraryHeight(uint256 aliceNetHeight_) returns()
	PauseConsensusOnArbitraryHeight(opts *bind.TransactOpts, aliceNetHeight_ *big.Int) (*types.Transaction, error)
	// RegisterValidators is a paid mutator transaction binding the contract method 0x65bd91af.
	//
	// Solidity: function registerValidators(address[] validators_, uint256[] stakerTokenIDs_) returns()
	RegisterValidators(opts *bind.TransactOpts, validators_ []common.Address, stakerTokenIDs_ []*big.Int) (*types.Transaction, error)
	// ScheduleMaintenance is a paid mutator transaction binding the contract method 0x2380db1a.
	//
	// Solidity: function scheduleMaintenance() returns()
	ScheduleMaintenance(opts *bind.TransactOpts) (*types.Transaction, error)
	// SetDisputerReward is a paid mutator transaction binding the contract method 0x7d907284.
	//
	// Solidity: function setDisputerReward(uint256 disputerReward_) returns()
	SetDisputerReward(opts *bind.TransactOpts, disputerReward_ *big.Int) (*types.Transaction, error)
	// SetLocation is a paid mutator transaction binding the contract method 0x827bfbdf.
	//
	// Solidity: function setLocation(string ip_) returns()
	SetLocation(opts *bind.TransactOpts, ip_ string) (*types.Transaction, error)
	// SetMaxIntervalWithoutSnapshots is a paid mutator transaction binding the contract method 0x564a7005.
	//
	// Solidity: function setMaxIntervalWithoutSnapshots(uint256 maxIntervalWithoutSnapshots) returns()
	SetMaxIntervalWithoutSnapshots(opts *bind.TransactOpts, maxIntervalWithoutSnapshots *big.Int) (*types.Transaction, error)
	// SetMaxNumValidators is a paid mutator transaction binding the contract method 0x6c0da0b4.
	//
	// Solidity: function setMaxNumValidators(uint256 maxNumValidators_) returns()
	SetMaxNumValidators(opts *bind.TransactOpts, maxNumValidators_ *big.Int) (*types.Transaction, error)
	// SetStakeAmount is a paid mutator transaction binding the contract method 0x43808c50.
	//
	// Solidity: function setStakeAmount(uint256 stakeAmount_) returns()
	SetStakeAmount(opts *bind.TransactOpts, stakeAmount_ *big.Int) (*types.Transaction, error)
	// SkimExcessEth is a paid mutator transaction binding the contract method 0x971b505b.
	//
	// Solidity: function skimExcessEth(address to_) returns(uint256 excess)
	SkimExcessEth(opts *bind.TransactOpts, to_ common.Address) (*types.Transaction, error)
	// SkimExcessToken is a paid mutator transaction binding the contract method 0x7aa507fb.
	//
	// Solidity: function skimExcessToken(address to_) returns(uint256 excess)
	SkimExcessToken(opts *bind.TransactOpts, to_ common.Address) (*types.Transaction, error)
	// UnregisterAllValidators is a paid mutator transaction binding the contract method 0xf6442e24.
	//
	// Solidity: function unregisterAllValidators() returns()
	UnregisterAllValidators(opts *bind.TransactOpts) (*types.Transaction, error)
	// UnregisterValidators is a paid mutator transaction binding the contract method 0xc6e86ad6.
	//
	// Solidity: function unregisterValidators(address[] validators_) returns()
	UnregisterValidators(opts *bind.TransactOpts, validators_ []common.Address) (*types.Transaction, error)
	// Receive is a paid mutator transaction binding the contract receive function.
	//
	// Solidity: receive() payable returns()
	Receive(opts *bind.TransactOpts) (*types.Transaction, error)
}
