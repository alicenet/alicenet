package dynamics

import (
	"bytes"
	"errors"
	"math/big"
	"strconv"
	"testing"
	"time"
)

func TestRawStorageMarshal(t *testing.T) {
	rs := &RawStorage{}
	_, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	s := &Storage{}
	_, err = s.rawStorage.Marshal()
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestRawStorageUnmarshal(t *testing.T) {
	rs := &RawStorage{}
	v, err := rs.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	rs2 := &RawStorage{}
	err = rs2.Unmarshal(v)
	if err != nil {
		t.Fatal(err)
	}

	v = []byte{}
	rs3 := &RawStorage{}
	err = rs3.Unmarshal(v)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	s := &Storage{}
	err = s.rawStorage.Unmarshal(v)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestRawStorageCopy(t *testing.T) {
	// Copy empty RawStorage
	rs1 := &RawStorage{}
	rs2, err := rs1.Copy()
	if err != nil {
		t.Fatal(err)
	}
	rs1Bytes, err := rs1.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	rs2Bytes, err := rs2.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rs1Bytes, rs2Bytes) {
		t.Fatal("Should have equal bytes (1)")
	}

	// Copy RawStorage with parameters
	rs1.standardParameters()
	rs2, err = rs1.Copy()
	if err != nil {
		t.Fatal(err)
	}
	rs1Bytes, err = rs1.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	rs2Bytes, err = rs2.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rs1Bytes, rs2Bytes) {
		t.Fatal("Should have equal bytes (2)")
	}

	// Copy RawStorage with some parameters set to zero
	rs1.MaxBytes = 0
	rs2, err = rs1.Copy()
	if err != nil {
		t.Fatal(err)
	}
	rs1Bytes, err = rs1.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	rs2Bytes, err = rs2.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rs1Bytes, rs2Bytes) {
		t.Fatal("Should have equal bytes (3)")
	}

	s := &Storage{}
	_, err = s.rawStorage.Copy()
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestRawStorageStandardParameters(t *testing.T) {
	rs := &RawStorage{}
	rs.standardParameters()

	retMaxBytes := rs.GetMaxBytes()
	if retMaxBytes != maxBytes {
		t.Fatal("Should be equal (1)")
	}

	retMaxProposalSize := rs.GetMaxProposalSize()
	if retMaxProposalSize != maxProposalSize {
		t.Fatal("Should be equal (2)")
	}

	retSrvrMsgTimeout := rs.GetSrvrMsgTimeout()
	if retSrvrMsgTimeout != srvrMsgTimeout {
		t.Fatal("Should be equal (3)")
	}

	retMsgTimeout := rs.GetMsgTimeout()
	if retMsgTimeout != msgTimeout {
		t.Fatal("Should be equal (4)")
	}

	retProposalTimeout := rs.GetProposalStepTimeout()
	if retProposalTimeout != proposalStepTO {
		t.Fatal("Should be equal (5)")
	}

	retPreVoteTimeout := rs.GetPreVoteStepTimeout()
	if retPreVoteTimeout != preVoteStepTO {
		t.Fatal("Should be equal (6)")
	}

	retPreCommitTimeout := rs.GetPreCommitStepTimeout()
	if retPreCommitTimeout != preCommitStepTO {
		t.Fatal("Should be equal (7)")
	}

	retDBRNRTO := rs.GetDeadBlockRoundNextRoundTimeout()
	if retDBRNRTO != dBRNRTO {
		t.Fatal("Should be equal (8)")
	}

	retDownloadTO := rs.GetDownloadTimeout()
	if retDownloadTO != downloadTO {
		t.Fatal("Should be equal (9)")
	}
}

