package events

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"

	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/crypto/bn256"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/snapshots"
	monInterfaces "github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/alicenet/alicenet/utils"
)

// ProcessSnapshotTaken handles receiving snapshots.
func ProcessSnapshotTaken(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, adminHandler monInterfaces.AdminHandler, taskRequestChan chan<- tasks.TaskRequest, exitFunc func()) error {
	logger = logger.WithField("method", "ProcessSnapshotTaken")
	logger.Info("Processing snapshots...")

	c := contracts.EthereumContracts()

	event, err := c.Snapshots().ParseSnapshotTaken(log)
	if err != nil {
		return err
	}

	epoch := event.Epoch
	safeToProceedConsensus := event.IsSafeToProceedConsensus

	logger = logger.WithFields(logrus.Fields{
		"ChainID":                  event.ChainId,
		"Epoch":                    epoch,
		"Height":                   event.Height,
		"Validator":                event.Validator.Hex(),
		"IsSafeToProceedConsensus": event.IsSafeToProceedConsensus})
	logger.Info("Snapshot taken")

	// put it back together
	bclaims := &objs.BClaims{
		ChainID:    event.BClaims.ChainId,
		Height:     event.BClaims.Height,
		TxCount:    event.BClaims.TxCount,
		PrevBlock:  event.BClaims.PrevBlock[:],
		TxRoot:     event.BClaims.TxRoot[:],
		StateRoot:  event.BClaims.StateRoot[:],
		HeaderRoot: event.BClaims.HeaderRoot[:],
	}

	header := &objs.BlockHeader{}
	header.BClaims = bclaims
	sigGroupBytes := []*big.Int{
		event.MasterPublicKey[0],
		event.MasterPublicKey[1],
		event.MasterPublicKey[2],
		event.MasterPublicKey[3],
		event.Signature[0],
		event.Signature[1],
	}
	header.SigGroup, err = bn256.MarshalBigIntSlice(sigGroupBytes)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal signature from snapshots")
		return err
	}
	header.TxHshLst = [][]byte{}

	// send the reconstituted header to a handler
	logger.Debug("invoking adminHandler.AddSnapshot")
	err = adminHandler.AddSnapshot(header, safeToProceedConsensus)
	if err != nil {
		return err
	}

	// kill any task that might still be trying to do this snapshot
	taskRequestChan <- tasks.NewKillTaskRequest(&snapshots.SnapshotTask{})

	//todo: this most likely wont be needed after adding the Canonical Version to the Snapshot event ----
	callOpts, err := eth.GetCallOpts(context.Background(), eth.GetDefaultAccount())
	if err != nil {
		logger.Errorf("Received and error during GetCallOpts: %v", err)
	}
	latestVersion, err := contracts.EthereumContracts().Dynamics().GetLatestAliceNetVersion(callOpts)
	if err != nil {
		logger.Errorf("Received and error during GetLatestAliceNetVersion: %v", err)
	}
	//todo: ---------------------------------------------------------------------------------------------

	if newMajorIsGreater, _, _, localVersion := utils.CompareCanonicalVersion(latestVersion); newMajorIsGreater {
		if uint32(epoch.Uint64()) > latestVersion.ExecutionEpoch {
			logger.Errorf("CRITICAL: your Major Canonical Node Version %d.%d.%d is lower than the latest %d.%d.%d and you exeeded the execution epoch %d. Please update your node before restart. Exiting!",
				localVersion.Major, localVersion.Minor, localVersion.Patch, latestVersion.Major, latestVersion.Minor, latestVersion.Patch, latestVersion.ExecutionEpoch)
			exitFunc()
		}
	}

	return nil
}
