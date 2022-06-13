package db

import (
	"bytes"
	"testing"
)

func TestMerkleProof(t *testing.T) {
	t.Parallel()
	mp0 := &MerkleProof{
		Included:   true,
		KeyHeight:  256,
		Key:        []byte("Key.HashHashHashHashHashHashHash"),
		ProofKey:   []byte("PKey.HashHashHashHashHashHashHas"),
		ProofValue: []byte("Val.HashHashHashHashHashHashHash"),
		Bitmap:     []byte("btmp"),
		Path:       [][]byte{[]byte("PathHashHashHashHashHashHashHash"), []byte("PathHashHashHashHashHashHashHash"), []byte("PathHashHashHashHashHashHashHash")},
	}
	mp0bin, err := mp0.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	mp1 := &MerkleProof{}
	err = mp1.UnmarshalBinary(mp0bin)
	if err != nil {
		t.Fatal(err)
	}
	if mp1.Included != mp0.Included {
		t.Fatalf("bad height: %t", mp1.Included)
	}
	if mp1.KeyHeight != mp0.KeyHeight {
		t.Fatalf("bad height: %d", mp1.KeyHeight)
	}
	if !bytes.Equal(mp1.Key, mp0.Key) {
		t.Fatalf("bad Key: %x", mp1.Key)
	}
	if !bytes.Equal(mp1.ProofKey, mp0.ProofKey) {
		t.Fatalf("bad ProofKey: %x", mp1.ProofKey)
	}
	if !bytes.Equal(mp1.ProofValue, mp0.ProofValue) {
		t.Fatalf("bad Next: %x", mp1.ProofValue)
	}
	if !bytes.Equal(mp1.Bitmap, mp0.Bitmap) {
		t.Fatalf("bad Bitmap: %x", mp1.Bitmap)
	}
	for i := 0; i < len(mp0.Path); i++ {
		if !bytes.Equal(mp1.Path[i], mp0.Path[i]) {
			t.Fatalf("bad Path: %s", mp1.Path[i])
		}
	}
}
