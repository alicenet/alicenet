package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/utils"
)

func TestTXInGood(t *testing.T) {
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
	sig := make([]byte, constants.CurveSecp256k1SigLen)
	txin := &TXIn{
		TXInLinker: txinl,
		Signature:  sig,
	}
	txin2 := &TXIn{}
	txinBytes, err := txin.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = txin2.UnmarshalBinary(txinBytes)
	if err != nil {
		t.Fatal(err)
	}
	txinEqual(t, txin, txin2)
}

func txinEqual(t *testing.T, txin1, txin2 *TXIn) {
	txinl1 := txin1.TXInLinker
	txinl2 := txin2.TXInLinker
	txinlEqual(t, txinl1, txinl2)
	if !bytes.Equal(txin1.Signature, txin2.Signature) {
		t.Fatal("Do not agree on Signature!")
	}
}

func TestTXInBad1(t *testing.T) {
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
	sig := make([]byte, constants.CurveSecp256k1SigLen)
	txin := &TXIn{
		TXInLinker: txinl,
		Signature:  sig,
	}
	txin2 := &TXIn{}
	txinBytes, err := txin.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = txin2.UnmarshalBinary(txinBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid TXInLinker!")
	}
}

func TestTXInBad2(t *testing.T) {
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
	sig := make([]byte, 0) // Invalid Signature
	txin := &TXIn{
		TXInLinker: txinl,
		Signature:  sig,
	}
	txin2 := &TXIn{}
	txinBytes, err := txin.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = txin2.UnmarshalBinary(txinBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid Signature: incorrect byte length!")
	}
}

func TestTXInMarshalBinary(t *testing.T) {
	txIn := &TXIn{}
	_, err := txIn.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error")
	}
}

func TestTXInUnmarshalBinary(t *testing.T) {
	txIn := &TXIn{}
	data := make([]byte, 0)
	err := txIn.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error")
	}
}

func TestTXInPreHash(t *testing.T) {
	txIn := &TXIn{}
	_, err := txIn.PreHash()
	if err == nil {
		t.Fatal("Should raise an error")
	}
}

func TestTXInUTXOID(t *testing.T) {
	txIn := &TXIn{}
	_, err := txIn.UTXOID()
	if err == nil {
		t.Fatal("Should raise an error")
	}
}

func TestTXInIsDeposit(t *testing.T) {
	txIn := &TXIn{}
	val := txIn.IsDeposit()
	if val {
		t.Fatal("Should be false")
	}
}

func TestTXInTxHash(t *testing.T) {
	txIn := &TXIn{}
	_, err := txIn.TxHash()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txIn.TXInLinker = &TXInLinker{}
	_, err = txIn.TxHash()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	txHashTrue := make([]byte, constants.HashLen)
	txIn.TXInLinker.TxHash = utils.CopySlice(txHashTrue)
	txHash, err := txIn.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(txHash, txHashTrue) {
		t.Fatal("TxHashes do not match")
	}
}

func TestTXInSetTxHash(t *testing.T) {
	txHashBad := make([]byte, 0)
	txIn := &TXIn{}
	err := txIn.SetTxHash(txHashBad)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txIn.TXInLinker = &TXInLinker{}
	err = txIn.SetTxHash(txHashBad)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	txHash := make([]byte, constants.HashLen)
	err = txIn.SetTxHash(txHash)
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	err = txIn.SetTxHash(txHash)
	if err == nil {
		t.Fatal("Should raise an error (4)")
	}
}

func TestTXInChainID(t *testing.T) {
	txIn := &TXIn{}
	_, err := txIn.ChainID()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txIn.TXInLinker = &TXInLinker{}
	_, err = txIn.ChainID()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestTXInConsumedTxIdx(t *testing.T) {
	txIn := &TXIn{}
	_, err := txIn.ConsumedTxIdx()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txIn.TXInLinker = &TXInLinker{}
	_, err = txIn.ConsumedTxIdx()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestTXInConsumedTxHash(t *testing.T) {
	txIn := &TXIn{}
	_, err := txIn.ConsumedTxHash()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	txIn.TXInLinker = &TXInLinker{}
	_, err = txIn.ConsumedTxHash()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}
