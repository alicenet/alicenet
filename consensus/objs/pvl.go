package objs

import (
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
)

type PreVoteList []*PreVote
type PreVoteNilList []bool

func (pvl PreVoteList) MakePreCommit(secpSigner *crypto.Secp256k1Signer) (*PreCommit, error) {
	sigs := [][]byte{}
	for _, pv := range pvl {
		s := utils.CopySlice(pv.Signature)
		sigs = append(sigs, s)
	}
	propBytes, err := pvl[0].Proposal.MarshalBinary()
	if err != nil {
		return nil, err
	}
	prop := &Proposal{}
	err = prop.UnmarshalBinary(propBytes)
	if err != nil {
		return nil, err
	}
	pc := &PreCommit{
		Proposal: prop,
		PreVotes: sigs,
	}
	err = pc.Sign(secpSigner)
	if err != nil {
		return nil, err
	}
	return pc, nil
}

func (pvl PreVoteList) GetProposal() (*Proposal, error) {
	propBytes, err := pvl[0].Proposal.MarshalBinary()
	if err != nil {
		return nil, err
	}
	prop := &Proposal{}
	err = prop.UnmarshalBinary(propBytes)
	if err != nil {
		return nil, err
	}
	return prop, nil
}
