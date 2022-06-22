package events

import (
	"fmt"
	"strings"

	"github.com/MadBase/MadNet/bridge/bindings"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/layer1"
	"github.com/MadBase/MadNet/layer1/executor/tasks"
	monInterfaces "github.com/MadBase/MadNet/layer1/monitor/interfaces"
	"github.com/MadBase/MadNet/layer1/monitor/objects"
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

func RegisterETHDKGEvents(em *objects.EventMap, monDB *db.Database, adminHandler monInterfaces.AdminHandler, taskRequestChan chan<- tasks.Task, taskKillChan chan<- string) {
	ethDkgEvents := GetETHDKGEvents()

	eventProcessorMap := make(map[string]objects.EventProcessor)
	eventProcessorMap["RegistrationOpened"] = func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessRegistrationOpened(eth, logger, log, state, monDB, taskRequestChan)
	}
	eventProcessorMap["AddressRegistered"] = func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessAddressRegistered(eth, logger, log, monDB)
	}
	eventProcessorMap["RegistrationComplete"] = func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessRegistrationComplete(eth, logger, log, monDB, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["SharesDistributed"] = func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessShareDistribution(eth, logger, log, monDB)
	}
	eventProcessorMap["ShareDistributionComplete"] = func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessShareDistributionComplete(eth, logger, log, monDB, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["KeyShareSubmitted"] = func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessKeyShareSubmitted(eth, logger, log, monDB)
	}
	eventProcessorMap["KeyShareSubmissionComplete"] = func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessKeyShareSubmissionComplete(eth, logger, log, monDB, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["MPKSet"] = func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessMPKSet(eth, logger, log, adminHandler, monDB, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["ValidatorMemberAdded"] = func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorMemberAdded(eth, logger, state, log, monDB)
	}
	eventProcessorMap["GPKJSubmissionComplete"] = func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessGPKJSubmissionComplete(eth, logger, log, monDB, taskRequestChan, taskKillChan)
	}
	eventProcessorMap["ValidatorSetCompleted"] = func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorSetCompleted(eth, logger, state, log, monDB, adminHandler)
	}

	// actually register the events
	for eventName, processor := range eventProcessorMap {
		// get the event from the ABI
		event, ok := ethDkgEvents[eventName]
		if !ok {
			panic(fmt.Errorf("%v event not found in ABI", eventName))
		}
		// register it
		if err := em.Register(event.ID.String(), eventName, processor); err != nil {
			panic(fmt.Errorf("could not register event %v", eventName))
		}
	}
}

func SetupEventMap(em *objects.EventMap, cdb *db.Database, monDB *db.Database, adminHandler monInterfaces.AdminHandler, depositHandler monInterfaces.DepositHandler, taskRequestChan chan<- tasks.Task, taskKillChan chan<- string) error {

	RegisterETHDKGEvents(em, monDB, adminHandler, taskRequestChan, taskKillChan)

	// MadByte.DepositReceived
	mbEvents := GetBTokenEvents()
	depositReceived, ok := mbEvents["DepositReceived"]
	if !ok {
		panic("could not find event MadByte.DepositReceived")
	}

	if err := em.Register(depositReceived.ID.String(), depositReceived.Name,
		func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return ProcessDepositReceived(eth, logger, log, cdb, monDB, depositHandler)
		}); err != nil {
		return err
	}

	// Snapshots.SnapshotTaken
	snapshotsEvents := GetSnapshotEvents()
	snapshotTakenEvent, ok := snapshotsEvents["SnapshotTaken"]
	if !ok {
		panic("could not find event Snapshots.SnapshotTaken")
	}

	if err := em.Register(snapshotTakenEvent.ID.String(), snapshotTakenEvent.Name,
		func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
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

	if err := em.Register(valueUpdatedEvent.ID.String(), valueUpdatedEvent.Name,
		func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return ProcessValueUpdated(eth, logger, log, monDB)
		}); err != nil {
		return err
	}

	// ValidatorPool.ValidatorMinorSlashed
	vpEvents := GetValidatorPoolEvents()

	// possible validator joined
	validatorJoinedEvent, ok := vpEvents["ValidatorJoined"]
	if !ok {
		panic("could not find event ValidatorPool.ValidatorJoined")
	}

	processValidatorJoinedFunc := func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorJoined(eth, logger, state, log)
	}
	if err := em.Register(validatorJoinedEvent.ID.String(), validatorJoinedEvent.Name, processValidatorJoinedFunc); err != nil {
		panic(fmt.Sprintf("couldn't register validator joined event:%v", err))
	}

	// possible validator left
	validatorLeftEvent, ok := vpEvents["ValidatorLeft"]
	if !ok {
		panic("could not find event ValidatorPool.ValidatorLeft")
	}

	processValidatorLeftFunc := func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorLeft(eth, logger, state, log)
	}
	if err := em.Register(validatorLeftEvent.ID.String(), validatorLeftEvent.Name, processValidatorLeftFunc); err != nil {
		panic(fmt.Sprintf("couldn't register validator left event:%v", err))
	}

	validatorMinorSlashedEvent, ok := vpEvents["ValidatorMinorSlashed"]
	if !ok {
		panic("could not find event ValidatorPool.ValidatorMinorSlashed")
	}

	processValidatorMinorSlashedFunc := func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorMinorSlashed(eth, logger, state, log)
	}
	if err := em.Register(validatorMinorSlashedEvent.ID.String(), validatorMinorSlashedEvent.Name, processValidatorMinorSlashedFunc); err != nil {
		panic(fmt.Sprintf("couldn't register validator minor slashed event:%v", err))
	}

	// ValidatorPool.ValidatorMajorSlashed
	validatorMajorSlashedEvent, ok := vpEvents["ValidatorMajorSlashed"]
	if !ok {
		panic("could not find event ValidatorPool.ValidatorMajorSlashed")
	}

	processValidatorMajorSlashedFunc := func(eth layer1.Client, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorMajorSlashed(eth, logger, state, log)
	}
	if err := em.Register(validatorMajorSlashedEvent.ID.String(), validatorMajorSlashedEvent.Name, processValidatorMajorSlashedFunc); err != nil {
		panic(err)
	}

	return nil
}
