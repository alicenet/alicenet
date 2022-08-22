package cloudflare

import (
	"math/big"
	"testing"
)

// Tests that negation works the same way on both assembly-optimized and pure Go
// implementation.
func TestGFpNeg(t *testing.T) {
	n := &gfP{0x0123456789abcdef, 0xfedcba9876543210, 0xdeadbeefdeadbeef, 0xfeebdaedfeebdaed}
	w := &gfP{0xfedcba9876543211, 0x0123456789abcdef, 0x2152411021524110, 0x0114251201142512}
	h := &gfP{}

	gfpNeg(h, n)
	if *h != *w {
		t.Errorf("negation mismatch: have %#x, want %#x", *h, *w)
	}
}

// Tests that addition works the same way on both assembly-optimized and pure Go
// implementation.
func TestGFpAdd(t *testing.T) {
	a := &gfP{0x0123456789abcdef, 0xfedcba9876543210, 0xdeadbeefdeadbeef, 0xfeebdaedfeebdaed}
	b := &gfP{0xfedcba9876543210, 0x0123456789abcdef, 0xfeebdaedfeebdaed, 0xdeadbeefdeadbeef}
	w := &gfP{0xc3df73e9278302b8, 0x687e956e978e3572, 0x254954275c18417f, 0xad354b6afc67f9b4}
	h := &gfP{}

	gfpAdd(h, a, b)
	if *h != *w {
		t.Errorf("addition mismatch: have %#x, want %#x", *h, *w)
	}
}

// Tests that subtraction works the same way on both assembly-optimized and pure Go
// implementation.
func TestGFpSub(t *testing.T) {
	a := &gfP{0x0123456789abcdef, 0xfedcba9876543210, 0xdeadbeefdeadbeef, 0xfeebdaedfeebdaed}
	b := &gfP{0xfedcba9876543210, 0x0123456789abcdef, 0xfeebdaedfeebdaed, 0xdeadbeefdeadbeef}
	w := &gfP{0x02468acf13579bdf, 0xfdb97530eca86420, 0xdfc1e401dfc1e402, 0x203e1bfe203e1bfd}
	h := &gfP{}

	gfpSub(h, a, b)
	if *h != *w {
		t.Errorf("subtraction mismatch: have %#x, want %#x", *h, *w)
	}
}

// Tests that multiplication works the same way on both assembly-optimized and pure Go
// implementation.
func TestGFpMul(t *testing.T) {
	a := &gfP{0x0123456789abcdef, 0xfedcba9876543210, 0xdeadbeefdeadbeef, 0xfeebdaedfeebdaed}
	b := &gfP{0xfedcba9876543210, 0x0123456789abcdef, 0xfeebdaedfeebdaed, 0xdeadbeefdeadbeef}
	w := &gfP{0xcbcbd377f7ad22d3, 0x3b89ba5d849379bf, 0x87b61627bd38b6d2, 0xc44052a2a0e654b2}
	h := &gfP{}

	gfpMul(h, a, b)
	if *h != *w {
		t.Errorf("multiplication mismatch: have %#x, want %#x", *h, *w)
	}
}

// Additional Tests

// Here are the standard numbers that we use to test various encodings.
// These numbers were generated using ProbablyPrime(50) from the Go
// library. They are all of the form 2^e + k and k was chosen for each e
// to be the smallest (probable) prime number. What is important
// about these numbers is that they > 2^64 and so are not int64 or uint64.
//
//		2^e + k
//
// 		e;  k
//
//		64;  13			128;  51		192; 133
// 		65; 131			129;  17		193;  65
// 		66;   9			130; 169		194;  27
// 		67;   3			131;  39		195;  35
// 		68;  33			132;  67		196;  21
// 		69;  29			133;  27		197; 107
// 		70;  25			134;  27		198;  15
// 		71;  11			135;  33		199; 101
// 		72;  15			136;  85		200; 235
// 		73;  29			137; 155		201; 351
// 		74;  37			138;  87		202;  67
// 		75;  33			139; 155		203;  15
//
// Thus, the entry (64; 13) means 2^64 + 13 is a (probable) prime.

