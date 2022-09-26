package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
)

func pnkEqual(t *testing.T, pnk, pnk2 *PendingNodeKey) {
	if !bytes.Equal(pnk.Prefix, pnk2.Prefix) {
		t.Fatal("fail")
	}
	if len(pnk.Prefix) != 2 {
		t.Fatal("fail")
	}
	if !bytes.Equal(pnk.Key, pnk2.Key) {
		t.Fatal("fail")
	}
	if len(pnk.Key) != constants.HashLen {
		t.Fatal("fail")
	}
}

func TestPendingNodeKey(t *testing.T) {
	prefix := []byte("Pr")
	key := crypto.Hasher([]byte("Key"))
	pnk := &PendingNodeKey{
		Prefix: prefix,
		Key:    key,
	}
	data, err := pnk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	pnk2 := &PendingNodeKey{}
	err = pnk2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	pnkEqual(t, pnk, pnk2)
}

func TestPendingNodeKeyBad(t *testing.T) {
	pnk := &PendingNodeKey{}
	_, err := pnk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	prefixBad := []byte("Prefix")
	prefixGood := []byte("Pr")
	keyBad := make([]byte, 33)
	keyGood := make([]byte, constants.HashLen)

	pnk = &PendingNodeKey{
		Prefix: prefixBad,
		Key:    keyGood,
	}
	_, err = pnk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	pnk = &PendingNodeKey{
		Prefix: prefixGood,
		Key:    keyBad,
	}
	_, err = pnk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	pnk = &PendingNodeKey{
		Prefix: prefixGood,
		Key:    keyGood,
	}
	_, err = pnk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	badData := make([]byte, 33)
	err = pnk.UnmarshalBinary(badData)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
}
