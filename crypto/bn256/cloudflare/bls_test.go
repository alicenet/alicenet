package cloudflare

import (
	"math/big"
	"testing"
)

func BenchmarkSignatureGeneration(b *testing.B) {
	privK := big.NewInt(3141592653589793) // 16 digits of Pi
	s := "MadHive Rocks!"
	msg := []byte(s)
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, err := Sign(msg, privK, HashToG1)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSignatureVerification(b *testing.B) {
	h2 := new(G2).ScalarBaseMult(big.NewInt(1))
	privK := big.NewInt(3141592653589793) // 16 digits of Pi
	pubK := new(G2).ScalarMult(h2, privK)
	s := "MadHive Rocks!"
	msg := []byte(s)
	sig, err := Sign(msg, privK, HashToG1)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, err = Verify(msg, sig, pubK, HashToG1)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestSignAndVerify(t *testing.T) {
	h2 := new(G2).ScalarBaseMult(big.NewInt(1))

	privK := big.NewInt(3141592653589793) // 16 digits of Pi
	pubK := new(G2).ScalarMult(h2, privK)

	s := "MadHive Rocks!"
	msg := []byte(s)
	sigMH, err := Sign(msg, privK, HashToG1)
	if err != nil {
		t.Fatal("Error occurred when signing MH:", err)
	}
	val, err := Verify(msg, sigMH, pubK, HashToG1)
	if err != nil {
		t.Fatal("Error occurred when verifying MH:", err)
	}
	if !val {
		t.Fatal("Verify failed to validate valid Sign MH signature")
	}

	s = "Cryptography is great"
	msg = []byte(s)
	sigCIG, err := Sign(msg, privK, HashToG1)
	if err != nil {
		t.Fatal("Error occurred when signing CIG:", err)
	}
	val, err = Verify(msg, sigCIG, pubK, HashToG1)
	if err != nil {
		t.Fatal("Error occurred when verifying CIG:", err)
	}
	if !val {
		t.Fatal("Verify failed to validate valid Sign CIG signature")
	}

	val, _ = Verify(msg, sigMH, pubK, HashToG1)
	if val {
		t.Fatal("Verify is incorrect; computes invalid signature as valid")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("Sign or Verify changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("Sign or Verify changed twistGen")
	}
}

func TestLagrangeInterpolationG1(t *testing.T) {
	threshold := 2

	secret1, _ := new(big.Int).SetString("3141592653589793238462643383279502884197169399375105820974944592307816406286", 10) // 76
	secret2, _ := new(big.Int).SetString("2718281828459045235360287471352662497757247093699959574966967627724076630353", 10) // 76
	secret3, _ := new(big.Int).SetString("1618033988749894848204586834365638117720309179805762862135448622705260462818", 10) // 76
	secret4, _ := new(big.Int).SetString("1414213562373095048801688724209698078569671875376948073176679737990732478462", 10) // 76

	msk := big.NewInt(0)
	msk.Add(msk, secret1)
	msk.Add(msk, secret2)
	msk.Add(msk, secret3)
	msk.Add(msk, secret4)
	msk.Mod(msk, Order)
	mpk := &G2{}
	mpk.ScalarBaseMult(msk)

	// What we want to try to interpolate
	g1mpk := &G1{}
	g1mpk.ScalarBaseMult(msk)

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

	listOfSS1 := []*big.Int{share1to1, share2to1, share3to1, share4to1}
	gsk1 := GenerateGroupSecretKeyPortion(listOfSS1)
	g1gsk1 := &G1{}
	g1gsk1.ScalarBaseMult(gsk1)

	listOfSS2 := []*big.Int{share1to2, share2to2, share3to2, share4to2}
	gsk2 := GenerateGroupSecretKeyPortion(listOfSS2)
	g1gsk2 := &G1{}
	g1gsk2.ScalarBaseMult(gsk2)

	listOfSS3 := []*big.Int{share1to3, share2to3, share3to3, share4to3}
	gsk3 := GenerateGroupSecretKeyPortion(listOfSS3)
	g1gsk3 := &G1{}
	g1gsk3.ScalarBaseMult(gsk3)

	listOfSS4 := []*big.Int{share1to4, share2to4, share3to4, share4to4}
	gsk4 := GenerateGroupSecretKeyPortion(listOfSS4)
	g1gsk4 := &G1{}
	g1gsk4.ScalarBaseMult(gsk4)

	listOfStuffBad := []*G1{g1gsk1, g1gsk2, g1gsk3, g1gsk4}
	badIndices := []int{1, 2, 3}
	_, err := LagrangeInterpolationG1(listOfStuffBad, badIndices, threshold)
	if err == nil {
		t.Fatal("Failed to raise flag for strings not present")
	}

	listOfG1s1 := []*G1{g1gsk1, g1gsk2, g1gsk3}
	indices1 := []int{1, 2, 3}
	testG1MPK1, err := LagrangeInterpolationG1(listOfG1s1, indices1, threshold)
	if err != nil {
		t.Fatal(err)
	}
	if !testG1MPK1.IsEqual(g1mpk) {
		t.Fatal("Failed to construct g1mpk from list1")
	}

	listOfG1s2 := []*G1{g1gsk1, g1gsk2, g1gsk3, g1gsk4}
	indices2 := []int{1, 2, 3, 4}
	testG1MPK2, err := LagrangeInterpolationG1(listOfG1s2, indices2, threshold)
	if err != nil {
		t.Fatal(err)
	}
	if !testG1MPK2.IsEqual(g1mpk) {
		t.Fatal("Failed to construct g1mpk from list2")
	}

	listOfG1s3 := []*G1{g1gsk3, g1gsk2, g1gsk1}
	indices3 := []int{3, 2, 1}
	testG1MPK3, err := LagrangeInterpolationG1(listOfG1s3, indices3, threshold)
	if err != nil {
		t.Fatal(err)
	}
	if !testG1MPK3.IsEqual(g1mpk) {
		t.Fatal("Failed to construct g1mpk from list3")
	}

	listOfG1s4 := []*G1{g1gsk2, g1gsk1, g1gsk4}
	indices4 := []int{2, 1, 4}
	testG1MPK4, err := LagrangeInterpolationG1(listOfG1s4, indices4, threshold)
	if err != nil {
		t.Fatal(err)
	}
	if !testG1MPK4.IsEqual(g1mpk) {
		t.Fatal("Failed to construct g1mpk from list4")
	}

	listOfG1s5 := []*G1{g1gsk2, g1gsk2, g1gsk4}
	indices5 := []int{2, 2, 4}
	testG1MPK5, _ := LagrangeInterpolationG1(listOfG1s5, indices5, threshold)
	if testG1MPK5.IsEqual(g1mpk) {
		t.Fatal("Constructed g1mpk from list5 when should have failed")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("LIG1 changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("LIG1 changed twistGen")
	}
}

func BenchmarkAggregateSignatures(b *testing.B) {
	threshold := 8

	gsk1, _ := new(big.Int).SetString("12345678910111213141516171819201", 10)
	gsk2, _ := new(big.Int).SetString("12345678910111213141516171819202", 10)
	gsk3, _ := new(big.Int).SetString("12345678910111213141516171819203", 10)
	gsk4, _ := new(big.Int).SetString("12345678910111213141516171819204", 10)
	gsk5, _ := new(big.Int).SetString("12345678910111213141516171819205", 10)
	gsk6, _ := new(big.Int).SetString("12345678910111213141516171819206", 10)
	gsk7, _ := new(big.Int).SetString("12345678910111213141516171819207", 10)
	gsk8, _ := new(big.Int).SetString("12345678910111213141516171819208", 10)
	gsk9, _ := new(big.Int).SetString("12345678910111213141516171819209", 10)

	s := "MadHive Rocks!"
	msg := []byte(s)

	sig1, err := Sign(msg, gsk1, HashToG1)
	if err != nil {
		b.Fatal(err)
	}
	sig2, err := Sign(msg, gsk2, HashToG1)
	if err != nil {
		b.Fatal(err)
	}
	sig3, err := Sign(msg, gsk3, HashToG1)
	if err != nil {
		b.Fatal(err)
	}
	sig4, err := Sign(msg, gsk4, HashToG1)
	if err != nil {
		b.Fatal(err)
	}
	sig5, err := Sign(msg, gsk5, HashToG1)
	if err != nil {
		b.Fatal(err)
	}
	sig6, err := Sign(msg, gsk6, HashToG1)
	if err != nil {
		b.Fatal(err)
	}
	sig7, err := Sign(msg, gsk7, HashToG1)
	if err != nil {
		b.Fatal(err)
	}
	sig8, err := Sign(msg, gsk8, HashToG1)
	if err != nil {
		b.Fatal(err)
	}
	sig9, err := Sign(msg, gsk9, HashToG1)
	if err != nil {
		b.Fatal(err)
	}

	sigs := []*G1{sig1, sig2, sig3, sig4, sig5, sig6, sig7, sig8, sig9}

	indices := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, err = AggregateSignatures(sigs, indices, threshold)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestAggregateSignatures(t *testing.T) {
	threshold := 2

	secret1, _ := new(big.Int).SetString("3141592653589793238462643383279502884197169399375105820974944592307816406286", 10) // 76
	secret2, _ := new(big.Int).SetString("2718281828459045235360287471352662497757247093699959574966967627724076630353", 10) // 76
	secret3, _ := new(big.Int).SetString("1618033988749894848204586834365638117720309179805762862135448622705260462818", 10) // 76
	secret4, _ := new(big.Int).SetString("1414213562373095048801688724209698078569671875376948073176679737990732478462", 10) // 76

	mskTrue, _ := new(big.Int).SetString("8892122033171828370829206413207501578244397548257776331254040580727885977919", 10)

	msk := big.NewInt(0)
	msk.Add(msk, secret1)
	msk.Add(msk, secret2)
	msk.Add(msk, secret3)
	msk.Add(msk, secret4)
	msk.Mod(msk, Order)
	mpk := &G2{}
	mpk.ScalarBaseMult(msk)

	if msk.Cmp(mskTrue) != 0 {
		t.Fatal("Error in initial definition of msk")
	}

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

	listOfSS1 := []*big.Int{share1to1, share2to1, share3to1, share4to1}
	gsk1 := GenerateGroupSecretKeyPortion(listOfSS1)
	gpk1 := &G2{}
	gpk1.ScalarBaseMult(gsk1)

	listOfSS2 := []*big.Int{share1to2, share2to2, share3to2, share4to2}
	gsk2 := GenerateGroupSecretKeyPortion(listOfSS2)
	gpk2 := &G2{}
	gpk2.ScalarBaseMult(gsk2)

	listOfSS3 := []*big.Int{share1to3, share2to3, share3to3, share4to3}
	gsk3 := GenerateGroupSecretKeyPortion(listOfSS3)
	gpk3 := &G2{}
	gpk3.ScalarBaseMult(gsk3)

	listOfSS4 := []*big.Int{share1to4, share2to4, share3to4, share4to4}
	gsk4 := GenerateGroupSecretKeyPortion(listOfSS4)
	gpk4 := &G2{}
	gpk4.ScalarBaseMult(gsk4)

	s := "MadHive Rocks!"
	msg := []byte(s)
	trueGroupSigMH, err := Sign(msg, msk, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with msk:", err)
	}

	trueSigMH1, err := Sign(msg, secret1, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with secret1:", err)
	}

	trueSigMH2, err := Sign(msg, secret2, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with secret2:", err)
	}

	trueSigMH3, err := Sign(msg, secret3, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with secret3:", err)
	}

	trueSigMH4, err := Sign(msg, secret4, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with secret4:", err)
	}

	prodTrueSigsMH := &G1{}
	prodTrueSigsMH.Add(trueSigMH1, trueSigMH2)
	prodTrueSigsMH.Add(prodTrueSigsMH, trueSigMH3)
	prodTrueSigsMH.Add(prodTrueSigsMH, trueSigMH4)
	if !trueGroupSigMH.IsEqual(prodTrueSigsMH) {
		t.Fatal("Error occurred when computing trueGroupSigMH:", err)
	}

	partialSigMH1, err := Sign(msg, gsk1, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with gsk1:", err)
	}
	val, err := Verify(msg, partialSigMH1, gpk1, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying MH with gpk1:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH gpk1 signature")
	}

	partialSigMH2, err := Sign(msg, gsk2, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with gsk2")
	}
	val, err = Verify(msg, partialSigMH2, gpk2, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying MH with gpk2:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH gpk2 signature")
	}

	partialSigMH3, err := Sign(msg, gsk3, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with gsk3")
	}
	val, err = Verify(msg, partialSigMH3, gpk3, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying MH with gpk3:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH gpk3 signature")
	}

	partialSigMH4, err := Sign(msg, gsk4, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with gsk4")
	}
	val, err = Verify(msg, partialSigMH4, gpk4, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying MH with gpk4:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH gpk4 signature")
	}

	list1Sigs := []*G1{partialSigMH1, partialSigMH2, partialSigMH3}
	indices1 := []int{1, 2, 3}
	grpSig, err := AggregateSignatures(list1Sigs, indices1, threshold)
	if err != nil {
		t.Fatal("Error occurred in AggregateSignatures for MH and list1:", err)
	}
	val, err = Verify(msg, grpSig, mpk, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in group signature MH 1 verification:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH 1 group signature")
	}

	list2Sigs := []*G1{partialSigMH1, partialSigMH2, partialSigMH3, partialSigMH4}
	indices2 := []int{1, 2, 3, 4}
	grpSig2, err := AggregateSignatures(list2Sigs, indices2, threshold)
	if err != nil {
		t.Fatal("Error occurred in AggregateSignatures for MH and list2:", err)
	}
	val, err = Verify(msg, grpSig2, mpk, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in group signature MH 2 verification:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH 2 group signature")
	}

	list3Sigs := []*G1{partialSigMH1, partialSigMH2}
	indices3 := []int{1, 2, 3, 4}
	_, err = AggregateSignatures(list3Sigs, indices3, threshold)
	if err == nil {
		t.Fatal("Failed to recognize array size MH difference")
	}

	list4Sigs := []*G1{partialSigMH1, partialSigMH2}
	indices4 := []int{1, 2}
	_, err = AggregateSignatures(list4Sigs, indices4, threshold)
	if err == nil {
		t.Fatal("Failed to recognize threshold failure MH")
	}

	s = "Cryptography is great"
	msgCIG := []byte(s)
	trueGroupSigCIG, err := Sign(msgCIG, msk, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing CIG with msk:", err)
	}

	trueSigCIG1, err := Sign(msgCIG, secret1, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing CIG with secret1:", err)
	}

	trueSigCIG2, err := Sign(msgCIG, secret2, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing CIG with secret2:", err)
	}

	trueSigCIG3, err := Sign(msgCIG, secret3, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing CIG with secret3:", err)
	}

	trueSigCIG4, err := Sign(msgCIG, secret4, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing CIG with secret4:", err)
	}

	prodTrueSigsCIG := &G1{}
	prodTrueSigsCIG.Add(trueSigCIG1, trueSigCIG2)
	prodTrueSigsCIG.Add(prodTrueSigsCIG, trueSigCIG3)
	prodTrueSigsCIG.Add(prodTrueSigsCIG, trueSigCIG4)
	if !trueGroupSigCIG.IsEqual(prodTrueSigsCIG) {
		t.Fatal("Error occurred when computing trueGroupSigCIG")
	}

	partialSigCIG1, err := Sign(msgCIG, gsk1, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing CIG with gsk1:", err)
	}
	val, err = Verify(msgCIG, partialSigCIG1, gpk1, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying CIG with gpk1:", err)
	}
	if !val {
		t.Fatal("Failed to verify CIG gpk1 signature")
	}

	partialSigCIG2, err := Sign(msgCIG, gsk2, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing CIG with gsk2:", err)
	}
	val, err = Verify(msgCIG, partialSigCIG2, gpk2, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying CIG with gpk2:", err)
	}
	if !val {
		t.Fatal("Failed to verify CIG gpk2 signature")
	}

	partialSigCIG3, err := Sign(msgCIG, gsk3, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing CIG with gsk3")
	}
	val, err = Verify(msgCIG, partialSigCIG3, gpk3, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying CIG with gpk3", err)
	}
	if !val {
		t.Fatal("Failed to verify CIG gpk3 signature")
	}

	partialSigCIG4, err := Sign(msgCIG, gsk4, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing CIG with gsk4:", err)
	}
	val, err = Verify(msgCIG, partialSigCIG4, gpk4, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying CIG with gpk4:", err)
	}
	if !val {
		t.Fatal("Failed to verify CIG gpk4 signature")
	}

	list1Sigs = []*G1{partialSigCIG1, partialSigCIG2, partialSigCIG3}
	indices1 = []int{1, 2, 3}
	grpSig, err = AggregateSignatures(list1Sigs, indices1, threshold)
	if err != nil {
		t.Fatal("Error occurred in AggregateSignatures for CIG and list1:", err)
	}
	val, err = Verify(msgCIG, grpSig, mpk, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in group signature CIG 1 verification:", err)
	}
	if !val {
		t.Fatal("Failed to verify CIG 1 group signature")
	}

	list2Sigs = []*G1{partialSigCIG1, partialSigCIG2, partialSigCIG3, partialSigCIG4}
	indices2 = []int{1, 2, 3, 4}
	grpSig2, err = AggregateSignatures(list2Sigs, indices2, threshold)
	if err != nil {
		t.Fatal("Error occurred in AggregateSignatures for CIG and list2:", err)
	}
	val, err = Verify(msgCIG, grpSig2, mpk, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in group signature CIG 2 verification:", err)
	}
	if !val {
		t.Fatal("Failed to verify CIG 2 group signature")
	}

	list3Sigs = []*G1{partialSigCIG1, partialSigCIG2}
	indices3 = []int{1, 2, 3, 4}
	_, err = AggregateSignatures(list3Sigs, indices3, threshold)
	if err == nil {
		t.Fatal("Failed to recognize array size difference CIG")
	}

	list4Sigs = []*G1{partialSigCIG1, partialSigCIG2}
	indices4 = []int{1, 2}
	_, err = AggregateSignatures(list4Sigs, indices4, threshold)
	if err == nil {
		t.Fatal("Failed to recognize threshold failure CIG")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("AggregateSignatures changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("AggregateSignatures changed twistGen")
	}
}

func TestMarshalBLS(t *testing.T) {
	blsMarshalSingle(t)
	blsMarshalAgg(t)
}

func blsMarshalSingle(t *testing.T) {
	t.Helper()
	privK := big.NewInt(1414213562373095) // 16 digits of sqrt(2)
	pubK := new(G2).ScalarBaseMult(privK)

	s := "MadHive Rocks!"
	msg := []byte(s)
	sig, err0 := Sign(msg, privK, HashToG1)
	if err0 != nil {
		t.Fatal("Error occurred in signing MH with privK:", err0)
	}

	sigCopy := &G1{}
	sigCopy.Set(sig)
	pubKCopy := &G2{}
	pubKCopy.Set(pubK)

	totalMarsh, err1 := MarshalSignature(sig, pubK)
	if err1 != nil {
		t.Fatal("Error in MarshalSignature", err1)
	}
	pubKUnmarsh, sigUnmarsh, err2 := UnmarshalSignature(totalMarsh)
	if err2 != nil {
		t.Fatal("Error occurred in UnmarshalSignature for MH", err2)
	}
	if !sigUnmarsh.IsEqual(sig) {
		t.Fatal("Failed to reproduce sig")
	}
	if !pubKUnmarsh.IsEqual(pubK) {
		t.Fatal("Failed to reproduce pubK")
	}

	badMarsh1 := make([]byte, 5*numBytes)
	_, _, err3 := UnmarshalSignature(badMarsh1)
	if err3 == nil {
		t.Fatal("Failed to raise error in UnmarshalSignature for invalid size of byte slice")
	}

	// Invalid byte slice for GFp unmarshal
	breakBytes := make([]byte, numBytes)
	breakBytes[0] = byte(255) // >= 49 will work

	// Raise invalid unmarshal for sig
	badMarsh2 := make([]byte, 6*numBytes)
	for k := 0; k < numBytes; k++ {
		badMarsh2[k] = breakBytes[k]
	}
	_, _, err4 := UnmarshalSignature(badMarsh2)
	if err4 == nil {
		t.Fatal("Failed to raise error in UnmarshalSignature for invalid sig")
	}

	// Raise invalid unmarshal for pubK
	badMarsh3 := make([]byte, 6*numBytes)
	for k := 0; k < numBytes; k++ {
		badMarsh3[4*numBytes+k] = breakBytes[k]
	}
	_, _, err5 := UnmarshalSignature(badMarsh3)
	if err5 == nil {
		t.Fatal("Failed to raise error in UnmarshalSignature for invalid pubK")
	}

	badMarsh4 := make([]byte, 5*numBytes)
	_, _, err6 := SplitPubkeySig(badMarsh4)
	if err6 == nil {
		t.Fatal("Failed to raise error in SplitPubkeySig for invalid marshal")
	}

	_, _, err7 := SplitPubkeySig(totalMarsh)
	if err7 != nil {
		t.Fatal("Error occurred in SplitPubkeySig:", err7)
	}

	badMarsh5 := make([]byte, 5*numBytes)
	_, err8 := PubkeyFromSig(badMarsh5)
	if err8 == nil {
		t.Fatal("Failed to raise error in PubkeyFromSig for invalid marshalled sig")
	}

	_, err9 := PubkeyFromSig(totalMarsh)
	if err9 != nil {
		t.Fatal("Error occurred in PubkeyFromSig:", err9)
	}
}

func blsMarshalAgg(t *testing.T) {
	t.Helper()
	threshold := 2

	secret1, _ := new(big.Int).SetString("3141592653589793238462643383279502884197169399375105820974944592307816406286", 10) // 76
	secret2, _ := new(big.Int).SetString("2718281828459045235360287471352662497757247093699959574966967627724076630353", 10) // 76
	secret3, _ := new(big.Int).SetString("1618033988749894848204586834365638117720309179805762862135448622705260462818", 10) // 76
	secret4, _ := new(big.Int).SetString("1414213562373095048801688724209698078569671875376948073176679737990732478462", 10) // 76

	mskTrue, _ := new(big.Int).SetString("8892122033171828370829206413207501578244397548257776331254040580727885977919", 10)

	msk := big.NewInt(0)
	msk.Add(msk, secret1)
	msk.Add(msk, secret2)
	msk.Add(msk, secret3)
	msk.Add(msk, secret4)
	msk.Mod(msk, Order)
	mpk := &G2{}
	mpk.ScalarBaseMult(msk)

	if msk.Cmp(mskTrue) != 0 {
		t.Fatal("Error in initial definition of msk")
	}

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

	listOfSS1 := []*big.Int{share1to1, share2to1, share3to1, share4to1}
	gsk1 := GenerateGroupSecretKeyPortion(listOfSS1)
	gpk1 := &G2{}
	gpk1.ScalarBaseMult(gsk1)

	listOfSS2 := []*big.Int{share1to2, share2to2, share3to2, share4to2}
	gsk2 := GenerateGroupSecretKeyPortion(listOfSS2)
	gpk2 := &G2{}
	gpk2.ScalarBaseMult(gsk2)

	listOfSS3 := []*big.Int{share1to3, share2to3, share3to3, share4to3}
	gsk3 := GenerateGroupSecretKeyPortion(listOfSS3)
	gpk3 := &G2{}
	gpk3.ScalarBaseMult(gsk3)

	listOfSS4 := []*big.Int{share1to4, share2to4, share3to4, share4to4}
	gsk4 := GenerateGroupSecretKeyPortion(listOfSS4)
	gpk4 := &G2{}
	gpk4.ScalarBaseMult(gsk4)

	s := "MadHive Rocks!"
	msg := []byte(s)
	trueGroupSigMH, err := Sign(msg, msk, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with msk:", err)
	}

	trueSigMH1, err := Sign(msg, secret1, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with secret1:", err)
	}

	trueSigMH2, err := Sign(msg, secret2, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with secret2:", err)
	}

	trueSigMH3, err := Sign(msg, secret3, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with secret3:", err)
	}

	trueSigMH4, err := Sign(msg, secret4, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with secret4:", err)
	}

	prodTrueSigsMH := &G1{}
	prodTrueSigsMH.Add(trueSigMH1, trueSigMH2)
	prodTrueSigsMH.Add(prodTrueSigsMH, trueSigMH3)
	prodTrueSigsMH.Add(prodTrueSigsMH, trueSigMH4)
	if !trueGroupSigMH.IsEqual(prodTrueSigsMH) {
		t.Fatal("Error occurred when computing trueGroupSigMH")
	}

	partialSigMH1, err := Sign(msg, gsk1, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with gsk1:", err)
	}
	val, err := Verify(msg, partialSigMH1, gpk1, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying MH with gpk1:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH gpk1 signature")
	}
	marshSig1, err := MarshalSignature(partialSigMH1, gpk1)
	if err != nil {
		t.Fatal("Error occurred in MarshalSig for pubK1:", err)
	}

	partialSigMH2, err := Sign(msg, gsk2, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with gsk2:", err)
	}
	val, err = Verify(msg, partialSigMH2, gpk2, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying MH with gpk2:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH gpk2 signature")
	}
	marshSig2, err := MarshalSignature(partialSigMH2, gpk2)
	if err != nil {
		t.Fatal("Error occurred in MarshalSig for pubK2:", err)
	}

	partialSigMH3, err := Sign(msg, gsk3, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with gsk3:", err)
	}
	val, err = Verify(msg, partialSigMH3, gpk3, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying MH with gpk3:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH gpk3 signature")
	}
	marshSig3, err := MarshalSignature(partialSigMH3, gpk3)
	if err != nil {
		t.Fatal("Error occurred in MarshalSig for pubK3:", err)
	}

	partialSigMH4, err := Sign(msg, gsk4, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with gsk4:", err)
	}
	val, err = Verify(msg, partialSigMH4, gpk4, HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying MH with gpk4:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH gpk4 signature")
	}
	marshSig4, err := MarshalSignature(partialSigMH4, gpk4)
	if err != nil {
		t.Fatal("Error occurred in MarshalSig for pubK4:", err)
	}

	listOfPubKsMarsh := [][]byte{gpk1.Marshal(), gpk2.Marshal(), gpk3.Marshal(), gpk4.Marshal()}
	marshalledSigs1 := [][]byte{marshSig1, marshSig2, marshSig3, marshSig4}
	grpsig1, err := AggregateMarshalledSignatures(marshalledSigs1, listOfPubKsMarsh, threshold)
	if err != nil {
		t.Fatal("Error in AggMarshSigs:", err)
	}
	if !grpsig1.IsEqual(prodTrueSigsMH) {
		t.Fatal("Error occurred in AggMarshSigs for MH")
	}

	badMarshalledSigs1 := [][]byte{marshSig1, marshSig2}
	_, err = AggregateMarshalledSignatures(badMarshalledSigs1, listOfPubKsMarsh, threshold)
	if err == nil {
		t.Fatal("Failed to raise threshold error")
	}

	badByteSlice := make([]byte, 5*numBytes)
	badMarshalledSigs2 := [][]byte{marshSig1, marshSig2, badByteSlice}
	_, err = AggregateMarshalledSignatures(badMarshalledSigs2, listOfPubKsMarsh, threshold)
	if err == nil {
		t.Fatal("Failed to raise error for bad marshalled sig")
	}

	badByteSlice2 := make([]byte, 3*numBytes)
	listOfPubKsMarshBad1 := [][]byte{gpk1.Marshal(), gpk2.Marshal(), gpk3.Marshal(), badByteSlice2}
	badMarshalledSigs3 := [][]byte{marshSig1, marshSig2, marshSig3, marshSig4}
	_, err = AggregateMarshalledSignatures(badMarshalledSigs3, listOfPubKsMarshBad1, threshold)
	if err == nil {
		t.Fatal("Failed to raise error for bad marshalled sig in ordered pubKs")
	}

	badByteSlice3 := make([]byte, 4*numBytes)
	listOfPubKsMarshBad2 := [][]byte{gpk1.Marshal(), gpk2.Marshal(), gpk3.Marshal(), badByteSlice3}
	badMarshalledSigs4 := [][]byte{marshSig1, marshSig2, marshSig3, marshSig4}
	_, err = AggregateMarshalledSignatures(badMarshalledSigs4, listOfPubKsMarshBad2, threshold)
	if err == nil {
		t.Fatal("Failed to raise error for being unable to find index")
	}

	cG := returnCurveGen()
	if !cG.IsEqual(curveGen) {
		t.Fatal("AggregateMarshalledSignatures changed curveGen")
	}
	tG := returnTwistGen()
	if !tG.IsEqual(twistGen) {
		t.Fatal("AggregateMarshalledSignatures changed twistGen")
	}
}
