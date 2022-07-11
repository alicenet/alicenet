package admin

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/dynamics"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

// Todo: Retry logic on snapshot submission; this will cause deadlock
//		 if snapshots are taken too close together

// Handlers Is a private bus for internal service use.
// At this time the only reason to use this bus is to
// enable blockchain events to be fed into the system
// and for accusations to be fed out of the system and
// into the Ethereum blockchain.
type Handlers struct {
	sync.RWMutex
	ctx         context.Context
	cancelFunc  func()
	closeOnce   sync.Once
	database    *db.Database
	isInit      bool
	isSync      bool
	secret      []byte
	chainID     uint32
	logger      *logrus.Logger
	ethAcct     []byte
	ethPubk     []byte
	appHandler  interfaces.Application
	storage     dynamics.StorageGetter
	ReceiveLock chan interfaces.Lockable
}

// Init creates all fields and binds external services
func (ah *Handlers) Init(chainID uint32, database *db.Database, secret []byte, appHandler interfaces.Application, ethPubk []byte, storage dynamics.StorageGetter) {
	ctx := context.Background()
	subCtx, cancelFunc := context.WithCancel(ctx)
	ah.ctx = subCtx
	ah.cancelFunc = cancelFunc
	ah.logger = logging.GetLogger(constants.LoggerConsensus)
	ah.database = database
	ah.chainID = chainID
	ah.appHandler = appHandler
	ah.ethPubk = ethPubk
	ah.secret = utils.CopySlice(secret)
	ah.ethAcct = crypto.GetAccount(ethPubk)
	ah.ReceiveLock = make(chan interfaces.Lockable)
	ah.storage = storage
}

// Close shuts down all workers
func (ah *Handlers) Close() {
	ah.closeOnce.Do(func() {
		ah.cancelFunc()
	})
}

func (ah *Handlers) getLock() (interfaces.Lockable, bool) {
	select {
	case lock := <-ah.ReceiveLock:
		return lock, true
	case <-ah.ctx.Done():
		return nil, false
	}
}

// AddValidatorSet adds a validator set to the db
// This function also creates the first block and initializes
// the genesis state when the first validator set is written
func (ah *Handlers) AddValidatorSet(v *objs.ValidatorSet) error {
	mutex, ok := ah.getLock()
	if !ok {
		return nil
	}
	mutex.Lock()
	defer mutex.Unlock()
	return ah.database.Update(func(txn *badger.Txn) error {
		// Checking if we can exit earlier (mainly when reconstructing the chain
		// from ethereum state)
		{
			height := uint32(1)
			if v.NotBefore >= 1 {
				height = v.NotBefore
			}

			vSet, err := ah.database.GetValidatorSet(txn, height)
			if err != nil {
				if err != badger.ErrKeyNotFound {
					utils.DebugTrace(ah.logger, err)
					return err
				}
				// do nothing
			}
			bhHeight := height - 1
			if v.NotBefore == 0 {
				bhHeight = 1
			}
			bh, err := ah.database.GetCommittedBlockHeader(txn, bhHeight)
			if err != nil {
				if err != badger.ErrKeyNotFound {
					utils.DebugTrace(ah.logger, err)
					return err
				}
			}
			// If we have a committed blocker header, and the current validator
			// set in memory is equal to the validator set that we are
			// receiving, we are good and we don't need to execute the steps
			// below
			if bh != nil && vSet != nil && bytes.Equal(v.GroupKey, vSet.GroupKey) {
				return nil
			}
		}
		// Adding new validators in case of epoch boundary
		if v.NotBefore%constants.EpochLength == 0 {
			return ah.epochBoundaryValidator(txn, v)
		}
		// reset case (we received from ethereum an event with group key fields
		// all zeros).
		if bytes.Equal(v.GroupKey, make([]byte, len(v.GroupKey))) {
			return ah.database.SetValidatorSet(txn, v)
		}
		// Setting a new set of validator outside the epoch boundaries and after
		// a the reset case above
		return ah.AddValidatorSetEdgecase(txn, v)
	})
}

