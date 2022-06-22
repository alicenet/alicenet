package accusation

import (
	"context"
	"encoding/gob"
	"testing"

	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type AccusationTest1 struct {
	UUID uuid.UUID
}

func (a *AccusationTest1) SubmitToSmartContracts() error {
	return nil
}

func (a *AccusationTest1) GetUUID() uuid.UUID {
	return a.UUID
}

func (a *AccusationTest1) SetUUID(uuid uuid.UUID) {
	a.UUID = uuid
}

func (a *AccusationTest1) IsProcessed() bool {
	return false
}

var _ objs.Accusation = &AccusationTest1{}

func TestPersistenceUnknownImpl(t *testing.T) {
	ctx := context.Background()
	nodeCtx, cf := context.WithCancel(ctx)
	defer cf()

	// Initialize consensus db: stores all state the consensus mechanism requires to work
	rawConsensusDb, err := utils.OpenBadger(nodeCtx.Done(), "", true)
	assert.Nil(t, err)
	defer rawConsensusDb.Close()

	db := &db.Database{}
	db.Init(rawConsensusDb)

	uuid := uuid.New()
	acc := &AccusationTest1{}
	acc.SetUUID(uuid)

	err = db.Update(func(txn *badger.Txn) error {
		err := db.SetAccusation(txn, acc)
		assert.NotNil(t, err)

		_, err = db.GetAccusation(txn, uuid)
		assert.NotNil(t, err)
		//assert.Equal(t, acc.GetUUID().String(), acc2.GetUUID().String())

		return nil
	})

	assert.Nil(t, err)
}

///////////////////////////////////

type AccusationTest2 struct {
	UUID uuid.UUID
}

func (a *AccusationTest2) SubmitToSmartContracts() error {
	return nil
}

func (a *AccusationTest2) GetUUID() uuid.UUID {
	return a.UUID
}

func (a *AccusationTest2) SetUUID(uuid uuid.UUID) {
	a.UUID = uuid
}

func (a *AccusationTest2) IsProcessed() bool {
	return false
}

var _ objs.Accusation = &AccusationTest2{}

func TestPersistenceKnownImpl(t *testing.T) {
	// register AccusationTest2 as a known implementation of Accusation into gob
	gob.Register(&AccusationTest2{})

	ctx := context.Background()
	nodeCtx, cf := context.WithCancel(ctx)
	defer cf()

	// Initialize consensus db: stores all state the consensus mechanism requires to work
	rawConsensusDb, err := utils.OpenBadger(nodeCtx.Done(), "", true)
	assert.Nil(t, err)
	defer rawConsensusDb.Close()

	db := &db.Database{}
	db.Init(rawConsensusDb)

	uuid := uuid.New()
	acc := &AccusationTest2{}
	acc.SetUUID(uuid)

	err = db.Update(func(txn *badger.Txn) error {
		err := db.SetAccusation(txn, acc)
		assert.Nil(t, err)

		acc2, err := db.GetAccusation(txn, uuid)
		assert.Nil(t, err)
		assert.Equal(t, acc.GetUUID().String(), acc2.GetUUID().String())

		acc3, ok := acc2.(*AccusationTest2)
		assert.True(t, ok)
		assert.Equal(t, acc.GetUUID().String(), acc3.GetUUID().String())

		return nil
	})

	assert.Nil(t, err)
}
