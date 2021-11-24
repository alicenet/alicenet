package dkgtasks

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeTask stores the data required to dispute shares
type DisputeTask struct {
	OriginalRegistrationEnd uint64
	State                   *objects.DkgState
	Success                 bool
}

// NewDisputeTask creates a new task
func NewDisputeTask(state *objects.DkgState) *DisputeTask {
	return &DisputeTask{
		OriginalRegistrationEnd: state.RegistrationEnd, // If these quit being equal, this task should be abandoned
		State:                   state,
	}
}

// Initialize begins the setup phase for DisputeTask.
// It determines if the shares previously distributed are valid.
// If any are invalid, disputes will be issued.
func (t *DisputeTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	dkgState, validState := state.(*objects.DkgState)
	if !validState {
		panic(fmt.Errorf("%w invalid state type", objects.ErrCanNotContinue))
	}

	t.State = dkgState

	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("StateLocation", fmt.Sprintf("%p", t.State)).Info("Initialize()...")

	if !t.State.ShareDistribution {
		return fmt.Errorf("%w because share distribution not successful", objects.ErrCanNotContinue)
	}

	// Loop through all participants and check to see if shares are valid
	for idx := 0; idx < len(t.State.Participants); idx++ {
		participant := t.State.Participants[idx]
		logger.Infof("t.State.Index: %v\n", t.State.Index)
		logger.Infof("participant.Index: %v\n", participant.Index)
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
func (t *DisputeTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is subsequent attempts at disputing distributed shares
func (t *DisputeTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *DisputeTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

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
		// issuer issued bad information; this is participant.
		// disputer is disputing them; *you* are disputing the information.
		issuer := participant.Address
		// NOTE: The smart contract uses index-0 in this function call.
		//		 We are using index-1 when storing, so we need to subtract 1 here.
		issuerListIdx := big.NewInt(int64(participant.Index - 1))
		disputerListIdx := big.NewInt(int64(t.State.Index - 1))
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

		txn, err := eth.Contracts().Ethdkg().SubmitDispute(txnOpts, issuer, issuerListIdx, disputerListIdx, encryptedShares, commitments, sharedKey, sharedKeyProof)
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
func (t *DisputeTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	// This wraps the retry logic for every phase, _except_ registration
	return GeneralTaskShouldRetry(ctx, t.State.Account, logger, eth,
		t.State.TransportPublicKey, t.OriginalRegistrationEnd, t.State.DisputeEnd)
}

// DoDone creates a log entry saying task is complete
func (t *DisputeTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Info("done")

	t.State.Dispute = t.Success
}
