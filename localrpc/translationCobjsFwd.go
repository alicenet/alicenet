package localrpc

import (
	"encoding/hex"
	"errors"

	from "github.com/alicenet/alicenet/consensus/objs"
	to "github.com/alicenet/alicenet/proto"
	"github.com/alicenet/alicenet/utils"
)

func ForwardTranslateBlockHeader(f *from.BlockHeader) (*to.BlockHeader, error) {
	t := &to.BlockHeader{}
	if f == nil {
		return nil, errors.New("blockHeader object should not be nil")
	}

	if f.BClaims != nil {
		BClaims, err := ForwardTranslateBClaims(f.BClaims)
		if err != nil {
			return nil, err
		}
		t.BClaims = BClaims
	}

	SigGroup := ForwardTranslateByte(f.SigGroup)

	t.SigGroup = SigGroup

	TxHshLst, err := ForwardTranslateByteSlice(f.TxHshLst)
	if err != nil {
		return nil, err
	}
	t.TxHshLst = TxHshLst
	return t, nil
}

func ForwardTranslateNRClaimsNRClaims(f *from.NRClaims) (*to.NRClaims, error) {
	t := &to.NRClaims{}
	if f == nil {
		return nil, errors.New("NRClaims object should not be nil")
	}

	if f.RCert != nil {
		RCert, err := ForwardTranslateRCert(f.RCert)
		if err != nil {
			return nil, err
		}
		t.RCert = RCert
	}

	if f.RClaims != nil {
		RClaims, err := ForwardTranslateRClaims(f.RClaims)
		if err != nil {
			return nil, err
		}
		t.RClaims = RClaims
	}

	SigShare := ForwardTranslateByte(f.SigShare)

	t.SigShare = SigShare
	return t, nil
}

func ForwardTranslatePreVoteNil(f *from.PreVoteNil) (*to.PreVoteNil, error) {
	t := &to.PreVoteNil{}
	if f == nil {
		return nil, errors.New("PreVoteNil object should not be nil")
	}

	if f.RCert != nil {
		RCert, err := ForwardTranslateRCert(f.RCert)
		if err != nil {
			return nil, err
		}
		t.RCert = RCert
	}

	Signature := ForwardTranslateByte(f.Signature)

	t.Signature = Signature
	return t, nil
}

func ForwardTranslateNextHeight(f *from.NextHeight) (*to.NextHeight, error) {
	t := &to.NextHeight{}
	if f == nil {
		return nil, errors.New("NextHeight object should not be nil")
	}

	if f.NHClaims != nil {
		NHClaims, err := ForwardTranslateNHClaimsNHClaims(f.NHClaims)
		if err != nil {
			return nil, err
		}
		t.NHClaims = NHClaims
	}

	PreCommits, err := ForwardTranslateByteSlice(f.PreCommits)
	if err != nil {
		return nil, err
	}
	t.PreCommits = PreCommits

	Signature := ForwardTranslateByte(f.Signature)

	t.Signature = Signature
	return t, nil
}

func ForwardTranslateRCert(f *from.RCert) (*to.RCert, error) {
	t := &to.RCert{}
	if f == nil {
		return nil, errors.New("RCert object should not be nil")
	}

	if f.RClaims != nil {
		RClaims, err := ForwardTranslateRClaims(f.RClaims)
		if err != nil {
			return nil, err
		}
		t.RClaims = RClaims
	}

	SigGroup := ForwardTranslateByte(f.SigGroup)

	t.SigGroup = SigGroup
	return t, nil
}

func ForwardTranslateNextRound(f *from.NextRound) (*to.NextRound, error) {
	t := &to.NextRound{}
	if f == nil {
		return nil, errors.New("NextRound object should not be nil")
	}

	if f.NRClaims != nil {
		NRClaims, err := ForwardTranslateNRClaimsNRClaims(f.NRClaims)
		if err != nil {
			return nil, err
		}
		t.NRClaims = NRClaims
	}

	Signature := ForwardTranslateByte(f.Signature)

	t.Signature = Signature
	return t, nil
}

