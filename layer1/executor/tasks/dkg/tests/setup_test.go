package tests

import (
	"context"
	"errors"
	"fmt"
	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/crypto/bn256"
	"github.com/alicenet/alicenet/crypto/bn256/cloudflare"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	testUtils "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/tests/utils"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/utils"
	"github.com/alicenet/alicenet/layer1/monitor/events"
	"github.com/alicenet/alicenet/layer1/monitor/objects"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"math/big"
	"os"
	"strings"
	"testing"
)

var HardHat *tests.Hardhat

func TestMain(m *testing.M) {
	hardhat, err := tests.StartHardHatNodeWithDefaultHost()
	if err != nil {
		panic(err)
	}
	HardHat = hardhat
	code := m.Run()
	hardhat.Close()
	os.Exit(code)
}

func setupEthereum(t *testing.T, n int) *tests.ClientFixture {
	logger := logging.GetLogger("test").WithField("test", t.Name())
	fixture := tests.NewClientFixture(HardHat, 0, n, logger, true, true, true)
	assert.NotNil(t, fixture)

	eth := fixture.Client
	assert.NotNil(t, eth)
	assert.Equal(t, n, len(eth.GetKnownAccounts()))

	t.Cleanup(func() {
		fixture.Close()
		ethereum.CleanGlobalVariables(t)
	})

	return fixture
}

type TestSuite struct {
	Eth                          layer1.Client
	DKGStatesDbs                 []*db.Database
	regTasks                     []*dkg.RegisterTask
	DispMissingRegTasks          []*dkg.DisputeMissingRegistrationTask
	ShareDistTasks               []*dkg.ShareDistributionTask
	DisputeMissingShareDistTasks []*dkg.DisputeMissingShareDistributionTask
	DisputeShareDistTasks        [][]*dkg.DisputeShareDistributionTask
	KeyshareSubmissionTasks      []*dkg.KeyShareSubmissionTask
	DisputeMissingKeyshareTasks  []*dkg.DisputeMissingKeySharesTask
	MpkSubmissionTasks           []*dkg.MPKSubmissionTask
	GpkjSubmissionTasks          []*dkg.GPKjSubmissionTask
	DisputeMissingGPKjTasks      []*dkg.DisputeMissingGPKjTask
	DisputeGPKjTasks             [][]*dkg.DisputeGPKjTask
	CompletionTasks              []*dkg.CompletionTask
	BadAddresses                 map[common.Address]bool
}

func GetDKGDb(t *testing.T) *db.Database {
	db := mocks.NewTestDB()
	t.Cleanup(func() {
		db.DB().Close()
	})
	return db
}

func SubscribeAndWaitReceipt(ctx context.Context, fixture *tests.ClientFixture, txn *types.Transaction) (*types.Receipt, error) {
	rcptResponse, err := fixture.Watcher.Subscribe(ctx, txn, nil)
	if err != nil {
		return nil, err
	}

	tests.MineFinalityDelayBlocks(fixture.Client)

	rcpt, err := rcptResponse.GetReceiptBlocking(ctx)
	if err != nil {
		return nil, err
	}
	if rcpt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("receipt status indicate failure: %v", rcpt.Status)
	}

	return rcpt, nil
}

func SetETHDKGPhaseLength(length uint16, fixture *tests.ClientFixture, callOpts *bind.TransactOpts, ctx context.Context) (*types.Transaction, *types.Receipt, error) {
	// Shorten ethdkg phase for testing purposes
	ethdkgABI, err := abi.JSON(strings.NewReader(bindings.ETHDKGMetaData.ABI))
	if err != nil {
		return nil, nil, err
	}

	input, err := ethdkgABI.Pack("setPhaseLength", uint16(length))
	if err != nil {
		return nil, nil, err
	}

	txn, err := ethereum.GetContracts().ContractFactory().CallAny(callOpts, ethereum.GetContracts().EthdkgAddress(), big.NewInt(0), input)
	if err != nil {
		return nil, nil, err
	}
	if txn == nil {
		return nil, nil, errors.New("non existent transaction ContractFactory.CallAny(ethdkg, setPhaseLength(...))")
	}

	rcpt, err := SubscribeAndWaitReceipt(ctx, fixture, txn)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting receipt for ContractFactory.CallAny(ethdkg, setPhaseLength(...)) err: %v", err)
	}
	return txn, rcpt, nil
}

