package cloudflare

import (
	"math/big"
	"testing"
)

func TestG1SerializeUncompressed(t *testing.T) {
	g1 := &G1{}
	ret1True := g1.Marshal()
	ret1 := g1.SerializeUncompressed()
	// Ensure format flag is correct
	if ret1[0] != g1UncompFlag {
		t.Fatal("Invalid byte format (1)")
	}
	// Ensure length is correct
	if len(ret1True)+1 != len(ret1) {
		t.Fatal("Incorrect byte length (1)")
	}
	// Ensure byte slice values are correct
	for j := 0; j < len(ret1True); j++ {
		if ret1True[j] != ret1[j+1] {
			t.Fatal("Invalid uncompressed serialization (1)")
		}
	}
	// Check deserialization returns same point
	g1New := &G1{}
	g1New.Deserialize(ret1)
	if !g1.IsEqual(g1New) {
		t.Fatal("Not Equal (1)")
	}

	g2 := &G1{}
	g2.ScalarBaseMult(big.NewInt(1))
	ret2True := g2.Marshal()
	ret2 := g2.SerializeUncompressed()
	// Ensure format flag is correct
	if ret2[0] != g1UncompFlag {
		t.Fatal("Invalid byte format (2)")
	}
	// Ensure length is correct
	if len(ret2True)+1 != len(ret2) {
		t.Fatal("Incorrect byte length (2)")
	}
	// Ensure byte slice values are correct
	for j := 0; j < len(ret2True); j++ {
		if ret2True[j] != ret2[j+1] {
			t.Fatal("Invalid uncompressed serialization (2)")
		}
	}
	// Check deserialization returns same point
	g2New := &G1{}
	g2New.Deserialize(ret2)
	if !g2.IsEqual(g2New) {
		t.Fatal("Not Equal (2)")
	}

	g3 := &G1{}
	g3.ScalarBaseMult(big.NewInt(123456789))
	ret3True := g3.Marshal()
	ret3 := g3.SerializeUncompressed()
	// Ensure format flag is correct
	if ret3[0] != g1UncompFlag {
		t.Fatal("Invalid byte format (3)")
	}
	// Ensure length is correct
	if len(ret3True)+1 != len(ret3) {
		t.Fatal("Incorrect byte length (3)")
	}
	// Ensure byte slice values are correct
	for j := 0; j < len(ret3True); j++ {
		if ret3True[j] != ret3[j+1] {
			t.Fatal("Invalid uncompressed serialization (3)")
		}
	}
	// Check deserialization returns same point
	g3New := &G1{}
	g3New.Deserialize(ret3)
	if !g3.IsEqual(g3New) {
		t.Fatal("Not Equal (3)")
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

func TestG1SerializeCompressed(t *testing.T) {
	g1Identity := &G1{}
	g1Identity.p = &curvePoint{}
	g1Identity.p.SetInfinity()

	g1 := &G1{}
	ret1 := g1.SerializeCompressed()
	// Ensure format flag is correct
	if ret1[0] != g1CompFlag {
		t.Fatal("Invalid byte format (1)")
	}
	// Ensure length is correct
	if numBytes+1 != len(ret1) {
		t.Fatal("Incorrect byte length (1)")
	}
	// Ensure byte slice values for x coordinate are correct
	for j := 1; j < len(ret1); j++ {
		if ret1[j] != 0 {
			t.Fatal("Invalid compressed serialization (1)")
		}
	}
	// Check deserialization returns same point
	g1New := &G1{}
	g1New.Deserialize(ret1)
	if !g1.IsEqual(g1New) {
		t.Fatal("Not Equal (1)")
	}

	g2 := &G1{}
	g2.ScalarBaseMult(big.NewInt(1))
	ret2True := g2.Marshal()
	ret2 := g2.SerializeCompressed()
	// Ensure format flag is correct; y is even
	if ret2[0] != g1CompFlag {
		t.Fatal("Invalid byte format (2)")
	}
	// Ensure length is correct
	if numBytes+1 != len(ret2) {
		t.Fatal("Incorrect byte length (2)")
	}
	// Ensure byte slice values for x coordinate are correct
	for j := 0; j < numBytes; j++ {
		if ret2True[j] != ret2[j+1] {
			t.Fatal("Invalid compressed serialization (2)")
		}
	}
	// Check deserialization returns same point
	g2New := &G1{}
	g2New.Deserialize(ret2)
	if !g2.IsEqual(g2New) {
		t.Fatal("Not Equal (2)")
	}

	// Want to negate return value and ensure deserializes properly;
	// thus, adding points together should equal identity element
	ret2Neg := g2.SerializeCompressed()
	v2 := byte(uint8(ret2Neg[0]))
	ret2Neg[0] = ((v2 & g1YOddFlag) ^ g1YOddFlag) | g1CompFlag
	g2Neg := &G1{}
	err := g2Neg.Deserialize(ret2Neg)
	if err != nil {
		t.Fatal(err)
	}
	g2Add := &G1{}
	g2Add.Add(g2, g2Neg)
	if !g2Add.IsEqual(g1Identity) {
		t.Fatal("Should equal Identity (2)")
	}

	g3 := &G1{}
	g3.ScalarBaseMult(big.NewInt(3))
	ret3True := g3.Marshal()
	ret3 := g3.SerializeCompressed()
	// Ensure format flag is correct
	if ret3[0] != g1CompFlag|g1YOddFlag {
		t.Fatal("Invalid byte format (3)")
	}
	// Ensure length is correct
	if numBytes+1 != len(ret3) {
		t.Fatal("Incorrect byte length (3)")
	}
	// Ensure byte slice values for x coordinate are correct
	for j := 0; j < numBytes; j++ {
		if ret3True[j] != ret3[j+1] {
			t.Fatal("Invalid uncompressed serialization (3)")
		}
	}
	// Check deserialization returns same point
	g3New := &G1{}
	g3New.Deserialize(ret3)
	if !g3.IsEqual(g3New) {
		t.Fatal("Not Equal (3)")
	}

	// Want to negate return value and ensure deserializes properly;
	// thus, adding points together should equal identity element
	ret3Neg := g3.SerializeCompressed()
	v3 := byte(uint8(ret3Neg[0]))
	ret3Neg[0] = ((v3 & g1YOddFlag) ^ g1YOddFlag) | g1CompFlag
	g3Neg := &G1{}
	err = g3Neg.Deserialize(ret3Neg)
	if err != nil {
		t.Fatal(err)
	}
	g3Add := &G1{}
	g3Add.Add(g3, g3Neg)
	if !g3Add.IsEqual(g1Identity) {
		t.Fatal("Should equal Identity (3)")
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

func TestG1Deserialize(t *testing.T) {
	g1 := &G1{}
	m1 := make([]byte, 0)
	err1 := g1.Deserialize(m1)
	// Should raise error due to slice length 0
	if err1 == nil {
		t.Fatal("Should have raised error (1)")
	}

	g2 := &G1{}
	m2 := []byte{255}
	err2 := g2.Deserialize(m2)
	// Should raise error due to invalid format byte
	if err2 == nil {
		t.Fatal("Should have raised error (2)")
	}

	g3 := &G1{}
	m3 := []byte{g1UncompFlag}
	err3 := g3.Deserialize(m3)
	// Should raise error due to incorrect slice length
	if err3 == nil {
		t.Fatal("Should have raised error (3)")
	}

	g4 := &G1{}
	m4 := []byte{g1CompFlag}
	err4 := g4.Deserialize(m4)
	// Should raise error due to incorrect slice length
	if err4 == nil {
		t.Fatal("Should have raised error (4)")
	}

	g5 := &G1{}
	m5 := []byte{g1CompFlag | g1YOddFlag}
	err5 := g5.Deserialize(m5)
	// Should raise error due to incorrect slice length
	if err5 == nil {
		t.Fatal("Should have raised error (5)")
	}

	// Invalid data for uncompressed; invalid point
	g6 := &G1{}
	m6 := make([]byte, 1+2*numBytes, 1+2*numBytes)
	m6[0] = g1UncompFlag
	m6[1] = 1
	err6 := g6.Deserialize(m6)
	if err6 == nil {
		t.Fatal("Should have raised error (6)")
	}

	// Valid data for compressed identity element
	g7 := &G1{}
	m7 := make([]byte, 1+numBytes, 1+numBytes)
	m7[0] = g1CompFlag
	err7 := g7.Deserialize(m7)
	if err7 != nil {
		t.Fatal(err7)
	}

	// Invalid data for compressed identity element: yOddFlag set
	g8 := &G1{}
	m8 := make([]byte, 1+numBytes, 1+numBytes)
	m8[0] = g1CompFlag | g1YOddFlag
	err8 := g8.Deserialize(m8)
	if err8 == nil {
		t.Fatal("Should have raised error (8)")
	}

	// Invalid data for compressed data: x coordinate too large
	g9 := &G1{}
	m9 := make([]byte, 1+numBytes, 1+numBytes)
	m9[0] = g1CompFlag
	m9[1] = 255
	err9 := g9.Deserialize(m9)
	if err9 == nil {
		t.Fatal("Should have raised error (9)")
	}

	// Invalid data for compressed data: no point has x == 4
	g10 := &G1{}
	m10 := make([]byte, 1+numBytes, 1+numBytes)
	m10[0] = g1CompFlag
	m10[numBytes] = 4
	err10 := g10.Deserialize(m10)
	if err10 == nil {
		t.Fatal("Should have raised error (10)")
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

func TestComputeG1YValue(t *testing.T) {
	x := newGFp(1)
	yIsOdd := false
	y := computeG1YValue(x, yIsOdd)
	yTrue := newGFp(2)
	if !y.IsEqual(yTrue) {
		t.Fatal("Should be equal (1)")
	}

	yIsOdd = true
	y = computeG1YValue(x, yIsOdd)
	gfpNeg(yTrue, newGFp(2))
	if !y.IsEqual(yTrue) {
		t.Fatal("Should be equal (2)")
	}
}

func TestG2SerializeUncompressed(t *testing.T) {
	g1 := &G2{}
	ret1True := g1.Marshal()
	ret1 := g1.SerializeUncompressed()
	// Ensure format flag is correct
	if ret1[0] != g2UncompFlag {
		t.Fatal("Invalid byte format (1)")
	}
	// Ensure length is correct
	if len(ret1True)+1 != len(ret1) {
		t.Fatal("Incorrect byte length (1)")
	}
	// Ensure byte slice values are correct
	for j := 0; j < len(ret1True); j++ {
		if ret1True[j] != ret1[j+1] {
			t.Fatal("Invalid uncompressed serialization (1)")
		}
	}
	// Check deserialization returns same point
	g1New := &G2{}
	g1New.Deserialize(ret1)
	if !g1.IsEqual(g1New) {
		t.Fatal("Not Equal (1)")
	}

	g2 := &G2{}
	g2.ScalarBaseMult(big.NewInt(1))
	ret2True := g2.Marshal()
	ret2 := g2.SerializeUncompressed()
	// Ensure format flag is correct
	if ret2[0] != g2UncompFlag {
		t.Fatal("Invalid byte format (2)")
	}
	// Ensure length is correct
	if len(ret2True)+1 != len(ret2) {
		t.Fatal("Incorrect byte length (2)")
	}
	// Ensure byte slice values are correct
	for j := 0; j < len(ret2True); j++ {
		if ret2True[j] != ret2[j+1] {
			t.Fatal("Invalid uncompressed serialization (2)")
		}
	}
	// Check deserialization returns same point
	g2New := &G2{}
	g2New.Deserialize(ret2)
	if !g2.IsEqual(g2New) {
		t.Fatal("Not Equal (2)")
	}

	g3 := &G2{}
	g3.ScalarBaseMult(big.NewInt(123456789))
	ret3True := g3.Marshal()
	ret3 := g3.SerializeUncompressed()
	// Ensure format flag is correct
	if ret3[0] != g2UncompFlag {
		t.Fatal("Invalid byte format (3)")
	}
	// Ensure length is correct
	if len(ret3True)+1 != len(ret3) {
		t.Fatal("Incorrect byte length (3)")
	}
	// Ensure byte slice values are correct
	for j := 0; j < len(ret3True); j++ {
		if ret3True[j] != ret3[j+1] {
			t.Fatal("Invalid uncompressed serialization (3)")
		}
	}
	// Check deserialization returns same point
	g3New := &G2{}
	g3New.Deserialize(ret3)
	if !g3.IsEqual(g3New) {
		t.Fatal("Not Equal (3)")
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

func TestG2SerializeCompressed(t *testing.T) {
	g2Identity := &G2{}
	g2Identity.p = &twistPoint{}
	g2Identity.p.SetInfinity()

	g1 := &G2{}
	ret1 := g1.SerializeCompressed()
	// Ensure format flag is correct
	if ret1[0] != g2CompFlag {
		t.Fatal("Invalid byte format (1)")
	}
	// Ensure length is correct
	if 2*numBytes+1 != len(ret1) {
		t.Fatal("Incorrect byte length (1)")
	}
	// Ensure byte slice values for x coordinate are correct
	for j := 1; j < len(ret1); j++ {
		if ret1[j] != 0 {
			t.Fatal("Invalid compressed serialization (1)")
		}
	}
	// Check deserialization returns same point
	g1New := &G2{}
	g1New.Deserialize(ret1)
	if !g1.IsEqual(g1New) {
		t.Fatal("Not Equal (1)")
	}

	g2 := &G2{}
	g2.ScalarBaseMult(big.NewInt(1))
	ret2True := g2.Marshal()
	ret2 := g2.SerializeCompressed()
	// Ensure format flag is correct; y is even
	if ret2[0] != (g2CompFlag | g2YXOddFlag) {
		t.Fatal("Invalid byte format (2)")
	}
	// Ensure length is correct
	if 2*numBytes+1 != len(ret2) {
		t.Fatal("Incorrect byte length (2)")
	}
	// Ensure byte slice values for x coordinate are correct
	for j := 0; j < 2*numBytes; j++ {
		if ret2True[j] != ret2[j+1] {
			t.Fatal("Invalid compressed serialization (2)")
		}
	}
	// Check deserialization returns same point
	g2New := &G2{}
	g2New.Deserialize(ret2)
	if !g2.IsEqual(g2New) {
		t.Fatal("Not Equal (2)")
	}

	// Want to negate return value and ensure deserializes properly;
	// thus, adding points together should equal identity element
	ret2Neg := g2.SerializeCompressed()
	v2 := byte(uint8(ret2Neg[0]))
	ret2Neg[0] = ((v2 & (g2YXOddFlag | g2YYOddFlag)) ^ (g2YXOddFlag | g2YYOddFlag)) | g2CompFlag
	g2Neg := &G2{}
	err := g2Neg.Deserialize(ret2Neg)
	if err != nil {
		t.Fatal(err)
	}
	g2Add := &G2{}
	g2Add.Add(g2, g2Neg)
	if !g2Add.IsEqual(g2Identity) {
		t.Fatal("Should equal Identity (2)")
	}

	g3 := &G2{}
	g3.ScalarBaseMult(big.NewInt(2))
	ret3True := g3.Marshal()
	ret3 := g3.SerializeCompressed()
	// Ensure format flag is correct
	if ret3[0] != g2CompFlag {
		t.Fatal("Invalid byte format (3)")
	}
	// Ensure length is correct
	if 2*numBytes+1 != len(ret3) {
		t.Fatal("Incorrect byte length (3)")
	}
	// Ensure byte slice values for x coordinate are correct
	for j := 0; j < 2*numBytes; j++ {
		if ret3True[j] != ret3[j+1] {
			t.Fatal("Invalid uncompressed serialization (3)")
		}
	}
	// Check deserialization returns same point
	g3New := &G2{}
	g3New.Deserialize(ret3)
	if !g3.IsEqual(g3New) {
		t.Fatal("Not Equal (3)")
	}

	// Want to negate return value and ensure deserializes properly;
	// thus, adding points together should equal identity element
	ret3Neg := g3.SerializeCompressed()
	v3 := byte(uint8(ret3Neg[0]))
	ret3Neg[0] = ((v3 & (g2YXOddFlag | g2YYOddFlag)) ^ (g2YXOddFlag | g2YYOddFlag)) | g2CompFlag
	g3Neg := &G2{}
	err = g3Neg.Deserialize(ret3Neg)
	if err != nil {
		t.Fatal(err)
	}
	g3Add := &G2{}
	g3Add.Add(g3, g3Neg)
	if !g3Add.IsEqual(g2Identity) {
		t.Fatal("Should equal Identity (3)")
	}

	g4 := &G2{}
	g4.ScalarBaseMult(big.NewInt(3))
	ret4True := g4.Marshal()
	ret4 := g4.SerializeCompressed()
	// Ensure format flag is correct
	if ret4[0] != (g2CompFlag | g2YYOddFlag) {
		t.Fatal("Invalid byte format (4)")
	}
	// Ensure length is correct
	if 2*numBytes+1 != len(ret4) {
		t.Fatal("Incorrect byte length (4)")
	}
	// Ensure byte slice values for x coordinate are correct
	for j := 0; j < 2*numBytes; j++ {
		if ret4True[j] != ret4[j+1] {
			t.Fatal("Invalid uncompressed serialization (4)")
		}
	}
	// Check deserialization returns same point
	g4New := &G2{}
	g4New.Deserialize(ret4)
	if !g4.IsEqual(g4New) {
		t.Fatal("Not Equal (4)")
	}

	// Want to negate return value and ensure deserializes properly;
	// thus, adding points together should equal identity element
	ret4Neg := g4.SerializeCompressed()
	v4 := byte(uint8(ret4Neg[0]))
	ret4Neg[0] = ((v4 & (g2YXOddFlag | g2YYOddFlag)) ^ (g2YXOddFlag | g2YYOddFlag)) | g2CompFlag
	g4Neg := &G2{}
	err = g4Neg.Deserialize(ret4Neg)
	if err != nil {
		t.Fatal(err)
	}
	g4Add := &G2{}
	g4Add.Add(g4, g4Neg)
	if !g4Add.IsEqual(g2Identity) {
		t.Fatal("Should equal Identity (4)")
	}

	g5 := &G2{}
	g5.ScalarBaseMult(big.NewInt(10))
	ret5True := g5.Marshal()
	ret5 := g5.SerializeCompressed()
	// Ensure format flag is correct
	if ret5[0] != (g2CompFlag | g2YXOddFlag | g2YYOddFlag) {
		t.Fatal("Invalid byte format (5)")
	}
	// Ensure length is correct
	if 2*numBytes+1 != len(ret5) {
		t.Fatal("Incorrect byte length (5)")
	}
	// Ensure byte slice values for x coordinate are correct
	for j := 0; j < 2*numBytes; j++ {
		if ret5True[j] != ret5[j+1] {
			t.Fatal("Invalid uncompressed serialization (5)")
		}
	}
	// Check deserialization returns same point
	g5New := &G2{}
	g5New.Deserialize(ret5)
	if !g5.IsEqual(g5New) {
		t.Fatal("Not Equal (5)")
	}

	// Want to negate return value and ensure deserializes properly;
	// thus, adding points together should equal identity element
	ret5Neg := g5.SerializeCompressed()
	v5 := byte(uint8(ret5Neg[0]))
	ret5Neg[0] = ((v5 & (g2YXOddFlag | g2YYOddFlag)) ^ (g2YXOddFlag | g2YYOddFlag)) | g2CompFlag
	g5Neg := &G2{}
	err = g5Neg.Deserialize(ret5Neg)
	if err != nil {
		t.Fatal(err)
	}
	g5Add := &G2{}
	g5Add.Add(g5, g5Neg)
	if !g5Add.IsEqual(g2Identity) {
		t.Fatal("Should equal Identity (5)")
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

func TestG2Deserialize(t *testing.T) {
	g1 := &G2{}
	m1 := make([]byte, 0)
	err1 := g1.Deserialize(m1)
	// Should raise error due to slice length 0
	if err1 == nil {
		t.Fatal("Should have raised error (1)")
	}

	g2 := &G2{}
	m2 := []byte{255}
	err2 := g2.Deserialize(m2)
	// Should raise error due to invalid format byte
	if err2 == nil {
		t.Fatal("Should have raised error (2)")
	}

	g3 := &G2{}
	m3 := []byte{g2UncompFlag}
	err3 := g3.Deserialize(m3)
	// Should raise error due to incorrect slice length
	if err3 == nil {
		t.Fatal("Should have raised error (3)")
	}

	g4 := &G2{}
	m4 := []byte{g2CompFlag}
	err4 := g4.Deserialize(m4)
	// Should raise error due to incorrect slice length
	if err4 == nil {
		t.Fatal("Should have raised error (4)")
	}

	g5 := &G2{}
	m5 := []byte{g2CompFlag | g2YXOddFlag}
	err5 := g5.Deserialize(m5)
	// Should raise error due to incorrect slice length
	if err5 == nil {
		t.Fatal("Should have raised error (5)")
	}

	g6 := &G2{}
	m6 := []byte{g2CompFlag | g2YYOddFlag}
	err6 := g6.Deserialize(m6)
	// Should raise error due to incorrect slice length
	if err6 == nil {
		t.Fatal("Should have raised error (6)")
	}

	g7 := &G2{}
	m7 := []byte{g2CompFlag | g2YXOddFlag | g2YYOddFlag}
	err7 := g7.Deserialize(m7)
	// Should raise error due to incorrect slice length
	if err7 == nil {
		t.Fatal("Should have raised error (7)")
	}

	// Invalid data for uncompressed; invalid point
	g8 := &G2{}
	m8 := make([]byte, 1+4*numBytes, 1+4*numBytes)
	m8[0] = g2UncompFlag
	m8[1] = 1
	err8 := g8.Deserialize(m8)
	if err8 == nil {
		t.Fatal("Should have raised error (8)")
	}

	// Valid data for compressed identity element
	g9 := &G2{}
	m9 := make([]byte, 1+2*numBytes, 1+2*numBytes)
	m9[0] = g2CompFlag
	err9 := g9.Deserialize(m9)
	if err9 != nil {
		t.Fatal(err9)
	}

	// Invalid data for compressed identity element: yXOddFlag set
	g10 := &G2{}
	m10 := make([]byte, 1+2*numBytes, 1+2*numBytes)
	m10[0] = g2CompFlag | g2YXOddFlag
	err10 := g10.Deserialize(m10)
	if err10 == nil {
		t.Fatal("Should have raised error (10)")
	}

	// Invalid data for compressed identity element: yYOddFlag set
	g11 := &G2{}
	m11 := make([]byte, 1+2*numBytes, 1+2*numBytes)
	m11[0] = g2CompFlag | g2YYOddFlag
	err11 := g11.Deserialize(m11)
	if err11 == nil {
		t.Fatal("Should have raised error (11)")
	}

	// Invalid data for compressed identity element: yXOddFlag and yYOddFlag set
	g12 := &G2{}
	m12 := make([]byte, 1+2*numBytes, 1+2*numBytes)
	m12[0] = g2CompFlag | g2YXOddFlag | g2YYOddFlag
	err12 := g12.Deserialize(m12)
	if err12 == nil {
		t.Fatal("Should have raised error (12)")
	}

	// Invalid data for compressed data: x.x coordinate too large
	g13 := &G2{}
	m13 := make([]byte, 1+2*numBytes, 1+2*numBytes)
	m13[0] = g2CompFlag
	m13[1] = 255
	err13 := g13.Deserialize(m13)
	if err13 == nil {
		t.Fatal("Should have raised error (13)")
	}

	// Invalid data for compressed data: x.y coordinate too large
	g14 := &G2{}
	m14 := make([]byte, 1+2*numBytes, 1+2*numBytes)
	m14[0] = g2CompFlag
	m14[1+numBytes] = 255
	err14 := g14.Deserialize(m14)
	if err14 == nil {
		t.Fatal("Should have raised error (14)")
	}

	// Invalid data for compressed data: no point has x == 4
	g15 := &G2{}
	m15 := make([]byte, 1+2*numBytes, 1+2*numBytes)
	m15[0] = g2CompFlag
	m15[2*numBytes] = 1
	err15 := g15.Deserialize(m15)
	if err15 == nil {
		t.Fatal("Should have raised error (15)")
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

func TestComputeG2YValue(t *testing.T) {
	tG := returnTwistGen()
	x := &gfP2{}
	x.Set(&tG.x)
	yXIsOdd := true
	yYIsOdd := false
	y := computeG2YValue(x, yXIsOdd, yYIsOdd)
	yTrue := &gfP2{}
	yTrue.Set(&tG.y)
	if !y.IsEqual(yTrue) {
		t.Fatal("Should be equal (1)")
	}

	yXIsOdd = false
	yYIsOdd = true
	y = computeG2YValue(x, yXIsOdd, yYIsOdd)
	yTrue.Neg(yTrue)
	if !y.IsEqual(yTrue) {
		t.Fatal("Should be equal (2)")
	}
}