func TestRawStorageMaxBytes(t *testing.T) {
	rs := &RawStorage{}
	retMaxBytes0 := rs.GetMaxBytes()
	if retMaxBytes0 != 0 {
		t.Fatal("Should be zero")
	}

	value := uint32(10000)
	rs.SetMaxBytes(value)
	retMaxBytes := rs.GetMaxBytes()
	if retMaxBytes != value {
		t.Fatal("Should be equal (1)")
	}

	retMaxProposalSize := rs.GetMaxProposalSize()
	if retMaxProposalSize != value {
		t.Fatal("Should be equal (2)")
	}
}

func TestRawStorageMaxProposalSize(t *testing.T) {
	rs := &RawStorage{}
	retMaxProposalSize0 := rs.GetMaxProposalSize()
	if retMaxProposalSize0 != 0 {
		t.Fatal("Should be zero (2)")
	}

	value := uint32(10000)
	rs.SetMaxBytes(value)
	retMaxProposalSize := rs.GetMaxProposalSize()
	if retMaxProposalSize != value {
		t.Fatal("Should be equal (2)")
	}
}

func TestRawStorageMsgTimeout(t *testing.T) {
	rs := &RawStorage{}
	retMsgTimeout0 := rs.GetMsgTimeout()
	if retMsgTimeout0 != 0 {
		t.Fatal("Should be zero")
	}

	value := time.Second
	rs.SetMsgTimeout(value)
	retMsgTimeout := rs.GetMsgTimeout()
	if retMsgTimeout != value {
		t.Fatal("Should be equal (1)")
	}

	valueSrvrMsg := (3 * value) / 4
	retSrvrMsgTimeout := rs.GetSrvrMsgTimeout()
	if retSrvrMsgTimeout != valueSrvrMsg {
		t.Fatal("Should be equal (2)")
	}
}

func TestRawStorageSrvrMsgTimeout(t *testing.T) {
	rs := &RawStorage{}
	retSrvrMsgTimeout0 := rs.GetSrvrMsgTimeout()
	if retSrvrMsgTimeout0 != 0 {
		t.Fatal("Should be zero")
	}

	value := time.Second
	rs.SetMsgTimeout(value)
	valueSrvrMsg := (3 * value) / 4
	retSrvrMsgTimeout := rs.GetSrvrMsgTimeout()
	if retSrvrMsgTimeout != valueSrvrMsg {
		t.Fatal("Should be equal")
	}
}

