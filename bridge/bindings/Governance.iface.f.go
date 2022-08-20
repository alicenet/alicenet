// Generated by ifacemaker. DO NOT EDIT.

package bindings

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// IGovernanceFilterer ...
type IGovernanceFilterer interface {
	// FilterSnapshotTaken is a free log retrieval operation binding the contract event 0x24b0dff7469a7007db81d741ef90d7966936fb78bc19d667f4575ecbf56ab350.
	//
	// Solidity: event SnapshotTaken(uint256 chainId, uint256 indexed epoch, uint256 height, address indexed validator, bool isSafeToProceedConsensus, bytes signatureRaw)
	FilterSnapshotTaken(opts *bind.FilterOpts, epoch []*big.Int, validator []common.Address) (*GovernanceSnapshotTakenIterator, error)
	// WatchSnapshotTaken is a free log subscription operation binding the contract event 0x24b0dff7469a7007db81d741ef90d7966936fb78bc19d667f4575ecbf56ab350.
	//
	// Solidity: event SnapshotTaken(uint256 chainId, uint256 indexed epoch, uint256 height, address indexed validator, bool isSafeToProceedConsensus, bytes signatureRaw)
	WatchSnapshotTaken(opts *bind.WatchOpts, sink chan<- *GovernanceSnapshotTaken, epoch []*big.Int, validator []common.Address) (event.Subscription, error)
	// ParseSnapshotTaken is a log parse operation binding the contract event 0x24b0dff7469a7007db81d741ef90d7966936fb78bc19d667f4575ecbf56ab350.
	//
	// Solidity: event SnapshotTaken(uint256 chainId, uint256 indexed epoch, uint256 height, address indexed validator, bool isSafeToProceedConsensus, bytes signatureRaw)
	ParseSnapshotTaken(log types.Log) (*GovernanceSnapshotTaken, error)
	// FilterValueUpdated is a free log retrieval operation binding the contract event 0x36dcd0e03525dedd9d5c21a263ef5f35d030298b5c48f1a713006aefc064ad05.
	//
	// Solidity: event ValueUpdated(uint256 indexed epoch, uint256 indexed key, bytes32 indexed value, address who)
	FilterValueUpdated(opts *bind.FilterOpts, epoch []*big.Int, key []*big.Int, value [][32]byte) (*GovernanceValueUpdatedIterator, error)
	// WatchValueUpdated is a free log subscription operation binding the contract event 0x36dcd0e03525dedd9d5c21a263ef5f35d030298b5c48f1a713006aefc064ad05.
	//
	// Solidity: event ValueUpdated(uint256 indexed epoch, uint256 indexed key, bytes32 indexed value, address who)
	WatchValueUpdated(opts *bind.WatchOpts, sink chan<- *GovernanceValueUpdated, epoch []*big.Int, key []*big.Int, value [][32]byte) (event.Subscription, error)
	// ParseValueUpdated is a log parse operation binding the contract event 0x36dcd0e03525dedd9d5c21a263ef5f35d030298b5c48f1a713006aefc064ad05.
	//
	// Solidity: event ValueUpdated(uint256 indexed epoch, uint256 indexed key, bytes32 indexed value, address who)
	ParseValueUpdated(log types.Log) (*GovernanceValueUpdated, error)
}
