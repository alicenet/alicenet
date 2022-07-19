package accusation

import (
	"context"
	"encoding/gob"
	"testing"
	"time"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
)

type AccusationTest1 struct {
	objs.BaseAccusation
}

func (a *AccusationTest1) SubmitToSmartContracts() error {
	return objs.ErrNotImpl
}

func (a *AccusationTest1) GetID() [32]byte {
	return a.ID
}

func (a *AccusationTest1) SetID(id [32]byte) {
	a.ID = id
}

func (a *AccusationTest1) GetPersistenceTimestamp() uint64 {
	return a.PersistenceTimestamp
}

func (a *AccusationTest1) SetPersistenceTimestamp(timestamp uint64) {
	a.PersistenceTimestamp = timestamp
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

	var id [32]byte
	copy(id[:], crypto.Hasher([]byte("test")))
	acc := &AccusationTest1{}
	acc.SetID(id)

	err = db.Update(func(txn *badger.Txn) error {
		err := db.SetAccusation(txn, acc)
		assert.NotNil(t, err)

		_, err = db.GetAccusation(txn, id)
		assert.NotNil(t, err)

		return nil
	})

	assert.Nil(t, err)
}

///////////////////////////////////

type AccusationTest2 struct {
	objs.BaseAccusation
}

func (a *AccusationTest2) SubmitToSmartContracts() error {
	return nil
}

func (a *AccusationTest2) GetID() [32]byte {
	return a.ID
}

func (a *AccusationTest2) SetID(id [32]byte) {
	a.ID = id
}

func (a *AccusationTest2) GetPersistenceTimestamp() uint64 {
	return a.PersistenceTimestamp
}

func (a *AccusationTest2) SetPersistenceTimestamp(timestamp uint64) {
	a.PersistenceTimestamp = timestamp
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

	var id [32]byte
	copy(id[:], crypto.Hasher([]byte("test")))
	acc := &AccusationTest2{}
	acc.SetID(id)

	err = db.Update(func(txn *badger.Txn) error {
		err := db.SetAccusation(txn, acc)
		assert.Nil(t, err)

		acc2, err := db.GetAccusation(txn, id)
		assert.Nil(t, err)
		assert.Equal(t, acc.GetID(), acc2.GetID())

		acc3, ok := acc2.(*AccusationTest2)
		assert.True(t, ok)
		assert.Equal(t, acc.GetID(), acc3.GetID())

		return nil
	})

	assert.Nil(t, err)
}

