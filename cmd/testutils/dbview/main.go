package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/objs"
	mnutils "github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

// Command is the cobra.Command specifically for running as a node
func initDatabase(ctx context.Context, path string, inMemory bool) *badger.DB {
	db, err := mnutils.OpenBadger(ctx.Done(), path, inMemory)
	if err != nil {
		panic(err)
	}
	return db
}
func main() {
	dbpath := flag.String("p", "", "Path to the db")
	cmd := flag.String("c", "", "function to run")
	height := flag.String("h", "", "height")
	stop := flag.String("s", "", "Stop height")
	flag.Parse()
	// create execution context for application
	ctx := context.Background()
	nodeCtx, cf := context.WithCancel(ctx)
	defer cf()
	// Initialize consensus db: stores all state the consensus mechanism requires to work
	rawConsensusDb := initDatabase(nodeCtx, *dbpath, false)
	defer rawConsensusDb.Close()
	db := &db.Database{}
	db.Init(rawConsensusDb)
	err := rawConsensusDb.View(func(txn *badger.Txn) error {
		switch *cmd {
		case "GetValidatorSet":
			hb, ok := new(big.Int).SetString(*height, 10)
			if !ok {
				panic("bad height")
			}
			h64 := hb.Int64()
			h := uint32(h64)
			vs, err := db.GetValidatorSet(txn, h)
			if err != nil {
				panic(err)
			}
			fmt.Println("Validators")
			for i := 0; i < len(vs.Validators); i++ {
				fmt.Printf("    %x\n", vs.Validators[i].VAddr)
				fmt.Printf("    %x\n", vs.Validators[i].GroupShare)
			}
			fmt.Printf("GroupKey\n    %x\n", vs.GroupKey)
			fmt.Printf("NotBefore\n    %v\n", vs.NotBefore)
		case "GetLastSnapshot":
			bh, err := db.GetLastSnapshot(txn)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Snapshot Block Header %#v\n", bh)
		case "GetSnapshotByHeight":
			hb, ok := new(big.Int).SetString(*height, 10)
			if !ok {
				panic("bad height")
			}
			h64 := hb.Int64()
			h := uint32(h64)
			bh, err := db.GetSnapshotByHeight(txn, h)
			if err != nil {
				panic(err)
			}
			printBH(bh)
		case "BlockHeader":
			hb, ok := new(big.Int).SetString(*height, 10)
			if !ok {
				panic("bad height")
			}
			h64 := hb.Int64()
			h := uint32(h64)
			sb, ok := new(big.Int).SetString(*stop, 10)
			if !ok {
				panic("bad height")
			}
			s64 := sb.Int64()
			s := uint32(s64)
			for i := h; i <= s; i++ {
				bh, err := db.GetCommittedBlockHeader(txn, i)
				if err != nil {
					panic(err)
				}
				printBH(bh)
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
func printBH(bh *objs.BlockHeader) {
	fmt.Printf("   CHAINID: %v\n", bh.BClaims.ChainID)
	fmt.Printf("    HEIGHT: %v\n", bh.BClaims.Height)
	fmt.Printf("   TXCOUNT: %v\n", bh.BClaims.TxCount)
	fmt.Printf("    TXROOT: %x\n", bh.BClaims.TxRoot)
	fmt.Printf(" STATEROOT: %x\n", bh.BClaims.StateRoot)
	fmt.Printf("HEADERROOT: %x\n", bh.BClaims.HeaderRoot)
	fmt.Printf(" PREVBLOCK: %x\n", bh.BClaims.PrevBlock)
	fmt.Printf(" SIGNATURE: %x\n", bh.SigGroup)
	fmt.Printf(" TXS: %x", bh.TxHshLst)
	fmt.Printf("\n")
	fmt.Printf("\n")
}
