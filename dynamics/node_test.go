package dynamics

import (
	"bytes"
	"errors"
	"testing"

	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/utils"
	"github.com/stretchr/testify/assert"
)

func TestNodeMakeKeys(t *testing.T) {
	t.Parallel()
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
}

func TestNodeMarshal(t *testing.T) {
	t.Parallel()
	node := &Node{}
	_, err := node.Marshal()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	epoch := uint32(1)
	_, dv := GetStandardDynamicValue()
	node, _, err = CreateLinkedList(epoch, dv)
	if err != nil {
		t.Fatal(err)
	}

	nodeBytes, err := node.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	node2 := &Node{}
	err = node2.Unmarshal(nodeBytes)
	if err != nil {
		t.Fatal(err)
	}
	if node.thisEpoch != node2.thisEpoch {
		t.Fatal("invalid thisEpoch")
	}
	if node.prevEpoch != node2.prevEpoch {
		t.Fatal("invalid prevEpoch")
	}
	if node.nextEpoch != node2.nextEpoch {
		t.Fatal("invalid nextEpoch")
	}
	dvBytes, err := node.dynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	dv2Bytes, err := node2.dynamicValues.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dvBytes, dv2Bytes) {
		t.Fatal("invalid Dynamic Value")
	}

	v := []byte{}
	err = node.Unmarshal(v)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	v = make([]byte, 12)
	err = node.Unmarshal(v)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	v = make([]byte, 13)
	err = node.Unmarshal(v)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
}

func TestNodeCopy(t *testing.T) {
	t.Parallel()
	n := &Node{}
	_, err := n.Copy()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	n.prevEpoch = 1
	n.thisEpoch = 1
	n.nextEpoch = 1
	n.dynamicValues = &DynamicValues{}
	n2, err := n.Copy()
	if err != nil {
		t.Fatal(err)
	}

	nBytes, err := n.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	n2Bytes, err := n2.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(nBytes, n2Bytes) {
		t.Fatal("nodes do not match")
	}
}

type wNode struct {
	node *Node
}

func TestNodeIsValid(t *testing.T) {
	t.Parallel()
	wNode := &wNode{}
	err := wNode.node.Validate()
	if !errors.Is(err, ErrNodeValueNilPointer) {
		t.Fatal("Node should not be valid (0)")
	}

	node := &Node{}
	err = node.Validate()
	if !errors.Is(err, ErrZeroEpoch) {
		t.Fatal("Node should not be valid (1)")
	}

	node.prevEpoch = 3
	node.thisEpoch = 2
	node.nextEpoch = 3
	err = node.Validate()
	expectedErr := &ErrInvalidNode{}
	if !errors.As(err, &expectedErr) {
		t.Fatal("Node should not be valid (2)")
	}

	node.prevEpoch = 1
	node.thisEpoch = 3
	node.nextEpoch = 2
	err = node.Validate()
	if !errors.As(err, &expectedErr) {
		t.Fatalf("Node should not be valid (3):%v", err)
	}

	node.prevEpoch = 1
	node.thisEpoch = 2
	node.nextEpoch = 3
	err = node.Validate()
	if !errors.Is(err, ErrDynamicValueNilPointer) {
		t.Fatal("Node should not be valid (4)")
	}

	node.dynamicValues = &DynamicValues{}
	err = node.Validate()
	if !errors.Is(err, ErrValueIsEmpty) {
		t.Fatalf("Node should not be valid (5): %v", err)
	}
	// valid node

	node.prevEpoch = 1
	node.thisEpoch = 2
	node.nextEpoch = 3
	_, node.dynamicValues = GetDynamicValueWithFees()
	err = node.Validate()
	assert.Nil(t, err)
}

func TestNodeIsHead(t *testing.T) {
	t.Parallel()
	node := &Node{}
	node.prevEpoch = 0
	node.nextEpoch = 0
	node.thisEpoch = 1
	_, node.dynamicValues = GetDynamicValueWithFees()
	if !node.IsHead() {
		t.Fatal("Should be Head")
	}

	node.prevEpoch = 2
	if node.IsHead() {
		t.Fatal("Should not be Head")
	}
}

func TestNodeIsTail(t *testing.T) {
	t.Parallel()
	node := &Node{}
	node.prevEpoch = 0
	node.nextEpoch = 0
	node.thisEpoch = 1
	_, node.dynamicValues = GetDynamicValueWithFees()
	if !node.IsTail() {
		t.Fatal("Should be Tail")
	}

	node.thisEpoch = 2
	node.nextEpoch = 3
	if node.IsTail() {
		t.Fatal("Should not be Tail")
	}
}

