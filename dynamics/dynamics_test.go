package dynamics

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
)

func ChangeDynamicValues(s *Storage, epoch uint32, raw []byte) {
	err := s.database.rawDB.Update(func(txn *badger.Txn) error {
		err := s.ChangeDynamicValues(txn, epoch, raw)
		return err
	})
	if err != nil {
		panic(err)
	}
}

func UpdateCurrentDynamicValue(s *Storage, epoch uint32) {
	err := s.database.rawDB.Update(func(txn *badger.Txn) error {
		err := s.UpdateCurrentDynamicValue(txn, epoch)
		return err
	})
	if err != nil {
		panic(err)
	}
}

func InitializeStorage() *Storage {
	storageLogger := newLogger()
	mock := NewTestDB()

	s := &Storage{}
	err := s.Init(mock, storageLogger)
	if err != nil {
		panic(err)
	}

	return s
}

func InitializeStorageWithFirstNode() *Storage {
	s := InitializeStorage()
	epoch := uint32(1)
	raw, _ := GetDynamicValueMostZeros()
	ChangeDynamicValues(s, epoch, raw)
	return s
}

func InitializeStorageWithStandardNode() *Storage {
	storageLogger := newLogger()
	mock := NewTestDB()

	s := &Storage{}
	err := s.Init(mock, storageLogger)
	if err != nil {
		panic(err)
	}
	raw, _ := GetStandardDynamicValue()
	ChangeDynamicValues(s, 1, raw)
	return s

}

func GetDynamicValueMostZeros() ([]byte, *DynamicValues) {
	data, err := hex.DecodeString("000000000000000000000000002dc6c00000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		panic(err)
	}
	dv, err := DecodeDynamicValues(data)
	if err != nil {
		panic(err)
	}
	return data, dv
}

func GetDynamicValueWithFees() ([]byte, *DynamicValues) {
	data, err := hex.DecodeString("00000fa000000bb800000bb8002dc6c00000000000000bb80000000000000bb800000000000000000000000000000fa0")
	if err != nil {
		panic(err)
	}
	dv, err := DecodeDynamicValues(data)
	if err != nil {
		panic(err)
	}
	return data, dv
}

func GetStandardDynamicValue() ([]byte, *DynamicValues) {
	data, err := hex.DecodeString("00000fa000000bb800000bb8002dc6c00000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		panic(err)
	}
	dv, err := DecodeDynamicValues(data)
	if err != nil {
		panic(err)
	}
	return data, dv
}

// Test Storage Init with nothing initialized
func TestStorageInit1(t *testing.T) {
	t.Parallel()
	s := InitializeStorageWithStandardNode()

	_, dv := GetStandardDynamicValue()
	dvBytes, err := dv.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	// Check dynamicValues == standardParameters
	storageRSBytes, err := s.DynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dvBytes, storageRSBytes) {
		t.Fatal("dynamicValues values do not match")
	}

	select {
	case <-s.startChan:
	default:
		t.Fatal("Starting channel should be closed")
	}
}

func TestStorageInitFromScratch(t *testing.T) {
	t.Parallel()
	s := &Storage{}
	s.Init(NewTestDB(), newLogger())

	// since there's no node, the starting channel should be closed and no request
	// against the storage should be allowed
	select {
	case <-s.startChan:
		t.Fatal("Starting channel should be closed")
	default:
	}
}

// Test Storage Init with nothing initialized
func TestStorageInitWithPersistance(t *testing.T) {
	t.Parallel()
	s := InitializeStorageWithStandardNode()

	_, dv := GetStandardDynamicValue()
	dvBytes, err := dv.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	s2 := &Storage{}
	s2.Init(s.database.rawDB, s.logger)

	// Check dynamicValues == standardParameters
	storageDVBytes, err := s2.DynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dvBytes, storageDVBytes) {
		t.Fatal("dynamicValues values do not match")
	}

	select {
	case <-s2.startChan:
	default:
		t.Fatal("Starting channel should be closed")
	}

	assert.Equal(t, s2.GetMaxBlockSize(), uint32(3_000_000))

}

func TestStorageStartGood(t *testing.T) {
	t.Parallel()
	storageLogger := newLogger()
	mock := NewTestDB()

	s := &Storage{}
	err := s.Init(mock, storageLogger)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-s.startChan:
		t.Fatal("Starting channel should be closed")
	default:
	}
	raw, _ := GetStandardDynamicValue()
	ChangeDynamicValues(s, 1, raw)
	select {
	case <-s.startChan:
	default:
		t.Fatal("Starting channel should be closed")
	}
}

