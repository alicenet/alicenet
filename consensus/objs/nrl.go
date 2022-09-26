package objs

import (
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
)

// NextRoundList ...
type NextRoundList []*NextRound

func (nrl NextRoundList) MakeRoundCert(bns *crypto.BNGroupSigner, groupShares [][]byte) (*RCert, error) {
	sigs := [][]byte{}
	for _, nr := range nrl {
		sigs = append(sigs, utils.CopySlice(nr.NRClaims.SigShare))
	}
	SigGroup, err := bns.Aggregate(sigs, groupShares)
	if err != nil {
		return nil, err
	}
	PrevBlock := utils.CopySlice(nrl[0].NRClaims.RClaims.PrevBlock)
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
