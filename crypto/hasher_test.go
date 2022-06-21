package crypto

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
)

func TestHasher(t *testing.T) {
	emptyByte := []byte{}
	res := Hasher(emptyByte)

	emptyByteHash := make([]byte, constants.HashLen)

	emptyByteHash[0] = 197
	emptyByteHash[1] = 210
	emptyByteHash[2] = 70
	emptyByteHash[3] = 1
	emptyByteHash[4] = 134
	emptyByteHash[5] = 247
	emptyByteHash[6] = 35
	emptyByteHash[7] = 60

	emptyByteHash[8] = 146
	emptyByteHash[9] = 126
	emptyByteHash[10] = 125
	emptyByteHash[11] = 178
	emptyByteHash[12] = 220
	emptyByteHash[13] = 199
	emptyByteHash[14] = 3
	emptyByteHash[15] = 192

	emptyByteHash[16] = 229
	emptyByteHash[17] = 0
	emptyByteHash[18] = 182
	emptyByteHash[19] = 83
	emptyByteHash[20] = 202
	emptyByteHash[21] = 130
	emptyByteHash[22] = 39
	emptyByteHash[23] = 59

	emptyByteHash[24] = 123
	emptyByteHash[25] = 250
	emptyByteHash[26] = 216
	emptyByteHash[27] = 4
	emptyByteHash[28] = 93
	emptyByteHash[29] = 133
	emptyByteHash[30] = 164
	emptyByteHash[31] = 112

	if !bytes.Equal(res, emptyByteHash) {
		t.Fatal("HashFunc256 changed; invalid hash result for empty byte slice")
	}
}
