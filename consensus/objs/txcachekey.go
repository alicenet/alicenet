package objs

import (
	"bytes"

	"github.com/MadBase/MadNet/errorz"

	"github.com/MadBase/MadNet/constants"
	gUtils "github.com/MadBase/MadNet/utils"
)

// TxCacheKey ...
type TxCacheKey struct {
	Prefix []byte
	Height uint32
	TxHash []byte
}

// MarshalBinary takes the TxCacheKey object and returns
// the canonical byte slice
func (b *TxCacheKey) MarshalBinary() ([]byte, error) {
	if b == nil || b.Height == 0 || len(b.TxHash) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	key := []byte{}
	Prefix := gUtils.CopySlice(b.Prefix)
	TxHash := gUtils.CopySlice(b.TxHash)
	Height := gUtils.MarshalUint32(b.Height)
	key = append(key, Prefix...)
	key = append(key, []byte("|")...)
	key = append(key, Height...)
	key = append(key, []byte("|")...)
	key = append(key, TxHash...)
	return key, nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// TxCacheKey object
func (b *TxCacheKey) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	splitData := bytes.Split(data, []byte("|"))
	if len(splitData) != 3 {
		return errorz.ErrCorrupt
	}
	b.Prefix = splitData[0]
	Height, err := gUtils.UnmarshalUint32(splitData[1])
	if err != nil {
		return err
	}
	if Height == 0 {
		return errorz.ErrInvalid{}.New("invalid height for unmarshalling")
	}
	b.Height = Height
	TxHash := gUtils.CopySlice(splitData[2])
	if len(TxHash) != constants.HashLen {
		return errorz.ErrInvalid{}.New("invalid txhash for unmarshalling; incorrect length")
	}
	b.TxHash = TxHash
	return nil
}

// MakeIterKey ...
func (b *TxCacheKey) MakeIterKey() ([]byte, error) {
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