func InitializeETHDKG(fixture *tests.ClientFixture, callOpts *bind.TransactOpts, ctx context.Context) (*types.Transaction, *types.Receipt, error) {
	// Shorten ethdkg phase for testing purposes
	validatorPoolABI, err := abi.JSON(strings.NewReader(bindings.ValidatorPoolMetaData.ABI))
	if err != nil {
		return nil, nil, err
	}

	input, err := validatorPoolABI.Pack("initializeETHDKG")
	if err != nil {
		return nil, nil, err
	}

	txn, err := ethereum.GetContracts().ContractFactory().CallAny(callOpts, ethereum.GetContracts().ValidatorPoolAddress(), big.NewInt(0), input)
	if err != nil {
		return nil, nil, err
	}
	if txn == nil {
		return nil, nil, errors.New("non existent transaction ContractFactory.CallAny(validatorPool, initializeETHDKG())")
	}

	rcpt, err := SubscribeAndWaitReceipt(ctx, fixture, txn)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting receipt for ContractFactory.CallAny(validatorPool, initializeETHDKG()) err: %v", err)
	}

	return txn, rcpt, nil
}

func StartFromRegistrationOpenPhase(t *testing.T, fixture *tests.ClientFixture, unregisteredValidators int, phaseLength uint16) *TestSuite {

	eth := fixture.Client
	ctx := context.Background()
	owner := eth.GetDefaultAccount()

	// Start EthDKG
	ownerOpts, err := eth.GetTransactionOpts(ctx, owner)
	assert.Nil(t, err)

	// Shorten ethdkg phase for testing purposes
	_, _, err = SetETHDKGPhaseLength(phaseLength, fixture, ownerOpts, ctx)
	assert.Nil(t, err)

	// init ETHDKG on ValidatorPool, through ContractFactory
	_, rcpt, err := InitializeETHDKG(fixture, ownerOpts, ctx)
	assert.Nil(t, err)

	event, err := testUtils.GetETHDKGRegistrationOpened(rcpt.Logs, eth)
	assert.Nil(t, err)
	assert.NotNil(t, event)

	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err)
	var validatorAddresses []common.Address
	// all known addresses must be validators at this point
	for _, acc := range eth.GetKnownAccounts() {
		validatorAddresses = append(validatorAddresses, acc.Address)
	}

	phase, err := ethereum.GetContracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint8(state.RegistrationOpen), phase)

	n := len(validatorAddresses)
	valCount, err := ethereum.GetContracts().ValidatorPool().GetValidatorsCount(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, uint64(n), valCount.Uint64())

	// Do Register task
	regTasks := make([]*dkg.RegisterTask, n)
	dispMissingRegTasks := make([]*dkg.DisputeMissingRegistrationTask, n)
	dkgStatesDbs := make([]*db.Database, n)
	accounts := eth.GetKnownAccounts()
	var receiptResponses []transaction.ReceiptResponse

	for idx := 0; idx < n; idx++ {
		logger := logging.GetLogger("test").WithField("Validator", accounts[idx].Address.String())
		// Set Registration success to true
		dkgState, regTask, dispMissingRegTask := events.UpdateStateOnRegistrationOpened(
			accounts[idx],
			event.StartBlock.Uint64(),
			event.PhaseLength.Uint64(),
			event.ConfirmationLength.Uint64(),
			event.Nonce.Uint64(),
			true,
			validatorAddresses,
		)

		dkgDb := GetDKGDb(t)
		dkgStatesDbs[idx] = dkgDb
		err := state.SaveDkgState(dkgDb, dkgState)
		assert.Nil(t, err)
		regTasks[idx] = regTask
		dispMissingRegTasks[idx] = dispMissingRegTask

		err = regTasks[idx].Initialize(ctx, nil, dkgDb, logger, eth, "RegistrationTask", fmt.Sprintf("%v", idx), nil)
		assert.Nil(t, err)
		err = regTasks[idx].Prepare(ctx)
		assert.Nil(t, err)

		if idx >= n-unregisteredValidators {
			continue
		}

		txn, err := regTasks[idx].Execute(ctx)
		assert.Nil(t, err)

		rcptResponse, err := fixture.Watcher.Subscribe(ctx, txn, nil)
		assert.Nil(t, err)
		receiptResponses = append(receiptResponses, rcptResponse)
	}

	tests.WaitGroupReceipts(t, eth, receiptResponses)

	// simulate receiving AddressRegistered event
	for i := 0; i < n; i++ {
		dkgState, err := state.GetDkgState(dkgStatesDbs[i])
		assert.Nil(t, err)

		if i >= n-unregisteredValidators {
			continue
		}

		for j := 0; j < n; j++ {
			otherDkgState, err := state.GetDkgState(dkgStatesDbs[j])
			assert.Nil(t, err)
			otherDkgState.OnAddressRegistered(dkgState.Account.Address, i+1, dkgState.Nonce, dkgState.TransportPublicKey)
			err = state.SaveDkgState(dkgStatesDbs[j], otherDkgState)
			assert.Nil(t, err)
		}
	}

	shareDistributionTasks := make([]*dkg.ShareDistributionTask, n)
	disputeMissingShareDistributionTasks := make([]*dkg.DisputeMissingShareDistributionTask, n)
	disputeShareDistTasks := make([][]*dkg.DisputeShareDistributionTask, n)

	if unregisteredValidators == 0 {
		height, err := eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)

		for idx := 0; idx < n; idx++ {
			dkgState, err := state.GetDkgState(dkgStatesDbs[idx])
			assert.Nil(t, err)
			shareDistributionTask, disputeMissingShareDistributionTask, disputeShareDistTask := events.UpdateStateOnRegistrationComplete(dkgState, height)
			shareDistributionTasks[idx] = shareDistributionTask
			disputeMissingShareDistributionTasks[idx] = disputeMissingShareDistributionTask
			disputeShareDistTasks[idx] = disputeShareDistTask
			err = state.SaveDkgState(dkgStatesDbs[idx], dkgState)
			assert.Nil(t, err)
		}

		// skip all the way to ShareDistribution phase
		tests.AdvanceTo(eth, shareDistributionTasks[0].Start)
	} else {
		dkgState, err := state.GetDkgState(dkgStatesDbs[0])
		assert.Nil(t, err)
		// this means some validators did not register, and the next phase is DisputeMissingRegistration
		tests.AdvanceTo(eth, dkgState.PhaseStart+dkgState.PhaseLength)
	}

	return &TestSuite{
		Eth:                          eth,
		DKGStatesDbs:                 dkgStatesDbs,
		regTasks:                     regTasks,
		DispMissingRegTasks:          dispMissingRegTasks,
		ShareDistTasks:               shareDistributionTasks,
		DisputeMissingShareDistTasks: disputeMissingShareDistributionTasks,
		DisputeShareDistTasks:        disputeShareDistTasks,
	}
}

