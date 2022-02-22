package objs

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/ostate"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// OwnState ...
type OwnState struct {
	VAddr             []byte
	GroupKey          []byte
	SyncToBH          *BlockHeader
	MaxBHSeen         *BlockHeader
	CanonicalSnapShot *BlockHeader
	PendingSnapShot   *BlockHeader
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// OwnState object
func (b *OwnState) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("OwnState.UnmarshalBinary; os not initialized")
	}
	bh, err := ostate.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *OwnState) UnmarshalCapn(bh mdefs.OwnState) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("OwnState.UnmarshalCapn; os not initialized")
	}
	b.SyncToBH = &BlockHeader{}
	b.MaxBHSeen = &BlockHeader{}
	b.PendingSnapShot = &BlockHeader{}
	b.CanonicalSnapShot = &BlockHeader{}
	err := ostate.Validate(bh)
	if err != nil {
		return err
	}
	b.VAddr = bh.VAddr()
	b.GroupKey = bh.GroupKey()
	err = b.SyncToBH.UnmarshalCapn(bh.SyncToBH())
	if err != nil {
		return err
	}
	err = b.MaxBHSeen.UnmarshalCapn(bh.MaxBHSeen())
	if err != nil {
		return err
	}
	err = b.CanonicalSnapShot.UnmarshalCapn(bh.CanonicalSnapShot())
	if err != nil {
		return err
	}
	err = b.PendingSnapShot.UnmarshalCapn(bh.PendingSnapShot())
	if err != nil {
		return err
	}
	return nil
}

// MarshalBinary takes the OwnState object and returns the canonical
// byte slice
func (b *OwnState) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("OwnState.MarshalBinary; os not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return ostate.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *OwnState) MarshalCapn(seg *capnp.Segment) (mdefs.OwnState, error) {
	if b == nil {
		return mdefs.OwnState{}, errorz.ErrInvalid{}.New("OwnState.MarshalCapn; os not initialized")
	}
	var bh mdefs.OwnState
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootOwnState(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewOwnState(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	SyncToBH, err := b.SyncToBH.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	MaxBHSeen, err := b.MaxBHSeen.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	CanonicalSnapShot, err := b.CanonicalSnapShot.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	PendingSnapShot, err := b.PendingSnapShot.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	err = bh.SetVAddr(b.VAddr)
	if err != nil {
		return mdefs.OwnState{}, err
	}
	err = bh.SetGroupKey(b.GroupKey)
	if err != nil {
		return mdefs.OwnState{}, err
	}
	err = bh.SetSyncToBH(SyncToBH)
	if err != nil {
		return mdefs.OwnState{}, err
	}
	err = bh.SetMaxBHSeen(MaxBHSeen)
	if err != nil {
		return mdefs.OwnState{}, err
	}
	err = bh.SetCanonicalSnapShot(CanonicalSnapShot)
	if err != nil {
		return mdefs.OwnState{}, err
	}
	err = bh.SetPendingSnapShot(PendingSnapShot)
	if err != nil {
		return mdefs.OwnState{}, err
	}
	return bh, nil
}

// Copy creates a copy of OwnState
func (b *OwnState) Copy() (*OwnState, error) {
	bdat, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	nobj := &OwnState{}
	err = nobj.UnmarshalBinary(bdat)
	if err != nil {
		return nil, err
	}
	return nobj, nil
}

// IsSync returns true if we are synced to the current block height
func (b *OwnState) IsSync() bool {
	if b == nil {
		return false
	}
	if b.MaxBHSeen.BClaims.Height-b.SyncToBH.BClaims.Height <= 1 {
		return true
	}
	return false
}
