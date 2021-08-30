package cloudflare

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"testing"
)

func TestG1Marshal(t *testing.T) {
	_, Ga, err := RandomG1(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ma := Ga.Marshal()

	Gb := new(G1)
	_, err = Gb.Unmarshal(ma)
	if err != nil {
		t.Fatal(err)
	}
	mb := Gb.Marshal()

	if !bytes.Equal(ma, mb) {
		t.Fatal("bytes are different")
	}
}

func TestG2Marshal(t *testing.T) {
	_, Ga, err := RandomG2(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ma := Ga.Marshal()

	Gb := new(G2)
	_, err = Gb.Unmarshal(ma)
	if err != nil {
		t.Fatal(err)
	}
	mb := Gb.Marshal()

	if !bytes.Equal(ma, mb) {
		t.Fatal("bytes are different")
	}
}

func TestBilinearity(t *testing.T) {
	for i := 0; i < 2; i++ {
		a, p1, _ := RandomG1(rand.Reader)
		b, p2, _ := RandomG2(rand.Reader)
		e1 := Pair(p1, p2)

		e2 := Pair(&G1{curveGen}, &G2{twistGen})
		e2.ScalarMult(e2, a)
		e2.ScalarMult(e2, b)

		if *e1.p != *e2.p {
			t.Fatalf("bad pairing result: %s", e1)
		}
	}
}

func TestTripartiteDiffieHellman(t *testing.T) {
	a, _ := rand.Int(rand.Reader, Order)
	b, _ := rand.Int(rand.Reader, Order)
	c, _ := rand.Int(rand.Reader, Order)

	pa, pb, pc := new(G1), new(G1), new(G1)
	qa, qb, qc := new(G2), new(G2), new(G2)

	_, err := pa.Unmarshal(new(G1).ScalarBaseMult(a).Marshal())
	if err != nil {
		t.Fatal(err)
	}
	_, err = qa.Unmarshal(new(G2).ScalarBaseMult(a).Marshal())
	if err != nil {
		t.Fatal(err)
	}
	_, err = pb.Unmarshal(new(G1).ScalarBaseMult(b).Marshal())
	if err != nil {
		t.Fatal(err)
	}
	_, err = qb.Unmarshal(new(G2).ScalarBaseMult(b).Marshal())
	if err != nil {
		t.Fatal(err)
	}
	_, err = pc.Unmarshal(new(G1).ScalarBaseMult(c).Marshal())
	if err != nil {
		t.Fatal(err)
	}
	_, err = qc.Unmarshal(new(G2).ScalarBaseMult(c).Marshal())
	if err != nil {
		t.Fatal(err)
	}

	k1 := Pair(pb, qc)
	k1.ScalarMult(k1, a)
	k1Bytes := k1.Marshal()

	k2 := Pair(pc, qa)
	k2.ScalarMult(k2, b)
	k2Bytes := k2.Marshal()

	k3 := Pair(pa, qb)
	k3.ScalarMult(k3, c)
	k3Bytes := k3.Marshal()

	if !bytes.Equal(k1Bytes, k2Bytes) || !bytes.Equal(k2Bytes, k3Bytes) {
		t.Fatalf("keys didn't agree")
	}
}

func BenchmarkG1(b *testing.B) {
	x, _ := rand.Int(rand.Reader, Order)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		new(G1).ScalarBaseMult(x)
	}
}

func BenchmarkG2(b *testing.B) {
	x, _ := rand.Int(rand.Reader, Order)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		new(G2).ScalarBaseMult(x)
	}
}
func BenchmarkPairing(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Pair(&G1{curveGen}, &G2{twistGen})
	}
}

// Additional Tests

