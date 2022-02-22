package objs

import (
	"errors"

	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/rstate"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

var errConflict = errors.New("conflict vote")

// RoundState ...
type RoundState struct {
	VAddr                 []byte
	GroupKey              []byte
	GroupShare            []byte
	GroupIdx              uint8
	RCert                 *RCert
	ConflictingRCert      *RCert
	Proposal              *Proposal
	ConflictingProposal   *Proposal
	PreVote               *PreVote
	ConflictingPreVote    *PreVote
	PreVoteNil            *PreVoteNil
	ImplicitPVN           bool
	PreCommit             *PreCommit
	ConflictingPreCommit  *PreCommit
	PreCommitNil          *PreCommitNil
	ImplicitPCN           bool
	NextRound             *NextRound
	NextHeight            *NextHeight
	ConflictingNextHeight *NextHeight
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// RoundState object
func (b *RoundState) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("RoundState.UnmarshalBinary; rs not initialized")
	}
	bh, err := rstate.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *RoundState) UnmarshalCapn(bh mdefs.RoundState) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("RoundState.UnmarshalCapn; rs not initialized")
	}
	err := rstate.Validate(bh)
	if err != nil {
		return err
	}
	b.VAddr = utils.CopySlice(bh.VAddr())
	b.GroupKey = utils.CopySlice(bh.GroupKey())
	b.GroupShare = utils.CopySlice(bh.GroupShare())
	b.GroupIdx = bh.GroupIdx()
	b.ImplicitPVN = bh.ImplicitPVN()
	b.ImplicitPCN = bh.ImplicitPCN()
	b.RCert = &RCert{}
	err = b.RCert.UnmarshalCapn(bh.RCert())
	if err != nil {
		return err
	}
	if bh.HasConflictingRCert() {
		b.ConflictingRCert = &RCert{}
		err := b.ConflictingRCert.UnmarshalCapn(bh.ConflictingRCert())
		if err != nil {
			return err
		}
	}
	if bh.HasProposal() {
		b.Proposal = &Proposal{}
		err := b.Proposal.UnmarshalCapn(bh.Proposal())
		if err != nil {
			return err
		}
	}
	if bh.HasConflictingProposal() {
		b.ConflictingProposal = &Proposal{}
		err := b.ConflictingProposal.UnmarshalCapn(bh.ConflictingProposal())
		if err != nil {
			return err
		}
	}
	if bh.HasPreVote() {
		b.PreVote = &PreVote{}
		err := b.PreVote.UnmarshalCapn(bh.PreVote())
		if err != nil {
			return err
		}
	}
	if bh.HasConflictingPreVote() {
		b.ConflictingPreVote = &PreVote{}
		err := b.ConflictingPreVote.UnmarshalCapn(bh.ConflictingPreVote())
		if err != nil {
			return err
		}
	}
	if bh.HasPreVoteNil() {
		b.PreVoteNil = &PreVoteNil{}
		err := b.PreVoteNil.UnmarshalCapn(bh.PreVoteNil())
		if err != nil {
			return err
		}
	}
	if bh.HasPreCommit() {
		b.PreCommit = &PreCommit{}
		err := b.PreCommit.UnmarshalCapn(bh.PreCommit())
		if err != nil {
			return err
		}
	}
	if bh.HasConflictingPreCommit() {
		b.ConflictingPreCommit = &PreCommit{}
		err := b.ConflictingPreCommit.UnmarshalCapn(bh.ConflictingPreCommit())
		if err != nil {
			return err
		}
	}
	if bh.HasPreCommitNil() {
		b.PreCommitNil = &PreCommitNil{}
		err := b.PreCommitNil.UnmarshalCapn(bh.PreCommitNil())
		if err != nil {
			return err
		}
	}
	if bh.HasNextRound() {
		b.NextRound = &NextRound{}
		err := b.NextRound.UnmarshalCapn(bh.NextRound())
		if err != nil {
			return err
		}
	}
	if bh.HasNextHeight() {
		b.NextHeight = &NextHeight{}
		err := b.NextHeight.UnmarshalCapn(bh.NextHeight())
		if err != nil {
			return err
		}
	}
	if bh.HasConflictingNextHeight() {
		b.ConflictingNextHeight = &NextHeight{}
		err := b.ConflictingNextHeight.UnmarshalCapn(bh.ConflictingNextHeight())
		if err != nil {
			return err
		}
	}
	return nil
}