func StartFromShareDistributionPhase(t *testing.T, fixture *tests.ClientFixture, undistributedSharesIdx []int, badSharesIdx []int, phaseLength uint16) *TestSuite {
	suite := StartFromRegistrationOpenPhase(t, fixture, 0, phaseLength)
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")
	n := len(suite.Eth.GetKnownAccounts())

	callOpts, err := suite.Eth.GetCallOpts(ctx, suite.Eth.GetDefaultAccount())
	assert.Nil(t, err)
	phase, err := ethereum.GetContracts().Ethdkg().GetETHDKGPhase(callOpts)
	assert.Nil(t, err)
	assert.Equal(t, phase, uint8(state.ShareDistribution))

	height, err := suite.Eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, height, suite.ShareDistTasks[0].Start)
	var receiptResponses []transaction.ReceiptResponse
	// Do Share Distribution task
	for idx := 0; idx < n; idx++ {
		var skipLoop = false

		for _, undistIdx := range undistributedSharesIdx {
			if idx == undistIdx {
				skipLoop = true
			}
		}

		if skipLoop {
			continue
		}

		shareDistTask := suite.ShareDistTasks[idx]

		err = shareDistTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], logger, suite.Eth, "ShareDistributionTask", fmt.Sprintf("%v", idx), nil)
		assert.Nil(t, err)
		err = shareDistTask.Prepare(ctx)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)

		if len(badSharesIdx) > 0 {
			suite.BadAddresses = make(map[common.Address]bool)
			for _, badIdx := range badSharesIdx {
				accounts := suite.Eth.GetKnownAccounts()
				// inject bad shares
				for _, s := range dkgState.Participants[accounts[badIdx].Address].EncryptedShares {
					s.Set(big.NewInt(0))
				}
				err = state.SaveDkgState(suite.DKGStatesDbs[idx], dkgState)
				assert.Nil(t, err)
				suite.BadAddresses[accounts[badIdx].Address] = true
			}
		}

		txn, err := shareDistTask.Execute(ctx)
		assert.Nil(t, err)

		rcptResponse, err := fixture.Watcher.Subscribe(ctx, txn, nil)
		assert.Nil(t, err)
		receiptResponses = append(receiptResponses, rcptResponse)

		// event
		for j := 0; j < n; j++ {
			// simulate receiving event for all participants
			participantDkgState, err := state.GetDkgState(suite.DKGStatesDbs[j])
			assert.Nil(t, err)

			err = participantDkgState.OnSharesDistributed(
				logger,
				dkgState.Account.Address,
				dkgState.Participants[dkgState.Account.Address].EncryptedShares,
				dkgState.Participants[dkgState.Account.Address].Commitments,
			)
			assert.Nil(t, err)
			err = state.SaveDkgState(suite.DKGStatesDbs[j], participantDkgState)
			assert.Nil(t, err)
		}

	}

	tests.WaitGroupReceipts(t, suite.Eth, receiptResponses)

	disputeShareDistributionTasks := make([][]*dkg.DisputeShareDistributionTask, n)
	keyshareSubmissionTasks := make([]*dkg.KeyShareSubmissionTask, n)
	disputeMissingKeySharesTasks := make([]*dkg.DisputeMissingKeySharesTask, n)

	if len(undistributedSharesIdx) == 0 {
		height, err := suite.Eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)
		var dispShareDistStartBlock uint64

		// this means all validators distributed their shares and now the phase is
		// set phase to DisputeShareDistribution
		for i := 0; i < n; i++ {
			dkgState, err := state.GetDkgState(suite.DKGStatesDbs[i])
			assert.Nil(t, err)
			disputeShareDistributionTask, keyshareSubmissionTask, disputeMissingKeySharesTask := events.UpdateStateOnShareDistributionComplete(dkgState, height)

			dispShareDistStartBlock = disputeShareDistributionTask[0].GetStart()

			disputeShareDistributionTasks[i] = disputeShareDistributionTask
			keyshareSubmissionTasks[i] = keyshareSubmissionTask
			disputeMissingKeySharesTasks[i] = disputeMissingKeySharesTask
			err = state.SaveDkgState(suite.DKGStatesDbs[i], dkgState)
			assert.Nil(t, err)
		}

		suite.DisputeShareDistTasks = disputeShareDistributionTasks
		suite.KeyshareSubmissionTasks = keyshareSubmissionTasks
		suite.DisputeMissingKeyshareTasks = disputeMissingKeySharesTasks

		// skip all the way to DisputeShareDistribution phase
		tests.AdvanceTo(suite.Eth, dispShareDistStartBlock)
	} else {
		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[0])
		assert.Nil(t, err)
		// this means some validators did not distribute shares, and the next phase is DisputeMissingShareDistribution
		tests.AdvanceTo(suite.Eth, dkgState.PhaseStart+dkgState.PhaseLength)
	}

	return suite
}

