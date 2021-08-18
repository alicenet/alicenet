package objs

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/prevote"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	gUtils "github.com/MadBase/MadNet/utils"
	capnp "zombiezen.com/go/capnproto2"
)

// PreVote ...
type PreVote struct {
	Proposal  *Proposal
	Signature []byte
	// Not Part of actual object below this line
	Voter    []byte
	GroupKey []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// PreVote object
func (b *PreVote) UnmarshalBinary(data []byte) error {
	bh, err := prevote.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *PreVote) UnmarshalCapn(bh mdefs.PreVote) error {
	b.Proposal = &Proposal{}
	err := prevote.Validate(bh)
	if err != nil {
		return err
	}
	err = b.Proposal.UnmarshalCapn(bh.Proposal())
	if err != nil {
		return err
	}
	b.Signature = gUtils.CopySlice(bh.Signature())
	return nil
}

// MarshalBinary takes the PreVote object and returns the canonical
// byte slice
func (b *PreVote) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return prevote.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *PreVote) MarshalCapn(seg *capnp.Segment) (mdefs.PreVote, error) {
	if b == nil {
		return mdefs.PreVote{}, errorz.ErrInvalid{}.New("not initialized")
	}
	var bh mdefs.PreVote
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootPreVote(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewPreVote(seg)
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
		return mdefs.PreVote{}, err
	}
	err = bh.SetSignature(b.Signature)
	if err != nil {
		return mdefs.PreVote{}, err
	}
	return bh, nil
}

func (b *PreVote) ValidateSignatures(secpVal *crypto.Secp256k1Validator, bnVal *crypto.BNGroupValidator) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	err := b.Proposal.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		return err
	}
	canonicalEncoding, err := b.Proposal.PClaims.MarshalBinary()
	if err != nil {
		return err
	}
	CE := []byte{}
	CE = append(CE, PreVoteSigDesignator()...)
	CE = append(CE, canonicalEncoding...)
	voter, err := secpVal.Validate(CE, b.Signature)
	if err != nil {
		return err
	}
	b.Voter = crypto.GetAccount(voter)
	b.GroupKey = b.Proposal.PClaims.RCert.GroupKey
	return nil
}

func (b *PreVote) Sign(secpSigner *crypto.Secp256k1Signer) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	canonicalEncoding, err := b.Proposal.PClaims.MarshalBinary()
	if err != nil {
		return err
	}
	CE := []byte{}
	CE = append(CE, PreVoteSigDesignator()...)
	CE = append(CE, canonicalEncoding...)
	sig, err := secpSigner.Sign(CE)
	if err != nil {
		return err
	}
	b.Signature = sig
	return nil
}
