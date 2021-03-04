package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/MadBase/MadNet/constants/dbprefix"

	trie "github.com/MadBase/MadNet/badgerTrie"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

// Database is an abstraction of the header trie and the object storage
type Database struct {
	sync.Mutex
	rawDB  *rawDataBase
	trie   *headerTrie
	logger *logrus.Logger
}

// Init will initialize the database
func (db *Database) Init(DB *badger.DB) error {
	logger := logging.GetLogger(constants.LoggerDB)
	db.logger = logger
	db.rawDB = &rawDataBase{db: DB, logger: logger}
	hdrTrie := &headerTrie{}
	hdrTrie.init()
	db.trie = hdrTrie
	return nil
}

func (db *Database) DB() *badger.DB {
	return db.rawDB.db
}

func (db *Database) View(fn TxnFunc) error {
	return db.rawDB.View(fn)
}

func (db *Database) Update(fn TxnFunc) error {
	db.Lock()
	defer db.Unlock()
	return db.rawDB.Update(fn)
}

func (db *Database) Sync() error {
	return db.rawDB.Sync()
}

func (db *Database) GarbageCollect() error {
	return db.rawDB.GarbageCollect()
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeEncryptedStoreKey(name []byte) []byte {
	Name := utils.CopySlice(name)
	prefix := dbprefix.PrefixEncryptedStore()
	key := []byte{}
	key = append(key, prefix...)
	key = append(key, Name...)
	return key
}

func (db *Database) SetEncryptedStore(txn *badger.Txn, v *objs.EncryptedStore) error {
	key := db.makeEncryptedStoreKey(v.Name)
	err := db.rawDB.SetEncryptedStore(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetEncryptedStore(txn *badger.Txn, name []byte) (*objs.EncryptedStore, error) {
	key := db.makeEncryptedStoreKey(name)
	result, err := db.rawDB.GetEncryptedStore(txn, key)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return nil, err
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// index current by VAddr
func (db *Database) makeCurrentRoundStateKey(vAddr []byte) ([]byte, error) {
	key := &objs.RoundStateCurrentKey{
		Prefix: dbprefix.PrefixCurrentRoundState(),
		VAddr:  utils.CopySlice(vAddr),
	}
	return key.MarshalBinary()
}

func (db *Database) SetCurrentRoundState(txn *badger.Txn, v *objs.RoundState) error {
	key, err := db.makeCurrentRoundStateKey(v.VAddr)
	if err != nil {
		return err
	}
	if err := db.rawDB.SetRoundState(txn, key, v); err != nil {
		return err
	}
	return db.SetHistoricRoundState(txn, v)
}

func (db *Database) GetCurrentRoundState(txn *badger.Txn, vaddr []byte) (*objs.RoundState, error) {
	key, err := db.makeCurrentRoundStateKey(vaddr)
	if err != nil {
		return nil, err
	}
	result, err := db.rawDB.GetRoundState(txn, key)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return nil, err
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// index historic by by height|round|vkey
func (db *Database) makeHistoricRoundStateKey(vaddr []byte, height uint32, round uint32) ([]byte, error) {
	key := &objs.RoundStateHistoricKey{
		Prefix: dbprefix.PrefixHistoricRoundState(),
		Height: height,
		Round:  round,
		VAddr:  utils.CopySlice(vaddr),
	}
	return key.MarshalBinary()
}

func (db *Database) makeHistoricRoundStateIterKey(height uint32) ([]byte, error) {
	key := &objs.RoundStateHistoricKey{
		Prefix: dbprefix.PrefixHistoricRoundState(),
		Height: height,
	}
	return key.MakeIterKey()
}

func (db *Database) SetHistoricRoundState(txn *badger.Txn, v *objs.RoundState) error {
	key, err := db.makeHistoricRoundStateKey(v.VAddr, v.RCert.RClaims.Height, v.RCert.RClaims.Round)
	if err != nil {
		return err
	}
	err = db.rawDB.SetRoundState(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetHistoricRoundState(txn *badger.Txn, vaddr []byte, height uint32, round uint32) (*objs.RoundState, error) {
	key, err := db.makeHistoricRoundStateKey(vaddr, height, round)
	if err != nil {
		return nil, err
	}
	result, err := db.rawDB.GetRoundState(txn, key)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return nil, err
	}
	return result, nil
}

func (db *Database) DeleteBeforeHistoricRoundState(txn *badger.Txn, height uint32, maxnum int) error {
	prefix, err := db.makeHistoricRoundStateIterKey(height)
	if err != nil {
		return err
	}
	opts := badger.DefaultIteratorOptions
	it := txn.NewIterator(opts)
	defer it.Close()
	keys := [][]byte{}
	count := 0
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		k := item.KeyCopy(nil)
		keys = append(keys, k)
		count++
		if count >= maxnum {
			break
		}
	}
	for i := 0; i < len(keys); i++ {
		k := keys[i]
		err := utils.DeleteValue(txn, utils.CopySlice(k))
		if err != nil {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeValidatorSetKey(notBefore uint32) ([]byte, error) {
	key := &objs.ValidatorSetKey{
		Prefix:    dbprefix.PrefixValidatorSet(),
		NotBefore: notBefore,
	}
	return key.MarshalBinary()
}

func (db *Database) makeValidatorSetIterKey() []byte {
	return dbprefix.PrefixValidatorSet()
}

func (db *Database) SetValidatorSet(txn *badger.Txn, v *objs.ValidatorSet) error {
	key, err := db.makeValidatorSetKey(v.NotBefore)
	if err != nil {
		return err
	}
	err = db.rawDB.SetValidatorSet(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetValidatorSet(txn *badger.Txn, height uint32) (*objs.ValidatorSet, error) {
	prefix := db.makeValidatorSetIterKey()
	seek := []byte{}
	seek = append(seek, prefix...)
	seek = append(seek, []byte{255, 255, 255, 255, 255}...)
	opts := badger.DefaultIteratorOptions
	opts.Reverse = true
	opts.Prefix = prefix
	opts.PrefetchValues = false
	var lastkey []byte
	func() {
		it := txn.NewIterator(opts)
		defer it.Close()
		it.Seek(seek)
		if it.Valid() {
			item := it.Item()
			k := item.KeyCopy(nil)
			lastkey = k
		}
	}()
	if lastkey == nil {
		return nil, badger.ErrKeyNotFound
	}
	result, err := db.rawDB.GetValidatorSet(txn, lastkey)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return nil, err
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) MakeHeaderTrieKeyFromHeight(height uint32) []byte {
	return makeTrieKeyFromHeight(height)
}

func (db *Database) makeCurrentHeaderRootKey() []byte {
	return dbprefix.PrefixBlockHeaderTrieRootCurrent()
}

func (db *Database) makeCommittedBlockHeaderKey(height uint32) ([]byte, error) {
	prefix := dbprefix.PrefixCommittedBlockHeader()
	key := &objs.BlockHeaderHeightKey{
		Prefix: prefix,
		Height: height,
	}
	return key.MarshalBinary()
}

func (db *Database) makeCommittedBlockHeaderHashIndexKey(blockHash []byte) ([]byte, error) {
	prefix := dbprefix.PrefixCommittedBlockHeaderHashIndex()
	key := &objs.BlockHeaderHashIndexKey{
		Prefix:    prefix,
		BlockHash: blockHash,
	}
	return key.MarshalBinary()
}

func (db *Database) makeHistoricHeaderRootKey(height uint32) ([]byte, error) {
	prefix := dbprefix.PrefixBlockHeaderTrieRootHistoric()
	key := &objs.BlockHeaderHeightKey{
		Prefix: prefix,
		Height: height,
	}
	return key.MarshalBinary()
}

func (db *Database) GetHeaderTrieRoot(txn *badger.Txn, height uint32) ([]byte, error) {
	key, err := db.makeHistoricHeaderRootKey(height)
	if err != nil {
		return nil, err
	}
	rt, err := db.rawDB.getValue(txn, key)
	if err != nil {
		utils.DebugTrace(db.logger, err, fmt.Sprintf("key that is not being found is %v", key))
		return nil, err
	}
	return rt, nil
}

func (db *Database) SetHeaderTrieRoot(txn *badger.Txn, height uint32, root []byte) error {
	key, err := db.makeHistoricHeaderRootKey(height)
	if err != nil {
		return err
	}
	err = db.rawDB.SetValue(txn, key, root)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) UpdateHeaderTrieRootFastSync(txn *badger.Txn, v *objs.BlockHeader) error {
	if err := db.finalizeSnapShotHdrRoot(txn, v.BClaims.HeaderRoot, v.BClaims.Height-1); err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	newHdrRoot, err := db.trie.ApplyState(txn, v, 0)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	if err := db.finalizeSnapShotHdrRoot(txn, newHdrRoot, v.BClaims.Height); err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) finalizeSnapShotHdrRoot(txn *badger.Txn, root []byte, height uint32) error {
	headerRootKey := db.makeCurrentHeaderRootKey()
	err := db.rawDB.SetValue(txn, headerRootKey, root)
	if err != nil {
		return err
	}
	historicTrieKey, err := db.makeHistoricHeaderRootKey(height)
	if err != nil {
		return err
	}
	err = db.rawDB.SetValue(txn, historicTrieKey, root)
	if err != nil {
		return err
	}
	return db.trie.FinalizeSnapShotHdrRoot(txn, root, height)
}

func (db *Database) SetCommittedBlockHeader(txn *badger.Txn, v *objs.BlockHeader) error {
	headerRoot, err := db.trie.ApplyState(txn, v, 0)
	if err != nil {
		return err
	}
	headerRootKey := db.makeCurrentHeaderRootKey()
	err = db.rawDB.SetValue(txn, headerRootKey, headerRoot)
	if err != nil {
		return err
	}
	historicTrieKey, err := db.makeHistoricHeaderRootKey(v.BClaims.Height)
	if err != nil {
		return err
	}
	err = db.rawDB.SetValue(txn, historicTrieKey, headerRoot)
	if err != nil {
		return err
	}
	return db.setCommittedBlockHeaderInternal(txn, v)
}

func (db *Database) SetCommittedBlockHeaderFastSync(txn *badger.Txn, v *objs.BlockHeader) error {
	return db.setCommittedBlockHeaderInternal(txn, v)
}

func (db *Database) setCommittedBlockHeaderInternal(txn *badger.Txn, v *objs.BlockHeader) error {
	key, err := db.makeCommittedBlockHeaderKey(v.BClaims.Height)
	if err != nil {
		return err
	}
	bHash, err := v.BlockHash()
	if err != nil {
		return err
	}
	indKey, err := db.makeCommittedBlockHeaderHashIndexKey(bHash)
	if err != nil {
		return err
	}
	err = db.rawDB.SetValue(txn, indKey, key)
	if err != nil {
		return err
	}
	err = db.rawDB.SetBlockHeader(txn, key, v)
	if err != nil {
		return err
	}
	if v.BClaims.Height == 1 || v.BClaims.Height%constants.EpochLength == 0 {
		if err := db.SetSnapshotBlockHeader(txn, v); err != nil {
			return err
		}
	}
	db.logger.Tracef(`
    BlockHeader{
      BClaims{
        ChainID:    %v
        Height:     %v
        TxCount:    %v
        TxRoot:     %x
        StateRoot:  %x
        HeaderRoot: %x
        PrevBlock:  %x
      }
      Sig:          %x ... %x
      TxList:       %x
    }`, v.BClaims.ChainID, v.BClaims.Height, v.BClaims.TxCount, v.BClaims.TxRoot, v.BClaims.StateRoot, v.BClaims.HeaderRoot, v.BClaims.PrevBlock, v.SigGroup[0:16], v.SigGroup[len(v.SigGroup)-11:len(v.SigGroup)-1], v.TxHshLst)
	return nil
}

func (db *Database) GetHeaderRootForProposal(txn *badger.Txn) ([]byte, error) {
	key := db.makeCurrentHeaderRootKey()
	result, err := db.rawDB.getValue(txn, key)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return nil, err
	}
	return result, nil
}

func (db *Database) DeleteCommittedBlockHeader(txn *badger.Txn, height uint32) error {
	headerRoot, err := db.trie.ApplyState(txn, nil, height)
	if err != nil {
		return err
	}
	key, err := db.makeCommittedBlockHeaderKey(height)
	if err != nil {
		return err
	}
	v, err := db.rawDB.GetBlockHeader(txn, key)
	if err != nil {
		return err
	}
	bHash, err := v.BlockHash()
	if err != nil {
		return err
	}
	indKey, err := db.makeCommittedBlockHeaderHashIndexKey(bHash)
	if err != nil {
		return err
	}
	err = utils.DeleteValue(txn, indKey)
	if err != nil {
		return err
	}
	err = utils.DeleteValue(txn, key)
	if err != nil {
		return err
	}
	headerRootKey := db.makeCurrentHeaderRootKey()
	err = db.rawDB.SetValue(txn, headerRootKey, headerRoot)
	if err != nil {
		return err
	}
	historicTrieKey, err := db.makeHistoricHeaderRootKey(v.BClaims.Height)
	if err != nil {
		return err
	}
	err = db.rawDB.SetValue(txn, historicTrieKey, headerRoot)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) ValidateCommittedBlockHeaderWithProof(txn *badger.Txn, root []byte, blockHeader *objs.BlockHeader, proof []byte) (bool, error) {
	rootHash, err := db.GetHeaderRootForProposal(txn)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return false, err
		}
		rootHash = nil
	}
	return db.trie.VerifyProof(txn, rootHash, root, proof, blockHeader)
}

func (db *Database) GetCommittedBlockHeaderWithProof(txn *badger.Txn, root []byte, blockHeight uint32) (*objs.BlockHeader, []byte, error) {
	hdr, err := db.GetCommittedBlockHeader(txn, blockHeight)
	if err != nil {
		return nil, nil, err
	}
	rootHash, err := db.GetHeaderRootForProposal(txn)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return nil, nil, err
		}
		rootHash = nil
	}
	_, proof, err := db.trie.GetProof(txn, rootHash, root, blockHeight)
	if err != nil {
		return nil, nil, err
	}
	return hdr, proof, nil
}

func (db *Database) GetCommittedBlockHeader(txn *badger.Txn, height uint32) (*objs.BlockHeader, error) {
	key, err := db.makeCommittedBlockHeaderKey(height)
	if err != nil {
		return nil, err
	}
	result, err := db.rawDB.GetBlockHeader(txn, key)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return nil, err
	}
	return result, nil
}

func (db *Database) GetCommittedBlockHeaderByHash(txn *badger.Txn, hash []byte) (*objs.BlockHeader, error) {
	indKey, err := db.makeCommittedBlockHeaderHashIndexKey(hash)
	if err != nil {
		return nil, err
	}
	key, err := db.rawDB.getValue(txn, indKey)
	if err != nil {
		return nil, err
	}
	result, err := db.rawDB.GetBlockHeader(txn, key)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return nil, err
	}
	return result, nil
}

func (db *Database) GetMostRecentCommittedBlockHeaderFastSync(txn *badger.Txn) (*objs.BlockHeader, error) {
	prefix := dbprefix.PrefixCommittedBlockHeader()
	seek := []byte{}
	seek = append(seek, prefix...)
	seek = append(seek, []byte{255, 255, 255, 255, 255}...)
	opts := badger.DefaultIteratorOptions
	opts.Reverse = true
	opts.Prefix = prefix
	opts.PrefetchValues = false
	var lastkey []byte
	func() {
		it := txn.NewIterator(opts)
		defer it.Close()
		it.Seek(seek)
		if it.Valid() {
			item := it.Item()
			k := item.KeyCopy(nil)
			lastkey = k
		}
	}()
	if lastkey == nil {
		return nil, badger.ErrKeyNotFound
	}
	result, err := db.rawDB.GetBlockHeader(txn, lastkey)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return nil, err
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeOwnStateKey() []byte {
	return dbprefix.PrefixOwnState()
}

func (db *Database) SetOwnState(txn *badger.Txn, v *objs.OwnState) error {
	key := db.makeOwnStateKey()
	err := db.rawDB.SetOwnState(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetOwnState(txn *badger.Txn) (*objs.OwnState, error) {
	key := db.makeOwnStateKey()
	result, err := db.rawDB.GetOwnState(txn, key)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return nil, err
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeOwnValidatingStateKey() []byte {
	return dbprefix.PrefixOwnValidatingState()
}

func (db *Database) SetOwnValidatingState(txn *badger.Txn, v *objs.OwnValidatingState) error {
	key := db.makeOwnValidatingStateKey()
	err := db.rawDB.SetOwnValidatingState(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetOwnValidatingState(txn *badger.Txn) (*objs.OwnValidatingState, error) {
	key := db.makeOwnValidatingStateKey()
	result, err := db.rawDB.GetOwnValidatingState(txn, key)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return nil, err
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeBroadcastBlockHeaderKey() []byte {
	return dbprefix.PrefixBroadcastBlockHeader()
}

func (db *Database) SetBroadcastBlockHeader(txn *badger.Txn, v *objs.BlockHeader) error {
	if v.BClaims.Height == 1 {
		return nil
	}
	key := db.makeBroadcastBlockHeaderKey()
	err := db.rawDB.SetBlockHeader(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetBroadcastBlockHeader(txn *badger.Txn) (*objs.BlockHeader, error) {
	key := db.makeBroadcastBlockHeaderKey()
	result, err := db.rawDB.GetBlockHeader(txn, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *Database) SubscribeBroadcastBlockHeader(ctx context.Context, cb func([]byte) error) {
	db.rawDB.subscribeToPrefix(ctx, dbprefix.PrefixBroadcastBlockHeader(), cb)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeBroadcastRCertKey() []byte {
	return dbprefix.PrefixBroadcastRCert()
}

func (db *Database) SetBroadcastRCert(txn *badger.Txn, v *objs.RCert) error {
	key := db.makeBroadcastRCertKey()
	err := db.rawDB.SetRCert(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetBroadcastRCert(txn *badger.Txn) (*objs.RCert, error) {
	key := db.makeBroadcastRCertKey()
	result, err := db.rawDB.GetRCert(txn, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *Database) SubscribeBroadcastRCert(ctx context.Context, cb func([]byte) error) {
	db.rawDB.subscribeToPrefix(ctx, dbprefix.PrefixBroadcastRCert(), cb)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeBroadcastTransactionKey() []byte {
	return dbprefix.PrefixBroadcastTransaction()
}

func (db *Database) SetBroadcastTransaction(txn *badger.Txn, v []byte) error {
	key := db.makeBroadcastTransactionKey()
	return utils.SetValue(txn, key, v)
}

func (db *Database) SubscribeBroadcastTransaction(ctx context.Context, cb func([]byte) error) {
	db.rawDB.subscribeToPrefix(ctx, dbprefix.PrefixBroadcastTransaction(), cb)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeBroadcastProposalKey() []byte {
	return dbprefix.PrefixBroadcastProposal()
}

func (db *Database) SetBroadcastProposal(txn *badger.Txn, v *objs.Proposal) error {
	key := db.makeBroadcastProposalKey()
	err := db.rawDB.SetProposal(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetBroadcastProposal(txn *badger.Txn) (*objs.Proposal, error) {
	key := db.makeBroadcastProposalKey()
	result, err := db.rawDB.GetProposal(txn, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *Database) SubscribeBroadcastProposal(ctx context.Context, cb func([]byte) error) {
	db.rawDB.subscribeToPrefix(ctx, dbprefix.PrefixBroadcastProposal(), cb)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeBroadcastPreVoteKey() []byte {
	return dbprefix.PrefixBroadcastPreVote()
}

func (db *Database) SetBroadcastPreVote(txn *badger.Txn, v *objs.PreVote) error {
	key := db.makeBroadcastPreVoteKey()
	err := db.rawDB.SetPreVote(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetBroadcastPreVote(txn *badger.Txn) (*objs.PreVote, error) {
	key := db.makeBroadcastPreVoteKey()
	result, err := db.rawDB.GetPreVote(txn, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *Database) SubscribeBroadcastPreVote(ctx context.Context, cb func([]byte) error) {
	db.rawDB.subscribeToPrefix(ctx, dbprefix.PrefixBroadcastPreVote(), cb)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeBroadcastPreVoteNilKey() []byte {
	return dbprefix.PrefixBroadcastPreVoteNil()
}

func (db *Database) SetBroadcastPreVoteNil(txn *badger.Txn, v *objs.PreVoteNil) error {
	key := db.makeBroadcastPreVoteNilKey()
	err := db.rawDB.SetPreVoteNil(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetBroadcastPreVoteNil(txn *badger.Txn) (*objs.PreVoteNil, error) {
	key := db.makeBroadcastPreVoteNilKey()
	result, err := db.rawDB.GetPreVoteNil(txn, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *Database) SubscribeBroadcastPreVoteNil(ctx context.Context, cb func([]byte) error) {
	db.rawDB.subscribeToPrefix(ctx, dbprefix.PrefixBroadcastPreVoteNil(), cb)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeBroadcastPreCommitKey() []byte {
	return dbprefix.PrefixBroadcastPreCommit()
}

func (db *Database) SetBroadcastPreCommit(txn *badger.Txn, v *objs.PreCommit) error {
	key := db.makeBroadcastPreCommitKey()
	err := db.rawDB.SetPreCommit(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetBroadcastPreCommit(txn *badger.Txn) (*objs.PreCommit, error) {
	key := db.makeBroadcastPreCommitKey()
	result, err := db.rawDB.GetPreCommit(txn, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *Database) SubscribeBroadcastPreCommit(ctx context.Context, cb func([]byte) error) {
	db.rawDB.subscribeToPrefix(ctx, dbprefix.PrefixBroadcastPreCommit(), cb)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeBroadcastPreCommitNilKey() []byte {
	return dbprefix.PrefixBroadcastPreCommitNil()
}

func (db *Database) SetBroadcastPreCommitNil(txn *badger.Txn, v *objs.PreCommitNil) error {
	key := db.makeBroadcastPreCommitNilKey()
	err := db.rawDB.SetPreCommitNil(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetBroadcastPreCommitNil(txn *badger.Txn) (*objs.PreCommitNil, error) {
	key := db.makeBroadcastPreCommitNilKey()
	result, err := db.rawDB.GetPreCommitNil(txn, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *Database) SubscribeBroadcastPreCommitNil(ctx context.Context, cb func([]byte) error) {
	db.rawDB.subscribeToPrefix(ctx, dbprefix.PrefixBroadcastPreCommitNil(), cb)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeBroadcastNextHeightKey() []byte {
	return dbprefix.PrefixBroadcastNextHeight()
}

func (db *Database) SetBroadcastNextHeight(txn *badger.Txn, v *objs.NextHeight) error {
	key := db.makeBroadcastNextHeightKey()
	err := db.rawDB.SetNextHeight(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetBroadcastNextHeight(txn *badger.Txn) (*objs.NextHeight, error) {
	key := db.makeBroadcastNextHeightKey()
	result, err := db.rawDB.GetNextHeight(txn, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *Database) SubscribeBroadcastNextHeight(ctx context.Context, cb func([]byte) error) {
	db.rawDB.subscribeToPrefix(ctx, dbprefix.PrefixBroadcastNextHeight(), cb)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeBroadcastNextRoundKey() []byte {
	return dbprefix.PrefixBroadcastNextRound()
}

func (db *Database) SetBroadcastNextRound(txn *badger.Txn, v *objs.NextRound) error {
	key := db.makeBroadcastNextRoundKey()
	err := db.rawDB.SetNextRound(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetBroadcastNextRound(txn *badger.Txn) (*objs.NextRound, error) {
	key := db.makeBroadcastNextRoundKey()
	result, err := db.rawDB.GetNextRound(txn, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *Database) SubscribeBroadcastNextRound(ctx context.Context, cb func([]byte) error) {
	db.rawDB.subscribeToPrefix(ctx, dbprefix.PrefixBroadcastNextRound(), cb)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeSnapshotBlockHeaderKey(height uint32) ([]byte, error) {
	prefix := dbprefix.PrefixSnapshotBlockHeader()
	key := &objs.BlockHeaderHeightKey{
		Prefix: prefix,
		Height: height,
	}
	return key.MarshalBinary()
}

func (db *Database) makeSnapshotBlockHeaderIterKey() []byte {
	prefix := dbprefix.PrefixSnapshotBlockHeader()
	return prefix
}

func (db *Database) SetSnapshotBlockHeader(txn *badger.Txn, v *objs.BlockHeader) error {
	key, err := db.makeSnapshotBlockHeaderKey(v.BClaims.Height)
	if err != nil {
		return err
	}
	err = db.rawDB.SetBlockHeader(txn, key, v)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return err
	}
	return nil
}

func (db *Database) GetSnapshotBlockHeader(txn *badger.Txn, height uint32) (*objs.BlockHeader, error) {
	key, err := db.makeSnapshotBlockHeaderKey(height)
	if err != nil {
		return nil, err
	}
	result, err := db.rawDB.GetBlockHeaderUnsafe(txn, key)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return nil, err
	}
	return result, nil
}

func (db *Database) GetLastSnapshot(txn *badger.Txn) (*objs.BlockHeader, error) {
	prefix := db.makeSnapshotBlockHeaderIterKey()
	seek := []byte{}
	seek = append(seek, prefix...)
	seek = append(seek, []byte{255, 255, 255, 255, 255}...)
	opts := badger.DefaultIteratorOptions
	opts.Reverse = true
	opts.Prefix = prefix
	opts.PrefetchValues = false
	var lastkey []byte
	func() {
		it := txn.NewIterator(opts)
		defer it.Close()
		it.Seek(seek)
		if it.Valid() {
			item := it.Item()
			k := item.KeyCopy(nil)
			lastkey = k
		}
	}()
	if lastkey == nil {
		return nil, badger.ErrKeyNotFound
	}
	result, err := db.rawDB.GetBlockHeaderUnsafe(txn, lastkey)
	if err != nil {
		utils.DebugTrace(db.logger, err)
		return nil, err
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeTxCacheKey(txHash []byte, height uint32) ([]byte, error) {
	prefix := dbprefix.PrefixTxCache()
	key := &objs.TxCacheKey{
		Prefix: prefix,
		Height: height,
		TxHash: utils.CopySlice(txHash),
	}
	return key.MarshalBinary()
}

func (db *Database) makeTxCacheIterKey(height uint32) ([]byte, error) {
	prefix := dbprefix.PrefixTxCache()
	key := &objs.TxCacheKey{
		Prefix: prefix,
		Height: height,
	}
	return key.MakeIterKey()
}

func (db *Database) SetTxCacheItem(txn *badger.Txn, height uint32, txHash []byte, tx []byte) error {
	key, err := db.makeTxCacheKey(txHash, height)
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, tx)
}

func (db *Database) GetTxCacheItem(txn *badger.Txn, height uint32, txHash []byte) ([]byte, error) {
	key, err := db.makeTxCacheKey(txHash, height)
	if err != nil {
		return nil, err
	}
	return db.rawDB.getValue(txn, key)
}

func (db *Database) TxCacheDropBefore(txn *badger.Txn, beforeHeight uint32, maxKeys int) error {
	keys := [][]byte{}
	prefix, err := db.makeTxCacheIterKey(beforeHeight)
	if err != nil {
		return err
	}
	opts := badger.DefaultIteratorOptions
	it := txn.NewIterator(opts)
	defer it.Close()
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		k := item.KeyCopy(nil)
		key := &objs.TxCacheKey{}
		err := key.UnmarshalBinary(k)
		if err != nil {
			keys = append(keys, utils.CopySlice(k))
			continue
		}
		if key.Height > beforeHeight {
			break
		}
		if key.Height < beforeHeight {
			keys = append(keys, utils.CopySlice(k))
		}
		if len(keys) >= maxKeys {
			break
		}
	}
	for i := 0; i < len(keys); i++ {
		k := keys[i]
		err := utils.DeleteValue(txn, utils.CopySlice(k))
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *Database) GetTxCacheSet(txn *badger.Txn, height uint32) ([][]byte, [][]byte, error) {
	txHashes := [][]byte{}
	txs := [][]byte{}
	prefix, err := db.makeTxCacheIterKey(height)
	if err != nil {
		return nil, nil, err
	}
	opts := badger.DefaultIteratorOptions
	it := txn.NewIterator(opts)
	defer it.Close()
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		k := item.KeyCopy(nil)
		key := &objs.TxCacheKey{}
		err := key.UnmarshalBinary(k)
		if err != nil {
			return nil, nil, err
		}
		if key.Height == height {
			tx, err := item.ValueCopy(nil)
			if err != nil {
				return nil, nil, err
			}
			txs = append(txs, tx)
			txHashes = append(txHashes, key.TxHash)
		}
	}
	return txs, txHashes, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makePendingHdrNodeKey(nodeKey []byte) ([]byte, error) {
	prefix := dbprefix.PrefixPendingHdrNodeKey()
	key := &objs.PendingNodeKey{
		Prefix: prefix,
		Key:    utils.CopySlice(nodeKey),
	}
	return key.MarshalBinary()
}

func (db *Database) makePendingHdrNodeKeyIterKey() []byte {
	prefix := dbprefix.PrefixPendingHdrNodeKey()
	return prefix
}

func (db *Database) DropPendingHdrNodeKeys() error {
	return db.rawDB.DropPrefix(db.makePendingHdrNodeKeyIterKey())
}

func (db *Database) SetPendingHdrNodeKey(txn *badger.Txn, nodeKey []byte, layer int) error {
	pnkey, err := db.makePendingHdrNodeKey(nodeKey)
	if err != nil {
		return err
	}
	layerBytes := utils.MarshalUint16(uint16(layer))
	return utils.SetValue(txn, pnkey, layerBytes)
}

func (db *Database) GetPendingHdrNodeKey(txn *badger.Txn, nodeKey []byte) (int, error) {
	pnkey, err := db.makePendingHdrNodeKey(nodeKey)
	if err != nil {
		return 0, err
	}
	layerBytes, err := utils.GetValue(txn, pnkey)
	if err != nil {
		return 0, err
	}
	layer, err := utils.UnmarshalUint16(layerBytes)
	if err != nil {
		return 0, err
	}
	return int(layer), nil
}

func (db *Database) DeletePendingHdrNodeKey(txn *badger.Txn, nodeKey []byte) error {
	pnkey, err := db.makePendingHdrNodeKey(nodeKey)
	if err != nil {
		return err
	}
	return utils.DeleteValue(txn, pnkey)
}

func (db *Database) CountPendingHdrNodeKeys(txn *badger.Txn) (int, error) {
	count := 0
	iter := db.GetPendingHdrNodeKeysIter(txn)
	defer iter.Close()
	for {
		_, _, isDone, err := iter.Next()
		if err != nil {
			return 0, err
		}
		if isDone {
			break
		}
		count++
	}
	return count, nil
}

func (db *Database) GetPendingHdrNodeKeysIter(txn *badger.Txn) *PendingHdrNodeIter {
	prefix := db.makePendingHdrNodeKeyIterKey()
	opts := badger.IteratorOptions{
		PrefetchSize:   100,
		PrefetchValues: true,
		Prefix:         prefix,
	}
	it := txn.NewIterator(opts)
	seek := []byte{}
	seek = append(seek, utils.CopySlice(prefix)...)
	seek = append(seek, make([]byte, 32)...)
	it.Seek(seek)
	return &PendingHdrNodeIter{it: it, prefixLen: len(prefix)}
}

type PendingHdrNodeIter struct {
	it        *badger.Iterator
	prefixLen int
}

func (pni *PendingHdrNodeIter) Next() ([]byte, int, bool, error) {
	var key []byte
	var value int
	var isDone bool
	err := func() error {
		if !pni.it.Valid() {
			isDone = true
			return nil
		}
		defer pni.it.Next()
		itm := pni.it.Item()
		tmpKey := itm.KeyCopy(nil)
		valueBytes, err := itm.ValueCopy(nil)
		if err != nil {
			return err
		}
		tmpValue, err := utils.UnmarshalUint16(valueBytes)
		if err != nil {
			return err
		}
		key = tmpKey[pni.prefixLen:]
		value = int(tmpValue)
		return nil
	}()
	if err != nil {
		return nil, 0, isDone, err
	}
	return key, value, isDone, nil
}

func (pni *PendingHdrNodeIter) Close() {
	pni.it.Close()
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makePendingNodeKey(nodeKey []byte) ([]byte, error) {
	prefix := dbprefix.PrefixPendingNodeKey()
	key := &objs.PendingNodeKey{
		Prefix: prefix,
		Key:    utils.CopySlice(nodeKey),
	}
	return key.MarshalBinary()
}

func (db *Database) makePendingNodeKeyIterKey() []byte {
	prefix := dbprefix.PrefixPendingNodeKey()
	return prefix
}

func (db *Database) DropPendingNodeKeys() error {
	return db.rawDB.DropPrefix(db.makePendingNodeKeyIterKey())
}

func (db *Database) SetPendingNodeKey(txn *badger.Txn, nodeKey []byte, layer int) error {
	pnkey, err := db.makePendingNodeKey(nodeKey)
	if err != nil {
		return err
	}
	layerBytes := utils.MarshalUint16(uint16(layer))
	return utils.SetValue(txn, pnkey, layerBytes)
}

func (db *Database) GetPendingNodeKey(txn *badger.Txn, nodeKey []byte) (int, error) {
	pnkey, err := db.makePendingNodeKey(nodeKey)
	if err != nil {
		return 0, err
	}
	layerBytes, err := utils.GetValue(txn, pnkey)
	if err != nil {
		return 0, err
	}
	layer, err := utils.UnmarshalUint16(layerBytes)
	if err != nil {
		return 0, err
	}
	return int(layer), nil
}

func (db *Database) DeletePendingNodeKey(txn *badger.Txn, nodeKey []byte) error {
	pnkey, err := db.makePendingNodeKey(nodeKey)
	if err != nil {
		return err
	}
	return utils.DeleteValue(txn, pnkey)
}

func (db *Database) CountPendingNodeKeys(txn *badger.Txn) (int, error) {
	count := 0
	iter := db.GetPendingNodeKeysIter(txn)
	defer iter.Close()
	for {
		_, _, isDone, err := iter.Next()
		if err != nil {
			return 0, err
		}
		if isDone {
			break
		}
		count++
	}
	return count, nil
}

func (db *Database) GetPendingNodeKeysIter(txn *badger.Txn) *PendingNodeIter {
	prefix := db.makePendingNodeKeyIterKey()
	opts := badger.IteratorOptions{
		PrefetchSize:   100,
		PrefetchValues: true,
		Prefix:         prefix,
	}
	it := txn.NewIterator(opts)
	seek := []byte{}
	seek = append(seek, utils.CopySlice(prefix)...)
	seek = append(seek, make([]byte, 32)...)
	it.Seek(seek)
	return &PendingNodeIter{it: it, prefixLen: len(prefix)}
}

type PendingNodeIter struct {
	it        *badger.Iterator
	prefixLen int
}

func (pni *PendingNodeIter) Next() ([]byte, int, bool, error) {
	var key []byte
	var value int
	var isDone bool
	err := func() error {
		if !pni.it.Valid() {
			isDone = true
			return nil
		}
		defer pni.it.Next()
		itm := pni.it.Item()
		tmpKey := itm.KeyCopy(nil)
		valueBytes, err := itm.ValueCopy(nil)
		if err != nil {
			return err
		}
		tmpValue, err := utils.UnmarshalUint16(valueBytes)
		if err != nil {
			return err
		}
		key = tmpKey[pni.prefixLen:]
		value = int(tmpValue)
		return nil
	}()
	if err != nil {
		return nil, 0, isDone, err
	}
	return key, value, isDone, nil
}

func (pni *PendingNodeIter) Close() {
	pni.it.Close()
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makePendingLeafKey(leafKey []byte) ([]byte, error) {
	prefix := dbprefix.PrefixPendingLeafKey()
	key := &objs.PendingLeafKey{
		Prefix: prefix,
		Key:    utils.CopySlice(leafKey),
	}
	return key.MarshalBinary()
}

func (db *Database) makePendingLeafKeyIterKey() []byte {
	prefix := dbprefix.PrefixPendingLeafKey()
	return prefix
}

func (db *Database) DropPendingLeafKeys() error {
	return db.rawDB.DropPrefix(db.makePendingLeafKeyIterKey())
}

func (db *Database) SetPendingLeafKey(txn *badger.Txn, leafKey []byte, value []byte) error {
	plkey, err := db.makePendingLeafKey(leafKey)
	if err != nil {
		return err
	}
	return utils.SetValue(txn, plkey, value)
}

func (db *Database) GetPendingLeafKey(txn *badger.Txn, leafKey []byte) ([]byte, error) {
	plkey, err := db.makePendingLeafKey(leafKey)
	if err != nil {
		return nil, err
	}
	value, err := utils.GetValue(txn, plkey)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (db *Database) DeletePendingLeafKey(txn *badger.Txn, leafKey []byte) error {
	plkey, err := db.makePendingLeafKey(leafKey)
	if err != nil {
		return err
	}
	return utils.DeleteValue(txn, plkey)
}

func (db *Database) CountPendingLeafKeys(txn *badger.Txn) (int, error) {
	count := 0
	iter := db.GetPendingLeafKeysIter(txn)
	defer iter.Close()
	for {
		_, _, isDone, err := iter.Next()
		if err != nil {
			return 0, err
		}
		if isDone {
			break
		}
		count++
	}
	return count, nil
}

func (db *Database) GetPendingLeafKeysIter(txn *badger.Txn) *PendingLeafIter {
	prefix := db.makePendingLeafKeyIterKey()
	opts := badger.IteratorOptions{
		PrefetchSize:   100,
		PrefetchValues: true,
		Prefix:         prefix,
	}
	it := txn.NewIterator(opts)
	seek := []byte{}
	seek = append(seek, utils.CopySlice(prefix)...)
	seek = append(seek, make([]byte, 32)...)
	it.Seek(seek)
	return &PendingLeafIter{it: it, prefixLen: len(prefix)}
}

type PendingLeafIter struct {
	it        *badger.Iterator
	prefixLen int
}

func (pni *PendingLeafIter) Next() ([]byte, []byte, bool, error) {
	var key []byte
	var value []byte
	var isDone bool
	err := func() error {
		if !pni.it.Valid() {
			isDone = true
			return nil
		}
		defer pni.it.Next()
		itm := pni.it.Item()
		tmpKey := itm.KeyCopy(nil)
		tmpValue, err := itm.ValueCopy(nil)
		if err != nil {
			return err
		}
		key = tmpKey[pni.prefixLen:]
		value = tmpValue
		return nil
	}()
	if err != nil {
		return nil, nil, isDone, err
	}
	return key, value, isDone, nil
}

func (pni *PendingLeafIter) Close() {
	pni.it.Close()
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) SetSafeToProceed(txn *badger.Txn, height uint32, isSafe bool) error {
	key := &objs.SafeToProceedKey{Prefix: dbprefix.PrefixSafeToProceed(), Height: height}
	k, err := key.MarshalBinary()
	if err != nil {
		return err
	}
	if isSafe {
		return utils.SetValue(txn, k, []byte{1})
	}
	return utils.SetValue(txn, k, []byte{0})
}

func (db *Database) GetSafeToProceed(txn *badger.Txn, height uint32) (bool, error) {
	key := &objs.SafeToProceedKey{Prefix: dbprefix.PrefixSafeToProceed(), Height: height}
	k, err := key.MarshalBinary()
	if err != nil {
		return false, err
	}
	v, err := db.rawDB.getValue(txn, k)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return false, err
		}
		return false, nil
	}
	if len(v) != 1 {
		return false, nil
	}
	switch v[0] {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) ContainsSnapShotHdrNode(txn *badger.Txn, root []byte) (bool, error) {
	node, err := db.GetSnapShotHdrNode(txn, root)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return false, err
		}
		return false, nil
	}
	if len(node) == 0 {
		return false, nil
	}
	return true, nil
}

func (db *Database) GetSnapShotHdrNode(txn *badger.Txn, root []byte) ([]byte, error) {
	return db.trie.GetSnapShotHdrNode(txn, root)
}

func (db *Database) SetSnapShotHdrNode(txn *badger.Txn, batch []byte, root []byte, layer int) ([][]byte, int, []trie.LeafNode, error) {
	return db.trie.StoreSnapShotHdrNode(txn, batch, root, layer)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makePendingHdrLeafKey(hdrLeafKey []byte) ([]byte, error) {
	prefix := dbprefix.PrefixPendingHdrLeafKey()
	key := &objs.PendingHdrLeafKey{
		Prefix: prefix,
		Key:    utils.CopySlice(hdrLeafKey),
	}
	return key.MarshalBinary()
}

func (db *Database) makePendingHdrLeafKeyIterKey() []byte {
	prefix := dbprefix.PrefixPendingHdrLeafKey()
	return prefix
}

func (db *Database) DropPendingHdrLeafKeys() error {
	return db.rawDB.DropPrefix(db.makePendingHdrLeafKeyIterKey())
}

func (db *Database) SetPendingHdrLeafKey(txn *badger.Txn, hdrLeafKey []byte, value []byte) error {
	phlkey, err := db.makePendingHdrLeafKey(hdrLeafKey)
	if err != nil {
		return err
	}
	return utils.SetValue(txn, phlkey, value)
}

func (db *Database) GetPendingHdrLeafKey(txn *badger.Txn, hdrLeafKey []byte) ([]byte, error) {
	phlkey, err := db.makePendingHdrLeafKey(hdrLeafKey)
	if err != nil {
		return nil, err
	}
	value, err := utils.GetValue(txn, phlkey)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (db *Database) DeletePendingHdrLeafKey(txn *badger.Txn, hdrLeafKey []byte) error {
	phlkey, err := db.makePendingHdrLeafKey(hdrLeafKey)
	if err != nil {
		return err
	}
	return utils.DeleteValue(txn, phlkey)
}

func (db *Database) CountPendingHdrLeafKeys(txn *badger.Txn) (int, error) {
	count := 0
	iter := db.GetPendingHdrLeafKeysIter(txn)
	defer iter.Close()
	for {
		_, _, isDone, err := iter.Next()
		if err != nil {
			return 0, err
		}
		if isDone {
			break
		}
		count++
	}
	return count, nil
}

func (db *Database) GetPendingHdrLeafKeysIter(txn *badger.Txn) *PendingHdrLeafIter {
	prefix := db.makePendingHdrLeafKeyIterKey()
	opts := badger.IteratorOptions{
		PrefetchSize:   100,
		PrefetchValues: true,
		Prefix:         prefix,
	}
	it := txn.NewIterator(opts)
	seek := []byte{}
	seek = append(seek, utils.CopySlice(prefix)...)
	seek = append(seek, make([]byte, 32)...)
	it.Seek(seek)
	return &PendingHdrLeafIter{it: it, prefixLen: len(prefix)}
}

type PendingHdrLeafIter struct {
	it        *badger.Iterator
	prefixLen int
}

func (phli *PendingHdrLeafIter) Next() ([]byte, []byte, bool, error) {
	var key []byte
	var value []byte
	var isDone bool
	err := func() error {
		if !phli.it.Valid() {
			isDone = true
			return nil
		}
		defer phli.it.Next()
		itm := phli.it.Item()
		tmpKey := itm.KeyCopy(nil)
		tmpValue, err := itm.ValueCopy(nil)
		if err != nil {
			return err
		}
		key = tmpKey[phli.prefixLen:]
		value = tmpValue
		return nil
	}()
	if err != nil {
		return nil, nil, isDone, err
	}
	return key, value, isDone, nil
}

func (phli *PendingHdrLeafIter) Close() {
	phli.it.Close()
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (db *Database) makeStagedBlockHeaderKey(height uint32) ([]byte, error) {
	prefix := dbprefix.PrefixStagedBlockHeaderKey()
	heightBytes := utils.MarshalUint32(height)
	key := &objs.StagedBlockHeaderKey{
		Prefix: prefix,
		Key:    heightBytes,
	}
	return key.MarshalBinary()
}

func (db *Database) makeStagedBlockHeaderKeyIterKey() []byte {
	prefix := dbprefix.PrefixStagedBlockHeaderKey()
	return prefix
}

func (db *Database) DropStagedBlockHeaderKeys() error {
	return db.rawDB.DropPrefix(db.makeStagedBlockHeaderKeyIterKey())
}

func (db *Database) SetStagedBlockHeader(txn *badger.Txn, height uint32, value []byte) error {
	sbhkey, err := db.makeStagedBlockHeaderKey(height)
	if err != nil {
		return err
	}
	return utils.SetValue(txn, sbhkey, value)
}

func (db *Database) GetStagedBlockHeader(txn *badger.Txn, height uint32) (*objs.BlockHeader, error) {
	sbhkey, err := db.makeStagedBlockHeaderKey(height)
	if err != nil {
		return nil, err
	}
	vBytes, err := utils.GetValue(txn, sbhkey)
	if err != nil {
		return nil, err
	}
	v := &objs.BlockHeader{}
	err = v.UnmarshalBinary(vBytes)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (db *Database) DeleteStagedBlockHeaderKey(txn *badger.Txn, height uint32) error {
	sbhkey, err := db.makeStagedBlockHeaderKey(height)
	if err != nil {
		return err
	}
	return utils.DeleteValue(txn, sbhkey)
}

func (db *Database) CountStagedBlockHeaderKeys(txn *badger.Txn) (int, error) {
	count := 0
	iter := db.GetStagedBlockHeaderKeyIter(txn)
	defer iter.Close()
	for {
		_, _, isDone, err := iter.Next()
		if err != nil {
			return 0, err
		}
		if isDone {
			break
		}
		count++
	}
	return count, nil
}

type StagedBlockHeaderKeyIter struct {
	it        *badger.Iterator
	prefixLen int
}

func (db *Database) GetStagedBlockHeaderKeyIter(txn *badger.Txn) *StagedBlockHeaderKeyIter {
	prefix := db.makeStagedBlockHeaderKeyIterKey()
	opts := badger.IteratorOptions{
		PrefetchSize:   100,
		PrefetchValues: true,
		Prefix:         prefix,
	}
	it := txn.NewIterator(opts)
	seek := []byte{}
	seek = append(seek, utils.CopySlice(prefix)...)
	seek = append(seek, []byte{0, 0, 0, 0}...)
	it.Seek(seek)
	return &StagedBlockHeaderKeyIter{it: it, prefixLen: len(prefix)}
}

func (sbhki *StagedBlockHeaderKeyIter) Next() (uint32, *objs.BlockHeader, bool, error) {
	var height uint32
	var bhdr *objs.BlockHeader
	var isDone bool
	err := func() error {
		if !sbhki.it.Valid() {
			isDone = true
			return nil
		}
		defer sbhki.it.Next()
		itm := sbhki.it.Item()
		key := itm.KeyCopy(nil)
		value, err := itm.ValueCopy(nil)
		if err != nil {
			return err
		}
		tmpHeight, err := utils.UnmarshalUint32(key[sbhki.prefixLen:])
		if err != nil {
			return err
		}
		tmpBhdr := &objs.BlockHeader{}
		err = tmpBhdr.UnmarshalBinary(value)
		if err != nil {
			return err
		}
		height = tmpHeight
		bhdr = tmpBhdr
		return nil
	}()
	if err != nil {
		return 0, nil, isDone, err
	}
	return height, bhdr, isDone, nil
}

func (sbhki *StagedBlockHeaderKeyIter) Close() {
	sbhki.it.Close()
}