func TestPersistAccusation(t *testing.T) {
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

	var id [32]byte
	copy(id[:], crypto.Hasher([]byte("test")))
	acc := &AccusationTest2{}
	acc.SetID(id)
	acc.SetPersistenceTimestamp(uint64(time.Now().Unix()))
	acc.SetState(objs.Persisted)

	err = db.Update(func(txn *badger.Txn) error {
		err := db.SetAccusation(txn, acc)
		assert.Nil(t, err)

		// check the retrieved accusation has the same values as the original
		acc2, err := db.GetAccusation(txn, id)
		assert.Nil(t, err)
		assert.Equal(t, acc.GetID(), acc2.GetID())
		assert.Equal(t, acc.GetState(), acc2.GetState())
		assert.Equal(t, acc.GetPersistenceTimestamp(), acc2.GetPersistenceTimestamp())

		// check the retrieved accusation is of type AccusationTest2
		acc3, ok := acc2.(*AccusationTest2)
		assert.True(t, ok)
		assert.Equal(t, acc.GetID(), acc3.GetID())

		// get all accusations without filters
		accs, err := db.GetAccusations(txn, nil)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(accs))
		assert.Equal(t, acc.GetID(), accs[0].GetID())

		// get persisted but unsheduled accusations
		accs, err = db.GetPersistedButUnscheduledAccusations(txn)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(accs))
		assert.Equal(t, acc.GetID(), accs[0].GetID())

		// get scheduled but incomplete accusations
		accs, err = db.GetScheduledButIncompleteAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		// get completed accusations
		accs, err = db.GetCompletedAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		//////////////
		// set accusation state to ScheduledForExecution and persist it
		//////////////
		acc.SetState(objs.ScheduledForExecution)
		err = db.SetAccusation(txn, acc)
		assert.Nil(t, err)

		// check the retrieved accusation has the same values as the original
		acc2, err = db.GetAccusation(txn, id)
		assert.Nil(t, err)
		assert.Equal(t, acc.GetID(), acc2.GetID())
		assert.Equal(t, acc.GetState(), acc2.GetState())
		assert.Equal(t, acc.GetPersistenceTimestamp(), acc2.GetPersistenceTimestamp())

		// get all accusations without filters
		accs, err = db.GetAccusations(txn, nil)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(accs))
		assert.Equal(t, acc.GetID(), accs[0].GetID())

		// get persisted but unscheduled accusations
		accs, err = db.GetPersistedButUnscheduledAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		// get scheduled but incomplete accusations
		accs, err = db.GetScheduledButIncompleteAccusations(txn)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(accs))
		assert.Equal(t, acc.GetID(), accs[0].GetID())

		// get completed accusations
		accs, err = db.GetCompletedAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		//////////////
		// set accusation state to Completed and persist it
		//////////////
		acc.SetState(objs.Completed)
		err = db.SetAccusation(txn, acc)
		assert.Nil(t, err)

		// check the retrieved accusation has the same values as the original
		acc2, err = db.GetAccusation(txn, id)
		assert.Nil(t, err)
		assert.Equal(t, acc.GetID(), acc2.GetID())
		assert.Equal(t, acc.GetState(), acc2.GetState())
		assert.Equal(t, acc.GetPersistenceTimestamp(), acc2.GetPersistenceTimestamp())

		// get all accusations without filters
		accs, err = db.GetAccusations(txn, nil)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(accs))
		assert.Equal(t, acc.GetID(), accs[0].GetID())

		// get persisted but unscheduled accusations
		accs, err = db.GetPersistedButUnscheduledAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		// get scheduled but incomplete accusations
		accs, err = db.GetScheduledButIncompleteAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		// get completed accusations
		accs, err = db.GetCompletedAccusations(txn)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(accs))
		assert.Equal(t, acc.GetID(), accs[0].GetID())

		return nil
	})

	assert.Nil(t, err)
}

