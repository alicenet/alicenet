package state

import (
	"encoding/json"
	"fmt"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
)

// asserting that SnapshotState struct implements interface tasks.ITaskState
var _ tasks.TaskState = &SnapshotState{}

type SnapshotState struct {
	Account            accounts.Account
	RawBClaims         []byte
	RawSigGroup        []byte
	BlockHeader        *objs.BlockHeader
	LastSnapshotHeight int
	DesperationDelay   int
	DesperationFactor  int
	RandomSeedHash     []byte
}

func (state *SnapshotState) PersistState(txn *badger.Txn) error {
	logger := logging.GetLogger("staterecover").WithField("State", "snapshotState")
	rawData, err := json.Marshal(state)
	if err != nil {
		return err
	}

	key := dbprefix.PrefixEthereumSnapshotState()
	logger.WithField("Key", string(key)).Debug("Saving state in the database")
	if err = utils.SetValue(txn, key, rawData); err != nil {
		return err
	}

	return nil
}

func (state *SnapshotState) LoadState(txn *badger.Txn) error {
	logger := logging.GetLogger("staterecover").WithField("State", "snapshotState")
	key := dbprefix.PrefixEthereumSnapshotState()
	logger.WithField("Key", string(key)).Debug("Loading state from database")
	rawData, err := utils.GetValue(txn, key)
	if err != nil {
		return err
	}

	err = json.Unmarshal(rawData, state)
	if err != nil {
		return err
	}

	return nil

}

func GetSnapshotState(monDB *db.Database) (*SnapshotState, error) {
	snapshotState := &SnapshotState{}
	err := monDB.View(func(txn *badger.Txn) error {
		return snapshotState.LoadState(txn)
	})
	if err != nil {
		return nil, err
	}
	return snapshotState, nil
}

func SaveSnapshotState(monDB *db.Database, snapshotState *SnapshotState) error {
	err := monDB.Update(func(txn *badger.Txn) error {
		return snapshotState.PersistState(txn)
	})
	if err != nil {
		return err
	}
	if err = monDB.Sync(); err != nil {
		return fmt.Errorf("Failed to set sync of snapshotState: %v", err)
	}
	return nil
}
