package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/utils"
)

func TestTXInLinkerGood(t *testing.T) {
	cid := uint32(2)
	consTxIdx := uint32(0)
	consTxHash := make([]byte, constants.HashLen)
	txinp := &TXInPreImage{
		ChainID:        cid,
		ConsumedTxIdx:  consTxIdx,
		ConsumedTxHash: consTxHash,
	}
	txHash := make([]byte, constants.HashLen)
	txinl := &TXInLinker{
		TXInPreImage: txinp,
		TxHash:       txHash,
	}
	txinl2 := &TXInLinker{}
	txinlBytes, err := txinl.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = txinl2.UnmarshalBinary(txinlBytes)
	if err != nil {
		t.Fatal(err)
	}
	txinlEqual(t, txinl, txinl2)
}

func txinlEqual(t *testing.T, txinl1, txinl2 *TXInLinker) {
	txinpi1 := txinl1.TXInPreImage
	txinpi2 := txinl2.TXInPreImage
	txinpiEqual(t, txinpi1, txinpi2)
	if !bytes.Equal(txinl1.TxHash, txinl2.TxHash) {
		t.Fatal("Do not agree on TxHash!")
	}
}

func TestTXInLinkerBad1(t *testing.T) {
	cid := uint32(0) // Invalid ChainID
	consTxIdx := uint32(0)
	consTxHash := make([]byte, constants.HashLen)
	txinp := &TXInPreImage{
		ChainID:        cid,
		ConsumedTxIdx:  consTxIdx,
		ConsumedTxHash: consTxHash,
	}
	txHash := make([]byte, constants.HashLen)
	txinl := &TXInLinker{
		TXInPreImage: txinp,
		TxHash:       txHash,
	}
	txinl2 := &TXInLinker{}
	txinlBytes, err := txinl.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = txinl2.UnmarshalBinary(txinlBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid TXInPreImage!")
	}
}

func TestTXInLinkerBad2(t *testing.T) {
	cid := uint32(2)
	consTxIdx := uint32(0)
	consTxHash := make([]byte, constants.HashLen)
	txinp := &TXInPreImage{
		ChainID:        cid,
		ConsumedTxIdx:  consTxIdx,
		ConsumedTxHash: consTxHash,
	}
	txHash := make([]byte, constants.HashLen+1) // Invalid ChainID
	txinl := &TXInLinker{
		TXInPreImage: txinp,
		TxHash:       txHash,
	}
	txinl2 := &TXInLinker{}
	txinlBytes, err := txinl.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = txinl2.UnmarshalBinary(txinlBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid TxHash: incorrect byte length!")
	}
}

func TestTXInLinkerMarshalBinary(t *testing.T) {
	txin := &TXIn{}
	_, err := txin.TXInLinker.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txinl := &TXInLinker{}
	_, err = txinl.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestTXInLinkerUnmarshalBinary(t *testing.T) {
	txin := &TXIn{}
	data := make([]byte, 0)
	err := txin.TXInLinker.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txinl := &TXInLinker{}
	err = txinl.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestTXInLinkerPreHash(t *testing.T) {
	txin := &TXIn{}
	_, err := txin.TXInLinker.PreHash()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txinl := &TXInLinker{}
	_, err = txinl.PreHash()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestTXInLinkerUTXOID(t *testing.T) {
	txin := &TXIn{}
	_, err := txin.TXInLinker.UTXOID()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txinl := &TXInLinker{}
	_, err = txinl.UTXOID()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestTXInLinkerIsDeposit(t *testing.T) {
	txin := &TXIn{}
	val := txin.TXInLinker.IsDeposit()
	if val {
		t.Fatal("Should be false (1)")
	}

	txinl := &TXInLinker{}
	val = txinl.IsDeposit()
	if val {
		t.Fatal("Should be false (2)")
	}
}

func TestTXInLinkerChainID(t *testing.T) {
	txin := &TXIn{}
	_, err := txin.TXInLinker.ChainID()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txinl := &TXInLinker{}
	_, err = txinl.ChainID()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	txinl.TXInPreImage = &TXInPreImage{}
	_, err = txinl.ChainID()
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	chainID := uint32(17)
	txinl.TXInPreImage.ChainID = chainID
	retChainID, err := txinl.ChainID()
	if err != nil {
		t.Fatal(err)
	}
	if retChainID != chainID {
		t.Fatal("ChainIDs do not match")
	}
}

func TestTXInLinkerSetTxHash(t *testing.T) {
	txHashBad := make([]byte, 0)
	txin := &TXIn{}
	err := txin.TXInLinker.SetTxHash(txHashBad)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txinl := &TXInLinker{}
	err = txinl.SetTxHash(txHashBad)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	txinl.TXInPreImage = &TXInPreImage{}
	err = txinl.SetTxHash(txHashBad)
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	txHash := make([]byte, constants.HashLen)
	err = txin.TXInLinker.SetTxHash(txHash)
	if err == nil {
		t.Fatal("Should raise an error (4)")
	}

	err = txinl.SetTxHash(txHash)
	if err != nil {
		t.Fatal("Should not raise an error")
	}
}

func TestTXInLinkerConsumedTxIdx(t *testing.T) {
	txin := &TXIn{}
	_, err := txin.TXInLinker.ConsumedTxIdx()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txinl := &TXInLinker{}
	_, err = txinl.ConsumedTxIdx()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	consumedTxIdx := uint32(25519)
	txinl.TXInPreImage = &TXInPreImage{}
	txinl.TXInPreImage.ConsumedTxIdx = consumedTxIdx
	txIdx, err := txinl.ConsumedTxIdx()
	if err != nil {
		t.Fatal(err)
	}
	if txIdx != consumedTxIdx {
		t.Fatal("ConsumedTxIdxes do not match")
	}
}

func TestTXInLinkerConsumedTxHash(t *testing.T) {
	txin := &TXIn{}
	_, err := txin.TXInLinker.ConsumedTxHash()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txinl := &TXInLinker{}
	_, err = txinl.ConsumedTxHash()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	txhashTrue := make([]byte, constants.HashLen)
	txinl.TXInPreImage = &TXInPreImage{}
	txinl.TXInPreImage.ConsumedTxHash = utils.CopySlice(txhashTrue)
	txhash, err := txinl.ConsumedTxHash()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(txhash, txhashTrue) {
		t.Fatal("ConsumedTxHashes do not match")
	}
}
