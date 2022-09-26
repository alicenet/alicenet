package events

import (
	"errors"

	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"

	aobjs "github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/monitor/interfaces"
)

func ProcessDepositReceived(eth layer1.Client, contracts layer1.AllSmartContracts, logger *logrus.Entry, log types.Log, cdb, monDB *db.Database, depositHandler interfaces.DepositHandler, chainID uint32) error {
	event, err := contracts.EthereumContracts().BToken().ParseDepositReceived(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"DepositID": event.DepositID,
		"Depositor": event.Depositor,
		"Amount":    event.Amount,
	}).Debug("Deposit received")

	err = cdb.Update(func(txn *badger.Txn) error {
		depositNonce := event.DepositID.Bytes()
		account := event.Depositor.Bytes()
		owner := &aobjs.Owner{}
		// todo: evaluate sec concern of non-validated CurveSpec if any
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