func TestRawStorageConsensusTimeouts(t *testing.T) {
	rs := &RawStorage{}

	retPropTOv0 := rs.GetProposalStepTimeout()
	if retPropTOv0 != 0 {
		t.Fatal("Should be zero (1)")
	}
	retPreVoteTOv0 := rs.GetPreVoteStepTimeout()
	if retPreVoteTOv0 != 0 {
		t.Fatal("Should be zero (2)")
	}
	retPreCommitTOv0 := rs.GetPreCommitStepTimeout()
	if retPreCommitTOv0 != 0 {
		t.Fatal("Should be zero (3)")
	}
	retDownloadTOv0 := rs.GetDownloadTimeout()
	if retDownloadTOv0 != 0 {
		t.Fatal("Should be zero (4)")
	}
	retDBRNRTOv0 := rs.GetDeadBlockRoundNextRoundTimeout()
	if retDBRNRTOv0 != 0 {
		t.Fatal("Should be zero (5)")
	}

	propValue := 10 * time.Second
	rs.SetProposalStepTimeout(propValue)

	retPropTOv1 := rs.GetProposalStepTimeout()
	if retPropTOv1 != propValue {
		t.Fatal("Should be equal (1)")
	}
	retPreVoteTOv1 := rs.GetPreVoteStepTimeout()
	if retPreVoteTOv1 != 0 {
		t.Fatal("Should be zero (6)")
	}
	retPreCommitTOv1 := rs.GetPreCommitStepTimeout()
	if retPreCommitTOv1 != 0 {
		t.Fatal("Should be zero (7)")
	}
	retDownloadTOv1 := rs.GetDownloadTimeout()
	if retDownloadTOv1 != propValue {
		t.Fatal("Should be equal (2)")
	}
	retDBRNRTOv1 := rs.GetDeadBlockRoundNextRoundTimeout()
	if retDBRNRTOv1 != ((5 * propValue) / 2) {
		t.Fatal("Should be equal (3)")
	}

	preVoteValue := 20 * time.Second
	rs.SetPreVoteStepTimeout(preVoteValue)

	retPropTOv2 := rs.GetProposalStepTimeout()
	if retPropTOv2 != propValue {
		t.Fatal("Should be equal (4)")
	}
	retPreVoteTOv2 := rs.GetPreVoteStepTimeout()
	if retPreVoteTOv2 != preVoteValue {
		t.Fatal("Should be equal (5)")
	}
	retPreCommitTOv2 := rs.GetPreCommitStepTimeout()
	if retPreCommitTOv2 != 0 {
		t.Fatal("Should be zero (8)")
	}
	retDownloadTOv2 := rs.GetDownloadTimeout()
	if retDownloadTOv2 != (propValue + preVoteValue) {
		t.Fatal("Should be equal (6)")
	}
	retDBRNRTOv2 := rs.GetDeadBlockRoundNextRoundTimeout()
	if retDBRNRTOv2 != ((5 * (propValue + preVoteValue)) / 2) {
		t.Fatal("Should be equal (7)")
	}

	preCommitValue := 30 * time.Second
	rs.SetPreCommitStepTimeout(preCommitValue)

	retPropTOv3 := rs.GetProposalStepTimeout()
	if retPropTOv3 != propValue {
		t.Fatal("Should be equal (8)")
	}
	retPreVoteTOv3 := rs.GetPreVoteStepTimeout()
	if retPreVoteTOv3 != preVoteValue {
		t.Fatal("Should be equal (9)")
	}
	retPreCommitTOv3 := rs.GetPreCommitStepTimeout()
	if retPreCommitTOv3 != preCommitValue {
		t.Fatal("Should be equal (10)")
	}
	retDownloadTOv3 := rs.GetDownloadTimeout()
	if retDownloadTOv3 != (propValue + preVoteValue + preCommitValue) {
		t.Fatal("Should be equal (11)")
	}
	retDBRNRTOv3 := rs.GetDeadBlockRoundNextRoundTimeout()
	if retDBRNRTOv3 != ((5 * (propValue + preVoteValue + preCommitValue)) / 2) {
		t.Fatal("Should be equal (12)")
	}
}

func TestRawStorageMinTxFee(t *testing.T) {
	rs1 := &RawStorage{}

	v1 := rs1.GetMinTxFeeCostRatio()
	if v1.Sign() != 0 {
		t.Fatal("minTxFeeCostRatio should be 0")
	}

	rs2 := &RawStorage{}
	err := rs2.SetMinTxFeeCostRatio(nil)
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatal("Should have raised ErrInvalidValue")
	}

	rs3 := &RawStorage{}
	value := new(big.Int).SetInt64(-1)
	err = rs3.SetMinTxFeeCostRatio(value)
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatal("Should have raised ErrInvalidValue")
	}

	rs4 := &RawStorage{}
	value4int := int64(1234567890)
	value4 := new(big.Int).SetInt64(value4int)
	err = rs4.SetMinTxFeeCostRatio(value4)
	if err != nil {
		t.Fatal(err)
	}
	v4 := rs4.GetMinTxFeeCostRatio()
	if v4.Cmp(big.NewInt(value4int)) != 0 {
		t.Fatal("incorrect minTxFeeCostRatio value")
	}
}

func TestRawStorageTxValidVersion(t *testing.T) {
	rs1 := &RawStorage{}
	v1 := rs1.GetTxValidVersion()
	if v1 != 0 {
		t.Fatal("invalid TxValidVersion")
	}

	rs2 := &RawStorage{}
	version2 := uint32(7919)
	rs2.SetTxValidVersion(version2)
	v2 := rs2.GetTxValidVersion()
	if v2 != version2 {
		t.Fatal("TxValidVersions do not match")
	}
}

