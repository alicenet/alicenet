package dkgtasks

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeShareDistributionTask stores the data required to dispute shares
type DisputeShareDistributionTask struct {
	State   *objects.DkgState
	Success bool
}

// NewDisputeShareDistributionTask creates a new task
func NewDisputeShareDistributionTask(state *objects.DkgState) *DisputeShareDistributionTask {
	return &DisputeShareDistributionTask{
		State: state,
	}
}

// Initialize begins the setup phase for DisputeShareDistributionTask.
// It determines if the shares previously distributed are valid.
// If any are invalid, disputes will be issued.
func (t *DisputeShareDistributionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	dkgState, validState := state.(*objects.DkgState)
	if !validState {
		panic(fmt.Errorf("%w invalid state type", objects.ErrCanNotContinue))
	}

	t.State = dkgState

	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("StateLocation", fmt.Sprintf("%p", t.State)).Info("DisputeShareDistributionTask Initialize()")

	if t.State.Phase != objects.DisputeShareDistribution {
		return fmt.Errorf("%w because it's not DisputeShareDistribution phase", objects.ErrCanNotContinue)
	}

	// Loop through all participants and check to see if shares are valid
	for idx := 0; idx < len(t.State.Participants); idx++ {
		participant := t.State.Participants[idx]
		//logger.Infof("t.State.Index: %v\n", t.State.Index)
		logger.Infof("participant idx: %v:%v:%v\n", idx, participant.Index, t.State.Index)
		valid, present, err := math.VerifyDistributedShares(t.State, participant)
		if err != nil {
			// A major error occured; we cannot continue
			logger.Errorf("VerifyDistributedShares broke; Participant Address: %v", participant.Address.Hex())
			return fmt.Errorf("VerifyDistributedShares broke: %v; %v", err.Error(), objects.ErrCanNotContinue)
		}
		if !present {
			logger.Warningf("No share from %v", participant.Address.Hex())
			continue
		}
		if !valid {
			logger.Warningf("Invalid share from %v", participant.Address.Hex())
			t.State.BadShares[participant.Address] = participant
		}
	}

	return nil
}

// DoWork is the first attempt at disputing distributed shares
func (t *DisputeShareDistributionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is subsequent attempts at disputing distributed shares
func (t *DisputeShareDistributionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *DisputeShareDistributionTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeShareDistributionTask doTask()")

	for _, participant := range t.State.BadShares {
		/*
			function submit_dispute(
				address issuer,
				uint256 issuer_list_idx,
				uint256 disputer_list_idx,
				uint256[] memory encrypted_shares,
				uint256[2][] memory commitments,
				uint256[2] memory shared_key,
				uint256[2] memory shared_key_correctness_proof
			) */
		// eth.Contracts().Ethdkg().SubmitDispute()

		// Initial information;
		// dishonestAddress issued bad information; this is participant.
		// disputer is disputing them; *you* are disputing the information.
		dishonestAddress := participant.Address
		encryptedShares := t.State.EncryptedShares[participant.Address]
		commitments := t.State.Commitments[participant.Address]

		// Construct shared key
		disputePublicKeyG1, err := bn256.BigIntArrayToG1(participant.PublicKey)
		if err != nil {
			return err
		}
		sharedKeyG1 := cloudflare.GenerateSharedSecretG1(t.State.TransportPrivateKey, disputePublicKeyG1)
		sharedKey, err := bn256.G1ToBigIntArray(sharedKeyG1)
		if err != nil {
			return err
		}

		// Construct shared key proof
		g1Base := new(cloudflare.G1).ScalarBaseMult(common.Big1)
		transportPublicKeyG1 := new(cloudflare.G1).ScalarBaseMult(t.State.TransportPrivateKey)
		sharedKeyProof, err := cloudflare.GenerateDLEQProofG1(
			g1Base, transportPublicKeyG1, disputePublicKeyG1, sharedKeyG1, t.State.TransportPrivateKey, rand.Reader)
		if err != nil {
			return err
		}

		txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
		}

		txn, err := eth.Contracts().Ethdkg().AccuseParticipantDistributedBadShares(txnOpts, dishonestAddress, encryptedShares, commitments, sharedKey, sharedKeyProof)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "submit share dispute failed: %v", err)
		}

		// Waiting for receipt
		receipt, err := eth.Queue().QueueAndWait(ctx, txn)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "waiting for receipt failed: %v", err)
		}
		if receipt == nil {
			return dkg.LogReturnErrorf(logger, "missing share dispute receipt")
		}

		// Check receipt to confirm we were successful
		if receipt.Status != uint64(1) {
			return dkg.LogReturnErrorf(logger, "share dispute status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
		}
	}

	t.Success = true
	return nil
}

// ShouldRetry checks if it makes sense to try again
func (t *DisputeShareDistributionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeShareDistributionTask ShouldRetry()")
	callOpts := eth.GetCallOpts(ctx, t.State.Account)
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	if err != nil {
		logger.Error("could not get BadParticipants")
	}

	logger.WithFields(logrus.Fields{
		"state.BadShares":     len(t.State.BadShares),
		"eth.badParticipants": badParticipants,
	}).Info("DisputeShareDistributionTask ShouldRetry2()")

	return len(t.State.BadShares) != int(badParticipants.Int64())

	// This wraps the retry logic for every phase, _except_ registration
	//return GeneralTaskShouldRetry(ctx, t.State.Account, logger, eth,
	//	t.State.TransportPublicKey, t.OriginalRegistrationEnd, t.State.DisputeShareDistributionEnd)
}

// DoDone creates a log entry saying task is complete
func (t *DisputeShareDistributionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Info("DisputeShareDistributionTask done")
}
