package lstate

import (
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/dgraph-io/badger/v2"
)

type Store struct {
	database db.DatabaseIface
}

func (ss *Store) Init(database db.DatabaseIface) error {
	ss.database = database
	return nil
}

func (ss *Store) LoadLocalState(txn *badger.Txn) (*RoundStates, error) {
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

func (ss *Store) LoadState(txn *badger.Txn, rcert *objs.RCert) (*RoundStates, error) {
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

func (ss *Store) WriteState(rs *RoundStates) error {
	err := ss.database.SetOwnState(rs.txn, rs.OwnState)
	if err != nil {
		return err
	}
	for _, valObj := range rs.ValidatorSet.Validators {
		r := rs.PeerStateMap[string(valObj.VAddr)]
		if err := ss.database.SetCurrentRoundState(rs.txn, r); err != nil {
			return err
		}
	}
	if err := ss.database.SetCurrentRoundState(rs.txn, rs.OwnRoundState()); err != nil {
		return err
	}
	err = ss.database.SetOwnValidatingState(rs.txn, rs.OwnValidatingState)
	if err != nil {
		return err
	}
	var conflictCheckerValue *objs.Proposal
	if rs.LockedValueCurrent() {
		conflictCheckerValue = rs.LockedValue()
	} else if rs.ValidValueCurrent() {
		conflictCheckerValue = rs.ValidValue()
	}
	for _, vobj := range rs.ValidatorSet.Validators {
		vAddr := vobj.VAddr
		s := rs.PeerStateMap[string(vAddr)]
		if conflictCheckerValue != nil {
			s.TrackExternalConflicts(conflictCheckerValue)
		}
		if err := ss.database.SetCurrentRoundState(rs.txn, s); err != nil {
			return err
		}
	}
	return nil
}

// GetDropData ...
func (ss *Store) GetDropData(txn *badger.Txn) (isValidator bool, isSync bool, chainID uint32, height uint32, round uint32, err error) {
	ownState, err := ss.database.GetOwnState(txn)
	if err != nil {
		return isValidator, isSync, chainID, height, round, err
	}
	chainID = ownState.SyncToBH.BClaims.ChainID
	height = ownState.SyncToBH.BClaims.Height
	isSync = ownState.IsSync()
	vs, err := ss.database.GetValidatorSet(txn, ownState.SyncToBH.BClaims.Height)
	if err != nil {
		return isValidator, isSync, chainID, height, round, err
	}
	if vs.ValidatorVAddrSet[string(ownState.VAddr)] {
		isValidator = true
	} else {
		isValidator = false
	}
	if isValidator {
		rs, err := ss.database.GetCurrentRoundState(txn, ownState.VAddr)
		if err != nil {
			return isValidator, isSync, chainID, height, round, err
		}
		round = rs.RCert.RClaims.Round
	} else {
		round = 1
	}
	return isValidator, isSync, chainID, height, round, err
}

// GetGossipValues ...
func (ss *Store) GetGossipValues() (*objs.Proposal, *objs.PreVote, *objs.PreVoteNil, *objs.PreCommit, *objs.PreCommitNil, *objs.NextRound, *objs.NextHeight, error) {
	var p *objs.Proposal
	var pv *objs.PreVote
	var pvn *objs.PreVoteNil
	var pc *objs.PreCommit
	var pcn *objs.PreCommitNil
	var nr *objs.NextRound
	var nh *objs.NextHeight

	err := ss.database.View(func(txn *badger.Txn) error {
		var err error

		p, err = ss.database.GetBroadcastProposal(txn)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			p = nil
		}

		pv, err = ss.database.GetBroadcastPreVote(txn)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			pv = nil
		}

		pvn, err = ss.database.GetBroadcastPreVoteNil(txn)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			pvn = nil
		}

		pc, err = ss.database.GetBroadcastPreCommit(txn)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			pc = nil
		}

		pcn, err = ss.database.GetBroadcastPreCommitNil(txn)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			pcn = nil
		}

		nr, err = ss.database.GetBroadcastNextRound(txn)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			nr = nil
		}

		nh, err = ss.database.GetBroadcastNextHeight(txn)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			nh = nil
		}

		return nil
	})
	return p, pv, pvn, pc, pcn, nr, nh, err
}

func (ss *Store) GetSyncToBH(txn *badger.Txn) (*objs.BlockHeader, error) {
	os, err := ss.database.GetOwnState(txn)
	if err != nil {
		return nil, err
	}
	return os.SyncToBH, nil
}

func (ss *Store) GetMaxBH(txn *badger.Txn) (*objs.BlockHeader, error) {
	os, err := ss.database.GetOwnState(txn)
	if err != nil {
		return nil, err
	}
	return os.MaxBHSeen, nil
}

func (ss *Store) IsSync(txn *badger.Txn) (bool, error) {
	mbhs, err := ss.GetMaxBH(txn)
	if err != nil {
		return false, err
	}
	stbh, err := ss.GetSyncToBH(txn)
	if err != nil {
		return false, err
	}
	if objs.RelateH(mbhs, stbh) == 0 {
		return true, nil
	}
	return false, nil
}
