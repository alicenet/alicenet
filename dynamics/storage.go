package dynamics

import (
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

// Ensuring interface check
var _ StorageGetter = (*Storage)(nil)

/*
PROPOSAL ON CHAIN
PROPOSAL GETS VOTED ON
IF PROPOSAL PASSES IT BECOMES ACTIVE IN FUTURE ( EPOCH OF ACTIVE > EPOCH OF FINAL VOTE + 1 )
WHEN PROPOSAL PASSES AN EVENT IS EMITTED FROM THE GOVERNANCE CONTRACT
THIS EVENT IS OBSERVED BY THE NODES
THE NODES FETCH THE NEW VALUES AND STORE IN THE DATABASE FOR FUTURE USE
ON THE EPOCH BOUNDARY OF NOT ACTIVE TO ACTIVE, THE STORAGE STRUCT MUST BE UPDATED IN MEMORY FROM
 THE VALUES STORED IN THE DB
*/

// Dynamics contains the list of "constants" which may be changed
// dynamically to reflect protocol updates.
// The point is that these values are essentially constant but may be changed
// in future.

// StorageGetter is the interface that all Storage structs must match
// to be valid. These will be used to store the constants which may change
// each epoch as governance determines.
type StorageGetter interface {
	GetMaxBytes() uint32
	GetMaxProposalSize() uint32

	GetProposalStepTimeout() time.Duration
	GetPreVoteStepTimeout() time.Duration
	GetPreCommitStepTimeout() time.Duration
	GetDeadBlockRoundNextRoundTimeout() time.Duration
	GetDownloadTimeout() time.Duration
	GetSrvrMsgTimeout() time.Duration
	GetMsgTimeout() time.Duration

	UpdateStorage(*badger.Txn, Updater) error
	LoadStorage(*badger.Txn, uint32) error

	GetDataStoreEpochFee() *big.Int
	GetDataStoreValidVersion() uint32

	GetValueStoreFee() *big.Int
	GetValueStoreValidVersion() uint32

	GetAtomicSwapFee() *big.Int
	GetAtomicSwapValidStopEpoch() uint32

	GetMinTxFee() *big.Int
	GetTxValidVersion() uint32
}

// Storage is the struct which will implement the StorageGetter interface.
type Storage struct {
	sync.RWMutex
	database   *Database
	startChan  chan struct{}
	startOnce  sync.Once
	rawStorage *RawStorage
	logger     *logrus.Logger
}

// checkUpdate confirms the specified update is valid.
func checkUpdate(update Updater) error {
	if update.Epoch() == 0 {
		return ErrInvalidUpdateValue
	}
	rs := &RawStorage{}
	err := rs.UpdateValue(update)
	if err != nil {
		return err
	}
	return nil
}

// Init initializes the Storage structure.
func (s *Storage) Init(rawDB rawDataBase, logger *logrus.Logger) error {
	// initialize channel
	s.startChan = make(chan struct{})

	// initialize database
	s.database = &Database{rawDB: rawDB}

	// initialize logger
	s.logger = logger
	return nil
}

// Start allows normal operations to begin. This MUST be called after Init
// and can only be called once.
func (s *Storage) Start() {
	s.Lock()
	defer s.Unlock()
	s.startOnce.Do(func() {
		close(s.startChan)
	})
	s.rawStorage = &RawStorage{}
	s.rawStorage.standardParameters()
}

// UpdateStorage updates the database to include changes that must be made
// to the database
func (s *Storage) UpdateStorage(txn *badger.Txn, update Updater) error {
	<-s.startChan

	s.Lock()
	defer s.Unlock()

	err := checkUpdate(update)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}

	// Need to add code to check if initialization has been performed;
	// that is, is this the first call to UpdateStorage?
	_, err = s.database.GetLinkedList(txn)
	if err != nil {
		if !errors.Is(err, ErrKeyNotPresent) {
			utils.DebugTrace(s.logger, err)
			return err
		}
		rs := &RawStorage{}
		rs.standardParameters()
		node, ll, err := CreateLinkedList(1, rs)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
		err = s.database.SetLinkedList(txn, ll)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
		err = s.database.SetNode(txn, node)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
	}

	err = s.updateStorageValue(txn, update)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	return nil
}

