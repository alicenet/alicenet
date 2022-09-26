package crypto

import "github.com/alicenet/alicenet/crypto/bn256/cloudflare"

// Hasher is the default hasher and calls the hash function defined
// in our cloudflare library. It has 32 byte (256 bit) output.
func Hasher(data ...[]byte) []byte {
	return cloudflare.HashFunc256(data...)
}