func StartFromKeyShareSubmissionPhase(t *testing.T, fixture *tests.ClientFixture, undistributedShares int, phaseLength uint16) *TestSuite {
	suite := StartFromShareDistributionPhase(t, fixture, []int{}, []int{}, phaseLength)
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")
	n := len(suite.Eth.GetKnownAccounts())

	keyshareSubmissionStartBlock := suite.KeyshareSubmissionTasks[0].Start
	tests.AdvanceTo(suite.Eth, keyshareSubmissionStartBlock)
	var receiptResponses []transaction.ReceiptResponse
	// Do key share submission task
	for idx := 0; idx < n; idx++ {
		if idx >= n-undistributedShares {
			continue
		}

		keyshareSubmissionTask := suite.KeyshareSubmissionTasks[idx]

		err := keyshareSubmissionTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], logger, suite.Eth, "KeyShareSubmissionTask", fmt.Sprintf("%v", idx), nil)
		assert.Nil(t, err)
		err = keyshareSubmissionTask.Prepare(ctx)
		assert.Nil(t, err)

		txn, err := keyshareSubmissionTask.Execute(ctx)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)

		rcptResponse, err := fixture.Watcher.Subscribe(ctx, txn, nil)
		assert.Nil(t, err)
		receiptResponses = append(receiptResponses, rcptResponse)

		// event
		for j := 0; j < n; j++ {
			participantDkgState, err := state.GetDkgState(suite.DKGStatesDbs[j])
			assert.Nil(t, err)
			// simulate receiving event for all participants
			participantDkgState.OnKeyShareSubmitted(
				dkgState.Account.Address,
				dkgState.Participants[dkgState.Account.Address].KeyShareG1s,
				dkgState.Participants[dkgState.Account.Address].KeyShareG1CorrectnessProofs,
				dkgState.Participants[dkgState.Account.Address].KeyShareG2s,
			)
			err = state.SaveDkgState(suite.DKGStatesDbs[j], participantDkgState)
			assert.Nil(t, err)
		}
	}

	tests.WaitGroupReceipts(t, suite.Eth, receiptResponses)

	mpkSubmissionTasks := make([]*dkg.MPKSubmissionTask, n)

	if undistributedShares == 0 {
		// at this point all the validators submitted their key shares
		height, err := suite.Eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)

		// this means all validators submitted their respective key shares and now the phase is
		// set phase to MPK
		var mpkSubmissionTaskStart uint64
		for i := 0; i < n; i++ {
			dkgState, err := state.GetDkgState(suite.DKGStatesDbs[i])
			assert.Nil(t, err)
			mpkSubmissionTask := events.UpdateStateOnKeyShareSubmissionComplete(dkgState, height)
			mpkSubmissionTaskStart = mpkSubmissionTask.GetStart()

			mpkSubmissionTasks[i] = mpkSubmissionTask
			err = state.SaveDkgState(suite.DKGStatesDbs[i], dkgState)
			assert.Nil(t, err)
		}

		// skip all the way to MPKSubmission phase
		tests.AdvanceTo(suite.Eth, mpkSubmissionTaskStart)
	} else {
		// this means some validators did not submit key shares, and the next phase is DisputeMissingKeyShares
		tests.AdvanceTo(suite.Eth, suite.DisputeMissingKeyshareTasks[0].Start)
	}

	suite.MpkSubmissionTasks = mpkSubmissionTasks

	return suite
}

