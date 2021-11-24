package bn256

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
)

func TestMarshalBigInt(t *testing.T) {
	// Test 0
	big0 := big.NewInt(0)
	bytes0True := make([]byte, numBytes)
	bytes0, err := MarshalBigInt(big0)
	if err != nil {
		t.Fatal("Unexpected error occurred in MarshalBigInt (0)")
	}
	if !bytes.Equal(bytes0True, bytes0) {
		t.Fatal("Should have equality for 0")
	}

	// Test 1
	big1 := big.NewInt(1)
	bytes1True := make([]byte, numBytes)
	bytes1True[31] = 1
	bytes1, err := MarshalBigInt(big1)
	if err != nil {
		t.Fatal("Unexpected error occurred in MarshalBigInt (1)")
	}
	if !bytes.Equal(bytes1True, bytes1) {
		t.Fatal("Should have equality for 1")
	}

	// Test 257 == 2^8 + 1
	big257 := big.NewInt(257)
	bytes257True := make([]byte, numBytes)
	bytes257True[31] = 1
	bytes257True[30] = 1
	bytes257, err := MarshalBigInt(big257)
	if err != nil {
		t.Fatal("Unexpected error occurred in MarshalBigInt (257)")
	}
	if !bytes.Equal(bytes257True, bytes257) {
		t.Fatal("Should have equality for 257")
	}

	// Test 65537 == 2^16 + 1
	big65537 := big.NewInt(65537)
	bytes65537True := make([]byte, numBytes)
	bytes65537True[31] = 1
	bytes65537True[29] = 1
	bytes65537, err := MarshalBigInt(big65537)
	if err != nil {
		t.Fatal("Unexpected error occurred in MarshalBigInt (65537)")
	}
	if !bytes.Equal(bytes65537True, bytes65537) {
		t.Fatal("Should have equality for 65537")
	}

	// Test 2^255 - 19
	big25519 := big.NewInt(1)
	big19 := big.NewInt(19)
	big25519.Lsh(big25519, 255)
	big25519.Sub(big25519, big19)
	bytes25519True := make([]byte, numBytes)
	bytes25519True[0] = 127
	for k := 1; k < (numBytes - 1); k++ {
		bytes25519True[k] = 255
	}
	bytes25519True[31] = 237
	bytes25519, err := MarshalBigInt(big25519)
	if err != nil {
		t.Fatal("Unexpected error occurred in MarshalBigInt (2^255 - 19)")
	}
	if !bytes.Equal(bytes25519True, bytes25519) {
		t.Fatal("Should have equality for 2^255 - 19")
	}

	// This test will fail
	tooBigInt, _ := new(big.Int).SetString("1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111", 10)
	_, err = MarshalBigInt(tooBigInt)
	if err != ErrNotUint256 {
		t.Fatal("Should have raised ErrNotUint256 error!")
	}
}

func TestMarshalBigIntBad(t *testing.T) {
	_, err := MarshalBigInt(nil)
	if err == nil {
		t.Fatal("Should raise an error")
	}
}

