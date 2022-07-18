package ethereum

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

var (
	once      sync.Once
	contracts *Contracts
)

// Contracts contains bindings to smart contract system
type Contracts struct {
	isInitialized           bool
	allAddresses            map[common.Address]bool
	eth                     *Client
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

func GetContracts() *Contracts {
	if contracts == nil || !contracts.isInitialized {
		panic("Ethereum smart contracts not initialized or not found")
	}
	return contracts
}

/// Set the contractFactoryAddress and looks up for all the contracts that we
/// need that were deployed via the factory. It's only executed once. Other call
/// to this functions are no-op.
func NewContracts(eth *Client, contractFactoryAddress common.Address) {
	once.Do(func() {
		contracts = getNewContractInstance(eth, contractFactoryAddress)
	})
}

func getNewContractInstance(eth *Client, contractFactoryAddress common.Address) *Contracts {
	tempContracts := &Contracts{
		allAddresses:           make(map[common.Address]bool),
		eth:                    eth,
		contractFactoryAddress: contractFactoryAddress,
	}
	err := tempContracts.lookupContracts()
	if err != nil {
		panic(err)
	}
	tempContracts.isInitialized = true
	return tempContracts
}

// LookupContracts uses the registry to lookup and create bindings for all required contracts
func (c *Contracts) lookupContracts() error {
	networkCtx, cf := context.WithCancel(context.Background())
	defer cf()
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
		contractFactory, err := bindings.NewAliceNetFactory(c.contractFactoryAddress, eth.internalClient)
		if err != nil {
			return err
		}
		c.contractFactory = contractFactory

		// todo: replace lookup with deterministic address compute

		callOpts, err := eth.GetCallOpts(networkCtx, eth.defaultAccount)
		if err != nil {
			logger.Errorf("Failed to generate call options for lookup %v", err)
		}

		// Just a help for looking up other contracts
		lookup := func(name string) (common.Address, error) {
			salt := utils.StringToBytes32(name)
			addr, err := contractFactory.Lookup(callOpts, salt)
			if err != nil {
				logger.Errorf("Failed lookup of \"%v\": %v", name, err)
			} else {
				logger.Infof("Lookup up of \"%v\" is 0x%x", name, addr)
			}
			c.allAddresses[addr] = true
			return addr, err
		}

		// ETHDKG
		c.ethdkgAddress, err = lookup("ETHDKG")
		logAndEat(logger, err)
		if bytes.Equal(c.ethdkgAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.ethdkg, err = bindings.NewETHDKG(c.ethdkgAddress, eth.internalClient)
		logAndEat(logger, err)

		// ValidatorPool
		c.validatorPoolAddress, err = lookup("ValidatorPool")
		logAndEat(logger, err)
		if bytes.Equal(c.validatorPoolAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorPool, err = bindings.NewValidatorPool(c.validatorPoolAddress, eth.internalClient)
		logAndEat(logger, err)

		// BToken
		c.bTokenAddress, err = lookup("BToken")
		logAndEat(logger, err)
		if bytes.Equal(c.bTokenAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.bToken, err = bindings.NewBToken(c.bTokenAddress, eth.internalClient)
		logAndEat(logger, err)

		// AToken
		c.aTokenAddress, err = lookup("AToken")
		logAndEat(logger, err)
		if bytes.Equal(c.aTokenAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.aToken, err = bindings.NewAToken(c.aTokenAddress, eth.internalClient)
		logAndEat(logger, err)

		// PublicStaking
		c.publicStakingAddress, err = lookup("PublicStaking")
		logAndEat(logger, err)
		if bytes.Equal(c.publicStakingAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.publicStaking, err = bindings.NewPublicStaking(c.publicStakingAddress, eth.internalClient)
		logAndEat(logger, err)

		// ValidatorStaking
		c.validatorStakingAddress, err = lookup("ValidatorStaking")
		logAndEat(logger, err)
		if bytes.Equal(c.validatorStakingAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorStaking, err = bindings.NewValidatorStaking(c.validatorStakingAddress, eth.internalClient)
		logAndEat(logger, err)

		// Governance
		c.governanceAddress, err = lookup("Governance")
		logAndEat(logger, err)
		if bytes.Equal(c.governanceAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.governance, err = bindings.NewGovernance(c.governanceAddress, eth.internalClient)
		logAndEat(logger, err)

		// Snapshots
		c.snapshotsAddress, err = lookup("Snapshots")
		logAndEat(logger, err)
		if bytes.Equal(c.snapshotsAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.snapshots, err = bindings.NewSnapshots(c.snapshotsAddress, eth.internalClient)
		logAndEat(logger, err)

		break
	}

	return nil
}

// return all addresses from all contracts in the contract struct
func (c *Contracts) GetAllAddresses() []common.Address {
	var allAddresses []common.Address
	for addr := range c.allAddresses {
		allAddresses = append(allAddresses, addr)
	}
	return allAddresses
}

func (c *Contracts) Ethdkg() bindings.IETHDKG {
	return c.ethdkg
}

func (c *Contracts) EthdkgAddress() common.Address {
	return c.ethdkgAddress
}

func (c *Contracts) AToken() bindings.IAToken {
	return c.aToken
}

func (c *Contracts) ATokenAddress() common.Address {
	return c.aTokenAddress
}

func (c *Contracts) BToken() bindings.IBToken {
	return c.bToken
}

func (c *Contracts) BTokenAddress() common.Address {
	return c.bTokenAddress
}

func (c *Contracts) PublicStaking() bindings.IPublicStaking {
	return c.publicStaking
}

func (c *Contracts) PublicStakingAddress() common.Address {
	return c.publicStakingAddress
}

func (c *Contracts) ValidatorStaking() bindings.IValidatorStaking {
	return c.validatorStaking
}

func (c *Contracts) ValidatorStakingAddress() common.Address {
	return c.validatorStakingAddress
}

func (c *Contracts) ContractFactory() bindings.IAliceNetFactory {
	return c.contractFactory
}

func (c *Contracts) ContractFactoryAddress() common.Address {
	return c.contractFactoryAddress
}

func (c *Contracts) Snapshots() bindings.ISnapshots {
	return c.snapshots
}

func (c *Contracts) SnapshotsAddress() common.Address {
	return c.snapshotsAddress
}

func (c *Contracts) ValidatorPool() bindings.IValidatorPool {
	return c.validatorPool
}

func (c *Contracts) ValidatorPoolAddress() common.Address {
	return c.validatorPoolAddress
}

func (c *Contracts) Governance() bindings.IGovernance {
	return c.governance
}

func (c *Contracts) GovernanceAddress() common.Address {
	return c.governanceAddress
}

// utils function to log an error
func logAndEat(logger *logrus.Logger, err error) {
	if err != nil {
		logger.Error(err)
	}
}

// Auxiliary function to clean the global variables that will allow the
// deployment and bindings of multiple contracts during other unit tests running
// in sequence. DON'T USE THIS FUNCTION OUTSIDE THE UNIT TESTS
func CleanGlobalVariables(t *testing.T) {
	contracts = nil
	once = sync.Once{}
}