func TestPersistMultipleAccusations(t *testing.T) {
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

	var idA [32]byte
	copy(idA[:], crypto.Hasher([]byte("idA")))
	accA := &AccusationTest2{}
	accA.SetID(idA)
	accA.SetPersistenceTimestamp(uint64(time.Now().Unix()))
	accA.SetState(objs.Persisted)

	var idB [32]byte
	copy(idB[:], crypto.Hasher([]byte("idB")))
	accB := &AccusationTest2{}
	accB.SetID(idB)
	accB.SetPersistenceTimestamp(uint64(time.Now().Unix() + 300))
	accB.SetState(objs.Persisted)
	accusations := make(map[[32]byte]objs.Accusation)
	accusations[idA] = accA
	accusations[idB] = accB

	err = db.Update(func(txn *badger.Txn) error {
		for _, acc := range accusations {
			err = db.SetAccusation(txn, acc)
			assert.Nil(t, err)
		}

		// check the retrieved accusation has the same values as the original
		accA2, err := db.GetAccusation(txn, idA)
		assert.Nil(t, err)
		assert.Equal(t, accA.GetID(), accA2.GetID())
		assert.Equal(t, accA.GetState(), accA2.GetState())
		assert.Equal(t, accA.GetPersistenceTimestamp(), accA2.GetPersistenceTimestamp())

		// check the retrieved accusation has the same values as the original
		accB2, err := db.GetAccusation(txn, idB)
		assert.Nil(t, err)
		assert.Equal(t, accB.GetID(), accB2.GetID())
		assert.Equal(t, accB.GetState(), accB2.GetState())
		assert.Equal(t, accB.GetPersistenceTimestamp(), accB2.GetPersistenceTimestamp())

		// get all accusations without filters
		accs, err := db.GetAccusations(txn, nil)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(accs))
		for _, acc := range accs {
			a, ok := accusations[acc.GetID()]
			assert.True(t, ok)
			assert.Equal(t, a.GetID(), acc.GetID())
		}

		// get persisted but unsheduled accusations
		accs, err = db.GetPersistedButUnscheduledAccusations(txn)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(accs))
		for _, acc := range accs {
			a, ok := accusations[acc.GetID()]
			assert.True(t, ok)
			assert.Equal(t, a.GetID(), acc.GetID())
		}

		// get scheduled but incomplete accusations
		accs, err = db.GetScheduledButIncompleteAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		// get completed accusations
		accs, err = db.GetCompletedAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		//////////////
		// set accusation state to ScheduledForExecution and persist it
		//////////////
		accA.SetState(objs.ScheduledForExecution)
		err = db.SetAccusation(txn, accA)
		assert.Nil(t, err)

		// check the retrieved accusation has the same values as the original
		accA2, err = db.GetAccusation(txn, idA)
		assert.Nil(t, err)
		assert.Equal(t, accA.GetID(), accA2.GetID())
		assert.Equal(t, accA.GetState(), accA2.GetState())
		assert.Equal(t, accA.GetPersistenceTimestamp(), accA2.GetPersistenceTimestamp())

		// get all accusations without filters
		accs, err = db.GetAccusations(txn, nil)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(accs))
		for _, acc := range accs {
			a, ok := accusations[acc.GetID()]
			assert.True(t, ok)
			assert.Equal(t, a.GetID(), acc.GetID())
		}

		// get persisted but unscheduled accusations
		accs, err = db.GetPersistedButUnscheduledAccusations(txn)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(accs))
		assert.Equal(t, accB.GetID(), accs[0].GetID())

		// get scheduled but incomplete accusations
		accs, err = db.GetScheduledButIncompleteAccusations(txn)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(accs))
		assert.Equal(t, accA.GetID(), accs[0].GetID())

		// get completed accusations
		accs, err = db.GetCompletedAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		//////////////
		// set accusation state to Completed and persist it
		//////////////
		accA.SetState(objs.Completed)
		err = db.SetAccusation(txn, accA)
		assert.Nil(t, err)

		// check the retrieved accusation has the same values as the original
		accA2, err = db.GetAccusation(txn, idA)
		assert.Nil(t, err)
		assert.Equal(t, accA.GetID(), accA2.GetID())
		assert.Equal(t, accA.GetState(), accA2.GetState())
		assert.Equal(t, accA.GetPersistenceTimestamp(), accA2.GetPersistenceTimestamp())

		// get all accusations without filters
		accs, err = db.GetAccusations(txn, nil)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(accs))
		for _, acc := range accs {
			a, ok := accusations[acc.GetID()]
			assert.True(t, ok)
			assert.Equal(t, a.GetID(), acc.GetID())
		}

		// get persisted but unscheduled accusations
		accs, err = db.GetPersistedButUnscheduledAccusations(txn)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(accs))
		assert.Equal(t, accB.GetID(), accs[0].GetID())

		// get scheduled but incomplete accusations
		accs, err = db.GetScheduledButIncompleteAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		// get completed accusations
		accs, err = db.GetCompletedAccusations(txn)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(accs))
		assert.Equal(t, accA.GetID(), accs[0].GetID())

		return nil
	})

	assert.Nil(t, err)
}

func TestPersistEmptyAccusationDB(t *testing.T) {
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

	err = db.View(func(txn *badger.Txn) error {

		// get all accusations without filters
		accs, err := db.GetAccusations(txn, nil)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		// get persisted but unsheduled accusations
		accs, err = db.GetPersistedButUnscheduledAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		// get scheduled but incomplete accusations
		accs, err = db.GetScheduledButIncompleteAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		// get completed accusations
		accs, err = db.GetCompletedAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		return nil
	})

	assert.Nil(t, err)
}