// Test Montgomery encoding and decoding
func TestEcDc(t *testing.T) {
	var k int64
	var start int64 = -10
	var stop int64 = 11

	// Test int64
	for k = start; k < stop; k++ {
		ecDcInt64(t, k)
	}

	// Test outside of int64 range
	eArr := []uint{64, 65, 66, 67, 128, 129, 130, 131, 192, 193, 194, 195}
	kArr := []uint{13, 131, 9, 3, 51, 17, 169, 39, 133, 65, 27, 35}
	if len(eArr) != len(kArr) {
		t.Fatal("eArr and kArr have different lengths")
	}
	for ell := 0; ell < len(eArr); ell++ {
		e := eArr[ell]
		k := kArr[ell]
		ecDcBig(t, e, k)
	}
}

// Test encoding and decoding for int64
func ecDcInt64(t *testing.T, g int64) {
	gGFp := newGFp(g)
	dec := newGFp(0)
	montDecode(dec, gGFp)
	bigDec := big.NewInt(0)
	bigDec.SetString(dec.String(), 16)
	bigG := big.NewInt(g)
	bigG.Mod(bigG, P)
	if bigDec.Cmp(bigG) != 0 {
		t.Fatal("Failed to properly encode", g)
		return
	}
	enc := newGFp(0)
	montEncode(enc, dec)
	if !enc.IsEqual(gGFp) {
		t.Fatal("Failed to properly re-encode", g)
		return
	}
}

func ecDcBig(t *testing.T, e uint, k uint) {
	bigVal := makeSpecialBig(e, k)
	bigValGFpDec := bigToDecodedGFp(bigVal)
	gfpVal := makeDecodedSpecialGFp(e, k)
	if !gfpVal.IsEqual(bigValGFpDec) {
		t.Fatal("Something major is wrong")
	}

	// At this point, we have the same decoded values.
	gfpValEnc := &gfP{}
	montEncode(gfpValEnc, gfpVal)
	bigValEnc := montEncodeBig(bigVal)
	bigGFpValEnc, _ := new(big.Int).SetString(gfpValEnc.String(), 16)
	if bigGFpValEnc.Cmp(bigValEnc) != 0 {
		t.Fatal("Failed to properly encode", bigVal)
	}

	gfpValEncDec := &gfP{}
	montDecode(gfpValEncDec, gfpValEnc)
	if !gfpVal.IsEqual(gfpValEncDec) {
		t.Fatal("Failed to properly decode", bigVal)
	}
}

func makeSpecialBig(e uint, k uint) *big.Int {
	// Make big.Int 2^e + k
	val := big.NewInt(0)
	if e < 256 {
		val.SetInt64(1)
		val.Lsh(val, e)
	}
	val.Add(val, big.NewInt(int64(k)))
	return val
}

func montEncodeBig(a *big.Int) *big.Int {
	// Given a, the Montgomery encoded form is aR mod P, where
	// R = 2^256 mod P
	R := big.NewInt(1)
	R.Lsh(R, 256)
	val := new(big.Int).Mul(a, R)
	val.Mod(val, P)
	return val
}

func bigToDecodedGFp(b *big.Int) *gfP {
	var ret *gfP
	// We assume b is nonnegative
	c := new(big.Int).Abs(b)
	R := big.NewInt(1)
	R.Lsh(R, 256)
	if c.Cmp(R) >= 0 {
		// This means c >= 2^256, so too large
		return ret
	}
	modVal := big.NewInt(1)
	modVal.Lsh(modVal, 64)
	d := big.NewInt(0)
	var d0, d1, d2, d3 uint64

	d.Mod(c, modVal)
	d0 = d.Uint64()
	c.Rsh(c, 64)
	d.Mod(c, modVal)
	d1 = d.Uint64()
	c.Rsh(c, 64)
	d.Mod(c, modVal)
	d2 = d.Uint64()
	c.Rsh(c, 64)
	d.Mod(c, modVal)
	d3 = d.Uint64()
	ret = &gfP{d0, d1, d2, d3}
	return ret
}

