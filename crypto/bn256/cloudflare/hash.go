package cloudflare

// This file contains hash-to-curve algorithms implementing the 2012
// LatinCrypto conference paper
//
// 		Indifferentiable Hashing to Barreto–Naehrig Curves
//		by Pierre-Alain Fouque and Mehdi Tibouchi
//
// with some additions from 2019 paper
//
// 		Fast and simple constant-time hashing to the BLS12-381 elliptic curve
// 		by Riad S. Wahby and Dan Boneh.
//
// We use Wahby and Boneh 2019 to develop a similar HashToG2 algorithm;
// it is essentially the same as our HashToG1 algorithm.

import (
	"crypto/subtle"
	"encoding/binary"
	"math/big"
)

// HashToG1 maps a byte slice into G1; we use the algorithm from Fouque
// and Tibouchi 2012. Domain separation in hashToBase allows us to compute
// t0 and t1, two elements of the base field F_p (gfP) which are from
// independent hash functions. From there, t0 is mapped to h0 and t1 is
// mapped to h1, with h0 and h1 elements of G1. Adding them together gives
// us our desired HashToG1 function. Taking the sum of two elements is
// required so that the hash function is uniform over G1; otherwise,
// there is a significant portion of the elliptic curve that is missed,
// which would mean it could not be considered a random oracle.
// The modification we made following Wahby and Boneh 2019 to improve
// efficiency invalidates the proof in Fouque and Tibouchi 2012 proving
// HashToG1 is indistinguishable from a random oracle. It is not currently
// known if the proof could be modified to match the implementation.
//
// We throw an error whenever we map to Infinity (the identity element),
// the generator, or the generator's inverse. This breaks security
// assumptions: when we map to Infinity, anyone can be proven to sign
// anything; when we map to the generator or its inverse, it is possible
// to falsely prove the group signed the message (this happens due to the
// distributed key generation protocol). Thus, these cases should be
// handled carefully, and we instead choose to return an error.
func HashToG1(msg []byte) (*G1, error) {
	g1HashPoint := &G1{}
	const dsp0 byte = 0x00
	const dsp1 byte = 0x01
	const dsp2 byte = 0x02
	const dsp3 byte = 0x03
	fieldElement0 := hashToBase(msg, dsp0, dsp1)
	fieldElement1 := hashToBase(msg, dsp2, dsp3)
	g1Point0, err0 := baseToG1(fieldElement0)
	if err0 != nil {
		return nil, err0
	}
	g1Point1, err1 := baseToG1(fieldElement1)
	if err1 != nil {
		return nil, err1
	}
	g1HashPoint.Add(g1Point0, g1Point1)
	if !safeSigningPoint(g1HashPoint) {
		return nil, ErrDangerousPoint
	}
	return g1HashPoint, nil
}

// safeSigningPoint ensures that we only return safe points from HashToG1.
// In particular, we do not sign Infinity (the identity element), the
// generator, or the generator's inverse (negation).
func safeSigningPoint(g *G1) bool {
	p := &curvePoint{}
	p.SetInfinity()
	if g.p.IsEqual(p) {
		return false
	}
	p.Set(curveGen)
	if g.p.IsEqual(p) {
		return false
	}
	p.Neg(p)
	return !g.p.IsEqual(p)
}

