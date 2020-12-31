package objs

import (
	"testing"
)

func TestBlockHeaderHeightKey(t *testing.T) {
	prefix := []byte("Nw")
	height := uint32(1)
	bhhk := &BlockHeaderHeightKey{
		Prefix: prefix,
		Height: height,
	}
	data, err := bhhk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	bhhk = &BlockHeaderHeightKey{
		Prefix: nil,
		Height: height,
	}
	_, err = bhhk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	bhhk2 := &BlockHeaderHeightKey{}
	err = bhhk2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}

	tooLong := make([]byte, 7)
	err = bhhk2.UnmarshalBinary(tooLong)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	zeros := make([]byte, 6)
	err = bhhk2.UnmarshalBinary(zeros)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

}