func makeDecodedSpecialGFp(e, k uint) *gfP {
	var val *gfP
	var eShift uint
	if e < 64 {
		// We assume this does not overflow!
		s := (1 << e) + k
		val = &gfP{uint64(s), 0, 0, 0}
	} else if e < 128 {
		eShift = e - 64
		val = &gfP{uint64(k), 1 << eShift, 0, 0}
	} else if e < 192 {
		eShift = e - 128
		val = &gfP{uint64(k), 0, 1 << eShift, 0}
	} else if e < 256 {
		eShift = e - 192
		val = &gfP{uint64(k), 0, 0, 1 << eShift}
	} else {
		// ``Overflow''
		val = &gfP{0, 0, 0, 0}
	}
	return val
}

func makeSpecialGFp(e uint, k uint) *gfP {
	v := makeDecodedSpecialGFp(e, k)
	montEncode(v, v)
	return v
}

func BenchmarkBigToGFp(b *testing.B) {
	bigVal, _ := new(big.Int).SetString("3141592653589793238462643383279502884197169399375105820974944592307816406286", 10) // 76
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_ = bigToGFp(bigVal)
	}
}

func TestBigToGFp(t *testing.T) {
	var k int64
	var start int64 = -10
	var stop int64 = 11

	// Test int64
	for k = start; k < stop; k++ {
		bigToGFpInt64(t, k)
	}

	// Look at 2^j + 1
	var pow2p1 int64
	var startPow2 uint = 1
	var stopPow2 uint = 63
	for j := startPow2; j < stopPow2; j++ {
		pow2p1 = (1 << j) + 1
		bigToGFpInt64(t, pow2p1)
	}

	// Test outside int64 range
	eArr := []uint{64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75,
		128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139,
		192, 193, 194, 195, 196, 197, 198, 199, 200, 201, 202, 203}
	kArr := []uint{13, 131, 9, 3, 33, 29, 25, 11, 15, 29, 37, 33,
		51, 17, 169, 39, 67, 27, 27, 33, 85, 155, 87, 155,
		133, 65, 27, 35, 21, 107, 15, 101, 235, 351, 67, 15}
	if len(eArr) != len(kArr) {
		t.Fatal("eArr and kArr have different lengths")
	}
	for ell := 0; ell < len(eArr); ell++ {
		e := eArr[ell]
		k := kArr[ell]
		bigToGFpOutside(t, e, k)
	}
}

func bigToGFpInt64(t *testing.T, g int64) {
	bigG := big.NewInt(g)
	btgG := bigToGFp(bigG)
	gfpG := newGFp(g)
	if !btgG.IsEqual(gfpG) {
		t.Fatal("bigToGFp failed produce correct int64", g)
		return
	}
}

func bigToGFpOutside(t *testing.T, e uint, k uint) {
	bigVal := makeSpecialBig(e, k)
	bigValGFp := bigToGFp(bigVal)
	gfpVal := makeSpecialGFp(e, k)
	if !bigValGFp.IsEqual(gfpVal) {
		t.Fatal("bigToGFp failed to produce correct big", bigVal)
	}
}

// Test basic arithmetic (addition, subtraction, multiplication)
func TestBasicArithmetic(t *testing.T) {
	var i, j, h int64
	var start int64 = -10
	var stop int64 = 11

	// Test int64
	for i = start; i < stop; i++ {
		for j = start; j < stop; j++ {
			h = i + j
			addInt64(t, i, j, h)
			h = i - j
			subInt64(t, i, j, h)
			h = i * j
			multiplyInt64(t, i, j, h)
		}
	}

	// Test outside int64 range
	eArr := []uint{64, 65, 66, 67, 128, 129, 130, 131, 192, 193, 194, 195}
	kArr := []uint{13, 131, 9, 3, 51, 17, 169, 39, 133, 65, 27, 35}
	if len(eArr) != len(kArr) {
		t.Fatal("eArr and kArr have different lengths")
	}
	for m := 0; m < len(eArr); m++ {
		e := eArr[m]
		k := kArr[m]
		for n := 0; n < len(kArr); n++ {
			f := eArr[n]
			ell := kArr[n]
			addOutside(t, e, k, f, ell)
			subOutside(t, e, k, f, ell)
			multiplyOutside(t, e, k, f, ell)
		}
	}
}

