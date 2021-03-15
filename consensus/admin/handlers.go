package admin

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/MadBase/MadNet/consensus/appmock"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/interfaces"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
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
	appHandler  appmock.Application
	RequestLock chan struct{}
	ReceiveLock chan interfaces.Lockable
}

// Init creates all fields and binds external services
func (ah *Handlers) Init(chainID uint32, database *db.Database, secret []byte, appHandler appmock.Application, ethPubk []byte) error {
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
	ah.RequestLock = make(chan struct{})
	ah.ReceiveLock = make(chan interfaces.Lockable)
	return nil
}

// Close shuts down all workers
func (ah *Handlers) Close() {
	ah.closeOnce.Do(func() {
		ah.cancelFunc()
	})
}

func (ah *Handlers) getLock() (interfaces.Lockable, bool) {
	select {
	case ah.RequestLock <- struct{}{}:
		select {
		case lock := <-ah.ReceiveLock:
			return lock, true
		case <-ah.ctx.Done():
			return nil, false
		}
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
	// ah.logger.Error("!!! OPEN addval TXN")
	// defer func() { ah.logger.Error("!!! CLOSE addval TXN") }()
	return ah.database.Update(func(txn *badger.Txn) error {
		// build round states
		bh, err := ah.database.GetLastSnapshot(txn)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				utils.DebugTrace(ah.logger, err)
				return err
			}
		}
		if bh == nil {
			stateRoot, err := ah.appHandler.ApplyState(txn, ah.chainID, 1, nil)
			if err != nil {
				utils.DebugTrace(ah.logger, err)
				return err
			}
			txRoot, err := objs.MakeTxRoot([][]byte{})
			if err != nil {
				utils.DebugTrace(ah.logger, err)
				return err
			}
			vlst := [][]byte{}
			for i := 0; i < len(v.Validators); i++ {
				val := v.Validators[i]
				vlst = append(vlst, crypto.Hasher(val.VAddr))
			}
			prevBlock, err := objs.MakeTxRoot(vlst)
			if err != nil {
				utils.DebugTrace(ah.logger, err)
				return err
			}
			bh = &objs.BlockHeader{
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
				return err
			}
			if err := ah.database.SetCommittedBlockHeader(txn, bh); err != nil {
				utils.DebugTrace(ah.logger, err)
				return err
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
				return err
			}
			ownValidatingState := &objs.OwnValidatingState{
				VAddr:    ah.ethAcct,
				GroupKey: v.GroupKey,
			}
			ownValidatingState.SetRoundStarted()
			if err := ah.database.SetOwnValidatingState(txn, ownValidatingState); err != nil {
				utils.DebugTrace(ah.logger, err)
				return err
			}
		}
		rcert, err := bh.GetRCert()
		if err != nil {
			utils.DebugTrace(ah.logger, err)
			return err
		}
		if rcert.RClaims.Height == 1 {
			rcert.RClaims.Height = 2
		}
		for i := 0; i < len(v.Validators); i++ {
			val := v.Validators[i]
			_, err := ah.database.GetCurrentRoundState(txn, val.VAddr)
			if err != nil {
				if err == badger.ErrKeyNotFound {
					rs := &objs.RoundState{
						VAddr:      utils.CopySlice(val.VAddr),
						GroupKey:   utils.CopySlice(v.GroupKey),
						GroupShare: utils.CopySlice(val.GroupShare),
						GroupIdx:   uint8(i),
						RCert:      rcert,
					}
					err = ah.database.SetCurrentRoundState(txn, rs)
					if err != nil {
						utils.DebugTrace(ah.logger, err)
						return err
					}
				} else {
					utils.DebugTrace(ah.logger, err)
					return err
				}
			}
		}
		_, err = ah.database.GetCurrentRoundState(txn, ah.ethAcct)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				rs := &objs.RoundState{
					VAddr:      ah.ethAcct,
					GroupKey:   v.GroupKey,
					GroupShare: make([]byte, constants.CurveBN256EthPubkeyLen),
					GroupIdx:   0,
					RCert:      rcert,
				}
				err = ah.database.SetCurrentRoundState(txn, rs)
				if err != nil {
					utils.DebugTrace(ah.logger, err)
					return err
				}
			} else {
				utils.DebugTrace(ah.logger, err)
				return err
			}
		}
		if rcert.RClaims.Height <= 2 {
			v.NotBefore = 1
		} else {
			v.NotBefore = rcert.RClaims.Height
		}
		err = ah.database.SetValidatorSet(txn, v)
		if err != nil {
			utils.DebugTrace(ah.logger, err)
			return err
		}
		err = ah.database.SetSafeToProceed(txn, rcert.RClaims.Height, true)
		if err != nil {
			utils.DebugTrace(ah.logger, err)
			return err
		}
		return nil
	})
}

// AddSnapshot stores a snapshot to the database
func (ah *Handlers) AddSnapshot(bh *objs.BlockHeader, startingEthDKG bool) error {
	mutex, ok := ah.getLock()
	if !ok {
		return nil
	}
	mutex.Lock()
	defer mutex.Unlock()
	// ah.logger.Error("!!! OPEN AddSnapshot TXN")
	// defer func() { ah.logger.Error("!!! CLOSE AddSnapshot TXN") }()
	return ah.database.Update(func(txn *badger.Txn) error {
		safeToProceed, err := ah.database.GetSafeToProceed(txn, bh.BClaims.Height)
		if err != nil {
			utils.DebugTrace(ah.logger, err)
			return err
		}
		if !safeToProceed {
			if err := ah.database.SetSafeToProceed(txn, bh.BClaims.Height, !startingEthDKG); err != nil {
				utils.DebugTrace(ah.logger, err)
				return err
			}
		}
		err = ah.database.SetSnapshotBlockHeader(txn, bh)
		if err != nil {
			utils.DebugTrace(ah.logger, err)
			return err
		}
		return nil
	})
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
			signer.SetPrivk(privk)
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
		// TODO: Do we now want to delete ClearText? Is there any concern
		// 		 about this being leaked?
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
					utils.DebugTrace(ah.logger, err)
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
