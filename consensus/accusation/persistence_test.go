package accusation

import (
	"context"
	"encoding/gob"
	"testing"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/layer1/executor/marshaller"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

type AccusationTest1 struct {
	*tasks.BaseTask
}

var _ tasks.Task = &AccusationTest1{}

func NewAccusationTest1Task(id string) tasks.Task {
	t := &AccusationTest1{
		BaseTask: tasks.NewBaseTask(0, 0, true, nil),
	}
	t.Id = id
	return t
}

func (t *AccusationTest1) Prepare(ctx context.Context) *tasks.TaskErr {
	return nil
}

func (t *AccusationTest1) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	return nil, nil
}

func (t *AccusationTest1) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	return true, nil
}

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

	var idString string = common.Bytes2Hex(crypto.Hasher([]byte("test")))
	acc := NewAccusationTest1Task(idString)
	accRaw, err := marshaller.GobMarshalBinary(acc)
	assert.NotNil(t, err)
	assert.Empty(t, accRaw)
}

///////////////////////////////////

type AccusationTest2 struct {
	*tasks.BaseTask
}

var _ tasks.Task = &AccusationTest2{}

func NewAccusationTest2Task(id string) tasks.Task {
	t := &AccusationTest2{
		BaseTask: tasks.NewBaseTask(0, 0, true, nil),
	}
	t.Id = id
	return t
}

func (t *AccusationTest2) Prepare(ctx context.Context) *tasks.TaskErr {
	return nil
}

func (t *AccusationTest2) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	return nil, nil
}

func (t *AccusationTest2) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	return true, nil
}

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

	var idString string = common.Bytes2Hex(crypto.Hasher([]byte("test")))
	var id [32]byte = utils.HexToBytes32(idString)
	acc := NewAccusationTest2Task(idString)
	accRaw, err := marshaller.GobMarshalBinary(acc)
	assert.Nil(t, err)

	err = db.Update(func(txn *badger.Txn) error {
		err := db.SetAccusationRaw(txn, id, accRaw)
		assert.Nil(t, err)

		acc2Raw, err := db.GetAccusationRaw(txn, id)
		assert.Nil(t, err)
		assert.NotEmpty(t, acc2Raw)
		acc2, err := marshaller.GobUnmarshalBinary(acc2Raw)
		assert.Nil(t, err)
		assert.Equal(t, acc.GetId(), acc2.GetId())

		acc3, ok := acc2.(*AccusationTest2)
		assert.True(t, ok)
		assert.Equal(t, acc.GetId(), acc3.GetId())

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

	var idString string = common.Bytes2Hex(crypto.Hasher([]byte("test")))
	var id [32]byte = utils.HexToBytes32(idString)
	acc := NewAccusationTest2Task(idString)
	accRaw, err := marshaller.GobMarshalBinary(acc)
	assert.Nil(t, err)

	err = db.Update(func(txn *badger.Txn) error {
		err := db.SetAccusationRaw(txn, id, accRaw)
		assert.Nil(t, err)

		// check the retrieved accusation has the same values as the original
		acc2Raw, err := db.GetAccusationRaw(txn, id)
		assert.Nil(t, err)
		assert.NotEmpty(t, acc2Raw)
		acc2, err := marshaller.GobUnmarshalBinary(acc2Raw)
		assert.Nil(t, err)
		assert.Equal(t, acc.GetId(), acc2.GetId())

		// check the retrieved accusation is of type AccusationTest2
		acc3, ok := acc2.(*AccusationTest2)
		assert.True(t, ok)
		assert.Equal(t, acc.GetId(), acc3.GetId())

		// get all accusations
		accs, err := db.GetAccusations(txn)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(accs))
		accs0, err := marshaller.GobUnmarshalBinary(accs[0])
		assert.Nil(t, err)
		assert.Equal(t, acc.GetId(), accs0.GetId())

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

	var idAString string = common.Bytes2Hex(crypto.Hasher([]byte("idA")))
	var idA [32]byte = utils.HexToBytes32(idAString)
	accA := NewAccusationTest2Task(idAString)

	var idBString string = common.Bytes2Hex(crypto.Hasher([]byte("idB")))
	var idB [32]byte = utils.HexToBytes32(idBString)
	accB := NewAccusationTest2Task(idBString)

	accusations := make(map[[32]byte]tasks.Task)
	accusations[idA] = accA
	accusations[idB] = accB

	err = db.Update(func(txn *badger.Txn) error {
		for id, acc := range accusations {
			accRaw, err := marshaller.GobMarshalBinary(acc)
			assert.Nil(t, err)
			err = db.SetAccusationRaw(txn, id, accRaw)
			assert.Nil(t, err)
		}

		// check the retrieved accusation has the same values as the original
		accA2Raw, err := db.GetAccusationRaw(txn, idA)
		assert.Nil(t, err)
		accA2, err := marshaller.GobUnmarshalBinary(accA2Raw)
		assert.Nil(t, err)
		assert.Equal(t, accA.GetId(), accA2.GetId())

		// check the retrieved accusation has the same values as the original
		accB2Raw, err := db.GetAccusationRaw(txn, idB)
		assert.Nil(t, err)
		accB2, err := marshaller.GobUnmarshalBinary(accB2Raw)
		assert.Nil(t, err)
		assert.Equal(t, accB.GetId(), accB2.GetId())

		// get all accusations
		accs, err := db.GetAccusations(txn)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(accs))
		for _, accRaw := range accs {
			acc, err := marshaller.GobUnmarshalBinary(accRaw)
			assert.Nil(t, err)

			var id [32]byte = utils.HexToBytes32(acc.GetId())
			a, ok := accusations[id]
			assert.True(t, ok)
			assert.Equal(t, a.GetId(), acc.GetId())
		}

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
		accs, err := db.GetAccusations(txn)
		assert.Nil(t, err)
		assert.Empty(t, accs)

		return nil
	})

	assert.Nil(t, err)
}
