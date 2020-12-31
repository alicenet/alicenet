package objs

import (
	"github.com/MadBase/MadNet/crypto"
	gUtils "github.com/MadBase/MadNet/utils"
)

// NextRoundList ...
type NextRoundList []*NextRound

func (nrl NextRoundList) MakeRoundCert(bns *crypto.BNGroupSigner, groupShares [][]byte) (*RCert, error) {
	sigs := [][]byte{}
	for _, nr := range nrl {
		sigs = append(sigs, gUtils.CopySlice(nr.NRClaims.SigShare))
	}
	SigGroup, err := bns.Aggregate(sigs, groupShares)
	if err != nil {
		return nil, err
	}
	PrevBlock := gUtils.CopySlice(nrl[0].NRClaims.RClaims.PrevBlock)
	rc := &RCert{
		RClaims: &RClaims{
			ChainID:   nrl[0].NRClaims.RClaims.ChainID,
			Height:    nrl[0].NRClaims.RClaims.Height,
			Round:     nrl[0].NRClaims.RClaims.Round,
			PrevBlock: PrevBlock,
		},
		SigGroup: SigGroup,
	}
	return rc, nil
}
