package lstate

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"

	"github.com/MadBase/MadNet/consensus/appmock"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/consensus/request"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

var errExpiredCtx = errors.New("ctx canceled")

type txResult struct {
	logger     *logrus.Logger
	appHandler appmock.Application
	txs        []interfaces.Transaction
	txHashes   map[string]bool
}

func (t *txResult) init(txHashes [][]byte) {
	t.txs = []interfaces.Transaction{}
	t.txHashes = make(map[string]bool)
	for _, txHash := range txHashes {
		t.txHashes[string(txHash)] = false
	}
}

func (t *txResult) missing() [][]byte {
	var missing [][]byte
	for txHash, haveIt := range t.txHashes {
		if !haveIt {
			missing = append(missing, utils.CopySlice([]byte(txHash)))
		}
	}
	return missing
}

func (t *txResult) add(tx interfaces.Transaction) error {
	txHash, err := tx.TxHash()
	if err != nil {
		return err
	}
	haveIt, shouldHaveIt := t.txHashes[string(txHash)]
	if !haveIt && shouldHaveIt {
		t.txs = append(t.txs, tx)
		t.txHashes[string(txHash)] = true
	}
	return nil
}

func (t *txResult) addMany(txs []interfaces.Transaction) error {
	var err error
	for i := 0; i < len(txs); i++ {
		e := t.add(txs[i])
		if e != nil {
			err = e
			utils.DebugTrace(t.logger, err)
		}
	}
	return err
}

func (t *txResult) addManyRaw(txs [][]byte) error {
	var err error
	for i := 0; i < len(txs); i++ {
		e := t.addRaw(txs[i])
		if e != nil {
			err = e
			utils.DebugTrace(t.logger, err)
		}
	}
	return err
}

func (t *txResult) addRaw(txb []byte) error {
	tx, err := t.appHandler.UnmarshalTx(utils.CopySlice(txb))
	if err != nil {
		utils.DebugTrace(t.logger, err)
		return err
	}
	return t.add(tx)
}

type DMan struct {
	sync.RWMutex
	database       db.DatabaseIface
	ctx            context.Context
	cf             func()
	appHandler     appmock.Application
	requestBus     *request.Client
	dlc            *dLCache
	txc            *txCache
	bhc            *bHCache
	height         uint32
	logger         *logrus.Logger
	targetBlockHDR *objs.BlockHeader
}

func (dm *DMan) Init(database db.DatabaseIface, app appmock.Application, reqBus *request.Client) error {
	ctx := context.Background()
	subCtx, cf := context.WithCancel(ctx)
	dm.ctx = subCtx
	dm.cf = cf
	dm.logger = logging.GetLogger(constants.LoggerDMan)
	dm.database = database
	dm.appHandler = app
	dm.requestBus = reqBus
	dm.dlc = &dLCache{}
	err := dm.dlc.init(constants.DownloadTO)
	if err != nil {
		utils.DebugTrace(dm.logger, err)
		return err
	}

	dm.txc = &txCache{
		app: dm.appHandler,
	}
	err = dm.txc.init()
	if err != nil {
		utils.DebugTrace(dm.logger, err)
		return err
	}

	dm.bhc = &bHCache{}
	err = dm.bhc.init()
	if err != nil {
		utils.DebugTrace(dm.logger, err)
		return err
	}
	return nil
}

func (dm *DMan) reset(txn *badger.Txn, height uint32) error {
	dm.cf()
	ctx := context.Background()
	subCtx, cf := context.WithCancel(ctx)
	dm.ctx = subCtx
	dm.cf = cf
	if height > 10 {
		if err := dm.database.TxCacheDropBefore(txn, height-5, 1000); err != nil {
			utils.DebugTrace(dm.logger, err)
			return err
		}
	}
	return nil
}

func (dm *DMan) AddTxs(txn *badger.Txn, height uint32, txs []interfaces.Transaction, allowReset bool) error {
	dm.Lock()
	defer dm.Unlock()
	if height > dm.height && allowReset {
		if err := dm.reset(txn, height); err != nil {
			dm.logger.Debugf("Error in DMan.AddTxs at dm.reset: %v", err)
			return err
		}
		dm.height = height
	}
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
		err = dm.txc.add(tx)
		if err != nil {
			utils.DebugTrace(dm.logger, err)
			return err
		}
	}
	return nil
}