func addInt64(t *testing.T, k int64, j int64, h int64) {
	gfpK := newGFp(k)
	gfpJ := newGFp(j)
	gfpH := newGFp(h)
	gfpR := newGFp(0)
	gfpAdd(gfpR, gfpK, gfpJ)
	if !gfpH.IsEqual(gfpR) {
		t.Fatal(k, "+", j, "!=", h)
		return
	}
}

func addOutside(t *testing.T, e uint, k uint, f uint, ell uint) {
	bigV1 := makeSpecialBig(e, k)
	bigV2 := makeSpecialBig(f, ell)
	bigRes := new(big.Int).Add(bigV1, bigV2)
	bigRes.Mod(bigRes, P)
	bigResGFp := bigToGFp(bigRes)
	gfpV1 := makeSpecialGFp(e, k)
	gfpV2 := makeSpecialGFp(f, ell)
	gfpRes := &gfP{}
	gfpAdd(gfpRes, gfpV1, gfpV2)
	if !gfpRes.IsEqual(bigResGFp) {
		t.Fatal(bigV1, "+", bigV2, "!=", bigRes)
	}
}

func subInt64(t *testing.T, k int64, j int64, h int64) {
	gfpK := newGFp(k)
	gfpJ := newGFp(j)
	gfpH := newGFp(h)
	gfpR := newGFp(0)
	gfpSub(gfpR, gfpK, gfpJ)
	if !gfpH.IsEqual(gfpR) {
		t.Fatal(k, "-", j, "!=", h)
		return
	}
}

func subOutside(t *testing.T, e uint, k uint, f uint, ell uint) {
	bigV1 := makeSpecialBig(e, k)
	bigV2 := makeSpecialBig(f, ell)
	bigRes := new(big.Int).Sub(bigV1, bigV2)
	bigRes.Mod(bigRes, P)
	bigResGFp := bigToGFp(bigRes)
	gfpV1 := makeSpecialGFp(e, k)
	gfpV2 := makeSpecialGFp(f, ell)
	gfpRes := &gfP{}
	gfpSub(gfpRes, gfpV1, gfpV2)
	if !gfpRes.IsEqual(bigResGFp) {
		t.Fatal(bigV1, "-", bigV2, "!=", bigRes)
	}
}

func multiplyInt64(t *testing.T, k int64, j int64, h int64) {
	gfpK := newGFp(k)
	gfpJ := newGFp(j)
	gfpH := newGFp(h)
	gfpR := newGFp(0)
	gfpMul(gfpR, gfpK, gfpJ)
	if !gfpH.IsEqual(gfpR) {
		t.Fatal(k, "*", j, "!=", h)
		return
	}
}

func multiplyOutside(t *testing.T, e uint, k uint, f uint, ell uint) {
	bigV1 := makeSpecialBig(e, k)
	bigV2 := makeSpecialBig(f, ell)
	bigRes := new(big.Int).Mul(bigV1, bigV2)
	bigRes.Mod(bigRes, P)
	bigResGFp := bigToGFp(bigRes)
	gfpV1 := makeSpecialGFp(e, k)
	gfpV2 := makeSpecialGFp(f, ell)
	gfpRes := &gfP{}
	gfpMul(gfpRes, gfpV1, gfpV2)
	if !gfpRes.IsEqual(bigResGFp) {
		t.Fatal(bigV1, "*", bigV2, "!=", bigRes)
	}
}

func TestInvert(t *testing.T) {
	var j int64
	var start int64 = -10
	var stop int64 = 11

	// Test int64
	for j = start; j < stop; j++ {
		if j != 0 {
			invertInt64(t, j)
		} else if j == 0 {
			gfpZero := newGFp(0)
			gfpRes := newGFp(0)
			gfpRes.Invert(gfpZero)
			if !gfpRes.IsEqual(gfpZero) {
				t.Fatal("Failed to `invert` of 0 as 0")
			}
		}
	}

	// Test outside int64 range
	eArr := []uint{64, 65, 66, 67, 128, 129, 130, 131, 192, 193, 194, 195}
	kArr := []uint{13, 131, 9, 3, 51, 17, 169, 39, 133, 65, 27, 35}
	if len(eArr) != len(kArr) {
		t.Fatal("eArr and kArr have different lengths")
	}
	for ell := 0; ell < len(eArr); ell++ {
		e := eArr[ell]
		k := kArr[ell]
		invertOutside(t, e, k)
	}
}

