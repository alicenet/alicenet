package lstate

import (
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/dgraph-io/badger/v2"
)

// MockStore .
type MockStore struct {
	database *db.MockDatabaseIface
}

func NewMockStore(mdb *db.MockDatabaseIface) *MockStore {
	return &MockStore{database: mdb}
}

func (ss *MockStore) LoadLocalState(txn *badger.Txn) (*RoundStates, error) {
	ownState, err := ss.database.GetOwnState(txn)
	if err != nil {
		return nil, err
	}
	rs, err := ss.database.GetCurrentRoundState(txn, ownState.VAddr)
	if err != nil {
		return nil, err
	}
	return ss.LoadState(txn, rs.RCert)
}

func (ss *MockStore) LoadState(txn *badger.Txn, rcert *objs.RCert) (*RoundStates, error) {
	rs := &RoundStates{
		height:       rcert.RClaims.Height,
		round:        rcert.RClaims.Round,
		PeerStateMap: make(map[string]*objs.RoundState),
	}
	validatorSet, err := ss.database.GetValidatorSet(txn, rcert.RClaims.Height)
	if err != nil {
		return nil, err
	}
	rs.ValidatorSet = validatorSet
	ownState, err := ss.database.GetOwnState(txn)
	if err != nil {
		return nil, err
	}
	rs.OwnState = ownState
	rstate, err := ss.database.GetCurrentRoundState(txn, ownState.VAddr)
	if err != nil {
		return nil, err
	}
	if rs.round == 0 {
		rs.round = rstate.RCert.RClaims.Round
	}
	rs.PeerStateMap[string(ownState.VAddr)] = rstate
	groupKey := rs.ValidatorSet.GroupKey
	for idx := 0; idx < len(rs.ValidatorSet.Validators); idx++ {
		val := rs.ValidatorSet.Validators[idx].VAddr
		rstate, err := ss.database.GetCurrentRoundState(txn, val)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return nil, err
			}
		}
		if rstate == nil {
			rcBytes, err := rcert.MarshalBinary()
			if err != nil {
				return nil, err
			}
			rc := &objs.RCert{}
			err = rc.UnmarshalBinary(rcBytes)
			if err != nil {
				return nil, err
			}
			groupShare := rs.ValidatorSet.Validators[idx].GroupShare
			rstate = &objs.RoundState{
				VAddr:      val,
				GroupKey:   groupKey,
				GroupShare: groupShare,
				GroupIdx:   uint8(idx),
				RCert:      rc,
			}
		}
		if rs.round == 0 {
			if rs.IsMe(val) {
				rs.round = rstate.RCert.RClaims.Round
			}
		}
		rs.PeerStateMap[string(val)] = rstate
	}
	ownValidatingState, err := ss.database.GetOwnValidatingState(txn)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return nil, err
		}
	}
	rs.OwnValidatingState = ownValidatingState
	return rs, nil
}
