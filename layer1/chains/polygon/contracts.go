package polygon

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	mbindings "github.com/alicenet/alicenet/bridge/bindings/multichain"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/common"
)

var _ layer1.MultichainContracts = &Contracts{}

// Contracts contains bindings to smart contract system.
type Contracts struct {
	allAddresses           map[common.Address]bool
	client                 layer1.Client
	contractFactory        mbindings.IAliceNetFactory
	contractFactoryAddress common.Address
	lightSnapshots         mbindings.ILightSnapshots
	lightSnapshotsAddress  common.Address
}

// Set the contractFactoryAddress and looks up for all the contracts that we
// need that were deployed via the factory. It's only executed once. Other call
// to this functions are no-op.
func NewContracts(client layer1.Client, contractFactoryAddress common.Address) *Contracts {
	newContracts := &Contracts{
		allAddresses:           make(map[common.Address]bool),
		client:                 client,
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
		contractFactory, err := mbindings.NewAliceNetFactory(
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

		// LightSnapshots
		c.lightSnapshotsAddress, err = lookup("LightSnapshots")
		utils.LogAndEat(logger, err)
		if bytes.Equal(c.lightSnapshotsAddress.Bytes(), make([]byte, 20)) {
			continue
		}

		c.lightSnapshots, err = mbindings.NewLightSnapshots(c.lightSnapshotsAddress, eth.GetInternalClient())
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

func (c *Contracts) ContractFactory() mbindings.IAliceNetFactory {
	return c.contractFactory
}

func (c *Contracts) ContractFactoryAddress() common.Address {
	return c.contractFactoryAddress
}

func (c *Contracts) LightSnapshots() mbindings.ILightSnapshots {
	return c.lightSnapshots
}

func (c *Contracts) LightSnapshotsAddress() common.Address {
	return c.lightSnapshotsAddress
}
