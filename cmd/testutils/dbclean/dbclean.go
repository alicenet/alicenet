package main

import (
	"flag"
	"fmt"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/logging"
	"github.com/dgraph-io/badger/v2"
)

func main() {
	dbpath := flag.String("p", "", "Path to the db")
	ratio := flag.Float64("r", .01, "GC Ratio")
	num := flag.Int("n", 1000, "Num iterations")
	flag.Parse()

	// create execution context for application
	logger := logging.GetLogger(constants.LoggerBadger)
	opts := badger.DefaultOptions(*dbpath)
	opts.CompactL0OnClose = true
	opts.NumLevelZeroTables = 1
	opts.NumLevelZeroTablesStall = 2
	opts.ValueLogFileSize = 1024 * 1024 * 10
	opts.Logger = logger

	rawConsensusDb, err := badger.Open(opts)
	if err != nil {
		logger.Errorf("Could not open database: %v", err)
		panic(err)
	}
	defer rawConsensusDb.Close()

	// Initialize consensus db: stores all state the consensus mechanism requires to work
	err = rawConsensusDb.Flatten(4)
	if err != nil {
		panic(err)
	}
	for i := 0; i < *num; i++ {
		fmt.Printf("Running iteration: %v\n", i)
		if err := rawConsensusDb.RunValueLogGC(*ratio); err != nil {
			if err != badger.ErrNoRewrite {
				panic(err)
			}
			fmt.Println("Done")
			break
		}
	}
	err = rawConsensusDb.Flatten(4)
	if err != nil {
		panic(err)
	}
}
