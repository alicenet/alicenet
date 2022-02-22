package monevents

import (
	"context"
	"errors"

	aobjs "github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	bobjs "github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/errorz"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func ProcessDepositReceived(eth interfaces.Ethereum, logger *logrus.Entry, state *bobjs.MonitorState, log types.Log,
	cdb *db.Database, depositHandler interfaces.DepositHandler) error {

	logger.Info("ProcessDepositReceived() ...")

	event, err := eth.Contracts().MadByte().ParseDepositReceived(log)
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

	err = cdb.Update(func(txn *badger.Txn) error {
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

	if err != nil {
		e := errorz.ErrInvalid{}.New("")
		if !errors.As(err, &e) {
			return err
		}
	}
	return nil
}

/*
// ProcessValueUpdated handles a dynamic value updating coming from our smart contract
func ProcessValueUpdated(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log,
	adminHandler interfaces.AdminHandler) error {

	logger.Info("ProcessValueUpdated() ...")
	panic("unimplemented function!")

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
} */

// ProcessSnapshotTaken handles receiving snapshots
func ProcessSnapshotTaken(eth interfaces.Ethereum, logger *logrus.Entry, state *bobjs.MonitorState, log types.Log,
	adminHandler interfaces.AdminHandler) error {

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

	callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())

	rawBClaims, err := c.Snapshots().GetBlockClaimsFromLatestSnapshot(callOpts)
	if err != nil {
		return err
	}

	signature, err := c.Snapshots().GetSignatureFromLatestSnapshot(callOpts)
	if err != nil {
		return err
	}

	sig, err := bn256.MarshalBigIntSlice(signature[:])
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
	header.SigGroup = sig
	header.TxHshLst = [][]byte{}

	// send the reconstituted header to a handler
	err = adminHandler.AddSnapshot(header, safeToProceedConsensus)
	if err != nil {
		return err
	}

	return nil
}
