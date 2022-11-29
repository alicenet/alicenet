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

func GetDynamicValueInThePast(s *Storage, epoch uint32) (uint32, *DynamicValues) {
	var value *DynamicValues
	var executionEpoch uint32
	err := s.database.rawDB.Update(func(txn *badger.Txn) error {
		var err error
		executionEpoch, value, err = s.GetDynamicValueInThePast(txn, epoch)
		return err
	})
	if err != nil {
		panic(err)
	}
	return executionEpoch, value
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
	err := s.Init(NewTestDB(), newLogger())
	if err != nil {
		t.Fatal("Database not initialized")
	}

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
	err = s2.Init(s.database.rawDB, s.logger)
	if err != nil {
		t.Fatal("unable to initialize storage")
	}

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

func TestStorageUpdateCurrentDynamicValueWithChange(t *testing.T) {
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
	UpdateCurrentDynamicValue(s, epoch+1)

	dvBytes, err = s.DynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dvBytes, dvTrueBytes) {
		t.Fatal("dynamicValues values do not match (2)")
	}
}

func TestStorageUpdateCurrentDynamicValueWithALotOfUpdates(t *testing.T) {
	t.Parallel()
	s := InitializeStorageWithStandardNode()

	epoch := uint32(255)

	newValueRawWithFee, _ := GetDynamicValueWithFees()
	for i := uint8(0); i < 10; i++ {
		// changing the last byte of the value of max block size
		newValueRawWithFee[15]++
		// adding the new node
		ChangeDynamicValues(s, epoch+uint32(i*10), newValueRawWithFee)
	}
	expectedValue := uint32(3_000_000)
	// check the value before
	assert.Equal(t, s.GetMaxBlockSize(), expectedValue)
	for i := uint8(0); i < 10; i++ {
		// value should not change at the boundary
		UpdateCurrentDynamicValue(s, epoch+uint32(i*10))
		// check the value after
		assert.Equal(t, s.GetMaxBlockSize(), expectedValue)

		// executing again with a epoch after the boundary, the value should change
		UpdateCurrentDynamicValue(s, epoch+uint32(i*10)+1)
		expectedValue++
		// check the value after
		assert.Equal(t, s.GetMaxBlockSize(), expectedValue)

		//executing again with the same previous epoch should not change the value
		UpdateCurrentDynamicValue(s, epoch+uint32(i*10)+1)
		// check the value after
		assert.Equal(t, s.GetMaxBlockSize(), expectedValue)
	}
	// a value way in the future, should be able to update the last valid dynamic
	// value in 1 iteration
	UpdateCurrentDynamicValue(s, epoch+100000)
	// check the value after
	assert.Equal(t, s.GetMaxBlockSize(), uint32(3_000_010))
}

func TestStorageUpdateCurrentDynamicValueWithALotOfUpdatesInSequence(t *testing.T) {
	t.Parallel()
	s := InitializeStorageWithStandardNode()

	epoch := uint32(255)

	newValueRawWithFee, _ := GetDynamicValueWithFees()
	for i := uint8(0); i < 10; i++ {
		// changing the last byte of the value of max block size
		newValueRawWithFee[15]++
		// adding the new node
		ChangeDynamicValues(s, epoch+uint32(i), newValueRawWithFee)
	}
	expectedValue := uint32(3_000_000)
	// check the value before
	assert.Equal(t, s.GetMaxBlockSize(), expectedValue)
	for i := uint8(0); i < 10; i++ {
		// value should not change at the boundary
		UpdateCurrentDynamicValue(s, epoch+uint32(i))
		// check the value after
		assert.Equal(t, s.GetMaxBlockSize(), expectedValue)

		// executing again with a epoch after the boundary, the value should change
		UpdateCurrentDynamicValue(s, epoch+uint32(i)+1)
		expectedValue++
		// check the value after
		assert.Equal(t, s.GetMaxBlockSize(), expectedValue)

		//executing again with the same previous epoch should not change the value
		UpdateCurrentDynamicValue(s, epoch+uint32(i)+1)
		// check the value after
		assert.Equal(t, s.GetMaxBlockSize(), expectedValue)
	}
}

