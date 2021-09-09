package objs

import (
	"bytes"
	"testing"

	"github.com/MadBase/MadNet/application/objs/uint256"
)

func TestTFPreImageGood(t *testing.T) {
	cid := uint32(2)
	fee, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	tfp := &TFPreImage{
		ChainID:  cid,
		Fee:      fee,
		TXOutIdx: txoid,
	}
	tfp2 := &TFPreImage{}
	tfpBytes, err := tfp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = tfp2.UnmarshalBinary(tfpBytes)
	if err != nil {
		t.Fatal(err)
	}
	tfpiEqual(t, tfp, tfp2)
}

func tfpiEqual(t *testing.T, tfpi1, tfpi2 *TFPreImage) {
	if tfpi1.ChainID != tfpi2.ChainID {
		t.Fatal("Do not agree on ChainID!")
	}
	if !tfpi1.Fee.Eq(tfpi2.Fee) {
		t.Fatal("Do not agree on Next!")
	}
	if tfpi1.TXOutIdx != tfpi2.TXOutIdx {
		t.Fatal("Do not agree on TXOutIdx!")
	}
}

func TestTFPreImageBad1(t *testing.T) {
	cid := uint32(0) // Invalid ChainID
	fee, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	tfp := &TFPreImage{
		ChainID:  cid,
		Fee:      fee,
		TXOutIdx: txoid,
	}
	tfp2 := &TFPreImage{}
	tfpBytes, err := tfp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = tfp2.UnmarshalBinary(tfpBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid ChainID!")
	}
}

func TestTFPreImageMarshalBinary(t *testing.T) {
	tf := &TxFee{}
	_, err := tf.TFPreImage.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	tfp := &TFPreImage{}
	_, err = tfp.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestTFPreImageUnmarshalBinary(t *testing.T) {
	data := make([]byte, 0)
	tf := &TxFee{}
	err := tf.TFPreImage.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	tfp := &TFPreImage{}
	err = tfp.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestTFPreImagePreHash(t *testing.T) {
	tf := &TxFee{}
	_, err := tf.TFPreImage.PreHash()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	tfp := &TFPreImage{}
	_, err = tfp.PreHash()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	// preHash is present; should not fail
	tfpGoodPH := &TFPreImage{}
	tfpGoodPH.preHash = make([]byte, 32)
	tfpGoodPHOut, err := tfpGoodPH.PreHash()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tfpGoodPHOut, tfpGoodPH.preHash) {
		t.Fatal("PreHashes do not match (1)")
	}

	// Make new
	cid := uint32(2)
	fee, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	tfp = &TFPreImage{
		ChainID:  cid,
		Fee:      fee,
		TXOutIdx: txoid,
	}
	out, err := tfp.PreHash()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, tfp.preHash) {
		t.Fatal("PreHashes do not match (2)")
	}
}
