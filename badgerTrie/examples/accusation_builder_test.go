package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"
	"testing"

	"github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/objs/uint256"
	utxo "github.com/alicenet/alicenet/application/utxohandler"
	trie "github.com/alicenet/alicenet/badgerTrie"
	"github.com/alicenet/alicenet/consensus/db"
	cobjs "github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

// Create deposits (valueStore) to be spend to create UTXOs
func makeDeposit(t *testing.T, s objs.Signer, chainID uint32, i int, value *uint256.Uint256) *objs.ValueStore {
	pubkey, err := s.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	vs := &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			TXOutIdx: constants.MaxUint32,
			Value:    value,
			ChainID:  chainID,
			Owner:    &objs.ValueStoreOwner{SVA: objs.ValueStoreSVA, CurveSpec: constants.CurveSecp256k1, Account: crypto.GetAccount(pubkey)},
		},
		TxHash: utils.ForceSliceToLength([]byte(strconv.Itoa(i)), constants.HashLen),
	}
	return vs
}

// Create transactions consuming deposits
func makeTxs(t *testing.T, s objs.Signer, v *objs.ValueStore) *objs.Tx {
	txIn, err := v.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	value, err := v.Value()
	if err != nil {
		t.Fatal(err)
	}
	chainID, err := txIn.ChainID()
	if err != nil {
		t.Fatal(err)
	}
	pubkey, err := s.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	tx := &objs.Tx{}
	tx.Vin = []*objs.TXIn{txIn}
	newValueStore := &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			ChainID:  chainID,
			Value:    value,
			Owner:    &objs.ValueStoreOwner{SVA: objs.ValueStoreSVA, CurveSpec: constants.CurveSecp256k1, Account: crypto.GetAccount(pubkey)},
			Fee:      new(uint256.Uint256).SetZero(),
			TXOutIdx: 0,
		},
		TxHash: make([]byte, 32),
	}
	newUTXO := &objs.TXOut{}
	err = newUTXO.NewValueStore(newValueStore)
	if err != nil {
		t.Fatal(err)
	}
	tx.Fee = new(uint256.Uint256).SetZero()
	tx.Vout = append(tx.Vout, newUTXO)
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	err = v.Sign(tx.Vin[0], s)
	if err != nil {
		t.Fatal(err)
	}
	return tx
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
func PersistentTxHash(txn *badger.Txn, b *objs.Tx) (*trie.SMT, []byte, error) {
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
		id := objs.MakeUTXOID(utils.CopySlice(hsh), uint32(idx))
		keys = append(keys, id)
		values = append(values, hsh)
	}
	for i, _ := range keys {
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

// Create a new UTXO spending a previous UTXO
func createNewUTXO(chainID uint32, value *uint256.Uint256, senderPubkey []byte, txHash []byte, txOutIdx uint32) (*objs.TXOut, error) {
	newValueStoreSender := &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			ChainID:  chainID,
			Value:    value,
			Owner:    &objs.ValueStoreOwner{SVA: objs.ValueStoreSVA, CurveSpec: constants.CurveSecp256k1, Account: crypto.GetAccount(senderPubkey)},
			Fee:      new(uint256.Uint256).SetZero(),
			TXOutIdx: txOutIdx,
		},
		TxHash: txHash,
	}

	newUTXO := &objs.TXOut{}
	err := newUTXO.NewValueStore(newValueStoreSender)
	if err != nil {
		return nil, err
	}
	return newUTXO, nil
}

