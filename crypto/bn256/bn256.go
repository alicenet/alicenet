package bn256

import (
	"errors"
	"math/big"

	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
)

// numBytes specifies the number of bytes in a GFp object
const numBytes = 32

// ErrNotUint256 occurs when we work with a uint with more than 256 bits
var ErrNotUint256 = errors.New("big.Ints are not at most 256-bit unsigned integers")

// ErrInvalidData occurs state is invalid
var ErrInvalidData = errors.New("invalid state")

// MarshalBigInt converts a 256-bit uint into a byte slice.
func MarshalBigInt(x *big.Int) ([]byte, error) {
	if x == nil {
		return nil, ErrInvalidData
	}
	xBytes := x.Bytes()
	xBytesLen := len(xBytes)
	if xBytesLen > numBytes {
		return nil, ErrNotUint256
	}
	byteSlice := make([]byte, numBytes)
	for j := 1; j <= xBytesLen; j++ {
		byteSlice[numBytes-j] = xBytes[xBytesLen-j]
	}
	return byteSlice, nil
}

// MarshalG1Big is used to compare the result from Go code generated
// by Solidity with the original Go code in cloudflare directory.
func MarshalG1Big(hashPoint [2]*big.Int) ([]byte, error) {
	// Note: assuming hashPoint is a value G1 point
	bigZero := big.NewInt(0)
	hashMarsh := make([]byte, 2*numBytes)
	x := hashPoint[0]
	y := hashPoint[1]
	if x == nil || y == nil {
		return nil, ErrInvalidData
	}
	if x.Cmp(bigZero) == 0 {
		return hashMarsh, nil
	}
	tmpBytes, err := MarshalBigInt(x)
	if err != nil {
		return nil, err
	}
	for k := 0; k < numBytes; k++ {
		hashMarsh[k] = tmpBytes[k]
	}
	tmpBytes, err = MarshalBigInt(y)
	if err != nil {
		return nil, err
	}
	for k := 0; k < numBytes; k++ {
		hashMarsh[numBytes+k] = tmpBytes[k]
	}
	return hashMarsh, nil
}

// MarshalG2Big is used to compare the result from Go code generated
// by Solidity with the original Go code in cloudflare directory.
func MarshalG2Big(hashPoint [4]*big.Int) ([]byte, error) {
	// Note: assuming hashPoint is a value G2 point
	bigZero := big.NewInt(0)
	hashMarsh := make([]byte, 4*numBytes)
	xi := hashPoint[0]
	x := hashPoint[1]
	yi := hashPoint[2]
	y := hashPoint[3]
	if x == nil || xi == nil || y == nil || yi == nil {
		return nil, ErrInvalidData
	}
	if xi.Cmp(bigZero) == 0 {
		return hashMarsh, nil
	}
	tmpBytes, err := MarshalBigInt(xi)
	if err != nil {
		return nil, err
	}
	for k := 0; k < numBytes; k++ {
		hashMarsh[k] = tmpBytes[k]
	}
	tmpBytes, err = MarshalBigInt(x)
	if err != nil {
		return nil, err
	}
	for k := 0; k < numBytes; k++ {
		hashMarsh[numBytes+k] = tmpBytes[k]
	}
	tmpBytes, err = MarshalBigInt(yi)
	if err != nil {
		return nil, err
	}
	for k := 0; k < numBytes; k++ {
		hashMarsh[2*numBytes+k] = tmpBytes[k]
	}
	tmpBytes, err = MarshalBigInt(y)
	if err != nil {
		return nil, err
	}
	for k := 0; k < numBytes; k++ {
		hashMarsh[3*numBytes+k] = tmpBytes[k]
	}
	return hashMarsh, nil
}

// G1ToBigIntArray converts cloudflare.G2 into big.Int array for testing purposes.
func G1ToBigIntArray(g1 *cloudflare.G1) ([2]*big.Int, error) {
	if g1 == nil {
		return [2]*big.Int{}, ErrInvalidData
	}
	g1Bytes := g1.Marshal()
	g1X := new(big.Int).SetBytes(g1Bytes[:numBytes])
	g1Y := new(big.Int).SetBytes(g1Bytes[numBytes : 2*numBytes])
	g1BigInt := [2]*big.Int{g1X, g1Y}
	return g1BigInt, nil
}

