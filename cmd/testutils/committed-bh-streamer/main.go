package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/objs"

	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

func main() {

	stateDbPath := flag.String("path", "", "path to db")
	flag.Parse()
	// create execution context for application
	ctx := context.Background()
	nodeCtx, cf := context.WithCancel(ctx)
	defer cf()

	stateDb, err := utils.OpenBadger(
		nodeCtx.Done(),
		*stateDbPath,
		false,
	)
	if err != nil {
		panic(err)
	}
	defer stateDb.Close()
	conDB := &db.Database{}
	// Initialize consensus database
	conDB.Init(stateDb)

	err = stateDb.View(func(txn *badger.Txn) error {
		for i := uint32(1); ; i++ {
			bh, err := conDB.GetCommittedBlockHeader(txn, i)
			if err != nil {
				if err != badger.ErrKeyNotFound {
					return err
				}
				break
			}
			err = printBClaims(bh)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func printBClaims(bh *objs.BlockHeader) error {
	bcc := bh.BClaims
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
