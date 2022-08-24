package transport

import (
	"encoding/binary"
	"encoding/hex"
)

// Converts a uint32 into a hex string.
func uint32ToHexString(v uint32) string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b[:], v)
	return hex.EncodeToString(b)
}

// Converts a hex string into a uint32.
// If the string will overflow a uint32,
// an error is raised.
func uint32FromHexString(v string) (uint32, error) {
	hb, err := hex.DecodeString(v)
	if err != nil {
		return 0, err
	}
	if len(hb) > 4 {
		return 0, ErrUint32Overflow
	}
	return binary.BigEndian.Uint32(hb[:]), nil
}