// AddValidatorSetEdgecase adds a validator set to the db if we have the
// expected block at the height 'v.NotBefore-1' (e.g syncing from the ethereum
// data). Otherwise, it will mark the change to happen in the future once we
// have the required block
func (ah *Handlers) AddValidatorSetEdgecase(txn *badger.Txn, v *objs.ValidatorSet) error {
	bh, err := ah.database.GetCommittedBlockHeader(txn, v.NotBefore-1)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.DebugTrace(ah.logger, err)
			return err
		}
		return ah.database.SetValidatorSetPostApplication(txn, v, v.NotBefore)
	}
	rcert, err := bh.GetRCert()
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return err
	}
	isValidator, err := ah.initValidatorsRoundState(txn, v, rcert)
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return err
	}
	// If we are not a validator we need to start our Round State
	if !isValidator {
		err = ah.initOwnRoundState(txn, v, rcert)
		if err != nil {
			utils.DebugTrace(ah.logger, err)
			return err
		}
	}
	err = ah.database.SetValidatorSet(txn, v)
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return err
	}
	return nil
}

// AddSnapshot stores a snapshot to the database
func (ah *Handlers) AddSnapshot(bh *objs.BlockHeader, safeToProceedConsensus bool) error {
	ah.logger.Debugf("inside adminHandler.AddSnapshot")
	mutex, ok := ah.getLock()
	if !ok {
		return errors.New("could not get adminHandler lock")
	}
	mutex.Lock()
	defer mutex.Unlock()
	err := ah.database.Update(func(txn *badger.Txn) error {
		safeToProceed, err := ah.database.GetSafeToProceed(txn, bh.BClaims.Height+1)
		if err != nil {
			utils.DebugTrace(ah.logger, err)
			return err
		}
		if !safeToProceed {
			ah.logger.Debugf("Did validators change in the previous epoch:%v Setting is safe to proceed for height %d to: %v", !safeToProceedConsensus, bh.BClaims.Height+1, safeToProceedConsensus)
			// set that it's safe to proceed to the next block
			if err := ah.database.SetSafeToProceed(txn, bh.BClaims.Height+1, safeToProceedConsensus); err != nil {
				utils.DebugTrace(ah.logger, err)
				return err
			}
		}
		if bh.BClaims.Height > 1 {
			err = ah.database.SetSnapshotBlockHeader(txn, bh)
			if err != nil {
				utils.DebugTrace(ah.logger, err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return err
	}
	ah.logger.Debugf("successfully saved state on adminHandler.AddSnapshot")
	return nil
}

// UpdateDynamicStorage updates dynamic storage values.
func (ah *Handlers) UpdateDynamicStorage(txn *badger.Txn, key, value string, epoch uint32) error {
	mutex, ok := ah.getLock()
	if !ok {
		return nil
	}
	mutex.Lock()
	defer mutex.Unlock()

	update, err := dynamics.NewUpdate(key, value, epoch)
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return err
	}
	err = ah.storage.UpdateStorage(txn, update)
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return err
	}
	return nil
}

// IsInitialized returns if the database has been initialized yet
func (ah *Handlers) IsInitialized() bool {
	ah.RLock()
	defer ah.RUnlock()
	return ah.isInit
}

// IsSynchronized returns if the ethereum BC has been/is synchronized
func (ah *Handlers) IsSynchronized() bool {
	ah.RLock()
	defer ah.RUnlock()
	return ah.isSync
}

// SetSynchronized allows the BC monitor to set the sync state for ethereum
func (ah *Handlers) SetSynchronized(v bool) {
	ah.Lock()
	defer ah.Unlock()
	ah.isSync = v
}

// RegisterSnapshotCallback allows a callback to be registered that will be called on snapshot blocks being
// added to the local db.
func (ah *Handlers) RegisterSnapshotCallback(fn func(bh *objs.BlockHeader) error) {
	wrapper := func(v []byte) error {
		bh := &objs.BlockHeader{}
		err := bh.UnmarshalBinary(v)
		if err != nil {
			return err
		}
		isValidator := false
		var syncToBH, maxBHSeen *objs.BlockHeader
		err = ah.database.View(func(txn *badger.Txn) error {
			vs, err := ah.database.GetValidatorSet(txn, bh.BClaims.Height)
			if err != nil {
				return err
			}
			os, err := ah.database.GetOwnState(txn)
			if err != nil {
				return err
			}
			for i := 0; i < len(vs.Validators); i++ {
				val := vs.Validators[i]
				if bytes.Equal(val.VAddr, os.VAddr) {
					isValidator = true
					break
				}
			}
			syncToBH = os.SyncToBH
			maxBHSeen = os.MaxBHSeen
			return nil
		})
		if err != nil {
			return err
		}
		if !isValidator {
			return nil
		}
		if maxBHSeen.BClaims.Height-syncToBH.BClaims.Height >= constants.EpochLength {
			return nil
		}
		if bh.BClaims.Height%constants.EpochLength == 0 {
			return fn(bh)
		}
		return nil
	}
	ah.database.SubscribeBroadcastBlockHeader(ah.ctx, wrapper)
}

// AddPrivateKey stores a private key from an EthDKG run into an encrypted
// keystore in the DB
func (ah *Handlers) AddPrivateKey(pk []byte, curveSpec constants.CurveSpec) error {
	mutex, ok := ah.getLock()
	if !ok {
		return nil
	}
	mutex.Lock()
	defer mutex.Unlock()
	// ah.logger.Error("!!! OPEN AddPrivateKey TXN")
	// defer func() { ah.logger.Error("!!! CLOSE AddPrivateKey TXN") }()
	err := ah.database.Update(func(txn *badger.Txn) error {
		switch curveSpec {
		case constants.CurveSecp256k1:
			privk := utils.CopySlice(pk)
			// secp key
			signer := crypto.Secp256k1Signer{}
			err := signer.SetPrivk(privk)
			if err != nil {
				return err
			}
			pubkey, err := signer.Pubkey()
			if err != nil {
				return err
			}
			name := crypto.GetAccount(pubkey)
			ec := &objs.EncryptedStore{
				Name:      name,
				ClearText: privk,
				Kid:       constants.AdminHandlerKid(),
			}
			err = ec.Encrypt(ah)
			if err != nil {
				return err
			}
			return ah.database.SetEncryptedStore(txn, ec)
		case constants.CurveBN256Eth:
			privk := utils.CopySlice(pk)
			// bn key
			signer := crypto.BNGroupSigner{}
			err := signer.SetPrivk(privk)
			if err != nil {
				return err
			}
			pubkey, err := signer.PubkeyShare()
			if err != nil {
				return err
			}
			ec := &objs.EncryptedStore{
				Name:      pubkey,
				ClearText: privk,
				Kid:       constants.AdminHandlerKid(),
			}
			err = ec.Encrypt(ah)
			if err != nil {
				return err
			}
			return ah.database.SetEncryptedStore(txn, ec)
		default:
			panic("not an allowed curve type")
		}
	})
	if err != nil {
		panic(err)
	}
	return nil
}

// GetPrivK returns an decrypted private key from an EthDKG run to the caller
func (ah *Handlers) GetPrivK(name []byte) ([]byte, error) {
	var privk []byte
	err := ah.database.View(func(txn *badger.Txn) error {
		ec, err := ah.database.GetEncryptedStore(txn, name)
		if err != nil {
			return err
		}
		err = ec.Decrypt(ah)
		if err != nil {
			return err
		}
		privk = utils.CopySlice(ec.ClearText)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return privk, nil
}

// GetKey allows the admin handler to act as a key resolver for decrypting
// stored private keys
func (ah *Handlers) GetKey(kid []byte) ([]byte, error) {
	out := make([]byte, len(ah.secret))
	copy(out[:], ah.secret)
	return out, nil
}

// InitializationMonitor polls the database for the existence of a snapshot
// It sets IsInitialized when one is found and returns
func (ah *Handlers) InitializationMonitor(closeChan <-chan struct{}) {
	ah.logger.Debug("InitializationMonitor loop starting")
	fn := func() {
		ah.logger.Debug("InitializationMonitor loop stopping")
	}
	defer fn()
	for {
		ok, err := func() (bool, error) {
			select {
			case <-closeChan:
				return false, errorz.ErrClosing
			case <-ah.ctx.Done():
				return false, nil
			case <-time.After(2 * time.Second):
				err := ah.database.View(func(txn *badger.Txn) error {
					_, err := ah.database.GetLastSnapshot(txn)
					if err != nil {
						return err
					}
					return nil
				})
				if err != nil {
					return false, nil
				}
				ah.Lock()
				ah.isInit = true
				ah.Unlock()
				return true, nil
			}
		}()
		if err != nil {
			return
		}
		if ok {
			return
		}
	}
}

func (ah *Handlers) epochBoundaryValidator(txn *badger.Txn, v *objs.ValidatorSet) error {
	bh, err := ah.database.GetSnapshotByHeight(txn, v.NotBefore)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.DebugTrace(ah.logger, err)
			return err
		}
	}
	if bh == nil || v.NotBefore == 0 {
		bh, err = ah.initDB(txn, v)
		if err != nil {
			utils.DebugTrace(ah.logger, err)
			return err
		}
	}
	rcert, err := bh.GetRCert()
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return err
	}
	isValidator, err := ah.initValidatorsRoundState(txn, v, rcert)
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return err
	}
	// If we are not a validator we need to start our Round State
	if !isValidator {
		err = ah.initOwnRoundState(txn, v, rcert)
		if err != nil {
			utils.DebugTrace(ah.logger, err)
			return err
		}
	}

	// fix zero epoch event in chain
	switch v.NotBefore {
	case 0:
		v.NotBefore = 1
	default:
		v.NotBefore = rcert.RClaims.Height
	}

	err = ah.database.SetValidatorSet(txn, v)
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return err
	}
	err = ah.database.SetSafeToProceed(txn, v.NotBefore, true)
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return err
	}
	return nil
}

