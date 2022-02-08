package objs

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/rcert"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// RCert ...
type RCert struct {
	RClaims  *RClaims
	SigGroup []byte
	// Not Part of actual object below this line
	GroupKey []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// RCert object
func (b *RCert) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("RCert.UnmarshalBinary; rcert not initialized")
	}
	bh, err := rcert.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *RCert) UnmarshalCapn(bh mdefs.RCert) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("RCert.UnmarshalCapn; rcert not initialized")
	}
	b.RClaims = &RClaims{}
	err := rcert.Validate(bh)
	if err != nil {
		return err
	}
	err = b.RClaims.UnmarshalCapn(bh.RClaims())
	if err != nil {
		return err
	}
	b.SigGroup = utils.CopySlice(bh.SigGroup())
	return nil
}

// MarshalBinary takes the RCert object and returns the canonical
// byte slice
func (b *RCert) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("RCert.MarshalBinary; rcert not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return rcert.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *RCert) MarshalCapn(seg *capnp.Segment) (mdefs.RCert, error) {
	if b == nil {
		return mdefs.RCert{}, errorz.ErrInvalid{}.New("RCert.MarshalCapn; rcert not initialized")
	}
	var bh mdefs.RCert
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootRCert(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewRCert(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	bc, err := b.RClaims.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	err = bh.SetRClaims(bc)
	if err != nil {
		return mdefs.RCert{}, err
	}
	err = bh.SetSigGroup(b.SigGroup)
	if err != nil {
		return mdefs.RCert{}, err
	}
	return bh, nil
}

// ValidateSignature validates the group signature on the RCert
func (b *RCert) ValidateSignature(bnVal *crypto.BNGroupValidator) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("RCert.ValidateSignatures; rcert not initialized")
	}
	if b.RClaims == nil {
		return errorz.ErrInvalid{}.New("RCert.ValidateSignatures; rclaims not initialized")
	}
	if b.RClaims.Height == 0 {
		return errorz.ErrInvalid{}.New("RCert.ValidateSignatures; height is zero")
	}
	if b.RClaims.ChainID == 0 {
		return errorz.ErrInvalid{}.New("RCert.ValidateSignatures; chainID is zero")
	}
	if b.RClaims.Round == 0 {
		return errorz.ErrInvalid{}.New("RCert.ValidateSignatures; round is zero")
	}
	if b.RClaims.Round > constants.DEADBLOCKROUND {
		return errorz.ErrInvalid{}.New("RCert.ValidateSignatures; round > DeadBlockRound")
	}
	if b.RClaims.Height == 1 && b.RClaims.Round == 1 {
		b.GroupKey = make([]byte, constants.CurveBN256EthPubkeyLen)
		return nil
	}
	if b.RClaims.Height == 1 && b.RClaims.Round > 1 {
		// No such RCert should exist, so raise an error
		return errorz.ErrInvalid{}.New("RCert.ValidateSignatures; RCert should not exist!")
	}
	if b.RClaims.Height == 2 && b.RClaims.Round == 1 {
		// There is nothing we can check because there is no group signature
		return nil
	}
	if len(b.RClaims.PrevBlock) != constants.HashLen {
		return errorz.ErrInvalid{}.New("RCert.ValidateSignatures; invalid PrevBlock")
	}
	if b.RClaims.Round > 1 {
		canonicalEncoding, err := b.RClaims.MarshalBinary()
		if err != nil {
			return err
		}
		groupKey, err := bnVal.Validate(canonicalEncoding, b.SigGroup)
		if err != nil {
			return err
		}
		b.GroupKey = groupKey
		return nil
	}
	groupKey, err := bnVal.Validate(b.RClaims.PrevBlock, b.SigGroup)
	if err != nil {
		return err
	}
	b.GroupKey = groupKey
	return nil
}

// PreVoteNil constructs a PreVoteNil object from RCert
func (b *RCert) PreVoteNil(secpSigner *crypto.Secp256k1Signer) (*PreVoteNil, error) {
	rce, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	pvn := &PreVoteNil{}
	pvn.RCert = &RCert{}
	err = pvn.RCert.UnmarshalBinary(rce)
	if err != nil {
		return nil, err
	}
	canonicalEncoding, err := pvn.RCert.MarshalBinary()
	if err != nil {
		return nil, err
	}
	PreVoteNilCE := []byte{}
	PreVoteNilCE = append(PreVoteNilCE, PreVoteNilSigDesignator()...)
	PreVoteNilCE = append(PreVoteNilCE, canonicalEncoding...)
	sig, err := secpSigner.Sign(PreVoteNilCE)
	if err != nil {
		return nil, err
	}
	pvn.Signature = sig
	return pvn, nil
}

// PreCommitNil constructs a PreCommitNil object from RCert
func (b *RCert) PreCommitNil(secpSigner *crypto.Secp256k1Signer) (*PreCommitNil, error) {
	rce, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	pcn := &PreCommitNil{}
	pcn.RCert = &RCert{}
	err = pcn.RCert.UnmarshalBinary(rce)
	if err != nil {
		return nil, err
	}
	canonicalEncoding, err := pcn.RCert.MarshalBinary()
	if err != nil {
		return nil, err
	}
	PreCommitNilCE := []byte{}
	PreCommitNilCE = append(PreCommitNilCE, PreCommitNilSigDesignator()...)
	PreCommitNilCE = append(PreCommitNilCE, canonicalEncoding...)
	sig, err := secpSigner.Sign(PreCommitNilCE)
	if err != nil {
		return nil, err
	}
	pcn.Signature = sig
	return pcn, nil
}

// NextRound constructs a NextRound object from RCert
func (b *RCert) NextRound(secpSigner *crypto.Secp256k1Signer, bnSigner *crypto.BNGroupSigner) (*NextRound, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("RCert.NextRound; rcert not initialized")
	}
	rcClaims, err := b.RClaims.MarshalBinary()
	if err != nil {
		return nil, err
	}
	nrrClaims := &RClaims{}
	err = nrrClaims.UnmarshalBinary(rcClaims)
	if err != nil {
		return nil, err
	}
	nrrClaims.Round++
	nrrc := &RCert{}
	rce, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = nrrc.UnmarshalBinary(rce)
	if err != nil {
		return nil, err
	}
	NRClaims := &NRClaims{
		RCert:   nrrc,
		RClaims: nrrClaims,
	}
	nr := &NextRound{
		NRClaims: NRClaims,
	}
	err = nr.Sign(secpSigner, bnSigner)
	if err != nil {
		return nil, err
	}
	return nr, nil
}
