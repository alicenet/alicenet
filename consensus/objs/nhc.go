package objs

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/nhclaims"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	capnp "zombiezen.com/go/capnproto2"
)

// NHClaims ...
type NHClaims struct {
	Proposal *Proposal
	SigShare []byte
	// Not Part of actual object below this line
	GroupShare []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// NHClaims object
func (b *NHClaims) UnmarshalBinary(data []byte) error {
	bh, err := nhclaims.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *NHClaims) UnmarshalCapn(bh mdefs.NHClaims) error {
	b.Proposal = &Proposal{}
	err := nhclaims.Validate(bh)
	if err != nil {
		return err
	}
	err = b.Proposal.UnmarshalCapn(bh.Proposal())
	if err != nil {
		return err
	}
	b.SigShare = bh.SigShare()
	return nil
}

// MarshalBinary takes the NHClaims object and returns the canonical
// byte slice
func (b *NHClaims) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return nhclaims.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *NHClaims) MarshalCapn(seg *capnp.Segment) (mdefs.NHClaims, error) {
	if b == nil {
		return mdefs.NHClaims{}, errorz.ErrInvalid{}.New("not initialized")
	}
	var bh mdefs.NHClaims
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootNHClaims(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewNHClaims(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	bc, err := b.Proposal.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	err = bh.SetProposal(bc)
	if err != nil {
		return mdefs.NHClaims{}, err
	}
	err = bh.SetSigShare(b.SigShare)
	if err != nil {
		return mdefs.NHClaims{}, err
	}
	return bh, nil
}

func (b *NHClaims) ValidateSignatures(secpVal *crypto.Secp256k1Validator, bnVal *crypto.BNGroupValidator) error {
	if b == nil || b.Proposal == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	err := b.Proposal.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		return err
	}
	canonicalEncoding, err := b.Proposal.PClaims.BClaims.BlockHash()
	if err != nil {
		return err
	}
	SigSharePubk, err := bnVal.Validate(canonicalEncoding, b.SigShare)
	if err != nil {
		return err
	}
	b.GroupShare = SigSharePubk
	return nil
}
