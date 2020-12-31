package bn256

import (
	"bytes"
	"crypto/ecdsa"
	"log"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/MadBase/MadNet/crypto/bn256/solidity"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

func SolidityContractSetup(t *testing.T, users int) (*solidity.Solidity, *backends.SimulatedBackend, []*ecdsa.PrivateKey, []*bind.TransactOpts) {
	if users < 1 {
		t.Fatal("Must have at least 1 user for contract setup")
	}
	// Generate a new random account and a funded simulator
	gasLimit := uint64(1000000000000000)
	genAlloc := make(core.GenesisAlloc)
	keyArray := make([]*ecdsa.PrivateKey, users)
	authArray := make([]*bind.TransactOpts, users)
	for k := 0; k < users; k++ {
		key, _ := crypto.GenerateKey()
		auth := bind.NewKeyedTransactor(key)
		genAlloc[auth.From] = core.GenesisAccount{Balance: big.NewInt(9223372036854775807)}
		keyArray[k] = key
		authArray[k] = auth
	}
	sim := backends.NewSimulatedBackend(genAlloc, gasLimit) // Deploy a token contract on the simulated blockchain
	_, _, c, err := solidity.DeploySolidity(authArray[0], sim)
	sim.Commit()
	if err != nil {
		log.Fatalf("Failed to deploy new token contract: %v", err)
	}
	return c, sim, keyArray, authArray
}

func TestHashSignVerify(t *testing.T) {
	users := 1
	c, sim, _, _ := SolidityContractSetup(t, users)
	defer sim.Close()

	msg := []byte{0x00, 0x01, 0x02, 0x03}
	trueHashG1, err := cloudflare.HashToG1(msg)
	if err != nil {
		t.Fatal("Error in cloudflare.HashToG1:", err)
	}
	trueHashMarsh := trueHashG1.Marshal()
	hashG1, err := c.HashToG1(&bind.CallOpts{}, msg)
	if err != nil {
		t.Fatal("Error in HashToG1:", err)
	}
	hashMarsh, err := MarshalG1Big(hashG1)
	if err != nil {
		t.Fatal("Unexpected error arose in MarshalG1Big")
	}
	if !bytes.Equal(hashMarsh, trueHashMarsh) {
		t.Fatal("HashToG1 fails to agree between Cloudflare and Solidity")
	}

	privK := big.NewInt(1234567890)
	trueSig, err := cloudflare.Sign(msg, privK, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error in cloudflare.Sign:", err)
	}
	trueSigMarsh := trueSig.Marshal()
	sig, err := c.Sign(&bind.CallOpts{}, msg, privK)
	if err != nil {
		t.Fatal("Error in Sign:", err)
	}
	sigMarsh, err := MarshalG1Big(sig)
	if err != nil {
		t.Fatal("Unexpected error arose in MarshalG1Big")
	}
	if !bytes.Equal(sigMarsh, trueSigMarsh) {
		t.Fatal("Sign fails to agree between Cloudflare and Solidity")
	}

	truePubK := new(cloudflare.G2).ScalarBaseMult(privK)
	trueRes, err := cloudflare.Verify(msg, trueSig, truePubK, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error in cloudflare.Verify:", err)
	}
	pubK := G2ToBigIntArray(truePubK)
	res, err := c.Verify(&bind.CallOpts{}, msg, sig, pubK)
	if err != nil {
		t.Fatal("Error in Verify:", err)
	}
	if trueRes != res {
		t.Fatal("Verify fails to agree between Cloudflare and Solidity")
	}

}

// Figure out how Infinity (the group identity element) is stored
func TestSafeSigningPoint(t *testing.T) {
	users := 1
	c, sim, _, _ := SolidityContractSetup(t, users)
	defer sim.Close()

	// G1 point setup
	g1 := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	g1Neg := new(cloudflare.G1).Neg(g1)
	inf := new(cloudflare.G1).Add(g1, g1Neg)
	validPoint := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1234567890))

	// big.Int array setup
	g1B := G1ToBigIntArray(g1)
	g1NegB := G1ToBigIntArray(g1Neg)
	infB := G1ToBigIntArray(inf)
	validPointB := G1ToBigIntArray(validPoint)

	signVal1, err1 := c.SafeSigningPoint(&bind.CallOpts{}, infB)
	if err1 != nil {
		t.Fatal("Error in SafeSigningPoint for Infinity:", err1)
	}
	if signVal1 {
		t.Fatal("Failed to return false for Infinity element")
	}

	signVal2, err2 := c.SafeSigningPoint(&bind.CallOpts{}, g1B)
	if err2 != nil {
		t.Fatal("Error in SafeSigningPoint for curveGen:", err2)
	}
	if signVal2 {
		t.Fatal("Failed to return false for generator element")
	}

	signVal3, err3 := c.SafeSigningPoint(&bind.CallOpts{}, g1NegB)
	if err3 != nil {
		t.Fatal("Error in SafeSigningPoint for negation of curveGen:", err3)
	}
	if signVal3 {
		t.Fatal("Failed to return false for negation of generator element")
	}

	signVal4, err4 := c.SafeSigningPoint(&bind.CallOpts{}, validPointB)
	if err4 != nil {
		t.Fatal("Error in SafeSigningPoint for validPoint:", err4)
	}
	if !signVal4 {
		t.Fatal("Failed to return true for validPoint")
	}

}

