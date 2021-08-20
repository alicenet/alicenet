package objs

import (
	"fmt"

	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/errorz"
)

type PaginationToken struct {
	LastPaginatedType LastPaginatedType
	TotalValue        *uint256.Uint256
	LastKey           []byte
}

type LastPaginatedType byte

const (
	LastPaginatedUtxo LastPaginatedType = iota
	LastPaginatedDeposit
)

// UnmarshalBinary takes a byte slice and returns the corresponding
// PaginationToken object
func (pt *PaginationToken) UnmarshalBinary(data []byte) error {
	if pt == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}

	if data == nil || len(data) < 65 || data[0] > 1 {
		return errorz.ErrInvalid{}.New("bytes invalid")
	}

	pt.LastPaginatedType = LastPaginatedType(data[0])

	TotalValue := &uint256.Uint256{}
	TotalValue.UnmarshalBinary(data[1:33])
	pt.TotalValue = TotalValue

	pt.LastKey = make([]byte, 0, 96)
	pt.LastKey = append(pt.LastKey, data[33:]...)

	return nil
}

// MarshalBinary takes the PaginationToken object and returns the canonical
// byte slice
func (pt *PaginationToken) MarshalBinary() ([]byte, error) {
	if pt == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}

	bTotalValue, err := pt.TotalValue.MarshalBinary()
	if err != nil {
		return nil, err
	}

	bytes := make([]byte, 33, 128)
	bytes[0] = byte(pt.LastPaginatedType)
	copy(bytes[1:33], bTotalValue)
	bytes = append(bytes, pt.LastKey...)

	return bytes, nil
}

func (b PaginationToken) String() string {
	return fmt.Sprintf("{LastPaginatedType: %d, TotalValue: %s, LastKey: 0x%x}", b.LastPaginatedType, b.TotalValue, b.LastKey)
}