func TestIsEqualG1(t *testing.T) {
	g1ten1 := new(G1).ScalarBaseMult(big.NewInt(10))
	g1ten2 := new(G1).ScalarBaseMult(big.NewInt(10))
	g1eleven := new(G1).ScalarBaseMult(big.NewInt(11))
	if !g1ten1.IsEqual(g1ten2) {
		t.Fatal("IsEqual (G1) failed to determine 10*genG1 == 10*genG1")
	}
	if g1ten1.IsEqual(g1eleven) {
		t.Fatal("IsEqual (G1) failed to determine 10*genG1 != 11*genG1")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("IsEqual (G1) changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("IsEqual (G1) changed twistGen")
	}
}

func TestIsEqualG2(t *testing.T) {
	g2ten1 := new(G2).ScalarBaseMult(big.NewInt(10))
	g2ten2 := new(G2).ScalarBaseMult(big.NewInt(10))
	g2eleven := new(G2).ScalarBaseMult(big.NewInt(11))
	if !g2ten1.IsEqual(g2ten2) {
		t.Fatal("IsEqual (G2) failed to determine 10*genG2 == 10*genG2")
	}
	if g2ten1.IsEqual(g2eleven) {
		t.Fatal("IsEqual (G2) failed to determine 10*genG2 != 11*genG2")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("IsEqual (G2) changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("IsEqual (G2) changed twistGen")
	}
}

func TestIsEqualGT(t *testing.T) {
	e := Pair(&G1{curveGen}, &G2{twistGen})
	g2ten1 := new(GT).ScalarMult(e, big.NewInt(10))
	g2ten2 := new(GT).ScalarMult(e, big.NewInt(10))
	g2eleven := new(GT).ScalarMult(e, big.NewInt(11))
	if !g2ten1.IsEqual(g2ten2) {
		t.Fatal("IsEqual (GT) failed to determine 10*genGT == 10*genGT")
	}
	if g2ten1.IsEqual(g2eleven) {
		t.Fatal("IsEqual (GT) failed to determine 10*genGT != 11*genGT")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("IsEqual (GT) changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("IsEqual (GT) changed twistGen")
	}
}

func TestG1Set(t *testing.T) {
	g1eleven := new(G1).ScalarBaseMult(big.NewInt(11))
	g1new := &G1{}
	g1new.Set(g1eleven)
	if !g1eleven.IsEqual(g1new) {
		t.Fatal("Set (G1) failed to set G1 elements equal")
	}
}

func TestG2Set(t *testing.T) {
	g2eleven := new(G2).ScalarBaseMult(big.NewInt(11))
	g2new := &G2{}
	g2new.Set(g2eleven)
	if !g2eleven.IsEqual(g2new) {
		t.Fatal("Set (G2) failed to set G2 elements equal")
	}
}

func TestGTSet(t *testing.T) {
	e := Pair(&G1{curveGen}, &G2{twistGen})
	gTeleven := &GT{}
	gTeleven.ScalarMult(e, big.NewInt(11))
	gTnew := &GT{}
	gTnew.Set(gTeleven)
	if !gTeleven.IsEqual(gTnew) {
		t.Fatal("Set (GT) failed to set GT elements equal")
	}
}

func TestG1Neg(t *testing.T) {
	g1eleven := new(G1).ScalarBaseMult(big.NewInt(11))
	g1neg := &G1{}
	g1neg.Neg(g1eleven)
	g1sum := &G1{}
	g1sum.Add(g1eleven, g1neg)
	if !g1sum.p.IsInfinity() {
		t.Fatal("Neg (G1) failed to compute negative of G1 element")
	}
}

func TestG2Neg(t *testing.T) {
	g2eleven := new(G2).ScalarBaseMult(big.NewInt(11))
	g2neg := &G2{}
	g2neg.Neg(g2eleven)
	g2sum := &G2{}
	g2sum.Add(g2eleven, g2neg)
	if !g2sum.p.IsInfinity() {
		t.Fatal("Neg (G2) failed to compute negative of G2 element")
	}
}

func TestGTNeg(t *testing.T) {
	e := Pair(&G1{curveGen}, &G2{twistGen})
	gTeleven := &GT{}
	gTeleven.ScalarMult(e, big.NewInt(11))
	gTneg := &GT{}
	gTneg.Neg(gTeleven)
	gTsum := &GT{}
	gTsum.Add(gTeleven, gTneg)
	if !gTsum.p.IsOne() {
		t.Fatal("Neg (GT) failed to compute negative of GT element")
	}
}

func TestMarshalG1(t *testing.T) {
	// More in-depth test than the original

	// Ensure non-initialized point is marshalled to zero byte slice
	g1 := &G1{}
	ret := g1.Marshal()
	for _, v := range ret {
		if v != 0 {
			t.Fatal("Incorrect G1 marshal result for nil")
		}
	}

	// Ensure non-initialized point is marshalled to zero byte slice
	// and unmarshalled to Infinity
	g2 := &G1{}
	g2.p = &curvePoint{}
	unmarsh2 := &G1{}
	ret2 := make([]byte, 2*numBytes)
	_, err2 := unmarsh2.Unmarshal(ret2)
	if err2 != nil {
		t.Fatal("Error in G1 Unmarshal with zero bytes")
	}
	g2.p.SetInfinity()
	if !unmarsh2.IsEqual(g2) {
		t.Fatal("Incorrect G1 unmarshal result for Infinity")
	}

	// Ensure correct marshalling and unmarshalling of G1 generator
	// as well using an initialized G1 element for unmarshalling
	g3 := &G1{}
	g3.ScalarBaseMult(big.NewInt(1))
	ret3 := g3.Marshal()
	unmarsh3 := &G1{}
	unmarsh3.ScalarBaseMult(big.NewInt(2))
	_, err3 := unmarsh3.Unmarshal(ret3)
	if err3 != nil {
		t.Fatal("Error in G1 Unmarshal of g1 (G1 generator)")
	}
	if !unmarsh3.IsEqual(g3) {
		t.Fatal("Incorrect G1 unmarshal result for g1 (G1 generator)")
	}

	// Ensure correct marshalling and unmarshalling of ``random'' G1 element
	g4 := &G1{}
	k := 123456789
	g4.ScalarBaseMult(big.NewInt(int64(k)))
	ret4 := g4.Marshal()
	unmarsh4 := &G1{}
	_, err4 := unmarsh4.Unmarshal(ret4)
	if err4 != nil {
		t.Fatal("Error in G1 Unmarshal of 123456789*g1")
	}
	if !unmarsh4.IsEqual(g4) {
		t.Fatal("Incorrect G1 unmarshal result for 123456789*g1")
	}

	// Ensure error is raised for insufficient byte slice
	ret5 := make([]byte, numBytes)
	unmarsh5 := &G1{}
	_, err5 := unmarsh5.Unmarshal(ret5)
	if err5 == nil {
		t.Fatal("Failed to raise G1 Unmarshal error for insufficient byte slice")
	}

	// Ensure error is raised for invalid G1 element
	g6 := &G1{}
	g6.p = &curvePoint{}
	g6.p.Set(curveGen)
	g6.p.x.Set(newGFp(2)) // mess up valid curvePoint
	ret6 := g6.Marshal()
	unmarsh6 := &G1{}
	_, err6 := unmarsh6.Unmarshal(ret6)
	if err6 == nil {
		t.Fatal("Failed to raise G1 Unmarshal error for non-G1 point")
	}

	// Invalid byte slice for GFp unmarshal
	breakBytes := make([]byte, numBytes)
	breakBytes[0] = byte(255) // >= 49 will work

	// Ensure error is raised for unmarshalling invalid x coordinate of curvePoint
	g7 := &G1{}
	g7.ScalarBaseMult(big.NewInt(1))
	ret7 := g7.Marshal()
	for k := 0; k < numBytes; k++ {
		ret7[k] = breakBytes[k]
	}
	unmarsh7 := &G1{}
	_, err7 := unmarsh7.Unmarshal(ret7)
	if err7 == nil {
		t.Fatal("Error in G1 Unmarshal of invalid x coordinate")
	}

	// Ensure error is raised for unmarshalling invalid y coordinate of curvePoint
	g8 := &G1{}
	g8.ScalarBaseMult(big.NewInt(1))
	ret8 := g8.Marshal()
	for k := 0; k < numBytes; k++ {
		ret8[numBytes+k] = breakBytes[k]
	}
	unmarsh8 := &G1{}
	_, err8 := unmarsh8.Unmarshal(ret8)
	if err8 == nil {
		t.Fatal("Error in G1 Unmarshal of invalid y coordinate")
	}
}

func TestMarshalG2(t *testing.T) {
	// More in-depth test than the original

	// Ensure non-initialized point is marshalled to zero byte slice
	g1 := &G2{}
	ret := g1.Marshal()
	for _, v := range ret {
		if v != 0 {
			t.Fatal("Incorrect G2 marshal result for nil")
		}
	}

	// Ensure non-initialized point is marshalled to zero byte slice
	// and unmarshalled to Infinity
	g2 := &G2{}
	g2.p = &twistPoint{}
	unmarsh2 := &G2{}
	ret2 := make([]byte, 4*numBytes)
	_, err2 := unmarsh2.Unmarshal(ret2)
	if err2 != nil {
		t.Fatal("Error in G2 Unmarshal with zero bytes")
	}
	g2.p.SetInfinity()
	if !unmarsh2.IsEqual(g2) {
		t.Fatal("Incorrect G2 unmarshal result for Infinity")
	}

	// Ensure correct marshalling and unmarshalling of G2 generator
	// as well using an initialized G2 element for unmarshalling
	g3 := &G2{}
	g3.ScalarBaseMult(big.NewInt(1))
	ret3 := g3.Marshal()
	unmarsh3 := &G2{}
	unmarsh3.ScalarBaseMult(big.NewInt(2))
	_, err3 := unmarsh3.Unmarshal(ret3)
	if err3 != nil {
		t.Fatal("Error in G2 Unmarshal of g2 (G2 generator)")
	}
	if !unmarsh3.IsEqual(g3) {
		t.Fatal("Incorrect G2 unmarshal result for g2 (G2 generator)")
	}

	// Ensure correct marshalling and unmarshalling of ``random'' G2 element
	g4 := &G2{}
	k := 123456789
	g4.ScalarBaseMult(big.NewInt(int64(k)))
	ret4 := g4.Marshal()
	unmarsh4 := &G2{}
	_, err4 := unmarsh4.Unmarshal(ret4)
	if err4 != nil {
		t.Fatal("Error in G2 Unmarshal of 123456789*g2")
	}
	if !unmarsh4.IsEqual(g4) {
		t.Fatal("Incorrect G2 unmarshal result for 123456789*g2")
	}

	// Ensure error is raised for insufficient byte slice
	ret5 := make([]byte, 3*numBytes)
	unmarsh5 := &G2{}
	_, err5 := unmarsh5.Unmarshal(ret5)
	if err5 == nil {
		t.Fatal("Failed to raise G2 Unmarshal error for insufficient byte slice")
	}

	// Ensure error is raised for invalid G2 element
	g6 := &G2{}
	g6.p = &twistPoint{}
	g6.p.Set(twistGen)
	g6.p.x.x.Set(newGFp(2)) // mess up valid twistPoint
	ret6 := g6.Marshal()
	unmarsh6 := &G2{}
	_, err6 := unmarsh6.Unmarshal(ret6)
	if err6 == nil {
		t.Fatal("Failed to raise G2 Unmarshal error for non-G2 point")
	}

	// Invalid byte slice for GFp unmarshal
	breakBytes := make([]byte, numBytes)
	breakBytes[0] = byte(255) // >= 49 will work

	// Ensure error is raised for unmarshalling invalid x.x coordinate of twistPoint
	g7 := &G2{}
	g7.ScalarBaseMult(big.NewInt(1))
	ret7 := g7.Marshal()
	for k := 0; k < numBytes; k++ {
		ret7[k] = breakBytes[k]
	}
	unmarsh7 := &G2{}
	_, err7 := unmarsh7.Unmarshal(ret7)
	if err7 == nil {
		t.Fatal("Error in G2 Unmarshal of invalid x.x coordinate")
	}

	// Ensure error is raised for unmarshalling invalid x.y coordinate of twistPoint
	g8 := &G2{}
	g8.ScalarBaseMult(big.NewInt(1))
	ret8 := g8.Marshal()
	for k := 0; k < numBytes; k++ {
		ret8[numBytes+k] = breakBytes[k]
	}
	unmarsh8 := &G2{}
	_, err8 := unmarsh8.Unmarshal(ret8)
	if err8 == nil {
		t.Fatal("Error in G2 Unmarshal of invalid x.y coordinate")
	}

	// Ensure error is raised for unmarshalling invalid y.x coordinate of twistPoint
	g9 := &G2{}
	g9.ScalarBaseMult(big.NewInt(1))
	ret9 := g9.Marshal()
	for k := 0; k < numBytes; k++ {
		ret9[2*numBytes+k] = breakBytes[k]
	}
	unmarsh9 := &G2{}
	_, err9 := unmarsh9.Unmarshal(ret9)
	if err9 == nil {
		t.Fatal("Error in G2 Unmarshal of invalid y.x coordinate")
	}

	// Ensure error is raised for unmarshalling invalid y.y coordinate of twistPoint
	g10 := &G2{}
	g10.ScalarBaseMult(big.NewInt(1))
	ret10 := g10.Marshal()
	for k := 0; k < numBytes; k++ {
		ret10[3*numBytes+k] = breakBytes[k]
	}
	unmarsh10 := &G2{}
	_, err10 := unmarsh10.Unmarshal(ret10)
	if err10 == nil {
		t.Fatal("Error in G2 Unmarshal of invalid y.y coordinate")
	}
}

func TestGTMarshal(t *testing.T) {
	e := Pair(&G1{curveGen}, &G2{twistGen})
	k := big.NewInt(123456789)
	Ga := &GT{}
	Ga.p = &gfP12{}
	Ga.ScalarMult(e, k)
	ma := Ga.Marshal()

	Gb := new(GT)
	_, err := Gb.Unmarshal(ma)
	if err != nil {
		t.Fatal(err)
	}
	mb := Gb.Marshal()

	if !bytes.Equal(ma, mb) {
		t.Fatal("bytes are different")
	}
}

func TestMarshalGT(t *testing.T) {
	e := Pair(&G1{curveGen}, &G2{twistGen})

	// More in-depth test than the original

	// Ensure non-initialized point is marshalled to one byte slice
	// (all zeros except for final byte, which is one)
	g1 := &GT{}
	ret := g1.Marshal()
	for k, v := range ret {
		if k != 383 {
			if v != 0 {
				t.Fatal("Incorrect GT marshal result for nil")
			}
		} else if k == 383 {
			if v != 1 {
				t.Fatal("Incorrect GT marshal result for nil")
			}
		}
	}

	// Ensure non-initialized point is marshalled to One byte slice
	// and unmarshalled to One
	g2 := &GT{}
	g2.p = &gfP12{}
	unmarsh2 := &GT{}
	ret2 := make([]byte, 12*numBytes)
	ret2[383] = 1
	_, err2 := unmarsh2.Unmarshal(ret2)
	if err2 != nil {
		t.Fatal("Error in GT Unmarshal with One bytes")
	}
	g2.p.SetOne()
	if !unmarsh2.IsEqual(g2) {
		t.Fatal("Incorrect GT unmarshal result for One")
	}

	// Ensure correct marshalling and unmarshalling of GT generator
	// as well using an initialized GT element for unmarshalling
	g3 := &GT{}
	g3.p = &gfP12{}
	g3.p.SetOne()
	ret3 := g3.Marshal()
	unmarsh3 := &GT{}
	unmarsh3.ScalarMult(e, big.NewInt(2))
	_, err3 := unmarsh3.Unmarshal(ret3)
	if err3 != nil {
		t.Fatal("Error in GT Unmarshal of gT (GT generator)")
	}
	if !unmarsh3.IsEqual(g3) {
		t.Fatal("Incorrect GT unmarshal result for gT (GT generator)")
	}

	// Ensure correct marshalling and unmarshalling of ``random'' G2 element
	g4 := &GT{}
	k := 123456789
	g4.ScalarMult(e, big.NewInt(int64(k)))
	ret4 := g4.Marshal()
	unmarsh4 := &GT{}
	_, err4 := unmarsh4.Unmarshal(ret4)
	if err4 != nil {
		t.Fatal("Error in GT Unmarshal of 123456789*gT")
	}
	if !unmarsh4.IsEqual(g4) {
		t.Fatal("Incorrect GT unmarshal result for 123456789*gT")
	}

	// Ensure error is raised for insufficient byte slice
	ret5 := make([]byte, 11*numBytes)
	unmarsh5 := &GT{}
	_, err5 := unmarsh5.Unmarshal(ret5)
	if err5 == nil {
		t.Fatal("Failed to raise GT Unmarshal error for insufficient byte slice")
	}

	// Ensure error is raised for invalid G2 element
	g6 := &GT{}
	g6.p = &gfP12{}
	g6.p.SetOne()
	g6.p.x.x.x.Set(newGFp(2)) // mess up valid twistPoint
	ret6 := g6.Marshal()
	unmarsh6 := &GT{}
	_, err6 := unmarsh6.Unmarshal(ret6)
	if err6 == nil {
		t.Fatal("Failed to raise GT Unmarshal error for non-GT point")
	}

	// Invalid byte slice for GFp unmarshal
	breakBytes := make([]byte, numBytes)
	breakBytes[0] = byte(255) // >= 49 will work

	// Ensure error is raised for unmarshalling invalid x.x.x coordinate of twistPoint
	g7 := &GT{}
	g7.p = &gfP12{}
	g7.p.SetOne()
	ret7 := g7.Marshal()
	for k := 0; k < numBytes; k++ {
		ret7[k] = breakBytes[k]
	}
	unmarsh7 := &GT{}
	_, err7 := unmarsh7.Unmarshal(ret7)
	if err7 == nil {
		t.Fatal("Error in GT Unmarshal of invalid x.x.x coordinate")
	}

	// Ensure error is raised for unmarshalling invalid x.x.y coordinate of twistPoint
	g8 := &GT{}
	g8.p = &gfP12{}
	g8.p.SetOne()
	ret8 := g8.Marshal()
	for k := 0; k < numBytes; k++ {
		ret8[numBytes+k] = breakBytes[k]
	}
	unmarsh8 := &GT{}
	_, err8 := unmarsh8.Unmarshal(ret8)
	if err8 == nil {
		t.Fatal("Error in GT Unmarshal of invalid x.x.y coordinate")
	}

	// Ensure error is raised for unmarshalling invalid x.y.x coordinate of twistPoint
	g9 := &GT{}
	g9.p = &gfP12{}
	g9.p.SetOne()
	ret9 := g9.Marshal()
	for k := 0; k < numBytes; k++ {
		ret9[2*numBytes+k] = breakBytes[k]
	}
	unmarsh9 := &GT{}
	_, err9 := unmarsh9.Unmarshal(ret9)
	if err9 == nil {
		t.Fatal("Error in GT Unmarshal of invalid x.y.x coordinate")
	}

	// Ensure error is raised for unmarshalling invalid x.y.y coordinate of twistPoint
	g10 := &GT{}
	g10.p = &gfP12{}
	g10.p.SetOne()
	ret10 := g10.Marshal()
	for k := 0; k < numBytes; k++ {
		ret10[3*numBytes+k] = breakBytes[k]
	}
	unmarsh10 := &GT{}
	_, err10 := unmarsh10.Unmarshal(ret10)
	if err10 == nil {
		t.Fatal("Error in GT Unmarshal of invalid x.y.y coordinate")
	}

	// Ensure error is raised for unmarshalling invalid x.z.x coordinate of twistPoint
	g11 := &GT{}
	g11.p = &gfP12{}
	g11.p.SetOne()
	ret11 := g11.Marshal()
	for k := 0; k < numBytes; k++ {
		ret11[4*numBytes+k] = breakBytes[k]
	}
	unmarsh11 := &GT{}
	_, err11 := unmarsh11.Unmarshal(ret11)
	if err11 == nil {
		t.Fatal("Error in GT Unmarshal of invalid x.z.x coordinate")
	}

	// Ensure error is raised for unmarshalling invalid x.z.y coordinate of twistPoint
	g12 := &GT{}
	g12.p = &gfP12{}
	g12.p.SetOne()
	ret12 := g12.Marshal()
	for k := 0; k < numBytes; k++ {
		ret12[5*numBytes+k] = breakBytes[k]
	}
	unmarsh12 := &GT{}
	_, err12 := unmarsh12.Unmarshal(ret12)
	if err12 == nil {
		t.Fatal("Error in GT Unmarshal of invalid x.z.y coordinate")
	}

	// Ensure error is raised for unmarshalling invalid y.x.x coordinate of twistPoint
	g13 := &GT{}
	g13.p = &gfP12{}
	g13.p.SetOne()
	ret13 := g13.Marshal()
	for k := 0; k < numBytes; k++ {
		ret13[6*numBytes+k] = breakBytes[k]
	}
	unmarsh13 := &GT{}
	_, err13 := unmarsh13.Unmarshal(ret13)
	if err13 == nil {
		t.Fatal("Error in GT Unmarshal of invalid y.x.x coordinate")
	}

	// Ensure error is raised for unmarshalling invalid y.x.y coordinate of twistPoint
	g14 := &GT{}
	g14.p = &gfP12{}
	g14.p.SetOne()
	ret14 := g14.Marshal()
	for k := 0; k < numBytes; k++ {
		ret14[7*numBytes+k] = breakBytes[k]
	}
	unmarsh14 := &GT{}
	_, err14 := unmarsh14.Unmarshal(ret14)
	if err14 == nil {
		t.Fatal("Error in GT Unmarshal of invalid y.x.y coordinate")
	}

	// Ensure error is raised for unmarshalling invalid y.y.x coordinate of twistPoint
	g15 := &GT{}
	g15.p = &gfP12{}
	g15.p.SetOne()
	ret15 := g15.Marshal()
	for k := 0; k < numBytes; k++ {
		ret15[8*numBytes+k] = breakBytes[k]
	}
	unmarsh15 := &GT{}
	_, err15 := unmarsh15.Unmarshal(ret15)
	if err15 == nil {
		t.Fatal("Error in GT Unmarshal of invalid y.y.x coordinate")
	}

	// Ensure error is raised for unmarshalling invalid y.y.y coordinate of twistPoint
	g16 := &GT{}
	g16.p = &gfP12{}
	g16.p.SetOne()
	ret16 := g16.Marshal()
	for k := 0; k < numBytes; k++ {
		ret16[9*numBytes+k] = breakBytes[k]
	}
	unmarsh16 := &GT{}
	_, err16 := unmarsh16.Unmarshal(ret16)
	if err16 == nil {
		t.Fatal("Error in GT Unmarshal of invalid y.y.y coordinate")
	}

	// Ensure error is raised for unmarshalling invalid y.z.x coordinate of twistPoint
	g17 := &GT{}
	g17.p = &gfP12{}
	g17.p.SetOne()
	ret17 := g17.Marshal()
	for k := 0; k < numBytes; k++ {
		ret17[10*numBytes+k] = breakBytes[k]
	}
	unmarsh17 := &GT{}
	_, err17 := unmarsh17.Unmarshal(ret17)
	if err17 == nil {
		t.Fatal("Error in GT Unmarshal of invalid y.z.x coordinate")
	}

	// Ensure error is raised for unmarshalling invalid y.z.y coordinate of twistPoint
	g18 := &GT{}
	g18.p = &gfP12{}
	g18.p.SetOne()
	ret18 := g18.Marshal()
	for k := 0; k < numBytes; k++ {
		ret18[11*numBytes+k] = breakBytes[k]
	}
	unmarsh18 := &GT{}
	_, err18 := unmarsh18.Unmarshal(ret18)
	if err18 == nil {
		t.Fatal("Error in GT Unmarshal of invalid y.z.y coordinate")
	}
}

func TestMillerFinalize(t *testing.T) {
	e1 := Pair(&G1{curveGen}, &G2{twistGen})
	e2 := Miller(&G1{curveGen}, &G2{twistGen})
	e2.Finalize()
	if !e2.IsEqual(e1) {
		t.Fatal("Error in Miller followed by Finalize")
	}
}

func TestPairingCheck(t *testing.T) {
	// Some examples copied from main_test
	a1 := new(G1).ScalarBaseMult(bigFromBase10("1"))
	a2 := new(G1).ScalarBaseMult(bigFromBase10("2"))
	a7 := new(G1).ScalarBaseMult(bigFromBase10("7"))
	an7 := new(G1).ScalarBaseMult(bigFromBase10("21888242871839275222246405745257275088548364400416034343698204186575808495610"))
	an2 := new(G1).ScalarBaseMult(bigFromBase10("21888242871839275222246405745257275088548364400416034343698204186575808495615"))
	an1 := new(G1).ScalarBaseMult(bigFromBase10("21888242871839275222246405745257275088548364400416034343698204186575808495616"))

	b0 := new(G2).ScalarBaseMult(bigFromBase10("0"))
	b1 := new(G2).ScalarBaseMult(bigFromBase10("1"))
	bn1 := new(G2).ScalarBaseMult(bigFromBase10("21888242871839275222246405745257275088548364400416034343698204186575808495616"))

	res1 := PairingCheck([]*G1{a1, an1, a1}, []*G2{b1, b1, b0})
	if !res1 {
		t.Fatal("Failed to determine valid pairing (1)")
	}

	res2 := PairingCheck([]*G1{a1, a1}, []*G2{b1, bn1})
	if !res2 {
		t.Fatal("Failed to determine valid pairing (2)")
	}

	a1neg := new(G1).Neg(a1)
	if !an1.IsEqual(a1neg) {
		t.Fatal("Failed a1 negation")
	}

	a2neg := new(G1).Neg(a2)
	if !an2.IsEqual(a2neg) {
		t.Fatal("Failed a2 negation")
	}

	a7neg := new(G1).Neg(a7)
	if !an7.IsEqual(a7neg) {
		t.Fatal("Failed a7 negation")
	}

	b1neg := new(G2).Neg(b1)
	if !bn1.IsEqual(b1neg) {
		t.Fatal("Failed b1 negation")
	}

	res3 := PairingCheck([]*G1{a1, a2, a2, a2, a7}, []*G2{b1, b1, b1, b1, bn1})
	if !res3 {
		t.Fatal("Failed to determine valid pairing (3)")
	}

	gT := Pair(a1, b1)
	if !gT.p.IsOnCurve() {
		t.Fatal("Some thing is wrong with gT (GT generator)")
	}

	cExp := big.NewInt(123456789101112)
	dExp := big.NewInt(3141592653589793) // 16 digits of Pi

	c1 := new(G1).ScalarBaseMult(cExp)
	d1 := new(G2).ScalarBaseMult(dExp)

	c1neg := new(G1).Neg(c1)
	sumC1 := new(G1).Add(c1, c1neg)
	if !sumC1.p.IsInfinity() {
		t.Fatal("c1 + (-c1) != Inf")
	}
	c1negExp := new(big.Int).Sub(Order, cExp)
	c1negCheck := new(G1).ScalarBaseMult(c1negExp)
	if !c1negCheck.IsEqual(c1neg) {
		t.Fatal("Error in c1neg exponent")
	}

	d1neg := new(G2).Neg(d1)
	sumD1 := new(G2).Add(d1, d1neg)
	if !sumD1.p.IsInfinity() {
		t.Fatal("d1 + (-d1) != Inf")
	}
	d1negExp := new(big.Int).Sub(Order, dExp)
	d1negCheck := new(G2).ScalarBaseMult(d1negExp)
	if !d1negCheck.IsEqual(d1neg) {
		t.Fatal("Error in d1neg exponent")
	}

	e1 := Pair(c1, d1)

	e1exp := new(big.Int).Mul(cExp, dExp)
	e1exp.Mod(e1exp, Order)
	e1check := new(GT).ScalarMult(gT, e1exp)
	if !e1check.IsEqual(e1) {
		t.Fatal("We have problem in e1 pairing")
	}

	e2 := Pair(c1neg, d1neg)
	e2exp := new(big.Int).Mul(c1negExp, d1negExp)
	e2exp.Mod(e2exp, Order)
	e2check := new(GT).ScalarMult(gT, e2exp)
	if !e2check.IsEqual(e2) {
		t.Fatal("We have problem in e2 pairing")
	}

	if !e1.IsEqual(e2) {
		t.Fatal("Problem in Pair")
	}

	e3 := Pair(c1, d1neg)
	e3exp := new(big.Int).Mul(cExp, d1negExp)
	e3exp.Mod(e3exp, Order)
	e3check := new(GT).ScalarMult(gT, e3exp)
	if !e3check.IsEqual(e3) {
		t.Fatal("We have problem in e3 pairing")
	}

	e3neg := new(GT).Neg(e3)
	if !e3neg.IsEqual(e1) {
		t.Fatal("Houston, we have a problem with e3neg")
	}

	e4 := Pair(c1neg, d1)
	e4exp := new(big.Int).Mul(c1negExp, dExp)
	e4exp.Mod(e4exp, Order)
	e4check := new(GT).ScalarMult(gT, e4exp)
	if !e4check.IsEqual(e4) {
		t.Fatal("We have problem in e4 pairing")
	}

	e4neg := new(GT).Neg(e4)
	if !e4neg.IsEqual(e1) {
		t.Fatal("Houston, we have a problem with e4neg")
	}

	if !PairingCheck([]*G1{c1, c1}, []*G2{d1, d1neg}) {
		t.Fatal("Error in PairingCheck")
	}

	alpha := big.NewInt(123456789101112)
	beta := big.NewInt(3141592653589793) // 16 digits of Pi
	negAlpha := new(big.Int).Sub(Order, alpha)
	negBeta := new(big.Int).Sub(Order, beta)

	g1 := new(G1).ScalarBaseMult(alpha)
	g1neg := new(G1).ScalarBaseMult(negAlpha)
	g2 := new(G1).ScalarBaseMult(beta)
	g2neg := new(G1).ScalarBaseMult(negBeta)
	h1 := new(G2).ScalarBaseMult(beta)
	h1neg := new(G2).ScalarBaseMult(negBeta)
	h2 := new(G2).ScalarBaseMult(alpha)
	h2neg := new(G2).ScalarBaseMult(negAlpha)

	if !PairingCheck([]*G1{g1, g2}, []*G2{h1neg, h2}) {
		t.Fatal("Error in PairingCheck (h1neg)")
	}

	if !PairingCheck([]*G1{g1, g2}, []*G2{h1, h2neg}) {
		t.Fatal("Error in PairingCheck (h2neg)")
	}

	if !PairingCheck([]*G1{g1neg, g2}, []*G2{h1, h2}) {
		t.Fatal("Error in PairingCheck (g1neg)")
	}

	if !PairingCheck([]*G1{g1, g2neg}, []*G2{h1, h2}) {
		t.Fatal("Error in PairingCheck (g2neg)")
	}
}

func TestG1Double(t *testing.T) {
	k := 1
	s := big.NewInt(int64(k))
	p := new(G1).ScalarBaseMult(s)
	if !p.p.IsOnCurve() {
		t.Fatal("p isn't on curve")
	}
	m := p.Add(p, p).Marshal()
	if _, err := p.Unmarshal(m); err != nil {
		t.Fatalf("p.Add(p, p) not in G1: %v", err)
	}
}

func TestG2Double(t *testing.T) {
	k := 1
	s := big.NewInt(int64(k))
	p := new(G2).ScalarBaseMult(s)
	if !p.p.IsOnCurve() {
		t.Fatal("p isn't on curve")
	}
	m := p.Add(p, p).Marshal()
	if _, err := p.Unmarshal(m); err != nil {
		t.Fatalf("p.Add(p, p) not in G2: %v", err)
	}
}

func TestGTDouble(t *testing.T) {
	k := 1
	s := big.NewInt(int64(k))
	g1 := new(G1).ScalarBaseMult(s)
	g2 := new(G2).ScalarBaseMult(s)
	gT := Pair(g1, g2)
	if gT.p.IsOne() {
		t.Fatal("p should not be 1")
	}
	if !gT.p.IsOnCurve() {
		t.Fatal("p isn't on curve")
	}
	m := gT.Add(gT, gT).Marshal()
	if _, err := gT.Unmarshal(m); err != nil {
		t.Fatalf("p.Add(p, p) not in GT: %v", err)
	}

	gT = Pair(g1, g2)
	if !gT.p.IsOnCurve() {
		t.Fatal("p isn't on curve")
	}
	gT.p.Square(gT.p)
	if !gT.p.IsOnCurve() {
		t.Fatal("p isn't on curve")
	}
	n := gT.Marshal()
	if _, err := gT.Unmarshal(n); err != nil {
		t.Fatalf("p.Add(p, p) not in GT (2): %v", err)
	}

	gT = Pair(g1, g2)
	if !gT.p.IsOnCurve() {
		t.Fatal("p isn't on curve")
	}
	gT.p.Mul(gT.p, gT.p)
	if !gT.p.IsOnCurve() {
		t.Fatal("p isn't on curve")
	}
	b := gT.Marshal()
	if _, err := gT.Unmarshal(b); err != nil {
		t.Fatalf("p.Add(p, p) not in GT (3): %v", err)
	}
}
