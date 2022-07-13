package accusation

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"
)

func detectMultipleProposal(rs *objs.RoundState, lrs *lstate.RoundStates) (objs.Accusation, bool) {

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

	// make sure proposer is a validator
	isValidator := lrs.ValidatorSet.IsVAddrValidator(rs.Proposal.Proposer)
	if !isValidator {
		return nil, false
	}

	// check if the proposal is being proposed by the correct validator
	if !lrs.IsProposerValid(rs.Proposal.Proposer) {
		return nil, false
	}

	os := lrs.OwnState
	prevBlock, err := os.MaxBHSeen.BlockHash()
	if err != nil {
		panic(fmt.Sprintf("detectMultipleProposal could not get os.MaxBHSeen.BlockHash: %v", err))
	}

	logging.GetLogger("accusations").WithFields(logrus.Fields{
		"os.height":          os.MaxBHSeen.BClaims.Height,
		"os.chain":           os.MaxBHSeen.BClaims.ChainID,
		"os.blockHash":       prevBlock,
		"proposal.round":     rs.Proposal.PClaims.RCert.RClaims.Round,
		"proposal.height":    rs.Proposal.PClaims.RCert.RClaims.Height,
		"proposal.chain":     rs.Proposal.PClaims.RCert.RClaims.ChainID,
		"proposal.prevBlock": rs.Proposal.PClaims.RCert.RClaims.PrevBlock,
	}).Debug("detectMultipleProposal")

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

	logging.GetLogger("accusations").WithFields(logrus.Fields{
		"sig0":      sig0,
		"sig0.3":    sig0Big.String(),
		"prop0.sig": hex.EncodeToString(rs.Proposal.Signature),
		"sig1":      sig1,
		"sig1.3":    sig1Big.String(),
		"prop1.sig": hex.EncodeToString(rs.ConflictingProposal.Signature),
	}).Warn("sigs")

	// submit both proposals and already validated that both RClaims are valid and sigs are different
	acc := &objs.MultipleProposalAccusation{
		Signature0: rs.Proposal.Signature,
		Proposal0:  rs.Proposal.PClaims,
		Signature1: rs.ConflictingProposal.Signature,
		Proposal1:  rs.ConflictingProposal.PClaims,
	}

	if sig0Big.Cmp(sig1Big) <= 0 {
		copy(
			acc.ID[:],
			crypto.Hasher(
				rs.Proposal.Signature,
				proposalPClaimsBin,
				rs.ConflictingProposal.Signature,
				conflictingProposalPClaimsBin,
			),
		)
	} else {
		copy(
			acc.ID[:],
			crypto.Hasher(
				rs.ConflictingProposal.Signature,
				conflictingProposalPClaimsBin,
				rs.Proposal.Signature,
				proposalPClaimsBin,
			),
		)
	}

	// todo: form Accusation task here

	return acc, true
}

// assert detectMultipleProposal is of type detector
var _ detector = detectMultipleProposal