// updateStorageValue updates the stored RawStorage values.
//
// We start at the Head of LinkedList and find the farthest point
// at which we need to update nodes.
// Once we find the beginning, we iterate forward and update all forward nodes.
func (s *Storage) updateStorageValue(txn *badger.Txn, update Updater) error {
	<-s.startChan

	epoch := update.Epoch()
	ll, err := s.database.GetLinkedList(txn)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	elu := ll.GetEpochLastUpdated()
	iterNode, err := s.database.GetNode(txn, elu)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}

	// firstNode denotes where we will begin looping forward
	// including the updated value;
	// this will not be used when we have a new Head
	firstNode := &Node{}
	// duplicateNode is used when we will copy values from the previous node
	// and then updating it to form a new node
	duplicateNode := &Node{}
	// newHead denotes if we need to update the Head of the LinkedList;
	// that is, if we need to update EpochLastUpdated
	newHead := false
	// addNode denotes whether we must add a node.
	// This will not occur if our update is on another node,
	// which could happen if multiple updates occur on one epoch.
	addNode := true

	// Loop backwards through the LinkedList to find firstNode and duplicateNode
	for {
		if epoch >= iterNode.thisEpoch {
			// We will use
			//
			// I = iterNode
			// U = updateNode
			// F = firstNode
			// D = duplicateNode
			// H = Head
			//
			// in our diagrams below.
			//
			// the update occurs in the current range
			if epoch == iterNode.thisEpoch {
				// the update occurs on a node; we do not need to add a node
				//
				//                         U
				//	                       F
				//	                       I
				// |---|---|---|---|---|---|---|---|---|---|---|---|
				firstNode, err = iterNode.Copy()
				if err != nil {
					utils.DebugTrace(s.logger, err)
					return err
				}
				addNode = false
			} else {
				// epoch > iterNode.thisEpoch
				if iterNode.IsHead() {
					// we will add a new node further in the future;
					// there will be no iteration.
					//
					//	       H
					//	       D
					//	       I               U
					// |---|---|---|---|---|---|---|---|---|---|---|---|
					newHead = true
				} else {
					// we start iterating at the node ahead.
					//
					//	       D
					//	       I               U               F
					// |---|---|---|---|---|---|---|---|---|---|---|---|
					firstNode, err = s.database.GetNode(txn, iterNode.nextEpoch)
					if err != nil {
						utils.DebugTrace(s.logger, err)
						return err
					}
				}
				duplicateNode, err = iterNode.Copy()
				if err != nil {
					utils.DebugTrace(s.logger, err)
					return err
				}
			}
			break
		}
		// If we have reached the tail node, then we do not have a node
		// for this specific epoch; we raise an error.
		if iterNode.IsTail() {
			// We cannot add an update before the first node
			return ErrInvalidUpdateValue
		}
		// We proceed backward in the linked list of nodes
		prevEpoch := iterNode.prevEpoch
		iterNode, err = s.database.GetNode(txn, prevEpoch)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
	}

	if addNode {
		// We need to add a new node, so we prepare
		node := &Node{
			thisEpoch: epoch,
		}
		// We compute the correct RawStorage value
		// We grab the RawStorage from duplicateNode and then update the value.
		rs, err := duplicateNode.rawStorage.Copy()
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
		err = rs.UpdateValue(update)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
		node.rawStorage = rs
		// We add the node to the database
		err = s.addNode(txn, node)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
	}

	if newHead {
		// We added a new Head, so we need to store this information
		// before we exit.
		err = ll.SetEpochLastUpdated(epoch)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
		err = s.database.SetLinkedList(txn, ll)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
		return nil
	}

	// We now iterate forward from firstNode and update all the nodes
	// to reflect the new values.
	iterNode, err = firstNode.Copy()
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}

	for {
		err = iterNode.rawStorage.UpdateValue(update)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
		err = s.database.SetNode(txn, iterNode)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
		if iterNode.IsHead() {
			break
		}
		nextEpoch := iterNode.nextEpoch
		iterNode, err = s.database.GetNode(txn, nextEpoch)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
	}

	return nil
}

// LoadStorage updates RawStorage to the correct value defined by the epoch.
//
// We will attempt to load the correct storage struct.
// If we receive ErrKeyNotPresent, then we return RawStorage
// with the standard parameters.
//
// We use Lock and Unlock rather than RLock and RUnlock because
// we modify Storage.
func (s *Storage) LoadStorage(txn *badger.Txn, epoch uint32) error {
	<-s.startChan

	s.Lock()
	defer s.Unlock()
	rs, err := s.loadStorage(txn, epoch)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	s.rawStorage, err = rs.Copy()
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	return nil
}

