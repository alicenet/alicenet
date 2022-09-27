package accusation

import (
	"bytes"
	"encoding/hex"
	"log"
	"math/big"

	aobjs "github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/utxohandler/utxotrie"
	trie "github.com/alicenet/alicenet/badgerTrie"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/executor/tasks/accusations"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

func detectInvalidUTXOConsumption(rs *objs.RoundState, lrs *lstate.RoundStates, db *db.Database) (tasks.Task, bool) {

	// if rs.Proposal.PClaims.BClaims.TxCount <= 0 {
	// 	return nil, false
	// }

	// rs.
	// if rs.Proposal.PClaims.BClaims.Height+1 != rs.Proposal.PClaims.BClaims.Height {
	// 	return nil, false
	// }
	logger := logging.GetLogger("accusations")
	var err error
	var accusation *accusations.InvalidUTXOConsumptionAccusationTask

	for _, txHash := range rs.Proposal.TxHshLst {
		// get Tx by hash from DB
		err = db.View(func(txn *badger.Txn) error {
			txBin, err := db.GetTxCacheItem(txn, rs.Proposal.PClaims.BClaims.Height, txHash)
			if err != nil {
				return err
			}

			tx := &aobjs.Tx{}
			err = tx.UnmarshalBinary(txBin)
			if err != nil {
				return err
			}

			var vins []*aobjs.TXIn = tx.Vin

			for _, vin := range vins {
				utxoID, err := vin.UTXOID()
				if err != nil {
					logger.Warnf("error getting utxoID: %v", err)
					return err
				}

				mStateRootProof, err := getStateMerkleProofs(txn, utxoID, logger)
				if err != nil {
					logger.Warnf("error getting merkel proof against state root: %v", err)
					return err
				}

				// if utxoID is included in state trie then continue bc all is fine
				// else generate proof of non-inclusion and accuse
				if !mStateRootProof.Included {
					pClaimsBin, err := rs.Proposal.PClaims.MarshalBinary()
					if err != nil {
						logger.Warnf("error marshalling PClaims: %v", err)
						return err
					}
					//rs.Proposal.Signature
					bClaimsBin, err := rs.Proposal.PClaims.BClaims.MarshalBinary()
					if err != nil {
						logger.Warnf("error marshalling BClaims: %v", err)
						return err
					}

					// TXInPreImage
					txInPreImageBin, err := vin.TXInLinker.TXInPreImage.MarshalBinary()
					if err != nil {
						logger.Warnf("error marshalling TXInPreImage: %v", err)
						return err
					}

					var proofs [3][]byte

					// proofs[0] = proofAgainstStateRoot
					proofs[0], err = mStateRootProof.MarshalBinary()
					if err != nil {
						logger.Warnf("error marshalling merkel proof against state root: %v", err)
						return err
					}

					// proofs[1] = proofInclusionTxRoot - ProofInclusionTxRoot against PClaims.BClaims.TxRoot.
					mTxRootProof, err := getTxRootMerkleProofs(txn, txHash, db)
					if err != nil {
						logger.Warnf("error getting merkel proof against tx root: %v", err)
						return err
					}
					proofs[1], err = mTxRootProof.MarshalBinary()
					if err != nil {
						logger.Warnf("error marshalling merkel proof against tx root: %v", err)
						return err
					}

					// proofs[2] = proofOfInclusionTxHash - ProofOfInclusionTxHash against the target hash from ProofInclusionTxRoot
					mTxHashProof, err := getTxHashProof(txn, tx, utxoID)
					if err != nil {
						logger.Warnf("error getting merkel hash against tx root: %v", err)
						return err
					}
					proofs[2], err = mTxHashProof.MarshalBinary()
					if err != nil {
						logger.Warnf("error marshalling merkel proof against tx hash: %v", err)
						return err
					}

					// prepare accusation task
					accusation = accusations.NewInvalidUTXOConsumptionAccusationTask(
						pClaimsBin,
						rs.Proposal.Signature,
						bClaimsBin,
						rs.Proposal.PClaims.RCert.SigGroup,
						txInPreImageBin,
						proofs,
					)

					return nil
				}
			}

			if err != nil {
				logger.Warnf("error processing vins: %v", err)
				return err
			}

			return nil
		})

		if err != nil {
			logger.Warnf("error processing txHashes: %v", err)
			// continue and find what we can
		}
	}

	if accusation != nil {
		// deterministic accusation ID
		var chainID []byte = common.LeftPadBytes(big.NewInt(int64(rs.Proposal.PClaims.RCert.RClaims.ChainID)).Bytes(), 4)
		var height []byte = common.LeftPadBytes(big.NewInt(int64(rs.Proposal.PClaims.RCert.RClaims.Height)).Bytes(), 4)
		var round []byte = common.LeftPadBytes(big.NewInt(int64(rs.Proposal.PClaims.RCert.RClaims.Round)).Bytes(), 4)
		var preSalt []byte = crypto.Hasher([]byte("AccusationMultipleProposal"))

		var id []byte = crypto.Hasher(
			rs.Proposal.Proposer,
			chainID,
			height,
			round,
			preSalt,
		)
		accusation.ID = hex.EncodeToString(id)

		return accusation, true
	}

	return nil, false
}

func getStateMerkleProofs(txn *badger.Txn, utxoID []byte, logger *logrus.Logger) (*db.MerkleProof, error) {
	root, err := utxotrie.GetCurrentStateRoot(txn)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.DebugTrace(logger, err)
			return nil, err
		}
	}
	if bytes.Equal(root, make([]byte, constants.HashLen)) {
		root = nil
	}
	stateTrie := trie.NewSMT(root, trie.Hasher, func() []byte { return dbprefix.PrefixUTXOTrie() })

	bitmap, auditPath, proofHeight, included, proofKey, proofVal, err := stateTrie.MerkleProofCompressed(txn, utxoID)
	if err != nil {
		return nil, err
	}
	//included := stateTrie.VerifyInclusionC(bitmap, utxoID, proofVal, auditPath, proofHeight)
	//logger.Debug("Is the proof compacted included in the trie: ", result)
	mproof := &db.MerkleProof{
		Included:   included,
		KeyHeight:  proofHeight,
		Key:        utxoID,
		ProofKey:   proofKey,
		ProofValue: proofVal,
		Bitmap:     bitmap,
		Path:       auditPath,
	}

	return mproof, nil
}

