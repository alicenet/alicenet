package objs

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/prevotenil"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// PreVoteNil ...
type PreVoteNil struct {
	RCert     *RCert
	Signature []byte
	// Not Part of actual object below this line
	Voter    []byte
	GroupKey []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// PreVoteNil object
func (b *PreVoteNil) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("PreVoteNil.UnmarshalBinary; pvn not initialized")
	}
	bh, err := prevotenil.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *PreVoteNil) UnmarshalCapn(bh mdefs.PreVoteNil) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("PreVoteNil.UnmarshalCapn; pvn not initialized")
	}
	b.RCert = &RCert{}
	err := prevotenil.Validate(bh)
	if err != nil {
		return err
	}
	err = b.RCert.UnmarshalCapn(bh.RCert())
	if err != nil {
		return err
	}
	b.Signature = utils.CopySlice(bh.Signature())
	return nil
}

// MarshalBinary takes the PreVoteNil object and returns the canonical
// byte slice
func (b *PreVoteNil) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("PreVoteNil.MarshalBinary; pvn not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return prevotenil.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *PreVoteNil) MarshalCapn(seg *capnp.Segment) (mdefs.PreVoteNil, error) {
	if b == nil {
		return mdefs.PreVoteNil{}, errorz.ErrInvalid{}.New("PreVoteNil.MarshalCapn; pvn not initialized")
	}
	var bh mdefs.PreVoteNil
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootPreVoteNil(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewPreVoteNil(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	bc, err := b.RCert.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	err = bh.SetRCert(bc)
	if err != nil {
		return mdefs.PreVoteNil{}, err
	}
	err = bh.SetSignature(b.Signature)
	if err != nil {
		return mdefs.PreVoteNil{}, err
	}
	return bh, nil
}

func (b *PreVoteNil) ValidateSignatures(secpVal *crypto.Secp256k1Validator, bnVal *crypto.BNGroupValidator) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("PreVoteNil.ValidateSignatures; pvn not initialized")
	}
	err := b.RCert.ValidateSignature(bnVal)
	if err != nil {
		return err
	}
	canonicalEncoding, err := b.RCert.MarshalBinary()
	if err != nil {
		return err
	}
	PreVoteNilCE := []byte{}
	PreVoteNilCE = append(PreVoteNilCE, PreVoteNilSigDesignator()...)
	PreVoteNilCE = append(PreVoteNilCE, canonicalEncoding...)
	voter, err := secpVal.Validate(PreVoteNilCE, b.Signature)
	if err != nil {
		return err
	}
	b.Voter = crypto.GetAccount(voter)
	b.GroupKey = b.RCert.GroupKey
	return nil
}
