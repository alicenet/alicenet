package dkgtasks_test

import (
	"context"
	"crypto/ecdsa"
	"math"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgevents"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"

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
	eth              interfaces.Ethereum
	dkgStates        []*objects.DkgState
	ecdsaPrivateKeys []*ecdsa.PrivateKey

	regTasks                     []*dkgtasks.RegisterTask
	dispMissingRegTasks          []*dkgtasks.DisputeMissingRegistrationTask
	shareDistTasks               []*dkgtasks.ShareDistributionTask
	disputeMissingShareDistTasks []*dkgtasks.DisputeMissingShareDistributionTask
	disputeShareDistTasks        []*dkgtasks.DisputeShareDistributionTask
	keyshareSubmissionTasks      []*dkgtasks.KeyshareSubmissionTask
	disputeMissingKeyshareTasks  []*dkgtasks.DisputeMissingKeySharesTask
	mpkSubmissionTasks           []*dkgtasks.MPKSubmissionTask
	gpkjSubmissionTasks          []*dkgtasks.GPKjSubmissionTask
	disputeMissingGPKjTasks      []*dkgtasks.DisputeMissingGPKjTask
	disputeGPKjTasks             []*dkgtasks.DisputeGPKjTask
	completionTasks              []*dkgtasks.CompletionTask
}

func StartFromRegistrationOpenPhase(t *testing.T, n int, unregisteredValidators int, phaseLength uint16) *TestSuite {
	ecdsaPrivateKeys, accounts := dtest.InitializePrivateKeysAndAccounts(n)

	eth := connectSimulatorEndpoint(t, ecdsaPrivateKeys, 1000*time.Millisecond)
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

	eventMap := monitor.GetETHDKGEvents()
	eventInfo, ok := eventMap["RegistrationOpened"]
	if !ok {
		t.Fatal("event not found: RegistrationOpened")
	}
	var event *bindings.ETHDKGRegistrationOpened
	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == eventInfo.ID.String() {
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
		regTasks[idx] = regTask
		dispMissingRegTasks[idx] = dispMissingRegTask
		err = regTasks[idx].Initialize(ctx, logger, eth, state)
		assert.Nil(t, err)

		if idx >= n-unregisteredValidators {
			continue
		}

		callOpts := eth.GetCallOpts(ctx, accounts[0])
		phase, err := eth.Contracts().Ethdkg().GetETHDKGPhase(callOpts)
		assert.Nil(t, err)
		assert.Equal(t, uint8(objects.RegistrationOpen), phase)

		nVal, err := eth.Contracts().Ethdkg().GetNumParticipants(callOpts)
		assert.Nil(t, err)
		assert.Equal(t, uint64(idx), nVal.Uint64())

		valCount, err := eth.Contracts().ValidatorPool().GetValidatorsCount(callOpts)
		assert.Nil(t, err)
		assert.Equal(t, uint64(n), valCount.Uint64())

		err = regTasks[idx].DoWork(ctx, logger, eth)
		assert.Nil(t, err)

		eth.Commit()
		assert.True(t, regTasks[idx].Success)
	}

	// simulate receiving AddressRegistered event
	for i := 0; i < n; i++ {
		state := dkgStates[i]

		if i >= n-unregisteredValidators {
			continue
		}

		for j := 0; j < n; j++ {
			dkgStates[j].OnAddressRegistered(state.Account.Address, i+1, state.Nonce, state.TransportPublicKey)
		}
	}

	shareDistributionTasks := make([]*dkgtasks.ShareDistributionTask, n)
	disputeMissingShareDistributionTasks := make([]*dkgtasks.DisputeMissingShareDistributionTask, n)
	disputeShareDistTasks := make([]*dkgtasks.DisputeShareDistributionTask, n)

	if unregisteredValidators == 0 {
		height, err := eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)

		for idx := 0; idx < n; idx++ {
			shareDistributionTask, _, _, disputeMissingShareDistributionTask, disputeShareDistTask, _, _ := dkgevents.UpdateStateOnRegistrationComplete(dkgStates[idx], height)

			shareDistributionTasks[idx] = shareDistributionTask
			disputeMissingShareDistributionTasks[idx] = disputeMissingShareDistributionTask
			disputeShareDistTasks[idx] = disputeShareDistTask
		}

		// skip all the way to ShareDistribution phase
		advanceTo(t, eth, shareDistributionTasks[0].Start)
	} else {
		// this means some validators did not register, and the next phase is DisputeMissingRegistration
		advanceTo(t, eth, dkgStates[0].PhaseStart+dkgStates[0].PhaseLength)
	}

	return &TestSuite{
		eth:                          eth,
		dkgStates:                    dkgStates,
		ecdsaPrivateKeys:             ecdsaPrivateKeys,
		regTasks:                     regTasks,
		dispMissingRegTasks:          dispMissingRegTasks,
		shareDistTasks:               shareDistributionTasks,
		disputeMissingShareDistTasks: disputeMissingShareDistributionTasks,
		disputeShareDistTasks:        disputeShareDistTasks,
	}
}

