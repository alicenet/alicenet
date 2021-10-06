package utils

import (
	"encoding/hex"
	"strings"
)

func DecodeHexString(h string) ([]byte, error) {
	h = strings.TrimPrefix(h, "0x")
	h = strings.TrimPrefix(h, "0X")
	return hex.DecodeString(h)
}

func EncodeHexString(h []byte) string {
	return hex.EncodeToString(h)
}