func invertInt64(t *testing.T, g int64) {
	gfpG := newGFp(g)
	gfpRes := newGFp(0)
	gfpProd := newGFp(0)
	gfpOne := newGFp(1)

	gfpRes.Invert(gfpG)
	gfpMul(gfpProd, gfpG, gfpRes)
	if !gfpProd.IsEqual(gfpOne) {
		t.Fatal("Failed to invert", g)
	}
}

func invertOutside(t *testing.T, e uint, k uint) {
	gfpOne := newGFp(1)
	bigVal := makeSpecialBig(e, k)
	gfpVal := makeSpecialGFp(e, k)
	gfpInv := &gfP{}
	gfpInv.Invert(gfpVal)
	gfpProd := &gfP{}
	gfpMul(gfpProd, gfpVal, gfpInv)
	if !gfpProd.IsEqual(gfpOne) {
		t.Fatal("Failed to invert", bigVal)
	}
}

// Test the Legrendre symbol
func TestLegendreGFP(t *testing.T) {
	var k int64
	var start int64 = -10
	var stop int64 = 11

	// Test int64
	for k = start; k < stop; k++ {
		legendreGFPInt64(t, k)
	}

	// Test outside int64 range
	eArr := []uint{64, 65, 66, 67, 128, 129, 130, 131, 192, 193, 194, 195}
	kArr := []uint{13, 131, 9, 3, 51, 17, 169, 39, 133, 65, 27, 35}
	if len(eArr) != len(kArr) {
		t.Fatal("eArr and kArr have different lengths")
	}
	for ell := 0; ell < len(eArr); ell++ {
		e := eArr[ell]
		k := kArr[ell]
		legendreGFPOutside(t, e, k)
	}
}

func legendreGFPInt64(t *testing.T, g int64) {
	bigG := big.NewInt(g)
	gfpG := newGFp(g)

	bigRes := big.Jacobi(bigG, P)
	gfpRes := gfpG.Legendre()

	if bigRes != gfpRes {
		t.Fatal("bigInt and GFp failed to agree on Legendre symbol for", g)
	}
}

func legendreGFPOutside(t *testing.T, e uint, k uint) {
	bigVal := makeSpecialBig(e, k)
	gfpVal := makeSpecialGFp(e, k)
	gfpLeg := gfpVal.Legendre()
	bigLeg := big.Jacobi(bigVal, P)
	if bigLeg != gfpLeg {
		t.Fatal("bigInt and GFp failed to agree on Legendre symbol for", bigVal)
	}
}

func TestSqrtGFP(t *testing.T) {
	var k int64
	var start int64 = -10
	var stop int64 = 11

	// Test int64
	for k = start; k < stop; k++ {
		sqrtGFPInt64(t, k)
	}
	// Test outside int64 range
	eArr := []uint{64, 65, 66, 67, 128, 129, 130, 131, 192, 193, 194, 195}
	kArr := []uint{13, 131, 9, 3, 51, 17, 169, 39, 133, 65, 27, 35}
	if len(eArr) != len(kArr) {
		t.Fatal("eArr and kArr have different lengths")
	}
	for ell := 0; ell < len(eArr); ell++ {
		e := eArr[ell]
		k := kArr[ell]
		sqrtGFPOutside(t, e, k)
	}
}

func sqrtGFPInt64(t *testing.T, g int64) {
	gfpG := newGFp(g)
	if gfpG.Legendre() != 1 {
		// No square root exists; exit
		return
	}
	gfpSqrt := newGFp(0)
	gfpTest := newGFp(0)
	gfpSqrt.Sqrt(gfpG)
	gfpMul(gfpTest, gfpSqrt, gfpSqrt)
	if !gfpTest.IsEqual(gfpG) {
		t.Fatal("Error computing square root of", g)
	}
}

