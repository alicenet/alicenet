package objs

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/precommitnil"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// PreCommitNil ...
type PreCommitNil struct {
	RCert     *RCert
	Signature []byte
	// Not Part of actual object below this line
	Voter    []byte
	GroupKey []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// PreCommitNil object
func (b *PreCommitNil) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("PreCommitNil.UnmarshalBinary; pcn not initialized")
	}
	bh, err := precommitnil.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *PreCommitNil) UnmarshalCapn(bh mdefs.PreCommitNil) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("PreCommitNil.UnmarshalCapn; pcn not initialized")
	}
	b.RCert = &RCert{}
	err := precommitnil.Validate(bh)
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

// MarshalBinary takes the PreCommitNil object and returns the canonical
// byte slice
func (b *PreCommitNil) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("PreCommitNil.MarshalBinary; pcn not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return precommitnil.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *PreCommitNil) MarshalCapn(seg *capnp.Segment) (mdefs.PreCommitNil, error) {
	if b == nil {
		return mdefs.PreCommitNil{}, errorz.ErrInvalid{}.New("PreCommitNil.MarshalCapn; pcn not initialized")
	}
	var bh mdefs.PreCommitNil
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootPreCommitNil(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewPreCommitNil(seg)
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
		return mdefs.PreCommitNil{}, err
	}
	err = bh.SetSignature(b.Signature)
	if err != nil {
		return mdefs.PreCommitNil{}, err
	}
	return bh, nil
}

func (b *PreCommitNil) ValidateSignatures(secpVal *crypto.Secp256k1Validator, bnVal *crypto.BNGroupValidator) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("PreCommitNil.ValidateSignatures; pcn not initialized")
	}
	err := b.RCert.ValidateSignature(bnVal)
	if err != nil {
		return err
	}
	canonicalEncoding, err := b.RCert.MarshalBinary()
	if err != nil {
		return err
	}
	PreCommitNilCE := []byte{}
	PreCommitNilCE = append(PreCommitNilCE, PreCommitNilSigDesignator()...)
	PreCommitNilCE = append(PreCommitNilCE, canonicalEncoding...)
	voter, err := secpVal.Validate(PreCommitNilCE, b.Signature)
	if err != nil {
		return err
	}
	addr := crypto.GetAccount(voter)
	b.Voter = addr
	b.GroupKey = b.RCert.GroupKey
	return nil
}
