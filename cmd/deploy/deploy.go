package deploy

import (
	"context"
	"math/big"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/config"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/bridge/bindings"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/spf13/cobra"
)

var RequiredWei = big.NewInt(8_000_000_000_000)

// Command is the cobra.Command specifically for running as an edge node, i.e. not a validator or relay
var Command = cobra.Command{
	Use:   "deploy",
	Short: "Deploys required smart contracts to Ethereum",
	Long:  "deploy uses Go bindings to deploy required smart contracts.",
	Run:   deployNode}

func deployNode(cmd *cobra.Command, args []string) {
	logger := logging.GetLogger("deploy")
	logger.Info("Deploying contracts...")

	eth, err := blockchain.NewEthereumEndpoint(
		config.Configuration.Ethereum.Endpoint,
		config.Configuration.Ethereum.Keystore,
		config.Configuration.Ethereum.Passcodes,
		config.Configuration.Ethereum.DefaultAccount,
		config.Configuration.Ethereum.Timeout,
		config.Configuration.Ethereum.RetryCount,
		config.Configuration.Ethereum.RetryDelay,
		config.Configuration.Ethereum.FinalityDelay)
	if err != nil {
		logger.Errorf("Could not connect to Ethereum: %v", err)
	}
	c := eth.Contracts()

	acct := eth.GetDefaultAccount()
	err = eth.UnlockAccount(acct)
	if err != nil {
		logger.Fatal(err)
	}

	bal, err := eth.GetBalance(acct.Address)
	if err != nil {
		logger.Warnf("Could not get balance for %v: %v", acct.Address.Hex(), err)
	}
	logger.Infof("DeployAccount: %v Balance: %v", acct.Address.Hex(), bal.String())

	if bal.Cmp(RequiredWei) < 0 {
		logger.Warnf("Probably insufficient gas, but trying anyway")
	}
	_, _, err = c.DeployContracts(context.Background(), acct)
	if err != nil {
		logger.Errorf("Could not deploy contracts: %v", err)
	}

	if config.Configuration.Deploy.Migrations {
		err = deployMigrations(eth)
		if err != nil {
			logger.Error(err)
		}
	}
}

func deployMigrations(eth blockchain.Ethereum) error {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	account := eth.GetDefaultAccount()
	logger := logging.GetLogger("deploy")
	client := eth.GetGethClient()
	c := eth.Contracts()

	txnOpts, err := eth.GetTransactionOpts(ctx, account)
	if err != nil {
		return err
	}

	txnQueue := queue.New(10)
	q := func(tx *types.Transaction) {
		if tx != nil {
			logger.Debugf("Queueing transaction %v", tx.Hash().String())
			txnQueue.Put(tx)
		} else {
			logger.Warn("Ignoring nil transaction")
		}
	}

	flushQ := func(queue *queue.Queue) {
		logger.Debugf("waiting for txns...")
		for txns, err := queue.Get(1); !queue.Empty(); txns, err = queue.Get(1) {
			if err != nil {
				logger.Errorf("failure: %v", err)
			}
			tx := txns[0].(*types.Transaction)
			logger.Debugf("waiting for txn: %v", tx.Hash().String())
			eth.WaitForReceipt(ctx, tx)
		}
	}

	//
	var txn *types.Transaction

	// Deploy all the migration contracts
	migrateStakingAddr, txn, _, err := bindings.DeployMigrateStakingFacet(txnOpts, client)
	if err != nil {
		logger.Errorf("Failed to deploy migrateStakingAddr...")
		return err
	}
	q(txn)
	logger.Infof("Deploy migrateStakingAddr = \"0x%0.40x\" gas = %v", migrateStakingAddr, txn.Gas())

	migrateSnapshotsAddr, txn, _, err := bindings.DeployMigrateSnapshotsFacet(txnOpts, client)
	if err != nil {
		logger.Errorf("Failed to deploy migrateSnapshotsAddr...")
		return err
	}
	q(txn)
	logger.Infof("Deploy migrateSnapshotsAddr = \"0x%0.40x\" gas = %v", migrateSnapshotsAddr, txn.Gas())

	migrateParticipantsAddr, txn, _, err := bindings.DeployMigrateParticipantsFacet(txnOpts, client)
	if err != nil {
		logger.Errorf("Failed to deploy migrateParticipantsAddr...")
		return err
	}
	q(txn)
	logger.Infof("Deploy migrateParticipantsAddr = \"0x%0.40x\" gas = %v", migrateParticipantsAddr, txn.Gas())

	migrateEthDKGAddr, txn, _, err := bindings.DeployMigrateETHDKG(txnOpts, client)
	if err != nil {
		logger.Errorf("Failed to deploy migrateEthDKGAddr...")
		return err
	}
	q(txn)
	logger.Infof("Deploy migrateEthDKGAddr = \"0x%0.40x\" gas = %v", migrateEthDKGAddr, txn.Gas())

	// Wire validators migration contracts
	validatorUpdater, err := bindings.NewDiamondUpdateFacet(c.ValidatorsAddress, client)
	if err != nil {
		logger.Errorf("failed to bind diamond updater to validators: %v", err)
		return err
	}
	vu := &blockchain.Updater{Updater: validatorUpdater, TxnOpts: txnOpts, Logger: logger}
	q(vu.Add("setBalancesFor(address,uint256,uint256,uint256)", migrateStakingAddr))
	q(vu.Add("snapshot(uint256,bytes,bytes)", migrateSnapshotsAddr))

	q(vu.Add("addValidatorImmediate(address,uint256[2])", migrateParticipantsAddr))
	q(vu.Add("removeValidatorImmediate(address,uint256[2])", migrateParticipantsAddr))

	// Wire EthDKG migration contract
	ethdkgUpdater, err := bindings.NewDiamondUpdateFacet(c.EthdkgAddress, client)
	if err != nil {
		logger.Errorf("failed to bind diamond updater to ethdkg: %v", err)
		return err
	}
	eu := &blockchain.Updater{Updater: ethdkgUpdater, TxnOpts: txnOpts, Logger: logger}
	q(eu.Add("migrate(uint256,uint32,uint32,uint256[4],address[],uint256[4][])", migrateEthDKGAddr))

	flushQ(txnQueue)

	return nil
}