// Creates a transaction transfering money from signer 1 to signer2
func makeTransfer(t *testing.T, sender objs.Signer, receiver objs.Signer, transferAmount uint64, v *objs.TXOut) *objs.Tx {
	txIn, err := v.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	txInHash, err := txIn.PreHash()
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("TXIn PreHash: %x\n", txInHash)
	value, err := v.Value()
	vuint64, err := value.ToUint64()
	returnedAmount := vuint64 - transferAmount
	value = &uint256.Uint256{}
	_, _ = value.FromUint64(returnedAmount)
	value2 := &uint256.Uint256{}
	_, _ = value2.FromUint64(transferAmount)

	if err != nil {
		t.Fatal(err)
	}
	chainID, err := txIn.ChainID()
	if err != nil {
		t.Fatal(err)
	}
	receiverPubkey, err := receiver.Pubkey()
	if err != nil {
		t.Fatal(err)
	}

	senderPubkey, err := sender.Pubkey()
	if err != nil {
		t.Fatal(err)
	}

	tx := &objs.Tx{}
	tx.Vin = []*objs.TXIn{txIn}
	newValueStoreSender := &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			ChainID:  chainID,
			Value:    value,
			Owner:    &objs.ValueStoreOwner{SVA: objs.ValueStoreSVA, CurveSpec: constants.CurveSecp256k1, Account: crypto.GetAccount(senderPubkey)},
			Fee:      new(uint256.Uint256).SetZero(),
			TXOutIdx: 0,
		},
		TxHash: make([]byte, 32),
	}

	// the new utxo that will be generated by this transaction
	newValueStoreReceiver := &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			ChainID:  chainID,
			Value:    value2,
			Owner:    &objs.ValueStoreOwner{SVA: objs.ValueStoreSVA, CurveSpec: constants.CurveSecp256k1, Account: crypto.GetAccount(receiverPubkey)},
			Fee:      new(uint256.Uint256).SetZero(),
			TXOutIdx: 1,
		},
		TxHash: make([]byte, 32),
	}
	newUTXOSender := &objs.TXOut{}
	err = newUTXOSender.NewValueStore(newValueStoreSender)
	if err != nil {
		t.Fatal(err)
	}
	txOutPreHash, err := newUTXOSender.PreHash()
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("TXOut PreHash: %x\n", txOutPreHash)
	newUTXOReceiver := &objs.TXOut{}
	err = newUTXOReceiver.NewValueStore(newValueStoreReceiver)
	if err != nil {
		t.Fatal(err)
	}
	tx.Fee = new(uint256.Uint256).SetZero()
	tx.Vout = append(tx.Vout, newUTXOSender, newUTXOReceiver)
	err = tx.SetTxHash() // <- compute the root from the TxHash smt
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

// Generate a block (bclaims) and append it to the chain (list of linked
// BClaims). This function is used to build the blockchain used by the tests.
func GenerateBlock(chain []*cobjs.BClaims, stateRoot []byte, txHshLst [][]byte) ([]*cobjs.BClaims, error) {
	var prevBlock []byte
	var headerRoot []byte
	if len(chain) == 0 {
		chain = []*cobjs.BClaims{}
		prevBlock = crypto.Hasher([]byte("foo"))
		headerRoot = crypto.Hasher([]byte(""))
	} else {
		_prevBlock, err := chain[len(chain)-1].BlockHash()
		if err != nil {
			return nil, err
		}
		prevBlock = _prevBlock
		headerRoot = crypto.Hasher([]byte("")) // todo: how to generate the block smt
	}
	txRoot, err := cobjs.MakeTxRoot(txHshLst) // generating the smt root
	log.Printf("txRoot height: (%d): %x\n", len(chain)+1, txRoot)
	if err != nil {
		if err != nil {
			return nil, err
		}
	}
	bclaims := &cobjs.BClaims{
		ChainID:    1,
		Height:     uint32(len(chain) + 1),
		TxCount:    uint32(len(txHshLst)),
		PrevBlock:  prevBlock,
		TxRoot:     txRoot,
		StateRoot:  stateRoot,
		HeaderRoot: headerRoot,
	}
	chain = append(chain, bclaims)

	log.Printf(
		"\nBlock: {\n\tChainID: %d\n\tHeight: %d\n\tTxCount: %d\n\tPrevBlock: %x\n\tTxRoot: %x\n\tStateRoot: %x\n\tHeaderRoot: %x\n}\n\n",
		bclaims.ChainID,
		bclaims.Height,
		bclaims.TxCount,
		bclaims.PrevBlock,
		bclaims.TxRoot,
		bclaims.StateRoot,
		bclaims.HeaderRoot,
	)
	return chain, nil
}

