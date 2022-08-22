package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/stretchr/testify/assert"
)

func TestMakeTxRoot(t *testing.T) {
	hashNull := crypto.Hasher([]byte{})
	retHash, err := MakeTxRoot(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(hashNull, retHash) {
		t.Fatal("invalid TxRoot")
	}

	hashesBad := [][]byte{}
	hashesBad = append(hashesBad, []byte{0})
	_, err = MakeTxRoot(hashesBad)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestGetProposerIdx(t *testing.T) {
	height := uint32(1)
	round := uint32(1)
	numv := 4
	propIdx := uint8(1)
	retIdx := GetProposerIdx(numv, height, round)
	if retIdx != propIdx {
		t.Fatal("Invalid proposer index (1)")
	}
	height = 5
	retIdx = GetProposerIdx(numv, height, round)
	if retIdx != propIdx {
		t.Fatal("Invalid proposer index (2)")
	}
	round = 2
	propIdx = 2
	retIdx = GetProposerIdx(numv, height, round)
	if retIdx != propIdx {
		t.Fatal("Invalid proposer index (3)")
	}
}

func TestSplitBlob(t *testing.T) {
	data := make([]byte, 10)
	blen := 3
	_, err := SplitBlob(data, blen)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	data = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	blen = 2
	trueBuf := [][]byte{}
	trueBuf = append(trueBuf, []byte{1, 2})
	trueBuf = append(trueBuf, []byte{3, 4})
	trueBuf = append(trueBuf, []byte{5, 6})
	trueBuf = append(trueBuf, []byte{7, 8})
	trueBuf = append(trueBuf, []byte{9, 10})
	buf, err := SplitBlob(data, blen)
	if err != nil {
		t.Fatal(err)
	}
	if len(buf) != len(trueBuf) {
		t.Fatal("invalid split")
	}
	for k := 0; k < len(trueBuf); k++ {
		if !bytes.Equal(buf[k], trueBuf[k]) {
			t.Fatalf("invalid buf split")
		}
	}
}

func TestExtractHR(t *testing.T) {
	trueCid := uint32(42)
	trueHeight := uint32(137)
	trueRound := uint32(1)
	bclaims := &BClaims{
		ChainID:    trueCid,
		Height:     trueHeight,
		TxCount:    0,
		PrevBlock:  crypto.Hasher([]byte("Genesis")),
		TxRoot:     crypto.Hasher([]byte("")),
		StateRoot:  crypto.Hasher([]byte("")),
		HeaderRoot: crypto.Hasher([]byte("")),
	}
	height, round := ExtractHR(bclaims)
	if height != trueHeight {
		t.Fatal("Invalid height")
	}
	if round != trueRound {
		t.Fatal("Invalid round")
	}
}

func TestExtractHRBad(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Should panic")
		}
	}()
	ExtractHR(nil)
}

func TestExtractHCID(t *testing.T) {
	trueCid := uint32(42)
	trueHeight := uint32(137)
	bclaims := &BClaims{
		ChainID:    trueCid,
		Height:     trueHeight,
		TxCount:    0,
		PrevBlock:  crypto.Hasher([]byte("Genesis")),
		TxRoot:     crypto.Hasher([]byte("")),
		StateRoot:  crypto.Hasher([]byte("")),
		HeaderRoot: crypto.Hasher([]byte("")),
	}
	height, cid := ExtractHCID(bclaims)
	if height != trueHeight {
		t.Fatal("Invalid height")
	}
	if cid != trueCid {
		t.Fatal("Invalid ChainID")
	}
}

func TestExtractHCIDBad(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Should panic")
		}
	}()
	ExtractHCID(nil)
}

func TestExtractRCertAny(t *testing.T) {
	cid := uint32(1)
	height := uint32(2)
	prevBlock := crypto.Hasher([]byte("Genesis"))
	txRoot := crypto.Hasher([]byte("TxRoot"))
	stateRoot := crypto.Hasher([]byte("StateRoot"))
	headerRoot := crypto.Hasher([]byte("HeaderRoot"))
	bc := &BClaims{
		ChainID:    cid,
		Height:     height,
		TxCount:    0,
		PrevBlock:  prevBlock,
		TxRoot:     txRoot,
		StateRoot:  stateRoot,
		HeaderRoot: headerRoot,
	}
	sigGroup := make([]byte, constants.CurveBN256EthSigLen)
	bh := &BlockHeader{
		BClaims:  bc,
		SigGroup: sigGroup,
	}
	bhsh, err := bh.BlockHash()
	if err != nil {
		t.Fatal(err)
	}

	rc, err := ExtractRCertAny(bh)
	if err != nil {
		t.Fatal(err)
	}
	if rc.RClaims.ChainID != cid {
		t.Fatal("invalid ChainID")
	}
	if rc.RClaims.Height != height+1 {
		t.Fatal("invalid Height")
	}
	if rc.RClaims.Round != 1 {
		t.Fatal("invalid Round")
	}
	if !bytes.Equal(bhsh, rc.RClaims.PrevBlock) {
		t.Fatal("invalid PrevBlock")
	}
}

