package dynamics

import (
	"bytes"
	"errors"
	"strconv"
	"testing"
	"time"
)

func initializeStorage() *Storage {
	storageLogger := newLogger()
	mock := &mockRawDB{}
	mock.rawDB = make(map[string]string)

	s := &Storage{}
	err := s.Init(mock, storageLogger)
	if err != nil {
		panic(err)
	}
	s.Start()
	return s
}

func initializeStorageWithFirstNode() *Storage {
	s := initializeStorage()
	field := "maxBytes"
	value := "3000000"
	epoch := uint32(1)
	update, err := NewUpdate(field, value, epoch)
	if err != nil {
		panic(err)
	}
	err = s.UpdateStorage(nil, update)
	if err != nil {
		panic(err)
	}
	return s
}

// Test Storage Init with nothing initialized
func TestStorageInit1(t *testing.T) {
	storageLogger := newLogger()
	mock := &mockRawDB{}
	mock.rawDB = make(map[string]string)

	s := &Storage{}
	err := s.Init(mock, storageLogger)
	if err != nil {
		t.Fatal(err)
	}
	s.Start()

	rs := &RawStorage{}
	rs.standardParameters()
	rsBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	// Check rawStorage == standardParameters
	storageRSBytes, err := s.rawStorage.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rsBytes, storageRSBytes) {
		t.Fatal("rawStorage values do not match")
	}
}

func TestStorageStartGood(t *testing.T) {
	storageLogger := newLogger()
	mock := &mockRawDB{}
	mock.rawDB = make(map[string]string)

	s := &Storage{}
	err := s.Init(mock, storageLogger)
	if err != nil {
		t.Fatal(err)
	}
	s.Start()
}

// Test ensures we panic when running Start before Init.
// This happens from attempting to close a closed channel.
func TestStorageStartFail(t *testing.T) {
	s := &Storage{}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Should panic")
		}
	}()
	s.Start()
}

// Test ensures storage has is initialized to the correct values.
func TestStorageInitialized(t *testing.T) {
	s := initializeStorage()

	maxBytesReturned := s.GetMaxBytes()
	if maxBytesReturned != maxBytes {
		t.Fatal("Incorrect MaxBytes")
	}

	maxProposalSizeReturned := s.GetMaxProposalSize()
	if maxProposalSizeReturned != maxProposalSize {
		t.Fatal("Incorrect MaxProposalSize")
	}

	srvrMsgTimeoutReturned := s.GetSrvrMsgTimeout()
	if srvrMsgTimeoutReturned != srvrMsgTimeout {
		t.Fatal("Incorrect srvrMsgTimeout")
	}

	msgTimeoutReturned := s.GetMsgTimeout()
	if msgTimeoutReturned != msgTimeout {
		t.Fatal("Incorrect msgTimeout")
	}

	proposalStepTimeoutReturned := s.GetProposalStepTimeout()
	if proposalStepTimeoutReturned != proposalStepTO {
		t.Fatal("Incorrect proposalStepTO")
	}

	preVoteStepTimeoutReturned := s.GetPreVoteStepTimeout()
	if preVoteStepTimeoutReturned != preVoteStepTO {
		t.Fatal("Incorrect preVoteStepTO")
	}

	preCommitStepTimeoutReturned := s.GetPreCommitStepTimeout()
	if preCommitStepTimeoutReturned != preCommitStepTO {
		t.Fatal("Incorrect preCommitStepTO")
	}

	deadBlockRoundNextRoundTimeoutReturned := s.GetDeadBlockRoundNextRoundTimeout()
	if deadBlockRoundNextRoundTimeoutReturned != dBRNRTO {
		t.Fatal("Incorrect deadBlockRoundNextRoundTimeout")
	}

	downloadTimeoutReturned := s.GetDownloadTimeout()
	if downloadTimeoutReturned != downloadTO {
		t.Fatal("Incorrect downloadTimeout")
	}

	minTxFeeCostRatioReturned := s.GetMinTxFeeCostRatio()
	if minTxFeeCostRatioReturned.Cmp(minTxFeeCostRatio) != 0 {
		t.Fatal("Incorrect minTxFeeCostRatio")
	}

	txValidVersion := s.GetTxValidVersion()
	if txValidVersion != 0 {
		t.Fatal("Incorrect txValidVersion")
	}

	vsFeeReturned := s.GetValueStoreFee()
	if vsFeeReturned.Cmp(valueStoreFee) != 0 {
		t.Fatal("Incorrect valueStoreFee")
	}

	vsValidVersion := s.GetValueStoreValidVersion()
	if vsValidVersion != 0 {
		t.Fatal("Incorrect valueStoreValidVersion")
	}

	asFeeReturned := s.GetAtomicSwapFee()
	if asFeeReturned.Cmp(atomicSwapFee) != 0 {
		t.Fatal("Incorrect atomicSwapFee")
	}

	asStopEpoch := s.GetAtomicSwapValidStopEpoch()
	if asStopEpoch != 0 {
		t.Fatal("Incorrect atomicSwapValidStopEpoch")
	}

	dsEpochFeeReturned := s.GetDataStoreEpochFee()
	if dsEpochFeeReturned.Cmp(dataStoreEpochFee) != 0 {
		t.Fatal("Incorrect dataStoreEpochFee")
	}

	dsValidVersion := s.GetDataStoreValidVersion()
	if dsValidVersion != 0 {
		t.Fatal("Incorrect dataStoreValidVersion")
	}
}

