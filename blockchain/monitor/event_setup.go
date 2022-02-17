package monitor

import (
	"fmt"
	"strings"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgevents"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/monitor/monevents"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/bridge/bindings"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func GetETHDKGEvents() map[string]abi.Event {
	ethDkgABI, err := abi.JSON(strings.NewReader(bindings.IETHDKGEventsMetaData.ABI))
	if err != nil {
		panic(err)
	}

	return ethDkgABI.Events
}

func RegisterETHDKGEvents(em *objects.EventMap, adminHandler interfaces.AdminHandler) {
	ethDkgEvents := GetETHDKGEvents()

	var eventProcessorMap map[string]objects.EventProcessor = make(map[string]objects.EventProcessor)
	eventProcessorMap["RegistrationOpened"] = dkgevents.ProcessRegistrationOpened
	eventProcessorMap["AddressRegistered"] = dkgevents.ProcessAddressRegistered
	eventProcessorMap["RegistrationComplete"] = dkgevents.ProcessRegistrationComplete
	eventProcessorMap["SharesDistributed"] = dkgevents.ProcessShareDistribution
	eventProcessorMap["ShareDistributionComplete"] = dkgevents.ProcessShareDistributionComplete
	eventProcessorMap["KeyShareSubmitted"] = dkgevents.ProcessKeyShareSubmitted
	eventProcessorMap["KeyShareSubmissionComplete"] = dkgevents.ProcessKeyShareSubmissionComplete
	eventProcessorMap["MPKSet"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return dkgevents.ProcessMPKSet(eth, logger, state, log, adminHandler)
	}
	eventProcessorMap["ValidatorMemberAdded"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return monevents.ProcessValidatorMemberAdded(eth, logger, state, log, adminHandler)
	}
	eventProcessorMap["GPKJSubmissionComplete"] = dkgevents.ProcessGPKJSubmissionComplete
	eventProcessorMap["ValidatorSetCompleted"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return monevents.ProcessValidatorSetCompleted(eth, logger, state, log, adminHandler)
	}

	// actually register the events
	for eventName, processor := range eventProcessorMap {
		// get the event from the ABI
		event, ok := ethDkgEvents[eventName]
		if !ok {
			panic(fmt.Errorf("%v event not found in ABI", eventName))
		}
		// register it
		if err := em.RegisterLocked(event.ID.String(), eventName, processor); err != nil {
			panic(fmt.Errorf("could not register event %v", eventName))
		}
	}
}

func SetupEventMap(em *objects.EventMap, cdb *db.Database, adminHandler interfaces.AdminHandler, depositHandler interfaces.DepositHandler) error {

	RegisterETHDKGEvents(em, adminHandler)

	// Events to pass through to side chain
	if err := em.RegisterLocked("0x5b063c6569a91e8133fc6cd71d31a4ca5c65c652fd53ae093f46107754f08541", "DepositReceived",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessDepositReceived(eth, logger, state, log, cdb, depositHandler)
		}); err != nil {
		return err
	}

	if err := em.RegisterLocked("0x6d438b6b835d16cdae6efdc0259fdfba17e6aa32dae81863a2467866f85f724a", "SnapshotTaken",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessSnapshotTaken(eth, logger, state, log, adminHandler)
		}); err != nil {
		return err
	}

	if err := em.RegisterLocked("0x36dcd0e03525dedd9d5c21a263ef5f35d030298b5c48f1a713006aefc064ad05", "ValueUpdated",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessValueUpdated(eth, logger, state, log, adminHandler)
		}); err != nil {
		return err
	}

	// Registering just for informational purposes
	if err := em.RegisterLocked("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925", "DSTokenApproval", nil); err != nil {
		return err
	}
	if err := em.RegisterLocked("0xce241d7ca1f669fee44b6fc00b8eba2df3bb514eed0f6f668f8f89096e81ed94", "LogSetOwner", nil); err != nil {
		return err
	}
	if err := em.RegisterLocked("0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885", "Mint", nil); err != nil {
		return err
	}
	if err := em.RegisterLocked("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", "Transfer", nil); err != nil {
		return err
	}
	if err := em.RegisterLocked("0x8c25e214c5693ebaf8008875bacedeb9e0aafd393864a314ed1801b2a4e13dd9", "ValidatorJoined", nil); err != nil {
		return err
	}
	if err := em.RegisterLocked("0x319bbadb03b94aedc69babb34a28675536a9cb30f4bbde343e1d0018c44ebd94", "ValidatorLeft", nil); err != nil {
		return err
	}
	if err := em.RegisterLocked("0x1de2f07b0a1c69916a8b25b889051644192307ea08444a2e11f8654d1db3ab0c", "LockedStake", nil); err != nil {
		return err
	}

	return nil
}