func sqrtGFPOutside(t *testing.T, e uint, k uint) {
	bigVal := makeSpecialBig(e, k)
	gfpVal := makeSpecialGFp(e, k)
	if gfpVal.Legendre() != 1 {
		// No square root exists; exit
		return
	}
	gfpSqrt := newGFp(0)
	gfpTest := newGFp(0)
	gfpSqrt.Sqrt(gfpVal)
	gfpMul(gfpTest, gfpSqrt, gfpSqrt)
	if !gfpTest.IsEqual(gfpVal) {
		t.Fatal("Error computing square root of", bigVal)
	}
}

func TestExponentiation(t *testing.T) {
	var g int64

	for _, g = range []int64{2, 3, 5, 7, 11, 13, 17, 19, 23} {
		exponentiationInt64(t, g)
	}
	// Test outside int64 range
	eArr := []uint{64, 67, 128, 131, 192, 195}
	kArr := []uint{13, 3, 51, 39, 133, 35}
	if len(eArr) != len(kArr) {
		t.Fatal("eArr and kArr have different lengths")
	}
	for ell := 0; ell < len(eArr); ell++ {
		e := eArr[ell]
		k := kArr[ell]
		exponentiationOutside(t, e, k)
	}
}

func exponentiationInt64(t *testing.T, g int64) {
	var k uint
	gfpG := newGFp(g)
	gfpRes := newGFp(0)
	bigG := big.NewInt(g)
	bigPower := big.NewInt(0)
	bigRes := big.NewInt(0)
	for k = 0; k < 256; k++ {
		// loop through all powers
		b := big.NewInt(1)
		bigPower.Lsh(b, k)
		bits := convertBigToUint64Array(bigPower)
		bigRes.Exp(bigG, bigPower, P)
		gfpRes.exp(gfpG, bits)
		gfpBR := bigToGFp(bigRes)
		if !gfpRes.IsEqual(gfpBR) {
			t.Fatal("Failed to agree on exponentiation with g =", g, "and k =", k)
		}
	}
}

func exponentiationOutside(t *testing.T, e uint, k uint) {
	var j uint
	bigVal := makeSpecialBig(e, k)
	gfpVal := makeSpecialGFp(e, k)
	gfpRes := newGFp(0)
	bigPower := big.NewInt(0)
	bigRes := big.NewInt(0)
	for j = 0; j < 256; j++ {
		// loop through all powers
		b := big.NewInt(1)
		bigPower.Lsh(b, j)
		bits := convertBigToUint64Array(bigPower)
		bigRes.Exp(bigVal, bigPower, P)
		gfpRes.exp(gfpVal, bits)
		gfpBR := bigToGFp(bigRes)
		if !gfpRes.IsEqual(gfpBR) {
			t.Fatal("Failed to agree on exponentiation with", bigVal, "and j =", j)
		}
	}
}

// Test the sign0 function for determining sign of gfP
func TestSign0(t *testing.T) {
	gfpTwo := newGFp(2)
	if sign0(gfpTwo) != 1 {
		t.Fatal("Failed to compute correct sign of 2")
	}

	gfpOne := newGFp(1)
	if sign0(gfpOne) != 1 {
		t.Fatal("Failed to compute correct sign of 1")
	}

	gfpZero := newGFp(0)
	if sign0(gfpZero) != 1 {
		t.Fatal("Failed to compute correct sign of 0")
	}

	gfpNegOne := newGFp(-1)
	if sign0(gfpNegOne) != -1 {
		t.Fatal("Failed to compute correct sign of -1")
	}

	gfpNegTwo := newGFp(-2)
	if sign0(gfpNegTwo) != -1 {
		t.Fatal("Failed to compute correct sign of -2")
	}

	gfpPM1O2 := bigToGFp(pMinus1Over2Big)
	if sign0(gfpPM1O2) != 1 {
		t.Fatal("Failed to compute correct sign of (P-1)/2")
	}

	gfpPM1O2p1 := &gfP{}
	gfpAdd(gfpPM1O2p1, gfpPM1O2, gfpOne)
	if sign0(gfpPM1O2p1) != -1 {
		t.Fatal("Failed to compute correct sign of (P-1)/2+1")
	}

	gfpPM1O2p2 := &gfP{}
	gfpAdd(gfpPM1O2p2, gfpPM1O2, gfpTwo)
	if sign0(gfpPM1O2p2) != -1 {
		t.Fatal("Failed to compute correct sign of (P-1)/2+2")
	}

	gfpPM1O2n1 := &gfP{}
	gfpAdd(gfpPM1O2n1, gfpPM1O2, gfpNegOne)
	if sign0(gfpPM1O2n1) != 1 {
		t.Fatal("Failed to compute correct sign of (P-1)/2-1")
	}

	gfpPM1O2n2 := &gfP{}
	gfpAdd(gfpPM1O2n2, gfpPM1O2, gfpNegTwo)
	if sign0(gfpPM1O2n2) != 1 {
		t.Fatal("Failed to compute correct sign of (P-1)/2-2")
	}
}

