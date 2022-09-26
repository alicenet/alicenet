package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
)

func plkEqual(t *testing.T, plk, plk2 *PendingLeafKey) {
	if !bytes.Equal(plk.Prefix, plk2.Prefix) {
		t.Fatal("fail")
	}
	if len(plk.Prefix) != 2 {
		t.Fatal("fail")
	}
	if !bytes.Equal(plk.Key, plk2.Key) {
		t.Fatal("fail")
	}
	if len(plk.Key) != constants.HashLen {
		t.Fatal("fail")
	}
}

func TestPendingLeafKey(t *testing.T) {
	prefix := []byte("Pr")
	key := crypto.Hasher([]byte("Key"))
	plk := &PendingLeafKey{
		Prefix: prefix,
		Key:    key,
	}
	data, err := plk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	plk2 := &PendingLeafKey{}
	err = plk2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	plkEqual(t, plk, plk2)
}

func TestPendingLeafKeyBad(t *testing.T) {
	plk := &PendingLeafKey{}
	_, err := plk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	prefixBad := []byte("Prefix")
	prefixGood := []byte("Pr")
	keyBad := make([]byte, 33)
	keyGood := make([]byte, constants.HashLen)

	plk = &PendingLeafKey{
		Prefix: prefixBad,
		Key:    keyGood,
	}
	_, err = plk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	plk = &PendingLeafKey{
		Prefix: prefixGood,
		Key:    keyBad,
	}
	_, err = plk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	plk = &PendingLeafKey{
		Prefix: prefixGood,
		Key:    keyGood,
	}
	_, err = plk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	badData := make([]byte, 33)
	err = plk.UnmarshalBinary(badData)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
}
