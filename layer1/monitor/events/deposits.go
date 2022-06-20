package events

import (
	"errors"

	aobjs "github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/layer1"
	"github.com/MadBase/MadNet/layer1/ethereum"
	"github.com/MadBase/MadNet/layer1/executor/tasks/dkg/state"
	"github.com/MadBase/MadNet/layer1/monitor/interfaces"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func ProcessDepositReceived(eth layer1.Client, logger *logrus.Entry, log types.Log,
	cdb *db.Database, monDB *db.Database, depositHandler interfaces.DepositHandler) error {

	logger.Info("ProcessDepositReceived() ...")

	dkgState := &state.DkgState{}
	var err error
	err = monDB.View(func(txn *badger.Txn) error {
		err = dkgState.LoadState(txn)
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

	event, err := ethereum.GetContracts().BToken().ParseDepositReceived(log)
	if err != nil {
		return err
	}

	bigChainID := eth.GetChainID()
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