func getTxRootMerkleProofs(txn *badger.Txn, txHash []byte, consDB *db.Database) (*db.MerkleProof, error) {
	newTxn := consDB.DB().NewTransaction(true)
	defer newTxn.Discard()
	txRootSMT, _, err := MakePersistentTxRoot(newTxn, [][]byte{txHash})
	if err != nil {
		return nil, err
	}
	bitmap, auditPath, proofHeight, included, proofKey, proofVal, err := txRootSMT.MerkleProofCompressed(txn, txHash)
	if err != nil {
		return nil, err
	}

	mproof := &db.MerkleProof{
		Included:   included,
		KeyHeight:  proofHeight,
		Key:        txHash,
		ProofKey:   proofKey,
		ProofValue: proofVal,
		Bitmap:     bitmap,
		Path:       auditPath,
	}

	return mproof, nil
}

// Creates a persistent TxRoot Merkle Trie from a list of transactions. This
// function is necessary since this Trie only lives on memory normally.
func MakePersistentTxRoot(txn *badger.Txn, txHashes [][]byte) (*trie.SMT, []byte, error) {
	if len(txHashes) == 0 {
		return nil, crypto.Hasher([]byte{}), nil
	}
	values := [][]byte{}
	for i := 0; i < len(txHashes); i++ {
		txHash := txHashes[i]
		values = append(values, crypto.Hasher(txHash))
	}
	// new in persistent smt
	smt := trie.NewSMT(nil, trie.Hasher, func() []byte { return []byte("mtr") })
	// smt update
	txHashesSorted, valuesSorted, err := utils.SortKVs(txHashes, values)
	if err != nil {
		return nil, nil, err
	}
	rootHash, err := smt.Update(txn, txHashesSorted, valuesSorted)
	if err != nil {
		return nil, nil, err
	}
	return smt, rootHash, nil
}

// Creates a persistent TxHash Merkle Trie from a transaction. This function is
// necessary since this Trie only lives on memory normally.
func PersistentTxHash(txn *badger.Txn, b *aobjs.Tx, utxoID []byte) (*trie.SMT, []byte, error) {
	if b == nil {
		return nil, nil, errorz.ErrInvalid{}.New("tx not initialized in txHash")
	}
	if err := b.Vout.SetTxOutIdx(); err != nil {
		return nil, nil, err
	}
	keys := [][]byte{}
	values := [][]byte{}
	for _, txIn := range b.Vin {
		id, err := txIn.UTXOID()
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, id)
		hsh, err := txIn.PreHash()
		if err != nil {
			return nil, nil, err
		}
		values = append(values, hsh)
	}
	for idx, txOut := range b.Vout {
		hsh, err := txOut.PreHash()
		if err != nil {
			return nil, nil, err
		}
		id := aobjs.MakeUTXOID(utils.CopySlice(hsh), uint32(idx))
		keys = append(keys, id)
		values = append(values, hsh)
	}
	for i := range keys {
		log.Printf("TxHash Key: %x\n", keys[i])
		log.Printf("TxHash Value: %x\n", values[i])
	}
	// new in persistent smt
	smt := trie.NewSMT(nil, trie.Hasher, func() []byte { return []byte("mtt") })
	// smt update
	keysSorted, valuesSorted, err := utils.SortKVs(keys, values)
	if err != nil {
		return nil, nil, err
	}
	if len(keysSorted) == 0 && len(valuesSorted) == 0 {
		rootHash := crypto.Hasher([][]byte{}...)
		return smt, utils.CopySlice(rootHash), nil
	}
	rootHash, err := smt.Update(txn, keysSorted, valuesSorted)
	if err != nil {
		return nil, nil, err
	}
	return smt, utils.CopySlice(rootHash), nil
}

func getTxHashProof(txn *badger.Txn, tx *aobjs.Tx, utxoID []byte) (*db.MerkleProof, error) {
	txHashSMT, _, err := PersistentTxHash(txn, tx, utxoID)
	if err != nil {
		return nil, err
	}
	bitmap, auditPath, proofHeight, included, proofKey, proofVal, err := txHashSMT.MerkleProofCompressed(txn, utxoID)
	if err != nil {
		return nil, err
	}

	mproof := &db.MerkleProof{
		Included:   included,
		KeyHeight:  proofHeight,
		Key:        utxoID,
		ProofKey:   proofKey,
		ProofValue: proofVal,
		Bitmap:     bitmap,
		Path:       auditPath,
	}

	return mproof, nil
}

// assert detectInvalidUTXOConsumption is of type detector
var _ detector = detectInvalidUTXOConsumption
