package dkgtasks

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/alicenet/alicenet/blockchain/dkg"
	"github.com/alicenet/alicenet/blockchain/dkg/math"
	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/alicenet/alicenet/crypto/bn256"
	"github.com/alicenet/alicenet/crypto/bn256/cloudflare"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// DisputeShareDistributionTask stores the data required to dispute shares
type DisputeShareDistributionTask struct {
	*ExecutionData
}

// asserting that DisputeShareDistributionTask struct implements interface interfaces.Task
var _ interfaces.Task = &DisputeShareDistributionTask{}

// NewDisputeShareDistributionTask creates a new task
func NewDisputeShareDistributionTask(state *objects.DkgState, start uint64, end uint64) *DisputeShareDistributionTask {
	return &DisputeShareDistributionTask{
		ExecutionData: NewExecutionData(state, start, end),
	}
}

// Initialize begins the setup phase for DisputeShareDistributionTask.
// It determines if the shares previously distributed are valid.
// If any are invalid, disputes will be issued.
func (t *DisputeShareDistributionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {

	logger.Info("DisputeShareDistributionTask Initialize()")

	dkgData, ok := state.(objects.ETHDKGTaskData)
	if !ok {
		return objects.ErrCanNotContinue
	}

	unlock := dkgData.LockState()
	defer unlock()
	if dkgData.State != t.State {
		t.State = dkgData.State
	}

	if t.State.Phase != objects.DisputeShareDistribution && t.State.Phase != objects.ShareDistribution {
		return fmt.Errorf("%w because it's not DisputeShareDistribution phase", objects.ErrCanNotContinue)
	}

	var participantsList = t.State.GetSortedParticipants()
	// Loop through all participants and check to see if shares are valid
	for idx := 0; idx < t.State.NumberOfValidators; idx++ {
		participant := participantsList[idx]

		var emptyHash [32]byte
		if participant.DistributedSharesHash == emptyHash {
			continue
		}

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

	callOpts := eth.GetCallOpts(ctx, t.State.Account)

	txnOpts, err := eth.GetTransactionOpts(ctx, t.State.Account)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// If the TxOpts exists, meaning the Tx replacement timeout was reached,
	// we increase the Gas to have priority for the next blocks
	if t.TxOpts != nil && t.TxOpts.Nonce != nil {
		logger.Info("txnOpts Replaced")
		txnOpts.Nonce = t.TxOpts.Nonce
		txnOpts.GasFeeCap = t.TxOpts.GasFeeCap
		txnOpts.GasTipCap = t.TxOpts.GasTipCap
	}

	for _, participant := range t.State.BadShares {

		isValidator, err := eth.Contracts().ValidatorPool().IsValidator(callOpts, participant.Address)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "getting isValidator failed: %v", err)
		}

		if !isValidator {
			continue
		}

		dishonestAddress := participant.Address
		encryptedShares := t.State.Participants[participant.Address].EncryptedShares
		commitments := t.State.Participants[participant.Address].Commitments

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

		// Accuse participant
		txn, err := eth.Contracts().Ethdkg().AccuseParticipantDistributedBadShares(txnOpts, dishonestAddress, encryptedShares, commitments, sharedKey, sharedKeyProof)
		if err != nil {
			return dkg.LogReturnErrorf(logger, "submit share dispute failed: %v", err)
		}
		t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
		t.TxOpts.GasFeeCap = txn.GasFeeCap()
		t.TxOpts.GasTipCap = txn.GasTipCap()
		t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

		logger.WithFields(logrus.Fields{
			"GasFeeCap": t.TxOpts.GasFeeCap,
			"GasTipCap": t.TxOpts.GasTipCap,
			"Nonce":     t.TxOpts.Nonce,
		}).Info("bad share dispute fees")

		// Queue transaction
		eth.Queue().QueueTransaction(ctx, txn)
	}

	t.Success = true
	return nil
}

// ShouldRetry checks if it makes sense to try again
// if the DKG process is in the right phase and blocks
// range and there still someone to accuse, the retry
// is executed
func (t *DisputeShareDistributionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("DisputeShareDistributionTask ShouldRetry()")

	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	if t.State.Phase != objects.DisputeShareDistribution {
		return false
	}

	callOpts := eth.GetCallOpts(ctx, t.State.Account)
	badParticipants, err := eth.Contracts().Ethdkg().GetBadParticipants(callOpts)
	if err != nil {
		logger.Error("could not get BadParticipants")
	}

	logger.WithFields(logrus.Fields{
		"state.BadShares":     len(t.State.BadShares),
		"eth.badParticipants": badParticipants,
	}).Debug("DisputeShareDistributionTask ShouldRetry2()")

	// if there is someone that wasn't accused we need to retry
	return len(t.State.BadShares) != int(badParticipants.Int64())
}

// DoDone creates a log entry saying task is complete
func (t *DisputeShareDistributionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Info("DisputeShareDistributionTask done")
}

func (t *DisputeShareDistributionTask) GetExecutionData() interface{} {
	return t.ExecutionData
}
