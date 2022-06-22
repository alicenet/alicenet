package lstate

import (
	"bytes"
	"fmt"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/dman"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

type Handlers struct {
	database *db.Database
	sstore   *Store
	secpVal  *crypto.Secp256k1Validator
	bnVal    *crypto.BNGroupValidator
	dm       *dman.DMan
	logger   *logrus.Logger
}

// Init initializes the Handlers object
func (mb *Handlers) Init(database *db.Database, dm *dman.DMan) {
	mb.logger = logging.GetLogger(constants.LoggerConsensus)
	mb.sstore = &Store{}
	mb.sstore.Init(database)
	mb.database = database
	mb.dm = dm
}

// AddProposal stores a proposal to the database
func (mb *Handlers) AddProposal(v *objs.Proposal) error {
	return mb.Store(v)
}

// AddPreVote stores a preVote to the database
func (mb *Handlers) AddPreVote(v *objs.PreVote) error {
	return mb.Store(v)
}

// AddPreVoteNil stores a preVoteNil to the database
func (mb *Handlers) AddPreVoteNil(v *objs.PreVoteNil) error {
	return mb.Store(v)
}

// AddPreCommit stores a preCommit to the database
func (mb *Handlers) AddPreCommit(v *objs.PreCommit) error {
	return mb.Store(v)
}

// AddPreCommitNil stores a preCommitNil to the database
func (mb *Handlers) AddPreCommitNil(v *objs.PreCommitNil) error {
	return mb.Store(v)
}

// AddNextRound stores a nextRound object to the database
func (mb *Handlers) AddNextRound(v *objs.NextRound) error {
	return mb.Store(v)
}

// AddNextHeight stores a nextHeight object to the database
func (mb *Handlers) AddNextHeight(v *objs.NextHeight) error {
	return mb.Store(v)
}

// AddBlockHeader stores a blockHeader object to the database
func (mb *Handlers) AddBlockHeader(v *objs.BlockHeader) error {
	return mb.Store(v)
}

// Store updates database to include the object; it stores the object.
func (mb *Handlers) Store(v interface{}) error {
	return mb.database.Update(func(txn *badger.Txn) error {
		rc, err := objs.ExtractRCertAny(v)
		if err != nil {
			return err
		}
		roundState, err := mb.sstore.LoadState(txn, rc)
		if err != nil {
			return err
		}
		switch obj := v.(type) {
		case *objs.Proposal:
			txHshLst := obj.TxHshLst
			err = roundState.SetProposal(obj)
			if err != nil {
				return err
			}
			// if !roundState.IsCurrentValidator() {
			// 	err := mb.database.SetBroadcastProposal(txn, obj)
			// 	if err != nil {
			// 		return err
			// 	}
			// }
			go mb.dm.DownloadTxs(roundState.height, roundState.round, txHshLst)
		case *objs.PreVote:
			err = roundState.SetPreVote(obj)
			if err != nil {
				return err
			}
			// if !roundState.IsCurrentValidator() {
			// err := mb.database.SetBroadcastPreVote(txn, obj)
			// if err != nil {
			// return err
			// }
			// }
		case *objs.PreVoteNil:
			err = roundState.SetPreVoteNil(obj)
			if err != nil {
				return err
			}
			// if !roundState.IsCurrentValidator() {
			// err := mb.database.SetBroadcastPreVoteNil(txn, obj)
			// if err != nil {
			// return err
			// }
			// }
		case *objs.PreCommit:
			err = roundState.SetPreCommit(obj)
			if err != nil {
				return err
			}
			// if !roundState.IsCurrentValidator() {
			// 	err := mb.database.SetBroadcastPreCommit(txn, obj)
			// 	if err != nil {
			// 		return err
			// 	}
			// }
		case *objs.PreCommitNil:
			err = roundState.SetPreCommitNil(obj)
			if err != nil {
				return err
			}
			// if !roundState.IsCurrentValidator() {
			// 	err := mb.database.SetBroadcastPreCommitNil(txn, obj)
			// 	if err != nil {
			// 		return err
			// 	}
			// }
		case *objs.NextRound:
			err = roundState.SetNextRound(obj)
			if err != nil {
				return err
			}
			// if !roundState.IsCurrentValidator() {
			// 	err := mb.database.SetBroadcastNextRound(txn, obj)
			// 	if err != nil {
			// 		return err
			// 	}
			// }
		case *objs.NextHeight:
			err = roundState.SetNextHeight(obj)
			if err != nil {
				return err
			}
			// if !roundState.IsCurrentValidator() {
			// 	err := mb.database.SetBroadcastNextHeight(txn, obj)
			// 	if err != nil {
			// 		return err
			// 	}
			// }
		case *objs.BlockHeader:
			ownState := roundState.OwnState
			if obj.BClaims.Height <= ownState.MaxBHSeen.BClaims.Height {
				return errorz.ErrInvalid{}.New("stale bh  - <= MaxBHSeen")
			}
			if obj.BClaims.Height <= ownState.SyncToBH.BClaims.Height {
				return errorz.ErrInvalid{}.New("stale bh - <= SyncTOBH ")
			}
			ownState.MaxBHSeen = obj
		}
		return mb.sstore.WriteState(txn, roundState)
	})
}

// PreValidate checks a message for validity and performs cryptographic
// validation
func (mb *Handlers) PreValidate(v interface{}) error {
	var Voter []byte
	var Proposer []byte
	var GroupShare []byte
	var GroupKey []byte
	var CoSigners [][]byte
	var round uint32
	_, chainID := objs.ExtractHCID(v)
	err := mb.database.View(func(txn *badger.Txn) error {
		os, err := mb.database.GetOwnState(txn)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		rs, err := mb.database.GetCurrentRoundState(txn, os.VAddr)
		if err != nil {
			utils.DebugTrace(mb.logger, err)
			return err
		}
		cid := rs.RCert.RClaims.ChainID
		if cid != chainID {
			return errorz.ErrInvalid{}.New("cid mismatch")
		}
		h := rs.RCert.RClaims.Height
		r := rs.RCert.RClaims.Round
		var height uint32
		switch obj := v.(type) {
		case *objs.Proposal:
			height = obj.PClaims.RCert.RClaims.Height
			if height < h {
				return errorz.ErrStale{}.New("Proposal h<h-1: OwnH:%v ObjH:%v", h, height)
			}
			round := obj.PClaims.RCert.RClaims.Round
			if round < r {
				errorz.ErrStale{}.New("Proposal r<r-1: OwnR:%v ObjR:%v", r, round)
			}
			if err := obj.ValidateSignatures(mb.secpVal, mb.bnVal); err != nil {
				return err
			}
			//Voter = nil
			Proposer = obj.Proposer
			//GroupShare = nil
			GroupKey = obj.GroupKey
			//CoSigners = nil
			round = obj.PClaims.RCert.RClaims.Round

			if err := mb.ProposalValidation(txn, GroupKey, Proposer, mb.subOneNoZero(height), height, round); err != nil {
				return err
			}
			return nil

		case *objs.PreVote:
			height = obj.Proposal.PClaims.RCert.RClaims.Height
			if height < h {
				return errorz.ErrStale{}.New("PreVote h<h-1: OwnH:%v ObjH:%v", h, height)
			}
			round = obj.Proposal.PClaims.RCert.RClaims.Round
			if round < r {
				errorz.ErrStale{}.New("PreVote r<r-1: OwnR:%v ObjR:%v", r, round)
			}
			if err := obj.ValidateSignatures(mb.secpVal, mb.bnVal); err != nil {
				return err
			}
			Voter = obj.Voter
			Proposer = obj.Proposal.Proposer
			//GroupShare = nil
			GroupKey = obj.GroupKey
			//CoSigners = nil

			if err := mb.VoteValidation(txn, GroupKey, Voter, Proposer, nil, mb.subOneNoZero(height), height, round); err != nil {
				return err
			}
			return nil

		case *objs.PreVoteNil:
			height = obj.RCert.RClaims.Height
			if height < h {
				return errorz.ErrStale{}.New("PreVoteNil h<h-1: OwnH:%v ObjH:%v", h, height)
			}
			round = obj.RCert.RClaims.Round
			if round < r {
				errorz.ErrStale{}.New("PreVoteNil r<r-1: OwnR:%v ObjR:%v", r, round)
			}
			if err := obj.ValidateSignatures(mb.secpVal, mb.bnVal); err != nil {
				return err
			}
			Voter = obj.Voter
			//Proposer = nil
			//GroupShare = nil
			GroupKey = obj.GroupKey
			//CoSigners = nil
			if err := mb.VoteNilValidation(txn, GroupKey, Voter, mb.subOneNoZero(height), height, round); err != nil {
				return err
			}
			return nil

		case *objs.PreCommit:
			height = obj.Proposal.PClaims.RCert.RClaims.Height
			if height < h {
				return errorz.ErrStale{}.New("PreCommit h<h-1: OwnH:%v ObjH:%v", h, height)
			}
			round = obj.Proposal.PClaims.RCert.RClaims.Round
			if round < r {
				errorz.ErrStale{}.New("PreCommit r<r-1: OwnR:%v ObjR:%v", r, round)
			}
			if err := obj.ValidateSignatures(mb.secpVal, mb.bnVal); err != nil {
				return err
			}
			Voter = obj.Voter
			Proposer = obj.Proposer
			//GroupShare = nil
			GroupKey = obj.GroupKey
			CoSigners = obj.Signers

			if err := mb.VoteValidation(txn, GroupKey, Voter, Proposer, CoSigners, mb.subOneNoZero(height), height, round); err != nil {
				return err
			}
			return nil

		case *objs.PreCommitNil:
			height = obj.RCert.RClaims.Height
			if height < h {
				return errorz.ErrStale{}.New("PreCommitNil h<h-1: OwnH:%v ObjH:%v", h, height)
			}
			round = obj.RCert.RClaims.Round
			if round < r {
				errorz.ErrStale{}.New("PreCommitNil r<r-1: OwnR:%v ObjR:%v", r, round)
			}
			if err := obj.ValidateSignatures(mb.secpVal, mb.bnVal); err != nil {
				return err
			}
			Voter = obj.Voter
			//Proposer = nil
			//GroupShare = nil
			GroupKey = obj.GroupKey
			//CoSigners = nil
			if err := mb.VoteNilValidation(txn, GroupKey, Voter, mb.subOneNoZero(height), height, round); err != nil {
				return err
			}
			return nil

		case *objs.NextRound:
			height = obj.NRClaims.RCert.RClaims.Height
			if height < h {
				return errorz.ErrStale{}.New("NextRound h<h-1: OwnH:%v ObjH:%v", h, height)
			}
			round = obj.NRClaims.RCert.RClaims.Round
			if round < r {
				errorz.ErrStale{}.New("NextRound r<r-1: OwnR:%v ObjR:%v", r, round)
			}
			if err := obj.ValidateSignatures(mb.secpVal, mb.bnVal); err != nil {
				return err
			}
			Voter = obj.Voter
			//Proposer = nil
			GroupShare = obj.GroupShare
			GroupKey = obj.GroupKey
			//CoSigners = nil
			if err := mb.VoteNextValidation(txn, GroupKey, GroupShare, Voter, Proposer, nil, mb.subOneNoZero(height), height, round); err != nil {
				return err
			}
			return nil

		case *objs.NextHeight:
			height = obj.NHClaims.Proposal.PClaims.RCert.RClaims.Height
			if height < h {
				return errorz.ErrStale{}.New("NextHeight h<h-1: OwnH:%v ObjH:%v", h, height)
			}
			if err := obj.ValidateSignatures(mb.secpVal, mb.bnVal); err != nil {
				return err
			}
			Voter = obj.Voter
			Proposer = obj.NHClaims.Proposal.Proposer
			GroupShare = obj.GroupShare
			GroupKey = obj.GroupKey
			CoSigners = obj.Signers
			round = obj.NHClaims.Proposal.PClaims.RCert.RClaims.Round
			if err := mb.VoteNextValidation(txn, GroupKey, GroupShare, Voter, Proposer, CoSigners, mb.subOneNoZero(height), height, round); err != nil {
				return err
			}
			return nil

		case *objs.BlockHeader:
			height = obj.BClaims.Height
			if err := obj.ValidateSignatures(mb.bnVal); err != nil {
				return err
			}
			//Voter = nil
			//Proposer = nil
			//GroupShare = nil
			GroupKey = obj.GroupKey
			//CoSigners = nil
			round = 1
			if err := mb.BlockHeaderValidation(txn, GroupKey, height, height+1, round); err != nil {
				return err
			}
			return nil

		default:
			panic("Unknown type")
		}
	})
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return err
	}
	return nil
}

