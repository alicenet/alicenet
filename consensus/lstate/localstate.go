package lstate

import (
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
)

// this is the bottom of call stack all methods in this file are the setters
// for votes on the local state.

func (ce *Engine) setMostRecentRCert(rs *RoundStates, v *objs.RCert) error {
	rs.OwnValidatingState.SetRoundStarted()
	if err := rs.OwnRoundState().SetRCert(v); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

func (ce *Engine) setMostRecentProposal(rs *RoundStates, v *objs.Proposal) error {
	ok, err := rs.OwnRoundState().SetProposal(v)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if !ok {
		return errorz.ErrCorrupt
	}
	return nil
}

func (ce *Engine) setMostRecentPreVote(rs *RoundStates, v *objs.PreVote) error {
	rs.OwnValidatingState.SetPreVoteStepStarted()
	ok, err := rs.OwnRoundState().SetPreVote(v)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if !ok {
		return errorz.ErrCorrupt
	}
	return nil
}

func (ce *Engine) setMostRecentPreVoteNil(rs *RoundStates, v *objs.PreVoteNil) error {
	rs.OwnValidatingState.SetPreVoteStepStarted()
	ok, err := rs.OwnRoundState().SetPreVoteNil(v)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if !ok {
		return errorz.ErrCorrupt
	}
	return nil
}

func (ce *Engine) setMostRecentPreCommit(rs *RoundStates, v *objs.PreCommit) error {
	rs.OwnValidatingState.SetPreCommitStepStarted()
	ok, err := rs.OwnRoundState().SetPreCommit(v)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if !ok {
		return errorz.ErrCorrupt
	}
	if err := ce.setMostRecentLockedValue(rs, v.Proposal); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

func (ce *Engine) setMostRecentPreCommitNil(rs *RoundStates, v *objs.PreCommitNil) error {
	rs.OwnValidatingState.SetPreCommitStepStarted()
	ok, err := rs.OwnRoundState().SetPreCommitNil(v)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if !ok {
		return errorz.ErrCorrupt
	}
	return nil
}

func (ce *Engine) setMostRecentNextRound(rs *RoundStates, v *objs.NextRound) error {
	ok, err := rs.OwnRoundState().SetNextRound(v)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if !ok {
		return errorz.ErrCorrupt
	}
	return nil
}

func (ce *Engine) setMostRecentNextHeight(rs *RoundStates, v *objs.NextHeight) error {
	ok, err := rs.OwnRoundState().SetNextHeight(v)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if !ok {
		return errorz.ErrCorrupt
	}
	if err := ce.setMostRecentLockedValue(rs, v.NHClaims.Proposal); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

func (ce *Engine) setMostRecentLockedValue(rs *RoundStates, v *objs.Proposal) error {
	rs.OwnValidatingState.ValidValue = v
	rs.OwnValidatingState.LockedValue = v
	return nil
}

func (ce *Engine) setMostRecentValidValue(rs *RoundStates, v *objs.Proposal) error {
	rs.OwnValidatingState.ValidValue = v
	return nil
}

func (ce *Engine) setMostRecentBlockHeader(rs *RoundStates, v *objs.BlockHeader) error {
	// ce.sstore.WriteState()
	// not really sure what to do for this and other functions here that don't pass rs
	// using rs.txn for now
	if err := ce.applyState(rs, v.BClaims.ChainID, v.TxHshLst); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	rc, err := v.GetRCert()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentRCert(rs, rc); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.database.SetCommittedBlockHeader(rs.txn, v); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	rs.OwnState.SyncToBH = v
	if rs.OwnState.MaxBHSeen.BClaims.Height < rs.OwnState.SyncToBH.BClaims.Height {
		rs.OwnState.MaxBHSeen = rs.OwnState.SyncToBH
	}
	if v.BClaims.Height%constants.EpochLength == 0 {
		rs.OwnState.CanonicalSnapShot = rs.OwnState.PendingSnapShot
		rs.OwnState.PendingSnapShot = v
	}
	if err := ce.database.SetBroadcastBlockHeader(rs.txn, v); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

func (ce *Engine) setMostRecentBlockHeaderFastSync(rs *RoundStates, v *objs.BlockHeader) error {
	rc, err := v.GetRCert()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentRCert(rs, rc); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	rs.OwnState.SyncToBH = v
	if rs.OwnState.MaxBHSeen.BClaims.Height < rs.OwnState.SyncToBH.BClaims.Height {
		rs.OwnState.MaxBHSeen = rs.OwnState.SyncToBH
	}
	if v.BClaims.Height%constants.EpochLength == 0 {
		rs.OwnState.CanonicalSnapShot = rs.OwnState.PendingSnapShot
		rs.OwnState.PendingSnapShot = v
	}
	return nil
}
