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

var prefixList = [][]byte{
	[]byte("ag"),
	[]byte("ah"),
	[]byte("ai"),
	[]byte("aj"),
	[]byte("ak"),
	[]byte("al"),
	[]byte("am"),
	[]byte("an"),
	[]byte("ao"),
}

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
	height := flag.String("h", "", "height")
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

	err := rawConsensusDb.Update(func(txn *badger.Txn) error {
		for i := 0; i < len(prefixList); i++ {
			err := db.DB().DropPrefix(prefixList[i])
			if err != nil {
				panic(err)
			}
		}
		resetOvs(db, txn)
		hb, ok := new(big.Int).SetString(*height, 10)
		if !ok {
			panic("bad height")
		}
		h64 := hb.Int64()
		h := uint32(h64)
		resetRs(db, txn, h)
		return nil
	})
	if err != nil {
		panic(err)
	}

}

func resetOvs(db *db.Database, txn *badger.Txn) {
	ovs, err := db.GetOwnValidatingState(txn)
	if err != nil {
		panic(err)
	}
	ovs = ovsFn(ovs)
	err = db.SetOwnValidatingState(txn, ovs)
	if err != nil {
		panic(err)
	}
}

func resetRs(db *db.Database, txn *badger.Txn, height uint32) {
	vss, err := db.GetValidatorSet(txn, height)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(vss.Validators); i++ {
		rs, err := db.GetCurrentRoundState(txn, vss.Validators[i].VAddr)
		if err != nil {
			panic(err)
		}
		rs = rsFn(rs)
		err = db.SetCurrentRoundState(txn, rs)
		if err != nil {
			panic(err)
		}
		for j := uint32(0); j < 5; j++ {
			rs, err := db.GetHistoricRoundState(txn, vss.Validators[i].VAddr, height, j)
			if err == nil {
				rs = rsFn(rs)
				err = db.SetHistoricRoundState(txn, rs)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func ovsFn(ovs *objs.OwnValidatingState) *objs.OwnValidatingState {
	ovsin := &objs.OwnValidatingState{
		RoundStarted:         ovs.RoundStarted,
		PreVoteStepStarted:   ovs.PreVoteStepStarted,
		PreCommitStepStarted: ovs.PreCommitStepStarted,
	}
	return ovsin
}

func rsFn(rs *objs.RoundState) *objs.RoundState {
	rsin := &objs.RoundState{
		VAddr:      rs.VAddr,
		GroupKey:   rs.GroupKey,
		GroupShare: rs.GroupShare,
		GroupIdx:   rs.GroupIdx,
		RCert:      rs.RCert,
	}
	return rsin
}

//nolint:unused,deadcode
func PrefixOwnValidatingState() []byte {
	return []byte("aa")
}

//nolint:unused,deadcode
func PrefixCurrentRoundState() []byte {
	return []byte("ab")
}

//nolint:unused,deadcode
func PrefixHistoricRoundState() []byte {
	return []byte("ac")
}

//nolint:unused,deadcode
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