func TestRawStorageValueStoreFee(t *testing.T) {
	rs1 := &RawStorage{}

	v1 := rs1.GetValueStoreFee()
	if v1.Sign() != 0 {
		t.Fatal("minValueStoreBurnedFee should be 0")
	}

	rs2 := &RawStorage{}
	err := rs2.SetValueStoreFee(nil)
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatal("Should have raised ErrInvalidValue")
	}

	rs3 := &RawStorage{}
	value := new(big.Int).SetInt64(-1)
	err = rs3.SetValueStoreFee(value)
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatal("Should have raised ErrInvalidValue")
	}

	rs4 := &RawStorage{}
	value4int := int64(1234567890)
	value4 := new(big.Int).SetInt64(value4int)
	err = rs4.SetValueStoreFee(value4)
	if err != nil {
		t.Fatal(err)
	}
	v4 := rs4.GetValueStoreFee()
	if v4.Cmp(big.NewInt(value4int)) != 0 {
		t.Fatal("incorrect minValueStoreBurnedFee value")
	}
}

func TestRawStorageValueStoreValidVersion(t *testing.T) {
	rs1 := &RawStorage{}
	v1 := rs1.GetValueStoreValidVersion()
	if v1 != 0 {
		t.Fatal("invalid ValueStoreValidVersion")
	}

	rs2 := &RawStorage{}
	version2 := uint32(7919)
	rs2.SetValueStoreValidVersion(version2)
	v2 := rs2.GetValueStoreValidVersion()
	if v2 != version2 {
		t.Fatal("ValueStoreTxValidVersions do not match")
	}
}

func TestRawStorageAtomicSwapFee(t *testing.T) {
	rs1 := &RawStorage{}

	v1 := rs1.GetAtomicSwapFee()
	if v1.Sign() != 0 {
		t.Fatal("atomicSwapFee should be 0")
	}

	rs2 := &RawStorage{}
	err := rs2.SetAtomicSwapFee(nil)
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatal("Should have raised ErrInvalidValue")
	}

	rs3 := &RawStorage{}
	value := new(big.Int).SetInt64(-1)
	err = rs3.SetAtomicSwapFee(value)
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatal("Should have raised ErrInvalidValue")
	}

	rs4 := &RawStorage{}
	value4int := int64(1234567890)
	value4 := new(big.Int).SetInt64(value4int)
	err = rs4.SetAtomicSwapFee(value4)
	if err != nil {
		t.Fatal(err)
	}
	v4 := rs4.GetAtomicSwapFee()
	if v4.Cmp(big.NewInt(value4int)) != 0 {
		t.Fatal("incorrect atomicSwapFee value")
	}
}

func TestRawStorageAtomicSwapValidStopEpoch(t *testing.T) {
	rs1 := &RawStorage{}
	v1 := rs1.GetAtomicSwapValidStopEpoch()
	if v1 != 0 {
		t.Fatal("invalid AtomicSwapValidStopEpoch")
	}

	rs2 := &RawStorage{}
	version2 := uint32(7919)
	rs2.SetAtomicSwapValidStopEpoch(version2)
	v2 := rs2.GetAtomicSwapValidStopEpoch()
	if v2 != version2 {
		t.Fatal("AtomicSwapValidStopEpochs do not match")
	}
}

