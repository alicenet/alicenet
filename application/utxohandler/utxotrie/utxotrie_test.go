package utxotrie

import (
	"errors"
	"fmt"
	"github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

/*
import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/alicenet/alicenet/crypto"
	"github.com/dgraph-io/badger/v2"
)

func TestUTXOTrie(t *testing.T) {
	height := uint32(1)
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
	//////////////////////////////////////////////////////////////////////////////
	//////////////////////////////////////////////////////////////////////////////
	hndlr := NewUTXOTrie(db)

	hndlr.Init(1)
	ok, err := hndlr.Contains(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("did not fail")
	}
	db.Update(func(txn *badger.Txn) error {
		utxoID1 := crypto.Hasher([]byte("utxoID1"))
		utxoHash1 := crypto.Hasher([]byte("utxoHash1"))

		utxoID2 := crypto.Hasher([]byte("utxoID2"))
		utxoHash2 := crypto.Hasher([]byte("utxoHash2"))

		utxoID3 := crypto.Hasher([]byte("utxoID3"))
		utxoHash3 := crypto.Hasher([]byte("utxoHash3"))

		utxoID4 := crypto.Hasher([]byte("utxoID4"))
		utxoHash4 := crypto.Hasher([]byte("utxoHash4"))

		utxoID5 := crypto.Hasher([]byte("utxoID5"))
		utxoHash5 := crypto.Hasher([]byte("utxoHash5"))

		utxoID6 := crypto.Hasher([]byte("utxoID6"))
		utxoHash6 := crypto.Hasher([]byte("utxoHash6"))

		newUTXOIDs := [][]byte{}
		newUTXOIDs = append(newUTXOIDs, utxoID1)
		newUTXOIDs = append(newUTXOIDs, utxoID2)
		newUTXOIDs = append(newUTXOIDs, utxoID3)
		newUTXOHashes := [][]byte{}
		newUTXOHashes = append(newUTXOHashes, utxoHash1)
		newUTXOHashes = append(newUTXOHashes, utxoHash2)
		newUTXOHashes = append(newUTXOHashes, utxoHash3)
		stateRootProposal, err := hndlr.GetStateRootForProposal(txn, newUTXOIDs, newUTXOHashes, [][]byte{})
		if err != nil {
			t.Fatal(err)
		}
		stateRoot, err := hndlr.ApplyState(txn, newUTXOIDs, newUTXOHashes, [][]byte{}, height)
		if err != nil {
			t.Fatal(err)
		}
		height++
		if !bytes.Equal(stateRoot, stateRootProposal) {
			t.Fatal("roots not equal")
		}
		ok, err = hndlr.Contains(nil, utxoID1)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("missing utxo")
		}
		ok, err = hndlr.Contains(nil, utxoID2)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("missing utxo")
		}
		ok, err = hndlr.Contains(nil, utxoID3)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("missing utxo")
		}
		newUTXOIDs = [][]byte{}
		newUTXOIDs = append(newUTXOIDs, utxoID4)
		newUTXOIDs = append(newUTXOIDs, utxoID5)
		newUTXOIDs = append(newUTXOIDs, utxoID6)
		newUTXOHashes = [][]byte{}
		newUTXOHashes = append(newUTXOHashes, utxoHash4)
		newUTXOHashes = append(newUTXOHashes, utxoHash5)
		newUTXOHashes = append(newUTXOHashes, utxoHash6)
		consumedUTXOIDs := [][]byte{}
		consumedUTXOIDs = append(consumedUTXOIDs, utxoID1)
		consumedUTXOIDs = append(consumedUTXOIDs, utxoID2)
		consumedUTXOIDs = append(consumedUTXOIDs, utxoID3)
		stateRootProposal, err = hndlr.GetStateRootForProposal(txn, newUTXOIDs, newUTXOHashes, consumedUTXOIDs)
		if err != nil {
			t.Fatal(err)
		}
		stateRoot, err = hndlr.ApplyState(txn, newUTXOIDs, newUTXOHashes, consumedUTXOIDs, height)
		if err != nil {
			t.Fatal(err)
		}
		height++
		if !bytes.Equal(stateRoot, stateRootProposal) {
			t.Fatal("roots not equal")
		}
		ok, err = hndlr.Contains(nil, utxoID1)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatalf("did not fail")
		}
		ok, err = hndlr.Contains(nil, utxoID2)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatalf("did not fail")
		}
		ok, err = hndlr.Contains(nil, utxoID3)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatalf("did not fail")
		}
		ok, err = hndlr.Contains(nil, utxoID4)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("missing utxo")
		}
		ok, err = hndlr.Contains(nil, utxoID5)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("missing utxo")
		}
		ok, err = hndlr.Contains(nil, utxoID6)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("missing utxo")
		}
		return nil
	})

}
*/

func TestUTXOTrie_ApplyState_CorruptedTrie(t *testing.T) {
	db := mocks.NewTestDB()
	trie := NewUTXOTrie(db.DB())

	for i := 1; i < 24; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(i))
			return err
		})
		require.Nil(t, err)
	}

	// Height 24, tx 0
	tx := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 21),
	}

	tx.Vin[0] = createVinTx(
		t,
		"01011b8f4e22353081e3870bde9171a6fd62924f6221fcb4cba1941896214fb210d92855a64346ca14a8bc9b343b649da07dd38dd94c3881ad305f3af91372ef5ea901",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		"0000000000000000000000000000000000000000000000000000000000000002",
		4294967295,
	)

	tx.Vout[0] = createValueStoreVoutTx(t, "82a978b3f5962a5b0957d9ee9eef472ee55b42f1", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(0), 1)
	tx.Vout[1] = createValueStoreVoutTx(t, "ba3a40133ff3b69424eec55ea54b94650f4b68eb", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(1), 2)
	tx.Vout[2] = createValueStoreVoutTx(t, "dceceaf3fc5c0a63d195d69b1a90011b7b19650d", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(2), 1)
	tx.Vout[3] = createValueStoreVoutTx(t, "d8052e80236e5e3f2372152d9d85951ee4fd8a8d", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(3), 2)
	tx.Vout[4] = createValueStoreVoutTx(t, "13cbb8d99c6c4e0f2728c7d72606e78a29c4e224", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(4), 1)
	tx.Vout[5] = createValueStoreVoutTx(t, "4602d163042e3e516bfbf73b59f7977f16a7e80c", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(5), 2)
	tx.Vout[6] = createValueStoreVoutTx(t, "24143873e0e0815fdcbcffdbe09c979cbf9ad013", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(6), 1)
	tx.Vout[7] = createValueStoreVoutTx(t, "caa569cb54e512837b6b5dfb1b7675eea57b1412", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(7), 2)
	tx.Vout[8] = createValueStoreVoutTx(t, "e0fc04fa2d34a66b779fd5cee748268032a146c0", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(8), 1)
	tx.Vout[9] = createValueStoreVoutTx(t, "f290c40531551726e0fa9cb66ec3f99fc5ebec81", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(9), 2)
	tx.Vout[10] = createValueStoreVoutTx(t, "1817465453315a39d62de436e8ae8134e4d9c2cd", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(10), 1)
	tx.Vout[11] = createValueStoreVoutTx(t, "572008ad1b49aa8d0d7a988a2b220560eeb3db97", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(11), 2)
	tx.Vout[12] = createValueStoreVoutTx(t, "7aa22a1a0672a54a819659dffc1c4ca5383114dc", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(12), 1)
	tx.Vout[13] = createValueStoreVoutTx(t, "96d4a79174d52d646ce4b80c24f7f181062c4395", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(13), 2)
	tx.Vout[14] = createValueStoreVoutTx(t, "00c40fe2095423509b9fd9b754323158af2310f3", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(14), 1)
	tx.Vout[15] = createValueStoreVoutTx(t, "fae29ec318a32ed8c2106da0e900baf9ef997ec3", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(15), 2)
	tx.Vout[16] = createValueStoreVoutTx(t, "8643d6c2e40f1a3a8d77a4021397e9f0dfba3eaf", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(16), 1)
	tx.Vout[17] = createValueStoreVoutTx(t, "3b7c2dd0756474c9297afb6eec9fab54306b1c78", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(17), 2)
	tx.Vout[18] = createValueStoreVoutTx(t, "1427f29745e5b8ebe70afb6646e73fd194da7a7a", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(18), 1)
	tx.Vout[19] = createValueStoreVoutTx(t, "702a6fee5b4be6aaec22b4c4fd0b7902c6cb1dfe", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "5000000000000000000", uint32(19), 2)
	tx.Vout[20] = createValueStoreVoutTx(t, "546f99f244b7b58b855330ae0e2bc1b30b41302f", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", "999900000000000000000000", uint32(20), 1)
	err := tx.SetTxHash()
	require.Nil(t, err)

	err = db.Update(func(txn *badger.Txn) error {
		_, err := trie.ApplyState(txn, objs.TxVec{tx}, 24)
		return err
	})
	require.Nil(t, err)

	for i := 25; i < 29; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(i))
			return err
		})
		require.Nil(t, err)
	}

	// Height 29, tx 0
	tx_29_0 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_29_0.Vin[0] = createVinTx(
		t,
		"01013414296a12e4f11d9b9f7ab76aa888287a17ab66ae77ddcbc90b9cfc1cb3514a3499566524fb751e465d58b02523c5bcd56dd163edcb54d87fe3846ad7af605601",
		"4bad7dcca377a058376005a612d45194137a1b8e77bee22aa89284434159600e",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		6,
	)
	tx_29_0.Vout[0] = createDataStoreVoutTx(
		t,
		"24143873e0e0815fdcbcffdbe09c979cbf9ad013",
		"4bad7dcca377a058376005a612d45194137a1b8e77bee22aa89284434159600e",
		"bdf9ccc3aa2dea9a3fd5f8f48bb374faf48be0c23bec177bff587894855d8106",
		"575d6a69425b306f3158",
		"3088",
		1,
		0,
		1,
		"7e24a8e08d42497f5ddcc318416738bb86e5286347107e644f114aa76c8cadbe2199609e334fb22c27f64d52c82e74806c9507e64984d615a416f1236d5d526e00",
	)
	tx_29_0.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"4bad7dcca377a058376005a612d45194137a1b8e77bee22aa89284434159600e",
		"4999999999999996912",
		1,
		1,
	)
	err = tx_29_0.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 1
	tx_29_1 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_29_1.Vin[0] = createVinTx(
		t,
		"0102068412c7ac2e92d81ef48080b197cf39512a6951e21a94a63652a5e63239873f3031eefbdd4afb127c48763fd57cd3b98c6d260bb3d46fc2bf0f85d125d3199b02933cc3f7ab805720067ca9d1efff7b2af7f75a683dab8d7777f8d3fdf44bdd27b288a5d0af37e53d0b3ba5e21a0cb7d4f66c3999c4dfd8dee524260af45ef30546576dca5ff20f94f087bc8371e6b3bd26fde4daf04d0f2a998997a7b718be15c4762742835cc96825a6116417783694df9db73c3baf9d092b8af4d51b7b99",
		"f020c514060c7b969cc9d457ddb8fb57b1aeda32d9ba89bf8e2273e10786e12a",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		1,
	)
	tx_29_1.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"f020c514060c7b969cc9d457ddb8fb57b1aeda32d9ba89bf8e2273e10786e12a",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_29_1.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 2
	tx_29_2 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_29_2.Vin[0] = createVinTx(
		t,
		"0102217427e194b5656db76a0d4a4790999dcbe2a27cacc53990dfabeadda45ce1df1ca27dab28fe2e6bbf7f4d903ce9760342993fdd4422e5a7603addd7aef7cd852c71a73bd6cebe0c852c5830c645e0f36e81b7b6dea12116df8e2fdafa657e3f03733a05b8e32d68d690aa2842f116fa5bcb9b2b7290b3e7b4e0af1043e9d29b079ca24b37435d7d0af062834d54887f1a12ea9ca1e09bfb65ea381a109fb00e0613ee387dbd69fc6f13290e486d6d660176f88e4a95d090cf0601abbb9e9b9d",
		"5ac7cc937e43623d211526994a622f2fb8307a474d9d92ef3919e8a5069527f2",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		13,
	)
	tx_29_2.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"5ac7cc937e43623d211526994a622f2fb8307a474d9d92ef3919e8a5069527f2",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_29_2.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 3
	tx_29_3 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_29_3.Vin[0] = createVinTx(
		t,
		"0101944dc7f2c60d4336a0a89d8c7f2922fa28dc6800ae696c42797612fc2364dac269893cf7ec5ddfb8ba8355beb68b44b9c1e657d94422de259bb464cd07e3c83101",
		"89f90ba0a43e50de81acfa7785601e54b56723bb8a731a5543709d8824304c7a",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		14,
	)
	tx_29_3.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"89f90ba0a43e50de81acfa7785601e54b56723bb8a731a5543709d8824304c7a",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_29_3.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 4
	tx_29_4 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_29_4.Vin[0] = createVinTx(
		t,
		"01022a48801c0c33b044cd2e3d6880f850d63e351c4978e617c3705f0c28bac6a09521cffe41c0bc9b9f0ccd8f7a118fb9aaf67f59ebb6005eefe3841378e6a1c1f4225ae2490d11202be589c6ab8f0533d46bc0c1ee067ce7996636b39ee76860812ecd1b828102a864a8c6f39942476385b89699a4b29b25559d13e910fcdc591d255b408b5d52dfefc365d5c4b22a821781d70f13617b4a143b309f3f9a42f18229cd47ccb96dc24a1ad10e0cc0ef72d83cc4c48f49ce19d07cd5e6cd025bf403",
		"15406f892afdb0e0d051759293e788a1215bc96b45a96205372f843ee0de9b55",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		3,
	)
	tx_29_4.Vout[0] = createDataStoreVoutTx(
		t,
		"d8052e80236e5e3f2372152d9d85951ee4fd8a8d",
		"15406f892afdb0e0d051759293e788a1215bc96b45a96205372f843ee0de9b55",
		"6e56d0c2f63a2a1234ef0ffcf06db83362118c9b3e17ed9c95fdbbaed4bfdad1",
		"567544414155532e58533165",
		"1552",
		1,
		0,
		2,
		"2a48801c0c33b044cd2e3d6880f850d63e351c4978e617c3705f0c28bac6a09521cffe41c0bc9b9f0ccd8f7a118fb9aaf67f59ebb6005eefe3841378e6a1c1f4225ae2490d11202be589c6ab8f0533d46bc0c1ee067ce7996636b39ee76860812ecd1b828102a864a8c6f39942476385b89699a4b29b25559d13e910fcdc591d1d4bf83ec551d4e3512a8e0917c7b96695df87dcbbf84d08fe56d0e04e2b667922f7c73fc576455b0729502c4abdad5e61aa924c18386d29bc9a59e47e8269cd",
	)
	tx_29_4.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"15406f892afdb0e0d051759293e788a1215bc96b45a96205372f843ee0de9b55",
		"4999999999999998448",
		1,
		1,
	)
	err = tx_29_4.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 5
	tx_29_5 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_29_5.Vin[0] = createVinTx(
		t,
		"010114d8544c0d748ac1796b6b837a93c45bebf2f8d45a2d0d7482a65571fffb880b1983269fac9eaf4c42b1911579986ee288acc0c8a33401693cb15a62c27a36c301",
		"10dbd0765d9da5e0d61791574329ff04d4fadabf749d2550b6b5a3a906180b31",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		16,
	)
	tx_29_5.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"10dbd0765d9da5e0d61791574329ff04d4fadabf749d2550b6b5a3a906180b31",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_29_5.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 6
	tx_29_6 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_29_6.Vin[0] = createVinTx(
		t,
		"0101502fd41a0672b85c601b92dbfb86298fe1b1939aa3c9f12b06291684b09da81559b0195473bc125c9692bb3d30ad2e36a547e245edc3d7a3e92779b8066607c901",
		"962dbd4f1c5a1338697987e54c0467d80d87999f7d4fdadc95c382c4ea37abbf",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		4,
	)
	tx_29_6.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"962dbd4f1c5a1338697987e54c0467d80d87999f7d4fdadc95c382c4ea37abbf",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_29_6.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 7
	tx_29_7 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_29_7.Vin[0] = createVinTx(
		t,
		"01021b350a8792909849017b6eec15b76a5820445ca6c93fa2b6a4be8ddcec4d5b3d2c24f3beae1001d740b26a2f9b9128e2a106b14c25344c27b23f8034bbc51809117e3fdc4e9d757a407042bd071921774be346d3096e5386ee0be154edd6a9322c0fc505a2e7669101e8c349c8cf93c74f0356fbe69d307c82eb7b0dd65400340e2f4547f82a10fb4a8ba72c885ae3160db9bad0fed8eb5632d4521467e8bec60965fd066dacfb78dbb604315959a5e8546db891764a63818a288a92573b2caa",
		"9bd0ccecde5be25f12df135c4f81c63ac0fc314be77017c7e7efbadc6a00cd5f",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		15,
	)
	tx_29_7.Vout[0] = createDataStoreVoutTx(
		t,
		"fae29ec318a32ed8c2106da0e900baf9ef997ec3",
		"9bd0ccecde5be25f12df135c4f81c63ac0fc314be77017c7e7efbadc6a00cd5f",
		"a5f2c2ed2c61195fe106ed18eeafacadab5f1ed94514f675ef937bce163f480a",
		"697a733041457c2e4958495a6c4e442f5067",
		"4334",
		1,
		0,
		2,
		"1b350a8792909849017b6eec15b76a5820445ca6c93fa2b6a4be8ddcec4d5b3d2c24f3beae1001d740b26a2f9b9128e2a106b14c25344c27b23f8034bbc51809117e3fdc4e9d757a407042bd071921774be346d3096e5386ee0be154edd6a9322c0fc505a2e7669101e8c349c8cf93c74f0356fbe69d307c82eb7b0dd6540034295de70c527b160d0e0b0decf3171c2de4c4363c66f8ca9f3b405ce8973c45b310e407dd7a7d3c434da23b79c2f570c80d101b5220e6a67a37636a2b70723a1e",
	)
	tx_29_7.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"9bd0ccecde5be25f12df135c4f81c63ac0fc314be77017c7e7efbadc6a00cd5f",
		"4999999999999995666",
		1,
		1,
	)
	err = tx_29_7.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 8
	tx_29_8 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_29_8.Vin[0] = createVinTx(
		t,
		"0102163bea1e569438743fcf7572e939ab08d9d8899cd05a384699a828183bb5cc6c00bb82f05869536a9823e8f92fdd2b673c74fbf44ace82b887a75e98d7d765ab218d29262da3bdb88ef3110ce69f1134ac3fd3a3dbe850a6cd67eccb2e4e69bc2abc299a966a06140262a593023751729fc96ef2f704bd5e5b2a780cdeeca9f31c68637b96f02f3184c98b8f21b028b824fde0bddd13325a4ae3e7471de45a85047e397467a6e0c599ec64d91e088a168645bb1728c333c8f1737e0eacb54193",
		"c86894e8a25598461e5095d09c1d4340c11c82d4e0fa220b619a9b22bf1c1a19",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		7,
	)
	tx_29_8.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"c86894e8a25598461e5095d09c1d4340c11c82d4e0fa220b619a9b22bf1c1a19",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_29_8.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 9
	tx_29_9 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_29_9.Vin[0] = createVinTx(
		t,
		"0101e3aad92108575bb767a9e600a7a939c1e331c4deaf986db6a05679848bdfa4b2304b2914a129f10fcd073cfe65573c73c7d44edc8bfa5bfa4d2731a774f7cf6e01",
		"38d1b1dce516903e2d29fd1663eb980e8d7d806e7f0afc99db5b5d462dc0f7fe",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		8,
	)
	tx_29_9.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"38d1b1dce516903e2d29fd1663eb980e8d7d806e7f0afc99db5b5d462dc0f7fe",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_29_9.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 10
	tx_29_10 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_29_10.Vin[0] = createVinTx(
		t,
		"010220fd387988b9b827bcce3876873bbdf6ac71d99c96d92ff46d881ec38929bad10099df0edbc699b232d7c5870ff4158ec8e60353dcf61a8f5e4bade7366422c2231bea61a2876d21d756cdccb7355fd4f56a57fb78d7f19ca7896ea25292ea2b217e4af4dad36bd9e0614ec5a00b06aae23c515a3e07596d08b7e6705504ccdf1074280331cc579c42a2c6589f699e1eeeab280f02418d0e97836e187954a52b190361ce58e94fc4f011eb503649525241977ffec718f34cf1803cf15db23de3",
		"3b80710dbc1429339db7da796d7bff8f8519f16743e11f2bd82aa2ca996a3c5f",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		9,
	)
	tx_29_10.Vout[0] = createDataStoreVoutTx(
		t,
		"f290c40531551726e0fa9cb66ec3f99fc5ebec81",
		"3b80710dbc1429339db7da796d7bff8f8519f16743e11f2bd82aa2ca996a3c5f",
		"96b365bba07da12fb305fd07a28a2590103abcf0acf240fef73d7fa7131d7061",
		"564d465364595653365f",
		"3860",
		1,
		0,
		2,
		"20fd387988b9b827bcce3876873bbdf6ac71d99c96d92ff46d881ec38929bad10099df0edbc699b232d7c5870ff4158ec8e60353dcf61a8f5e4bade7366422c2231bea61a2876d21d756cdccb7355fd4f56a57fb78d7f19ca7896ea25292ea2b217e4af4dad36bd9e0614ec5a00b06aae23c515a3e07596d08b7e6705504ccdf0478df1a523b8c035fd49a6479feca5b430e9cb8e8d210f546c077d2c58708ed2130cf23f2895b545b0241520c1c7381284ca0d22c43b84a591d72fc17164a9b",
	)
	tx_29_10.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"3b80710dbc1429339db7da796d7bff8f8519f16743e11f2bd82aa2ca996a3c5f",
		"4999999999999996140",
		1,
		1,
	)
	err = tx_29_10.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 11
	tx_29_11 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_29_11.Vin[0] = createVinTx(
		t,
		"0101ae10dcfa1dfe3826c4c569e1b960bf39e0b4a63f5e8df76f4188d03f06f8a43b4c2d1db6afda79dff052e09c4f4b97adda71d763a9942388c8f1ec4f720c8af800",
		"3e080a34556ba5f6992b9cc622e38aef8f4b3d850420badbc5a431c373c25f7b",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		10,
	)
	tx_29_11.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"3e080a34556ba5f6992b9cc622e38aef8f4b3d850420badbc5a431c373c25f7b",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_29_11.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 12
	tx_29_12 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_29_12.Vin[0] = createVinTx(
		t,
		"0101fe765d90b2f3cc44c1d0c4f3301d8b95ab7475180211357894e8b4c2cd0915b72d91a3ce501249ce38eedb0650f27d1c853378c34e4e58d7c446558d66b0c6ea00",
		"f52dc40af73127ce58faf87c8ff48e00001dfc3d70ea41d16c3e395e887eae6c",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		0,
	)
	tx_29_12.Vout[0] = createDataStoreVoutTx(
		t,
		"82a978b3f5962a5b0957d9ee9eef472ee55b42f1",
		"f52dc40af73127ce58faf87c8ff48e00001dfc3d70ea41d16c3e395e887eae6c",
		"19102b7b3e0c0b1af368e6d1def1a38ae36177fed54bbe214fd461526f85df06",
		"5941524653786578315f35646f63373d4c",
		"3537",
		1,
		0,
		1,
		"e3406b25e9276bc89878dd18390bfe6ea6a8a1eb927d364a8faba73cc68ea10f63f811e0648467861da9ddac2fb34f96cdc021433c3bf33aec1332f872c21f4a00",
	)
	tx_29_12.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"f52dc40af73127ce58faf87c8ff48e00001dfc3d70ea41d16c3e395e887eae6c",
		"4999999999999996463",
		1,
		1,
	)
	err = tx_29_12.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 13
	tx_29_13 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_29_13.Vin[0] = createVinTx(
		t,
		"0101215e86920fab0e95f34aa4c58f1bbc2d7ced5059bf505dcfc6627da5827917c62d1059efaf29aee4cf2859430f6152da3573e18e75f695cafabf447127d8f58a00",
		"dfe2a85bac0630eaaf2c8d6a86edaa1a7e6e303dbba0b6ea1bc11e9fa6fb6081",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		12,
	)
	tx_29_13.Vout[0] = createDataStoreVoutTx(
		t,
		"7aa22a1a0672a54a819659dffc1c4ca5383114dc",
		"dfe2a85bac0630eaaf2c8d6a86edaa1a7e6e303dbba0b6ea1bc11e9fa6fb6081",
		"b80836e997e344d0873a3081f219d52d5d784f1efcc9796153a28d327fd282c1",
		"7e66455443223b637d644a",
		"2322",
		1,
		0,
		1,
		"2f1d7c77f5bdb362c6e8a03efb9c3171437b48dd37d266ca51d5db731ac4b9d93464f7b2edbb56ae1936c337d74d03bbe9ccda26b427fedfc7ac46465eb8082401",
	)
	tx_29_13.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"dfe2a85bac0630eaaf2c8d6a86edaa1a7e6e303dbba0b6ea1bc11e9fa6fb6081",
		"4999999999999997678",
		1,
		1,
	)
	err = tx_29_13.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 14
	tx_29_14 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_29_14.Vin[0] = createVinTx(
		t,
		"01022f7e14b7c437d984b60cc2d474762f709f8419b5d1f1e30d5a86679c6f9ba5a422cfeb117087279099b47ff3b6aaf7be784d82c531f37b0a9d32fe660037cd2806b982a197b3fea0fb62bc5a4ef2d3c4cf75bbb19e326177fd7923cfe5a8c44907b7db485cac4a13801fa1702aae68b1b5d34c8efd5ce28df3ee710d8fc66fab105263e59c7f005fd7ecb1c1675ad33992c26835a77b3812f3c96a4e818854ef2c8fecf65171b3740299fca3676be40c0b4ed05d7348c90e8a0efb3a70bc9b3e",
		"db4eb6993fc5db2f03146af41bb60f453fa629a8c754575c0f770aeee6bcf9b8",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		17,
	)
	tx_29_14.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"db4eb6993fc5db2f03146af41bb60f453fa629a8c754575c0f770aeee6bcf9b8",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_29_14.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 15
	tx_29_15 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_29_15.Vin[0] = createVinTx(
		t,
		"0101ed3cd769805c22898c55530154b3661ab84fb906abf8267bf9ee2c44dd670f3804ec6e9b1fff5b2f3b7cd19b202c3a2b43687bb9aa631386b011cf1cbf08a3a500",
		"3c2134f7b71e5a3488853cf1dd5817fbcd011e5a47e89b10c90807e32bf1b0cf",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		18,
	)
	tx_29_15.Vout[0] = createDataStoreVoutTx(
		t,
		"1427f29745e5b8ebe70afb6646e73fd194da7a7a",
		"3c2134f7b71e5a3488853cf1dd5817fbcd011e5a47e89b10c90807e32bf1b0cf",
		"90d8510e5ee2b61176d54254bf67c34c46df594caebf17de68685a35801d1f3f",
		"2e733d507b4a53324a2e6c48",
		"3880",
		1,
		0,
		1,
		"da7d2aa170176c752f722ff69021911f7ac7194fc0237a0a816dc14458b1890f762e8d8aed008a15d1dcf56cfab237cb3d3dd5f4ecbb054cf74eaa5d5482227301",
	)
	tx_29_15.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"3c2134f7b71e5a3488853cf1dd5817fbcd011e5a47e89b10c90807e32bf1b0cf",
		"4999999999999996120",
		1,
		1,
	)
	err = tx_29_15.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 16
	tx_29_16 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_29_16.Vin[0] = createVinTx(
		t,
		"010230175e700747c6e5e3b6a04ef449cd7162aa214f3f5923c74b863adbad8b743e087dccb85469dc644c4cfcf5783297bdb935372a895354f314ac2e81b4011e590c1a186d9df4a8b7883fb4b46c5791d4ab6596c14a0a8e72f567a54a0d604cff2cb9b45a51afe9cd8dcea3c706241fbf11e45078c1d636e9654f8e91460980fa04e1b9f4de367005dd860f6573d9133f6089ae9d0b6f43c2ac2b39fde3b5448903eb8c622e1442a7bfbf7fdae74b8024a0903f717a62c99bf10c10f759c52e5e",
		"addf921d48f46de1c5d719f7588bbcf977761005cb28bb13e7f6ed2c925d0ce3",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		19,
	)
	tx_29_16.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"addf921d48f46de1c5d719f7588bbcf977761005cb28bb13e7f6ed2c925d0ce3",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_29_16.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 17
	tx_29_17 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_29_17.Vin[0] = createVinTx(
		t,
		"01022247362be8a30896ca6277788f455061daa8d1938a9ef8d977e7864a1a3f9af0012dcba8f27209d1f42d4bd2c3d48bdd1c1b535b02937b0409b80e9e163c79972e69bc04e266e5c7a3ea7de87356433e4a470a25dc93935449f1459ca046e56c0fa8bb69dce46832e0ba64229605aa86401d07b8747532e6626be54736a4ca2819e7f9f5c1265bb247520efc330e7c6fe65baad1731f5aac9fa6e219927f0b4d0b3e3f76ade27a5b5462077a09901502640f9685b5812c4ccc57c9fd33dbb103",
		"b5e5544e753dfc8585a174c328c21be7815b8755b342246cb13b7ffa3d1bc5fe",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		11,
	)
	tx_29_17.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"b5e5544e753dfc8585a174c328c21be7815b8755b342246cb13b7ffa3d1bc5fe",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_29_17.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 18
	tx_29_18 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_29_18.Vin[0] = createVinTx(
		t,
		"0101248a96574c27a1157b5e9b17f551c5892688ec5d27a3e77e1037b32dbc3bca54615ea608eca1ee3043ec4656b774594274393ff3534cfce54df6fc660ba975b800",
		"59a053830d9f9988e6313e7d4e62b7bf3caa7fc4eb19a509710679346fce75cc",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		2,
	)
	tx_29_18.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"59a053830d9f9988e6313e7d4e62b7bf3caa7fc4eb19a509710679346fce75cc",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_29_18.SetTxHash()
	require.Nil(t, err)

	// Height 29, tx 19
	tx_29_19 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_29_19.Vin[0] = createVinTx(
		t,
		"01022ecc8aac03b38c0b0f45919a1491bc2e2be9d4856c14ff0dcff3f630086a40502ea9e97e4b20556ec18ff449766a24dfb78da717761b866dcecbaa50465b214910417547aa9b4c9df2366974d21f36ec0e6c03fa15aa28a1579d3ed9d92736091c8d9fd6fd438e2b725594bcf028d39fea460825ad4d923d436e114a43209df60e93150ee3701e7061af4c580fc4aa8cb4f24b65ec6b0669ea9807697a6b13bc3053f03b449e8be719eb1df3b2b6abebe97316e79cfe9aca58d3514c80a6a5a3",
		"0ecb0457c22c9b6eca878cb1ab08f74e984b353153881a42a8af83b6266e1118",
		"fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd",
		5,
	)
	tx_29_19.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"0ecb0457c22c9b6eca878cb1ab08f74e984b353153881a42a8af83b6266e1118",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_29_19.SetTxHash()
	require.Nil(t, err)

	err = db.Update(func(txn *badger.Txn) error {
		_, err := trie.ApplyState(txn, objs.TxVec{
			tx_29_0,
			tx_29_1,
			tx_29_2,
			tx_29_3,
			tx_29_4,
			tx_29_5,
			tx_29_6,
			tx_29_7,
			tx_29_8,
			tx_29_9,
			tx_29_10,
			tx_29_11,
			tx_29_12,
			tx_29_13,
			tx_29_14,
			tx_29_15,
			tx_29_16,
			tx_29_17,
			tx_29_18,
			tx_29_19,
		}, 29)
		return err
	})
	require.Nil(t, err)

	for i := 30; i < 33; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(i))
			return err
		})
		require.Nil(t, err)
	}

	// Height 33, tx 0
	tx_33_0 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 21),
		Vout: make([]*objs.TXOut, 21),
	}

	tx_33_0.Vin[0] = createVinTx(t, "01013c6f1ad8e86a2f0e0b675f671977a27384dc5be71efcc4420dc7339f796f376b7c4b69887f46e43dcd54fae2b736358f52dca4069739dea4c565529a0760f7f900", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "9bd0ccecde5be25f12df135c4f81c63ac0fc314be77017c7e7efbadc6a00cd5f", 1)
	tx_33_0.Vin[1] = createVinTx(t, "0101ba4d75f4ecf2e64c278f6a5e864a05af13a65fb05b6c664be9a973d96b0264940a001253328fa3343fc0d0b05626cba90e21bd221d6bfd5e4901c4aaf3f98a9001", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "3c2134f7b71e5a3488853cf1dd5817fbcd011e5a47e89b10c90807e32bf1b0cf", 1)
	tx_33_0.Vin[2] = createVinTx(t, "01013e126cbe9280c98e47dfff47d2b35127fda0f68a2b52ca21199a6472939029a658be5257987a818ae55884338664f3ec4f415baaa646ff01da5c36a23cdc9d5001", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "3b80710dbc1429339db7da796d7bff8f8519f16743e11f2bd82aa2ca996a3c5f", 1)
	tx_33_0.Vin[3] = createVinTx(t, "010179d6ba3c6baa9252b3a6d504c1c35e96ee4b555f211adcc866c8ddb72beaa70b4f8ce4a1d3ad74dd8ec212c1578b3b844b2f9ca9c7ef7be7c97eed74ddebf95e00", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "f52dc40af73127ce58faf87c8ff48e00001dfc3d70ea41d16c3e395e887eae6c", 1)
	tx_33_0.Vin[4] = createVinTx(t, "010155c07c2b40952d0022bee8db77d28deed62201e4f5bff8a0bd47d8d5a8bc98fd3514a59f91ec9f6c15bf5522a45748abebdf6c230e5092cc6478d2b95d233cfa01", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "4bad7dcca377a058376005a612d45194137a1b8e77bee22aa89284434159600e", 1)
	tx_33_0.Vin[5] = createVinTx(t, "01010a168276c7ad7d105601990a2681bf8063df9341255c5d0e68beae5df180ae1b6053ec5fcb3dbe24bbc8abb615137376d3ccec9301338d10a2afaa244def5e9000", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "dfe2a85bac0630eaaf2c8d6a86edaa1a7e6e303dbba0b6ea1bc11e9fa6fb6081", 1)
	tx_33_0.Vin[6] = createVinTx(t, "01018fce8159f8c559c63cda2ce482980fa4e1ba45e8af0e4d6304c3322bb12b63746dddf7d19b0dc8854388b9ffbef18ec4a44b948dd7348e7afc538c731765293f00", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "15406f892afdb0e0d051759293e788a1215bc96b45a96205372f843ee0de9b55", 1)
	tx_33_0.Vin[7] = createVinTx(t, "01019e9f89f70e9f2f2f1876fd96cd750e8efb7058a02916d62f2755e2550235daf33418189cc4908868506e43fd72eb7756834e43091e1beb7113f85e91f3c7f68c01", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "38d1b1dce516903e2d29fd1663eb980e8d7d806e7f0afc99db5b5d462dc0f7fe", 0)
	tx_33_0.Vin[8] = createVinTx(t, "01012816ffc5684e50643224c4614dc0dddf378403f811c0b239d01777e72ac6a9a80ce85c5a6340b3828e956b3f69041224fa4b33bafe33fd535776f525afe4930c01", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "0ecb0457c22c9b6eca878cb1ab08f74e984b353153881a42a8af83b6266e1118", 0)
	tx_33_0.Vin[9] = createVinTx(t, "01017b9793a3daa62ff90982d1a75a2aaaa93807785fa7c4984b2e3795d91720ca923b2d8ef76f342a084cf2d58722c7c0680e67e63d040d7af255912d601bd30def00", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "10dbd0765d9da5e0d61791574329ff04d4fadabf749d2550b6b5a3a906180b31", 0)
	tx_33_0.Vin[10] = createVinTx(t, "01013ea96ef8549ab8684ea1526fadf50e0b742302495deaa4760fad4a7c4d85f8d317eea45e8ca9281af8d923156d191c5040ac9593d98e327f16b3be41b9beb94701", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "db4eb6993fc5db2f03146af41bb60f453fa629a8c754575c0f770aeee6bcf9b8", 0)
	tx_33_0.Vin[11] = createVinTx(t, "01013842f36226694483eb16785e25ca2468e7739f68606633ddbb9f432b049aebb6074f7928c9c69efb940a346a3304b64cd99e1e2ab340cd4117d510158dd8790801", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "962dbd4f1c5a1338697987e54c0467d80d87999f7d4fdadc95c382c4ea37abbf", 0)
	tx_33_0.Vin[12] = createVinTx(t, "0101c15b9e6e43ab1bdbc6a62bd0dbfb0f0bacbb395d618a303dff0807155d571be42998b8426b27dde8f1562db2c22b8154c59464e24cc5b2557108087d9cdc218a01", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "89f90ba0a43e50de81acfa7785601e54b56723bb8a731a5543709d8824304c7a", 0)
	tx_33_0.Vin[13] = createVinTx(t, "01016958e8bddd93168b9e0f5d6d35c386180e12a97688ec6ee4b72180cfe84e820568d91dc9ea76e04a2ba7492ae466de2f07d0d6c1c571503971e10fd48b2cc10300", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5ac7cc937e43623d211526994a622f2fb8307a474d9d92ef3919e8a5069527f2", 0)
	tx_33_0.Vin[14] = createVinTx(t, "0101e820b33cf51f19a96523701b87233666c900bd43155732e09b55e019ce193b2b20ebe4802a34c8fbd0d924c5e3e73628bf08f127830f41c7fb348332475dca8301", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "f020c514060c7b969cc9d457ddb8fb57b1aeda32d9ba89bf8e2273e10786e12a", 0)
	tx_33_0.Vin[15] = createVinTx(t, "01012fbb5f7e3052fa0770071d22b4232d93045780ddb8bbfb45b2429be5f1d6371b48714908ea0cfcc0c5d7f76ff4c50bafd352bd87bc300f5f85846f6e731d951c00", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "b5e5544e753dfc8585a174c328c21be7815b8755b342246cb13b7ffa3d1bc5fe", 0)
	tx_33_0.Vin[16] = createVinTx(t, "0101d4e286ff59b4e3670ebe34e0a2053ce912bcbad2a07e7ba07c524827313572bf306b435e8e7e03435fb48e32564f719340e2db3d865c717f4c79d4f525b0568100", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "59a053830d9f9988e6313e7d4e62b7bf3caa7fc4eb19a509710679346fce75cc", 0)
	tx_33_0.Vin[17] = createVinTx(t, "0101cb5fd330932d0ff0d2a80e2a4e77655c7aba9061fef506f20c4a797ab665c5593a6da0bd46cabe2bbda45a57d6949174bc7a9f2fcf8e557e837df153827cf8e401", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "addf921d48f46de1c5d719f7588bbcf977761005cb28bb13e7f6ed2c925d0ce3", 0)
	tx_33_0.Vin[18] = createVinTx(t, "010174584d79716858e23e5cac06cd759f0627bc78fdf5965da31bd19954e06064dc09c1c6013cf010f3dcd0651e277d3ca416425a1a8fc6f56c3eb174340cc88ebb00", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "3e080a34556ba5f6992b9cc622e38aef8f4b3d850420badbc5a431c373c25f7b", 0)
	tx_33_0.Vin[19] = createVinTx(t, "0101a13f2edfa7eec15cddef4baf257d446d353673602c02d23d27d89404f731c10b0f6307295eb184a838d9cdb9e8f8cb2cea4323d2cb1ffba110e9b4f380c5045301", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "c86894e8a25598461e5095d09c1d4340c11c82d4e0fa220b619a9b22bf1c1a19", 0)
	tx_33_0.Vin[20] = createVinTx(t, "01015823e189b49731982cc7117090d3d500db8718bb1623f445cef61badd31b4fc43d6b2bc450dcde855204f18d868f0d9b2f7ec5d7b6075cc26491b6155388050f01", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "fd7e1f1a57d9d244133fcc9457ecc981c2f11e1ea015d704e7b6d6d4a4d727dd", 20)

	tx_33_0.Vout[0] = createValueStoreVoutTx(t, "daff1aa9698b743af62b001317d62eec2a8bf712", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(0), 1)
	tx_33_0.Vout[1] = createValueStoreVoutTx(t, "09b4f445d496edfbfb8bd2c1daad34f4041c4147", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(1), 2)
	tx_33_0.Vout[2] = createValueStoreVoutTx(t, "27bf5398a931bcc0214d485e80b3ede8d9323c81", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(2), 1)
	tx_33_0.Vout[3] = createValueStoreVoutTx(t, "39a25f9727d36527b396b1b762e5ea26279b1943", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(3), 2)
	tx_33_0.Vout[4] = createValueStoreVoutTx(t, "5846fa79c42154fd9426a02a959f3e17f4aaf5b9", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(4), 1)
	tx_33_0.Vout[5] = createValueStoreVoutTx(t, "41e8c6a0f2d23ea2658a7e0aa208a495a408d1eb", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(5), 2)
	tx_33_0.Vout[6] = createValueStoreVoutTx(t, "ec3c82fe677a267c88b66b5b708e687fcefec0a4", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(6), 1)
	tx_33_0.Vout[7] = createValueStoreVoutTx(t, "caec083f2d205acc050f3e79f59beee870870cee", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(7), 2)
	tx_33_0.Vout[8] = createValueStoreVoutTx(t, "8286079ac03a79c82ce1e1c5d4f15b5d68c2ac66", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(8), 1)
	tx_33_0.Vout[9] = createValueStoreVoutTx(t, "9aaba6d90cb9f942cfa2965c5db6bbe8294e543d", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(9), 2)
	tx_33_0.Vout[10] = createValueStoreVoutTx(t, "a8b91f451f5785b160fbfa9b0728f0281707f536", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(10), 1)
	tx_33_0.Vout[11] = createValueStoreVoutTx(t, "7da036bd9035fb4ed72e73f4a4838b826bcfc489", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(11), 2)
	tx_33_0.Vout[12] = createValueStoreVoutTx(t, "23b6ae42a51e73554caf60f714505da6b675d433", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(12), 1)
	tx_33_0.Vout[13] = createValueStoreVoutTx(t, "e0dc37f05ae1bfee75936bd58679768ec6f8958a", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(13), 2)
	tx_33_0.Vout[14] = createValueStoreVoutTx(t, "fd1d8ebb7d20594108fea8d4270f592d6689084c", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(14), 1)
	tx_33_0.Vout[15] = createValueStoreVoutTx(t, "970b8037ef59a2be2276335a7adbc32d2c88806d", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(15), 2)
	tx_33_0.Vout[16] = createValueStoreVoutTx(t, "827cc3cc277a704efc95d57276d23a39af18d504", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(16), 1)
	tx_33_0.Vout[17] = createValueStoreVoutTx(t, "d9428e09265a87cc5455ef16f075f84fdb524c8d", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(17), 2)
	tx_33_0.Vout[18] = createValueStoreVoutTx(t, "7f2ae714fd4365c900500064531aa37f57975323", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(18), 1)
	tx_33_0.Vout[19] = createValueStoreVoutTx(t, "ffbabc81378c515a644e1b89b38a456aeffffff9", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "5000000000000000000", uint32(19), 2)
	tx_33_0.Vout[20] = createValueStoreVoutTx(t, "546f99f244b7b58b855330ae0e2bc1b30b41302f", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", "999899999999999999977427", uint32(20), 1)
	err = tx_33_0.SetTxHash()
	require.Nil(t, err)

	err = db.Update(func(txn *badger.Txn) error {
		_, err := trie.ApplyState(txn, objs.TxVec{tx_33_0}, 33)
		return err
	})
	require.Nil(t, err)

	for i := 34; i < 39; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(i))
			return err
		})
		require.Nil(t, err)
	}

	// Height 39, tx 0
	tx_39_0 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_39_0.Vin[0] = createVinTx(
		t,
		"01020025d7a6a3b0fc0e2cafa70d184d95c5ec23338a46f4b1bb44538971c4ee060a24cd449bcb7d13cb1a74a45741502a9d1e8ebdce440f77d87989f28f2bdb92e5063c48158ec38b1601e81ff6ca912bd6ca8a0e69403ed34ef7c3bec2ba25d84a2d62e485696587cf2b19136bad842cb090139a315f3845730bf223f3b94795e82ef9f3c1624b43ce93f20cc9ddf91a4458c6887b0eefb6dd686e043e43a89cd72d7402fee7869556e6d74f2d081426084712acb3e58ccf3abc4afa35a88684f2",
		"08340ecc4a88b24d03c5a0efe6cf47b3e310584a0b40ef0dabb8552e78974806",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		9,
	)
	tx_39_0.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"08340ecc4a88b24d03c5a0efe6cf47b3e310584a0b40ef0dabb8552e78974806",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_39_0.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 1
	tx_39_1 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_39_1.Vin[0] = createVinTx(
		t,
		"01021887f77650e10016f6c2b5bceb20adebd1e5d37ae9b3f6cfd9c08724a2981695209328b95281b0ca0d19cade67730370e796cc9bf13e63efc5296db4645b3713263fe19721c92591263c1f20d866f545bbe07f95bd978b84c9ea6dd5982253e310e5325e63cfff5ca6beb6586eae1b97f41e0e8e142570b4e17ebf14c09a48ef04fdaa5c5a0f1a6fb1b234ca1e12cb7f7110cafaf0ae2642e366d45aea69096411bd0c7aecfe04812b340c3f9035b598d22f5bacf65f07c32db12e52f4b27728",
		"da98ed500d0cb2e9c694cee64f01826f74dabcca37f53046bd188e5055249600",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		19,
	)
	tx_39_1.Vout[0] = createDataStoreVoutTx(
		t,
		"ffbabc81378c515a644e1b89b38a456aeffffff9",
		"da98ed500d0cb2e9c694cee64f01826f74dabcca37f53046bd188e5055249600",
		"ab31f5e8dac72e92e3426d9a0580dd3c3bac5f7a6335f91edeb711ccef873772",
		"495a6349376d7c2f354a6176",
		"4656",
		2,
		0,
		2,
		"1887f77650e10016f6c2b5bceb20adebd1e5d37ae9b3f6cfd9c08724a2981695209328b95281b0ca0d19cade67730370e796cc9bf13e63efc5296db4645b3713263fe19721c92591263c1f20d866f545bbe07f95bd978b84c9ea6dd5982253e310e5325e63cfff5ca6beb6586eae1b97f41e0e8e142570b4e17ebf14c09a48ef2cbf218996636e4ce00984cf9188e023b9183b5375249d74d4206740951df4af236f58e43c28c68c2cade2314cbf0b5ade8e40cb5b0a000caac3e43ced8dd295",
	)
	tx_39_1.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"da98ed500d0cb2e9c694cee64f01826f74dabcca37f53046bd188e5055249600",
		"4999999999999995344",
		1,
		1,
	)
	err = tx_39_1.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 2
	tx_39_2 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_39_2.Vin[0] = createVinTx(
		t,
		"01013c58902f6654abf7dbe3d4c08d5c007b4c194f4b2212e9108150bb48e889585a3be8f84319e4edda079eb2344bcfaea72310057919bc3508599e8fba0dec37e401",
		"f484756c96d40121fc53b30f7f187858f8c4ce3218d1c00374e6f3adda83da79",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		10,
	)
	tx_39_2.Vout[0] = createDataStoreVoutTx(
		t,
		"a8b91f451f5785b160fbfa9b0728f0281707f536",
		"f484756c96d40121fc53b30f7f187858f8c4ce3218d1c00374e6f3adda83da79",
		"f63b970c503a11ed5115dc480ef56a64be9b54ac81d60113f923c58a9ff44a61",
		"7468434e706f72434e57",
		"3860",
		2,
		0,
		1,
		"b9a0a5c691830093241b7e6349e9d9ded927859ec3e1fdbaf0808ddb56db6ed63e9cbe46233aeab579ebb2e8dc22f783dd89b7e788fc45b2011b89fa631861e300",
	)
	tx_39_2.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"f484756c96d40121fc53b30f7f187858f8c4ce3218d1c00374e6f3adda83da79",
		"4999999999999996140",
		1,
		1,
	)
	err = tx_39_2.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 3
	tx_39_3 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_39_3.Vin[0] = createVinTx(
		t,
		"01015fa6a750d6a4f010be8c974dff96dafdd037f6c61ea629335db7afca6eaae1de3ab55b3d3f59cf02c3c39adb002adba02dce552aed5106f39cc87d9e50ba6ece01",
		"0d82e50b781801ec35744ce9c7534fd589d94d86f73f8ca90e49d76c25fca4b6",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		14,
	)
	tx_39_3.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"0d82e50b781801ec35744ce9c7534fd589d94d86f73f8ca90e49d76c25fca4b6",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_39_3.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 4
	tx_39_4 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_39_4.Vin[0] = createVinTx(
		t,
		"01018bba5c1ec204dbb8e91930839e3ed936a30b3633f1dd1b12104ad571593c91451ad6c5de7595ebf88a1eb3182b016ebd07c90b6fa3e9b2cc1214ebc8ce950c0c01",
		"c49ae2710615db148b6057a87b2023007655f5193f14b4877e36eb0f8c68a7b7",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		4,
	)
	tx_39_4.Vout[0] = createDataStoreVoutTx(
		t,
		"5846fa79c42154fd9426a02a959f3e17f4aaf5b9",
		"c49ae2710615db148b6057a87b2023007655f5193f14b4877e36eb0f8c68a7b7",
		"e9884073fb06ebd06703266590185213c061a2ee0086d6bb0ad1695cfd7df2bd",
		"555f77794930537d554168344635765e",
		"4704",
		2,
		0,
		1,
		"bf0df3215afe0bceb9ac420bc99ab766413d8bcf5e225d29ed7e2d7817c3345a579c369eb1a989217abccdd32de7d6d76a77042bec5214489e3e631ff5d7684101",
	)
	tx_39_4.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"c49ae2710615db148b6057a87b2023007655f5193f14b4877e36eb0f8c68a7b7",
		"4999999999999995296",
		1,
		1,
	)
	err = tx_39_4.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 5
	tx_39_5 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_39_5.Vin[0] = createVinTx(
		t,
		"01022aa9815562fd588cd79e9407f64c0c436e34560d6a4ad91108b36bb68dfe03f30b0d5a6b417edd7e39b727a07fa1087c54609af7f62c230aab61d274e9f4ed1d13f4d9d1e11544793980925baed6d9efc46972155b389c7298c1ac4ce952328c11edbea2a4c386d71cb3abe051368f2d2a2aa6b98fbea3913980e66ad44adcbc0682426c0fdb4dabe0485c66e986dacf5a8fed9587bf44bbe7766661f91b5469284009e32a2031f97732e6b6cd37d7af5b7853f30001691a2093de49b5736f32",
		"bdd2e7653f07cb957ff7940265e8ab0252c10ff0a447cf4ac7a29c3e271b25a7",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		15,
	)
	tx_39_5.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"bdd2e7653f07cb957ff7940265e8ab0252c10ff0a447cf4ac7a29c3e271b25a7",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_39_5.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 6
	tx_39_6 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_39_6.Vin[0] = createVinTx(
		t,
		"01021235d8910c413355d388b05f4870b2caa7bd3014a01019a1402883e582f295930bed81df14c07358eeaf4b8ef8dd399be02ab2e80dd8147e7f633ac438c3a30f235283f0e582513d9c4606de90fc079644c6ecb141362dac4d1a0fdd1cf809ab0ca13a18f92e0ba080e1bc6b833c6c4f3fd83c4a82f257fe037b8da7ee421df80e403597dab1386082c14021317688323071d7ba4698cfa08f9528148a2543511f20965a63b025239172c7d26ad41e591b50d397dd7c6e9a0ee7d19c35a2bd1b",
		"fbc846180774451c7ff6bc0fc53cca2d53a195388693109fcba4a5037395823d",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		13,
	)
	tx_39_6.Vout[0] = createDataStoreVoutTx(
		t,
		"e0dc37f05ae1bfee75936bd58679768ec6f8958a",
		"fbc846180774451c7ff6bc0fc53cca2d53a195388693109fcba4a5037395823d",
		"30aa61e43f806b28cd55dcf22df0f085543c1d137acb8da89f8dd3c0c8ee4911",
		"4f505076394b572f714d534a626530382c4a",
		"1970",
		2,
		0,
		2,
		"1235d8910c413355d388b05f4870b2caa7bd3014a01019a1402883e582f295930bed81df14c07358eeaf4b8ef8dd399be02ab2e80dd8147e7f633ac438c3a30f235283f0e582513d9c4606de90fc079644c6ecb141362dac4d1a0fdd1cf809ab0ca13a18f92e0ba080e1bc6b833c6c4f3fd83c4a82f257fe037b8da7ee421df80b86135fb7e38b97f335270a7396f3f9ccf27cac9ddf1a39aeee9a798aa6f6b902cca3d9856aa3e5f57b5a1a41e7fc5f82472d2af0f128503ccd953b301eab35",
	)
	tx_39_6.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"fbc846180774451c7ff6bc0fc53cca2d53a195388693109fcba4a5037395823d",
		"4999999999999998030",
		1,
		1,
	)
	err = tx_39_6.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 7
	tx_39_7 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_39_7.Vin[0] = createVinTx(
		t,
		"0101bc36a203624df34f87589c4bbe4dfa940a870a90c87301b2c2ef120246c1e9e27e8a961dbd7c33584500d08e3101f5c76c349e4a69dc28572a83cf9628259d8401",
		"74441c1828c6eb31eecc91939da6dfbc479e3418040ab57740fad78f4467815c",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		16,
	)
	tx_39_7.Vout[0] = createDataStoreVoutTx(
		t,
		"827cc3cc277a704efc95d57276d23a39af18d504",
		"74441c1828c6eb31eecc91939da6dfbc479e3418040ab57740fad78f4467815c",
		"e971bcccb1cb270ce46aa0094d5ff3cea611f52b6580091463c4c1351df28a26",
		"57347430307e75625153",
		"2316",
		2,
		0,
		1,
		"b30e12ef7bb02cd80afa2795501d9ad4fbb5d61f6bcde520aecc44446e961e7e370fda108ef235f2a3bb33b16ef5f26af4e55ac8346946b3d9a3c8aea6d251bf01",
	)
	tx_39_7.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"74441c1828c6eb31eecc91939da6dfbc479e3418040ab57740fad78f4467815c",
		"4999999999999997684",
		1,
		1,
	)
	err = tx_39_7.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 8
	tx_39_8 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_39_8.Vin[0] = createVinTx(
		t,
		"01020306051930408eebea4043dac8dedc9da0aa571d03589b82b91c8ac610ef90e226e865b7fac3ef71246e204d2d5c2e8010e68f9ddfb226cd8a5229018a82e50423043e9a4e9c7467347d0cafc91ac2d66e8d54d191a8490e10c2b0c5e2f319072ce617202634446a13e79eaeef5f93409a898ae25e22eb4691a68cb84704be8804e7e100099d261164db6b3feff2644641463f8fb9bb70c2d755cd3bc85b03de138a87e30a2d560fd6c5eb9bee5afa9e8fd30d95f2b4d166dfade0fe03598a48",
		"0e21cbaf4a20c6d16cc4d5189421d94a1ba4b03c7730cc659d2a0533c19a49ad",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		7,
	)
	tx_39_8.Vout[0] = createDataStoreVoutTx(
		t,
		"caec083f2d205acc050f3e79f59beee870870cee",
		"0e21cbaf4a20c6d16cc4d5189421d94a1ba4b03c7730cc659d2a0533c19a49ad",
		"5d3ba2afe0ba96d2232a1445acdbf0d6d1ee0ac43b83ec04f03eb7e5fe0d90dc",
		"3b51712d455e526274685a67",
		"3104",
		2,
		0,
		2,
		"0306051930408eebea4043dac8dedc9da0aa571d03589b82b91c8ac610ef90e226e865b7fac3ef71246e204d2d5c2e8010e68f9ddfb226cd8a5229018a82e50423043e9a4e9c7467347d0cafc91ac2d66e8d54d191a8490e10c2b0c5e2f319072ce617202634446a13e79eaeef5f93409a898ae25e22eb4691a68cb84704be882e78f631634f3464107404f56b253a9c12e9a8064c3a8c0aae023515a6eb768815319cbf2a8c551a149ed8ca6dc7e6ecc0454c0fb722706cb8c1c92f89ef803e",
	)
	tx_39_8.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"0e21cbaf4a20c6d16cc4d5189421d94a1ba4b03c7730cc659d2a0533c19a49ad",
		"4999999999999996896",
		1,
		1,
	)
	err = tx_39_8.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 9
	tx_39_9 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_39_9.Vin[0] = createVinTx(
		t,
		"01014a3c9192134396fb8fc7f7e96df9ffcf2e09146005a25101fbc38302b7de07bc14f89917a3a6fe550efff679693f0fbe9cb6a3a03100fd01c3e4f64506abd15100",
		"0cd750e424060f4d6ad2e6b32a037603ec27eed8004c05fb09af5f6d0eaa0c94",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		18,
	)
	tx_39_9.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"0cd750e424060f4d6ad2e6b32a037603ec27eed8004c05fb09af5f6d0eaa0c94",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_39_9.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 10
	tx_39_10 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_39_10.Vin[0] = createVinTx(
		t,
		"010151dcffc276f38187bbc4ab5804b428006cb1c757fc585ec98d39d559cc1ca91c5a024b97be356df01166fc1e52f08ccc05bbb02da0dadf2ac453d27cf882fcef00",
		"3bebc51687f9fcb6ff2d55e016b818f85393eb85aa9ce67a57553cad6fc6f320",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		0,
	)
	tx_39_10.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"3bebc51687f9fcb6ff2d55e016b818f85393eb85aa9ce67a57553cad6fc6f320",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_39_10.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 11
	tx_39_11 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_39_11.Vin[0] = createVinTx(
		t,
		"010227ac85377de8d1561e5d55d67115b9c656263d93c2763e5987e5bc718d107d45015e807c38bce73598a172cd05bf953f623697645eaec5e2ae379967543993ab2ee110cfb955c164a1ebf72f628e0eadebcb4b2767dc77f0b072023833ff8eff1066b7734d369557d303c94429b6401bb8378d19c0008ee677baa609360610e91d2409eb9119a7ce1a900ddab81d5d5a852a290bc7054776e2026711475fed32087e7441fa5fbfb9ae205bb45ef35a066ebe4102a616f4c078939f84f9cab613",
		"73a7cf8506d9382a3ae91b337020d461dbbadc7ed95e78d370bf000a3c0dd2f7",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		11,
	)
	tx_39_11.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"73a7cf8506d9382a3ae91b337020d461dbbadc7ed95e78d370bf000a3c0dd2f7",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_39_11.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 12
	tx_39_12 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 2),
	}
	tx_39_12.Vin[0] = createVinTx(
		t,
		"0102209b9493e7f20f3c56ad2f6dcb11dea11904d26d5c374339821cbba820d6a3170853162b2450e6f754858500098e7898021887329ba259ef01a859234da91eec24e936e46291e27cac1bdbf5d09335149bfea6eec4e98fdc119c4ad6c7e8503510ad747e4b40367ff66bbff993dde5084e17b1539046b2fe875730ec175fa950079acf4d7795e4da0560085b4496e2a800bd3d2da7e3312f2a0a1a3bf4d8fc36282b9e91b15ae3c69807a3211da83d7595beb772513bb2712156c7999d5f4f6a",
		"602bed90699d2ff4c104bea29f9e4151e0d40a214f7f07c954cf5315962a85e5",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		1,
	)
	tx_39_12.Vout[0] = createDataStoreVoutTx(
		t,
		"09b4f445d496edfbfb8bd2c1daad34f4041c4147",
		"602bed90699d2ff4c104bea29f9e4151e0d40a214f7f07c954cf5315962a85e5",
		"761b53b0aaffab50eda81008275d42c12b7ad78516d02ac2d41cf5f9cd34a356",
		"72734b5546794c4c792d646e4a52",
		"1560",
		2,
		0,
		2,
		"209b9493e7f20f3c56ad2f6dcb11dea11904d26d5c374339821cbba820d6a3170853162b2450e6f754858500098e7898021887329ba259ef01a859234da91eec24e936e46291e27cac1bdbf5d09335149bfea6eec4e98fdc119c4ad6c7e8503510ad747e4b40367ff66bbff993dde5084e17b1539046b2fe875730ec175fa950055943e1a4c615b95be7e1b51b7953ec5a28f72098ebdcc7c7274442fe4a453612e653483d08df11666a76432fa11dae968df3d9d39b58d54d8b176fa246b3e4",
	)
	tx_39_12.Vout[1] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"602bed90699d2ff4c104bea29f9e4151e0d40a214f7f07c954cf5315962a85e5",
		"4999999999999998440",
		1,
		1,
	)
	err = tx_39_12.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 13
	tx_39_13 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_39_13.Vin[0] = createVinTx(
		t,
		"01013330137ceedf426910837e930eb481d94e66d974508752c6a232334559bd60670b72b0db7e49b8e8e0b568d8914975150b28ced17ebfb3e7b249daa559d8e32001",
		"e413d0bc8b7f86759e09fcb04c6e015bd13c6c3297f7112305f162160259e941",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		12,
	)
	tx_39_13.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"e413d0bc8b7f86759e09fcb04c6e015bd13c6c3297f7112305f162160259e941",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_39_13.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 14
	tx_39_14 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_39_14.Vin[0] = createVinTx(
		t,
		"0102083b1cbb965268494d80df6403107b97e4265888cc847bdf4af2b03442e20a960fb8b66f9783f0c7dfc5ed0ff15d1c3a134bdad3f2e06d343bef98d495cf0ea024c5bca16ccf7c0428a968cadfb22592be6b93c8bec0e5def8b1178fae75e8be1430fe117c8e3aff95e14382d258325cd067b0311038af5e7b427f5a65a9a4fb2c7066de9b8c1d667425203804f49ca0f20663097842362ad99181027368f29f016c5ac2de4a4aac35be1e37afb74eb5cb6192d2ce6257a6dc208176b6efdfb7",
		"a5f8b51946089599b5b1489b5eef08026ed5b3e19601fc0fb29be51d8c727d5c",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		3,
	)
	tx_39_14.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"a5f8b51946089599b5b1489b5eef08026ed5b3e19601fc0fb29be51d8c727d5c",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_39_14.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 15
	tx_39_15 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_39_15.Vin[0] = createVinTx(
		t,
		"01020a3a6e16f4c13fc600c3ee83232b771f38c6145a26d08d7caaafd634e1eaec4b2a2cba93d2141a4c94fe601ac65f57f31de47a5b393fb68e5bfaab588ed22019001a13694e88b5aa0d7a87355fadbb23fa4f63e143de7997955572ee4e03ae3a224b59f7a52dcc1e372b77eea9af798f68ace9fa061998ffb2429e45edb695c022a148024733272eeb4232d17ffbb8c765b20f9295cf459f2cf674e7814ab27529148fbdf480f287473157b1db9fd569bee998d442b521303c36626e65151290",
		"8ab7f9340b8e3302ca0c36fe85a1fb9853c407e49bedcfc7a5c5a855a538a56b",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		5,
	)
	tx_39_15.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"8ab7f9340b8e3302ca0c36fe85a1fb9853c407e49bedcfc7a5c5a855a538a56b",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_39_15.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 16
	tx_39_16 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_39_16.Vin[0] = createVinTx(
		t,
		"01018de75197d2cfb5b32f224a6e0bc326c54732dd40866d0ce928ed5cbcc097b4030465cdaa572ce0e94459a988839cb487f5820dd5d2f2e436ab3e7cafc34eea6800",
		"2f873e1c0a08f14c9af2afaf0ddd07974fcf63201f23459bf8d7d3807fc58638",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		2,
	)
	tx_39_16.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"2f873e1c0a08f14c9af2afaf0ddd07974fcf63201f23459bf8d7d3807fc58638",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_39_16.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 17
	tx_39_17 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_39_17.Vin[0] = createVinTx(
		t,
		"01010277f0c5098d159d9f3a6f3d6895ced693d2157fd04877ec9aeb77c77fdebc4a76bb112bd2529c476d5bf99bdea918528f13912c1dca06e39fd2211cb9b5c2b901",
		"0f9cf88d61ff2f03898521837d51aea739d0917cd63f7bbfe9471ed8130e071a",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		6,
	)
	tx_39_17.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"0f9cf88d61ff2f03898521837d51aea739d0917cd63f7bbfe9471ed8130e071a",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_39_17.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 18
	tx_39_18 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_39_18.Vin[0] = createVinTx(
		t,
		"01022e4055664df90c3f015d4d970b74efd8f4ea50149886076ccef20c1d57e1f97d0878d4599a4c59ff61efd03c812db10e8455d8bfd64484b954018c733182bcab115a03904132f618680db49c001907b978e64ac2524fd17e21c1e394f7aab23b0d72f8c15c39e034f4bb1d8af25321025426e5fa4f9f861e497aa83fd6cbf9fb2e085a2a896fb88ba54aa7ac6d649f52fef6b8140952cf9c2c31a79c963aa6fd243fc86953599da496ce0a61b8c9d04bc6255ed7d510f1de465ca4202bb6cd8b",
		"30bcf2eaca7d4e26381c957a33451503167f4117a6ace74db4bd5fad6f163e4d",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		17,
	)
	tx_39_18.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"30bcf2eaca7d4e26381c957a33451503167f4117a6ace74db4bd5fad6f163e4d",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_39_18.SetTxHash()
	require.Nil(t, err)

	// Height 39, tx 19
	tx_39_19 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 1),
		Vout: make([]*objs.TXOut, 1),
	}
	tx_39_19.Vin[0] = createVinTx(
		t,
		"01016c5479906644265f52ba22245d5ddac7427020f1897ee5101f849745b6345b2c23e84323f26ae9351493f5de8b8945658813bed75b3c917f61ea98670ce53ea100",
		"1ac3dbf5c032234b82cbeae88c76ad300bdbf41f1b4e8392137d1efe81e082c6",
		"40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221",
		8,
	)
	tx_39_19.Vout[0] = createValueStoreVoutTx(
		t,
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"1ac3dbf5c032234b82cbeae88c76ad300bdbf41f1b4e8392137d1efe81e082c6",
		"5000000000000000000",
		0,
		1,
	)
	err = tx_39_19.SetTxHash()
	require.Nil(t, err)

	err = db.Update(func(txn *badger.Txn) error {
		_, err := trie.ApplyState(txn, objs.TxVec{
			tx_39_0,
			tx_39_1,
			tx_39_2,
			tx_39_3,
			tx_39_4,
			tx_39_5,
			tx_39_6,
			tx_39_7,
			tx_39_8,
			tx_39_9,
			tx_39_10,
			tx_39_11,
			tx_39_12,
			tx_39_13,
			tx_39_14,
			tx_39_15,
			tx_39_16,
			tx_39_17,
			tx_39_18,
			tx_39_19,
		}, 39)
		return err
	})
	require.Nil(t, err)

	for i := 40; i < 44; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(i))
			return err
		})
		require.Nil(t, err)
	}

	// Height 33, tx 0
	tx_44_0 := &objs.Tx{
		Fee:  uint256.Zero(),
		Vin:  make([]*objs.TXIn, 21),
		Vout: make([]*objs.TXOut, 21),
	}

	tx_44_0.Vin[0] = createVinTx(t, "0101a93c61c19c91a24ff534ce1f83be52b189c6df723f3bbc6125c03633d33e9f7e5007360ef65f2fe56ffb6b5de3717acf0656b66f04823a84edbd76e4caa224b200", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "c49ae2710615db148b6057a87b2023007655f5193f14b4877e36eb0f8c68a7b7", 1)
	tx_44_0.Vin[1] = createVinTx(t, "010163709eceb263b4c2988343ad7fa2ba092085614658da268bde9475568b8a2bcd2b851e9f3978ac8c4f277d4efb1fa7c7f66de464b8203829c7c1ff766de4c4f600", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "da98ed500d0cb2e9c694cee64f01826f74dabcca37f53046bd188e5055249600", 1)
	tx_44_0.Vin[2] = createVinTx(t, "0101ef253d214adf45485731689c242119fb6143729a37c4677b9c529bbebd76b8ea2099d2a25b56c271ccf958e0959a216b45de1c5ecd187742c48a5e3d723819bf00", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "f484756c96d40121fc53b30f7f187858f8c4ce3218d1c00374e6f3adda83da79", 1)
	tx_44_0.Vin[3] = createVinTx(t, "0101a05cdb7fa1c3d32d477291e909d09631c48c480c8a5a14b0b2d5a159d87886f077c2ca402367b44a396ef4a22af8d50a963922359eefae2afc93891aa8fa1ef601", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "0e21cbaf4a20c6d16cc4d5189421d94a1ba4b03c7730cc659d2a0533c19a49ad", 1)
	tx_44_0.Vin[4] = createVinTx(t, "01013af2b1a8a9d49e4922e0f5dacadd22dd645bee63b658b554a5352dfce8ae839f0748a347454ee835d82194d5deabcc2f6cc30bebfaa7a6ae83244edd0dbc525801", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "74441c1828c6eb31eecc91939da6dfbc479e3418040ab57740fad78f4467815c", 1)
	tx_44_0.Vin[5] = createVinTx(t, "0101bf61fac60efeb7e4611eb37538664a21c30a1e94ddcf033b6f48a5827d7178d876fdd88e5daa91827a650809a32bb3a572e768895cd26f45fc0997e3be03250d00", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "fbc846180774451c7ff6bc0fc53cca2d53a195388693109fcba4a5037395823d", 1)
	tx_44_0.Vin[6] = createVinTx(t, "0101175624e7e80bf3bb67778491a2d80eea9726cf9fc68b7ec0ae3b37bd60792a915d6caf5f3482d4a302124396c8ecf253867585c7efab8f02c9059a4a67a5749a01", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "602bed90699d2ff4c104bea29f9e4151e0d40a214f7f07c954cf5315962a85e5", 1)
	tx_44_0.Vin[7] = createVinTx(t, "0101a7a4c844bbb8c71e3e44203e519c2109384b0bdc07fa7c3c325e0398905cd7804c74e2794e24a6de8b9710a573530e50ce7e8784d7bbe91ffaf47caf5657533300", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "73a7cf8506d9382a3ae91b337020d461dbbadc7ed95e78d370bf000a3c0dd2f7", 0)
	tx_44_0.Vin[8] = createVinTx(t, "0101927c3ab0608741e7d1b27608ecf1530422f7a3aa6773461a8204d0e41718153b0a9da694dabf91dc43216384926a1f2dd705aaea11ba09bda0f46296f3a1e52c00", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "3bebc51687f9fcb6ff2d55e016b818f85393eb85aa9ce67a57553cad6fc6f320", 0)
	tx_44_0.Vin[9] = createVinTx(t, "0101ac26cbbfa1200f21e384330feaabb20cd3cdbf76ed366a1cd4bf8ddd2545e853306390d4495b58215a9e26728c6e1316405a9f28303619699e3e2d37c7d7977601", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "bdd2e7653f07cb957ff7940265e8ab0252c10ff0a447cf4ac7a29c3e271b25a7", 0)
	tx_44_0.Vin[10] = createVinTx(t, "010173b733cf8c13d40a8fba3a006e907e774729d6aadc19a2a5bb43694259cdcad14ca0fc908d216059d63815a41035e8d735e93196565ba2881641e03e71e0e1af00", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "0d82e50b781801ec35744ce9c7534fd589d94d86f73f8ca90e49d76c25fca4b6", 0)
	tx_44_0.Vin[11] = createVinTx(t, "01015299e80e6a7f36928e630b50462bf52ef08ce3f57abc083c1ab3e57e402ee5ef17d49c0f0ccf19c7e8acbf9a82d31ec67619a72396f39a190911fe45a766edb801", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "8ab7f9340b8e3302ca0c36fe85a1fb9853c407e49bedcfc7a5c5a855a538a56b", 0)
	tx_44_0.Vin[12] = createVinTx(t, "01015cfd7c73f85c6b848f9ef37ec8bfc7be00287cccc303f520c0bced8ac8426cc90593eb9ae8707aaeb082b8b498200b7ed37cdd10366f64bbd5cb7cbeca55666600", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "e413d0bc8b7f86759e09fcb04c6e015bd13c6c3297f7112305f162160259e941", 0)
	tx_44_0.Vin[13] = createVinTx(t, "01019e5282512035c938432def31502e682489f57254d732511e09da74baee0bfd1f655bf41e9e4c9482cb59e5a543d2e4f854e5d515773390c1b3ea1f4cc6e48eb101", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "1ac3dbf5c032234b82cbeae88c76ad300bdbf41f1b4e8392137d1efe81e082c6", 0)
	tx_44_0.Vin[14] = createVinTx(t, "010141c6945006874bc0e7086ee18f0a3455d20b2e1b6dc49c134a093adffb33ff390ff986569e46ea88635823d244413057f6db2651649fc2c1b2beeda9592b03ae00", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "a5f8b51946089599b5b1489b5eef08026ed5b3e19601fc0fb29be51d8c727d5c", 0)
	tx_44_0.Vin[15] = createVinTx(t, "01010051dc656dc3971b7a53814616ea30107969265fddb06b9827db5528e6dec77a75a1dc7545ace0ba6e09f099bdac2ea875eb9e55fd4809eac818d0b9f0d53fd200", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "0f9cf88d61ff2f03898521837d51aea739d0917cd63f7bbfe9471ed8130e071a", 0)
	tx_44_0.Vin[16] = createVinTx(t, "01017dbd5901395ee7b1e0b1089d3e9b6c438a23439056f86560cbf6f72b90043be967d5fd127593ce86a4ede7f6e60a03491f974a86765128262e0cdd22d115c82401", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "2f873e1c0a08f14c9af2afaf0ddd07974fcf63201f23459bf8d7d3807fc58638", 0)
	tx_44_0.Vin[17] = createVinTx(t, "010115b98cb09f186a318b2c33c782e679f5c93189f38c87a836240838b62042bab267051330f76700442a7d8eb28d34562b79e56421f34d74bf7d0b2ef4c30bd86301", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "0cd750e424060f4d6ad2e6b32a037603ec27eed8004c05fb09af5f6d0eaa0c94", 0)
	tx_44_0.Vin[18] = createVinTx(t, "01011e8f4091003a64402f97ec1010a08f05f116634075c373d498463909301501b42efb3910fa11f764f77fd4ee433ac079e1312a1b03d35f5affd02def2f6a4d1500", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "08340ecc4a88b24d03c5a0efe6cf47b3e310584a0b40ef0dabb8552e78974806", 0)
	tx_44_0.Vin[19] = createVinTx(t, "01013127adbb77cb16202e0f679a1551969c205fc7d9e519f0ee75f5d7a7358fd378116b08155d85af435ac223f86e25aaf29bea41003f40b8deae374e273c5b551201", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "30bcf2eaca7d4e26381c957a33451503167f4117a6ace74db4bd5fad6f163e4d", 0)
	tx_44_0.Vin[20] = createVinTx(t, "010181f55f942e2dff5222add869e7d4098f8091afac329f64163ba53f5fd911134904aa3d44d900b696baf75fc832a3723283692d947ef7b376f3cffd217c6e03a801", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "40dd848b38b5700793bb9179eba2fd48dab811297e405b5206b1a89a6c032221", 20)

	tx_44_0.Vout[0] = createValueStoreVoutTx(t, "a75e73cc5e38ae62afaeb132cd39c6192a8d53ec", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(0), 1)
	tx_44_0.Vout[1] = createValueStoreVoutTx(t, "8507aaadc753a4412401bc63d5623a7e6d574e15", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(1), 2)
	tx_44_0.Vout[2] = createValueStoreVoutTx(t, "a069d04587d0553cbd12f50b2dc98d3b29b9ae8c", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(2), 1)
	tx_44_0.Vout[3] = createValueStoreVoutTx(t, "6070e405342649ae6bfc887bdbe37a43ac0bbc90", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(3), 2)
	tx_44_0.Vout[4] = createValueStoreVoutTx(t, "63d211a0918e3cec11f91bfe3050c063b6e15dc9", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(4), 1)
	tx_44_0.Vout[5] = createValueStoreVoutTx(t, "1247a47aca6adabbd52124b9acb911f115c6f60f", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(5), 2)
	tx_44_0.Vout[6] = createValueStoreVoutTx(t, "4d9c2f8448cdd1403f41418f30d40178ae540f91", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(6), 1)
	tx_44_0.Vout[7] = createValueStoreVoutTx(t, "1b388eea55630968da468d2c5e3bf87d2e06e258", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(7), 2)
	tx_44_0.Vout[8] = createValueStoreVoutTx(t, "476d2c2b1a5aa79874e25e9f50ca3dc24a7993ac", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(8), 1)
	tx_44_0.Vout[9] = createValueStoreVoutTx(t, "29f1a1872056b88924d2fd1e2b9f2f3832dc7cec", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(9), 2)
	tx_44_0.Vout[10] = createValueStoreVoutTx(t, "6c65ed80f65e706ab7677682930ec3b8e3fc911f", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(10), 1)
	tx_44_0.Vout[11] = createValueStoreVoutTx(t, "beecd78bde305238959210c89030dd4b8d220a79", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(11), 2)
	tx_44_0.Vout[12] = createValueStoreVoutTx(t, "2e956e0d303438ab8e536b2e7bbbb14dd9518a37", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(12), 1)
	tx_44_0.Vout[13] = createValueStoreVoutTx(t, "9fef65820ba66f1130be8ae06a2bdaef290c1941", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(13), 2)
	tx_44_0.Vout[14] = createValueStoreVoutTx(t, "25f1e58bef4135077df4daa91262b2d0c975f1ad", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(14), 1)
	tx_44_0.Vout[15] = createValueStoreVoutTx(t, "fa2243091b65e5f8bce43e23649ba453475d4c63", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(15), 2)
	tx_44_0.Vout[16] = createValueStoreVoutTx(t, "f7b49cb77898b000bada8ca0dca0e5bb2ea07232", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(16), 1)
	tx_44_0.Vout[17] = createValueStoreVoutTx(t, "1a59a4fc1e76142f1755e2c811cfc7ab08592c70", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(17), 2)
	tx_44_0.Vout[18] = createValueStoreVoutTx(t, "a4d14903dd37fc461bbbcf2485eff4e7f45707da", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(18), 1)
	tx_44_0.Vout[19] = createValueStoreVoutTx(t, "4a8b24894a6d4ea816aa9000a7f6d8ada4b6fceb", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "5000000000000000000", uint32(19), 2)
	tx_44_0.Vout[20] = createValueStoreVoutTx(t, "546f99f244b7b58b855330ae0e2bc1b30b41302f", "a4bf43c0ac89a13de4e4a4a65dc106501291bcb57baba3b8326c9a158d8c7390", "999899999999999999955257", uint32(20), 1)
	err = tx_44_0.SetTxHash()
	require.Nil(t, err)

	err = db.Update(func(txn *badger.Txn) error {
		_, err := trie.ApplyState(txn, objs.TxVec{tx_44_0}, 44)
		return err
	})
	require.Nil(t, err)

	err = db.Update(func(txn *badger.Txn) error {
		utxoID, err := utils.DecodeHexString("fe032f675e56b4b0889ea0f9279735c0437d1c8ac736d0408c8ff67bd2e3e67a")
		require.Nil(t, err)
		_, missing, err := trie.Get(txn, [][]byte{utxoID})
		if len(missing) != 0 {
			return errors.New("not found")
		}
		return err
	})
	require.Nil(t, err)
}

