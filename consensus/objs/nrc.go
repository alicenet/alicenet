package objs

import (
	"bytes"

	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/nrclaims"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// NRClaims ...
type NRClaims struct {
	RCert    *RCert
	RClaims  *RClaims
	SigShare []byte
	// Not Part of actual object below this line
	GroupShare []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// NRClaims object
func (b *NRClaims) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("NRClaims.UnmarshalBinary; nrclaims not initialized")
	}
	bh, err := nrclaims.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *NRClaims) UnmarshalCapn(bh mdefs.NRClaims) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("NRClaims.UnmarshalCapn; nrclaims not initialized")
	}
	b.RCert = &RCert{}
	b.RClaims = &RClaims{}
	err := nrclaims.Validate(bh)
	if err != nil {
		return err
	}
	err = b.RCert.UnmarshalCapn(bh.RCert())
	if err != nil {
		return err
	}
	err = b.RClaims.UnmarshalCapn(bh.RClaims())
	if err != nil {
		return err
	}
	b.SigShare = bh.SigShare()
	if b.RCert.RClaims.Height != b.RClaims.Height {
		return errorz.ErrInvalid{}.New("NRClaims.UnmarshalCapn; nrclaims height mismatch")
	}
	if b.RCert.RClaims.Round+1 != b.RClaims.Round {
		return errorz.ErrInvalid{}.New("NRClaims.UnmarshalCapn; nrclaims round not plus 1")
	}
	if b.RCert.RClaims.ChainID != b.RClaims.ChainID {
		return errorz.ErrInvalid{}.New("NRClaims.UnmarshalCapn; nrclaims chainID != chainID")
	}
	if !bytes.Equal(b.RCert.RClaims.PrevBlock, b.RClaims.PrevBlock) {
		return errorz.ErrInvalid{}.New("NRClaims.UnmarshalCapn; nrclaims prevBlock != prevBlock")
	}
	return nil
}

// MarshalBinary takes the NRClaims object and returns the canonical
// byte slice
func (b *NRClaims) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("NRClaims.MarshalBinary; nrclaims not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return nrclaims.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *NRClaims) MarshalCapn(seg *capnp.Segment) (mdefs.NRClaims, error) {
	if b == nil {
		return mdefs.NRClaims{}, errorz.ErrInvalid{}.New("NRClaims.MarshalCapn; nrclaims not initialized")
	}
	var bh mdefs.NRClaims
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootNRClaims(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewNRClaims(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	rc, err := b.RCert.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	rs, err := b.RClaims.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	err = bh.SetRCert(rc)
	if err != nil {
		return mdefs.NRClaims{}, err
	}
	err = bh.SetRClaims(rs)
	if err != nil {
		return mdefs.NRClaims{}, err
	}
	err = bh.SetSigShare(b.SigShare[:])
	if err != nil {
		return mdefs.NRClaims{}, err
	}
	return bh, nil
}

func (b *NRClaims) ValidateSignatures(bnVal *crypto.BNGroupValidator) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("NRClaims.ValidateSignatures; nrclaims not initialized")
	}
	err := b.RCert.ValidateSignature(bnVal)
	if err != nil {
		return err
	}
	canonicalEncoding, err := b.RClaims.MarshalBinary()
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

func (b *NRClaims) Sign(bnSigner *crypto.BNGroupSigner) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("NRClaims.Sign; nrclaims not initialized")
	}
	canonicalEncoding, err := b.RClaims.MarshalBinary()
	if err != nil {
		return err
	}
	sig, err := bnSigner.Sign(canonicalEncoding)
	if err != nil {
		return err
	}
	b.SigShare = sig
	return nil
}