func TestRawStorageDataStoreEpochFee(t *testing.T) {
	rs1 := &RawStorage{}

	v1 := rs1.GetDataStoreEpochFee()
	if v1.Sign() != 0 {
		t.Fatal("dataStoreEpochFee should be 0")
	}

	rs2 := &RawStorage{}
	err := rs2.SetDataStoreEpochFee(nil)
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatal("Should have raised ErrInvalidValue")
	}

	rs3 := &RawStorage{}
	value := new(big.Int).SetInt64(-1)
	err = rs3.SetDataStoreEpochFee(value)
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatal("Should have raised ErrInvalidValue")
	}

	rs4 := &RawStorage{}
	value4int := int64(1234567890)
	value4 := new(big.Int).SetInt64(value4int)
	err = rs4.SetDataStoreEpochFee(value4)
	if err != nil {
		t.Fatal(err)
	}
	v4 := rs4.GetDataStoreEpochFee()
	if v4.Cmp(big.NewInt(value4int)) != 0 {
		t.Fatal("incorrect dataStoreEpochFee value")
	}
}

func TestRawStorageDataStoreValidVersion(t *testing.T) {
	rs1 := &RawStorage{}
	v1 := rs1.GetDataStoreValidVersion()
	if v1 != 0 {
		t.Fatal("invalid DataStoreTxValidVersion")
	}

	rs2 := &RawStorage{}
	version2 := uint32(7919)
	rs2.SetDataStoreValidVersion(version2)
	v2 := rs2.GetDataStoreValidVersion()
	if v2 != version2 {
		t.Fatal("DataStoreTxValidVersions do not match")
	}
}

