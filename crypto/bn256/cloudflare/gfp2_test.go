package cloudflare

import (
	"fmt"
	"testing"
)

func TestIsEqualGFP2(t *testing.T) {
	z1 := &gfP2{}
	z2 := &gfP2{}
	one := &gfP2{}
	z1.SetZero()
	z2.SetZero()
	one.SetOne()
	if !z1.IsEqual(z2) {
		t.Fatal("IsEqual (GFp2) failed to show 0 == 0")
	}
	if z1.IsEqual(one) {
		t.Fatal("IsEqual (GFp2) showed 0 == 1")
	}
}

func TestEcDcGfP2(t *testing.T) {
	one := &gfP2{}
	one.SetOne()
	tmp := gfP2Decode(one)
	oneCopy := gfP2Encode(tmp)
	if !one.IsEqual(oneCopy) {
		t.Fatal("GFP2 Encode and Decode Failed!")
	}
}

func TestLegendreGFP2(t *testing.T) {
	// Testing Legendre call on GFP2
	gfp2Zero := &gfP2{}
	gfp2Zero.SetZero()
	zeroLeg := gfp2Zero.Legendre()
	if zeroLeg != 0 {
		t.Fatal("Failed to compute correct gfP2 legendre symbol for 0")
	}

	gfp2One := &gfP2{}
	gfp2One.SetOne()
	oneLeg := gfp2One.Legendre()
	if oneLeg != 1 {
		t.Fatal("Failed to compute correct gfP2 legendre symbol for 1")
	}

	gfp2NegOne := &gfP2{}
	gfp2NegOne.Neg(gfp2One)
	negOneLeg := gfp2NegOne.Legendre()
	if negOneLeg != 1 {
		t.Fatal("Failed to compute correct gfP2 legendre symbol for -1")
	}

	gfp2Elem := &gfP2{}
	gfp2Elem.x.Set(newGFp(1))
	gfp2Elem.y.Set(newGFp(2))
	leg := gfp2Elem.Legendre()
	if leg != -1 {
		t.Fatal("Failed to compute correct gfP2 legendre symbol for 1i + 2")
	}

	// xi = i + 9 is the nonsquare noncube from the twist
	gfp2Xi := &gfP2{}
	gfp2Xi.x.Set(newGFp(1))
	gfp2Xi.y.Set(newGFp(9))
	leg = gfp2Xi.Legendre()
	if leg != -1 {
		t.Fatal("Failed to compute correct gfP2 legendre symbol for xi = 1i + 9")
	}
}

func TestExpGFP2(t *testing.T) {
	gfp2Zero := &gfP2{}
	gfp2Zero.SetZero()
	gfp2ZeroRes := &gfP2{}
	gfp2ZeroRes.exp(gfp2Zero, pMinus3Over4)
	if !gfp2Zero.IsEqual(gfp2ZeroRes) {
		t.Fatal("Failed to exponentiate 0^(P-3)/4 in gfp2")
	}

	gfp2One := &gfP2{}
	gfp2One.SetOne()
	gfp2OneRes := &gfP2{}
	gfp2OneRes.exp(gfp2One, pMinus3Over4)
	if !gfp2One.IsEqual(gfp2OneRes) {
		t.Fatal("Failed to exponentiate 1^(P-3)/4 in gfp2")
	}

	gfp2NegOne := &gfP2{}
	gfp2NegOne.Neg(gfp2One)
	gfp2NegOneRes := &gfP2{}
	gfp2NegOneRes.exp(gfp2NegOne, pMinus3Over4)
	if !gfp2NegOne.IsEqual(gfp2NegOneRes) {
		t.Fatal("Failed to exponentiate (-1)^(P-3)/4 in gfp2")
	}

	gfp2NegOneRes.exp(gfp2NegOne, pPlus1Over4)
	if !gfp2One.IsEqual(gfp2NegOneRes) {
		t.Fatal("Failed to exponentiate (-1)^(P+1)/4 in gfp2")
	}

	gfp2NegOneRes.exp(gfp2NegOne, pPlus1)
	if !gfp2One.IsEqual(gfp2NegOneRes) {
		t.Fatal("Failed to exponentiate (-1)^(P+1) in gfp2")
	}

	xi := &gfP2{}
	xi.SetOne()
	xi.MulXi(xi)
	xiConj := &gfP2{}
	xiConj.Conjugate(xi)
	xiRes := &gfP2{}
	xiRes.exp(xi, pPlus1)
	xiNorm := &gfP2{}
	xiNorm.Mul(xi, xiConj)
	if !xiRes.IsEqual(xiNorm) {
		t.Fatal("Failed to exponentiate (xi)^(P+1) in gfp2")
	}
}

