package dkg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	exConstants "github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
	exUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/utils"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/sirupsen/logrus"
)

// MPKSubmissionTask stores the data required to submit the mpk
type MPKSubmissionTask struct {
	*objects.Task
}

// asserting that MPKSubmissionTask struct implements interface interfaces.Task
var _ interfaces.ITask = &MPKSubmissionTask{}

// NewMPKSubmissionTask creates a new task
func NewMPKSubmissionTask(dkgState *state.DkgState, start uint64, end uint64) *MPKSubmissionTask {
	return &MPKSubmissionTask{
		Task: objects.NewTask(dkgState, exConstants.MPKSubmissionTaskName, start, end),
	}
}

// Initialize prepares for work to be done in MPKSubmission phase.
// Here we load all key shares and construct the master public key
// to submit in DoWork.
func (t *MPKSubmissionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) error {

	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("MPKSubmissionTask Initialize()...")

	taskState, ok := t.State.(*state.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	if taskState.Phase != state.MPKSubmission {
		return fmt.Errorf("%w because it's not in MPKSubmission phase", objects.ErrCanNotContinue)
	}

	// compute MPK if not yet computed
	if taskState.MasterPublicKey[0] == nil ||
		taskState.MasterPublicKey[1] == nil ||
		taskState.MasterPublicKey[2] == nil ||
		taskState.MasterPublicKey[3] == nil ||
		(taskState.MasterPublicKey[0].Cmp(big.NewInt(0)) == 0 &&
			taskState.MasterPublicKey[1].Cmp(big.NewInt(0)) == 0 &&
			taskState.MasterPublicKey[2].Cmp(big.NewInt(0)) == 0 &&
			taskState.MasterPublicKey[3].Cmp(big.NewInt(0)) == 0) {

		// setup leader election
		block, err := eth.GetClient().BlockByNumber(ctx, big.NewInt(int64(t.Start)))
		if err != nil {
			return fmt.Errorf("MPKSubmissionTask Initialize(): error getting block by number: %v", err)
		}

		logger.Infof("block hash: %v\n", block.Hash())
		t.StartBlockHash.SetBytes(block.Hash().Bytes())

		// prepare MPK
		g1KeyShares := make([][2]*big.Int, taskState.NumberOfValidators)
		g2KeyShares := make([][4]*big.Int, taskState.NumberOfValidators)

		var participantsList = taskState.GetSortedParticipants()
		validMPK := true
		for idx, participant := range participantsList {
			// Bringing these in from state but could directly query contract
			g1KeyShares[idx] = taskState.Participants[participant.Address].KeyShareG1s
			g2KeyShares[idx] = taskState.Participants[participant.Address].KeyShareG2s

			logger.Debugf("INIT idx:%v pidx:%v address:%v g1:%v g2:%v", idx, participant.Index, participant.Address.Hex(), g1KeyShares[idx], g2KeyShares[idx])

			for i := range g1KeyShares[idx] {
				if g1KeyShares[idx][i] == nil {
					logger.Errorf("Missing g1Keyshare[%v][%v] for %v.", idx, i, participant.Address.Hex())
					validMPK = false
				}
			}

			for i := range g2KeyShares[idx] {
				if g2KeyShares[idx][i] == nil {
					logger.Errorf("Missing g2Keyshare[%v][%v] for %v.", idx, i, participant.Address.Hex())
					validMPK = false
				}
			}
		}

		logger.Infof("# Participants: %v\n", len(taskState.Participants))

		mpk, err := state.GenerateMasterPublicKey(g1KeyShares, g2KeyShares)
		if err != nil && validMPK {
			return utils.LogReturnErrorf(logger, "Failed to generate master public key:%v", err)
		}

		if !validMPK {
			mpk = [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
		}

		// Master public key is all we generate here so save it
		taskState.MasterPublicKey = mpk
	} else {
		logger.Infof("MPKSubmissionTask Initialize(): mpk already defined")
	}

	return nil
}

// DoWork is the first attempt at submitting the mpk
func (t *MPKSubmissionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at submitting the mpk
func (t *MPKSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) error {
	return t.doTask(ctx, logger, eth)
}

func (t *MPKSubmissionTask) doTask(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) error {
	t.State.Lock()
	defer t.State.Unlock()

	taskState, ok := t.State.(*state.DkgState)
	if !ok {
		return objects.ErrCanNotContinue
	}

	logger.Info("MPKSubmissionTask doTask()")

	if !t.shouldSubmitMPK(ctx, eth, logger) {
		t.Success = true
		return nil
	}

	// submit if I'm a leader for this task
	if !t.AmILeading(ctx, eth, logger, taskState) {
		return errors.New("not leading MPK submission yet")
	}

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, taskState.Account)
	if err != nil {
		return utils.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// If the TxOpts exists, meaning the Tx replacement timeout was reached,
	// we increase the Gas to have priority for the next blocks
	if t.TxOpts != nil && t.TxOpts.Nonce != nil {
		logger.Info("txnOpts Replaced")
		txnOpts.Nonce = t.TxOpts.Nonce
		txnOpts.GasFeeCap = t.TxOpts.GasFeeCap
		txnOpts.GasTipCap = t.TxOpts.GasTipCap
	}

	// Submit MPK
	logger.Infof("submitting master public key:%v", taskState.MasterPublicKey)
	txn, err := eth.Contracts().Ethdkg().SubmitMasterPublicKey(txnOpts, taskState.MasterPublicKey)
	if err != nil {
		return utils.LogReturnErrorf(logger, "submitting master public key failed: %v", err)
	}
	t.TxOpts.TxHashes = append(t.TxOpts.TxHashes, txn.Hash())
	t.TxOpts.GasFeeCap = txn.GasFeeCap()
	t.TxOpts.GasTipCap = txn.GasTipCap()
	t.TxOpts.Nonce = big.NewInt(int64(txn.Nonce()))

	logger.WithFields(logrus.Fields{
		"GasFeeCap": t.TxOpts.GasFeeCap,
		"GasTipCap": t.TxOpts.GasTipCap,
		"Nonce":     t.TxOpts.Nonce,
	}).Info("MPK submission fees")

	// Queue transaction
	watcher := transaction.WatcherFromNetwork(eth)
	watcher.Subscribe(ctx, txn)
	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
func (t *MPKSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth ethereum.Network) bool {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("MPKSubmissionTask ShouldRetry()")

	generalRetry := exUtils.GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	taskState, ok := t.State.(*state.DkgState)
	if !ok {
		logger.Error("Invalid convertion of taskState object")
		return false
	}

	if taskState.Phase != state.MPKSubmission {
		return false
	}

	return t.shouldSubmitMPK(ctx, eth, logger)
}

// DoDone creates a log entry saying task is complete
func (t *MPKSubmissionTask) DoDone(logger *logrus.Entry) {
	t.State.Lock()
	defer t.State.Unlock()

	logger.WithField("Success", t.Success).Infof("MPKSubmissionTask done")
}

func (t *MPKSubmissionTask) GetExecutionData() interfaces.ITaskExecutionData {
	return t.Task
}

func (t *MPKSubmissionTask) shouldSubmitMPK(ctx context.Context, eth ethereum.Network, logger *logrus.Entry) bool {

	taskState, ok := t.State.(*state.DkgState)
	if !ok {
		logger.Error("MPKSubmissionTask ShouldRetry() invalid convertion of taskState object")
		return true
	}

	if taskState.MasterPublicKey[0].Cmp(big.NewInt(0)) == 0 &&
		taskState.MasterPublicKey[1].Cmp(big.NewInt(0)) == 0 &&
		taskState.MasterPublicKey[2].Cmp(big.NewInt(0)) == 0 &&
		taskState.MasterPublicKey[3].Cmp(big.NewInt(0)) == 0 {
		return false
	}

	callOpts, err := eth.GetCallOpts(ctx, taskState.Account)
	if err != nil {
		logger.Error(fmt.Sprintf("MPKSubmissionTask ShouldRetry() failed getting call options: %v", err))
		return true
	}

	mpkHash, err := eth.Contracts().Ethdkg().GetMasterPublicKeyHash(callOpts)
	if err != nil {
		return true
	}

	logger.WithField("Method", "shouldSubmitMPK").Debugf("mpkHash received")

	mpkHashBin, err := bn256.MarshalBigIntSlice(taskState.MasterPublicKey[:])
	if err != nil {
		return true
	}
	mpkHashSlice := crypto.Hasher(mpkHashBin)

	if bytes.Equal(mpkHash[:], mpkHashSlice) {
		logger.WithField("Method", "shouldSubmitMPK").Debugf("state mpkHash is different from the received")
		return false
	}

	logger.WithField("Method", "shouldSubmitMPK").Debugf("state mpkHash is equal to the received")
	return true
}

func (t *MPKSubmissionTask) AmILeading(ctx context.Context, eth ethereum.Network, logger *logrus.Entry, taskState *state.DkgState) bool {
	// check if I'm a leader for this task
	currentHeight, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return false
	}

	blocksSinceDesperation := int(currentHeight) - int(t.Start) - constants.ETHDKGDesperationDelay
	amILeading := utils.AmILeading(taskState.NumberOfValidators, taskState.Index-1, blocksSinceDesperation, t.StartBlockHash.Bytes(), logger)

	logger.WithFields(logrus.Fields{
		"currentHeight":                    currentHeight,
		"t.Start":                          t.Start,
		"constants.ETHDKGDesperationDelay": constants.ETHDKGDesperationDelay,
		"blocksSinceDesperation":           blocksSinceDesperation,
		"amILeading":                       amILeading,
	}).Infof("dkg.AmILeading")

	return amILeading
}