// loadStorage wraps loadRawStorage and ensures that a valid RawStorage
// value is returned if possible.
//
// When loadStorage is called by other functions, we should only need
// those functions to have RLock and RUnlock because we do not modify Storage.
func (s *Storage) loadStorage(txn *badger.Txn, epoch uint32) (*RawStorage, error) {
	rs, err := s.loadRawStorage(txn, epoch)
	if err != nil {
		if !errors.Is(err, ErrKeyNotPresent) {
			utils.DebugTrace(s.logger, err)
			return nil, err
		}
		rs = &RawStorage{}
		rs.standardParameters()
	}
	return rs, nil
}

// loadRawStorage looks for the appropriate RawStorage value in the database
// and returns that value.
//
// We start at the most updated epoch and proceed backwards until we arrive
// at the node with
//		epoch >= node.thisEpoch
func (s *Storage) loadRawStorage(txn *badger.Txn, epoch uint32) (*RawStorage, error) {
	if epoch == 0 {
		return nil, ErrZeroEpoch
	}
	ll, err := s.database.GetLinkedList(txn)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return nil, err
	}
	elu := ll.GetEpochLastUpdated()
	currentNode, err := s.database.GetNode(txn, elu)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return nil, err
	}

	// Loop backwards through the LinkedList
	for {
		if epoch >= currentNode.thisEpoch {
			rs, err := currentNode.rawStorage.Copy()
			if err != nil {
				utils.DebugTrace(s.logger, err)
				return nil, err
			}
			return rs, nil
		}
		// If we have reached the tail node, then we do not have a node
		// for this specific epoch; we raise an error.
		if currentNode.IsTail() {
			utils.DebugTrace(s.logger, ErrInvalid)
			return nil, ErrInvalid
		}
		// We proceed backward in the linked list of nodes
		prevEpoch := currentNode.prevEpoch
		currentNode, err = s.database.GetNode(txn, prevEpoch)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return nil, err
		}
	}
}

// addNode adds an additional node to the database.
// This node can be added anywhere.
// If the node is added at the head, then LinkedList must be updated
// to reflect this change.
func (s *Storage) addNode(txn *badger.Txn, node *Node) error {
	<-s.startChan

	// Ensure node.rawStorage and node.thisEpoch are valid;
	// other parameters should not be set.
	// This ensure that node.thisEpoch != 0
	if !node.IsPreValid() {
		return ErrInvalid
	}

	// Get LinkedList and Head
	ll, err := s.database.GetLinkedList(txn)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	elu := ll.GetEpochLastUpdated()
	currentNode, err := s.database.GetNode(txn, elu)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}

	if node.thisEpoch > currentNode.thisEpoch {
		// node to be added is strictly ahead of ELU
		err = s.addNodeHead(txn, node, currentNode)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
		return nil
	}

	if node.thisEpoch == currentNode.thisEpoch {
		// Node is already present; raise error
		return ErrInvalid
	}

	if currentNode.IsTail() {
		// The first node is always at epoch 1;
		// we cannot add a node before that.
		return ErrInvalid
	}

	// prevNode := &Node{}

	// Loop backwards through the LinkedList
	for {
		// Get previous node
		prevNode, err := s.database.GetNode(txn, currentNode.prevEpoch)
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
		if prevNode.thisEpoch < node.thisEpoch && node.thisEpoch < currentNode.thisEpoch {
			// We need to add node in between prevNode and currentNode
			err = s.addNodeSplit(txn, node, prevNode, currentNode)
			if err != nil {
				utils.DebugTrace(s.logger, err)
				return err
			}
			return nil
		}
		if node.thisEpoch == prevNode.thisEpoch {
			// Node is already present; raise error
			return ErrInvalid
		}
		if prevNode.IsTail() {
			// The first node is always at epoch 1;
			// we cannot add a node before that.
			return ErrInvalid
		}
		currentNode, err = prevNode.Copy()
		if err != nil {
			utils.DebugTrace(s.logger, err)
			return err
		}
	}
}

