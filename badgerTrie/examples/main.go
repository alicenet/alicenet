package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	trie "github.com/MadBase/MadNet/badgerTrie"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/crypto"
)

func hash(things ...[]byte) []byte {
	var data []byte
	for i := 0; i < len(things); i++ {
		data = append(data, things[i]...)
	}
	hf := crypto.Keccak256Hash(data)
	return hf.Bytes()
}

// should print out the hash of a leaf, the complete merkle proof, and the hash value of the root
// the values printed out can currently be pasted into the proof_verifier solidity file to verify that the proofs are correct
// for now the solidity contract could be deployed on Remix, but should be able to call it automatically eventually
func getValuesforVerifier(membershipProof bool) {

	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	smt := trie.NewSMT(nil, trie.Hasher, nil)

	var keys, values [][]byte
	bs := make([]byte, 32)
	binary.BigEndian.PutUint32(bs, 0)
	keys = [][]byte{bs}
	bs = make([]byte, 32)
	bs[31] = 1
	keys = append(keys, bs)
	unaddedKey := make([]byte, 32)
	unaddedKey[31] = 5
	values = [][]byte{hash([]byte("thirtythree")), hash([]byte("test"))}

	fmt.Println("updating the tree with keys and values")
	fn := func(txn *badger.Txn) error {
		// Add data to empty trie
		smt.Update(txn, keys, values)

		var mp [][]byte
		var bitset []byte
		var height int
		var included bool
		var rkey, rval []byte

		// this allows for testing the two different cases: proof of membership or non membership
		if membershipProof {
			fmt.Println("hash", hex.EncodeToString(values[0]))
			bitset, mp, height, included, rkey, rval, err = smt.MerkleProofCompressed(txn, keys[0])
			if err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Println("hash", hex.EncodeToString(trie.DefaultLeaf))
			bitset, mp, height, included, rkey, rval, err = smt.MerkleProofCompressed(txn, unaddedKey)
			if err != nil {
				log.Fatal(err)
			}
		}

		fmt.Printf("d leaf %x\n", trie.DefaultLeaf)
		fmt.Printf("height %v\n", height)
		fmt.Printf("included %v\n", included)
		fmt.Printf("rkey %x\n", rkey)
		fmt.Printf("rval %x\n", rval)
		// fmt.Println("bitset", hex.EncodeToString(bitset))
		data := binary.LittleEndian.Uint64(bitset)
		fmt.Println("bitset", data)
		fmt.Printf("proof ")
		for i := 0; i < len(mp); i++ {
			fmt.Printf("%v", hex.EncodeToString(mp[i]))
		}
		fmt.Println()
		fmt.Println("root", hex.EncodeToString(smt.Root))

		verified := smt.VerifyInclusionC(bitset, keys[0], values[0], mp, height)
		fmt.Println("verified", verified)

		// smt.Commit(txn)

		return nil
	}

	err = db.Update(fn)
	if err != nil {
		panic(err)
	}

}

func main() {
	getValuesforVerifier(true)
}
