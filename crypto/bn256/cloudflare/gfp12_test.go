package cloudflare

import "testing"

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
