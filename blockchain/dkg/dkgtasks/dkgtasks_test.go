package dkgtasks_test

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/monitor"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const SETUP_GROUP int = 13

func connectSimulatorEndpoint(t *testing.T, accountAddresses []string) blockchain.Ethereum {
	eth, err := blockchain.NewEthereumSimulator(
		"../../../assets/test/keys",
		"../../../assets/test/passcodes.txt",
		6,
		1*time.Second,
		5*time.Second,
		0,
		big.NewInt(9223372036854775807),
		accountAddresses...)

	assert.Nil(t, err, "Failed to build Ethereum endpoint...")
	assert.True(t, eth.IsEthereumAccessible(), "Web3 endpoint is not available.")

	// Mine a block once a second
	go func() {
		for true {
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
		txn, err := c.StakingToken.Transfer(txnOpts, acct.Address, big.NewInt(10_000_000))
		assert.Nilf(t, err, "Failed on transfer %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)

		o, err := eth.GetTransactionOpts(ctx, acct)
		assert.Nil(t, err)

		// 2. Allow system to take tokens from 'acct' for staking
		txn, err = c.StakingToken.Approve(o, c.ValidatorsAddress, big.NewInt(10_000_000))
		assert.Nilf(t, err, "Failed on approval %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)

		// 3. Tell system to take tokens from 'acct' for staking
		txn, err = c.Staking.LockStake(o, big.NewInt(1_000_000))
		assert.Nilf(t, err, "Failed on lock %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)

		// 4. Tell system 'acct' wants to join as validator
		var validatorId [2]*big.Int
		validatorId[0] = big.NewInt(int64(idx))
		validatorId[1] = big.NewInt(int64(idx * 2))

		txn, err = c.Validators.AddValidator(o, acct.Address, validatorId)
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

func connectRemoteEndpoint(t *testing.T, accountAddresses []string) blockchain.Ethereum {
	eth, err := blockchain.NewEthereumEndpoint(
		"http://192.168.86.29:8545",
		"keystore_test",
		"assets_test/passcodes.txt",
		accountAddresses[0],
		3*time.Second, // This is the timeout for blocking actions
		30,            // Let's do lots of retries
		1*time.Second, // This is the retry delay
		2)             // For testing finality is 2 blocks
	assert.Nil(t, err)

	return eth
}

func validator(t *testing.T, idx int, eth blockchain.Ethereum, validatorAcct accounts.Account, wg *sync.WaitGroup) {
	name := fmt.Sprintf("validator%2d", idx)
	logger := logging.GetLogger(name)
	c := eth.Contracts()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startBlock := uint64(1)

	currentBlock := uint64(startBlock)
	lastBlock := uint64(startBlock)
	addresses := []common.Address{c.EthdkgAddress}

	taskManager := tasks.NewManager()

	state := dkg.NewEthDKGState(validatorAcct)

	scheduler := monitor.NewSequentialSchedule()
	defer scheduler.Status(logging.GetLogger(fmt.Sprintf("scheduler%v", idx)))

	for currentBlock < startBlock+50 {
		var err error
		currentBlock, err = eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)

		if currentBlock != lastBlock {
			logger.Infof("Block %d -> %d", lastBlock, currentBlock)

			logs, err := eth.GetEvents(ctx, lastBlock+1, currentBlock, addresses)
			assert.Nil(t, err)

			for _, log := range logs {

				// RegistrationOpen
				if log.Topics[0] == common.HexToHash("0x9c6f8368fe7e77e8cb9438744581403bcb3f53298e517f04c1b8475487402e97") {
					event, err := c.Ethdkg.ParseRegistrationOpen(log)
					assert.Nil(t, err)

					logger.Debugf("open event:%+v", event)

					scheduler.Purge()

					state.PopulateSchedule(event) // TODO make this better, pure mutation functions are awkward

					scheduler.Schedule(state.RegistrationStart, state.RegistrationEnd, dkgtasks.NewRegisterTask(state))                    // Registration
					scheduler.Schedule(state.ShareDistributionStart, state.ShareDistributionEnd, dkgtasks.NewShareDistributionTask(state)) // ShareDistribution
					scheduler.Schedule(state.DisputeStart, state.DisputeEnd, dkgtasks.NewDisputeTask(state))                               // DisputeShares
					scheduler.Schedule(state.KeyShareSubmissionStart, state.KeyShareSubmissionEnd, dkgtasks.NewPlaceHolder(state))         // KeyShareSubmission
					scheduler.Schedule(state.MPKSubmissionStart, state.MPKSubmissionEnd, dkgtasks.NewPlaceHolder(state))                   // MasterPublicKeySubmission
					scheduler.Schedule(state.GPKJSubmissionStart, state.GPKJSubmissionEnd, dkgtasks.NewPlaceHolder(state))                 // GroupPublicKeySubmission
					scheduler.Schedule(state.GPKJGroupAccusationStart, state.GPKJGroupAccusationEnd, dkgtasks.NewPlaceHolder(state))       // DisputeGroupPublicKey
					scheduler.Schedule(state.CompleteStart, state.CompleteEnd, dkgtasks.NewPlaceHolder(state))                             // Complete

					logger.Debugf("ethdkg state:%+v", state)
				}
			}

			for block := lastBlock + 1; block <= currentBlock; block++ {
				uuid, _ := scheduler.Find(block)
				if uuid != nil {
					task, _ := scheduler.Retrieve(uuid)
					handler := taskManager.NewTaskHandler(logger, eth, task)

					handler.Start()

					scheduler.Remove(uuid)
				}
			}

			lastBlock = currentBlock
		}

		time.Sleep(time.Second)
	}

	wg.Done()
}

func TestRegisterSuccess(t *testing.T) {

	var accountAddresses []string = []string{
		"0x546F99F244b7B58B855330AE0E2BC1b30b41302F", "0x9AC1c9afBAec85278679fF75Ef109217f26b1417",
		"0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac", "0x615695C4a4D6a60830e5fca4901FbA099DF26271",
		"0x63a6627b79813A7A43829490C4cE409254f64177"}

	logging.GetLogger("ethsim").SetLevel(logrus.InfoLevel)

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

	t.Logf("  ethdkg address: %v", c.EthdkgAddress.Hex())
	t.Logf("registry address: %v", c.RegistryAddress.Hex())

	callOpts := eth.GetCallOpts(ctx, ownerAccount)
	txnOpts, err := eth.GetTransactionOpts(context.Background(), ownerAccount)
	assert.Nil(t, err)

	// Start validators running
	wg := sync.WaitGroup{}
	for i := 0; i < 4; i++ {
		acct, err := eth.GetAccount(common.HexToAddress(accountAddresses[i+1]))
		assert.Nil(t, err)

		wg.Add(1)
		go validator(t, i, eth, acct, &wg)
	}

	// Kick off a round of ethdkg
	txn, err := c.Ethdkg.UpdatePhaseLength(txnOpts, big.NewInt(5))
	assert.Nil(t, err)
	eth.Queue().QueueAndWait(ctx, txn)

	txn, err = c.Ethdkg.InitializeState(txnOpts)
	assert.Nil(t, err)
	eth.Queue().QueueAndWait(ctx, txn)

	// Now we know ethdkg is running, let's find out when registration has to happen
	// TODO this should be based on an OpenRegistration event
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	t.Logf("currentHeight:%v", currentHeight)

	endingHeight, err := c.Ethdkg.TREGISTRATIONEND(callOpts)
	assert.Nil(t, err)
	t.Logf("endingHeight:%v", endingHeight)

	// Wait for validators to complete
	wg.Wait()

}
