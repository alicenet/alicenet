package main

import (
	"encoding/hex"
	"log"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestVerifierContract(t *testing.T) {
	// Generate a new random account and a funded simulator
	gasLimit := uint64(1000000000000000)
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	genAlloc := make(core.GenesisAlloc)
	genAlloc[auth.From] = core.GenesisAccount{Balance: big.NewInt(9223372036854775807)}
	sim := backends.NewSimulatedBackend(genAlloc, gasLimit) // Deploy a token contract on the simulated blockchain
	_, _, c, err := DeployMerkleProof(auth, sim)
	sim.Commit()
	if err != nil {
		log.Fatalf("Failed to deploy new token contract: %v", err)
	}

	// hash
	// s := "e14816b9ec614728dd43f604e5c7e4e76cd92fca200ec69703345ee9aeb7510f"
	// new version uses the leaf hash which gets a hash with the leaf height
	s := "e5133d9d634f9b97f5d043003e3c83f8dc7dcf813c85cfcf6e904ba15fcb5392"
	hsh, err := hex.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}

	// proof
	// s = "9c22ff5f21f0b81b113e63f7db6da94fedef11b2119b4088b89664fb9a3cb658"
	s = "bc64dfca2e4e3f7cf0fc2bc3df54991f11772ce0f56c422568be9d1536dc6258"
	prf, err := hex.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}

	// root
	// s = "c8a241e0fd87db47cdd7d55ab01e81e3bfbe2d899ab57391cbab01b66424d72f"
	s = "be2fa5aaa236896fecb7db3936ea10a72a3b6497a618c77494285348304366bb"
	root, err := hex.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}

	var rootArr, hshArr [32]byte
	copy(rootArr[:], root[:])
	copy(hshArr[:], hsh[:])

	pkey := big.NewInt(0)
	pbs := big.NewInt(1)
	height := big.NewInt(256)

	result, err := c.CheckProof(&bind.CallOpts{}, prf, rootArr, hshArr, pkey, pbs, height)
	if err != nil {
		log.Fatal(err)
	}

	if result != true {
		t.Fatal("test failed: proof did not pass the verification check")
	}

}