func TestReplicateCorruptTrie(t *testing.T) {
	db := mocks.NewTestDB()
	trie := NewUTXOTrie(db.DB())

	for i := 1; i < 24; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(i))
			return err
		})
		require.Nil(t, err)
	}

	err := db.Update(func(txn *badger.Txn) error {
		addkeys := [][]byte{}
		addvalues := [][]byte{}

		addkeys = append(addkeys, decodeHexString(t, "0000000000000000000000000000000000000000000000000000000000000002"))
		addvalues = append(addvalues, decodeHexString(t, "4f5889db505a98d91600c0f9284de960a2c77987540195febadbd5a325763906"))

		addkeys = append(addkeys, decodeHexString(t, "028e494f106f90215d86b1a6fa149c77a41bf91e3a493f3a52538f06cccff47e"))
		addvalues = append(addvalues, decodeHexString(t, "0427401c02f11517275aebb28a185566775d52e86bbd49740841d5621cc67ded"))

		addkeys = append(addkeys, decodeHexString(t, "140189a557ffa75bb94c8faa00d199b9fac721b55a69a254017991e45dfcde91"))
		addvalues = append(addvalues, decodeHexString(t, "5aae93300431510e012b5085334e48e35ce5b4c3b3e5390ff1512f810204dd61"))

		addkeys = append(addkeys, decodeHexString(t, "1b88bf743110e7269a222336244fb7d93bbb211ffc8910484b78bcbbc4062e4b"))
		addvalues = append(addvalues, decodeHexString(t, "ae8027e86b45d73ffd835df96efe761a95c437969d8867e1dcd483251b54b9b4"))

		addkeys = append(addkeys, decodeHexString(t, "244b1e24e2b59139c4257fe304ecd5c2c86b80d4a13362616126b4d193eeae11"))
		addvalues = append(addvalues, decodeHexString(t, "bd1fcf5d3b455c9574dc818b00cbb72b7b2d2d0371a14fe0f14eb496f6f24ddc"))

		addkeys = append(addkeys, decodeHexString(t, "3913238275c4dce15872510b5fac65890e1b93f7df9cec4b20eb20f664b5bc66"))
		addvalues = append(addvalues, decodeHexString(t, "0fbb0bc74ef9b5dd060979128f5f94f6c8a1a1c193013d4ea0a5b054b6dbf73f"))

		addkeys = append(addkeys, decodeHexString(t, "481941ae5b1ff4dd3a1b7e17ee5e806b4c82830ccb87f85c886c84cd2525f744"))
		addvalues = append(addvalues, decodeHexString(t, "c0c433990d95a5e2dc38b8a92a5d85f56902a2d2d552288d6fde977336181b41"))

		addkeys = append(addkeys, decodeHexString(t, "4cd11f91ae66a5b677ce6ae43d0dd54199e9fd60d57eb43a93f92c3309736581"))
		addvalues = append(addvalues, decodeHexString(t, "fa4964d3726eaeb934d2af6873f198fa8ea7fe0caef57886caca6f9937127fb6"))

		addkeys = append(addkeys, decodeHexString(t, "68359fac40cda2256b66b0cbc8f8798a9d15f72b7a397416bb6c32cc7a914c7b"))
		addvalues = append(addvalues, decodeHexString(t, "ad9e4a2f1699684cda0ae409d95a44e71917a64e4fe08d1c13e098f80cacb70a"))

		addkeys = append(addkeys, decodeHexString(t, "6b9b82254d67c630b74cbeb6be33f3716b3076e31d137312201dff36d949ab5c"))
		addvalues = append(addvalues, decodeHexString(t, "71960cb03343341b17ae722b94da8766a5e2012e3f26cfffbb144d800bd5905f"))

		addkeys = append(addkeys, decodeHexString(t, "824d53e2abb97dee87a805b38e81ef933d72b907679d0a7e46702ebf70b5194c"))
		addvalues = append(addvalues, decodeHexString(t, "a296a0fab6544259d88f345a7d3063d09d90ab78ab011a6fb4b0057c730e1902"))

		addkeys = append(addkeys, decodeHexString(t, "852c8e1ef05fcce0bafb21923763a6b1793b83677f16b014d1a05fa1e6eeb4db"))
		addvalues = append(addvalues, decodeHexString(t, "584f78ae218436fcf0e01a31163a2de958a1ae1e238afe90f09f2f4947161055"))

		addkeys = append(addkeys, decodeHexString(t, "8d5270b53f9946d8129aa48647d96ff1dd3d2ebd20b2b6dd9883dc0e4b227a0d"))
		addvalues = append(addvalues, decodeHexString(t, "2d189a2a662aa633396a497aa73d51a31cbe2958082ef539dc6c8696424c54ce"))

		addkeys = append(addkeys, decodeHexString(t, "90bbab7c40edb5304b3197a64fb52a19653719dc2fc797625fc8a54d46113d9f"))
		addvalues = append(addvalues, decodeHexString(t, "62109b305cf965f1e59705c3e368c729c8351741169c0dc5007686248dcb37fb"))

		addkeys = append(addkeys, decodeHexString(t, "a469e42aabe2ecde16127c1a5599f7768c66874e4d6b4ad7cea1254753f311f7"))
		addvalues = append(addvalues, decodeHexString(t, "631e87fd35959f7c250d4a7a93f96db9cfa1bdcb78aaa6efee20407edb3b762c"))

		addkeys = append(addkeys, decodeHexString(t, "ba6d99c18a79d1af6b50a099458658422e2452ac238052ee4bdc917ab0af1814"))
		addvalues = append(addvalues, decodeHexString(t, "6065c454e501e3af0bed8536aedad1152afa14bdc791b1fb8d905728271983eb"))

		addkeys = append(addkeys, decodeHexString(t, "bc3044a4cbba136615f00857d66e0cd5e22ba9db55344de73a4897cdec57a220"))
		addvalues = append(addvalues, decodeHexString(t, "38a23f6290987b410f4d4ce565abbc0120b0a39445470780f2e5ff74bc4554b0"))

		addkeys = append(addkeys, decodeHexString(t, "bfb504b298739058c27104533236e62b7110daa706199a3b0b0f80d756fb5894"))
		addvalues = append(addvalues, decodeHexString(t, "28dea933894ade6b9e4d2d15f5a55021ff4f310c155e47decccdadcd5571aab4"))

		addkeys = append(addkeys, decodeHexString(t, "ca122f122489d5c2f77646ce79582842aa85d4cac222b00c6cbbcb6e8ef423fd"))
		addvalues = append(addvalues, decodeHexString(t, "f2ec427fe4c25f85e3308bafe94c9c24b96aa10e8b55e8b7bc8fe3a7f9248ce4"))

		addkeys = append(addkeys, decodeHexString(t, "cf7aa0e8e27867d128b5b61f2c942ddff90ae320e5361d127b8de0c17841cfed"))
		addvalues = append(addvalues, decodeHexString(t, "b054ddeb73e879c43b43c75ee58ea5014616415681ae5b04e5d699461362283d"))

		addkeys = append(addkeys, decodeHexString(t, "ee251e6bfbfbf3997026d08f01e617ccf82edc9361a03c895129543b3667528c"))
		addvalues = append(addvalues, decodeHexString(t, "3c0794d727dca8a23f719268cc7e97649ce2bc897955d49887e7a4ae454e953d"))

		addkeys = append(addkeys, decodeHexString(t, "fabbab9ab237a1dbc38b9203939c12b1d05539b6af8f14beb7dd2a58cfe86d62"))
		addvalues = append(addvalues, decodeHexString(t, "e7189139a3823eec4f43714bd487211ffee76cee361bf5f8260f8b25bf698522"))

		current, err := trie.GetCurrentTrie(txn)
		if err != nil {
			return err
		}

		stateRoot, err := current.Update(txn, addkeys, addvalues)
		if err != nil {
			return err
		}

		if err := setRootForHeight(txn, 24, stateRoot); err != nil {
			return err
		}
		if err := SetCurrentStateRoot(txn, stateRoot); err != nil {
			return err
		}
		_, err = current.Commit(txn, 24)
		if err != nil {
			return err
		}

		return nil
	})
	require.Nil(t, err)

	for i := 25; i < 29; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(i))
			return err
		})
		require.Nil(t, err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		addkeys := [][]byte{}
		addvalues := [][]byte{}

		addkeys = append(addkeys, decodeHexString(t, "028e494f106f90215d86b1a6fa149c77a41bf91e3a493f3a52538f06cccff47e"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "03aa98c9509e4e346ce52601e234d30c66920b68db29d85e16b2718b9684ec4f"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "0c3e622418e84e6deaa8ba89ef118a2953bf428289ecdf4a07156ccef7b9edd8"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "140189a557ffa75bb94c8faa00d199b9fac721b55a69a254017991e45dfcde91"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "16c750dabba88f7a40b8ad2e874e27176e93293944353bffb88bc5a931d23359"))
		addvalues = append(addvalues, decodeHexString(t, "db229dbf5866f2bf9d91d77d9dc8c8076b1c0162418f311c1f2a184640c83697"))

		addkeys = append(addkeys, decodeHexString(t, "18046eaae3c5ac7bcf516cecd66c91d93b7d353e47557d6692589bc07116865e"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "1b88bf743110e7269a222336244fb7d93bbb211ffc8910484b78bcbbc4062e4b"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "244b1e24e2b59139c4257fe304ecd5c2c86b80d4a13362616126b4d193eeae11"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "2f6fd4218dcdc8062cb8120622799a45d555ea10f85dfd5fcec5a93bff18511d"))
		addvalues = append(addvalues, decodeHexString(t, "e4e2f84c732b74a1e3ccde57fd8ced595c76c8ea1c8c2c2ef7c74e495323e4aa"))

		addkeys = append(addkeys, decodeHexString(t, "37ddd07605b7a38be323d887d4d71842fe48274480a5bd48f35aaed6af8eeb4c"))
		addvalues = append(addvalues, decodeHexString(t, "923cb351408b6104e4a86f847d0a3ddb5256fba6091d8bbf477dee7f62276068"))

		addkeys = append(addkeys, decodeHexString(t, "3913238275c4dce15872510b5fac65890e1b93f7df9cec4b20eb20f664b5bc66"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "3d0ddc83994b82f4a37a0fd56a28bfecb31e5f98b21597237b59dbef20c45f0d"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "481941ae5b1ff4dd3a1b7e17ee5e806b4c82830ccb87f85c886c84cd2525f744"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "4cd11f91ae66a5b677ce6ae43d0dd54199e9fd60d57eb43a93f92c3309736581"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "53dbd950edfc5c7baa6ad8d7025735c28b0158d2e623792352d9ca2c79041fcc"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "57d234e6cbcafce3b0e9dfee0b3cf0fe831241686482d8f4633dcec41f4d4e61"))
		addvalues = append(addvalues, decodeHexString(t, "02ac205eb8a372e915b0d067305158c6283b9a1b33cc8f5a89896b60ddb2001d"))

		addkeys = append(addkeys, decodeHexString(t, "61f4008720140193b83a947862ea7134c58688899178f17671170a0aa1074ae3"))
		addvalues = append(addvalues, decodeHexString(t, "41cbd27456f5f125fd15119372b2a4a46c7c03d964f2601a4c3b75fd2aa72687"))

		addkeys = append(addkeys, decodeHexString(t, "68359fac40cda2256b66b0cbc8f8798a9d15f72b7a397416bb6c32cc7a914c7b"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "6b9b82254d67c630b74cbeb6be33f3716b3076e31d137312201dff36d949ab5c"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "766cee4a0ffd0ce36eecf9cec622ab80863fcc4584fe40f179c744143f2beb62"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "7bb456481c1d20c26ba78a5d7615e39b40933e45a39bd07945c290256e1797d9"))
		addvalues = append(addvalues, decodeHexString(t, "53685a6bf3d40d6198ccf22bdc04dc41854d5ab14ba1b103648a2964192a4822"))

		addkeys = append(addkeys, decodeHexString(t, "80f615e14071f3911e88c5a47f55641cba240332edb2c308db6b2362ded6e25c"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "824d53e2abb97dee87a805b38e81ef933d72b907679d0a7e46702ebf70b5194c"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "8486e6cb1c68e517cb9207ec87b6648c23df4396ad6c03092abdfc89268a67b9"))
		addvalues = append(addvalues, decodeHexString(t, "d489d10af579d9557bad91c7fe6c369f89656b36f230d975a977e0f0aab19fbc"))

		addkeys = append(addkeys, decodeHexString(t, "852c8e1ef05fcce0bafb21923763a6b1793b83677f16b014d1a05fa1e6eeb4db"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "8d5270b53f9946d8129aa48647d96ff1dd3d2ebd20b2b6dd9883dc0e4b227a0d"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "90bbab7c40edb5304b3197a64fb52a19653719dc2fc797625fc8a54d46113d9f"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "91d91c86f8eec290a15d61e25868b99e818529e611c9426210c22aade57b6267"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "9678f115ffcb338fbf37d55f3b6417fde7317cad96df49528f889d3314bce817"))
		addvalues = append(addvalues, decodeHexString(t, "6f85507adfd2549927d9a4ab417f856c9d145e93e45284406934816f1f737933"))

		addkeys = append(addkeys, decodeHexString(t, "a0b6c0ee92e36141e3d61098ee248e3f94a454d755fe9f01593314228aaec600"))
		addvalues = append(addvalues, decodeHexString(t, "162d35b843acc9a761c7254675e0988994ff32129feb05fa477e8b09d0bc6482"))

		addkeys = append(addkeys, decodeHexString(t, "b1cfeaabd4b1d27063b2ca1f74caca07aa04c1c11f917c6210a68722a8215735"))
		addvalues = append(addvalues, decodeHexString(t, "b80e81f751b9fc2d1775a058087e5c9ebd62dc7be208b5024a7d3e2f21e732ed"))

		addkeys = append(addkeys, decodeHexString(t, "b8442b25f94569a40c2dc763d076ab05454913cc9347fd192ab08099ee933471"))
		addvalues = append(addvalues, decodeHexString(t, "d189292474494f108a5fc86f6f627ae77be66007df7353a70a65f6bc8ed2c136"))

		addkeys = append(addkeys, decodeHexString(t, "ba6d99c18a79d1af6b50a099458658422e2452ac238052ee4bdc917ab0af1814"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "bb94fc1a62829b893d78365405fdaed7e9b8e68c3a53f1d2ea0ecdb0f0b38d69"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "bc3044a4cbba136615f00857d66e0cd5e22ba9db55344de73a4897cdec57a220"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "bfb504b298739058c27104533236e62b7110daa706199a3b0b0f80d756fb5894"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "c0cce73a20860f3c9aba9967afb78e78cb293cd27e666546da9f748c4d6411cd"))
		addvalues = append(addvalues, decodeHexString(t, "bf836079cfc15dd7624de68f00f42d914618505d79d284b6211f21c0bb562ba4"))

		addkeys = append(addkeys, decodeHexString(t, "ca122f122489d5c2f77646ce79582842aa85d4cac222b00c6cbbcb6e8ef423fd"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "cb8520fadd30361c5c89140eccdc274ede9bd33c7374a8170c173459b807d0db"))
		addvalues = append(addvalues, decodeHexString(t, "baa902d6385f36efa359fe9acd422875f726b3f6cc6d832bdc0b54107ed0b445"))

		addkeys = append(addkeys, decodeHexString(t, "cf7aa0e8e27867d128b5b61f2c942ddff90ae320e5361d127b8de0c17841cfed"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "cfeab0d7ca71731dfa640163c3e24db602b89647f0220c6f5f0aa1de5a8dce10"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "d0ff52435c39d02c3844a3655a07fa74064fd4d8d4256a422a6922b1b88e3d2f"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "de398bc854201b51c9800227a1c3d9eb7d0c920224461139235bef4b6055dec1"))
		addvalues = append(addvalues, decodeHexString(t, "3f53c6e428a91ad06ade73008395a59906c98c3f8ada7e1598daa480affb7a3a"))

		addkeys = append(addkeys, decodeHexString(t, "ee251e6bfbfbf3997026d08f01e617ccf82edc9361a03c895129543b3667528c"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "fabbab9ab237a1dbc38b9203939c12b1d05539b6af8f14beb7dd2a58cfe86d62"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "fc0d14c8b904c7d5d32b9ce4cd44ab18c3eda4476429d44c900d0b9930b938f6"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "ff0d1a8d8ffec77dd5e2d13ffad0313fca2e99d1331801b281b4007bbe5f68c2"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		current, err := trie.GetCurrentTrie(txn)
		if err != nil {
			return err
		}

		stateRoot, err := current.Update(txn, addkeys, addvalues)
		if err != nil {
			return err
		}

		if err := setRootForHeight(txn, 29, stateRoot); err != nil {
			return err
		}
		if err := SetCurrentStateRoot(txn, stateRoot); err != nil {
			return err
		}
		_, err = current.Commit(txn, 29)
		if err != nil {
			return err
		}

		return nil
	})
	require.Nil(t, err)

	for i := 30; i < 33; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(i))
			return err
		})
		require.Nil(t, err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		addkeys := [][]byte{}
		addvalues := [][]byte{}

		addkeys = append(addkeys, decodeHexString(t, "03aa98c9509e4e346ce52601e234d30c66920b68db29d85e16b2718b9684ec4f"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "0c3e622418e84e6deaa8ba89ef118a2953bf428289ecdf4a07156ccef7b9edd8"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "18046eaae3c5ac7bcf516cecd66c91d93b7d353e47557d6692589bc07116865e"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "19b2e21a4895c925ed396a05000aced962f87e17e998d6f36f37bbc9d3d92eae"))
		addvalues = append(addvalues, decodeHexString(t, "9c0c0af5ac5348861cd0d555cc5b4d226954b65ecd8158ca68653f0690bba8dd"))

		addkeys = append(addkeys, decodeHexString(t, "263a850e681cce49b783cf4dd683f5a8eff8b2714553ddb23874a33ff2717f4d"))
		addvalues = append(addvalues, decodeHexString(t, "402b52166fc31f71f5bb472aaf3a63525691292342abee2000058e892c723d58"))

		addkeys = append(addkeys, decodeHexString(t, "2825b8ae9842ed6cfd878bf92e38001de8c9e57a10aa3401e2175356c220fe52"))
		addvalues = append(addvalues, decodeHexString(t, "51a42b3b7d23d3f605cb35320de61739a69ddaa73efad65fed8145ee37593bbb"))

		addkeys = append(addkeys, decodeHexString(t, "2a47d574da86e86bdd1901e50379bc681617ede0f1a65e69b1e191d080f7dae4"))
		addvalues = append(addvalues, decodeHexString(t, "39e3d02033d68ce416505614f9dbaae9e09b2569a414e633706b6dc994db7570"))

		addkeys = append(addkeys, decodeHexString(t, "2af2279160ab633fa42d5934e0cf3a2a7d6143827981f8a562ce6af2d2860cea"))
		addvalues = append(addvalues, decodeHexString(t, "28ae46636cb34fc3ec6ba10b616f9d7802bea1d0e06be9c97771a14a85fb2433"))

		addkeys = append(addkeys, decodeHexString(t, "2f6fd4218dcdc8062cb8120622799a45d555ea10f85dfd5fcec5a93bff18511d"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "310c80bd4d2dabf80e8cb330289bc89c9bb1aa90b19b22aad892064078ca1206"))
		addvalues = append(addvalues, decodeHexString(t, "2d56c6b4750f7bbea05c3c6076ee8faf0bfce0b390c4dd4c0e1f7cfa5a31e4fb"))

		addkeys = append(addkeys, decodeHexString(t, "37ddd07605b7a38be323d887d4d71842fe48274480a5bd48f35aaed6af8eeb4c"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "3c2cd2ca690e3081bd451ff1fe1a693d32bf34b9c4488eea49332776f78d343c"))
		addvalues = append(addvalues, decodeHexString(t, "f008c47ff012b1812cb14d65f7be6a539b3b19488f6e5b0fe507e741d9190b0e"))

		addkeys = append(addkeys, decodeHexString(t, "3d0ddc83994b82f4a37a0fd56a28bfecb31e5f98b21597237b59dbef20c45f0d"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "3e639b962fbcaa28d476ff64d954148753dcf682ae683c1a42934afbc6f0c47e"))
		addvalues = append(addvalues, decodeHexString(t, "1297b39fef9926d1238ca41a49db8ace247167a7d3b8d86842eff8dd3dffda16"))

		addkeys = append(addkeys, decodeHexString(t, "5281a54995869ed27171e94437cc1e6d0d6201d98e29a299a92a2a8fd3629d22"))
		addvalues = append(addvalues, decodeHexString(t, "599213c5278f8166593289b3017b4a4196aa15edd23d36ac61f63e92d50ebab6"))

		addkeys = append(addkeys, decodeHexString(t, "53dbd950edfc5c7baa6ad8d7025735c28b0158d2e623792352d9ca2c79041fcc"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "5e69b28b6da5d5b1ff05e3d111706846fdd2aa0aa7d7251111496c42e3035be8"))
		addvalues = append(addvalues, decodeHexString(t, "09f5c8141dc5699e6b8cd0011d3218a87df59b210a0f60342f233e1b22845829"))

		addkeys = append(addkeys, decodeHexString(t, "762f6adc57cadf1f028c3241bf9841f2543833145c7a0b7f85cf43546e22c475"))
		addvalues = append(addvalues, decodeHexString(t, "15327b9d0666fe217dc8cf81af0e9db6f943e2f9f1fad8bc17049ed790645304"))

		addkeys = append(addkeys, decodeHexString(t, "766cee4a0ffd0ce36eecf9cec622ab80863fcc4584fe40f179c744143f2beb62"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "80f615e14071f3911e88c5a47f55641cba240332edb2c308db6b2362ded6e25c"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "8486e6cb1c68e517cb9207ec87b6648c23df4396ad6c03092abdfc89268a67b9"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "91d91c86f8eec290a15d61e25868b99e818529e611c9426210c22aade57b6267"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "958ea68cd1fdad355b93c45fc1e3ea5dff13c2dd9c5ab3044a10cb9205184476"))
		addvalues = append(addvalues, decodeHexString(t, "4bb2ace5019b7f41ae7140b74f5a4c37aca1689617bf3c851abbc6ce51d3a253"))

		addkeys = append(addkeys, decodeHexString(t, "9595802baffc61fa32afbeb7706eaeb5b7def4ba3c0c7e87f6834f58bcd1f267"))
		addvalues = append(addvalues, decodeHexString(t, "bc7ee0f85a80b308ccd8b0244458f1bf94f92bc739320634e2ab98196d38b9f4"))

		addkeys = append(addkeys, decodeHexString(t, "9678f115ffcb338fbf37d55f3b6417fde7317cad96df49528f889d3314bce817"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "a469e42aabe2ecde16127c1a5599f7768c66874e4d6b4ad7cea1254753f311f7"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "b1cfeaabd4b1d27063b2ca1f74caca07aa04c1c11f917c6210a68722a8215735"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "b4d52fb831a2815154ca59dbef7f71d08a9a02290e36211032657386c7956e2f"))
		addvalues = append(addvalues, decodeHexString(t, "61daed6aebc65a42f449014c15c4cb5e9371243292d8b048365d30253e1a4146"))

		addkeys = append(addkeys, decodeHexString(t, "b7aea2b41e9b84847da87a3407b09a4c454da03d9afbccf95c2b72e86fd3d4ed"))
		addvalues = append(addvalues, decodeHexString(t, "c076b38bc29ab16ae8e7c290d3ab12c8d1f3ed94e463a4d0067413d211df675f"))

		addkeys = append(addkeys, decodeHexString(t, "b8442b25f94569a40c2dc763d076ab05454913cc9347fd192ab08099ee933471"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "bb94fc1a62829b893d78365405fdaed7e9b8e68c3a53f1d2ea0ecdb0f0b38d69"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "bca3223a5181035fe631cc1546e9f30e3ac3ad63db8dffd8a6b2e07c509ea5f8"))
		addvalues = append(addvalues, decodeHexString(t, "e77df29bd2ff5158c329edc07b47ec1227fbb76f6ea53e6efbf9007080ba259a"))

		addkeys = append(addkeys, decodeHexString(t, "cb8520fadd30361c5c89140eccdc274ede9bd33c7374a8170c173459b807d0db"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "ccb7985389122163b5a16b744ce16e7a8fcd5cdd2db1f1858024debb727fe0b8"))
		addvalues = append(addvalues, decodeHexString(t, "714ae1aefb7f71840745073f12ca961acc86c1382259fbd711e2009647e50265"))

		addkeys = append(addkeys, decodeHexString(t, "cfeab0d7ca71731dfa640163c3e24db602b89647f0220c6f5f0aa1de5a8dce10"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "d0ff52435c39d02c3844a3655a07fa74064fd4d8d4256a422a6922b1b88e3d2f"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "d7d705a683d323c3a51cf6a7dcf0fe575a5e2d929bc0086a69f42cece83f5259"))
		addvalues = append(addvalues, decodeHexString(t, "809f28936f88776833fa7dfb3e98a57673c11d909e6f70b50966cf3a7d036a7b"))

		addkeys = append(addkeys, decodeHexString(t, "e15d5cdc48f1effbd9fc03c33fe18204fca09fa2fa5c5fda8cfa3dcddc0a45e0"))
		addvalues = append(addvalues, decodeHexString(t, "cf4725fd3707d80bad405584667ce231d95fe8c26710cc0ed9ad3595392ee2d3"))

		addkeys = append(addkeys, decodeHexString(t, "e462c88d614f6d94b4a5a2026deee6bb5f183842f6132cfe7592d603b7b0b87d"))
		addvalues = append(addvalues, decodeHexString(t, "f4f3e02a7ba21689204b6fcc2ef1dca24dad5619b5dc68fb90c7fbb67ec66b88"))

		addkeys = append(addkeys, decodeHexString(t, "f3bf49d5f116e995678d651508cdd688a396c381caea0830af3f33cbe3f4e827"))
		addvalues = append(addvalues, decodeHexString(t, "5024b0227fc53ab634a3cd009435492a3ad212ed6c8db3dafc7c8ca1e44e53de"))

		addkeys = append(addkeys, decodeHexString(t, "fc0d14c8b904c7d5d32b9ce4cd44ab18c3eda4476429d44c900d0b9930b938f6"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "ff0d1a8d8ffec77dd5e2d13ffad0313fca2e99d1331801b281b4007bbe5f68c2"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		current, err := trie.GetCurrentTrie(txn)
		if err != nil {
			return err
		}

		stateRoot, err := current.Update(txn, addkeys, addvalues)
		if err != nil {
			return err
		}

		if err := setRootForHeight(txn, 33, stateRoot); err != nil {
			return err
		}
		if err := SetCurrentStateRoot(txn, stateRoot); err != nil {
			return err
		}
		_, err = current.Commit(txn, 33)
		if err != nil {
			return err
		}

		return nil
	})
	require.Nil(t, err)

	for i := 34; i < 39; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(i))
			return err
		})
		require.Nil(t, err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		addkeys := [][]byte{}
		addvalues := [][]byte{}

		addkeys = append(addkeys, decodeHexString(t, "0111be2cb7f44efd0bb1d6af63244b91804c0c241eeb6e6d8f4cc90413f968c9"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "022894f05601e2a53f5a2cbba398b57c38dd504b924bec1be1e928b49fca0171"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "08cf04f2398dff64e725425bf90c42fc26dd96f943620c91af839987ec22714b"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "19b2e21a4895c925ed396a05000aced962f87e17e998d6f36f37bbc9d3d92eae"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "21175bba1dd338187636b7fd8786780dd769e51fbb92f698653d6b8c9d1e0f47"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "263a850e681cce49b783cf4dd683f5a8eff8b2714553ddb23874a33ff2717f4d"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "2825b8ae9842ed6cfd878bf92e38001de8c9e57a10aa3401e2175356c220fe52"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "2a47d574da86e86bdd1901e50379bc681617ede0f1a65e69b1e191d080f7dae4"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "2af2279160ab633fa42d5934e0cf3a2a7d6143827981f8a562ce6af2d2860cea"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "310c80bd4d2dabf80e8cb330289bc89c9bb1aa90b19b22aad892064078ca1206"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "3c2cd2ca690e3081bd451ff1fe1a693d32bf34b9c4488eea49332776f78d343c"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "3e639b962fbcaa28d476ff64d954148753dcf682ae683c1a42934afbc6f0c47e"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "43ac92d4742b31c0eb0e742b4dd9b3538e26487b19e5a35e063b683631f2cb16"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "500059d2c413c249a6548abb23016966b5ba030f30f6062140b0d3e736934430"))
		addvalues = append(addvalues, decodeHexString(t, "1428bf583a873607c40f2ada13ce7a592d4cb3b920270518fb38b503960a02c0"))

		addkeys = append(addkeys, decodeHexString(t, "511a4eac18ddcffecd2b98fad9c282c094cc9501c0d18fb448387208595a34af"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "595219ef9fc50525a458cffc7805361b2bff0dba1589474c327ed66f6c210be9"))
		addvalues = append(addvalues, decodeHexString(t, "62780965cc64d3a005b596a82dd1a8d925d71c33c1bfd08e7e9bbb048a0f32f3"))

		addkeys = append(addkeys, decodeHexString(t, "5e69b28b6da5d5b1ff05e3d111706846fdd2aa0aa7d7251111496c42e3035be8"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "6179d3c843235a771d196d4fa7fd02656e0944dd85e3c24b2f3ddb30711537bb"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "6d7b421db0a677a8a10ddf8d54bbfaeada9dd9074e9079d8acdc3b76497e6aa2"))
		addvalues = append(addvalues, decodeHexString(t, "cd315c8d51278f91548ec7f49d3793f1ca801a325fb155ab09592b3ecddc94e2"))

		addkeys = append(addkeys, decodeHexString(t, "762f6adc57cadf1f028c3241bf9841f2543833145c7a0b7f85cf43546e22c475"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "78ac23c04645b46309589d8f78c7f747d52b59136cc634bc0cd332c2bdaf5c6f"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "7ab1455d169c666be8ea6646ab0364bebdb22c6ebc825d7d09da2c37a605b454"))
		addvalues = append(addvalues, decodeHexString(t, "4fd14f80ac3d191b63b94e186f8dee8bafc22f9badc9bdacc3825bfef103d277"))

		addkeys = append(addkeys, decodeHexString(t, "8c3dbf9efed235b2d9010546ebb4f232dfab1e09d02ad87d755f3d1bd23d402d"))
		addvalues = append(addvalues, decodeHexString(t, "923cb351408b6104e4a86f847d0a3ddb5256fba6091d8bbf477dee7f62276068"))

		addkeys = append(addkeys, decodeHexString(t, "91039ef0e0e1b4bd51e38622c28f18f80ab573c719784631868b7b85e5614d14"))
		addvalues = append(addvalues, decodeHexString(t, "12dd09db647feb91f32a093c308564428a00f6bc0bbd43850de24e5151a7271c"))

		addkeys = append(addkeys, decodeHexString(t, "958ea68cd1fdad355b93c45fc1e3ea5dff13c2dd9c5ab3044a10cb9205184476"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "9595802baffc61fa32afbeb7706eaeb5b7def4ba3c0c7e87f6834f58bcd1f267"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "9bbdcc3b251d04793d77fbca6da8de06d33b8ae095b43bec4d43b20d4ba2acd4"))
		addvalues = append(addvalues, decodeHexString(t, "3135e14969b5670b77b2004b181610984d28efa6c8fe28e71bd1f8120a823375"))

		addkeys = append(addkeys, decodeHexString(t, "9eec33d131daa7a1ab29134746f4d1b6514fe0eff74a069c2fee7e35829df2f0"))
		addvalues = append(addvalues, decodeHexString(t, "b818d3be9e328e9c94222e46baaffc062e39330ff3ca753d633fd3081ef9f96f"))

		addkeys = append(addkeys, decodeHexString(t, "a6be7164c8fb5a17f6f42eb7911b0192bfae29d71d96840de7144357f3ccd985"))
		addvalues = append(addvalues, decodeHexString(t, "1fd70ea0b2067adc01593ce1a00438add0271924b2ab76b7ae14300b80b10c57"))

		addkeys = append(addkeys, decodeHexString(t, "b122c932ab06fb3c17a2295c290159b26a33a204cc1e7a1fe71e6bd782e28ac0"))
		addvalues = append(addvalues, decodeHexString(t, "ede0d97e850a4e223ba3e73cfc0a4fa247447d0b9a628f5a6729b5f497215859"))

		addkeys = append(addkeys, decodeHexString(t, "b4d52fb831a2815154ca59dbef7f71d08a9a02290e36211032657386c7956e2f"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "b7aea2b41e9b84847da87a3407b09a4c454da03d9afbccf95c2b72e86fd3d4ed"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "b99972b45505254c5653356c82a8fe2190e220219061a0e5af700bd14f453fae"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "bca3223a5181035fe631cc1546e9f30e3ac3ad63db8dffd8a6b2e07c509ea5f8"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "c8e251171a45c7c7abec92e367d70c48a7680fabf1f6a7e8a8ed4917f1b387b5"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "ca4f07f15d0174b1a43edcb2ac88443f014dbb77aa3ea97f2d380eb7a571e194"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "ccb7985389122163b5a16b744ce16e7a8fcd5cdd2db1f1858024debb727fe0b8"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "d7d705a683d323c3a51cf6a7dcf0fe575a5e2d929bc0086a69f42cece83f5259"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "e0977ffc8bed16e419af04050304e2a73ec2a03ceeb7feb4c66a99cb29c45b31"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "e15d5cdc48f1effbd9fc03c33fe18204fca09fa2fa5c5fda8cfa3dcddc0a45e0"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "e462c88d614f6d94b4a5a2026deee6bb5f183842f6132cfe7592d603b7b0b87d"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "f26784ebae43c82fb79ffa8e2802a35232e9d45fe0855ea8220da471dc115b3d"))
		addvalues = append(addvalues, decodeHexString(t, "f56c97ab252feb5a956e7dc2677907f138a31d8ff96b799226a08c71a792087c"))

		addkeys = append(addkeys, decodeHexString(t, "f3bf49d5f116e995678d651508cdd688a396c381caea0830af3f33cbe3f4e827"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "f6c89266dd46b56d58a976cdc8144196fe5df27839730925125261b440f6b3bc"))
		addvalues = append(addvalues, decodeHexString(t, "a1d3e42b3286e57c66d9788338eaaf692ba827e2526e90374c35cbc4457a39ea"))

		addkeys = append(addkeys, decodeHexString(t, "f7050844f3999eda71673d8f01a2e72aab849f431a616f2e1eefcc7dca065734"))
		addvalues = append(addvalues, decodeHexString(t, "d293b522eeae4876e4f2fbe4ac76f1764fc75fea016cf202369b6e2d54281b8d"))

		addkeys = append(addkeys, decodeHexString(t, "f8ed08f75871c4d633bc104adb8576f34ab1abdf8f390b3e01dc39c85e3d0e54"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		addkeys = append(addkeys, decodeHexString(t, "f9961deb68062674d5cbf1af1761d0d5f977035d0167cb3f997618a8545fb3e1"))
		addvalues = append(addvalues, decodeHexString(t, "ee149e50ec684b36b35e19e214cbc6493c0d0431226ff59b3d85c21d3d9f7b5b"))

		current, err := trie.GetCurrentTrie(txn)
		if err != nil {
			return err
		}

		stateRoot, err := current.Update(txn, addkeys, addvalues)
		if err != nil {
			return err
		}

		if err := setRootForHeight(txn, 39, stateRoot); err != nil {
			return err
		}
		if err := SetCurrentStateRoot(txn, stateRoot); err != nil {
			return err
		}
		_, err = current.Commit(txn, 39)
		if err != nil {
			return err
		}

		return nil
	})
	require.Nil(t, err)

	for i := 40; i < 44; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(i))
			return err
		})
		require.Nil(t, err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		addkeys := [][]byte{}
		addvalues := [][]byte{}

		addkeys = append(addkeys, decodeHexString(t, "0111be2cb7f44efd0bb1d6af63244b91804c0c241eeb6e6d8f4cc90413f968c9"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "022894f05601e2a53f5a2cbba398b57c38dd504b924bec1be1e928b49fca0171"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "062c43f3a083260cc84e63c76b5f7570dbd48b1471c98b6b8b4ec3c781190468"))
		addvalues = append(addvalues, decodeHexString(t, "cd20c3347abf099b31fa5b4f1af49e817b7904412443c42a70cbb58b7d5cc58d"))

		addkeys = append(addkeys, decodeHexString(t, "08cf04f2398dff64e725425bf90c42fc26dd96f943620c91af839987ec22714b"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "0cbcba5844ddb5b88b081f2d3ff5b28e7f89889ad7b036949e01389983fe20df"))
		addvalues = append(addvalues, decodeHexString(t, "66bf9b8f15360f28ea1e2d6c331f4c691e41021cb1151239a3d0676ec5914078"))

		addkeys = append(addkeys, decodeHexString(t, "21175bba1dd338187636b7fd8786780dd769e51fbb92f698653d6b8c9d1e0f47"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "228df16570a5fc7ff06d1db7b86ba0fe5b185765234baaf5d757fadc235bb9d4"))
		addvalues = append(addvalues, decodeHexString(t, "393f4a8ff5b73cc9250128788a79604a4e74afdcb1c428f254728b20cf9a0874"))

		addkeys = append(addkeys, decodeHexString(t, "32f23092c196810ff8be8cf7b27bcc513c54910d9e97cc7f11b5ea0d36526a82"))
		addvalues = append(addvalues, decodeHexString(t, "b7fba14877cc46ba69f0993333a167c7010d077d3ba3f7fe967a8ae0174f7103"))

		addkeys = append(addkeys, decodeHexString(t, "3e4e4a81d6fb4c5349dd5bb26e85db7e31889919a76eb4cb7eeb95cd75c558c4"))
		addvalues = append(addvalues, decodeHexString(t, "179cd704ddb9395d6f19405292225a0eadc13bab8a5d01c5e6531234d5d5f8bc"))

		addkeys = append(addkeys, decodeHexString(t, "43ac92d4742b31c0eb0e742b4dd9b3538e26487b19e5a35e063b683631f2cb16"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "49e16d91098f6f46a690b94e4962043e38b62d71678c510a9f7d874c0aecf271"))
		addvalues = append(addvalues, decodeHexString(t, "02ecb4a249c513891f22a4f2ac34c1fba18e153e01da6f1637332365f9329976"))

		addkeys = append(addkeys, decodeHexString(t, "500059d2c413c249a6548abb23016966b5ba030f30f6062140b0d3e736934430"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "511a4eac18ddcffecd2b98fad9c282c094cc9501c0d18fb448387208595a34af"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "5281a54995869ed27171e94437cc1e6d0d6201d98e29a299a92a2a8fd3629d22"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "595219ef9fc50525a458cffc7805361b2bff0dba1589474c327ed66f6c210be9"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "59ca80c3c03e35426ae7510997b4a2886ceaa8cd6f0f2bfadc05d20133ca8e52"))
		addvalues = append(addvalues, decodeHexString(t, "ae85d8406402b6fecf0b0a042a446f4700c8dfeafd7182ee8293a125a0a4a553"))

		addkeys = append(addkeys, decodeHexString(t, "6179d3c843235a771d196d4fa7fd02656e0944dd85e3c24b2f3ddb30711537bb"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "6faf77116ae27ee954d5e7c3a7a7ba6d2342d61145fa03c9887ba50efe6a9134"))
		addvalues = append(addvalues, decodeHexString(t, "7f87dcf6d2d0fd61bc74bd72d4190029e348a8ce8be59f21b65db366630aa2fd"))

		addkeys = append(addkeys, decodeHexString(t, "71a4d8ff7c6d2d4c4a8ed80d59033339f046b5f3d116d85d80fd30039bb6c005"))
		addvalues = append(addvalues, decodeHexString(t, "cb34106f2b8b7909e7f9345052f909a7eada7b1dd1eaa4cd9c13bf98202a08ba"))

		addkeys = append(addkeys, decodeHexString(t, "78ac23c04645b46309589d8f78c7f747d52b59136cc634bc0cd332c2bdaf5c6f"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "857fb0c34ba67c742470684c56516b5d1060d9aa41cdd18dbe38f0b39710da5b"))
		addvalues = append(addvalues, decodeHexString(t, "7795bfad19fc77cf33eb8eb738fb924df19dd87d39d342c1063c296b95b32f2c"))

		addkeys = append(addkeys, decodeHexString(t, "88c696e5b897ff374bb047985843b957307967a75fa61d108cc76675fb398a18"))
		addvalues = append(addvalues, decodeHexString(t, "f9fac21fe8e8fec5192708731a34a894b6d9cbd10c947bf6c2c5da8c9b6167a8"))

		addkeys = append(addkeys, decodeHexString(t, "88e1047220c0ff3ef98255b932ed0ac666a8613e2779361e97cc5c2823ee9117"))
		addvalues = append(addvalues, decodeHexString(t, "aefdaaacb6eadf3ead5143fa3a62f34046ae54cfb729917891d9922b8880b3dc"))

		addkeys = append(addkeys, decodeHexString(t, "8c3dbf9efed235b2d9010546ebb4f232dfab1e09d02ad87d755f3d1bd23d402d"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "91039ef0e0e1b4bd51e38622c28f18f80ab573c719784631868b7b85e5614d14"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "9348e8e4b3107afd330c188e1436ff8ef45831cb727d8c621f3e8c4c81006b74"))
		addvalues = append(addvalues, decodeHexString(t, "15631b7e9f32f41aff0bc1807aa3596ef77cb9b0fade6c856f51ee8873b96d80"))

		addkeys = append(addkeys, decodeHexString(t, "a6be7164c8fb5a17f6f42eb7911b0192bfae29d71d96840de7144357f3ccd985"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "b99972b45505254c5653356c82a8fe2190e220219061a0e5af700bd14f453fae"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "bee35168fc988745c79d6dc276c64d4cd4a35f0fd6456cf5d45c685a4dde878b"))
		addvalues = append(addvalues, decodeHexString(t, "09436664de572f6f9ad7d5742a8c42787431ded1d3e84b983428d3bda24e0ed5"))

		addkeys = append(addkeys, decodeHexString(t, "c8e251171a45c7c7abec92e367d70c48a7680fabf1f6a7e8a8ed4917f1b387b5"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "ca4f07f15d0174b1a43edcb2ac88443f014dbb77aa3ea97f2d380eb7a571e194"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "ccd213505dc4b5a27114b2f2f7d342d78680d4609e579c90018303d303b4128a"))
		addvalues = append(addvalues, decodeHexString(t, "a7790e3224714c1a69149425424d325950244326a5bdce0549aa0e08a89eeed9"))

		addkeys = append(addkeys, decodeHexString(t, "d5ad8e33d3f651348e601aa123615dbfc850fda244f031d257cc006653e7c4f7"))
		addvalues = append(addvalues, decodeHexString(t, "b88f30d1fef07d65fbde0c4f1705d962d2d7fef5a59a33f2aa46b8bbf498d469"))

		addkeys = append(addkeys, decodeHexString(t, "d6753121ce34abe24a5002ac3fcdf65e6eb0a5623794c04eabf517e8dbe44c10"))
		addvalues = append(addvalues, decodeHexString(t, "4801ce61cfde0b7e2534ffaf0c562113052616ca733359da7aa05c655d07d876"))

		addkeys = append(addkeys, decodeHexString(t, "e0977ffc8bed16e419af04050304e2a73ec2a03ceeb7feb4c66a99cb29c45b31"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "f00d5d38a6cb2ce40d530b0c774543138c53bdf10397ff1fbbc4706bbac06f30"))
		addvalues = append(addvalues, decodeHexString(t, "f5a9062a3d16cc14b21edc2850535cd5f5c9d1e802016be8017217741124a4d4"))

		addkeys = append(addkeys, decodeHexString(t, "f26784ebae43c82fb79ffa8e2802a35232e9d45fe0855ea8220da471dc115b3d"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "f83e8e24a545040fa11df16b2a0397a4012ca721a5cb945fcfe0e5ef47b4f84e"))
		addvalues = append(addvalues, decodeHexString(t, "37904adc8acb466aa5345b17fdd1a126f71035a62b836f84b841a9b4bc29c9cc"))

		addkeys = append(addkeys, decodeHexString(t, "f892433b68b33e7e544fe5a703a1a1325f36f6101adf10efe3d3d6fbee2879b9"))
		addvalues = append(addvalues, decodeHexString(t, "f36013cb01b35a05416fd242aa53f1c8bc2a424851a778847c5daf581a31c9aa"))

		addkeys = append(addkeys, decodeHexString(t, "f8ed08f75871c4d633bc104adb8576f34ab1abdf8f390b3e01dc39c85e3d0e54"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "f9961deb68062674d5cbf1af1761d0d5f977035d0167cb3f997618a8545fb3e1"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "fe032f675e56b4b0889ea0f9279735c0437d1c8ac736d0408c8ff67bd2e3e67a"))
		addvalues = append(addvalues, decodeHexString(t, "f2554f06a804d367b4682726b1b96f81c6bedf0e4994d7babd680036060d8d2b"))

		current, err := trie.GetCurrentTrie(txn)
		if err != nil {
			return err
		}

		stateRoot, err := current.Update(txn, addkeys, addvalues)
		if err != nil {
			return err
		}

		if err := setRootForHeight(txn, 44, stateRoot); err != nil {
			return err
		}
		if err := SetCurrentStateRoot(txn, stateRoot); err != nil {
			return err
		}
		_, err = current.Commit(txn, 44)
		if err != nil {
			return err
		}

		return nil
	})
	require.Nil(t, err)

	err = db.Update(func(txn *badger.Txn) error {
		utxoID, err := utils.DecodeHexString("fe032f675e56b4b0889ea0f9279735c0437d1c8ac736d0408c8ff67bd2e3e67a")
		require.Nil(t, err)
		_, missing, err := trie.Get(txn, [][]byte{utxoID})
		if len(missing) != 0 {
			return errors.New("not found")
		}
		return err
	})
	require.Nil(t, err)
}

