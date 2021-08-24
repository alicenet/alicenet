package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"

	"github.com/MadBase/MadNet/utils"
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
	if err := conDB.Init(stateDb); err != nil {
		panic(err)
	}

	err = stateDb.View(func(txn *badger.Txn) error {
		iter := conDB.GetStagedBlockHeaderKeyIter(txn)
		defer iter.Close()
		for {
			_, bh, isDone, err := iter.Next()
			if err != nil {
				panic(err)
			}
			if isDone {
				break
			}
			if err := printBClaims(bh); err != nil {
				panic(err)
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
