package events

import (
	"fmt"
	"strings"

	ethereumInterfaces "github.com/MadBase/MadNet/blockchain/ethereum/interfaces"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/monitor/objects"
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
	eventProcessorMap["RegistrationOpened"] = func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessRegistrationOpened(eth, logger, log, cdb, taskRequestChan)
	}
	eventProcessorMap["AddressRegistered"] = func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessAddressRegistered(eth, logger, log, cdb)
	}
	eventProcessorMap["RegistrationComplete"] = func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessRegistrationComplete(eth, logger, log, cdb, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["SharesDistributed"] = func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessShareDistribution(eth, logger, log, cdb)
	}
	eventProcessorMap["ShareDistributionComplete"] = func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessShareDistributionComplete(eth, logger, log, cdb, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["KeyShareSubmitted"] = func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessKeyShareSubmitted(eth, logger, log, cdb)
	}
	eventProcessorMap["KeyShareSubmissionComplete"] = func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessKeyShareSubmissionComplete(eth, logger, log, cdb, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["MPKSet"] = func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessMPKSet(eth, logger, log, adminHandler, cdb, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["ValidatorMemberAdded"] = func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorMemberAdded(eth, logger, state, log, cdb)
	}
	eventProcessorMap["GPKJSubmissionComplete"] = func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessGPKJSubmissionComplete(eth, logger, log, cdb, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["ValidatorSetCompleted"] = func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorSetCompleted(eth, logger, state, log, cdb, adminHandler)
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
		func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return ProcessDepositReceived(eth, logger, log, cdb, depositHandler)
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
		func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return ProcessSnapshotTaken(eth, logger, log, adminHandler)
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
		func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return ProcessValueUpdated(eth, logger, log, cdb)
		}); err != nil {
		return err
	}

	// ValidatorPool.ValidatorMinorSlashed
	vpEvents := GetValidatorPoolEvents()
	validatorMinorSlashedEvent, ok := vpEvents["ValidatorMinorSlashed"]
	if !ok {
		panic("could not find event ValidatorPool.ValidatorMinorSlashed")
	}

	processValidatorMinorSlashedFunc := func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorMinorSlashed(eth, logger, log)
	}
	if err := em.RegisterLocked(validatorMinorSlashedEvent.ID.String(), validatorMinorSlashedEvent.Name, processValidatorMinorSlashedFunc); err != nil {
		panic(err)
	}

	// ValidatorPool.ValidatorMajorSlashed
	validatorMajorSlashedEvent, ok := vpEvents["ValidatorMajorSlashed"]
	if !ok {
		panic("could not find event ValidatorPool.ValidatorMajorSlashed")
	}

	processValidatorMajorSlashedFunc := func(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorMajorSlashed(eth, logger, log)
	}
	if err := em.RegisterLocked(validatorMajorSlashedEvent.ID.String(), validatorMajorSlashedEvent.Name, processValidatorMajorSlashedFunc); err != nil {
		panic(err)
	}

	return nil
}