func TestLagrangeInterpolationG1(t *testing.T) {
	users := 1
	c, sim, _, _ := SolidityContractSetup(t, users)
	defer sim.Close()

	s1 := big.NewInt(123456790)
	s2 := big.NewInt(234567901)
	s3 := big.NewInt(345679012)
	s4 := big.NewInt(456790123)
	g1p1 := new(cloudflare.G1).ScalarBaseMult(s1)
	g1p2 := new(cloudflare.G1).ScalarBaseMult(s2)
	g1p3 := new(cloudflare.G1).ScalarBaseMult(s3)
	g1p4 := new(cloudflare.G1).ScalarBaseMult(s4)
	pointsG1 := []*cloudflare.G1{g1p1, g1p2, g1p3, g1p4}

	g1p1Big := G1ToBigIntArray(g1p1)
	g1p2Big := G1ToBigIntArray(g1p2)
	g1p3Big := G1ToBigIntArray(g1p3)
	g1p4Big := G1ToBigIntArray(g1p4)
	pointsG1Big := [][2]*big.Int{g1p1Big, g1p2Big, g1p3Big, g1p4Big}

	indices := []int{1, 3, 8, 7}
	indicesBig := make([]*big.Int, len(indices))
	for i, val := range indices {
		indicesBig[i] = big.NewInt(int64(val))
	}
	maxIndex := indices[0]
	for _, idx := range indices {
		if idx > maxIndex {
			maxIndex = idx
		}
	}
	invArray := make([]*big.Int, maxIndex-1)
	invArray[0] = big.NewInt(1)
	for k := 1; k < len(invArray); k++ {
		bigK := big.NewInt(int64(k + 1))
		invArray[k] = new(big.Int).ModInverse(bigK, cloudflare.Order)
	}
	threshold := 3
	thresholdBig := big.NewInt(int64(threshold))

	g1LIG1res, err := cloudflare.LagrangeInterpolationG1(pointsG1, indices, threshold)
	if err != nil {
		t.Fatal("Something is wrong in cloudflare.LagrangeInterpolationG1")
	}
	g1Bytes := g1LIG1res.Marshal()
	bigLIG1res, err := c.LagrangeInterpolationG1(&bind.CallOpts{}, pointsG1Big, indicesBig, thresholdBig, invArray)
	if err != nil {
		t.Fatal("Something is wrong in crypto.LagrangeInterpolationG1")
	}
	bigBytes, err := MarshalG1Big(bigLIG1res)
	if err != nil {
		t.Fatal("Unexpected error arose in MarshalG1Big")
	}
	if !bytes.Equal(g1Bytes, bigBytes) {
		t.Fatal("LIG1 do not agree")
	}

	indices = []int{2, 3, 4, 5}
	indicesBig = make([]*big.Int, len(indices))
	for i, val := range indices {
		indicesBig[i] = big.NewInt(int64(val))
	}
	threshold = 2
	thresholdBig = big.NewInt(int64(threshold))

	g1LIG1res, err = cloudflare.LagrangeInterpolationG1(pointsG1, indices, threshold)
	if err != nil {
		t.Fatal("Something is wrong in cloudflare.LagrangeInterpolationG1")
	}
	g1Bytes = g1LIG1res.Marshal()
	bigLIG1res, err = c.LagrangeInterpolationG1(&bind.CallOpts{}, pointsG1Big, indicesBig, thresholdBig, invArray)
	if err != nil {
		t.Fatal("Something is wrong in crypto.LagrangeInterpolationG1")
	}
	bigBytes, err = MarshalG1Big(bigLIG1res)
	if err != nil {
		t.Fatal("Unexpected error arose in MarshalG1Big")
	}
	if !bytes.Equal(g1Bytes, bigBytes) {
		t.Fatal("LIG1 do not agree")
	}
}

