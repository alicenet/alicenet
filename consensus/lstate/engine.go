package lstate

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"

	"github.com/MadBase/MadNet/consensus/admin"
	"github.com/MadBase/MadNet/consensus/appmock"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/consensus/request"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/logging"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

// Engine is the consensus algorithm parent object.
type Engine struct {
	ctx       context.Context
	cancelCtx func()

	database *db.Database
	sstore   *Store

	RequestBus *request.Client
	appHandler appmock.Application

	logger     *logrus.Logger
	secpSigner *crypto.Secp256k1Signer
	bnSigner   *crypto.BNGroupSigner

	AdminBus *admin.Handlers

	fastSync *SnapShotManager

	ethAcct []byte
	EthPubk []byte

	dm *DMan
}

// Init will initialize the Consensus Engine and all sub modules
func (ce *Engine) Init(database *db.Database, dm *DMan, app appmock.Application, signer *crypto.Secp256k1Signer, adminHandlers *admin.Handlers, publicKey []byte, rbusClient *request.Client) error {
	background := context.Background()
	ctx, cf := context.WithCancel(background)
	ce.cancelCtx = cf
	ce.ctx = ctx
	ce.secpSigner = signer
	ce.database = database
	ce.AdminBus = adminHandlers
	ce.EthPubk = publicKey
	ce.RequestBus = rbusClient
	ce.appHandler = app
	ce.sstore = &Store{}
	err := ce.sstore.Init(database)
	if err != nil {
		return err
	}
	ce.dm = dm
	if len(ce.EthPubk) > 0 {
		ce.ethAcct = crypto.GetAccount(ce.EthPubk)
	}
	ce.logger = logging.GetLogger(constants.LoggerConsensus)
	ce.fastSync = &SnapShotManager{
		appHandler: app,
		requestBus: ce.RequestBus,
	}
	if err := ce.fastSync.Init(database); err != nil {
		return err
	}
	return nil
}

// Status .
func (ce *Engine) Status(status map[string]interface{}) (map[string]interface{}, error) {
	var rs *RoundStates
	err := ce.database.View(func(txn *badger.Txn) error {
		rss, err := ce.sstore.LoadLocalState(txn)
		if err != nil {
			return err
		}
		rs = rss
		return nil
	})
	if err != nil {
		return nil, err
	}
	bhsh, err := rs.OwnState.SyncToBH.BlockHash()
	if err != nil {
		return nil, err
	}
	if rs.OwnState.MaxBHSeen.BClaims.Height-rs.OwnState.SyncToBH.BClaims.Height < 2 {
		status[constants.StatusBlkRnd] = fmt.Sprintf("%d/%d", rs.OwnState.SyncToBH.BClaims.Height, rs.OwnRoundState().RCert.RClaims.Round)
		status[constants.StatusBlkHsh] = fmt.Sprintf("%x..%x", bhsh[0:2], bhsh[len(bhsh)-2:])
		status[constants.StatusTxCt] = rs.OwnState.SyncToBH.BClaims.TxCount
		return status, nil
	}
	status[constants.StatusBlkRnd] = fmt.Sprintf("%d/%v", rs.OwnState.MaxBHSeen.BClaims.Height, "-")
	status[constants.StatusBlkHsh] = fmt.Sprintf("%x..%x", bhsh[0:2], bhsh[len(bhsh)-2:])
	status[constants.StatusSyncToBlk] = fmt.Sprintf("%d", rs.OwnState.SyncToBH.BClaims.Height)
	return status, nil
}

