package dynamics

import (
	"bytes"
	"testing"
)

func TestNodeMarshal(t *testing.T) {
	node := &Node{}
	_, err := node.Marshal()
	if err == nil {
		t.Fatal("Should have raied error (1)")
	}

	epoch := uint32(1)
	rs := &RawStorage{}
	rs.standardParameters()
	node, _, err = CreateLinkedList(epoch, rs)
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
	rsBytes, err := node.rawStorage.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	rs2Bytes, err := node2.rawStorage.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rsBytes, rs2Bytes) {
		t.Fatal("invalid RawStroage")
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
	n := &Node{}
	_, err := n.Copy()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	n.prevEpoch = 1
	n.thisEpoch = 1
	n.nextEpoch = 1
	n.rawStorage = &RawStorage{}
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
	wNode := &wNode{}
	if wNode.node.IsValid() {
		t.Fatal("Node should not be valid (0)")
	}

	node := &Node{}
	if node.IsValid() {
		t.Fatal("Node should not be valid (1)")
	}

	node.prevEpoch = 3
	node.thisEpoch = 2
	node.nextEpoch = 3
	if node.IsValid() {
		t.Fatal("Node should not be valid (2)")
	}

	node.prevEpoch = 1
	node.thisEpoch = 3
	node.nextEpoch = 2
	if node.IsValid() {
		t.Fatal("Node should not be valid (3)")
	}

	node.prevEpoch = 1
	node.thisEpoch = 2
	node.nextEpoch = 3
	if node.IsValid() {
		t.Fatal("Node should not be valid (4)")
	}

	node.rawStorage = &RawStorage{}
	if !node.IsValid() {
		t.Fatal("Node should be valid")
	}
}

func TestNodeIsPreValid(t *testing.T) {
	wNode := &wNode{}
	if wNode.node.IsPreValid() {
		t.Fatal("Node should not be prevalid (0)")
	}

	node := &Node{}
	if node.IsPreValid() {
		t.Fatal("Node should not be prevalid (1)")
	}

	node.prevEpoch = 0
	node.thisEpoch = 0
	node.nextEpoch = 0
	if node.IsPreValid() {
		t.Fatal("Node should not be prevalid (2)")
	}

	node.prevEpoch = 1
	node.thisEpoch = 1
	node.nextEpoch = 0
	if node.IsPreValid() {
		t.Fatal("Node should not be prevalid (3)")
	}

	node.prevEpoch = 0
	node.thisEpoch = 1
	node.nextEpoch = 1
	if node.IsPreValid() {
		t.Fatal("Node should not be prevalid (4)")
	}

	node.prevEpoch = 0
	node.thisEpoch = 1
	node.nextEpoch = 0
	if node.IsPreValid() {
		t.Fatal("Node should not be prevalid (5)")
	}

	node.rawStorage = &RawStorage{}
	if !node.IsPreValid() {
		t.Fatal("Node should be prevalid")
	}
}

func TestNodeIsHead(t *testing.T) {
	node := &Node{}
	if node.IsHead() {
		t.Fatal("Node invalid; should be false")
	}

	node.prevEpoch = 1
	node.thisEpoch = 1
	node.nextEpoch = 1
	node.rawStorage = &RawStorage{}
	if !node.IsHead() {
		t.Fatal("Should be Head")
	}

	node.nextEpoch = 2
	if node.IsHead() {
		t.Fatal("Should not be Head")
	}
}

func TestNodeIsTail(t *testing.T) {
	node := &Node{}
	if node.IsTail() {
		t.Fatal("Node invalid; should be false")
	}

	node.prevEpoch = 1
	node.thisEpoch = 1
	node.nextEpoch = 1
	node.rawStorage = &RawStorage{}
	if !node.IsTail() {
		t.Fatal("Should be Tail")
	}

	node.thisEpoch = 2
	node.nextEpoch = 2
	if node.IsTail() {
		t.Fatal("Should not be Tail")
	}
}

