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

func GetGovernanceEvents() map[string]abi.Event {
	governanceABI, err := abi.JSON(strings.NewReader(bindings.GovernanceMetaData.ABI))
	if err != nil {
		panic(err)
	}

	return governanceABI.Events
}

func GetMadByteEvents() map[string]abi.Event {
	madByteABI, err := abi.JSON(strings.NewReader(bindings.MadByteMetaData.ABI))
	if err != nil {
		panic(err)
	}

	return madByteABI.Events
}

func GetSnapshotEvents() map[string]abi.Event {
	snapshotsABI, err := abi.JSON(strings.NewReader(bindings.SnapshotsMetaData.ABI))
	if err != nil {
		panic(err)
	}

	return snapshotsABI.Events
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

	// MadByte.DepositReceived
	mbEvents := GetMadByteEvents()
	depositReceived, ok := mbEvents["DepositReceived"]
	if !ok {
		panic("could not find event MadByte.DepositReceived")
	}

	if err := em.RegisterLocked(depositReceived.ID.String(), depositReceived.Name,
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessDepositReceived(eth, logger, state, log, cdb, depositHandler)
		}); err != nil {
		return err
	}

	// Snapshots.SnapshotTaken
	snapshotsEvents := GetSnapshotEvents()
	snapshotTakenEvent, ok := snapshotsEvents["SnapshotTaken"]
	if !ok {
		panic("could not find event Snapshots.SnapshotTaken")
	}

	if err := em.RegisterLocked(snapshotTakenEvent.ID.String(), snapshotTakenEvent.Name,
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessSnapshotTaken(eth, logger, state, log, adminHandler)
		}); err != nil {
		return err
	}

	// Governance.ValueUpdated
	govEvents := GetGovernanceEvents()
	valueUpdatedEvent, ok := govEvents["ValueUpdated"]
	if !ok {
		panic("could not find event Governance.ValueUpdated")
	}

	if err := em.RegisterLocked(valueUpdatedEvent.ID.String(), valueUpdatedEvent.Name,
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessValueUpdated(eth, logger, state, log, adminHandler)
		}); err != nil {
		return err
	}

	return nil
}
