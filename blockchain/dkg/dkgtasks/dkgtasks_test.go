package dkgtasks_test

import (
	"context"
	"crypto/ecdsa"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgevents"
	"math"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/MadBase/bridge/bindings"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/dkg/dtest"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/monitor"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/accounts"
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

func connectSimulatorEndpoint(t *testing.T, privateKeys []*ecdsa.PrivateKey, blockInterval time.Duration) interfaces.Ethereum {
	eth, err := blockchain.NewEthereumSimulator(
		privateKeys,
		6,
		1*time.Second,
		5*time.Second,
		0,
		big.NewInt(math.MaxInt64))

	assert.Nil(t, err, "Failed to build Ethereum endpoint...")
	assert.True(t, eth.IsEthereumAccessible(), "Web3 endpoint is not available.")

	// Mine a block once a second
	if blockInterval > 1*time.Millisecond {
		go func() {
			for {
				time.Sleep(blockInterval)
				eth.Commit()
			}
		}()
	}

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

	accountList := eth.GetKnownAccounts()
	for idx, acct := range accountList {
		//	for idx := 1; idx < len(accountAddresses); idx++ {
		// acct, err := eth.GetAccount(common.HexToAddress(accountAddresses[idx]))
		//assert.Nil(t, err)
		err = eth.UnlockAccount(acct)
		assert.Nil(t, err)

		eth.Commit()

		t.Logf("# unlocked %v of %v", idx+1, len(accountList))

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
		//var validatorId [2]*big.Int
		//validatorId[0] = big.NewInt(int64(idx))
		//validatorId[1] = big.NewInt(int64(idx * 2))

		txn, err = c.ValidatorPool().AddValidator(o, acct.Address)
		assert.Nilf(t, err, "Failed on register %v", idx)
		assert.NotNil(t, txn)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)
		t.Logf("Finished loop %v of %v", idx+1, len(accountList))
		eth.Commit()
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
		phase := dkgState.Phase
		done = phase == objects.Completion
		logger.WithFields(logrus.Fields{
			"Phase":                 phase,
			"HighestBlockProcessed": monitorState.HighestBlockProcessed,
			"HighestBlockFinalized": monitorState.HighestBlockFinalized,
			"Done":                  done,
		}).Info("Completion check")
		dkgState.RUnlock()
	}

	// Make sure we used the admin handler
	assert.True(t, adminHandler.privateKeyCalled)
	assert.Equal(t, objects.Completion, dkgState.Phase)
}

func SetupTasks(tr *objects.TypeRegistry) {
	tr.RegisterInstanceType(&dkgtasks.PlaceHolder{})
	tr.RegisterInstanceType(&dkgtasks.RegisterTask{})
	tr.RegisterInstanceType(&dkgtasks.ShareDistributionTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeShareDistributionTask{})
	tr.RegisterInstanceType(&dkgtasks.KeyshareSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.MPKSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.GPKjSubmissionTask{})
	tr.RegisterInstanceType(&dkgtasks.DisputeGPKjTask{})
	tr.RegisterInstanceType(&dkgtasks.CompletionTask{})
}

func advanceTo(t *testing.T, eth interfaces.Ethereum, target uint64) {
	height, err := eth.GetFinalizedHeight(context.Background())
	assert.Nil(t, err)

	distance := target - height

	for i := uint64(0); i < distance; i++ {
		eth.Commit()
	}
}

