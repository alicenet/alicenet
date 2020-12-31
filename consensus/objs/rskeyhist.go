package objs

import (
	"bytes"
	"encoding/hex"

	"github.com/MadBase/MadNet/errorz"

	gUtils "github.com/MadBase/MadNet/utils"
)

// RoundStateHistoricKey ...
type RoundStateHistoricKey struct {
	Prefix []byte
	Height uint32
	Round  uint32
	VAddr  []byte
}

// MarshalBinary takes the RoundStateHistoricKey object and returns
// the canonical byte slice
func (b *RoundStateHistoricKey) MarshalBinary() ([]byte, error) {
	if b == nil || b.Height == 0 || b.Round == 0 {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	key := []byte{}
	Prefix := gUtils.CopySlice(b.Prefix)
	VAddr := make([]byte, hex.EncodedLen(len(b.VAddr)))
	_ = hex.Encode(VAddr, b.VAddr)
	Height := gUtils.MarshalUint32(b.Height)
	Round := gUtils.MarshalUint32(b.Round)
	key = append(key, Prefix...)
	key = append(key, []byte("|")...)
	key = append(key, Height...)
	key = append(key, []byte("|")...)
	key = append(key, Round...)
	key = append(key, []byte("|")...)
	key = append(key, VAddr...)
	return key, nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// RoundStateHistoricKey object
func (b *RoundStateHistoricKey) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	splitData := bytes.Split(data, []byte("|"))
	if len(splitData) != 4 {
		return errorz.ErrCorrupt
	}
	b.Prefix = splitData[0]
	Height, err := gUtils.UnmarshalUint32(splitData[1])
	if err != nil {
		return err
	}
	if Height == 0 {
		return errorz.ErrInvalid{}.New("invalid height in unmarshalling")
	}
	b.Height = Height
	Round, err := gUtils.UnmarshalUint32(splitData[2])
	if err != nil {
		return err
	}
	if Round == 0 {
		return errorz.ErrInvalid{}.New("invalid round in unmarshalling")
	}
	b.Round = Round
	VAddr := make([]byte, hex.DecodedLen(len(splitData[3])))
	_, err = hex.Decode(VAddr, splitData[3])
	if err != nil {
		return err
	}
	b.VAddr = VAddr
	return nil
}

// MakeIterKey ...
func (b *RoundStateHistoricKey) MakeIterKey() ([]byte, error) {
	if b == nil || b.Height == 0 {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	key := []byte{}
	Prefix := gUtils.CopySlice(b.Prefix)
	Height := gUtils.MarshalUint32(b.Height)
	key = append(key, Prefix...)
	key = append(key, []byte("|")...)
	key = append(key, Height...)
	key = append(key, []byte("|")...)
	return key, nil
}
