package dynamics

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/constants/dbprefix"
)

func TestLinkedListMakeKeys(t *testing.T) {
	llk := makeLinkedListKey()
	if llk.epoch != 0 {
		t.Fatal("epoch should be 0")
	}
	if !bytes.Equal(llk.prefix, dbprefix.PrefixStorageNodeKey()) {
		t.Fatal("prefixes do not match")
	}
	llkBytes, err := llk.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	llkTrue := []byte{}
	llkTrue = append(llkTrue, dbprefix.PrefixStorageNodeKey()...)
	llkTrue = append(llkTrue, 0, 0, 0, 0)
	if !bytes.Equal(llkBytes, llkTrue) {
		t.Fatal("marshalled bytes do not match")
	}
}

func TestLinkedListMarshal(t *testing.T) {
	ll := &LinkedList{}
	if ll.IsValid() {
		t.Fatal("Should not have valid LinkedList")
	}

	invalidBytes := []byte{0, 1, 2, 3, 4}
	err := ll.Unmarshal(invalidBytes)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	invalidBytes2 := make([]byte, 4)
	err = ll.Unmarshal(invalidBytes2)
	if err != nil {
		t.Fatal(err)
	}
	if ll.epochLastUpdated != 0 {
		t.Fatal("Should have raised error (3)")
	}

	v := []byte{255, 255, 255, 255}
	err = ll.Unmarshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if ll.epochLastUpdated != constants.MaxUint32 {
		t.Fatal("Invalid LinkedList (1)")
	}

	retBytes := ll.Marshal()
	if !bytes.Equal(retBytes, v) {
		t.Fatal("invalid marshalled bytes")
	}
}

func TestLinkedListGetSet(t *testing.T) {
	ll := &LinkedList{}
	err := ll.SetEpochLastUpdated(0)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	elu := uint32(123456)
	err = ll.SetEpochLastUpdated(elu)
	if err != nil {
		t.Fatal(err)
	}
	retElu := ll.GetEpochLastUpdated()
	if retElu != elu {
		t.Fatal("Invalid EpochLastUpdated")
	}

	if !ll.IsValid() {
		t.Fatal("LinkedList should be valid")
	}
}

func TestCreateLinkedList(t *testing.T) {
	epoch := uint32(0)
	_, _, err := CreateLinkedList(epoch, nil)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	epoch = 1
	_, _, err = CreateLinkedList(epoch, nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	rs := &RawStorage{}
	rs.standardParameters()
	node, linkedlist, err := CreateLinkedList(epoch, rs)
	if err != nil {
		t.Fatal(err)
	}
	if node.thisEpoch != epoch {
		t.Fatal("invalid thisEpoch")
	}
	if node.prevEpoch != epoch {
		t.Fatal("invalid prevEpoch")
	}
	if node.nextEpoch != epoch {
		t.Fatal("invalid nextEpoch")
	}
	if linkedlist.epochLastUpdated != epoch {
		t.Fatal("invalid epochLastUpdated")
	}
}
