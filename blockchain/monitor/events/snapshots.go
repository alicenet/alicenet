package events

import (
	"context"
	monInterfaces "github.com/MadBase/MadNet/blockchain/monitor/interfaces"

	ethereumInterfaces "github.com/MadBase/MadNet/blockchain/ethereum/interfaces"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// ProcessSnapshotTaken handles receiving snapshots
func ProcessSnapshotTaken(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, log types.Log,
	adminHandler monInterfaces.IAdminHandler) error {

	logger.Info("ProcessSnapshotTaken() ...")

	c := eth.Contracts()

	event, err := c.Snapshots().ParseSnapshotTaken(log)
	if err != nil {
		return err
	}

	epoch := event.Epoch
	safeToProceedConsensus := event.IsSafeToProceedConsensus

	logger.WithFields(logrus.Fields{
		"ChainID":                  event.ChainId,
		"Epoch":                    epoch,
		"Height":                   event.Height,
		"Validator":                event.Validator.Hex(),
		"IsSafeToProceedConsensus": event.IsSafeToProceedConsensus}).Infof("Snapshot taken")

	// Retrieve snapshot information from contract
	ctx, cancel := context.WithTimeout(context.Background(), eth.Timeout())
	defer cancel()

	callOpts, err := eth.GetCallOpts(ctx, eth.GetDefaultAccount())
	if err != nil {
		return err
	}

	rawBClaims, err := c.Snapshots().GetBlockClaimsFromSnapshot(callOpts, epoch)
	if err != nil {
		return err
	}

	// put it back together
	bclaims := &objs.BClaims{
		ChainID:    rawBClaims.ChainId,
		Height:     rawBClaims.Height,
		TxCount:    rawBClaims.TxCount,
		PrevBlock:  rawBClaims.PrevBlock[:],
		TxRoot:     rawBClaims.TxRoot[:],
		StateRoot:  rawBClaims.StateRoot[:],
		HeaderRoot: rawBClaims.HeaderRoot[:],
	}

	header := &objs.BlockHeader{}
	header.BClaims = bclaims
	header.SigGroup = event.SignatureRaw
	header.TxHshLst = [][]byte{}

	// send the reconstituted header to a handler
	logger.Debugf("invoking adminHandler.AddSnapshot")
	err = adminHandler.AddSnapshot(header, safeToProceedConsensus)
	if err != nil {
		return err
	}

	return nil
}
