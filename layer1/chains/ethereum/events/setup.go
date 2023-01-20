package events

import (
	"fmt"
	"strings"

	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor"
	monInterfaces "github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/alicenet/alicenet/layer1/monitor/objects"
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

func GetALCBEvents() map[string]abi.Event {
	alcbABI, err := abi.JSON(strings.NewReader(bindings.ALCBMetaData.ABI))
	if err != nil {
		panic(err)
	}

	return alcbABI.Events
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

func GetDynamicsEvents() map[string]abi.Event {
	snapshotsABI, err := abi.JSON(strings.NewReader(bindings.DynamicsMetaData.ABI))
	if err != nil {
		panic(err)
	}

	return snapshotsABI.Events
}

func RegisterETHDKGEvents(em *objects.EventMap, monDB *db.Database, adminHandler monInterfaces.AdminHandler, taskHandler executor.TaskHandler) {
	ethDkgEvents := GetETHDKGEvents()

	eventProcessorMap := make(map[string]objects.EventProcessor)

	eventProcessorMap["RegistrationOpened"] = func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessRegistrationOpened(eth, contracts, logger, log, state, monDB, taskHandler)
	}
	eventProcessorMap["AddressRegistered"] = func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessAddressRegistered(contracts, logger, log, monDB)
	}
	eventProcessorMap["RegistrationComplete"] = func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessRegistrationComplete(contracts, logger, log, monDB, taskHandler)
	}
	eventProcessorMap["SharesDistributed"] = func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessShareDistribution(contracts, logger, log, monDB)
	}
	eventProcessorMap["ShareDistributionComplete"] = func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessShareDistributionComplete(contracts, logger, log, monDB, taskHandler)
	}
	eventProcessorMap["KeyShareSubmitted"] = func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessKeyShareSubmitted(contracts, logger, log, monDB)
	}
	eventProcessorMap["KeyShareSubmissionComplete"] = func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessKeyShareSubmissionComplete(contracts, logger, log, monDB, taskHandler)
	}
	eventProcessorMap["MPKSet"] = func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessMPKSet(contracts, logger, log, adminHandler, monDB, taskHandler)
	}
	eventProcessorMap["ValidatorMemberAdded"] = func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorMemberAdded(eth, contracts, logger, state, log, monDB)
	}
	eventProcessorMap["GPKJSubmissionComplete"] = func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessGPKJSubmissionComplete(contracts, logger, log, monDB, taskHandler)
	}
	eventProcessorMap["ValidatorSetCompleted"] = func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorSetCompleted(eth, contracts, logger, state, log, monDB, adminHandler)
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

func SetupEventMap(em *objects.EventMap, cdb, monDB *db.Database, adminHandler monInterfaces.AdminHandler, depositHandler monInterfaces.DepositHandler, taskHandler executor.TaskHandler, exitFunc func(), chainID uint32) error {
	RegisterETHDKGEvents(em, monDB, adminHandler, taskHandler)

	// MadByte.DepositReceived
	mbEvents := GetALCBEvents()
	depositReceived, ok := mbEvents["DepositReceived"]
	if !ok {
		panic("could not find event MadByte.DepositReceived")
	}

	if err := em.Register(depositReceived.ID.String(), depositReceived.Name,
		func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return ProcessDepositReceived(eth, contracts, logger, log, cdb, monDB, depositHandler, chainID)
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
		func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return ProcessSnapshotTaken(contracts, logger, log, adminHandler, taskHandler)
		}); err != nil {
		return err
	}

	// Governance.ValueUpdated
	govEvents := GetGovernanceEvents()
	snapshotTakenOldEvent, ok := govEvents["SnapshotTaken"]
	if !ok {
		panic("could not find event Snapshots.SnapshotTakenOld")
	}

	if err := em.Register(snapshotTakenOldEvent.ID.String(), snapshotTakenOldEvent.Name,
		func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return ProcessSnapshotTakenOld(eth, contracts, logger, log, adminHandler, taskHandler)
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

	processValidatorJoinedFunc := func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorJoined(eth, contracts, logger, state, log)
	}
	if err := em.Register(validatorJoinedEvent.ID.String(), validatorJoinedEvent.Name, processValidatorJoinedFunc); err != nil {
		panic(fmt.Sprintf("couldn't register validator joined event:%v", err))
	}

	// possible validator left
	validatorLeftEvent, ok := vpEvents["ValidatorLeft"]
	if !ok {
		panic("could not find event ValidatorPool.ValidatorLeft")
	}

	processValidatorLeftFunc := func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorLeft(eth, contracts, logger, state, log)
	}
	if err := em.Register(validatorLeftEvent.ID.String(), validatorLeftEvent.Name, processValidatorLeftFunc); err != nil {
		panic(fmt.Sprintf("couldn't register validator left event:%v", err))
	}

	validatorMinorSlashedEvent, ok := vpEvents["ValidatorMinorSlashed"]
	if !ok {
		panic("could not find event ValidatorPool.ValidatorMinorSlashed")
	}

	processValidatorMinorSlashedFunc := func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorMinorSlashed(eth, contracts, logger, state, log)
	}
	if err := em.Register(validatorMinorSlashedEvent.ID.String(), validatorMinorSlashedEvent.Name, processValidatorMinorSlashedFunc); err != nil {
		panic(fmt.Sprintf("couldn't register validator minor slashed event:%v", err))
	}

	// ValidatorPool.ValidatorMajorSlashed
	validatorMajorSlashedEvent, ok := vpEvents["ValidatorMajorSlashed"]
	if !ok {
		panic("could not find event ValidatorPool.ValidatorMajorSlashed")
	}

	processValidatorMajorSlashedFunc := func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return ProcessValidatorMajorSlashed(eth, contracts, logger, state, log)
	}
	if err := em.Register(validatorMajorSlashedEvent.ID.String(), validatorMajorSlashedEvent.Name, processValidatorMajorSlashedFunc); err != nil {
		panic(err)
	}

	dynamicsEvents := GetDynamicsEvents()
	newAliceNetNodeVersionAvailableEvent, ok := dynamicsEvents["NewAliceNetNodeVersionAvailable"]
	if !ok {
		panic("could not find event Dynamics.NewAliceNetNodeVersionAvailable")
	}

	if err := em.Register(newAliceNetNodeVersionAvailableEvent.ID.String(), newAliceNetNodeVersionAvailableEvent.Name,
		func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return ProcessNewAliceNetNodeVersionAvailable(contracts, logger, log, state, taskHandler)
		}); err != nil {
		return err
	}

	newCanonicalAliceNetNodeVersionEvent, ok := dynamicsEvents["NewCanonicalAliceNetNodeVersion"]
	if !ok {
		panic("could not find event Dynamics.NewCanonicalAliceNetNodeVersion")
	}

	if err := em.Register(newCanonicalAliceNetNodeVersionEvent.ID.String(), newCanonicalAliceNetNodeVersionEvent.Name,
		func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return ProcessNewCanonicalAliceNetNodeVersion(contracts, logger, log, exitFunc)
		}); err != nil {
		return err
	}

	dynamicValueChangedEvent, ok := dynamicsEvents["DynamicValueChanged"]
	if !ok {
		panic("could not find event Dynamics.DynamicValueChanged")
	}

	if err := em.Register(dynamicValueChangedEvent.ID.String(), dynamicValueChangedEvent.Name,
		func(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return ProcessDynamicValueChanged(contracts, logger, log, adminHandler)
		}); err != nil {
		return err
	}

	return nil
}
