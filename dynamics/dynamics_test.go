package dynamics

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"
	"testing"
	"time"
)

func InitializeStorage() *Storage {
	storageLogger := newLogger()
	mock := &MockRawDB{}
	mock.rawDB = make(map[string]string)

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
	data, err := hex.DecodeString("000000000000000000000000002dc6c000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		panic(err)
	}
	err = s.ChangeDynamicValues(nil, epoch, data)
	if err != nil {
		panic(err)
	}
	return s
}

func InitializeStorageWithStandardNode() *Storage {
	storageLogger := newLogger()
	mock := &MockRawDB{}
	mock.rawDB = make(map[string]string)

	s := &Storage{}
	err := s.Init(mock, storageLogger)
	if err != nil {
		panic(err)
	}

	err = s.ChangeDynamicValues(nil, 1, GetStandardDynamicValueRaw())
	if err != nil {
		panic(err)
	}
	return s

}

func GetStandardDynamicValue() *DynamicValues {
	data, err := hex.DecodeString("00000fa000000bb800000bb8002dc6c000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		panic(err)
	}
	dv, err := DecodeDynamicValues(data)
	if err != nil {
		panic(err)
	}
	return dv
}

func GetStandardDynamicValueRaw() []byte {
	data, err := hex.DecodeString("00000fa000000bb800000bb8002dc6c000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		panic(err)
	}
	return data
}

// Test Storage Init with nothing initialized
func TestStorageInit1(t *testing.T) {
	storageLogger := newLogger()
	mock := &MockRawDB{}
	mock.rawDB = make(map[string]string)

	s := &Storage{}
	err := s.Init(mock, storageLogger)
	if err != nil {
		t.Fatal(err)
	}

	dv := GetStandardDynamicValue()
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
}

func TestStorageStartGood(t *testing.T) {
	storageLogger := newLogger()
	mock := &MockRawDB{}
	mock.rawDB = make(map[string]string)

	s := &Storage{}
	err := s.Init(mock, storageLogger)
	if err != nil {
		t.Fatal(err)
	}

	s.ChangeDynamicValues(nil, 1, GetStandardDynamicValueRaw())

}

// Test ensures we panic when trying to add a value before Init.
// This happens from attempting to close a closed channel.
func TestStorageStartFail(t *testing.T) {
	s := &Storage{}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Should panic")
		}
	}()
	s.ChangeDynamicValues(nil, 1, GetStandardDynamicValueRaw())
}

// Test ensures storage has is initialized to the correct values.
func TestStorageInitialized(t *testing.T) {
	s := InitializeStorageWithStandardNode()

	InitialMaxBlockSize := uint32(30000)
	InitialProposalTimeout := time.Duration(4000 * time.Millisecond)
	InitialPreVoteTimeout := time.Duration(3000 * time.Millisecond)
	InitialPreCommitTimeout := time.Duration(3000 * time.Millisecond)

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
	if minTxFeeReturned.Cmp(new(big.Int).SetInt64(0)) != 0 {
		t.Fatal("Incorrect minTxFee")
	}

	vsFeeReturned := s.GetValueStoreFee()
	if vsFeeReturned.Cmp(new(big.Int).SetInt64(0)) != 0 {
		t.Fatal("Incorrect valueStoreFee")
	}

	dsEpochFeeReturned := s.GetDataStoreFee()
	if dsEpochFeeReturned.Cmp(new(big.Int).SetInt64(0)) != 0 {
		t.Fatal("Incorrect dataStoreEpochFee")
	}

}

// Test success of LoadStorage
func TestStorageLoadStorageGood1(t *testing.T) {
	s := InitializeStorageWithStandardNode()
	epoch := uint32(25519)

	dvTrue := GetStandardDynamicValue()
	rsTrueBytes, err := dvTrue.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	err = s.UpdateCurrentDynamicValue(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}
	dvBytes, err := s.DynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dvBytes, rsTrueBytes) {
		t.Fatal("dynamicValues values do not match")
	}
}

