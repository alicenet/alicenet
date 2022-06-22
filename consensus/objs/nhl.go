package objs

import (
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
)

// NextHeightList ...
type NextHeightList []*NextHeight

func (nhl NextHeightList) MakeBlockHeader(bns *crypto.BNGroupSigner, groupShares [][]byte) (*BlockHeader, *RCert, error) {
	sigs := [][]byte{}
	for _, nh := range nhl {
		sigs = append(sigs, utils.CopySlice(nh.NHClaims.SigShare))
	}
	SigGroup, err := bns.Aggregate(sigs, groupShares)
	if err != nil {
		return nil, nil, err
	}
	bcBytes, err := nhl[0].NHClaims.Proposal.PClaims.BClaims.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}
	bclaims := &BClaims{}
	err = bclaims.UnmarshalBinary(bcBytes)
	if err != nil {
		return nil, nil, err
	}
	txHshlst := [][]byte{}
	for _, hsh := range nhl[0].NHClaims.Proposal.TxHshLst {
		chsh := utils.CopySlice(hsh)
		txHshlst = append(txHshlst, chsh)
	}

	bh := &BlockHeader{
		BClaims:  bclaims,
		TxHshLst: txHshlst,
		SigGroup: SigGroup,
	}
	PrevBlock, err := bclaims.BlockHash()
	if err != nil {
		return nil, nil, err
	}
	PrevBlockCopy := utils.CopySlice(PrevBlock)
	SigGroupCopy := utils.CopySlice(SigGroup)
	rc := &RCert{
		RClaims: &RClaims{
			ChainID:   bclaims.ChainID,
			Height:    bclaims.Height + 1,
			Round:     1,
			PrevBlock: PrevBlockCopy,
		},
		SigGroup: SigGroupCopy,
	}
	return bh, rc, nil
}