func StartFromMPKSubmissionPhase(t *testing.T, fixture *tests.ClientFixture, phaseLength uint16) *TestSuite {
	suite := StartFromKeyShareSubmissionPhase(t, fixture, 0, phaseLength)
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")
	n := len(suite.Eth.GetKnownAccounts())

	// Do MPK Submission task (once is enough)
	var receiptResponses []transaction.ReceiptResponse
	for idx := 0; idx < n; idx++ {
		task := suite.MpkSubmissionTasks[idx]
		err := task.Initialize(ctx, nil, suite.DKGStatesDbs[idx], fixture.Logger, suite.Eth, "MPKSubmissionTask", fmt.Sprintf("%v", idx), nil)
		assert.Nil(t, err)
		err = task.Prepare(ctx)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)
		if utils.AmILeading(suite.Eth, ctx, logger, int(task.GetStart()), task.StartBlockHash[:], n, dkgState.Index) {
			txn, err := task.Execute(ctx)
			assert.Nil(t, err)

			rcptResponse, subsErr := fixture.Watcher.Subscribe(ctx, txn, nil)
			assert.Nil(t, subsErr)
			receiptResponses = append(receiptResponses, rcptResponse)
		}
	}

	tests.WaitGroupReceipts(t, suite.Eth, receiptResponses)

	height, err := suite.Eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	gpkjSubmissionTasks := make([]*dkg.GPKjSubmissionTask, n)
	disputeMissingGPKjTasks := make([]*dkg.DisputeMissingGPKjTask, n)
	disputeGPKjTasks := make([][]*dkg.DisputeGPKjTask, n)

	for idx := 0; idx < n; idx++ {
		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)
		gpkjSubmissionTask, disputeMissingGPKjTask, disputeGPKjTask := events.UpdateStateOnMPKSet(dkgState, height, mocks.NewMockAdminHandler())

		gpkjSubmissionTasks[idx] = gpkjSubmissionTask
		disputeMissingGPKjTasks[idx] = disputeMissingGPKjTask
		disputeGPKjTasks[idx] = disputeGPKjTask
		err = state.SaveDkgState(suite.DKGStatesDbs[idx], dkgState)
		assert.Nil(t, err)
	}

	suite.GpkjSubmissionTasks = gpkjSubmissionTasks
	suite.DisputeMissingGPKjTasks = disputeMissingGPKjTasks
	suite.DisputeGPKjTasks = disputeGPKjTasks

	return suite
}

