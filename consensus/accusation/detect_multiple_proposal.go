package accusation

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/accusations"
	"github.com/ethereum/go-ethereum/common"
)

func detectMultipleProposal(rs *objs.RoundState, lrs *lstate.RoundStates, consDB *db.Database) (tasks.Task, bool) {
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

	// submit both proposals and already validated that both RClaims are valid and sigs are different
	acc := accusations.NewMultipleProposalAccusationTask(
		rs.Proposal.Signature,
		rs.Proposal.PClaims,
		rs.ConflictingProposal.Signature,
		rs.ConflictingProposal.PClaims,
		rs.Proposal.Proposer,
	)

	// deterministic accusation ID
	var chainID []byte = common.LeftPadBytes(big.NewInt(int64(rs.Proposal.PClaims.RCert.RClaims.ChainID)).Bytes(), 4)
	var height []byte = common.LeftPadBytes(big.NewInt(int64(rs.Proposal.PClaims.RCert.RClaims.Height)).Bytes(), 4)
	var round []byte = common.LeftPadBytes(big.NewInt(int64(rs.Proposal.PClaims.RCert.RClaims.Round)).Bytes(), 4)
	var preSalt []byte = crypto.Hasher([]byte("AccusationMultipleProposal"))

	var id []byte = crypto.Hasher(
		rs.Proposal.Proposer,
		chainID,
		height,
		round,
		preSalt,
	)
	acc.ID = hex.EncodeToString(id)

	return acc, true
}

// assert detectMultipleProposal is of type detector
var _ detector = detectMultipleProposal
