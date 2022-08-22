package main

import (
	"context"
	"os/user"
	"path/filepath"

	"flag"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	var path = flag.String("dbpath", "", "Path to validator state db.")
	flag.Parse()
	ctx := context.Background()
	nodeCtx, cf := context.WithCancel(ctx)
	defer cf()
	logger := logging.GetLogger("test")
	stateDb := openBadger(nodeCtx, logger, *path, false)
	//monitorDb := monitor.NewDatabase(nodeCtx, config.Configuration.Chain.MonitorDbPath, config.Configuration.Chain.MonitorDbInMemory)
	prefixList := [][]byte{
		[]byte("aa"),
		[]byte("ab"),
		[]byte("ac"),
		[]byte("ae"),
		[]byte("af"),
		[]byte("ag"),
		[]byte("ah"),
		[]byte("ai"),
		[]byte("aj"),
		[]byte("ak"),
		[]byte("al"),
		[]byte("am"),
		[]byte("an"),
		[]byte("ao"),
		[]byte("ap"),
		[]byte("aq"),
		[]byte("ar"),
		[]byte("as"),
		[]byte("at"),
		[]byte("au"),
		[]byte("av"),
		[]byte("aw"),
		[]byte("ax"),
		[]byte("ay"),
		[]byte("az"),
		[]byte("a1"),
		[]byte("a2"),
		[]byte("a3"),
		[]byte("a4"),
		[]byte("a5"),
		[]byte("Ay"),
		[]byte("Az"),
		[]byte("A1"),
		[]byte("A2"),
		[]byte("A3"),
		[]byte("na"),
		[]byte("nb"),
		[]byte("nc"),
		[]byte("nd"),
		[]byte("nl"),
		[]byte("nm"),
		[]byte("nn"),
		[]byte("no"),
		[]byte("np"),
		[]byte("nq"),
		[]byte("nr"),
		[]byte("ns"),
		[]byte("nt"),
		[]byte("nu"),
		[]byte("nv"),
		[]byte("nw"),
		[]byte("nx"),
		[]byte("ny"),
		[]byte("nz"),
		[]byte("n0"),
		[]byte("n1"),
		[]byte("n2"),
		[]byte("n3"),
		[]byte("n4"),
		[]byte("n5"),
		[]byte("n6"),
		[]byte("n7"),
	}

	for _, pf := range prefixList {
		if err := stateDb.DropPrefix(utils.CopySlice(pf)); err != nil {
			panic(err)
		}
	}

}

func openBadger(ctx context.Context, logger *logrus.Logger, directoryName string, inMemory bool) *badger.DB {
	if len(directoryName) >= 2 {
		if directoryName[0:2] == "~/" {
			usr, err := user.Current()
			if err != nil {
				panic(err)
			}
			directoryName = filepath.Join(usr.HomeDir, directoryName[1:])
			logger.Infof("Directory:%v", directoryName)
		}
	}

	logger.Infof("Opening badger DB... In-Memory:%v Directory:%v", inMemory, directoryName)

	opts := badger.DefaultOptions(directoryName).WithInMemory(inMemory).WithSyncWrites(true)
	opts.Logger = logging.GetLogger(constants.LoggerBadger)

	db, err := badger.Open(opts)
	if err != nil {
		logger.Panicf("Could not open database: %v", err)
	}

	go func() {
		defer db.Close()
		<-ctx.Done()
	}()

	return db
}