func IgnoreTestDkgSuccess(t *testing.T) {
	for _, logger := range logging.GetKnownLoggers() {
		logger.SetLevel(logrus.DebugLevel)
	}

	dkgStates, ecdsaPrivateKeys := dtest.InitializeNewDetDkgStateInfo(5)

	t.Logf("dkgStates:%v", dkgStates)
	t.Logf("ecdsaPrivateKeys:%v", ecdsaPrivateKeys)

	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 100*time.Millisecond)
	defer eth.Close()

	accountList := eth.GetKnownAccounts()
	var ownerAccount accounts.Account

	for idx, account := range accountList {
		err := eth.UnlockAccount(account)
		assert.Nil(t, err)

		if idx == 0 {
			ownerAccount = account
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := eth.Contracts()

	txnOpts, err := eth.GetTransactionOpts(context.Background(), ownerAccount)
	assert.Nil(t, err)

	// Start validators running
	wg := sync.WaitGroup{}
	tr := &objects.TypeRegistry{}
	adminHandlers := make([]*adminHandlerMock, len(accountList))
	SetupTasks(tr)
	//	for i := 1; i < len(accountList); i++ {
	for i, account := range accountList {
		adminHandlers[i] = new(adminHandlerMock)
		wg.Add(1)
		go validator(t, i, eth, account, adminHandlers[i], &wg, tr)
	}

	// Kick off a round of ethdkg
	txn, err := c.Ethdkg().SetPhaseLength(txnOpts, 10)
	assert.Nil(t, err)
	eth.Queue().QueueAndWait(ctx, txn)

	txn, err = c.Ethdkg().InitializeETHDKG(txnOpts)
	assert.Nil(t, err)
	eth.Queue().QueueAndWait(ctx, txn)

	// Now we know ethdkg is running, let's find out when registration has to happen
	// TODO this should be based on an OpenRegistration event
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	t.Logf("Current Height:%v", currentHeight)

	// Wait for validators to complete
	wg.Wait()
}

type TestSuite struct {
	eth                 interfaces.Ethereum
	dkgStates           []*objects.DkgState
	ecdsaPrivateKeys    []*ecdsa.PrivateKey
	regTasks            []*dkgtasks.RegisterTask
	dispMissingRegTasks []*dkgtasks.DisputeMissingRegistrationTask
}

func StartFromRegistrationOpenPhase(t *testing.T, n int, unregisteredValidators int, phaseLength uint16) *TestSuite {
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 333*time.Millisecond)
	assert.NotNil(t, eth)

	ctx := context.Background()
	owner := accounts[0]

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	txn, err := eth.Contracts().Ethdkg().SetPhaseLength(ownerOpts, phaseLength)
	assert.Nil(t, err)
	_, err = eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	txn, err = eth.Contracts().ValidatorPool().InitializeETHDKG(ownerOpts)
	assert.Nil(t, err)

	eth.Commit()
	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	assert.Nil(t, err)

	var event *bindings.ETHDKGRegistrationOpened
	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == "0xbda431b9b63510f1398bf33d700e013315bcba905507078a1780f13ea5b354b9" {
			event, err = eth.Contracts().Ethdkg().ParseRegistrationOpened(*log)
			assert.Nil(t, err)
			break
		}
	}
	assert.NotNil(t, event)

	// Do Register task
	regTasks := make([]*dkgtasks.RegisterTask, n)
	dispMissingRegTasks := make([]*dkgtasks.DisputeMissingRegistrationTask, n)
	dkgStates := make([]*objects.DkgState, n)
	logger := logging.GetLogger("test").WithField("Validator", accounts[0].Address.String())
	for idx := 0; idx < n; idx++ {
		// Set Registration success to true
		state, _, regTask, dispMissingRegTask := dkgevents.UpdateStateOnRegistrationOpened(
			accounts[idx],
			event.StartBlock.Uint64(),
			event.PhaseLength.Uint64(),
			event.ConfirmationLength.Uint64(),
			event.Nonce.Uint64())

		dkgStates[idx] = state
		regTasks[idx] = regTask.(*dkgtasks.RegisterTask)
		dispMissingRegTasks[idx] = dispMissingRegTask.(*dkgtasks.DisputeMissingRegistrationTask)
		err = regTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)

		if idx >= n-unregisteredValidators {
			continue
		}

		err = regTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, regTasks[idx].Success)
	}

	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		if idx >= n-unregisteredValidators {
			continue
		}
		for i := 0; i < n; i++ {
			dkgStates[i].Participants[state.Account.Address] = &objects.Participant{
				Address:   state.Account.Address,
				Index:     idx + 1,
				PublicKey: state.TransportPublicKey,
				Phase:     uint8(objects.RegistrationOpen),
				Nonce:     state.Nonce,
			}
		}
	}

	advanceTo(t, eth, event.StartBlock.Uint64()+event.PhaseLength.Uint64())

	return &TestSuite{
		eth:                 eth,
		dkgStates:           dkgStates,
		ecdsaPrivateKeys:    ecdsaPrivateKeys,
		regTasks:            regTasks,
		dispMissingRegTasks: dispMissingRegTasks,
	}
}

func StartFromShareDistributionPhase(t *testing.T, n int, undistributedShares int, phaseLength uint16) *TestSuite {
	suite := StartFromRegistrationOpenPhase(t, n, 0, phaseLength)
	accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	currentHeight, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	// Do Share Distribution task
	shareDistributionTasks := make([]*dkgtasks.ShareDistributionTask, n)
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())

		// set phase
		state.Phase = objects.ShareDistribution
		state.PhaseStart = currentHeight + state.ConfirmationLength
		phaseEnd := state.PhaseStart + state.PhaseLength

		if idx >= n-undistributedShares {
			continue
		}

		shareDistributionTasks[idx] = dkgtasks.NewShareDistributionTask(state, state.PhaseStart, phaseEnd)
		err := shareDistributionTasks[idx].Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)

		err = shareDistributionTasks[idx].DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, shareDistributionTasks[idx].Success)

		// event
		for j := 0; j < n; j++ {
			suite.dkgStates[j].Participants[state.Account.Address].Phase = uint8(objects.ShareDistribution)
		}

	}
	// Ensure all participants have valid share information
	dtest.PopulateEncryptedSharesAndCommitments(t, suite.dkgStates)

	if undistributedShares == 0 {
		// this means all validators distributed their shares and now the phase is
		// set phase to DisputeShareDistribution
		for i := 0; i < n; i++ {
			suite.dkgStates[i].Phase = objects.DisputeShareDistribution
		}
	} else {
		// this means some validators did not distribute shares, and the next phase is DisputeMissingShareDistribution
		advanceTo(t, suite.eth, suite.dkgStates[0].PhaseStart+suite.dkgStates[0].PhaseLength)
	}

	return suite
}
