package utils

import (
	"encoding/hex"
	"fmt"
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

func EncodeArrayOfHexStrings(h [][]byte) string {
	encodedStr := "[ "
	for _, v := range h {
		encodedStr += fmt.Sprintf("%s ", EncodeHexString(v))
	}
	encodedStr += "]"
	return encodedStr
}
