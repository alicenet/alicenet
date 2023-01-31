package txhash

import (
	"bytes"

	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
)

func LeafHash(data []byte) []byte {
	return crypto.Hasher(crypto.Hasher(data))
}

func HashPair(data1, data2 []byte) []byte {
	if bytes.Compare(data1, data2) < 0 {
		return crypto.Hasher(data1, data2)
	} else {
		return crypto.Hasher(data2, data1)
	}
}

func ComputeMerkleRoot(data [][]byte) []byte {
	leaves := ProcessData(data)
	roothash := ProcessLeaves(leaves)
	return roothash
}

func ProcessData(data [][]byte) [][]byte {
	n := len(data)
	leaves := make([][]byte, n)
	for k := 0; k < n; k++ {
		leaves[k] = LeafHash(data[k])
	}
	return leaves
}

func ProcessLeaves(leaves [][]byte) []byte {
	n := len(leaves)
	tree := make([][]byte, 2*n-1)
	for k := 0; k < n; k++ {
		tree[2*n-1-k] = utils.CopySlice(leaves[k])
	}
	for k := n - 1; k >= 0; k-- {
		tree[k] = HashPair(tree[2*k+1], tree[2*k+2])
	}
	return utils.CopySlice(tree[0])
}
