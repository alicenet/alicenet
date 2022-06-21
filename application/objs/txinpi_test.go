package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
)

func TestTXInPreImageGood(t *testing.T) {
	cid := uint32(2)
	consTxIdx := uint32(0)
	consTxHash := make([]byte, constants.HashLen)
	txinp := &TXInPreImage{
		ChainID:        cid,
		ConsumedTxIdx:  consTxIdx,
		ConsumedTxHash: consTxHash,
	}
	txinp2 := &TXInPreImage{}
	txinpBytes, err := txinp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = txinp2.UnmarshalBinary(txinpBytes)
	if err != nil {
		t.Fatal(err)
	}
	txinpiEqual(t, txinp, txinp2)
}

func txinpiEqual(t *testing.T, txinpi1, txinpi2 *TXInPreImage) {
	if txinpi1.ChainID != txinpi2.ChainID {
		t.Fatal("Do not agree on ChainID!")
	}
	if txinpi1.ConsumedTxIdx != txinpi2.ConsumedTxIdx {
		t.Fatal("Do not agree on IssuedAt!")
	}
	if !bytes.Equal(txinpi1.ConsumedTxHash, txinpi2.ConsumedTxHash) {
		t.Fatal("Do not agree on Index!")
	}
}

func TestTXInPreImageBad1(t *testing.T) {
	cid := uint32(2)
	consTxIdx := uint32(0)
	consTxHash := make([]byte, constants.HashLen+1) // Invalid TxHash
	txinp := &TXInPreImage{
		ChainID:        cid,
		ConsumedTxIdx:  consTxIdx,
		ConsumedTxHash: consTxHash,
	}
	_, err := txinp.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise error for invalid TxHash: incorrect byte length!")
	}
}

func TestTXInPreImageMarshalBinary(t *testing.T) {
	txinl := &TXInLinker{}
	_, err := txinl.TXInPreImage.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txinp := &TXInPreImage{}
	_, err = txinp.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestTXInPreImageUnmarshalBinary(t *testing.T) {
	txinl := &TXInLinker{}
	data := make([]byte, 0)
	err := txinl.TXInPreImage.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txinp := &TXInPreImage{}
	err = txinp.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestTXInPreImagePreHash(t *testing.T) {
	txinl := &TXInLinker{}
	_, err := txinl.TXInPreImage.PreHash()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txinp := &TXInPreImage{}
	_, err = txinp.PreHash()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestTXInPreImageUTXOID(t *testing.T) {
	txinl := &TXInLinker{}
	_, err := txinl.TXInPreImage.UTXOID()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txinp := &TXInPreImage{}
	_, err = txinp.UTXOID()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestTXInPreImageIsDeposit(t *testing.T) {
	txinl := &TXInLinker{}
	val := txinl.TXInPreImage.IsDeposit()
	if val {
		t.Fatal("Should be false (1)")
	}

	txinp := &TXInPreImage{}
	val = txinp.IsDeposit()
	if val {
		t.Fatal("Should be false (2)")
	}

	txinp.ChainID = 1
	txinp.ConsumedTxIdx = constants.MaxUint32
	txinp.ConsumedTxHash = crypto.Hasher([]byte{})
	val = txinp.IsDeposit()
	if !val {
		t.Fatal("Should be true")
	}
}