func TestReplicateCorruptTrieSimplified(t *testing.T) {
	db := mocks.NewTestDB()
	trie := NewUTXOTrie(db.DB())

	for i := 1; i < 33; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(i))
			return err
		})
		require.Nil(t, err)
	}

	err := db.Update(func(txn *badger.Txn) error {
		addkeys := [][]byte{}
		addvalues := [][]byte{}

		addkeys = append(addkeys, decodeHexString(t, "3bf49d5f116e995678d651508cdd688a396c381caea0830af3f33cbe3f4e8270"))
		addvalues = append(addvalues, decodeHexString(t, "5024b0227fc53ab634a3cd009435492a3ad212ed6c8db3dafc7c8ca1e44e53de"))

		current, err := trie.GetCurrentTrie(txn)
		if err != nil {
			return err
		}

		stateRoot, err := current.Update(txn, addkeys, addvalues)
		if err != nil {
			return err
		}

		if err := setRootForHeight(txn, 33, stateRoot); err != nil {
			return err
		}
		if err := SetCurrentStateRoot(txn, stateRoot); err != nil {
			return err
		}
		_, err = current.Commit(txn, 33)
		if err != nil {
			return err
		}

		return nil
	})
	require.Nil(t, err)

	for i := 34; i < 39; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(i))
			return err
		})
		require.Nil(t, err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		addkeys := [][]byte{}
		addvalues := [][]byte{}

		addkeys = append(addkeys, decodeHexString(t, "26784ebae43c82fb79ffa8e2802a35232e9d45fe0855ea8220da471dc115b3d0"))
		addvalues = append(addvalues, decodeHexString(t, "f56c97ab252feb5a956e7dc2677907f138a31d8ff96b799226a08c71a792087c"))

		addkeys = append(addkeys, decodeHexString(t, "3bf49d5f116e995678d651508cdd688a396c381caea0830af3f33cbe3f4e8270"))
		addvalues = append(addvalues, decodeHexString(t, "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"))

		addkeys = append(addkeys, decodeHexString(t, "8ed08f75871c4d633bc104adb8576f34ab1abdf8f390b3e01dc39c85e3d0e540"))
		addvalues = append(addvalues, decodeHexString(t, "10dc1043aec198fd7c47d853399f9fe51885e45f4f02095e6c3f27c9e3cd3f78"))

		current, err := trie.GetCurrentTrie(txn)
		if err != nil {
			return err
		}

		stateRoot, err := current.Update(txn, addkeys, addvalues)
		if err != nil {
			return err
		}

		if err := setRootForHeight(txn, 39, stateRoot); err != nil {
			return err
		}
		if err := SetCurrentStateRoot(txn, stateRoot); err != nil {
			return err
		}
		_, err = current.Commit(txn, 39)
		if err != nil {
			return err
		}

		return nil
	})
	require.Nil(t, err)

	for i := 40; i < 44; i++ {
		err := db.Update(func(txn *badger.Txn) error {
			_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(i))
			return err
		})
		require.Nil(t, err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		addkeys := [][]byte{}
		addvalues := [][]byte{}

		addkeys = append(addkeys, decodeHexString(t, "a002f675e56b4b0889ea0f9279735c0437d1c8ac736d0408c8ff67bd2e3e67a0"))
		addvalues = append(addvalues, decodeHexString(t, "f2554f06a804d367b4682726b1b96f81c6bedf0e4994d7babd680036060d8d2b"))

		current, err := trie.GetCurrentTrie(txn)
		if err != nil {
			return err
		}

		stateRoot, err := current.Update(txn, addkeys, addvalues)
		if err != nil {
			return err
		}

		if err := setRootForHeight(txn, 44, stateRoot); err != nil {
			return err
		}
		if err := SetCurrentStateRoot(txn, stateRoot); err != nil {
			return err
		}
		_, err = current.Commit(txn, 44)
		if err != nil {
			return err
		}

		return nil
	})
	require.Nil(t, err)

	err = db.Update(func(txn *badger.Txn) error {
		utxoID, err := utils.DecodeHexString("a002f675e56b4b0889ea0f9279735c0437d1c8ac736d0408c8ff67bd2e3e67a0")
		require.Nil(t, err)
		_, missing, err := trie.Get(txn, [][]byte{utxoID})
		if len(missing) != 0 {
			return fmt.Errorf("not found: %s", "a002f675e56b4b0889ea0f9279735c0437d1c8ac736d0408c8ff67bd2e3e67a0")
		}
		return err
	})
	require.Nil(t, err)
}

