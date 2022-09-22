package accusation

import (
	"context"
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/dgraph-io/badger/v2"
	"testing"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/utils"
	"github.com/stretchr/testify/assert"
)

func setupInvalidUTXOConsumptionAccusationTest(t *testing.T) *db.Database {
	ctx := context.Background()
	nodeCtx, cf := context.WithCancel(ctx)
	t.Cleanup(cf)

	// Initialize consensus db: stores all state the consensus mechanism requires to work
	rawConsensusDb, err := utils.OpenBadger(nodeCtx.Done(), "", true)
	assert.Nil(t, err)
	var closeDB func() = func() {
		err := rawConsensusDb.Close()
		if err != nil {
			t.Errorf("error closing rawConsensusDb: %v", err)
		}
	}
	t.Cleanup(closeDB)

	db := &db.Database{}
	db.Init(rawConsensusDb)

	return db
}

// TestNoMultipleProposalBehavior tests no bad behavior detected
func TestNoInvalidUTXOConsumptionBehavior(t *testing.T) {
	consDB := setupInvalidUTXOConsumptionAccusationTest(t)
	_ = GenerateAccusationsForInvalidUTXO(t, consDB)
	sstore := lstate.New(consDB)

	err := consDB.View(func(txn *badger.Txn) error {
		//os, err := consDB.GetOwnState(txn)
		//if err != nil {
		//	return err
		//}
		//
		//rs, err := consDB.GetCurrentRoundState(txn, os.VAddr)
		//if err != nil {
		//	return err
		//}

		rss, err := sstore.LoadLocalState(txn)
		if err != nil {
			return err
		}

		//assert.NotNil(t, os)
		//assert.NotNil(t, rs)
		assert.NotNil(t, rss)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
