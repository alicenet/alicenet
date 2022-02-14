package objs

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/nhclaims"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
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
	if b == nil {
		return errorz.ErrInvalid{}.New("NHClaims.UnmarshalBinary; nhclaims not initialized")
	}
	bh, err := nhclaims.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *NHClaims) UnmarshalCapn(bh mdefs.NHClaims) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("NHClaims.UnmarshalCapn; nhclaims not initialized")
	}
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
		return nil, errorz.ErrInvalid{}.New("NHClaims.MarshalBinary; nhclaims not initialized")
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
		return mdefs.NHClaims{}, errorz.ErrInvalid{}.New("NHClaims.MarshalCapn; nhclaims not initialized")
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
	if b == nil {
		return errorz.ErrInvalid{}.New("NHClaims.ValidateSignatures; nhclaims not initialized")
	}
	if b.Proposal == nil {
		return errorz.ErrInvalid{}.New("NHClaims.ValidateSignatures; proposal not initialized")
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