// MarshalBinary takes the RoundState object and returns the canonical
// byte slice
func (b *RoundState) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("RoundState.MarshalBinary; rs not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return rstate.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *RoundState) MarshalCapn(seg *capnp.Segment) (mdefs.RoundState, error) {
	if b == nil {
		return mdefs.RoundState{}, errorz.ErrInvalid{}.New("RoundState.MarshalCapn; rs not initialized")
	}
	var bh mdefs.RoundState
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootRoundState(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewRoundState(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	RCert, err := b.RCert.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	err = bh.SetRCert(RCert)
	if err != nil {
		return mdefs.RoundState{}, err
	}
	if b.ConflictingRCert != nil {
		ConflictingRCert, err := b.ConflictingRCert.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetConflictingRCert(ConflictingRCert)
		if err != nil {
			return mdefs.RoundState{}, err
		}
	}
	if b.Proposal != nil {
		Proposal, err := b.Proposal.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetProposal(Proposal)
		if err != nil {
			return mdefs.RoundState{}, err
		}
	}
	if b.ConflictingProposal != nil {
		ConflictingProposal, err := b.ConflictingProposal.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetConflictingProposal(ConflictingProposal)
		if err != nil {
			return mdefs.RoundState{}, err
		}
	}
	if b.PreVote != nil {
		PreVote, err := b.PreVote.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetPreVote(PreVote)
		if err != nil {
			return mdefs.RoundState{}, err
		}
	}
	if b.ConflictingPreVote != nil {
		ConflictingPreVote, err := b.ConflictingPreVote.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetConflictingPreVote(ConflictingPreVote)
		if err != nil {
			return mdefs.RoundState{}, err
		}
	}
	if b.PreVoteNil != nil {
		PreVoteNil, err := b.PreVoteNil.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetPreVoteNil(PreVoteNil)
		if err != nil {
			return mdefs.RoundState{}, err
		}
	}
	if b.PreCommit != nil {
		PreCommit, err := b.PreCommit.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetPreCommit(PreCommit)
		if err != nil {
			return mdefs.RoundState{}, err
		}
	}
	if b.ConflictingPreCommit != nil {
		ConflictingPreCommit, err := b.ConflictingPreCommit.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetConflictingPreCommit(ConflictingPreCommit)
		if err != nil {
			return mdefs.RoundState{}, err
		}
	}
	if b.PreCommitNil != nil {
		PreCommitNil, err := b.PreCommitNil.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetPreCommitNil(PreCommitNil)
		if err != nil {
			return mdefs.RoundState{}, err
		}
	}
	if b.NextRound != nil {
		NextRound, err := b.NextRound.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetNextRound(NextRound)
		if err != nil {
			return mdefs.RoundState{}, err
		}
	}
	if b.NextHeight != nil {
		NextHeight, err := b.NextHeight.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetNextHeight(NextHeight)
		if err != nil {
			return mdefs.RoundState{}, err
		}
	}
	if b.ConflictingNextHeight != nil {
		ConflictingNextHeight, err := b.ConflictingNextHeight.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetConflictingNextHeight(ConflictingNextHeight)
		if err != nil {
			return mdefs.RoundState{}, err
		}
	}
	err = bh.SetVAddr(b.VAddr)
	if err != nil {
		return mdefs.RoundState{}, err
	}
	err = bh.SetGroupKey(b.GroupKey)
	if err != nil {
		return mdefs.RoundState{}, err
	}
	err = bh.SetGroupShare(b.GroupShare)
	if err != nil {
		return mdefs.RoundState{}, err
	}
	bh.SetGroupIdx(b.GroupIdx)
	bh.SetImplicitPVN(b.ImplicitPVN)
	bh.SetImplicitPCN(b.ImplicitPCN)
	return bh, nil
}

func (b *RoundState) Reset() {
	b.RCert = nil
	b.ConflictingRCert = nil
	b.Proposal = nil
	b.ConflictingProposal = nil
	b.PreVote = nil
	b.ConflictingPreVote = nil
	b.PreVoteNil = nil
	b.ImplicitPVN = false
	b.PreCommit = nil
	b.ConflictingPreCommit = nil
	b.PreCommitNil = nil
	b.ImplicitPCN = false
	b.NextRound = nil
}

func (b *RoundState) CurrentHR(a *RCert) bool {
	relation := RelateHR(a, b)
	return relation == 0
}

func (b *RoundState) FutureHR(a *RCert) bool {
	relation := RelateHR(a, b.RCert)
	return relation == -1
}

func (b *RoundState) CurrentH(a *RCert) bool {
	relation := RelateH(a, b.RCert)
	return relation == 0
}

func (b *RoundState) FutureH(a *RCert) bool {
	relation := RelateH(a, b.RCert)
	return relation == -1
}

func (b *RoundState) PCurrent(a *RCert) bool {
	if b.Proposal != nil {
		relation := RelateHR(a, b.Proposal)
		return relation == 0
	}
	return false
}

func (b *RoundState) PVCurrent(a *RCert) bool {
	if b.PreVote != nil {
		relation := RelateHR(a, b.PreVote)
		if relation == 0 {
			return true
		}
	}
	return false
}

func (b *RoundState) PVNCurrent(a *RCert) bool {
	if b.PreVoteNil != nil {
		relation := RelateHR(a, b.PreVoteNil)
		if relation == 0 {
			return true
		}
	}
	return false
}

func (b *RoundState) PCCurrent(a *RCert) bool {
	if b.PreCommit != nil {
		relation := RelateHR(a, b.PreCommit)
		if relation == 0 {
			return true
		}
	}
	return false
}

func (b *RoundState) PCNCurrent(a *RCert) bool {
	if b.PreCommitNil != nil {
		relation := RelateHR(a, b.PreCommitNil)
		if relation == 0 {
			return true
		}
	}
	return false
}

func (b *RoundState) NRCurrent(a *RCert) bool {
	if b.NextRound != nil {
		relation := RelateHR(a, b.NextRound)
		if relation == 0 {
			return true
		}
	}
	return false
}

func (b *RoundState) NHCurrent(a *RCert) bool {
	if b.NextHeight != nil {
		// if we are in DBR
		if IsDeadBlockRound(a) || IsDeadBlockRound(b) {
			//ignore a NH from before DBR
			if RelateHR(a, b.NextHeight) == 0 {
				// count all NH from DBR
				return true
			}
		} else {
			// count all NH messages from any round in this height
			if RelateH(a, b.NextHeight) == 0 {
				// count all NH from DBR
				return true
			}
		}
	}
	return false
}

func (b *RoundState) TrackExternalConflicts(v *Proposal) {
	// from current height
	relationHR := RelateHR(b, v)
	if relationHR == 1 { // from prev round
		return
	}
	if relationHR == -1 { // from future round
		return
	}
	// is current
	b.checkStaleAndConflict(v, false)
}

func (b *RoundState) SetRCert(rc *RCert) error {
	b.setReset(rc)
	return nil
}

func (b *RoundState) SetProposal(v *Proposal) (bool, error) {
	if IsDeadBlockRound(v) {
		if len(v.TxHshLst) != 0 {
			return false, errorz.ErrInvalid{}.New("RoundState.SetProposal; tx hash in DBR set proposal")
		}
	}
	ok, err := b.genericSet(v)
	if err != nil {
		return false, err
	}
	if !ok {
		if IsDeadBlockRound(v) {
			return false, errorz.ErrInvalid{}.New("RoundState.SetProposal; corrupt p in dbr")
		}
	}
	return ok, nil
}

func (b *RoundState) SetPreVote(v *PreVote) (bool, error) {
	if IsDeadBlockRound(v) {
		if len(v.Proposal.TxHshLst) != 0 {
			return false, errorz.ErrInvalid{}.New("RoundState.SetPreVote; tx hash in dbr set pv")
		}
	}
	ok, err := b.genericSet(v)
	if err != nil {
		return false, err
	}
	if !ok {
		if IsDeadBlockRound(v) {
			return false, errorz.ErrInvalid{}.New("RoundState.SetPreVote; corrupt pv in dbr")
		}
	}
	return ok, nil
}

func (b *RoundState) SetPreVoteNil(v *PreVoteNil) (bool, error) {
	if IsDeadBlockRound(v) {
		return false, errorz.ErrInvalid{}.New("RoundState.SetPreVoteNil; dbr tx hash in set pvn")
	}
	ok, err := b.genericSet(v)
	if err != nil {
		return false, err
	}
	if !ok {
		if IsDeadBlockRound(v) {
			return false, errorz.ErrInvalid{}.New("RoundState.SetPreVoteNil; corrupt pvn in dbr")
		}
	}
	return ok, nil
}

func (b *RoundState) SetPreCommit(v *PreCommit) (bool, error) {
	if IsDeadBlockRound(v) {
		if len(v.Proposal.TxHshLst) != 0 {
			return false, errorz.ErrInvalid{}.New("RoundState.SetPreCommit; dbr tx hash in set pc")
		}
	}
	ok, err := b.genericSet(v)
	if err != nil {
		return false, err
	}
	if !ok {
		if IsDeadBlockRound(v) {
			return false, errorz.ErrInvalid{}.New("RoundState.SetPreCommit; corrupt pc in dbr")
		}
	}
	return ok, nil
}

func (b *RoundState) SetPreCommitNil(v *PreCommitNil) (bool, error) {
	if IsDeadBlockRound(v) {
		return false, errorz.ErrInvalid{}.New("RoundState.SetPreCommitNil; dbr tx hash in pcn")
	}
	ok, err := b.genericSet(v)
	if err != nil {
		return false, err
	}
	if !ok {
		if IsDeadBlockRound(v) {
			return false, errorz.ErrInvalid{}.New("RoundState.SetPreCommitNil; corrupt pcn in dbr")
		}
	}
	return ok, nil
}

func (b *RoundState) SetNextRound(v *NextRound) (bool, error) {
	if IsDeadBlockRound(v) {
		return false, errorz.ErrInvalid{}.New("RoundState.SetNextRound; nr in dbr")
	}
	ok, err := b.genericSet(v)
	if err != nil {
		return false, err
	}
	if !ok {
		if IsDeadBlockRound(v) {
			return false, errorz.ErrInvalid{}.New("RoundState.SetNextRound; corrupt nr in dbr")
		}
	}
	return ok, nil
}

func (b *RoundState) SetNextHeight(v *NextHeight) (bool, error) {
	if IsDeadBlockRound(v) {
		if len(v.NHClaims.Proposal.TxHshLst) != 0 {
			return false, errorz.ErrInvalid{}.New("RoundState.SetNextHeight; set nh dbr tx hash")
		}
	}
	relationH := RelateH(b, v)
	if relationH == 1 { // from prev height
		return false, errorz.ErrStale{}.New("RoundState.SetNextHeight; set nh relation == 1")
	}
	if relationH == -1 { // from future height
		b.setReset(v)
		return true, nil
	}
	// from current height
	ok, err := b.checkStaleAndConflict(v, true)
	if err != nil || !ok {
		return ok, err
	}
	b.setType(v)
	return true, nil
}

func (b *RoundState) genericSet(v interface{}) (bool, error) {
	relationH := RelateH(b, v)
	if relationH == 1 { // from prev height
		return false, errorz.ErrStale{}.New("RoundState.genericSet; relationH == 1")
	}
	if relationH == -1 { // from future height
		b.setReset(v)
		return true, nil
	}
	// from current height

	relationHR := RelateHR(b, v)
	if relationHR == 1 { // from prev round
		switch v.(type) {
		case *NextHeight:
			if b.NextHeight != nil {
				if RelateHR(b.NextHeight, v) != -1 {
					return false, errorz.ErrStale{}.New("RoundState.genericSet; stale next height")
				}
			}
		default:
			return false, errorz.ErrStale{}.New("RoundState.genericSet; relationHR == 1")
		}
	}
	if relationHR == -1 { // from future round
		b.setReset(v)
		return true, nil
	}
	// is current
	ok, err := b.checkStaleAndConflict(v, true)
	if err != nil || !ok {
		return ok, err
	}
	b.setType(v)
	return true, nil
}

func (b *RoundState) checkStaleAndConflict(a interface{}, internal bool) (bool, error) {
	err := b.checkConflict(a, internal)
	if err != nil {
		if err == errConflict {
			b.setTypeConflict(a, internal)
			return false, nil
		}
		return false, err
	}
	err = b.checkTypeStale(a, internal)
	if err != nil {
		return false, err
	}
	return true, err
}

func (b *RoundState) checkConflict(a interface{}, internal bool) error {
	if internal {
		b.resetNHForDBR(a)
		if err := b.checkSameTypeConflict(a); err != nil {
			return err
		}
	}
	if !PrevBlockEqual(b, a) {
		return errConflict
	}
	var hasBClaims bool
	switch a.(type) {
	case *RCert:
		hasBClaims = false
	case *Proposal:
		hasBClaims = true
	case *PreVote:
		hasBClaims = true
	case *PreVoteNil:
		hasBClaims = false
	case *PreCommit:
		hasBClaims = true
	case *PreCommitNil:
		hasBClaims = false
	case *NextRound:
		hasBClaims = false
	case *NextHeight:
		hasBClaims = true
	default:
		panic("RoundState.checkConflict; bad type in hash bclaims check")
	}
	if b.Proposal != nil && hasBClaims {
		ok, err := BClaimsEqual(a, b.Proposal)
		if err != nil {
			return err
		}
		if !ok {
			return errConflict
		}
	}
	if b.PreVote != nil && hasBClaims {
		ok, err := BClaimsEqual(a, b.PreVote)
		if err != nil {
			return err
		}
		if !ok {
			return errConflict
		}
	}
	if b.PreVoteNil != nil {
		ok := PrevBlockEqual(a, b.PreVoteNil)
		if !ok {
			return errConflict
		}
	}
	if b.PreCommit != nil && hasBClaims {
		ok, err := BClaimsEqual(a, b.PreCommit)
		if err != nil {
			return err
		}
		if !ok {
			return errConflict
		}
	}
	if b.PreCommitNil != nil {
		ok := PrevBlockEqual(a, b.PreCommitNil)
		if !ok {
			return errConflict
		}
	}
	if b.NextRound != nil {
		ok := PrevBlockEqual(a, b.NextRound)
		if !ok {
			return errConflict
		}
	}
	if b.NextHeight != nil && hasBClaims {
		ok, err := BClaimsEqual(a, b.NextHeight)
		if err != nil {
			return err
		}
		if !ok {
			return errConflict
		}
	}
	return nil
}

func (b *RoundState) resetNHForDBR(v interface{}) {
	if IsDeadBlockRound(v) {
		if b.NextHeight != nil {
			if !IsDeadBlockRound(b.NextHeight) {
				b.NextHeight = nil
			}
		}
		if b.ConflictingNextHeight != nil {
			if !IsDeadBlockRound(b.ConflictingNextHeight) {
				b.ConflictingNextHeight = nil
			}
		}
	}
}

func (b *RoundState) checkSameTypeConflict(any interface{}) error {
	if b.ImplicitPVN || b.ImplicitPCN {
		return errorz.ErrInvalid{}.New("RoundState.checkSameTypeConflict; pvn or pcn implicit nil set")
	}
	if b.ConflictingRCert != nil {
		return errorz.ErrInvalid{}.New("RoundState.checkSameTypeConflict; conflicting rc")
	}
	switch any.(type) {
	case *Proposal:
		if b.ConflictingProposal != nil {
			return errorz.ErrInvalid{}.New("RoundState.checkSameTypeConflict; conflicting p set")
		}
	case *PreVote:
		if b.PreVoteNil != nil {
			return errorz.ErrInvalid{}.New("RoundState.checkSameTypeConflict; conflicting pvn set")
		}
		if b.ConflictingPreVote != nil {
			return errorz.ErrInvalid{}.New("RoundState.checkSameTypeConflict; conflicting pv set")
		}
	case *PreVoteNil:
		if b.PreVote != nil {
			return errorz.ErrInvalid{}.New("RoundState.checkSameTypeConflict; pv and pvn")
		}
	case *PreCommit:
		if b.PreVoteNil != nil {
			return errorz.ErrInvalid{}.New("RoundState.checkSameTypeConflict; pc and pvn")
		}
		if b.PreCommitNil != nil {
			return errorz.ErrInvalid{}.New("RoundState.checkSameTypeConflict; pc and pcn")
		}
		if b.ConflictingPreCommit != nil {
			return errorz.ErrInvalid{}.New("RoundState.checkSameTypeConflict; pc and conflicting pc set")
		}
	case *PreCommitNil:
		if b.PreCommit != nil {
			return errorz.ErrInvalid{}.New("RoundState.checkSameTypeConflict; pc and pcn on nil side")
		}
	case *NextRound:
	// check rcert done above
	case *NextHeight:
		if b.ConflictingNextHeight != nil {
			return errorz.ErrInvalid{}.New("RoundState.checkSameTypeConflict; nh conflicting set")
		}
	default:
		panic("RoundState.checkSameTypeConflict; bad type in check same conflict")
	}
	return nil
}

func (b *RoundState) setReset(any interface{}) {
	// run twice
	var isFutureHeight bool
	if RelateH(b, any) == -1 {
		isFutureHeight = true
	}
	_, round := ExtractHR(any)
	switch any.(type) {
	case *RCert:
		if round == constants.DEADBLOCKROUND || isFutureHeight {
			b.NextHeight = nil
			b.ConflictingNextHeight = nil
		}
	case *Proposal:
		if round == constants.DEADBLOCKROUND || isFutureHeight {
			b.NextHeight = nil
			b.ConflictingNextHeight = nil
		}
	case *PreVote:
		if round == constants.DEADBLOCKROUND || isFutureHeight {
			b.NextHeight = nil
			b.ConflictingNextHeight = nil
		}
	case *PreVoteNil:
		if round == constants.DEADBLOCKROUND || isFutureHeight {
			b.NextHeight = nil
			b.ConflictingNextHeight = nil
		}
	case *PreCommit:
		if round == constants.DEADBLOCKROUND || isFutureHeight {
			b.NextHeight = nil
			b.ConflictingNextHeight = nil
		}
	case *PreCommitNil:
		if round == constants.DEADBLOCKROUND || isFutureHeight {
			b.NextHeight = nil
			b.ConflictingNextHeight = nil
		}
	case *NextRound:
		if round == constants.DEADBLOCKROUNDNR || isFutureHeight {
			b.NextHeight = nil
			b.ConflictingNextHeight = nil
		}
	case *NextHeight:
		if round == constants.DEADBLOCKROUND || isFutureHeight {
			b.NextHeight = nil
			b.ConflictingNextHeight = nil
		}
	default:
		panic("RoundState.setReset; bad type in set type")
	}
	b.Reset()
	b.setType(ExtractRCert(any))
	b.setType(any)
}

func (b *RoundState) setTypeConflict(any interface{}, internal bool) {
	if internal {
		switch v := any.(type) {
		case *Proposal:
			b.ConflictingProposal = v
		case *PreVote:
			b.ConflictingPreVote = v
		case *PreVoteNil:
			b.ConflictingRCert = ExtractRCert(v)
		case *PreCommit:
			b.ConflictingPreCommit = v
		case *PreCommitNil:
			b.ConflictingRCert = ExtractRCert(v)
		case *NextRound:
			b.ConflictingRCert = ExtractRCert(v)
		case *NextHeight:
			b.ConflictingNextHeight = v
		default:
			panic("RoundState.setTypeConflict; bad type in set conflict")
		}
	}
	b.ImplicitPVN = true
	b.ImplicitPCN = true
}

func (b *RoundState) setType(any interface{}) {
	switch v := any.(type) {
	case *RCert:
		b.RCert = v
	case *Proposal:
		b.Proposal = v
	case *PreVote:
		b.PreVote = v
	case *PreVoteNil:
		b.PreVoteNil = v
	case *PreCommit:
		b.PreCommit = v
	case *PreCommitNil:
		b.PreCommitNil = v
	case *NextRound:
		b.NextRound = v
	case *NextHeight:
		b.NextHeight = v
	default:
		panic("RoundState.setType; bad type in set type")
	}
}

func (b *RoundState) checkTypeStale(any interface{}, internal bool) error {
	if !internal {
		return nil
	}
	switch any.(type) {
	case *Proposal:
		if b.Proposal != nil {
			return errorz.ErrStale{}.New("RoundState.checkTypeStale; prop already set")
		}
	case *PreVote:
		if b.PreVote != nil {
			return errorz.ErrStale{}.New("RoundState.checkTypeStale; pv already set")
		}
	case *PreVoteNil:
		if b.PreVoteNil != nil {
			return errorz.ErrStale{}.New("RoundState.checkTypeStale; pvn already set")
		}
	case *PreCommit:
		if b.PreCommit != nil {
			return errorz.ErrStale{}.New("RoundState.checkTypeStale; pc already set")
		}
	case *PreCommitNil:
		if b.PreCommitNil != nil {
			return errorz.ErrStale{}.New("RoundState.checkTypeStale; pcn already set")
		}
	case *NextHeight:
		if b.NextHeight != nil {
			return errorz.ErrStale{}.New("RoundState.checkTypeStale; nh already set")
		}
	case *NextRound:
		if b.NextRound != nil {
			return errorz.ErrStale{}.New("RoundState.checkTypeStale; nr already set")
		}
	default:
		panic("RoundState.checkTypeStale; bad type in check stale")
	}
	return nil
}