func (dm *DMan) GetTxs(txn *badger.Txn, height uint32, txLst [][]byte) ([]interfaces.Transaction, [][]byte, error) {
	dm.RLock()
	defer dm.RUnlock()
	return dm.getTxsInternal(txn, height, txLst)
}

func (dm *DMan) getTxsInternal(txn *badger.Txn, height uint32, txLst [][]byte) ([]interfaces.Transaction, [][]byte, error) {
	result := &txResult{appHandler: dm.appHandler, logger: dm.logger}
	result.init(txLst)

	// get from cache
	found, _ := dm.txc.getMany(txLst)
	if err := result.addMany(found); err != nil {
		utils.DebugTrace(dm.logger, err)
		return nil, nil, err
	}

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

	if err := result.addMany(found); err != nil {
		utils.DebugTrace(dm.logger, err)
		return nil, nil, err
	}

	missing = result.missing()
	if len(missing) > 0 {
		dm.downloadTxsInternal(dm.ctx, dm.dlc, dm.txc, missing)
	}
	missing = result.missing()
	return result.txs, missing, nil
}

// SyncOneBH syncs one blockheader and its transactions
// the initialization of prevBH from SyncToBH implies SyncToBH must be updated to
// the canonical bh before we begin unless we are syncing from a height gt the
// canonical bh
func (dm *DMan) SyncOneBH(txn *badger.Txn, rs *RoundStates) ([]interfaces.Transaction, *objs.BlockHeader, error) {
	// get the lock and setup the release
	dm.Lock()
	defer dm.Unlock()

	// create the signature validator
	bnVal := &crypto.BNGroupValidator{}

	// assign the target height
	targetHeight := rs.OwnState.SyncToBH.BClaims.Height + 1
	if targetHeight > dm.height {
		dm.height = targetHeight
	}

	if dm.targetBlockHDR != nil {
		if dm.targetBlockHDR.BClaims.Height < targetHeight {
			dm.targetBlockHDR = nil
		}
	}

	if dm.targetBlockHDR == nil {

		// create a nested context with timeout for request
		ctx, cancelFunc := context.WithTimeout(context.Background(), constants.MsgTimeout)
		defer cancelFunc()

		// do the request
		bhLst, err := dm.requestBus.RequestP2PGetBlockHeaders(ctx, []uint32{targetHeight})
		if err != nil {
			utils.DebugTrace(dm.logger, err)
			return nil, nil, errorz.ErrInvalid{}.New("get BlockHeaders failed")
		}

		// if we got back too many block headers then return
		if len(bhLst) != 1 {
			return nil, nil, errorz.ErrInvalid{}.New("len(bhLst) != 1")
		}

		// bh is element zero of bhLst
		bh := bhLst[0]

		// check the chainID of bh
		if bh.BClaims.ChainID != rs.OwnState.SyncToBH.BClaims.ChainID {
			return nil, nil, errorz.ErrInvalid{}.New("Wrong chainID")
		}

		// check the height of the bh
		if bh.BClaims.Height != targetHeight {
			return nil, nil, errorz.ErrInvalid{}.New("Wrong block height")
		}
		prevBHsh, err := rs.OwnState.SyncToBH.BlockHash() // get block hash
		if err != nil {
			utils.DebugTrace(dm.logger, err)
			return nil, nil, err
		}
		// compare to prevBlock from bh
		if !bytes.Equal(bh.BClaims.PrevBlock, prevBHsh) {
			return nil, nil, errorz.ErrInvalid{}.New("BlockHash does not match previous!")
		}

		// verify the signature and group key
		GroupKey := bh.GroupKey
		if err := bh.ValidateSignatures(bnVal); err != nil {
			utils.DebugTrace(dm.logger, err)
			return nil, nil, errorz.ErrInvalid{}.New(err.Error())
		}
		if !bytes.Equal(GroupKey, rs.ValidatorSet.GroupKey) {
			return nil, nil, errorz.ErrInvalid{}.New("group key does not match expected")
		}
		dm.targetBlockHDR = bh
	}

	if dm.targetBlockHDR != nil {
		txs, missing, err := dm.getTxsInternal(txn, dm.targetBlockHDR.BClaims.Height, dm.targetBlockHDR.TxHshLst)
		if err != nil {
			utils.DebugTrace(dm.logger, err)
			return nil, nil, err
		}
		if len(missing) > 0 {
			utils.DebugTrace(dm.logger, err)
			return nil, nil, errorz.ErrMissingTransactions
		}
		return txs, dm.targetBlockHDR, nil
	}
	return nil, nil, errorz.ErrMissingTransactions
}

