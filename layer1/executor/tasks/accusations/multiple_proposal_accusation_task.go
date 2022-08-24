package accusations

import (
	"context"
	"fmt"

	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/core/types"
)

// MultipleProposalAccusationTask executes a simple example task
type MultipleProposalAccusationTask struct {
	// All tasks should start by being composed by the BaseTask
	*tasks.BaseTask
	// everything from here onwards are fields that are unique for this task. If
	// a field is exposed it will be serialized, persisted and restored during
	// eventual crashes.
	Signature0 []byte
	Proposal0  *objs.PClaims
	Signature1 []byte
	Proposal1  *objs.PClaims
}

// asserting that MultipleProposalTask struct implements interface tasks.Task, all
// tasks should conform to this interface
var _ tasks.Task = &MultipleProposalAccusationTask{}

func NewMultipleProposalAccusationTask(signature0 []byte, proposal0 *objs.PClaims, signature1 []byte, proposal1 *objs.PClaims) *MultipleProposalAccusationTask {
	multipleProposalTask := &MultipleProposalAccusationTask{
		BaseTask:   tasks.NewBaseTask(0, 0, true, nil),
		Signature0: signature0,
		Proposal0:  proposal0,
		Signature1: signature1,
		Proposal1:  proposal1,
	}
	return multipleProposalTask
}

// Prepare prepares for work to be done by the task. Put in here any preparation
// logic to execute the task action. E.g preparing a snapshot, computing ethdkg
// data to send to smart contracts. This function should be simple. ALWAYS USE
// THE CTX THAT IT'S PASSED TO THIS FUNCTION. IN CASE YOU NEED ANOTHER CONTEXT
// (E.G WithTimeout) CREATE THE CONTEXT FROM THE CTX PASSED TO THIS FUNCTION.
// ALWAYS MAKE SURE THAT THERE'S NO POSSIBILITY OF INFINITE LOOP THAT DOESN'T
// CHECK THE CTX IN HERE.
func (t *MultipleProposalAccusationTask) Prepare(ctx context.Context) *tasks.TaskErr {
	// if you function needs to persist state, or share/carry over state with other
	// tasks, use the database that is shared with the tasks objects (monitor
	// database). The initial state can be created here, or externally, a new 2
	// letters database entry should be added to the `constants/dbprefix/layer1.go`
	// to save this task state.

	// Simple example (look snapshots or ethdkg tasks for more information):

	/*
		// Get GetExampleState is an auxiliary function to retrieve the state from db
		// and check errors. Check `layer1/executor/tasks/state/snapshots` for more information.
		exampleState, err := state.GetExampleState(s.GetDB())
		if err != nil {
			return tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorDuringPreparation, err), false)
		}
		// do cool stuff and preparation
		exampleState.Foo = 42
		// SaveExampleState is an auxiliary function to save the state in the db
		// and check errors. Check `layer1/executor/tasks/state/snapshots` task for more examples.
		err = state.SaveExampleState(s.GetDB(), exampleState)
		if err != nil {
			return tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorDuringPreparation, err), false)
		}
	*/

	// todo: inform accusation manager that this task is persisted and it's now crash resillient. or update directly in DB

	return nil
}

// Execute executes the task business logic (main action). This function may or
// may not call the layer 1 smart contracts. In case this function doesn't call
// the smart contracts it should return nil, otherwise it should return the
// transaction. ALWAYS USE THE CTX THAT IT'S PASSED TO THIS FUNCTION. IN CASE
// YOU NEED ANOTHER CONTEXT (E.G WithTimeout) CREATE THE CONTEXT FROM THE CTX
// PASSED TO THIS FUNCTION. ALWAYS MAKE SURE THAT THERE'S NO POSSIBILITY OF
// INFINITE LOOP THAT DOESN'T CHECK THE CTX IN HERE.
func (t *MultipleProposalAccusationTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {

	client := t.GetClient()
	txnOpts, err := client.GetTransactionOpts(ctx, client.GetDefaultAccount())
	if err != nil {
		// if it failed here, it means that we are not willing to pay the tx costs based on config or we
		// failed to retrieve tx fee data from the ethereum node
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingTxnOpts, err), true)
	}

	prop0, err := t.Proposal0.MarshalBinary()
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("MultipleProposalAccusationTask: failed to marshal proposal0: %v", err), false)
	}

	prop1, err := t.Proposal1.MarshalBinary()
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("MultipleProposalAccusationTask: failed to marshal proposal1: %v", err), false)
	}

	t.GetLogger().Infof("MultipleProposalAccusationTask: trying to call smart contract to accuse of multiple proposals, id: %s", t.GetId())
	txn, err := t.GetContractsHandler().EthereumContracts().MultipleProposalAccusation().AccuseMultipleProposal(
		txnOpts,
		t.Signature0,
		prop0,
		t.Signature1,
		prop1,
	)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("MultipleProposalAccusationTask: failed to accuse: %v", err), true)
	}
	return txn, nil
}

// ShouldExecute checks if it makes sense to execute the task. This function
// should contains the logic to see if the actions were already performed (by
// someone else or by this node) or if the execute function succeeded (e.g the
// data was committed to a layer 1 smart contract state). This function will be
// used to check if the execute operation succeeded in addition to the receipt
// of the txn created by the Execute method (in case an execution happened).
// This function should return true in case we should execute or false if the
// completion requirements were fulfilled and there's no need to execute. ALWAYS
// USE THE CTX THAT IT'S PASSED TO THIS FUNCTION. IN CASE YOU NEED ANOTHER
// CONTEXT (E.G WithTimeout) CREATE THE CONTEXT FROM THE CTX PASSED TO THIS
// FUNCTION. ALWAYS MAKE SURE THAT THERE'S NO POSSIBILITY OF INFINITE LOOP THAT
// DOESN'T CHECK THE CTX IN HERE.
func (t *MultipleProposalAccusationTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {

	client := t.GetClient()
	// The GetCallOpts will retrieve the information in layer 1 smart contracts with
	// a delay of FINALITY_BLOCKS. I.e, the information will be there after `X`
	// blocks has passed after the transaction was executed. We do this, to avoid
	// chain re-orgs.
	callOpts, err := client.GetCallOpts(ctx, t.GetClient().GetDefaultAccount())
	if err != nil {
		return true, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingCallOpts, err), true)
	}
	var id [32]byte = utils.HexToBytes32(t.GetId())

	isAccused, err := t.GetContractsHandler().EthereumContracts().MultipleProposalAccusation().IsAccused(callOpts, id)
	if err != nil {
		t.GetLogger().Infof("MultipleProposalAccusationTask: error checking IsAccused: %v", err)
		return true, tasks.NewTaskErr(fmt.Sprintf("MultipleProposalAccusationTask: error getting foo in the example task: %v", err), true)
	}

	t.GetLogger().Infof("MultipleProposalAccusationTask: accusation \"0x%x\" done: %v", id, isAccused)

	return !isAccused, nil
}
