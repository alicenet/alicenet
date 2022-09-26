package main

import (
	"context"
	"flag"
	"fmt"

	"time"

	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/localrpc"
)

func main() {
	hPtr := flag.String("host", "127.0.0.1:8884", "Host to connect to.")
	startPtr := flag.Int("start", 1, "Start block.")
	stopPtr := flag.Int("stop", 0, "Start block. Will print all known blocks if not included or set to zero.")
	followModePtr := flag.Bool("follow", false, "Follow mode when present.")
	flag.Parse()
	client := &localrpc.Client{Address: *hPtr, TimeOut: time.Second * 3}
	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		panic(err)
	}
	start := uint32(*startPtr)
	for {
		bn, err := client.GetBlockNumber(ctx)
		if err != nil {
			panic(err)
		}
		if start > bn || start <= 0 {
			start = 1
		}
		if *stopPtr > 0 {
			if bn < uint32(*stopPtr) {
				bn = uint32(*stopPtr)
			}
		}
		for i := start + 1; i <= bn; i++ {
			if *stopPtr > 0 {
				if uint32(*stopPtr) == start {
					bn = uint32(*stopPtr)
				}
			}
			start++
			bh, err := client.GetBlockHeader(ctx, i)
			if err != nil {
				panic(err)
			}
			printBH(bh)
		}
		time.Sleep(3 * time.Second)
		start = bn
		if !*followModePtr {
			break
		}
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
