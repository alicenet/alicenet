package dynamics

import (
	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/utils"
)

// LinkedList is a doubly linked list which will store nodes corresponding
// to changes to dynamic parameters.
// We store the largest epoch which has been updated.
type LinkedList struct {
	epochLastUpdated uint32
}

func makeLinkedListKey() *NodeKey {
	nk := &NodeKey{
		prefix: dbprefix.PrefixStorageNodeKey(),
		epoch:  0,
	}
	return nk
}

// GetEpochLastUpdated returns highest epoch with changes
func (ll *LinkedList) GetEpochLastUpdated() uint32 {
	return ll.epochLastUpdated
}

// SetEpochLastUpdated returns highest epoch with changes
func (ll *LinkedList) SetEpochLastUpdated(epoch uint32) error {
	if epoch == 0 {
		return ErrZeroEpoch
	}
	ll.epochLastUpdated = epoch
	return nil
}

// Marshal marshals LinkedList
func (ll *LinkedList) Marshal() []byte {
	eluBytes := utils.MarshalUint32(ll.epochLastUpdated)
	return eluBytes
}

// Unmarshal unmarshals LinkedList
func (ll *LinkedList) Unmarshal(v []byte) error {
	if len(v) != 4 {
		return ErrInvalidNode
	}
	elu, _ := utils.UnmarshalUint32(v[0:4])
	ll.epochLastUpdated = elu
	return nil
}

// IsValid returns true if LinkedList is valid
func (ll *LinkedList) IsValid() bool {
	return ll.epochLastUpdated != 0
}

// CreateLinkedList creates the first node in a LinkedList.
// These values can then be stored in the database.
func CreateLinkedList(epoch uint32, rs *RawStorage) (*Node, *LinkedList, error) {
	if epoch == 0 {
		return nil, nil, ErrZeroEpoch
	}
	rsCopy, err := rs.Copy()
	if err != nil {
		return nil, nil, err
	}
	node := &Node{
		thisEpoch:  epoch,
		prevEpoch:  epoch,
		nextEpoch:  epoch,
		rawStorage: rsCopy,
	}
	linkedList := &LinkedList{
		epochLastUpdated: epoch,
	}
	return node, linkedList, nil
}