func TestAggregateSignatures(t *testing.T) {
	users := 1
	c, sim, _, _ := SolidityContractSetup(t, users)
	defer sim.Close()

	threshold := 2
	thresholdBig := big.NewInt(int64(threshold))

	invArray := make([]*big.Int, 4)
	invArray[0] = big.NewInt(1)
	for k := 1; k < len(invArray); k++ {
		bigK := big.NewInt(int64(k + 1))
		invArray[k] = new(big.Int).ModInverse(bigK, cloudflare.Order)
	}

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
	msk.Mod(msk, cloudflare.Order)
	mpk := &cloudflare.G2{}
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

	share1to1 := cloudflare.PrivatePolyEval(privCoefs1, 1)
	share1to2 := cloudflare.PrivatePolyEval(privCoefs1, 2)
	share1to3 := cloudflare.PrivatePolyEval(privCoefs1, 3)
	share1to4 := cloudflare.PrivatePolyEval(privCoefs1, 4)
	share2to1 := cloudflare.PrivatePolyEval(privCoefs2, 1)
	share2to2 := cloudflare.PrivatePolyEval(privCoefs2, 2)
	share2to3 := cloudflare.PrivatePolyEval(privCoefs2, 3)
	share2to4 := cloudflare.PrivatePolyEval(privCoefs2, 4)
	share3to1 := cloudflare.PrivatePolyEval(privCoefs3, 1)
	share3to2 := cloudflare.PrivatePolyEval(privCoefs3, 2)
	share3to3 := cloudflare.PrivatePolyEval(privCoefs3, 3)
	share3to4 := cloudflare.PrivatePolyEval(privCoefs3, 4)
	share4to1 := cloudflare.PrivatePolyEval(privCoefs4, 1)
	share4to2 := cloudflare.PrivatePolyEval(privCoefs4, 2)
	share4to3 := cloudflare.PrivatePolyEval(privCoefs4, 3)
	share4to4 := cloudflare.PrivatePolyEval(privCoefs4, 4)

	listOfSS1 := []*big.Int{share1to1, share2to1, share3to1, share4to1}
	gsk1 := cloudflare.GenerateGroupSecretKeyPortion(listOfSS1)
	gpk1 := &cloudflare.G2{}
	gpk1.ScalarBaseMult(gsk1)

	listOfSS2 := []*big.Int{share1to2, share2to2, share3to2, share4to2}
	gsk2 := cloudflare.GenerateGroupSecretKeyPortion(listOfSS2)
	gpk2 := &cloudflare.G2{}
	gpk2.ScalarBaseMult(gsk2)

	listOfSS3 := []*big.Int{share1to3, share2to3, share3to3, share4to3}
	gsk3 := cloudflare.GenerateGroupSecretKeyPortion(listOfSS3)
	gpk3 := &cloudflare.G2{}
	gpk3.ScalarBaseMult(gsk3)

	listOfSS4 := []*big.Int{share1to4, share2to4, share3to4, share4to4}
	gsk4 := cloudflare.GenerateGroupSecretKeyPortion(listOfSS4)
	gpk4 := &cloudflare.G2{}
	gpk4.ScalarBaseMult(gsk4)

	s := "MadHive Rocks!"
	msg := []byte(s)

	partialSigMH1, err := cloudflare.Sign(msg, gsk1, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with gsk1:", err)
	}
	val, err := cloudflare.Verify(msg, partialSigMH1, gpk1, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying MH with gpk1:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH gpk1 signature")
	}
	ethPSigMH1 := G1ToBigIntArray(partialSigMH1)

	partialSigMH2, err := cloudflare.Sign(msg, gsk2, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with gsk2")
	}
	val, err = cloudflare.Verify(msg, partialSigMH2, gpk2, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying MH with gpk2:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH gpk2 signature")
	}
	ethPSigMH2 := G1ToBigIntArray(partialSigMH2)

	partialSigMH3, err := cloudflare.Sign(msg, gsk3, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with gsk3")
	}
	val, err = cloudflare.Verify(msg, partialSigMH3, gpk3, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying MH with gpk3:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH gpk3 signature")
	}
	ethPSigMH3 := G1ToBigIntArray(partialSigMH3)

	partialSigMH4, err := cloudflare.Sign(msg, gsk4, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in signing MH with gsk4")
	}
	val, err = cloudflare.Verify(msg, partialSigMH4, gpk4, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in verifying MH with gpk4:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH gpk4 signature")
	}
	ethPSigMH4 := G1ToBigIntArray(partialSigMH4)

	list1Sigs := [][2]*big.Int{ethPSigMH1, ethPSigMH2, ethPSigMH3}
	indices1 := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}
	grpsigBig1, err := c.AggregateSignatures(&bind.CallOpts{}, list1Sigs, indices1, thresholdBig, invArray)
	if err != nil {
		t.Fatal("Error occurred in AggregateSignatures for MH and list1:", err)
	}
	grpsig := new(cloudflare.G1)
	grpsigBytes, err := MarshalG1Big(grpsigBig1)
	if err != nil {
		t.Fatal("Unexpected error arose in MarshalG1Big")
	}
	_, err = grpsig.Unmarshal(grpsigBytes)
	if err != nil {
		t.Fatal("Error occurred in converting [2]*big.Int to G1")
	}
	val, err = cloudflare.Verify(msg, grpsig, mpk, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in group signature MH 1 verification:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH 1 group signature")
	}

	list2Sigs := [][2]*big.Int{ethPSigMH1, ethPSigMH2, ethPSigMH3, ethPSigMH4}
	indices2 := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4)}
	grpsigBig2, err := c.AggregateSignatures(&bind.CallOpts{}, list2Sigs, indices2, thresholdBig, invArray)
	if err != nil {
		t.Fatal("Error occurred in AggregateSignatures for MH and list2:", err)
	}
	grpsigBytes, err = MarshalG1Big(grpsigBig2)
	if err != nil {
		t.Fatal("Unexpected error arose in MarshalG1Big")
	}
	_, err = grpsig.Unmarshal(grpsigBytes)
	if err != nil {
		t.Fatal("Error occurred in converting [2]*big.Int to G1")
	}
	val, err = cloudflare.Verify(msg, grpsig, mpk, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in group signature MH 2 verification:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH 2 group signature")
	}

	list3Sigs := [][2]*big.Int{ethPSigMH3, ethPSigMH2, ethPSigMH1, ethPSigMH4}
	indices3 := []*big.Int{big.NewInt(3), big.NewInt(2), big.NewInt(1), big.NewInt(4)}
	grpsigBig3, err := c.AggregateSignatures(&bind.CallOpts{}, list3Sigs, indices3, thresholdBig, invArray)
	if err != nil {
		t.Fatal("Error occurred in AggregateSignatures for MH and list3:", err)
	}
	grpsigBytes, err = MarshalG1Big(grpsigBig3)
	if err != nil {
		t.Fatal("Unexpected error arose in MarshalG1Big")
	}
	_, err = grpsig.Unmarshal(grpsigBytes)
	if err != nil {
		t.Fatal("Error occurred in converting [2]*big.Int to G1")
	}
	val, err = cloudflare.Verify(msg, grpsig, mpk, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in group signature MH 3 verification:", err)
	}
	if !val {
		t.Fatal("Failed to verify MH 3 group signature")
	}

	// Signature and index array length mismatch
	list4Sigs := [][2]*big.Int{ethPSigMH1, ethPSigMH2}
	indices4 := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4)}
	_, err = c.AggregateSignatures(&bind.CallOpts{}, list4Sigs, indices4, thresholdBig, invArray)
	if err == nil {
		t.Fatal("Error should have occurred in AggregateSignatures for mismatch arrays", err)
	}

	// Insufficient signature length; below threshold.
	list5Sigs := [][2]*big.Int{ethPSigMH1, ethPSigMH2}
	indices5 := []*big.Int{big.NewInt(1), big.NewInt(2)}
	_, err = c.AggregateSignatures(&bind.CallOpts{}, list5Sigs, indices5, thresholdBig, invArray)
	if err == nil {
		t.Fatal("Error should have occurred in AggregateSignatures for not meeting threshold", err)
	}

	// Invalid inverses in invArray
	invArray[0] = big.NewInt(2)
	_, err = c.AggregateSignatures(&bind.CallOpts{}, list1Sigs, indices1, thresholdBig, invArray)
	if err == nil {
		t.Fatal("Error should have occurred in AggregateSignatures for invalid invArray", err)
	}
	invArray[0] = big.NewInt(1) // Change it back

	// Insufficient number of inverses; invArray not long enough.
	invArray = invArray[:2]
	_, err = c.AggregateSignatures(&bind.CallOpts{}, list3Sigs, indices3, thresholdBig, invArray)
	if err == nil {
		t.Fatal("Error should have occurred in AggregateSignatures for invalid invArray", err)
	}
}
