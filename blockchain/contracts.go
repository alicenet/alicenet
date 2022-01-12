package blockchain

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/bridge/bindings"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ContractDetails contains bindings to smart contract system
type ContractDetails struct {
	eth                  *EthereumDetails
	crypto               *bindings.Crypto
	cryptoAddress        common.Address
	deposit              *bindings.Deposit
	depositAddress       common.Address
	governor             *bindings.Governor
	governorAddress      common.Address
	ethdkg               *bindings.ETHDKG
	ethdkgAddress        common.Address
	participants         *bindings.Participants
	registry             *bindings.Registry
	registryAddress      common.Address
	snapshots            *bindings.Snapshots
	staking              *bindings.Staking
	stakingToken         *bindings.Token
	stakingTokenAddress  common.Address
	utilityToken         *bindings.Token
	utilityTokenAddress  common.Address
	validators           *bindings.Validators
	validatorsAddress    common.Address
	validatorPool        *bindings.ValidatorPool
	validatorPoolAddress common.Address
}

// LookupContracts uses the registry to lookup and create bindings for all required contracts
func (c *ContractDetails) LookupContracts(ctx context.Context, registryAddress common.Address) error {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-signals:
			return errors.New("GoodBye from lookup contracts!")
		case <-time.After(1 * time.Second):
		}

		eth := c.eth
		logger := eth.logger

		// Load the registry first
		registry, err := bindings.NewRegistry(registryAddress, eth.client)
		if err != nil {
			return err
		}
		c.registry = registry
		c.registryAddress = registryAddress

		// Just a help for looking up other contracts
		lookup := func(name string) (common.Address, error) {
			addr, err := registry.Lookup(eth.GetCallOpts(ctx, eth.defaultAccount), name)
			if err != nil {
				logger.Errorf("Failed lookup of \"%v\": %v", name, err)
			} else {
				logger.Infof("Lookup up of \"%v\" is 0x%x", name, addr)
			}
			return addr, err
		}

		// Lookup up governance address and bind to it
		c.governorAddress, err = lookup("governance/v1")
		logAndEat(logger, err)
		if bytes.Equal(c.governorAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.governor, err = bindings.NewGovernor(c.governorAddress, eth.client)
		logAndEat(logger, err)

		// Lookup up deposit address and bind to it
		c.depositAddress, err = lookup("deposit/v1")
		logAndEat(logger, err)
		if bytes.Equal(c.depositAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.deposit, err = bindings.NewDeposit(c.depositAddress, eth.client)
		logAndEat(logger, err)

		c.ethdkgAddress, err = lookup("ethdkg/v1")
		logAndEat(logger, err)
		if bytes.Equal(c.ethdkgAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.ethdkg, err = bindings.NewETHDKG(c.ethdkgAddress, eth.client)
		logAndEat(logger, err)

		c.stakingTokenAddress, err = lookup("stakingToken/v1")
		logAndEat(logger, err)
		if bytes.Equal(c.stakingTokenAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.stakingToken, err = bindings.NewToken(c.stakingTokenAddress, eth.client)
		logAndEat(logger, err)

		c.utilityTokenAddress, err = lookup("utilityToken/v1")
		logAndEat(logger, err)
		if bytes.Equal(c.utilityTokenAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.utilityToken, err = bindings.NewToken(c.utilityTokenAddress, eth.client)
		logAndEat(logger, err)

		c.validatorsAddress, err = lookup("validators/v1")
		logAndEat(logger, err)
		if bytes.Equal(c.validatorsAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorPoolAddress, err = lookup("validatorPool/v1")
		logAndEat(logger, err)
		if bytes.Equal(c.validatorPoolAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorPool, err = bindings.NewValidatorPool(c.validatorPoolAddress, eth.client)
		logAndEat(logger, err)

		// These all call the ValidatorsDiamond contract but we need various interfaces to keep API
		c.validators, err = bindings.NewValidators(c.validatorsAddress, eth.client)
		logAndEat(logger, err)

		c.participants, err = bindings.NewParticipants(c.validatorsAddress, eth.client)
		logAndEat(logger, err)

		c.snapshots, err = bindings.NewSnapshots(c.validatorsAddress, eth.client)
		logAndEat(logger, err)

		stakingAddress, err := lookup("staking/v1")
		logAndEat(logger, err)
		if bytes.Equal(stakingAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.staking, err = bindings.NewStaking(stakingAddress, eth.client)
		logAndEat(logger, err)
		break
	}

	return nil
}

// DeployContracts deploys and does basic setup for all contracts. It returns a binding to the registry, it's address or an error.
func (c *ContractDetails) DeployContracts(ctx context.Context, account accounts.Account) (*bindings.Registry, common.Address, error) {
	eth := c.eth
	logger := eth.logger
	eth.contracts = c

	txnOpts, err := eth.GetTransactionOpts(ctx, account)
	if err != nil {
		return nil, common.Address{}, err
	}

	logger.Debug("Deploying contracts...")
	q := eth.Queue()

	deployGroup := 111
	facetConfigGroup := 222

	var txn *types.Transaction

	// Deploy registry
	c.registryAddress, txn, c.registry, err = bindings.DeployRegistry(txnOpts, eth.client)
	if err != nil {
		logger.Errorf("Failed to deploy registry...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, deployGroup, txn)
	logger.Infof("* registryAddress = \"0x%0.40x\"", c.registryAddress)

	// Deploy staking token
	c.stakingTokenAddress, txn, c.stakingToken, err = bindings.DeployToken(txnOpts, eth.client, StringToBytes32("STK"), StringToBytes32("MadNet Staking"))
	if err != nil {
		logger.Errorf("Failed to deploy stakingToken...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, deployGroup, txn)
	logger.Infof("  stakingTokenAddress = \"0x%0.40x\"", c.stakingTokenAddress)

	// Deploy reference crypto contract
	c.cryptoAddress, txn, c.crypto, err = bindings.DeployCrypto(txnOpts, eth.client)
	if err != nil {
		logger.Errorf("Failed to deploy crypto...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, deployGroup, txn)
	logger.Infof("        cryptoAddress = \"0x%0.40x\"", c.cryptoAddress)

	// Deploy governor
	c.governorAddress, txn, _, err = bindings.DeployDirectGovernance(txnOpts, eth.client)
	if err != nil {
		logger.Errorf("Failed to deploy governance contract...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, deployGroup, txn)
	logger.Infof("    governanceAddress = \"0x%0.40x\"", c.governorAddress)

	c.governor, err = bindings.NewGovernor(c.governorAddress, eth.client)
	logAndEat(logger, err)

	// Deploy utility token
	c.utilityTokenAddress, txn, c.utilityToken, err = bindings.DeployToken(txnOpts, eth.client, StringToBytes32("UTL"), StringToBytes32("MadNet Utility"))
	if err != nil {
		logger.Errorf("Failed to deploy utilityToken...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, deployGroup, txn)
	logger.Infof("  utilityTokenAddress = \"0x%0.40x\"", c.utilityTokenAddress)

	// Deploy Deposit contract
	c.depositAddress, txn, c.deposit, err = bindings.DeployDeposit(txnOpts, eth.client, c.registryAddress)
	if err != nil {
		logger.Errorf("Failed to deploy deposit...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, deployGroup, txn)
	logger.Infof("  depositAddress = \"0x%0.40x\"", c.depositAddress)

	// Deploy ValidatorsDiamond
	c.validatorsAddress, txn, _, err = bindings.DeployValidatorsDiamond(txnOpts, eth.client) // Deploy the core diamond
	if err != nil {
		logger.Errorf("Failed to deploy validators diamond...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, deployGroup, txn)

	// Deploy validators facets
	participantsFacet, txn, _, err := bindings.DeployParticipantsFacet(txnOpts, eth.client)
	if err != nil {
		logger.Error("Failed to deploy participants facet...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, deployGroup, txn)

	// Deploy Snapshot facet
	snapshotsFacet, txn, _, err := bindings.DeploySnapshotsFacet(txnOpts, eth.client)
	if err != nil {
		logger.Error("Failed to deploy snapshots facet...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, deployGroup, txn)

	// Deploy staking facet
	stakingFacet, txn, _, err := bindings.DeployStakingFacet(txnOpts, eth.client)
	if err != nil {
		logger.Error("Failed to deploy staking facet...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, deployGroup, txn)

	c.participants, err = bindings.NewParticipants(c.validatorsAddress, eth.client)
	logAndEat(logger, err)

	c.snapshots, err = bindings.NewSnapshots(c.validatorsAddress, eth.client)
	logAndEat(logger, err)

	c.staking, err = bindings.NewStaking(c.validatorsAddress, eth.client)
	logAndEat(logger, err)

	c.validators, err = bindings.NewValidators(c.validatorsAddress, eth.client)
	if err != nil {
		logger.Errorf("Failed to deploy validators...")
		return nil, common.Address{}, err
	}
	logger.Infof("  validatorsAddress = \"0x%0.40x\"", c.validatorsAddress)

	validatorsUpdate, err := bindings.NewDiamondUpdateFacet(c.validatorsAddress, eth.client)
	if err != nil {
		logger.Errorf("Failed to bind validators update  ..")
		return nil, common.Address{}, err
	}

	// Wait for all the deploys to finish
	eth.commit()

	q.WaitGroupTransactions(ctx, deployGroup)

	// Register all the validators facets
	vu := &Updater{Updater: validatorsUpdate, TxnOpts: txnOpts, Logger: logger}

	// Staking maintenance
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("initializeStaking(address)", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("balanceReward()", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("balanceRewardFor(address)", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("balanceStake()", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("balanceStakeFor(address)", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("balanceUnlocked()", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("balanceUnlockedFor(address)", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("balanceUnlockedReward()", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("balanceUnlockedRewardFor(address)", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("lockStake(uint256)", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("majorFine(address)", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("majorStakeFine()", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("minimumStake()", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("minorFine(address)", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("minorStakeFine()", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("requestUnlockStake()", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("rewardAmount()", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("rewardBonus()", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("setMajorStakeFine(uint256)", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("setMinimumStake(uint256)", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("setMinorStakeFine(uint256)", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("setRewardAmount(uint256)", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("setRewardBonus(uint256)", stakingFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("unlockStake(uint256)", stakingFacet))

	// Snapshot maintenance
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("initializeSnapshots(address)", snapshotsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("snapshot(bytes,bytes)", snapshotsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("setMinEthSnapshotSize(uint256)", snapshotsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("minEthSnapshotSize()", snapshotsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("setMinMadSnapshotSize(uint256)", snapshotsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("minMadSnapshotSize()", snapshotsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("setEpoch(uint256)", snapshotsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("epoch()", snapshotsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("getChainIdFromSnapshot(uint256)", snapshotsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("getRawBlockClaimsSnapshot(uint256)", snapshotsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("getRawSignatureSnapshot(uint256)", snapshotsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("getHeightFromSnapshot(uint256)", snapshotsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("getMadHeightFromSnapshot(uint256)", snapshotsFacet))

	// Validator maintenance
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("initializeParticipants(address)", participantsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("addValidator(address,uint256[2])", participantsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("removeValidator(address,uint256[2])", participantsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("isValidator(address)", participantsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("getValidatorPublicKey(address)", participantsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("confirmValidators()", participantsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("validatorMaxCount()", participantsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("validatorCount()", participantsFacet))
	q.QueueGroupTransaction(ctx, facetConfigGroup, vu.Add("setValidatorMaxCount(uint8)", participantsFacet))

	c.validatorPoolAddress, txn, _, err = bindings.DeployValidatorPool(txnOpts, eth.client, make([]byte, 0))
	if err != nil {
		logger.Errorf("Failed to deploy Validator Pool contract...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)
	logger.Infof(" Gas = %0.10v Validator Pool = \"0x%0.40x\"", txn.Gas(), c.ValidatorPoolAddress())

	ethdkgAccusationAddress, txn, _, err := bindings.DeployETHDKGAccusations(txnOpts, eth.client)
	if err != nil {
		logger.Errorf("Failed to deploy ETHDKGAccusation contract...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)
	logger.Infof(" Gas = %0.10v Ethdkg Accusation = \"0x%0.40x\"", txn.Gas(), ethdkgAccusationAddress)

	ethdkgPhasesAddress, txn, _, err := bindings.DeployETHDKGPhases(txnOpts, eth.client)
	if err != nil {
		logger.Errorf("Failed to deploy ETHDKGPhases contract...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)
	logger.Infof(" Gas = %0.10v ETHDKG Phases = \"0x%0.40x\"", txn.Gas(), ethdkgPhasesAddress)

	c.ethdkgAddress, txn, _, err = bindings.DeployETHDKG(txnOpts, eth.client, c.validatorPoolAddress, ethdkgAccusationAddress, ethdkgPhasesAddress, make([]byte, 0))
	if err != nil {
		logger.Errorf("Failed to deploy EthDKG...")
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)
	logger.Infof(" Gas = %0.10v EthDKG  = \"0x%0.40x\"", txn.Gas(), c.EthdkgAddress())

	c.ethdkg, err = bindings.NewETHDKG(c.ethdkgAddress, eth.client)
	logAndEat(logger, err)

	c.validatorPool, err = bindings.NewValidatorPool(c.validatorPoolAddress, eth.client)
	logAndEat(logger, err)

	// Wait for all the deploys to finish
	eth.commit()

	q.WaitGroupTransactions(ctx, facetConfigGroup)
	// flushQ(txnQueue)

	txn, err = c.ValidatorPool().SetETHDKG(txnOpts, c.ethdkgAddress)
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)

	txn, err = c.registry.Register(txnOpts, "deposit/v1", c.depositAddress)
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)

	txn, err = c.registry.Register(txnOpts, "ethdkg/v1", c.ethdkgAddress)
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)

	txn, err = c.registry.Register(txnOpts, "crypto/v1", c.cryptoAddress)
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)

	txn, err = c.registry.Register(txnOpts, "governance/v1", c.governorAddress)
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)

	txn, err = c.registry.Register(txnOpts, "staking/v1", c.validatorsAddress)
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)

	txn, err = c.registry.Register(txnOpts, "stakingToken/v1", c.stakingTokenAddress)
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)

	txn, err = c.registry.Register(txnOpts, "utilityToken/v1", c.utilityTokenAddress)
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)

	txn, err = c.registry.Register(txnOpts, "validators/v1", c.validatorsAddress)
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)

	txn, err = c.registry.Register(txnOpts, "validatorPool/v1", c.validatorPoolAddress)
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, txn)

	eth.commit()

	// Wait for all the deploys to finish
	q.WaitGroupTransactions(ctx, facetConfigGroup)

	// Initialize Snapshots facet
	tx, err := c.snapshots.InitializeSnapshots(txnOpts, c.registryAddress)
	if err != nil {
		logger.Errorf("Failed to initialize SnapshotsFacet: %v", err)
		return nil, common.Address{}, err
	}
	eth.commit()

	rcpt, err := eth.Queue().QueueAndWait(ctx, tx)
	if err != nil {
		logger.Errorf("Failed to get receipt for initializing Snapshots facet: %v", err)
		return nil, common.Address{}, err
	}
	if rcpt != nil {
		logger.Infof("Snapshots update status: %v", rcpt.Status)
	} else {
		logger.Errorf("Snapshots update receipt is nil")
	}

	tx, err = c.snapshots.SetEpoch(txnOpts, big.NewInt(1))
	if err != nil {
		logger.Errorf("Failed to initialize Snapshots facet next snapshot: %v", err)
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, facetConfigGroup, tx)

	// Default staking values
	tx, err = c.staking.SetMinimumStake(txnOpts, big.NewInt(1000000))
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, tx)

	tx, err = c.staking.SetMajorStakeFine(txnOpts, big.NewInt(200000))
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, tx)

	tx, err = c.staking.SetMinorStakeFine(txnOpts, big.NewInt(50000))
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, tx)

	tx, err = c.staking.SetRewardAmount(txnOpts, big.NewInt(1000))
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, tx)

	tx, err = c.staking.SetRewardBonus(txnOpts, big.NewInt(1000))
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, tx)

	tx, err = c.snapshots.SetMinMadSnapshotSize(txnOpts, big.NewInt(int64(constants.EpochLength)))
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, tx)

	tx, err = c.snapshots.SetMinEthSnapshotSize(txnOpts, big.NewInt(int64(constants.EpochLength/8)))
	logAndEat(logger, err)
	q.QueueGroupTransaction(ctx, facetConfigGroup, tx)

	eth.commit()

	q.WaitGroupTransactions(ctx, facetConfigGroup)

	// Initialize Participants facet
	tx, err = c.participants.InitializeParticipants(txnOpts, c.registryAddress)
	if err != nil {
		logger.Errorf("Failed to initialize Participants facet: %v", err)
		return nil, common.Address{}, err
	}
	eth.commit()

	rcpt, err = eth.Queue().QueueAndWait(ctx, tx)
	if err != nil {
		logger.Errorf("Failed to get receipt for initializing Participants facet: %v", err)
		return nil, common.Address{}, err
	}
	if rcpt != nil {
		logger.Infof("Participants update status: %v", rcpt.Status)
	} else {
		logger.Errorf("Participants update receipt is nil")
	}

	tx, err = c.participants.SetValidatorMaxCount(txnOpts, 10)
	if err != nil {
		logger.Errorf("Failed to initialize Participants facet: %v", err)
		return nil, common.Address{}, err
	}
	q.QueueGroupTransaction(ctx, facetConfigGroup, tx)
	eth.commit()
	q.WaitGroupTransactions(ctx, facetConfigGroup)

	// Staking updates
	tx, err = c.staking.InitializeStaking(txnOpts, c.registryAddress)
	if err != nil {
		logger.Errorf("Failed to update staking contract references: %v", err)
		return nil, common.Address{}, err
	}
	eth.Queue().QueueTransaction(ctx, tx)

	eth.commit()

	rcpt, err = eth.Queue().WaitTransaction(ctx, tx)
	if err != nil {
		logger.Errorf("Failed to get receipt for staking update: %v", err)
		return nil, common.Address{}, err

	}
	if rcpt != nil {
		logger.Infof("staking update status: %v", rcpt.Status)
	} else {
		logger.Errorf("staking receipt is nil")
	}

	// Deposit updates
	tx, err = c.deposit.ReloadRegistry(txnOpts)
	if err != nil {
		logger.Errorf("Failed to update deposit contract references: %v", err)
		return nil, common.Address{}, err
	}
	eth.commit()

	rcpt, err = eth.Queue().QueueAndWait(ctx, tx)
	if err != nil {
		logger.Errorf("Failed to get receipt for deposit update: %v", err)
		return nil, common.Address{}, err
	} else if rcpt != nil {
		logger.Infof("deposit update status: %v", rcpt.Status)
	}

	eth.commit()

	rcpt, err = eth.Queue().QueueAndWait(ctx, tx)
	if err != nil {
		logger.Errorf("Failed to get receipt for ethdkg update: %v", err)
		return nil, common.Address{}, err
	} else if rcpt != nil {
		logger.Infof("ethdkg update status: %v", rcpt.Status)
	}

	// //START: If we want to change the phase length, this is how:
	// tx, err = c.ethdkg.UpdatePhaseLength(txnOpts, big.NewInt(8))
	// if err != nil {
	// 	logger.Errorf("Failed to update ethdkg phase length references: %v", err)
	// 	return nil, common.Address{}, err
	// }

	// eth.commit()

	// rcpt, err = eth.Queue().QueueAndWait(ctx, tx)
	// if err != nil {
	// 	logger.Errorf("Failed to get receipt for ethdkg update: %v", err)
	// 	return nil, common.Address{}, err
	// } else if rcpt != nil {
	// 	logger.Infof("ethdkg update status: %v", rcpt.Status)
	// }
	// //END: If we want to change the phase length

	return c.registry, c.registryAddress, nil
}

func (c *ContractDetails) Crypto() *bindings.Crypto {
	return c.crypto
}

func (c *ContractDetails) CryptoAddress() common.Address {
	return c.cryptoAddress
}

func (c *ContractDetails) Deposit() *bindings.Deposit {
	return c.deposit
}

func (c *ContractDetails) DepositAddress() common.Address {
	return c.depositAddress
}

func (c *ContractDetails) Ethdkg() *bindings.ETHDKG {
	return c.ethdkg
}

func (c *ContractDetails) EthdkgAddress() common.Address {
	return c.ethdkgAddress
}

func (c *ContractDetails) Governor() *bindings.Governor {
	return c.governor
}

func (c *ContractDetails) GovernorAddress() common.Address {
	return c.governorAddress
}

func (c *ContractDetails) Participants() *bindings.Participants {
	return c.participants
}

func (c *ContractDetails) Registry() *bindings.Registry {
	return c.registry
}

func (c *ContractDetails) RegistryAddress() common.Address {
	return c.registryAddress
}

func (c *ContractDetails) Snapshots() *bindings.Snapshots {
	return c.snapshots
}

func (c *ContractDetails) Staking() *bindings.Staking {
	return c.staking
}

func (c *ContractDetails) StakingToken() *bindings.Token {
	return c.stakingToken
}

func (c *ContractDetails) StakingTokenAddress() common.Address {
	return c.stakingTokenAddress
}

func (c *ContractDetails) UtilityToken() *bindings.Token {
	return c.utilityToken
}

func (c *ContractDetails) UtilityTokenAddress() common.Address {
	return c.utilityTokenAddress
}

func (c *ContractDetails) Validators() *bindings.Validators {
	return c.validators
}

func (c *ContractDetails) ValidatorsAddress() common.Address {
	return c.validatorsAddress
}

func (c *ContractDetails) ValidatorPool() *bindings.ValidatorPool {
	return c.validatorPool
}

func (c *ContractDetails) ValidatorPoolAddress() common.Address {
	return c.validatorPoolAddress
}