func TestSqrtGFP2(t *testing.T) {
	var k int64
	var start int64 = -10
	var stop int64 = 11
	for k = start; k < stop; k++ {
		sqrtGFP2Int64(t, k)
	}

	// Test i+1 (sqrt exists)
	gfp2Elem := &gfP2{}
	gfp2Elem.x.Set(newGFp(1))
	gfp2Elem.y.Set(newGFp(1))
	gfp2ElemSqrt := &gfP2{}
	gfp2ElemSqrt.Sqrt(gfp2Elem)
	gfp2ElemSqrtSquared := &gfP2{}
	gfp2ElemSqrtSquared.Square(gfp2ElemSqrt)
	if !gfp2ElemSqrtSquared.IsEqual(gfp2Elem) {
		t.Fatal("Failed to compute gfp2 sqrt of", gfp2Elem.String())
	}

	// Test -i+1 (sqrt exists)
	gfp2Elem.Conjugate(gfp2Elem)
	gfp2ElemSqrt.Sqrt(gfp2Elem)
	gfp2ElemSqrtSquared.Square(gfp2ElemSqrt)
	if !gfp2ElemSqrtSquared.IsEqual(gfp2Elem) {
		t.Fatal("Failed to compute gfp2 sqrt of", gfp2Elem.String())
	}

	// Test i-1 (sqrt exists)
	gfp2Elem.x.Set(newGFp(1))
	gfp2Elem.y.Set(newGFp(-1))
	gfp2ElemSqrt.Sqrt(gfp2Elem)
	gfp2ElemSqrtSquared.Square(gfp2ElemSqrt)
	if !gfp2ElemSqrtSquared.IsEqual(gfp2Elem) {
		t.Fatal("Failed to compute gfp2 sqrt of", gfp2Elem.String())
	}

	// Test -i-1 (sqrt exists)
	gfp2Elem.Conjugate(gfp2Elem)
	gfp2ElemSqrt.Sqrt(gfp2Elem)
	gfp2ElemSqrtSquared.Square(gfp2ElemSqrt)
	if !gfp2ElemSqrtSquared.IsEqual(gfp2Elem) {
		t.Fatal("Failed to compute gfp2 sqrt of", gfp2Elem.String())
	}

}

func sqrtGFP2Int64(t *testing.T, k int64) {
	// These square roots always exist
	gfp2RealK := &gfP2{}
	gfp2RealK.x.Set(newGFp(0))
	gfp2RealK.y.Set(newGFp(k))
	gfp2RealKSqrt := &gfP2{}
	gfp2RealKSqrt.Sqrt(gfp2RealK)
	gfp2RealKSqrtSquared := &gfP2{}
	gfp2RealKSqrtSquared.Square(gfp2RealKSqrt)
	if !gfp2RealKSqrtSquared.IsEqual(gfp2RealK) {
		t.Fatal("Failed to compute gfp2 sqrt of", gfp2RealK.String())
	}

	gfp2ImagK := &gfP2{}
	gfp2ImagK.x.Set(newGFp(k))
	gfp2ImagK.y.Set(newGFp(0))
	gfp2ImagKSqrt := &gfP2{}
	gfp2ImagKSqrt.Sqrt(gfp2ImagK)
	gfp2ImagKSqrtSquared := &gfP2{}
	gfp2ImagKSqrtSquared.Square(gfp2ImagKSqrt)
	if !gfp2ImagKSqrtSquared.IsEqual(gfp2ImagK) {
		t.Fatal("Failed to compute gfp2 sqrt of", gfp2ImagK.String())
	}
}

