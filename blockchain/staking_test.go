package blockchain_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StakingTestSuite struct {
	suite.Suite
	eth      interfaces.Ethereum
	callOpts *bind.CallOpts
	txnOpts  *bind.TransactOpts
}

var InitialAllowance = big.NewInt(1000000000000000000)

func (s *StakingTestSuite) SetupTest() {
	t := s.T()
	var err error
	s.eth, err = setupEthereum(t, 4)
	assert.Nil(t, err)
	eth := s.eth
	c := s.eth.Contracts()
	ctx := context.TODO()

	acct := eth.GetDefaultAccount()
	txnOpts, _ := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	_, err = c.StakingToken().Transfer(txnOpts, acct.Address, InitialAllowance)
	assert.Nilf(t, err, "Initial transfer of %v to %v failed: %v", InitialAllowance, acct.Address.Hex(), err)

	// Tester needs to approve those for Staking
	s.txnOpts, err = eth.GetTransactionOpts(ctx, acct)
	assert.Nil(t, err, "Can't build txnOpts")

	_, err = c.StakingToken().Approve(s.txnOpts, c.ValidatorsAddress(), InitialAllowance)
	assert.Nilf(t, err, "Initial approval of %v to %v failed: %v", InitialAllowance, c.ValidatorsAddress().Hex(), err)
	s.eth.Commit()

	s.callOpts = eth.GetCallOpts(ctx, acct)

	// Tell staking we're in the 1st epoch
	_, err = c.Snapshots().SetEpoch(txnOpts, big.NewInt(1)) // Must be deploy account
	assert.Nil(t, err)
	s.eth.Commit()
}

func (s *StakingTestSuite) TestStakeEvent() {
	t := s.T()

	eth := s.eth
	c := eth.Contracts()

	balance, err := c.Staking().BalanceStake(s.callOpts)
	assert.Truef(t, err == nil, "Failed to check balance:%v", err)
	assert.Truef(t, big.NewInt(10).Cmp(balance) > 0, "Allowance %v insufficient", balance)

	stakeAmount := big.NewInt(1000000)
	txn, err := c.Staking().LockStake(s.txnOpts, stakeAmount)
	assert.Nil(t, err, "Failed to post stake")
	assert.NotNil(t, txn, "Staking transaction is nil")
	s.eth.Commit()

	rcpt, err := eth.Queue().QueueAndWait(context.Background(), txn)
	assert.True(t, err == nil, "Couldn't parse event log:%v", err)

	events := rcpt.Logs
	assert.Equal(t, 2, len(events), "Should be 2 events.")

	foundStakeEvent := false
	for _, event := range events {
		stakeEvent, err := c.Staking().ParseLockedStake(*event)
		if err == nil {
			foundStakeEvent = true
			assert.Equal(t, stakeAmount, stakeEvent.Amount, "Stake amount incorrect")
		}
	}
	assert.True(t, foundStakeEvent)
}

func (s *StakingTestSuite) TestUnlocked() {
	stakeAmount := big.NewInt(1000000)

	t := s.T()
	eth := s.eth
	c := eth.Contracts()
	ctx := context.TODO()

	// Start by making sure unlocked balance and stake are both 0
	unlocked, err := c.Staking().BalanceUnlocked(s.callOpts)
	assert.Truef(t, err == nil, "Failed to get unlocked balance: %v", err)
	assert.Truef(t, big.NewInt(0).Cmp(unlocked) == 0, "Initial unlocked balance should be 0 but is %v", unlocked)
	s.eth.Commit()

	staked, err := c.Staking().BalanceStake(s.callOpts)
	assert.Truef(t, err == nil, "Failed to get stake balance: %v", err)
	assert.Truef(t, big.NewInt(0).Cmp(staked) == 0, "Initial stake should be 0 but is %v", staked)
	s.eth.Commit()

	// Now we lock some - this pulls from token balance based on approvals
	_, err = c.Staking().LockStake(s.txnOpts, stakeAmount)
	assert.True(t, err == nil, "Failed to post stake:%v", err)
	s.eth.Commit()

	// Make sure stake shows the increase and unlocked balance has no change
	staked, err = c.Staking().BalanceStake(s.callOpts)
	assert.Truef(t, err == nil, "Failed to get stake balance: %v", err)
	assert.Truef(t, stakeAmount.Cmp(staked) == 0, "Stake should be %v but is %v", stakeAmount, staked)
	t.Logf("staked balance is %v", staked)

	unlocked, err = c.Staking().BalanceUnlocked(s.callOpts)
	assert.Truef(t, err == nil, "Failed to get unlocked balance: %v", err)
	assert.Truef(t, big.NewInt(0).Cmp(unlocked) == 0, "Unlocked balance should be 0 but is %v", unlocked)
	t.Logf("unlocked balance is %v", unlocked)

	// Request stake be unlockable
	_, err = c.Staking().RequestUnlockStake(s.txnOpts)
	assert.Truef(t, err == nil, "Failed to request unlock of stake: %v", err)
	s.eth.Commit()

	// Set clock ahead - requires privileged account (contract owner/operator)
	ownerAuth, _ := eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	_, err = c.Snapshots().SetEpoch(ownerAuth, big.NewInt(5))
	assert.Truef(t, err == nil, "Failed to set clock forward: %v", err)
	s.eth.Commit()

	// Now we can actually unlock stake
	txn, err := c.Staking().UnlockStake(s.txnOpts, stakeAmount)
	assert.Truef(t, err == nil, "Failed to unlock stake: %v", err)
	s.eth.Commit()

	// Just making sure the unlock completes
	_, err = eth.Queue().QueueAndWait(context.Background(), txn)
	if err != nil {
		t.Fatal(err)
	}
	// Now unlocked balance contains what was formerly staked
	unlocked, err = c.Staking().BalanceUnlocked(s.callOpts)
	assert.Truef(t, err == nil, "Failed to get stake balance: %v", err)
	assert.Truef(t, stakeAmount.Cmp(unlocked) == 0, "Unlocked balance should be %v but is %v", stakeAmount, unlocked)
}

func (s *StakingTestSuite) TestBalanceUnlockedFor() {
	t := s.T()
	eth := s.eth
	c := eth.Contracts()

	balance, err := c.Staking().BalanceUnlockedFor(s.callOpts, c.ValidatorsAddress())
	assert.Nilf(t, err, "Failed: balanceUnlockedFor()")
	assert.Truef(t, big.NewInt(0).Cmp(balance) == 0, "Allowance initially should be %v but is %v", InitialAllowance, balance)
}

func TestStakingTestSuite(t *testing.T) {
	suite.Run(t, new(StakingTestSuite))
}