// Test ensures we panic when trying to add a value before Init.
// This happens from attempting to close a closed channel.
func TestStorageStartFail(t *testing.T) {
	t.Parallel()
	s := &Storage{}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Should panic")
		}
	}()
	raw, _ := GetStandardDynamicValue()
	ChangeDynamicValues(s, 1, raw)
}

// Test ensures storage has is initialized to the correct values.
func TestStorageInitialized(t *testing.T) {
	t.Parallel()
	s := InitializeStorage()
	raw, _ := GetDynamicValueWithFees()
	ChangeDynamicValues(s, 1, raw)

	InitialMaxBlockSize := uint32(3_000_000)
	InitialProposalTimeout := time.Duration(4000 * time.Millisecond)
	InitialPreVoteTimeout := time.Duration(3000 * time.Millisecond)
	InitialPreCommitTimeout := time.Duration(3000 * time.Millisecond)
	InitialDataStoreFee := new(big.Int).SetInt64(3000)
	InitialValueStoreFee := new(big.Int).SetInt64(3000)
	InitialMinTxFee := new(big.Int).SetInt64(4000)

	maxBytesReturned := s.GetMaxBlockSize()
	if maxBytesReturned != InitialMaxBlockSize {
		t.Fatal("Incorrect MaxBytes")
	}

	maxProposalSizeReturned := s.GetMaxProposalSize()
	if maxProposalSizeReturned != InitialMaxBlockSize {
		t.Fatal("Incorrect MaxProposalSize")
	}

	proposalTimeoutReturned := s.GetProposalTimeout()
	if proposalTimeoutReturned != InitialProposalTimeout {
		t.Fatal("Incorrect proposalTO")
	}

	preVoteTimeoutReturned := s.GetPreVoteTimeout()
	if preVoteTimeoutReturned != InitialPreVoteTimeout {
		t.Fatal("Incorrect preVoteTO")
	}

	preCommitTimeoutReturned := s.GetPreCommitTimeout()
	if preCommitTimeoutReturned != InitialPreCommitTimeout {
		t.Fatal("Incorrect preCommitStepTO")
	}

	deadBlockRoundNextRoundTimeoutReturned := s.GetDeadBlockRoundNextRoundTimeout()
	sum := InitialProposalTimeout + InitialPreVoteTimeout + InitialPreCommitTimeout
	dBRNRTO := (5 * sum) / 2
	if deadBlockRoundNextRoundTimeoutReturned != dBRNRTO {
		t.Fatal("Incorrect deadBlockRoundNextRoundTimeout")
	}

	downloadTimeoutReturned := s.GetDownloadTimeout()
	if downloadTimeoutReturned != sum {
		t.Fatal("Incorrect downloadTimeout")
	}

	minTxFeeReturned := s.GetMinScaledTransactionFee()
	if minTxFeeReturned.Cmp(InitialMinTxFee) != 0 {
		t.Fatal("Incorrect minTxFee")
	}

	vsFeeReturned := s.GetValueStoreFee()
	if vsFeeReturned.Cmp(InitialValueStoreFee) != 0 {
		t.Fatal("Incorrect valueStoreFee")
	}

	dsEpochFeeReturned := s.GetDataStoreFee()
	if dsEpochFeeReturned.Cmp(InitialDataStoreFee) != 0 {
		t.Fatal("Incorrect dataStoreEpochFee")
	}

}

// Test success of UpdateCurrentDynamicValue
func TestStorageUpdateCurrentDynamicValueGood1(t *testing.T) {
	t.Parallel()
	s := InitializeStorageWithStandardNode()
	epoch := uint32(25519)

	_, dvTrue := GetStandardDynamicValue()
	dvTrueBytes, err := dvTrue.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	dvBytes, err := s.DynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dvBytes, dvTrueBytes) {
		t.Fatal("dynamicValues values do not match")
	}
	// nothing should happen, since there's no node update
	UpdateCurrentDynamicValue(s, epoch)

	dvBytes, err = s.DynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dvBytes, dvTrueBytes) {
		t.Fatal("dynamicValues values do not match")
	}
}

func TestStorageUpdateCurrentDynamicValueWithUpdate(t *testing.T) {
	t.Parallel()
	s := InitializeStorageWithStandardNode()
	_, dvTrue := GetStandardDynamicValue()
	dvTrueBytes, err := dvTrue.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	dvBytes, err := s.DynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dvBytes, dvTrueBytes) {
		t.Fatal("dynamicValues values do not match")
	}

	epoch := uint32(25519)
	newValueRaw, dvTrue := GetDynamicValueMostZeros()
	dvTrueBytes, err = dvTrue.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	// adding the new node
	ChangeDynamicValues(s, epoch, newValueRaw)

	// after we reached that epoch values should be updated
	UpdateCurrentDynamicValue(s, epoch)

	dvBytes, err = s.DynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dvBytes, dvTrueBytes) {
		t.Fatal("dynamicValues values do not match (2)")
	}
}

