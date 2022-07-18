package dkg

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/crypto/bn256"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	"github.com/alicenet/alicenet/utils"
)

// MPKSubmissionTask stores the data required to submit the mpk
type MPKSubmissionTask struct {
	*tasks.BaseTask
	// variables that are unique only for this task
	StartBlockHash common.Hash `json:"startBlockHash"`
}

// asserting that MPKSubmissionTask struct implements interface tasks.Task
var _ tasks.Task = &MPKSubmissionTask{}

// NewMPKSubmissionTask creates a new task
func NewMPKSubmissionTask(start uint64, end uint64) *MPKSubmissionTask {
	return &MPKSubmissionTask{
		BaseTask: tasks.NewBaseTask(start, end, false, nil),
	}
}

// Prepare prepares for work to be done in the MPKSubmissionTask
// Here we load all key shares and construct the master public key
// to submit in DoWork.
func (t *MPKSubmissionTask) Prepare(ctx context.Context) *tasks.TaskErr {
	logger := t.GetLogger().WithField("method", "Prepare()")
	logger.Debugf("preparing task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorDuringPreparation, err), false)
	}

	if dkgState.Phase != state.MPKSubmission {
		return tasks.NewTaskErr("it's not in MPKSubmission phase", false)
	}

	// compute MPK if not yet computed
	if isMasterPublicKeyEmpty(dkgState.MasterPublicKey) {
		client := t.GetClient()
		// setup leader election
		block, err := client.GetBlockByNumber(ctx, big.NewInt(int64(t.GetStart())))
		if err != nil {
			return tasks.NewTaskErr(fmt.Sprintf("error getting block by number: %v", err), true)
		}

		logger.Debugf("block hash: %v\n", block.Hash())
		t.StartBlockHash = block.Hash()

		// prepare MPK
		g1KeyShares := make([][2]*big.Int, dkgState.NumberOfValidators)
		g2KeyShares := make([][4]*big.Int, dkgState.NumberOfValidators)

		var participantsList = dkgState.GetSortedParticipants()
		for idx, participant := range participantsList {
			// Bringing these in from state but could directly query contract
			g1KeyShares[idx] = dkgState.Participants[participant.Address].KeyShareG1s
			g2KeyShares[idx] = dkgState.Participants[participant.Address].KeyShareG2s

			logger.Debugf(
				"INIT idx:%v pidx:%v address:%v g1:%v g2:%v",
				idx, participant.Index,
				participant.Address.Hex(),
				g1KeyShares[idx],
				g2KeyShares[idx],
			)

			for i := range g1KeyShares[idx] {
				if g1KeyShares[idx][i] == nil {
					return tasks.NewTaskErr(fmt.Sprintf("Missing g1Keyshare[%v][%v] for %v.", idx, i, participant.Address.Hex()), false)
				}
			}

			for i := range g2KeyShares[idx] {
				if g2KeyShares[idx][i] == nil {
					return tasks.NewTaskErr(fmt.Sprintf("Missing g2Keyshare[%v][%v] for %v.", idx, i, participant.Address.Hex()), false)
				}
			}
		}

		logger.Debugf("# Participants: %v\n", len(dkgState.Participants))

		mpk, err := state.GenerateMasterPublicKey(g1KeyShares, g2KeyShares)
		if err != nil {
			return tasks.NewTaskErr(fmt.Sprintf("Failed to generate master public key:%v", err), false)
		}

		// Master public key is all we generate here so save it
		dkgState.MasterPublicKey = mpk

		err = state.SaveDkgState(t.GetDB(), dkgState)
		if err != nil {
			return tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorDuringPreparation, err), false)
		}
	} else {
		logger.Debugf("mpk already defined")
	}

	return nil
}

// Execute executes the task business logic
func (t *MPKSubmissionTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "Execute()")
	logger.Debug("initiate execution")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	// submit if I'm a leader for this task
	client := t.GetClient()
	isLeading, err := utils.AmILeading(
		client,
		ctx,
		logger,
		int(t.GetStart()),
		t.StartBlockHash.Bytes(),
		dkgState.NumberOfValidators,
		dkgState.Index-1,
		constants.ETHDKGDesperationFactor,
		constants.ETHDKGDesperationDelay,
	)
	if err != nil {
		return nil, tasks.NewTaskErr("error getting eth height for leader election", true)
	}

	if !isLeading {
		return nil, tasks.NewTaskErr("not leading MPK submission yet", true)
	}

	// Setup
	txnOpts, err := client.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingTxnOpts, err), true)
	}

	// Submit MPK
	logger.Infof("submitting master public key:%v", dkgState.MasterPublicKey)
	txn, err := ethereum.GetContracts().Ethdkg().SubmitMasterPublicKey(txnOpts, dkgState.MasterPublicKey)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("submitting master public key failed: %v", err), true)
	}

	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *MPKSubmissionTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Debug("should execute task")

	dkgState, err := state.GetDkgState(t.GetDB())
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.MPKSubmission {
		logger.Debugf("phase %v different from MPKSubmission", dkgState.Phase)
		return false, nil
	}

	// if the mpk is empty in the state that we loaded from db, it means that
	// something really bad happened (e.g initiate was not successful, data
	// corruption)
	if isMasterPublicKeyEmpty(dkgState.MasterPublicKey) {
		return false, tasks.NewTaskErr("empty master public key", false)
	}

	client := t.GetClient()
	callOpts, err := client.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingCallOpts, err), true)
	}

	mpkHash, err := ethereum.GetContracts().Ethdkg().GetMasterPublicKeyHash(callOpts)
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("failed to retrieve mpk from smart contracts: %v", err), true)
	}

	// If we fail here, it means that we had a data corruption or we stored wrong
	// data for dkgstate the master public key
	mpkHashBin, err := bn256.MarshalBigIntSlice(dkgState.MasterPublicKey[:])
	if err != nil {
		return false, tasks.NewTaskErr(fmt.Sprintf("failed to serialize internal mpk: %v", err), false)
	}

	mpkHashSlice := crypto.Hasher(mpkHashBin)
	if bytes.Equal(mpkHash[:], mpkHashSlice) {
		logger.Debug("state mpkHash is equal to the received")
		return false, nil
	}

	logger.Tracef("state mpkHash is not equal to the received, should execute")
	return true, nil
}

func isMasterPublicKeyEmpty(masterPublicKey [4]*big.Int) bool {
	isNil :=
		(masterPublicKey[0] == nil ||
			masterPublicKey[1] == nil ||
			masterPublicKey[2] == nil ||
			masterPublicKey[3] == nil)

	return isNil || (masterPublicKey[0].Cmp(big.NewInt(0)) == 0 &&
		masterPublicKey[1].Cmp(big.NewInt(0)) == 0 &&
		masterPublicKey[2].Cmp(big.NewInt(0)) == 0 &&
		masterPublicKey[3].Cmp(big.NewInt(0)) == 0)
}
