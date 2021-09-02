package cloudflare

import (
	"math/big"
	"testing"
)

func TestIsEqualGFP12(t *testing.T) {
	z1 := &gfP12{}
	z2 := &gfP12{}
	one := &gfP12{}
	z1.SetZero()
	z2.SetZero()
	one.SetOne()
	if !z1.IsEqual(z2) {
		t.Fatal("IsEqual (GFp12) failed to show 0 == 0")
	}
	if z1.IsEqual(one) {
		t.Fatal("IsEqual (GFp12) showed 0 == 1")
	}

	if !z1.IsZero() {
		t.Fatal("Should show z1 == 0!")
	}
	if z1.IsOne() {
		t.Fatal("Should show z1 != 1!")
	}
	if one.IsZero() {
		t.Fatal("Should show one != 0!")
	}
	if !one.IsOne() {
		t.Fatal("Should show one == 1!")
	}
}

func TestAddSubGFP12(t *testing.T) {
	zero := &gfP12{}
	zero.SetZero()
	one := &gfP12{}
	one.SetOne()
	negOne := &gfP12{}
	negOne.Neg(one)
	res := &gfP12{}
	res.Add(one, negOne)
	if !res.IsEqual(zero) {
		t.Fatal("Should show res == 0!")
	}

	res.Sub(one, one)
	if !res.IsEqual(zero) {
		t.Fatal("Should show res == 0!")
	}
}

func TestGFP12MulSquare(t *testing.T) {
	k := 1
	s := big.NewInt(int64(k))
	g1 := new(G1).ScalarBaseMult(s)
	g2 := new(G2).ScalarBaseMult(s)
	gT := Pair(g1, g2)
	p := &gfP12{}
	p.Set(gT.p)

	// Correct value
	a := &gfP12{}
	a.Set(p)
	b := &gfP12{}
	b.Set(p)
	c := &gfP12{}
	c.Mul(a, b)

	// Test multiplication
	q := &gfP12{}
	q.Set(p)
	q.Mul(q, q)
	if !q.IsEqual(c) {
		t.Fatal("gfP12 do not match (1)")
	}

	// Test squaring
	r := &gfP12{}
	r.Set(p)
	r.Square(r)
	if !r.IsEqual(c) {
		t.Fatal("gfP12 do not match (2)")
	}
}