func (mb *Handlers) ValidateRCERT(txn *badger.Txn, groupKey []byte, bHeight uint32, rHeight uint32, rNumber uint32) error {
	if rHeight <= 2 {
		return nil
	}
	height := normalizeHeightForRCERT(bHeight, rHeight, rNumber)
	vSet, err := mb.database.GetValidatorSet(txn, height)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return err
	}
	if !bytes.Equal(vSet.GroupKey, groupKey) {
		return errorz.ErrInvalid{}.New("bad rcert group key")
	}
	return nil
}

func (mb *Handlers) ProposalValidation(txn *badger.Txn, groupKey []byte, proposer []byte, bHeight uint32, rHeight uint32, rNumber uint32) error {
	// OBJECTS	GROUPKEY	VOTER	GROUPSHARE	PROPOSER	HEIGHT	ROUND_1VSET		ROUND>1VSET
	// PROPOSAL	TRUE	 	FALSE	FALSE		TRUE		N+1	    N	            N+1
	if err := mb.ValidateRCERT(txn, groupKey, bHeight, rHeight, rNumber); err != nil {
		return err
	}
	vSet, err := mb.database.GetValidatorSet(txn, rHeight)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return err
	}
	if !vSet.IsValidTuple(proposer, vSet.GroupKey) {
		return errorz.ErrInvalid{}.New("invalid proposer in state handlers")
	}
	if rNumber == constants.DEADBLOCKROUND {
		return nil
	}
	pidx := objs.GetProposerIdx(len(vSet.Validators), rHeight, rNumber)
	valObj := vSet.Validators[pidx]
	vAddr := valObj.VAddr
	if !bytes.Equal(proposer, vAddr) {
		return errorz.ErrInvalid{}.New("bad proposer")
	}
	return nil
}

