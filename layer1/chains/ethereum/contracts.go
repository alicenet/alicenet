package ethereum

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	ebindings "github.com/alicenet/alicenet/bridge/bindings/ethereum"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

var _ layer1.EthereumContracts = &Contracts{}

// Contracts contains bindings to smart contract system.
type Contracts struct {
	allAddresses            map[common.Address]bool
	client                  layer1.Client
	ethdkg                  ebindings.IETHDKG
	ethdkgAddress           common.Address
	alca                    ebindings.IALCA
	alcaAddress             common.Address
	alcb                    ebindings.IALCB
	alcbAddress             common.Address
	publicStaking           ebindings.IPublicStaking
	publicStakingAddress    common.Address
	validatorStaking        ebindings.IValidatorStaking
	validatorStakingAddress common.Address
	contractFactory         ebindings.IAliceNetFactory
	contractFactoryAddress  common.Address
	snapshots               ebindings.ISnapshots
	snapshotsAddress        common.Address
	validatorPool           ebindings.IValidatorPool
	validatorPoolAddress    common.Address
	governance              ebindings.IGovernance
	governanceAddress       common.Address
	dynamics                ebindings.IDynamics
	dynamicsAddress         common.Address
}

// Set the contractFactoryAddress and looks up for all the contracts that we
// need that were deployed via the factory. It's only executed once. Other call
// to this functions are no-op.
func NewContracts(eth layer1.Client, contractFactoryAddress common.Address) *Contracts {
	newContracts := &Contracts{
		allAddresses:           make(map[common.Address]bool),
		client:                 eth,
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

	eth := c.client
	logger := eth.GetLogger()
	logger.Infof("Looking up smart contracts on Ethereum...")
	for {
		select {
		case <-signals:
			return errors.New("goodBye from lookup contracts")
		case <-time.After(1 * time.Second):
		}

		// Load the contractFactory first
		contractFactory, err := ebindings.NewAliceNetFactory(
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
		utils.LogAndEat(logger, err)
		if bytes.Equal(c.ethdkgAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.ethdkg, err = ebindings.NewETHDKG(c.ethdkgAddress, eth.GetInternalClient())
		utils.LogAndEat(logger, err)

		// ValidatorPool
		c.validatorPoolAddress, err = lookup("ValidatorPool")
		utils.LogAndEat(logger, err)
		if bytes.Equal(c.validatorPoolAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorPool, err = ebindings.NewValidatorPool(
			c.validatorPoolAddress,
			eth.GetInternalClient(),
		)
		utils.LogAndEat(logger, err)

		// ALCB TODO: bring it back once we deploy ALCB c.alcbAddress, err = lookup("ALCB")
		// utils.LogAndEat(logger, err) if bytes.Equal(c.alcbAddress.Bytes(), make([]byte, 20)) {
		//  continue
		// }

		// workaround for now, just putting a random address, we should uncomment the code above
		// once we deploy ALCB
		c.alcbAddress = common.HexToAddress("0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5bc")
		logger.Infof("Lookup up of \"%v\" is 0x%x", "ALCB", c.alcbAddress)
		c.alcb, err = ebindings.NewALCB(c.alcbAddress, eth.GetInternalClient())
		utils.LogAndEat(logger, err)

		// ALCA
		c.alcaAddress, err = lookup("ALCA")
		utils.LogAndEat(logger, err)
		if bytes.Equal(c.alcaAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.alca, err = ebindings.NewALCA(c.alcaAddress, eth.GetInternalClient())
		utils.LogAndEat(logger, err)

		// PublicStaking
		c.publicStakingAddress, err = lookup("PublicStaking")
		utils.LogAndEat(logger, err)
		if bytes.Equal(c.publicStakingAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.publicStaking, err = ebindings.NewPublicStaking(
			c.publicStakingAddress,
			eth.GetInternalClient(),
		)
		utils.LogAndEat(logger, err)

		// ValidatorStaking
		c.validatorStakingAddress, err = lookup("ValidatorStaking")
		utils.LogAndEat(logger, err)
		if bytes.Equal(c.validatorStakingAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorStaking, err = ebindings.NewValidatorStaking(
			c.validatorStakingAddress,
			eth.GetInternalClient(),
		)
		utils.LogAndEat(logger, err)

		// Governance
		// c.governanceAddress, err = lookup("Governance")
		// utils.LogAndEat(logger, err)
		// if bytes.Equal(c.governanceAddress.Bytes(), make([]byte, 20)) {
		// 	continue
		// }

		// workaround for now, just putting a random address, should uncoment above code once we
		// deploy governance
		c.governanceAddress = common.HexToAddress("0x0b1F9c2b7bED6Db83295c7B5158E3806d67eC5be")
		logger.Infof("Lookup up of \"%v\" is 0x%x", "Governance", c.governanceAddress)

		c.governance, err = ebindings.NewGovernance(c.governanceAddress, eth.GetInternalClient())
		utils.LogAndEat(logger, err)

		// Snapshots
		c.snapshotsAddress, err = lookup("Snapshots")
		utils.LogAndEat(logger, err)
		if bytes.Equal(c.snapshotsAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.snapshots, err = ebindings.NewSnapshots(c.snapshotsAddress, eth.GetInternalClient())
		utils.LogAndEat(logger, err)

		// Dynamics
		c.dynamicsAddress, err = lookup("Dynamics")
		utils.LogAndEat(logger, err)
		if bytes.Equal(c.dynamicsAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.dynamics, err = ebindings.NewDynamics(c.dynamicsAddress, eth.GetInternalClient())
		utils.LogAndEat(logger, err)

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

func (c *Contracts) Ethdkg() ebindings.IETHDKG {
	return c.ethdkg
}

func (c *Contracts) EthdkgAddress() common.Address {
	return c.ethdkgAddress
}

func (c *Contracts) ALCA() ebindings.IALCA {
	return c.alca
}

func (c *Contracts) ALCAAddress() common.Address {
	return c.alcaAddress
}

func (c *Contracts) ALCB() ebindings.IALCB {
	return c.alcb
}

func (c *Contracts) ALCBAddress() common.Address {
	return c.alcbAddress
}

func (c *Contracts) Dynamics() ebindings.IDynamics {
	return c.dynamics
}

func (c *Contracts) DynamicsAddress() common.Address {
	return c.dynamicsAddress
}

func (c *Contracts) PublicStaking() ebindings.IPublicStaking {
	return c.publicStaking
}

func (c *Contracts) PublicStakingAddress() common.Address {
	return c.publicStakingAddress
}

func (c *Contracts) ValidatorStaking() ebindings.IValidatorStaking {
	return c.validatorStaking
}

func (c *Contracts) ValidatorStakingAddress() common.Address {
	return c.validatorStakingAddress
}

func (c *Contracts) ContractFactory() ebindings.IAliceNetFactory {
	return c.contractFactory
}

func (c *Contracts) ContractFactoryAddress() common.Address {
	return c.contractFactoryAddress
}

func (c *Contracts) Snapshots() ebindings.ISnapshots {
	return c.snapshots
}

func (c *Contracts) SnapshotsAddress() common.Address {
	return c.snapshotsAddress
}

func (c *Contracts) ValidatorPool() ebindings.IValidatorPool {
	return c.validatorPool
}

func (c *Contracts) ValidatorPoolAddress() common.Address {
	return c.validatorPoolAddress
}

func (c *Contracts) Governance() ebindings.IGovernance {
	return c.governance
}

func (c *Contracts) GovernanceAddress() common.Address {
	return c.governanceAddress
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
