package accusation_test

import (
	"testing"

	"github.com/MadBase/MadNet/consensus/objs"
)

func TestDoubleProposal(t *testing.T) {

	voo := &objs.RoundState{
		VAddr: []byte{},
		Proposal: &objs.Proposal{
			Signature: []byte{},
			PClaims: &objs.PClaims{
				RCert: &objs.RCert{
					SigGroup: []byte{},
					RClaims: &objs.RClaims{
						ChainID:   42,
						Height:    73,
						Round:     2,
						PrevBlock: []byte{},
					},
				},
				BClaims: &objs.BClaims{
					ChainID:   42,
					Height:    73,
					PrevBlock: []byte{},
				},
			},
		},
	}

	t.Logf("voo:%v", voo)

	p1 := voo.Proposal
	p2 := voo.ConflictingProposal

	// p1.PClaims.RCert.RClaims.Height

	t.Logf("p1:%v p2:%v", p1, p2)
}