// UpdateLocalState .
func (ce *Engine) UpdateLocalState() (bool, error) {
	var isSync bool
	updateLocalState := true
	// ce.logger.Error("!!! OPEN UpdateLocalState TXN")
	// defer func() { ce.logger.Error("!!! CLOSE UpdateLocalState TXN") }()
	err := ce.database.Update(func(txn *badger.Txn) error {
		ownState, err := ce.database.GetOwnState(txn)
		if err != nil {
			return err
		}
		height, _ := objs.ExtractHR(ownState.SyncToBH)
		vs, err := ce.database.GetValidatorSet(txn, height)
		if err != nil {
			return err
		}
		if !bytes.Equal(vs.GroupKey, ownState.GroupKey) {
			ownState.GroupKey = vs.GroupKey
			err = ce.database.SetOwnState(txn, ownState)
			if err != nil {
				return err
			}
		}
		ownValidatingState, err := ce.database.GetOwnValidatingState(txn)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
		}
		if ownValidatingState == nil {
			ownValidatingState = &objs.OwnValidatingState{}
		}
		if !bytes.Equal(ownValidatingState.VAddr, ownState.VAddr) || !bytes.Equal(ownValidatingState.GroupKey, ownState.GroupKey) {
			ovs := &objs.OwnValidatingState{
				VAddr:    ownState.VAddr,
				GroupKey: ownState.GroupKey,
			}
			ovs.SetRoundStarted()
			err := ce.database.SetOwnValidatingState(txn, ovs)
			if err != nil {
				return err
			}
		}
		roundState, err := ce.sstore.LoadLocalState(txn)
		if err != nil {
			return err
		}
		if roundState.OwnState.SyncToBH.BClaims.Height%constants.EpochLength == 0 {
			safe, err := ce.database.GetSafeToProceed(txn, roundState.OwnState.SyncToBH.BClaims.Height)
			if err != nil {
				utils.DebugTrace(ce.logger, err)
				return err
			}
			if !safe {
				utils.DebugTrace(ce.logger, nil, "not safe")
				updateLocalState = false
			}
		}
		if roundState.OwnState.SyncToBH.BClaims.Height < roundState.OwnState.MaxBHSeen.BClaims.Height {
			isSync = false
			updateLocalState = false
		}
		if updateLocalState {
			ok, err := ce.updateLocalStateInternal(txn, roundState)
			if err != nil {
				return err
			}
			isSync = ok
		}
		err = ce.sstore.WriteState(txn, roundState)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		return nil
	})
	if err != nil {
		e := errorz.ErrInvalid{}.New("")
		if !errors.As(err, &e) && err != errorz.ErrMissingTransactions {
			return false, err
		}
		return false, nil
	}
	err = ce.database.Sync()
	if err != nil {
		return false, err
	}
	return isSync, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// This changes state for local node
