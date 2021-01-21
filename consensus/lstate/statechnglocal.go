package lstate

import (
	"bytes"
	"errors"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"

	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/dgraph-io/badger/v2"
)

// These are the step handlers. They figure out how to take an action based on
// what action is determined as necessary in updateLocalStateInternal

func (ce *Engine) doPendingProposalStep(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	ce.logger.Debugf("doPendingProposalStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round)
	os := rs.OwnRoundState()
	rcert := os.RCert
	if rcert.RClaims.Round == constants.DEADBLOCKROUND {
		return nil
	}
	var chngHandler changeHandler
	// if not locked or valid form new proposal
	chngHandler = &dPPSProposeNewHandler{ce: ce, txn: txn, rs: rs}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// if not locked but valid known, propose valid value
	chngHandler = &dPPSProposeValidHandler{ce: ce, txn: txn, rs: rs}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// if locked, propose locked
	chngHandler = &dPPSProposeLockedHandler{ce: ce, txn: txn, rs: rs}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	return nil
}

type changeHandler interface {
	evalCriteria() bool
	evalLogic() error
}

type dPPSProposeNewHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
}

func (pn *dPPSProposeNewHandler) evalCriteria() bool {
	return !pn.rs.LockedValueCurrent() && !pn.rs.ValidValueCurrent() // 00 case
}

func (pn *dPPSProposeNewHandler) evalLogic() error {
	return pn.ce.dPPSProposeNewFunc(pn.txn, pn.rs)
}

func (ce *Engine) dPPSProposeNewFunc(txn *badger.Txn, rs *RoundStates) error {
	if err := ce.castNewProposalValue(txn, rs); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

type dPPSProposeValidHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
}

func (pv *dPPSProposeValidHandler) evalCriteria() bool {
	return !pv.rs.LockedValueCurrent() && pv.rs.ValidValueCurrent() // 01 case
}

func (pv *dPPSProposeValidHandler) evalLogic() error {
	if err := pv.ce.castProposalFromValue(pv.txn, pv.rs, pv.rs.ValidValue()); err != nil {
		utils.DebugTrace(pv.ce.logger, err)
		return err
	}
	return nil
}

type dPPSProposeLockedHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
}

func (pl *dPPSProposeLockedHandler) evalCriteria() bool {
	return pl.rs.LockedValueCurrent() // 10 or 11 case
}

func (pl *dPPSProposeLockedHandler) evalLogic() error {
	if err := pl.ce.castProposalFromValue(pl.txn, pl.rs, pl.rs.LockedValue()); err != nil {
		utils.DebugTrace(pl.ce.logger, err)
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
	ce.logger.Debugf("doPendingPreVoteStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round)
	var chngHandler changeHandler
	chngHandler = &dPPVSDeadBlockRoundHandler{ce: ce, txn: txn, rs: rs}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}

	p := rs.GetCurrentProposal()
	// proposal timeout hit
	// if there is a current proposal
	// Height, round, previous block hash

	// if we are not locked and there is no known valid value
	// check if the proposed value is valid, and if so
	// prevote this value
	//00 case
	chngHandler = &dPPVSPreVoteNewHandler{ce: ce, txn: txn, rs: rs, p: p}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// if we are locked on a valid value, only prevote the value if it is equal
	// to the lock
	//01 case
	chngHandler = &dPPVSPreVoteValidHandler{ce: ce, txn: txn, rs: rs, p: p}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// if we are locked on a locked value, only prevote the value if it is equal
	// to the lock
	//10 case
	//11 case
	chngHandler = &dPPVSPreVoteLockedHandler{ce: ce, txn: txn, rs: rs, p: p}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}

	// no current proposal known
	// so prevote nil
	chngHandler = &dPPVSPreVoteNilHandler{ce: ce, txn: txn, rs: rs, p: p}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	return nil
}

type dPPVSDeadBlockRoundHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
}

func (dbrh *dPPVSDeadBlockRoundHandler) evalCriteria() bool {
	os := dbrh.rs.OwnRoundState()
	rcert := os.RCert
	return rcert.RClaims.Round == constants.DEADBLOCKROUND
}

func (dbrh *dPPVSDeadBlockRoundHandler) evalLogic() error {
	return dbrh.ce.dPPSProposeNewFunc(dbrh.txn, dbrh.rs)
}