// hashToBase is a hash function which takes the msg byte slice and outputs
// a finite field value t (gfP). t will then be (deterministically) mapped
// into G1. The bytes byteI and byteJ allow for domain separation.
// HashToG1 (the full function above) requires two independent hash functions
// to the finite field.
//
// We use the output of two independent hashes to have a uniform 512-bit
// output. This enables us to have a distribution that is much closer
// to a uniform hashToBase than just using 256 bits. We give the rational now.
//
// We let s0 be uniformly distributed in uint256; that is, s0 is a uniformly
// distributed 256-bit unsigned integer. This can be obtained by converting
// the output of a 256-bit hash function as a 256-bit unsigned integer.
// Just using s0 mod P will lead to bias in the distribution.
// In particular, there is bias towards the lower 5% of the numbers in
// [0, P). The 1-norm error between s0 mod P and a uniform
// distribution is ~ 1/4. By itself, this 1-norm error is not too enlightening,
// but we will compare it with another distribution that has much smaller
// 1-norm error.
//
// To obtain a better distribution with less bias, we take 2 uint256 hash
// outputs (using byteI and byteJ for domain separation so the hashes are
// independent) and concatenate them to form a “uint512”. Of course,
// this is not possible in practice, so we view the combined output as
//
//	x == s0*2^256 + s1.
//
// This implies that x (combined from s0 and s1 in this way) is a
// 512-bit uint. If s0 and s1 are uniformly distributed modulo 2^256,
// then x is uniformly distributed modulo 2^512. We now want to reduce
// this modulo P. This is done as follows:
//
//	x mod P == [(s0 mod P)*(2^256 mod P)] + s1 mod P.
//
// This allows us easily compute the result without needing to implement
// higher precision in the EVM. The 1-norm error between x mod P and a uniform
// distribution is ~1e-77. This is a *significant* improvement from s0 mod P.
// For all practical purposes, there is no difference from a uniform
// distribution.
func hashToBase(msg []byte, dsp0 byte, dsp1 byte) *gfP {
	dsp0msg := append([]byte{dsp0}, msg...)
	hashResult0 := HashFunc256(dsp0msg)
	dsp1msg := append([]byte{dsp1}, msg...)
	hashResult1 := HashFunc256(dsp1msg)
	fieldElement0 := new(big.Int).SetBytes(hashResult0)
	fieldElement1 := new(big.Int).SetBytes(hashResult1)
	fieldElement0.Mul(fieldElement0, two256ModP)
	fieldElement0.Mod(fieldElement0, P)
	fieldElement1.Mod(fieldElement1, P)
	fieldElement := new(big.Int).Add(fieldElement0, fieldElement1)
	fieldElement.Mod(fieldElement, P)
	fieldElementGFp := bigToGFp(fieldElement)
	return fieldElementGFp
}

