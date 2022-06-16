package dkg

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	exConstants "github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/transaction"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/crypto/bn256"
)

// MPKSubmissionTask stores the data required to submit the mpk
type MPKSubmissionTask struct {
	*objects.Task
}

// asserting that MPKSubmissionTask struct implements interface interfaces.Task
var _ interfaces.ITask = &MPKSubmissionTask{}

// NewMPKSubmissionTask creates a new task
func NewMPKSubmissionTask(start uint64, end uint64) *MPKSubmissionTask {
	return &MPKSubmissionTask{
		Task: objects.NewTask(exConstants.MPKSubmissionTaskName, start, end, false, transaction.NewSubscribeOptions(true, constants.ETHDKGMaxStaleBlocks)),
	}
}

// Prepare prepares for work to be done in the MPKSubmissionTask
// Here we load all key shares and construct the master public key
// to submit in DoWork.
func (t *MPKSubmissionTask) Prepare() *interfaces.TaskErr {
	logger := t.GetLogger()
	logger.Debug("MPKSubmissionTask Prepare()...")

	dkgState := &state.DkgState{}
	var isRecoverable bool
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		if err != nil {
			isRecoverable = false
			return err
		}

		if dkgState.Phase != state.MPKSubmission {
			isRecoverable = false
			return errors.New("it's not in MPKSubmission phase")
		}

		// compute MPK if not yet computed
		if isMasterPublicKeyEmpty(dkgState.MasterPublicKey) {
			client := t.GetClient()
			ctx := t.GetCtx()
			// setup leader election
			block, err := client.GetBlockByNumber(ctx, big.NewInt(int64(t.GetStart())))
			if err != nil {
				isRecoverable = true
				return fmt.Errorf("error getting block by number: %v", err)
			}

			logger.Debugf("block hash: %v\n", block.Hash())
			t.SetStartBlockHash(block.Hash().Bytes())

			// prepare MPK
			g1KeyShares := make([][2]*big.Int, dkgState.NumberOfValidators)
			g2KeyShares := make([][4]*big.Int, dkgState.NumberOfValidators)

			var participantsList = dkgState.GetSortedParticipants()
			for idx, participant := range participantsList {
				// Bringing these in from state but could directly query contract
				g1KeyShares[idx] = dkgState.Participants[participant.Address].KeyShareG1s
				g2KeyShares[idx] = dkgState.Participants[participant.Address].KeyShareG2s

				logger.Debugf("INIT idx:%v pidx:%v address:%v g1:%v g2:%v", idx, participant.Index, participant.Address.Hex(), g1KeyShares[idx], g2KeyShares[idx])

				for i := range g1KeyShares[idx] {
					if g1KeyShares[idx][i] == nil {
						isRecoverable = false
						return fmt.Errorf("Missing g1Keyshare[%v][%v] for %v.", idx, i, participant.Address.Hex())
					}
				}

				for i := range g2KeyShares[idx] {
					if g2KeyShares[idx][i] == nil {
						isRecoverable = false
						return fmt.Errorf("Missing g2Keyshare[%v][%v] for %v.", idx, i, participant.Address.Hex())
					}
				}
			}

			logger.Debugf("# Participants: %v\n", len(dkgState.Participants))

			mpk, err := state.GenerateMasterPublicKey(g1KeyShares, g2KeyShares)
			if err != nil {
				isRecoverable = false
				return fmt.Errorf("Failed to generate master public key:%v", err)
			}

			// Master public key is all we generate here so save it
			dkgState.MasterPublicKey = mpk

			err = dkgState.PersistState(txn)
			if err != nil {
				isRecoverable = false
				return err
			}
		} else {
			logger.Debugf("mpk already defined")
		}

		return nil
	})

	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(constants.ErrorDuringPreparation, err), isRecoverable)
	}

	return nil
}

// Execute executes the task business logic
func (t *MPKSubmissionTask) Execute() ([]*types.Transaction, *interfaces.TaskErr) {
	logger := t.GetLogger()
	logger.Debug("MPKSubmissionTask Execute()...")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	// submit if I'm a leader for this task
	eth := t.GetClient()
	ctx := t.GetCtx()
	if !t.AmILeading(dkgState) {
		return nil, interfaces.NewTaskErr("not leading MPK submission yet", true)
	}

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingTxnOpts, err), true)
	}

	// Submit MPK
	logger.Infof("submitting master public key:%v", dkgState.MasterPublicKey)
	txn, err := eth.Contracts().Ethdkg().SubmitMasterPublicKey(txnOpts, dkgState.MasterPublicKey)
	if err != nil {
		return nil, interfaces.NewTaskErr(fmt.Sprintf("submitting master public key failed: %v", err), true)
	}

	return []*types.Transaction{txn}, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *MPKSubmissionTask) ShouldExecute() *interfaces.TaskErr {
	logger := t.GetLogger().WithField("method", "ShouldExecute()")
	logger.Trace("should execute task")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(constants.ErrorLoadingDkgState, err), false)
	}

	if dkgState.Phase != state.MPKSubmission {
		return interfaces.NewTaskErr(fmt.Sprintf("phase %v different from MPKSubmission", dkgState.Phase), false)
	}

	// if the mpk is empty in the state that we loaded from db, it means that
	// something really bad happened (ew.g initiate was not successful, data
	// corruption)
	if isMasterPublicKeyEmpty(dkgState.MasterPublicKey) {
		return interfaces.NewTaskErr(fmt.Sprintf("phase %v different from MPKSubmission", dkgState.Phase), false)
	}

	client := t.GetClient()
	ctx := t.GetCtx()
	callOpts, err := client.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf(constants.FailedGettingCallOpts, err), true)
	}

	mpkHash, err := client.Contracts().Ethdkg().GetMasterPublicKeyHash(callOpts)
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf("failed to retrieve mpk from smart contracts: %v", err), true)
	}

	// If we fail here, it means that we had a data corruption or we stored wrong
	// data for dkgstate the master public key
	mpkHashBin, err := bn256.MarshalBigIntSlice(dkgState.MasterPublicKey[:])
	if err != nil {
		return interfaces.NewTaskErr(fmt.Sprintf("failed to serialize internal mpk: %v", err), false)
	}

	mpkHashSlice := crypto.Hasher(mpkHashBin)
	if bytes.Equal(mpkHash[:], mpkHashSlice) {
		return interfaces.NewTaskErr("state mpkHash is equal to the received", false)
	}

	logger.Tracef("state mpkHash is not equal to the received, should execute")
	return nil
}

func isMasterPublicKeyEmpty(masterPublicKey [4]*big.Int) bool {
	isNil :=
		(masterPublicKey[0] == nil ||
			masterPublicKey[1] == nil ||
			masterPublicKey[2] == nil ||
			masterPublicKey[3] == nil)

	isAllZero := (masterPublicKey[0].Cmp(big.NewInt(0)) == 0 &&
		masterPublicKey[1].Cmp(big.NewInt(0)) == 0 &&
		masterPublicKey[2].Cmp(big.NewInt(0)) == 0 &&
		masterPublicKey[3].Cmp(big.NewInt(0)) == 0)

	return isNil || isAllZero
}
