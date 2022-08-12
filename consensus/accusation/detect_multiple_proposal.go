package accusation

import (
	"bytes"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/accusations"
	"github.com/alicenet/alicenet/utils"
)

func detectMultipleProposal(rs *objs.RoundState, lrs *lstate.RoundStates, db *db.Database) (tasks.Task, bool) {

	// rs.Proposal and rs.ConflictingProposal should both not be nil
	if rs.Proposal == nil || rs.ConflictingProposal == nil {
		return nil, false
	}

	if !bytes.Equal(rs.Proposal.Proposer, rs.ConflictingProposal.Proposer) {
		return nil, false
	}

	proposalPClaimsBin, err := rs.Proposal.PClaims.MarshalBinary()
	if err != nil {
		return nil, false
	}
	conflictingProposalPClaimsBin, err := rs.ConflictingProposal.PClaims.MarshalBinary()
	if err != nil {
		return nil, false
	}

	if bytes.Equal(proposalPClaimsBin, conflictingProposalPClaimsBin) {
		// we don't have multiple proposals, we can jump ship
		return nil, false
	}

	isValidator := lrs.ValidatorSet.IsVAddrValidator(rs.Proposal.Proposer)
	if !isValidator {
		return nil, false
	}

	// check if the proposal is being proposed by the correct validator
	if !lrs.IsProposerValid(rs.Proposal.Proposer) {
		return nil, false
	}

	// proposals must have same RClaims: PClaim -> rcert -> rclaims
	if !rs.Proposal.PClaims.RCert.RClaims.Equals(rs.ConflictingProposal.PClaims.RCert.RClaims) {
		return nil, false
	}

	// if signatures are different it means multiple proposals were sent
	err = rs.Proposal.ValidateSignatures(&crypto.Secp256k1Validator{}, &crypto.BNGroupValidator{})
	if err != nil {
		return nil, false
	}
	err = rs.ConflictingProposal.ValidateSignatures(&crypto.Secp256k1Validator{}, &crypto.BNGroupValidator{})
	if err != nil {
		return nil, false
	}

	/*
		to calculate the deterministic accusation ID based on signature sorting:

		convert sig0 and sig1 to uints and sort ASC
		if for example sig0 < sig1 then
		  id = keccak(sig0, pclaims0, sig1, pclaims1)
		no need to submit acc object with sig0 being the lowest sig and sig1 being the highest sig
	*/
	sig0 := uint256.Zero()
	err = sig0.UnmarshalBinary(crypto.Hasher(rs.Proposal.Signature))
	if err != nil {
		return nil, false
	}
	sig1 := uint256.Zero()
	err = sig1.UnmarshalBinary(crypto.Hasher(rs.ConflictingProposal.Signature))
	if err != nil {
		return nil, false
	}

	sig0Big, err := sig0.ToBigInt()
	if err != nil {
		return nil, false
	}

	sig1Big, err := sig1.ToBigInt()
	if err != nil {
		return nil, false
	}

	// submit both proposals and already validated that both RClaims are valid and sigs are different
	acc := accusations.NewMultipleProposalAccusationTask(
		rs.Proposal.Signature,
		rs.Proposal.PClaims,
		rs.ConflictingProposal.Signature,
		rs.ConflictingProposal.PClaims,
	)

	var id [32]byte

	// deterministic ID
	if sig0Big.Cmp(sig1Big) <= 0 {
		idBin := crypto.Hasher(
			rs.Proposal.Signature,
			proposalPClaimsBin,
			rs.ConflictingProposal.Signature,
			conflictingProposalPClaimsBin,
		)
		copy(id[:], idBin)
	} else {
		idBin := crypto.Hasher(
			rs.ConflictingProposal.Signature,
			conflictingProposalPClaimsBin,
			rs.Proposal.Signature,
			proposalPClaimsBin,
		)
		copy(id[:], idBin)
	}

	acc.Id = utils.Bytes32ToHex(id)

	return acc, true
}

// assert detectMultipleProposal is of type detector
var _ detector = detectMultipleProposal
