package accusation

import (
	"bytes"
	"fmt"

	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"
)

func detectMultipleProposal(rs *objs.RoundState, lrs *lstate.RoundStates) (objs.Accusation, bool) {

	if rs.Proposal == nil {
		return nil, false
	}

	os := lrs.OwnState

	prevBlock, err := os.MaxBHSeen.BlockHash()
	if err != nil {
		panic(fmt.Sprintf("detectMultipleProposal could not get os.MaxBHSeen.BlockHash: %v", err))
	}

	logging.GetLogger("accusations").WithFields(logrus.Fields{
		//"os.round":           os.MaxBHSeen.BClaims.,
		"os.height":          os.MaxBHSeen.BClaims.Height,
		"os.chain":           os.MaxBHSeen.BClaims.ChainID,
		"os.blockHash":       prevBlock,
		"proposal.round":     rs.Proposal.PClaims.RCert.RClaims.Round,
		"proposal.height":    rs.Proposal.PClaims.RCert.RClaims.Height,
		"proposal.chain":     rs.Proposal.PClaims.RCert.RClaims.ChainID,
		"proposal.prevBlock": rs.Proposal.PClaims.RCert.RClaims.PrevBlock,
	}).Debug("detectMultipleProposal")

	// rs.Proposal and rs.ConflictingProposal should both not be nil
	if rs.Proposal == nil || rs.ConflictingProposal == nil {
		return nil, false
	}

	// proposals must have same RClaims: PClaim -> rcert -> rclaims
	if !rs.Proposal.PClaims.RCert.RClaims.Equals(rs.ConflictingProposal.PClaims.RCert.RClaims) {
		return nil, false
	}

	// checking all fields inside rclaims for validity
	if rs.Proposal.PClaims.RCert.RClaims.ChainID != os.MaxBHSeen.BClaims.ChainID ||
		rs.Proposal.PClaims.RCert.RClaims.Height-1 != os.MaxBHSeen.BClaims.Height ||
		!bytes.Equal(rs.Proposal.PClaims.RCert.RClaims.PrevBlock, prevBlock) {
		return nil, false
	}

	// proposals must be valid
	// if signatures are different it means multiple proposals were sent
	err = rs.Proposal.ValidateSignatures(&crypto.Secp256k1Validator{}, &crypto.BNGroupValidator{})
	if err == nil {
		// signatures are equal, nothing else to do here
		return nil, false
	}

	// make sure proposer is a validator
	isValidator := lrs.ValidatorSet.IsVAddrValidator(rs.Proposal.Proposer)

	if !isValidator {
		return nil, false
	}

	// submit both proposals and already validated that both RClaims are valid and sigs are different
	acc := &objs.MultipleProposalAccusation{
		Signature0: rs.Proposal.Signature,
		Proposal0:  rs.Proposal.PClaims,
		Signature1: rs.ConflictingProposal.Signature,
		Proposal1:  rs.ConflictingProposal.PClaims,
	}

	// todo: form Accusation task here

	return acc, true
}

// assert detectMultipleProposal is of type detector
var _ detector = detectMultipleProposal
