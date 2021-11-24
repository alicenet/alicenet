package objects_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/blockchain/objects"
)

func TestParticipantCopy(t *testing.T) {
	p := &objects.Participant{}
	addrBytes := make([]byte, 20)
	addrBytes[0] = 255
	addrBytes[19] = 255
	p.Address.SetBytes(addrBytes)
	publicKey := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	p.PublicKey = publicKey
	index := 13
	p.Index = index

	c := p.Copy()
	pBytes := p.Address.Bytes()
	cBytes := c.Address.Bytes()
	if !bytes.Equal(pBytes, cBytes) {
		t.Fatal("bytes do not match")
	}
	pPubKey := p.PublicKey
	cPubKey := c.PublicKey
	if pPubKey[0].Cmp(cPubKey[0]) != 0 || pPubKey[1].Cmp(cPubKey[1]) != 0 {
		t.Fatal("public keys do not match")
	}
	if p.Index != c.Index {
		t.Fatal("Indices do not match")
	}

	pString := p.String()
	cString := c.String()
	if pString != cString {
		t.Fatal("strings do not match")
	}
}

func TestParticipantListExtractIndices(t *testing.T) {
	p1 := &objects.Participant{Index: 1}
	p2 := &objects.Participant{Index: 2}
	p3 := &objects.Participant{Index: 3}
	p4 := &objects.Participant{Index: 4}

	pl := objects.ParticipantList{p4, p2, p3, p1}
	indices := []int{4, 2, 3, 1}
	retIndices := pl.ExtractIndices()
	if len(indices) != len(retIndices) {
		t.Fatal("invalid indices")
	}
	for k := 0; k < len(indices); k++ {
		if indices[k] != retIndices[k] {
			t.Fatal("invalid indices when looping")
		}
	}
}
