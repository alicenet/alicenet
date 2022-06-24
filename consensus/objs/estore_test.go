package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/crypto"
)

type mockKeyResolver struct {
}

func (m *mockKeyResolver) GetKey(kid []byte) ([]byte, error) {
	_ = kid
	return crypto.Hasher([]byte("kidSecret")), nil
}

func TestEstore(t *testing.T) {
	mkr := &mockKeyResolver{}
	es := &EncryptedStore{
		Kid:       []byte("foo"),
		Name:      []byte("bar"),
		ClearText: []byte("secret"),
	}
	err := es.Encrypt(mkr)
	if err != nil {
		t.Fatal(err)
	}
	if es.ClearText != nil {
		t.Fatal("ClearText should be nil!")
	}
	esb, err := es.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	es2 := &EncryptedStore{}
	err = es2.UnmarshalBinary(esb)
	if err != nil {
		t.Fatal(err)
	}
	err = es2.Decrypt(mkr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(es2.ClearText, []byte("secret")) {
		t.Fatal("did not decrypt")
	}
}