// Re-Initializes our own Round State object
func (ah *Handlers) initOwnRoundState(txn *badger.Txn, v *objs.ValidatorSet, rcert *objs.RCert) error {
	rs, err := ah.database.GetCurrentRoundState(txn, ah.ethAcct)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return err
		}
	}
	if (rs == nil) || (!bytes.Equal(rs.GroupKey, v.GroupKey) && v.NotBefore >= rcert.RClaims.Height) {
		rs = &objs.RoundState{
			VAddr:      ah.ethAcct,
			GroupKey:   v.GroupKey,
			GroupShare: make([]byte, constants.CurveBN256EthPubkeyLen),
			GroupIdx:   0,
			RCert:      rcert,
		}
	}
	err = ah.database.SetCurrentRoundState(txn, rs)
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return err
	}
	return nil
}

// Re-Initializes all the validators Round State objects
func (ah *Handlers) initValidatorsRoundState(txn *badger.Txn, v *objs.ValidatorSet, rcert *objs.RCert) (bool, error) {
	isValidator := false
	for i := 0; i < len(v.Validators); i++ {
		val := v.Validators[i]
		rs, err := ah.database.GetCurrentRoundState(txn, val.VAddr)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(ah.logger, err)
				return false, err
			}
		}
		rcertTemp := rcert
		if rs != nil && rs.RCert.RClaims.Height > rcert.RClaims.Height {
			rcertTemp = rs.RCert
		}
		rs = &objs.RoundState{
			VAddr:      utils.CopySlice(val.VAddr),
			GroupKey:   utils.CopySlice(v.GroupKey),
			GroupShare: utils.CopySlice(val.GroupShare),
			GroupIdx:   uint8(i),
			RCert:      rcertTemp,
		}
		err = ah.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			utils.DebugTrace(ah.logger, err)
			return false, err
		}
		if bytes.Equal(rs.VAddr, ah.ethAcct) {
			isValidator = true
		}
	}
	return isValidator, nil
}

