package cloudflare

import (
	"encoding/binary"
	"fmt"
	"math/big"
)

type gfP [4]uint64

// newGFp creates new gfP from int64
func newGFp(x int64) (out *gfP) {
	if x >= 0 {
		out = &gfP{uint64(x)}
	} else {
		out = &gfP{uint64(-x)}
		gfpNeg(out, out)
	}

	montEncode(out, out)
	return out
}

// bigToGFp enables one to convert big.Int into gfP;
// it first computes the result modulo P before conversion
//
// This has been modified to use code from Cloudflare's hashToBase function
// currently available. It greatly simplifies things.
func bigToGFp(a *big.Int) *gfP {
	c := new(big.Int).Mod(a, P)
	v := c.Bytes()
	v32 := [32]byte{}
	for i := len(v) - 1; i >= 0; i-- {
		v32[len(v)-1-i] = v[i]
	}
	t := &gfP{
		binary.LittleEndian.Uint64(v32[0*8 : 1*8]),
		binary.LittleEndian.Uint64(v32[1*8 : 2*8]),
		binary.LittleEndian.Uint64(v32[2*8 : 3*8]),
		binary.LittleEndian.Uint64(v32[3*8 : 4*8]),
	}
	montEncode(t, t)
	return t
}

func (e *gfP) String() string {
	return fmt.Sprintf("%16.16x%16.16x%16.16x%16.16x", e[3], e[2], e[1], e[0])
}

func (e *gfP) IsEqual(f *gfP) bool {
	return e.String() == f.String()
}

func (e *gfP) Set(f *gfP) {
	e[0] = f[0]
	e[1] = f[1]
	e[2] = f[2]
	e[3] = f[3]
}

// exp performs modular exponentiation using square and multiply method.
func (e *gfP) exp(f *gfP, bits [4]uint64) {
	sum, power := &gfP{}, &gfP{}
	sum.Set(rN1)
	power.Set(f)

	for word := 0; word < 4; word++ {
		for bit := uint(0); bit < 64; bit++ {
			if (bits[word]>>bit)&1 == 1 {
				gfpMul(sum, sum, power)
			}
			gfpMul(power, power, power)
		}
	}

	gfpMul(sum, sum, r3)
	e.Set(sum)
}

// Invert computes the multiplicative inverse in gfP
func (e *gfP) Invert(f *gfP) {
	e.exp(f, pMinus2)
}

// Sqrt computes the square roots in gfP.
// This assumes a square root of e exists; use Legendre to determine existence.
// If square root does not exist, this call will compute square root of -e.
func (e *gfP) Sqrt(f *gfP) {
	e.exp(f, pPlus1Over4)
}

// Legendre computes the Legendre symbol.
// This determines whether or not a square root of e exists modulo P.
func (e *gfP) Legendre() int {
	f := &gfP{}
	f.exp(e, pMinus1Over2)
	montDecode(f, f)
	if *f != [4]uint64{} {
		return 2*int(f[0]&1) - 1
	}
	return 0
}

func (e *gfP) Marshal(out []byte) {
	for w := uint(0); w < 4; w++ {
		for b := uint(0); b < 8; b++ {
			out[8*w+b] = byte(e[3-w] >> (56 - 8*b))
		}
	}
}

func (e *gfP) Unmarshal(in []byte) error {
	// Unmarshal the bytes into little endian form
	for w := uint(0); w < 4; w++ {
		for b := uint(0); b < 8; b++ {
			e[3-w] += uint64(in[8*w+b]) << (56 - 8*b)
		}
	}
	// Ensure the point respects the curve modulus
	for i := 3; i >= 0; i-- {
		if e[i] < p2[i] {
			return nil
		}
		if e[i] > p2[i] {
			return ErrInvalidCoordinate
		}
	}
	return ErrInvalidCoordinate
}

func montEncode(c, a *gfP) { gfpMul(c, a, r2) }
func montDecode(c, a *gfP) { gfpMul(c, a, &gfP{1}) }

// sign0 computes the sign of gfP.
// We take the convention from Wahby and Boneh 2019 paper:
//
//
//		sign0(e) == { 1,	0 <= e <= (P-1)/2 	}
//					 -1,	(P-1)/2 < e <= P-1
//
// Wahby and Boneh use sign0 to replace their final Legendre function call in
// BaseToG1/G2 to save computation
func sign0(e *gfP) int {
	x := &gfP{}
	montDecode(x, e)
	for w := 3; w >= 0; w-- {
		if x[w] < pMinus1Over2[w] {
			return 1
		} else if x[w] > pMinus1Over2[w] {
			return -1
		}
	}
	return 1
}
