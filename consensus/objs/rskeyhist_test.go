package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
)

func rshkEqual(t *testing.T, rshk, rshk2 *RoundStateHistoricKey) {
	if !bytes.Equal(rshk.Prefix, rshk2.Prefix) {
		t.Fatal("fail")
	}
	if rshk.Height != rshk2.Height {
		t.Fatal("fail")
	}
	if rshk.Round != rshk2.Round {
		t.Fatal("fail")
	}
	if !bytes.Equal(rshk.VAddr, rshk2.VAddr) {
		t.Fatal("fail")
	}
}

func TestRoundStateHistoricKey(t *testing.T) {
	prefix := []byte("Prefix")
	height := uint32(13)
	round := uint32(1)
	vaddr := make([]byte, constants.OwnerLen)
	rshk := &RoundStateHistoricKey{
		Prefix: prefix,
		Height: height,
		Round:  round,
		VAddr:  vaddr,
	}
	data, err := rshk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rshk2 := &RoundStateHistoricKey{}
	err = rshk2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	rshkEqual(t, rshk, rshk2)

	_, err = rshk.MakeIterKey()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRoundStateHeightKeyBad(t *testing.T) {
	rshk := &RoundStateHistoricKey{}
	_, err := rshk.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}
	_, err = rshk.MakeIterKey()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	dataBad0 := []byte("00000000|00000000000000")
	err = rshk.UnmarshalBinary(dataBad0)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	dataHeightGood := make([]byte, 4)
	dataHeightGood[3] = 1
	dataRoundGood := make([]byte, 4)
	dataRoundGood[3] = 1

	dataPrefix := make([]byte, 6)
	dataHeightBad0 := make([]byte, 5)
	dataRoundBad0 := make([]byte, 5)
	dataVAddr := make([]byte, constants.OwnerLen)
	dataBad1 := []byte{}
	dataBad1 = append(dataBad1, dataPrefix...)
	dataBad1 = append(dataBad1, []byte("|")...)
	dataBad1 = append(dataBad1, dataHeightBad0...)
	dataBad1 = append(dataBad1, []byte("|")...)
	dataBad1 = append(dataBad1, dataRoundBad0...)
	dataBad1 = append(dataBad1, []byte("|")...)
	dataBad1 = append(dataBad1, dataVAddr...)
	err = rshk.UnmarshalBinary(dataBad1)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	dataHeightBad1 := make([]byte, 4)
	dataBad2 := []byte{}
	dataBad2 = append(dataBad2, dataPrefix...)
	dataBad2 = append(dataBad2, []byte("|")...)
	dataBad2 = append(dataBad2, dataHeightBad1...)
	dataBad2 = append(dataBad2, []byte("|")...)
	dataBad2 = append(dataBad2, dataRoundBad0...)
	dataBad2 = append(dataBad2, []byte("|")...)
	dataBad2 = append(dataBad2, dataVAddr...)
	err = rshk.UnmarshalBinary(dataBad2)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	dataBad3 := []byte{}
	dataBad3 = append(dataBad3, dataPrefix...)
	dataBad3 = append(dataBad3, []byte("|")...)
	dataBad3 = append(dataBad3, dataHeightGood...)
	dataBad3 = append(dataBad3, []byte("|")...)
	dataBad3 = append(dataBad3, dataRoundBad0...)
	dataBad3 = append(dataBad3, []byte("|")...)
	dataBad3 = append(dataBad3, dataVAddr...)
	err = rshk.UnmarshalBinary(dataBad3)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}

	dataRoundBad1 := make([]byte, 4)
	dataBad4 := []byte{}
	dataBad4 = append(dataBad4, dataPrefix...)
	dataBad4 = append(dataBad4, []byte("|")...)
	dataBad4 = append(dataBad4, dataHeightGood...)
	dataBad4 = append(dataBad4, []byte("|")...)
	dataBad4 = append(dataBad4, dataRoundBad1...)
	dataBad4 = append(dataBad4, []byte("|")...)
	dataBad4 = append(dataBad4, dataVAddr...)
	err = rshk.UnmarshalBinary(dataBad4)
	if err == nil {
		t.Fatal("Should have raised error (6)")
	}

	dataBad5 := []byte{}
	dataBad5 = append(dataBad5, dataPrefix...)
	dataBad5 = append(dataBad5, []byte("|")...)
	dataBad5 = append(dataBad5, dataHeightGood...)
	dataBad5 = append(dataBad5, []byte("|")...)
	dataBad5 = append(dataBad5, dataRoundGood...)
	dataBad5 = append(dataBad5, []byte("|")...)
	dataBad5 = append(dataBad5, dataVAddr...)
	err = rshk.UnmarshalBinary(dataBad5)
	if err == nil {
		t.Fatal("Should have raised error (7)")
	}
}
