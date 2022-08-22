package cloudflare

import (
	"crypto/rand"
	"math/big"
	"testing"
)

func TestPrivatePolyEval(t *testing.T) {
	privCoefs := []*big.Int{big.NewInt(314), big.NewInt(159), big.NewInt(265)}

	true1 := big.NewInt(738)  // 738 == 314 + 159*1 + 265*1^2
	true2 := big.NewInt(1692) // 1692 == 314 + 159*2 + 265*2^2
	true3 := big.NewInt(3176) // 3176 == 314 + 159*3 + 265*3^2

	res1 := PrivatePolyEval(privCoefs, 1)
	if res1.Cmp(true1) != 0 {
		t.Fatal("Error occurred in PrivatePolyEval for j = 1")
	}
	res2 := PrivatePolyEval(privCoefs, 2)
	if res2.Cmp(true2) != 0 {
		t.Fatal("Error occurred in PrivatePolyEval for j = 2")
	}
	res3 := PrivatePolyEval(privCoefs, 3)
	if res3.Cmp(true3) != 0 {
		t.Fatal("Error occurred in PrivatePolyEval for j = 3")
	}
	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("PrivatePolyEval changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("PrivatePolyEval changed twistGen")
	}
}

func TestGeneratePublicCoefs(t *testing.T) {
	privCoefs := []*big.Int{big.NewInt(314), big.NewInt(159), big.NewInt(265)}
	gC0 := new(G1).ScalarBaseMult(privCoefs[0])
	gC1 := new(G1).ScalarBaseMult(privCoefs[1])
	gC2 := new(G1).ScalarBaseMult(privCoefs[2])

	truePubCoefs := []*G1{gC0, gC1, gC2}

	pubCoefs := GeneratePublicCoefs(privCoefs)
	for j := 0; j < len(privCoefs); j++ {
		if !pubCoefs[j].IsEqual(truePubCoefs[j]) {
			t.Fatal("Error in GeneratePublicCoefs for j =", j)
		}
	}
	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("GeneratePublicCoefs changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("GeneratePublicCoefs changed twistGen")
	}
}

