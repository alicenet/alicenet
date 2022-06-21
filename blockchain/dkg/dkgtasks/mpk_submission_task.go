package dkgtasks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/crypto/bn256"
	"math/big"

	"github.com/alicenet/alicenet/blockchain/dkg"
	"github.com/alicenet/alicenet/blockchain/dkg/math"
	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/alicenet/alicenet/constants"
	"github.com/sirupsen/logrus"
)

// MPKSubmissionTask stores the data required to submit the mpk
type MPKSubmissionTask struct {
	*ExecutionData
}

// asserting that MPKSubmissionTask struct implements interface interfaces.Task
var _ interfaces.Task = &MPKSubmissionTask{}

// NewMPKSubmissionTask creates a new task
func NewMPKSubmissionTask(state *objects.DkgState, start uint64, end uint64) *MPKSubmissionTask {
	return &MPKSubmissionTask{
		ExecutionData: NewExecutionData(state, start, end),
	}
}

// Initialize prepares for work to be done in MPKSubmission phase.
// Here we load all key shares and construct the master public key
// to submit in DoWork.
func (t *MPKSubmissionTask) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {

	logger.Info("MPKSubmissionTask Initialize()...")

	dkgData, ok := state.(objects.ETHDKGTaskData)
	if !ok {
		return objects.ErrCanNotContinue
	}

	unlock := dkgData.LockState()
	defer unlock()
	if dkgData.State != t.State {
		t.State = dkgData.State
	}

	if t.State.Phase != objects.MPKSubmission {
		return fmt.Errorf("%w because it's not in MPKSubmission phase", objects.ErrCanNotContinue)
	}

	// compute MPK if not yet computed
	if t.State.MasterPublicKey[0] == nil ||
		t.State.MasterPublicKey[1] == nil ||
		t.State.MasterPublicKey[2] == nil ||
		t.State.MasterPublicKey[3] == nil ||
		(t.State.MasterPublicKey[0].Cmp(big.NewInt(0)) == 0 &&
			t.State.MasterPublicKey[1].Cmp(big.NewInt(0)) == 0 &&
			t.State.MasterPublicKey[2].Cmp(big.NewInt(0)) == 0 &&
			t.State.MasterPublicKey[3].Cmp(big.NewInt(0)) == 0) {

		// setup leader election
		block, err := eth.GetGethClient().BlockByNumber(ctx, big.NewInt(int64(t.Start)))
		if err != nil {
			return fmt.Errorf("MPKSubmissionTask Initialize(): error getting block by number: %v", err)
		}

		logger.Infof("block hash: %v\n", block.Hash())
		t.StartBlockHash.SetBytes(block.Hash().Bytes())

		// prepare MPK
		g1KeyShares := make([][2]*big.Int, t.State.NumberOfValidators)
		g2KeyShares := make([][4]*big.Int, t.State.NumberOfValidators)

		var participantsList = t.State.GetSortedParticipants()
		validMPK := true
		for idx, participant := range participantsList {
			// Bringing these in from state but could directly query contract
			g1KeyShares[idx] = t.State.Participants[participant.Address].KeyShareG1s
			g2KeyShares[idx] = t.State.Participants[participant.Address].KeyShareG2s

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

		logger.Infof("# Participants: %v\n", len(t.State.Participants))

		mpk, err := math.GenerateMasterPublicKey(g1KeyShares, g2KeyShares)
		if err != nil && validMPK {
			return dkg.LogReturnErrorf(logger, "Failed to generate master public key:%v", err)
		}

		if !validMPK {
			mpk = [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
		}

		// Master public key is all we generate here so save it
		t.State.MasterPublicKey = mpk

		unlock()
		dkgData.PersistStateCB()
	} else {
		logger.Infof("MPKSubmissionTask Initialize(): mpk already defined")
	}

	return nil
}

// DoWork is the first attempt at submitting the mpk
func (t *MPKSubmissionTask) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

// DoRetry is all subsequent attempts at submitting the mpk
func (t *MPKSubmissionTask) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	return t.doTask(ctx, logger, eth)
}

func (t *MPKSubmissionTask) doTask(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("MPKSubmissionTask doTask()")

	if !t.shouldSubmitMPK(ctx, eth, logger) {
		t.Success = true
		return nil
	}

	// submit if I'm a leader for this task
	if !t.AmILeading(ctx, eth, logger) {
		return errors.New("not leading MPK submission yet")
	}

	// Setup
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

	// Submit MPK
	logger.Infof("submitting master public key:%v", t.State.MasterPublicKey)
	txn, err := eth.Contracts().Ethdkg().SubmitMasterPublicKey(txnOpts, t.State.MasterPublicKey)
	if err != nil {
		return dkg.LogReturnErrorf(logger, "submitting master public key failed: %v", err)
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
	eth.Queue().QueueTransaction(ctx, txn)
	t.Success = true

	return nil
}

// ShouldRetry checks if it makes sense to try again
// Predicates:
// -- we haven't passed the last block
func (t *MPKSubmissionTask) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	t.State.Lock()
	defer t.State.Unlock()

	logger.Info("MPKSubmissionTask ShouldRetry()")

	generalRetry := GeneralTaskShouldRetry(ctx, logger, eth, t.Start, t.End)
	if !generalRetry {
		return false
	}

	if t.State.Phase != objects.MPKSubmission {
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

func (t *MPKSubmissionTask) GetExecutionData() interface{} {
	return t.ExecutionData
}

func (t *MPKSubmissionTask) shouldSubmitMPK(ctx context.Context, eth interfaces.Ethereum, logger *logrus.Entry) bool {
	if t.State.MasterPublicKey[0].Cmp(big.NewInt(0)) == 0 &&
		t.State.MasterPublicKey[1].Cmp(big.NewInt(0)) == 0 &&
		t.State.MasterPublicKey[2].Cmp(big.NewInt(0)) == 0 &&
		t.State.MasterPublicKey[3].Cmp(big.NewInt(0)) == 0 {
		return false
	}

	callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())

	mpkHash, err := eth.Contracts().Ethdkg().GetMasterPublicKeyHash(callOpts)
	if err != nil {
		return true
	}

	logger.WithField("Method", "shouldSubmitMPK").Debugf("mpkHash received")

	mpkHashBin, err := bn256.MarshalBigIntSlice(t.State.MasterPublicKey[:])
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

func (t *MPKSubmissionTask) AmILeading(ctx context.Context, eth interfaces.Ethereum, logger *logrus.Entry) bool {
	// check if I'm a leader for this task
	currentHeight, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		return false
	}

	blocksSinceDesperation := int(currentHeight) - int(t.Start) - constants.ETHDKGDesperationDelay
	amILeading := dkg.AmILeading(t.State.NumberOfValidators, t.State.Index-1, blocksSinceDesperation, t.StartBlockHash.Bytes(), logger)

	logger.WithFields(logrus.Fields{
		"currentHeight":                    currentHeight,
		"t.Start":                          t.Start,
		"constants.ETHDKGDesperationDelay": constants.ETHDKGDesperationDelay,
		"blocksSinceDesperation":           blocksSinceDesperation,
		"amILeading":                       amILeading,
	}).Infof("dkg.AmILeading")

	return amILeading
}
