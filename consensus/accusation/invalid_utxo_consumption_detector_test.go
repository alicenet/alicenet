package accusation

import (
	"context"
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
	setupInvalidUTXOConsumptionAccusationTest(t)
}
