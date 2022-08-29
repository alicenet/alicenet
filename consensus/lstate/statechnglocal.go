package lstate

import (
	"bytes"
	"errors"

	"github.com/dgraph-io/badger/v2"

	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/utils"
)

// These are the step handlers. They figure out how to take an action based on
// what action is determined as necessary in updateLocalStateInternal

func (ce *Engine) doPendingProposalStep(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	ce.logger.Debugf("doPendingProposalStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v	GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])
	os := rs.OwnRoundState()
	rcert := os.RCert
	if rcert.RClaims.Round == constants.DEADBLOCKROUND {
		return nil
	}
	// if not locked or valid form new proposal
	if !rs.LockedValueCurrent() && !rs.ValidValueCurrent() { // 00 case
		if err := ce.castNewProposalValue(txn, rs); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		return nil
	}
	// if not locked but valid known, propose valid value
	if !rs.LockedValueCurrent() && rs.ValidValueCurrent() { // 01 case
		if err := ce.castProposalFromValue(txn, rs, rs.ValidValue()); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		return nil
	}
	// if locked, propose locked
	// 10
	// 11 case
	if err := ce.castProposalFromValue(txn, rs, rs.LockedValue()); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doPendingPreVoteStep(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	ce.logger.Debugf("doPendingPreVoteStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v	GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])
	os := rs.OwnRoundState()
	rcert := os.RCert
	if rcert.RClaims.Round == constants.DEADBLOCKROUND {
		// Safely form EmptyBlock PreVote
		rs.OwnValidatingState.ValidValue = nil
		rs.OwnValidatingState.LockedValue = nil
		TxRoot, err := objs.MakeTxRoot([][]byte{})
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		bclaims := rs.OwnState.SyncToBH.BClaims
		PrevBlock := utils.CopySlice(rcert.RClaims.PrevBlock)
		headerRoot, err := ce.database.GetHeaderTrieRoot(txn, rs.OwnState.SyncToBH.BClaims.Height)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		StateRoot := utils.CopySlice(bclaims.StateRoot)
		p := &objs.Proposal{
			PClaims: &objs.PClaims{
				BClaims: &objs.BClaims{
					PrevBlock:  PrevBlock,
					HeaderRoot: headerRoot,
					StateRoot:  StateRoot,
					TxRoot:     TxRoot,
					ChainID:    rcert.RClaims.ChainID,
					Height:     rcert.RClaims.Height,
				},
				RCert: rcert,
			},
		}
		if err := p.Sign(ce.secpSigner); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		if err := ce.castPreVote(txn, rs, p); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		return nil
	}
	p := rs.GetCurrentProposal()
	// proposal timeout hit
	// if there is a current proposal
	// Height, round, previous block hash
	if p != nil {
		// if we are not locked and there is no known valid value
		// check if the proposed value is valid, and if so
		// prevote this value
		// 00 case
		if !rs.LockedValueCurrent() && !rs.ValidValueCurrent() {
			txs, _, err := ce.dm.GetTxs(txn, p.PClaims.BClaims.Height, rs.round, p.TxHshLst)
			if err == nil {
				ok, err := ce.isValid(txn, rs, p.PClaims.BClaims.ChainID, p.PClaims.BClaims.StateRoot, p.PClaims.BClaims.HeaderRoot, txs)
				if err != nil {
					var e *errorz.ErrInvalid
					if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
						utils.DebugTrace(ce.logger, err)
						return err
					}
				}
				if ok { // proposal is valid
					if err := ce.castPreVote(txn, rs, p); err != nil {
						utils.DebugTrace(ce.logger, err)
						return err
					}
					return nil
				}
			} // proposal is not valid
			if err := ce.castPreVoteNil(txn, rs); err != nil {
				utils.DebugTrace(ce.logger, err)
				return err
			}
			return nil
		}
		// if we are locked on a valid value, only prevote the value if it is equal
		// to the lock
		// 01 case
		if !rs.LockedValueCurrent() && rs.ValidValueCurrent() {
			if err := ce.castPreVoteWithLock(txn, rs, rs.ValidValue(), p); err != nil {
				utils.DebugTrace(ce.logger, err)
				return err
			}
			return nil
		}
		// if we are locked on a locked value, only prevote the value if it is equal
		// to the lock
		// 10 case
		// 11 case
		if err := ce.castPreVoteWithLock(txn, rs, rs.LockedValue(), p); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		return nil
	} // no current proposal known
	// so prevote nil
	if err := ce.castPreVoteNil(txn, rs); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doPreVoteStep(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	// local node must have cast a preVote to get here
	// count the prevotes and prevote nils
	pvl, pvnl, err := rs.GetCurrentPreVotes()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	ce.logger.Debugf("doPreVoteStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v   PV:%v   PVN:%v	GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, len(pvl), len(pvnl), rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])
	// if we have enough prevotes, cast a precommit
	// this will update the locked value
	if len(pvl) >= rs.GetCurrentThreshold() {
		if err := ce.castPreCommit(txn, rs, pvl); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
	}
	// no thresholds met, so do nothing
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doPreVoteNilStep(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	// local node must have cast a preVote nil
	// count the preVotes and prevote nils
	pvl, pvnl, err := rs.GetCurrentPreVotes()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	ce.logger.Debugf("doPreVoteNilStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v   PV:%v   PVN:%v	GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, len(pvl), len(pvnl), rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])
	if len(pvl) >= rs.GetCurrentThreshold() {
		p, err := pvl.GetProposal()
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		err = ce.updateValidValue(txn, rs, p)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
	}
	// if greater than threshold prevote nils, cast the prevote nil
	if len(pvnl) >= rs.GetCurrentThreshold() {
		if err := ce.castPreCommitNil(txn, rs); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doPendingPreCommit(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	os := rs.OwnRoundState()
	rcert := os.RCert
	// prevote timeout hit with no clear consensus in either direction
	// during cycle before timeout
	// count the prevotes
	pvl, pvnl, err := rs.GetCurrentPreVotes()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	ce.logger.Debugf("doPendingPreCommit:    MAXBH:%v    STBH:%v    RH:%v    RN:%v   PV:%v   PVN:%v	GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, len(pvl), len(pvnl), rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])
	// if we have prevote consensus now
	if len(pvl) >= rs.GetCurrentThreshold() {
		// if we preVoted for the proposal then preCommit
		if rs.LocalPreVoteCurrent() {
			if err := ce.castPreCommit(txn, rs, pvl); err != nil {
				utils.DebugTrace(ce.logger, err)
				return err
			}
			return nil
		}
		// update the valid value
		p, err := pvl.GetProposal()
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		if err := ce.updateValidValue(txn, rs, p); err != nil {
			var e *errorz.ErrInvalid
			if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
				utils.DebugTrace(ce.logger, err)
				return err
			}
		}
	} // fallthrough to precommit nil
	// since the timeout has expired,
	// free to cast preCommitNil without
	// clear consensus if the total votes is
	// greater than threshold
	if rcert.RClaims.Round != constants.DEADBLOCKROUND {
		if err := ce.castPreCommitNil(txn, rs); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
	}
	// threshold not met as of yet
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doPreCommitStep(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	// local node cast a precommit this round
	// count the precommits
	pcl, pcnl, err := rs.GetCurrentPreCommits()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	ce.logger.Debugf("doPreCommitStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v   PC:%v   PCN:%v	GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, len(pcl), len(pcnl), rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])
	// if we have consensus and can verify
	// cast the nextHeight
	if len(pcl) >= rs.GetCurrentThreshold() {
		p, err := pcl.GetProposal()
		if err != nil {
			return err
		}
		if err := ce.updateValidValue(txn, rs, p); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		if err := ce.castNextHeightFromPreCommits(txn, rs, pcl); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
	}
	// no consensus, wait for more votes
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doPreCommitNilStep(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	// local node cast a precommit nil this round
	// count the precommits
	pcl, pcnl, err := rs.GetCurrentPreCommits()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	ce.logger.Debugf("doPreCommitNilStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v   PC:%v   PCN:%v	  GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, len(pcl), len(pcnl), rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])
	// if we have a preCommit consensus,
	// move directly to next height
	if len(pcl) >= rs.GetCurrentThreshold() {
		p, err := pcl.GetProposal()
		if err != nil {
			return err
		}
		if err := ce.updateValidValue(txn, rs, p); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		if err := ce.castNextHeightFromPreCommits(txn, rs, pcl); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		return nil
	}
	// if we have a consensus for a precommit nil,
	// cast a next round
	if len(pcnl) >= rs.GetCurrentThreshold() {
		if rs.Round() != constants.DEADBLOCKROUNDNR {
			if err := ce.castNextRound(txn, rs); err != nil {
				utils.DebugTrace(ce.logger, err)
				return err
			}
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doPendingNext(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	os := rs.OwnRoundState()
	rcert := os.RCert
	// the precommit timeout has been hit
	// count the precommits
	pcl, pcnl, err := rs.GetCurrentPreCommits()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	ce.logger.Debugf("doPendingNext:    MAXBH:%v    STBH:%v    RH:%v    RN:%v   PC:%v   PCN:%v	GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, len(pcl), len(pcnl), rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])
	// if greater than threshold precommits,
	// use our own precommit if we did a precommit this round
	// if not use a random precommit. This is safe due to
	// locking of vote additions.
	errorFree := true
	if len(pcl) >= rs.GetCurrentThreshold() {
		p, err := pcl.GetProposal()
		if err != nil {
			return err
		}

		if err := ce.updateValidValue(txn, rs, p); err != nil {
			var e *errorz.ErrInvalid
			if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
				utils.DebugTrace(ce.logger, err)
				return err
			}
			errorFree = false
		}

		if errorFree {
			if err := ce.castNextHeightFromPreCommits(txn, rs, pcl); err != nil {
				var e *errorz.ErrInvalid
				if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
					utils.DebugTrace(ce.logger, err)
					return err
				}
				errorFree = false
			}
		}

		if errorFree {
			return nil
		}
	}
	// if the combination of votes is greater than the
	// threshold without the precommits being enough
	// cast a next round
	if rcert.RClaims.Round != constants.DEADBLOCKROUND {
		if rcert.RClaims.Round == constants.DEADBLOCKROUNDNR {
			dbrnrTO := ce.storage.GetDeadBlockRoundNextRoundTimeout()
			if rs.OwnValidatingState.DBRNRExpired(dbrnrTO) {
				// Wait a long time before moving into Dead Block Round
				if len(pcl)+len(pcnl) >= rs.GetCurrentThreshold() {
					if err := ce.castNextRound(txn, rs); err != nil {
						utils.DebugTrace(ce.logger, err)
						return err
					}
					return nil
				}
			}
		} else {
			if len(pcl)+len(pcnl) >= rs.GetCurrentThreshold() {
				if err := ce.castNextRound(txn, rs); err != nil {
					utils.DebugTrace(ce.logger, err)
					return err
				}
				return nil
			}
		}
	}
	// threshold votes have not been observed
	// do nothing and wait for more votes
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doNextRoundStep(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	// count the precommit messages from this round
	pcl, pcnl, err := rs.GetCurrentPreCommits()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	// check for enough preCommits in current round to cast nextHeight
	if len(pcl) >= rs.GetCurrentThreshold() {
		ce.logger.Debugf("doNextRoundStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v   PC:%v   PCN:%v	GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, len(pcl), len(pcnl), rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])
		p, err := pcl.GetProposal()
		if err != nil {
			return err
		}
		errorFree := true
		if err := ce.updateValidValue(txn, rs, p); err != nil {
			var e *errorz.ErrInvalid
			if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
				utils.DebugTrace(ce.logger, err)
				return err
			}
			errorFree = false
		}
		if errorFree {
			if err := ce.castNextHeightFromPreCommits(txn, rs, pcl); err != nil {
				var e *errorz.ErrInvalid
				if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
					utils.DebugTrace(ce.logger, err)
					return err
				}
				errorFree = false
			}
		}
		if errorFree {
			return nil
		}
	}
	// last of all count next round messages from this round only
	nhl, nrl, err := rs.GetCurrentNext()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	ce.logger.Debugf("doNextRoundStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v   NH:%v   NR:%v	  GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, len(nhl), len(nrl), rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])

	// form a new round cert if we have enough
	if len(nrl) >= rs.GetCurrentThreshold() {
		err := ce.castNextRoundRCert(txn, rs, nrl)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
	}
	// if we do not have enough yet,
	// do nothing and wait for more votes
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doRoundJump(txn *badger.Txn, rs *RoundStates, rc *objs.RCert) error {
	ce.logger.Debugf("doRoundJump:    MAXBH:%v    STBH:%v    RH:%v    RN:%v		GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])
	// local node cast a precommit nil this round
	// count the precommits
	pcl, _, err := rs.GetCurrentPreCommits()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	// if local node knows of even a single
	// precommit, update the valid value
	if len(pcl) > rs.GetCurrentThreshold() {
		p, err := pcl.GetProposal()
		if err != nil {
			return err
		}
		if err := ce.updateValidValue(txn, rs, p); err != nil {
			var e *errorz.ErrInvalid
			if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
				utils.DebugTrace(ce.logger, err)
				return err
			}
		}
	}
	if err := ce.setMostRecentRCert(rs, rc); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

// doCheckValidValue is called by non-validating nodes to update ValidValue
// and perform other related tasks.
func (ce *Engine) doCheckValidValue(txn *badger.Txn, rs *RoundStates) error {
	ce.logger.Debugf("doCheckValidValue:    MAXBH:%v    STBH:%v    RH:%v    RN:%v	GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])

	nhl, _, err := rs.GetCurrentNext()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}

	// if we have a threshold
	// make a new round cert and form the new block header
	// proceed to next height
	if len(nhl) >= rs.GetCurrentThreshold() {
		err := ce.updateValidValue(txn, rs, nhl[0].NHClaims.Proposal)
		if err != nil {
			var e *errorz.ErrInvalid
			if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
				utils.DebugTrace(ce.logger, err)
				return err
			}
		} else {
			if err := ce.castNewCommittedBlockHeader(txn, rs, nhl); err != nil {
				utils.DebugTrace(ce.logger, err)
				var e *errorz.ErrInvalid
				if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
					return err
				}
			}
		}
		return nil
	}

	pcl, _, err := rs.GetCurrentPreCommits()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	// if local node knows of even a single
	// precommit, update the valid value
	if len(pcl) > rs.GetCurrentThreshold() {
		p, err := pcl.GetProposal()
		if err != nil {
			return err
		}
		if err := ce.updateValidValue(txn, rs, p); err != nil {
			var e *errorz.ErrInvalid
			if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
				utils.DebugTrace(ce.logger, err)
				return err
			}
		}
		return nil
	}

	pvl, _, err := rs.GetCurrentPreVotes()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	// if we have prevote consensus now
	if len(pvl) >= rs.GetCurrentThreshold() {
		// update the valid value
		p, err := pvl.GetProposal()
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		if err := ce.updateValidValue(txn, rs, p); err != nil {
			var e *errorz.ErrInvalid
			if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
				utils.DebugTrace(ce.logger, err)
				return err
			}
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doNextHeightStep(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	// we cast a next height
	// we are stuck here until consensus
	// count the next height messages from any round
	nhl, _, err := rs.GetCurrentNext()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	ce.logger.Debugf("doNextHeightStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v    NHs:%d	GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, len(nhl), rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])

	// if we have a threshold
	// make a new round cert and form the new block header
	// proceed to next height
	if len(nhl) >= rs.GetCurrentThreshold() {
		if err := ce.castNewCommittedBlockHeader(txn, rs, nhl); err != nil {
			utils.DebugTrace(ce.logger, err)
			var e *errorz.ErrInvalid
			if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
				return err
			}
		}
	}
	// we have not gotten enough next height messages yet
	// do nothing and wait for more messages
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doHeightJumpStep(txn *badger.Txn, rs *RoundStates, rcert *objs.RCert) (bool, error) {
	ce.logger.Debugf("doHeightJumpStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v	GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])
	// get the last element of the sorted future height
	// if we can just jump up to this height, do so.
	// if the height is only one more, we can simply move to this
	// height if everything looks okay
	if rcert.RClaims.Height != rs.Height()+1 {
		return false, nil
	}
	if rcert.RClaims.Round != 1 {
		return false, nil
	}
	// if we have a valid value, check if the valid value
	// matches the previous blockhash of block n+1
	// if so, form the block and jump up to this level
	// this is safe because we call isValid on all values
	// before storing in a lock
	if rs.ValidValueCurrent() {
		bhsh, err := rs.ValidValue().PClaims.BClaims.BlockHash()
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return false, err
		}
		if bytes.Equal(bhsh, rcert.RClaims.PrevBlock) {
			vv := rs.ValidValue()
			err := ce.castNewCommittedBlockFromProposalAndRCert(txn, rs, vv, rcert)
			if err != nil {
				var e *errorz.ErrInvalid
				if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
					utils.DebugTrace(ce.logger, err)
					return false, err
				}
				return false, nil
			}
			rs.OwnValidatingState.ValidValue = nil
			rs.OwnValidatingState.LockedValue = nil
			return true, nil
		}
		// TODO: handle case (else case) in which the ValidValue does not match
		//		 with local ValidValue; may or may not be possible.
	}
	// could not use valid value - only option left is to check if the block
	// had no tx's in it - if so we can build it with no additional information
	// that what we have in scope by forming bclaims for empty block and
	// checking if it has correct hash
	os := rs.OwnRoundState()
	ownRcert := os.RCert
	txs := [][]byte{}
	TxRoot, err := objs.MakeTxRoot(txs)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	bclaims := rs.OwnState.SyncToBH.BClaims
	PrevBlock := utils.CopySlice(ownRcert.RClaims.PrevBlock)
	headerRoot, err := ce.database.GetHeaderTrieRoot(txn, rs.OwnState.SyncToBH.BClaims.Height)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	StateRoot := utils.CopySlice(bclaims.StateRoot)
	bh := &objs.BlockHeader{
		BClaims: &objs.BClaims{
			PrevBlock:  PrevBlock,
			HeaderRoot: headerRoot,
			StateRoot:  StateRoot,
			TxRoot:     TxRoot,
			ChainID:    ownRcert.RClaims.ChainID,
			Height:     ownRcert.RClaims.Height,
		},
		TxHshLst: txs,
		SigGroup: utils.CopySlice(rcert.SigGroup),
	}
	bhash, err := bh.BlockHash()
	if err != nil {
		return false, err
	}
	if !bytes.Equal(bhash, rcert.RClaims.PrevBlock) {
		return false, nil
	}
	ok, err := ce.isValid(txn, rs, bh.BClaims.ChainID, bh.BClaims.StateRoot, bh.BClaims.HeaderRoot, []interfaces.Transaction{})
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	if !ok {
		return false, nil
	}
	if err := bh.ValidateSignatures(&crypto.BNGroupValidator{}); err != nil {
		return false, nil // if sig is not valid then move to sync mode
	}
	if err := ce.setMostRecentBlockHeader(txn, rs, bh); err != nil {
		return false, err
	}
	return true, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) updateValidValue(txn *badger.Txn, rs *RoundStates, p *objs.Proposal) error {
	ce.logger.Debugf("updateValidValue:    MAXBH:%v    STBH:%v    RH:%v    RN:%v	GRPK:%x..%x", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, rs.ValidatorSet.GroupKey[0:4], rs.ValidatorSet.GroupKey[len(rs.ValidatorSet.GroupKey)-5:len(rs.ValidatorSet.GroupKey)-1])
	txs, _, err := ce.dm.GetTxs(txn, p.PClaims.BClaims.Height, rs.round, p.TxHshLst)
	if err != nil {
		if err != errorz.ErrMissingTransactions {
			utils.DebugTrace(ce.logger, err)
			return err
		}
	}
	// check if the proposal is valid
	ok, err := ce.isValid(txn, rs, p.PClaims.BClaims.ChainID, p.PClaims.BClaims.StateRoot, p.PClaims.BClaims.HeaderRoot, txs)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if !ok { // proposal is invalid
		return errorz.ErrInvalid{}.New("proposal is invalid in update vv")
	}
	// update the valid value
	if err := ce.setMostRecentValidValue(rs, p); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}
