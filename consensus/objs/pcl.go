package objs

import (
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
)

type PreCommitList []*PreCommit
type PreCommitNilList []bool

func (pcl PreCommitList) MakeNextHeight(secpSigner *crypto.Secp256k1Signer, bnSigner *crypto.BNGroupSigner) (*NextHeight, error) {
	propBytes, err := pcl[0].Proposal.MarshalBinary()
	if err != nil {
		return nil, err
	}
	prop := &Proposal{}
	err = prop.UnmarshalBinary(propBytes)
	if err != nil {
		return nil, err
	}
	sigs := [][]byte{}
	for _, pc := range pcl {
		s := utils.CopySlice(pc.Signature)
		sigs = append(sigs, s)
	}
	nh := &NextHeight{
		NHClaims: &NHClaims{
			Proposal: prop,
		},
		PreCommits: sigs,
	}
	err = nh.Sign(secpSigner, bnSigner)
	if err != nil {
		return nil, err
	}
	return nh, nil
}

func (pcl PreCommitList) GetProposal() (*Proposal, error) {
	propBytes, err := pcl[0].Proposal.MarshalBinary()
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
