package lstate

import (
	"bytes"
	"errors"

	"github.com/MadBase/MadNet/errorz"
	"github.com/dgraph-io/badger/v2"

	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/crypto"
	gUtils "github.com/MadBase/MadNet/utils"
)

type RoundStates struct {
	height             uint32
	round              uint32
	txn                *badger.Txn
	OwnState           *objs.OwnState
	ValidatorSet       *objs.ValidatorSet
	OwnValidatingState *objs.OwnValidatingState
	PeerStateMap       map[string]*objs.RoundState
}

func (r *RoundStates) OwnRoundState() *objs.RoundState {
	return r.PeerStateMap[string(r.OwnState.VAddr)]
}

func (r *RoundStates) IsMe(vAddr []byte) bool {
	if len(vAddr) != 20 {
		return false
	}
	if string(r.OwnState.VAddr) == string(vAddr) {
		return true
	}
	return false
}

func (r *RoundStates) LocalIsProposer() bool {
	ownVAddr := r.OwnState.VAddr
	idx := objs.GetProposerIdx(len(r.ValidatorSet.Validators), r.height, r.round)
	proposerValObj := r.ValidatorSet.Validators[idx]
	vAddr := proposerValObj.VAddr
	return bytes.Equal(vAddr, ownVAddr)
}

func (r *RoundStates) IsCurrentValidator() bool {
	vs := r.ValidatorSet
	vsvvs := vs.ValidatorVAddrSet
	os := r.OwnState
	return vsvvs[string(os.VAddr)]
}

func (r *RoundStates) GetRoundState(vAddr []byte) *objs.RoundState {
	return r.PeerStateMap[string(vAddr)]
}

func (r *RoundStates) GetCurrentProposal() *objs.Proposal {
	idx := objs.GetProposerIdx(len(r.ValidatorSet.ValidatorVAddrMap), r.height, r.round)
	proposerValObj := r.ValidatorSet.Validators[idx]
	vAddr := proposerValObj.VAddr
	proposer := r.PeerStateMap[string(vAddr)]
	if proposer != nil {
		if proposer.Proposal != nil {
			rcert := r.OwnRoundState().RCert
			if proposer.PCurrent(rcert) {
				return proposer.Proposal
			}
		}
	}
	return nil
}

func (r *RoundStates) GetCurrentPreVotes() (objs.PreVoteList, objs.PreVoteNilList, error) {
	pvl := objs.PreVoteList{}
	pvnl := objs.PreVoteNilList{}
	rcert := r.OwnRoundState().RCert
	for _, valObj := range r.ValidatorSet.Validators {
		peerState := r.PeerStateMap[string(valObj.VAddr)]
		if peerState.PVCurrent(rcert) {
			pvl = append(pvl, peerState.PreVote)
		}
		if peerState.PVNCurrent(rcert) {
			pvnl = append(pvnl, true)
		}
	}
	return pvl, pvnl, nil
}

func (r *RoundStates) GetCurrentPreCommits() (objs.PreCommitList, objs.PreCommitNilList, error) {
	pvl := objs.PreCommitList{}
	pvnl := objs.PreCommitNilList{}
	rcert := r.OwnRoundState().RCert
	for _, valObj := range r.ValidatorSet.Validators {
		peerState := r.PeerStateMap[string(valObj.VAddr)]
		if peerState.PCCurrent(rcert) {
			pvl = append(pvl, peerState.PreCommit)
		}
		if peerState.PCNCurrent(rcert) {
			pvnl = append(pvnl, true)
		}
	}
	return pvl, pvnl, nil
}

func (r *RoundStates) GetCurrentNext() (objs.NextHeightList, objs.NextRoundList, error) {
	pvl := objs.NextHeightList{}
	pvnl := objs.NextRoundList{}
	rcert := r.OwnRoundState().RCert
	for _, valObj := range r.ValidatorSet.Validators {
		peerState := r.PeerStateMap[string(valObj.VAddr)]
		if peerState.NHCurrent(rcert) {
			pvl = append(pvl, peerState.NextHeight)
		}
		if peerState.NRCurrent(rcert) {
			pvnl = append(pvnl, peerState.NextRound)
		}
	}
	return pvl, pvnl, nil
}

func (r *RoundStates) SetProposal(p *objs.Proposal) error {
	rs := r.GetRoundState(p.Proposer)
	if rs == nil {
		return errorz.ErrInvalid{}.New("round state is nil in set prop")
	}
	_, err := rs.SetProposal(p)
	if err != nil {
		return err
	}
	return nil
}

func (r *RoundStates) SetPreVote(pv *objs.PreVote) error {
	err := r.SetProposal(pv.Proposal)
	if err != nil {
		etest := &errorz.ErrStale{}
		if !errors.As(err, &etest) {
			return err
		}
	}
	rs := r.GetRoundState(pv.Voter)
	if rs == nil {
		return errorz.ErrInvalid{}.New("rs nil in set prevote")
	}
	_, err = rs.SetPreVote(pv)
	if err != nil {
		return err
	}
	return nil
}

