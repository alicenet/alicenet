//

package db

import (
	"context"

	trie "github.com/MadBase/MadNet/badgerTrie"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/dgraph-io/badger/v2"
)

// DatabaseIface .
type DatabaseIface interface {
	// Init will initialize the database
	Init(DB *badger.DB) error
	DB() *badger.DB
	View(fn TxnFunc) error
	Update(fn TxnFunc) error
	Sync() error
	GarbageCollect() error
	SetEncryptedStore(txn *badger.Txn, v *objs.EncryptedStore) error
	GetEncryptedStore(txn *badger.Txn, name []byte) (*objs.EncryptedStore, error)
	SetCurrentRoundState(txn *badger.Txn, v *objs.RoundState) error
	GetCurrentRoundState(txn *badger.Txn, vaddr []byte) (*objs.RoundState, error)
	SetHistoricRoundState(txn *badger.Txn, v *objs.RoundState) error
	GetHistoricRoundState(txn *badger.Txn, vaddr []byte, height uint32, round uint32) (*objs.RoundState, error)
	DeleteBeforeHistoricRoundState(txn *badger.Txn, height uint32, maxnum int) error
	SetValidatorSet(txn *badger.Txn, v *objs.ValidatorSet) error
	GetValidatorSet(txn *badger.Txn, height uint32) (*objs.ValidatorSet, error)
	MakeHeaderTrieKeyFromHeight(height uint32) []byte
	GetHeaderTrieRoot(txn *badger.Txn, height uint32) ([]byte, error)
	UpdateHeaderTrieRootFastSync(txn *badger.Txn, v *objs.BlockHeader) error
	SetCommittedBlockHeader(txn *badger.Txn, v *objs.BlockHeader) error
	SetCommittedBlockHeaderFastSync(txn *badger.Txn, v *objs.BlockHeader) error
	GetHeaderRootForProposal(txn *badger.Txn) ([]byte, error)
	DeleteCommittedBlockHeader(txn *badger.Txn, height uint32) error
	ValidateCommittedBlockHeaderWithProof(txn *badger.Txn, root []byte, blockHeader *objs.BlockHeader, proof []byte) (bool, error)
	GetCommittedBlockHeaderWithProof(txn *badger.Txn, root []byte, blockHeight uint32) (*objs.BlockHeader, []byte, error)
	GetCommittedBlockHeader(txn *badger.Txn, height uint32) (*objs.BlockHeader, error)
	GetCommittedBlockHeaderByHash(txn *badger.Txn, hash []byte) (*objs.BlockHeader, error)
	GetMostRecentCommittedBlockHeaderFastSync(txn *badger.Txn) (*objs.BlockHeader, error)
	SetOwnState(txn *badger.Txn, v *objs.OwnState) error
	GetOwnState(txn *badger.Txn) (*objs.OwnState, error)
	SetOwnValidatingState(txn *badger.Txn, v *objs.OwnValidatingState) error
	GetOwnValidatingState(txn *badger.Txn) (*objs.OwnValidatingState, error)
	SetBroadcastBlockHeader(txn *badger.Txn, v *objs.BlockHeader) error
	GetBroadcastBlockHeader(txn *badger.Txn) (*objs.BlockHeader, error)
	SubscribeBroadcastBlockHeader(ctx context.Context, cb func([]byte) error)
	SetBroadcastRCert(txn *badger.Txn, v *objs.RCert) error
	GetBroadcastRCert(txn *badger.Txn) (*objs.RCert, error)
	SubscribeBroadcastRCert(ctx context.Context, cb func([]byte) error)
	SetBroadcastTransaction(txn *badger.Txn, v []byte) error
	SubscribeBroadcastTransaction(ctx context.Context, cb func([]byte) error)
	SetBroadcastProposal(txn *badger.Txn, v *objs.Proposal) error
	GetBroadcastProposal(txn *badger.Txn) (*objs.Proposal, error)
	SubscribeBroadcastProposal(ctx context.Context, cb func([]byte) error)
	SetBroadcastPreVote(txn *badger.Txn, v *objs.PreVote) error
	GetBroadcastPreVote(txn *badger.Txn) (*objs.PreVote, error)
	SubscribeBroadcastPreVote(ctx context.Context, cb func([]byte) error)
	SetBroadcastPreVoteNil(txn *badger.Txn, v *objs.PreVoteNil) error
	GetBroadcastPreVoteNil(txn *badger.Txn) (*objs.PreVoteNil, error)
	SubscribeBroadcastPreVoteNil(ctx context.Context, cb func([]byte) error)
	SetBroadcastPreCommit(txn *badger.Txn, v *objs.PreCommit) error
	GetBroadcastPreCommit(txn *badger.Txn) (*objs.PreCommit, error)
	SubscribeBroadcastPreCommit(ctx context.Context, cb func([]byte) error)
	SetBroadcastPreCommitNil(txn *badger.Txn, v *objs.PreCommitNil) error
	GetBroadcastPreCommitNil(txn *badger.Txn) (*objs.PreCommitNil, error)
	SubscribeBroadcastPreCommitNil(ctx context.Context, cb func([]byte) error)
	SetBroadcastNextHeight(txn *badger.Txn, v *objs.NextHeight) error
	GetBroadcastNextHeight(txn *badger.Txn) (*objs.NextHeight, error)
	SubscribeBroadcastNextHeight(ctx context.Context, cb func([]byte) error)
	SetBroadcastNextRound(txn *badger.Txn, v *objs.NextRound) error
	GetBroadcastNextRound(txn *badger.Txn) (*objs.NextRound, error)
	SubscribeBroadcastNextRound(ctx context.Context, cb func([]byte) error)
	SetSnapshotBlockHeader(txn *badger.Txn, v *objs.BlockHeader) error
	GetSnapshotBlockHeader(txn *badger.Txn, height uint32) (*objs.BlockHeader, error)
	GetLastSnapshot(txn *badger.Txn) (*objs.BlockHeader, error)
	SetTxCacheItem(txn *badger.Txn, height uint32, txHash []byte, tx []byte) error
	GetTxCacheItem(txn *badger.Txn, height uint32, txHash []byte) ([]byte, error)
	TxCacheDropBefore(txn *badger.Txn, beforeHeight uint32, maxKeys int) error
	GetTxCacheSet(txn *badger.Txn, height uint32) ([][]byte, [][]byte, error)
	DropPendingHdrNodeKeys() error
	SetPendingHdrNodeKey(txn *badger.Txn, nodeKey []byte, layer int) error
	GetPendingHdrNodeKey(txn *badger.Txn, nodeKey []byte) (int, error)
	DeletePendingHdrNodeKey(txn *badger.Txn, nodeKey []byte) error
	CountPendingHdrNodeKeys(txn *badger.Txn) (int, error)
	GetPendingHdrNodeKeysIter(txn *badger.Txn) *PendingHdrNodeIter
	DropPendingNodeKeys() error
	SetPendingNodeKey(txn *badger.Txn, nodeKey []byte, layer int) error
	GetPendingNodeKey(txn *badger.Txn, nodeKey []byte) (int, error)
	DeletePendingNodeKey(txn *badger.Txn, nodeKey []byte) error
	CountPendingNodeKeys(txn *badger.Txn) (int, error)
	GetPendingNodeKeysIter(txn *badger.Txn) *PendingNodeIter
	DropPendingLeafKeys() error
	SetPendingLeafKey(txn *badger.Txn, leafKey []byte, value []byte) error
	GetPendingLeafKey(txn *badger.Txn, leafKey []byte) ([]byte, error)
	DeletePendingLeafKey(txn *badger.Txn, leafKey []byte) error
	CountPendingLeafKeys(txn *badger.Txn) (int, error)
	GetPendingLeafKeysIter(txn *badger.Txn) *PendingLeafIter
	SetSafeToProceed(txn *badger.Txn, height uint32, isSafe bool) error
	GetSafeToProceed(txn *badger.Txn, height uint32) (bool, error)
	ContainsSnapShotHdrNode(txn *badger.Txn, root []byte) (bool, error)
	GetSnapShotHdrNode(txn *badger.Txn, root []byte) ([]byte, error)
	SetSnapShotHdrNode(txn *badger.Txn, batch []byte, root []byte, layer int) ([][]byte, int, []trie.LeafNode, error)
	DropPendingHdrLeafKeys() error
	SetPendingHdrLeafKey(txn *badger.Txn, hdrLeafKey []byte, value []byte) error
	GetPendingHdrLeafKey(txn *badger.Txn, hdrLeafKey []byte) ([]byte, error)
	DeletePendingHdrLeafKey(txn *badger.Txn, hdrLeafKey []byte) error
	CountPendingHdrLeafKeys(txn *badger.Txn) (int, error)
	GetPendingHdrLeafKeysIter(txn *badger.Txn) *PendingHdrLeafIter
	DropStagedBlockHeaderKeys() error
	SetStagedBlockHeader(txn *badger.Txn, height uint32, value []byte) error
	GetStagedBlockHeader(txn *badger.Txn, height uint32) (*objs.BlockHeader, error)
	DeleteStagedBlockHeaderKey(txn *badger.Txn, height uint32) error
	CountStagedBlockHeaderKeys(txn *badger.Txn) (int, error)
	GetStagedBlockHeaderKeyIter(txn *badger.Txn) *StagedBlockHeaderKeyIter
}
