package utils

import (
	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

func DecodeHexString(h string) ([]byte, error) {
	h = strings.TrimPrefix(h, "0x")
	h = strings.TrimPrefix(h, "0X")
	return hex.DecodeString(h)
}

func EncodeHexString(h []byte) string {
	return hex.EncodeToString(h)
}

func HexToBytes32(data string) [32]byte {
	var bin []byte = common.Hex2Bytes(data)
	var bin32 [32]byte
	copy(bin32[:], bin)
	return bin32
}

func Bytes32ToHex(data [32]byte) string {
	return common.Bytes2Hex(data[:])
}
