package objs

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/nextround"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// NextRound ...
type NextRound struct {
	NRClaims  *NRClaims
	Signature []byte
	// Not Part of actual object below this line
	Voter      []byte
	GroupKey   []byte
	GroupShare []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// NextRound object
func (b *NextRound) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("NextRound.UnmarshalBinary; nr not initialized")
	}
	bh, err := nextround.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *NextRound) UnmarshalCapn(bh mdefs.NextRound) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("NextRound.UnmarshalCapn; nr not initialized")
	}
	b.NRClaims = &NRClaims{}
	err := nextround.Validate(bh)
	if err != nil {
		return err
	}
	err = b.NRClaims.UnmarshalCapn(bh.NRClaims())
	if err != nil {
		return err
	}
	b.Signature = bh.Signature()
	return nil
}

// MarshalBinary takes the NextRound object and returns the canonical
// byte slice
func (b *NextRound) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("NextRound.MarshalBinary; nr not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return nextround.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *NextRound) MarshalCapn(seg *capnp.Segment) (mdefs.NextRound, error) {
	if b == nil {
		return mdefs.NextRound{}, errorz.ErrInvalid{}.New("NextRound.MarshalCapn; nr not initialized")
	}
	var bh mdefs.NextRound
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootNextRound(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewNextRound(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	rc, err := b.NRClaims.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	err = bh.SetNRClaims(rc)
	if err != nil {
		return mdefs.NextRound{}, err
	}
	err = bh.SetSignature(b.Signature)
	if err != nil {
		return mdefs.NextRound{}, err
	}
	return bh, nil
}

func (b *NextRound) Sign(secpSigner *crypto.Secp256k1Signer, bnSigner *crypto.BNGroupSigner) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("NextRound.Sign; nr not initialized")
	}
	err := b.NRClaims.Sign(bnSigner)
	if err != nil {
		return err
	}
	canonicalEncoding, err := b.NRClaims.MarshalBinary()
	if err != nil {
		return err
	}
	NextRoundCE := []byte{}
	NextRoundCE = append(NextRoundCE, NextRoundSigDesignator()...)
	NextRoundCE = append(NextRoundCE, canonicalEncoding...)
	sig, err := secpSigner.Sign(NextRoundCE)
	if err != nil {
		return err
	}
	b.Signature = sig
	return nil
}

func (b *NextRound) ValidateSignatures(secpVal *crypto.Secp256k1Validator, bnVal *crypto.BNGroupValidator) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("NextRound.ValidateSignatures; nr not initialized")
	}
	err := b.NRClaims.ValidateSignatures(bnVal)
	if err != nil {
		return err
	}
	canonicalEncoding, err := b.NRClaims.MarshalBinary()
	if err != nil {
		return err
	}
	NextRoundCE := []byte{}
	NextRoundCE = append(NextRoundCE, NextRoundSigDesignator()...)
	NextRoundCE = append(NextRoundCE, canonicalEncoding...)
	voter, err := secpVal.Validate(NextRoundCE, b.Signature)
	if err != nil {
		return err
	}
	addr := crypto.GetAccount(voter)
	b.Voter = addr
	b.GroupKey = b.NRClaims.RCert.GroupKey
	b.GroupShare = b.NRClaims.GroupShare
	return nil
}
