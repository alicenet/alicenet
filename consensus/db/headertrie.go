package db

import (
	"errors"

	"github.com/MadBase/MadNet/constants/dbprefix"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/sirupsen/logrus"

	trie "github.com/MadBase/MadNet/badgerTrie"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/dgraph-io/badger/v2"
)

// on completion of block one headerRoot_1 is determined
//    block one has the headerRoot of all ONES
//    block one is inserted into trie during initialization
// on block two a proposal is made, this includes headerRoot_1
//    block two is  inserted into trie after NextHeight is finalized
//    the trie now has both headerRoot_1 and headerRoot_2
// on block two a proposal is made, this includes headerRoot_2
//  ...

type headerTrie struct {
	logger *logrus.Logger
}

func (ht *headerTrie) init() {
	logger := logging.GetLogger(constants.LoggerDB)
	ht.logger = logger
}

func (ht *headerTrie) VerifyProof(txn *badger.Txn, rootHash []byte, root []byte, proof []byte, blockHeader *objs.BlockHeader) (bool, error) {
	tr := trie.NewSMT(rootHash, crypto.Hasher, dbprefix.PrefixBlockHeaderTrie)
	mproof := &MerkleProof{}
	err := mproof.UnmarshalBinary(proof)
	if err != nil {
		return false, err
	}
	if !mproof.Included {
		return false, nil
	}
	heightBytes := utils.MarshalUint32(blockHeader.BClaims.Height)
	key := make([]byte, constants.HashLen)
	copy(key[:], heightBytes)
	return tr.VerifyInclusionCR(root, mproof.Bitmap, key, mproof.ProofValue, mproof.Path, mproof.KeyHeight), nil
}

func (ht *headerTrie) GetProof(txn *badger.Txn, rootHash []byte, root []byte, height uint32) (bool, []byte, error) {
	tr := trie.NewSMT(rootHash, crypto.Hasher, dbprefix.PrefixBlockHeaderTrie)
	heightBytes := utils.MarshalUint32(height)
	key := make([]byte, constants.HashLen)
	copy(key[:], heightBytes)
	bitmap, mp, keyheight, included, proofKey, proofVal, err := tr.MerkleProofCompressedR(txn, key, root)
	if err != nil {
		return false, nil, err
	}
	mproof := &MerkleProof{
		Included:   included,
		KeyHeight:  keyheight,
		Key:        key,
		ProofKey:   proofKey,
		ProofValue: proofVal,
		Bitmap:     bitmap,
		Path:       mp,
	}
	mpbytes, err := mproof.MarshalBinary()
	if err != nil {
		return false, nil, err
	}
	return included, mpbytes, nil
}

func (ht *headerTrie) Get(txn *badger.Txn, rootHash []byte, height uint32) ([]byte, error) {
	tr := trie.NewSMT(rootHash, crypto.Hasher, dbprefix.PrefixBlockHeaderTrie)
	heightBytes := utils.MarshalUint32(height)
	key := make([]byte, constants.HashLen)
	copy(key[:], heightBytes)
	hsh, err := tr.Get(txn, key)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return nil, err
		}
	}
	return hsh, nil
}

func (ht *headerTrie) Contains(txn *badger.Txn, rootHash []byte, height uint32) (bool, error) {
	tr := trie.NewSMT(rootHash, crypto.Hasher, dbprefix.PrefixBlockHeaderTrie)
	heightBytes := utils.MarshalUint32(height)
	key := make([]byte, constants.HashLen)
	copy(key[:], heightBytes)
	hsh, err := tr.Get(txn, key)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return false, err
		}
		return false, nil
	}
	if len(hsh) > 0 {
		return true, nil
	}
	return false, nil
}

func (ht *headerTrie) ApplyState(txn *badger.Txn, newHeader *objs.BlockHeader, deleteHeader uint32) ([]byte, error) {
	if newHeader != nil && deleteHeader != 0 {
		return nil, errors.New("only one operation allowed at a time")
	}
	if newHeader == nil && deleteHeader == 0 {
		return nil, errors.New("header and deleteHeader are no-op values")
	}
	res, err := ht.update(txn, newHeader, deleteHeader)
	if err != nil {
		utils.DebugTrace(ht.logger, err)
		return nil, err
	}
	return res, nil
}

func (ht *headerTrie) update(txn *badger.Txn, newHeader *objs.BlockHeader, deleteHeader uint32) ([]byte, error) {
	var tr *trie.SMT
	if newHeader != nil {
		if newHeader.BClaims.Height == 1 {
			tr = trie.NewSMT(nil, crypto.Hasher, dbprefix.PrefixBlockHeaderTrie)
		} else {
			tmp, err := trie.NewSMTForHeight(txn, newHeader.BClaims.Height-1, crypto.Hasher, dbprefix.PrefixBlockHeaderTrie)
			if err != nil {
				ht.logger.Errorf("Error in ht.update at trie.NewSMTFortHeight (1): %v", err)
				return nil, err
			}
			tr = tmp
		}
	}
	if deleteHeader > 0 {
		if deleteHeader == 1 {
			tr = trie.NewSMT(nil, crypto.Hasher, dbprefix.PrefixBlockHeaderTrie)
			return tr.Root, nil
		} else {
			tmp, err := trie.NewSMTForHeight(txn, deleteHeader-1, crypto.Hasher, dbprefix.PrefixBlockHeaderTrie)
			if err != nil {
				utils.DebugTrace(ht.logger, err)
				return nil, err
			}
			tr = tmp
			return tr.Root, nil
		}
	}
	if newHeader != nil && tr != nil {
		hsh, err := newHeader.BClaims.BlockHash()
		if err != nil {
			utils.DebugTrace(ht.logger, err)
			return nil, err
		}
		height := newHeader.BClaims.Height
		if height == 0 {
			utils.DebugTrace(ht.logger, err)
			return nil, errorz.ErrInvalid{}.New("height zero in hdr trie")
		}
		key := makeTrieKeyFromHeight(height)
		if _, err := tr.Update(txn, [][]byte{key}, [][]byte{hsh}); err != nil {
			utils.DebugTrace(ht.logger, err)
			return nil, err
		}
		return tr.Commit(txn, newHeader.BClaims.Height)
	}
	panic("headerTrie no-op")
}

func (ht *headerTrie) StoreSnapShotHdrNode(txn *badger.Txn, batch []byte, root []byte, layer int) ([][]byte, int, []trie.LeafNode, error) {
	t := trie.NewSMT(root, crypto.Hasher, dbprefix.PrefixBlockHeaderTrie)
	return t.StoreSnapShotNode(txn, batch, root, layer)
}

func (ht *headerTrie) GetSnapShotHdrNode(txn *badger.Txn, root []byte) ([]byte, error) {
	return trie.GetNodeDB(txn, dbprefix.PrefixBlockHeaderTrie(), root)
}

func (ht *headerTrie) FinalizeSnapShotHdrRoot(txn *badger.Txn, root []byte, height uint32) error {
	t := trie.NewSMT(root, crypto.Hasher, dbprefix.PrefixBlockHeaderTrie)
	return t.FinalizeSnapShotRoot(txn, root, height)
}

func makeTrieKeyFromHeight(height uint32) []byte {
	heightBytes := utils.MarshalUint32(height)
	key := make([]byte, constants.HashLen)
	copy(key[:], heightBytes)
	return key
}
