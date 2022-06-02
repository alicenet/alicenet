package events

import (
	"errors"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/objects"

	aobjs "github.com/MadBase/MadNet/application/objs"
	ethereumInterfaces "github.com/MadBase/MadNet/blockchain/ethereum/interfaces"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func ProcessDepositReceived(eth ethereumInterfaces.IEthereum, logger *logrus.Entry, log types.Log,
	cdb *db.Database, depositHandler interfaces.DepositHandler) error {

	logger.Info("ProcessDepositReceived() ...")

	dkgState := &dkgObjects.DkgState{}
	var err error
	err = cdb.View(func(txn *badger.Txn) error {
		dkgState, err = objects.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	if !dkgState.IsValidator {
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
