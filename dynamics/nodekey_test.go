package dynamics

import (
	"bytes"
	"errors"
	"testing"

	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/utils"
)

func TestNodeMakeKeys(t *testing.T) {
	epoch := uint32(0)
	_, err := makeNodeKey(epoch)
	if !errors.Is(err, ErrZeroEpoch) {
		t.Fatal("Should have returned error for zero epoch")
	}

	nk := &NodeKey{}
	_, err = nk.Marshal()
	if err == nil {
		t.Fatal("Should have raised error")
	}

	epoch = 1
	nk, err = makeNodeKey(epoch)
	if err != nil {
		t.Fatal(err)
	}
	if nk.epoch != epoch {
		t.Fatal("epochs do not match")
	}
	if !bytes.Equal(nk.prefix, dbprefix.PrefixStorageNodeKey()) {
		t.Fatal("prefixes do not match (1)")
	}
	nkBytes, err := nk.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	nkTrue := []byte{}
	nkTrue = append(nkTrue, dbprefix.PrefixStorageNodeKey()...)
	epochBytes := utils.MarshalUint32(epoch)
	nkTrue = append(nkTrue, epochBytes...)
	if !bytes.Equal(nkBytes, nkTrue) {
		t.Fatal("invalid marshalling")
	}

	llk := makeLinkedListKey()
	if llk.epoch != 0 {
		t.Fatal("epoch should be 0")
	}
	if !bytes.Equal(nk.prefix, dbprefix.PrefixStorageNodeKey()) {
		t.Fatal("prefixes do not match (2)")
	}
}