// order of ops is as follows:
//
//  check for height jump
//  check for dead block round round jump
//  follow dead block round next round if signed by self
//  follow f+1 other dead block round next round messages
//  do own next height if in dead block round
//  follow a next height from any round in the same height as us
//      this is safe due to how we count next heights to filter
//      dead block round
//  follow a round jump to any non dead block round
//  do a possible next round in same round
//  do a possible precommit/pendingnext in same round
//  do a possible precommit/pendingnext in same round
//  do a possible prevote/pendingprecommit in same round
//  do a possible prevotenil/pendingprecommit in same round
//  do a possible pending prevote
//  do a possible do a proposal if not already proposed and is proposer
//  do nothing if not any of above is true
func (ce *Engine) updateLocalStateInternal(txn *badger.Txn, rs *RoundStates) (bool, error) {
	if err := ce.loadValidationKey(rs); err != nil {
		return false, nil
	}
	os := rs.OwnRoundState()

	// extract the round cert for use
	rcert := os.RCert

	// create three vectors that may overlap
	// these vectors sort all current validators by height/round
	// as is determined by their respective rcert
	// these vectors are:
	//  Current height future round
	//  Current height any round
	//  Future height any round
	ChFr := []*objs.RoundState{}
	FH := []*objs.RoundState{}
	for i := 0; i < len(rs.ValidatorSet.Validators); i++ {
		vObj := rs.ValidatorSet.Validators[i]
		vAddr := vObj.VAddr
		vroundState := rs.PeerStateMap[string(vAddr)]
		relationH := objs.RelateH(rcert, vroundState.RCert)
		if relationH == 0 {
			relationHR := objs.RelateHR(rcert, vroundState.RCert)
			if relationHR == -1 {
				ChFr = append(ChFr, vroundState)
			}
		} else if relationH == -1 {
			FH = append(FH, vroundState)
		}
	}

	// if there are ANY peers in a future height, try to follow
	// we should try to follow the max height possible
	if len(FH) > 0 {
		var maxHR *objs.RoundState
		maxHeight := uint32(0)
		for _, vroundState := range FH {
			if vroundState.RCert.RClaims.Height > maxHeight {
				maxHR = vroundState
				maxHeight = vroundState.RCert.RClaims.Height
			}
		}
		currentHandler := fhHandler{ce: ce, txn: txn, rs: rs, maxHR: maxHR}
		if currentHandler.evalCriteria() {
			return currentHandler.evalLogic()
		}
	}
	// at this point no height jump is possible
	// otherwise we would have followed it

	////////////////////////////////////////////////////////////////////////////////

	// look for a round certificate from the dead block round
	// if one exists, we should follow it
	if len(ChFr) > 0 {
		var maxRCert *objs.RCert
		for _, vroundState := range ChFr {
			if vroundState.RCert.RClaims.Round == constants.DEADBLOCKROUND {
				maxRCert = vroundState.RCert
				break
			}
		}
		currentHandler := rCertHandler{ce: ce, txn: txn, rs: rs, maxRCert: maxRCert}
		if currentHandler.evalCriteria() {
			return currentHandler.evalLogic()
		}
	}

	// if we have voted for the next round in round preceding the
	// dead block round, goto do next round step
	if rs.OwnRoundState().NextRound != nil {
		if rs.OwnRoundState().NRCurrent(rcert) {
			currentHandler := doNrHandler{ce: ce, txn: txn, rs: rs, rcert: rcert}
			if currentHandler.evalCriteria() {
				return currentHandler.evalLogic()
			}
		}
	}

	////////////////////////////////////////////////////////////////////////////////

	// check for next height messages
	// if one exists, follow it
	NHs, _, err := rs.GetCurrentNext()
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	NHCurrent := os.NHCurrent(rcert)
	var currentHandler handler
	currentHandler = castNhHandler{ce: ce, txn: txn, rs: rs, NHs: NHs, NHCurrent: NHCurrent}
	if currentHandler.evalCriteria() {
		return currentHandler.evalLogic()
	}

	currentHandler = doNextHeightHandler{ce: ce, txn: txn, rs: rs, NHs: NHs, NHCurrent: NHCurrent}
	if currentHandler.evalCriteria() {
		return currentHandler.evalLogic()
	}

	// determine if there is a round jump
	// if so, make sure it is the max round possible.
	currentHandler = doRoundJumpHandler{ce: ce, txn: txn, rs: rs, ChFr: ChFr}
	if currentHandler.evalCriteria() {
		return currentHandler.evalLogic()
	}

	// iterate all possibles from nextRound down to proposal
	// and take that action
	PCurrent := os.PCurrent(rcert)
	PVCurrent := os.PVCurrent(rcert)
	PVNCurrent := os.PVNCurrent(rcert)
	PCCurrent := os.PCCurrent(rcert)
	PCNCurrent := os.PCNCurrent(rcert)
	NRCurrent := os.NRCurrent(rcert)

	//
	//	Order:
	//		NR, PC, PCN, PV, PVN, ProposalTO, IsProposer

	ceHandlers := []handler{
		nrCurrentHandler{ce: ce, txn: txn, rs: rs, NRCurrent: NRCurrent},
		pcCurrentHandler{ce: ce, txn: txn, rs: rs, PCCurrent: PCCurrent},
		pcnCurrentHandler{ce: ce, txn: txn, rs: rs, PCNCurrent: PCNCurrent},
		pvCurrentHandler{ce: ce, txn: txn, rs: rs, PVCurrent: PVCurrent},
		pvnCurrentHandler{ce: ce, txn: txn, rs: rs, PVNCurrent: PVNCurrent},
		ptoExpiredHandler{ce: ce, txn: txn, rs: rs},
		validPropHandler{ce: ce, txn: txn, rs: rs, PCurrent: PCurrent},
	}

	for i := 0; i < len(ceHandlers); i++ {
		if ceHandlers[i].evalCriteria() {
			return ceHandlers[i].evalLogic()
		}
	}

	return true, nil
}

