package dynamics

import (
	"github.com/alicenet/alicenet/utils"
)

// LinkedList is a doubly linked list which will store nodes corresponding to
// changes to dynamic parameters. We store the latest epoch which has been
// updated and the most future epoch (tail).
type LinkedList struct {
	currentValue uint32
	tail         uint32
}

// GetEpochLastUpdated returns highest epoch with changes
func (ll *LinkedList) GetEpochLastUpdated() uint32 {
	return ll.currentValue
}

// GetEpochLastUpdated returns most highest epoch with changes
func (ll *LinkedList) GetMostFutureUpdate() uint32 {
	return ll.tail
}

// SetEpochLastUpdated returns highest epoch with changes
func (ll *LinkedList) SetEpochLastUpdated(epoch uint32) error {
	if epoch == 0 {
		return ErrZeroEpoch
	}
	ll.currentValue = epoch
	return nil
}

// SetEpochLastUpdated returns highest epoch with changes
func (ll *LinkedList) SetMostFutureUpdate(epoch uint32) error {
	if epoch == 0 {
		return ErrZeroEpoch
	}
	ll.tail = epoch
	return nil
}

// Marshal marshals LinkedList
func (ll *LinkedList) Marshal() []byte {
	headBytes := utils.MarshalUint32(ll.currentValue)
	tailBytes := utils.MarshalUint32(ll.tail)
	v := []byte{}
	v = append(v, headBytes...)
	v = append(v, tailBytes...)
	return v
}

// Unmarshal unmarshals LinkedList
func (ll *LinkedList) Unmarshal(v []byte) error {
	if len(v) != 8 {
		return ErrInvalidLinkedList
	}
	head, err := utils.UnmarshalUint32(v[0:4])
	if err != nil {
		return err
	}
	tail, err := utils.UnmarshalUint32(v[4:8])
	if err != nil {
		return err
	}
	ll.currentValue = head
	ll.tail = tail
	return nil
}

// IsValid returns true if LinkedList is valid. A linkedList is valid if it has
// at least one node (head).
func (ll *LinkedList) IsValid() bool {
	return ll.currentValue != 0
}

// CreateLinkedList creates the first node in a LinkedList.
// These values can then be stored in the database.
func CreateLinkedList(epoch uint32, dv *DynamicValues) (*Node, *LinkedList, error) {
	node, err := CreateNode(epoch, dv)
	if err != nil {
		return nil, nil, err
	}
	linkedList := &LinkedList{
		currentValue: epoch,
		tail:         epoch,
	}
	return node, linkedList, nil
}
