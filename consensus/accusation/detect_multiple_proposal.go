package accusation

import "github.com/alicenet/alicenet/consensus/objs"

func detectMultipleProposal(rs *objs.RoundState) (objs.Accusation, bool) {
	// rs.Proposal and rs.ConflictingProposal should both not be nil
	if rs.Proposal == nil || rs.ConflictingProposal == nil {
		return nil, false
	}

	// proposals must have same height and round: PClaim -> rcert -> rclaims (better to check the whole rclaims equality) (also check all fields inside rclaims for validity - especially PrevBlock)
	if rs.Proposal.PClaims.RCert.RClaims.Height != rs.ConflictingProposal.PClaims.BClaims.Height || rs.Proposal.PClaims.RCert.RClaims.Round != rs.ConflictingProposal.PClaims.RCert.RClaims.Round {
		return nil, false
	}

	// proposals must be valid prop.ValidateSignatures(crypto.Secp256k1Validator{}, crypto.BNGroupValidator{})
	// signatures are different it means double spend

	// proposer []byte
	// make sure proposer is a validator
	// submit both proposals and already validated that both RClaims are valid and sigs are different
	return nil, false
}

// assert detectMultipleProposal is of type detector
var _ detector = detectMultipleProposal
