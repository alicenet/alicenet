package lstate

import (
	"bytes"
	"testing"

	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/crypto"
)

func maketestBH() *objs.BlockHeader {
	bh := &objs.BlockHeader{
		BClaims: &objs.BClaims{
			ChainID:    1,
			Height:     1,
			TxCount:    0,
			PrevBlock:  make([]byte, 32),
			TxRoot:     crypto.Hasher([]byte{}),
			StateRoot:  make([]byte, 32),
			HeaderRoot: make([]byte, 32),
		},
		TxHshLst: [][]byte{},
		SigGroup: make([]byte, 192),
	}
	return bh
}

func TestBHAdd(t *testing.T) {
	bhc := &bHCache{}
	err := bhc.init()
	if err != nil {
		t.Fatal(err)
	}
	bh := maketestBH()
	err = bhc.add(bh)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBHAddGet(t *testing.T) {
	bhc := &bHCache{}
	err := bhc.init()
	if err != nil {
		t.Fatal(err)
	}
	bh := maketestBH()
	err = bhc.add(bh)
	if err != nil {
		t.Fatal(err)
	}
	bHsh, err := bh.BlockHash()
	if err != nil {
		t.Fatal(err)
	}
	bh2, ok := bhc.get(bHsh)
	if !ok {
		t.Fatal("not ok")
	}
	bh2Bytes, err := bh2.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	bhBytes, err := bh.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bh2Bytes, bhBytes) {
		t.Fatal("not equal")
	}
}

func TestBHAddContainsRemove(t *testing.T) {
	bhc := &bHCache{}
	err := bhc.init()
	if err != nil {
		t.Fatal(err)
	}
	bh := maketestBH()
	bHsh, err := bh.BClaims.BlockHash()
	if err != nil {
		t.Fatal(err)
	}
	err = bhc.add(bh)
	if err != nil {
		t.Fatal(err)
	}
	present := bhc.removeBlockHash(bHsh)
	if !present {
		t.Fatal("Contains bh after it should have been removed!")
	}
	present = bhc.containsBlockHash(bHsh)
	if present {
		t.Fatal("Contains bh!")
	}
}

func TestBHPurge(t *testing.T) {
	bhc := &bHCache{}
	err := bhc.init()
	if err != nil {
		t.Fatal(err)
	}
	bh := maketestBH()
	err = bhc.add(bh)
	if err != nil {
		t.Fatal(err)
	}
	bHsh, err := bh.BClaims.BlockHash()
	if err != nil {
		t.Fatal(err)
	}
	present := bhc.containsBlockHash(bHsh)
	if !present {
		t.Fatal("Does not contain bh!")
	}
	bhc.purge()
	present = bhc.containsBlockHash(bHsh)
	if present {
		t.Fatal("Contains bh!")
	}
}
