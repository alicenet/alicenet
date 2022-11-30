package dynamics

import (
	"bytes"

	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/utils"
)

// NodeKey stores the necessary information to load a Node
type NodeKey struct {
	prefix []byte
	epoch  uint32
}

func makeNodeKey(epoch uint32) (*NodeKey, error) {
	if epoch == 0 {
		return nil, ErrZeroEpoch
	}
	nk := &NodeKey{
		prefix: dbprefix.PrefixStorageNodeKey(),
		epoch:  epoch,
	}
	return nk, nil
}

// Marshal converts NodeKey into the byte slice
func (nk *NodeKey) Marshal() ([]byte, error) {
	if !bytes.Equal(nk.prefix, dbprefix.PrefixStorageNodeKey()) {
		return nil, ErrInvalidNodeKey
	}
	epochBytes := utils.MarshalUint32(nk.epoch)
	key := []byte{}
	key = append(key, nk.prefix...)
	key = append(key, epochBytes...)
	return key, nil
}

// Node contains necessary information about DynamicValues;
// it also points to the epoch of the previous node and next node
// in the doubly linked list.
type Node struct {
	thisEpoch     uint32
	prevEpoch     uint32
	nextEpoch     uint32
	dynamicValues *DynamicValues
}

// CreateNode creates a unlinked node.
func CreateNode(epoch uint32, dv *DynamicValues) (*Node, error) {
	if epoch == 0 {
		return nil, ErrZeroEpoch
	}
	dvCopy, err := dv.Copy()
	if err != nil {
		return nil, err
	}
	node := &Node{
		thisEpoch:     epoch,
		prevEpoch:     0,
		nextEpoch:     0,
		dynamicValues: dvCopy,
	}
	return node, nil
}

// Marshal marshals a Node
func (n *Node) Marshal() ([]byte, error) {
	dvBytes, err := n.dynamicValues.Marshal()
	if err != nil {
		return nil, err
	}
	teBytes := utils.MarshalUint32(n.thisEpoch)
	peBytes := utils.MarshalUint32(n.prevEpoch)
	neBytes := utils.MarshalUint32(n.nextEpoch)
	v := []byte{}
	v = append(v, teBytes...)
	v = append(v, peBytes...)
	v = append(v, neBytes...)
	v = append(v, dvBytes...)
	return v, nil
}

// Unmarshal unmarshals a Node
func (n *Node) Unmarshal(v []byte) error {
	if len(v) < 12 {
		return ErrInvalid
	}
	thisEpoch, _ := utils.UnmarshalUint32(v[0:4])
	prevEpoch, _ := utils.UnmarshalUint32(v[4:8])
	nextEpoch, _ := utils.UnmarshalUint32(v[8:12])
	n.thisEpoch = thisEpoch
	n.prevEpoch = prevEpoch
	n.nextEpoch = nextEpoch
	n.dynamicValues = &DynamicValues{}
	err := n.dynamicValues.Unmarshal(v[12:])
	if err != nil {
		return err
	}
	return nil
}

// IsValid returns true if Node is valid
func (n *Node) Validate() error {
	if n == nil {
		return ErrNodeValueNilPointer
	}
	if n.thisEpoch == 0 {
		// node has not set values; invalid
		return ErrZeroEpoch
	}
	if n.prevEpoch >= n.thisEpoch || (n.nextEpoch != 0 && n.thisEpoch >= n.nextEpoch) {
		return &ErrInvalidNode{n}
	}
	return n.dynamicValues.Validate()
}

// Copy makes a copy of Node
func (n *Node) Copy() (*Node, error) {
	nodeBytes, err := n.Marshal()
	if err != nil {
		return nil, err
	}
	nodeCopy := &Node{}
	err = nodeCopy.Unmarshal(nodeBytes)
	if err != nil {
		return nil, err
	}
	return nodeCopy, nil
}

// SetEpochs sets n.prevEpoch and n.nextEpoch.
func (n *Node) SetEpochs(prevNode *Node, nextNode *Node) error {
	if err := prevNode.Validate(); err != nil {
		return err
	}

	if nextNode != nil {
		if err := nextNode.Validate(); err != nil {
			return err
		}
		if prevNode.thisEpoch < n.thisEpoch && n.thisEpoch < nextNode.thisEpoch {
			// In this setting, we want to add a new node in between prevNode and nextNode
			//
			// Update prevNode;
			// must point forward to n
			prevNode.nextEpoch = n.thisEpoch
			// Update epochs for n;
			// must point backward to prevNode and forward to nextNode
			n.prevEpoch = prevNode.thisEpoch
			n.nextEpoch = nextNode.thisEpoch
			// Update  nextNode;
			// must point backward to n
			nextNode.prevEpoch = n.thisEpoch
			return nil
		}
	} else {
		if prevNode.thisEpoch < n.thisEpoch && prevNode.IsTail() {
			// n is the new tail
			// Update prevNode.nextEpoch
			prevNode.nextEpoch = n.thisEpoch
			// Update epochs for n;
			// must point backward to prevNode and forward to zero
			n.prevEpoch = prevNode.thisEpoch
			n.nextEpoch = 0
			return nil
		}
	}

	return ErrInvalid
}

// IsHead returns true if Node is begging of linked list;
// in this case, n.prevEpoch == n.thisEpoch
func (n *Node) IsHead() bool {
	return n.prevEpoch == 0
}

// IsTail returns true if Node is end of linked list;
// in this case, n.nextEpoch == n.thisEpoch
func (n *Node) IsTail() bool {
	return n.nextEpoch == 0
}
