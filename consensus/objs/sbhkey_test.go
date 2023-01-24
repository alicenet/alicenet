package objs

import (
	"bytes"
	"testing"
)

func sbhkEqual(t *testing.T, sbhk, sbhk2 *StagedBlockHeaderKey) {
	t.Helper()
	if !bytes.Equal(sbhk.Prefix, sbhk2.Prefix) {
		t.Fatal("fail")
	}
	if !bytes.Equal(sbhk.Key, sbhk2.Key) {
		t.Fatal("fail")
	}
}

func TestStagedBlockHeaderKey(t *testing.T) {
	prefix := []byte("Pr")
	key := []byte("Key0")
	sbhk := &StagedBlockHeaderKey{
		Prefix: prefix,
		Key:    key,
	}
	data, err := sbhk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	sbhk2 := &StagedBlockHeaderKey{}
	err = sbhk2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	sbhkEqual(t, sbhk, sbhk2)
}

func TestStagedBlockHeaderKeyBad(t *testing.T) {
	sbhk := &StagedBlockHeaderKey{}
	data := make([]byte, 5)
	err := sbhk.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	prefixBad := []byte("BadPrefix")
	keyGood := []byte("Key0")
	sbhk = &StagedBlockHeaderKey{
		Prefix: prefixBad,
		Key:    keyGood,
	}
	_, err = sbhk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	prefixGood := []byte("Pr")
	keyBad := []byte("BadKey")
	sbhk = &StagedBlockHeaderKey{
		Prefix: prefixGood,
		Key:    keyBad,
	}
	_, err = sbhk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}
