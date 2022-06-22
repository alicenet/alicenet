package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/utils"
)

func rclaimsEqual(t *testing.T, rclaims, rclaims2 *RClaims) {
	if rclaims.ChainID != rclaims2.ChainID {
		t.Fatal("fail")
	}
	if rclaims.Height != rclaims2.Height {
		t.Fatal("fail")
	}
	if rclaims.Round != rclaims2.Round {
		t.Fatal("fail")
	}
	if !bytes.Equal(rclaims.PrevBlock, rclaims2.PrevBlock) {
		t.Fatal("fail")
	}
}

func TestRClaims(t *testing.T) {
	cid := uint32(66)
	height := uint32(2)
	round := uint32(1)
	prevHash := make([]byte, constants.HashLen)
	rcl := &RClaims{
		ChainID:   cid,
		Height:    height,
		Round:     round,
		PrevBlock: prevHash,
	}
	data, err := rcl.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rcl2 := &RClaims{}
	err = rcl2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	rclaimsEqual(t, rcl, rcl2)
}

func TestRClaimsBad(t *testing.T) {
	cid := uint32(66)
	height := uint32(2)
	round := uint32(1)
	prevHash := make([]byte, constants.HashLen)
	rcl2 := &RClaims{}

	rcl := &RClaims{
		ChainID:   0,
		Height:    height,
		Round:     round,
		PrevBlock: prevHash,
	}
	_, err := rcl.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	rcl = &RClaims{
		ChainID:   cid,
		Height:    0,
		Round:     round,
		PrevBlock: prevHash,
	}
	_, err = rcl.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	rcl = &RClaims{
		ChainID:   cid,
		Height:    height,
		Round:     0,
		PrevBlock: prevHash,
	}
	_, err = rcl.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	prevHashBad := make([]byte, constants.HashLen+1)
	rcl = &RClaims{
		ChainID:   cid,
		Height:    height,
		Round:     round,
		PrevBlock: prevHashBad,
	}
	dataBad2, err := rcl.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = rcl2.UnmarshalBinary(dataBad2)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
}

func TestRClaimsBad2(t *testing.T) {
	cid := uint32(66)
	height := uint32(113)
	round := uint32(1)
	prevHash := make([]byte, constants.HashLen)

	rcl := &RClaims{
		ChainID:   cid,
		Height:    height,
		Round:     round,
		PrevBlock: prevHash,
	}
	data, err := rcl.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	rcl2 := &RClaims{}
	// Raise error for bad ChainID
	dataBad1 := utils.CopySlice(data)
	dataBad1[8] = 0
	err = rcl2.UnmarshalBinary(dataBad1)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	// Raise error for bad Height
	dataBad2 := utils.CopySlice(data)
	dataBad2[12] = 0
	err = rcl2.UnmarshalBinary(dataBad2)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	// Raise error for bad Round
	dataBad3 := utils.CopySlice(data)
	dataBad3[16] = 0
	err = rcl2.UnmarshalBinary(dataBad3)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}