func TestExtractRCertAnyBad(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Should panic")
		}
	}()
	rcert, err := ExtractRCertAny(nil)
	assert.Nil(t, rcert)
	assert.NotNil(t, err)
}

func TestExtractRCertBad(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Should panic")
		}
	}()
	ExtractRCert(nil)
}

func TestExtractBClaimsBad(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Should panic")
		}
	}()
	ExtractBClaims(nil)
}

func TestRelateHR(t *testing.T) {
	cid := uint32(66)
	height1 := uint32(2)
	round1 := uint32(1)
	prevHash1 := make([]byte, constants.HashLen)
	rc1 := &RClaims{
		ChainID:   cid,
		Height:    height1,
		Round:     round1,
		PrevBlock: prevHash1,
	}

	height2 := uint32(2)
	round2 := uint32(2)
	prevHash2 := make([]byte, constants.HashLen)
	rc2 := &RClaims{
		ChainID:   cid,
		Height:    height2,
		Round:     round2,
		PrevBlock: prevHash2,
	}

	rel := RelateHR(rc1, rc2)
	if rel != -1 {
		t.Fatal("Should return -1")
	}

	rel = RelateHR(rc2, rc1)
	if rel != 1 {
		t.Fatal("Should return 1")
	}

	rel = RelateHR(rc1, rc1)
	if rel != 0 {
		t.Fatal("Should return 0")
	}

	height3 := uint32(3)
	round3 := uint32(1)
	prevHash3 := make([]byte, constants.HashLen)
	rc3 := &RClaims{
		ChainID:   cid,
		Height:    height3,
		Round:     round3,
		PrevBlock: prevHash3,
	}

	height4 := uint32(4)
	round4 := uint32(2)
	prevHash4 := make([]byte, constants.HashLen)
	rc4 := &RClaims{
		ChainID:   cid,
		Height:    height4,
		Round:     round4,
		PrevBlock: prevHash4,
	}

	rel = RelateHR(rc3, rc4)
	if rel != -1 {
		t.Fatal("Should return -1")
	}
	rel = RelateHR(rc4, rc3)
	if rel != 1 {
		t.Fatal("Should return 1")
	}
}

func TestBClaimsEqual(t *testing.T) {
	height1 := uint32(1)
	cid := uint32(1)
	bc1 := &BClaims{
		ChainID:    cid,
		Height:     height1,
		TxCount:    0,
		PrevBlock:  crypto.Hasher([]byte("Genesis")),
		TxRoot:     crypto.Hasher([]byte("")),
		StateRoot:  crypto.Hasher([]byte("")),
		HeaderRoot: crypto.Hasher([]byte("")),
	}
	bh1 := &BlockHeader{
		BClaims: bc1,
	}

	height2 := uint32(2)
	bc2 := &BClaims{
		ChainID:    cid,
		Height:     height2,
		TxCount:    0,
		PrevBlock:  crypto.Hasher([]byte("Genesis")),
		TxRoot:     crypto.Hasher([]byte("")),
		StateRoot:  crypto.Hasher([]byte("")),
		HeaderRoot: crypto.Hasher([]byte("")),
	}
	bh2 := &BlockHeader{
		BClaims: bc2,
	}

	equal, err := BClaimsEqual(bh1, bh2)
	if err != nil {
		t.Fatal(err)
	}
	if equal {
		t.Fatal("Should not be equal")
	}
	equal, err = BClaimsEqual(bh1, bh1)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatal("Should be equal")
	}

	bcBad := &BClaims{}
	bhBad := &BlockHeader{
		BClaims: bcBad,
	}
	_, err = BClaimsEqual(bh1, bhBad)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	_, err = BClaimsEqual(bhBad, bh1)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}