func (s *Storage) addNodeHead(txn *badger.Txn, node, headNode *Node) error {
	if !node.IsPreValid() || !headNode.IsValid() {
		return ErrInvalid
	}
	if !headNode.IsHead() || node.thisEpoch <= headNode.thisEpoch {
		// We require headNode to be head and node.thisEpoch < headNode.thisEpoch
		return ErrInvalid
	}
	err := node.SetEpochs(headNode, nil)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	// Store the nodes after changes have been made
	err = s.database.SetNode(txn, headNode)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	err = s.database.SetNode(txn, node)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}

	// Update EpochLastUpdated
	ll, err := s.database.GetLinkedList(txn)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	// We need to update EpochLastUpdated
	err = ll.SetEpochLastUpdated(node.thisEpoch)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	err = s.database.SetLinkedList(txn, ll)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	return nil
}

func (s *Storage) addNodeSplit(txn *badger.Txn, node, prevNode, nextNode *Node) error {
	if !node.IsPreValid() || !prevNode.IsValid() || !nextNode.IsValid() {
		return ErrInvalid
	}
	if (prevNode.thisEpoch >= node.thisEpoch) || (node.thisEpoch >= nextNode.thisEpoch) {
		return ErrInvalid
	}
	err := node.SetEpochs(prevNode, nextNode)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	// Store the nodes after changes have been made
	err = s.database.SetNode(txn, prevNode)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	err = s.database.SetNode(txn, nextNode)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	err = s.database.SetNode(txn, node)
	if err != nil {
		utils.DebugTrace(s.logger, err)
		return err
	}
	return nil
}

// GetMaxBytes returns the maximum allowed bytes
func (s *Storage) GetMaxBytes() uint32 {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetMaxBytes()
}

// GetMaxProposalSize returns the maximum size of bytes allowed in a proposal
func (s *Storage) GetMaxProposalSize() uint32 {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetMaxProposalSize()
}

// GetSrvrMsgTimeout returns the time before timeout of server message
func (s *Storage) GetSrvrMsgTimeout() time.Duration {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetSrvrMsgTimeout()
}

// GetMsgTimeout returns the timeout to receive a message
func (s *Storage) GetMsgTimeout() time.Duration {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetMsgTimeout()
}

// GetProposalStepTimeout returns the proposal step timeout
func (s *Storage) GetProposalStepTimeout() time.Duration {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetProposalStepTimeout()
}

// GetPreVoteStepTimeout returns the prevote step timeout
func (s *Storage) GetPreVoteStepTimeout() time.Duration {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetPreVoteStepTimeout()
}

// GetPreCommitStepTimeout returns the precommit step timeout
func (s *Storage) GetPreCommitStepTimeout() time.Duration {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetPreCommitStepTimeout()
}

// GetDeadBlockRoundNextRoundTimeout returns the timeout required before
// moving into the DeadBlockRound
func (s *Storage) GetDeadBlockRoundNextRoundTimeout() time.Duration {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetDeadBlockRoundNextRoundTimeout()
}

// GetDownloadTimeout returns the timeout for downloads
func (s *Storage) GetDownloadTimeout() time.Duration {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetDownloadTimeout()
}

// GetMinTxFee returns the minimum transaction fee.
func (s *Storage) GetMinTxFee() *big.Int {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetMinTxFee()
}

// GetTxValidVersion returns the transaction valid version
func (s *Storage) GetTxValidVersion() uint32 {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetTxValidVersion()
}

// GetValueStoreFee returns the transaction fee for ValueStore
func (s *Storage) GetValueStoreFee() *big.Int {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetValueStoreFee()
}

// GetValueStoreValidVersion returns the ValueStore valid version
func (s *Storage) GetValueStoreValidVersion() uint32 {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetValueStoreValidVersion()
}

// GetAtomicSwapFee returns the transaction fee for AtomicSwap
func (s *Storage) GetAtomicSwapFee() *big.Int {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetAtomicSwapFee()
}

// GetAtomicSwapValidStopEpoch returns the last epoch at which AtomicSwap is valid
func (s *Storage) GetAtomicSwapValidStopEpoch() uint32 {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetAtomicSwapValidStopEpoch()
}

// GetDataStoreEpochFee returns the DataStore fee per epoch
func (s *Storage) GetDataStoreEpochFee() *big.Int {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetDataStoreEpochFee()
}

// GetDataStoreValidVersion returns the DataStore valid version
func (s *Storage) GetDataStoreValidVersion() uint32 {
	<-s.startChan

	s.RLock()
	defer s.RUnlock()
	return s.rawStorage.GetDataStoreValidVersion()
}
