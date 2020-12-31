package utils

import (
	"errors"
	"math/big"
	"sort"
)

type kvs struct {
	keys   [][]byte
	values [][]byte
}

func (s kvs) Len() int {
	return len(s.keys)
}
func (s kvs) Swap(i, j int) {
	s.keys[i], s.keys[j] = s.keys[j], s.keys[i]
	s.values[i], s.values[j] = s.values[j], s.values[i]
}
func (s kvs) Less(i, j int) bool {
	kv1 := new(big.Int).SetBytes(s.keys[i])
	kv2 := new(big.Int).SetBytes(s.keys[j])
	cmp := kv1.Cmp(kv2)
	return cmp <= 0
}

// SortKVs allows a pair of slices, one containing keys and one containing
// values to be sorted by treating each key as a big.Int and preserving the
// relationship between keys and values during the sort. SortKVs is a copy
// before sort algorithm. The original slices will remain unsorted.
func SortKVs(keys [][]byte, values [][]byte) ([][]byte, [][]byte, error) {
	if len(keys) != len(values) {
		return nil, nil, errors.New("length mismatch")
	}
	keysCopy := make([][]byte, len(keys))
	valuesCopy := make([][]byte, len(values))
	for i := 0; i < len(keys); i++ {
		keysCopy[i] = CopySlice(keys[i])
		valuesCopy[i] = CopySlice(values[i])
	}
	kVs := kvs{keys: keysCopy, values: valuesCopy}
	sort.Sort(kVs)
	return keysCopy, valuesCopy, nil
}
