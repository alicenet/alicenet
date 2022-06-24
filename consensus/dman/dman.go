package dman

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/errorz"

	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

type DMan struct {
	downloadActor *RootActor
	database      databaseView
	appHandler    interfaces.Application
	bnVal         *crypto.BNGroupValidator
	logger        *logrus.Logger
}

func (dm *DMan) Init(database databaseView, app interfaces.Application, reqBus reqBusView) {
	dm.logger = logging.GetLogger(constants.LoggerDMan)
	dm.database = database
	dm.appHandler = app
	dm.bnVal = &crypto.BNGroupValidator{}
	proxy := &typeProxy{
		app,
		reqBus,
		database,
	}
	dm.downloadActor = &RootActor{}
	dm.downloadActor.Init(dm.logger, proxy)
}

func (dm *DMan) Start() {
	dm.downloadActor.Start()
}

func (dm *DMan) Close() {
	dm.downloadActor.wg.Wait()
}

func (dm *DMan) FlushCacheToDisk(txn *badger.Txn, height uint32) error {
	return dm.downloadActor.FlushCacheToDisk(txn, height)
}

func (dm *DMan) CleanCache(txn *badger.Txn, height uint32) error {
	return dm.downloadActor.CleanCache(txn, height)
}

func (dm *DMan) AddTxs(txn *badger.Txn, height uint32, txs []interfaces.Transaction) error {
	for i := 0; i < len(txs); i++ {
		tx := txs[i]
		txHash, err := tx.TxHash()
		if err != nil {
			utils.DebugTrace(dm.logger, err)
			return err
		}
		txb, err := tx.MarshalBinary()
		if err != nil {
			utils.DebugTrace(dm.logger, err)
			return err
		}
		if err := dm.database.SetTxCacheItem(txn, height, utils.CopySlice(txHash), utils.CopySlice(txb)); err != nil {
			utils.DebugTrace(dm.logger, err)
			return err
		}
	}
	return nil
}

func (dm *DMan) GetTxs(txn *badger.Txn, height, round uint32, txLst [][]byte) ([]interfaces.Transaction, [][]byte, error) {
	result := &txResult{appHandler: dm.appHandler, logger: dm.logger}
	result.init(txLst)
	missing := result.missing()
	// get from the database
	for i := 0; i < len(missing); i++ {
		txHash := utils.CopySlice(missing[i])
		txb, err := dm.database.GetTxCacheItem(txn, height, txHash)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(dm.logger, err)
				return nil, nil, err
			}
			continue
		}
		if err := result.addRaw(txb); err != nil {
			utils.DebugTrace(dm.logger, err)
			return nil, nil, err
		}
	}
	missing = result.missing()
	// get from the pending store
	found, _, err := dm.appHandler.PendingTxGet(txn, height, missing)
	if err != nil {
		var e *errorz.ErrInvalid
		if err != errorz.ErrMissingTransactions && !errors.As(err, &e) && err != badger.ErrKeyNotFound {
			utils.DebugTrace(dm.logger, err)
			return nil, nil, err
		}
	}
	for i := 0; i < len(found); i++ {
		txh, err := found[i].TxHash()
		if err != nil {
			utils.DebugTrace(dm.logger, err)
			return nil, nil, err
		}
		txb, err := found[i].MarshalBinary()
		if err != nil {
			utils.DebugTrace(dm.logger, err)
			return nil, nil, err
		}
		err = dm.database.SetTxCacheItem(txn, height, txh, txb)
		if err != nil {
			utils.DebugTrace(dm.logger, err)
			return nil, nil, err
		}
	}
	if err := result.addMany(found); err != nil {
		utils.DebugTrace(dm.logger, err)
		return nil, nil, err
	}
	missing = result.missing()
	found = []interfaces.Transaction{}
	for i := 0; i < len(missing); i++ {
		txi, ok := dm.downloadActor.txc.Get(missing[i])
		if ok {
			found = append(found, txi)
		}
	}
	if err := result.addMany(found); err != nil {
		utils.DebugTrace(dm.logger, err)
		return nil, nil, err
	}
	missing = result.missing()
	if len(missing) > 0 {
		dm.DownloadTxs(height, round, missing)
	}
	missing = result.missing()
	return result.txs, missing, nil
}

