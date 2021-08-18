package objs

import (
	"bytes"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"

	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/proposal"
	"github.com/MadBase/MadNet/crypto"
	gUtils "github.com/MadBase/MadNet/utils"
	capnp "zombiezen.com/go/capnproto2"
)

//Proposal ...
type Proposal struct {
	PClaims   *PClaims
	Signature []byte
	TxHshLst  [][]byte
	// Not Part of actual object below this line
	Proposer []byte
	GroupKey []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// Proposal object
func (b *Proposal) UnmarshalBinary(data []byte) error {
	bh, err := proposal.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *Proposal) UnmarshalCapn(bh mdefs.Proposal) error {
	b.PClaims = &PClaims{}
	err := proposal.Validate(bh)
	if err != nil {
		return err
	}
	txHshLst := bh.TxHshLst()
	lst, err := SplitHashes(txHshLst)
	if err != nil {
		return err
	}
	err = b.PClaims.UnmarshalCapn(bh.PClaims())
	if err != nil {
		return err
	}
	b.TxHshLst = lst
	b.Signature = gUtils.CopySlice(bh.Signature())
	if b.PClaims.RCert.RClaims.Round == constants.DEADBLOCKROUND {
		if len(b.TxHshLst) != 0 {
			return errorz.ErrInvalid{}.New("non empty tx hash lst in deadblockround")
		}
	}
	return nil
}

// MarshalBinary takes the Proposal object and returns the canonical
// byte slice
func (b *Proposal) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return proposal.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *Proposal) MarshalCapn(seg *capnp.Segment) (mdefs.Proposal, error) {
	if b == nil {
		return mdefs.Proposal{}, errorz.ErrInvalid{}.New("not initialized")
	}
	var bh mdefs.Proposal
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootProposal(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewProposal(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	bc, err := b.PClaims.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	err = bh.SetPClaims(bc)
	if err != nil {
		return mdefs.Proposal{}, err
	}
	err = bh.SetSignature(b.Signature)
	if err != nil {
		return mdefs.Proposal{}, err
	}
	err = bh.SetTxHshLst(bytes.Join(b.TxHshLst, []byte("")))
	if err != nil {
		return mdefs.Proposal{}, err
	}
	return bh, nil
}

func (b *Proposal) RePropose(secpSigner *crypto.Secp256k1Signer, rc *RCert) (*Proposal, error) {
	pce, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	p := &Proposal{}
	err = p.UnmarshalBinary(pce)
	if err != nil {
		return nil, err
	}
	rcb, err := rc.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = p.PClaims.RCert.UnmarshalBinary(rcb)
	if err != nil {
		return nil, err
	}
	err = p.Sign(secpSigner)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (b *Proposal) Sign(secpSigner *crypto.Secp256k1Signer) error {
	if b == nil || b.PClaims == nil || b.PClaims.RCert == nil || b.PClaims.RCert.RClaims == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	if b.PClaims.RCert.RClaims.Round > constants.DEADBLOCKROUND {
		return errorz.ErrInvalid{}.New("proposal round > DBR")
	}
	canonicalEncoding, err := b.PClaims.MarshalBinary()
	if err != nil {
		return err
	}
	ProposalCE := []byte{}
	ProposalCE = append(ProposalCE, ProposalSigDesignator()...)
	ProposalCE = append(ProposalCE, canonicalEncoding...)
	sig, err := secpSigner.Sign(ProposalCE)
	if err != nil {
		return err
	}
	b.Signature = sig
	return nil
}

func (b *Proposal) ValidateSignatures(val *crypto.Secp256k1Validator, bnVal *crypto.BNGroupValidator) error {
	if b == nil || b.PClaims == nil || b.PClaims.BClaims == nil || b.PClaims.RCert == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	txRoot, err := MakeTxRoot(b.TxHshLst)
	if err != nil {
		return err
	}
	if !bytes.Equal(txRoot, b.PClaims.BClaims.TxRoot) {
		return errorz.ErrInvalid{}.New("proposal txroot mismatch")
	}
	err = b.PClaims.RCert.ValidateSignature(bnVal)
	if err != nil {
		return err
	}
	if b.PClaims.RCert.RClaims.Round > constants.DEADBLOCKROUND {
		return errorz.ErrInvalid{}.New("proposal rcert round > DBR")
	}
	canonicalEncoding, err := b.PClaims.MarshalBinary()
	if err != nil {
		return err
	}
	ProposalCE := []byte{}
	ProposalCE = append(ProposalCE, ProposalSigDesignator()...)
	ProposalCE = append(ProposalCE, canonicalEncoding...)
	proposer, err := val.Validate(ProposalCE, b.Signature)
	if err != nil {
		return err
	}
	b.Proposer = crypto.GetAccount(proposer)
	b.GroupKey = b.PClaims.RCert.GroupKey
	return nil
}

func (b *Proposal) PreVote(secpSigner *crypto.Secp256k1Signer) (*PreVote, error) {
	pcb, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	pv := &PreVote{}
	pv.Proposal = &Proposal{}
	err = pv.Proposal.UnmarshalBinary(pcb)
	if err != nil {
		return nil, err
	}
	err = pv.Sign(secpSigner)
	if err != nil {
		return nil, err
	}
	return pv, nil
}
