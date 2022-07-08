package objs

type MultipleProposalAccusation struct {
	BaseAccusation
	Signature0 []byte
	Proposal0  *PClaims
	Signature1 []byte
	Proposal1  *PClaims
}

// assert MultipleProposalAccusation implements Accusation
var _ Accusation = &MultipleProposalAccusation{}