// baseToG1 takes an element of the finite field and outputs an element of
// the underlying elliptic curve. This algorithm follows Fouque and Tibouchi's
// 2012 paper “Indifferentiable Hashing to Barreto--Naehrig Curves”.
// We also use some ideas from Wahby and Boneh's 2019 paper
// “Fast and simple constant-time hashing to the BLS12-381 elliptic curve”.
// The main idea is that given an input t, we produce points x1, x2, x3
// which are finite field elements. At least one of these points is the
// x-coordinate of a valid point on the elliptic curve. For uniqueness, we
// choose the first point which is valid.
// The Wahby and Boneh idea we use is to choose alpha so that we minimize
// exponentiations. Additionally, when alpha == 0, x1 is still a valid point,
// so we do not need a special case when t == 0.
// We also use sign0 from Wahby and Boneh as well in order to further reduce
// the required exponentiations.
// It is not currently known how much this affects the resulting hashing
// distribution.
//
// We let
//
//	g(x) == x^3 + curveB
func baseToG1(t *gfP) (*G1, error) {
	gfpOne := newGFp(1)

	// We have
	//
	//		alphaPart1 == t^2
	//
	//		alphaPart2 == g(1) + t^2
	//
	//		alpha 	   == (alphaPart1*alphaPart2)^(-1)
	//
	//		tmp 	   == (g(1) + t^2)^3
	//				   == alphaPart2^3
	//
	//		tFourth	   == t^4
	//
	// Inversion is computed using multiplication, so if
	// alphaPart1 == 0 or alphaPart2 == 0, then alpha == 0.
	// This still leads to a valid value.
	alphaPart1 := &gfP{}
	gfpMul(alphaPart1, t, t)
	alphaPart2 := &gfP{}
	gfpAdd(alphaPart2, alphaPart1, g1HashConst4)
	alpha := &gfP{} // alpha == 0 when t == 0
	gfpMul(alpha, alphaPart1, alphaPart2)
	alpha.Invert(alpha)
	tFourth := &gfP{}
	gfpMul(tFourth, alphaPart1, alphaPart1)
	tmp := &gfP{}
	gfpMul(tmp, alphaPart2, alphaPart2)
	gfpMul(tmp, tmp, alphaPart2)

	// x1, x2, and x3 are possible x coordinates for
	// this baseToG1 function. We have
	//
	//		x1 == g1HashConst1 - g1HashConst2*tFourth*alpha
	//
	//		x2 == -1 - x1
	//
	//		x3 == 1 - g1HashConst3*tmp*alpha
	x1 := &gfP{}
	gfpMul(x1, g1HashConst2, tFourth)
	gfpMul(x1, x1, alpha)
	gfpNeg(x1, x1)
	gfpAdd(x1, g1HashConst1, x1)

	x2 := &gfP{}
	gfpAdd(x2, x1, gfpOne)
	gfpNeg(x2, x2)

	x3 := &gfP{}
	gfpMul(x3, g1HashConst3, tmp)
	gfpMul(x3, x3, alpha)
	gfpNeg(x3, x3)
	gfpAdd(x3, x3, gfpOne)

	// We now look at x1 and x2 to determine if either of these coordinates
	// form valid points on the elliptic curve. We have
	//
	//		gX1 == g(x1)
	//			== x1^3 + curveB
	//		gX2 == g(x2)
	//			== x2^3 + curveB
	gX1 := newGFp(0)
	gfpMul(gX1, x1, x1)
	gfpMul(gX1, gX1, x1)
	gfpAdd(gX1, gX1, curveB)

	gX2 := newGFp(0)
	gfpMul(gX2, x2, x2)
	gfpMul(gX2, gX2, x2)
	gfpAdd(gX2, gX2, curveB)

	// We compute the Legendre symbols of gX1 and gX2 in order to determine
	// which value from {x1, x2, x3} will be the correct x coordinate.
	residue1 := gX1.Legendre()
	residue2 := gX2.Legendre()

	// We compute the index and determine which coefficients are 0 or 1.
	index := (residue1-1)*(residue2-3)/4 + 1
	coef1 := subtle.ConstantTimeEq(int32(1), int32(index))
	coef2 := subtle.ConstantTimeEq(int32(2), int32(index))
	coef3 := subtle.ConstantTimeEq(int32(3), int32(index))

	gfpCoef1 := newGFp(int64(coef1))
	gfpCoef2 := newGFp(int64(coef2))
	gfpCoef3 := newGFp(int64(coef3))

	potentialX1 := &gfP{}
	gfpMul(potentialX1, gfpCoef1, x1)
	potentialX2 := &gfP{}
	gfpMul(potentialX2, gfpCoef2, x2)
	potentialX3 := &gfP{}
	gfpMul(potentialX3, gfpCoef3, x3)

	// We compute the correct x coordinate from
	//
	//		x == coef1*x1 + coef2*x2 + coef3*x3
	x := &gfP{}
	gfpAdd(x, x, potentialX1)
	gfpAdd(x, x, potentialX2)
	gfpAdd(x, x, potentialX3)

	// We have
	//
	//		g == g(x)
	//		  == x^3 + curveB
	//
	// By design, x is the x-coordinate of a valid curve point and so
	// g has a square root.
	g := &gfP{}
	gfpMul(g, x, x)
	gfpMul(g, g, x)
	gfpAdd(g, g, curveB)

	// Use of sign0 modification of FT paper; change from Wahby and Boneh 2019
	y := &gfP{}
	y.Sqrt(g)
	ySign := newGFp(int64(sign0(t)))
	gfpMul(y, y, ySign)

	p := &curvePoint{}
	p.x.Set(x)
	p.y.Set(y)
	p.z.Set(gfpOne)
	p.t.Set(gfpOne)

	h := &G1{}
	h.p = &curvePoint{}
	h.p.Set(p)
	if !h.p.IsOnCurve() {
		return nil, ErrInvalidPoint
	}

	return h, nil
}

