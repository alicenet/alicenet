package interfaces

import (
	trie "github.com/MadBase/MadNet/badgerTrie"
	"github.com/dgraph-io/badger/v2"
)

// Application ...
type Application interface {
	// UnmarshalTx allows a transaction to be unmarshalled into a transaction
	// interface for use by the consensus algorithm
	UnmarshalTx(txb []byte) (Transaction, error)
	// IsValid returns true if the list of transactions is a valid transformation
	// and false if the list is not valid. If an error is returned, it indicates
	// a low level failure that should stop the main loop.
	IsValid(txn *badger.Txn, chainID uint32, height uint32, stateHash []byte, tx []Transaction) (bool, error)
	// GetValidProposal is a function that returns a list of transactions
	// that will cause a valid state transition function for the local node's
	// current state. This is the function used to create a new proposal.
	// comes from application logic
	GetValidProposal(txn *badger.Txn, chainID, height, maxBytes uint32) ([]Transaction, []byte, error)
	// ApplyState is a function that returns a list of transactions
	// that will cause a valid state transition function for the local node's
	// current state. This is the function used to create a new proposal.
	// comes from application logic
	ApplyState(txn *badger.Txn, chainID uint32, height uint32, txs []Transaction) (stateHash []byte, err error)
	// PendingTxAdd adds a transaction to the txPool and cleans up any stale
	// tx as a result.
	PendingTxAdd(txn *badger.Txn, chainID uint32, height uint32, txs []Transaction) error
	// MinedTxGet returns a list of mined transactions and a list of missing
	// transaction hashes for mined transactions
	MinedTxGet(txn *badger.Txn, txHash [][]byte) ([]Transaction, [][]byte, error)
	// PendingTxGet returns a list of transactions and a list of missing
	// transaction hashes from the pending transaction pool
	PendingTxGet(txn *badger.Txn, height uint32, txHashes [][]byte) ([]Transaction, [][]byte, error)
	// PendingTxContains returns a list of missing transaction hashes
	// from the pending tx pool
	PendingTxContains(txn *badger.Txn, height uint32, txHashes [][]byte) ([][]byte, error)
	// StoreSnapShotNodes stores a snapshot node into the state trie of the application
	StoreSnapShotNode(txn *badger.Txn, batch []byte, root []byte, layer int) ([][]byte, int, []trie.LeafNode, error)
	// GetSnapShotNode returns a snapshot node from the state trie to a peer
	GetSnapShotNode(txn *badger.Txn, height uint32, key []byte) ([]byte, error)
	// StoreSnapShotStateData stores a snapshot state element to the database
	StoreSnapShotStateData(txn *badger.Txn, key []byte, value []byte, data []byte) error
	// GetSnapShotStateData retrieves value corresponding to key from the State Data
	GetSnapShotStateData(txn *badger.Txn, key []byte) ([]byte, error)
	// FinalizeSnapShotRoot validates snapshot root and corresponding state
	FinalizeSnapShotRoot(txn *badger.Txn, root []byte, height uint32) error
	// BeginSnapShotSync deletes the entries of the database in preparation
	// for fast synchronization
	BeginSnapShotSync(txn *badger.Txn) error
	// FinalizeSync performs any final logic needed to terminate a fast sync
	FinalizeSync(txn *badger.Txn) error
}