// Test failure of UpdateCurrentDynamicValue
func TestStorageUpdateCurrentDynamicValueBad1(t *testing.T) {
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
func TestStorageUpdateCurrentDynamicValueBad2(t *testing.T) {
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

func TestGetValuesInThePast(t *testing.T) {
	t.Parallel()
	s := InitializeStorageWithStandardNode()

	epoch := uint32(255)

	newValueRawWithFee, _ := GetDynamicValueWithFees()
	expectedValue := uint32(3_000_000)
	for i := uint8(0); i < 10; i++ {
		// changing the last byte of the value of max block size
		newValueRawWithFee[15]++
		// adding the new node
		ChangeDynamicValues(s, epoch+uint32(i*10), newValueRawWithFee)
		assert.Equal(t, s.GetMaxBlockSize(), expectedValue)
		// now update the values
		UpdateCurrentDynamicValue(s, epoch+uint32(i*10)+1)
		expectedValue++
		// check the value after
		assert.Equal(t, s.GetMaxBlockSize(), expectedValue)
	}

	//  get values in the past
	expectedValue = uint32(3_000_000)
	// get the value of first epoch
	executionEpoch, dv := GetDynamicValueInThePast(s, 1)
	assert.Equal(t, dv.GetMaxBlockSize(), expectedValue)
	// should be the height of the head node
	assert.Equal(t, executionEpoch, uint32(1))

	// before the first update
	_, dv = GetDynamicValueInThePast(s, 254)
	assert.Equal(t, dv.GetMaxBlockSize(), expectedValue)
	// should be the height of the head node
	assert.Equal(t, executionEpoch, uint32(1))

	for i := uint8(0); i < 10; i++ {
		// before the update
		_, dv = GetDynamicValueInThePast(s, epoch+uint32(i*10)-5)
		assert.Equal(t, dv.GetMaxBlockSize(), expectedValue)
		expectedValue++
		executionEpoch, dv = GetDynamicValueInThePast(s, epoch+uint32(i*10)+1)
		assert.Equal(t, dv.GetMaxBlockSize(), expectedValue)
		assert.Equal(t, executionEpoch, epoch+uint32(i*10))
	}

	// get value way in the future
	_, dv = GetDynamicValueInThePast(s, 254000)
	assert.Equal(t, dv.GetMaxBlockSize(), expectedValue)
}

// Test failure of UpdateCurrentDynamicValue
func TestValueInThePastBad1(t *testing.T) {
	t.Parallel()
	s := InitializeStorageWithFirstNode()
	err := s.database.rawDB.Update(func(txn *badger.Txn) error {
		_, _, err := s.getDynamicValueInThePast(txn, 0)
		return err
	})
	if !errors.Is(err, ErrZeroEpoch) {
		t.Fatalf("Should have raised ErrZeroEpoch raise instead: %v", err)
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

func TestStorageAddEmptyNodeTail(t *testing.T) {
	t.Parallel()
	origEpoch := uint32(1)
	s := InitializeStorageWithFirstNode()

	err := s.database.rawDB.Update(func(txn *badger.Txn) error {
		ll, err := s.database.GetLinkedList(txn)
		if err != nil {
			t.Fatal(err)
		}
		headNode, err := s.database.GetNode(txn, origEpoch)
		if err != nil {
			t.Fatal(err)
		}
		err = headNode.Validate()
		if err != nil {
			t.Fatal("headNode should be valid")
		}
		if !headNode.IsTail() || !headNode.IsHead() {
			t.Fatal("first should be tail and head")
		}
		err = s.addNode(txn, ll, origEpoch+1, &DynamicValues{})
		if err == nil {
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

}

func TestStorageAddNodeToPreExistingEpoch(t *testing.T) {
	origEpoch := uint32(1)
	s := InitializeStorageWithFirstNode()

	err := s.database.rawDB.Update(func(txn *badger.Txn) error {
		ll, err := s.database.GetLinkedList(txn)
		if err != nil {
			t.Fatal(err)
		}
		headNode, err := s.database.GetNode(txn, origEpoch)
		if err != nil {
			t.Fatal(err)
		}
		if err := headNode.Validate(); err != nil {
			t.Fatal("headNode should be valid")
		}

		_, dv := GetStandardDynamicValue()
		err = s.addNode(txn, ll, 1, dv)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Try to add a node that is not after the current tail
func TestStorageAddNodeToMiddleEpoch(t *testing.T) {
	s := InitializeStorageWithFirstNode()
	epoch := uint32(25519)
	newValueRaw, dvTrue := GetDynamicValueMostZeros()
	_, err := dvTrue.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	// adding the new node away in the future
	ChangeDynamicValues(s, epoch, newValueRaw)
	raw, _ := GetStandardDynamicValue()
	err = s.database.rawDB.Update(func(txn *badger.Txn) error {
		err := s.ChangeDynamicValues(txn, 500, raw)
		return err
	})
	expectedErr := &ErrInvalidNode{}
	if !errors.As(err, &expectedErr) {
		t.Fatal(err)
	}
}