func ForwardTranslateProposal(f *from.Proposal) (*to.Proposal, error) {
	t := &to.Proposal{}
	if f == nil {
		return nil, errors.New("proposal object should not be nil")
	}

	if f.PClaims != nil {
		PClaims, err := ForwardTranslatePClaims(f.PClaims)
		if err != nil {
			return nil, err
		}
		t.PClaims = PClaims
	}

	Signature := ForwardTranslateByte(f.Signature)

	t.Signature = Signature

	TxHshLst, err := ForwardTranslateByteSlice(f.TxHshLst)
	if err != nil {
		return nil, err
	}
	t.TxHshLst = TxHshLst
	return t, nil
}

func ForwardTranslatePreCommit(f *from.PreCommit) (*to.PreCommit, error) {
	t := &to.PreCommit{}
	if f == nil {
		return nil, errors.New("PreCommit object should not be nil")
	}

	PreVotes, err := ForwardTranslateByteSlice(f.PreVotes)
	if err != nil {
		return nil, err
	}
	t.PreVotes = PreVotes

	if f.Proposal != nil {
		Proposal, err := ForwardTranslateProposal(f.Proposal)
		if err != nil {
			return nil, err
		}
		t.Proposal = Proposal
	}

	Signature := ForwardTranslateByte(f.Signature)

	t.Signature = Signature
	return t, nil
}

func ForwardTranslatePreVote(f *from.PreVote) (*to.PreVote, error) {
	t := &to.PreVote{}
	if f == nil {
		return nil, errors.New("PreVote object should not be nil")
	}

	if f.Proposal != nil {
		Proposal, err := ForwardTranslateProposal(f.Proposal)
		if err != nil {
			return nil, err
		}
		t.Proposal = Proposal
	}

	Signature := ForwardTranslateByte(f.Signature)

	t.Signature = Signature
	return t, nil
}

func ForwardTranslateBClaims(f *from.BClaims) (*to.BClaims, error) {
	t := &to.BClaims{}
	if f == nil {
		return nil, errors.New("BClaims object should not be nil")
	}

	t.ChainID = f.ChainID

	HeaderRoot := ForwardTranslateByte(f.HeaderRoot)

	t.HeaderRoot = HeaderRoot

	t.Height = f.Height

	PrevBlock := ForwardTranslateByte(f.PrevBlock)

	t.PrevBlock = PrevBlock

	StateRoot := ForwardTranslateByte(f.StateRoot)

	t.StateRoot = StateRoot

	t.TxCount = f.TxCount

	TxRoot := ForwardTranslateByte(f.TxRoot)

	t.TxRoot = TxRoot
	return t, nil
}

func ForwardTranslateNHClaimsNHClaims(f *from.NHClaims) (*to.NHClaims, error) {
	t := &to.NHClaims{}
	if f == nil {
		return nil, errors.New("NHClaims object should not be nil")
	}

	if f.Proposal != nil {
		Proposal, err := ForwardTranslateProposal(f.Proposal)
		if err != nil {
			return nil, err
		}
		t.Proposal = Proposal
	}

	SigShare := ForwardTranslateByte(f.SigShare)

	t.SigShare = SigShare
	return t, nil
}

func ForwardTranslatePClaims(f *from.PClaims) (*to.PClaims, error) {
	t := &to.PClaims{}
	if f == nil {
		return nil, errors.New("PClaims object should not be nil")
	}

	if f.BClaims != nil {
		BClaims, err := ForwardTranslateBClaims(f.BClaims)
		if err != nil {
			return nil, err
		}
		t.BClaims = BClaims
	}

	if f.RCert != nil {
		RCert, err := ForwardTranslateRCert(f.RCert)
		if err != nil {
			return nil, err
		}
		t.RCert = RCert
	}
	return t, nil
}

func ForwardTranslateRClaims(f *from.RClaims) (*to.RClaims, error) {
	t := &to.RClaims{}
	if f == nil {
		return nil, errors.New("RClaims object should not be nil")
	}

	t.ChainID = f.ChainID
	t.Height = f.Height

	PrevBlock := ForwardTranslateByte(f.PrevBlock)

	t.PrevBlock = PrevBlock

	t.Round = f.Round
	return t, nil
}

func ForwardTranslateByte(in []byte) string {
	return utils.EncodeHexString(in)
}

func ForwardTranslateByteSlice(in [][]byte) ([]string, error) {
	out := []string{}
	for _, b := range in {
		out = append(out, hex.EncodeToString(utils.CopySlice(b)))
	}
	return out, nil
}
