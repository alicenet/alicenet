package objs

import (
	"bytes"

	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/nextheight"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// NextHeight ...
type NextHeight struct {
	NHClaims   *NHClaims
	Signature  []byte
	PreCommits [][]byte
	// Not Part of actual object below this line
	Voter      []byte
	GroupKey   []byte
	GroupShare []byte
	Signers    [][]byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// NextHeight object
func (b *NextHeight) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("NextHeight.UnmarshalBinary; nh not initialized")
	}
	bh, err := nextheight.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *NextHeight) UnmarshalCapn(bh mdefs.NextHeight) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("NextHeight.UnmarshalCapn; nh not initialized")
	}
	b.NHClaims = &NHClaims{}
	err := nextheight.Validate(bh)
	if err != nil {
		return err
	}
	sigLst := bh.PreCommits()
	lst, err := SplitSignatures(sigLst)
	if err != nil {
		return err
	}
	b.PreCommits = lst
	err = b.NHClaims.UnmarshalCapn(bh.NHClaims())
	if err != nil {
		return err
	}
	b.Signature = utils.CopySlice(bh.Signature())
	return nil
}

// MarshalBinary takes the NextHeight object and returns the canonical
// byte slice
func (b *NextHeight) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("NextHeight.MarshalBinary; nh not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return nextheight.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *NextHeight) MarshalCapn(seg *capnp.Segment) (mdefs.NextHeight, error) {
	if b == nil {
		return mdefs.NextHeight{}, errorz.ErrInvalid{}.New("NextHeight.MarshalCapn; nh not initialized")
	}
	var bh mdefs.NextHeight
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootNextHeight(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewNextHeight(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	bc, err := b.NHClaims.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	if err := bh.SetNHClaims(bc); err != nil {
		return bh, err
	}
	if err := bh.SetSignature(b.Signature); err != nil {
		return bh, err
	}
	if err := bh.SetPreCommits(bytes.Join(b.PreCommits, []byte(""))); err != nil {
		return bh, err
	}
	return bh, nil
}

func (b *NextHeight) ValidateSignatures(secpVal *crypto.Secp256k1Validator, bnVal *crypto.BNGroupValidator) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("NextHeight.ValidateSignatures; nh not initialized")
	}
	if b.NHClaims == nil {
		return errorz.ErrInvalid{}.New("NextHeight.ValidateSignatures; nhclaims not initialized")
	}
	if b.NHClaims.Proposal == nil {
		return errorz.ErrInvalid{}.New("NextHeight.ValidateSignatures; proposal not initialized")
	}
	if b.NHClaims.Proposal.PClaims == nil {
		return errorz.ErrInvalid{}.New("NextHeight.ValidateSignatures; pclaims not initialized")
	}
	err := b.NHClaims.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		return err
	}
	canonicalEncoding, err := b.NHClaims.Proposal.PClaims.MarshalBinary()
	if err != nil {
		return err
	}
	NextHeightCE := []byte{}
	NextHeightCE = append(NextHeightCE, NextHeightSigDesignator()...)
	NextHeightCE = append(NextHeightCE, canonicalEncoding...)
	voter, err := secpVal.Validate(NextHeightCE, b.Signature)
	if err != nil {
		return err
	}
	addr := crypto.GetAccount(voter)
	b.Voter = addr

	for _, sig := range b.PreCommits {
		PreCommitCE := []byte{}
		PreCommitCE = append(PreCommitCE, PreCommitSigDesignator()...)
		PreCommitCE = append(PreCommitCE, canonicalEncoding...)
		pubkey, err := secpVal.Validate(PreCommitCE, utils.CopySlice(sig))
		if err != nil {
			return err
		}
		addr := crypto.GetAccount(pubkey)
		b.Signers = append(b.Signers, addr)
	}
	b.GroupKey = b.NHClaims.Proposal.PClaims.RCert.GroupKey
	b.GroupShare = b.NHClaims.GroupShare
	return nil
}

func (b *NextHeight) Plagiarize(secpSigner *crypto.Secp256k1Signer, bnSigner *crypto.BNGroupSigner) (*NextHeight, error) {
	nhb, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	nh := &NextHeight{}
	if err := nh.UnmarshalBinary(nhb); err != nil {
		return nil, err
	}
	if err := nh.Sign(secpSigner, bnSigner); err != nil {
		return nil, err
	}
	return nh, nil
}

func (b *NextHeight) Sign(secpSigner *crypto.Secp256k1Signer, bnSigner *crypto.BNGroupSigner) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("NextHeight.Sign; nh not initialized")
	}
	if b.NHClaims == nil {
		return errorz.ErrInvalid{}.New("NextHeight.Sign; nhclaims not initialized")
	}
	if b.NHClaims.Proposal == nil {
		return errorz.ErrInvalid{}.New("NextHeight.Sign; proposal not initialized")
	}
	if b.NHClaims.Proposal.PClaims == nil {
		return errorz.ErrInvalid{}.New("NextHeight.Sign; pclaims not initialized")
	}
	canonicalEncoding, err := b.NHClaims.Proposal.PClaims.MarshalBinary()
	if err != nil {
		return err
	}
	CE := []byte{}
	CE = append(CE, NextHeightSigDesignator()...)
	CE = append(CE, canonicalEncoding...)
	sig, err := secpSigner.Sign(CE)
	if err != nil {
		return err
	}
	b.Signature = sig
	bclaimsCanEnc, err := b.NHClaims.Proposal.PClaims.BClaims.BlockHash()
	if err != nil {
		return err
	}
	SigShare, err := bnSigner.Sign(bclaimsCanEnc)
	if err != nil {
		return err
	}
	b.NHClaims.SigShare = SigShare
	return nil
}
