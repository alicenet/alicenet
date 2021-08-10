package monevents

import (
	"context"
	"fmt"

	aobjs "github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func ProcessDepositReceived(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log,
	consensusDb *db.Database, depositHandler interfaces.DepositHandler) error {

	logger.Info("ProcessDepositReceived() ...")

	event, err := eth.Contracts().Deposit().ParseDepositReceived(log)
	if err != nil {
		return err
	}

	bigChainID := eth.ChainID()
	//TODO check to make sure chainID fits into a uint32
	chainID := uint32(bigChainID.Uint64())

	logger.WithFields(logrus.Fields{
		"DepositID": event.DepositID,
		"Depositor": event.Depositor,
		"Amount":    event.Amount,
	}).Info("Deposit received")

	return consensusDb.Update(func(txn *badger.Txn) error {
		depositNonce := event.DepositID.Bytes()
		account := event.Depositor.Bytes()
		owner := &aobjs.Owner{}
		err := owner.New(account, constants.CurveSecp256k1)
		if err != nil {
			logger.Debugf("Error in Services.ProcessDepositReceived at owner.New: %v", err)
			return err
		}
		return depositHandler.Add(txn, chainID, depositNonce, event.Amount, owner)
	})
}

// ProcessValueUpdated handles a dynamic value updating coming from our smart contract
func ProcessValueUpdated(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log,
	adminHandler interfaces.AdminHandler) error {

	logger.Info("ProcessValueUpdated() ...")

	event, err := eth.Contracts().Governor().ParseValueUpdated(log)
	if err != nil {
		return err
	}

	logger = logger.WithFields(logrus.Fields{
		"Epoch": event.Epoch.Uint64(),
		"Key":   event.Key.String(),
		"Value": fmt.Sprintf("0x%x", event.Value),
	})

	logger.Infof("Value updated")

	logger.Warnf("Dropping dynamic value on the floor")
	return nil
}

// ProcessSnapshotTaken handles receiving snapshots
func ProcessSnapshotTaken(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log,
	adminHandler interfaces.AdminHandler) error {

	logger.Info("ProcessSnapshotTaken() ...")

	c := eth.Contracts()

	event, err := c.Validators().ParseSnapshotTaken(log)
	if err != nil {
		return err
	}

	epoch := event.Epoch
	ethDkgStarted := event.StartingETHDKG

	logger.WithFields(logrus.Fields{
		"ChainID":        event.ChainId,
		"Epoch":          epoch,
		"Height":         event.Height,
		"Validator":      event.Validator.Hex(),
		"StartingEthDKG": event.StartingETHDKG}).Infof("Snapshot taken")

	// Retrieve snapshot information from contract
	ctx, cancel := context.WithTimeout(context.Background(), eth.Timeout())
	defer cancel()

	callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())

	rawBClaims, err := c.Validators().GetRawBlockClaimsSnapshot(callOpts, epoch)
	if err != nil {
		return err
	}

	rawSignature, err := c.Validators().GetRawSignatureSnapshot(callOpts, epoch)
	if err != nil {
		return err
	}

	// put it back together
	bclaims := &objs.BClaims{}
	err = bclaims.UnmarshalBinary(rawBClaims)
	if err != nil {
		return err
	}
	header := &objs.BlockHeader{}
	header.BClaims = bclaims
	header.SigGroup = rawSignature

	// send the reconstituted header to a handler
	err = adminHandler.AddSnapshot(header, ethDkgStarted)
	if err != nil {
		return err
	}

	return nil
}