// SetNode with prevNode at Head
func TestNodeSetEpochsGood1(t *testing.T) {
	t.Parallel()
	_, dv := GetStandardDynamicValue()
	nodeEpoch := uint32(25519)
	node := &Node{
		prevEpoch:     0,
		thisEpoch:     nodeEpoch,
		nextEpoch:     0,
		dynamicValues: dv,
	}
	dvsNew, err := dv.Copy()
	if err != nil {
		t.Fatal(err)
	}
	dvsNew.MaxBlockSize = 1234567890
	first := uint32(1)
	last := uint32(257)
	prevEpoch := last
	prevNode := &Node{
		prevEpoch:     first,
		thisEpoch:     last,
		nextEpoch:     0,
		dynamicValues: dvsNew,
	}
	err = prevNode.Validate()
	if err != nil {
		t.Fatal("prevNode should be Valid")
	}
	if prevNode.thisEpoch >= node.thisEpoch {
		t.Fatal("Should have prevNode.thisEpoch < node.thisEpoch")
	}
	err = node.SetEpochs(prevNode, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Now need to confirm all epochs are good.
	if prevNode.prevEpoch != first {
		t.Fatal("prevNode.prevEpoch is incorrect")
	}
	if prevNode.thisEpoch != prevEpoch {
		t.Fatal("prevNode.thisEpoch is incorrect")
	}
	if prevNode.nextEpoch != nodeEpoch {
		t.Fatal("prevNode.nextEpoch is incorrect; it does not point to new nodeEpoch")
	}
	if node.prevEpoch != prevEpoch {
		t.Fatal("prevNode.prevEpoch is incorrect; it does not equal prevEpoch")
	}
	if node.thisEpoch != nodeEpoch {
		t.Fatal("node.thisEpoch is incorrect")
	}
	if node.nextEpoch != 0 {
		t.Fatal("node.nextEpoch is incorrect; it does not point to zero")
	}
}

// SetNode in between prevNode and nextNode
func TestNodeSetEpochsGood2(t *testing.T) {
	t.Parallel()
	_, rs := GetStandardDynamicValue()
	nodeEpoch := uint32(25519)
	node := &Node{
		prevEpoch:     0,
		thisEpoch:     nodeEpoch,
		nextEpoch:     0,
		dynamicValues: rs,
	}

	rsNew, err := rs.Copy()
	if err != nil {
		t.Fatal(err)
	}
	rsNew.MaxBlockSize = 1234567890

	first := uint32(1)
	last := uint32(1234567890)
	prevNode := &Node{
		prevEpoch:     0,
		thisEpoch:     first,
		nextEpoch:     last,
		dynamicValues: rsNew,
	}
	err = prevNode.Validate()
	if err != nil {
		t.Fatal("prevNode should be Valid")
	}
	if node.thisEpoch < prevNode.thisEpoch {
		t.Fatal("Should have node.thisEpoch < nextNode.thisEpoch")
	}

	nextNode := &Node{
		prevEpoch:     first,
		thisEpoch:     last,
		nextEpoch:     0,
		dynamicValues: rsNew,
	}
	err = nextNode.Validate()
	if err != nil {
		t.Fatal("nextNode should be Valid")
	}
	if node.thisEpoch >= nextNode.thisEpoch {
		t.Fatal("Should have node.thisEpoch < nextNode.thisEpoch")
	}

	err = node.SetEpochs(prevNode, nextNode)
	if err != nil {
		t.Fatal(err)
	}

	// Now need to confirm all epochs are good.
	if prevNode.prevEpoch != 0 {
		t.Fatal("nextNode.prevEpoch is incorrect")
	}
	if prevNode.thisEpoch != first {
		t.Fatal("nextNode.thisEpoch is incorrect")
	}
	if prevNode.nextEpoch != nodeEpoch {
		t.Fatal("nextNode.nextEpoch is incorrect")
	}

	if node.prevEpoch != first {
		t.Fatal("node.prevEpoch is incorrect")
	}
	if node.thisEpoch != nodeEpoch {
		t.Fatal("node.thisEpoch is incorrect")
	}
	if node.nextEpoch != last {
		t.Fatal("node.nextEpoch is incorrect")
	}

	if nextNode.prevEpoch != nodeEpoch {
		t.Fatal("nextNode.prevEpoch is incorrect")
	}
	if nextNode.thisEpoch != last {
		t.Fatal("nextNode.thisEpoch is incorrect")
	}
	if nextNode.nextEpoch != 0 {
		t.Fatal("nextNode.nextEpoch is incorrect")
	}
}

// We should raise an error when having node not Valid
func TestNodeSetEpochsBad1(t *testing.T) {
	t.Parallel()
	node := &Node{}
	err := node.SetEpochs(nil, nil)
	if !errors.Is(err, ErrNodeValueNilPointer) {
		t.Fatalf("Should have raised error: got %v", err)
	}
}

// We should raise error for prevNode being invalid
func TestNodeSetEpochsBad2(t *testing.T) {
	t.Parallel()
	_, rs := GetStandardDynamicValue()
	node := &Node{
		prevEpoch:     0,
		thisEpoch:     25519,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	// prevEpoch is greater than nextEpoch
	prevNode := &Node{
		prevEpoch:     2,
		thisEpoch:     50,
		nextEpoch:     1,
		dynamicValues: rs,
	}
	err := prevNode.Validate()
	expectedErr := &ErrInvalidNode{}
	if !errors.As(err, &expectedErr) {
		t.Fatalf("prev node should not be valid: %v", err)
	}
	// trying invalid prev node
	err = node.SetEpochs(prevNode, nil)
	if !errors.As(err, &expectedErr) {
		t.Fatalf("prev node should not be valid(2) %v", err)
	}
	// trying empty node
	err = node.SetEpochs(&Node{}, nil)
	if !errors.Is(err, ErrZeroEpoch) {
		t.Fatalf("prev node should not be valid(3) %v", err)
	}
}

// We should raise error for nextNode not nil
func TestNodeSetEpochsBad3(t *testing.T) {
	t.Parallel()
	_, dv := GetStandardDynamicValue()
	node := &Node{
		prevEpoch:     0,
		thisEpoch:     25519,
		nextEpoch:     0,
		dynamicValues: dv,
	}
	prevNode := &Node{
		prevEpoch:     1,
		thisEpoch:     257,
		nextEpoch:     123456789,
		dynamicValues: dv,
	}
	err := prevNode.Validate()
	if err != nil {
		t.Fatalf("prev node should be valid: %v", err)
	}
	nextNode := &Node{}
	err = node.SetEpochs(prevNode, nextNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// We should raise error for prevNode.thisEpoch >= node.thisEpoch
func TestNodeSetEpochsBad4(t *testing.T) {
	t.Parallel()
	_, rs := GetStandardDynamicValue()
	node := &Node{
		prevEpoch:     0,
		thisEpoch:     25519,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	prevNode := &Node{
		prevEpoch:     1,
		thisEpoch:     25519,
		nextEpoch:     123456789,
		dynamicValues: rs,
	}
	err := prevNode.Validate()
	if err != nil {
		t.Fatal("prevNode should be Valid")
	}
	// should not be able to overwrite a node
	err = node.SetEpochs(prevNode, nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	// node is not tail
	prevNode2 := &Node{
		prevEpoch:     1,
		thisEpoch:     25518,
		nextEpoch:     123456789,
		dynamicValues: rs,
	}
	// should not be able to overwrite a node
	err = node.SetEpochs(prevNode2, nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	prevNode3 := &Node{
		prevEpoch:     1,
		thisEpoch:     25520,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	// should not be able to overwrite a node
	err = node.SetEpochs(prevNode3, nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// We should raise error for prevNode not nil
func TestNodeSetEpochsBad5(t *testing.T) {
	t.Parallel()
	_, rs := GetStandardDynamicValue()
	node := &Node{
		prevEpoch:     0,
		thisEpoch:     25519,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	prevNode := &Node{}
	err := prevNode.Validate()
	if err == nil {
		t.Fatal("prevNode should not be valid")
	}
	nextNode := &Node{}
	err = node.SetEpochs(prevNode, nextNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// We should raise error for prevNode not nil
func TestNodeSetEpochsBad6(t *testing.T) {
	t.Parallel()
	_, rs := GetStandardDynamicValue()
	node := &Node{
		prevEpoch:     0,
		thisEpoch:     25519,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	nextNode := &Node{}
	err := nextNode.Validate()
	if err == nil {
		t.Fatal("nextNode should not be valid")
	}
	err = node.SetEpochs(nil, nextNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// We should raise error for prevNode not nil
func TestNodeSetEpochsBad7(t *testing.T) {
	t.Parallel()
	_, rs := GetStandardDynamicValue()
	node := &Node{
		prevEpoch:     0,
		thisEpoch:     25519,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	rsNew, err := rs.Copy()
	if err != nil {
		t.Fatal(err)
	}
	rsNew.MaxBlockSize = 1234567890
	prevNode := &Node{
		prevEpoch:     1,
		thisEpoch:     257,
		nextEpoch:     25519,
		dynamicValues: rsNew,
	}
	nextNode := &Node{
		prevEpoch:     257,
		thisEpoch:     25518,
		nextEpoch:     0,
		dynamicValues: rsNew,
	}
	if node.thisEpoch < nextNode.thisEpoch {
		t.Fatal("We should not have node.thisEpoch >= nextNode.thisEpoch to raise error")
	}
	err = node.SetEpochs(prevNode, nextNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestNodeSetEpochsBad8(t *testing.T) {
	t.Parallel()
	_, rs := GetStandardDynamicValue()
	node := &Node{
		prevEpoch:     0,
		thisEpoch:     25519,
		nextEpoch:     0,
		dynamicValues: rs,
	}
	rsNew, err := rs.Copy()
	if err != nil {
		t.Fatal(err)
	}
	rsNew.MaxBlockSize = 1234567890
	nextNode := &Node{
		prevEpoch:     1,
		thisEpoch:     123456,
		nextEpoch:     123456789,
		dynamicValues: rsNew,
	}
	err = nextNode.Validate()
	if err != nil {
		t.Fatal("nextNode should be valid")
	}
	if node.thisEpoch >= nextNode.thisEpoch {
		t.Fatal("We should have node.thisEpoch >= nextNode.thisEpoch")
	}
	if nextNode.IsTail() {
		t.Fatal("We should not have nextNode is tail to raise error")
	}
	err = node.SetEpochs(nil, nextNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}
