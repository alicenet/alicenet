package state

import (
	"encoding/json"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants/dbprefix"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
)

// asserting that SnapshotState struct implements interface interfaces.ITaskState
var _ interfaces.ITaskState = &SnapshotState{}

type SnapshotState struct {
	Account     accounts.Account
	RawBClaims  []byte
	RawSigGroup []byte
	BlockHeader *objs.BlockHeader
}

func (state *SnapshotState) PersistState(txn *badger.Txn) error {
	rawData, err := json.Marshal(state)
	if err != nil {
		return err
	}

	key := dbprefix.PrefixSnapshotState()
	if err = utils.SetValue(txn, key, rawData); err != nil {
		return err
	}

	return nil
}

func (state *SnapshotState) LoadState(txn *badger.Txn) error {
	key := dbprefix.PrefixSnapshotState()
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
