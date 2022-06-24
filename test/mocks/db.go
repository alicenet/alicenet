package mocks

import (
	"context"
	"io/ioutil"

	"github.com/alicenet/alicenet/consensus/db"
	constants "github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

func NewTestRawDB() *badger.DB {
	logging.GetLogger(constants.LoggerBadger).SetOutput(ioutil.Discard)
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