// SetNode with prevNode at Head
func TestNodeSetEpochsGood1(t *testing.T) {
	rs := &RawStorage{}
	rs.standardParameters()
	nodeEpoch := uint32(25519)
	node := &Node{
		prevEpoch:  0,
		thisEpoch:  nodeEpoch,
		nextEpoch:  0,
		rawStorage: rs,
	}
	if !node.IsPreValid() {
		t.Fatal("node should be preValid")
	}
	rsNew, err := rs.Copy()
	if err != nil {
		t.Fatal(err)
	}
	rsNew.MaxBytes = 1234567890
	first := uint32(1)
	last := uint32(257)
	prevEpoch := last
	prevNode := &Node{
		prevEpoch:  first,
		thisEpoch:  prevEpoch,
		nextEpoch:  last,
		rawStorage: rsNew,
	}
	if !prevNode.IsValid() {
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
	if node.nextEpoch != nodeEpoch {
		t.Fatal("node.nextEpoch is incorrect; it does not point to self")
	}
}

// SetNode in between prevNode and nextNode
func TestNodeSetEpochsGood2(t *testing.T) {
	rs := &RawStorage{}
	rs.standardParameters()
	nodeEpoch := uint32(25519)
	node := &Node{
		prevEpoch:  0,
		thisEpoch:  nodeEpoch,
		nextEpoch:  0,
		rawStorage: rs,
	}
	if !node.IsPreValid() {
		t.Fatal("node should be preValid")
	}

	rsNew, err := rs.Copy()
	if err != nil {
		t.Fatal(err)
	}
	rsNew.MaxBytes = 1234567890

	first := uint32(1)
	last := uint32(1234567890)
	prevNode := &Node{
		prevEpoch:  first,
		thisEpoch:  first,
		nextEpoch:  last,
		rawStorage: rsNew,
	}
	if !prevNode.IsValid() {
		t.Fatal("prevNode should be Valid")
	}
	if node.thisEpoch < prevNode.thisEpoch {
		t.Fatal("Should have node.thisEpoch < nextNode.thisEpoch")
	}

	nextNode := &Node{
		prevEpoch:  first,
		thisEpoch:  last,
		nextEpoch:  last,
		rawStorage: rsNew,
	}
	if !nextNode.IsValid() {
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
	if prevNode.prevEpoch != first {
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
	if nextNode.nextEpoch != last {
		t.Fatal("nextNode.nextEpoch is incorrect")
	}
}

// We should raise an error when having node not PreValid
func TestNodeSetEpochsBad1(t *testing.T) {
	node := &Node{}
	err := node.SetEpochs(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// We should raise error for prevNode being invalid
func TestNodeSetEpochsBad2(t *testing.T) {
	rs := &RawStorage{}
	rs.standardParameters()
	node := &Node{
		prevEpoch:  0,
		thisEpoch:  25519,
		nextEpoch:  0,
		rawStorage: rs,
	}
	if !node.IsPreValid() {
		t.Fatal("node should be preValid")
	}
	prevNode := &Node{}
	if prevNode.IsValid() {
		t.Fatal("prevNode should not be valid")
	}
	err := node.SetEpochs(prevNode, nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// We should raise error for nextNode not nil
func TestNodeSetEpochsBad3(t *testing.T) {
	rs := &RawStorage{}
	rs.standardParameters()
	node := &Node{
		prevEpoch:  0,
		thisEpoch:  25519,
		nextEpoch:  0,
		rawStorage: rs,
	}
	if !node.IsPreValid() {
		t.Fatal("node should be preValid")
	}
	prevNode := &Node{
		prevEpoch:  1,
		thisEpoch:  257,
		nextEpoch:  123456789,
		rawStorage: rs,
	}
	if !prevNode.IsValid() {
		t.Fatal("prevNode should be Valid")
	}
	nextNode := &Node{}
	err := node.SetEpochs(prevNode, nextNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// We should raise error for prevNode.thisEpoch >= node.thisEpoch
func TestNodeSetEpochsBad4(t *testing.T) {
	rs := &RawStorage{}
	rs.standardParameters()
	node := &Node{
		prevEpoch:  0,
		thisEpoch:  25519,
		nextEpoch:  0,
		rawStorage: rs,
	}
	if !node.IsPreValid() {
		t.Fatal("node should be preValid")
	}
	prevNode := &Node{
		prevEpoch:  1,
		thisEpoch:  25519,
		nextEpoch:  123456789,
		rawStorage: rs,
	}
	if !prevNode.IsValid() {
		t.Fatal("prevNode should be Valid")
	}
	if prevNode.thisEpoch < node.thisEpoch {
		t.Fatal("Should have prevNode.thisEpoch >= node.thisEpoch")
	}
	err := node.SetEpochs(prevNode, nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// We should raise error for prevNode not nil
func TestNodeSetEpochsBad5(t *testing.T) {
	rs := &RawStorage{}
	rs.standardParameters()
	node := &Node{
		prevEpoch:  0,
		thisEpoch:  25519,
		nextEpoch:  0,
		rawStorage: rs,
	}
	if !node.IsPreValid() {
		t.Fatal("node should be preValid")
	}
	prevNode := &Node{}
	if prevNode.IsValid() {
		t.Fatal("prevNode should not be valid")
	}
	nextNode := &Node{}
	err := node.SetEpochs(prevNode, nextNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// We should raise error for prevNode not nil
func TestNodeSetEpochsBad6(t *testing.T) {
	rs := &RawStorage{}
	rs.standardParameters()
	node := &Node{
		prevEpoch:  0,
		thisEpoch:  25519,
		nextEpoch:  0,
		rawStorage: rs,
	}
	nextNode := &Node{}
	if nextNode.IsValid() {
		t.Fatal("nextNode should not be valid")
	}
	err := node.SetEpochs(nil, nextNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// We should raise error for prevNode not nil
func TestNodeSetEpochsBad7(t *testing.T) {
	rs := &RawStorage{}
	rs.standardParameters()
	node := &Node{
		prevEpoch:  0,
		thisEpoch:  25519,
		nextEpoch:  0,
		rawStorage: rs,
	}
	rsNew, err := rs.Copy()
	if err != nil {
		t.Fatal(err)
	}
	rsNew.MaxBytes = 1234567890
	nextNode := &Node{
		prevEpoch:  1,
		thisEpoch:  257,
		nextEpoch:  123456789,
		rawStorage: rsNew,
	}
	if !nextNode.IsValid() {
		t.Fatal("nextNode should not be valid")
	}
	if node.thisEpoch < nextNode.thisEpoch {
		t.Fatal("We should not have node.thisEpoch >= nextNode.thisEpoch to raise error")
	}
	err = node.SetEpochs(nil, nextNode)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// We should raise error for prevNode not nil
func TestNodeSetEpochsBad8(t *testing.T) {
	rs := &RawStorage{}
	rs.standardParameters()
	node := &Node{
		prevEpoch:  0,
		thisEpoch:  25519,
		nextEpoch:  0,
		rawStorage: rs,
	}
	rsNew, err := rs.Copy()
	if err != nil {
		t.Fatal(err)
	}
	rsNew.MaxBytes = 1234567890
	nextNode := &Node{
		prevEpoch:  1,
		thisEpoch:  123456,
		nextEpoch:  123456789,
		rawStorage: rsNew,
	}
	if !nextNode.IsValid() {
		t.Fatal("nextNode should not be valid")
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
