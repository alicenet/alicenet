// Generated by ifacemaker. DO NOT EDIT.

package bindings

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// IPublicStakingFilterer ...
type IPublicStakingFilterer interface {
	// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
	//
	// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
	FilterApproval(opts *bind.FilterOpts, owner []common.Address, approved []common.Address, tokenId []*big.Int) (*PublicStakingApprovalIterator, error)
	// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
	//
	// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
	WatchApproval(opts *bind.WatchOpts, sink chan<- *PublicStakingApproval, owner []common.Address, approved []common.Address, tokenId []*big.Int) (event.Subscription, error)
	// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
	//
	// Solidity: event Approval(address indexed owner, address indexed approved, uint256 indexed tokenId)
	ParseApproval(log types.Log) (*PublicStakingApproval, error)
	// FilterApprovalForAll is a free log retrieval operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
	//
	// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
	FilterApprovalForAll(opts *bind.FilterOpts, owner []common.Address, operator []common.Address) (*PublicStakingApprovalForAllIterator, error)
	// WatchApprovalForAll is a free log subscription operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
	//
	// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
	WatchApprovalForAll(opts *bind.WatchOpts, sink chan<- *PublicStakingApprovalForAll, owner []common.Address, operator []common.Address) (event.Subscription, error)
	// ParseApprovalForAll is a log parse operation binding the contract event 0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31.
	//
	// Solidity: event ApprovalForAll(address indexed owner, address indexed operator, bool approved)
	ParseApprovalForAll(log types.Log) (*PublicStakingApprovalForAll, error)
	// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
	//
	// Solidity: event Initialized(uint8 version)
	FilterInitialized(opts *bind.FilterOpts) (*PublicStakingInitializedIterator, error)
	// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
	//
	// Solidity: event Initialized(uint8 version)
	WatchInitialized(opts *bind.WatchOpts, sink chan<- *PublicStakingInitialized) (event.Subscription, error)
	// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
	//
	// Solidity: event Initialized(uint8 version)
	ParseInitialized(log types.Log) (*PublicStakingInitialized, error)
	// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
	//
	// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
	FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*PublicStakingTransferIterator, error)
	// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
	//
	// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
	WatchTransfer(opts *bind.WatchOpts, sink chan<- *PublicStakingTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error)
	// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
	//
	// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
	ParseTransfer(log types.Log) (*PublicStakingTransfer, error)
}