func (mb *Handlers) VoteValidation(txn *badger.Txn, groupKey []byte, voter []byte, proposer []byte, coSigners [][]byte, bHeight uint32, rHeight uint32, rNumber uint32) error {
	// OBJECTS		GROUPKEY	VOTER	GROUPSHARE	PROPOSER	HEIGHT	ROUND_1VSET		ROUND>1VSET
	// PREVOTE		TRUE		TRUE	FALSE		TRUE		N+1		N				N+1
	// PRECOMMIT	TRUE		TRUE	FALSE		TRUE		N+1		N				N+1
	if err := mb.ValidateRCERT(txn, groupKey, bHeight, rHeight, rNumber); err != nil {
		return err
	}
	if proposer != nil {
		if err := mb.ProposalValidation(txn, groupKey, proposer, bHeight, rHeight, rNumber); err != nil {
			return err
		}
	}
	vSet, err := mb.database.GetValidatorSet(txn, rHeight)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return err
	}
	if !vSet.IsValidTuple(voter, vSet.GroupKey) {
		correctgk := vSet.GroupKey
		vl := [][]byte{}
		for _, vobj := range vSet.Validators {
			vl = append(vl, vobj.VAddr)
		}
		return errorz.ErrInvalid{}.New(fmt.Sprintf("invalid tuple in state handlers: \nvoter:%x \nGroupKey:%x\ncorrectgk:%x\n%x\n%x\n%x\n%x", voter, groupKey, correctgk, vl[0], vl[1], vl[2], vl[3]))
	}
	for _, cs := range coSigners {
		if !vSet.IsValidTuple(utils.CopySlice(cs), vSet.GroupKey) {
			return errorz.ErrInvalid{}.New("bad co signer")
		}
	}
	return nil
}