func StartFromGPKjPhase(t *testing.T, fixture *tests.ClientFixture, undistributedGPKjIdx []int, badGPKjIdx []int, phaseLength uint16) *TestSuite {
	suite := StartFromMPKSubmissionPhase(t, fixture, phaseLength)
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("Validator", "")
	n := len(suite.Eth.GetKnownAccounts())
	var receiptResponses []transaction.ReceiptResponse
	suite.BadAddresses = make(map[common.Address]bool)
	// Do GPKj Submission task
	for idx := 0; idx < n; idx++ {
		var skipLoop = false

		for _, undistIdx := range undistributedGPKjIdx {
			if idx == undistIdx {
				skipLoop = true
			}
		}

		if skipLoop {
			continue
		}

		gpkjSubTask := suite.GpkjSubmissionTasks[idx]

		err := gpkjSubTask.Initialize(ctx, nil, suite.DKGStatesDbs[idx], logger, suite.Eth, "GPKjSubmissionTask", fmt.Sprintf("%v", idx), nil)
		assert.Nil(t, err)
		err = gpkjSubTask.Prepare(ctx)
		assert.Nil(t, err)

		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)

		for _, badIdx := range badGPKjIdx {
			if idx == badIdx {
				// inject bad shares
				// mess up with group private key (gskj)
				gskjBad := new(big.Int).Add(dkgState.GroupPrivateKey, big.NewInt(1))
				// here's the group public key
				gpkj := new(cloudflare.G2).ScalarBaseMult(gskjBad)
				gpkjBad, err := bn256.G2ToBigIntArray(gpkj)
				assert.Nil(t, err)

				dkgState.GroupPrivateKey = gskjBad
				dkgState.Participants[dkgState.Account.Address].GPKj = gpkjBad
				err = state.SaveDkgState(suite.DKGStatesDbs[idx], dkgState)
				assert.Nil(t, err)
				suite.BadAddresses[dkgState.Account.Address] = true
			}
		}

		txn, err := gpkjSubTask.Execute(ctx)
		assert.Nil(t, err)

		rcptResponse, subsErr := fixture.Watcher.Subscribe(ctx, txn, nil)
		assert.Nil(t, subsErr)
		receiptResponses = append(receiptResponses, rcptResponse)

		// event
		for j := 0; j < n; j++ {
			participantDkgState, err := state.GetDkgState(suite.DKGStatesDbs[j])
			assert.Nil(t, err)
			// simulate receiving event for all participants
			participantDkgState.OnGPKjSubmitted(
				dkgState.Account.Address,
				dkgState.Participants[dkgState.Account.Address].GPKj,
			)
			err = state.SaveDkgState(suite.DKGStatesDbs[j], participantDkgState)
			assert.Nil(t, err)
		}

	}
	tests.WaitGroupReceipts(t, suite.Eth, receiptResponses)

	disputeGPKjTasks := make([][]*dkg.DisputeGPKjTask, n)
	completionTasks := make([]*dkg.CompletionTask, n)

	if len(undistributedGPKjIdx) == 0 {
		height, err := suite.Eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)
		var dispGPKjStartBlock uint64

		// this means all validators submitted their GPKjs and now the phase is
		// set phase to DisputeGPKjDistribution
		for i := 0; i < n; i++ {
			dkgState, err := state.GetDkgState(suite.DKGStatesDbs[i])
			assert.Nil(t, err)
			disputeGPKjTask, completionTask := events.UpdateStateOnGPKJSubmissionComplete(dkgState, height)

			dispGPKjStartBlock = disputeGPKjTask[0].GetStart()

			disputeGPKjTasks[i] = disputeGPKjTask
			completionTasks[i] = completionTask
			err = state.SaveDkgState(suite.DKGStatesDbs[i], dkgState)
			assert.Nil(t, err)
		}

		suite.DisputeGPKjTasks = disputeGPKjTasks
		suite.CompletionTasks = completionTasks

		// skip all the way to DisputeGPKj phase
		tests.AdvanceTo(suite.Eth, dispGPKjStartBlock)
	} else {
		// this means some validators did not submit their GPKjs, and the next phase is DisputeMissingGPKj
		tests.AdvanceTo(suite.Eth, suite.DisputeMissingGPKjTasks[0].Start)
	}

	return suite
}