// BigIntArrayToG1 converts Ethereum big.Int G1 arrays into cloudflare.G1
// elements for computing purposes.
func BigIntArrayToG1(g1BigInt [2]*big.Int) (*cloudflare.G1, error) {
	g1Bytes, err := MarshalG1Big(g1BigInt)
	if err != nil {
		return nil, err
	}
	g1 := new(cloudflare.G1)
	_, err = g1.Unmarshal(g1Bytes)
	if err != nil {
		return nil, err
	}
	return g1, nil
}

// BigIntArrayToG2 converts Ethereum big.Int G2 arrays into cloudflare.G2
// elements for computing purposes.
func BigIntArrayToG2(g2BigInt [4]*big.Int) (*cloudflare.G2, error) {
	g2Bytes, err := MarshalG2Big(g2BigInt)
	if err != nil {
		return nil, err
	}
	g2 := new(cloudflare.G2)
	_, err = g2.Unmarshal(g2Bytes)
	if err != nil {
		return nil, err
	}
	return g2, nil
}

// BigIntArraySliceToG1 converts Ethereum big.Int G1 array slice into
// cloudflare.G1 slice for computing purposes.
func BigIntArraySliceToG1(g1BigIntArray [][2]*big.Int) ([]*cloudflare.G1, error) {
	m := len(g1BigIntArray)
	g1Array := make([]*cloudflare.G1, m)
	for j := 0; j < m; j++ {
		g1Big := g1BigIntArray[j]
		g1, err := BigIntArrayToG1(g1Big)
		if err != nil {
			return nil, err
		}
		g1Array[j] = g1
	}
	return g1Array, nil
}

// G2ToBigIntArray converts cloudflare.G2 into big.Int array for testing purposes.
func G2ToBigIntArray(g2 *cloudflare.G2) ([4]*big.Int, error) {
	if g2 == nil {
		return [4]*big.Int{}, ErrInvalidData
	}
	g2Bytes := g2.Marshal()
	g2XI := new(big.Int).SetBytes(g2Bytes[:numBytes])
	g2X := new(big.Int).SetBytes(g2Bytes[numBytes : 2*numBytes])
	g2YI := new(big.Int).SetBytes(g2Bytes[2*numBytes : 3*numBytes])
	g2Y := new(big.Int).SetBytes(g2Bytes[3*numBytes : 4*numBytes])
	g2BigInt := [4]*big.Int{g2XI, g2X, g2YI, g2Y}
	return g2BigInt, nil
}

// MarshalBigIntSlice returns a byte slice for encoding that we will use to
// check that we hash to the correct value. All of these values are assumed
// to be uint256.
func MarshalBigIntSlice(bigSlice []*big.Int) ([]byte, error) {
	n := len(bigSlice)
	byteSlice := make([]byte, numBytes*n)
	for k := 0; k < n; k++ {
		x := bigSlice[k]
		tmpBytes, err := MarshalBigInt(x)
		if err != nil {
			return nil, err
		}
		for j := 0; j < numBytes; j++ {
			byteSlice[k*numBytes+j] = tmpBytes[j]
		}
	}
	return byteSlice, nil
}

// MarshalG1BigSlice creates a byte slice from G1Big slice.
// This is used in testing purposes.
func MarshalG1BigSlice(g1BigSlice [][2]*big.Int) ([]byte, error) {
	n := len(g1BigSlice)
	byteSlice := make([]byte, 2*numBytes*n)
	for k := 0; k < n; k++ {
		x := g1BigSlice[k]
		tmpBytes, err := MarshalG1Big(x)
		if err != nil {
			return nil, err
		}
		for j := 0; j < 2*numBytes; j++ {
			byteSlice[k*2*numBytes+j] = tmpBytes[j]
		}
	}
	return byteSlice, nil
}