// Init the validators DB and objects
func (ah *Handlers) initDB(txn *badger.Txn, v *objs.ValidatorSet) (*objs.BlockHeader, error) {
	stateRoot, err := ah.appHandler.ApplyState(txn, ah.chainID, 1, nil)
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return nil, err
	}
	txRoot, err := objs.MakeTxRoot([][]byte{})
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return nil, err
	}
	vlst := [][]byte{}
	for i := 0; i < len(v.Validators); i++ {
		val := v.Validators[i]
		vlst = append(vlst, crypto.Hasher(val.VAddr))
	}
	prevBlock, err := objs.MakeTxRoot(vlst)
	if err != nil {
		utils.DebugTrace(ah.logger, err)
		return nil, err
	}
	bh := &objs.BlockHeader{
		BClaims: &objs.BClaims{
			ChainID:    ah.chainID,
			Height:     1,
			PrevBlock:  prevBlock,
			StateRoot:  stateRoot,
			HeaderRoot: make([]byte, constants.HashLen),
			TxRoot:     txRoot,
		},
		SigGroup: make([]byte, constants.CurveBN256EthSigLen),
		TxHshLst: [][]byte{},
	}
	if err := ah.database.SetSnapshotBlockHeader(txn, bh); err != nil {
		utils.DebugTrace(ah.logger, err)
		return nil, err
	}
	if err := ah.database.SetCommittedBlockHeader(txn, bh); err != nil {
		utils.DebugTrace(ah.logger, err)
		return nil, err
	}
	ownState := &objs.OwnState{
		VAddr:             ah.ethAcct,
		SyncToBH:          bh,
		MaxBHSeen:         bh,
		CanonicalSnapShot: bh,
		PendingSnapShot:   bh,
	}
	if err := ah.database.SetOwnState(txn, ownState); err != nil {
		utils.DebugTrace(ah.logger, err)
		return nil, err
	}
	ownValidatingState := new(objs.OwnValidatingState)
	ownValidatingState.SetRoundStarted()
	if err := ah.database.SetOwnValidatingState(txn, ownValidatingState); err != nil {
		utils.DebugTrace(ah.logger, err)
		return nil, err
	}
	return bh, nil
}