// HashToG2 maps a byte slice into G2; while baseToTwist maps only to the
// twist curve, we clear the cofactor to ensure the hash point lies on G2.
// Because we map from Fp2 to Twist in the same way as above in HashToG1,
// it is likely the proof carries over, ensuring that “HashToTwist”
// (HashToG2 without clearing the cofactor) is indistinguishable from
// a random oracle on Twist. If this were the case, it would follow
// that HashToG2 is indistinguishable from a random oracle.
// It is not currently known if HashToG2 is indistinguishable from
// a random oracle.
func HashToG2(msg []byte) (*G2, error) {
	g2HashPoint := &G2{}
	g2HashPoint.p = &twistPoint{}
	twistHashPoint := &twistPoint{}
	const dsp0 byte = 0x04
	const dsp1 byte = 0x05
	const dsp2 byte = 0x06
	const dsp3 byte = 0x07
	const dsp4 byte = 0x08
	const dsp5 byte = 0x09
	const dsp6 byte = 0x0a
	const dsp7 byte = 0x0b
	fieldElement0X := hashToBase(msg, dsp0, dsp1)
	fieldElement0Y := hashToBase(msg, dsp2, dsp3)
	fieldElement1X := hashToBase(msg, dsp4, dsp5)
	fieldElement1Y := hashToBase(msg, dsp6, dsp7)
	fieldElement0 := &gfP2{}
	fieldElement0.x.Set(fieldElement0X)
	fieldElement0.y.Set(fieldElement0Y)
	fieldElement1 := &gfP2{}
	fieldElement1.x.Set(fieldElement1X)
	fieldElement1.y.Set(fieldElement1Y)
	twistPoint0, err0 := baseToTwist(fieldElement0)
	if err0 != nil {
		return nil, err0
	}
	twistPoint1, err1 := baseToTwist(fieldElement1)
	if err1 != nil {
		return nil, err1
	}
	twistHashPoint.Add(twistPoint0, twistPoint1)
	twistHashPoint.ClearCofactor(twistHashPoint)
	if !twistHashPoint.IsOnCurve() {
		return nil, ErrInvalidPoint
	}
	g2HashPoint.p.Set(twistHashPoint)
	return g2HashPoint, nil
}

