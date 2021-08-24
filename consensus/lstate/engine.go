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
	"github.com/MadBase/MadNet/consensus/dman"
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

	dm *dman.DMan
}

// Init will initialize the Consensus Engine and all sub modules
func (ce *Engine) Init(database *db.Database, dm *dman.DMan, app appmock.Application, signer *crypto.Secp256k1Signer, adminHandlers *admin.Handlers, publicKey []byte, rbusClient *request.Client) error {
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
		err = ce.dm.FlushCacheToDisk(txn, height)
		if err != nil {
			return err
		}
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
		if roundState.OwnState.SyncToBH.BClaims.Height+1 < roundState.OwnState.MaxBHSeen.BClaims.Height {
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
		if err := ce.dm.CleanCache(txn, height); err != nil {
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
		if maxHR != nil {
			err := ce.doHeightJumpStep(txn, rs, maxHR.RCert)
			if err != nil {
				utils.DebugTrace(ce.logger, err)
				return false, err
			}
			return true, nil
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
		if maxRCert != nil {
			err := ce.doRoundJump(txn, rs, maxRCert)
			if err != nil {
				utils.DebugTrace(ce.logger, err)
				return false, err
			}
			return true, nil
		}
	}

	// if we have voted for the next round in round preceding the
	// dead block round, goto do next round step
	if rs.OwnRoundState().NextRound != nil {
		if rs.OwnRoundState().NRCurrent(rcert) {
			if rcert.RClaims.Round == constants.DEADBLOCKROUNDNR {
				err := ce.doNextRoundStep(txn, rs)
				if err != nil {
					utils.DebugTrace(ce.logger, err)
					return false, err
				}
				return true, nil
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
	if len(NHs) > 0 && !os.NHCurrent(rcert) {
		err := ce.castNextHeightFromNextHeight(txn, rs, NHs[0])
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return false, err
		}
		return true, nil
	}
	if len(NHs) > 0 && os.NHCurrent(rcert) {
		err := ce.doNextHeightStep(txn, rs)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return false, err
		}
		return true, nil
	}

	// determine if there is a round jump
	// if so, make sure it is the max round possible.
	if len(ChFr) > 0 {
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

	// iterate all possibles from nextRound down to proposal
	// and take that action
	ISProposer := rs.LocalIsProposer()
	PCurrent := os.PCurrent(rcert)
	PVCurrent := os.PVCurrent(rcert)
	PVNCurrent := os.PVNCurrent(rcert)
	PCCurrent := os.PCCurrent(rcert)
	PCNCurrent := os.PCNCurrent(rcert)
	NRCurrent := os.NRCurrent(rcert)
	PTOExpired := rs.OwnValidatingState.PTOExpired()
	PVTOExpired := rs.OwnValidatingState.PVTOExpired()
	PCTOExpired := rs.OwnValidatingState.PCTOExpired()

	// dispatch to handlers
	if NRCurrent {
		err := ce.doNextRoundStep(txn, rs)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return false, err
		}
		return true, nil
	}
	if PCCurrent {
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
	if PCNCurrent {
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

	if PVCurrent {
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
	if PVNCurrent {
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
	if PTOExpired {
		err := ce.doPendingPreVoteStep(txn, rs)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return false, err
		}
		return true, nil
	}
	if ISProposer && !PCurrent {
		err := ce.doPendingProposalStep(txn, rs)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return false, err
		}
		return true, nil
	}
	return true, nil
}

func (ce *Engine) Sync() (bool, error) {
	// see if sync is done
	// if yes exit
	syncDone := false
	err := ce.database.Update(func(txn *badger.Txn) error {
		rs, err := ce.sstore.LoadLocalState(txn)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(ce.logger, err)
				return err
			}
			return nil
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
				csbh, err := ce.database.GetSnapshotBlockHeader(txn, canonicalSnapShotHeight)
				if err != nil {
					if err != badger.ErrKeyNotFound {
						utils.DebugTrace(ce.logger, err)
						return err
					}
					return errorz.ErrInvalid{}.New("Snapshot header not available for sync")
				}
				fastSyncDone, err := ce.fastSync.Update(txn, csbh)
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
		txs, bh, ok, err := ce.dm.SyncOneBH(txn, rs.OwnState.SyncToBH, rs.OwnState.MaxBHSeen, rs.ValidatorSet)
		if err != nil {
			utils.DebugTrace(ce.logger, err)
			return err
		}
		if ok {
			ok, err = ce.isValid(txn, rs, bh.BClaims.ChainID, bh.BClaims.StateRoot, bh.BClaims.HeaderRoot, txs)
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
