package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
)

func txckEqual(t *testing.T, txck, txck2 *TxCacheKey) {
	if !bytes.Equal(txck.Prefix, txck2.Prefix) {
		t.Fatal("fail")
	}
	if txck.Height != txck2.Height {
		t.Fatal("fail")
	}
	if txck.Height == 0 {
		t.Fatal("fail")
	}
	if !bytes.Equal(txck.TxHash, txck2.TxHash) {
		t.Fatal("fail")
	}
	if len(txck.TxHash) != constants.HashLen {
		t.Fatal("fail")
	}
}

func TestTxCacheKey(t *testing.T) {
	prefix := []byte("prefix")
	height := uint32(1)
	txHash := crypto.Hasher([]byte("txHash"))
	txck := &TxCacheKey{
		Prefix: prefix,
		Height: height,
		TxHash: txHash,
	}
	data, err := txck.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	txck2 := &TxCacheKey{}
	err = txck2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	txckEqual(t, txck, txck2)

	hb := utils.MarshalUint32(height)
	iterKeyTrue := []byte{}
	iterKeyTrue = append(iterKeyTrue, prefix...)
	iterKeyTrue = append(iterKeyTrue, []byte("|")...)
	iterKeyTrue = append(iterKeyTrue, hb...)
	iterKeyTrue = append(iterKeyTrue, []byte("|")...)
	iterKey, err := txck.MakeIterKey()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(iterKey, iterKeyTrue) {
		t.Fatal("IterKeys do not match")
	}
}

func TestTxCacheKeyBad(t *testing.T) {
	txck := &TxCacheKey{}
	_, err := txck.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	_, err = txck.MakeIterKey()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	dataBad0 := []byte("00000000|00000000000000")
	err = txck.UnmarshalBinary(dataBad0)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	dataHeightGood := make([]byte, 4)
	dataHeightGood[3] = 1

	prefix := []byte("prefix")
	dataHeightBad0 := make([]byte, 5)
	dataTxHashBad := make([]byte, constants.HashLen+1)
	dataBad1 := []byte{}
	dataBad1 = append(dataBad1, prefix...)
	dataBad1 = append(dataBad1, []byte("|")...)
	dataBad1 = append(dataBad1, dataHeightBad0...)
	dataBad1 = append(dataBad1, []byte("|")...)
	dataBad1 = append(dataBad1, dataTxHashBad...)
	err = txck.UnmarshalBinary(dataBad1)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	dataHeightBad1 := make([]byte, 4)
	dataBad2 := []byte{}
	dataBad2 = append(dataBad2, prefix...)
	dataBad2 = append(dataBad2, []byte("|")...)
	dataBad2 = append(dataBad2, dataHeightBad1...)
	dataBad2 = append(dataBad2, []byte("|")...)
	dataBad2 = append(dataBad2, dataTxHashBad...)
	err = txck.UnmarshalBinary(dataBad2)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}

	dataBad3 := []byte{}
	dataBad3 = append(dataBad3, prefix...)
	dataBad3 = append(dataBad3, []byte("|")...)
	dataBad3 = append(dataBad3, dataHeightGood...)
	dataBad3 = append(dataBad3, []byte("|")...)
	dataBad3 = append(dataBad3, dataTxHashBad...)
	err = txck.UnmarshalBinary(dataBad3)
	if err == nil {
		t.Fatal("Should have raised error (6)")
	}
}
