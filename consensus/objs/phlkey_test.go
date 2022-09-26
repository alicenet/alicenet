package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
)

func phlkEqual(t *testing.T, phlk, phlk2 *PendingHdrLeafKey) {
	if !bytes.Equal(phlk.Prefix, phlk2.Prefix) {
		t.Fatal("fail")
	}
	if len(phlk.Prefix) != 2 {
		t.Fatal("fail")
	}
	if !bytes.Equal(phlk.Key, phlk2.Key) {
		t.Fatal("fail")
	}
	if len(phlk.Key) != constants.HashLen {
		t.Fatal("fail")
	}
}

func TestPendingHdrLeafKey(t *testing.T) {
	prefix := []byte("Pr")
	key := crypto.Hasher([]byte("Key"))
	phlk := &PendingHdrLeafKey{
		Prefix: prefix,
		Key:    key,
	}
	data, err := phlk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	phlk2 := &PendingHdrLeafKey{}
	err = phlk2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	phlkEqual(t, phlk, phlk2)
}

func TestPendingHdrLeafKeyBad(t *testing.T) {
	phlk := &PendingHdrLeafKey{}
	_, err := phlk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	prefixBad := []byte("Prefix")
	prefixGood := []byte("Pr")
	keyBad := make([]byte, 33)
	keyGood := make([]byte, constants.HashLen)

	phlk = &PendingHdrLeafKey{
		Prefix: prefixBad,
		Key:    keyGood,
	}
	_, err = phlk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	phlk = &PendingHdrLeafKey{
		Prefix: prefixGood,
		Key:    keyBad,
	}
	_, err = phlk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	phlk = &PendingHdrLeafKey{
		Prefix: prefixGood,
		Key:    keyGood,
	}
	_, err = phlk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	badData := make([]byte, 33)
	err = phlk.UnmarshalBinary(badData)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
}
