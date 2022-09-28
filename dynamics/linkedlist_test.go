package dynamics

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
)

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
	if ll.currentValue != 0 {
		t.Fatal("Should have raised error (3)")
	}

	if ll.tail != 0 {
		t.Fatal("Should have raised error (4)")
	}

	v := []byte{255, 255, 255, 255, 255, 255, 255, 255}
	err = ll.Unmarshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if ll.currentValue != constants.MaxUint32 {
		t.Fatal("Invalid LinkedList (1)")
	}
	if ll.tail != constants.MaxUint32 {
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

	mfu := uint32(123457)
	err = ll.SetMostFutureUpdate(mfu)
	if err != nil {
		t.Fatal(err)
	}
	retMfu := ll.GetMostFutureUpdate()
	if retMfu != mfu {
		t.Fatal("Invalid EpochLastUpdated")
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

	dv := GetStandardDynamicValue()
	node, linkedlist, err := CreateLinkedList(epoch, dv)
	if err != nil {
		t.Fatal(err)
	}
	if node.thisEpoch != epoch {
		t.Fatal("invalid thisEpoch")
	}
	if node.prevEpoch != 0 {
		t.Fatal("invalid prevEpoch")
	}
	if node.nextEpoch != 0 {
		t.Fatal("invalid nextEpoch")
	}
	if linkedlist.currentValue != epoch {
		t.Fatal("invalid currentValue")
	}
	if linkedlist.tail != epoch {
		t.Fatal("invalid currentValue")
	}
}