func TestMarshalG1(t *testing.T) {
	// True definitions and standard tests
	g1Gen := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	g1GenMarshTrue := g1Gen.Marshal()
	g1GenB := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	g1GenMarsh, err := MarshalG1Big(g1GenB)
	if err != nil {
		t.Fatal("No error should arise (gen)")
	}
	if !bytes.Equal(g1GenMarsh, g1GenMarshTrue) {
		t.Fatal("MarshalG1Big fails to agree between Cloudflare and bn256 on curveGen")
	}

	g1Inf := new(cloudflare.G1).ScalarBaseMult(cloudflare.Order)
	g1InfMarshTrue := g1Inf.Marshal()
	g1InfB := [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	g1InfMarsh, err := MarshalG1Big(g1InfB)
	if err != nil {
		t.Fatal("No error should arise (inf)")
	}
	if !bytes.Equal(g1InfMarsh, g1InfMarshTrue) {
		t.Fatal("MarshalG1Big fails to agree between Cloudflare and bn256 on Infinity")
	}

	// Look at an ``arbitrary'' G1 point;
	// chose first value such that both x and y coordinates < 2^248, so that
	// their left-most byte is 0.
	xArb, _ := new(big.Int).SetString("298773868438273703498850243504184491106691330065064935169702410332558258242", 10)
	yArb, _ := new(big.Int).SetString("277967159421498597465334996757796866882824842668492009308772372720306380794", 10)
	arb := [2]*big.Int{xArb, yArb}
	g1Arb := new(cloudflare.G1).ScalarBaseMult(big.NewInt(5055))
	g1ArbMarshTrue := g1Arb.Marshal()
	g1ArbMarsh, err := MarshalG1Big(arb)
	if err != nil {
		t.Fatal("No error should arise (arb)")
	}
	if !bytes.Equal(g1ArbMarsh, g1ArbMarshTrue) {
		t.Fatal("MarshalG1Big fails to agree between Cloudflare and bn256 on Arb")
	}

	tooBigInt, _ := new(big.Int).SetString("1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111", 10)
	badX := [2]*big.Int{tooBigInt, big.NewInt(1)}
	_, err = MarshalG1Big(badX)
	if err != ErrNotUint256 {
		t.Fatal("Error should have arisen in x coordinate")
	}

	badY := [2]*big.Int{big.NewInt(1), tooBigInt}
	_, err = MarshalG1Big(badY)
	if err != ErrNotUint256 {
		t.Fatal("Error should have arisen in y coordinate")
	}
}

func TestMarshalG1Bad(t *testing.T) {
	bad0 := [2]*big.Int{}
	_, err := MarshalG1Big(bad0)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	bad1 := [2]*big.Int{nil, nil}
	_, err = MarshalG1Big(bad1)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	big1 := big.NewInt(1)
	bad2 := [2]*big.Int{big1, nil}
	_, err = MarshalG1Big(bad2)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	bad3 := [2]*big.Int{nil, big1}
	_, err = MarshalG1Big(bad3)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func TestG1ToBigArray(t *testing.T) {
	// True definitions and standard tests
	g1Gen := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	g1GenBTrue := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	g1GenB, err := G1ToBigIntArray(g1Gen)
	if err != nil {
		t.Fatal(err)
	}
	if (g1GenB[0].Cmp(g1GenBTrue[0]) != 0) || (g1GenB[1].Cmp(g1GenBTrue[1]) != 0) {
		t.Fatal("G1ToBigArray failed for curveGen")
	}

	g1Inf := new(cloudflare.G1).ScalarBaseMult(cloudflare.Order)
	g1InfBTrue := [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	g1InfB, err := G1ToBigIntArray(g1Inf)
	if err != nil {
		t.Fatal(err)
	}
	if (g1InfB[0].Cmp(g1InfBTrue[0]) != 0) || (g1InfB[1].Cmp(g1InfBTrue[1]) != 0) {
		t.Fatal("G1ToBigArray failed for Infinity")
	}

	// Look at an ``arbitrary'' G1 point;
	// chose first value such that both x and y coordinates < 2^248, so that
	// their left-most byte is 0.
	g1Arb := new(cloudflare.G1).ScalarBaseMult(big.NewInt(5055))
	xArb, _ := new(big.Int).SetString("298773868438273703498850243504184491106691330065064935169702410332558258242", 10)
	yArb, _ := new(big.Int).SetString("277967159421498597465334996757796866882824842668492009308772372720306380794", 10)
	g1ArbBTrue := [2]*big.Int{xArb, yArb}
	g1ArbB, err := G1ToBigIntArray(g1Arb)
	if err != nil {
		t.Fatal(err)
	}
	if (g1ArbB[0].Cmp(g1ArbBTrue[0]) != 0) || (g1ArbB[1].Cmp(g1ArbBTrue[1]) != 0) {
		t.Fatal("G1ToBigArray failed for g1Arb")
	}

}

func TestG1ToBigArrayBad(t *testing.T) {
	_, err := G1ToBigIntArray(nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestBigIntArrayToG1(t *testing.T) {
	g1InfBig := [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	g1InfTrue := new(cloudflare.G1).ScalarBaseMult(cloudflare.Order)
	g1Inf, err := BigIntArrayToG1(g1InfBig)
	if err != nil {
		t.Fatal("Unexpected error occurred in converting BigInt to G1 (identity)")
	}
	if !g1InfTrue.IsEqual(g1Inf) {
		t.Fatal("Failed to produce g1Inf correctly")
	}

	g1GenBig := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	g1Gen, err := BigIntArrayToG1(g1GenBig)
	if err != nil {
		t.Fatal("Unexpected error occurred in converting BigInt to G1 (generator)")
	}
	g1GenTrue := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	if !g1GenTrue.IsEqual(g1Gen) {
		t.Fatal("Failed to produce g1Gen correctly")
	}

	// Look at an ``arbitrary'' G1 point;
	// chose first value such that both x and y coordinates < 2^248, so that
	// their left-most byte is 0.
	g1ArbTrue := new(cloudflare.G1).ScalarBaseMult(big.NewInt(5055))
	xArb, _ := new(big.Int).SetString("298773868438273703498850243504184491106691330065064935169702410332558258242", 10)
	yArb, _ := new(big.Int).SetString("277967159421498597465334996757796866882824842668492009308772372720306380794", 10)
	g1ArbBig := [2]*big.Int{xArb, yArb}
	g1Arb, err := BigIntArrayToG1(g1ArbBig)
	if err != nil {
		t.Fatal("Unexpected error occurred in converting BigInt to G1 (arbitrary)")
	}
	if !g1ArbTrue.IsEqual(g1Arb) {
		t.Fatal("Failed to produce g1Arb correctly")
	}

	// Not on curve, so conversion should fail
	g1BadBig := [2]*big.Int{big.NewInt(1), big.NewInt(3)}
	_, err = BigIntArrayToG1(g1BadBig)
	if err == nil {
		t.Fatal("Should have raised error for invalid G1 point")
	}

	// Raise an error when converting to G1
	tooBigInt, _ := new(big.Int).SetString("1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111", 10)
	g1BadBig2 := [2]*big.Int{tooBigInt, big.NewInt(1)}
	_, err = BigIntArrayToG1(g1BadBig2)
	if err == nil {
		t.Fatal("Should have raised error for invalid G1 point (big.Int too large)")
	}
}

func TestBigIntArrayToG1Bad(t *testing.T) {
	_, err := BigIntArrayToG1([2]*big.Int{})
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}
	_, err = BigIntArrayToG1([2]*big.Int{nil, nil})
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
}

func TestBigIntArraySliceToG1(t *testing.T) {
	g1InfBig := [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	g1InfTrue := new(cloudflare.G1).ScalarBaseMult(cloudflare.Order)

	g1GenBig := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	g1GenTrue := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))

	// Look at an ``arbitrary'' G1 point;
	// chose first value such that both x and y coordinates < 2^248, so that
	// their left-most byte is 0.
	g1ArbTrue := new(cloudflare.G1).ScalarBaseMult(big.NewInt(5055))
	xArb, _ := new(big.Int).SetString("298773868438273703498850243504184491106691330065064935169702410332558258242", 10)
	yArb, _ := new(big.Int).SetString("277967159421498597465334996757796866882824842668492009308772372720306380794", 10)
	g1ArbBig := [2]*big.Int{xArb, yArb}

	g1BigIntSlice := [][2]*big.Int{g1InfBig, g1GenBig, g1ArbBig}
	g1Slice, err := BigIntArraySliceToG1(g1BigIntSlice)
	if err != nil {
		t.Fatal("Unexpected error occurred in BigIntArraySliceToG1 creation")
	}

	g1Inf := g1Slice[0]
	if !g1InfTrue.IsEqual(g1Inf) {
		t.Fatal("Failed to produce g1Inf correctly")
	}
	g1Gen := g1Slice[1]
	if !g1GenTrue.IsEqual(g1Gen) {
		t.Fatal("Failed to produce g1Gen correctly")
	}
	g1Arb := g1Slice[2]
	if !g1ArbTrue.IsEqual(g1Arb) {
		t.Fatal("Failed to produce g1Arb correctly")
	}

	// Not on curve, so conversion should fail
	g1BadBig := [2]*big.Int{big.NewInt(1), big.NewInt(3)}
	g1BigIntSliceBad := [][2]*big.Int{g1InfBig, g1GenBig, g1ArbBig, g1BadBig}
	_, err = BigIntArraySliceToG1(g1BigIntSliceBad)
	if err == nil {
		t.Fatal("Should have raised error for invalid G1 point")
	}

}

func TestG2ToBigArray(t *testing.T) {
	g2Gen := new(cloudflare.G2).ScalarBaseMult(big.NewInt(1))
	g2GenXI, _ := new(big.Int).SetString("11559732032986387107991004021392285783925812861821192530917403151452391805634", 10)
	g2GenX, _ := new(big.Int).SetString("10857046999023057135944570762232829481370756359578518086990519993285655852781", 10)
	g2GenYI, _ := new(big.Int).SetString("4082367875863433681332203403145435568316851327593401208105741076214120093531", 10)
	g2GenY, _ := new(big.Int).SetString("8495653923123431417604973247489272438418190587263600148770280649306958101930", 10)
	g2GenBTrue := [4]*big.Int{g2GenXI, g2GenX, g2GenYI, g2GenY}
	g2GenB, err := G2ToBigIntArray(g2Gen)
	if err != nil {
		t.Fatal(err)
	}
	if (g2GenB[0].Cmp(g2GenBTrue[0]) != 0) || (g2GenB[1].Cmp(g2GenBTrue[1]) != 0) || (g2GenB[2].Cmp(g2GenBTrue[2]) != 0) || (g2GenB[3].Cmp(g2GenBTrue[3]) != 0) {
		t.Fatal("G2ToBigArray failed for twistGen")
	}

	g2Inf := new(cloudflare.G2).ScalarBaseMult(cloudflare.Order)
	g2InfBTrue := [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
	g2InfB, err := G2ToBigIntArray(g2Inf)
	if err != nil {
		t.Fatal(err)
	}
	if (g2InfB[0].Cmp(g2InfBTrue[0]) != 0) || (g2InfB[1].Cmp(g2InfBTrue[1]) != 0) || (g2InfB[2].Cmp(g2InfBTrue[2]) != 0) || (g2InfB[3].Cmp(g2InfBTrue[3]) != 0) {
		t.Fatal("G2ToBigArray failed for Infinity")
	}
}

func TestG2ToBigArrayBad(t *testing.T) {
	_, err := G2ToBigIntArray(nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestMarshalG2(t *testing.T) {
	// True definitions and standard tests
	g2Gen := new(cloudflare.G2).ScalarBaseMult(big.NewInt(1))
	g2GenMarshTrue := g2Gen.Marshal()
	g2GenXi, _ := new(big.Int).SetString("11559732032986387107991004021392285783925812861821192530917403151452391805634", 10)
	g2GenX, _ := new(big.Int).SetString("10857046999023057135944570762232829481370756359578518086990519993285655852781", 10)
	g2GenYi, _ := new(big.Int).SetString("4082367875863433681332203403145435568316851327593401208105741076214120093531", 10)
	g2GenY, _ := new(big.Int).SetString("8495653923123431417604973247489272438418190587263600148770280649306958101930", 10)
	g2GenB := [4]*big.Int{g2GenXi, g2GenX, g2GenYi, g2GenY}
	g2GenMarsh, err := MarshalG2Big(g2GenB)
	if err != nil {
		t.Fatal("No error should arise (gen)")
	}
	if !bytes.Equal(g2GenMarsh, g2GenMarshTrue) {
		t.Fatal("MarshalG2Big fails to agree between Cloudflare and bn256 on twistGen")
	}

	g2Inf := new(cloudflare.G2).ScalarBaseMult(cloudflare.Order)
	g2InfMarshTrue := g2Inf.Marshal()
	g2InfB := [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
	g2InfMarsh, err := MarshalG2Big(g2InfB)
	if err != nil {
		t.Fatal("No error should arise (inf)")
	}
	if !bytes.Equal(g2InfMarsh, g2InfMarshTrue) {
		t.Fatal("MarshalG1Big fails to agree between Cloudflare and bn256 on G2 Infinity")
	}

	tooBigInt, _ := new(big.Int).SetString("1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111", 10)

	badXi := [4]*big.Int{tooBigInt, big.NewInt(1), big.NewInt(1), big.NewInt(1)}
	_, err = MarshalG2Big(badXi)
	if err != ErrNotUint256 {
		t.Fatal("Error should have arisen in xi coordinate")
	}

	badX := [4]*big.Int{big.NewInt(1), tooBigInt, big.NewInt(1), big.NewInt(1)}
	_, err = MarshalG2Big(badX)
	if err != ErrNotUint256 {
		t.Fatal("Error should have arisen in x coordinate")
	}

	badYi := [4]*big.Int{big.NewInt(1), big.NewInt(1), tooBigInt, big.NewInt(1)}
	_, err = MarshalG2Big(badYi)
	if err != ErrNotUint256 {
		t.Fatal("Error should have arisen in yi coordinate")
	}

	badY := [4]*big.Int{big.NewInt(1), big.NewInt(1), big.NewInt(1), tooBigInt}
	_, err = MarshalG2Big(badY)
	if err != ErrNotUint256 {
		t.Fatal("Error should have arisen in y coordinate")
	}
}

func TestMarshalG2Bad(t *testing.T) {
	bad0 := [4]*big.Int{}
	_, err := MarshalG2Big(bad0)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	bad1 := [4]*big.Int{nil, nil, nil, nil}
	_, err = MarshalG2Big(bad1)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	big1 := big.NewInt(1)
	bad2 := [4]*big.Int{nil, big1, big1, big1}
	_, err = MarshalG2Big(bad2)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	bad3 := [4]*big.Int{big1, nil, big1, big1}
	_, err = MarshalG2Big(bad3)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	bad4 := [4]*big.Int{big1, big1, nil, big1}
	_, err = MarshalG2Big(bad4)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	bad5 := [4]*big.Int{big1, big1, big1, nil}
	_, err = MarshalG2Big(bad5)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}
}

func TestBigIntArrayToG2(t *testing.T) {
	g2InfBig := [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
	g2InfTrue := new(cloudflare.G2).ScalarBaseMult(cloudflare.Order)
	g2Inf, err := BigIntArrayToG2(g2InfBig)
	if err != nil {
		t.Fatal("Unexpected error occurred in converting BigInt to G2 (identity)")
	}
	if !g2InfTrue.IsEqual(g2Inf) {
		t.Fatal("Failed to produce g2Inf correctly")
	}

	g2GenXi, _ := new(big.Int).SetString("11559732032986387107991004021392285783925812861821192530917403151452391805634", 10)
	g2GenX, _ := new(big.Int).SetString("10857046999023057135944570762232829481370756359578518086990519993285655852781", 10)
	g2GenYi, _ := new(big.Int).SetString("4082367875863433681332203403145435568316851327593401208105741076214120093531", 10)
	g2GenY, _ := new(big.Int).SetString("8495653923123431417604973247489272438418190587263600148770280649306958101930", 10)
	g2GenBig := [4]*big.Int{g2GenXi, g2GenX, g2GenYi, g2GenY}
	g2Gen, err := BigIntArrayToG2(g2GenBig)
	if err != nil {
		t.Fatal("Unexpected error occurred in converting BigInt to G1 (generator)")
	}
	g2GenTrue := new(cloudflare.G2).ScalarBaseMult(big.NewInt(1))
	if !g2GenTrue.IsEqual(g2Gen) {
		t.Fatal("Failed to produce g2Gen correctly")
	}

	// Not on curve, so conversion should fail
	g2BadBig := [4]*big.Int{big.NewInt(1), big.NewInt(3), big.NewInt(5), big.NewInt(7)}
	_, err = BigIntArrayToG2(g2BadBig)
	if err == nil {
		t.Fatal("Should have raised error for invalid G2 point")
	}

	// Raise an error when converting to G2
	tooBigInt, _ := new(big.Int).SetString("1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111", 10)
	g2BadBig2 := [4]*big.Int{tooBigInt, big.NewInt(1), big.NewInt(1), big.NewInt(1)}
	_, err = BigIntArrayToG2(g2BadBig2)
	if err == nil {
		t.Fatal("Should have raised error for invalid G2 point (big.Int too large)")
	}
}

func TestBigIntArrayToG2Bad(t *testing.T) {
	_, err := BigIntArrayToG2([4]*big.Int{})
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}
	_, err = BigIntArrayToG2([4]*big.Int{nil, nil, nil, nil})
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
}

func TestMarshalBigIntSlice(t *testing.T) {
	bigSlice := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}
	_, err := MarshalBigIntSlice(bigSlice)
	if err != nil {
		t.Fatal("big.Int slice to byte slice should have succeeded")
	}

	tooBigInt, _ := new(big.Int).SetString("1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111", 10)
	bigSliceBad := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3), tooBigInt}
	_, err = MarshalBigIntSlice(bigSliceBad)
	if err != ErrNotUint256 {
		t.Fatal("ErrNotUint256 should have occurred")
	}

	res, err := MarshalBigIntSlice([]*big.Int{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 0 {
		t.Fatal("incorrect result")
	}
}

func TestMarshalBigIntSliceBad(t *testing.T) {
	_, err := MarshalBigIntSlice([]*big.Int{nil})
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestMarshalG1BigSlice(t *testing.T) {
	g1BigInf := [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	g1BigGen := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	g1BigSlice := [][2]*big.Int{g1BigGen, g1BigInf}
	g1BytesTrue := make([]byte, 4*numBytes)
	g1BytesTrue[31] = 1
	g1BytesTrue[63] = 2
	g1Bytes, err := MarshalG1BigSlice(g1BigSlice)
	if err != nil {
		t.Fatal("Unexpected error occurred")
	}
	if !bytes.Equal(g1Bytes, g1BytesTrue) {
		t.Fatal("Byte slices should be equal")
	}

	tooBigInt, _ := new(big.Int).SetString("1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111", 10)
	g1BigBad := [2]*big.Int{tooBigInt, tooBigInt}
	g1BigSliceBad := [][2]*big.Int{g1BigBad, g1BigGen, g1BigInf}
	_, err = MarshalG1BigSlice(g1BigSliceBad)
	if err == nil {
		t.Fatal("Should have raised an error due to invalid g1Big (too large)")
	}
}

func TestMarshalG1BigSliceBad(t *testing.T) {
	g1BigInf := [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	g1BigGen := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	g1BigBad0 := [2]*big.Int{}
	g1BigSlice0 := [][2]*big.Int{g1BigGen, g1BigInf, g1BigBad0}
	_, err := MarshalG1BigSlice(g1BigSlice0)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}
	g1BigBad1 := [2]*big.Int{nil, nil}
	g1BigSlice1 := [][2]*big.Int{g1BigGen, g1BigInf, g1BigBad1}
	_, err = MarshalG1BigSlice(g1BigSlice1)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

}
