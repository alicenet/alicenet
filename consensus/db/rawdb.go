package db

import (
	"context"
	"sync"
	"time"

	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

type Txn struct {
	sync.Mutex
	badger.Txn
	callbacks []TxnFunc
}

func (txn *Txn) AddCallback(fn TxnFunc) {
	txn.Lock()
	defer txn.Unlock()
	if txn.callbacks == nil {
		txn.callbacks = []TxnFunc{}
	}
	txn.callbacks = append(txn.callbacks, fn)
}

type TxnFunc func(txn *badger.Txn) error

type rawDataBase struct {
	db     *badger.DB
	logger *logrus.Logger
}

func (db *rawDataBase) View(fn func(txn *badger.Txn) error) error {
	return db.db.View(fn)
}

func (db *rawDataBase) Update(fn func(txn *badger.Txn) error) error {
	txn := db.db.NewTransaction(true)
	defer txn.Discard()
	if err := fn(txn); err != nil {
		return err
	}
	err := txn.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (db *rawDataBase) Sync() error {
	return db.db.Sync()
}

func (db *rawDataBase) GarbageCollect() error {
	db.db.RunValueLogGC(constants.BadgerDiscardRatio)
	db.db.RunValueLogGC(constants.BadgerDiscardRatio)
	return nil
}

func (db *rawDataBase) DropPrefix(k []byte) error {
	return db.db.DropPrefix(k)
}

// subscribe to prefix is used to form the proposal subscription
func (db *rawDataBase) subscribeToPrefix(ctx context.Context, prefix []byte, cb func([]byte) error) {
	fn := func(kvs *badger.KVList) error {
		for i := 0; i < len(kvs.Kv); i++ {
			kv := kvs.Kv[i]
			if kv.Value != nil {
				err := cb(kv.Value)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	fn2 := func() {
		err := db.db.Subscribe(ctx, fn, prefix)
		if err != nil && err != context.Canceled {
			db.logger.Warnf("terminating db subscription for prefix: %v", prefix)
		}
	}
	go fn2()
}

func (db *rawDataBase) getValue(txn *badger.Txn, key []byte) ([]byte, error) {
	return utils.GetValue(txn, key)
}

func (db *rawDataBase) SetValue(txn *badger.Txn, key []byte, value []byte) error {
	return utils.SetValue(txn, key, value)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) SetOwnValidatingState(txn *badger.Txn, key []byte, v *objs.OwnValidatingState) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetOwnValidatingState(txn *badger.Txn, key []byte) (*objs.OwnValidatingState, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	vv := &objs.OwnValidatingState{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) SetEncryptedStore(txn *badger.Txn, key []byte, v *objs.EncryptedStore) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetEncryptedStore(txn *badger.Txn, key []byte) (*objs.EncryptedStore, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	vv := &objs.EncryptedStore{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) SetOwnState(txn *badger.Txn, key []byte, v *objs.OwnState) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetOwnState(txn *badger.Txn, key []byte) (*objs.OwnState, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	vv := &objs.OwnState{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) SetRoundState(txn *badger.Txn, key []byte, v *objs.RoundState) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetRoundState(txn *badger.Txn, key []byte) (*objs.RoundState, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	vv := &objs.RoundState{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) SetValidatorSet(txn *badger.Txn, key []byte, v *objs.ValidatorSet) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetValidatorSet(txn *badger.Txn, key []byte) (*objs.ValidatorSet, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	vv := &objs.ValidatorSet{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) SetBlockHeader(txn *badger.Txn, key []byte, v *objs.BlockHeader) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetBlockHeader(txn *badger.Txn, key []byte) (*objs.BlockHeader, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	vv := &objs.BlockHeader{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) SetRCert(txn *badger.Txn, key []byte, v *objs.RCert) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetRCert(txn *badger.Txn, key []byte) (*objs.RCert, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	vv := &objs.RCert{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) SetProposal(txn *badger.Txn, key []byte, v *objs.Proposal) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetProposal(txn *badger.Txn, key []byte) (*objs.Proposal, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	vv := &objs.Proposal{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) SetPreVote(txn *badger.Txn, key []byte, v *objs.PreVote) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetPreVote(txn *badger.Txn, key []byte) (*objs.PreVote, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	vv := &objs.PreVote{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

func (db *rawDataBase) SetPreVoteNil(txn *badger.Txn, key []byte, v *objs.PreVoteNil) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetPreVoteNil(txn *badger.Txn, key []byte) (*objs.PreVoteNil, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	vv := &objs.PreVoteNil{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) SetPreCommit(txn *badger.Txn, key []byte, v *objs.PreCommit) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetPreCommit(txn *badger.Txn, key []byte) (*objs.PreCommit, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	vv := &objs.PreCommit{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

func (db *rawDataBase) SetPreCommitNil(txn *badger.Txn, key []byte, v *objs.PreCommitNil) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetPreCommitNil(txn *badger.Txn, key []byte) (*objs.PreCommitNil, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	vv := &objs.PreCommitNil{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) SetNextHeight(txn *badger.Txn, key []byte, v *objs.NextHeight) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetNextHeight(txn *badger.Txn, key []byte) (*objs.NextHeight, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	vv := &objs.NextHeight{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

func (db *rawDataBase) SetNextRound(txn *badger.Txn, key []byte, v *objs.NextRound) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetNextRound(txn *badger.Txn, key []byte) (*objs.NextRound, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	vv := &objs.NextRound{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) SetU32(txn *badger.Txn, key []byte, v uint32) error {
	vv := utils.MarshalUint32(v)
	return utils.SetValue(txn, key, vv)
}

func (db *rawDataBase) GetU32(txn *badger.Txn, key []byte) (uint32, error) {
	vv, err := db.getValue(txn, key)
	if err != nil {
		return 0, err
	}
	return utils.UnmarshalUint32(vv)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) GetInt(txn *badger.Txn, key []byte) (int64, error) {
	v, err := db.getValue(txn, key)
	if err != nil {
		return 0, err
	}
	if v == nil {
		return 0, nil
	}
	vv, err := utils.UnmarshalInt64(v)
	if err != nil {
		return 0, err
	}
	return vv, nil
}

func (db *rawDataBase) SetInt(txn *badger.Txn, key []byte, v int) error {
	tbytes := utils.MarshalInt64(int64(v))
	return utils.SetValue(txn, key, tbytes)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) GetDuration(txn *badger.Txn, key []byte) (time.Duration, error) {
	v, err := db.GetInt(txn, key)
	if err != nil {
		return 0, err
	}
	return time.Duration(v), nil
}

func (db *rawDataBase) SetDuration(txn *badger.Txn, key []byte, v time.Duration) error {
	return db.SetInt(txn, key, int(v))
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) GetTime(txn *badger.Txn, key []byte) (time.Time, error) {
	t := &time.Time{}
	v, err := db.getValue(txn, key)
	if err != nil {
		return time.Time{}, err
	}
	err = t.UnmarshalBinary(v)
	if err != nil {
		return time.Time{}, err
	}
	return *t, nil
}

func (db *rawDataBase) SetTime(txn *badger.Txn, key []byte, t time.Time) error {
	val, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, val)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *rawDataBase) incrementCounter(txn *badger.Txn, k []byte) error {
	v, err := db.getCounter(txn, k)
	if err != nil {
		return err
	}
	v++
	return db.SetInt(txn, k, v)
}

func (db *rawDataBase) decrementCounter(txn *badger.Txn, k []byte) error {
	v, err := db.getCounter(txn, k)
	if err != nil {
		return err
	}
	v--
	return db.SetInt(txn, k, v)
}

func (db *rawDataBase) zeroCounter(txn *badger.Txn, k []byte) error {
	return db.SetInt(txn, k, 0)
}

func (db *rawDataBase) getCounter(txn *badger.Txn, k []byte) (int, error) {
	v, err := db.GetInt(txn, k)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return 0, err
		}
		v = 0
	}
	return int(v), nil
}
