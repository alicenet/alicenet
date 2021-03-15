package lstate

import (
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/utils"
)

// These are intermediate handlers. Once the step handlers have decided on how
// to perform an action one of these is called to perform the action. Every
// method in this file will call a setter at termination.

func (ce *Engine) castNewProposalValue(rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	txList, txRoot, stateRoot, headerRoot, err := ce.getValidValue(rs)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if stateRoot == nil {
		stateRoot = make([]byte, constants.HashLen)
	}
	p := &objs.Proposal{
		PClaims: &objs.PClaims{
			BClaims: &objs.BClaims{
				ChainID:    rs.OwnState.SyncToBH.BClaims.ChainID,
				Height:     rs.OwnState.SyncToBH.BClaims.Height + 1,
				PrevBlock:  rs.PrevBlock(),
				HeaderRoot: headerRoot,
				StateRoot:  stateRoot,
				TxRoot:     txRoot,
				TxCount:    uint32(len(txList)),
			},
			RCert: &objs.RCert{
				SigGroup: rs.OwnRoundState().RCert.SigGroup,
				RClaims: &objs.RClaims{
					ChainID:   rs.OwnState.SyncToBH.BClaims.ChainID,
					Height:    rs.OwnState.SyncToBH.BClaims.Height + 1,
					PrevBlock: rs.PrevBlock(),
					Round:     rs.Round(),
				},
			},
		},
		TxHshLst: txList,
	}
	ce.logger.Tracef(`
    Proposal{
      PClaims{
        BClaims{
          ChainID:      %v
          Height:       %v
          PrevBlock:    %x
          HeaderRoot:   %x
          StateRoot:    %x
          TxRoot:       %x
          TxCount:      %v
        }
        RCert{
          SigGroup:     %x ... %x
          RClaims{
            ChainID:    %v
            Height:     %v
            PrevBlock:  %x
            Round:      %v
          }
        }
      }
      TxHshLst:         %x
    }`, rs.OwnState.SyncToBH.BClaims.ChainID, rs.OwnState.SyncToBH.BClaims.Height+1, rs.PrevBlock(), headerRoot, stateRoot, txRoot, len(txList), rs.OwnRoundState().RCert.SigGroup[0:16], rs.OwnRoundState().RCert.SigGroup[len(rs.OwnRoundState().RCert.SigGroup)-11:], rs.OwnState.SyncToBH.BClaims.ChainID, rs.OwnState.SyncToBH.BClaims.Height+1, rs.PrevBlock(), rs.Round(), txList)
	err = p.Sign(ce.secpSigner)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := p.ValidateSignatures(&crypto.Secp256k1Validator{}, &crypto.BNGroupValidator{}); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	err = ce.updateValidValue(rs, p)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentProposal(rs, p); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.database.SetBroadcastProposal(rs.txn, p); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

func (ce *Engine) castProposalFromValue(rs *RoundStates, prop *objs.Proposal) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	// we have either a locked or valid value
	rcert := rs.OwnRoundState().RCert
	p, err := prop.RePropose(ce.secpSigner, rcert)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := p.ValidateSignatures(&crypto.Secp256k1Validator{}, &crypto.BNGroupValidator{}); err != nil {
		ce.logger.Debugf("Error in castProposalFromValue at ValidateSignatures: %v", err)
		return err
	}
	if err := ce.setMostRecentProposal(rs, p); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.database.SetBroadcastProposal(rs.txn, p); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) castPreVoteWithLock(rs *RoundStates, lock *objs.Proposal, p *objs.Proposal) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	ok, err := objs.BClaimsEqual(lock, p)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if ok {
		if err := ce.castPreVote(rs, p); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		return nil
	}
	if err := ce.castPreVoteNil(rs); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