func decodeHexString(t *testing.T, value string) []byte {
	result, err := utils.DecodeHexString(value)
	require.Nil(t, err)
	return utils.CopySlice(result)
}

func createVinTx(t *testing.T, sign, hash, consumedTxHash string, consumedTxIdx uint32) *objs.TXIn {
	chainID := uint32(1337)
	hexSign, err := utils.DecodeHexString(sign)
	require.Nil(t, err)
	hexHash, err := utils.DecodeHexString(hash)
	require.Nil(t, err)
	hexConsumedTxHash, err := utils.DecodeHexString(consumedTxHash)
	require.Nil(t, err)
	return &objs.TXIn{
		Signature: hexSign,
		TXInLinker: &objs.TXInLinker{
			TxHash: hexHash,
			TXInPreImage: &objs.TXInPreImage{
				ChainID:        chainID,
				ConsumedTxIdx:  consumedTxIdx,
				ConsumedTxHash: hexConsumedTxHash,
			},
		},
	}
}

func createValueStoreVoutTx(t *testing.T, account, txHash, value string, tXOutIdx uint32, curveSpec constants.CurveSpec) *objs.TXOut {
	chainID := uint32(1337)
	txOut := &objs.TXOut{}

	hexAccount, err := utils.DecodeHexString(account)
	require.Nil(t, err)
	hexTxHash, err := utils.DecodeHexString(txHash)
	require.Nil(t, err)
	bigintValue, ok := new(big.Int).SetString(value, 10)
	require.True(t, ok)
	uint256Value, err := new(uint256.Uint256).FromBigInt(bigintValue)
	require.Nil(t, err)
	newValueStore := &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			ChainID: chainID,
			Value:   uint256Value,
			Owner: &objs.ValueStoreOwner{
				SVA:       1,
				CurveSpec: curveSpec,
				Account:   hexAccount,
			},
			TXOutIdx: tXOutIdx,
			Fee:      uint256.Zero(),
		},
		TxHash: hexTxHash,
	}

	err = txOut.NewValueStore(newValueStore)
	require.Nil(t, err)

	return txOut
}