func (r *RoundStates) SetPreVoteNil(pvn *objs.PreVoteNil) error {
	rs := r.GetRoundState(pvn.Voter)
	if rs == nil {
		return errorz.ErrInvalid{}.New("rs nil in pvn")
	}
	_, err := rs.SetPreVoteNil(pvn)
	if err != nil {
		return err
	}
	return nil
}

func (r *RoundStates) SetPreCommit(pc *objs.PreCommit) error {
	err := r.SetProposal(pc.Proposal)
	if err != nil {
		etest := &errorz.ErrStale{}
		if !errors.As(err, &etest) {
			return err
		}
	}
	pvl, err := pc.MakeImplPreVotes()
	if err != nil {
		return err
	}
	for _, pv := range pvl {
		rs := r.GetRoundState(pv.Voter)
		if rs == nil {
			return errorz.ErrInvalid{}.New("rs nil in set prevote")
		}
		_, err = rs.SetPreVote(pv)
		if err != nil {
			etest := &errorz.ErrStale{}
			if !errors.As(err, &etest) {
				rs := r.GetRoundState(pc.Voter)
				if rs == nil {
					return errorz.ErrInvalid{}.New("rs nil in pc")
				}
				rs.ImplicitPCN = true
				return nil
			}
		}
	}
	rs := r.GetRoundState(pc.Voter)
	if rs == nil {
		return errorz.ErrInvalid{}.New("rs nil in pc")
	}
	_, err = rs.SetPreCommit(pc)
	if err != nil {
		return err
	}
	return nil
}

func (r *RoundStates) SetPreCommitNil(pcn *objs.PreCommitNil) error {
	rs := r.GetRoundState(pcn.Voter)
	if rs == nil {
		return errorz.ErrInvalid{}.New("rs nil in pcn")
	}
	_, err := rs.SetPreCommitNil(pcn)
	if err != nil {
		return err
	}
	return nil
}

func (r *RoundStates) SetNextHeight(pc *objs.NextHeight) error {
	r.SetProposal(pc.NHClaims.Proposal)
	rs := r.GetRoundState(pc.Voter)
	if rs == nil {
		return errorz.ErrInvalid{}.New("rs nil in nh")
	}
	_, err := rs.SetNextHeight(pc)
	if err != nil {
		return err
	}
	return nil
}

func (r *RoundStates) SetNextRound(pc *objs.NextRound) error {
	rs := r.GetRoundState(pc.Voter)
	if rs == nil {
		return errorz.ErrInvalid{}.New("rs nil in nr")
	}
	_, err := rs.SetNextRound(pc)
	if err != nil {
		return err
	}
	return nil
}

func (r *RoundStates) LockedValueCurrent() bool {
	vv := r.LockedValue()
	if vv == nil {
		return false
	}
	relation := objs.RelateH(r.OwnRoundState(), vv)
	return relation == 0
}

func (r *RoundStates) ValidValueCurrent() bool {
	vv := r.ValidValue()
	if vv == nil {
		return false
	}
	relation := objs.RelateH(r.OwnRoundState(), vv)
	return relation == 0
}

func (r *RoundStates) LockedValue() *objs.Proposal {
	if r.OwnValidatingState == nil {
		return nil
	}
	return r.OwnValidatingState.LockedValue
}

func (r *RoundStates) ValidValue() *objs.Proposal {
	if r.OwnValidatingState == nil {
		return nil
	}
	return r.OwnValidatingState.ValidValue
}

func (r *RoundStates) ChainID() uint32 {
	return r.OwnRoundState().RCert.RClaims.ChainID
}

func (r *RoundStates) Height() uint32 {
	return r.OwnRoundState().RCert.RClaims.Height
}

func (r *RoundStates) Round() uint32 {
	return r.OwnRoundState().RCert.RClaims.Round
}

func (r *RoundStates) RCert() *objs.RCert {
	return r.OwnRoundState().RCert
}

func (r *RoundStates) PrevBlock() []byte {
	prevBlock := gUtils.CopySlice(r.OwnRoundState().RCert.RClaims.PrevBlock)
	return prevBlock
}

func (r *RoundStates) GetCurrentThreshold() int {
	return crypto.CalcThreshold(len(r.PeerStateMap)) + 1
}

func (r *RoundStates) LocalPreVoteCurrent() bool {
	return r.OwnRoundState().PVCurrent(r.OwnRoundState().RCert)
}

func (r *RoundStates) LocalPreCommitCurrent() bool {
	return r.OwnRoundState().PCCurrent(r.OwnRoundState().RCert)
}