func TestCompareSharedSecret(t *testing.T) {
	privCoefs := []*big.Int{big.NewInt(314), big.NewInt(159), big.NewInt(265)}
	truePubCoefs := GeneratePublicCoefs(privCoefs)

	true1 := big.NewInt(738)  // 738 == 314 + 159*1 + 265*1^2
	true2 := big.NewInt(1692) // 1692 == 314 + 159*2 + 265*2^2
	true3 := big.NewInt(3176) // 3176 == 314 + 159*3 + 265*3^2

	j := 1
	valid1, err1 := CompareSharedSecret(true1, j, truePubCoefs)
	if err1 != nil {
		t.Fatal("CompareSharedSecret raised an error for j = 1")
	}
	if !valid1 {
		t.Fatal("CompareSharedSecret failed to return valid for j = 1")
	}

	j = 2
	valid2, err2 := CompareSharedSecret(true2, j, truePubCoefs)
	if err2 != nil {
		t.Fatal("CompareSharedSecret raised an error for j = 2")
	}
	if !valid2 {
		t.Fatal("CompareSharedSecret failed to return valid for j = 2")
	}

	j = 3
	valid3, err3 := CompareSharedSecret(true3, j, truePubCoefs)
	if err3 != nil {
		t.Fatal("CompareSharedSecret raised an error for j = 3")
	}
	if !valid3 {
		t.Fatal("CompareSharedSecret failed to return valid for j = 3")
	}

	validBad, errBad := CompareSharedSecret(true1, j, truePubCoefs)
	if errBad != nil {
		t.Fatal("CompareSharedSecret failed to return nil for bad comparison")
	}
	if validBad {
		t.Fatal("CompareSharedSecret failed to return false for bad comparison")
	}

	_, errBad2 := CompareSharedSecret(true1, 0, truePubCoefs)
	if errBad2 == nil {
		t.Fatal("Should have raised error")
	}

	_, err := CompareSharedSecret(nil, j, truePubCoefs)
	if err == nil {
		t.Fatal("Should have raised error for invalid secret")
	}

	_, err = CompareSharedSecret(true1, j, []*G1{nil})
	if err == nil {
		t.Fatal("Should have raised error for invalid public coefficients")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("CompareSharedSecret changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("CompareSharedSecret changed twistGen")
	}
}

func TestPrivPubKeyGen(t *testing.T) {
	_, _, err := GeneratePrivatePublicKeys(rand.Reader)
	if err != nil {
		t.Fatal("Error occurred in GenPrivPubKeys")
	}
}

func TestConstructPrivPolyCoefs(t *testing.T) {
	threshold := 0
	_, err := ConstructPrivatePolyCoefs(rand.Reader, threshold)
	if err == nil {
		t.Fatal("Failed to determine invalid threshold value")
	}

	threshold = 4
	_, err = ConstructPrivatePolyCoefs(rand.Reader, threshold)
	if err != nil {
		t.Fatal("Failed to compute coefs")
	}
}

func TestGenerateSharedSecret(t *testing.T) {
	privK := big.NewInt(1)
	pubK := new(G1).ScalarBaseMult(privK)
	kX, kY := GenerateSharedSecret(privK, pubK)
	big1 := big.NewInt(1)
	big2 := big.NewInt(2)
	if (kX.Cmp(big1) != 0) || (kY.Cmp(big2) != 0) {
		t.Fatal("Failed to correctly generate shared secrets")
	}

	privK1 := big.NewInt(1234567890)
	privK2 := big.NewInt(2357111317)
	pubK2 := new(G1).ScalarBaseMult(privK2)
	kXRand, kYRand := GenerateSharedSecret(privK1, pubK2)
	sharedExp := new(big.Int).Mul(privK1, privK2)
	sharedExp.Mod(sharedExp, Order)
	sharedPoint := new(G1).ScalarBaseMult(sharedExp)
	sharedPointBytes := sharedPoint.Marshal()
	sharedXBytes := sharedPointBytes[:numBytes]
	sharedYBytes := sharedPointBytes[numBytes : 2*numBytes]
	sharedX := new(big.Int).SetBytes(sharedXBytes)
	sharedY := new(big.Int).SetBytes(sharedYBytes)
	if (kXRand.Cmp(sharedX) != 0) || (kYRand.Cmp(sharedY) != 0) {
		t.Fatal("Failed to correctly generate shared secrets for random points")
	}
}

func TestGenerateSecretShares(t *testing.T) {
	privK1 := big.NewInt(10)
	pubK1 := new(G1).ScalarBaseMult(privK1)
	privCoefs1 := []*big.Int{big.NewInt(100), big.NewInt(1), big.NewInt(2)}

	privK2 := big.NewInt(11)
	pubK2 := new(G1).ScalarBaseMult(privK2)
	privCoefs2 := []*big.Int{big.NewInt(101), big.NewInt(1), big.NewInt(2)}

	privK3 := big.NewInt(12)
	pubK3 := new(G1).ScalarBaseMult(privK3)
	privCoefs3 := []*big.Int{big.NewInt(102), big.NewInt(1), big.NewInt(2)}

	privK4 := big.NewInt(13)
	pubK4 := new(G1).ScalarBaseMult(privK4)
	privCoefs4 := []*big.Int{big.NewInt(103), big.NewInt(1), big.NewInt(2)}

	pubKs := []*G1{pubK1, pubK2, pubK3, pubK4}

	secretValues1, err := GenerateSecretShares(pubK1, privCoefs1, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secrets 1")
	}
	// Secrets to share:
	//
	//		share1to1 = 100 + 1*1 + 2*1^2 = 103
	//		share1to2 = 100 + 1*2 + 2*2^2 = 110
	//		share1to3 = 100 + 1*3 + 2*3^2 = 121
	//		share1to4 = 100 + 1*4 + 2*4^2 = 136
	//share1to1 := big.NewInt(103)
	share1to2 := big.NewInt(110)
	share1to3 := big.NewInt(121)
	share1to4 := big.NewInt(136)
	secretValues1to2 := secretValues1[0]
	secretValues1to3 := secretValues1[1]
	secretValues1to4 := secretValues1[2]
	if secretValues1to2.Cmp(share1to2) != 0 {
		t.Fatal("Error in GenerateSecretShares for 1 to 2")
	}
	if secretValues1to3.Cmp(share1to3) != 0 {
		t.Fatal("Error in GenerateSecretShares for 1 to 3")
	}
	if secretValues1to4.Cmp(share1to4) != 0 {
		t.Fatal("Error in GenerateSecretShares for 1 to 4")
	}

	secretValues2, err := GenerateSecretShares(pubK2, privCoefs2, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secrets 2")
	}
	// Secrets to share:
	//
	//		share2to1 = 101 + 1*1 + 2*1^2 = 104
	//		share2to2 = 101 + 1*2 + 2*2^2 = 111
	//		share2to3 = 101 + 1*3 + 2*3^2 = 122
	//		share2to4 = 101 + 1*4 + 2*4^2 = 137
	share2to1 := big.NewInt(104)
	//share2to2 := big.NewInt(111)
	share2to3 := big.NewInt(122)
	share2to4 := big.NewInt(137)
	secretValues2to1 := secretValues2[0]
	secretValues2to3 := secretValues2[1]
	secretValues2to4 := secretValues2[2]
	if secretValues2to1.Cmp(share2to1) != 0 {
		t.Fatal("Error in GenerateSecretShares for 2 to 1")
	}
	if secretValues2to3.Cmp(share2to3) != 0 {
		t.Fatal("Error in GenerateSecretShares for 2 to 3")
	}
	if secretValues2to4.Cmp(share2to4) != 0 {
		t.Fatal("Error in GenerateSecretShares for 2 to 4")
	}

	secretValues3, err := GenerateSecretShares(pubK3, privCoefs3, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secrets 3")
	}
	// Secrets to share:
	//
	//		share3to1 = 102 + 1*1 + 2*1^2 = 105
	//		share3to2 = 102 + 1*2 + 2*2^2 = 112
	//		share3to3 = 102 + 1*3 + 2*3^2 = 123
	//		share3to4 = 102 + 1*4 + 2*4^2 = 138
	share3to1 := big.NewInt(105)
	share3to2 := big.NewInt(112)
	//share3to3 := big.NewInt(123)
	share3to4 := big.NewInt(138)
	secretValues3to1 := secretValues3[0]
	secretValues3to2 := secretValues3[1]
	secretValues3to4 := secretValues3[2]
	if secretValues3to1.Cmp(share3to1) != 0 {
		t.Fatal("Error in GenerateSecretShares for 3 to 1")
	}
	if secretValues3to2.Cmp(share3to2) != 0 {
		t.Fatal("Error in GenerateSecretShares for 3 to 2")
	}
	if secretValues3to4.Cmp(share3to4) != 0 {
		t.Fatal("Error in GenerateSecretShares for 3 to 4")
	}

	secretValues4, err := GenerateSecretShares(pubK4, privCoefs4, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secrets 4")
	}
	// Secrets to share:
	//
	//		share4to1 = 103 + 1*1 + 2*1^2 = 106
	//		share4to2 = 103 + 1*2 + 2*2^2 = 113
	//		share4to3 = 103 + 1*3 + 2*3^2 = 124
	//		share4to4 = 103 + 1*4 + 2*4^2 = 139
	share4to1 := big.NewInt(106)
	share4to2 := big.NewInt(113)
	share4to3 := big.NewInt(124)
	//share4to4 := big.NewInt(139)
	secretValues4to1 := secretValues4[0]
	secretValues4to2 := secretValues4[1]
	secretValues4to3 := secretValues4[2]
	if secretValues4to1.Cmp(share4to1) != 0 {
		t.Fatal("Error in GenerateSecretShares for 4 to 1")
	}
	if secretValues4to2.Cmp(share4to2) != 0 {
		t.Fatal("Error in GenerateSecretShares for 4 to 2")
	}
	if secretValues4to3.Cmp(share4to3) != 0 {
		t.Fatal("Error in GenerateSecretShares for 4 to 3")
	}
}

func TestGenerateSecretSharesFail(t *testing.T) {
	privK1 := big.NewInt(10)
	pubK1 := new(G1).ScalarBaseMult(privK1)
	privCoefs1 := []*big.Int{big.NewInt(100), big.NewInt(1), big.NewInt(2)}

	privK2 := big.NewInt(11)
	pubK2 := new(G1).ScalarBaseMult(privK2)

	privK3 := big.NewInt(12)
	pubK3 := new(G1).ScalarBaseMult(privK3)

	privK4 := big.NewInt(13)
	pubK4 := new(G1).ScalarBaseMult(privK4)

	pubKs := []*G1{pubK2, pubK3, pubK4}

	_, err := GenerateSecretShares(pubK1, privCoefs1, pubKs)
	if err == nil {
		t.Fatal("Error should have occurred; pubK1 is missing from pubKs")
	}
}

func TestGenerateEncryptedShares(t *testing.T) {
	part1 := 1
	privK1 := big.NewInt(10)
	pubK1 := new(G1).ScalarBaseMult(privK1)
	privCoefs1 := []*big.Int{big.NewInt(100), big.NewInt(1), big.NewInt(2)}

	part2 := 2
	privK2 := big.NewInt(11)
	pubK2 := new(G1).ScalarBaseMult(privK2)
	privCoefs2 := []*big.Int{big.NewInt(101), big.NewInt(1), big.NewInt(2)}

	part3 := 3
	privK3 := big.NewInt(12)
	pubK3 := new(G1).ScalarBaseMult(privK3)
	privCoefs3 := []*big.Int{big.NewInt(102), big.NewInt(1), big.NewInt(2)}

	part4 := 4
	privK4 := big.NewInt(13)
	pubK4 := new(G1).ScalarBaseMult(privK4)
	privCoefs4 := []*big.Int{big.NewInt(103), big.NewInt(1), big.NewInt(2)}

	pubKs := []*G1{pubK1, pubK2, pubK3, pubK4}

	secretValues1, err := GenerateSecretShares(pubK1, privCoefs1, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secret shares 1")
	}
	encryptedValues1, err := GenerateEncryptedShares(secretValues1, privK1, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating commitments 1")
	}
	enc1to2 := encryptedValues1[0]
	enc1to3 := encryptedValues1[1]
	enc1to4 := encryptedValues1[2]
	// Secrets to share:
	//
	//		share1to1 = 100 + 1*1 + 2*1^2 = 103
	//		share1to2 = 100 + 1*2 + 2*2^2 = 110
	//		share1to3 = 100 + 1*3 + 2*3^2 = 121
	//		share1to4 = 100 + 1*4 + 2*4^2 = 136
	//share1to1 := big.NewInt(103)
	share1to2 := big.NewInt(110)
	share1to3 := big.NewInt(121)
	share1to4 := big.NewInt(136)
	decrypted1to2 := Decrypt(enc1to2, privK2, pubK1, part2)
	if decrypted1to2.Cmp(share1to2) != 0 {
		t.Fatal("Error in sharing 1 to 2")
	}
	decrypted1to3 := Decrypt(enc1to3, privK3, pubK1, part3)
	if decrypted1to3.Cmp(share1to3) != 0 {
		t.Fatal("Error in sharing 1 to 3")
	}
	decrypted1to4 := Decrypt(enc1to4, privK4, pubK1, part4)
	if decrypted1to4.Cmp(share1to4) != 0 {
		t.Fatal("Error in sharing 1 to 4")
	}

	secretValues2, err := GenerateSecretShares(pubK2, privCoefs2, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secret shares 2")
	}
	encryptedValues2, err := GenerateEncryptedShares(secretValues2, privK2, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating commitments 2")
	}
	enc2to1 := encryptedValues2[0]
	enc2to3 := encryptedValues2[1]
	enc2to4 := encryptedValues2[2]
	// Secrets to share:
	//
	//		share2to1 = 101 + 1*1 + 2*1^2 = 104
	//		share2to2 = 101 + 1*2 + 2*2^2 = 111
	//		share2to3 = 101 + 1*3 + 2*3^2 = 122
	//		share2to4 = 101 + 1*4 + 2*4^2 = 137
	share2to1 := big.NewInt(104)
	//share2to2 := big.NewInt(111)
	share2to3 := big.NewInt(122)
	share2to4 := big.NewInt(137)
	decrypted2to1 := Decrypt(enc2to1, privK1, pubK2, part1)
	if decrypted2to1.Cmp(share2to1) != 0 {
		t.Fatal("Error in sharing 2 to 1")
	}
	decrypted2to3 := Decrypt(enc2to3, privK3, pubK2, part3)
	if decrypted2to3.Cmp(share2to3) != 0 {
		t.Fatal("Error in sharing 2 to 3")
	}
	decrypted2to4 := Decrypt(enc2to4, privK4, pubK2, part4)
	if decrypted2to4.Cmp(share2to4) != 0 {
		t.Fatal("Error in sharing 2 to 4")
	}

	secretValues3, err := GenerateSecretShares(pubK3, privCoefs3, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secret shares 3")
	}
	encryptedValues3, err := GenerateEncryptedShares(secretValues3, privK3, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating commitments 3")
	}
	enc3to1 := encryptedValues3[0]
	enc3to2 := encryptedValues3[1]
	enc3to4 := encryptedValues3[2]
	// Secrets to share:
	//
	//		share3to1 = 102 + 1*1 + 2*1^2 = 105
	//		share3to2 = 102 + 1*2 + 2*2^2 = 112
	//		share3to3 = 102 + 1*3 + 2*3^2 = 123
	//		share3to4 = 102 + 1*4 + 2*4^2 = 138
	share3to1 := big.NewInt(105)
	share3to2 := big.NewInt(112)
	//share3to3 := big.NewInt(123)
	share3to4 := big.NewInt(138)
	decrypted3to1 := Decrypt(enc3to1, privK1, pubK3, part1)
	if decrypted3to1.Cmp(share3to1) != 0 {
		t.Fatal("Error in sharing 3 to 1")
	}
	decrypted3to2 := Decrypt(enc3to2, privK2, pubK3, part2)
	if decrypted3to2.Cmp(share3to2) != 0 {
		t.Fatal("Error in sharing 3 to 2")
	}
	decrypted3to4 := Decrypt(enc3to4, privK4, pubK3, part4)
	if decrypted3to4.Cmp(share3to4) != 0 {
		t.Fatal("Error in sharing 3 to 4")
	}

	secretValues4, err := GenerateSecretShares(pubK4, privCoefs4, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secret shares 4")
	}
	encryptedValues4, err := GenerateEncryptedShares(secretValues4, privK4, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating commitments 4")
	}
	enc4to1 := encryptedValues4[0]
	enc4to2 := encryptedValues4[1]
	enc4to3 := encryptedValues4[2]
	// Secrets to share:
	//
	//		share4to1 = 103 + 1*1 + 2*1^2 = 106
	//		share4to2 = 103 + 1*2 + 2*2^2 = 113
	//		share4to3 = 103 + 1*3 + 2*3^2 = 124
	//		share4to4 = 103 + 1*4 + 2*4^2 = 139
	share4to1 := big.NewInt(106)
	share4to2 := big.NewInt(113)
	share4to3 := big.NewInt(124)
	//share4to4 := big.NewInt(139)
	decrypted4to1 := Decrypt(enc4to1, privK1, pubK4, part1)
	if decrypted4to1.Cmp(share4to1) != 0 {
		t.Fatal("Error in sharing 4 to 1")
	}
	decrypted4to2 := Decrypt(enc4to2, privK2, pubK4, part2)
	if decrypted4to2.Cmp(share4to2) != 0 {
		t.Fatal("Error in sharing 4 to 2")
	}
	decrypted4to3 := Decrypt(enc4to3, privK3, pubK4, part3)
	if decrypted4to3.Cmp(share4to3) != 0 {
		t.Fatal("Error in sharing 4 to 3")
	}
}

func TestGenerateEncryptedSharesFail(t *testing.T) {
	privK1 := big.NewInt(10)
	pubK1 := new(G1).ScalarBaseMult(privK1)
	privCoefs1 := []*big.Int{big.NewInt(100), big.NewInt(1), big.NewInt(2)}

	privK2 := big.NewInt(11)
	pubK2 := new(G1).ScalarBaseMult(privK2)

	privK3 := big.NewInt(12)
	pubK3 := new(G1).ScalarBaseMult(privK3)

	privK4 := big.NewInt(13)
	pubK4 := new(G1).ScalarBaseMult(privK4)

	pubKs := []*G1{pubK1, pubK2, pubK3, pubK4}

	secretValues1, err := GenerateSecretShares(pubK1, privCoefs1, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secret shares 1")
	}

	pubKs = []*G1{pubK2, pubK3, pubK4}
	_, err = GenerateEncryptedShares(secretValues1, privK1, pubKs)
	if err == nil {
		t.Fatal("Error should have occurred; pubK1 missing from pubKs")
	}
}

func TestCondenseCommitments(t *testing.T) {
	privK1 := big.NewInt(10)
	pubK1 := new(G1).ScalarBaseMult(privK1)
	privCoefs1 := []*big.Int{big.NewInt(100), big.NewInt(1), big.NewInt(2)}

	privK2 := big.NewInt(11)
	pubK2 := new(G1).ScalarBaseMult(privK2)
	privCoefs2 := []*big.Int{big.NewInt(101), big.NewInt(1), big.NewInt(2)}

	privK3 := big.NewInt(12)
	pubK3 := new(G1).ScalarBaseMult(privK3)
	privCoefs3 := []*big.Int{big.NewInt(102), big.NewInt(1), big.NewInt(2)}

	privK4 := big.NewInt(13)
	pubK4 := new(G1).ScalarBaseMult(privK4)
	privCoefs4 := []*big.Int{big.NewInt(103), big.NewInt(1), big.NewInt(2)}

	pubKs := []*G1{pubK1, pubK2, pubK3, pubK4}

	secretValues1, err := GenerateSecretShares(pubK1, privCoefs1, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secret shares 1")
	}
	encryptedValues1, err := GenerateEncryptedShares(secretValues1, privK1, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating commitments 1")
	}
	enc1to2 := encryptedValues1[0]
	enc1to3 := encryptedValues1[1]
	enc1to4 := encryptedValues1[2]

	secretValues2, err := GenerateSecretShares(pubK2, privCoefs2, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secret shares 2")
	}
	encryptedValues2, err := GenerateEncryptedShares(secretValues2, privK2, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating commitments 2")
	}
	enc2to1 := encryptedValues2[0]
	enc2to3 := encryptedValues2[1]
	enc2to4 := encryptedValues2[2]

	secretValues3, err := GenerateSecretShares(pubK3, privCoefs3, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secret shares 3")
	}
	encryptedValues3, err := GenerateEncryptedShares(secretValues3, privK3, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating commitments 3")
	}
	enc3to1 := encryptedValues3[0]
	enc3to2 := encryptedValues3[1]
	enc3to4 := encryptedValues3[2]

	secretValues4, err := GenerateSecretShares(pubK4, privCoefs4, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secret shares 4")
	}
	encryptedValues4, err := GenerateEncryptedShares(secretValues4, privK4, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating commitments 4")
	}
	enc4to1 := encryptedValues4[0]
	enc4to2 := encryptedValues4[1]
	enc4to3 := encryptedValues4[2]

	trueEncValuesFor1 := []*big.Int{enc2to1, enc3to1, enc4to1}
	trueEncValuesFor2 := []*big.Int{enc1to2, enc3to2, enc4to2}
	trueEncValuesFor3 := []*big.Int{enc1to3, enc2to3, enc4to3}
	trueEncValuesFor4 := []*big.Int{enc1to4, enc2to4, enc3to4}

	combinedCommitments1 := make([][]*big.Int, 4)
	// Not doing anything for 0
	combinedCommitments1[1] = make([]*big.Int, 3)
	combinedCommitments1[1] = encryptedValues2
	combinedCommitments1[2] = make([]*big.Int, 3)
	combinedCommitments1[2] = encryptedValues3
	combinedCommitments1[3] = make([]*big.Int, 3)
	combinedCommitments1[3] = encryptedValues4
	encValuesFor1, err := CondenseCommitments(pubK1, combinedCommitments1, pubKs)
	if err != nil {
		t.Fatal("Error when condensing commitments for 1")
	}
	if len(encValuesFor1) != len(trueEncValuesFor1) {
		t.Fatal("encValuesFor1 has incorrect length")
	}
	for j := 0; j < len(encValuesFor1); j++ {
		trueEncVal := trueEncValuesFor1[j]
		encVal := encValuesFor1[j]
		if trueEncVal.Cmp(encVal) != 0 {
			t.Fatal("Error in condensing encrypted values 1")
		}
	}

	combinedCommitments2 := make([][]*big.Int, 4)
	combinedCommitments2[0] = make([]*big.Int, 3)
	combinedCommitments2[0] = encryptedValues1
	// Not doing anything for 1
	combinedCommitments2[2] = make([]*big.Int, 3)
	combinedCommitments2[2] = encryptedValues3
	combinedCommitments2[3] = make([]*big.Int, 3)
	combinedCommitments2[3] = encryptedValues4
	encValuesFor2, err := CondenseCommitments(pubK2, combinedCommitments2, pubKs)
	if err != nil {
		t.Fatal("Error when condensing commitments for 2")
	}
	if len(encValuesFor2) != len(trueEncValuesFor2) {
		t.Fatal("encValuesFor2 has incorrect length")
	}
	for j := 0; j < len(encValuesFor2); j++ {
		trueEncVal := trueEncValuesFor2[j]
		encVal := encValuesFor2[j]
		if trueEncVal.Cmp(encVal) != 0 {
			t.Fatal("Error in condensing encrypted values 2")
		}
	}

	combinedCommitments3 := make([][]*big.Int, 4)
	combinedCommitments3[0] = make([]*big.Int, 3)
	combinedCommitments3[0] = encryptedValues1
	combinedCommitments3[1] = make([]*big.Int, 3)
	combinedCommitments3[1] = encryptedValues2
	// Not doing anything for 2
	combinedCommitments3[3] = make([]*big.Int, 3)
	combinedCommitments3[3] = encryptedValues4
	encValuesFor3, err := CondenseCommitments(pubK3, combinedCommitments3, pubKs)
	if err != nil {
		t.Fatal("Error when condensing commitments for 3")
	}
	if len(encValuesFor3) != len(trueEncValuesFor3) {
		t.Fatal("encValuesFor3 has incorrect length")
	}
	for j := 0; j < len(encValuesFor3); j++ {
		trueEncVal := trueEncValuesFor3[j]
		encVal := encValuesFor3[j]
		if trueEncVal.Cmp(encVal) != 0 {
			t.Fatal("Error in condensing encrypted values 3")
		}
	}

	combinedCommitments4 := make([][]*big.Int, 4)
	combinedCommitments4[0] = make([]*big.Int, 3)
	combinedCommitments4[0] = encryptedValues1
	combinedCommitments4[1] = make([]*big.Int, 3)
	combinedCommitments4[1] = encryptedValues2
	combinedCommitments4[2] = make([]*big.Int, 3)
	combinedCommitments4[2] = encryptedValues3
	// Not doing anything for 3
	encValuesFor4, err := CondenseCommitments(pubK4, combinedCommitments4, pubKs)
	if err != nil {
		t.Fatal("Error when condensing commitments for 4")
	}
	if len(encValuesFor4) != len(trueEncValuesFor4) {
		t.Fatal("encValuesFor3 has incorrect length")
	}
	for j := 0; j < len(encValuesFor4); j++ {
		trueEncVal := trueEncValuesFor4[j]
		encVal := encValuesFor4[j]
		if trueEncVal.Cmp(encVal) != 0 {
			t.Fatal("Error in condensing encrypted values 4")
		}
	}
}

func TestCondenseCommitmentsFail(t *testing.T) {
	privK1 := big.NewInt(10)
	pubK1 := new(G1).ScalarBaseMult(privK1)
	privK2 := big.NewInt(11)
	pubK2 := new(G1).ScalarBaseMult(privK2)
	privK3 := big.NewInt(12)
	pubK3 := new(G1).ScalarBaseMult(privK3)
	privK4 := big.NewInt(13)
	pubK4 := new(G1).ScalarBaseMult(privK4)
	pubKs := []*G1{pubK1, pubK2, pubK3, pubK4}
	pubKsBad := []*G1{pubK2, pubK3, pubK4}

	combinedCommitmentsGood := make([][]*big.Int, 4)
	combinedCommitmentsGood[0] = make([]*big.Int, 3)
	combinedCommitmentsGood[1] = make([]*big.Int, 3)
	combinedCommitmentsGood[2] = make([]*big.Int, 3)

	combinedCommitmentsBad1 := make([][]*big.Int, 5)
	combinedCommitmentsBad1[0] = make([]*big.Int, 3)
	combinedCommitmentsBad1[1] = make([]*big.Int, 3)
	combinedCommitmentsBad1[2] = make([]*big.Int, 3)

	combinedCommitmentsBad2 := make([][]*big.Int, 4)
	combinedCommitmentsBad2[0] = make([]*big.Int, 2)
	combinedCommitmentsBad2[1] = make([]*big.Int, 3)
	combinedCommitmentsBad2[2] = make([]*big.Int, 3)

	_, err := CondenseCommitments(pubK1, combinedCommitmentsGood, pubKsBad)
	if err != ErrMissingIndex {
		t.Fatal("Should have raised error for missing index")
	}
	_, err = CondenseCommitments(pubK1, combinedCommitmentsBad1, pubKs)
	if err != ErrArrayMismatch {
		t.Fatal("Should have raised error for bad commitments length 1")
	}
	_, err = CondenseCommitments(pubK1, combinedCommitmentsBad2, pubKs)
	if err != ErrArrayMismatch {
		t.Fatal("Should have raised error for bad commitments length 2")
	}
}

func TestGenerateDecryptedShares(t *testing.T) {
	privK1 := big.NewInt(10)
	pubK1 := new(G1).ScalarBaseMult(privK1)
	privCoefs1 := []*big.Int{big.NewInt(100), big.NewInt(1), big.NewInt(2)}

	privK2 := big.NewInt(11)
	pubK2 := new(G1).ScalarBaseMult(privK2)
	privCoefs2 := []*big.Int{big.NewInt(101), big.NewInt(1), big.NewInt(2)}

	privK3 := big.NewInt(12)
	pubK3 := new(G1).ScalarBaseMult(privK3)
	privCoefs3 := []*big.Int{big.NewInt(102), big.NewInt(1), big.NewInt(2)}

	privK4 := big.NewInt(13)
	pubK4 := new(G1).ScalarBaseMult(privK4)
	privCoefs4 := []*big.Int{big.NewInt(103), big.NewInt(1), big.NewInt(2)}

	pubKs := []*G1{pubK1, pubK2, pubK3, pubK4}

	secretValues1, err := GenerateSecretShares(pubK1, privCoefs1, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secret shares 1")
	}
	share1to2 := secretValues1[0]
	share1to3 := secretValues1[1]
	share1to4 := secretValues1[2]
	encryptedValues1, err := GenerateEncryptedShares(secretValues1, privK1, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating commitments 1")
	}

	secretValues2, err := GenerateSecretShares(pubK2, privCoefs2, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secret shares 2")
	}
	share2to1 := secretValues2[0]
	share2to3 := secretValues2[1]
	share2to4 := secretValues2[2]
	encryptedValues2, err := GenerateEncryptedShares(secretValues2, privK2, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating commitments 2")
	}

	secretValues3, err := GenerateSecretShares(pubK3, privCoefs3, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secret shares 3")
	}
	share3to1 := secretValues3[0]
	share3to2 := secretValues3[1]
	share3to4 := secretValues3[2]
	encryptedValues3, err := GenerateEncryptedShares(secretValues3, privK3, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating commitments 3")
	}

	secretValues4, err := GenerateSecretShares(pubK4, privCoefs4, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating secret shares 4")
	}
	share4to1 := secretValues4[0]
	share4to2 := secretValues4[1]
	share4to3 := secretValues4[2]
	encryptedValues4, err := GenerateEncryptedShares(secretValues4, privK4, pubKs)
	if err != nil {
		t.Fatal("Error occurred when generating commitments 4")
	}

	trueSecretValuesFor1 := []*big.Int{share2to1, share3to1, share4to1}
	trueSecretValuesFor2 := []*big.Int{share1to2, share3to2, share4to2}
	trueSecretValuesFor3 := []*big.Int{share1to3, share2to3, share4to3}
	trueSecretValuesFor4 := []*big.Int{share1to4, share2to4, share3to4}

	combinedCommitments1 := make([][]*big.Int, 4)
	// Not doing anything for 0
	combinedCommitments1[1] = make([]*big.Int, 3)
	combinedCommitments1[1] = encryptedValues2
	combinedCommitments1[2] = make([]*big.Int, 3)
	combinedCommitments1[2] = encryptedValues3
	combinedCommitments1[3] = make([]*big.Int, 3)
	combinedCommitments1[3] = encryptedValues4
	encValuesFor1, err := CondenseCommitments(pubK1, combinedCommitments1, pubKs)
	if err != nil {
		t.Fatal("Error when condensing commitments for 1")
	}
	secretValuesFor1, err := GenerateDecryptedShares(privK1, encValuesFor1, pubKs)
	if err != nil {
		t.Fatal("Error when generating decrypted shares for 1")
	}
	if len(secretValuesFor1) != len(trueSecretValuesFor1) {
		t.Fatal("secretValuesFor1 has incorrect length")
	}
	for j := 0; j < len(secretValuesFor1); j++ {
		secretValue := secretValuesFor1[j]
		trueSecretValue := trueSecretValuesFor1[j]
		if secretValue.Cmp(trueSecretValue) != 0 {
			t.Fatal("Error when computing secret values for 1")
		}
	}

	combinedCommitments2 := make([][]*big.Int, 4)
	combinedCommitments2[0] = make([]*big.Int, 3)
	combinedCommitments2[0] = encryptedValues1
	// Not doing anything for 1
	combinedCommitments2[2] = make([]*big.Int, 3)
	combinedCommitments2[2] = encryptedValues3
	combinedCommitments2[3] = make([]*big.Int, 3)
	combinedCommitments2[3] = encryptedValues4
	encValuesFor2, err := CondenseCommitments(pubK2, combinedCommitments2, pubKs)
	if err != nil {
		t.Fatal("Error when condensing commitments for 2")
	}
	secretValuesFor2, err := GenerateDecryptedShares(privK2, encValuesFor2, pubKs)
	if err != nil {
		t.Fatal("Error when generating decrypted shares for 2")
	}
	if len(secretValuesFor2) != len(trueSecretValuesFor2) {
		t.Fatal("secretValuesFor2 has incorrect length")
	}
	for j := 0; j < len(secretValuesFor2); j++ {
		secretValue := secretValuesFor2[j]
		trueSecretValue := trueSecretValuesFor2[j]
		if secretValue.Cmp(trueSecretValue) != 0 {
			t.Fatal("Error when computing secret values for 2")
		}
	}

	combinedCommitments3 := make([][]*big.Int, 4)
	combinedCommitments3[0] = make([]*big.Int, 3)
	combinedCommitments3[0] = encryptedValues1
	combinedCommitments3[1] = make([]*big.Int, 3)
	combinedCommitments3[1] = encryptedValues2
	// Not doing anything for 2
	combinedCommitments3[3] = make([]*big.Int, 3)
	combinedCommitments3[3] = encryptedValues4
	encValuesFor3, err := CondenseCommitments(pubK3, combinedCommitments3, pubKs)
	if err != nil {
		t.Fatal("Error when condensing commitments for 3")
	}
	secretValuesFor3, err := GenerateDecryptedShares(privK3, encValuesFor3, pubKs)
	if err != nil {
		t.Fatal("Error when generating decrypted shares for 3")
	}
	if len(secretValuesFor3) != len(trueSecretValuesFor3) {
		t.Fatal("secretValuesFor3 has incorrect length")
	}
	for j := 0; j < len(secretValuesFor3); j++ {
		secretValue := secretValuesFor3[j]
		trueSecretValue := trueSecretValuesFor3[j]
		if secretValue.Cmp(trueSecretValue) != 0 {
			t.Fatal("Error when computing secret values for 3")
		}
	}

	combinedCommitments4 := make([][]*big.Int, 4)
	combinedCommitments4[0] = make([]*big.Int, 3)
	combinedCommitments4[0] = encryptedValues1
	combinedCommitments4[1] = make([]*big.Int, 3)
	combinedCommitments4[1] = encryptedValues2
	combinedCommitments4[2] = make([]*big.Int, 3)
	combinedCommitments4[2] = encryptedValues3
	// Not doing anything for 3
	encValuesFor4, err := CondenseCommitments(pubK4, combinedCommitments4, pubKs)
	if err != nil {
		t.Fatal("Error when condensing commitments for 4")
	}
	secretValuesFor4, err := GenerateDecryptedShares(privK4, encValuesFor4, pubKs)
	if err != nil {
		t.Fatal("Error when generating decrypted shares for 4")
	}
	if len(secretValuesFor4) != len(trueSecretValuesFor4) {
		t.Fatal("secretValuesFor4 has incorrect length")
	}
	for j := 0; j < len(secretValuesFor4); j++ {
		secretValue := secretValuesFor4[j]
		trueSecretValue := trueSecretValuesFor4[j]
		if secretValue.Cmp(trueSecretValue) != 0 {
			t.Fatal("Error when computing secret values for 4")
		}
	}
}

func TestGenerateDecryptedSharesFail(t *testing.T) {
	privK1 := big.NewInt(10)
	pubK1 := new(G1).ScalarBaseMult(privK1)

	privK2 := big.NewInt(11)
	pubK2 := new(G1).ScalarBaseMult(privK2)

	privK3 := big.NewInt(12)
	pubK3 := new(G1).ScalarBaseMult(privK3)

	privK4 := big.NewInt(13)
	pubK4 := new(G1).ScalarBaseMult(privK4)

	pubKs := []*G1{pubK1, pubK2, pubK3, pubK4}
	pubKsBad := []*G1{pubK2, pubK3, pubK4}

	encryptedValuesBad := make([]*big.Int, 5)
	encryptedValuesBad2 := make([]*big.Int, 2)

	_, err := GenerateDecryptedShares(privK1, encryptedValuesBad, pubKs)
	if err != ErrArrayMismatch {
		t.Fatal("Wrong error occurred in GenerateDecryptedShares; expected array mismatch")
	}

	_, err = GenerateDecryptedShares(privK1, encryptedValuesBad2, pubKsBad)
	if err != ErrMissingIndex {
		t.Fatal("Wrong error occurred in GenerateDecryptedShares; expected missing index")
	}
}

func TestGenerateGSKJ(t *testing.T) {
	secret1, _ := new(big.Int).SetString("3141592653589793238462643383279502884197169399375105820974944592307816406286", 10) // 76
	secret2, _ := new(big.Int).SetString("2718281828459045235360287471352662497757247093699959574966967627724076630353", 10) // 76
	secret3, _ := new(big.Int).SetString("1618033988749894848204586834365638117720309179805762862135448622705260462818", 10) // 76
	secret4, _ := new(big.Int).SetString("1414213562373095048801688724209698078569671875376948073176679737990732478462", 10) // 76

	big1 := big.NewInt(1)
	big2 := big.NewInt(2)

	privCoefs1 := []*big.Int{secret1, big1, big2}
	privCoefs2 := []*big.Int{secret2, big1, big2}
	privCoefs3 := []*big.Int{secret3, big1, big2}
	privCoefs4 := []*big.Int{secret4, big1, big2}

	share1to1 := PrivatePolyEval(privCoefs1, 1)
	share1to2 := PrivatePolyEval(privCoefs1, 2)
	share1to3 := PrivatePolyEval(privCoefs1, 3)
	share1to4 := PrivatePolyEval(privCoefs1, 4)
	share2to1 := PrivatePolyEval(privCoefs2, 1)
	share2to2 := PrivatePolyEval(privCoefs2, 2)
	share2to3 := PrivatePolyEval(privCoefs2, 3)
	share2to4 := PrivatePolyEval(privCoefs2, 4)
	share3to1 := PrivatePolyEval(privCoefs3, 1)
	share3to2 := PrivatePolyEval(privCoefs3, 2)
	share3to3 := PrivatePolyEval(privCoefs3, 3)
	share3to4 := PrivatePolyEval(privCoefs3, 4)
	share4to1 := PrivatePolyEval(privCoefs4, 1)
	share4to2 := PrivatePolyEval(privCoefs4, 2)
	share4to3 := PrivatePolyEval(privCoefs4, 3)
	share4to4 := PrivatePolyEval(privCoefs4, 4)

	trueGSK1 := big.NewInt(0)
	trueGSK1.Add(trueGSK1, share1to1)
	trueGSK1.Add(trueGSK1, share2to1)
	trueGSK1.Add(trueGSK1, share3to1)
	trueGSK1.Add(trueGSK1, share4to1)
	trueGSK1.Mod(trueGSK1, Order)
	listOfSS1 := []*big.Int{share1to1, share2to1, share3to1, share4to1}
	resSS1 := GenerateGroupSecretKeyPortion(listOfSS1)
	if resSS1.Cmp(trueGSK1) != 0 {
		t.Fatal("Error in computing gsk1")
	}

	trueGSK2 := big.NewInt(0)
	trueGSK2.Add(trueGSK2, share1to2)
	trueGSK2.Add(trueGSK2, share2to2)
	trueGSK2.Add(trueGSK2, share3to2)
	trueGSK2.Add(trueGSK2, share4to2)
	trueGSK2.Mod(trueGSK2, Order)
	listOfSS2 := []*big.Int{share1to2, share2to2, share3to2, share4to2}
	resSS2 := GenerateGroupSecretKeyPortion(listOfSS2)
	if resSS2.Cmp(trueGSK2) != 0 {
		t.Fatal("Error in computing gsk2")
	}

	trueGSK3 := big.NewInt(0)
	trueGSK3.Add(trueGSK3, share1to3)
	trueGSK3.Add(trueGSK3, share2to3)
	trueGSK3.Add(trueGSK3, share3to3)
	trueGSK3.Add(trueGSK3, share4to3)
	trueGSK3.Mod(trueGSK3, Order)
	listOfSS3 := []*big.Int{share1to3, share2to3, share3to3, share4to3}
	resSS3 := GenerateGroupSecretKeyPortion(listOfSS3)
	if resSS3.Cmp(trueGSK3) != 0 {
		t.Fatal("Error in computing gsk3")
	}

	trueGSK4 := big.NewInt(0)
	trueGSK4.Add(trueGSK4, share1to4)
	trueGSK4.Add(trueGSK4, share2to4)
	trueGSK4.Add(trueGSK4, share3to4)
	trueGSK4.Add(trueGSK4, share4to4)
	trueGSK4.Mod(trueGSK4, Order)
	listOfSS4 := []*big.Int{share1to4, share2to4, share3to4, share4to4}
	resSS4 := GenerateGroupSecretKeyPortion(listOfSS4)
	if resSS4.Cmp(trueGSK4) != 0 {
		t.Fatal("Error in computing gsk4")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("GenerateGSK changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("GenerateGSK changed twistGen")
	}
}
