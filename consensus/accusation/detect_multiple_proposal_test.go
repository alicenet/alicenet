package accusation

import (
	"strconv"
	"testing"

	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/crypto"
	"github.com/stretchr/testify/assert"
)

// TestMultipleProposalAccusation tests detection of multiple proposals
func TestMultipleProposalAccusation(t *testing.T) {
	lrs := &lstate.RoundStates{}
	lrs.OwnState = &objs.OwnState{
		MaxBHSeen: &objs.BlockHeader{},
	}
	//txHashes := [][][]byte{}
	txhash := crypto.Hasher([]byte(strconv.Itoa(1)))
	txHshLst := [][]byte{txhash}
	txRoot, err := objs.MakeTxRoot(txHshLst)
	assert.Nil(t, err)
	//txHashes = append(txHashes, txHshLst)
	bclaims := &objs.BClaims{
		ChainID:    1,
		Height:     1,
		TxCount:    1,
		PrevBlock:  crypto.Hasher([]byte("foo")),
		TxRoot:     txRoot,
		StateRoot:  crypto.Hasher([]byte("")),
		HeaderRoot: crypto.Hasher([]byte("")),
	}
	lrs.OwnState.MaxBHSeen.BClaims = bclaims
	vSetMap := make(map[string]bool)
	vSetMap["aaaaa"] = true
	lrs.ValidatorSet = &objs.ValidatorSet{
		Validators: []*objs.Validator{
			{
				VAddr: []byte("aaaaa"),
			},
		},
		ValidatorVAddrSet: vSetMap,
	}
	bh, err := lrs.OwnState.MaxBHSeen.BlockHash()
	assert.Nil(t, err)

	// round state
	rclaims := &objs.RClaims{
		ChainID:   1,
		Height:    2,
		Round:     1,
		PrevBlock: bh,
	}
	rs := &objs.RoundState{}
	proposal0 := &objs.Proposal{
		Proposer: []byte("aaaaa"),
		PClaims: &objs.PClaims{
			RCert: &objs.RCert{
				RClaims: rclaims,
			},
		},
	}
	proposal1 := &objs.Proposal{
		PClaims: &objs.PClaims{
			RCert: &objs.RCert{
				RClaims: rclaims,
			},
		},
	}
	rs.Proposal = proposal0
	rs.ConflictingProposal = proposal1

	acc, found := detectMultipleProposal(rs, lrs)
	assert.False(t, found)
	assert.Nil(t, acc)
}
