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
	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

var _ layer1.EthereumContracts = &Contracts{}

// Contracts contains bindings to smart contract system.
type Contracts struct {
	allAddresses                          map[common.Address]bool
	eth                                   *Client
	ethdkg                                bindings.IETHDKG
	ethdkgAddress                         common.Address
	aToken                                bindings.IAToken
	aTokenAddress                         common.Address
	bToken                                bindings.IBToken
	bTokenAddress                         common.Address
	publicStaking                         bindings.IPublicStaking
	publicStakingAddress                  common.Address
	validatorStaking                      bindings.IValidatorStaking
	validatorStakingAddress               common.Address
	contractFactory                       bindings.IAliceNetFactory
	contractFactoryAddress                common.Address
	snapshots                             bindings.ISnapshots
	snapshotsAddress                      common.Address
	validatorPool                         bindings.IValidatorPool
	validatorPoolAddress                  common.Address
	governance                            bindings.IGovernance
	governanceAddress                     common.Address
	accusationMultipleProposal            bindings.IAccusationMultipleProposal
	accusationMultipleProposalAddress     common.Address
	accusationInvalidTxConsumption        bindings.IAccusationInvalidTxConsumption
	accusationInvalidTxConsumptionAddress common.Address
	dynamics                              bindings.IDynamics
	dynamicsAddress                       common.Address
}