// Test the sign0GFp2 function for determining sign of gfP2 (Real part)
func TestSign0GFp2Real(t *testing.T) {
	gfpOne := newGFp(1)
	gfp2One := &gfP2{}
	gfp2One.y.Set(gfpOne)
	if sign0GFp2(gfp2One) != 1 {
		fmt.Println("Sign(1):", sign0(gfpOne))
		t.Fatal("Failed to compute correct sign of 1 (GFp2)")
	}

	gfp2Zero := &gfP2{}
	if sign0GFp2(gfp2Zero) != 1 {
		fmt.Println("Sign(0):", sign0GFp2(gfp2Zero))
		t.Fatal("Failed to compute correct sign of 0 (GFp2)")
	}

	gfpNegOne := newGFp(-1)
	gfp2NegOne := &gfP2{}
	gfp2NegOne.y.Set(gfpNegOne)
	if sign0GFp2(gfp2NegOne) != -1 {
		fmt.Println("Sign(-1):", sign0GFp2(gfp2NegOne))
		t.Fatal("Failed to compute correct sign of -1 (GFp2)")
	}

	gfpPM1O2 := bigToGFp(pMinus1Over2Big)
	gfp2PM1O2 := &gfP2{}
	gfp2PM1O2.y.Set(gfpPM1O2)
	if sign0GFp2(gfp2PM1O2) != 1 {
		fmt.Println("Sign(PM1O2):", sign0GFp2(gfp2PM1O2))
		t.Fatal("Failed to compute correct sign of (P-1)/2 (GFp2)")
	}

	gfp2PM1O2p1 := new(gfP2).Add(gfp2PM1O2, gfp2One)
	if sign0GFp2(gfp2PM1O2p1) != -1 {
		fmt.Println("Sign(PM1O2p1):", sign0GFp2(gfp2PM1O2p1))
		t.Fatal("Failed to compute correct sign of (P-1)/2+1 (GFp2)")
	}

	gfp2PM1O2n1 := new(gfP2).Add(gfp2PM1O2, gfp2NegOne)
	if sign0GFp2(gfp2PM1O2n1) != 1 {
		fmt.Println("Sign(PM1O2n1):", sign0GFp2(gfp2PM1O2n1))
		t.Fatal("Failed to compute correct sign of (P-1)/2-1 (GFp2)")
	}
}

// Test the sign0GFp2 function for determining sign of gfP2 (Imag part)
func TestSign0GFp2Imag(t *testing.T) {
	gfpOne := newGFp(1)
	gfp2One := &gfP2{}
	gfp2One.x.Set(gfpOne)
	if sign0GFp2(gfp2One) != 1 {
		fmt.Println("Sign(1):", sign0(gfpOne))
		t.Fatal("Failed to compute correct sign of 1*i (GFp2)")
	}

	gfpNegOne := newGFp(-1)
	gfp2NegOne := &gfP2{}
	gfp2NegOne.x.Set(gfpNegOne)
	if sign0GFp2(gfp2NegOne) != -1 {
		fmt.Println("Sign(-1):", sign0GFp2(gfp2NegOne))
		t.Fatal("Failed to compute correct sign of -1*i (GFp2)")
	}

	gfpPM1O2 := bigToGFp(pMinus1Over2Big)
	gfp2PM1O2 := &gfP2{}
	gfp2PM1O2.x.Set(gfpPM1O2)
	if sign0GFp2(gfp2PM1O2) != 1 {
		fmt.Println("Sign(PM1O2):", sign0GFp2(gfp2PM1O2))
		t.Fatal("Failed to compute correct sign of (P-1)/2*i (GFp2)")
	}

	gfp2PM1O2p1 := new(gfP2).Add(gfp2PM1O2, gfp2One)
	if sign0GFp2(gfp2PM1O2p1) != -1 {
		fmt.Println("Sign(PM1O2p1):", sign0GFp2(gfp2PM1O2p1))
		t.Fatal("Failed to compute correct sign of [(P-1)/2+1]*i (GFp2)")
	}

	gfp2PM1O2n1 := new(gfP2).Add(gfp2PM1O2, gfp2NegOne)
	if sign0GFp2(gfp2PM1O2n1) != 1 {
		fmt.Println("Sign(PM1O2n1):", sign0GFp2(gfp2PM1O2n1))
		t.Fatal("Failed to compute correct sign of [(P-1)/2-1]*i (GFp2)")
	}
}

