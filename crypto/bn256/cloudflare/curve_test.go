package cloudflare

import (
	"testing"
)

func TestReturnCurveGen(t *testing.T) {
	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("returnCurveGen failed")
	}
}

func TestIsEqualCurve(t *testing.T) {
	p1 := &curvePoint{}
	p1.Set(curveGen)
	p2 := &curvePoint{}
	p2.Set(curveGen)
	if !p1.IsEqual(p2) {
		t.Fatal("Did not determine that curveGen == curveGen")
	}

	p2.SetInfinity()
	if p1.IsEqual(p2) {
		t.Fatal("Did not determine that curveGen != Infinity")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("IsEqualCurve changed curveGen")
	}
}

func TestIsOnCurve(t *testing.T) {
	p := &curvePoint{}

	p.Set(curveGen)
	if !p.IsOnCurve() {
		t.Fatal("Did not determine curveGen is on curve")
	}

	p.Neg(p)
	if !p.IsOnCurve() {
		t.Fatal("Did not determine -curveGen is on curve")
	}

	p.SetInfinity()
	if !p.IsOnCurve() {
		t.Fatal("Did not determine Infinite is on curve")
	}

	s := "MadHive Rocks!"
	msg := []byte(s)
	g, err := HashToG1(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !g.p.IsOnCurve() {
		t.Fatal("HashToG1 of MH failed")
	}

	s = "Cryptography is great"
	msg = []byte(s)
	g, err = HashToG1(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !g.p.IsOnCurve() {
		t.Fatal("HashToG1 of CIG failed")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("IsOnCurve changed curveGen")
	}
}

func TestAdd(t *testing.T) {
	c := &curvePoint{}
	a := &curvePoint{}
	b := &curvePoint{}

	a.Set(curveGen)
	b.SetInfinity()
	c.Add(a, b)
	if !c.IsEqual(curveGen) {
		t.Fatal("Failed to determine curveGen + Inf == curveGen")
	}

	a.SetInfinity()
	b.Set(curveGen)
	c.Add(a, b)
	if !c.IsEqual(curveGen) {
		t.Fatal("Failed to determine Inf + curveGen == curveGen")
	}

	a.SetInfinity()
	b.SetInfinity()
	c.Add(a, b)
	if !c.IsEqual(a) {
		t.Fatal("Failed to determine Inf + Inf == Inf")
	}

	a.Set(curveGen)
	b.Add(a, a)
	c.Double(a)
	if !b.IsEqual(c) {
		t.Fatal("Failed to determine curveGen + curveGen == 2*curveGen")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("curvePoint Add changed curveGen")
	}
}

func TestCurveDouble(t *testing.T) {
	c := &curvePoint{}
	c.Add(curveGen, curveGen)
	if !c.IsOnCurve() {
		t.Fatal("curvePoint should be on curve (1)")
	}

	a := &curvePoint{}
	a.Set(curveGen)
	a.Add(a, a)
	if !a.IsOnCurve() {
		t.Fatal("curvePoint should be on curve (2)")
	}
	if !c.IsEqual(a) {
		t.Fatal("curvePoints should be equal (2)")
	}

	b := &curvePoint{}
	b.Set(curveGen)
	b.Double(b)
	if !b.IsOnCurve() {
		t.Fatal("curvePoint should be on curve (3)")
	}
	if !c.IsEqual(b) {
		t.Fatal("curvePoints should be equal (3)")
	}
}
