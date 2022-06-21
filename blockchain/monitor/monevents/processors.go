package monevents

import (
	"context"
	"errors"
	"fmt"

	aobjs "github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/blockchain/objects"
	bobjs "github.com/alicenet/alicenet/blockchain/objects"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/errorz"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func ProcessDepositReceived(eth interfaces.Ethereum, logger *logrus.Entry, state *bobjs.MonitorState, log types.Log,
	cdb *db.Database, depositHandler interfaces.DepositHandler) error {

	logger.Info("ProcessDepositReceived() ...")

	if !state.EthDKG.IsValidator {
		return nil
	}

	event, err := eth.Contracts().BToken().ParseDepositReceived(log)
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
		// todo: evvaluate sec concern of non-validated CurveSpec if any
		if err := owner.New(account, constants.CurveSpec(event.AccountType)); err != nil {
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

// ProcessValueUpdated handles a dynamic value updating coming from our smart contract
func ProcessValueUpdated(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log,
	adminHandler interfaces.AdminHandler) error {

	logger.Info("ProcessValueUpdated() ...")

	if !state.EthDKG.IsValidator {
		return nil
	}

	event, err := eth.Contracts().Governance().ParseValueUpdated(log)
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

// ProcessValidatorMinorSlashed handles the Minor Slash event
func ProcessValidatorMinorSlashed(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessValidatorMinorSlashed() ...")

	event, err := eth.Contracts().ValidatorPool().ParseValidatorMinorSlashed(log)
	if err != nil {
		return err
	}

	logger = logger.WithFields(logrus.Fields{
		"Account":               event.Account.String(),
		"PublicStaking.TokenID": event.PublicStakingTokenID.Uint64(),
	})

	logger.Infof("ValidatorMinorSlashed")

	return nil
}

// ProcessValidatorMajorSlashed handles the Major Slash event
func ProcessValidatorMajorSlashed(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {

	logger.Info("ProcessValidatorMajorSlashed() ...")

	event, err := eth.Contracts().ValidatorPool().ParseValidatorMajorSlashed(log)
	if err != nil {
		return err
	}

	logger = logger.WithFields(logrus.Fields{
		"Account": event.Account.String(),
	})

	logger.Infof("ValidatorMajorSlashed")

	return nil
}