func TestInvertGFP2(t *testing.T) {
	var k int64
	var start int64 = -10
	var stop int64 = 11
	for k = start; k < stop; k++ {
		invertGFP2Int64(t, k)
	}

	gfp2One := &gfP2{}
	gfp2One.SetOne()

	// Test i+1
	gfp2Elem := &gfP2{}
	gfp2Elem.x.Set(newGFp(1))
	gfp2Elem.y.Set(newGFp(1))
	gfp2ElemInverse := &gfP2{}
	gfp2ElemInverse.Invert(gfp2Elem)
	gfp2ElemInverseMul := &gfP2{}
	gfp2ElemInverseMul.Mul(gfp2Elem, gfp2ElemInverse)
	if !gfp2ElemInverseMul.IsEqual(gfp2One) {
		t.Fatal("Failed to compute gfp2 inverse of", gfp2Elem.String())
	}

	// Test -i+1
	gfp2Elem.Conjugate(gfp2Elem)
	gfp2ElemInverse.Invert(gfp2Elem)
	gfp2ElemInverseMul.Mul(gfp2Elem, gfp2ElemInverse)
	if !gfp2ElemInverseMul.IsEqual(gfp2One) {
		t.Fatal("Failed to compute gfp2 inverse of", gfp2Elem.String())
	}

	// Test i-1
	gfp2Elem.x.Set(newGFp(1))
	gfp2Elem.y.Set(newGFp(-1))
	gfp2ElemInverse.Invert(gfp2Elem)
	gfp2ElemInverseMul.Mul(gfp2Elem, gfp2ElemInverse)
	if !gfp2ElemInverseMul.IsEqual(gfp2One) {
		t.Fatal("Failed to compute gfp2 inverse of", gfp2Elem.String())
	}

	// Test -i-1
	gfp2Elem.Conjugate(gfp2Elem)
	gfp2ElemInverse.Invert(gfp2Elem)
	gfp2ElemInverseMul.Mul(gfp2Elem, gfp2ElemInverse)
	if !gfp2ElemInverseMul.IsEqual(gfp2One) {
		t.Fatal("Failed to compute gfp2 inverse of", gfp2Elem.String())
	}
}

func invertGFP2Int64(t *testing.T, k int64) {
	if k == 0 {
		return
	}
	gfp2One := &gfP2{}
	gfp2One.SetOne()
	gfp2RealK := &gfP2{}
	gfp2RealK.x.Set(newGFp(0))
	gfp2RealK.y.Set(newGFp(k))
	gfp2RealKInverse := &gfP2{}
	gfp2RealKInverse.Invert(gfp2RealK)
	gfp2RealKInverseMul := &gfP2{}
	gfp2RealKInverseMul.Mul(gfp2RealK, gfp2RealKInverse)
	if !gfp2RealKInverseMul.IsEqual(gfp2One) {
		t.Fatal("Failed to compute gfp2 inverse of", gfp2RealK.String())
	}

	gfp2ImagK := &gfP2{}
	gfp2ImagK.x.Set(newGFp(k))
	gfp2ImagK.y.Set(newGFp(0))
	gfp2ImagKInverse := &gfP2{}
	gfp2ImagKInverse.Invert(gfp2ImagK)
	gfp2ImagKInverseMul := &gfP2{}
	gfp2ImagKInverseMul.Mul(gfp2ImagK, gfp2ImagKInverse)
	if !gfp2ImagKInverseMul.IsEqual(gfp2One) {
		t.Fatal("Failed to compute gfp2 inverse of", gfp2ImagK.String())
	}
}

func TestGFP2MulSquare(t *testing.T) {
	p := &gfP2{}
	two := newGFp(2)
	three := newGFp(3)
	p.x.Set(two)
	p.y.Set(three)

	// Correct value
	a := &gfP2{}
	a.Set(p)
	b := &gfP2{}
	b.Set(p)
	c := &gfP2{}
	c.Mul(a, b)

	// Test multiplication
	q := &gfP2{}
	q.Set(p)
	q.Mul(q, q)
	if !q.IsEqual(c) {
		t.Fatal("gfP2 do not match (1)")
	}

	// Test squaring
	r := &gfP2{}
	r.Set(p)
	r.Square(r)
	if !r.IsEqual(c) {
		t.Fatal("gfP2 do not match (2)")
	}
}
