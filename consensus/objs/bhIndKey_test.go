package objs

import (
	"testing"

	"github.com/alicenet/alicenet/constants"
)

func TestBlockHeaderHashIndexKey(t *testing.T) {
	prefix := []byte("Nu")
	hash := make([]byte, constants.HashLen)
	bhik := &BlockHeaderHashIndexKey{
		Prefix:    prefix,
		BlockHash: hash,
	}
	data, err := bhik.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	bhik = &BlockHeaderHashIndexKey{
		Prefix:    nil,
		BlockHash: hash,
	}
	_, err = bhik.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	bhik = &BlockHeaderHashIndexKey{
		Prefix:    prefix,
		BlockHash: nil,
	}
	_, err = bhik.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	bhik2 := &BlockHeaderHashIndexKey{}
	err = bhik2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}

	tooLong := make([]byte, constants.HashLen+3)
	err = bhik2.UnmarshalBinary(tooLong)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}
