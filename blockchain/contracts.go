package blockchain

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MadBase/MadNet/bridge/bindings"
	"github.com/ethereum/go-ethereum/common"
)

// ContractDetails contains bindings to smart contract system
type ContractDetails struct {
	eth                    *EthereumDetails
	ethdkg                 *bindings.ETHDKG
	ethdkgAddress          common.Address
	madToken               *bindings.MadToken
	madTokenAddress        common.Address
	madByte                *bindings.MadByte
	madByteAddress         common.Address
	stakeNFT               *bindings.StakeNFT
	stakeNFTAddress        common.Address
	validatorNFT           *bindings.ValidatorNFT
	validatorNFTAddress    common.Address
	contractFactory        *bindings.MadnetFactory
	contractFactoryAddress common.Address
	snapshots              *bindings.Snapshots
	snapshotsAddress       common.Address
	validatorPool          *bindings.ValidatorPool
	validatorPoolAddress   common.Address
	// factory        *bindings.Factory
	// factoryAddress common.Address
	governance        *bindings.Governance
	governanceAddress common.Address
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
		contractFactory, err := bindings.NewMadnetFactory(contractFactoryAddress, eth.client)
		if err != nil {
			return err
		}
		c.contractFactory = contractFactory
		c.contractFactoryAddress = contractFactoryAddress

		// todo: replace lookup with deterministic address compute

		// Just a help for looking up other contracts
		lookup := func(name string) (common.Address, error) {
			addr, err := contractFactory.Lookup(eth.GetCallOpts(ctx, eth.defaultAccount), name)
			if err != nil {
				logger.Errorf("Failed lookup of \"%v\": %v", name, err)
			} else {
				logger.Infof("Lookup up of \"%v\" is 0x%x", name, addr)
			}
			return addr, err
		}

		/*
			- "MadnetFactoryBase"
			- "Foundation"
			+ "MadByte".
			+ "MadToken".
			+ "StakeNFT".
			- "StakeNFTLP"
			+ "ValidatorNFT".
			- "ETHDKGAccusations"
			- "ETHDKGPhases"
			+ "ETHDKG".
			+ "Governance".
			+ "Snapshots".
			+ "ValidatorPool".
		*/

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

		// MadByte
		c.madByteAddress, err = lookup("MadByte")
		logAndEat(logger, err)
		if bytes.Equal(c.madByteAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.madByte, err = bindings.NewMadByte(c.madByteAddress, eth.client)
		logAndEat(logger, err)

		// MadToken
		c.madTokenAddress, err = lookup("MadToken")
		logAndEat(logger, err)
		if bytes.Equal(c.madTokenAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.madToken, err = bindings.NewMadToken(c.madTokenAddress, eth.client)
		logAndEat(logger, err)

		// StakeNFT
		c.stakeNFTAddress, err = lookup("StakeNFT")
		logAndEat(logger, err)
		if bytes.Equal(c.stakeNFTAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.stakeNFT, err = bindings.NewStakeNFT(c.stakeNFTAddress, eth.client)
		logAndEat(logger, err)

		// ValidatorNFT
		c.validatorNFTAddress, err = lookup("ValidatorNFT")
		logAndEat(logger, err)
		if bytes.Equal(c.validatorNFTAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorNFT, err = bindings.NewValidatorNFT(c.validatorNFTAddress, eth.client)
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

func (c *ContractDetails) Ethdkg() *bindings.ETHDKG {
	return c.ethdkg
}

func (c *ContractDetails) EthdkgAddress() common.Address {
	return c.ethdkgAddress
}

func (c *ContractDetails) MadToken() *bindings.MadToken {
	return c.madToken
}

func (c *ContractDetails) MadTokenAddress() common.Address {
	return c.madTokenAddress
}

func (c *ContractDetails) MadByte() *bindings.MadByte {
	return c.madByte
}

func (c *ContractDetails) MadByteAddress() common.Address {
	return c.madByteAddress
}

func (c *ContractDetails) StakeNFT() *bindings.StakeNFT {
	return c.stakeNFT
}

func (c *ContractDetails) StakeNFTAddress() common.Address {
	return c.stakeNFTAddress
}

func (c *ContractDetails) ValidatorNFT() *bindings.ValidatorNFT {
	return c.validatorNFT
}

func (c *ContractDetails) ValidatorNFTAddress() common.Address {
	return c.validatorNFTAddress
}

func (c *ContractDetails) ContractFactory() *bindings.MadnetFactory {
	return c.contractFactory
}

func (c *ContractDetails) ContractFactoryAddress() common.Address {
	return c.contractFactoryAddress
}

func (c *ContractDetails) Snapshots() *bindings.Snapshots {
	return c.snapshots
}

func (c *ContractDetails) SnapshotsAddress() common.Address {
	return c.snapshotsAddress
}

func (c *ContractDetails) ValidatorPool() *bindings.ValidatorPool {
	return c.validatorPool
}

func (c *ContractDetails) ValidatorPoolAddress() common.Address {
	return c.validatorPoolAddress
}

func (c *ContractDetails) Governance() *bindings.Governance {
	return c.governance
}

func (c *ContractDetails) GovernanceAddress() common.Address {
	return c.governanceAddress
}

// func (c *ContractDetails) Factory() *bindings.Factory {
// 	return c.factory
// }

// func (c *ContractDetails) FactoryAddress() common.Address {
// 	return c.factoryAddress
// }