// Test the deserialization of the Merkle Proof strucs from the binary blob
func testMerkleProofDeserialization(mpbytes []byte, mproof *db.MerkleProof) error {
	mp1 := &db.MerkleProof{}
	err := mp1.UnmarshalBinary(mpbytes)
	if err != nil {
		return err
	}
	if mp1.Included != mproof.Included {
		return errors.New(fmt.Sprintf("bad height: %t Expected: %t", mp1.Included, mproof.Included))
	}
	if mp1.KeyHeight != mproof.KeyHeight {
		return errors.New(fmt.Sprintf("bad height: %d Expected: %d", mp1.KeyHeight, mproof.KeyHeight))
	}
	if !bytes.Equal(mp1.Key, mproof.Key) {
		return errors.New(fmt.Sprintf("bad Key: %x Expected %x", mp1.Key, mproof.Key))
	}
	if !bytes.Equal(mp1.ProofKey, mproof.ProofKey) {
		return errors.New(fmt.Sprintf("bad ProofKey: %x Expected: %x", mp1.ProofKey, mproof.ProofKey))
	}
	if !bytes.Equal(mp1.ProofValue, mproof.ProofValue) {
		return errors.New(fmt.Sprintf("bad Next: %x Expected: %x", mp1.ProofValue, mproof.ProofValue))
	}
	if !bytes.Equal(mp1.Bitmap, mproof.Bitmap) {
		return errors.New(fmt.Sprintf("bad Bitmap: %x Expected: %x", mp1.Bitmap, mproof.Bitmap))
	}
	for i := 0; i < len(mproof.Path); i++ {
		if !bytes.Equal(mp1.Path[i], mproof.Path[i]) {
			return errors.New(fmt.Sprintf("bad Path: %s Expected: %x", mp1.Path[i], mproof.Path[i]))
		}
	}
	return nil
}

