package localrpc

import (
	"bytes"
	"testing"

	"github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/application/objs/uint256"
)

func testHash() []byte {
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(i)
	}
	return b
}

func TestTXFeeTranslation(t *testing.T) {
	fee := &uint256.Uint256{}
	fee.FromUint64(12349872)
	obj1 := &objs.TxFee{TFPreImage: &objs.TFPreImage{ChainID: 42, TXOutIdx: 77, Fee: fee}, TxHash: testHash()}
	proto1, err := ForwardTranslateTxFee(obj1)
	if err != nil {
		t.Fatal("Failed to serialize txfee", err)
	}
	obj2, err := ReverseTranslateTxFee(proto1)
	if err != nil {
		t.Fatal("Failed to deserialize txfee", err)
	}

	// test transitivity
	if !obj1.TFPreImage.Fee.Eq(obj2.TFPreImage.Fee) ||
		obj1.TFPreImage.TXOutIdx != obj2.TFPreImage.TXOutIdx ||
		obj1.TFPreImage.ChainID != obj2.TFPreImage.ChainID ||
		!bytes.Equal(obj1.TxHash, obj2.TxHash) {
		t.Fatal("back and forth serialization should yield the same obj:", obj1, obj2)
	}

	// test fields between serialized and deserialized version
	h, err := ReverseTranslateByte(proto1.TxHash)
	if err != nil {
		t.Fatal("Failed to deserialize fee", err)
	}
	f := (&uint256.Uint256{})
	err = f.UnmarshalString(proto1.TFPreImage.Fee)
	if err != nil {
		t.Fatal("Failed to deserialize txfee", err)
	}
	if obj1.TFPreImage.ChainID != proto1.TFPreImage.ChainID ||
		obj1.TFPreImage.TXOutIdx != proto1.TFPreImage.TXOutIdx ||
		!obj1.TFPreImage.Fee.Eq(f) ||
		!bytes.Equal(obj1.TxHash, h) {
		t.Fatal("should have same fields after serialization")
	}
}