// Test success of LoadStorage again
func TestStorageLoadStorageGood2(t *testing.T) {
	s := InitializeStorageWithFirstNode()
	epoch := uint32(25519)

	rsTrue := GetStandardDynamicValue()
	rsTrueBytes, err := rsTrue.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	err = s.LoadStorage(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}
	rsBytes, err := s.DynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rsBytes, rsTrueBytes) {
		t.Fatal("dynamicValues values do not match")
	}
}

// Test failure of LoadStorage
func TestStorageLoadStorageBad1(t *testing.T) {
	s := InitializeStorage()
	// We attempt to load the zero epoch;
	// this should raise an error.
	err := s.LoadStorage(nil, 0)
	if !errors.Is(err, ErrZeroEpoch) {
		t.Fatal("Should have raised ErrZeroEpoch")
	}
}

// Test success of loadDynamicValues.
func TestStorageLoadDynamicValuesGood1(t *testing.T) {
	s := InitializeStorageWithFirstNode()
	rsTrue := GetStandardDynamicValue()
	rsTrueBytes, err := rsTrue.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	// We attempt to load an epoch from an empty database;
	// this should raise an error.
	rs, err := s.loadDynamicValues(nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	rsBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rsBytes, rsTrueBytes) {
		t.Fatal("DynamicValues values do not match")
	}
}

// Test success of loadDynamicValues again
func TestStorageLoadDynamicValuesGood2(t *testing.T) {
	s := InitializeStorageWithFirstNode()
	rsTrue := GetStandardDynamicValue()
	rsTrueBytes, err := rsTrue.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	field := "maxBytes"
	value := "123456789"
	epoch := uint32(257)
	update, err := NewUpdate(field, value, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = s.UpdateStorage(nil, update)
	if err != nil {
		t.Fatal(err)
	}

	// We attempt to load an epoch from an empty database;
	// this should raise an error.
	rs, err := s.loadDynamicValues(nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	rsBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rsBytes, rsTrueBytes) {
		t.Fatal("DynamicValues values do not match")
	}
}

// Test failure of loadDynamicValues.
// We raise an error for attempting to load epoch 0
func TestStorageLoadDynamicValuesBad1(t *testing.T) {
	s := InitializeStorageWithFirstNode()

	// We attempt to load an epoch from an empty database;
	// this should raise an error.
	_, err := s.loadDynamicValues(nil, 0)
	if !errors.Is(err, ErrZeroEpoch) {
		t.Fatal("Should have raised ErrZeroEpoch")
	}
}

// Test failure of loadDynamicValues.
// We raise an error for not having LinkedList present.
func TestStorageLoadDynamicValuesBad2(t *testing.T) {
	storageLogger := newLogger()
	database := initializeDB()
	s := &Storage{}
	s.startChan = make(chan struct{})
	s.database = database
	s.logger = storageLogger

	// We attempt to load an epoch from an empty database;
	// this should raise an error.
	_, err := s.loadDynamicValues(nil, 1)
	if !errors.Is(err, ErrKeyNotPresent) {
		t.Fatal("Should have raised ErrKeyNotPresent")
	}
}

// Test failure of loadDynamicValues again.
// It should not be possible to reach this configuration.
func TestStorageLoadStorageBad3(t *testing.T) {
	storageLogger := newLogger()
	database := initializeDB()
	ll := &LinkedList{
		currentValue: 1,
	}
	err := database.SetLinkedList(nil, ll)
	if err != nil {
		t.Fatal(err)
	}

	s := &Storage{}
	s.startChan = make(chan struct{})
	s.database = database
	s.logger = storageLogger

	// We attempt to load an epoch from an empty database (without nodes but LinkedList set);
	// this should raise an error.
	_, err = s.loadDynamicValues(nil, 1)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeHeadGood(t *testing.T) {
	// Initialize storage and have standard node at epoch 1
	s := InitializeStorageWithFirstNode()
	epoch := uint32(1)
	headNode, err := s.database.GetNode(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}

	newEpoch := epoch + 1
	rsNew := &DynamicValues{}
	rsNew.standardParameters()
	rsBytes, err := rsNew.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	newMaxBytes := uint32(1234567)
	rsNew.SetMaxBytes(newMaxBytes)
	node := &Node{
		thisEpoch:     newEpoch,
		dynamicValues: rsNew,
	}
	if !node.IsPreValid() {
		t.Fatal("node should be prevalid")
	}
	rsNewBytes, err := rsNew.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	err = s.addNodeHead(nil, node, headNode)
	if err != nil {
		t.Fatal(err)
	}

	origNode, err := s.database.GetNode(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}
	if origNode.prevEpoch != epoch || origNode.thisEpoch != epoch || origNode.nextEpoch != newEpoch {
		t.Fatal("origNode invalid (1)")
	}
	rsOrigBytes, err := origNode.dynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rsOrigBytes, rsBytes) {
		t.Fatal("origNode invalid (2)")
	}

	retNode, err := s.database.GetNode(nil, newEpoch)
	if err != nil {
		t.Fatal(err)
	}
	if retNode.prevEpoch != epoch || retNode.thisEpoch != newEpoch || retNode.nextEpoch != newEpoch {
		t.Fatal("retNode invalid (1)")
	}
	retBytes, err := retNode.dynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retBytes, rsNewBytes) {
		t.Fatal("retNode invalid (2)")
	}
}

