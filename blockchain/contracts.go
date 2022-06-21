package blockchain

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/ethereum/go-ethereum/common"
)

// ContractDetails contains bindings to smart contract system
type ContractDetails struct {
	eth                     *EthereumDetails
	ethdkg                  bindings.IETHDKG
	ethdkgAddress           common.Address
	aToken                  bindings.IAToken
	aTokenAddress           common.Address
	bToken                  bindings.IBToken
	bTokenAddress           common.Address
	publicStaking           bindings.IPublicStaking
	publicStakingAddress    common.Address
	validatorStaking        bindings.IValidatorStaking
	validatorStakingAddress common.Address
	contractFactory         bindings.IAliceNetFactory
	contractFactoryAddress  common.Address
	snapshots               bindings.ISnapshots
	snapshotsAddress        common.Address
	validatorPool           bindings.IValidatorPool
	validatorPoolAddress    common.Address
	governance              bindings.IGovernance
	governanceAddress       common.Address
}

// LookupContracts uses the registry to lookup and create bindings for all required contracts
func (c *ContractDetails) LookupContracts(ctx context.Context, contractFactoryAddress common.Address) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-signals:
			return errors.New("goodBye from lookup contracts")
		case <-time.After(1 * time.Second):
		}

		eth := c.eth
		logger := eth.logger

		// Load the contractFactory first
		contractFactory, err := bindings.NewAliceNetFactory(contractFactoryAddress, eth.client)
		if err != nil {
			return err
		}
		c.contractFactory = contractFactory
		c.contractFactoryAddress = contractFactoryAddress

		// todo: replace lookup with deterministic address compute

		// Just a help for looking up other contracts
		lookup := func(name string) (common.Address, error) {
			salt := StringToBytes32(name)
			addr, err := contractFactory.Lookup(eth.GetCallOpts(ctx, eth.defaultAccount), salt)
			if err != nil {
				logger.Errorf("Failed lookup of \"%v\": %v", name, err)
			} else {
				logger.Infof("Lookup up of \"%v\" is 0x%x", name, addr)
			}
			return addr, err
		}

		// ETHDKG
		c.ethdkgAddress, err = lookup("ETHDKG")
		logAndEat(logger, err)
		if bytes.Equal(c.ethdkgAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.ethdkg, err = bindings.NewETHDKG(c.ethdkgAddress, eth.client)
		logAndEat(logger, err)

		// ValidatorPool
		c.validatorPoolAddress, err = lookup("ValidatorPool")
		logAndEat(logger, err)
		if bytes.Equal(c.validatorPoolAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorPool, err = bindings.NewValidatorPool(c.validatorPoolAddress, eth.client)
		logAndEat(logger, err)

		// BToken
		c.bTokenAddress, err = lookup("BToken")
		logAndEat(logger, err)
		if bytes.Equal(c.bTokenAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.bToken, err = bindings.NewBToken(c.bTokenAddress, eth.client)
		logAndEat(logger, err)

		// AToken
		c.aTokenAddress, err = lookup("AToken")
		logAndEat(logger, err)
		if bytes.Equal(c.aTokenAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.aToken, err = bindings.NewAToken(c.aTokenAddress, eth.client)
		logAndEat(logger, err)

		// PublicStaking
		c.publicStakingAddress, err = lookup("PublicStaking")
		logAndEat(logger, err)
		if bytes.Equal(c.publicStakingAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.publicStaking, err = bindings.NewPublicStaking(c.publicStakingAddress, eth.client)
		logAndEat(logger, err)

		// ValidatorStaking
		c.validatorStakingAddress, err = lookup("ValidatorStaking")
		logAndEat(logger, err)
		if bytes.Equal(c.validatorStakingAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorStaking, err = bindings.NewValidatorStaking(c.validatorStakingAddress, eth.client)
		logAndEat(logger, err)

		// Governance
		c.governanceAddress, err = lookup("Governance")
		logAndEat(logger, err)
		if bytes.Equal(c.governanceAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.governance, err = bindings.NewGovernance(c.governanceAddress, eth.client)
		logAndEat(logger, err)

		// Snapshots
		c.snapshotsAddress, err = lookup("Snapshots")
		logAndEat(logger, err)
		if bytes.Equal(c.snapshotsAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.snapshots, err = bindings.NewSnapshots(c.snapshotsAddress, eth.client)
		logAndEat(logger, err)

		break
	}

	return nil
}

func (c *ContractDetails) Ethdkg() bindings.IETHDKG {
	return c.ethdkg
}

func (c *ContractDetails) EthdkgAddress() common.Address {
	return c.ethdkgAddress
}

func (c *ContractDetails) AToken() bindings.IAToken {
	return c.aToken
}

func (c *ContractDetails) ATokenAddress() common.Address {
	return c.aTokenAddress
}

func (c *ContractDetails) BToken() bindings.IBToken {
	return c.bToken
}

func (c *ContractDetails) BTokenAddress() common.Address {
	return c.bTokenAddress
}

func (c *ContractDetails) PublicStaking() bindings.IPublicStaking {
	return c.publicStaking
}

func (c *ContractDetails) PublicStakingAddress() common.Address {
	return c.publicStakingAddress
}

func (c *ContractDetails) ValidatorStaking() bindings.IValidatorStaking {
	return c.validatorStaking
}

func (c *ContractDetails) ValidatorStakingAddress() common.Address {
	return c.validatorStakingAddress
}

func (c *ContractDetails) ContractFactory() bindings.IAliceNetFactory {
	return c.contractFactory
}

func (c *ContractDetails) ContractFactoryAddress() common.Address {
	return c.contractFactoryAddress
}

func (c *ContractDetails) Snapshots() bindings.ISnapshots {
	return c.snapshots
}

func (c *ContractDetails) SnapshotsAddress() common.Address {
	return c.snapshotsAddress
}

func (c *ContractDetails) ValidatorPool() bindings.IValidatorPool {
	return c.validatorPool
}

func (c *ContractDetails) ValidatorPoolAddress() common.Address {
	return c.validatorPoolAddress
}

func (c *ContractDetails) Governance() bindings.IGovernance {
	return c.governance
}

func (c *ContractDetails) GovernanceAddress() common.Address {
	return c.governanceAddress
}