type handler interface {
	evalCriteria() bool
	evalLogic() (bool, error)
}

type fhHandler struct {
	ce    *Engine
	txn   *badger.Txn
	rs    *RoundStates
	maxHR *objs.RoundState
}

func (fhh fhHandler) evalCriteria() bool {
	return fhh.maxHR != nil
}

func (fhh fhHandler) evalLogic() (bool, error) {
	return fhh.ce.fhFunc(fhh.txn, fhh.rs, fhh.maxHR)
}

func (ce *Engine) fhFunc(txn *badger.Txn, rs *RoundStates, maxHR *objs.RoundState) (bool, error) {
	err := ce.doHeightJumpStep(txn, rs, maxHR.RCert)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

type rCertHandler struct {
	ce       *Engine
	txn      *badger.Txn
	rs       *RoundStates
	maxRCert *objs.RCert
}

func (rch rCertHandler) evalCriteria() bool {
	return rch.maxRCert != nil
}

func (rch rCertHandler) evalLogic() (bool, error) {
	return rch.ce.rCertFunc(rch.txn, rch.rs, rch.maxRCert)
}

func (ce *Engine) rCertFunc(txn *badger.Txn, rs *RoundStates, maxRCert *objs.RCert) (bool, error) {
	err := ce.doRoundJump(txn, rs, maxRCert)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

type doNrHandler struct {
	ce    *Engine
	txn   *badger.Txn
	rs    *RoundStates
	rcert *objs.RCert
}

func (dnrh doNrHandler) evalCriteria() bool {
	return dnrh.rcert.RClaims.Round == constants.DEADBLOCKROUNDNR
}

func (dnrh doNrHandler) evalLogic() (bool, error) {
	return dnrh.ce.doNrFunc(dnrh.txn, dnrh.rs, dnrh.rcert)
}

func (ce *Engine) doNrFunc(txn *badger.Txn, rs *RoundStates, rcert *objs.RCert) (bool, error) {
	err := ce.doNextRoundStep(txn, rs)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

type castNhHandler struct {
	ce        *Engine
	txn       *badger.Txn
	rs        *RoundStates
	NHs       objs.NextHeightList
	NHCurrent bool
}

func (cnhh castNhHandler) evalCriteria() bool {
	return len(cnhh.NHs) > 0 && !cnhh.NHCurrent
}

func (cnhh castNhHandler) evalLogic() (bool, error) {
	return cnhh.ce.castNhFunc(cnhh.txn, cnhh.rs, cnhh.NHs)
}

func (ce *Engine) castNhFunc(txn *badger.Txn, rs *RoundStates, NHs objs.NextHeightList) (bool, error) {
	err := ce.castNextHeightFromNextHeight(txn, rs, NHs[0])
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

type doNextHeightHandler struct {
	ce        *Engine
	txn       *badger.Txn
	rs        *RoundStates
	NHs       objs.NextHeightList
	NHCurrent bool
}

func (dnhh doNextHeightHandler) evalCriteria() bool {
	return len(dnhh.NHs) > 0 && dnhh.NHCurrent
}

func (dnhh doNextHeightHandler) evalLogic() (bool, error) {
	return dnhh.ce.doNextHeightFunc(dnhh.txn, dnhh.rs, dnhh.NHs)
}

func (ce *Engine) doNextHeightFunc(txn *badger.Txn, rs *RoundStates, NHs objs.NextHeightList) (bool, error) {
	err := ce.doNextHeightStep(txn, rs)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

type doRoundJumpHandler struct {
	ce   *Engine
	txn  *badger.Txn
	rs   *RoundStates
	ChFr []*objs.RoundState
}

func (drjh doRoundJumpHandler) evalCriteria() bool {
	return len(drjh.ChFr) > 0
}

func (drjh doRoundJumpHandler) evalLogic() (bool, error) {
	return drjh.ce.doRoundJumpFunc(drjh.txn, drjh.rs, drjh.ChFr)
}

func (ce *Engine) doRoundJumpFunc(txn *badger.Txn, rs *RoundStates, ChFr []*objs.RoundState) (bool, error) {
	var maxRCert *objs.RCert
	for _, vroundState := range ChFr {
		if maxRCert == nil {
			maxRCert = vroundState.RCert
		}
		if vroundState.RCert.RClaims.Round > maxRCert.RClaims.Round {
			maxRCert = vroundState.RCert
		}
	}
	err := ce.doRoundJump(txn, rs, maxRCert)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

type nrCurrentHandler struct {
	ce        *Engine
	txn       *badger.Txn
	rs        *RoundStates
	NRCurrent bool
}

func (nrch nrCurrentHandler) evalCriteria() bool {
	return nrch.NRCurrent
}

func (nrch nrCurrentHandler) evalLogic() (bool, error) {
	return nrch.ce.nrCurrentFunc(nrch.txn, nrch.rs)
}

func (ce *Engine) nrCurrentFunc(txn *badger.Txn, rs *RoundStates) (bool, error) {
	err := ce.doNextRoundStep(txn, rs)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

type pcCurrentHandler struct {
	ce        *Engine
	txn       *badger.Txn
	rs        *RoundStates
	PCCurrent bool
}

func (pcch pcCurrentHandler) evalCriteria() bool {
	return pcch.PCCurrent
}

func (pcch pcCurrentHandler) evalLogic() (bool, error) {
	PCTOExpired := pcch.rs.OwnValidatingState.PCTOExpired()
	return pcch.ce.pcCurrentFunc(pcch.txn, pcch.rs, PCTOExpired)
}

func (ce *Engine) pcCurrentFunc(txn *badger.Txn, rs *RoundStates, PCTOExpired bool) (bool, error) {
	if PCTOExpired {
		err := ce.doPendingNext(txn, rs)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return false, err
		}
		return true, nil
	}
	err := ce.doPreCommitStep(txn, rs)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

type pcnCurrentHandler struct {
	ce         *Engine
	txn        *badger.Txn
	rs         *RoundStates
	PCNCurrent bool
}

func (pcnch pcnCurrentHandler) evalCriteria() bool {
	return pcnch.PCNCurrent
}

func (pcnch pcnCurrentHandler) evalLogic() (bool, error) {
	PCTOExpired := pcnch.rs.OwnValidatingState.PCTOExpired()
	return pcnch.ce.pcnCurrentFunc(pcnch.txn, pcnch.rs, PCTOExpired)
}

func (ce *Engine) pcnCurrentFunc(txn *badger.Txn, rs *RoundStates, PCTOExpired bool) (bool, error) {
	if PCTOExpired {
		err := ce.doPendingNext(txn, rs)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return false, err
		}
		return true, nil
	}
	err := ce.doPreCommitNilStep(txn, rs)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

type pvCurrentHandler struct {
	ce        *Engine
	txn       *badger.Txn
	rs        *RoundStates
	PVCurrent bool
}

func (pvch pvCurrentHandler) evalCriteria() bool {
	return pvch.PVCurrent
}

func (pvch pvCurrentHandler) evalLogic() (bool, error) {
	PVTOExpired := pvch.rs.OwnValidatingState.PVTOExpired()
	return pvch.ce.pvCurrentFunc(pvch.txn, pvch.rs, PVTOExpired)
}

func (ce *Engine) pvCurrentFunc(txn *badger.Txn, rs *RoundStates, PVTOExpired bool) (bool, error) {
	if PVTOExpired {
		err := ce.doPendingPreCommit(txn, rs)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return false, err
		}
		return true, nil
	}
	err := ce.doPreVoteStep(txn, rs)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

type pvnCurrentHandler struct {
	ce         *Engine
	txn        *badger.Txn
	rs         *RoundStates
	PVNCurrent bool
}

func (pvnch pvnCurrentHandler) evalCriteria() bool {
	return pvnch.PVNCurrent
}

func (pvnch pvnCurrentHandler) evalLogic() (bool, error) {
	PVTOExpired := pvnch.rs.OwnValidatingState.PVTOExpired()
	return pvnch.ce.pvnCurrentFunc(pvnch.txn, pvnch.rs, PVTOExpired)
}

func (ce *Engine) pvnCurrentFunc(txn *badger.Txn, rs *RoundStates, PVTOExpired bool) (bool, error) {
	if PVTOExpired {
		err := ce.doPendingPreCommit(txn, rs)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return false, err
		}
		return true, nil
	}
	err := ce.doPreVoteNilStep(txn, rs)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

type ptoExpiredHandler struct {
	ce  *Engine
	txn *badger.Txn
	rs  *RoundStates
}

func (ptoeh ptoExpiredHandler) evalCriteria() bool {
	PTOExpired := ptoeh.rs.OwnValidatingState.PTOExpired()
	return PTOExpired
}

func (ptoeh ptoExpiredHandler) evalLogic() (bool, error) {
	return ptoeh.ce.ptoExpiredFunc(ptoeh.txn, ptoeh.rs)
}

func (ce *Engine) ptoExpiredFunc(txn *badger.Txn, rs *RoundStates) (bool, error) {
	err := ce.doPendingPreVoteStep(txn, rs)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

type validPropHandler struct {
	ce       *Engine
	txn      *badger.Txn
	rs       *RoundStates
	PCurrent bool
}

func (vph validPropHandler) evalCriteria() bool {
	IsProposer := vph.rs.LocalIsProposer()
	return IsProposer && !vph.PCurrent
}

func (vph validPropHandler) evalLogic() (bool, error) {
	return vph.ce.validPropFunc(vph.txn, vph.rs)
}

func (ce *Engine) validPropFunc(txn *badger.Txn, rs *RoundStates) (bool, error) {
	err := ce.doPendingProposalStep(txn, rs)
	if err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return true, nil
}

// Sync .
func (ce *Engine) Sync() (bool, error) {
	// see if sync is done
	// if yes exit
	syncDone := false
	// ce.logger.Error("!!! OPEN SYNC TXN")
	// defer func() { ce.logger.Error("!!! CLOSE SYNC TXN") }()
	err := ce.database.Update(func(txn *badger.Txn) error {
		rs, err := ce.sstore.LoadLocalState(txn)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		// begin handling logic
		if rs.OwnState.MaxBHSeen.BClaims.Height == rs.OwnState.SyncToBH.BClaims.Height {
			syncDone = true
			return nil
		}
		if rs.OwnState.MaxBHSeen.BClaims.Height > constants.EpochLength*2 {
			if rs.OwnState.SyncToBH.BClaims.Height <= rs.OwnState.MaxBHSeen.BClaims.Height-constants.EpochLength*2 {
				// Guard against the short first epoch causing errors in the sync logic
				// by escaping early and just waiting for the MaxBHSeen to increase.
				if rs.OwnState.MaxBHSeen.BClaims.Height%constants.EpochLength == 0 {
					return nil
				}
				epochOfMaxBHSeen := utils.Epoch(rs.OwnState.MaxBHSeen.BClaims.Height)
				canonicalEpoch := epochOfMaxBHSeen - 2
				canonicalSnapShotHeight := canonicalEpoch * constants.EpochLength
				mrcbh, err := ce.database.GetMostRecentCommittedBlockHeaderFastSync(txn)
				if err != nil {
					utils.DebugTrace(ce.logger, err)
					return err
				}
				csbh, err := ce.database.GetSnapshotBlockHeader(txn, canonicalSnapShotHeight)
				if err != nil {
					utils.DebugTrace(ce.logger, err)
					return err
				}
				canonicalBlockHash, err := csbh.BlockHash()
				if err != nil {
					utils.DebugTrace(ce.logger, err)
					return err
				}
				fastSyncDone, err := ce.fastSync.Update(txn, csbh.BClaims.Height, mrcbh.BClaims.Height, csbh.BClaims.StateRoot, csbh.BClaims.HeaderRoot, canonicalBlockHash)
				if err != nil {
					utils.DebugTrace(ce.logger, err)
					return err
				}
				if fastSyncDone {
					if err := ce.setMostRecentBlockHeaderFastSync(txn, rs, csbh); err != nil {
						utils.DebugTrace(ce.logger, err)
						return err
					}
					err = ce.sstore.WriteState(txn, rs)
					if err != nil {
						utils.DebugTrace(ce.logger, err)
						return err
					}
					return nil
				}
				return nil
			}
		}
		ce.logger.Debugf("SyncOneBH:  MBHS:%v  STBH:%v", rs.OwnState.MaxBHSeen.BClaims.Height, rs.OwnState.SyncToBH.BClaims.Height)
		txs, bh, err := ce.dm.SyncOneBH(txn, rs)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		ok, err := ce.isValid(txn, rs, bh.BClaims.ChainID, bh.BClaims.StateRoot, bh.BClaims.HeaderRoot, txs)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		if !ok {
			return nil
		}
		err = ce.setMostRecentBlockHeader(txn, rs, bh)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		err = ce.sstore.WriteState(txn, rs)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		return nil
	})
	if err != nil {
		e := errorz.ErrInvalid{}.New("")
		if errors.As(err, &e) {
			utils.DebugTrace(ce.logger, err)
			return false, nil
		}
		if err == errorz.ErrMissingTransactions {
			utils.DebugTrace(ce.logger, err)
			return false, nil
		}
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	if err := ce.database.Sync(); err != nil {
		utils.DebugTrace(ce.logger, err)
		return false, err
	}
	return syncDone, nil
}

func (ce *Engine) loadValidationKey(rs *RoundStates) error {
	if rs.IsCurrentValidator() {
		if !bytes.Equal(rs.ValidatorSet.GroupKey, rs.OwnValidatingState.GroupKey) || ce.bnSigner == nil {
			for _, v := range rs.ValidatorSet.Validators {
				if bytes.Equal(v.VAddr, rs.OwnState.VAddr) {
					name := make([]byte, len(v.GroupShare))
					copy(name[:], v.GroupShare)
					pk, err := ce.AdminBus.GetPrivK(name)
					if err != nil {
						utils.DebugTrace(ce.logger, err)
						return nil
					}
					signer := &crypto.BNGroupSigner{}
					signer.SetPrivk(pk)
					err = signer.SetGroupPubk(rs.ValidatorSet.GroupKey)
					if err != nil {
						utils.DebugTrace(ce.logger, err)
						return err
					}
					ce.bnSigner = signer
					pubk, err := ce.bnSigner.PubkeyShare()
					if err != nil {
						return err
					}
					if !bytes.Equal(name, pubk) {
						utils.DebugTrace(ce.logger, err)
						return err
					}
					break
				}
			}
			rs.OwnValidatingState.GroupKey = rs.ValidatorSet.GroupKey
		}
	}
	return nil
}