func (mb *Handlers) VoteNilValidation(txn *badger.Txn, groupKey []byte, voter []byte, bHeight uint32, rHeight uint32, rNumber uint32) error {
	// OBJECTS		GROUPKEY	VOTER	GROUPSHARE	PROPOSER	HEIGHT	ROUND_1VSET		ROUND>1VSET
	// PREVOTENIL	TRUE		TRUE	FALSE		FALSE		N+1		N				N+1
	// PRECOMMITNIL	TRUE		TRUE	FALSE		FALSE		N+1		N				N+1
	if err := mb.ValidateRCERT(txn, groupKey, bHeight, rHeight, rNumber); err != nil {
		return err
	}
	vSet, err := mb.database.GetValidatorSet(txn, rHeight)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return err
	}
	if !vSet.IsValidTuple(voter, vSet.GroupKey) {
		correctgk := vSet.GroupKey
		vl := [][]byte{}
		for _, vobj := range vSet.Validators {
			vl = append(vl, vobj.VAddr)
		}
		return errorz.ErrInvalid{}.New(fmt.Sprintf("invalid tuple in state handlers: \nvoter:%x \nGroupKey:%x\ncorrectgk:%x\n%x\n%x\n%x\n%x", voter, groupKey, correctgk, vl[0], vl[1], vl[2], vl[3]))
	}
	return nil
}

