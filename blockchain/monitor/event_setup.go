package monitor

import (
	"fmt"
	"strings"

	"github.com/alicenet/alicenet/blockchain/dkg/dkgevents"
	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/blockchain/monitor/monevents"
	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func GetETHDKGEvents() map[string]abi.Event {
	ethDkgABI, err := abi.JSON(strings.NewReader(bindings.ETHDKGMetaData.ABI))
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

func GetBTokenEvents() map[string]abi.Event {
	bTokenABI, err := abi.JSON(strings.NewReader(bindings.BTokenMetaData.ABI))
	if err != nil {
		panic(err)
	}

	return bTokenABI.Events
}

func GetSnapshotEvents() map[string]abi.Event {
	snapshotsABI, err := abi.JSON(strings.NewReader(bindings.SnapshotsMetaData.ABI))
	if err != nil {
		panic(err)
	}

	return snapshotsABI.Events
}

func GetValidatorPoolEvents() map[string]abi.Event {
	validatorPoolABI, err := abi.JSON(strings.NewReader(bindings.ValidatorPoolMetaData.ABI))
	if err != nil {
		panic(err)
	}

	return validatorPoolABI.Events
}

func GetPublicStakingEvents() map[string]abi.Event {
	publicStakingABI, err := abi.JSON(strings.NewReader(bindings.PublicStakingMetaData.ABI))
	if err != nil {
		panic(err)
	}

	return publicStakingABI.Events
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
	mbEvents := GetBTokenEvents()
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

	// ValidatorPool.ValidatorMinorSlashed
	vpEvents := GetValidatorPoolEvents()
	validatorMinorSlashedEvent, ok := vpEvents["ValidatorMinorSlashed"]
	if !ok {
		panic("could not find event ValidatorPool.ValidatorMinorSlashed")
	}

	if err := em.RegisterLocked(validatorMinorSlashedEvent.ID.String(), validatorMinorSlashedEvent.Name, monevents.ProcessValidatorMinorSlashed); err != nil {
		panic(err)
	}

	// ValidatorPool.ValidatorMajorSlashed
	validatorMajorSlashedEvent, ok := vpEvents["ValidatorMajorSlashed"]
	if !ok {
		panic("could not find event ValidatorPool.ValidatorMajorSlashed")
	}

	if err := em.RegisterLocked(validatorMajorSlashedEvent.ID.String(), validatorMajorSlashedEvent.Name, monevents.ProcessValidatorMajorSlashed); err != nil {
		panic(err)
	}

	return nil
}
