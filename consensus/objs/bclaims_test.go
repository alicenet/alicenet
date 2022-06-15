package objs

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/utils"
)

func generateChain(length int) ([]*BClaims, [][][]byte, error) {
	chain := []*BClaims{}
	txHashes := [][][]byte{}
	txhash := crypto.Hasher([]byte(strconv.Itoa(1)))
	txHshLst := [][]byte{txhash}
	txRoot, err := MakeTxRoot(txHshLst)
	if err != nil {
		return nil, nil, err
	}
	txHashes = append(txHashes, txHshLst)
	bclaims := &BClaims{
		ChainID:    1,
		Height:     1,
		TxCount:    1,
		PrevBlock:  crypto.Hasher([]byte("foo")),
		TxRoot:     txRoot,
		StateRoot:  crypto.Hasher([]byte("")),
		HeaderRoot: crypto.Hasher([]byte("")),
	}
	chain = append(chain, bclaims)
	for i := 1; i < length; i++ {
		bhsh, err := chain[i-1].BlockHash()
		if err != nil {
			return nil, nil, err
		}
		txhash := crypto.Hasher([]byte(strconv.Itoa(i)))
		txHshLst := [][]byte{txhash}
		txRoot, err := MakeTxRoot(txHshLst)
		if err != nil {
			return nil, nil, err
		}
		txHashes = append(txHashes, txHshLst)
		bclaims := &BClaims{
			ChainID:    1,
			Height:     uint32(len(chain) + 1),
			TxCount:    1,
			PrevBlock:  bhsh,
			TxRoot:     txRoot,
			StateRoot:  chain[i-1].StateRoot,
			HeaderRoot: chain[i-1].HeaderRoot,
		}
		chain = append(chain, bclaims)
	}
	return chain, txHashes, nil
}

func TestBClaims(t *testing.T) {
	bclaimsList, _, err := generateChain(2)
	if err != nil {
		t.Fatal(err)
	}
	bclaims := bclaimsList[0]
	bclaimsBytes, err := bclaims.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	bclaims2 := &BClaims{}
	err = bclaims2.UnmarshalBinary(bclaimsBytes)
	if err != nil {
		t.Fatal(err)
	}
	bclaimsEqual(t, bclaims, bclaims2)
}

func TestBClaimsBad(t *testing.T) {
	bclaimsList, _, err := generateChain(2)
	if err != nil {
		t.Fatal(err)
	}
	bclaims := bclaimsList[0]
	_, err = bclaims.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	height := bclaims.Height
	bclaims.Height = 0
	_, err = bclaims.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	bclaims.Height = height

	cid := bclaims.ChainID
	bclaims.ChainID = 0
	_, err = bclaims.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	bclaims.ChainID = cid

	bclaims.Height = 0
	_, err = bclaims.BlockHash()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func bclaimsEqual(t *testing.T, bclaims, bclaims2 *BClaims) {
	if bclaims.ChainID != bclaims2.ChainID {
		t.Fatal("fail")
	}
	if bclaims.Height != bclaims2.Height {
		t.Fatal("fail")
	}
	if bclaims.TxCount != bclaims2.TxCount {
		t.Fatal("fail")
	}
	if !bytes.Equal(bclaims.PrevBlock, bclaims2.PrevBlock) {
		t.Fatal("fail")
	}
	if !bytes.Equal(bclaims.TxRoot, bclaims2.TxRoot) {
		t.Fatal("fail")
	}
	if !bytes.Equal(bclaims.StateRoot, bclaims2.StateRoot) {
		t.Fatal("fail")
	}
	if !bytes.Equal(bclaims.HeaderRoot, bclaims2.HeaderRoot) {
		t.Fatal("fail")
	}
}

func TestBClaimsMarshal(t *testing.T) {
	bh := &BlockHeader{}
	_, err := bh.BClaims.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	bclaims := &BClaims{}
	_, err = bclaims.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	bclaims = &BClaims{
		ChainID: 1,
	}
	_, err = bclaims.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	bclaims = &BClaims{
		ChainID: 1,
		Height:  1,
		TxCount: 0,
	}
	data, err := bclaims.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	bc2 := &BClaims{}
	err = bc2.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	bclaims = &BClaims{
		ChainID:   1,
		Height:    1,
		TxCount:   0,
		PrevBlock: crypto.Hasher([]byte("Genesis")),
	}
	data, err = bclaims.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = bc2.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	bclaims = &BClaims{
		ChainID:   1,
		Height:    1,
		TxCount:   0,
		PrevBlock: crypto.Hasher([]byte("Genesis")),
		TxRoot:    crypto.Hasher([]byte("")),
	}
	data, err = bclaims.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = bc2.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}

	bclaims = &BClaims{
		ChainID:   1,
		Height:    1,
		TxCount:   0,
		PrevBlock: crypto.Hasher([]byte("Genesis")),
		TxRoot:    crypto.Hasher([]byte("")),
		StateRoot: crypto.Hasher([]byte("")),
	}
	data, err = bclaims.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = bc2.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error (6)")
	}

	bclaims = &BClaims{
		ChainID:    223,
		Height:     225,
		TxCount:    0,
		PrevBlock:  crypto.Hasher([]byte("Genesis")),
		TxRoot:     crypto.Hasher([]byte("")),
		StateRoot:  crypto.Hasher([]byte("")),
		HeaderRoot: crypto.Hasher([]byte("")),
	}
	data, err = bclaims.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = bc2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}

	// Take valid marshaled state and raise an error for invalid ChainID
	dataBad1 := utils.CopySlice(data)
	dataBad1[8] = 0
	bc3 := &BClaims{}
	err = bc3.UnmarshalBinary(dataBad1)
	if err == nil {
		t.Fatal("Should have raised error (7)")
	}

	// Take valid marshaled state and raise an error for invalid Height
	dataBad2 := utils.CopySlice(data)
	dataBad2[12] = 0
	bc4 := &BClaims{}
	err = bc4.UnmarshalBinary(dataBad2)
	if err == nil {
		t.Fatal("Should have raised error (8)")
	}
}