// baseToTwist follows the same idea as baseToG1 with a few changes.
// The major change is that all sqrt and inversion calculations are done
// by arithmetic in gfP2 (naturally). The square root calculations can be
// found in Gora Adj and Francisco Rodrı́guez-Henrı́quez's 2012 paper
// “Square root computation over even extension fields”.
// That paper also contains how to compute the Legendre symbol in gfP2.
//
// We let
//
//	g'(x) == x^3 + twistB
func baseToTwist(t *gfP2) (*twistPoint, error) {
	gfp2One := &gfP2{}
	gfp2One.SetOne()

	// We have
	//
	//		alphaPart1 == t^2
	//
	//		alphaPart2 == g'(1) + t^2
	//
	//		alpha 	   == (alphaPart1*alphaPart2)^(-1)
	//
	//		tmp 	   == (g'(1) + t^2)^3
	//				   == alphaPart2^3
	//
	//		tFourth	   == t^4
	//
	// Inversion is computed using multiplication, so if
	// alphaPart1 == 0 or alphaPart2 == 0, then alpha == 0.
	// This still leads to a valid value.
	alphaPart1 := &gfP2{}
	alphaPart1.Square(t)
	alphaPart2 := &gfP2{}
	alphaPart2.Add(alphaPart1, g2HashConst4)
	alpha := &gfP2{} // alpha == 0 when t == 0 or t^2 + g'(1) == 0
	alpha.Mul(alphaPart1, alphaPart2)
	alpha.Invert(alpha)
	tFourth := &gfP2{}
	tFourth.Square(alphaPart1)
	tmp := &gfP2{}
	tmp.Square(alphaPart2)
	tmp.Mul(tmp, alphaPart2)

	// x1, x2, and x3 are possible x coordinates for
	// this baseToTwist function. We have
	//
	//		x1 == g2HashConst1 - g2HashConst2*tFourth*alpha
	//
	//		x2 == -1 - x1
	//
	//		x3 == 1 - g2HashConst3*tmp*alpha
	x1 := &gfP2{}
	x1.Mul(g2HashConst2, tFourth)
	x1.Mul(x1, alpha)
	x1.Neg(x1)
	x1.Add(g2HashConst1, x1)

	x2 := &gfP2{}
	x2.Add(x1, gfp2One)
	x2.Neg(x2)

	x3 := &gfP2{}
	x3.Mul(g2HashConst3, tmp)
	x3.Mul(x3, alpha)
	x3.Neg(x3)
	x3.Add(x3, gfp2One)

	// We now look at x1 and x2 to determine if either of these coordinates
	// form valid points on the elliptic curve. We have
	//
	//		gPrimeX1 == g'(x1)
	//				 == x1^3 + twistB
	//		gPrimeX2 == g'(x2)
	//				 == x2^3 + twistB
	gPrimeX1 := &gfP2{}
	gPrimeX1.Square(x1)
	gPrimeX1.Mul(gPrimeX1, x1)
	gPrimeX1.Add(gPrimeX1, twistB)

	gPrimeX2 := &gfP2{}
	gPrimeX2.Square(x2)
	gPrimeX2.Mul(gPrimeX2, x2)
	gPrimeX2.Add(gPrimeX2, twistB)

	// We compute the Legendre symbols of gX1 and gX2 in order to determine
	// which value from {x1, x2, x3} will be the correct x coordinate.
	residue1 := gPrimeX1.Legendre()
	residue2 := gPrimeX2.Legendre()

	// We compute the index and determine which coefficients are 0 or 1.
	index := (residue1-1)*(residue2-3)/4 + 1
	coef1 := subtle.ConstantTimeEq(int32(1), int32(index))
	coef2 := subtle.ConstantTimeEq(int32(2), int32(index))
	coef3 := subtle.ConstantTimeEq(int32(3), int32(index))

	gfpCoef1 := newGFp(int64(coef1))
	gfpCoef2 := newGFp(int64(coef2))
	gfpCoef3 := newGFp(int64(coef3))

	potentialX1 := &gfP2{}
	potentialX1.MulScalar(x1, gfpCoef1)
	potentialX2 := &gfP2{}
	potentialX2.MulScalar(x2, gfpCoef2)
	potentialX3 := &gfP2{}
	potentialX3.MulScalar(x3, gfpCoef3)

	// We compute the correct x coordinate from
	//
	//		x == coef1*x1 + coef2*x2 + coef3*x3
	x := &gfP2{}
	x.SetZero()
	x.Add(x, potentialX1)
	x.Add(x, potentialX2)
	x.Add(x, potentialX3)

	// We have
	//
	//		gPrime == g'(x)
	//		  	   == x^3 + twistB
	//
	// By design, x is the x-coordinate of a valid twist point and so
	// g has a square root.
	gPrime := &gfP2{}
	gPrime.Square(x)
	gPrime.Mul(gPrime, x)
	gPrime.Add(gPrime, twistB)

	// Use of sign0 modification of FT paper; change from Wahby and Boneh 2019
	y := &gfP2{}
	y.Sqrt(gPrime)
	ySign := newGFp(int64(sign0GFp2(t)))
	y.MulScalar(y, ySign)

	p := &twistPoint{}
	p.x.Set(x)
	p.y.Set(y)
	p.z.Set(gfp2One)
	p.t.Set(gfp2One)

	if !p.IsOnTwist() {
		return nil, ErrInvalidPoint
	}

	return p, nil
}

// convertBigToUint64Array converts big.Int by splitting the least significant
// 256 bits into an array of four uint64; this is only used in tests.
func convertBigToUint64Array(b *big.Int) [4]uint64 {
	one := big.NewInt(1)
	twoTo256 := new(big.Int).Lsh(one, uint(256))
	bMod := new(big.Int).Mod(b, twoTo256)

	v := bMod.Bytes()
	v32 := [32]byte{}
	for i := len(v) - 1; i >= 0; i-- {
		v32[len(v)-1-i] = v[i]
	}
	bits := [4]uint64{
		binary.LittleEndian.Uint64(v32[0*8 : 1*8]),
		binary.LittleEndian.Uint64(v32[1*8 : 2*8]),
		binary.LittleEndian.Uint64(v32[2*8 : 3*8]),
		binary.LittleEndian.Uint64(v32[3*8 : 4*8]),
	}
	return bits
}
