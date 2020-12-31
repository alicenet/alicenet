package cloudflare

import (
	"testing"
)

func TestTwistGen(t *testing.T) {
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("returnTwistGen failed")
	}

	if !tG.IsEqual(twistGenBig) {
		t.Fatal("Error with twistGenBig; not equal to twistGen")
	}
}

func TestIsEqualTwist(t *testing.T) {
	p1 := &twistPoint{}
	p1.Set(twistGen)
	p2 := &twistPoint{}
	p2.Set(twistGen)
	if !p1.IsEqual(p2) {
		t.Fatal("Did not determine that twistGen == twistGen")
	}

	p2.SetInfinity()
	if p1.IsEqual(p2) {
		t.Fatal("Did not determine that twistGen != Infinity")
	}

	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("IsEqualTwist changed twistGen")
	}
}

func TestTwistGenNeg(t *testing.T) {
	sum := &twistPoint{}
	sum.Add(twistGen, twistGenNeg)
	if !sum.IsInfinity() {
		t.Fatal("Error with twistGen and twistGenNeg; do not sum to identity")
	}

	if !twistGenNeg.IsEqual(twistGenNegBig) {
		t.Fatal("Error with twistGenNegBig; not equal to twistGenNeg")
	}

	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("IsEqualTwist changed twistGen")
	}
}

func TestIsOnCurveG2(t *testing.T) {
	p := &twistPoint{}

	p.Set(twistGen)
	if !p.IsOnCurve() {
		t.Fatal("Did not determine twistGen is on G2")
	}

	p.SetInfinity()
	if !p.IsOnCurve() {
		t.Fatal("Did not determine Infinity is on G2")
	}

	p.Set(twistGenNeg)
	if !p.IsOnCurve() {
		t.Fatal("Did not determine twistGenNeg is on G2")
	}

	gfp2One := &gfP2{}
	gfp2One.SetOne()
	gfp2Two := &gfP2{}
	gfp2Two.Add(gfp2One, gfp2One)
	p.x.Set(gfp2One)
	p.y.Set(gfp2Two)
	p.z.Set(gfp2One)
	p.t.Set(gfp2One)
	if p.IsOnCurve() {
		t.Fatal("Failed to determine (1,2) is not on G2")
	}

	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("IsOnCurve (twist) changed twistGen")
	}
}

func TestIsOnTwist(t *testing.T) {
	p := &twistPoint{}

	p.Set(twistGen)
	if !p.IsOnTwist() {
		t.Fatal("Did not determine twistGen is on twist")
	}

	p.SetInfinity()
	if !p.IsOnTwist() {
		t.Fatal("Did not determine Infinity is on twist")
	}

	gfp2Zero := &gfP2{}
	p, err := baseToTwist(gfp2Zero)
	if err != nil {
		t.Fatal(err)
	}
	if !p.IsOnTwist() {
		t.Fatal("Error occurred in baseToTwist for t = 0")
	}

	gfp2One := &gfP2{}
	gfp2One.SetOne()
	gfp2Two := &gfP2{}
	gfp2Two.Add(gfp2One, gfp2One)
	p.x.Set(gfp2One)
	p.y.Set(gfp2Two)
	p.z.Set(gfp2One)
	p.t.Set(gfp2One)
	if p.IsOnTwist() {
		t.Fatal("Failed to determine (1,2) is not on twist")
	}

	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("IsOnTwist changed twistGen")
	}
}

func TestClearCofactor(t *testing.T) {
	var k int64
	for k = 0; k < 10; k++ {
		gfp2K := &gfP2{}
		gfp2K.x.Set(newGFp(k))
		g, err := baseToTwist(gfp2K)
		if err != nil {
			t.Fatal(err)
		}
		g.ClearCofactor(g)
		if !g.IsOnTwist() {
			t.Fatal("Did not map to G2 from baseToTwist for t = i *", k)
		}
	}
	for k = 0; k < 10; k++ {
		gfp2K := &gfP2{}
		gfp2K.y.Set(newGFp(k))
		g, err := baseToTwist(gfp2K)
		if err != nil {
			t.Fatal(err)
		}
		g.ClearCofactor(g)
		if !g.IsOnTwist() {
			t.Fatal("Did not map to G2 from baseToTwist for t =", k)
		}
	}

	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("IsOnTwist changed twistGen")
	}
}

func TestAddTwist(t *testing.T) {
	c := &twistPoint{}
	a := &twistPoint{}
	b := &twistPoint{}

	a.Set(twistGen)
	b.SetInfinity()
	c.Add(a, b)
	if !c.IsEqual(twistGen) {
		t.Fatal("Failed to determine twistGen + Inf == twistGen")
	}

	a.SetInfinity()
	b.Set(twistGen)
	c.Add(a, b)
	if !c.IsEqual(twistGen) {
		t.Fatal("Failed to determine Inf + twistGen == twistGen")
	}

	a.SetInfinity()
	b.SetInfinity()
	c.Add(a, b)
	if !c.IsEqual(a) {
		t.Fatal("Failed to determine Inf + Inf == Inf")
	}

	a.Set(twistGen)
	b.Add(a, a)
	c.Double(a)
	if !b.IsEqual(c) {
		t.Fatal("Failed to determine twistGen + twistGen == 2*twistGen")
	}

	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("Twist Add changed twistGen")
	}
}
