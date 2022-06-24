package db

import (
	"bytes"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
)

// MerkleProof is a structure which holds a proof of inclusion or exclusion
// for in a Merkle trie
type MerkleProof struct {
	Included   bool
	KeyHeight  int
	Key        []byte
	ProofKey   []byte
	ProofValue []byte
	Bitmap     []byte
	Path       [][]byte
}

// MarshalBinary takes the MerkleProof object and returns the canonical
// byte slice
func (mp *MerkleProof) MarshalBinary() ([]byte, error) {
	out := []byte{}
	Included := make([]byte, 1)
	if mp.Included {
		Included[0] = uint8(1)
	}
	out = append(out, Included...)

	Height := utils.MarshalUint16(uint16(mp.KeyHeight))
	out = append(out, Height...)

	Key := make([]byte, constants.HashLen)
	copy(Key[:], mp.Key)
	out = append(out, Key...)

	ProofKey := make([]byte, constants.HashLen)
	copy(ProofKey[:], mp.ProofKey)
	out = append(out, ProofKey...)

	ProofValue := make([]byte, constants.HashLen)
	copy(ProofValue[:], mp.ProofValue)
	out = append(out, ProofValue...)

	BitMapLength := utils.MarshalUint16(uint16(len(mp.Bitmap)))
	out = append(out, BitMapLength...)

	PathLength := utils.MarshalUint16(uint16(len(mp.Path)))
	out = append(out, PathLength...)

	Bitmap := make([]byte, len(mp.Bitmap))
	copy(Bitmap, mp.Bitmap)
	out = append(out, Bitmap...)

	for i := 0; i < len(mp.Path); i++ {
		key := mp.Path[i]
		out = append(out, utils.CopySlice(key)...)
	}
	return out, nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// MerkleProof object
func (mp *MerkleProof) UnmarshalBinary(data []byte) error {
	if len(data) < 7+3*constants.HashLen {
		return errorz.ErrInvalid{}.New("bad merkle proof len")
	}

	Included := data[0:1]
	if Included[0] == 1 {
		mp.Included = true
	}

	HeightBytes := data[1:3]
	Height, err := utils.UnmarshalUint16(HeightBytes)
	if err != nil {
		return err
	}
	mp.KeyHeight = int(Height)

	Key := make([]byte, constants.HashLen)
	copy(Key[:], data[3:3+constants.HashLen])
	mp.Key = Key

	ProofKey := make([]byte, constants.HashLen)
	copy(ProofKey[:], data[3+constants.HashLen:3+2*constants.HashLen])
	mp.ProofKey = ProofKey

	ProofValue := make([]byte, constants.HashLen)
	copy(ProofValue, data[3+2*constants.HashLen:3+3*constants.HashLen])
	mp.ProofValue = ProofValue

	BitMapLengthBytes := data[3+3*constants.HashLen : 5+3*constants.HashLen]
	BitMapLength, err := utils.UnmarshalUint16(BitMapLengthBytes)
	if err != nil {
		return err
	}
	BitMapLengthValue := int(BitMapLength)

	PathLengthBytes := data[5+3*constants.HashLen : 7+3*constants.HashLen]
	PathLength, err := utils.UnmarshalUint16(PathLengthBytes)
	if err != nil {
		return err
	}
	PathLengthValue := int(PathLength)
	if len(data) != 7+3*constants.HashLen+BitMapLengthValue+PathLengthValue*constants.HashLen {
		return errorz.ErrInvalid{}.New("bad merkle proof len")
	}

	Bitmap := make([]byte, BitMapLengthValue)
	copy(Bitmap, data[7+3*constants.HashLen:7+3*constants.HashLen+BitMapLengthValue])
	mp.Bitmap = Bitmap

	mp.Path = make([][]byte, PathLengthValue)
	idx := 0
	start := 7 + BitMapLengthValue + 3*constants.HashLen
	stop := 7 + BitMapLengthValue + 4*constants.HashLen
	for i := 0; i < constants.HashLen*PathLengthValue; i += constants.HashLen {
		Key := make([]byte, constants.HashLen)
		copy(Key, data[start+i:stop+i])
		mp.Path[idx] = Key
		idx++
	}
	zeroBuf := make([]byte, constants.HashLen)
	if bytes.Equal(mp.ProofKey, zeroBuf) {
		mp.ProofKey = nil
	}
	if bytes.Equal(mp.ProofValue, zeroBuf) {
		mp.ProofValue = nil
	}
	return nil
}
