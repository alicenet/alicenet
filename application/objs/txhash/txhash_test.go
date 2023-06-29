package txhash

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
)

func TestLeafHash(t *testing.T) {
	data0 := make([]byte, constants.HashLen)
	trueLeafhash0 := crypto.Hasher(crypto.Hasher(data0))
	retLeafhash0 := LeafHash(data0)
	if !bytes.Equal(retLeafhash0, trueLeafhash0) {
		t.Fatal("invalid leafhash 0")
	}

	data1 := make([]byte, constants.HashLen)
	for k := 0; k < len(data1); k++ {
		data1[k] = 1
	}
	trueLeafhash1 := crypto.Hasher(crypto.Hasher(data1))
	retLeafhash1 := LeafHash(data1)
	if !bytes.Equal(retLeafhash1, trueLeafhash1) {
		t.Fatal("invalid leafhash 1")
	}

	data2 := make([]byte, constants.HashLen)
	for k := 0; k < len(data2); k++ {
		data2[k] = 2
	}
	trueLeafhash2 := crypto.Hasher(crypto.Hasher(data2))
	retLeafhash2 := LeafHash(data2)
	if !bytes.Equal(retLeafhash2, trueLeafhash2) {
		t.Fatal("invalid leafhash 2")
	}
}

func TestHashPair(t *testing.T) {
	data0 := make([]byte, constants.HashLen)
	data1 := make([]byte, constants.HashLen)
	for k := 0; k < len(data1); k++ {
		data1[k] = 1
	}
	trueHashPair := crypto.Hasher(data0, data1)
	retHashPair1 := HashPair(data0, data1)
	if !bytes.Equal(retHashPair1, trueHashPair) {
		t.Fatal("invalid leafhash 1")
	}
	retHashPair2 := HashPair(data1, data0)
	if !bytes.Equal(retHashPair2, trueHashPair) {
		t.Fatal("invalid leafhash 2")
	}
}

func TestProcessData(t *testing.T) {
	data0 := make([]byte, constants.HashLen)
	data1 := make([]byte, constants.HashLen)
	for k := 0; k < len(data1); k++ {
		data1[k] = 1
	}
	data2 := make([]byte, constants.HashLen)
	for k := 0; k < len(data1); k++ {
		data1[k] = 2
	}
	data := [][]byte{data0, data1, data2}

	leaf0 := LeafHash(data0)
	leaf1 := LeafHash(data1)
	leaf2 := LeafHash(data2)
	leaves := [][]byte{leaf0, leaf1, leaf2}

	retLeaves := ProcessData(data)
	if len(leaves) != len(retLeaves) {
		t.Fatal("invalid length")
	}
	for k := 0; k < len(leaves); k++ {
		if !bytes.Equal(retLeaves[k], leaves[k]) {
			t.Fatalf("invalid leaf at %d", k)
		}
	}
}

func TestProcessLeaves1(t *testing.T) {
	data0 := make([]byte, constants.HashLen)
	data1 := make([]byte, constants.HashLen)
	for k := 0; k < len(data1); k++ {
		data1[k] = 1
	}
	data2 := make([]byte, constants.HashLen)
	for k := 0; k < len(data2); k++ {
		data2[k] = 2
	}
	//data := [][]byte{data0, data1, data2}

	leaf0 := LeafHash(data0)
	leaf1 := LeafHash(data1)
	leaf2 := LeafHash(data2)
	leaves := [][]byte{leaf0, leaf1, leaf2}

	int1 := HashPair(leaf1, leaf0)
	trueRootHash := HashPair(int1, leaf2)

	retRootHash := ProcessLeaves(leaves)
	if !bytes.Equal(retRootHash, trueRootHash) {
		t.Fatal("invalid root hash")
	}
}

func TestProcessLeaves2(t *testing.T) {
	data0 := make([]byte, constants.HashLen)
	data1 := make([]byte, constants.HashLen)
	for k := 0; k < len(data1); k++ {
		data1[k] = 1
	}
	data2 := make([]byte, constants.HashLen)
	for k := 0; k < len(data2); k++ {
		data2[k] = 2
	}
	data3 := make([]byte, constants.HashLen)
	for k := 0; k < len(data3); k++ {
		data3[k] = 3
	}
	data4 := make([]byte, constants.HashLen)
	for k := 0; k < len(data4); k++ {
		data4[k] = 4
	}

	leaf0 := LeafHash(data0)
	leaf1 := LeafHash(data1)
	leaf2 := LeafHash(data2)
	leaf3 := LeafHash(data3)
	leaf4 := LeafHash(data4)
	leaves := [][]byte{leaf0, leaf1, leaf2, leaf3, leaf4}

	int3 := HashPair(leaf1, leaf0)
	int2 := HashPair(leaf3, leaf2)
	int1 := HashPair(int3, leaf4)
	trueRootHash := HashPair(int1, int2)

	retRootHash := ProcessLeaves(leaves)
	if !bytes.Equal(retRootHash, trueRootHash) {
		t.Fatal("invalid root hash")
	}
}

func TestComputeMerkleRoot1(t *testing.T) {
	data0 := make([]byte, constants.HashLen)
	data1 := make([]byte, constants.HashLen)
	for k := 0; k < len(data1); k++ {
		data1[k] = 1
	}
	data2 := make([]byte, constants.HashLen)
	for k := 0; k < len(data2); k++ {
		data2[k] = 2
	}
	data := [][]byte{data0, data1, data2}

	leaf0 := LeafHash(data0)
	leaf1 := LeafHash(data1)
	leaf2 := LeafHash(data2)

	int1 := HashPair(leaf1, leaf0)
	trueRootHash := HashPair(int1, leaf2)

	retRootHash := ComputeMerkleRoot(data)
	if !bytes.Equal(retRootHash, trueRootHash) {
		t.Fatal("invalid root hash")
	}
}

func TestComputeMerkleRoot2(t *testing.T) {
	data0 := make([]byte, constants.HashLen)
	data1 := make([]byte, constants.HashLen)
	for k := 0; k < len(data1); k++ {
		data1[k] = 1
	}
	data2 := make([]byte, constants.HashLen)
	for k := 0; k < len(data2); k++ {
		data2[k] = 2
	}
	data3 := make([]byte, constants.HashLen)
	for k := 0; k < len(data3); k++ {
		data3[k] = 3
	}
	data4 := make([]byte, constants.HashLen)
	for k := 0; k < len(data4); k++ {
		data4[k] = 4
	}
	data := [][]byte{data0, data1, data2, data3, data4}

	leaf0 := LeafHash(data0)
	leaf1 := LeafHash(data1)
	leaf2 := LeafHash(data2)
	leaf3 := LeafHash(data3)
	leaf4 := LeafHash(data4)

	int3 := HashPair(leaf1, leaf0)
	int2 := HashPair(leaf3, leaf2)
	int1 := HashPair(int3, leaf4)
	trueRootHash := HashPair(int1, int2)

	retRootHash := ComputeMerkleRoot(data)
	if !bytes.Equal(retRootHash, trueRootHash) {
		t.Fatal("invalid root hash")
	}
}
