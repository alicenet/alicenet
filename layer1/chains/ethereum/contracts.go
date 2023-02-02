package ethereum

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/evm"
	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

var _ layer1.EthereumContracts = &Contracts{}

// Contracts contains bindings to smart contract system.
type Contracts struct {
	allAddresses            map[common.Address]bool
	eth                     *evm.Client
	ethdkg                  bindings.IETHDKG
	ethdkgAddress           common.Address
	alca                    bindings.IALCA
	alcaAddress             common.Address
	alcb                    bindings.IALCB
	alcbAddress             common.Address
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
	dynamics                bindings.IDynamics
	dynamicsAddress         common.Address
}

// Set the contractFactoryAddress and looks up for all the contracts that we
// need that were deployed via the factory. It's only executed once. Other call
// to this functions are no-op.
func NewContracts(eth *evm.Client, contractFactoryAddress common.Address) *Contracts {
	newContracts := &Contracts{
		allAddresses:           make(map[common.Address]bool),
		eth:                    eth,
		contractFactoryAddress: contractFactoryAddress,
	}
	err := newContracts.lookupContracts()
	if err != nil {
		panic(err)
	}
	return newContracts
}

// LookupContracts uses the registry to lookup and create bindings for all required contracts.
func (c *Contracts) lookupContracts() error {
	networkCtx, cf := context.WithCancel(context.Background())
	defer cf()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	eth := c.eth
	logger := eth.GetLogger()
	logger.Infof("Looking up smart contracts on Ethereum...")
	for {
		select {
		case <-signals:
			return errors.New("goodBye from lookup contracts")
		case <-time.After(1 * time.Second):
		}

		// Load the contractFactory first
		contractFactory, err := bindings.NewAliceNetFactory(
			c.contractFactoryAddress,
			eth.GetInternalClient(),
		)
		if err != nil {
			return err
		}
		c.contractFactory = contractFactory

		// todo: replace lookup with deterministic address compute

		callOpts, err := eth.GetCallOpts(networkCtx, eth.GetDefaultAccount())
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

		c.ethdkg, err = bindings.NewETHDKG(c.ethdkgAddress, eth.GetInternalClient())
		logAndEat(logger, err)

		// ValidatorPool
		c.validatorPoolAddress, err = lookup("ValidatorPool")
		logAndEat(logger, err)
		if bytes.Equal(c.validatorPoolAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorPool, err = bindings.NewValidatorPool(
			c.validatorPoolAddress,
			eth.GetInternalClient(),
		)
		logAndEat(logger, err)

		// ALCB
		c.alcbAddress, err = lookup("ALCB")
		logAndEat(logger, err)
		if bytes.Equal(c.alcbAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.alcb, err = bindings.NewALCB(c.alcbAddress, eth.GetInternalClient())
		logAndEat(logger, err)

		// ALCA
		c.alcaAddress, err = lookup("ALCA")
		logAndEat(logger, err)
		if bytes.Equal(c.alcaAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.alca, err = bindings.NewALCA(c.alcaAddress, eth.GetInternalClient())
		logAndEat(logger, err)

		// PublicStaking
		c.publicStakingAddress, err = lookup("PublicStaking")
		logAndEat(logger, err)
		if bytes.Equal(c.publicStakingAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.publicStaking, err = bindings.NewPublicStaking(
			c.publicStakingAddress,
			eth.GetInternalClient(),
		)
		logAndEat(logger, err)

		// ValidatorStaking
		c.validatorStakingAddress, err = lookup("ValidatorStaking")
		logAndEat(logger, err)
		if bytes.Equal(c.validatorStakingAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorStaking, err = bindings.NewValidatorStaking(
			c.validatorStakingAddress,
			eth.GetInternalClient(),
		)
		logAndEat(logger, err)

		// Governance
		c.governanceAddress, err = lookup("Governance")
		logAndEat(logger, err)
		if bytes.Equal(c.governanceAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.governance, err = bindings.NewGovernance(c.governanceAddress, eth.GetInternalClient())
		logAndEat(logger, err)

		// Snapshots
		c.snapshotsAddress, err = lookup("Snapshots")
		logAndEat(logger, err)
		if bytes.Equal(c.snapshotsAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.snapshots, err = bindings.NewSnapshots(c.snapshotsAddress, eth.GetInternalClient())
		logAndEat(logger, err)

		// Dynamics
		c.dynamicsAddress, err = lookup("Dynamics")
		logAndEat(logger, err)
		if bytes.Equal(c.dynamicsAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.dynamics, err = bindings.NewDynamics(c.dynamicsAddress, eth.GetInternalClient())
		logAndEat(logger, err)

		break
	}

	return nil
}

// return all addresses from all contracts in the contract struct.
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

func (c *Contracts) ALCA() bindings.IALCA {
	return c.alca
}

func (c *Contracts) ALCAAddress() common.Address {
	return c.alcaAddress
}

func (c *Contracts) ALCB() bindings.IALCB {
	return c.alcb
}

func (c *Contracts) ALCBAddress() common.Address {
	return c.alcbAddress
}

func (c *Contracts) Dynamics() bindings.IDynamics {
	return c.dynamics
}

func (c *Contracts) DynamicsAddress() common.Address {
	return c.dynamicsAddress
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

// utils function to log an error.
func logAndEat(logger *logrus.Logger, err error) {
	if err != nil {
		logger.Error(err)
	}
}

// Get the current validators.
func GetValidators(
	eth layer1.Client,
	contracts layer1.EthereumContracts,
	logger *logrus.Logger,
	ctx context.Context,
) ([]common.Address, error) {
	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	if err != nil {
		return nil, err
	}
	validatorAddresses, err := contracts.ValidatorPool().GetValidatorsAddresses(callOpts)
	if err != nil {
		logger.Warnf("Could not call contract:%v", err)
		return nil, err
	}

	return validatorAddresses, nil
}
