package objs

import (
	"time"

	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/ovstate"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// OwnValidatingState ...
type OwnValidatingState struct {
	RoundStarted         int64
	PreVoteStepStarted   int64
	PreCommitStepStarted int64
	ValidValue           *Proposal
	LockedValue          *Proposal
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// OwnValidatingState object
func (b *OwnValidatingState) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("OwnValidatingState.UnmarshalBinary; ovs not initialized")
	}
	bh, err := ovstate.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *OwnValidatingState) UnmarshalCapn(bh mdefs.OwnValidatingState) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("OwnValidatingState.UnmarshalCapn; ovs not initialized")
	}
	err := ovstate.Validate(bh)
	if err != nil {
		return err
	}
	b.RoundStarted = bh.RoundStarted()
	b.PreVoteStepStarted = bh.PreVoteStepStarted()
	b.PreCommitStepStarted = bh.PreCommitStepStarted()
	if bh.HasLockedValue() {
		b.LockedValue = &Proposal{}
		err := b.LockedValue.UnmarshalCapn(bh.LockedValue())
		if err != nil {
			return err
		}
	}
	if bh.HasValidValue() {
		b.ValidValue = &Proposal{}
		err := b.ValidValue.UnmarshalCapn(bh.ValidValue())
		if err != nil {
			return err
		}
	}
	return nil
}

// MarshalBinary takes the OwnValidatingState object and returns the canonical
// byte slice
func (b *OwnValidatingState) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("OwnValidatingState.MarshalBinary; ovs not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return ovstate.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *OwnValidatingState) MarshalCapn(seg *capnp.Segment) (mdefs.OwnValidatingState, error) {
	if b == nil {
		return mdefs.OwnValidatingState{}, errorz.ErrInvalid{}.New("OwnValidatingState.MarshalCapn; ovs not initialized")
	}
	var bh mdefs.OwnValidatingState
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootOwnValidatingState(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewOwnValidatingState(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	bh.SetRoundStarted(b.RoundStarted)
	bh.SetPreVoteStepStarted(b.PreVoteStepStarted)
	bh.SetPreCommitStepStarted(b.PreCommitStepStarted)
	if b.LockedValue != nil {
		LockedValue, err := b.LockedValue.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetLockedValue(LockedValue)
		if err != nil {
			return mdefs.OwnValidatingState{}, err
		}
	}
	if b.ValidValue != nil {
		ValidValue, err := b.ValidValue.MarshalCapn(seg)
		if err != nil {
			return bh, err
		}
		err = bh.SetValidValue(ValidValue)
		if err != nil {
			return mdefs.OwnValidatingState{}, err
		}
	}
	return bh, nil
}

func (b *OwnValidatingState) PTOExpired(proposalStepTO time.Duration) bool {
	return time.Unix(b.RoundStarted, 0).Add(proposalStepTO).Before(time.Now())
}

func (b *OwnValidatingState) PVTOExpired(preVoteStepTO time.Duration) bool {
	return time.Unix(b.PreVoteStepStarted, 0).Add(preVoteStepTO).Before(time.Now())
}

func (b *OwnValidatingState) PCTOExpired(preCommitStepTO time.Duration) bool {
	return time.Unix(b.PreCommitStepStarted, 0).Add(preCommitStepTO).Before(time.Now())
}

func (b *OwnValidatingState) DBRNRExpired(dbrnrTO time.Duration) bool {
	return time.Unix(b.PreCommitStepStarted, 0).Add(dbrnrTO).Before(time.Now())
}

func (b *OwnValidatingState) SetRoundStarted() {
	now := time.Now()
	b.RoundStarted = now.Unix()
	b.PreVoteStepStarted = 0
	b.PreCommitStepStarted = 0
}

func (b *OwnValidatingState) SetPreVoteStepStarted() {
	now := time.Now()
	b.PreVoteStepStarted = now.Unix()
	b.PreCommitStepStarted = 0
}

func (b *OwnValidatingState) SetPreCommitStepStarted() {
	now := time.Now()
	b.PreCommitStepStarted = now.Unix()
}
