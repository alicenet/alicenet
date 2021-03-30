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
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/MadBase/MadNet/blockchain/tasks/dkgtasks"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var accountAddresses []string = []string{
	"0x546F99F244b7B58B855330AE0E2BC1b30b41302F", "0x9AC1c9afBAec85278679fF75Ef109217f26b1417",
	"0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac", "0x615695C4a4D6a60830e5fca4901FbA099DF26271",
	"0x63a6627b79813A7A43829490C4cE409254f64177"}

func connectSimulatorEndpoint(t *testing.T) blockchain.Ethereum {
	eth, err := blockchain.NewEthereumSimulator(
		"../../../assets/test/keys",
		"../../../assets/test/passcodes.txt",
		6,
		1*time.Second,
		0,
		big.NewInt(9223372036854775807),
		accountAddresses...)
	assert.Nil(t, err)

	go func() {
		for true {
			time.Sleep(1 * time.Second)
			eth.Commit()
		}
	}()

	return eth
}

func connectRemoteEndpoint(t *testing.T) blockchain.Ethereum {
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

func joinValidatorSet(t *testing.T, eth blockchain.Ethereum, ownerAcct accounts.Account, validatorAcct accounts.Account) {
	c := eth.Contracts()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create txn opts for owner and validator
	ownerTxnOpts, err := eth.GetTransactionOpts(ctx, ownerAcct)
	assert.Nil(t, err)

	txnOpts, err := eth.GetTransactionOpts(ctx, validatorAcct)
	assert.Nil(t, err)

	// Transfer tokens from owner to validator
	txn, err := c.StakingToken.Transfer(ownerTxnOpts, validatorAcct.Address, big.NewInt(10000000))
	assert.Nil(t, err)

	rcpt, err := eth.WaitForReceipt(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt)
	assert.Equal(t, rcpt.Status, uint64(1))

	// Approve tokens for staking contract to withdraw
	txn, err = c.StakingToken.Approve(txnOpts, c.ValidatorsAddress, big.NewInt(1000000))
	assert.Nil(t, err)

	rcpt, err = eth.WaitForReceipt(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt)
	assert.Equal(t, rcpt.Status, uint64(1))

	// Tell staking contract to lock stake
	txn, err = c.Staking.LockStake(txnOpts, big.NewInt(1000000))
	assert.Nil(t, err)

	rcpt, err = eth.WaitForReceipt(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt)
	assert.Equal(t, rcpt.Status, uint64(1))

	// Tell validators we want to join
	txn, err = c.Validators.AddValidator(txnOpts, validatorAcct.Address, [2]*big.Int{big.NewInt(1), big.NewInt(2)})
	assert.Nil(t, err)

	rcpt, err = eth.WaitForReceipt(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt)
	assert.Equal(t, rcpt.Status, uint64(1))
}

func TestRegisterSuccess(t *testing.T) {
	eth := connectSimulatorEndpoint(t)
	// eth := connectRemoteEndpoint(t)
	logging.GetLogger("ethereum").SetLevel(logrus.InfoLevel)

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

	_, _, err := c.DeployContracts(ctx, ownerAccount)
	// err := c.LookupContracts(common.HexToAddress("0xe83043E6fCafda1664254C393D6c151EadE85Cc0"))
	assert.Nil(t, err)

	t.Logf("  ethdkg address: %v", c.EthdkgAddress.Hex())
	t.Logf("registry address: %v", c.RegistryAddress.Hex())

	callOpts := eth.GetCallOpts(ctx, ownerAccount)
	txnOpts, err := eth.GetTransactionOpts(context.Background(), ownerAccount)
	assert.Nil(t, err)

	// Kick off a round of ethdkg
	txn, err := c.Ethdkg.InitializeState(txnOpts)
	assert.Nil(t, err)

	rcpt, err := eth.WaitForReceipt(ctx, txn)
	assert.Nil(t, err)
	assert.NotNil(t, rcpt)
	assert.Equal(t, rcpt.Status, uint64(1), "receipt status shows transaction failure")

	// Now we know ethdkg is running, let's find out when registration has to happen
	// TODO this should be based on an OpenRegistration event
	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)
	t.Logf("currentHeight:%v", currentHeight)

	endingHeight, err := c.Ethdkg.TREGISTRATIONEND(callOpts)
	assert.Nil(t, err)
	t.Logf("endingHeight:%v", endingHeight)

	logging.GetLogger("ethsim").SetLevel(logrus.WarnLevel)

	// This will be slow
	for i := 0; i < 4; i++ {
		acct, err := eth.GetAccount(common.HexToAddress(accountAddresses[i+1]))
		assert.Nil(t, err)

		joinValidatorSet(t, eth, ownerAccount, acct)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 4; i++ {
		acct, err := eth.GetAccount(common.HexToAddress(accountAddresses[i+1]))
		assert.Nil(t, err)

		wg.Add(1)
		go validator(t, i, eth, acct, &wg)
	}
	wg.Wait()

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
	taskManager := tasks.NewManager(logger)

	handlers := make(map[uint64]func())

	for currentBlock < startBlock+60 {
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

					priv, pub, err := dkg.GenerateKeys()
					assert.Nil(t, err)

					registrationEnds := event.RegistrationEnds.Uint64()
					distributionEnds := event.ShareDistributionEnds.Uint64()

					registrationTask := dkgtasks.NewRegisterTask(pub, registrationEnds)
					th := taskManager.NewTaskHandler(10*time.Minute, time.Second, registrationTask)

					th.Start()

					handlers[registrationEnds+1] = func() {

						callOpts := eth.GetCallOpts(ctx, validatorAcct)
						assert.Nil(t, err)

						// Number participants in key generation
						bigN, err := c.Ethdkg.NumberOfRegistrations(callOpts)
						assert.Nil(t, err)

						n := bigN.Uint64()

						threshold, _ := dkg.ThresholdForUserCount(int(n))

						// Make n participants
						participants := []*dkg.Participant{}
						for idx := uint64(0); idx < n; idx++ {

							addr, err := c.Ethdkg.Addresses(callOpts, new(big.Int).SetUint64(idx))
							assert.Nil(t, err)

							var publicKey [2]*big.Int
							publicKey[0], err = c.Ethdkg.PublicKeys(callOpts, addr, big.NewInt(0))
							assert.Nil(t, err)

							publicKey[1], err = c.Ethdkg.PublicKeys(callOpts, addr, big.NewInt(1))
							assert.Nil(t, err)

							participant := &dkg.Participant{
								Address:   addr,
								Index:     int(idx + 1),
								PublicKey: publicKey}

							participants = append(participants, participant)
						}
						shares, coeff, commitments, err := dkg.GenerateShares(priv, pub, participants, threshold)
						t.Logf("coeff:%v", coeff)

						distroTask := dkgtasks.NewShareDistributionTask(pub, shares, commitments, registrationEnds, distributionEnds)
						dth := taskManager.NewTaskHandler(time.Second, time.Second, distroTask)

						dth.Start()
					}
				}

			}

			if fn, present := handlers[currentBlock]; present {
				fn()
				delete(handlers, currentBlock)
			}

			lastBlock = currentBlock
		}

		time.Sleep(time.Second)
	}

	wg.Done()
}
