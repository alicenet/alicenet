package secp256k1

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

const (
	pubkeyBytesLenCompressed = 33
)

const (
	pubkeyCompressed byte = 0x02 // y_bit + x coord
)

// S256 returns set of parameters for secp256k1
func S256() *secp256k1.BitCurve {
	return secp256k1.S256()
}

// PrivateKey is the private key
type PrivateKey ecdsa.PrivateKey

// PubKey returns the public key corresponding to the private key
func (priv *PrivateKey) PubKey() *PublicKey {
	return (*PublicKey)(&priv.PublicKey)
}

// Serialize returns the byte slice of the private key
func (priv *PrivateKey) Serialize() []byte {
	return priv.D.Bytes()
}

// NewPrivateKey returns a new private key
func NewPrivateKey(curve *secp256k1.BitCurve) (*PrivateKey, error) {
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}
	return (*PrivateKey)(priv), nil
}

// PrivKeyFromBytes returns the PrivateKey and PublicKey
// corresponding to privKey
func PrivKeyFromBytes(curve *secp256k1.BitCurve, privKey []byte) (*PrivateKey, *PublicKey) {
	x, y := curve.ScalarBaseMult(privKey)
	d := new(big.Int).SetBytes(privKey)
	d.Mod(d, curve.N)
	priv := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: d,
	}
	return (*PrivateKey)(priv), (*PublicKey)(&priv.PublicKey)
}

// PublicKey is the public key
type PublicKey ecdsa.PublicKey

// SerializeCompressed returns the compressed form of the public key
func (pub *PublicKey) SerializeCompressed() []byte {
	pubkeyBytes := make([]byte, 0, pubkeyBytesLenCompressed)
	if pub.X == nil || pub.Y == nil || !isOnCurve(S256(), pub.X, pub.Y) {
		return pubkeyBytes
	}
	format := pubkeyCompressed
	if isOdd(pub.Y) {
		format |= 0x01
	}
	pubkeyBytes = append(pubkeyBytes, format)
	pubkeyBytes = append(pubkeyBytes, utils.ForceSliceToLength(pub.X.Bytes(), pubkeyBytesLenCompressed-1)...)
	return pubkeyBytes
}

// ParsePubKey parses a byte slice into a PublicKey
func ParsePubKey(pubkeyBytes []byte, curve *secp256k1.BitCurve) (*PublicKey, error) {
	pub := &PublicKey{}
	pub.Curve = curve

	if len(pubkeyBytes) == 0 {
		return nil, errors.New("invalid public key: zero byte length")
	}

	format := pubkeyBytes[0]
	yBit := (format & 0x01) == 0x01
	format &= ^byte(0x01)

	switch len(pubkeyBytes) {
	case pubkeyBytesLenCompressed:
		if format != pubkeyCompressed {
			return nil, errors.New("invalid public key: incorrect format bit")
		}
		var err error
		pub.X = new(big.Int).SetBytes(pubkeyBytes[1:])
		pub.Y, err = decompressPoint(curve, pub.X, yBit)
		if err != nil {
			return nil, err
		}
		return pub, nil
	default:
		return nil, errors.New("invalid public key: incorrect byte length")
	}
}

func decompressPoint(curve *secp256k1.BitCurve, x *big.Int, yBit bool) (*big.Int, error) {
	if curve == nil || x == nil {
		return nil, errors.New("invalid curve point: bad curve or x coordiate")
	}
	x3 := new(big.Int).Mul(x, x)
	x3.Mod(x3, curve.P)
	x3.Mul(x3, x)
	x3.Mod(x3, curve.P)
	x3.Add(x3, curve.B)
	x3.Mod(x3, curve.P)
	y := new(big.Int).ModSqrt(x3, curve.P)
	if y == nil {
		return nil, errors.New("invalid curve point: bad y coordinate")
	}
	if yBit != isOdd(y) {
		y.Sub(curve.P, y)
		y.Mod(y, curve.P)
	}
	if !isOnCurve(curve, x, y) {
		return nil, errors.New("invalid curve point: not on curve")
	}
	if yBit != isOdd(y) {
		return nil, errors.New("invalid curve point: invalid oddness")
	}
	return y, nil
}

func isOnCurve(curve *secp256k1.BitCurve, x *big.Int, y *big.Int) bool {
	if curve == nil || x == nil || y == nil {
		return false
	}
	y2 := new(big.Int).Exp(y, big.NewInt(2), curve.P)
	x3 := new(big.Int).Exp(x, big.NewInt(3), curve.P)
	x3.Add(x3, curve.B)
	x3.Mod(x3, curve.P)
	return y2.Cmp(x3) == 0
}

func isOdd(z *big.Int) bool {
	return z.Bit(0) == 1
}