func (ce *Engine) castPreVote(rs *RoundStates, p *objs.Proposal) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	pv, err := p.PreVote(ce.secpSigner)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := pv.ValidateSignatures(&crypto.Secp256k1Validator{}, &crypto.BNGroupValidator{}); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentValidValue(rs, p); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentPreVote(rs, pv); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.database.SetBroadcastPreVote(rs.txn, pv); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) castPreVoteNil(rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	rcert := rs.OwnRoundState().RCert
	pvn, err := rcert.PreVoteNil(ce.secpSigner)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := pvn.ValidateSignatures(&crypto.Secp256k1Validator{}, &crypto.BNGroupValidator{}); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.database.SetBroadcastPreVoteNil(rs.txn, pvn); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentPreVoteNil(rs, pvn); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) castPreCommit(rs *RoundStates, preVotes []*objs.PreVote) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	pvl := objs.PreVoteList(preVotes)
	pc, err := pvl.MakePreCommit(ce.secpSigner)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := pc.ValidateSignatures(&crypto.Secp256k1Validator{}, &crypto.BNGroupValidator{}); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.database.SetBroadcastPreCommit(rs.txn, pc); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentPreCommit(rs, pc); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) castPreCommitNil(rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	rcert := rs.OwnRoundState().RCert
	pcn, err := rcert.PreCommitNil(ce.secpSigner)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := pcn.ValidateSignatures(&crypto.Secp256k1Validator{}, &crypto.BNGroupValidator{}); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.database.SetBroadcastPreCommitNil(rs.txn, pcn); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentPreCommitNil(rs, pcn); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) castNextRound(rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	rcert := rs.OwnRoundState().RCert
	rcBytes, err := rcert.MarshalBinary()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	rc := &objs.RCert{}
	err = rc.UnmarshalBinary(rcBytes)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	nr, err := rc.NextRound(ce.secpSigner, ce.bnSigner)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := nr.ValidateSignatures(&crypto.Secp256k1Validator{}, &crypto.BNGroupValidator{}); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.database.SetBroadcastNextRound(rs.txn, nr); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentNextRound(rs, nr); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) castNextRoundRCert(rs *RoundStates, NextRounds []*objs.NextRound) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	nrl := objs.NextRoundList(NextRounds)
	shares := [][]byte{}
	for _, val := range rs.ValidatorSet.Validators {
		shares = append(shares, val.GroupShare)
	}
	rc, err := nrl.MakeRoundCert(ce.bnSigner, shares)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := rc.ValidateSignature(&crypto.BNGroupValidator{}); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.database.SetBroadcastRCert(rs.txn, rc); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentRCert(rs, rc); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) castNextHeightFromNextHeight(rs *RoundStates, nh *objs.NextHeight) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	newNh, err := nh.Plagiarize(ce.secpSigner, ce.bnSigner)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := newNh.ValidateSignatures(&crypto.Secp256k1Validator{}, &crypto.BNGroupValidator{}); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.database.SetBroadcastNextHeight(rs.txn, newNh); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentNextHeight(rs, newNh); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

func (ce *Engine) castNextHeightFromPreCommits(rs *RoundStates, preCommits []*objs.PreCommit) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	pcl := objs.PreCommitList(preCommits)
	nh, err := pcl.MakeNextHeight(ce.secpSigner, ce.bnSigner)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := nh.ValidateSignatures(&crypto.Secp256k1Validator{}, &crypto.BNGroupValidator{}); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.database.SetBroadcastNextHeight(rs.txn, nh); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentNextHeight(rs, nh); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) castNewCommittedBlockHeader(rs *RoundStates, nextHeights []*objs.NextHeight) error {
	nhl := objs.NextHeightList(nextHeights)
	shares := make([][]byte, len(rs.ValidatorSet.Validators))
	for i := 0; i < len(rs.ValidatorSet.Validators); i++ {
		vobj := rs.ValidatorSet.Validators[i]
		shares[i] = vobj.GroupShare
	}
	bh, rc, err := nhl.MakeBlockHeader(ce.bnSigner, shares)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := rc.ValidateSignature(&crypto.BNGroupValidator{}); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := bh.ValidateSignatures(&crypto.BNGroupValidator{}); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentBlockHeader(rs, bh); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

func (ce *Engine) castNewCommittedBlockFromProposalAndRCert(rs *RoundStates, p *objs.Proposal, rc *objs.RCert) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	bh := &objs.BlockHeader{
		SigGroup: rc.SigGroup,
		BClaims:  p.PClaims.BClaims,
		TxHshLst: p.TxHshLst,
	}
	if err := bh.ValidateSignatures(&crypto.BNGroupValidator{}); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if err := ce.setMostRecentBlockHeader(rs, bh); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}
