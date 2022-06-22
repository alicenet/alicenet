package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"strings"

	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/consensus/objs/bclaims"
)

func main() {
	capObj := flag.String("obj", "", "Block header to decode.")
	sol := flag.Bool("sol", false, "hex from solidity.")
	flag.Parse()
	if strings.HasPrefix(*capObj, "0x") {
		capObjd := *capObj
		capObjd = capObjd[2:]
		capObj = &capObjd
	}
	if *sol {
		capObjd := *capObj
		capObjd = capObjd[len(capObjd)-384:]
		capObj = &capObjd
	}
	err := printBClaims(*capObj)
	if err != nil {
		panic(err)
	}
}

func printBClaims(h string) error {
	hb, err := hex.DecodeString(h)
	if err != nil {
		return err
	}
	bc, err := bclaims.Unmarshal(hb)
	if err != nil {
		return err
	}
	bcc := &objs.BClaims{}
	err = bcc.UnmarshalCapn(bc)
	if err != nil {
		return err
	}
	p := `
BClaims: 
	ChainID:    %v,
	Height:     %v,
	PrevBlock:  %x,
	HeaderRoot: %x,
	StateRoot:  %x,
	TxRoot:     %x,
	TxCount:    %v,
`
	fmt.Printf(p, bcc.ChainID, bcc.Height, bcc.PrevBlock, bcc.HeaderRoot, bcc.StateRoot, bcc.TxRoot, bcc.TxCount)
	return nil
}