func createDataStoreVoutTx(t *testing.T, account, txHash, index, rawData, value string, IssuedAt, tXOutIdx uint32, curveSpec constants.CurveSpec, signature string) *objs.TXOut {
	chainID := uint32(1337)
	txOut := &objs.TXOut{}

	hexAccount, err := utils.DecodeHexString(account)
	require.Nil(t, err)
	hexTxHash, err := utils.DecodeHexString(txHash)
	require.Nil(t, err)
	hexIndex, err := utils.DecodeHexString(index)
	require.Nil(t, err)
	hexRawData, err := utils.DecodeHexString(rawData)
	require.Nil(t, err)
	bigintValue, ok := new(big.Int).SetString(value, 10)
	require.True(t, ok)
	uint256Value, err := new(uint256.Uint256).FromBigInt(bigintValue)
	require.Nil(t, err)
	hexSignature, err := utils.DecodeHexString(signature)
	require.Nil(t, err)

	newDataStore := &objs.DataStore{
		DSLinker: &objs.DSLinker{
			DSPreImage: &objs.DSPreImage{
				ChainID:  chainID,
				Index:    hexIndex,
				IssuedAt: IssuedAt,
				Deposit:  uint256Value,
				RawData:  hexRawData,
				TXOutIdx: tXOutIdx,
				Owner: &objs.DataStoreOwner{
					SVA:       3,
					CurveSpec: curveSpec,
					Account:   hexAccount,
				},
				Fee: uint256.Zero(),
			},
			TxHash: hexTxHash,
		},
		Signature: &objs.DataStoreSignature{
			SVA:       3,
			CurveSpec: curveSpec,
			Signature: hexSignature,
		},
	}

	err = txOut.NewDataStore(newDataStore)
	require.Nil(t, err)

	return txOut
}
