package db

import (
	"bytes"
	"testing"

	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/dgraph-io/badger/v2"
)

type testParams struct {
	TxRoot     []byte
	StateRoot  []byte
	PrevBlock  []byte
	HeaderRoot []byte
	ChainID    uint32
	NumTx      int
	Height     uint32
	Round      uint32
}

func makeTestParams() *testParams {
	return &testParams{
		TxRoot:     make([]byte, constants.HashLen),
		StateRoot:  make([]byte, constants.HashLen),
		PrevBlock:  make([]byte, constants.HashLen),
		HeaderRoot: make([]byte, constants.HashLen),
		ChainID:    1,
		Round:      1,
		Height:     1,
	}
}

func createDatabase(t *testing.T) (*Database, *testParams) {
	t.Helper()
	dir := t.TempDir()
	opts := badger.DefaultOptions(dir)
	bdb, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := bdb.Close(); err != nil {
			t.Error(err)
		}
	})
	db := &Database{}
	db.Init(bdb)
	params := makeTestParams()
	return db, params
}

func TestRoundState(t *testing.T) {
	t.Parallel()
	groupSigner := &crypto.BNGroupSigner{}
	err := groupSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	groupKey, _ := groupSigner.PubkeyShare()

	bnSigner := &crypto.BNGroupSigner{}
	err = bnSigner.SetPrivk(crypto.Hasher([]byte("secret2")))
	if err != nil {
		t.Fatal(err)
	}
	bnKey, _ := groupSigner.PubkeyShare()

	secpSigner := &crypto.Secp256k1Signer{}
	err = secpSigner.SetPrivk(crypto.Hasher([]byte("secret3")))
	if err != nil {
		t.Fatal(err)
	}
	secpKey, _ := secpSigner.Pubkey()
	vAddr := crypto.GetAccount(secpKey)

	db, p := createDatabase(t)
	badgerD := db.rawDB.db
	err = badgerD.Update(func(txn *badger.Txn) error {
		sig, err := groupSigner.Sign(p.PrevBlock)
		if err != nil {
			t.Fatal(err)
		}
		rs := &objs.RoundState{
			VAddr:      vAddr, // change done
			GroupKey:   groupKey,
			GroupShare: bnKey,
			GroupIdx:   127,
			RCert: &objs.RCert{
				SigGroup: sig,
				RClaims: &objs.RClaims{
					ChainID:   p.ChainID,
					Height:    p.Height,
					PrevBlock: p.PrevBlock,
					Round:     p.Round,
				},
			},
		}
		err = db.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatal(err)
		}
		rs2, err := db.GetCurrentRoundState(txn, vAddr)
		if err != nil {
			t.Fatal(err)
		}
		if rs2.GroupIdx != rs.GroupIdx {
			t.Fatal("GroupIdx does not agree between RoundState rs and CurrentRoundState rs2!")
		}
		rs2, err = db.GetHistoricRoundState(txn, vAddr, 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if rs2.GroupIdx != rs.GroupIdx {
			t.Fatal("GroupIdx does not agree between RoundState rs and HistoricRoundState rs2!")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBlockHeader(t *testing.T) {
	t.Parallel()
	groupSigner := &crypto.BNGroupSigner{}
	err := groupSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}

	bnSigner := &crypto.BNGroupSigner{}
	err = bnSigner.SetPrivk(crypto.Hasher([]byte("secret2")))
	if err != nil {
		t.Fatal(err)
	}

	db, p := createDatabase(t)
	badgerD := db.rawDB.db
	err = badgerD.Update(func(txn *badger.Txn) error {
		sig, err := groupSigner.Sign(p.PrevBlock)
		if err != nil {
			t.Fatal(err)
		}
		rs := &objs.BlockHeader{
			SigGroup: sig,
			BClaims: &objs.BClaims{
				ChainID:    p.ChainID,
				Height:     p.Height,
				PrevBlock:  p.PrevBlock,
				HeaderRoot: p.HeaderRoot,
				StateRoot:  p.StateRoot,
				TxRoot:     p.TxRoot,
			},
		}
		err = db.SetCommittedBlockHeader(txn, rs)
		if err != nil {
			t.Fatal(err)
		}
		rs.BClaims.Height = 2
		err = db.SetCommittedBlockHeader(txn, rs)
		if err != nil {
			t.Fatal(err)
		}
		hdrRootBeforeDelete, err := db.GetHeaderRootForProposal(txn)
		if err != nil {
			t.Fatal(err)
		}
		rs.BClaims.Height = 3
		err = db.SetCommittedBlockHeader(txn, rs)
		if err != nil {
			t.Fatal(err)
		}
		rs.BClaims.Height = 1
		bhsh, err := rs.BlockHash()
		if err != nil {
			t.Fatal(err)
		}
		rs4, err := db.GetCommittedBlockHeaderByHash(txn, bhsh)
		if err != nil {
			t.Fatal(err)
		}
		if rs4.BClaims.Height != rs.BClaims.Height {
			t.Fatalf("BlockHeight does not match between BlockHeader rs and CommittedBlockHeader rs4! %d %d", rs.BClaims.Height, rs4.BClaims.Height)
		}
		rs2, err := db.GetCommittedBlockHeader(txn, 1)
		if err != nil {
			t.Fatal(err)
		}
		if rs2.BClaims.Height != rs.BClaims.Height {
			t.Fatalf("BlockHeight does not match between BlockHeader rs and CommittedBlockHeader rs2! %d %d", rs.BClaims.Height, rs2.BClaims.Height)
		}
		hdrRoot, err := db.GetHeaderRootForProposal(txn)
		if err != nil {
			t.Fatal(err)
		}
		rs3, prf, err := db.GetCommittedBlockHeaderWithProof(txn, hdrRoot, 3)
		if err != nil {
			t.Fatal(err)
		}
		ok, err := db.ValidateCommittedBlockHeaderWithProof(txn, hdrRoot, rs3, prf)
		if err != nil {
			t.Error(err)
		}
		if !ok {
			t.Fatal("failed to validate proof")
		}
		hdrRootPendingDelete, err := db.GetHeaderRootForProposal(txn)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Equal(hdrRootPendingDelete, hdrRootBeforeDelete) {
			t.Fatal("hdrRoot should be different!")
		}
		err = db.DeleteCommittedBlockHeader(txn, 3)
		if err != nil {
			t.Fatal("Failed to delete committed block header!")
		}
		_, _, err = db.GetCommittedBlockHeaderWithProof(txn, hdrRoot, 3)
		if err != badger.ErrKeyNotFound && err != nil {
			t.Fatal(err)
		}
		_, err = db.GetCommittedBlockHeader(txn, 3)
		if err != badger.ErrKeyNotFound && err != nil {
			t.Fatal(err)
		}
		rootHash, err := db.GetHeaderRootForProposal(txn)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				t.Fatal(err)
			}
			rootHash = nil
		}
		value, err := db.trie.Get(txn, rootHash, 3)
		if err != badger.ErrKeyNotFound && err != nil {
			t.Fatal(err)
		}
		if len(value) > 0 {
			t.Fatal("Should be an empty slice!")
		}
		hdrRootAfterDelete, err := db.GetHeaderRootForProposal(txn)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(hdrRootAfterDelete, hdrRootBeforeDelete) {
			t.Fatal("hdr root mismatch")
		}
		rt, err := db.GetHeaderTrieRoot(txn, 2)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(rt, hdrRootBeforeDelete) {
			t.Fatal("hdr root mismatch does not match before delete")
		}
		rt3, err := db.GetHeaderTrieRoot(txn, 3)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(hdrRootAfterDelete, rt3) {
			t.Fatal("hdr root mismatch does not match after delete")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEncryptedStore(t *testing.T) {
	t.Parallel()
	groupSigner := &crypto.BNGroupSigner{}
	err := groupSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}

	bnSigner := &crypto.BNGroupSigner{}
	err = bnSigner.SetPrivk(crypto.Hasher([]byte("secret2")))
	if err != nil {
		t.Fatal(err)
	}

	db, _ := createDatabase(t)
	badgerD := db.rawDB
	err = badgerD.Update(func(txn *badger.Txn) error {
		name := []byte("foo")
		ec := &objs.EncryptedStore{
			Name: name,
		}
		err := db.SetEncryptedStore(txn, ec)
		if err != nil {
			t.Fatal(err)
		}
		ec2, err := db.GetEncryptedStore(txn, name)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(ec2.Name, ec.Name) {
			t.Fatal("name mismatch: did not unmarshal correctly")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestValidatorSet(t *testing.T) {
	t.Parallel()
	groupSigner := &crypto.BNGroupSigner{}
	err := groupSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	groupKey, _ := groupSigner.PubkeyShare()

	bnSigner := &crypto.BNGroupSigner{}
	err = bnSigner.SetPrivk(crypto.Hasher([]byte("secret2")))
	if err != nil {
		t.Fatal(err)
	}

	db, _ := createDatabase(t)
	badgerD := db.rawDB
	err = badgerD.Update(func(txn *badger.Txn) error {
		vkey0 := crypto.Hasher([]byte("s0"))[12:]
		gShare0 := crypto.Hasher([]byte("g0"))
		val0 := &objs.Validator{
			VAddr:      vkey0,
			GroupShare: gShare0,
		}
		vkey1 := crypto.Hasher([]byte("s1"))[12:]
		gShare1 := crypto.Hasher([]byte("g1"))
		val1 := &objs.Validator{
			VAddr:      vkey1,
			GroupShare: gShare1,
		}
		vkey2 := crypto.Hasher([]byte("s2"))[12:]
		gShare2 := crypto.Hasher([]byte("g2"))
		val2 := &objs.Validator{
			VAddr:      vkey2,
			GroupShare: gShare2,
		}
		vkey3 := crypto.Hasher([]byte("s3"))[12:]
		gShare3 := crypto.Hasher([]byte("g3"))
		val3 := &objs.Validator{
			VAddr:      vkey3,
			GroupShare: gShare3,
		}

		notBefore := uint32(1)
		vSet := &objs.ValidatorSet{
			Validators: []*objs.Validator{val0, val1, val2, val3},
			GroupKey:   groupKey,
			NotBefore:  notBefore,
		}

		vSetBytes, err := vSet.MarshalBinary()
		if err != nil {
			t.Fatal("Error when marshalling vSet!")
		}

		err = db.SetValidatorSet(txn, vSet)
		if err != nil {
			t.Fatal("Error in SetValidatorSet")
		}

		height := uint32(1)
		vSetTest, err := db.GetValidatorSet(txn, height)
		if err != nil {
			t.Error(err)
			t.Fatal("Error in GetValidatorSet")
		}
		vSetTestBytes, err := vSetTest.MarshalBinary()
		if err != nil {
			t.Fatal("Error when marshalling vSetTest!")
		}

		if !bytes.Equal(vSetBytes, vSetTestBytes) {
			t.Fatal("vSetBytes and vSetTestBytes are not equal!")
		}

		_, err = db.makeValidatorSetKey(notBefore)
		if err != nil {
			t.Fatal("Error in makeValidatorSetKey!")
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSnapShotMany(t *testing.T) {
	t.Parallel()
	groupSigner := &crypto.BNGroupSigner{}
	err := groupSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}

	bnSigner := &crypto.BNGroupSigner{}
	err = bnSigner.SetPrivk(crypto.Hasher([]byte("secret2")))
	if err != nil {
		t.Fatal(err)
	}

	db, p := createDatabase(t)
	badgerD := db.rawDB
	var bhash []byte
	err = badgerD.Update(func(txn *badger.Txn) error {
		sig, err := groupSigner.Sign(p.PrevBlock)
		if err != nil {
			t.Fatal(err)
		}
		for i := uint32(1); i < constants.EpochLength+1; i++ {
			bh := &objs.BlockHeader{
				SigGroup: sig,
				BClaims: &objs.BClaims{
					ChainID:    p.ChainID,
					Height:     i,
					PrevBlock:  p.PrevBlock,
					HeaderRoot: p.HeaderRoot,
					StateRoot:  p.StateRoot,
					TxRoot:     p.TxRoot,
				},
			}
			tmpBhash, err := bh.BlockHash()
			if err != nil {
				t.Fatal(err)
			}
			bhash = tmpBhash
			if i == 1 || i%constants.EpochLength == 0 {
				err = db.SetSnapshotBlockHeader(txn, bh)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = badgerD.Update(func(txn *badger.Txn) error {
		newbh, err := db.GetLastSnapshot(txn)
		if err != nil {
			t.Fatal(err)
		}
		newbhash, err := newbh.BlockHash()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(newbhash, bhash) {
			t.Fatal("hashes not equal")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSnapShotOne(t *testing.T) {
	t.Parallel()
	groupSigner := &crypto.BNGroupSigner{}
	err := groupSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}

	bnSigner := &crypto.BNGroupSigner{}
	err = bnSigner.SetPrivk(crypto.Hasher([]byte("secret2")))
	if err != nil {
		t.Fatal(err)
	}

	db, p := createDatabase(t)
	badgerD := db.rawDB
	var bhash []byte
	err = badgerD.Update(func(txn *badger.Txn) error {
		sig, err := groupSigner.Sign(p.PrevBlock)
		if err != nil {
			t.Fatal(err)
		}
		bh := &objs.BlockHeader{
			SigGroup: sig,
			BClaims: &objs.BClaims{
				ChainID:    p.ChainID,
				Height:     p.Height,
				PrevBlock:  p.PrevBlock,
				HeaderRoot: p.HeaderRoot,
				StateRoot:  p.StateRoot,
				TxRoot:     p.TxRoot,
			},
		}
		tmpBhash, err := bh.BlockHash()
		if err != nil {
			t.Fatal(err)
		}
		bhash = tmpBhash
		err = db.SetSnapshotBlockHeader(txn, bh)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = badgerD.Update(func(txn *badger.Txn) error {
		newbh, err := db.GetLastSnapshot(txn)
		if err != nil {
			t.Fatal(err)
		}
		newbhash, err := newbh.BlockHash()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(newbhash, bhash) {
			t.Fatal("hashes not equal")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetLastCommittedBHFSMany(t *testing.T) {
	t.Parallel()
	groupSigner := &crypto.BNGroupSigner{}
	err := groupSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}

	bnSigner := &crypto.BNGroupSigner{}
	err = bnSigner.SetPrivk(crypto.Hasher([]byte("secret2")))
	if err != nil {
		t.Fatal(err)
	}

	db, p := createDatabase(t)
	badgerD := db.rawDB
	var bhash []byte
	err = badgerD.Update(func(txn *badger.Txn) error {
		sig, err := groupSigner.Sign(p.PrevBlock)
		if err != nil {
			t.Fatal(err)
		}
		for i := uint32(1); i < constants.EpochLength+1; i++ {
			bh := &objs.BlockHeader{
				SigGroup: sig,
				BClaims: &objs.BClaims{
					ChainID:    p.ChainID,
					Height:     i,
					PrevBlock:  p.PrevBlock,
					HeaderRoot: p.HeaderRoot,
					StateRoot:  p.StateRoot,
					TxRoot:     p.TxRoot,
				},
			}
			tmpBhash, err := bh.BlockHash()
			if err != nil {
				t.Fatal(err)
			}
			bhash = tmpBhash
			err = db.SetCommittedBlockHeaderFastSync(txn, bh)
			if err != nil {
				t.Fatal(err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = badgerD.Update(func(txn *badger.Txn) error {
		newbh, err := db.GetMostRecentCommittedBlockHeaderFastSync(txn)
		if err != nil {
			t.Fatal(err)
		}
		newbhash, err := newbh.BlockHash()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(newbhash, bhash) {
			t.Fatal("hashes not equal")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetLastCommittedBHFSOne(t *testing.T) {
	t.Parallel()
	groupSigner := &crypto.BNGroupSigner{}
	err := groupSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}

	bnSigner := &crypto.BNGroupSigner{}
	err = bnSigner.SetPrivk(crypto.Hasher([]byte("secret2")))
	if err != nil {
		t.Fatal(err)
	}

	db, p := createDatabase(t)
	badgerD := db.rawDB
	var bhash []byte
	err = badgerD.Update(func(txn *badger.Txn) error {
		sig, err := groupSigner.Sign(p.PrevBlock)
		if err != nil {
			t.Fatal(err)
		}
		bh := &objs.BlockHeader{
			SigGroup: sig,
			BClaims: &objs.BClaims{
				ChainID:    p.ChainID,
				Height:     p.Height,
				PrevBlock:  p.PrevBlock,
				HeaderRoot: p.HeaderRoot,
				StateRoot:  p.StateRoot,
				TxRoot:     p.TxRoot,
			},
		}
		tmpBhash, err := bh.BlockHash()
		if err != nil {
			t.Fatal(err)
		}
		bhash = tmpBhash
		err = db.SetCommittedBlockHeaderFastSync(txn, bh)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = badgerD.Update(func(txn *badger.Txn) error {
		newbh, err := db.GetMostRecentCommittedBlockHeaderFastSync(txn)
		if err != nil {
			t.Fatal(err)
		}
		newbhash, err := newbh.BlockHash()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(newbhash, bhash) {
			t.Fatal("hashes not equal")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func getValidatorSet(t *testing.T, height uint32, seed string) (*objs.ValidatorSet, error) {
	groupSigner := &crypto.BNGroupSigner{}
	err := groupSigner.SetPrivk(crypto.Hasher([]byte(seed)))
	if err != nil {
		t.Fatal(err)
	}
	groupKey, _ := groupSigner.PubkeyShare()

	bnSigner := &crypto.BNGroupSigner{}
	err = bnSigner.SetPrivk(crypto.Hasher([]byte(seed + "2")))
	if err != nil {
		t.Fatal(err)
	}
	vkey0 := crypto.Hasher([]byte("s0"))[12:]
	gShare0 := crypto.Hasher([]byte("g0"))
	val0 := &objs.Validator{
		VAddr:      vkey0,
		GroupShare: gShare0,
	}
	vkey1 := crypto.Hasher([]byte("s1"))[12:]
	gShare1 := crypto.Hasher([]byte("g1"))
	val1 := &objs.Validator{
		VAddr:      vkey1,
		GroupShare: gShare1,
	}
	vkey2 := crypto.Hasher([]byte("s2"))[12:]
	gShare2 := crypto.Hasher([]byte("g2"))
	val2 := &objs.Validator{
		VAddr:      vkey2,
		GroupShare: gShare2,
	}
	vkey3 := crypto.Hasher([]byte("s3"))[12:]
	gShare3 := crypto.Hasher([]byte("g3"))
	val3 := &objs.Validator{
		VAddr:      vkey3,
		GroupShare: gShare3,
	}

	notBefore := uint32(height)
	vSet := &objs.ValidatorSet{
		Validators: []*objs.Validator{val0, val1, val2, val3},
		GroupKey:   groupKey,
		NotBefore:  notBefore,
	}
	return vSet, nil
}

func TestValidatorSet2(t *testing.T) {
	t.Parallel()
	db, _ := createDatabase(t)
	badgerD := db.rawDB
	err := badgerD.Update(func(txn *badger.Txn) error {
		vSet1, err := getValidatorSet(t, 1, "v1")
		if err != nil {
			t.Fatal("Error in Vset")
		}
		vSet2, err := getValidatorSet(t, 15, "v15")
		if err != nil {
			t.Fatal("Error in Vset")
		}
		err = db.SetValidatorSet(txn, vSet1)
		if err != nil {
			t.Fatal("Error in SetValidatorSet1")
		}
		err = db.SetValidatorSet(txn, vSet2)
		if err != nil {
			t.Fatal("Error in SetValidatorSet2")
		}

		hList := []uint32{12, 13, 14, 15, 16, 17}
		rList := []*objs.ValidatorSet{vSet1, vSet1, vSet1, vSet2, vSet2, vSet2}
		for index, h := range hList {
			r := rList[index]
			vs, err := db.GetValidatorSet(txn, h)
			if err != nil {
				t.Fatal(err)
			}
			ok, err := compareObject(r, vs)
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Fatalf("VSet %x doesn't match db entry for height: %d", r.GroupKey, h)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

type mb interface {
	MarshalBinary() ([]byte, error)
}

func compareObject(a mb, b mb) (bool, error) {
	aBin, err := a.MarshalBinary()
	if err != nil {
		return false, err
	}
	bBin, err := b.MarshalBinary()
	if err != nil {
		return false, err
	}
	return bytes.Equal(aBin, bBin), nil
}
