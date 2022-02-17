package deploy

import (
	"bytes"
	"context"
	"math/big"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/config"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/bridge/bindings"
	"github.com/spf13/cobra"
)

const MIGRATION_GRP = 973

var RequiredWei = big.NewInt(8_000_000_000_000)

// Command is the cobra.Command specifically for running deploying contracts
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

	// If flagged we'll deploy and configure the facets used in migrations
	if config.Configuration.Deploy.Migrations {
		err = deployMigrations(eth)
		if err != nil {
			logger.Error(err)
		}

		// If flagged we issue a migration transaction
		if config.Configuration.Deploy.TestMigrations {
			err = testMigrations(eth)
			if err != nil {
				logger.Error(err)
			}
		}
	}

}

func testMigrations(eth interfaces.Ethereum) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	account := eth.GetDefaultAccount()
	logger := logging.GetLogger("test")
	client := eth.GetGethClient()
	c := eth.Contracts()

	txnOpts, err := eth.GetTransactionOpts(ctx, account)
	if err != nil {
		return err
	}

	// Try staking first
	migrateStaking, err := bindings.NewMigrateStakingFacet(c.ValidatorsAddress(), client)
	if err != nil {
		logger.Errorf("binding migrateStaking failed: %v", err)
		return err
	}

	txn, err := migrateStaking.SetBalancesFor(txnOpts, account.Address, big.NewInt(111111), big.NewInt(222222), big.NewInt(333333))
	if err != nil {
		logger.Errorf("setting stake balances failed: %v", err)
		return err
	}
	eth.Queue().QueueAndWait(ctx, txn)
	logger.Infof("setting balances for %v. Gas = %v", account.Address.Hex(), txn.Gas())

	// Try snapshoting
	migrateSnapshots, err := bindings.NewMigrateSnapshotsFacet(c.ValidatorsAddress(), client)
	if err != nil {
		logger.Errorf("binding migrateSnapshots failed: %v", err)
		return err
	}

	sig := bytes.Repeat([]byte{1}, 192)
	bc := bytes.Repeat([]byte{1}, 176)

	txn, err = migrateSnapshots.Snapshot(txnOpts, big.NewInt(1), sig, bc)
	if err != nil {
		logger.Errorf("creating snapshot failed: %v", err)
		return err
	}
	eth.Queue().QueueAndWait(ctx, txn)
	logger.Infof("creating snapshot Gas = %v", txn.Gas())

	txn, err = c.Snapshots().SetEpoch(txnOpts, big.NewInt(2))
	if err != nil {
		logger.Errorf("setting epoch failed: %v", err)
		return err
	}
	eth.Queue().QueueAndWait(ctx, txn)

	// Try Participants

	return nil
}

func deployMigrations(eth interfaces.Ethereum) error {

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

	//
	logger.Infof("Deploying migrations...")

	// Deploy all the migration contracts
	migrateStakingAddr, txn, _, err := bindings.DeployMigrateStakingFacet(txnOpts, client)
	if err != nil {
		logger.Errorf("Failed to deploy migrateStakingAddr...")
		return err
	}
	eth.Queue().QueueGroupTransaction(ctx, MIGRATION_GRP, txn)
	logger.Infof("Deploy migrateStakingAddr = \"0x%0.40x\" gas = %v", migrateStakingAddr, txn.Gas())

	migrateSnapshotsAddr, txn, _, err := bindings.DeployMigrateSnapshotsFacet(txnOpts, client)
	if err != nil {
		logger.Errorf("Failed to deploy migrateSnapshotsAddr...")
		return err
	}
	eth.Queue().QueueGroupTransaction(ctx, MIGRATION_GRP, txn)
	logger.Infof("Deploy migrateSnapshotsAddr = \"0x%0.40x\" gas = %v", migrateSnapshotsAddr, txn.Gas())

	migrateParticipantsAddr, txn, _, err := bindings.DeployMigrateParticipantsFacet(txnOpts, client)
	if err != nil {
		logger.Errorf("Failed to deploy migrateParticipantsAddr...")
		return err
	}
	eth.Queue().QueueGroupTransaction(ctx, MIGRATION_GRP, txn)
	logger.Infof("Deploy migrateParticipantsAddr = \"0x%0.40x\" gas = %v", migrateParticipantsAddr, txn.Gas())

	migrateEthDKGAddr, txn, _, err := bindings.DeployMigrateETHDKG(txnOpts, client)
	if err != nil {
		logger.Errorf("Failed to deploy migrateEthDKGAddr...")
		return err
	}
	eth.Queue().QueueGroupTransaction(ctx, MIGRATION_GRP, txn)
	logger.Infof("Deploy migrateEthDKGAddr = \"0x%0.40x\" gas = %v", migrateEthDKGAddr, txn.Gas())

	// Wire validators migration contracts
	validatorUpdater, err := bindings.NewDiamondUpdateFacet(c.ValidatorsAddress(), client)
	if err != nil {
		logger.Errorf("failed to bind diamond updater to validators: %v", err)
		return err
	}
	vu := &blockchain.Updater{Updater: validatorUpdater, TxnOpts: txnOpts, Logger: logger}

	//
	eth.Queue().QueueGroupTransaction(ctx, MIGRATION_GRP, vu.Add("setBalancesFor(address,uint256,uint256,uint256)", migrateStakingAddr))
	eth.Queue().QueueGroupTransaction(ctx, MIGRATION_GRP, vu.Add("snapshot(uint256,bytes,bytes)", migrateSnapshotsAddr))

	eth.Queue().QueueGroupTransaction(ctx, MIGRATION_GRP, vu.Add("addValidatorImmediate(address,uint256[2])", migrateParticipantsAddr))
	eth.Queue().QueueGroupTransaction(ctx, MIGRATION_GRP, vu.Add("removeValidatorImmediate(address,uint256[2])", migrateParticipantsAddr))

	eth.Queue().WaitGroupTransactions(ctx, MIGRATION_GRP)

	return nil
}