func StartFromShareDistributionPhase(t *testing.T, n int, undistributedSharesIdx []int, badSharesIdx []int, phaseLength uint16) *TestSuite {
	suite := StartFromRegistrationOpenPhase(t, n, 0, phaseLength)
	//accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do Share Distribution task
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]

		var skipLoop = false

		for _, undistIdx := range undistributedSharesIdx {
			if idx == undistIdx {
				skipLoop = true
			}
		}

		if skipLoop {
			continue
		}

		shareDistTask := suite.shareDistTasks[idx]

		err := shareDistTask.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)

		for _, badIdx := range badSharesIdx {
			if idx == badIdx {
				// inject bad shares
				for _, s := range state.Participants[state.Account.Address].EncryptedShares {
					s.Set(big.NewInt(0))
				}
			}
		}

		err = shareDistTask.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, shareDistTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			err = suite.dkgStates[j].OnSharesDistributed(
				logger,
				state.Account.Address,
				state.Participants[state.Account.Address].EncryptedShares,
				state.Participants[state.Account.Address].Commitments,
			)
			assert.Nil(t, err)
		}

	}

	disputeShareDistributionTasks := make([]*dkgtasks.DisputeShareDistributionTask, n)
	keyshareSubmissionTasks := make([]*dkgtasks.KeyshareSubmissionTask, n)
	disputeMissingKeySharesTasks := make([]*dkgtasks.DisputeMissingKeySharesTask, n)

	if len(undistributedSharesIdx) == 0 {
		height, err := suite.eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)
		var dispShareDistStartBlock uint64

		// this means all validators distributed their shares and now the phase is
		// set phase to DisputeShareDistribution
		for i := 0; i < n; i++ {
			disputeShareDistributionTask, dispShareStartBlock, _, keyshareSubmissionTask, _, _, disputeMissingKeySharesTask, _, _ := dkgevents.UpdateStateOnShareDistributionComplete(suite.dkgStates[i], logger, height)

			dispShareDistStartBlock = dispShareStartBlock

			disputeShareDistributionTasks[i] = disputeShareDistributionTask
			keyshareSubmissionTasks[i] = keyshareSubmissionTask
			disputeMissingKeySharesTasks[i] = disputeMissingKeySharesTask
		}

		suite.disputeShareDistTasks = disputeShareDistributionTasks
		suite.keyshareSubmissionTasks = keyshareSubmissionTasks
		suite.disputeMissingKeyshareTasks = disputeMissingKeySharesTasks

		// skip all the way to DisputeShareDistribution phase
		advanceTo(t, suite.eth, dispShareDistStartBlock)
	} else {
		// this means some validators did not distribute shares, and the next phase is DisputeMissingShareDistribution
		advanceTo(t, suite.eth, suite.dkgStates[0].PhaseStart+suite.dkgStates[0].PhaseLength)
	}

	return suite
}

func StartFromKeyShareSubmissionPhase(t *testing.T, n int, undistributedShares int, phaseLength uint16) *TestSuite {
	suite := StartFromShareDistributionPhase(t, n, []int{}, []int{}, phaseLength)
	//accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")

	keyshareSubmissionStartBlock := suite.keyshareSubmissionTasks[0].Start
	advanceTo(t, suite.eth, keyshareSubmissionStartBlock)

	// Do key share submission task
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]

		if idx >= n-undistributedShares {
			continue
		}

		keyshareSubmissionTask := suite.keyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)

		err = keyshareSubmissionTask.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, keyshareSubmissionTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			suite.dkgStates[j].OnKeyShareSubmitted(
				state.Account.Address,
				state.Participants[state.Account.Address].KeyShareG1s,
				state.Participants[state.Account.Address].KeyShareG1CorrectnessProofs,
				state.Participants[state.Account.Address].KeyShareG2s,
			)
		}
	}

	mpkSubmissionTasks := make([]*dkgtasks.MPKSubmissionTask, n)

	if undistributedShares == 0 {
		// at this point all the validators submitted their key shares
		height, err := suite.eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)

		// this means all validators submitted their respective key shares and now the phase is
		// set phase to MPK
		var mpkSubmissionTaskStart uint64
		for i := 0; i < n; i++ {
			mpkSubmissionTask, taskStart, _ := dkgevents.UpdateStateOnKeyShareSubmissionComplete(suite.dkgStates[i], logger, height)
			mpkSubmissionTaskStart = taskStart

			mpkSubmissionTasks[i] = mpkSubmissionTask
		}

		// skip all the way to MPKSubmission phase
		advanceTo(t, suite.eth, mpkSubmissionTaskStart)
	} else {
		// this means some validators did not submit key shares, and the next phase is DisputeMissingKeyShares
		advanceTo(t, suite.eth, suite.dkgStates[0].PhaseStart+suite.dkgStates[0].PhaseLength)
	}

	suite.mpkSubmissionTasks = mpkSubmissionTasks

	return suite
}

