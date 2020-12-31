package objs

import (
	"bytes"
	"testing"
)

func vskEqual(t *testing.T, vsk, vsk2 *ValidatorSetKey) {
	if !bytes.Equal(vsk.Prefix, vsk2.Prefix) {
		t.Fatal("fail")
	}
	if len(vsk.Prefix) != 2 {
		t.Fatal("fail")
	}
	if vsk.NotBefore != vsk2.NotBefore {
		t.Fatal("fail")
	}
	if vsk.NotBefore == 0 {
		t.Fatal("fail")
	}
}

func TestValidatorSetKey(t *testing.T) {
	prefix := []byte("Pr")
	nb := uint32(1)
	vsk := &ValidatorSetKey{
		Prefix:    prefix,
		NotBefore: nb,
	}
	data, err := vsk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	vsk2 := &ValidatorSetKey{}
	err = vsk2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	vskEqual(t, vsk, vsk2)
}

func TestValidatorSetKeyBad(t *testing.T) {
	vsk := &ValidatorSetKey{}
	_, err := vsk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	dataBad0 := make([]byte, 7)
	err = vsk.UnmarshalBinary(dataBad0)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	dataBad1 := make([]byte, 6)
	err = vsk.UnmarshalBinary(dataBad1)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	notBeforeBad := uint32(0)
	notBeforeGood := uint32(1)
	prefixBad := []byte("Prefix")
	prefixGood := []byte("Pr")
	vsk = &ValidatorSetKey{
		Prefix:    prefixBad,
		NotBefore: notBeforeGood,
	}
	_, err = vsk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	vsk = &ValidatorSetKey{
		Prefix:    prefixGood,
		NotBefore: notBeforeBad,
	}
	_, err = vsk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}

	vsk = &ValidatorSetKey{
		Prefix:    prefixGood,
		NotBefore: notBeforeGood,
	}
	_, err = vsk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
}