func TestIsEqualGFP(t *testing.T) {
	z1 := newGFp(0)
	z2 := newGFp(0)
	one := newGFp(1)
	if !z1.IsEqual(z2) {
		t.Fatal("IsEqual (GFp) failed to show 0 == 0")
	}
	if z1.IsEqual(one) {
		t.Fatal("IsEqual (GFp) showed 0 == 1")
	}
}

func TestMarshalGFP(t *testing.T) {
	for _, k := range []int{-1, 0, 1} {
		gfpMarshalTest(k, t)
	}
	gfpMarshalTestExceed(t)
	gfpMarshalTestEqual(t)
}

func gfpMarshalTest(k int, t *testing.T) {
	gfpK := newGFp(int64(k))
	retK := make([]byte, numBytes)
	gfpK.Marshal(retK)
	gfpKcopy := newGFp(int64(k))
	unmarshK := &gfP{}
	err := unmarshK.Unmarshal(retK)
	if err != nil {
		t.Fatalf("Error in unmarshalling %d in GFp", k)
	}
	if !gfpKcopy.IsEqual(unmarshK) {
		t.Fatalf("Failed to unmarshal to %d in GFp", k)
	}
}

func gfpMarshalTestExceed(t *testing.T) {
	breakBytesExceedGFp := &gfP{}
	breakBytesExceed := make([]byte, numBytes)
	breakBytesExceed[0] = byte(255) // >= 49 will work
	errBBE := breakBytesExceedGFp.Unmarshal(breakBytesExceed)
	if errBBE == nil {
		t.Fatal("Failed to raise error; coordinate exceeds modulus")
	}
}

func gfpMarshalTestEqual(t *testing.T) {
	// The exact coordinates of the modulus;
	// this can be computed by starting with p2[3] and taking the largest byte
	// (that is, the top 2 hex digits) and working down the uint64 before heading
	// to p2[2] and down.
	breakGFp := &gfP{}
	breakBytes := make([]byte, numBytes)

	breakBytes[0] = 48
	breakBytes[1] = 100
	breakBytes[2] = 78
	breakBytes[3] = 114
	breakBytes[4] = 225
	breakBytes[5] = 49
	breakBytes[6] = 160
	breakBytes[7] = 41

	breakBytes[8] = 184
	breakBytes[9] = 80
	breakBytes[10] = 69
	breakBytes[11] = 182
	breakBytes[12] = 129
	breakBytes[13] = 129
	breakBytes[14] = 88
	breakBytes[15] = 93

	breakBytes[16] = 151
	breakBytes[17] = 129
	breakBytes[18] = 106
	breakBytes[19] = 145
	breakBytes[20] = 104
	breakBytes[21] = 113
	breakBytes[22] = 202
	breakBytes[23] = 141

	breakBytes[24] = 60
	breakBytes[25] = 32
	breakBytes[26] = 140
	breakBytes[27] = 22
	breakBytes[28] = 216
	breakBytes[29] = 124
	breakBytes[30] = 253
	breakBytes[31] = 71

	err := breakGFp.Unmarshal(breakBytes)
	if err == nil {
		t.Fatal("Failed to raise error; coordinate equals modulus")
	}
}
