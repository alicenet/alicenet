package localrpc

import (
	"encoding/hex"

	to "github.com/alicenet/alicenet/consensus/objs"
	from "github.com/alicenet/alicenet/proto"
	"github.com/alicenet/alicenet/utils"
)

func ReverseTranslateBlockHeader(f *from.BlockHeader) (*to.BlockHeader, error) {
	t := &to.BlockHeader{}
	if f.BClaims != nil {
		BClaims, err := ReverseTranslateBClaims(f.BClaims)
		if err != nil {
			return nil, err
		}
		t.BClaims = BClaims
	}

	SigGroup, err := ReverseTranslateByte(f.SigGroup)
	if err != nil {
		return nil, err
	}
	t.SigGroup = SigGroup

	TxHshLst, err := ReverseTranslateByteSlice(f.TxHshLst)
	if err != nil {
		return nil, err
	}
	t.TxHshLst = TxHshLst
	return t, nil
}

func ReverseTranslateNRClaims(f *from.NRClaims) (*to.NRClaims, error) {
	t := &to.NRClaims{}
	if f.RCert != nil {
		RCert, err := ReverseTranslateRCert(f.RCert)
		if err != nil {
			return nil, err
		}
		t.RCert = RCert
	}

	if f.RClaims != nil {
		RClaims, err := ReverseTranslateRClaims(f.RClaims)
		if err != nil {
			return nil, err
		}
		t.RClaims = RClaims
	}

	SigShare, err := ReverseTranslateByte(f.SigShare)
	if err != nil {
		return nil, err
	}
	t.SigShare = SigShare
	return t, nil
}

func ReverseTranslatePreVoteNil(f *from.PreVoteNil) (*to.PreVoteNil, error) {
	t := &to.PreVoteNil{}
	if f.RCert != nil {
		RCert, err := ReverseTranslateRCert(f.RCert)
		if err != nil {
			return nil, err
		}
		t.RCert = RCert
	}

	Signature, err := ReverseTranslateByte(f.Signature)
	if err != nil {
		return nil, err
	}
	t.Signature = Signature
	return t, nil
}

func ReverseTranslateNextHeight(f *from.NextHeight) (*to.NextHeight, error) {
	t := &to.NextHeight{}
	if f.NHClaims != nil {
		NHClaims, err := ReverseTranslateNHClaims(f.NHClaims)
		if err != nil {
			return nil, err
		}
		t.NHClaims = NHClaims
	}

	PreCommits, err := ReverseTranslateByteSlice(f.PreCommits)
	if err != nil {
		return nil, err
	}
	t.PreCommits = PreCommits

	Signature, err := ReverseTranslateByte(f.Signature)
	if err != nil {
		return nil, err
	}
	t.Signature = Signature
	return t, nil
}

func ReverseTranslateRCert(f *from.RCert) (*to.RCert, error) {
	t := &to.RCert{}
	if f.RClaims != nil {
		RClaims, err := ReverseTranslateRClaims(f.RClaims)
		if err != nil {
			return nil, err
		}
		t.RClaims = RClaims
	}

	SigGroup, err := ReverseTranslateByte(f.SigGroup)
	if err != nil {
		return nil, err
	}
	t.SigGroup = SigGroup
	return t, nil
}

func ReverseTranslateNextRound(f *from.NextRound) (*to.NextRound, error) {
	t := &to.NextRound{}
	if f.NRClaims != nil {
		NRClaims, err := ReverseTranslateNRClaims(f.NRClaims)
		if err != nil {
			return nil, err
		}
		t.NRClaims = NRClaims
	}

	Signature, err := ReverseTranslateByte(f.Signature)
	if err != nil {
		return nil, err
	}
	t.Signature = Signature
	return t, nil
}

