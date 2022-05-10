package mocks

import (
	"context"
	"io/ioutil"

	"github.com/MadBase/MadNet/consensus/db"
	constants "github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
)

func NewMockRawDb() *badger.DB {
	logging.GetLogger(constants.LoggerBadger).SetOutput(ioutil.Discard)
	db, err := utils.OpenBadger(context.Background().Done(), "", true)
	if err != nil {
		panic(err)
	}
	return db
}

func NewMockDb() *db.Database {
	db := &db.Database{}
	db.Init(NewMockRawDb())
	return db
}
