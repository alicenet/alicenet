package accusations

import (
	"context"
	"fmt"

	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/core/types"
)

// InvalidUTXOConsumptionAccusationTask executes a simple example task
type InvalidUTXOConsumptionAccusationTask struct {
	// All tasks should start by being composed by the BaseTask
	*tasks.BaseTask
	// everything from here onwards are fields that are unique for this task. If
	// a field is exposed it will be serialized, persisted and restored during
	// eventual crashes.
	PClaims         []byte
	PClaimsSig      []byte
	BClaims         []byte
	BClaimsSigGroup []byte
	TxInPreImage    []byte
	Proofs          [3][]byte
}

// asserting that MultipleProposalTask struct implements interface tasks.Task, all
// tasks should conform to this interface
var _ tasks.Task = &InvalidUTXOConsumptionAccusationTask{}

func NewInvalidUTXOConsumptionAccusationTask(
	pClaims []byte,
	pClaimsSig []byte,
	bClaims []byte,
	bClaimsSigGroup []byte,
	txInPreImage []byte,
	proofs [3][]byte,
) *InvalidUTXOConsumptionAccusationTask {
	invalidUTXOConsumptionTask := &InvalidUTXOConsumptionAccusationTask{
		BaseTask:        tasks.NewBaseTask(0, 0, true, nil),
		PClaims:         pClaims,
		PClaimsSig:      pClaimsSig,
		BClaims:         bClaims,
		BClaimsSigGroup: bClaimsSigGroup,
		TxInPreImage:    txInPreImage,
		Proofs:          proofs,
	}
	return invalidUTXOConsumptionTask
}

// Prepare prepares for work to be done by the task. Put in here any preparation
// logic to execute the task action. E.g preparing a snapshot, computing ethdkg
// data to send to smart contracts. This function should be simple. ALWAYS USE
// THE CTX THAT IT'S PASSED TO THIS FUNCTION. IN CASE YOU NEED ANOTHER CONTEXT
// (E.G WithTimeout) CREATE THE CONTEXT FROM THE CTX PASSED TO THIS FUNCTION.
// ALWAYS MAKE SURE THAT THERE'S NO POSSIBILITY OF INFINITE LOOP THAT DOESN'T
// CHECK THE CTX IN HERE.
func (t *InvalidUTXOConsumptionAccusationTask) Prepare(ctx context.Context) *tasks.TaskErr {
	return nil
}

// Execute executes the task business logic (main action). This function may or
// may not call the layer 1 smart contracts. In case this function doesn't call
// the smart contracts it should return nil, otherwise it should return the
// transaction. ALWAYS USE THE CTX THAT IT'S PASSED TO THIS FUNCTION. IN CASE
// YOU NEED ANOTHER CONTEXT (E.G WithTimeout) CREATE THE CONTEXT FROM THE CTX
// PASSED TO THIS FUNCTION. ALWAYS MAKE SURE THAT THERE'S NO POSSIBILITY OF
// INFINITE LOOP THAT DOESN'T CHECK THE CTX IN HERE.
func (t *InvalidUTXOConsumptionAccusationTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {

	client := t.GetClient()
	txnOpts, err := client.GetTransactionOpts(ctx, client.GetDefaultAccount())
	if err != nil {
		// if it failed here, it means that we are not willing to pay the tx costs based on config or we
		// failed to retrieve tx fee data from the ethereum node
		return nil, tasks.NewTaskErr(fmt.Sprintf(tasks.FailedGettingTxnOpts, err), true)
	}

	t.GetLogger().Warnf("InvalidUTXOConsumptionAccusationTask: trying to call smart contract to accuse of invalid UTXO consumption, id: %s", t.GetId())
	txn, err := t.GetContractsHandler().EthereumContracts().AccusationInvalidTxConsumption().AccuseInvalidTransactionConsumption(
		txnOpts,
		t.PClaims,
		t.PClaimsSig,
		t.BClaims,
		t.BClaimsSigGroup,
		t.TxInPreImage,
		t.Proofs,
	)
	if err != nil {
		return nil, tasks.NewTaskErr(fmt.Sprintf("InvalidUTXOConsumptionAccusationTask: failed to accuse: %v", err), true)
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
func (t *InvalidUTXOConsumptionAccusationTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {

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

	isAccused, err := t.GetContractsHandler().EthereumContracts().AccusationInvalidTxConsumption().IsAccused(callOpts, id)
	if err != nil {
		t.GetLogger().Infof("InvalidUTXOConsumptionAccusationTask: error checking IsAccused: %v", err)
		return true, tasks.NewTaskErr(fmt.Sprintf("InvalidUTXOConsumptionAccusationTask: error getting foo in the example task: %v", err), true)
	}

	t.GetLogger().Infof("InvalidUTXOConsumptionAccusationTask: accusation \"0x%x\" done: %v", id, isAccused)

	return !isAccused, nil
}
