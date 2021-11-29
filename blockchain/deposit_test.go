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

type DepositTestSuite struct {
	suite.Suite
	eth      interfaces.Ethereum
	callOpts *bind.CallOpts
	txnOpts  *bind.TransactOpts
}

func (s *DepositTestSuite) SetupTest() {
	t := s.T()
	eth, err := setupEthereum(t, 4)
	assert.Nil(t, err)
	c := eth.Contracts()

	s.eth = eth
	ctx := context.TODO()

	testAcct := eth.GetDefaultAccount()

	err = eth.UnlockAccount(testAcct)
	if err != nil {
		panic(err)
	}

	bal, _ := eth.GetBalance(testAcct.Address)
	t.Logf("ether balance of %v is %v", testAcct.Address.Hex(), bal)

	// Deployer starts with tokens, so has to transfer
	txnOpts, _ := eth.GetTransactionOpts(ctx, testAcct)
	_, err = c.UtilityToken().Transfer(txnOpts, testAcct.Address, InitialAllowance)
	assert.Nil(t, err)
	eth.Commit()

	assert.Nilf(t, err, "Initial transfer of %v to %v failed: %v", InitialAllowance, testAcct.Address.Hex(), err)
	if err == nil {
		t.Logf("Initial transfer of %v tokens to %v succeeded.", InitialAllowance, testAcct.Address.Hex())
	}

	s.callOpts = eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	s.txnOpts, err = eth.GetTransactionOpts(ctx, eth.GetDefaultAccount())
	assert.Nil(t, err, "Failed to build txnOpts")
}

func (s *DepositTestSuite) TestDepositEvent() {
	t := s.T()
	eth := s.eth
	c := eth.Contracts()

	bal, _ := c.UtilityToken().BalanceOf(s.callOpts, eth.GetDefaultAccount().Address)
	t.Logf("utility token balance of %v is %v", eth.GetDefaultAccount().Address.Hex(), bal)

	bal, _ = eth.GetBalance(eth.GetDefaultAccount().Address)
	t.Logf("ether balance of %v is %v", eth.GetDefaultAccount().Address.Hex(), bal)

	// Approve deposit contract to withdrawh.GetDefaultAccount())
	txn, err := c.UtilityToken().Approve(s.txnOpts, c.DepositAddress(), big.NewInt(10000))
	assert.Nilf(t, err, "Approve failed by %v to %v", eth.GetDefaultAccount().Address.Hex(), c.DepositAddress().Hex())
	assert.NotNil(t, txn, "Approve failed: transaction is nil")
	s.eth.Commit()

	// Tell deposit contract to withdraw
	txn, err = c.Deposit().Deposit(s.txnOpts, big.NewInt(1000))
	assert.Nil(t, err, "Deposit failed")
	assert.NotNilf(t, txn, "Deposit failed: transaction is nil")
	s.eth.Commit()
}

func TestDepositTestSuite(t *testing.T) {
	suite.Run(t, new(DepositTestSuite))
}