func TestStorageCheckUpdate(t *testing.T) {
	fieldBad := "invalid"
	valueBad := "invalid"
	epochGood := uint32(25519)
	_, err := NewUpdate(fieldBad, valueBad, epochGood)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	fieldGood := "maxBytes"
	valueGood := "1234567890"
	update, err := NewUpdate(fieldGood, valueGood, epochGood)
	if err != nil {
		t.Fatal(err)
	}
	err = checkUpdate(update)
	if err != nil {
		t.Fatal(err)
	}

	epochBad := uint32(0)
	update, err = NewUpdate(fieldGood, valueGood, epochBad)
	if err != nil {
		t.Fatal(err)
	}
	err = checkUpdate(update)
	if !errors.Is(err, ErrInvalidUpdateValue) {
		t.Fatal("Should have raised error (2)")
	}

	update = &Update{epoch: 1}
	err = checkUpdate(update)
	if !errors.Is(err, ErrInvalidUpdateValue) {
		t.Fatal("Should have raised error (3)")
	}
}

// Test success of LoadStorage
func TestStorageLoadStorageGood1(t *testing.T) {
	s := initializeStorage()
	epoch := uint32(25519)

	rsTrue := &RawStorage{}
	rsTrue.standardParameters()
	rsTrueBytes, err := rsTrue.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	err = s.LoadStorage(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}
	rsBytes, err := s.rawStorage.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rsBytes, rsTrueBytes) {
		t.Fatal("rawStorage values do not match")
	}
}

// Test success of LoadStorage again
func TestStorageLoadStorageGood2(t *testing.T) {
	s := initializeStorageWithFirstNode()
	epoch := uint32(25519)

	rsTrue := &RawStorage{}
	rsTrue.standardParameters()
	rsTrueBytes, err := rsTrue.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	err = s.LoadStorage(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}
	rsBytes, err := s.rawStorage.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rsBytes, rsTrueBytes) {
		t.Fatal("rawStorage values do not match")
	}
}

// Test failure of LoadStorage
func TestStorageLoadStorageBad1(t *testing.T) {
	s := initializeStorage()
	// We attempt to load the zero epoch;
	// this should raise an error.
	err := s.LoadStorage(nil, 0)
	if !errors.Is(err, ErrZeroEpoch) {
		t.Fatal("Should have raised ErrZeroEpoch")
	}
}

// Test success of loadRawStorage.
func TestStorageLoadRawStorageGood1(t *testing.T) {
	s := initializeStorageWithFirstNode()
	rsTrue := &RawStorage{}
	rsTrue.standardParameters()
	rsTrueBytes, err := rsTrue.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	// We attempt to load an epoch from an empty database;
	// this should raise an error.
	rs, err := s.loadRawStorage(nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	rsBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rsBytes, rsTrueBytes) {
		t.Fatal("RawStorage values do not match")
	}
}

