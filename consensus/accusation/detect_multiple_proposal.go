package accusation

import "github.com/alicenet/alicenet/consensus/objs"

func detectMultipleProposal(rs *objs.RoundState) (objs.Accusation, bool) {
	// rs.Proposal and rs.ConflictingProposal should both not be nil
	// proposals must have same height and round: PLcainm -> rcert -> rclaims (better to check the whole rclaims equality) (also check all fields inside rclaims for validity - especially PrevBlock)
	// proposals must be valid prop.ValidateSignatures(crypto.Secp256k1Validator{}, crypto.BNGroupValidator{})
	// signatures are different it means double spend

	// proposer []byte
	// make sure proposer is a validator
	// submit both proposals and already validated that both RClaims are valid and sigs are different
	return nil, false
}

// assert detectMultipleProposal is of type detector
var _ detector = detectMultipleProposal