func (ce *Engine) dPPVSDeadBlockRoundFunc(txn *badger.Txn, rs *RoundStates) error {
	// Safely form EmptyBlock PreVote
	os := rs.OwnRoundState()
	rcert := os.RCert
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

type dPPVSPreVoteNewHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
	p   *objs.Proposal
}

func (pvnewh *dPPVSPreVoteNewHandler) evalCriteria() bool {
	cond1 := pvnewh.p != nil
	cond2 := !pvnewh.rs.LockedValueCurrent() && !pvnewh.rs.ValidValueCurrent()
	return cond1 && cond2
}

func (pvnewh *dPPVSPreVoteNewHandler) evalLogic() error {
	return pvnewh.ce.dPPVSPreVoteNewFunc(pvnewh.txn, pvnewh.rs, pvnewh.p)
}

func (ce *Engine) dPPVSPreVoteNewFunc(txn *badger.Txn, rs *RoundStates, p *objs.Proposal) error {
	txs, _, err := ce.dm.GetTxs(txn, p.PClaims.BClaims.Height, p.TxHshLst)
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

type dPPVSPreVoteValidHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
	p   *objs.Proposal
}

func (pvvh *dPPVSPreVoteValidHandler) evalCriteria() bool {
	cond1 := pvvh.p != nil
	cond2 := !pvvh.rs.LockedValueCurrent() && pvvh.rs.ValidValueCurrent()
	return cond1 && cond2
}

func (pvvh *dPPVSPreVoteValidHandler) evalLogic() error {
	return pvvh.ce.dPPVSPreVoteValidFunc(pvvh.txn, pvvh.rs, pvvh.p)
}

func (ce *Engine) dPPVSPreVoteValidFunc(txn *badger.Txn, rs *RoundStates, p *objs.Proposal) error {
	if err := ce.castPreVoteWithLock(txn, rs, rs.ValidValue(), p); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

type dPPVSPreVoteLockedHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
	p   *objs.Proposal
}

func (pvlh *dPPVSPreVoteLockedHandler) evalCriteria() bool {
	cond1 := pvlh.p != nil
	cond2 := pvlh.rs.LockedValueCurrent()
	return cond1 && cond2
}

func (pvlh *dPPVSPreVoteLockedHandler) evalLogic() error {
	return pvlh.ce.dPPVSPreVoteLockedFunc(pvlh.txn, pvlh.rs, pvlh.p)
}

func (ce *Engine) dPPVSPreVoteLockedFunc(txn *badger.Txn, rs *RoundStates, p *objs.Proposal) error {
	if err := ce.castPreVoteWithLock(txn, rs, rs.LockedValue(), p); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

type dPPVSPreVoteNilHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
	p   *objs.Proposal
}

func (pvnh *dPPVSPreVoteNilHandler) evalCriteria() bool {
	return pvnh.p == nil
}

func (pvnh *dPPVSPreVoteNilHandler) evalLogic() error {
	return pvnh.ce.dPPVSPreVoteNilFunc(pvnh.txn, pvnh.rs)
}

