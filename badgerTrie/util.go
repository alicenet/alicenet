/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package trie

import (
	"bytes"
	"crypto/rand"
	"sort"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
)

var (
	// DefaultLeaf is the value that may be passed to Update in order to delete
	// a key from the database.
	DefaultLeaf = Hasher([]byte{0})
)

// Hash is used to convert a hash into a byte array
type Hash [constants.HashLen]byte

// GetFreshData gets size number of byte slice where each slice should have a length of length
func GetFreshData(size, length int) [][]byte {
	var data [][]byte
	for i := 0; i < size; i++ {
		key := make([]byte, constants.HashLen)
		_, err := rand.Read(key)
		if err != nil {
			panic(err)
		}
		data = append(data, Hasher(key)[:length])
	}
	sort.Sort(DataArray(data))
	return data
}

// GetFreshDataSorted gets size number of byte slice where each slice should have a length of length
func GetFreshDataUnsorted(size, length int) [][]byte {
	var data [][]byte
	for i := 0; i < size; i++ {
		key := make([]byte, constants.HashLen)
		_, err := rand.Read(key)
		if err != nil {
			panic(err)
		}
		data = append(data, Hasher(key)[:length])
	}
	return data
}

func convNilToBytes(byteArray []byte) []byte {
	if byteArray == nil {
		return []byte{}
	}
	return byteArray
}

func bitIsSet(bits []byte, i int) bool {
	return bits[i/8]&(1<<uint(7-i%8)) != 0
}

func bitSet(bits []byte, i int) {
	bits[i/8] |= 1 << uint(7-i%8)
}

// Hasher is a default hasher to use and calls the hash function defined
// in our crypto library. It has 32 byte (256 bit) output.
func Hasher(data ...[]byte) []byte {
	return crypto.Hasher(data...)
}

// DataArray is for sorting
type DataArray [][]byte

func (d DataArray) Len() int {
	return len(d)
}
func (d DataArray) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
func (d DataArray) Less(i, j int) bool {
	return bytes.Compare(d[i], d[j]) == -1
}