func TestRawStorageUpdateValueBad(t *testing.T) {
	rs := &RawStorage{}
	fieldBad := "invalid"
	valueBad := ""
	update := &Update{
		name:  fieldBad,
		value: valueBad,
	}
	err := rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestRawStorageUpdateValueMaxBytes(t *testing.T) {
	rs := &RawStorage{}
	field := "maxBytes"
	valueBad1 := ""
	epoch := uint32(1)
	update, err := NewUpdate(field, valueBad1, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	retValue := rs.GetMaxBytes()
	if retValue != 0 {
		t.Fatal("Incorrect MaxBytes (1)")
	}
	retValue = rs.GetMaxProposalSize()
	if retValue != 0 {
		t.Fatal("Incorrect MaxProposalSize (1)")
	}

	valueBad2 := "-1"
	update, err = NewUpdate(field, valueBad2, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	retValue = rs.GetMaxBytes()
	if retValue != 0 {
		t.Fatal("Incorrect MaxBytes (2)")
	}
	retValue = rs.GetMaxProposalSize()
	if retValue != 0 {
		t.Fatal("Incorrect MaxProposalSize (2)")
	}

	valueGood := "1000"
	valueTrue64, err := strconv.ParseUint(valueGood, 10, 32)
	if err != nil {
		t.Fatal(err)
	}
	valueTrue := uint32(valueTrue64)
	update, err = NewUpdate(field, valueGood, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err != nil {
		t.Fatal(err)
	}
	retValue = rs.GetMaxBytes()
	if retValue != valueTrue {
		t.Fatal("Incorrect MaxBytes (3)")
	}
	retValue = rs.GetMaxProposalSize()
	if retValue != valueTrue {
		t.Fatal("Incorrect MaxProposalSize (3)")
	}
}

func TestRawStorageUpdateValueProposalStepTimeout(t *testing.T) {
	rs := &RawStorage{}

	retProposalStepTO := rs.GetProposalStepTimeout()
	if retProposalStepTO != 0 {
		t.Fatal("Incorrect ProposalStepTO (1)")
	}

	field := "proposalStepTimeout"
	valueBad1 := ""
	epoch := uint32(1)
	update, err := NewUpdate(field, valueBad1, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	valueBad2 := "-1"
	update, err = NewUpdate(field, valueBad2, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if !errors.Is(err, ErrInvalid) {
		t.Fatal("Should have raised error (2)")
	}

	valueGood := "1000000000"
	update, err = NewUpdate(field, valueGood, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue64, err := strconv.ParseInt(valueGood, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	propTrue := time.Duration(valueTrue64)
	retProposalStepTO = rs.GetProposalStepTimeout()
	if retProposalStepTO != propTrue {
		t.Fatal("Incorrect ProposalStepTO (2)")
	}
}

func TestRawStorageUpdateValuePreVoteStepTimeout(t *testing.T) {
	rs := &RawStorage{}

	retPreVoteStepTO := rs.GetPreVoteStepTimeout()
	if retPreVoteStepTO != 0 {
		t.Fatal("Incorrect PreVoteStepTO (1)")
	}

	field := "preVoteStepTimeout"
	valueBad1 := ""
	epoch := uint32(1)
	update, err := NewUpdate(field, valueBad1, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	valueBad2 := "-1"
	update, err = NewUpdate(field, valueBad2, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if !errors.Is(err, ErrInvalid) {
		t.Fatal("Should have raised error (2)")
	}

	valueGood := "1000000000"
	update, err = NewUpdate(field, valueGood, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue64, err := strconv.ParseInt(valueGood, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	preVoteTrue := time.Duration(valueTrue64)
	retPreVoteStepTO = rs.GetPreVoteStepTimeout()
	if retPreVoteStepTO != preVoteTrue {
		t.Fatal("Incorrect PreVoteStepTO (2)")
	}
}

func TestRawStorageUpdateValuePreCommitStepTimeout(t *testing.T) {
	rs := &RawStorage{}

	retPreCommitStepTO := rs.GetPreCommitStepTimeout()
	if retPreCommitStepTO != 0 {
		t.Fatal("Incorrect PreCommitStepTO (1)")
	}

	field := "preCommitStepTimeout"
	valueBad1 := ""
	epoch := uint32(1)
	update, err := NewUpdate(field, valueBad1, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	valueBad2 := "-1"
	update, err = NewUpdate(field, valueBad2, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if !errors.Is(err, ErrInvalid) {
		t.Fatal("Should have raised error (2)")
	}

	valueGood := "1000000000"
	update, err = NewUpdate(field, valueGood, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue64, err := strconv.ParseInt(valueGood, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	preCommitTrue := time.Duration(valueTrue64)
	retPreCommitStepTO = rs.GetPreCommitStepTimeout()
	if retPreCommitStepTO != preCommitTrue {
		t.Fatal("Incorrect PreCommitStepTO (2)")
	}
}

func TestRawStorageUpdateValueMsgTimeout(t *testing.T) {
	rs := &RawStorage{}

	retMsgTO := rs.GetMsgTimeout()
	if retMsgTO != 0 {
		t.Fatal("Incorrect MsgTimeout (1)")
	}

	field := "msgTimeout"
	valueBad1 := ""
	epoch := uint32(1)
	update, err := NewUpdate(field, valueBad1, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	valueBad2 := "-1"
	update, err = NewUpdate(field, valueBad2, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if !errors.Is(err, ErrInvalid) {
		t.Fatal("Should have raised error (2)")
	}

	valueGood := "1000000000"
	update, err = NewUpdate(field, valueGood, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue64, err := strconv.ParseInt(valueGood, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	msgTrue := time.Duration(valueTrue64)
	retMsgTO = rs.GetMsgTimeout()
	if retMsgTO != msgTrue {
		t.Fatal("Incorrect MsgTimeout (2)")
	}
}

func TestRawStorageUpdateMinTxBurnedFee(t *testing.T) {
	rs := &RawStorage{}

	retMinTxFee := rs.GetMinTxFeeCostRatio()
	if retMinTxFee.Sign() != 0 {
		t.Fatal("Incorrect MinTxFee (1)")
	}

	field := "minTxFeeCostRatio"
	valueBad1 := ""
	epoch := uint32(1)
	update, err := NewUpdate(field, valueBad1, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	valueBad2 := "-1"
	update, err = NewUpdate(field, valueBad2, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if !errors.Is(err, ErrInvalid) {
		t.Fatal("Should have raised ErrInvalidUpdateValue error")
	}

	valueGood := "1000000000"
	update, err = NewUpdate(field, valueGood, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue, ok := new(big.Int).SetString(valueGood, 10)
	if !ok {
		t.Fatal("SetString failed")
	}
	retMinTxFee = rs.GetMinTxFeeCostRatio()
	if retMinTxFee.Cmp(valueTrue) != 0 {
		t.Fatal("Incorrect MinTxBurnedFee (2)")
	}
}

func TestRawStorageUpdateTxValidVersion(t *testing.T) {
	rs := &RawStorage{}

	retTxValidVersion := rs.GetTxValidVersion()
	if retTxValidVersion != 0 {
		t.Fatal("Incorrect TxValidVersion (1)")
	}

	field := "txValidVersion"
	valueBad1 := ""
	epoch := uint32(1)
	update, err := NewUpdate(field, valueBad1, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	valueBad2 := "-1"
	update, err = NewUpdate(field, valueBad2, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	valueGood := "1000000000"
	update, err = NewUpdate(field, valueGood, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue64, err := strconv.ParseInt(valueGood, 10, 32)
	if err != nil {
		t.Fatal(err)
	}
	txValidTrue := uint32(valueTrue64)
	retTxValidVersion = rs.GetTxValidVersion()
	if retTxValidVersion != txValidTrue {
		t.Fatal("Incorrect TxValidVersion (2)")
	}
}

func TestRawStorageUpdateValueStoreFee(t *testing.T) {
	rs := &RawStorage{}

	retMinVSFee := rs.GetValueStoreFee()
	if retMinVSFee.Sign() != 0 {
		t.Fatal("Incorrect MinValueStoreBurnedFee (1)")
	}

	field := "valueStoreFee"
	valueBad1 := ""
	epoch := uint32(1)
	update, err := NewUpdate(field, valueBad1, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	valueBad2 := "-1"
	update, err = NewUpdate(field, valueBad2, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if !errors.Is(err, ErrInvalid) {
		t.Fatal("Should have raised ErrInvalidUpdateValue error")
	}

	valueGood := "1000000000"
	update, err = NewUpdate(field, valueGood, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue, ok := new(big.Int).SetString(valueGood, 10)
	if !ok {
		t.Fatal("SetString failed")
	}
	retMinVSFee = rs.GetValueStoreFee()
	if retMinVSFee.Cmp(valueTrue) != 0 {
		t.Fatal("Incorrect MinValueStoreBurnedFee (2)")
	}
}

func TestRawStorageUpdateValueStoreValidVersion(t *testing.T) {
	rs := &RawStorage{}

	retVSValidVersion := rs.GetValueStoreValidVersion()
	if retVSValidVersion != 0 {
		t.Fatal("Incorrect ValueStoreValidVersion (1)")
	}

	field := "valueStoreValidVersion"
	valueBad1 := ""
	epoch := uint32(1)
	update, err := NewUpdate(field, valueBad1, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	valueBad2 := "-1"
	update, err = NewUpdate(field, valueBad2, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	valueGood := "1000000000"
	update, err = NewUpdate(field, valueGood, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue64, err := strconv.ParseInt(valueGood, 10, 32)
	if err != nil {
		t.Fatal(err)
	}
	vsTxValidTrue := uint32(valueTrue64)
	retVSValidVersion = rs.GetValueStoreValidVersion()
	if retVSValidVersion != vsTxValidTrue {
		t.Fatal("Incorrect ValueStoreValidVersion (2)")
	}
}

func TestRawStorageUpdateAtomicSwapFee(t *testing.T) {
	rs := &RawStorage{}

	retMinASFee := rs.GetAtomicSwapFee()
	if retMinASFee.Sign() != 0 {
		t.Fatal("Incorrect AtomicSwapFee (1)")
	}

	field := "atomicSwapFee"
	valueBad1 := ""
	epoch := uint32(1)
	update, err := NewUpdate(field, valueBad1, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	valueBad2 := "-1"
	update, err = NewUpdate(field, valueBad2, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if !errors.Is(err, ErrInvalid) {
		t.Fatal("Should have raised ErrInvalid error")
	}

	valueGood := "1000000000"
	update, err = NewUpdate(field, valueGood, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue, ok := new(big.Int).SetString(valueGood, 10)
	if !ok {
		t.Fatal("SetString failed")
	}
	retMinASFee = rs.GetAtomicSwapFee()
	if retMinASFee.Cmp(valueTrue) != 0 {
		t.Fatal("Incorrect AtomicSwapFee (2)")
	}
}

func TestRawStorageUpdateAtomicSwapStopEpoch(t *testing.T) {
	rs := &RawStorage{}

	retAtomicSwapStopEpoch := rs.GetAtomicSwapValidStopEpoch()
	if retAtomicSwapStopEpoch != 0 {
		t.Fatal("Incorrect AtomicSwapValidStopEpoch (1)")
	}

	field := "atomicSwapValidStopEpoch"
	valueBad1 := ""
	epoch := uint32(1)
	update, err := NewUpdate(field, valueBad1, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	valueBad2 := "-1"
	update, err = NewUpdate(field, valueBad2, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	valueGood := "1000000000"
	update, err = NewUpdate(field, valueGood, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue64, err := strconv.ParseInt(valueGood, 10, 32)
	if err != nil {
		t.Fatal(err)
	}
	asTrue := uint32(valueTrue64)
	retAtomicSwapStopEpoch = rs.GetAtomicSwapValidStopEpoch()
	if retAtomicSwapStopEpoch != asTrue {
		t.Fatal("Incorrect AtomicSwapValidStopEpoch (2)")
	}
}

func TestRawStorageUpdateDataStoreEpochFee(t *testing.T) {
	rs := &RawStorage{}

	dsEpochFee := rs.GetDataStoreEpochFee()
	if dsEpochFee.Sign() != 0 {
		t.Fatal("Incorrect DataStoreEpochFee (1)")
	}

	field := "dataStoreEpochFee"
	valueBad1 := ""
	epoch := uint32(1)
	update, err := NewUpdate(field, valueBad1, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	valueBad2 := "-1"
	update, err = NewUpdate(field, valueBad2, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if !errors.Is(err, ErrInvalid) {
		t.Fatal("Should have raised ErrInvalid error")
	}

	valueGood := "1000000000"
	update, err = NewUpdate(field, valueGood, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue, ok := new(big.Int).SetString(valueGood, 10)
	if !ok {
		t.Fatal("SetString failed")
	}
	dsEpochFee = rs.GetDataStoreEpochFee()
	if dsEpochFee.Cmp(valueTrue) != 0 {
		t.Fatal("Incorrect DataStoreEpochFee (2)")
	}
}

func TestRawStorageUpdateDataStoreValidVersion(t *testing.T) {
	rs := &RawStorage{}

	retDSTxValidVersion := rs.GetDataStoreValidVersion()
	if retDSTxValidVersion != 0 {
		t.Fatal("Incorrect DataStoreValidVersion (1)")
	}

	field := "dataStoreValidVersion"
	valueBad1 := ""
	epoch := uint32(1)
	update, err := NewUpdate(field, valueBad1, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	valueBad2 := "-1"
	update, err = NewUpdate(field, valueBad2, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	valueGood := "1000000000"
	update, err = NewUpdate(field, valueGood, epoch)
	if err != nil {
		t.Fatal(err)
	}
	err = rs.UpdateValue(update)
	if err != nil {
		t.Fatal(err)
	}

	valueTrue64, err := strconv.ParseInt(valueGood, 10, 32)
	if err != nil {
		t.Fatal(err)
	}
	dsTxValidTrue := uint32(valueTrue64)
	retDSTxValidVersion = rs.GetDataStoreValidVersion()
	if retDSTxValidVersion != dsTxValidTrue {
		t.Fatal("Incorrect DataStoreTxValidVersion (2)")
	}
}
