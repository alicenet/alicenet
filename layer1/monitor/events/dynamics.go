package events

import (
	"fmt"

	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/dynamics"
	"github.com/alicenet/alicenet/layer1/monitor/objects"
	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// ProcessDynamicValueChanged handles a dynamic value updating coming from our smart contract.
func ProcessDynamicValueChanged(contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log) error {
	logger.Info("ProcessDynamicValueChanged() ...")

	event, err := contracts.EthereumContracts().Dynamics().ParseDynamicValueChanged(log)
	if err != nil {
		return err
	}

	logger = logger.WithFields(logrus.Fields{
		"Epoch": event.Epoch.Uint64(),
		"Value": fmt.Sprintf("0x%x", event.RawDynamicValues),
	})
	// TODO; decode and add the dynamic value in here
	logger.Infof("Value updated")

	return nil
}

func ProcessNewAliceNetNodeVersionAvailable(contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, monState *objects.MonitorState, taskRequestChan chan<- tasks.TaskRequest) error {
	logger = logger.WithField("method", "ProcessNewAliceNetNodeVersionAvailable")
	logger.Info("Processing new AliceNet node version...")

	event, err := contracts.EthereumContracts().Dynamics().ParseNewAliceNetNodeVersionAvailable(log)
	if err != nil {
		return err
	}

	logger = logger.WithFields(logrus.Fields{
		"ExecutionEpoch": event.Version.ExecutionEpoch,
		"Major":          event.Version.Major,
		"Minor":          event.Version.Minor,
		"Patch":          event.Version.Patch,
	})

	monState.CanonicalVersion = event.Version
	logger.Info("New AliceNet node version available!")

	// Killing previous task
	taskRequestChan <- tasks.NewKillTaskRequest(&dynamics.CanonicalVersionCheckTask{})

	// If any element of the new Version is greater, schedule the task
	newMajorIsGreater, newMinorIsGreater, newPatchIsGreater, _, err := utils.CompareCanonicalVersion(event.Version)
	if err != nil {
		return fmt.Errorf("processing new version available: %w", err)
	}

	if newMajorIsGreater || newMinorIsGreater || newPatchIsGreater {
		// Scheduling task with the new Canonical Version
		taskRequestChan <- tasks.NewScheduleTaskRequest(dynamics.NewVersionCheckTask(event.Version))
	}

	return nil
}

func ProcessNewCanonicalAliceNetNodeVersion(
	contracts layer1.AllSmartContracts,
	logger *logrus.Entry,
	log types.Log,
	monState *objects.MonitorState,
	taskRequestChan chan<- tasks.TaskRequest,
	exitFunc func(),
) error {
	logger = logger.WithField("method", "ProcessNewCanonicalAliceNetNodeVersion")
	logger.Info("Processing new AliceNet node version becoming canonical...")

	event, err := contracts.EthereumContracts().Dynamics().ParseNewCanonicalAliceNetNodeVersion(log)
	if err != nil {
		return err
	}

	logger = logger.WithFields(logrus.Fields{
		"ExecutionEpoch": event.Version.ExecutionEpoch,
		"Major":          event.Version.Major,
		"Minor":          event.Version.Minor,
		"Patch":          event.Version.Patch,
	})
	newMajorIsGreater, _, _, localVersion, err := utils.CompareCanonicalVersion(event.Version)
	if err != nil {
		return fmt.Errorf("processing new canonical version: %w", err)
	}

	if newMajorIsGreater {
		logger.Errorf(
			"CRITICAL: Exiting! Your Node Version %d.%d.%d is lower than the latest required version %d.%d.%d! Please update your node!",
			localVersion.Major,
			localVersion.Minor,
			localVersion.Patch,
			event.Version.Major,
			event.Version.Minor,
			event.Version.Patch,
		)
		exitFunc()
	}
	return nil
}