func StartFromMPKSubmissionPhase(t *testing.T, n int, phaseLength uint16) *TestSuite {
	suite := StartFromKeyShareSubmissionPhase(t, n, 0, phaseLength)
	//accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")
	dkgStates := suite.dkgStates
	eth := suite.eth

	// Do MPK Submission task (once is enough)
	task := suite.mpkSubmissionTasks[0]
	state := dkgStates[0]

	err := task.Initialize(ctx, logger, eth, state)
	assert.Nil(t, err)
	err = task.DoWork(ctx, logger, eth)
	assert.Nil(t, err)

	eth.Commit()

	height, err := suite.eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	// advanceTo(t, eth, height+dkgStates[0].ConfirmationLength)

	gpkjSubmissionTasks := make([]*dkgtasks.GPKjSubmissionTask, n)
	disputeMissingGPKjTasks := make([]*dkgtasks.DisputeMissingGPKjTask, n)
	disputeGPKjTasks := make([]*dkgtasks.DisputeGPKjTask, n)

	for idx := 0; idx < n; idx++ {
		state := dkgStates[idx]
		gpkjSubmissionTask, _, _, disputeMissingGPKjTask, disputeGPKjTask, _, _ := dkgevents.UpdateStateOnMPKSet(state, logger, height, new(adminHandlerMock))

		gpkjSubmissionTasks[idx] = gpkjSubmissionTask
		disputeMissingGPKjTasks[idx] = disputeMissingGPKjTask
		disputeGPKjTasks[idx] = disputeGPKjTask
	}

	suite.gpkjSubmissionTasks = gpkjSubmissionTasks
	suite.disputeMissingGPKjTasks = disputeMissingGPKjTasks
	suite.disputeGPKjTasks = disputeGPKjTasks

	return suite
}

func StartFromGPKjPhase(t *testing.T, n int, undistributedGPKjIdx []int, badGPKjIdx []int, phaseLength uint16) *TestSuite {
	suite := StartFromMPKSubmissionPhase(t, n, phaseLength)
	//accounts := suite.eth.GetKnownAccounts()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")

	// Do GPKj Submission task
	for idx := 0; idx < n; idx++ {
		state := suite.dkgStates[idx]

		var skipLoop = false

		for _, undistIdx := range undistributedGPKjIdx {
			if idx == undistIdx {
				skipLoop = true
			}
		}

		if skipLoop {
			continue
		}

		gpkjSubTask := suite.gpkjSubmissionTasks[idx]

		err := gpkjSubTask.Initialize(ctx, logger, suite.eth, state)
		assert.Nil(t, err)

		for _, badIdx := range badGPKjIdx {
			if idx == badIdx {
				// inject bad shares
				// mess up with group private key (gskj)
				gskjBad := new(big.Int).Add(state.GroupPrivateKey, big.NewInt(1))
				// here's the group public key
				gpkj := new(cloudflare.G2).ScalarBaseMult(gskjBad)
				gpkjBad, err := bn256.G2ToBigIntArray(gpkj)
				assert.Nil(t, err)

				state.GroupPrivateKey = gskjBad
				state.Participants[state.Account.Address].GPKj = gpkjBad
			}
		}

		err = gpkjSubTask.DoWork(ctx, logger, suite.eth)
		assert.Nil(t, err)

		suite.eth.Commit()
		assert.True(t, gpkjSubTask.Success)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			suite.dkgStates[j].OnGPKjSubmitted(
				state.Account.Address,
				state.Participants[state.Account.Address].GPKj,
			)
		}

	}

	disputeGPKjTasks := make([]*dkgtasks.DisputeGPKjTask, n)
	completionTasks := make([]*dkgtasks.CompletionTask, n)

	if len(undistributedGPKjIdx) == 0 {
		height, err := suite.eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)
		var dispGPKjStartBlock uint64

		// this means all validators submitted their GPKjs and now the phase is
		// set phase to DisputeGPKjDistribution
		for i := 0; i < n; i++ {
			disputeGPKjTask, disputeGPKjStartBlock, _, completionTask, _, _ := dkgevents.UpdateStateOnGPKJSubmissionComplete(suite.dkgStates[i], logger, height)

			dispGPKjStartBlock = disputeGPKjStartBlock

			disputeGPKjTasks[i] = disputeGPKjTask
			completionTasks[i] = completionTask
		}

		suite.disputeGPKjTasks = disputeGPKjTasks
		suite.completionTasks = completionTasks

		// skip all the way to DisputeGPKj phase
		advanceTo(t, suite.eth, dispGPKjStartBlock)
	} else {
		// this means some validators did not submit their GPKjs, and the next phase is DisputeMissingGPKj
		advanceTo(t, suite.eth, suite.dkgStates[0].PhaseStart+suite.dkgStates[0].PhaseLength)
	}

	return suite
}