func TestStorageAddNodeHeadBad1(t *testing.T) {
	origEpoch := uint32(1)
	s := InitializeStorageWithFirstNode()
	headNode, err := s.database.GetNode(nil, origEpoch)
	if err != nil {
		t.Fatal(err)
	}
	if !headNode.IsValid() {
		t.Fatal("headNode should be valid")
	}

	node := &Node{}
	if node.IsPreValid() {
		t.Fatal("node should not be prevalid")
	}

	err = s.addNodeHead(nil, node, headNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeHeadBad2(t *testing.T) {
	origEpoch := uint32(1)
	s := InitializeStorageWithFirstNode()
	headNode, err := s.database.GetNode(nil, origEpoch)
	if err != nil {
		t.Fatal(err)
	}
	if !headNode.IsValid() {
		t.Fatal("headNode should be valid")
	}

	rs := &DynamicValues{}
	node := &Node{
		thisEpoch:     1,
		dynamicValues: rs,
	}
	if !node.IsPreValid() {
		t.Fatal("node should be prevalid")
	}

	err = s.addNodeHead(nil, node, headNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeSplitGood(t *testing.T) {
	first := uint32(1)
	s := InitializeStorageWithFirstNode()
	prevNode, err := s.database.GetNode(nil, first)
	if err != nil {
		t.Fatal(err)
	}
	if !prevNode.IsValid() {
		t.Fatal("prevNode should be valid")
	}
	last := uint32(10)
	prevNode.nextEpoch = last
	err = s.database.SetNode(nil, prevNode)
	if err != nil {
		t.Fatal(err)
	}

	// Set up nextnode
	rs := &DynamicValues{}
	rs.standardParameters()
	rsBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	nextNode := &Node{
		prevEpoch:     first,
		thisEpoch:     last,
		nextEpoch:     last,
		dynamicValues: rs,
	}
	if !nextNode.IsValid() {
		t.Fatal("nextNode should be valid")
	}
	err = s.database.SetNode(nil, nextNode)
	if err != nil {
		t.Fatal(err)
	}

	// Set up node
	rsNew, err := rs.Copy()
	if err != nil {
		t.Fatal(err)
	}
	rsNew.MaxBytes = 123456
	rsNewBytes, err := rsNew.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	newEpoch := uint32(5)
	node := &Node{
		thisEpoch:     newEpoch,
		dynamicValues: rsNew,
	}
	if !node.IsPreValid() {
		t.Fatal("node should be preValid")
	}

	err = s.addNodeSplit(nil, node, prevNode, nextNode)
	if err != nil {
		t.Fatal(err)
	}

	// Check everything
	firstNode, err := s.database.GetNode(nil, first)
	if err != nil {
		t.Fatal(err)
	}
	if firstNode.prevEpoch != first || firstNode.thisEpoch != first || firstNode.nextEpoch != newEpoch {
		t.Fatal("firstNode invalid (1)")
	}
	retBytes, err := firstNode.dynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retBytes, rsBytes) {
		t.Fatal("firstNode invalid (2)")
	}

	middleNode, err := s.database.GetNode(nil, newEpoch)
	if err != nil {
		t.Fatal(err)
	}
	if middleNode.prevEpoch != first || middleNode.thisEpoch != newEpoch || middleNode.nextEpoch != last {
		t.Fatal("middleNode invalid (1)")
	}
	retBytes, err = middleNode.dynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retBytes, rsNewBytes) {
		t.Fatal("middleNode invalid (2)")
	}

	lastNode, err := s.database.GetNode(nil, last)
	if err != nil {
		t.Fatal(err)
	}
	if lastNode.prevEpoch != newEpoch || lastNode.thisEpoch != last || lastNode.nextEpoch != last {
		t.Fatal("lastNode invalid (1)")
	}
	retBytes, err = lastNode.dynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retBytes, rsBytes) {
		t.Fatal("lastNode invalid (2)")
	}
}

func TestStorageAddNodeSplitBad1(t *testing.T) {
	first := uint32(1)
	s := InitializeStorageWithFirstNode()
	prevNode, err := s.database.GetNode(nil, first)
	if err != nil {
		t.Fatal(err)
	}
	if !prevNode.IsValid() {
		t.Fatal("prevNode should be valid")
	}
	last := uint32(10)
	prevNode.nextEpoch = last
	err = s.database.SetNode(nil, prevNode)
	if err != nil {
		t.Fatal(err)
	}

	// Set up nodes
	rs := &DynamicValues{}
	rs.standardParameters()
	nextNode := &Node{
		prevEpoch:     first,
		thisEpoch:     last,
		nextEpoch:     last,
		dynamicValues: rs,
	}
	if !nextNode.IsValid() {
		t.Fatal("nextNode should be valid")
	}
	err = s.database.SetNode(nil, nextNode)
	if err != nil {
		t.Fatal(err)
	}

	node := &Node{}

	err = s.addNodeSplit(nil, node, prevNode, nextNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeSplitBad2(t *testing.T) {
	first := uint32(1)
	s := InitializeStorageWithFirstNode()
	prevNode, err := s.database.GetNode(nil, first)
	if err != nil {
		t.Fatal(err)
	}
	if !prevNode.IsValid() {
		t.Fatal("prevNode should be valid")
	}
	last := uint32(10)
	prevNode.nextEpoch = last
	err = s.database.SetNode(nil, prevNode)
	if err != nil {
		t.Fatal(err)
	}

	// Set up nodes
	rs := &DynamicValues{}
	rs.standardParameters()
	nextNode := &Node{
		prevEpoch:     first,
		thisEpoch:     last,
		nextEpoch:     last,
		dynamicValues: rs,
	}
	if !nextNode.IsValid() {
		t.Fatal("nextNode should be valid")
	}
	err = s.database.SetNode(nil, nextNode)
	if err != nil {
		t.Fatal(err)
	}

	rsNew, err := rs.Copy()
	if err != nil {
		t.Fatal(err)
	}
	rsNew.MaxBytes = 123456
	newEpoch := uint32(100)
	node := &Node{
		thisEpoch:     newEpoch,
		dynamicValues: rsNew,
	}
	if !node.IsPreValid() {
		t.Fatal("node should be preValid")
	}

	err = s.addNodeSplit(nil, node, prevNode, nextNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// Test addNode when adding to Head
func TestStorageAddNodeGood1(t *testing.T) {
	origEpoch := uint32(1)
	s := InitializeStorageWithFirstNode()
	rs := &DynamicValues{}
	rs.standardParameters()
	rsStandardBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	newMaxBytes := uint32(12345)
	rs.MaxBytes = newMaxBytes
	epoch := uint32(10)
	newNode := &Node{
		prevEpoch:     0,
		thisEpoch:     epoch,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	err = s.addNode(nil, newNode)
	if err != nil {
		t.Fatal(err)
	}
	rsNewBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	// Check everything
	origNode, err := s.database.GetNode(nil, origEpoch)
	if err != nil {
		t.Fatal(err)
	}
	if origNode.prevEpoch != origEpoch {
		t.Fatal("origNode.prevEpoch is invalid")
	}
	if origNode.thisEpoch != origEpoch {
		t.Fatal("origNode.thisEpoch is invalid")
	}
	if origNode.nextEpoch != epoch {
		t.Fatal("origNode.nextEpoch is invalid")
	}
	retRS, err := origNode.dynamicValues.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err := retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retRSBytes, rsStandardBytes) {
		t.Fatal("invalid DynamicValues")
	}

	addedNode, err := s.database.GetNode(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}
	if addedNode.prevEpoch != origEpoch {
		t.Fatal("addedNode.prevEpoch is invalid")
	}
	if addedNode.thisEpoch != epoch {
		t.Fatal("addedNode.thisEpoch is invalid")
	}
	if addedNode.nextEpoch != epoch {
		t.Fatal("addedNode.nextEpoch is invalid")
	}
	retRS, err = addedNode.dynamicValues.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err = retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retRSBytes, rsNewBytes) {
		t.Fatal("invalid DynamicValues (2)")
	}
}

// Test addNode when adding to Head and then in between
func TestStorageAddNodeGood2(t *testing.T) {
	origEpoch := uint32(1)
	s := InitializeStorageWithFirstNode()
	rs := &DynamicValues{}
	rs.standardParameters()
	rsStandardBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	newEpoch := uint32(100)
	newNode := &Node{
		prevEpoch:     0,
		thisEpoch:     newEpoch,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	err = s.addNode(nil, newNode)
	if err != nil {
		t.Fatal(err)
	}

	newEpoch2 := uint32(10)
	newMaxBytes := uint32(12345689)
	rs.SetMaxBytes(newMaxBytes)
	newNode2 := &Node{
		prevEpoch:     0,
		thisEpoch:     newEpoch2,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	err = s.addNode(nil, newNode2)
	if err != nil {
		t.Fatal(err)
	}
	rsNewBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	// Check everything
	origNode, err := s.database.GetNode(nil, origEpoch)
	if err != nil {
		t.Fatal(err)
	}
	if origNode.prevEpoch != origEpoch {
		t.Fatal("origNode.prevEpoch is invalid")
	}
	if origNode.thisEpoch != origEpoch {
		t.Fatal("origNode.thisEpoch is invalid")
	}
	if origNode.nextEpoch != newEpoch2 {
		t.Fatal("origNode.nextEpoch is invalid")
	}
	retRS, err := origNode.dynamicValues.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err := retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retRSBytes, rsStandardBytes) {
		t.Fatal("invalid DynamicValues")
	}

	addedNode, err := s.database.GetNode(nil, newEpoch)
	if err != nil {
		t.Fatal(err)
	}
	if addedNode.prevEpoch != newEpoch2 {
		t.Fatal("addedNode.prevEpoch is invalid")
	}
	if addedNode.thisEpoch != newEpoch {
		t.Fatal("addedNode.thisEpoch is invalid")
	}
	if addedNode.nextEpoch != newEpoch {
		t.Fatal("addedNode.nextEpoch is invalid")
	}
	retRS, err = addedNode.dynamicValues.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err = retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retRSBytes, rsStandardBytes) {
		t.Fatal("invalid DynamicValues (2)")
	}

	addedNode2, err := s.database.GetNode(nil, newEpoch2)
	if err != nil {
		t.Fatal(err)
	}
	if addedNode2.prevEpoch != origEpoch {
		t.Fatal("addedNode2.prevEpoch is invalid")
	}
	if addedNode2.thisEpoch != newEpoch2 {
		t.Fatal("addedNode2.thisEpoch is invalid")
	}
	if addedNode2.nextEpoch != newEpoch {
		t.Fatal("addedNode2.nextEpoch is invalid")
	}
	retRS, err = addedNode2.dynamicValues.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err = retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retRSBytes, rsNewBytes) {
		t.Fatal("invalid DynamicValues (3)")
	}
}

func TestStorageAddNodeBad1(t *testing.T) {
	s := InitializeStorage()
	rs := &DynamicValues{}
	newNode := &Node{
		prevEpoch:     0,
		thisEpoch:     0,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	err := s.addNode(nil, newNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeBad2(t *testing.T) {
	s := InitializeStorage()
	rs := &DynamicValues{}
	newNode := &Node{
		prevEpoch:     0,
		thisEpoch:     1,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	err := s.addNode(nil, newNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeBad3(t *testing.T) {
	s := InitializeStorageWithFirstNode()
	rs := &DynamicValues{}
	newNode := &Node{
		prevEpoch:     0,
		thisEpoch:     1,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	err := s.addNode(nil, newNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeBad4(t *testing.T) {
	s := InitializeStorageWithFirstNode()
	rs := &DynamicValues{}
	newNode := &Node{
		prevEpoch:     0,
		thisEpoch:     257,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	err := s.addNode(nil, newNode)
	if err != nil {
		t.Fatal(err)
	}

	newNode2 := &Node{
		prevEpoch:     0,
		thisEpoch:     1,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	err = s.addNode(nil, newNode2)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeBad5(t *testing.T) {
	s := InitializeStorageWithFirstNode()
	rs := &DynamicValues{}
	newNode := &Node{
		prevEpoch:     0,
		thisEpoch:     257,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	err := s.addNode(nil, newNode)
	if err != nil {
		t.Fatal(err)
	}

	newNode2 := &Node{
		prevEpoch:     0,
		thisEpoch:     25519,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	err = s.addNode(nil, newNode2)
	if err != nil {
		t.Fatal(err)
	}

	newNode3 := &Node{
		prevEpoch:     0,
		thisEpoch:     1,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	err = s.addNode(nil, newNode3)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeBad6(t *testing.T) {
	s := InitializeStorage()
	ll := &LinkedList{25519}
	err := s.database.SetLinkedList(nil, ll)
	if err != nil {
		t.Fatal(err)
	}

	rs := &DynamicValues{}
	newNode := &Node{
		prevEpoch:     0,
		thisEpoch:     1,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	err = s.addNode(nil, newNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// Test failure of UpdateStorage
func TestStorageUpdateStorageBad(t *testing.T) {
	s := InitializeStorage()
	epoch := uint32(25519)
	field := "invalid"
	value := ""
	update := &Update{
		name:  field,
		value: value,
		epoch: epoch,
	}
	err := s.UpdateStorage(nil, update)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// Test success of UpdateStorage
func TestStorageUpdateStorageGood1(t *testing.T) {
	s := InitializeStorage()
	epoch := uint32(1)
	field := "maxBytes"
	value := "123456789"
	update, err := NewUpdate(field, value, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = s.UpdateStorage(nil, update)
	if err != nil {
		t.Fatal(err)
	}

	rs := &DynamicValues{}
	rs.standardParameters()
	i, err := strconv.Atoi(value)
	if err != nil {
		t.Fatal(err)
	}
	rs.MaxBytes = uint32(i)
	rs.MaxProposalSize = uint32(i)
	rsBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	err = s.LoadStorage(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}
	retRS, err := s.DynamicValues.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err := retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(rsBytes, retRSBytes) {
		t.Log(string(rsBytes))
		t.Log(string(retRSBytes))
		t.Fatal("invalid DynamicValues")
	}
}

// Test success of UpdateStorage
func TestStorageUpdateStorageGood2(t *testing.T) {
	s := InitializeStorage()
	epoch := uint32(25519)
	field := "maxBytes"
	value := "123456789"
	update, err := NewUpdate(field, value, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = s.UpdateStorage(nil, update)
	if err != nil {
		t.Fatal(err)
	}

	rs := &DynamicValues{}
	rs.standardParameters()
	i, err := strconv.Atoi(value)
	if err != nil {
		t.Fatal(err)
	}
	rs.SetMaxBytes(uint32(i))
	rsBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	err = s.LoadStorage(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}
	retRS, err := s.DynamicValues.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err := retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(rsBytes, retRSBytes) {
		t.Log(string(rsBytes))
		t.Log(string(retRSBytes))
		t.Fatal("invalid DynamicValues")
	}
}

// Test success of UpdateStorage
func TestStorageUpdateStorageGood3(t *testing.T) {
	s := InitializeStorageWithFirstNode()

	// Add another epoch in the future
	epoch2 := uint32(25519)
	rs2 := &DynamicValues{}
	rs2.standardParameters()
	newPropTO := 13 * time.Second
	rs2.SetProposalStepTimeout(newPropTO)
	node2 := &Node{
		thisEpoch:     epoch2,
		dynamicValues: rs2,
	}
	err := s.addNode(nil, node2)
	if err != nil {
		t.Fatal(err)
	}

	// Update initial storage value
	epoch := uint32(1)
	field := "maxBytes"
	value := "123456789"
	update, err := NewUpdate(field, value, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = s.UpdateStorage(nil, update)
	if err != nil {
		t.Fatal(err)
	}
	rsTrue := GetStandardDynamicValue()
	newMaxBytesInt, err := strconv.Atoi(value)
	if err != nil {
		t.Fatal(err)
	}
	rsTrue.SetMaxBytes(uint32(newMaxBytesInt))
	rsTrueBytes, err := rsTrue.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	// Test Epoch1
	err = s.LoadStorage(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}
	retRS, err := s.DynamicValues.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err := retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rsTrueBytes, retRSBytes) {
		t.Log(string(rsTrueBytes))
		t.Log(string(retRSBytes))
		t.Fatal("invalid DynamicValues (1)")
	}

	rsTrue2 := &DynamicValues{}
	rsTrue2.standardParameters()
	rsTrue2.SetMaxBytes(uint32(newMaxBytesInt))
	rsTrue2.SetProposalStepTimeout(newPropTO)
	rsTrue2Bytes, err := rsTrue2.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	// Test Epoch2
	err = s.LoadStorage(nil, epoch2)
	if err != nil {
		t.Fatal(err)
	}
	retRS, err = s.DynamicValues.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err = retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rsTrue2Bytes, retRSBytes) {
		t.Log(string(rsTrue2Bytes))
		t.Log(string(retRSBytes))
		t.Fatal("invalid DynamicValues (2)")
	}
}

// Test failure of UpdateStorageValue
// Attempt to perform invalid update at future epoch
func TestStorageUpdateStorageValueBad1(t *testing.T) {
	s := InitializeStorage()
	epoch := uint32(25519)
	field := "invalid"
	value := ""
	update := &Update{
		name:  field,
		value: value,
		epoch: epoch,
	}
	err := s.updateStorageValue(nil, update)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// Test failure of UpdateStorageValue
// Attempt to perform invalid update at current epoch
func TestStorageUpdateStorageValueBad2(t *testing.T) {
	s := InitializeStorage()
	epoch := uint32(1)
	field := "invalid"
	value := ""
	update := &Update{
		name:  field,
		value: value,
		epoch: epoch,
	}
	err := s.updateStorageValue(nil, update)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// Test failure of UpdateStorageValue
// Attempt to perform invalid update at previous epoch
func TestStorageUpdateStorageValueBad3(t *testing.T) {
	s := InitializeStorage()
	epoch := uint32(1)
	field := "invalid"
	value := ""
	update := &Update{
		name:  field,
		value: value,
		epoch: epoch,
	}
	err := s.updateStorageValue(nil, update)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// Test failure of UpdateStorageValue
// Attempt to perform invalid update in between two valid storage values
func TestStorageUpdateStorageValueBad4(t *testing.T) {
	s := InitializeStorageWithFirstNode()
	rs := &DynamicValues{}
	rs.standardParameters()
	node := &Node{
		thisEpoch:     25519,
		dynamicValues: rs,
	}

	err := s.addNode(nil, node)
	if err != nil {
		t.Fatal(err)
	}

	epoch := uint32(257)
	field := "invalid"
	value := ""
	update := &Update{
		name:  field,
		value: value,
		epoch: epoch,
	}
	err = s.updateStorageValue(nil, update)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageGetMinTxFee(t *testing.T) {
	s := InitializeStorageWithFirstNode()
	txFee := s.GetMinTxFee()
	if txFee.Cmp(minTxFee) != 0 {
		t.Fatal("txFee incorrect")
	}

	epoch := uint32(25519)
	field := "minTxFee"
	value := "123456789"
	update, err := NewUpdate(field, value, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = s.UpdateStorage(nil, update)
	if err != nil {
		t.Fatal(err)
	}
	err = s.LoadStorage(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue, err := stringToBigInt(value)
	if err != nil {
		t.Fatal(err)
	}

	txFee = s.GetMinTxFee()
	if err != nil {
		t.Fatal(err)
	}
	if txFee.Cmp(valueTrue) != 0 {
		t.Fatal("incorrect txFee")
	}
}

func TestStorageGetDataStoreEpochFee(t *testing.T) {
	s := InitializeStorageWithFirstNode()
	dsEpochFee := s.GetDataStoreEpochFee()
	if dsEpochFee.Cmp(dataStoreEpochFee) != 0 {
		t.Fatal("dsEpochFee incorrect")
	}

	epoch := uint32(25519)
	field := "dataStoreEpochFee"
	value := "123456789"
	update, err := NewUpdate(field, value, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = s.UpdateStorage(nil, update)
	if err != nil {
		t.Fatal(err)
	}
	err = s.LoadStorage(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue, err := stringToBigInt(value)
	if err != nil {
		t.Fatal(err)
	}

	dsEpochFee = s.GetDataStoreEpochFee()
	if dsEpochFee.Cmp(valueTrue) != 0 {
		t.Fatal("incorrect dataStoreEpochFee")
	}
}

func TestStorageGetValueStoreFee(t *testing.T) {
	s := InitializeStorageWithFirstNode()
	vsFee := s.GetValueStoreFee()
	if vsFee.Cmp(valueStoreFee) != 0 {
		t.Fatal("vsFee incorrect")
	}

	epoch := uint32(25519)
	field := "valueStoreFee"
	value := "123456789"
	update, err := NewUpdate(field, value, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = s.UpdateStorage(nil, update)
	if err != nil {
		t.Fatal(err)
	}
	err = s.LoadStorage(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue, err := stringToBigInt(value)
	if err != nil {
		t.Fatal(err)
	}

	vsFee = s.GetValueStoreFee()
	if vsFee.Cmp(valueTrue) != 0 {
		t.Fatal("incorrect valueStoreFee")
	}
}
