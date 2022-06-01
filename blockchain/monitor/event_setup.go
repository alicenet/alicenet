package monitor

import (
	"fmt"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/dkgevents"
	"strings"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/monitor/monevents"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/bridge/bindings"
	"github.com/MadBase/MadNet/consensus/db"
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

func RegisterETHDKGEvents(em *objects.EventMap, cdb *db.Database, adminHandler interfaces.AdminHandler, taskRequestChan chan<- interfaces.ITask, taskKillChan chan<- string) {
	ethDkgEvents := GetETHDKGEvents()

	eventProcessorMap := make(map[string]objects.EventProcessor)
	eventProcessorMap["RegistrationOpened"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return dkgevents.ProcessRegistrationOpened(eth, logger, log, cdb, taskRequestChan)
	}
	eventProcessorMap["AddressRegistered"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return dkgevents.ProcessAddressRegistered(eth, logger, log, cdb)
	}
	eventProcessorMap["RegistrationComplete"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return dkgevents.ProcessRegistrationComplete(eth, logger, log, cdb, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["SharesDistributed"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return dkgevents.ProcessShareDistribution(eth, logger, log, cdb)
	}
	eventProcessorMap["ShareDistributionComplete"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return dkgevents.ProcessShareDistributionComplete(eth, logger, log, cdb, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["KeyShareSubmitted"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return dkgevents.ProcessKeyShareSubmitted(eth, logger, log, cdb)
	}
	eventProcessorMap["KeyShareSubmissionComplete"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return dkgevents.ProcessKeyShareSubmissionComplete(eth, logger, log, cdb, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["MPKSet"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return dkgevents.ProcessMPKSet(eth, logger, log, adminHandler, cdb, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["ValidatorMemberAdded"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return monevents.ProcessValidatorMemberAdded(eth, logger, state, log, cdb)
	}
	eventProcessorMap["GPKJSubmissionComplete"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return dkgevents.ProcessGPKJSubmissionComplete(eth, logger, log, cdb, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["ValidatorSetCompleted"] = func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return monevents.ProcessValidatorSetCompleted(eth, logger, state, log, cdb, adminHandler)
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

func SetupEventMap(em *objects.EventMap, cdb *db.Database, adminHandler interfaces.AdminHandler, depositHandler interfaces.DepositHandler, taskRequestChan chan<- interfaces.ITask, taskKillChan chan<- string) error {

	RegisterETHDKGEvents(em, cdb, adminHandler, taskRequestChan, taskKillChan)

	// MadByte.DepositReceived
	mbEvents := GetBTokenEvents()
	depositReceived, ok := mbEvents["DepositReceived"]
	if !ok {
		panic("could not find event MadByte.DepositReceived")
	}

	if err := em.RegisterLocked(depositReceived.ID.String(), depositReceived.Name,
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessDepositReceived(eth, logger, log, cdb, depositHandler)
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
			return monevents.ProcessSnapshotTaken(eth, logger, log, adminHandler)
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
			return monevents.ProcessValueUpdated(eth, logger, log, cdb)
		}); err != nil {
		return err
	}

	// ValidatorPool.ValidatorMinorSlashed
	vpEvents := GetValidatorPoolEvents()
	validatorMinorSlashedEvent, ok := vpEvents["ValidatorMinorSlashed"]
	if !ok {
		panic("could not find event ValidatorPool.ValidatorMinorSlashed")
	}

	processValidatorMinorSlashedFunc := func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return monevents.ProcessValidatorMinorSlashed(eth, logger, log)
	}
	if err := em.RegisterLocked(validatorMinorSlashedEvent.ID.String(), validatorMinorSlashedEvent.Name, processValidatorMinorSlashedFunc); err != nil {
		panic(err)
	}

	// ValidatorPool.ValidatorMajorSlashed
	validatorMajorSlashedEvent, ok := vpEvents["ValidatorMajorSlashed"]
	if !ok {
		panic("could not find event ValidatorPool.ValidatorMajorSlashed")
	}

	processValidatorMajorSlashedFunc := func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return monevents.ProcessValidatorMajorSlashed(eth, logger, log)
	}
	if err := em.RegisterLocked(validatorMajorSlashedEvent.ID.String(), validatorMajorSlashedEvent.Name, processValidatorMajorSlashedFunc); err != nil {
		panic(err)
	}

	return nil
}