func StartFromCompletion(t *testing.T, fixture *tests.ClientFixture, phaseLength uint16) *TestSuite {
	suite := StartFromGPKjPhase(t, fixture, []int{}, []int{}, phaseLength)
	dkgState, err := state.GetDkgState(suite.DKGStatesDbs[0])
	assert.Nil(t, err)
	// move to Completion phase
	tests.AdvanceTo(suite.Eth, suite.CompletionTasks[0].Start+dkgState.ConfirmationLength)

	return suite
}

func RegisterPotentialValidatorOnMonitor(t *testing.T, suite *TestSuite, accounts []accounts.Account) {
	monState := objects.NewMonitorState()
	for idx := 0; idx < len(accounts); idx++ {
		monState.PotentialValidators[accounts[idx].Address] = objects.PotentialValidator{
			Account: accounts[idx].Address,
		}
	}

	for idx := 0; idx < len(accounts); idx++ {
		err := monState.PersistState(suite.DKGStatesDbs[idx])
		assert.Nil(t, err)
	}
}

func CheckBadValidators(t *testing.T, badValidators []int, suite *TestSuite) {
	for _, badId := range badValidators {
		dkgState, err := state.GetDkgState(suite.DKGStatesDbs[badId])
		assert.Nil(t, err)

		callOpts, err := suite.Eth.GetCallOpts(context.Background(), dkgState.Account)
		assert.Nil(t, err)

		isValidator, err := ethereum.GetContracts().ValidatorPool().IsValidator(callOpts, dkgState.Account.Address)
		assert.Nil(t, err)
		assert.Equal(t, false, isValidator)
	}
}
