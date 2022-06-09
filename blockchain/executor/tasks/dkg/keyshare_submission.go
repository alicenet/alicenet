package dkg

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/executor/constants"
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"
	dkgUtils "github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/utils"
)

// KeyShareSubmissionTask is the task for submitting Keyshare information
type KeyShareSubmissionTask struct {
	*objects.Task
}

// asserting that KeyShareSubmissionTask struct implements interface interfaces.Task
var _ interfaces.ITask = &KeyShareSubmissionTask{}

// NewKeyShareSubmissionTask creates a new task
func NewKeyShareSubmissionTask(start uint64, end uint64) *KeyShareSubmissionTask {
	return &KeyShareSubmissionTask{
		Task: objects.NewTask(constants.KeyShareSubmissionTaskName, start, end, false),
	}
}

// Prepare prepares for work to be done in the KeyShareSubmissionTask.
// Here, the G1 key share, G1 proof, and G2 key share are constructed
// and stored for submission.
func (t *KeyShareSubmissionTask) Prepare() error {
	logger := t.GetLogger()
	logger.Info("KeyShareSubmissionTask Prepare()...")

	dkgState := &state.DkgState{}
	err := t.GetDB().Update(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		if err != nil {
			return err
		}

		me := dkgState.Account.Address

		// check if task already defined key shares
		if dkgState.Participants[me].KeyShareG1s[0] == nil ||
			dkgState.Participants[me].KeyShareG1s[1] == nil ||
			(dkgState.Participants[me].KeyShareG1s[0].Cmp(big.NewInt(0)) == 0 &&
				dkgState.Participants[me].KeyShareG1s[1].Cmp(big.NewInt(0)) == 0) {

			// Generate the key shares
			g1KeyShare, g1Proof, g2KeyShare, err := state.GenerateKeyShare(dkgState.SecretValue)
			if err != nil {
				return err
			}

			dkgState.Participants[me].KeyShareG1s = g1KeyShare
			dkgState.Participants[me].KeyShareG1CorrectnessProofs = g1Proof
			dkgState.Participants[me].KeyShareG2s = g2KeyShare

			err = dkgState.PersistState(txn)
			if err != nil {
				return err
			}
		} else {
			logger.Infof("KeyShareSubmissionTask Prepare(): key shares already defined")
		}

		return nil
	})

	if err != nil {
		return dkgUtils.LogReturnErrorf(logger, "KeyShareSubmissionTask.Prepare(): error during the preparation: %v", err)
	}

	return nil
}

// Execute executes the task business logic
func (t *KeyShareSubmissionTask) Execute() ([]*types.Transaction, error) {
	logger := t.GetLogger()
	logger.Info("KeyShareSubmissionTask doTask()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "KeyShareSubmissionTask.Execute(): error loading dkgState: %v", err)
	}

	// Setup
	me := dkgState.Account

	// Setup
	eth := t.GetEth()
	ctx := t.GetCtx()
	txnOpts, err := eth.GetTransactionOpts(ctx, dkgState.Account)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "getting txn opts failed: %v", err)
	}

	// Submit Keyshares
	logger.Infof("submitting key shares: %v %v %v %v",
		me.Address,
		dkgState.Participants[me.Address].KeyShareG1s,
		dkgState.Participants[me.Address].KeyShareG1CorrectnessProofs,
		dkgState.Participants[me.Address].KeyShareG2s)
	txn, err := eth.Contracts().Ethdkg().SubmitKeyShare(txnOpts,
		dkgState.Participants[me.Address].KeyShareG1s,
		dkgState.Participants[me.Address].KeyShareG1CorrectnessProofs,
		dkgState.Participants[me.Address].KeyShareG2s)
	if err != nil {
		return nil, dkgUtils.LogReturnErrorf(logger, "submitting keyshare failed: %v", err)
	}

	return []*types.Transaction{txn}, nil
}

// ShouldExecute checks if it makes sense to execute the task
func (t *KeyShareSubmissionTask) ShouldExecute() bool {
	logger := t.GetLogger()
	logger.Info("KeyShareSubmissionTask ShouldExecute()")

	dkgState := &state.DkgState{}
	err := t.GetDB().View(func(txn *badger.Txn) error {
		err := dkgState.LoadState(txn)
		return err
	})
	if err != nil {
		logger.Errorf("could not get dkgState with error %v", err)
		return true
	}

	eth := t.GetEth()
	ctx := t.GetCtx()
	me := dkgState.Account
	callOpts, err := eth.GetCallOpts(ctx, me)
	if err != nil {
		logger.Debugf("KeyShareSubmissionTask ShouldExecute failed getting call options: %v", err)
		return true
	}

	phase, err := eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
	if err != nil {
		logger.Infof("KeyShareSubmissionTask ShouldExecute GetETHDKGPhase error: %v", err)
		return true
	}

	// DisputeShareDistribution || KeyShareSubmission
	if phase != uint8(state.DisputeShareDistribution) && phase != uint8(state.KeyShareSubmission) {
		return false
	}

	// Check the key share submission status
	status, err := state.CheckKeyShare(ctx, eth.Contracts().Ethdkg(), logger, callOpts, me.Address, dkgState.Participants[me.Address].KeyShareG1s)
	if err != nil {
		logger.Errorf("KeyShareSubmissionTask ShouldExecute CheckKeyShare error: %v", err)
		return true
	}

	if status == state.KeyShared || status == state.BadKeyShared {
		return false
	}

	return true
}