// Test failure of LoadStorage
func TestStorageLoadStorageBad1(t *testing.T) {
	t.Parallel()
	s := InitializeStorageWithFirstNode()
	// We attempt to load the zero epoch;
	// this should raise an error.
	err := s.database.rawDB.Update(func(txn *badger.Txn) error {
		err := s.UpdateCurrentDynamicValue(txn, 0)
		return err
	})
	if !errors.Is(err, ErrZeroEpoch) {
		t.Fatalf("Should have raised ErrZeroEpoch raise instead: %v", err)
	}
}

// Test failure of loadDynamicValues again.
// It should not be possible to reach this configuration.
func TestStorageLoadStorageBad2(t *testing.T) {
	t.Parallel()
	storageLogger := newLogger()
	database := initializeDB()
	ll := &LinkedList{
		currentValue: 1,
		tail:         1,
	}
	err := database.rawDB.Update(func(txn *badger.Txn) error {
		return database.SetLinkedList(txn, ll)
	})
	assert.Nil(t, err)

	s := &Storage{}
	s.startChan = make(chan struct{})
	close(s.startChan)
	s.database = database
	s.logger = storageLogger

	// We attempt to load an epoch from an empty database (without nodes but LinkedList set);
	// this should raise an error.
	err = s.database.rawDB.Update(func(txn *badger.Txn) error {
		err := s.UpdateCurrentDynamicValue(txn, 1)
		return err
	})
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeTailGood(t *testing.T) {
	t.Parallel()
	// Initialize storage and have standard node at epoch 1
	s := InitializeStorageWithStandardNode()
	epoch := uint32(1)
	_, dv := GetStandardDynamicValue()
	dvBytes, err := dv.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	newEpoch := epoch + 2
	dvNewRaw, dvNew := GetDynamicValueWithFees()
	dvNewBytes, err := dvNew.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	ChangeDynamicValues(s, newEpoch, dvNewRaw)

	err = s.database.rawDB.View(func(txn *badger.Txn) error {
		origNode, err := s.database.GetNode(txn, epoch)
		if err != nil {
			t.Fatal(err)
		}
		if origNode.prevEpoch != 0 || origNode.thisEpoch != epoch || origNode.nextEpoch != newEpoch {
			t.Fatal("origNode invalid (1)")
		}
		dvOrigBytes, err := origNode.dynamicValues.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(dvOrigBytes, dvBytes) {
			t.Fatal("origNode invalid (2)")
		}

		retNode, err := s.database.GetNode(txn, newEpoch)
		if err != nil {
			t.Fatal(err)
		}
		if retNode.prevEpoch != epoch || retNode.thisEpoch != newEpoch || retNode.nextEpoch != 0 {
			t.Fatal("retNode invalid (1)")
		}
		retBytes, err := retNode.dynamicValues.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(retBytes, dvNewBytes) {
			t.Fatal("retNode invalid (2)")
		}
		return nil
	})
	assert.Nil(t, err)
}

// func TestStorageAddEmptyNodeHead(t *testing.T) {
// 	t.Parallel()
// 	origEpoch := uint32(1)
// 	s := InitializeStorageWithFirstNode()

// 	err := s.database.rawDB.Update(func(txn *badger.Txn) error {
// 		ll, err := s.database.GetLinkedList(txn)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		headNode, err := s.database.GetNode(txn, origEpoch)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		if !headNode.IsValid() {
// 			t.Fatal("headNode should be valid")
// 		}
// 		if !headNode.IsTail() || !headNode.IsHead() {
// 			t.Fatal("first should be tail and head")
// 		}
// 		err = s.addNode(txn, ll, origEpoch+1, &DynamicValues{})
// 		if err != nil {
// 			t.Fatalf("Should have raised error %v", err)
// 			t.Fatal("Should have raised error")
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// }

// func TestStorageAddNodeHeadBad2(t *testing.T) {
// 	origEpoch := uint32(1)
// 	s := InitializeStorageWithFirstNode()
// 	headNode, err := s.database.GetNode(nil, origEpoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !headNode.IsValid() {
// 		t.Fatal("headNode should be valid")
// 	}

// 	rs := &DynamicValues{}
// 	node := &Node{
// 		thisEpoch:     1,
// 		dynamicValues: rs,
// 	}
// 	if !node.IsPreValid() {
// 		t.Fatal("node should be prevalid")
// 	}

// 	err = s.addNodeHead(nil, node, headNode)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// func TestStorageAddNodeSplitGood(t *testing.T) {
// 	first := uint32(1)
// 	s := InitializeStorageWithFirstNode()
// 	prevNode, err := s.database.GetNode(nil, first)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !prevNode.IsValid() {
// 		t.Fatal("prevNode should be valid")
// 	}
// 	last := uint32(10)
// 	prevNode.nextEpoch = last
// 	err = s.database.SetNode(nil, prevNode)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Set up nextnode
// 	rs := &DynamicValues{}
// 	rs.standardParameters()
// 	rsBytes, err := rs.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	nextNode := &Node{
// 		prevEpoch:     first,
// 		thisEpoch:     last,
// 		nextEpoch:     last,
// 		dynamicValues: rs,
// 	}
// 	if !nextNode.IsValid() {
// 		t.Fatal("nextNode should be valid")
// 	}
// 	err = s.database.SetNode(nil, nextNode)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Set up node
// 	rsNew, err := rs.Copy()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	rsNew.MaxBytes = 123456
// 	rsNewBytes, err := rsNew.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	newEpoch := uint32(5)
// 	node := &Node{
// 		thisEpoch:     newEpoch,
// 		dynamicValues: rsNew,
// 	}
// 	if !node.IsPreValid() {
// 		t.Fatal("node should be preValid")
// 	}

// 	err = s.addNodeSplit(nil, node, prevNode, nextNode)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Check everything
// 	firstNode, err := s.database.GetNode(nil, first)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if firstNode.prevEpoch != first || firstNode.thisEpoch != first || firstNode.nextEpoch != newEpoch {
// 		t.Fatal("firstNode invalid (1)")
// 	}
// 	retBytes, err := firstNode.dynamicValues.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !bytes.Equal(retBytes, rsBytes) {
// 		t.Fatal("firstNode invalid (2)")
// 	}

// 	middleNode, err := s.database.GetNode(nil, newEpoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if middleNode.prevEpoch != first || middleNode.thisEpoch != newEpoch || middleNode.nextEpoch != last {
// 		t.Fatal("middleNode invalid (1)")
// 	}
// 	retBytes, err = middleNode.dynamicValues.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !bytes.Equal(retBytes, rsNewBytes) {
// 		t.Fatal("middleNode invalid (2)")
// 	}

// 	lastNode, err := s.database.GetNode(nil, last)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if lastNode.prevEpoch != newEpoch || lastNode.thisEpoch != last || lastNode.nextEpoch != last {
// 		t.Fatal("lastNode invalid (1)")
// 	}
// 	retBytes, err = lastNode.dynamicValues.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !bytes.Equal(retBytes, rsBytes) {
// 		t.Fatal("lastNode invalid (2)")
// 	}
// }

// func TestStorageAddNodeSplitBad1(t *testing.T) {
// 	first := uint32(1)
// 	s := InitializeStorageWithFirstNode()
// 	prevNode, err := s.database.GetNode(nil, first)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !prevNode.IsValid() {
// 		t.Fatal("prevNode should be valid")
// 	}
// 	last := uint32(10)
// 	prevNode.nextEpoch = last
// 	err = s.database.SetNode(nil, prevNode)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Set up nodes
// 	rs := &DynamicValues{}
// 	rs.standardParameters()
// 	nextNode := &Node{
// 		prevEpoch:     first,
// 		thisEpoch:     last,
// 		nextEpoch:     last,
// 		dynamicValues: rs,
// 	}
// 	if !nextNode.IsValid() {
// 		t.Fatal("nextNode should be valid")
// 	}
// 	err = s.database.SetNode(nil, nextNode)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	node := &Node{}

// 	err = s.addNodeSplit(nil, node, prevNode, nextNode)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// func TestStorageAddNodeSplitBad2(t *testing.T) {
// 	first := uint32(1)
// 	s := InitializeStorageWithFirstNode()
// 	prevNode, err := s.database.GetNode(nil, first)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !prevNode.IsValid() {
// 		t.Fatal("prevNode should be valid")
// 	}
// 	last := uint32(10)
// 	prevNode.nextEpoch = last
// 	err = s.database.SetNode(nil, prevNode)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Set up nodes
// 	rs := &DynamicValues{}
// 	rs.standardParameters()
// 	nextNode := &Node{
// 		prevEpoch:     first,
// 		thisEpoch:     last,
// 		nextEpoch:     last,
// 		dynamicValues: rs,
// 	}
// 	if !nextNode.IsValid() {
// 		t.Fatal("nextNode should be valid")
// 	}
// 	err = s.database.SetNode(nil, nextNode)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rsNew, err := rs.Copy()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	rsNew.MaxBytes = 123456
// 	newEpoch := uint32(100)
// 	node := &Node{
// 		thisEpoch:     newEpoch,
// 		dynamicValues: rsNew,
// 	}
// 	if !node.IsPreValid() {
// 		t.Fatal("node should be preValid")
// 	}

// 	err = s.addNodeSplit(nil, node, prevNode, nextNode)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// // Test addNode when adding to Head
// func TestStorageAddNodeGood1(t *testing.T) {
// 	origEpoch := uint32(1)
// 	s := InitializeStorageWithFirstNode()
// 	rs := &DynamicValues{}
// 	rs.standardParameters()
// 	rsStandardBytes, err := rs.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	newMaxBytes := uint32(12345)
// 	rs.MaxBytes = newMaxBytes
// 	epoch := uint32(10)
// 	newNode := &Node{
// 		prevEpoch:     0,
// 		thisEpoch:     epoch,
// 		nextEpoch:     0,
// 		dynamicValues: rs,
// 	}
// 	err = s.addNode(nil, newNode)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	rsNewBytes, err := rs.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Check everything
// 	origNode, err := s.database.GetNode(nil, origEpoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if origNode.prevEpoch != origEpoch {
// 		t.Fatal("origNode.prevEpoch is invalid")
// 	}
// 	if origNode.thisEpoch != origEpoch {
// 		t.Fatal("origNode.thisEpoch is invalid")
// 	}
// 	if origNode.nextEpoch != epoch {
// 		t.Fatal("origNode.nextEpoch is invalid")
// 	}
// 	retRS, err := origNode.dynamicValues.Copy()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	retRSBytes, err := retRS.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !bytes.Equal(retRSBytes, rsStandardBytes) {
// 		t.Fatal("invalid DynamicValues")
// 	}

// 	addedNode, err := s.database.GetNode(nil, epoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if addedNode.prevEpoch != origEpoch {
// 		t.Fatal("addedNode.prevEpoch is invalid")
// 	}
// 	if addedNode.thisEpoch != epoch {
// 		t.Fatal("addedNode.thisEpoch is invalid")
// 	}
// 	if addedNode.nextEpoch != epoch {
// 		t.Fatal("addedNode.nextEpoch is invalid")
// 	}
// 	retRS, err = addedNode.dynamicValues.Copy()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	retRSBytes, err = retRS.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !bytes.Equal(retRSBytes, rsNewBytes) {
// 		t.Fatal("invalid DynamicValues (2)")
// 	}
// }

// // Test addNode when adding to Head and then in between
// func TestStorageAddNodeGood2(t *testing.T) {
// 	origEpoch := uint32(1)
// 	s := InitializeStorageWithFirstNode()
// 	rs := &DynamicValues{}
// 	rs.standardParameters()
// 	rsStandardBytes, err := rs.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	newEpoch := uint32(100)
// 	newNode := &Node{
// 		prevEpoch:     0,
// 		thisEpoch:     newEpoch,
// 		nextEpoch:     0,
// 		dynamicValues: rs,
// 	}
// 	err = s.addNode(nil, newNode)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	newEpoch2 := uint32(10)
// 	newMaxBytes := uint32(12345689)
// 	rs.SetMaxBytes(newMaxBytes)
// 	newNode2 := &Node{
// 		prevEpoch:     0,
// 		thisEpoch:     newEpoch2,
// 		nextEpoch:     0,
// 		dynamicValues: rs,
// 	}
// 	err = s.addNode(nil, newNode2)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	rsNewBytes, err := rs.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Check everything
// 	origNode, err := s.database.GetNode(nil, origEpoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if origNode.prevEpoch != origEpoch {
// 		t.Fatal("origNode.prevEpoch is invalid")
// 	}
// 	if origNode.thisEpoch != origEpoch {
// 		t.Fatal("origNode.thisEpoch is invalid")
// 	}
// 	if origNode.nextEpoch != newEpoch2 {
// 		t.Fatal("origNode.nextEpoch is invalid")
// 	}
// 	retRS, err := origNode.dynamicValues.Copy()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	retRSBytes, err := retRS.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !bytes.Equal(retRSBytes, rsStandardBytes) {
// 		t.Fatal("invalid DynamicValues")
// 	}

// 	addedNode, err := s.database.GetNode(nil, newEpoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if addedNode.prevEpoch != newEpoch2 {
// 		t.Fatal("addedNode.prevEpoch is invalid")
// 	}
// 	if addedNode.thisEpoch != newEpoch {
// 		t.Fatal("addedNode.thisEpoch is invalid")
// 	}
// 	if addedNode.nextEpoch != newEpoch {
// 		t.Fatal("addedNode.nextEpoch is invalid")
// 	}
// 	retRS, err = addedNode.dynamicValues.Copy()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	retRSBytes, err = retRS.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !bytes.Equal(retRSBytes, rsStandardBytes) {
// 		t.Fatal("invalid DynamicValues (2)")
// 	}

// 	addedNode2, err := s.database.GetNode(nil, newEpoch2)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if addedNode2.prevEpoch != origEpoch {
// 		t.Fatal("addedNode2.prevEpoch is invalid")
// 	}
// 	if addedNode2.thisEpoch != newEpoch2 {
// 		t.Fatal("addedNode2.thisEpoch is invalid")
// 	}
// 	if addedNode2.nextEpoch != newEpoch {
// 		t.Fatal("addedNode2.nextEpoch is invalid")
// 	}
// 	retRS, err = addedNode2.dynamicValues.Copy()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	retRSBytes, err = retRS.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !bytes.Equal(retRSBytes, rsNewBytes) {
// 		t.Fatal("invalid DynamicValues (3)")
// 	}
// }

// func TestStorageAddNodeBad1(t *testing.T) {
// 	s := InitializeStorage()
// 	rs := &DynamicValues{}
// 	newNode := &Node{
// 		prevEpoch:     0,
// 		thisEpoch:     0,
// 		nextEpoch:     0,
// 		dynamicValues: rs,
// 	}
// 	err := s.addNode(nil, newNode)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// func TestStorageAddNodeBad2(t *testing.T) {
// 	s := InitializeStorage()
// 	rs := &DynamicValues{}
// 	newNode := &Node{
// 		prevEpoch:     0,
// 		thisEpoch:     1,
// 		nextEpoch:     0,
// 		dynamicValues: rs,
// 	}
// 	err := s.addNode(nil, newNode)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// func TestStorageAddNodeBad3(t *testing.T) {
// 	s := InitializeStorageWithFirstNode()
// 	rs := &DynamicValues{}
// 	newNode := &Node{
// 		prevEpoch:     0,
// 		thisEpoch:     1,
// 		nextEpoch:     0,
// 		dynamicValues: rs,
// 	}
// 	err := s.addNode(nil, newNode)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// func TestStorageAddNodeBad4(t *testing.T) {
// 	s := InitializeStorageWithFirstNode()
// 	rs := &DynamicValues{}
// 	newNode := &Node{
// 		prevEpoch:     0,
// 		thisEpoch:     257,
// 		nextEpoch:     0,
// 		dynamicValues: rs,
// 	}
// 	err := s.addNode(nil, newNode)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	newNode2 := &Node{
// 		prevEpoch:     0,
// 		thisEpoch:     1,
// 		nextEpoch:     0,
// 		dynamicValues: rs,
// 	}
// 	err = s.addNode(nil, newNode2)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// func TestStorageAddNodeBad5(t *testing.T) {
// 	s := InitializeStorageWithFirstNode()
// 	rs := &DynamicValues{}
// 	newNode := &Node{
// 		prevEpoch:     0,
// 		thisEpoch:     257,
// 		nextEpoch:     0,
// 		dynamicValues: rs,
// 	}
// 	err := s.addNode(nil, newNode)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	newNode2 := &Node{
// 		prevEpoch:     0,
// 		thisEpoch:     25519,
// 		nextEpoch:     0,
// 		dynamicValues: rs,
// 	}
// 	err = s.addNode(nil, newNode2)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	newNode3 := &Node{
// 		prevEpoch:     0,
// 		thisEpoch:     1,
// 		nextEpoch:     0,
// 		dynamicValues: rs,
// 	}
// 	err = s.addNode(nil, newNode3)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// func TestStorageAddNodeBad6(t *testing.T) {
// 	s := InitializeStorage()
// 	ll := &LinkedList{25519}
// 	err := s.database.SetLinkedList(nil, ll)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rs := &DynamicValues{}
// 	newNode := &Node{
// 		prevEpoch:     0,
// 		thisEpoch:     1,
// 		nextEpoch:     0,
// 		dynamicValues: rs,
// 	}
// 	err = s.addNode(nil, newNode)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// // Test failure of UpdateStorage
// func TestStorageUpdateStorageBad(t *testing.T) {
// 	s := InitializeStorage()
// 	epoch := uint32(25519)
// 	field := "invalid"
// 	value := ""
// 	update := &Update{
// 		name:  field,
// 		value: value,
// 		epoch: epoch,
// 	}
// 	err := s.UpdateStorage(nil, update)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// // Test success of UpdateStorage
// func TestStorageUpdateStorageGood1(t *testing.T) {
// 	s := InitializeStorage()
// 	epoch := uint32(1)
// 	field := "maxBytes"
// 	value := "123456789"
// 	update, err := NewUpdate(field, value, epoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = s.UpdateStorage(nil, update)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rs := &DynamicValues{}
// 	rs.standardParameters()
// 	i, err := strconv.Atoi(value)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	rs.MaxBytes = uint32(i)
// 	rs.MaxProposalSize = uint32(i)
// 	rsBytes, err := rs.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	err = s.UpdateCurrentDynamicValue(nil, epoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	retRS, err := s.DynamicValues.Copy()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	retRSBytes, err := retRS.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if !bytes.Equal(rsBytes, retRSBytes) {
// 		t.Log(string(rsBytes))
// 		t.Log(string(retRSBytes))
// 		t.Fatal("invalid DynamicValues")
// 	}
// }

// // Test success of UpdateStorage
// func TestStorageUpdateStorageGood2(t *testing.T) {
// 	s := InitializeStorage()
// 	epoch := uint32(25519)
// 	field := "maxBytes"
// 	value := "123456789"
// 	update, err := NewUpdate(field, value, epoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = s.UpdateStorage(nil, update)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rs := &DynamicValues{}
// 	rs.standardParameters()
// 	i, err := strconv.Atoi(value)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	rs.SetMaxBytes(uint32(i))
// 	rsBytes, err := rs.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	err = s.UpdateCurrentDynamicValue(nil, epoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	retRS, err := s.DynamicValues.Copy()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	retRSBytes, err := retRS.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if !bytes.Equal(rsBytes, retRSBytes) {
// 		t.Log(string(rsBytes))
// 		t.Log(string(retRSBytes))
// 		t.Fatal("invalid DynamicValues")
// 	}
// }

// // Test success of UpdateStorage
// func TestStorageUpdateStorageGood3(t *testing.T) {
// 	s := InitializeStorageWithFirstNode()

// 	// Add another epoch in the future
// 	epoch2 := uint32(25519)
// 	rs2 := &DynamicValues{}
// 	rs2.standardParameters()
// 	newPropTO := 13 * time.Second
// 	rs2.SetProposalStepTimeout(newPropTO)
// 	node2 := &Node{
// 		thisEpoch:     epoch2,
// 		dynamicValues: rs2,
// 	}
// 	err := s.addNode(nil, node2)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Update initial storage value
// 	epoch := uint32(1)
// 	field := "maxBytes"
// 	value := "123456789"
// 	update, err := NewUpdate(field, value, epoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = s.UpdateStorage(nil, update)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	_, rsTrue := GetStandardDynamicValue()
// 	newMaxBytesInt, err := strconv.Atoi(value)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	rsTrue.SetMaxBytes(uint32(newMaxBytesInt))
// 	rsTrueBytes, err := rsTrue.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Test Epoch1
// 	err = s.UpdateCurrentDynamicValue(nil, epoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	retRS, err := s.DynamicValues.Copy()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	retRSBytes, err := retRS.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !bytes.Equal(rsTrueBytes, retRSBytes) {
// 		t.Log(string(rsTrueBytes))
// 		t.Log(string(retRSBytes))
// 		t.Fatal("invalid DynamicValues (1)")
// 	}

// 	rsTrue2 := &DynamicValues{}
// 	rsTrue2.standardParameters()
// 	rsTrue2.SetMaxBytes(uint32(newMaxBytesInt))
// 	rsTrue2.SetProposalStepTimeout(newPropTO)
// 	rsTrue2Bytes, err := rsTrue2.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Test Epoch2
// 	err = s.UpdateCurrentDynamicValue(nil, epoch2)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	retRS, err = s.DynamicValues.Copy()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	retRSBytes, err = retRS.Marshal()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !bytes.Equal(rsTrue2Bytes, retRSBytes) {
// 		t.Log(string(rsTrue2Bytes))
// 		t.Log(string(retRSBytes))
// 		t.Fatal("invalid DynamicValues (2)")
// 	}
// }

// // Test failure of UpdateStorageValue
// // Attempt to perform invalid update at future epoch
// func TestStorageUpdateStorageValueBad1(t *testing.T) {
// 	s := InitializeStorage()
// 	epoch := uint32(25519)
// 	field := "invalid"
// 	value := ""
// 	update := &Update{
// 		name:  field,
// 		value: value,
// 		epoch: epoch,
// 	}
// 	err := s.updateStorageValue(nil, update)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// // Test failure of UpdateStorageValue
// // Attempt to perform invalid update at current epoch
// func TestStorageUpdateStorageValueBad2(t *testing.T) {
// 	s := InitializeStorage()
// 	epoch := uint32(1)
// 	field := "invalid"
// 	value := ""
// 	update := &Update{
// 		name:  field,
// 		value: value,
// 		epoch: epoch,
// 	}
// 	err := s.updateStorageValue(nil, update)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// // Test failure of UpdateStorageValue
// // Attempt to perform invalid update at previous epoch
// func TestStorageUpdateStorageValueBad3(t *testing.T) {
// 	s := InitializeStorage()
// 	epoch := uint32(1)
// 	field := "invalid"
// 	value := ""
// 	update := &Update{
// 		name:  field,
// 		value: value,
// 		epoch: epoch,
// 	}
// 	err := s.updateStorageValue(nil, update)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// // Test failure of UpdateStorageValue
// // Attempt to perform invalid update in between two valid storage values
// func TestStorageUpdateStorageValueBad4(t *testing.T) {
// 	s := InitializeStorageWithFirstNode()
// 	rs := &DynamicValues{}
// 	rs.standardParameters()
// 	node := &Node{
// 		thisEpoch:     25519,
// 		dynamicValues: rs,
// 	}

// 	err := s.addNode(nil, node)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	epoch := uint32(257)
// 	field := "invalid"
// 	value := ""
// 	update := &Update{
// 		name:  field,
// 		value: value,
// 		epoch: epoch,
// 	}
// 	err = s.updateStorageValue(nil, update)
// 	if err == nil {
// 		t.Fatal("Should have raised error")
// 	}
// }

// func TestStorageGetMinTxFee(t *testing.T) {
// 	s := InitializeStorageWithFirstNode()
// 	txFee := s.GetMinTxFee()
// 	if txFee.Cmp(minTxFee) != 0 {
// 		t.Fatal("txFee incorrect")
// 	}

// 	epoch := uint32(25519)
// 	field := "minTxFee"
// 	value := "123456789"
// 	update, err := NewUpdate(field, value, epoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = s.UpdateStorage(nil, update)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = s.UpdateCurrentDynamicValue(nil, epoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	valueTrue, err := stringToBigInt(value)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	txFee = s.GetMinTxFee()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if txFee.Cmp(valueTrue) != 0 {
// 		t.Fatal("incorrect txFee")
// 	}
// }

// func TestStorageGetDataStoreEpochFee(t *testing.T) {
// 	s := InitializeStorageWithFirstNode()
// 	dsEpochFee := s.GetDataStoreEpochFee()
// 	if dsEpochFee.Cmp(dataStoreEpochFee) != 0 {
// 		t.Fatal("dsEpochFee incorrect")
// 	}

// 	epoch := uint32(25519)
// 	field := "dataStoreEpochFee"
// 	value := "123456789"
// 	update, err := NewUpdate(field, value, epoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = s.UpdateStorage(nil, update)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = s.UpdateCurrentDynamicValue(nil, epoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	valueTrue, err := stringToBigInt(value)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	dsEpochFee = s.GetDataStoreEpochFee()
// 	if dsEpochFee.Cmp(valueTrue) != 0 {
// 		t.Fatal("incorrect dataStoreEpochFee")
// 	}
// }

// func TestStorageGetValueStoreFee(t *testing.T) {
// 	s := InitializeStorageWithFirstNode()
// 	vsFee := s.GetValueStoreFee()
// 	if vsFee.Cmp(valueStoreFee) != 0 {
// 		t.Fatal("vsFee incorrect")
// 	}

// 	epoch := uint32(25519)
// 	field := "valueStoreFee"
// 	value := "123456789"
// 	update, err := NewUpdate(field, value, epoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = s.UpdateStorage(nil, update)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = s.UpdateCurrentDynamicValue(nil, epoch)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	valueTrue, err := stringToBigInt(value)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	vsFee = s.GetValueStoreFee()
// 	if vsFee.Cmp(valueTrue) != 0 {
// 		t.Fatal("incorrect valueStoreFee")
// 	}
// }
