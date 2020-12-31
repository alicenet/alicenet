package db

import (
	"bytes"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
)

// MerkleProof is a structure which holds a proof of inclusion or exclusion
// for in a Merkle trie
type MerkleProof struct {
	Included  bool
	KeyHeight int
	Key       []byte
	Value     []byte
	Bitmap    []byte
	Path      [][]byte
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

	Value := make([]byte, constants.HashLen)
	copy(Value[:], mp.Value)
	out = append(out, Value...)

	BitMapLength := utils.MarshalUint16(uint16(len(mp.Bitmap)))
	out = append(out, BitMapLength...)

	Bitmap := make([]byte, len(mp.Bitmap))
	copy(Bitmap, mp.Bitmap)
	out = append(out, Bitmap...)

	PathLength := utils.MarshalUint16(uint16(len(mp.Path)))
	out = append(out, PathLength...)

	for i := 0; i < len(mp.Path); i++ {
		key := mp.Path[i]
		out = append(out, utils.CopySlice(key)...)
	}
	return out, nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// MerkleProof object
func (mp *MerkleProof) UnmarshalBinary(data []byte) error {
	if len(data) < 5+2*constants.HashLen {
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

	Value := make([]byte, constants.HashLen)
	copy(Value, data[3+constants.HashLen:3+2*constants.HashLen])
	mp.Value = Value

	BitMapLengthBytes := data[3+2*constants.HashLen : 5+2*constants.HashLen]
	BitMapLength, err := utils.UnmarshalUint16(BitMapLengthBytes)
	if err != nil {
		return err
	}
	BitMapLengthValue := int(BitMapLength)

	if len(data) < 7+2*constants.HashLen+BitMapLengthValue {
		return errorz.ErrInvalid{}.New("bad merkle proof len")
	}

	Bitmap := make([]byte, BitMapLengthValue)
	copy(Bitmap, data[5+2*constants.HashLen:5+2*constants.HashLen+BitMapLengthValue])
	mp.Bitmap = Bitmap

	PathLengthBytes := data[5+2*constants.HashLen+BitMapLengthValue : 7+2*constants.HashLen+BitMapLengthValue]
	PathLength, err := utils.UnmarshalUint16(PathLengthBytes)
	if err != nil {
		return err
	}
	PathLengthValue := int(PathLength)
	if len(data) != 7+2*constants.HashLen+BitMapLengthValue+PathLengthValue*constants.HashLen {
		return errorz.ErrInvalid{}.New("bad merkle proof len")
	}

	mp.Path = make([][]byte, PathLengthValue)
	idx := 0
	start := 7 + BitMapLengthValue + 2*constants.HashLen
	stop := 7 + BitMapLengthValue + 3*constants.HashLen
	for i := 0; i < constants.HashLen*PathLengthValue; i += constants.HashLen {
		Key := make([]byte, constants.HashLen)
		copy(Key, data[start+i:stop+i])
		mp.Path[idx] = Key
		idx++
	}
	zeroBuf := make([]byte, constants.HashLen)
	if bytes.Equal(mp.Key, zeroBuf) {
		mp.Key = nil
	}
	return nil
}
