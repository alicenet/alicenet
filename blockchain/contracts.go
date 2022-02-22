package blockchain

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MadBase/bridge/bindings"
	"github.com/ethereum/go-ethereum/common"
)

// ContractDetails contains bindings to smart contract system
type ContractDetails struct {
	eth                  *EthereumDetails
	ethdkg               *bindings.ETHDKG
	ethdkgAddress        common.Address
	madToken             *bindings.MadToken
	madTokenAddress      common.Address
	madByte              *bindings.MadByte
	madByteAddress       common.Address
	stakeNFT             *bindings.StakeNFT
	stakeNFTAddress      common.Address
	validatorNFT         *bindings.ValidatorNFT
	validatorNFTAddress  common.Address
	registry             *bindings.Registry
	registryAddress      common.Address
	snapshots            *bindings.Snapshots
	snapshotsAddress     common.Address
	validatorPool        *bindings.ValidatorPool
	validatorPoolAddress common.Address
	// factory        *bindings.Factory
	// factoryAddress common.Address
}

// LookupContracts uses the registry to lookup and create bindings for all required contracts
func (c *ContractDetails) LookupContracts(ctx context.Context, registryAddress common.Address) error {
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

		c.ethdkgAddress, err = lookup("ethdkg/v1")
		logAndEat(logger, err)
		if bytes.Equal(c.ethdkgAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.ethdkg, err = bindings.NewETHDKG(c.ethdkgAddress, eth.client)
		logAndEat(logger, err)

		c.validatorPoolAddress, err = lookup("validatorPool/v1")
		logAndEat(logger, err)
		if bytes.Equal(c.validatorPoolAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorPool, err = bindings.NewValidatorPool(c.validatorPoolAddress, eth.client)
		logAndEat(logger, err)

		c.snapshots, err = bindings.NewSnapshots(c.validatorPoolAddress, eth.client)
		logAndEat(logger, err)

		stakingAddress, err := lookup("staking/v1")
		logAndEat(logger, err)
		if bytes.Equal(stakingAddress.Bytes(), make([]byte, 20)) {
			continue
		}

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

func (c *ContractDetails) Registry() *bindings.Registry {
	return c.registry
}

func (c *ContractDetails) RegistryAddress() common.Address {
	return c.registryAddress
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

// func (c *ContractDetails) Factory() *bindings.Factory {
// 	return c.factory
// }

// func (c *ContractDetails) FactoryAddress() common.Address {
// 	return c.factoryAddress
// }