// Test success of loadRawStorage again
func TestStorageLoadRawStorageGood2(t *testing.T) {
	s := initializeStorageWithFirstNode()
	rsTrue := &RawStorage{}
	rsTrue.standardParameters()
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
	rs, err := s.loadRawStorage(nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	rsBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rsBytes, rsTrueBytes) {
		t.Fatal("RawStorage values do not match")
	}
}

// Test failure of loadRawStorage.
// We raise an error for attempting to load epoch 0
func TestStorageLoadRawStorageBad1(t *testing.T) {
	s := initializeStorageWithFirstNode()

	// We attempt to load an epoch from an empty database;
	// this should raise an error.
	_, err := s.loadRawStorage(nil, 0)
	if !errors.Is(err, ErrZeroEpoch) {
		t.Fatal("Should have raised ErrZeroEpoch")
	}
}

// Test failure of loadRawStorage.
// We raise an error for not having LinkedList present.
func TestStorageLoadRawStorageBad2(t *testing.T) {
	storageLogger := newLogger()
	database := initializeDB()
	s := &Storage{}
	s.startChan = make(chan struct{})
	s.database = database
	s.logger = storageLogger
	s.Start()

	// We attempt to load an epoch from an empty database;
	// this should raise an error.
	_, err := s.loadRawStorage(nil, 1)
	if !errors.Is(err, ErrKeyNotPresent) {
		t.Fatal("Should have raised ErrKeyNotPresent")
	}
}

