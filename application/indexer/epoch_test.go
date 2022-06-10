package indexer

import (
	"bytes"
	"testing"

	trie "github.com/MadBase/MadNet/badgerTrie"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/internal/testing/environment"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
)

func makeEpochConstrainedList() *EpochConstrainedList {
	prefix1 := func() []byte {
		return []byte("aa")
	}
	prefix2 := func() []byte {
		return []byte("ab")
	}
	index := NewEpochConstrainedIndex(prefix1, prefix2)
	return index
}

func TestEpochConstrainedListAppend(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	ecl := makeEpochConstrainedList()
	txHash := trie.Hasher([]byte("txHash"))
	epoch := uint32(1)
	err := db.Update(func(txn *badger.Txn) error {
		err := ecl.Append(txn, epoch, txHash)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEpochConstrainedListDrop(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	ecl := makeEpochConstrainedList()
	txHash := trie.Hasher([]byte("txHash"))
	epoch := uint32(1)
	err := db.Update(func(txn *badger.Txn) error {
		err := ecl.Drop(txn, txHash)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		err = ecl.Append(txn, epoch, txHash)
		if err != nil {
			t.Fatal(err)
		}
		err = ecl.Drop(txn, txHash)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEpochConstrainedListGetEpoch(t *testing.T) {
	t.Parallel()
	db := environment.SetupBadgerDatabase(t)

	ecl := makeEpochConstrainedList()
	txHash := trie.Hasher([]byte("txHash"))
	epoch := uint32(1)
	err := db.Update(func(txn *badger.Txn) error {
		_, err := ecl.GetEpoch(txn, txHash)
		if err == nil {
			t.Fatal("Should have raised error")
		}
		err = ecl.Append(txn, epoch, txHash)
		if err != nil {
			t.Fatal(err)
		}
		returnedEpoch, err := ecl.GetEpoch(txn, txHash)
		if err != nil {
			t.Fatal(err)
		}
		if epoch != returnedEpoch {
			t.Fatal("epochs do not match")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEpochConstrainedListMakeListKey(t *testing.T) {
	t.Parallel()

	ecl := makeEpochConstrainedList()
	txHash := trie.Hasher([]byte("txHash"))
	epoch := uint32(1)
	trueKey := []byte{}
	trueKey = append(trueKey, ecl.prefix()...)
	trueKey = append(trueKey, utils.MarshalUint32(epoch)...)
	trueKey = append(trueKey, txHash...)
	eclKey := ecl.makeKey(epoch, txHash)
	key := eclKey.MarshalBinary()
	if !bytes.Equal(key, trueKey) {
		t.Fatal("keys do not match")
	}
}

func TestEpochConstrainedListMakeListRefKey(t *testing.T) {
	t.Parallel()

	ecl := makeEpochConstrainedList()
	txHash := trie.Hasher([]byte("txHash"))
	trueKey := []byte{}
	trueKey = append(trueKey, ecl.refPrefix()...)
	trueKey = append(trueKey, txHash...)
	eclKey := ecl.makeRefKey(txHash)
	key := eclKey.MarshalBinary()
	if !bytes.Equal(key, trueKey) {
		t.Fatal("keys do not match")
	}
}

func TestEpochConstrainedListKey(t *testing.T) {
	eclKey := &EpochConstrainedListKey{}
	keyTrue := crypto.Hasher([]byte("key"))
	eclKey.key = keyTrue
	key := eclKey.MarshalBinary()
	if !bytes.Equal(key, keyTrue) {
		t.Fatal("keys do not match (1)")
	}
	eclKey2 := &EpochConstrainedListKey{}
	eclKey2.UnmarshalBinary(key)
	if !bytes.Equal(eclKey.key, eclKey2.key) {
		t.Fatal("keys do not match (2)")
	}

	eclRefkey := &EpochConstrainedListRefKey{}
	refkeyTrue := crypto.Hasher([]byte("refkey"))
	eclRefkey.refkey = refkeyTrue
	refkey := eclRefkey.MarshalBinary()
	if !bytes.Equal(refkey, refkeyTrue) {
		t.Fatal("refkeys do not match (1)")
	}
	eclRefkey2 := &EpochConstrainedListRefKey{}
	eclRefkey2.UnmarshalBinary(refkey)
	if !bytes.Equal(eclRefkey.refkey, eclRefkey2.refkey) {
		t.Fatal("refkeys do not match (2)")
	}
}