func ReverseTranslateProposal(f *from.Proposal) (*to.Proposal, error) {
	t := &to.Proposal{}
	if f.PClaims != nil {
		PClaims, err := ReverseTranslatePClaims(f.PClaims)
		if err != nil {
			return nil, err
		}
		t.PClaims = PClaims
	}

	Signature, err := ReverseTranslateByte(f.Signature)
	if err != nil {
		return nil, err
	}
	t.Signature = Signature

	TxHshLst, err := ReverseTranslateByteSlice(f.TxHshLst)
	if err != nil {
		return nil, err
	}
	t.TxHshLst = TxHshLst
	return t, nil
}

func ReverseTranslatePreCommit(f *from.PreCommit) (*to.PreCommit, error) {
	t := &to.PreCommit{}
	PreVotes, err := ReverseTranslateByteSlice(f.PreVotes)
	if err != nil {
		return nil, err
	}
	t.PreVotes = PreVotes

	if f.Proposal != nil {
		Proposal, err := ReverseTranslateProposal(f.Proposal)
		if err != nil {
			return nil, err
		}
		t.Proposal = Proposal
	}

	Signature, err := ReverseTranslateByte(f.Signature)
	if err != nil {
		return nil, err
	}
	t.Signature = Signature
	return t, nil
}

func ReverseTranslatePreVote(f *from.PreVote) (*to.PreVote, error) {
	t := &to.PreVote{}
	if f.Proposal != nil {
		Proposal, err := ReverseTranslateProposal(f.Proposal)
		if err != nil {
			return nil, err
		}
		t.Proposal = Proposal
	}

	Signature, err := ReverseTranslateByte(f.Signature)
	if err != nil {
		return nil, err
	}
	t.Signature = Signature
	return t, nil
}

func ReverseTranslateBClaims(f *from.BClaims) (*to.BClaims, error) {
	t := &to.BClaims{}
	t.ChainID = f.ChainID

	HeaderRoot, err := ReverseTranslateByte(f.HeaderRoot)
	if err != nil {
		return nil, err
	}
	t.HeaderRoot = HeaderRoot

	t.Height = f.Height

	PrevBlock, err := ReverseTranslateByte(f.PrevBlock)
	if err != nil {
		return nil, err
	}
	t.PrevBlock = PrevBlock

	StateRoot, err := ReverseTranslateByte(f.StateRoot)
	if err != nil {
		return nil, err
	}
	t.StateRoot = StateRoot

	t.TxCount = f.TxCount

	TxRoot, err := ReverseTranslateByte(f.TxRoot)
	if err != nil {
		return nil, err
	}
	t.TxRoot = TxRoot
	return t, nil
}

func ReverseTranslateNHClaims(f *from.NHClaims) (*to.NHClaims, error) {
	t := &to.NHClaims{}
	if f.Proposal != nil {
		Proposal, err := ReverseTranslateProposal(f.Proposal)
		if err != nil {
			return nil, err
		}
		t.Proposal = Proposal
	}

	SigShare, err := ReverseTranslateByte(f.SigShare)
	if err != nil {
		return nil, err
	}
	t.SigShare = SigShare
	return t, nil
}

func ReverseTranslatePClaims(f *from.PClaims) (*to.PClaims, error) {
	t := &to.PClaims{}
	if f.BClaims != nil {
		BClaims, err := ReverseTranslateBClaims(f.BClaims)
		if err != nil {
			return nil, err
		}
		t.BClaims = BClaims
	}

	if f.RCert != nil {
		RCert, err := ReverseTranslateRCert(f.RCert)
		if err != nil {
			return nil, err
		}
		t.RCert = RCert
	}
	return t, nil
}

func ReverseTranslateRClaims(f *from.RClaims) (*to.RClaims, error) {
	t := &to.RClaims{}

	t.ChainID = f.ChainID
	t.Height = f.Height

	PrevBlock, err := ReverseTranslateByte(f.PrevBlock)
	if err != nil {
		return nil, err
	}

	t.PrevBlock = PrevBlock
	t.Round = f.Round
	return t, nil
}

func ReverseTranslateByte(in string) ([]byte, error) {
	return utils.DecodeHexString(in)
}

func ReverseTranslateByteSlice(in []string) ([][]byte, error) {
	out := [][]byte{}
	for _, b := range in {
		s, err := hex.DecodeString(b)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}
