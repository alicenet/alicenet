package dkgtasks_test

import (
	"context"
	"math"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/monitor"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const SETUP_GROUP int = 13

type adminHandlerMock struct {
	snapshotCalled     bool
	privateKeyCalled   bool
	validatorSetCalled bool
	registerSnapshot   bool
	setSynchronized    bool
}

func (ah *adminHandlerMock) AddPrivateKey([]byte, constants.CurveSpec) error {
	ah.privateKeyCalled = true
	return nil
}

func (ah *adminHandlerMock) AddSnapshot(*objs.BlockHeader, bool) error {
	ah.snapshotCalled = true
	return nil
}

func (ah *adminHandlerMock) AddValidatorSet(*objs.ValidatorSet) error {
	ah.validatorSetCalled = true
	return nil
}

func (ah *adminHandlerMock) RegisterSnapshotCallback(func(*objs.BlockHeader) error) {
	ah.registerSnapshot = true
}

func (ah *adminHandlerMock) SetSynchronized(v bool) {
	ah.setSynchronized = true
}

func connectSimulatorEndpoint(t *testing.T, accountAddresses []string) interfaces.Ethereum {
	eth, err := blockchain.NewEthereumSimulator(
		"../../../assets/test/keys",
		"../../../assets/test/passcodes.txt",
		6,
		1*time.Second,
		5*time.Second,
		0,
		big.NewInt(math.MaxInt64),
		accountAddresses...)

	assert.Nil(t, err, "Failed to build Ethereum endpoint...")
	assert.True(t, eth.IsEthereumAccessible(), "Web3 endpoint is not available.")

	// Mine a block once a second
	go func() {
		for {
			time.Sleep(1 * time.Second)
			eth.Commit()
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Unlock the default account and use it to deploy contracts
	deployAccount := eth.GetDefaultAccount()
	err = eth.UnlockAccount(deployAccount)
	assert.Nil(t, err, "Failed to unlock default account")

	// Deploy all the contracts
	c := eth.Contracts()
	_, _, err = c.DeployContracts(ctx, deployAccount)
	assert.Nil(t, err, "Failed to deploy contracts...")

	// For each address passed set them up as a validator
	txnOpts, err := eth.GetTransactionOpts(ctx, deployAccount)
	assert.Nil(t, err, "Failed to create txn opts")
	for idx := 1; idx < len(accountAddresses); idx++ {
		acct, err := eth.GetAccount(common.HexToAddress(accountAddresses[idx]))
		assert.Nil(t, err)
		eth.UnlockAccount(acct)

		// 1. Give 'acct' tokens
		txn, err := c.StakingToken().Transfer(txnOpts, acct.Address, big.NewInt(10_000_000))
		assert.Nilf(t, err, "Failed on transfer %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)

		o, err := eth.GetTransactionOpts(ctx, acct)
		assert.Nil(t, err)

		// 2. Allow system to take tokens from 'acct' for staking
		txn, err = c.StakingToken().Approve(o, c.ValidatorsAddress(), big.NewInt(10_000_000))
		assert.Nilf(t, err, "Failed on approval %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)

		// 3. Tell system to take tokens from 'acct' for staking
		txn, err = c.Staking().LockStake(o, big.NewInt(1_000_000))
		assert.Nilf(t, err, "Failed on lock %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)

		// 4. Tell system 'acct' wants to join as validator
		var validatorId [2]*big.Int
		validatorId[0] = big.NewInt(int64(idx))
		validatorId[1] = big.NewInt(int64(idx * 2))

		txn, err = c.Validators().AddValidator(o, acct.Address, validatorId)
		assert.Nilf(t, err, "Failed on register %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)
	}

	// Wait for all transactions for all accounts to complete
	rcpts, err := eth.Queue().WaitGroupTransactions(ctx, SETUP_GROUP)
	assert.Nil(t, err)

	// Make sure all transactions were successful
	t.Logf("# rcpts: %v", len(rcpts))
	for _, rcpt := range rcpts {
		assert.Equal(t, uint64(1), rcpt.Status)
	}

	return eth
}

func validator(t *testing.T, idx int, eth interfaces.Ethereum, validatorAcct accounts.Account, adminHandler *adminHandlerMock, wg *sync.WaitGroup, tr *objects.TypeRegistry) {

	defer wg.Done()

	logger := logging.GetLogger("validator").
		WithField("Index", idx).
		WithField("Address", validatorAcct.Address.Hex())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dkgState := objects.NewDkgState(validatorAcct)
	schedule := objects.NewSequentialSchedule(tr, adminHandler)

	monitorState := objects.NewMonitorState(dkgState, schedule)
	monitorState.HighestBlockProcessed = 0
	monitorState.HighestBlockFinalized = 1

	events := objects.NewEventMap()

	monitor.SetupEventMap(events, nil, adminHandler, nil)

	var done bool

	for !done {
		err := monitor.MonitorTick(ctx, cancel, wg, eth, monitorState, logger, events, adminHandler, 10)
		assert.Nil(t, err)

		time.Sleep(time.Second)

		// Quit test when we either:
		// 1) complete successfully, or
		// 2) past the point when we possibly could. This means we aborted somewhere along the way and failed DKG
		dkgState.RLock()
		done = dkgState.Complete || (dkgState.CompleteEnd > 0 && monitorState.HighestBlockProcessed >= dkgState.CompleteEnd)
		logger.WithFields(logrus.Fields{
			"Complete":              dkgState.Complete,
			"CompleteEnd":           dkgState.CompleteEnd,
			"HighestBlockProcessed": monitorState.HighestBlockProcessed,
			"Done":                  done,
		}).Info("Completion check")
		dkgState.RUnlock()
	}

	// Make sure we used the admin handler
	assert.True(t, adminHandler.privateKeyCalled)
	assert.True(t, dkgState.Complete)
}

func TestDkgSuccess(t *testing.T) {

	var accountAddresses []string = []string{
		"0x546F99F244b7B58B855330AE0E2BC1b30b41302F", "0x9AC1c9afBAec85278679fF75Ef109217f26b1417",
		"0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac", "0x615695C4a4D6a60830e5fca4901FbA099DF26271",
		"0x63a6627b79813A7A43829490C4cE409254f64177", "0x16564cf3e880d9f5d09909f51b922941ebbbc24d"}

	for _, logger := range logging.GetKnownLoggers() {
		logger.SetLevel(logrus.DebugLevel)
	}

	eth := connectSimulatorEndpoint(t, accountAddresses)
	defer eth.Close()

	var ownerAccount accounts.Account

	for idx := range accountAddresses {
		a, err := eth.GetAccount(common.HexToAddress(accountAddresses[idx]))
		assert.Nil(t, err)
		err = eth.UnlockAccount(a)
		assert.Nil(t, err)

		if idx == 0 {
			ownerAccount = a
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := eth.Contracts()

	callOpts := eth.GetCallOpts(ctx, ownerAccount)
	txnOpts, err := eth.GetTransactionOpts(context.Background(), ownerAccount)
	assert.Nil(t, err)

	// Start validators running
	wg := sync.WaitGroup{}
	tr := &objects.TypeRegistry{}
	adminHandlers := make([]*adminHandlerMock, 0)
	SetupTasks(tr)
	for i := 0; i < 5; i++ {
		acct, err := eth.GetAccount(common.HexToAddress(accountAddresses[i+1]))
		assert.Nil(t, err)

		adminHandlers = append(adminHandlers, new(adminHandlerMock))
		wg.Add(1)
		go validator(t, i, eth, acct, adminHandlers[i], &wg, tr)
	}

	// Kick off a round of ethdkg
	txn, err := c.Ethdkg().UpdatePhaseLength(txnOpts, big.NewInt(4))
	assert.Nil(t, err)
	eth.Queue().QueueAndWait(ctx, txn)

	txn, err = c.Ethdkg().InitializeState(txnOpts)
	assert.Nil(t, err)
	eth.Queue().QueueAndWait(ctx, txn)

	// Now we know ethdkg is running, let's find out when registration has to happen
	// TODO this should be based on an OpenRegistration event
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	t.Logf("Current Height:%v", currentHeight)

	endingHeight, err := c.Ethdkg().TREGISTRATIONEND(callOpts)
	assert.Nil(t, err)
	t.Logf("Registration Close Height:%v", endingHeight)

	// Wait for validators to complete
	wg.Wait()
}

func TestFoo(t *testing.T) {
	m := make(map[string]int)
	assert.Equal(t, 0, len(m))

	m["a"] = 5
	assert.Equal(t, 1, len(m))

	m["b"] = 3
	assert.Equal(t, 2, len(m))

	t.Logf("m:%p", m)
}

func SetupTasks(tr *objects.TypeRegistry) {
	tr.RegisterInstanceType(&dkgtasks.PlaceHolder{})
	tr.RegisterInstanceType(&dkgtasks.RegisterTask{})
	tr.RegisterInstanceType(&dkgtasks.ShareDistributionTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeTask{})
	tr.RegisterInstanceType(&dkgtasks.KeyshareSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.MPKSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.GPKSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.GPKJDisputeTask{})
	tr.RegisterInstanceType(&dkgtasks.CompletionTask{})
}
