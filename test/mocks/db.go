package mocks

import (
	"context"
	"io"

	"github.com/dgraph-io/badger/v2"

	"github.com/alicenet/alicenet/consensus/db"
	constants "github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
)

func NewTestRawDB() *badger.DB {
	logging.GetLogger(constants.LoggerBadger).SetOutput(io.Discard)
	db, err := utils.OpenBadger(context.Background().Done(), "", true)
	if err != nil {
		panic(err)
	}
	return db
}

func NewTestDB() *db.Database {
	db := &db.Database{}
	db.Init(NewTestRawDB())
	return db
}
