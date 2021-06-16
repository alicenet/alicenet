package dkgtasks_test

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgevents"
	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/monitor"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const SETUP_GROUP int = 13

func connectSimulatorEndpoint(t *testing.T, accountAddresses []string) interfaces.Ethereum {
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

func connectRemoteEndpoint(t *testing.T, accountAddresses []string) interfaces.Ethereum {
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

func validator(t *testing.T, idx int, eth interfaces.Ethereum, validatorAcct accounts.Account, wg *sync.WaitGroup) {
	name := fmt.Sprintf("validator%2d", idx)
	logger := logging.GetLogger(name)
	c := eth.Contracts()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startBlock := uint64(1)

	currentBlock := uint64(startBlock)
	lastBlock := uint64(startBlock)
	addresses := []common.Address{c.EthdkgAddress()}

	dkgState := objects.NewDkgState(validatorAcct)
	scheduler := monitor.NewSequentialSchedule()

	monitorState := &objects.MonitorState{EthDKG: dkgState, Schedule: scheduler}

	taskManager := tasks.NewManager()
	events := objects.NewEventMap()

	SetupEventMap(events)
	SetupTasks()

	var done bool
	var err error

	for !done {
		currentBlock, err = eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)

		if currentBlock != lastBlock {
			logger.Infof("Block %d -> %d", lastBlock, currentBlock)

			logs, err := eth.GetEvents(ctx, lastBlock+1, currentBlock, addresses)
			assert.Nil(t, err)

			// Check all the logs for an event we want to process
			for _, log := range logs {

				eventID := log.Topics[0].String()

				info, present := events.Lookup(eventID)
				if present {
					err := info.Processor(eth, logger, monitorState, log)
					if err != nil {
						logger.Errorf("Failed processing event: %v", err)
					}

				}

			}

			// Check if any tasks are scheduled
			for block := lastBlock + 1; block <= currentBlock; block++ {
				uuid, err := scheduler.Find(block)
				if err == nil {
					task, _ := scheduler.Retrieve(uuid)

					handler := taskManager.StartTask(logger, eth, task)
					logger.Infof("handler:%p", handler)

					scheduler.Remove(uuid)
				}
			}

			lastBlock = currentBlock
		}

		time.Sleep(time.Second)

		dkgState.RLock()
		done = dkgState.Complete
		dkgState.RUnlock()
	}

	wg.Done()
}

func TestDkgSuccess(t *testing.T) {

	var accountAddresses []string = []string{
		"0x546F99F244b7B58B855330AE0E2BC1b30b41302F", "0x9AC1c9afBAec85278679fF75Ef109217f26b1417",
		"0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac", "0x615695C4a4D6a60830e5fca4901FbA099DF26271",
		"0x63a6627b79813A7A43829490C4cE409254f64177", "0x16564cf3e880d9f5d09909f51b922941ebbbc24d"}

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

	callOpts := eth.GetCallOpts(ctx, ownerAccount)
	txnOpts, err := eth.GetTransactionOpts(context.Background(), ownerAccount)
	assert.Nil(t, err)

	// Start validators running
	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		acct, err := eth.GetAccount(common.HexToAddress(accountAddresses[i+1]))
		assert.Nil(t, err)

		wg.Add(1)
		go validator(t, i, eth, acct, &wg)
	}

	// Kick off a round of ethdkg
	txn, err := c.Ethdkg().UpdatePhaseLength(txnOpts, big.NewInt(2))
	assert.Nil(t, err)
	eth.Queue().QueueAndWait(ctx, txn)

	txn, err = c.Ethdkg().InitializeState(txnOpts)
	assert.Nil(t, err)
	eth.Queue().QueueAndWait(ctx, txn)

	// Now we know ethdkg is running, let's find out when registration has to happen
	// TODO this should be based on an OpenRegistration event
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	t.Logf("currentHeight:%v", currentHeight)

	endingHeight, err := c.Ethdkg().TREGISTRATIONEND(callOpts)
	assert.Nil(t, err)
	t.Logf("endingHeight:%v", endingHeight)

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

func SetupTasks() {
	tasks.RegisterTask(&dkgtasks.PlaceHolder{})
	tasks.RegisterTask(&dkgtasks.RegisterTask{})
	tasks.RegisterTask(&dkgtasks.ShareDistributionTask{})
	tasks.RegisterTask(&dkgtasks.DisputeTask{})
	tasks.RegisterTask(&dkgtasks.KeyshareSubmissionTask{})
	tasks.RegisterTask(&dkgtasks.MPKSubmissionTask{})
	tasks.RegisterTask(&dkgtasks.GPKSubmissionTask{})
	tasks.RegisterTask(&dkgtasks.GPKJDisputeTask{})
	tasks.RegisterTask(&dkgtasks.CompletionTask{})
}

func SetupEventMap(em *objects.EventMap) error {

	// if err := em.RegisterLocked("0x3529eeacda732ca25cee203cc6382b6d0688ee079ec8e53fd2dcbf259bdd3fa1", "DepositReceived-Obsolete", nil); err != nil {
	// 	return err
	// }
	// if err := em.RegisterLocked("0x6bae01a1b82866e1dfe8d98c42383fc58df9b4adeb47d7ac24ee4b53d409da6c", "DepositReceived-Obsolete", nil); err != nil {
	// 	return err
	// }
	// if err := em.RegisterLocked("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925", "DSTokenApproval", nil); err != nil {
	// 	return err
	// }
	// if err := em.RegisterLocked("0xce241d7ca1f669fee44b6fc00b8eba2df3bb514eed0f6f668f8f89096e81ed94", "LogSetOwner", nil); err != nil {
	// 	return err
	// }
	// if err := em.RegisterLocked("0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885", "Mint", nil); err != nil {
	// 	return err
	// }
	// if err := em.RegisterLocked("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", "Transfer", nil); err != nil {
	// 	return err
	// }
	// if err := em.RegisterLocked("0x8c25e214c5693ebaf8008875bacedeb9e0aafd393864a314ed1801b2a4e13dd9", "ValidatorJoined", nil); err != nil {
	// 	return err
	// }
	// if err := em.RegisterLocked("0x319bbadb03b94aedc69babb34a28675536a9cb30f4bbde343e1d0018c44ebd94", "ValidatorLeft", nil); err != nil {
	// 	return err
	// }
	// if err := em.RegisterLocked("0x1de2f07b0a1c69916a8b25b889051644192307ea08444a2e11f8654d1db3ab0c", "LockedStake", nil); err != nil {
	// 	return err
	// }

	// Real event processors are below
	// if err := em.RegisterLocked("0x5b063c6569a91e8133fc6cd71d31a4ca5c65c652fd53ae093f46107754f08541", "DepositReceived", svcs.ProcessDepositReceived); err != nil {
	// 	return err
	// }
	// if err := em.RegisterLocked("0x113b129fac2dde341b9fbbec2bb79a95b9945b0e80fda711fc8ae5c7b0ea83b0", "ValidatorMember", svcs.ProcessValidatorMember); err != nil {
	// 	return err
	// }
	// if err := em.RegisterLocked("0x1c85ff1efe0a905f8feca811e617102cb7ec896aded693eb96366c8ef22bb09f", "ValidatorSet", svcs.ProcessValidatorSet); err != nil {
	// 	return err
	// }
	// if err := em.RegisterLocked("0x6d438b6b835d16cdae6efdc0259fdfba17e6aa32dae81863a2467866f85f724a", "SnapshotTaken", svcs.ProcessSnapshotTaken); err != nil {
	// 	return err
	// }
	if err := em.RegisterLocked("0xa84d294194d6169652a99150fd2ef10e18b0d2caa10beeea237bbddcc6e22b10", "ShareDistribution", dkgevents.ProcessShareDistribution); err != nil {
		return err
	}
	if err := em.RegisterLocked("0xb0ee36c3780de716eb6c83687f433ae2558a6923e090fd238b657fb6c896badc", "KeyShareSubmission", dkgevents.ProcessKeyShareSubmission); err != nil {
		return err
	}
	if err := em.RegisterLocked("0x9c6f8368fe7e77e8cb9438744581403bcb3f53298e517f04c1b8475487402e97", "RegistrationOpen", dkgevents.ProcessOpenRegistration); err != nil {
		return err
	}

	return nil
}
