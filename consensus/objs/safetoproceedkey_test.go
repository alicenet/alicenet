package objs

import (
	"bytes"
	"testing"
)

func stpkEqual(t *testing.T, stpk, stpk2 *SafeToProceedKey) {
	if !bytes.Equal(stpk.Prefix, stpk2.Prefix) {
		t.Fatal("fail")
	}
	if stpk.Height != stpk2.Height {
		t.Fatal("fail")
	}
}

func TestSafeToProceedKey(t *testing.T) {
	prefix := []byte("Prefix")
	height := uint32(13)
	stpk := &SafeToProceedKey{
		Prefix: prefix,
		Height: height,
	}
	data, err := stpk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	stpk2 := &SafeToProceedKey{}
	err = stpk2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	stpkEqual(t, stpk, stpk2)
}

func TestSafeToProceedKeyBad(t *testing.T) {
	stpk := &SafeToProceedKey{}
	_, err := stpk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}
	stpk.Prefix = []byte("Prefix")
	_, err = stpk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	stpk.Prefix = nil
	dataBad0 := []byte("0000000000000000000000")
	err = stpk.UnmarshalBinary(dataBad0)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	dataHeightGood := make([]byte, 4)
	dataHeightGood[3] = 1

	dataPrefix := make([]byte, 6)
	dataHeightBad0 := make([]byte, 5)
	dataNilBad := make([]byte, 1)
	dataBad1 := []byte{}
	dataBad1 = append(dataBad1, dataPrefix...)
	dataBad1 = append(dataBad1, []byte("|")...)
	dataBad1 = append(dataBad1, dataHeightBad0...)
	dataBad1 = append(dataBad1, []byte("|")...)
	dataBad1 = append(dataBad1, dataNilBad...)
	err = stpk.UnmarshalBinary(dataBad1)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	dataHeightBad1 := make([]byte, 4)
	dataBad2 := []byte{}
	dataBad2 = append(dataBad2, dataPrefix...)
	dataBad2 = append(dataBad2, []byte("|")...)
	dataBad2 = append(dataBad2, dataHeightBad1...)
	dataBad2 = append(dataBad2, []byte("|")...)
	dataBad2 = append(dataBad2, dataNilBad...)
	err = stpk.UnmarshalBinary(dataBad2)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	dataBad3 := []byte{}
	dataBad3 = append(dataBad3, dataPrefix...)
	dataBad3 = append(dataBad3, []byte("|")...)
	dataBad3 = append(dataBad3, dataHeightGood...)
	dataBad3 = append(dataBad3, []byte("|")...)
	dataBad3 = append(dataBad3, dataNilBad...)
	err = stpk.UnmarshalBinary(dataBad3)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}
}
