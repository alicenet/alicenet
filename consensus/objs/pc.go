package objs

import (
	"bytes"

	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/precommit"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// PreCommit ...
type PreCommit struct {
	Proposal  *Proposal
	Signature []byte
	PreVotes  [][]byte
	// Not Part of actual object below this line
	Voter    []byte
	Signers  [][]byte
	GroupKey []byte
	Proposer []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// PreCommit object
func (b *PreCommit) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("PreCommit.UnmarshalBinary; pc not initialized")
	}
	bh, err := precommit.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *PreCommit) UnmarshalCapn(bh mdefs.PreCommit) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("PreCommit.UnmarshalCapn; pc not initialized")
	}
	b.Proposal = &Proposal{}
	err := precommit.Validate(bh)
	if err != nil {
		return err
	}
	sigLst := bh.PreVotes()
	lst, err := SplitSignatures(sigLst)
	if err != nil {
		return err
	}
	b.PreVotes = lst
	err = b.Proposal.UnmarshalCapn(bh.Proposal())
	if err != nil {
		return err
	}
	b.Signature = bh.Signature()
	return nil
}

// MarshalBinary takes the PreCommit object and returns the canonical
// byte slice
func (b *PreCommit) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("PreCommit.MarshalBinary; pc not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return precommit.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *PreCommit) MarshalCapn(seg *capnp.Segment) (mdefs.PreCommit, error) {
	if b == nil {
		return mdefs.PreCommit{}, errorz.ErrInvalid{}.New("PreCommit.MarshalCapn; pc not initialized")
	}
	var bh mdefs.PreCommit
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootPreCommit(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewPreCommit(seg)
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
		return mdefs.PreCommit{}, err
	}
	err = bh.SetSignature(b.Signature)
	if err != nil {
		return mdefs.PreCommit{}, err
	}
	err = bh.SetPreVotes(bytes.Join(b.PreVotes, []byte("")))
	if err != nil {
		return mdefs.PreCommit{}, err
	}
	return bh, nil
}

func (b *PreCommit) ValidateSignatures(secpVal *crypto.Secp256k1Validator, bnVal *crypto.BNGroupValidator) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("PreCommit.ValidateSignatures; pc not initialized")
	}
	if b.Proposal == nil {
		return errorz.ErrInvalid{}.New("PreCommit.ValidateSignatures; proposal not initialized")
	}
	if b.Proposal.PClaims == nil {
		return errorz.ErrInvalid{}.New("PreCommit.ValidateSignatures; pclaims not initialized")
	}
	if b.Proposal.PClaims.RCert == nil {
		return errorz.ErrInvalid{}.New("PreCommit.ValidateSignatures; rcert not initialized")
	}
	err := b.Proposal.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		return err
	}
	b.Proposer = b.Proposal.Proposer
	b.GroupKey = b.Proposal.PClaims.RCert.GroupKey

	canonicalEncoding, err := b.Proposal.PClaims.MarshalBinary()
	if err != nil {
		return err
	}

	CE := []byte{}
	CE = append(CE, PreCommitSigDesignator()...)
	CE = append(CE, canonicalEncoding...)
	voter, err := secpVal.Validate(CE, b.Signature)
	if err != nil {
		return err
	}
	b.Voter = crypto.GetAccount(voter)

	canonicalEncoding, err = b.Proposal.PClaims.MarshalBinary()
	if err != nil {
		return err
	}
	CE = []byte{}
	CE = append(CE, PreVoteSigDesignator()...)
	CE = append(CE, canonicalEncoding...)
	for _, sig := range b.PreVotes {
		pubkey, err := secpVal.Validate(CE, utils.CopySlice(sig))
		if err != nil {
			return err
		}
		addr := crypto.GetAccount(pubkey)
		b.Signers = append(b.Signers, addr)
	}
	return nil
}

func (b *PreCommit) MakeImplPreVotes() (PreVoteList, error) {
	pvl := PreVoteList{}
	for idx, pv := range b.PreVotes {
		pcBytes, err := b.MarshalBinary()
		if err != nil {
			return nil, err
		}
		pc := &PreCommit{}
		err = pc.UnmarshalBinary(pcBytes)
		if err != nil {
			return nil, err
		}
		groupKey := utils.CopySlice(b.GroupKey)
		voter := utils.CopySlice(b.Signers[idx])
		sig := utils.CopySlice(pv)
		pV := &PreVote{
			Signature: sig,
			Proposal:  pc.Proposal,
			GroupKey:  groupKey,
			Voter:     voter,
		}
		pvl = append(pvl, pV)
	}
	return pvl, nil
}

func (b *PreCommit) Sign(secpSigner *crypto.Secp256k1Signer) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("PreCommit.Sign; pc not initialized")
	}
	if b.Proposal == nil {
		return errorz.ErrInvalid{}.New("PreCommit.Sign; proposal not initialized")
	}
	canonicalEncoding, err := b.Proposal.PClaims.MarshalBinary()
	if err != nil {
		return err
	}
	CE := []byte{}
	CE = append(CE, PreCommitSigDesignator()...)
	CE = append(CE, canonicalEncoding...)
	sig, err := secpSigner.Sign(CE)
	if err != nil {
		return err
	}
	b.Signature = sig
	return nil
}