func (mb *Handlers) VoteNextValidation(txn *badger.Txn, groupKey []byte, groupShare []byte, voter []byte, proposer []byte, coSigners [][]byte, bHeight uint32, rHeight uint32, rNumber uint32) error {
	// OBJECTS	GROUPKEY	VOTER	GROUPSHARE	PROPOSER	HEIGHT	ROUND_1VSET	ROUND>1VSET
	// NROUND	TRUE		TRUE	TRUE		FALSE		N+1		N			N+1
	// NHEIGHT	TRUE		TRUE	TRUE		TRUE		N+1		N			N+1
	if err := mb.ValidateRCERT(txn, groupKey, bHeight, rHeight, rNumber); err != nil {
		return err
	}
	if proposer != nil {
		if err := mb.ProposalValidation(txn, groupKey, proposer, bHeight, rHeight, rNumber); err != nil {
			return err
		}
	}
	vSet, err := mb.database.GetValidatorSet(txn, rHeight)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return err
	}
	if !vSet.IsValidTriplet(voter, groupShare, vSet.GroupKey) {
		correctgk := groupShare
		vl := [][]byte{}
		for _, vobj := range vSet.Validators {
			vl = append(vl, vobj.GroupShare)
		}
		correctShare, ok := vSet.ValidatorGroupShareMap[string(voter)]
		mb.logger.Errorf("Correct Share: %v\n Ok: %v", correctShare, ok)
		return errorz.ErrInvalid{}.New(fmt.Sprintf("invalid triplet in state handlers: \nvoter:%x \nGroupShare:%x\n%x\n%x\n%x\n%x", voter, correctgk, vl[0], vl[1], vl[2], vl[3]))
	}
	for _, cs := range coSigners {
		if !vSet.IsValidTuple(utils.CopySlice(cs), vSet.GroupKey) {
			return errorz.ErrInvalid{}.New("bad co signer")
		}
	}
	return nil
}

func (mb *Handlers) BlockHeaderValidation(txn *badger.Txn, groupKey []byte, bHeight uint32, rHeight uint32, rNumber uint32) error {
	// OBJECTS	GROUPKEY	VOTER	GROUPSHARE	PROPOSER	HEIGHT	ROUND_1VSET	ROUND>1VSET
	// BH		TRUE		FALSE	FALSE		FALSE		N		N			N+1
	if bHeight <= 1 {
		return nil
	}
	vSet, err := mb.database.GetValidatorSet(txn, bHeight)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return err
	}
	vspa, ok, err := mb.database.GetValidatorSetPostApplication(txn, vSet.NotBefore)
	if err != nil {
		utils.DebugTrace(mb.logger, err)
		return err
	}
	if ok {
		vSet = vspa
	}
	if !bytes.Equal(groupKey, vSet.GroupKey) {
		return errorz.ErrInvalid{}.New(fmt.Sprintf("group key mismatch in state handlers: Height: %d GroupKey: %x\nVsetGroupKey:%x", bHeight, groupKey, vSet.GroupKey))
	}
	return nil
}

func normalizeHeightForRCERT(bHeight uint32, rHeight uint32, rNumber uint32) uint32 {
	if bHeight <= 2 || rNumber == 1 {
		return bHeight
	}
	return rHeight
}

func (mb *Handlers) subOneNoZero(height uint32) uint32 {
	if height <= 1 {
		return 1
	}
	return height - 1
}
