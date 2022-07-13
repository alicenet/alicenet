package examples

import (
	"context"
	"time"

	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/ethereum/go-ethereum/core/types"
)

// SnapshotTask pushes a snapshot to Ethereum
type SimpleExampleTask struct {
	// All tasks should start by being composed by the BaseTask
	*tasks.BaseTask
	// everything from here onwards are fields that are exclusive for this task. If
	// a field is exposed it will be serialized, persisted and restored during
	// eventual crashes.
	Foo uint64
}

// asserting that SimpleExampleTask struct implements interface tasks.Task, all
// tasks should conform to this interface
var _ tasks.Task = &SimpleExampleTask{}

// OPTIONAL: Auxiliary function to create this task
func NewSimpleExampleTask(start uint64, end uint64) *SimpleExampleTask {
	snapshotTask := &SimpleExampleTask{
		// Parameters to start the base task. Check the tasks parameters to see the
		// description of each field.
		BaseTask: tasks.NewBaseTask(start, end, false, nil),
		Foo:      0,
	}
	return snapshotTask
}

// Prepare prepares for work to be done by the task. Put in here any preparation
// logic to execute the task action. E.g preparing an snapshot, computing ethdkg
// data to send to smart contracts. This function should be simple. ALWAYS USE
// THE CTX THAT IT'S PASSED TO THIS FUNCTION. IN CASE YOU NEED ANOTHER CONTEXT
// (E.G WithTimeout) CREATE THE CONTEXT FROM THE CTX PASSED TO THIS FUNCTION.
// ALWAYS MAKE SURE THAT THERE'S NO POSSIBILITY OF INFINITE LOOP THAT DOESN'T
// CHECK THE CTX IN HERE.
func (s *SimpleExampleTask) Prepare(ctx context.Context) *tasks.TaskErr {
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

	return nil
}

// Execute executes the task business logic (main action). This function may or
// may not call the layer 1 smart contracts. In case this function doesn't call
// the smart contracts it should return nil, otherwise it should return the
// transaction. ALWAYS USE THE CTX THAT IT'S PASSED TO THIS FUNCTION. IN CASE
// YOU NEED ANOTHER CONTEXT (E.G WithTimeout) CREATE THE CONTEXT FROM THE CTX
// PASSED TO THIS FUNCTION. ALWAYS MAKE SURE THAT THERE'S NO POSSIBILITY OF
// INFINITE LOOP THAT DOESN'T CHECK THE CTX IN HERE.
func (s *SimpleExampleTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {

	// If this function needs to read/write state and share with the other tasks use
	// the db. Check the `Prepare` documentation above for more information.

	/*

		exampleState, err := state.GetExampleState(s.GetDB())
		if err != nil {
			return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.ErrorLoadingDkgState, err), false)
		}

	*/

	// The `Execute` method may or may not call a layer1 smart contract. In case it
	// needs to call an smart contract a transaction should be returned. E.g

	/*

		txnOpts, err := client.GetTransactionOpts(ctx, exampleState.Account)
		if err != nil {
			// if it failed here, it means that we are not willing to pay the tx costs based on config or we
			// failed to retrieve tx fee data from the ethereum node
			return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingTxnOpts, err), true)
		}
		logger.Info("trying to call smart contract in the example")
		txn, err := ethereum.GetContracts().DummyContract().SetFoo(txnOpts, exampleState.Foo)
		if err != nil {
			return nil, tasks.NewTaskErr(fmt.Sprintf("failed to send snapshot: %v", err), true)
		}
		return txn, nil

	*/

	// In case we don't need to call an smart contract. Just return nil,nil when the
	// action was performed (e.g checked if a validator is honest, so no transaction
	// is needed). Sending nil here will terminate the task successfully. See
	// `layer1/executor/tasks/dkg/dispute_gpkj.go` for more details.

	// Simulating cool stuff
	<-time.After(10 * time.Second)
	return nil, nil
}

// ShouldExecute checks if it makes sense to execute the task. This function
// should contains the logic to see if the actions were already performed (by
// someone else or by this node) or if the execute function succeeded (e.g the
// data was committed to a layer 1 smart contract state). This function will be
// used to check if the execute operation succeeded in addition to the receipt
// of the txn created by the Execute method (in case an execution happened).
// This function should return true in case we should execute or false it the
// completion requirements were fulfilled and there's no need to execute. ALWAYS
// USE THE CTX THAT IT'S PASSED TO THIS FUNCTION. IN CASE YOU NEED ANOTHER
// CONTEXT (E.G WithTimeout) CREATE THE CONTEXT FROM THE CTX PASSED TO THIS
// FUNCTION. ALWAYS MAKE SURE THAT THERE'S NO POSSIBILITY OF INFINITE LOOP THAT
// DOESN'T CHECK THE CTX IN HERE.
func (s *SimpleExampleTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {

	// Check here if the action was performed successfully or if we should execute
	// at all (the requirement is already fulfilled, e.g someone already sent the
	// snapshot, we already registered ethdkg and the data is in the smart contract).

	/*

		client := t.GetClient()

		// The GetCallOpts will retrieve the information in layer 1 smart contracts with
		// a delay of FINALITY_BLOCKS. I.e, the information will be there after `X`
		// blocks has passed after the transaction was executed. We do this, to avoid
		// chain re-orgs.

		callOpts, err := client.GetCallOpts(ctx, s.GetClient().GetDefaultAccount())
		if err != nil {
			return false, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingCallOpts, err), true)
		}

		foo, err := ethereum.GetContracts().DummyContract().GetFoo(callOpts)
		if err != nil {
			return false, tasks.NewTaskErr(fmt.Sprintf("error getting foo in the example task: %v", err), true)
		}

		return foo == 42

	*/

	return true, nil
}