// Test failure of loadRawStorage again.
// It should not be possible to reach this configuration.
func TestStorageLoadStorageBad3(t *testing.T) {
	storageLogger := newLogger()
	database := initializeDB()
	ll := &LinkedList{
		epochLastUpdated: 1,
	}
	err := database.SetLinkedList(nil, ll)
	if err != nil {
		t.Fatal(err)
	}

	s := &Storage{}
	s.startChan = make(chan struct{})
	s.database = database
	s.logger = storageLogger
	s.Start()
	// We attempt to load an epoch from an empty database (without nodes but LinkedList set);
	// this should raise an error.
	_, err = s.loadRawStorage(nil, 1)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeHeadGood(t *testing.T) {
	// Initialize storage and have standard node at epoch 1
	s := initializeStorageWithFirstNode()
	epoch := uint32(1)
	headNode, err := s.database.GetNode(nil, epoch)
	if err != nil {
		t.Fatal(err)
	}

	newEpoch := epoch + 1
	rsNew := &RawStorage{}
	rsNew.standardParameters()
	rsBytes, err := rsNew.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	newMaxBytes := uint32(1234567)
	rsNew.SetMaxBytes(newMaxBytes)
	node := &Node{
		thisEpoch:  newEpoch,
		rawStorage: rsNew,
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
	rsOrigBytes, err := origNode.rawStorage.Marshal()
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
	retBytes, err := retNode.rawStorage.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retBytes, rsNewBytes) {
		t.Fatal("retNode invalid (2)")
	}
}

func TestStorageAddNodeHeadBad1(t *testing.T) {
	origEpoch := uint32(1)
	s := initializeStorageWithFirstNode()
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
	s := initializeStorageWithFirstNode()
	headNode, err := s.database.GetNode(nil, origEpoch)
	if err != nil {
		t.Fatal(err)
	}
	if !headNode.IsValid() {
		t.Fatal("headNode should be valid")
	}

	rs := &RawStorage{}
	node := &Node{
		thisEpoch:  1,
		rawStorage: rs,
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
	s := initializeStorageWithFirstNode()
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
	rs := &RawStorage{}
	rs.standardParameters()
	rsBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	nextNode := &Node{
		prevEpoch:  first,
		thisEpoch:  last,
		nextEpoch:  last,
		rawStorage: rs,
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
		thisEpoch:  newEpoch,
		rawStorage: rsNew,
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
	retBytes, err := firstNode.rawStorage.Marshal()
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
	retBytes, err = middleNode.rawStorage.Marshal()
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
	retBytes, err = lastNode.rawStorage.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retBytes, rsBytes) {
		t.Fatal("lastNode invalid (2)")
	}
}

func TestStorageAddNodeSplitBad1(t *testing.T) {
	first := uint32(1)
	s := initializeStorageWithFirstNode()
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
	rs := &RawStorage{}
	rs.standardParameters()
	nextNode := &Node{
		prevEpoch:  first,
		thisEpoch:  last,
		nextEpoch:  last,
		rawStorage: rs,
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
	s := initializeStorageWithFirstNode()
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
	rs := &RawStorage{}
	rs.standardParameters()
	nextNode := &Node{
		prevEpoch:  first,
		thisEpoch:  last,
		nextEpoch:  last,
		rawStorage: rs,
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
		thisEpoch:  newEpoch,
		rawStorage: rsNew,
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
	s := initializeStorageWithFirstNode()
	rs := &RawStorage{}
	rs.standardParameters()
	rsStandardBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	newMaxBytes := uint32(12345)
	rs.MaxBytes = newMaxBytes
	epoch := uint32(10)
	newNode := &Node{
		prevEpoch:  0,
		thisEpoch:  epoch,
		nextEpoch:  0,
		rawStorage: rs,
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
	retRS, err := origNode.rawStorage.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err := retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retRSBytes, rsStandardBytes) {
		t.Fatal("invalid RawStorage")
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
	retRS, err = addedNode.rawStorage.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err = retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retRSBytes, rsNewBytes) {
		t.Fatal("invalid RawStorage (2)")
	}
}

// Test addNode when adding to Head and then in between
func TestStorageAddNodeGood2(t *testing.T) {
	origEpoch := uint32(1)
	s := initializeStorageWithFirstNode()
	rs := &RawStorage{}
	rs.standardParameters()
	rsStandardBytes, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	newEpoch := uint32(100)
	newNode := &Node{
		prevEpoch:  0,
		thisEpoch:  newEpoch,
		nextEpoch:  0,
		rawStorage: rs,
	}
	err = s.addNode(nil, newNode)
	if err != nil {
		t.Fatal(err)
	}

	newEpoch2 := uint32(10)
	newMaxBytes := uint32(12345689)
	rs.SetMaxBytes(newMaxBytes)
	newNode2 := &Node{
		prevEpoch:  0,
		thisEpoch:  newEpoch2,
		nextEpoch:  0,
		rawStorage: rs,
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
	retRS, err := origNode.rawStorage.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err := retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retRSBytes, rsStandardBytes) {
		t.Fatal("invalid RawStorage")
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
	retRS, err = addedNode.rawStorage.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err = retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retRSBytes, rsStandardBytes) {
		t.Fatal("invalid RawStorage (2)")
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
	retRS, err = addedNode2.rawStorage.Copy()
	if err != nil {
		t.Fatal(err)
	}
	retRSBytes, err = retRS.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retRSBytes, rsNewBytes) {
		t.Fatal("invalid RawStorage (3)")
	}
}

func TestStorageAddNodeBad1(t *testing.T) {
	s := initializeStorage()
	rs := &RawStorage{}
	newNode := &Node{
		prevEpoch:  0,
		thisEpoch:  0,
		nextEpoch:  0,
		rawStorage: rs,
	}
	err := s.addNode(nil, newNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeBad2(t *testing.T) {
	s := initializeStorage()
	rs := &RawStorage{}
	newNode := &Node{
		prevEpoch:  0,
		thisEpoch:  1,
		nextEpoch:  0,
		rawStorage: rs,
	}
	err := s.addNode(nil, newNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeBad3(t *testing.T) {
	s := initializeStorageWithFirstNode()
	rs := &RawStorage{}
	newNode := &Node{
		prevEpoch:  0,
		thisEpoch:  1,
		nextEpoch:  0,
		rawStorage: rs,
	}
	err := s.addNode(nil, newNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeBad4(t *testing.T) {
	s := initializeStorageWithFirstNode()
	rs := &RawStorage{}
	newNode := &Node{
		prevEpoch:  0,
		thisEpoch:  257,
		nextEpoch:  0,
		rawStorage: rs,
	}
	err := s.addNode(nil, newNode)
	if err != nil {
		t.Fatal(err)
	}

	newNode2 := &Node{
		prevEpoch:  0,
		thisEpoch:  1,
		nextEpoch:  0,
		rawStorage: rs,
	}
	err = s.addNode(nil, newNode2)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeBad5(t *testing.T) {
	s := initializeStorageWithFirstNode()
	rs := &RawStorage{}
	newNode := &Node{
		prevEpoch:  0,
		thisEpoch:  257,
		nextEpoch:  0,
		rawStorage: rs,
	}
	err := s.addNode(nil, newNode)
	if err != nil {
		t.Fatal(err)
	}

	newNode2 := &Node{
		prevEpoch:  0,
		thisEpoch:  25519,
		nextEpoch:  0,
		rawStorage: rs,
	}
	err = s.addNode(nil, newNode2)
	if err != nil {
		t.Fatal(err)
	}

	newNode3 := &Node{
		prevEpoch:  0,
		thisEpoch:  1,
		nextEpoch:  0,
		rawStorage: rs,
	}
	err = s.addNode(nil, newNode3)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestStorageAddNodeBad6(t *testing.T) {
	s := initializeStorage()
	ll := &LinkedList{25519}
	err := s.database.SetLinkedList(nil, ll)
	if err != nil {
		t.Fatal(err)
	}

	rs := &RawStorage{}
	newNode := &Node{
		prevEpoch:  0,
		thisEpoch:  1,
		nextEpoch:  0,
		rawStorage: rs,
	}
	err = s.addNode(nil, newNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// Test failure of UpdateStorage
func TestStorageUpdateStorageBad(t *testing.T) {
	s := initializeStorage()
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
	s := initializeStorage()
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

	rs := &RawStorage{}
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
	retRS, err := s.rawStorage.Copy()
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
		t.Fatal("invalid RawStorage")
	}
}

// Test success of UpdateStorage
func TestStorageUpdateStorageGood2(t *testing.T) {
	s := initializeStorage()
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

	rs := &RawStorage{}
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
	retRS, err := s.rawStorage.Copy()
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
		t.Fatal("invalid RawStorage")
	}
}

// Test success of UpdateStorage
func TestStorageUpdateStorageGood3(t *testing.T) {
	s := initializeStorageWithFirstNode()

	// Add another epoch in the future
	epoch2 := uint32(25519)
	rs2 := &RawStorage{}
	rs2.standardParameters()
	newPropTO := 13 * time.Second
	rs2.SetProposalStepTimeout(newPropTO)
	node2 := &Node{
		thisEpoch:  epoch2,
		rawStorage: rs2,
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
	rsTrue := &RawStorage{}
	rsTrue.standardParameters()
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
	retRS, err := s.rawStorage.Copy()
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
		t.Fatal("invalid RawStorage (1)")
	}

	rsTrue2 := &RawStorage{}
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
	retRS, err = s.rawStorage.Copy()
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
		t.Fatal("invalid RawStorage (2)")
	}
}

// Test failure of UpdateStorageValue
// Attempt to perform invalid update at future epoch
func TestStorageUpdateStorageValueBad1(t *testing.T) {
	s := initializeStorage()
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
	s := initializeStorage()
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
	s := initializeStorage()
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
	s := initializeStorageWithFirstNode()
	rs := &RawStorage{}
	rs.standardParameters()
	node := &Node{
		thisEpoch:  25519,
		rawStorage: rs,
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
	s := initializeStorageWithFirstNode()
	txFeeCostRatio := s.GetMinTxFeeCostRatio()
	if txFeeCostRatio.Cmp(minTxFeeCostRatio) != 0 {
		t.Fatal("txFee incorrect")
	}

	epoch := uint32(25519)
	field := "minTxFeeCostRatio"
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

	txFeeCostRatio = s.GetMinTxFeeCostRatio()
	if err != nil {
		t.Fatal(err)
	}
	if txFeeCostRatio.Cmp(valueTrue) != 0 {
		t.Fatal("incorrect txFeeCostRatio")
	}
}

func TestStorageGetDataStoreEpochFee(t *testing.T) {
	s := initializeStorageWithFirstNode()
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
	s := initializeStorageWithFirstNode()
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

func TestStorageGetAtomicSwapFee(t *testing.T) {
	s := initializeStorageWithFirstNode()
	asFee := s.GetAtomicSwapFee()
	if asFee.Cmp(atomicSwapFee) != 0 {
		t.Fatal("asFee incorrect")
	}

	epoch := uint32(25519)
	field := "atomicSwapFee"
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

	asFee = s.GetAtomicSwapFee()
	if asFee.Cmp(valueTrue) != 0 {
		t.Fatal("incorrect atomicSwapFee")
	}
}