func (ce *Engine) dPPVSPreVoteNilFunc(txn *badger.Txn, rs *RoundStates) error {
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
	ce.logger.Debugf("doPreVoteStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round)
	// local node must have cast a preVote to get here
	// count the prevotes and prevote nils
	pvl, _, err := rs.GetCurrentPreVotes()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	// if we have enough prevotes, cast a precommit
	// this will update the locked value
	var chngHandler changeHandler
	chngHandler = &dPVSCastPCHandler{ce: ce, txn: txn, rs: rs, pvl: pvl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// no thresholds met, so do nothing
	return nil
}

type dPVSCastPCHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
	pvl []*objs.PreVote
}

func (pvcpc *dPVSCastPCHandler) evalCriteria() bool {
	return len(pvcpc.pvl) >= pvcpc.rs.GetCurrentThreshold()
}

func (pvcpc *dPVSCastPCHandler) evalLogic() error {
	return pvcpc.ce.dPVSCastPCFunc(pvcpc.txn, pvcpc.rs, pvcpc.pvl)
}

func (ce *Engine) dPVSCastPCFunc(txn *badger.Txn, rs *RoundStates, pvl []*objs.PreVote) error {
	if err := ce.castPreCommit(txn, rs, pvl); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doPreVoteNilStep(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	ce.logger.Debugf("doPreVoteNilStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round)
	// local node must have cast a preVote nil
	// count the preVotes and prevote nils
	pvl, pvnl, err := rs.GetCurrentPreVotes()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	var chngHandler changeHandler
	chngHandler = &dPVNSUpdateVVHandler{ce: ce, txn: txn, rs: rs, pvl: pvl, pvnl: pvnl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// if greater than threshold prevote nils, cast the prevote nil
	chngHandler = &dPVNSCastPCNHandler{ce: ce, txn: txn, rs: rs, pvnl: pvnl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	return nil
}

type dPVNSUpdateVVHandler struct {
	ce   *Engine
	txn  *badger.Txn
	rs   *RoundStates
	pvl  objs.PreVoteList
	pvnl objs.PreVoteNilList
}

func (pvnu *dPVNSUpdateVVHandler) evalCriteria() bool {
	return len(pvnu.pvl) >= pvnu.rs.GetCurrentThreshold()
}

func (pvnu *dPVNSUpdateVVHandler) evalLogic() error {
	return pvnu.ce.dPVNSUpdateVVFunc(pvnu.txn, pvnu.rs, pvnu.pvl)
}

func (ce *Engine) dPVNSUpdateVVFunc(txn *badger.Txn, rs *RoundStates, pvl objs.PreVoteList) error {
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
	return nil
}

type dPVNSCastPCNHandler struct {
	ce   *Engine
	txn  *badger.Txn
	rs   *RoundStates
	pvnl objs.PreVoteNilList
}

func (cpcn *dPVNSCastPCNHandler) evalCriteria() bool {
	return len(cpcn.pvnl) >= cpcn.rs.GetCurrentThreshold()
}

func (cpcn *dPVNSCastPCNHandler) evalLogic() error {
	return cpcn.ce.dPVNSCastPCNFunc(cpcn.txn, cpcn.rs)
}

func (ce *Engine) dPVNSCastPCNFunc(txn *badger.Txn, rs *RoundStates) error {
	if err := ce.castPreCommitNil(txn, rs); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
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
	ce.logger.Debugf("doPendingPreCommit:    MAXBH:%v    STBH:%v    RH:%v    RN:%v", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round)
	//os := rs.OwnRoundState()
	//rcert := os.RCert
	// prevote timeout hit with no clear consensus in either direction
	// during cycle before timeout
	// count the prevotes
	pvl, pvnl, err := rs.GetCurrentPreVotes()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	// if we have prevote consensus now
	var chngHandler changeHandler
	chngHandler = &dPPCCastPCHandler{ce: ce, txn: txn, rs: rs, pvl: pvl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	chngHandler = &dPPCUpdateVVHandler{ce: ce, txn: txn, rs: rs, pvl: pvl, pvnl: pvnl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// fallthrough to precommit nil
	// since the timeout has expired,
	// free to cast preCommitNil without
	// clear consensus if the total votes is
	// greater than threshold
	chngHandler = &dPPCNotDBRHandler{ce: ce, txn: txn, rs: rs, pvl: pvl, pvnl: pvnl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// threshold not met as of yet
	return nil
}

type dPPCCastPCHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
	pvl objs.PreVoteList
}

func (cpc *dPPCCastPCHandler) evalCriteria() bool {
	cond1 := len(cpc.pvl) >= cpc.rs.GetCurrentThreshold()
	cond2 := cpc.rs.LocalPreVoteCurrent()
	return cond1 && cond2
}

func (cpc *dPPCCastPCHandler) evalLogic() error {
	return cpc.ce.dPPCCastPCFunc(cpc.txn, cpc.rs, cpc.pvl)
}

func (ce *Engine) dPPCCastPCFunc(txn *badger.Txn, rs *RoundStates, pvl objs.PreVoteList) error {
	if err := ce.castPreCommit(txn, rs, pvl); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

type dPPCUpdateVVHandler struct {
	ce   *Engine
	txn  *badger.Txn
	rs   *RoundStates
	pvl  objs.PreVoteList
	pvnl objs.PreVoteNilList
}

func (uvv *dPPCUpdateVVHandler) evalCriteria() bool {
	cond1 := len(uvv.pvl) >= uvv.rs.GetCurrentThreshold()
	cond2 := !uvv.rs.LocalPreVoteCurrent()
	return cond1 && cond2
}

func (uvv *dPPCUpdateVVHandler) evalLogic() error {
	return uvv.ce.dPPCUpdateVVFunc(uvv.txn, uvv.rs, uvv.pvl, uvv.pvnl)
}

func (ce *Engine) dPPCUpdateVVFunc(txn *badger.Txn, rs *RoundStates, pvl objs.PreVoteList, pvnl objs.PreVoteNilList) error {
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
	// If we are here, then there are threshold PreCommits but we needed to
	// update ValidValue. We now MUST fall through and PreCommitNil.
	os := rs.OwnRoundState()
	rcert := os.RCert
	if rcert.RClaims.Round != constants.DEADBLOCKROUND {
		if len(pvl)+len(pvnl) >= rs.GetCurrentThreshold() {
			if err := ce.castPreCommitNil(txn, rs); err != nil {
				utils.DebugTrace(ce.logger, err)
				return err
			}
			return nil
		}
		return nil
	}
	return nil
}

type dPPCNotDBRHandler struct {
	ce   *Engine
	txn  *badger.Txn
	rs   *RoundStates
	pvl  objs.PreVoteList
	pvnl objs.PreVoteNilList
}

func (ndbr *dPPCNotDBRHandler) evalCriteria() bool {
	os := ndbr.rs.OwnRoundState()
	rcert := os.RCert
	cond1 := rcert.RClaims.Round != constants.DEADBLOCKROUND
	cond2 := len(ndbr.pvl)+len(ndbr.pvnl) >= ndbr.rs.GetCurrentThreshold()
	return cond1 && cond2
}

func (ndbr *dPPCNotDBRHandler) evalLogic() error {
	return ndbr.ce.dPPCNotDBRFunc(ndbr.txn, ndbr.rs)
}

func (ce *Engine) dPPCNotDBRFunc(txn *badger.Txn, rs *RoundStates) error {
	if err := ce.castPreCommitNil(txn, rs); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doPreCommitStep(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	ce.logger.Debugf("doPreCommitStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round)
	// local node cast a precommit this round
	// count the precommits
	pcl, _, err := rs.GetCurrentPreCommits()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	// if we have consensus and can verify
	// cast the nextHeight
	var chngHandler changeHandler
	chngHandler = &dPCSCastNHHandler{ce: ce, txn: txn, rs: rs, pcl: pcl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// no consensus, wait for more votes
	return nil
}

type dPCSCastNHHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
	pcl objs.PreCommitList
}

func (cnh *dPCSCastNHHandler) evalCriteria() bool {
	return len(cnh.pcl) >= cnh.rs.GetCurrentThreshold()
}

func (cnh *dPCSCastNHHandler) evalLogic() error {
	return cnh.ce.dPCSCastNHFunc(cnh.txn, cnh.rs, cnh.pcl)
}

func (ce *Engine) dPCSCastNHFunc(txn *badger.Txn, rs *RoundStates, pcl objs.PreCommitList) error {
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

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doPreCommitNilStep(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	ce.logger.Debugf("doPreCommitNilStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round)
	// local node cast a precommit nil this round
	// count the precommits
	pcl, pcnl, err := rs.GetCurrentPreCommits()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	// if we have a preCommit consensus,
	// move directly to next height
	var chngHandler changeHandler
	chngHandler = &dPCNSCastNHHandler{ce: ce, txn: txn, rs: rs, pcl: pcl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// if we have a consensus for a precommit nil,
	// cast a next round
	chngHandler = &dPCNSCastNRHandler{ce: ce, txn: txn, rs: rs, pcnl: pcnl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	return nil
}

type dPCNSCastNHHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
	pcl objs.PreCommitList
}

func (cnh *dPCNSCastNHHandler) evalCriteria() bool {
	return len(cnh.pcl) >= cnh.rs.GetCurrentThreshold()
}

func (cnh *dPCNSCastNHHandler) evalLogic() error {
	return cnh.ce.dPCNSCastNHFunc(cnh.txn, cnh.rs, cnh.pcl)
}

func (ce *Engine) dPCNSCastNHFunc(txn *badger.Txn, rs *RoundStates, pcl objs.PreCommitList) error {
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

type dPCNSCastNRHandler struct {
	ce   *Engine
	txn  *badger.Txn
	rs   *RoundStates
	pcnl objs.PreCommitNilList
}

func (cnr *dPCNSCastNRHandler) evalCriteria() bool {
	cond1 := len(cnr.pcnl) >= cnr.rs.GetCurrentThreshold()
	cond2 := cnr.rs.Round() != constants.DEADBLOCKROUNDNR
	return cond1 && cond2
}

func (cnr *dPCNSCastNRHandler) evalLogic() error {
	return cnr.ce.dPCNSCastNRFunc(cnr.txn, cnr.rs)
}

func (ce *Engine) dPCNSCastNRFunc(txn *badger.Txn, rs *RoundStates) error {
	if err := ce.castNextRound(txn, rs); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
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
	ce.logger.Debugf("doPendingNext:    MAXBH:%v    STBH:%v    RH:%v    RN:%v", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round)
	// the precommit timeout has been hit
	// count the precommits
	pcl, pcnl, err := rs.GetCurrentPreCommits()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	// if greater than threshold precommits,
	// use our own precommit if we did a precommit this round
	// if not use a random precommit. This is safe due to
	// locking of vote additions.
	var chngHandler changeHandler
	chngHandler = &dPNCastNextHeightHandler{ce: ce, txn: txn, rs: rs, pcl: pcl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// if the combination of votes is greater than the
	// threshold without the precommits being enough
	// cast a next round
	chngHandler = &dPNCastNextRoundHandler{ce: ce, txn: txn, rs: rs, pcl: pcl, pcnl: pcnl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// threshold votes have not been observed
	// do nothing and wait for more votes
	return nil
}

type dPNCastNextHeightHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
	pcl objs.PreCommitList
}

func (cnh *dPNCastNextHeightHandler) evalCriteria() bool {
	return len(cnh.pcl) >= cnh.rs.GetCurrentThreshold()
}

func (cnh *dPNCastNextHeightHandler) evalLogic() error {
	return cnh.ce.dPNCastNextHeightFunc(cnh.txn, cnh.rs, cnh.pcl)
}

func (ce *Engine) dPNCastNextHeightFunc(txn *badger.Txn, rs *RoundStates, pcl objs.PreCommitList) error {
	errorFree := true
	os := rs.OwnRoundState()
	rcert := os.RCert

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

	// if the combination of votes is greater than the
	// threshold without the precommits being enough
	// cast a next round
	if rcert.RClaims.Round != constants.DEADBLOCKROUND {
		if (rcert.RClaims.Round == constants.DEADBLOCKROUNDNR) && !rs.OwnValidatingState.DBRNRExpired() {
			// Wait a long time before moving into Dead Block Round
			return nil
		}
		if err := ce.castNextRound(txn, rs); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		return nil
	}
	return nil
}

type dPNCastNextRoundHandler struct {
	ce   *Engine
	txn  *badger.Txn
	rs   *RoundStates
	pcl  objs.PreCommitList
	pcnl objs.PreCommitNilList
}

func (cnr *dPNCastNextRoundHandler) evalCriteria() bool {
	os := cnr.rs.OwnRoundState()
	rcert := os.RCert
	cond1 := len(cnr.pcl) < cnr.rs.GetCurrentThreshold()
	cond2 := rcert.RClaims.Round != constants.DEADBLOCKROUND
	cond3 := len(cnr.pcl)+len(cnr.pcnl) >= cnr.rs.GetCurrentThreshold()
	return cond1 && cond2 && cond3
}

func (cnr *dPNCastNextRoundHandler) evalLogic() error {
	return cnr.ce.dPNCastNextRoundFunc(cnr.txn, cnr.rs)
}

func (ce *Engine) dPNCastNextRoundFunc(txn *badger.Txn, rs *RoundStates) error {
	os := rs.OwnRoundState()
	rcert := os.RCert
	// if the combination of votes is greater than the
	// threshold without the precommits being enough
	// cast a next round
	if (rcert.RClaims.Round == constants.DEADBLOCKROUNDNR) && !rs.OwnValidatingState.DBRNRExpired() {
		// Wait a long time before moving into Dead Block Round
		return nil
	}
	if err := ce.castNextRound(txn, rs); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doNextRoundStep(txn *badger.Txn, rs *RoundStates) error {
	if !rs.IsCurrentValidator() {
		return nil
	}
	ce.logger.Debugf("doNextRoundStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round)
	// count the precommit messages from this round
	pcl, _, err := rs.GetCurrentPreCommits()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	// check for enough preCommits in current round to cast nextHeight
	var chngHandler changeHandler
	chngHandler = &dNRSCastNextHeightHandler{ce: ce, txn: txn, rs: rs, pcl: pcl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// last of all count next round messages from this round only
	_, nrl, err := rs.GetCurrentNext()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}

	// form a new round cert if we have enough
	chngHandler = &dNRSCastNextRoundHandler{ce: ce, txn: txn, rs: rs, pcl: pcl, nrl: nrl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// if we do not have enough yet,
	// do nothing and wait for more votes
	return nil
}

type dNRSCastNextHeightHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
	pcl objs.PreCommitList
}

func (cnh *dNRSCastNextHeightHandler) evalCriteria() bool {
	return len(cnh.pcl) >= cnh.rs.GetCurrentThreshold()
}

func (cnh *dNRSCastNextHeightHandler) evalLogic() error {
	return cnh.ce.dNRSCastNextHeightFunc(cnh.txn, cnh.rs, cnh.pcl)
}

func (ce *Engine) dNRSCastNextHeightFunc(txn *badger.Txn, rs *RoundStates, pcl objs.PreCommitList) error {
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

	// last of all count next round messages from this round only
	_, nrl, err := rs.GetCurrentNext()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}

	// form a new round cert if we have enough
	if len(nrl) >= rs.GetCurrentThreshold() {
		if err := ce.castNextRoundRCert(txn, rs, nrl); err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
	}
	// if we do not have enough yet,
	// do nothing and wait for more votes
	return nil
}

type dNRSCastNextRoundHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
	pcl objs.PreCommitList
	nrl objs.NextRoundList
}

func (cnr *dNRSCastNextRoundHandler) evalCriteria() bool {
	cond1 := len(cnr.pcl) < cnr.rs.GetCurrentThreshold()
	cond2 := len(cnr.nrl) >= cnr.rs.GetCurrentThreshold()
	return cond1 && cond2
}

func (cnr *dNRSCastNextRoundHandler) evalLogic() error {
	return cnr.ce.dNRSCastNextRoundFunc(cnr.txn, cnr.rs, cnr.nrl)
}

func (ce *Engine) dNRSCastNextRoundFunc(txn *badger.Txn, rs *RoundStates, nrl objs.NextRoundList) error {
	if err := ce.castNextRoundRCert(txn, rs, nrl); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// TODO: look at the criteria to ensure it is correct!
func (ce *Engine) doRoundJump(txn *badger.Txn, rs *RoundStates, rc *objs.RCert) error {
	ce.logger.Debugf("doRoundJump:    MAXBH:%v    STBH:%v    RH:%v    RN:%v", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round)
	// local node cast a precommit nil this round
	// count the precommits
	pcl, _, err := rs.GetCurrentPreCommits()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	// if local node knows of even a single
	// precommit, update the valid value
	var chngHandler changeHandler
	chngHandler = &dRJUpdateVVHandler{ce: ce, txn: txn, rs: rs, rc: rc, pcl: pcl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	chngHandler = &dRJSetRCertHandler{ce: ce, rs: rs, rc: rc, pcl: pcl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	return nil
}

type dRJUpdateVVHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
	rc  *objs.RCert
	pcl objs.PreCommitList
}

func (uvv *dRJUpdateVVHandler) evalCriteria() bool {
	return len(uvv.pcl) > uvv.rs.GetCurrentThreshold()
}

func (uvv *dRJUpdateVVHandler) evalLogic() error {
	return uvv.ce.dRJUpdateVVFunc(uvv.txn, uvv.rs, uvv.rc, uvv.pcl)
}

func (ce *Engine) dRJUpdateVVFunc(txn *badger.Txn, rs *RoundStates, rc *objs.RCert, pcl objs.PreCommitList) error {
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
	if err := ce.setMostRecentRCert(rs, rc); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	return nil
}

type dRJSetRCertHandler struct {
	ce  *Engine
	rs  *RoundStates
	rc  *objs.RCert
	pcl objs.PreCommitList
}

func (src *dRJSetRCertHandler) evalCriteria() bool {
	return len(src.pcl) <= src.rs.GetCurrentThreshold()
}

func (src *dRJSetRCertHandler) evalLogic() error {
	return src.ce.dRJSetRCertFunc(src.rs, src.rc)
}

func (ce *Engine) dRJSetRCertFunc(rs *RoundStates, rc *objs.RCert) error {
	if err := ce.setMostRecentRCert(rs, rc); err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
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
	ce.logger.Debugf("doNextHeightStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v    NHs:%d", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round, len(nhl))

	// if we have a threshold
	// make a new round cert and form the new block header
	// proceed to next height
	var chngHandler changeHandler
	chngHandler = &dNHSCastBHHandler{ce: ce, txn: txn, rs: rs, nhl: nhl}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}
	// we have not gotten enough next height messages yet
	// do nothing and wait for more messages
	return nil
}

type dNHSCastBHHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
	nhl objs.NextHeightList
}

func (cbh *dNHSCastBHHandler) evalCriteria() bool {
	return len(cbh.nhl) >= cbh.rs.GetCurrentThreshold()
}

func (cbh *dNHSCastBHHandler) evalLogic() error {
	return cbh.ce.dNHSCastBHFunc(cbh.txn, cbh.rs, cbh.nhl)
}

func (ce *Engine) dNHSCastBHFunc(txn *badger.Txn, rs *RoundStates, nhl objs.NextHeightList) error {
	if err := ce.castNewCommittedBlockHeader(txn, rs, nhl); err != nil {
		utils.DebugTrace(ce.logger, err)
		var e *errorz.ErrInvalid
		if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) doHeightJumpStep(txn *badger.Txn, rs *RoundStates, rcert *objs.RCert) error {
	ce.logger.Debugf("doHeightJumpStep:    MAXBH:%v    STBH:%v    RH:%v    RN:%v", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round)
	// get the last element of the sorted future height
	// if we can just jump up to this height, do so.
	// if the height is only one more, we can simply move to this
	// height if everything looks okay

	// if we have a valid value, check if the valid value
	// matches the previous blockhash of block n+1
	// if so, form the block and jump up to this level
	// this is safe because we call isValid on all values
	// before storing in a lock
	var chngHandler changeHandler
	chngHandler = &dHJSJumpHandler{ce: ce, txn: txn, rs: rs, rcert: rcert}
	if chngHandler.evalCriteria() {
		return chngHandler.evalLogic()
	}

	// we can not do anything from here without a ton of work
	// so easier to just wait for the next block header to unsync us
	return nil
}

type dHJSJumpHandler struct {
	ce    *Engine
	txn   *badger.Txn
	rs    *RoundStates
	rcert *objs.RCert
}

func (jh *dHJSJumpHandler) evalCriteria() bool {
	cond1 := jh.rcert.RClaims.Height <= jh.rs.Height()+1
	cond2 := jh.rs.ValidValueCurrent()
	return cond1 && cond2
}

func (jh *dHJSJumpHandler) evalLogic() error {
	return jh.ce.dHJSJumpFunc(jh.txn, jh.rs, jh.rcert)
}

func (ce *Engine) dHJSJumpFunc(txn *badger.Txn, rs *RoundStates, rcert *objs.RCert) error {
	bhsh, err := rs.ValidValue().PClaims.BClaims.BlockHash()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return err
	}
	if bytes.Equal(bhsh, rcert.RClaims.PrevBlock) && rcert.RClaims.Round == 1 {
		vv := rs.ValidValue()
		err := ce.castNewCommittedBlockFromProposalAndRCert(txn, rs, vv, rcert)
		if err != nil {
			var e *errorz.ErrInvalid
			if err != errorz.ErrMissingTransactions && !errors.As(err, &e) {
				utils.DebugTrace(ce.logger, err)
				return err
			}
		}
		rs.OwnValidatingState.ValidValue = nil
		rs.OwnValidatingState.LockedValue = nil
		return nil
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (ce *Engine) updateValidValue(txn *badger.Txn, rs *RoundStates, p *objs.Proposal) error {
	ce.logger.Debugf("updateValidValue:    MAXBH:%v    STBH:%v    RH:%v    RN:%v", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Height, rs.OwnRoundState().RCert.RClaims.Round)
	txs, _, err := ce.dm.GetTxs(txn, p.PClaims.BClaims.Height, p.TxHshLst)
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
