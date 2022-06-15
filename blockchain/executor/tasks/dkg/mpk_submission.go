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
	executorInterfaces "github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
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
func (t *MPKSubmissionTask) Prepare() *executorInterfaces.TaskErr {
	logger := t.GetLogger()
	logger.Info("MPKSubmissionTask Prepare()...")

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
		return executorInterfaces.NewTaskErr(fmt.Sprintf(constants.ErrorDuringPreparation, err), isRecoverable)
	}

	return nil
}

// Execute executes the task business logic
func (t *MPKSubmissionTask) Execute() ([]*types.Transaction, *executorInterfaces.TaskErr) {
	logger := t.GetLogger()
	logger.Info("MPKSubmissionTask Execute()...")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("MPKSubmissionTask.Execute(): error loading dkgState: %v", err)
	}

	// submit if I'm a leader for this task
	eth := t.GetClient()
	ctx := t.GetCtx()
	if !t.AmILeading(dkgState) {
		return nil, errors.New("not leading MPK submission yet")
	}

	// Setup
	txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, utils.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Submit MPK
	logger.Infof("submitting master public key:%v", dkgState.MasterPublicKey)
	txn, err := eth.Contracts().Ethdkg().SubmitMasterPublicKey(txnOpts, dkgState.MasterPublicKey)
	if err != nil {
		return nil, utils.LogReturnErrorf(logger, "submitting master public key failed: %v", err)
	}

	return []*types.Transaction{txn}, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *MPKSubmissionTask) ShouldExecute() (bool, *executorInterfaces.TaskErr) {
	logger := t.GetLogger()
	logger.Info("MPKSubmissionTask ShouldExecute()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		logger.Errorf("could not get dkgState with error %v", err)
		return true
	}

	if dkgState.Phase != state.MPKSubmission {
		return false
	}

	return t.shouldSubmitMPK(dkgState)
}

func (t *MPKSubmissionTask) shouldSubmitMPK(dkgState *state.DkgState) bool {
	if isMasterPublicKeyEmpty(dkgState.MasterPublicKey) {
		return false
	}

	logger := t.GetLogger()
	eth := t.GetClient()
	ctx := t.GetCtx()
	callOpts, err := eth.GetCallOpts(ctx, dkgState.Account)
	if err != nil {
		logger.Error(fmt.Sprintf("MPKSubmissionTask shouldSubmitMPK() failed getting call options: %v", err))
		return true
	}

	mpkHash, err := eth.Contracts().Ethdkg().GetMasterPublicKeyHash(callOpts)
	if err != nil {
		return true
	}

	logger.WithField("Method", "shouldSubmitMPK").Debugf("mpkHash received")

	mpkHashBin, err := bn256.MarshalBigIntSlice(dkgState.MasterPublicKey[:])
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
