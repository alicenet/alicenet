package utils

import (
	"encoding/binary"

	"github.com/MadBase/MadNet/errorz"
)

// MarshalUint16 will marshal an uint16 object.
func MarshalUint16(v uint16) []byte {
	vv := make([]byte, 2)
	binary.BigEndian.PutUint16(vv, v)
	return vv
}

// UnmarshalUint16 will unmarshal an uint16 object.
func UnmarshalUint16(v []byte) (uint16, error) {
	if len(v) != 2 {
		return 0, errorz.ErrInvalid{}.New("UnmarshalUint16: invalid byte length; should be 2")
	}
	vv := binary.BigEndian.Uint16(v)
	return vv, nil
}

// MarshalUint32 will marshal an uint32 object.
func MarshalUint32(v uint32) []byte {
	vv := make([]byte, 4)
	binary.BigEndian.PutUint32(vv, v)
	return vv
}

// UnmarshalUint32 will unmarshal an uint32 object.
func UnmarshalUint32(v []byte) (uint32, error) {
	if len(v) != 4 {
		return 0, errorz.ErrInvalid{}.New("UnmarshalUint32: invalid byte length; should be 4")
	}
	vv := binary.BigEndian.Uint32(v)
	return vv, nil
}

// MarshalUint64 will marshal a uint64 object.
func MarshalUint64(v uint64) []byte {
	vv := make([]byte, 8)
	binary.BigEndian.PutUint64(vv, v)
	return vv
}

// UnmarshalUint64 will unmarshal a uint64 object.
func UnmarshalUint64(v []byte) (uint64, error) {
	if len(v) != 8 {
		return 0, errorz.ErrInvalid{}.New("UnmarshalUint64: invalid byte length; should be 8")
	}
	vv := binary.BigEndian.Uint64(v)
	return vv, nil
}

// MarshalInt64 will marshal an int64 object.
func MarshalInt64(v int64) []byte {
	vv := make([]byte, 8)
	binary.BigEndian.PutUint64(vv, uint64(v))
	return vv
}

// UnmarshalInt64 will unmarshal an int64 object.
func UnmarshalInt64(v []byte) (int64, error) {
	if len(v) != 8 {
		return 0, errorz.ErrInvalid{}.New("UnmarshalInt64: invalid byte length; should be 8")
	}
	vv := int64(binary.BigEndian.Uint64(v))
	return vv, nil
}