// Get all Merkle Proofs (inclusion) from all transactions inserted in a block
func getAllStateMerkleProofs(hndlr *utxo.UTXOHandler, txs []*objs.Tx) func(txn *badger.Txn) error {
	fn := func(txn *badger.Txn) error {
		stateTrie, err := hndlr.GetTrie().GetCurrentTrie(txn)
		if err != nil {
			return err
		}
		log.Printf("Trie height: %d\n", stateTrie.TrieHeight)
		for _, tx := range txs {
			txHash, err := tx.TxHash()
			if err != nil {
				return err
			}
			log.Println("===========Proof of inclusion=========")
			log.Printf("Tx: %x\n", txHash)
			log.Println("======================================")
			utxoIDs, err := tx.GeneratedUTXOID()
			if err != nil {
				return err
			}
			for i, utxoID := range utxoIDs {
				//auditPath, included, proofKey, proofVal, err := stateTrie.MerkleProof(txn, utxoID) // *badger.Txn, key []byte
				bitmap, auditPath, proofHeight, included, proofKey, proofVal, err := stateTrie.MerkleProofCompressed(txn, utxoID)
				if err != nil {
					return err
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
				mpbytes, err := mproof.MarshalBinary()
				if err != nil {
					return err
				}
				err = testMerkleProofDeserialization(mpbytes, mproof)
				if err != nil {
					return err
				}
				log.Printf("UTXOID: %x\n", utxoID)
				log.Printf("auditPath: %x\n", auditPath)
				log.Printf("Bitmap: %x\n", bitmap)
				log.Printf("Proof height: %x\n", proofHeight)
				log.Print("Included:", included)
				log.Printf("Proof key: %x\n", proofKey)
				log.Printf("Proof value: %x\n", proofVal)
				log.Printf("Proof capnproto: %x\n", mpbytes)
				if len(utxoIDs) > i+1 {
					log.Println("---------------------")
				}
			}
			log.Println("======================================")
			log.Println()

		}
		return nil
	}
	return fn
}

// Get the proof of Inclusion or Exclusion of a certain UTXOID in the state Trie.
func getStateMerkleProofs(hndlr *utxo.UTXOHandler, txs []*objs.Tx, utxoID []byte) func(txn *badger.Txn) error {
	fn := func(txn *badger.Txn) error {
		stateTrie, err := hndlr.GetTrie().GetCurrentTrie(txn)
		if err != nil {
			return err
		}
		log.Println("===========Proof of inclusion Separate Key =========")
		log.Printf("Trie height: %d\n", stateTrie.TrieHeight)
		bitmap, auditPath, proofHeight, included, proofKey, proofVal, err := stateTrie.MerkleProofCompressed(txn, utxoID)
		if err != nil {
			return err
		}
		result := stateTrie.VerifyInclusionC(bitmap, utxoID, proofVal, auditPath, proofHeight)
		log.Print("Is the proof compacted included in the trie: ", result)
		_, err = CreateMerkleProof(
			included,
			proofHeight,
			utxoID,
			proofKey,
			proofVal,
			bitmap,
			auditPath,
		)
		if err != nil {
			return err
		}
		return nil
	}
	return fn
}

// Create a Merkle Proof struct from the data returned by the MerkleProofCompressed methods.
func CreateMerkleProof(included bool, proofHeight int, key []byte, proofKey []byte, proofVal []byte, bitmap []byte, auditPath [][]byte) (*db.MerkleProof, error) {
	mproof := &db.MerkleProof{
		Included:   included,
		KeyHeight:  proofHeight,
		Key:        key,
		ProofKey:   proofKey,
		ProofValue: proofVal,
		Bitmap:     bitmap,
		Path:       auditPath,
	}
	mpbytes, err := mproof.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = testMerkleProofDeserialization(mpbytes, mproof)
	if err != nil {
		return nil, err
	}
	log.Println("=========== Merkle Proof =========")
	log.Printf("Bitmap: %x\n", mproof.Bitmap)
	log.Printf("auditPathCompacted: %x\n", mproof.Path)
	log.Printf("Proof height: %x\n", mproof.KeyHeight)
	log.Print("Included:", mproof.Included)
	log.Printf("Key: %x", mproof.Key)
	log.Printf("Proof key: %x\n", mproof.ProofKey)
	log.Printf("Proof value: %x\n", mproof.ProofValue)
	log.Printf("Proof capnproto: %x\n", mpbytes)
	log.Println("======================================")
	log.Println()

	return mproof, nil
}

// Main Test function to generate the accusations
func TestGenerateAccusations(t *testing.T) {
	////////////////// Database setup ////////////////
	log.Println("Starting the generation of accusation test objects")
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	//////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	signer := &crypto.Secp256k1Signer{}
	err = signer.SetPrivk(crypto.Hasher([]byte("secret")))

	if err != nil {
		t.Fatal(err)
	}

	signer2 := &crypto.Secp256k1Signer{}
	err = signer2.SetPrivk(crypto.Hasher([]byte("secret2")))

	if err != nil {
		t.Fatal(err)
	}

	hndlr := utxo.NewUTXOHandler(db)
	err = hndlr.Init(1)
	if err != nil {
		t.Fatal(err)
	}

	///////// Block 1 ////////////
	log.Println("Block 1:")
	// Creating First UTXO
	var txs []*objs.Tx
	var deposits []*objs.ValueStore
	var txHshLst [][]byte
	var txHshSMTs []*trie.SMT
	for i := uint64(0); i < 5; i++ {
		newTxn := db.NewTransaction(true)
		value := &uint256.Uint256{}
		tmp, ok := new(big.Int).SetString("ffffffff", 16)
		if !ok {
			t.Fatal(err)
		}
		_, _ = value.FromBigInt(tmp)
		deposits = append(deposits, makeDeposit(t, signer, 1, int(i), value)) // created pre-image object
		txs = append(txs, makeTxs(t, signer, deposits[i]))
		txHash, err := txs[i].TxHash()
		if err != nil {
			t.Fatal(err)
		}
		smtTxHsh, _, err := PersistentTxHash(newTxn, txs[i])
		if err != nil {
			t.Fatal(err)
		}
		log.Printf("Tx hash (%d): %x", i, txHash)
		txHshLst = append(txHshLst, txHash)
		txHshSMTs = append(txHshSMTs, smtTxHsh)
	}

	newTxn := db.NewTransaction(true)

	var stateRoot []byte
	err = db.Update(func(txn *badger.Txn) error {
		stateRoot, err = hndlr.ApplyState(txn, txs, 1)
		if err != nil {
			t.Fatal(err)
		}
		log.Printf("stateRoot: %x\n", stateRoot)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Update(getAllStateMerkleProofs(hndlr, txs))
	if err != nil {
		t.Fatal(err)
	}

	// // USE TO GENERATE Merkle proofs with an arbitrary key against the stateTrie trie
	// utxoID, err := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000030")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// err = db.Update(getStateMerkleProofs(hndlr, txs, utxoID))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// /////////////////////////////////////////////

	// Generating block 1
	chain, err := GenerateBlock(nil, stateRoot, txHshLst)
	if err != nil {
		t.Fatal(err)
	}

	newTxn = db.NewTransaction(true)
	_, txRoot, err := MakePersistentTxRoot(newTxn, txHshLst) // generating the smt root
	log.Printf("The transaction Root persisted: %x\n", txRoot)
	if err != nil {
		t.Fatal(err)
	}

	// /////////////////////////////////////////////
	// // USE to generate an arbitrary merkle proof against a txRoot trie
	// transactionIncluded, err := hex.DecodeString("77338383edde4a7477f32549672ce24ff2800d7b61bf71cac09fab6e9b008495")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// bitmap, auditPath, proofHeight, included, proofKey, proofVal, err := smtTxRoot.MerkleProofCompressed(newTxn, transactionIncluded)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// _, err = CreateMerkleProof(
	// 	included,
	// 	proofHeight,
	// 	transactionIncluded,
	// 	proofKey,
	// 	proofVal,
	// 	bitmap,
	// 	auditPath,
	// )
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// /////////////////////////////////////////////

	////////// Block 2 /////////////
	// this is consuming utxo generated on block 1
	log.Println("\n\n\n ========== Block 2: =============")
	value := &uint256.Uint256{}
	_, _ = value.FromUint64(100000000)

	// ////////////////// Generating accusations against invalid transactions ///////
	// /////////////// Consuming a non existing UTXO
	// signerPubKey, err := signer.Pubkey()
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// nonExistingUTXO, err := createNewUTXO(1, value, signerPubKey, crypto.Hasher([]byte("lol")), 0)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// tx2 := makeTransfer(t, signer, signer2, 1, nonExistingUTXO)
	// /////////////////////////////////////////////////////////

	// ////////////// Consuming a deposit that was already spent
	// tx2 := makeTxs(t, signer, deposits[0])
	// //////////////////////////////////////////////////////////

	// /////////////// Consuming a deposit that doesn't exist before /////
	// // todo:
	// /////

	// // /////////////// Consuming a valid UTXO
	tx2 := makeTransfer(t, signer, signer2, 1, txs[0].Vout[0])
	// // //////////////////////////////////////////////////////////

	txHash2, err := tx2.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Tx hash: %x", txHash2)
	tx2UTXOID2, err := tx2.Vout.UTXOID()
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("TX2 OUT UTXOID: %x", tx2UTXOID2)
	txHashSMT, _, err := PersistentTxHash(newTxn, tx2)
	if err != nil {
		t.Fatal(err)
	}
	tx2UTXOID, err := tx2.Vin.UTXOID()
	log.Printf("Vin UTXO ID: %x\n", tx2UTXOID[0])
	if err != nil {
		t.Fatal(err)
	}

	chain, err = GenerateBlock(chain, stateRoot, [][]byte{txHash2})
	if err != nil {
		t.Fatal(err)
	}

	//////////// Generating cobjs.BClaims and PClaims ////////////////////

	bnVal := &crypto.BNGroupValidator{}
	if err != nil {
		t.Fatal(err)
	}
	bclaims := chain[0]
	bhsh, err := bclaims.BlockHash()
	if err != nil {
		t.Fatal(err)
	}
	gk := &crypto.BNGroupSigner{}
	gk.SetPrivk(crypto.Hasher([]byte("secret")))
	sig, err := gk.Sign(bhsh)
	if err != nil {
		t.Fatal(err)
	}
	bh := &cobjs.BlockHeader{
		BClaims:  bclaims,
		SigGroup: sig,
		TxHshLst: txHshLst,
	}
	err = bh.ValidateSignatures(bnVal)
	if err != nil {
		t.Fatal(err)
	}
	rcert, err := bh.GetRCert()
	if err != nil {
		t.Fatal(err)
	}
	err = rcert.ValidateSignature(bnVal)
	if err != nil {
		t.Fatal(err)
	}

	pclms := &cobjs.PClaims{
		BClaims: chain[1],
		RCert:   rcert,
	}

	pClaimsBin, err := pclms.MarshalBinary()
	log.Printf("PClaims Block 2:\n%x\n\n", pClaimsBin)
	prop := &cobjs.Proposal{
		PClaims:  pclms,
		TxHshLst: [][]byte{txHash2},
	}
	err = prop.Sign(signer)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Sig PClaims Block 2:\n%x\n\n", prop.Signature)
	bclaimsBin, err := chain[0].MarshalBinary()
	log.Printf("BClaim block 1:\n%x\n\n", bclaimsBin)
	log.Printf("SigGrup Block 1:\n%x\n\n", rcert.SigGroup)
	log.Printf("\n\n ======== Creating the Merkle proof for the StateTrie =======")
	err = db.Update(getStateMerkleProofs(hndlr, []*objs.Tx{tx2}, tx2UTXOID[0]))
	if err != nil {
		t.Fatal(err)
	}

	txRootSMT, txRoot2, err := MakePersistentTxRoot(db.NewTransaction(true), [][]byte{txHash2}) // generating the smt root
	log.Println(" ======== Creating the merkle proof for the TxHASHROOT =====")
	bitmap, auditPath, proofHeight, included, proofKey, proofVal, err := txRootSMT.MerkleProofCompressed(newTxn, txHash2)
	if err != nil {
		t.Fatal(err)
	}
	_, err = CreateMerkleProof(
		included,
		proofHeight,
		txHash2,
		proofKey,
		proofVal,
		bitmap,
		auditPath,
	)
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("\n\n ========Creating the merkle proof for the TxHASHSMT ==========")
	bitmap, auditPath, proofHeight, included, proofKey, proofVal, err = txHashSMT.MerkleProofCompressed(newTxn, tx2UTXOID[0])
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	_, err = CreateMerkleProof(
		included,
		proofHeight,
		tx2UTXOID[0],
		proofKey,
		proofVal,
		bitmap,
		auditPath,
	)
	if err != nil {
		t.Fatal(err)
	}
	preImageBin, err := tx2.Vin[0].TXInLinker.TXInPreImage.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("\n\n ======== TxIn Preimage hash: ======= \n%x\n\n", preImageBin)
	log.Println("=========")
	log.Printf("The transaction Root persisted Block 2: %x\n", txRoot2)
	if err != nil {
		t.Fatal(err)
	}
}