func (dm *DMan) DownloadTxs(txHshLst [][]byte) bool {
	dm.Lock()
	defer dm.Unlock()
	return dm.downloadTxsInternal(dm.ctx, dm.dlc, dm.txc, txHshLst)
}

func (dm *DMan) downloadTxsInternal(ctx context.Context, dlc *dLCache, txc *txCache, txHshLst [][]byte) bool {
	missingCount := 0
	for i := 0; i < len(txHshLst); i++ {
		txHsh := txHshLst[i]
		if !txc.containsTxHsh(utils.CopySlice(txHsh)) {
			missingCount++
			if !dlc.containsTxHsh(utils.CopySlice(txHsh)) {
				err := dlc.add(utils.CopySlice(txHsh))
				if err != nil {
					utils.DebugTrace(dm.logger, err)
				}
				go dm.downloadWithRetry(ctx, dlc, txc, utils.CopySlice(txHsh))
			}
		}
	}
	return missingCount <= 0
}

// should be owned by dl cache
func (dm *DMan) downloadWithRetry(ctx context.Context, dlc *dLCache, txc *txCache, txHsh []byte) {
	defer dlc.cancelOne(utils.CopySlice(txHsh))
	subCtx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// check if we have the tx
			if txc.containsTxHsh(utils.CopySlice(txHsh)) {
				return
			}
			// check the cache to see if time is left
			if dlc.expired(utils.CopySlice(txHsh)) {
				return
			}
			tx, err := dm.downloadOne(subCtx, dlc, txc, utils.CopySlice(txHsh))
			if err != nil {
				utils.DebugTrace(dm.logger, err)
				// if an error was returned, wait a small timeout and continue
				time.Sleep(1 * time.Second)
				continue
			}
			txhshReturned, err := tx.TxHash()
			if err != nil {
				utils.DebugTrace(dm.logger, err)
				time.Sleep(1 * time.Second)
				continue
			}
			if !bytes.Equal(txhshReturned, txHsh) {
				time.Sleep(1 * time.Second)
				continue
			}
			err = func() error {
				dm.Lock()
				defer dm.Unlock()
				select {
				case <-ctx.Done():
					return nil
				default:
					if err := txc.add(tx); err != nil {
						utils.DebugTrace(dm.logger, err)
						return err
					}
				}
				return nil
			}()
			if err == nil {
				return
			}
		}
	}
}

// downloadOne will download one transaction
func (dm *DMan) downloadOne(ctx context.Context, dlc *dLCache, txc *txCache, txHsh []byte) (interfaces.Transaction, error) {
	select {
	case <-ctx.Done():
		return nil, errExpiredCtx
	default:
		subCtx, cancelFunc := context.WithTimeout(ctx, constants.MsgTimeout)
		defer cancelFunc()
		txLst, err := dm.requestBus.RequestP2PGetPendingTx(subCtx, [][]byte{utils.CopySlice(txHsh)})
		if err == nil && len(txLst) == 1 {
			tx, err := dm.appHandler.UnmarshalTx(txLst[0])
			if err == nil {
				return tx, nil
			}
		}
		// check if we have the tx
		if txc.containsTxHsh(utils.CopySlice(txHsh)) {
			return nil, errors.New("complete")
		}
		// check the cache to see if time is left
		if dlc.expired(utils.CopySlice(txHsh)) {
			return nil, errors.New("complete")
		}
		subCtx2, cancelFunc2 := context.WithTimeout(ctx, constants.MsgTimeout)
		defer cancelFunc2()
		txLst, err = dm.requestBus.RequestP2PGetMinedTxs(subCtx2, [][]byte{txHsh})
		if err != nil {
			utils.DebugTrace(dm.logger, err)
			return nil, errorz.ErrInvalid{}.New(err.Error())
		}
		if len(txLst) != 1 {
			return nil, errorz.ErrInvalid{}.New("Downloaded more than 1 txn when only should have 1")
		}
		tx, err := dm.appHandler.UnmarshalTx(utils.CopySlice(txLst[0]))
		if err != nil {
			utils.DebugTrace(dm.logger, err)
			return nil, err
		}
		return tx, nil
	}
}
