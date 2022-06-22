package dynamics

import (
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/constants"
)

func TestUpdateGet(t *testing.T) {
	update := &Update{}
	if update.Name() != "" {
		t.Fatal("Should have raised error (1)")
	}
	if update.Type() != UpdateType(0) {
		t.Fatal("Should have raised error (2)")
	}
	if update.Value() != "" {
		t.Fatal("Should have raised error (3)")
	}
	if update.Epoch() != 0 {
		t.Fatal("Should have raised error (4)")
	}
}

func TestUpdateNewUpdate(t *testing.T) {
	fieldBad := ""
	value := "123456789"
	epoch := uint32(1)
	_, err := NewUpdate(fieldBad, value, epoch)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	fieldGood := "maxBytes"
	update, err := NewUpdate(fieldGood, value, epoch)
	if err != nil {
		t.Fatal(err)
	}
	if update.Name() != fieldGood {
		t.Fatal("invalid update name")
	}
	if update.Type() != MaxBytesType {
		t.Fatal("invalid update type")
	}
	if update.Value() != value {
		t.Fatal("invalid update value")
	}
	if update.Epoch() != epoch {
		t.Fatal("invalid update epoch")
	}
}

func TestStringToInt32(t *testing.T) {
	value := ""
	_, err := stringToInt32(value)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	value = "0"
	v, err := stringToInt32(value)
	if err != nil {
		t.Fatal(err)
	}
	if v != 0 {
		t.Fatal("invalid conversion (1)")
	}

	value = "1"
	v, err = stringToInt32(value)
	if err != nil {
		t.Fatal(err)
	}
	if v != 1 {
		t.Fatal("invalid conversion (2)")
	}

	value = "2147483647" // maxInt32
	v, err = stringToInt32(value)
	if err != nil {
		t.Fatal(err)
	}
	if v != 2147483647 {
		t.Fatal("invalid conversion (3)")
	}

	value = "2147483648"
	_, err = stringToInt32(value)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestStringToUint32(t *testing.T) {
	value := ""
	_, err := stringToUint32(value)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	value = "0"
	v, err := stringToUint32(value)
	if err != nil {
		t.Fatal(err)
	}
	if v != 0 {
		t.Fatal("invalid conversion (1)")
	}

	value = "1"
	v, err = stringToUint32(value)
	if err != nil {
		t.Fatal(err)
	}
	if v != 1 {
		t.Fatal("invalid conversion (2)")
	}

	value = "4294967295"
	v, err = stringToUint32(value)
	if err != nil {
		t.Fatal(err)
	}
	if v != constants.MaxUint32 {
		t.Fatal("invalid conversion (3)")
	}

	value = "4294967296"
	_, err = stringToUint32(value)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestStringToTimeDuration(t *testing.T) {
	value := ""
	_, err := stringToTimeDuration(value)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	value = "0"
	v, err := stringToTimeDuration(value)
	if err != nil {
		t.Fatal(err)
	}
	if v != 0*time.Second {
		t.Fatal("invalid conversion (1)")
	}

	value = "1"
	v, err = stringToTimeDuration(value)
	if err != nil {
		t.Fatal(err)
	}
	if v != 1*time.Nanosecond {
		t.Fatal("invalid conversion (2)")
	}

	value = "1000"
	v, err = stringToTimeDuration(value)
	if err != nil {
		t.Fatal(err)
	}
	if v != 1*time.Microsecond {
		t.Fatal("invalid conversion (3)")
	}

	value = "1000000"
	v, err = stringToTimeDuration(value)
	if err != nil {
		t.Fatal(err)
	}
	if v != 1*time.Millisecond {
		t.Fatal("invalid conversion (4)")
	}

	value = "1000000000"
	v, err = stringToTimeDuration(value)
	if err != nil {
		t.Fatal(err)
	}
	if v != 1*time.Second {
		t.Fatal("invalid conversion (5)")
	}

	value = "9223372036854775807"
	v, err = stringToTimeDuration(value)
	if err != nil {
		t.Fatal(err)
	}
	if v != (9223372036*time.Second + 854775807*time.Nanosecond) {
		t.Fatal("invalid conversion (6)")
	}

	value = "9223372036854775808"
	_, err = stringToTimeDuration(value)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	value = "-1"
	_, err = stringToTimeDuration(value)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func TestStringToBigInt(t *testing.T) {
	value := ""
	_, err := stringToBigInt(value)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	vTrue := big.NewInt(0)
	value = "0"
	v, err := stringToBigInt(value)
	if err != nil {
		t.Fatal(err)
	}
	if v.Cmp(vTrue) != 0 {
		t.Fatal("invalid conversion (1)")
	}

	vTrue = big.NewInt(1)
	value = "1"
	v, err = stringToBigInt(value)
	if err != nil {
		t.Fatal(err)
	}
	if v.Cmp(vTrue) != 0 {
		t.Fatal("invalid conversion (2)")
	}

	vTrue = big.NewInt(25519)
	value = "25519"
	v, err = stringToBigInt(value)
	if err != nil {
		t.Fatal(err)
	}
	if v.Cmp(vTrue) != 0 {
		t.Fatal("invalid conversion (3)")
	}

	value = "-1"
	_, err = stringToBigInt(value)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func TestConvertFieldToType(t *testing.T) {
	field := ""
	_, err := convertFieldToType(field)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	field = "maxBytes"
	uType, err := convertFieldToType(field)
	if err != nil {
		t.Fatal(err)
	}
	if uType != MaxBytesType {
		t.Fatal("Incorrect UpdateType (1)")
	}

	field = "proposalStepTimeout"
	uType, err = convertFieldToType(field)
	if err != nil {
		t.Fatal(err)
	}
	if uType != ProposalStepTimeoutType {
		t.Fatal("Incorrect UpdateType (2)")
	}

	field = "preVoteStepTimeout"
	uType, err = convertFieldToType(field)
	if err != nil {
		t.Fatal(err)
	}
	if uType != PreVoteStepTimeoutType {
		t.Fatal("Incorrect UpdateType (3)")
	}

	field = "preCommitStepTimeout"
	uType, err = convertFieldToType(field)
	if err != nil {
		t.Fatal(err)
	}
	if uType != PreCommitStepTimeoutType {
		t.Fatal("Incorrect UpdateType (4)")
	}

	field = "msgTimeout"
	uType, err = convertFieldToType(field)
	if err != nil {
		t.Fatal(err)
	}
	if uType != MsgTimeoutType {
		t.Fatal("Incorrect UpdateType (5)")
	}

	field = "minTxFee"
	uType, err = convertFieldToType(field)
	if err != nil {
		t.Fatal(err)
	}
	if uType != MinTxFeeType {
		t.Fatal("Incorrect UpdateType (6)")
	}

	field = "txValidVersion"
	uType, err = convertFieldToType(field)
	if err != nil {
		t.Fatal(err)
	}
	if uType != TxValidVersionType {
		t.Fatal("Incorrect UpdateType (7)")
	}

	field = "valueStoreFee"
	uType, err = convertFieldToType(field)
	if err != nil {
		t.Fatal(err)
	}
	if uType != ValueStoreFeeType {
		t.Fatal("Incorrect UpdateType (8)")
	}

	field = "valueStoreValidVersion"
	uType, err = convertFieldToType(field)
	if err != nil {
		t.Fatal(err)
	}
	if uType != ValueStoreValidVersionType {
		t.Fatal("Incorrect UpdateType (9)")
	}

	field = "atomicSwapFee"
	uType, err = convertFieldToType(field)
	if err != nil {
		t.Fatal(err)
	}
	if uType != AtomicSwapFeeType {
		t.Fatal("Incorrect UpdateType (10)")
	}

	field = "atomicSwapValidStopEpoch"
	uType, err = convertFieldToType(field)
	if err != nil {
		t.Fatal(err)
	}
	if uType != AtomicSwapValidStopEpochType {
		t.Fatal("Incorrect UpdateType (11)")
	}

	field = "dataStoreEpochFee"
	uType, err = convertFieldToType(field)
	if err != nil {
		t.Fatal(err)
	}
	if uType != DataStoreEpochFeeType {
		t.Fatal("Incorrect UpdateType (12)")
	}

	field = "dataStoreValidVersion"
	uType, err = convertFieldToType(field)
	if err != nil {
		t.Fatal(err)
	}
	if uType != DataStoreValidVersionType {
		t.Fatal("Incorrect UpdateType (13)")
	}
}
