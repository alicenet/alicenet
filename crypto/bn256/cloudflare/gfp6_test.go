package cloudflare

import "testing"

func TestIsEqualGFP6(t *testing.T) {
	z1 := &gfP6{}
	z2 := &gfP6{}
	one := &gfP6{}
	z1.SetZero()
	z2.SetZero()
	one.SetOne()
	if !z1.IsEqual(z2) {
		t.Fatal("IsEqual (GFp6) failed to show 0 == 0")
	}
	if z1.IsEqual(one) {
		t.Fatal("IsEqual (GFp6) showed 0 == 1")
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