// SyncOneBH syncs one blockheader and its transactions
// the initialization of prevBH from SyncToBH implies SyncToBH must be updated to
// the canonical bh before we begin unless we are syncing from a height gt the
// canonical bh
func (dm *DMan) SyncOneBH(txn *badger.Txn, syncToBH *objs.BlockHeader, maxBHSeen *objs.BlockHeader, validatorSet *objs.ValidatorSet) ([]interfaces.Transaction, *objs.BlockHeader, bool, error) {
	targetHeight := syncToBH.BClaims.Height + 1

	currentHeight := dm.downloadActor.ba.getHeight()

	bhCache, inCache := dm.downloadActor.bhc.Get(targetHeight)
	if !inCache && currentHeight != targetHeight {
		for i := currentHeight + 1; i < currentHeight+constants.EpochLength; i++ {
			height := i
			if height > maxBHSeen.BClaims.Height {
				break
			}
			_, inCache := dm.downloadActor.bhc.Get(height)
			if !inCache {
				dm.downloadActor.DownloadBlockHeader(height, 1)
			} else {
				break
			}
		}
		dm.downloadActor.ba.updateHeight(syncToBH.BClaims.Height)
		return nil, nil, false, nil
	}

	dm.downloadActor.ba.updateHeight(syncToBH.BClaims.Height)

	// check the chainID of bh
	if bhCache.BClaims.ChainID != syncToBH.BClaims.ChainID {
		dm.downloadActor.bhc.Del(targetHeight)
		return nil, nil, false, errorz.ErrInvalid{}.New("Wrong chainID")
	}
	// check the height of the bh
	if bhCache.BClaims.Height != targetHeight {
		dm.downloadActor.bhc.Del(targetHeight)
		return nil, nil, false, errorz.ErrInvalid{}.New("Wrong block height")
	}
	// get prev bh
	prevBHsh, err := syncToBH.BlockHash()
	if err != nil {
		dm.downloadActor.bhc.Del(targetHeight)
		utils.DebugTrace(dm.logger, err)
		return nil, nil, false, err
	}
	// compare to prevBlock from bh
	if !bytes.Equal(bhCache.BClaims.PrevBlock, prevBHsh) {
		dm.downloadActor.bhc.Del(targetHeight)
		return nil, nil, false, errorz.ErrInvalid{}.New("BlockHash does not match previous!")
	}
	txs, missing, err := dm.GetTxs(txn, targetHeight, 1, bhCache.TxHshLst)
	if err != nil {
		utils.DebugTrace(dm.logger, err)
		return nil, nil, false, err
	}
	if len(missing) > 0 {
		return nil, nil, false, nil
	}
	// verify the signature and group key
	if err := bhCache.ValidateSignatures(dm.bnVal); err != nil {
		utils.DebugTrace(dm.logger, err)
		return nil, nil, false, errorz.ErrInvalid{}.New(err.Error())
	}
	if !bytes.Equal(bhCache.GroupKey, validatorSet.GroupKey) {
		return nil, nil, false, errorz.ErrInvalid{}.New(fmt.Sprintf("group key does not match expected: Height: %d bhCache.GroupKey: %x\nvalidatorSet.GroupKey:%x", bhCache.BClaims.Height, bhCache.GroupKey, validatorSet.GroupKey))
	}
	if err := dm.database.SetCommittedBlockHeader(txn, bhCache); err != nil {
		return nil, nil, false, err
	}
	if syncToBH.BClaims.Height > 10 {
		if err := dm.database.TxCacheDropBefore(txn, syncToBH.BClaims.Height-5, 1000); err != nil {
			utils.DebugTrace(dm.logger, err)
			return nil, nil, false, err
		}
	}

	return txs, bhCache, true, nil
}

func (dm *DMan) DownloadTxs(height, round uint32, txHshLst [][]byte) {
	for i := 0; i < len(txHshLst); i++ {
		txHsh := txHshLst[i]
		dm.downloadActor.DownloadTx(height, round, txHsh)
	}
}
