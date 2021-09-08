package objs

import (
	"bytes"
	"testing"

	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
)

func TestTxFeeGood(t *testing.T) {
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
	txHash := make([]byte, constants.HashLen)
	tf := &TxFee{
		TFPreImage: tfp,
		TxHash:     txHash,
	}
	tf2 := &TxFee{}
	tfBytes, err := tf.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = tf2.UnmarshalBinary(tfBytes)
	if err != nil {
		t.Fatal(err)
	}
	tfEqual(t, tf, tf2)
}

func tfEqual(t *testing.T, tf1, tf2 *TxFee) {
	tfpi1 := tf1.TFPreImage
	tfpi2 := tf2.TFPreImage
	tfpiEqual(t, tfpi1, tfpi2)
	if !bytes.Equal(tf1.TxHash, tf2.TxHash) {
		t.Fatal("Do not agree on TxHash!")
	}
}

func TestTxFeeBad1(t *testing.T) {
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
	txHash := make([]byte, constants.HashLen)
	tf := &TxFee{
		TFPreImage: tfp,
		TxHash:     txHash,
	}
	tf2 := &TxFee{}
	tfBytes, err := tf.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = tf2.UnmarshalBinary(tfBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid TFPreImage!")
	}
}

func TestTxFeeBad2(t *testing.T) {
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
	txHash := make([]byte, constants.HashLen+1) // Invalid TxHash
	tf := &TxFee{
		TFPreImage: tfp,
		TxHash:     txHash,
	}
	tf2 := &TxFee{}
	tfBytes, err := tf.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = tf2.UnmarshalBinary(tfBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid TxHash: incorrect byte length!")
	}
}

func TestTxFeeNew(t *testing.T) {
	utxo := &TXOut{}
	chainID := uint32(0)
	fee, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	err = utxo.txFee.New(chainID, nil)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	tf := &TxFee{}
	err = tf.New(chainID, nil)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	chainID = 1
	err = tf.New(chainID, nil)
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	err = tf.New(chainID, fee)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxFeeMarshalBinary(t *testing.T) {
	utxo := &TXOut{}
	_, err := utxo.txFee.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	_, err = utxo.txFee.MarshalCapn(nil)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
	tf := &TxFee{}
	_, err = tf.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}
	_, err = tf.MarshalCapn(nil)
	if err == nil {
		t.Fatal("Should raise an error (4)")
	}

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
	txHash := make([]byte, constants.HashLen)
	tf = &TxFee{
		TFPreImage: tfp,
		TxHash:     txHash,
	}
	tfBytes, err := tf.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	tf2 := &TxFee{}
	err = tf2.UnmarshalBinary(tfBytes)
	if err != nil {
		t.Fatal(err)
	}
	tfEqual(t, tf, tf2)
}

func TestTxFeeUnmarshalBinary(t *testing.T) {
	data := make([]byte, 0)
	utxo := &TXOut{}
	err := utxo.txFee.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	tf := &TxFee{}
	err = tf.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

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
	txHash := make([]byte, constants.HashLen)
	tf = &TxFee{
		TFPreImage: tfp,
		TxHash:     txHash,
	}
	tfBytes, err := tf.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	tf2 := &TxFee{}
	err = tf2.UnmarshalBinary(tfBytes)
	if err != nil {
		t.Fatal(err)
	}
	tfEqual(t, tf, tf2)
}

func TestTxFeePreHash(t *testing.T) {
	utxo := &TXOut{}
	_, err := utxo.txFee.PreHash()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	tf := &TxFee{}
	_, err = tf.PreHash()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}

func TestTxFeeTXOutIdx(t *testing.T) {
	utxo := &TXOut{}
	_, err := utxo.txFee.TXOutIdx()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	tf := &TxFee{}
	_, err = tf.TXOutIdx()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
	tf.TFPreImage = &TFPreImage{}
	txOutIdx := uint32(17)
	tf.TFPreImage.TXOutIdx = txOutIdx
	out, err := tf.TXOutIdx()
	if err != nil {
		t.Fatal(err)
	}
	if out != txOutIdx {
		t.Fatal("TXOutIdxes do not match")
	}
}

func TestTxFeeUTXOID(t *testing.T) {
	utxo := &TXOut{}
	_, err := utxo.txFee.UTXOID()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	tf := &TxFee{}
	_, err = tf.UTXOID()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

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
	tf = &TxFee{
		TFPreImage: tfp,
		TxHash:     nil,
	}
	_, err = tf.UTXOID()
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	txHash := make([]byte, constants.HashLen)
	tf.TxHash = txHash
	utxoID, err := tf.UTXOID()
	if err != nil {
		t.Fatal(err)
	}
	out, err := tf.UTXOID()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, utxoID) {
		t.Fatal("utxoIDs do not match")
	}
}

func TestTxFeeSetTXOutIdx(t *testing.T) {
	idx := uint32(0)
	utxo := &TXOut{}
	err := utxo.txFee.SetTXOutIdx(idx)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	tf := &TxFee{}
	err = tf.SetTXOutIdx(idx)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
	tf.TFPreImage = &TFPreImage{}
	err = tf.SetTXOutIdx(idx)
	if err != nil {
		t.Fatal(err)
	}
	out, err := tf.TXOutIdx()
	if err != nil {
		t.Fatal(err)
	}
	if out != idx {
		t.Fatal("TXOutIdxes do not match")
	}
}

func TestTxFeeSetTxHash(t *testing.T) {
	tf := &TxFee{}
	txHash := make([]byte, 0)
	err := tf.SetTxHash(txHash)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	tf.TFPreImage = &TFPreImage{}
	err = tf.SetTxHash(txHash)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	txHash = make([]byte, constants.HashLen)
	err = tf.SetTxHash(txHash)
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestTxFeeChainID(t *testing.T) {
	tf := &TxFee{}
	_, err := tf.ChainID()
	if err == nil {
		t.Fatal("Should raise an error")
	}
	tf.TFPreImage = &TFPreImage{}
	cid := uint32(17)
	tf.TFPreImage.ChainID = cid
	chainID, err := tf.ChainID()
	if err != nil {
		t.Fatal(err)
	}
	if cid != chainID {
		t.Fatal("ChainIDs do not match")
	}
}

func TestTxFeeFee(t *testing.T) {
	tf := &TxFee{}
	_, err := tf.Fee()
	if err == nil {
		t.Fatal("Should raise an error")
	}
	tf.TFPreImage = &TFPreImage{}
	feeTrue, err := new(uint256.Uint256).FromUint64(1234567890)
	if err != nil {
		t.Fatal(err)
	}
	tf.TFPreImage.Fee = feeTrue.Clone()
	fee, err := tf.Fee()
	if err != nil {
		t.Fatal(err)
	}
	if !fee.Eq(feeTrue) {
		t.Fatal("Fees do not match")
	}
}