// Set the contractFactoryAddress and looks up for all the contracts that we
// need that were deployed via the factory. It's only executed once. Other call
// to this functions are no-op.
func NewContracts(eth *Client, contractFactoryAddress common.Address) *Contracts {
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
	logger := eth.logger
	logger.Infof("Looking up smart contracts on Ethereum...")
	for {
		select {
		case <-signals:
			return errors.New("goodBye from lookup contracts")
		case <-time.After(1 * time.Second):
		}

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
			continue
		}

		// ETHDKG
		c.ethdkgAddress, err = c.lookupString(callOpts, "ETHDKG")
		logAndEat(logger, err)
		if bytes.Equal(c.ethdkgAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.ethdkg, err = bindings.NewETHDKG(c.ethdkgAddress, eth.internalClient)
		logAndEat(logger, err)

		// ValidatorPool
		c.validatorPoolAddress, err = c.lookupString(callOpts, "ValidatorPool")
		logAndEat(logger, err)
		if bytes.Equal(c.validatorPoolAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorPool, err = bindings.NewValidatorPool(c.validatorPoolAddress, eth.internalClient)
		logAndEat(logger, err)

		// BToken
		c.bTokenAddress, err = c.lookupString(callOpts, "BToken")
		logAndEat(logger, err)
		if bytes.Equal(c.bTokenAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.bToken, err = bindings.NewBToken(c.bTokenAddress, eth.internalClient)
		logAndEat(logger, err)

		// AToken
		c.aTokenAddress, err = c.lookupString(callOpts, "AToken")
		logAndEat(logger, err)
		if bytes.Equal(c.aTokenAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.aToken, err = bindings.NewAToken(c.aTokenAddress, eth.internalClient)
		logAndEat(logger, err)

		// PublicStaking
		c.publicStakingAddress, err = c.lookupString(callOpts, "PublicStaking")
		logAndEat(logger, err)
		if bytes.Equal(c.publicStakingAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.publicStaking, err = bindings.NewPublicStaking(c.publicStakingAddress, eth.internalClient)
		logAndEat(logger, err)

		// ValidatorStaking
		c.validatorStakingAddress, err = c.lookupString(callOpts, "ValidatorStaking")
		logAndEat(logger, err)
		if bytes.Equal(c.validatorStakingAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.validatorStaking, err = bindings.NewValidatorStaking(c.validatorStakingAddress, eth.internalClient)
		logAndEat(logger, err)

		// Governance
		c.governanceAddress, err = c.lookupString(callOpts, "Governance")
		logAndEat(logger, err)
		if bytes.Equal(c.governanceAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.governance, err = bindings.NewGovernance(c.governanceAddress, eth.internalClient)
		logAndEat(logger, err)

		// Snapshots
		c.snapshotsAddress, err = c.lookupString(callOpts, "Snapshots")
		logAndEat(logger, err)
		if bytes.Equal(c.snapshotsAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.snapshots, err = bindings.NewSnapshots(c.snapshotsAddress, eth.internalClient)
		logAndEat(logger, err)

		// Dynamics
		c.dynamicsAddress, err = c.lookupString(callOpts, "Dynamics")
		logAndEat(logger, err)
		if bytes.Equal(c.dynamicsAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.dynamics, err = bindings.NewDynamics(c.dynamicsAddress, eth.internalClient)
		logAndEat(logger, err)

		// AccusationMultipleProposal
		contractSaltComponents := []string{"AccusationMultipleProposal", "Accusation"}
		c.accusationMultipleProposalAddress, err = c.lookupRoleBasedSalt(callOpts, contractSaltComponents)
		logAndEat(logger, err)
		if bytes.Equal(c.accusationMultipleProposalAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.accusationMultipleProposal, err = bindings.NewAccusationMultipleProposal(c.accusationMultipleProposalAddress, eth.internalClient)
		logAndEat(logger, err)

		// AccusationInvalidTxConsumption
		contractSaltComponents = []string{"AccusationInvalidTxConsumption", "Accusation"}
		c.accusationInvalidTxConsumptionAddress, err = c.lookupRoleBasedSalt(callOpts, contractSaltComponents)
		logAndEat(logger, err)
		if bytes.Equal(c.accusationInvalidTxConsumptionAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.accusationInvalidTxConsumption, err = bindings.NewAccusationInvalidTxConsumption(c.accusationInvalidTxConsumptionAddress, eth.internalClient)
		logAndEat(logger, err)

		break
	}

	return nil
}

func (c *Contracts) lookupString(callOpts *bind.CallOpts, name string) (common.Address, error) {
	salt := utils.StringToBytes32(name)
	addr, err := c.lookupBytes32(callOpts, salt)

	if err != nil {
		c.eth.logger.Errorf("Failed lookup of contract \"%v\": %v", name, err)
		return common.Address{}, err
	}

	c.eth.logger.Infof("Lookup up of contract \"%v\" is 0x%x", name, addr)
	return addr, nil
}

func (c *Contracts) lookupRoleBasedSalt(callOpts *bind.CallOpts, saltComponents []string) (common.Address, error) {
	salt := utils.CalculateSaltFromComponents(saltComponents)
	addr, err := c.lookupBytes32(callOpts, salt)

	if err != nil {
		c.eth.logger.Errorf("Failed lookup of contract \"%v\" with salt \"0x%x\": %v", saltComponents[0], salt, err)
		return common.Address{}, err
	}

	c.eth.logger.Infof("Lookup up of contract \"%v\" is 0x%x", saltComponents[0], addr)
	return addr, nil
}

func (c *Contracts) lookupBytes32(callOpts *bind.CallOpts, salt [32]byte) (common.Address, error) {
	addr, err := c.contractFactory.Lookup(callOpts, salt)
	if err != nil {
		c.eth.logger.Errorf("Failed lookup of salt \"0x%x\": %v", salt, err)
		return common.Address{}, err
	}

	c.allAddresses[addr] = true
	return addr, nil
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

func (c *Contracts) MultipleProposalAccusation() bindings.IAccusationMultipleProposal {
	return c.accusationMultipleProposal
}

func (c *Contracts) MultipleProposalAccusationAddress() common.Address {
	return c.accusationMultipleProposalAddress
}

func (c *Contracts) AccusationInvalidTxConsumption() bindings.IAccusationInvalidTxConsumption {
	return c.accusationInvalidTxConsumption
}

func (c *Contracts) AccusationInvalidTxConsumptionAddress() common.Address {
	return c.accusationInvalidTxConsumptionAddress
}

// utils function to log an error.
func logAndEat(logger *logrus.Logger, err error) {
	if err != nil {
		logger.Error(err)
	}
}
