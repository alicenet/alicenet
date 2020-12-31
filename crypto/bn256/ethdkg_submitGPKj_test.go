package bn256

import (
	"bytes"
	"context"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// We successfully submit valid gpkj's.
func TestSubmitGPKjSuccess(t *testing.T) {
	n := 4
	threshold, _ := thresholdFromUsers(n) // threshold, k are return values
	_ = threshold                         // for linter
	//c, sim, keyArray, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	c, _, sim, _, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	defer sim.Close()
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error in getting RegistrationEnd")
	}
	AdvanceBlocksUntil(sim, registrationEnd)
	// Current block number is now 22 > 21 == T_REGISTRATION_END;
	// in Share Distribution phase

	// These are the standard secrets be used for testing purposes
	secretValuesArray := make([]*big.Int, n)
	secretBase := big.NewInt(100)
	for j := 0; j < n; j++ {
		secretValuesArray[j] = new(big.Int).Add(secretBase, big.NewInt(int64(j)))
	}

	// These are the standard private polynomial coefs for testing purposes
	basePrivatePolynomialCoefs := make([]*big.Int, threshold+1)
	for j := 1; j < threshold+1; j++ {
		basePrivatePolynomialCoefs[j] = big.NewInt(int64(j))
	}

	// Create private polynomials for all users
	privPolyCoefsArray := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		privPolyCoefsArray[ell] = make([]*big.Int, threshold+1)
		privPolyCoefsArray[ell][0] = secretValuesArray[ell]
		for j := 1; j < threshold+1; j++ {
			privPolyCoefsArray[ell][j] = basePrivatePolynomialCoefs[j]
		}
	}

	// Create public coefficients for all users
	pubCoefsArray := make([][]*cloudflare.G1, n)
	for ell := 0; ell < n; ell++ {
		pubCoefsArray[ell] = make([]*cloudflare.G1, threshold+1)
		pubCoefsArray[ell] = cloudflare.GeneratePublicCoefs(privPolyCoefsArray[ell])
	}

	// Create big.Int version of public coefficients
	pubCoefsBigArray := make([][][2]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		pubCoefsBigArray[ell] = make([][2]*big.Int, threshold+1)
		for j := 0; j < threshold+1; j++ {
			pubCoefsBigArray[ell][j] = G1ToBigIntArray(pubCoefsArray[ell][j])
		}
	}

	// Create encrypted shares to submit
	encSharesArray := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		privK := privKArray[ell]
		pubK := pubKArray[ell]
		encSharesArray[ell] = make([]*big.Int, n-1)
		secretsArray, err := cloudflare.GenerateSecretShares(pubK, privPolyCoefsArray[ell], pubKArray)
		if err != nil {
			t.Fatal("Error occurred while generating sharing secrets")
		}
		encSharesArray[ell], err = cloudflare.GenerateEncryptedShares(secretsArray, privK, pubKArray)
		if err != nil {
			t.Fatal("Error occurred while generating commitments")
		}
	}

	// Create arrays to hold submitted information
	// First index is participant receiving (n), then who from (n), then values (n-1);
	// note that this would have to be changed in practice
	rcvdEncShares := make([][][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		rcvdEncShares[ell] = make([][]*big.Int, n)
		for j := 0; j < n; j++ {
			rcvdEncShares[ell][j] = make([]*big.Int, n-1)
		}
	}
	rcvdCommitments := make([][][][2]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		rcvdCommitments[ell] = make([][][2]*big.Int, n)
		for j := 0; j < n; j++ {
			rcvdCommitments[ell][j] = make([][2]*big.Int, threshold+1)
		}
	}

	big0 := big.NewInt(0)
	big1 := big.NewInt(1)
	big2 := big.NewInt(2)
	big3 := big.NewInt(3)

	// Submit encrypted shares and commitments
	for ell := 0; ell < n; ell++ {
		auth := authArray[ell]
		encShares := encSharesArray[ell]
		pubCoefs := pubCoefsBigArray[ell]
		txOpt := &bind.TransactOpts{
			From:     auth.From,
			Nonce:    nil,
			Signer:   auth.Signer,
			Value:    nil,
			GasPrice: nil,
			GasLimit: gasLim,
			Context:  nil,
		}
		// Check public_key to ensure registered
		pubKBigRcvd0, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (0)")
		}
		pubKBigRcvd1, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (1)")
		}
		registeredPubK := (pubKBigRcvd0.Cmp(big0) != 0) || (pubKBigRcvd1.Cmp(big0) != 0)
		if !registeredPubK {
			t.Fatal("Public Key already exists")
		}
		txn, err := c.DistributeShares(txOpt, encShares, pubCoefs)
		if err != nil {
			t.Fatal("Unexpected error arose in DistributeShares submission")
		}
		sim.Commit()
		receipt, err := sim.TransactionReceipt(context.Background(), txn.Hash())
		if err != nil {
			t.Fatal("Unexpected error in TransactionReceipt")
		}
		shareDistEvent, err := c.ETHDKGFilterer.ParseShareDistribution(*receipt.Logs[0])
		if err != nil {
			t.Fatal("Unexpected error in ParseShareDistribution")
		}
		// Save values in arrays for everyone
		for j := 0; j < n; j++ {
			if j == ell {
				continue
			}
			rcvdEncShares[j][ell] = shareDistEvent.EncryptedShares
			rcvdCommitments[j][ell] = shareDistEvent.Commitments
		}
	}
	// Everything above is good but now we want to check stuff like events and logs

	ssArrayAll := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		ssArrayAll[ell] = make([]*big.Int, n-1)
	}

	// HERE WE GO
	for ell := 0; ell < n; ell++ {
		rcvdEncSharesEll := rcvdEncShares[ell]
		pubK := pubKArray[ell]
		privK := privKArray[ell]
		sharedEncryptedArray, err := cloudflare.CondenseCommitments(pubK, rcvdEncSharesEll, pubKArray)
		if err != nil {
			t.Fatal("Unexpected error occurred when condensing commitments")
		}
		sharedSecretsArray, err := cloudflare.GenerateDecryptedShares(privK, sharedEncryptedArray, pubKArray)
		if err != nil {
			t.Fatal("Unexpected error occurred when decrypting secrets")
		}
		ssArrayAll[ell] = sharedSecretsArray
	}

	gskjArray := make([]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		sharedSecretsArray := ssArrayAll[ell]
		privPolyCoefs := privPolyCoefsArray[ell]
		idx := ell + 1
		selfSecret := cloudflare.PrivatePolyEval(privPolyCoefs, idx)
		gskj := new(big.Int).Set(selfSecret)
		for j := 0; j < n-1; j++ {
			sharedSecret := sharedSecretsArray[j]
			gskj.Add(gskj, sharedSecret)
		}
		gskjArray[ell] = gskj
	}

	shareDistributionEnd, err := c.TSHAREDISTRIBUTIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting ShareDistributionEnd")
	}
	AdvanceBlocksUntil(sim, shareDistributionEnd)
	// Current block number is now 47 > 46 == T_SHARE_DISTRIBUTION_END;
	// in Dispute phase
	disputeEnd, err := c.TDISPUTEEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting DisputeEnd")
	}
	AdvanceBlocksUntil(sim, disputeEnd)
	// Current block number is now 72 > 71 == T_DISPUTE_END;
	// in Key Derivation phase

	// Check block number here
	curBlock := sim.Blockchain().CurrentBlock().Number()
	keyShareSubmissionEnd, err := c.TKEYSHARESUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting KeyShareSubmissionEnd")
	}
	validBlockNumber := (disputeEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(keyShareSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in Key Share Submission Phase")
	}

	// Now to submit key shares
	keyShareArrayG1 := make([]*cloudflare.G1, n)
	keyShareArrayG2 := make([]*cloudflare.G2, n)
	keyShareArrayDLEQProof := make([][2]*big.Int, n)

	h1BaseMsg := []byte("MadHive Rocks!")
	g1Base := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	h1Base, err := cloudflare.HashToG1(h1BaseMsg)
	if err != nil {
		t.Fatal("Error when computing HashToG1([]byte(\"MadHive Rock!\"))")
	}
	h2Base := new(cloudflare.G2).ScalarBaseMult(big.NewInt(1))
	orderMinus1, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495616", 10)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	for ell := 0; ell < n; ell++ {
		secretValue := secretValuesArray[ell]
		g1Value := new(cloudflare.G1).ScalarBaseMult(secretValue)
		keyShareG1 := new(cloudflare.G1).ScalarMult(h1Base, secretValue)
		keyShareG2 := new(cloudflare.G2).ScalarMult(h2Base, secretValue)

		// Generate and Verify DLEQ Proof
		keyShareDLEQProof, err := cloudflare.GenerateDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, secretValue, rand.Reader)
		if err != nil {
			t.Fatal("Error when generating DLEQ Proof")
		}
		err = cloudflare.VerifyDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, keyShareDLEQProof)
		if err != nil {
			t.Fatal("Invalid DLEQ h1Value proof")
		}

		// PairingCheck to ensure keyShareG1 and keyShareG2 form valid pair
		validPair := cloudflare.PairingCheck([]*cloudflare.G1{keyShareG1, h1Base}, []*cloudflare.G2{h2Neg, keyShareG2})
		if !validPair {
			t.Fatal("Error in PairingCheck")
		}

		auth := authArray[ell]
		txOpt := &bind.TransactOpts{
			From:     auth.From,
			Nonce:    nil,
			Signer:   auth.Signer,
			Value:    nil,
			GasPrice: nil,
			GasLimit: gasLim,
			Context:  nil,
		}

		// Check Key Shares to ensure not submitted
		keyShareBig0, err := c.KeyShares(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (0)")
		}
		keyShareBig1, err := c.KeyShares(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (1)")
		}
		zeroKeyShare := (keyShareBig0.Cmp(big0) == 0) && (keyShareBig1.Cmp(big0) == 0)
		if !zeroKeyShare {
			t.Fatal("Unexpected error: KeyShare is nonzero and already present")
		}

		// Check Share Distribution Hashes
		authHash, err := c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
		if err != nil {
			t.Fatal("Error when calling ShareDistributionHashes")
		}
		zeroBytes := make([]byte, numBytes)
		validHash := !bytes.Equal(authHash[:], zeroBytes)
		if !validHash {
			t.Fatal("Unexpected error: invalid hash")
		}

		keyShareG1Big := G1ToBigIntArray(keyShareG1)
		keyShareG2Big := G2ToBigIntArray(keyShareG2)

		_, err = c.SubmitKeyShare(txOpt, auth.From, keyShareG1Big, keyShareDLEQProof, keyShareG2Big)
		if err != nil {
			t.Fatal("Unexpected error occurred when submitting key shares")
		}

		keyShareArrayG1[ell] = keyShareG1
		keyShareArrayG2[ell] = keyShareG2
		keyShareArrayDLEQProof[ell] = keyShareDLEQProof
	}
	sim.Commit()

	// Need to check key share submission and confirm validity
	for ell := 0; ell < n; ell++ {
		auth := authArray[ell]
		keyShareG1 := keyShareArrayG1[ell]
		keyShareBig0, err := c.KeyShares(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (0)")
		}
		keyShareBig1, err := c.KeyShares(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (1)")
		}
		keyShareG1Rcvd, err := BigIntArrayToG1([2]*big.Int{keyShareBig0, keyShareBig1})
		if err != nil {
			t.Fatal("Error in BigIntArrayToG1 call")
		}
		if !keyShareG1.IsEqual(keyShareG1Rcvd) {
			t.Fatal("KeyShareG1 mismatch between submission and received!")
		}
	}

	AdvanceBlocksUntil(sim, keyShareSubmissionEnd)
	// Check block number here
	curBlock = sim.Blockchain().CurrentBlock().Number()
	mpkSubmissionEnd, err := c.TMPKSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting MPKSubmissionEnd")
	}
	validBlockNumber = (keyShareSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(mpkSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in MPK Submission Phase")
	}

	// Make Master Public Key (this is not how you would actually do this)
	mpk := new(cloudflare.G2).Add(keyShareArrayG2[0], keyShareArrayG2[1])
	for ell := 2; ell < n; ell++ {
		mpk.Add(mpk, keyShareArrayG2[ell])
	}
	mpkBig := G2ToBigIntArray(mpk)

	// For G1 version
	mpkG1 := new(cloudflare.G1).Add(keyShareArrayG1[0], keyShareArrayG1[1])
	for ell := 2; ell < n; ell++ {
		mpkG1.Add(mpkG1, keyShareArrayG1[ell])
	}

	// Perform PairingCheck on mpk and mpkG1 to ensure valid pair
	validPair := cloudflare.PairingCheck([]*cloudflare.G1{mpkG1, h1Base}, []*cloudflare.G2{h2Neg, mpk})
	if !validPair {
		t.Fatal("Error in PairingCheck for mpk")
	}

	auth := authArray[0]
	txOpt := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	_, err = c.SubmitMasterPublicKey(txOpt, mpkBig)
	if err != nil {
		t.Fatal("Unexpected error occurred when submitting master public key")
	}
	sim.Commit()

	mpkRcvd0, err := c.MasterPublicKey(&bind.CallOpts{}, big0)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (0)")
	}
	mpkRcvd1, err := c.MasterPublicKey(&bind.CallOpts{}, big1)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (1)")
	}
	mpkRcvd2, err := c.MasterPublicKey(&bind.CallOpts{}, big2)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (2)")
	}
	mpkRcvd3, err := c.MasterPublicKey(&bind.CallOpts{}, big3)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (3)")
	}
	mpkSubmittedMatchesRcvd := (mpkBig[0].Cmp(mpkRcvd0) == 0) && (mpkBig[1].Cmp(mpkRcvd1) == 0) && (mpkBig[2].Cmp(mpkRcvd2) == 0) && (mpkBig[3].Cmp(mpkRcvd3) == 0)
	if !mpkSubmittedMatchesRcvd {
		t.Fatal("mpk submitted does not match received!")
	}

	// We now proceed to submit gpkj's; they were created above

	// Check block number here
	AdvanceBlocksUntil(sim, mpkSubmissionEnd)
	gpkjSubmissionEnd, err := c.TGPKJSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting GPKJSubmissionEnd")
	}
	curBlock = sim.Blockchain().CurrentBlock().Number()
	validBlockNumber = (mpkSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(gpkjSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in GPKj Submission Phase")
	}

	initialMessage, err := c.InitialMessage(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error when getting InitialMessage for gpkj signature")
	}

	// Make and submit gpkj's
	initialSigArray := make([]*cloudflare.G1, n)
	gpkjArray := make([]*cloudflare.G2, n)
	for ell := 0; ell < n; ell++ {
		gskj := gskjArray[ell]
		gpkj := new(cloudflare.G2).ScalarBaseMult(gskj)
		initialSig, err := cloudflare.Sign(initialMessage, gskj, cloudflare.HashToG1)
		if err != nil {
			t.Fatal("Error occurred in cloudflare.Sign when signing initialMessage")
		}
		gpkjBig := G2ToBigIntArray(gpkj)
		initialSigBig := G1ToBigIntArray(initialSig)

		auth := authArray[ell]
		txOpt := &bind.TransactOpts{
			From:     auth.From,
			Nonce:    nil,
			Signer:   auth.Signer,
			Value:    nil,
			GasPrice: nil,
			GasLimit: gasLim,
			Context:  nil,
		}

		// Ensure no previous submission
		gpkjSubmission0, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Error when calling GpkjSubmission0")
		}
		gpkjSubmission1, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Error when calling GpkjSubmission1")
		}
		gpkjSubmission2, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big2)
		if err != nil {
			t.Fatal("Error when calling GpkjSubmission2")
		}
		gpkjSubmission3, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big3)
		if err != nil {
			t.Fatal("Error when calling GpkjSubmission3")
		}
		emptyGPKjSub := (gpkjSubmission0.Cmp(big0) == 0) && (gpkjSubmission1.Cmp(big0) == 0) && (gpkjSubmission2.Cmp(big0) == 0) && (gpkjSubmission3.Cmp(big0) == 0)
		if !emptyGPKjSub {
			t.Fatal("Unexpected error; gpkj already submitted")
		}

		// Verify signature
		validSig, err := cloudflare.Verify(initialMessage, initialSig, gpkj, cloudflare.HashToG1)
		if err != nil {
			t.Fatal("Error when calling cloudflare.Verify for (initialSig, gpkj) verification")
		}
		if !validSig {
			t.Fatal("Unexpected error; initialSig fails cloudflare.Verify")
		}

		_, err = c.SubmitGPKj(txOpt, gpkjBig, initialSigBig)
		if err != nil {
			t.Fatal("Error occurred when submitting gpkj")
		}

		initialSigArray[ell] = initialSig
		gpkjArray[ell] = gpkj
	}
	sim.Commit()

	// Confirm submissions
	for ell := 0; ell < n; ell++ {
		auth := authArray[ell]
		initialSig := initialSigArray[ell]
		initialSigBig := G1ToBigIntArray(initialSig)
		gpkj := gpkjArray[ell]
		gpkjBig := G2ToBigIntArray(gpkj)

		// Get Submission for gpkj and confirm
		gpkjRcvd0, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Error when calling GpkjSubmission0")
		}
		gpkjRcvd1, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Error when calling GpkjSubmission1")
		}
		gpkjRcvd2, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big2)
		if err != nil {
			t.Fatal("Error when calling GpkjSubmission2")
		}
		gpkjRcvd3, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big3)
		if err != nil {
			t.Fatal("Error when calling GpkjSubmission3")
		}
		matchGPKjSub := (gpkjRcvd0.Cmp(gpkjBig[0]) == 0) && (gpkjRcvd1.Cmp(gpkjBig[1]) == 0) && (gpkjRcvd2.Cmp(gpkjBig[2]) == 0) && (gpkjRcvd3.Cmp(gpkjBig[3]) == 0)
		if !matchGPKjSub {
			t.Fatal("Unexpected error; gpkjRcvd does not match submission")
		}

		// Get Submission for initialSig and confirm
		iSigRcvd0, err := c.InitialSignatures(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Error when calling GpkjSubmission0")
		}
		iSigRcvd1, err := c.InitialSignatures(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Error when calling GpkjSubmission1")
		}
		matchISigSub := (iSigRcvd0.Cmp(initialSigBig[0]) == 0) && (iSigRcvd1.Cmp(initialSigBig[1]) == 0)
		if !matchISigSub {
			t.Fatal("Unexpected error; iSigRcvd does not match submission")
		}
	}

	// Validate gpkj's by looking at aggregate signatures

	// Test first batch
	fbSigs := make([]*cloudflare.G1, threshold+1)
	fbIndices := make([]int, threshold+1)
	for ell := 0; ell < threshold+1; ell++ {
		fbSigs[ell] = initialSigArray[ell]
		fbIndices[ell] = ell + 1
	}
	fbGrpsig, err := cloudflare.AggregateSignatures(fbSigs, fbIndices, threshold)
	if err != nil {
		t.Fatal("Error in cloudflare.AggregateSignatures")
	}
	validGrpsigFB, err := cloudflare.Verify(initialMessage, fbGrpsig, mpk, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error in cloudflare.Verify")
	}
	if !validGrpsigFB {
		t.Fatal("First batch failed to form valid group signature")
	}

	// Test second batch
	sbSigs := make([]*cloudflare.G1, threshold+1)
	sbIndices := make([]int, threshold+1)
	for ell := 0; ell < threshold+1; ell++ {
		sbSigs[ell] = initialSigArray[ell+n-threshold-1]
		sbIndices[ell] = ell + n - threshold
	}
	sbGrpsig, err := cloudflare.AggregateSignatures(sbSigs, sbIndices, threshold)
	if err != nil {
		t.Fatal("Error in cloudflare.AggregateSignatures")
	}
	validGrpsigSB, err := cloudflare.Verify(initialMessage, sbGrpsig, mpk, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error in cloudflare.Verify")
	}
	if !validGrpsigSB {
		t.Fatal("Second batch failed to form valid group signature")
	}
}

// We attempt to submit gpkj but fail due to incorrect block number.
func TestSubmitGPKjFailBlockNumber(t *testing.T) {
	n := 4
	threshold, _ := thresholdFromUsers(n) // threshold, k are return values
	_ = threshold                         // for linter
	//c, sim, keyArray, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	c, _, sim, _, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	defer sim.Close()
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error in getting RegistrationEnd")
	}
	AdvanceBlocksUntil(sim, registrationEnd)
	// Current block number is now 22 > 21 == T_REGISTRATION_END;
	// in Share Distribution phase

	// These are the standard secrets be used for testing purposes
	secretValuesArray := make([]*big.Int, n)
	secretBase := big.NewInt(100)
	for j := 0; j < n; j++ {
		secretValuesArray[j] = new(big.Int).Add(secretBase, big.NewInt(int64(j)))
	}

	// These are the standard private polynomial coefs for testing purposes
	basePrivatePolynomialCoefs := make([]*big.Int, threshold+1)
	for j := 1; j < threshold+1; j++ {
		basePrivatePolynomialCoefs[j] = big.NewInt(int64(j))
	}

	// Create private polynomials for all users
	privPolyCoefsArray := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		privPolyCoefsArray[ell] = make([]*big.Int, threshold+1)
		privPolyCoefsArray[ell][0] = secretValuesArray[ell]
		for j := 1; j < threshold+1; j++ {
			privPolyCoefsArray[ell][j] = basePrivatePolynomialCoefs[j]
		}
	}

	// Create public coefficients for all users
	pubCoefsArray := make([][]*cloudflare.G1, n)
	for ell := 0; ell < n; ell++ {
		pubCoefsArray[ell] = make([]*cloudflare.G1, threshold+1)
		pubCoefsArray[ell] = cloudflare.GeneratePublicCoefs(privPolyCoefsArray[ell])
	}

	// Create big.Int version of public coefficients
	pubCoefsBigArray := make([][][2]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		pubCoefsBigArray[ell] = make([][2]*big.Int, threshold+1)
		for j := 0; j < threshold+1; j++ {
			pubCoefsBigArray[ell][j] = G1ToBigIntArray(pubCoefsArray[ell][j])
		}
	}

	// Create encrypted shares to submit
	encSharesArray := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		privK := privKArray[ell]
		pubK := pubKArray[ell]
		encSharesArray[ell] = make([]*big.Int, n-1)
		secretsArray, err := cloudflare.GenerateSecretShares(pubK, privPolyCoefsArray[ell], pubKArray)
		if err != nil {
			t.Fatal("Error occurred while generating sharing secrets")
		}
		encSharesArray[ell], err = cloudflare.GenerateEncryptedShares(secretsArray, privK, pubKArray)
		if err != nil {
			t.Fatal("Error occurred while generating commitments")
		}
	}

	// Create arrays to hold submitted information
	// First index is participant receiving (n), then who from (n), then values (n-1);
	// note that this would have to be changed in practice
	rcvdEncShares := make([][][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		rcvdEncShares[ell] = make([][]*big.Int, n)
		for j := 0; j < n; j++ {
			rcvdEncShares[ell][j] = make([]*big.Int, n-1)
		}
	}
	rcvdCommitments := make([][][][2]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		rcvdCommitments[ell] = make([][][2]*big.Int, n)
		for j := 0; j < n; j++ {
			rcvdCommitments[ell][j] = make([][2]*big.Int, threshold+1)
		}
	}

	big0 := big.NewInt(0)
	big1 := big.NewInt(1)
	big2 := big.NewInt(2)
	big3 := big.NewInt(3)

	// Submit encrypted shares and commitments
	for ell := 0; ell < n; ell++ {
		auth := authArray[ell]
		encShares := encSharesArray[ell]
		pubCoefs := pubCoefsBigArray[ell]
		txOpt := &bind.TransactOpts{
			From:     auth.From,
			Nonce:    nil,
			Signer:   auth.Signer,
			Value:    nil,
			GasPrice: nil,
			GasLimit: gasLim,
			Context:  nil,
		}
		// Check public_key to ensure registered
		pubKBigRcvd0, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (0)")
		}
		pubKBigRcvd1, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (1)")
		}
		registeredPubK := (pubKBigRcvd0.Cmp(big0) != 0) || (pubKBigRcvd1.Cmp(big0) != 0)
		if !registeredPubK {
			t.Fatal("Public Key already exists")
		}
		txn, err := c.DistributeShares(txOpt, encShares, pubCoefs)
		if err != nil {
			t.Fatal("Unexpected error arose in DistributeShares submission")
		}
		sim.Commit()
		receipt, err := sim.TransactionReceipt(context.Background(), txn.Hash())
		if err != nil {
			t.Fatal("Unexpected error in TransactionReceipt")
		}
		shareDistEvent, err := c.ETHDKGFilterer.ParseShareDistribution(*receipt.Logs[0])
		if err != nil {
			t.Fatal("Unexpected error in ParseShareDistribution")
		}
		// Save values in arrays for everyone
		for j := 0; j < n; j++ {
			if j == ell {
				continue
			}
			rcvdEncShares[j][ell] = shareDistEvent.EncryptedShares
			rcvdCommitments[j][ell] = shareDistEvent.Commitments
		}
	}
	// Everything above is good but now we want to check stuff like events and logs

	ssArrayAll := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		ssArrayAll[ell] = make([]*big.Int, n-1)
	}

	// HERE WE GO
	for ell := 0; ell < n; ell++ {
		rcvdEncSharesEll := rcvdEncShares[ell]
		pubK := pubKArray[ell]
		privK := privKArray[ell]
		sharedEncryptedArray, err := cloudflare.CondenseCommitments(pubK, rcvdEncSharesEll, pubKArray)
		if err != nil {
			t.Fatal("Unexpected error occurred when condensing commitments")
		}
		sharedSecretsArray, err := cloudflare.GenerateDecryptedShares(privK, sharedEncryptedArray, pubKArray)
		if err != nil {
			t.Fatal("Unexpected error occurred when decrypting secrets")
		}
		ssArrayAll[ell] = sharedSecretsArray
	}

	gskjArray := make([]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		sharedSecretsArray := ssArrayAll[ell]
		privPolyCoefs := privPolyCoefsArray[ell]
		idx := ell + 1
		selfSecret := cloudflare.PrivatePolyEval(privPolyCoefs, idx)
		gskj := new(big.Int).Set(selfSecret)
		for j := 0; j < n-1; j++ {
			sharedSecret := sharedSecretsArray[j]
			gskj.Add(gskj, sharedSecret)
		}
		gskjArray[ell] = gskj
	}

	shareDistributionEnd, err := c.TSHAREDISTRIBUTIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting ShareDistributionEnd")
	}
	AdvanceBlocksUntil(sim, shareDistributionEnd)
	// Current block number is now 47 > 46 == T_SHARE_DISTRIBUTION_END;
	// in Dispute phase
	disputeEnd, err := c.TDISPUTEEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting DisputeEnd")
	}
	AdvanceBlocksUntil(sim, disputeEnd)
	// Current block number is now 72 > 71 == T_DISPUTE_END;
	// in Key Derivation phase

	// Check block number here
	curBlock := sim.Blockchain().CurrentBlock().Number()
	keyShareSubmissionEnd, err := c.TKEYSHARESUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting KeyShareSubmissionEnd")
	}
	validBlockNumber := (disputeEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(keyShareSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in Key Share Submission Phase")
	}

	// Now to submit key shares
	keyShareArrayG1 := make([]*cloudflare.G1, n)
	keyShareArrayG2 := make([]*cloudflare.G2, n)
	keyShareArrayDLEQProof := make([][2]*big.Int, n)

	h1BaseMsg := []byte("MadHive Rocks!")
	g1Base := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	h1Base, err := cloudflare.HashToG1(h1BaseMsg)
	if err != nil {
		t.Fatal("Error when computing HashToG1([]byte(\"MadHive Rock!\"))")
	}
	h2Base := new(cloudflare.G2).ScalarBaseMult(big.NewInt(1))
	orderMinus1, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495616", 10)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	for ell := 0; ell < n; ell++ {
		secretValue := secretValuesArray[ell]
		g1Value := new(cloudflare.G1).ScalarBaseMult(secretValue)
		keyShareG1 := new(cloudflare.G1).ScalarMult(h1Base, secretValue)
		keyShareG2 := new(cloudflare.G2).ScalarMult(h2Base, secretValue)

		// Generate and Verify DLEQ Proof
		keyShareDLEQProof, err := cloudflare.GenerateDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, secretValue, rand.Reader)
		if err != nil {
			t.Fatal("Error when generating DLEQ Proof")
		}
		err = cloudflare.VerifyDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, keyShareDLEQProof)
		if err != nil {
			t.Fatal("Invalid DLEQ h1Value proof")
		}

		// PairingCheck to ensure keyShareG1 and keyShareG2 form valid pair
		validPair := cloudflare.PairingCheck([]*cloudflare.G1{keyShareG1, h1Base}, []*cloudflare.G2{h2Neg, keyShareG2})
		if !validPair {
			t.Fatal("Error in PairingCheck")
		}

		auth := authArray[ell]
		txOpt := &bind.TransactOpts{
			From:     auth.From,
			Nonce:    nil,
			Signer:   auth.Signer,
			Value:    nil,
			GasPrice: nil,
			GasLimit: gasLim,
			Context:  nil,
		}

		// Check Key Shares to ensure not submitted
		keyShareBig0, err := c.KeyShares(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (0)")
		}
		keyShareBig1, err := c.KeyShares(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (1)")
		}
		zeroKeyShare := (keyShareBig0.Cmp(big0) == 0) && (keyShareBig1.Cmp(big0) == 0)
		if !zeroKeyShare {
			t.Fatal("Unexpected error: KeyShare is nonzero and already present")
		}

		// Check Share Distribution Hashes
		authHash, err := c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
		if err != nil {
			t.Fatal("Error when calling ShareDistributionHashes")
		}
		zeroBytes := make([]byte, numBytes)
		validHash := !bytes.Equal(authHash[:], zeroBytes)
		if !validHash {
			t.Fatal("Unexpected error: invalid hash")
		}

		keyShareG1Big := G1ToBigIntArray(keyShareG1)
		keyShareG2Big := G2ToBigIntArray(keyShareG2)

		_, err = c.SubmitKeyShare(txOpt, auth.From, keyShareG1Big, keyShareDLEQProof, keyShareG2Big)
		if err != nil {
			t.Fatal("Unexpected error occurred when submitting key shares")
		}

		keyShareArrayG1[ell] = keyShareG1
		keyShareArrayG2[ell] = keyShareG2
		keyShareArrayDLEQProof[ell] = keyShareDLEQProof
	}
	sim.Commit()

	// Need to check key share submission and confirm validity
	for ell := 0; ell < n; ell++ {
		auth := authArray[ell]
		keyShareG1 := keyShareArrayG1[ell]
		keyShareBig0, err := c.KeyShares(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (0)")
		}
		keyShareBig1, err := c.KeyShares(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (1)")
		}
		keyShareG1Rcvd, err := BigIntArrayToG1([2]*big.Int{keyShareBig0, keyShareBig1})
		if err != nil {
			t.Fatal("Error in BigIntArrayToG1 call")
		}
		if !keyShareG1.IsEqual(keyShareG1Rcvd) {
			t.Fatal("KeyShareG1 mismatch between submission and received!")
		}
	}

	AdvanceBlocksUntil(sim, keyShareSubmissionEnd)
	// Check block number here
	curBlock = sim.Blockchain().CurrentBlock().Number()
	mpkSubmissionEnd, err := c.TMPKSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting MPKSubmissionEnd")
	}
	validBlockNumber = (keyShareSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(mpkSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in MPK Submission Phase")
	}

	// Make Master Public Key (this is not how you would actually do this)
	mpk := new(cloudflare.G2).Add(keyShareArrayG2[0], keyShareArrayG2[1])
	for ell := 2; ell < n; ell++ {
		mpk.Add(mpk, keyShareArrayG2[ell])
	}
	mpkBig := G2ToBigIntArray(mpk)

	// For G1 version
	mpkG1 := new(cloudflare.G1).Add(keyShareArrayG1[0], keyShareArrayG1[1])
	for ell := 2; ell < n; ell++ {
		mpkG1.Add(mpkG1, keyShareArrayG1[ell])
	}

	// Perform PairingCheck on mpk and mpkG1 to ensure valid pair
	validPair := cloudflare.PairingCheck([]*cloudflare.G1{mpkG1, h1Base}, []*cloudflare.G2{h2Neg, mpk})
	if !validPair {
		t.Fatal("Error in PairingCheck for mpk")
	}

	auth := authArray[0]
	txOpt := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	_, err = c.SubmitMasterPublicKey(txOpt, mpkBig)
	if err != nil {
		t.Fatal("Unexpected error occurred when submitting master public key")
	}
	sim.Commit()

	mpkRcvd0, err := c.MasterPublicKey(&bind.CallOpts{}, big0)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (0)")
	}
	mpkRcvd1, err := c.MasterPublicKey(&bind.CallOpts{}, big1)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (1)")
	}
	mpkRcvd2, err := c.MasterPublicKey(&bind.CallOpts{}, big2)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (2)")
	}
	mpkRcvd3, err := c.MasterPublicKey(&bind.CallOpts{}, big3)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (3)")
	}
	mpkSubmittedMatchesRcvd := (mpkBig[0].Cmp(mpkRcvd0) == 0) && (mpkBig[1].Cmp(mpkRcvd1) == 0) && (mpkBig[2].Cmp(mpkRcvd2) == 0) && (mpkBig[3].Cmp(mpkRcvd3) == 0)
	if !mpkSubmittedMatchesRcvd {
		t.Fatal("mpk submitted does not match received!")
	}

	// We now proceed to submit gpkj's; they were created above

	// Check block number here; will fail
	AdvanceBlocksUntil(sim, mpkSubmissionEnd)
	gpkjSubmissionEnd, err := c.TGPKJSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting GPKJSubmissionEnd")
	}
	AdvanceBlocksUntil(sim, gpkjSubmissionEnd)
	curBlock = sim.Blockchain().CurrentBlock().Number()
	validBlockNumber = (mpkSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(gpkjSubmissionEnd) <= 0)
	if validBlockNumber {
		t.Fatal("Unexpected error; not in GPKj Submission Phase")
	}

	initialMessage, err := c.InitialMessage(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error when getting InitialMessage for gpkj signature")
	}

	// Make and submit gpkj's; this will fail because not in correct block
	idx := 0
	gskj := gskjArray[idx]
	gpkj := new(cloudflare.G2).ScalarBaseMult(gskj)
	initialSig, err := cloudflare.Sign(initialMessage, gskj, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in cloudflare.Sign when signing initialMessage")
	}
	gpkjBig := G2ToBigIntArray(gpkj)
	initialSigBig := G1ToBigIntArray(initialSig)

	auth = authArray[idx]
	txOpt = &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	// Ensure no previous submission
	gpkjSubmission0, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission0")
	}
	gpkjSubmission1, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission1")
	}
	gpkjSubmission2, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big2)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission2")
	}
	gpkjSubmission3, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big3)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission3")
	}
	emptyGPKjSub := (gpkjSubmission0.Cmp(big0) == 0) && (gpkjSubmission1.Cmp(big0) == 0) && (gpkjSubmission2.Cmp(big0) == 0) && (gpkjSubmission3.Cmp(big0) == 0)
	if !emptyGPKjSub {
		t.Fatal("Unexpected error; gpkj already submitted")
	}

	// Verify signature
	validSig, err := cloudflare.Verify(initialMessage, initialSig, gpkj, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error when calling cloudflare.Verify for (initialSig, gpkj) verification")
	}
	if !validSig {
		t.Fatal("Unexpected error; initialSig fails cloudflare.Verify")
	}

	_, err = c.SubmitGPKj(txOpt, gpkjBig, initialSigBig)
	if err != nil {
		t.Fatal("Error occurred when submitting gpkj")
	}
	sim.Commit()

	// Confirm no submission occurred
	gpkjRcvd0, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission0")
	}
	gpkjRcvd1, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission1")
	}
	gpkjRcvd2, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big2)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission2")
	}
	gpkjRcvd3, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big3)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission3")
	}
	emptyGPKjRcvd := (gpkjRcvd0.Cmp(big0) == 0) && (gpkjRcvd1.Cmp(big0) == 0) && (gpkjRcvd2.Cmp(big0) == 0) && (gpkjRcvd3.Cmp(big0) == 0)
	if !emptyGPKjRcvd {
		t.Fatal("Unexpected error; gpkj submission should have failed due to block number")
	}
}

// We submit gpkj and then attempt to submit another (different) gpkj;
// this should fail.
func TestSubmitGPKjFailResubmit(t *testing.T) {
	n := 4
	threshold, _ := thresholdFromUsers(n) // threshold, k are return values
	_ = threshold                         // for linter
	//c, sim, keyArray, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	c, _, sim, _, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	defer sim.Close()
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error in getting RegistrationEnd")
	}
	AdvanceBlocksUntil(sim, registrationEnd)
	// Current block number is now 22 > 21 == T_REGISTRATION_END;
	// in Share Distribution phase

	// These are the standard secrets be used for testing purposes
	secretValuesArray := make([]*big.Int, n)
	secretBase := big.NewInt(100)
	for j := 0; j < n; j++ {
		secretValuesArray[j] = new(big.Int).Add(secretBase, big.NewInt(int64(j)))
	}

	// These are the standard private polynomial coefs for testing purposes
	basePrivatePolynomialCoefs := make([]*big.Int, threshold+1)
	for j := 1; j < threshold+1; j++ {
		basePrivatePolynomialCoefs[j] = big.NewInt(int64(j))
	}

	// Create private polynomials for all users
	privPolyCoefsArray := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		privPolyCoefsArray[ell] = make([]*big.Int, threshold+1)
		privPolyCoefsArray[ell][0] = secretValuesArray[ell]
		for j := 1; j < threshold+1; j++ {
			privPolyCoefsArray[ell][j] = basePrivatePolynomialCoefs[j]
		}
	}

	// Create public coefficients for all users
	pubCoefsArray := make([][]*cloudflare.G1, n)
	for ell := 0; ell < n; ell++ {
		pubCoefsArray[ell] = make([]*cloudflare.G1, threshold+1)
		pubCoefsArray[ell] = cloudflare.GeneratePublicCoefs(privPolyCoefsArray[ell])
	}

	// Create big.Int version of public coefficients
	pubCoefsBigArray := make([][][2]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		pubCoefsBigArray[ell] = make([][2]*big.Int, threshold+1)
		for j := 0; j < threshold+1; j++ {
			pubCoefsBigArray[ell][j] = G1ToBigIntArray(pubCoefsArray[ell][j])
		}
	}

	// Create encrypted shares to submit
	encSharesArray := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		privK := privKArray[ell]
		pubK := pubKArray[ell]
		encSharesArray[ell] = make([]*big.Int, n-1)
		secretsArray, err := cloudflare.GenerateSecretShares(pubK, privPolyCoefsArray[ell], pubKArray)
		if err != nil {
			t.Fatal("Error occurred while generating sharing secrets")
		}
		encSharesArray[ell], err = cloudflare.GenerateEncryptedShares(secretsArray, privK, pubKArray)
		if err != nil {
			t.Fatal("Error occurred while generating commitments")
		}
	}

	// Create arrays to hold submitted information
	// First index is participant receiving (n), then who from (n), then values (n-1);
	// note that this would have to be changed in practice
	rcvdEncShares := make([][][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		rcvdEncShares[ell] = make([][]*big.Int, n)
		for j := 0; j < n; j++ {
			rcvdEncShares[ell][j] = make([]*big.Int, n-1)
		}
	}
	rcvdCommitments := make([][][][2]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		rcvdCommitments[ell] = make([][][2]*big.Int, n)
		for j := 0; j < n; j++ {
			rcvdCommitments[ell][j] = make([][2]*big.Int, threshold+1)
		}
	}

	big0 := big.NewInt(0)
	big1 := big.NewInt(1)
	big2 := big.NewInt(2)
	big3 := big.NewInt(3)

	// Submit encrypted shares and commitments
	for ell := 0; ell < n; ell++ {
		auth := authArray[ell]
		encShares := encSharesArray[ell]
		pubCoefs := pubCoefsBigArray[ell]
		txOpt := &bind.TransactOpts{
			From:     auth.From,
			Nonce:    nil,
			Signer:   auth.Signer,
			Value:    nil,
			GasPrice: nil,
			GasLimit: gasLim,
			Context:  nil,
		}
		// Check public_key to ensure registered
		pubKBigRcvd0, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (0)")
		}
		pubKBigRcvd1, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (1)")
		}
		registeredPubK := (pubKBigRcvd0.Cmp(big0) != 0) || (pubKBigRcvd1.Cmp(big0) != 0)
		if !registeredPubK {
			t.Fatal("Public Key already exists")
		}
		txn, err := c.DistributeShares(txOpt, encShares, pubCoefs)
		if err != nil {
			t.Fatal("Unexpected error arose in DistributeShares submission")
		}
		sim.Commit()
		receipt, err := sim.TransactionReceipt(context.Background(), txn.Hash())
		if err != nil {
			t.Fatal("Unexpected error in TransactionReceipt")
		}
		shareDistEvent, err := c.ETHDKGFilterer.ParseShareDistribution(*receipt.Logs[0])
		if err != nil {
			t.Fatal("Unexpected error in ParseShareDistribution")
		}
		// Save values in arrays for everyone
		for j := 0; j < n; j++ {
			if j == ell {
				continue
			}
			rcvdEncShares[j][ell] = shareDistEvent.EncryptedShares
			rcvdCommitments[j][ell] = shareDistEvent.Commitments
		}
	}
	// Everything above is good but now we want to check stuff like events and logs

	ssArrayAll := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		ssArrayAll[ell] = make([]*big.Int, n-1)
	}

	// HERE WE GO
	for ell := 0; ell < n; ell++ {
		rcvdEncSharesEll := rcvdEncShares[ell]
		pubK := pubKArray[ell]
		privK := privKArray[ell]
		sharedEncryptedArray, err := cloudflare.CondenseCommitments(pubK, rcvdEncSharesEll, pubKArray)
		if err != nil {
			t.Fatal("Unexpected error occurred when condensing commitments")
		}
		sharedSecretsArray, err := cloudflare.GenerateDecryptedShares(privK, sharedEncryptedArray, pubKArray)
		if err != nil {
			t.Fatal("Unexpected error occurred when decrypting secrets")
		}
		ssArrayAll[ell] = sharedSecretsArray
	}

	gskjArray := make([]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		sharedSecretsArray := ssArrayAll[ell]
		privPolyCoefs := privPolyCoefsArray[ell]
		idx := ell + 1
		selfSecret := cloudflare.PrivatePolyEval(privPolyCoefs, idx)
		gskj := new(big.Int).Set(selfSecret)
		for j := 0; j < n-1; j++ {
			sharedSecret := sharedSecretsArray[j]
			gskj.Add(gskj, sharedSecret)
		}
		gskjArray[ell] = gskj
	}

	shareDistributionEnd, err := c.TSHAREDISTRIBUTIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting ShareDistributionEnd")
	}
	AdvanceBlocksUntil(sim, shareDistributionEnd)
	// Current block number is now 47 > 46 == T_SHARE_DISTRIBUTION_END;
	// in Dispute phase
	disputeEnd, err := c.TDISPUTEEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting DisputeEnd")
	}
	AdvanceBlocksUntil(sim, disputeEnd)
	// Current block number is now 72 > 71 == T_DISPUTE_END;
	// in Key Derivation phase

	// Check block number here
	curBlock := sim.Blockchain().CurrentBlock().Number()
	keyShareSubmissionEnd, err := c.TKEYSHARESUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting KeyShareSubmissionEnd")
	}
	validBlockNumber := (disputeEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(keyShareSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in Key Share Submission Phase")
	}

	// Now to submit key shares
	keyShareArrayG1 := make([]*cloudflare.G1, n)
	keyShareArrayG2 := make([]*cloudflare.G2, n)
	keyShareArrayDLEQProof := make([][2]*big.Int, n)

	h1BaseMsg := []byte("MadHive Rocks!")
	g1Base := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	h1Base, err := cloudflare.HashToG1(h1BaseMsg)
	if err != nil {
		t.Fatal("Error when computing HashToG1([]byte(\"MadHive Rock!\"))")
	}
	h2Base := new(cloudflare.G2).ScalarBaseMult(big.NewInt(1))
	orderMinus1, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495616", 10)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	for ell := 0; ell < n; ell++ {
		secretValue := secretValuesArray[ell]
		g1Value := new(cloudflare.G1).ScalarBaseMult(secretValue)
		keyShareG1 := new(cloudflare.G1).ScalarMult(h1Base, secretValue)
		keyShareG2 := new(cloudflare.G2).ScalarMult(h2Base, secretValue)

		// Generate and Verify DLEQ Proof
		keyShareDLEQProof, err := cloudflare.GenerateDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, secretValue, rand.Reader)
		if err != nil {
			t.Fatal("Error when generating DLEQ Proof")
		}
		err = cloudflare.VerifyDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, keyShareDLEQProof)
		if err != nil {
			t.Fatal("Invalid DLEQ h1Value proof")
		}

		// PairingCheck to ensure keyShareG1 and keyShareG2 form valid pair
		validPair := cloudflare.PairingCheck([]*cloudflare.G1{keyShareG1, h1Base}, []*cloudflare.G2{h2Neg, keyShareG2})
		if !validPair {
			t.Fatal("Error in PairingCheck")
		}

		auth := authArray[ell]
		txOpt := &bind.TransactOpts{
			From:     auth.From,
			Nonce:    nil,
			Signer:   auth.Signer,
			Value:    nil,
			GasPrice: nil,
			GasLimit: gasLim,
			Context:  nil,
		}

		// Check Key Shares to ensure not submitted
		keyShareBig0, err := c.KeyShares(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (0)")
		}
		keyShareBig1, err := c.KeyShares(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (1)")
		}
		zeroKeyShare := (keyShareBig0.Cmp(big0) == 0) && (keyShareBig1.Cmp(big0) == 0)
		if !zeroKeyShare {
			t.Fatal("Unexpected error: KeyShare is nonzero and already present")
		}

		// Check Share Distribution Hashes
		authHash, err := c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
		if err != nil {
			t.Fatal("Error when calling ShareDistributionHashes")
		}
		zeroBytes := make([]byte, numBytes)
		validHash := !bytes.Equal(authHash[:], zeroBytes)
		if !validHash {
			t.Fatal("Unexpected error: invalid hash")
		}

		keyShareG1Big := G1ToBigIntArray(keyShareG1)
		keyShareG2Big := G2ToBigIntArray(keyShareG2)

		_, err = c.SubmitKeyShare(txOpt, auth.From, keyShareG1Big, keyShareDLEQProof, keyShareG2Big)
		if err != nil {
			t.Fatal("Unexpected error occurred when submitting key shares")
		}

		keyShareArrayG1[ell] = keyShareG1
		keyShareArrayG2[ell] = keyShareG2
		keyShareArrayDLEQProof[ell] = keyShareDLEQProof
	}
	sim.Commit()

	// Need to check key share submission and confirm validity
	for ell := 0; ell < n; ell++ {
		auth := authArray[ell]
		keyShareG1 := keyShareArrayG1[ell]
		keyShareBig0, err := c.KeyShares(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (0)")
		}
		keyShareBig1, err := c.KeyShares(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (1)")
		}
		keyShareG1Rcvd, err := BigIntArrayToG1([2]*big.Int{keyShareBig0, keyShareBig1})
		if err != nil {
			t.Fatal("Error in BigIntArrayToG1 call")
		}
		if !keyShareG1.IsEqual(keyShareG1Rcvd) {
			t.Fatal("KeyShareG1 mismatch between submission and received!")
		}
	}

	AdvanceBlocksUntil(sim, keyShareSubmissionEnd)
	// Check block number here
	curBlock = sim.Blockchain().CurrentBlock().Number()
	mpkSubmissionEnd, err := c.TMPKSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting MPKSubmissionEnd")
	}
	validBlockNumber = (keyShareSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(mpkSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in MPK Submission Phase")
	}

	// Make Master Public Key (this is not how you would actually do this)
	mpk := new(cloudflare.G2).Add(keyShareArrayG2[0], keyShareArrayG2[1])
	for ell := 2; ell < n; ell++ {
		mpk.Add(mpk, keyShareArrayG2[ell])
	}
	mpkBig := G2ToBigIntArray(mpk)

	// For G1 version
	mpkG1 := new(cloudflare.G1).Add(keyShareArrayG1[0], keyShareArrayG1[1])
	for ell := 2; ell < n; ell++ {
		mpkG1.Add(mpkG1, keyShareArrayG1[ell])
	}

	// Perform PairingCheck on mpk and mpkG1 to ensure valid pair
	validPair := cloudflare.PairingCheck([]*cloudflare.G1{mpkG1, h1Base}, []*cloudflare.G2{h2Neg, mpk})
	if !validPair {
		t.Fatal("Error in PairingCheck for mpk")
	}

	auth := authArray[0]
	txOpt := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	_, err = c.SubmitMasterPublicKey(txOpt, mpkBig)
	if err != nil {
		t.Fatal("Unexpected error occurred when submitting master public key")
	}
	sim.Commit()

	mpkRcvd0, err := c.MasterPublicKey(&bind.CallOpts{}, big0)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (0)")
	}
	mpkRcvd1, err := c.MasterPublicKey(&bind.CallOpts{}, big1)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (1)")
	}
	mpkRcvd2, err := c.MasterPublicKey(&bind.CallOpts{}, big2)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (2)")
	}
	mpkRcvd3, err := c.MasterPublicKey(&bind.CallOpts{}, big3)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (3)")
	}
	mpkSubmittedMatchesRcvd := (mpkBig[0].Cmp(mpkRcvd0) == 0) && (mpkBig[1].Cmp(mpkRcvd1) == 0) && (mpkBig[2].Cmp(mpkRcvd2) == 0) && (mpkBig[3].Cmp(mpkRcvd3) == 0)
	if !mpkSubmittedMatchesRcvd {
		t.Fatal("mpk submitted does not match received!")
	}

	// We now proceed to submit gpkj's; they were created above

	// Check block number here
	AdvanceBlocksUntil(sim, mpkSubmissionEnd)
	gpkjSubmissionEnd, err := c.TGPKJSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting GPKJSubmissionEnd")
	}
	curBlock = sim.Blockchain().CurrentBlock().Number()
	validBlockNumber = (mpkSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(gpkjSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in GPKj Submission Phase")
	}

	initialMessage, err := c.InitialMessage(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error when getting InitialMessage for gpkj signature")
	}

	// Make and submit gpkj's
	idx := 0
	gskj := gskjArray[idx]
	gpkj := new(cloudflare.G2).ScalarBaseMult(gskj)
	initialSig, err := cloudflare.Sign(initialMessage, gskj, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in cloudflare.Sign when signing initialMessage")
	}
	gpkjBig := G2ToBigIntArray(gpkj)
	initialSigBig := G1ToBigIntArray(initialSig)

	auth = authArray[idx]
	txOpt = &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	// Ensure no previous submission
	gpkjSubmission0, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission0")
	}
	gpkjSubmission1, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission1")
	}
	gpkjSubmission2, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big2)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission2")
	}
	gpkjSubmission3, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big3)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission3")
	}
	emptyGPKjSub := (gpkjSubmission0.Cmp(big0) == 0) && (gpkjSubmission1.Cmp(big0) == 0) && (gpkjSubmission2.Cmp(big0) == 0) && (gpkjSubmission3.Cmp(big0) == 0)
	if !emptyGPKjSub {
		t.Fatal("Unexpected error; gpkj already submitted")
	}

	// Verify signature
	validSig, err := cloudflare.Verify(initialMessage, initialSig, gpkj, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error when calling cloudflare.Verify for (initialSig, gpkj) verification")
	}
	if !validSig {
		t.Fatal("Unexpected error; initialSig fails cloudflare.Verify")
	}

	_, err = c.SubmitGPKj(txOpt, gpkjBig, initialSigBig)
	if err != nil {
		t.Fatal("Error occurred when submitting gpkj")
	}
	sim.Commit()

	// Check submission
	gpkjRcvd0, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission0")
	}
	gpkjRcvd1, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission1")
	}
	gpkjRcvd2, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big2)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission2")
	}
	gpkjRcvd3, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big3)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission3")
	}
	matchGPKjSub := (gpkjRcvd0.Cmp(gpkjBig[0]) == 0) && (gpkjRcvd1.Cmp(gpkjBig[1]) == 0) && (gpkjRcvd2.Cmp(gpkjBig[2]) == 0) && (gpkjRcvd3.Cmp(gpkjBig[3]) == 0)
	if !matchGPKjSub {
		t.Fatal("Unexpected error; gpkj does not match submission")
	}

	// Make and submit (new) gpkj
	gskjNew := new(big.Int).Add(gskj, big1)
	gpkjNew := new(cloudflare.G2).ScalarBaseMult(gskjNew)
	initialSigNew, err := cloudflare.Sign(initialMessage, gskjNew, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error occurred in cloudflare.Sign when signing initialMessage (new)")
	}
	gpkjNewBig := G2ToBigIntArray(gpkjNew)
	initialSigNewBig := G1ToBigIntArray(initialSigNew)

	_, err = c.SubmitGPKj(txOpt, gpkjNewBig, initialSigNewBig)
	if err != nil {
		t.Fatal("Error occurred when submitting gpkj")
	}
	sim.Commit()

	// Attempt to resubmit; should fail
	gpkjRcvd0, err = c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission0")
	}
	gpkjRcvd1, err = c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission1")
	}
	gpkjRcvd2, err = c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big2)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission2")
	}
	gpkjRcvd3, err = c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big3)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission3")
	}
	matchGPKjNewSub := (gpkjRcvd0.Cmp(gpkjNewBig[0]) == 0) && (gpkjRcvd1.Cmp(gpkjNewBig[1]) == 0) && (gpkjRcvd2.Cmp(gpkjNewBig[2]) == 0) && (gpkjRcvd3.Cmp(gpkjNewBig[3]) == 0)
	if matchGPKjNewSub {
		t.Fatal("Unexpected error; gpkj should not match resubmission")
	}
}

// Attempt to submit false gpkj signature with signature off-curve.
func TestSubmitGPKjFailOffCurveSignature(t *testing.T) {
	n := 4
	threshold, _ := thresholdFromUsers(n) // threshold, k are return values
	_ = threshold                         // for linter
	//c, sim, keyArray, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	c, _, sim, _, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	defer sim.Close()
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error in getting RegistrationEnd")
	}
	AdvanceBlocksUntil(sim, registrationEnd)
	// Current block number is now 22 > 21 == T_REGISTRATION_END;
	// in Share Distribution phase

	// These are the standard secrets be used for testing purposes
	secretValuesArray := make([]*big.Int, n)
	secretBase := big.NewInt(100)
	for j := 0; j < n; j++ {
		secretValuesArray[j] = new(big.Int).Add(secretBase, big.NewInt(int64(j)))
	}

	// These are the standard private polynomial coefs for testing purposes
	basePrivatePolynomialCoefs := make([]*big.Int, threshold+1)
	for j := 1; j < threshold+1; j++ {
		basePrivatePolynomialCoefs[j] = big.NewInt(int64(j))
	}

	// Create private polynomials for all users
	privPolyCoefsArray := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		privPolyCoefsArray[ell] = make([]*big.Int, threshold+1)
		privPolyCoefsArray[ell][0] = secretValuesArray[ell]
		for j := 1; j < threshold+1; j++ {
			privPolyCoefsArray[ell][j] = basePrivatePolynomialCoefs[j]
		}
	}

	// Create public coefficients for all users
	pubCoefsArray := make([][]*cloudflare.G1, n)
	for ell := 0; ell < n; ell++ {
		pubCoefsArray[ell] = make([]*cloudflare.G1, threshold+1)
		pubCoefsArray[ell] = cloudflare.GeneratePublicCoefs(privPolyCoefsArray[ell])
	}

	// Create big.Int version of public coefficients
	pubCoefsBigArray := make([][][2]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		pubCoefsBigArray[ell] = make([][2]*big.Int, threshold+1)
		for j := 0; j < threshold+1; j++ {
			pubCoefsBigArray[ell][j] = G1ToBigIntArray(pubCoefsArray[ell][j])
		}
	}

	// Create encrypted shares to submit
	encSharesArray := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		privK := privKArray[ell]
		pubK := pubKArray[ell]
		encSharesArray[ell] = make([]*big.Int, n-1)
		secretsArray, err := cloudflare.GenerateSecretShares(pubK, privPolyCoefsArray[ell], pubKArray)
		if err != nil {
			t.Fatal("Error occurred while generating sharing secrets")
		}
		encSharesArray[ell], err = cloudflare.GenerateEncryptedShares(secretsArray, privK, pubKArray)
		if err != nil {
			t.Fatal("Error occurred while generating commitments")
		}
	}

	// Create arrays to hold submitted information
	// First index is participant receiving (n), then who from (n), then values (n-1);
	// note that this would have to be changed in practice
	rcvdEncShares := make([][][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		rcvdEncShares[ell] = make([][]*big.Int, n)
		for j := 0; j < n; j++ {
			rcvdEncShares[ell][j] = make([]*big.Int, n-1)
		}
	}
	rcvdCommitments := make([][][][2]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		rcvdCommitments[ell] = make([][][2]*big.Int, n)
		for j := 0; j < n; j++ {
			rcvdCommitments[ell][j] = make([][2]*big.Int, threshold+1)
		}
	}

	big0 := big.NewInt(0)
	big1 := big.NewInt(1)
	big2 := big.NewInt(2)
	big3 := big.NewInt(3)

	// Submit encrypted shares and commitments
	for ell := 0; ell < n; ell++ {
		auth := authArray[ell]
		encShares := encSharesArray[ell]
		pubCoefs := pubCoefsBigArray[ell]
		txOpt := &bind.TransactOpts{
			From:     auth.From,
			Nonce:    nil,
			Signer:   auth.Signer,
			Value:    nil,
			GasPrice: nil,
			GasLimit: gasLim,
			Context:  nil,
		}
		// Check public_key to ensure registered
		pubKBigRcvd0, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (0)")
		}
		pubKBigRcvd1, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (1)")
		}
		registeredPubK := (pubKBigRcvd0.Cmp(big0) != 0) || (pubKBigRcvd1.Cmp(big0) != 0)
		if !registeredPubK {
			t.Fatal("Public Key already exists")
		}
		txn, err := c.DistributeShares(txOpt, encShares, pubCoefs)
		if err != nil {
			t.Fatal("Unexpected error arose in DistributeShares submission")
		}
		sim.Commit()
		receipt, err := sim.TransactionReceipt(context.Background(), txn.Hash())
		if err != nil {
			t.Fatal("Unexpected error in TransactionReceipt")
		}
		shareDistEvent, err := c.ETHDKGFilterer.ParseShareDistribution(*receipt.Logs[0])
		if err != nil {
			t.Fatal("Unexpected error in ParseShareDistribution")
		}
		// Save values in arrays for everyone
		for j := 0; j < n; j++ {
			if j == ell {
				continue
			}
			rcvdEncShares[j][ell] = shareDistEvent.EncryptedShares
			rcvdCommitments[j][ell] = shareDistEvent.Commitments
		}
	}
	// Everything above is good but now we want to check stuff like events and logs

	ssArrayAll := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		ssArrayAll[ell] = make([]*big.Int, n-1)
	}

	// HERE WE GO
	for ell := 0; ell < n; ell++ {
		rcvdEncSharesEll := rcvdEncShares[ell]
		pubK := pubKArray[ell]
		privK := privKArray[ell]
		sharedEncryptedArray, err := cloudflare.CondenseCommitments(pubK, rcvdEncSharesEll, pubKArray)
		if err != nil {
			t.Fatal("Unexpected error occurred when condensing commitments")
		}
		sharedSecretsArray, err := cloudflare.GenerateDecryptedShares(privK, sharedEncryptedArray, pubKArray)
		if err != nil {
			t.Fatal("Unexpected error occurred when decrypting secrets")
		}
		ssArrayAll[ell] = sharedSecretsArray
	}

	gskjArray := make([]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		sharedSecretsArray := ssArrayAll[ell]
		privPolyCoefs := privPolyCoefsArray[ell]
		idx := ell + 1
		selfSecret := cloudflare.PrivatePolyEval(privPolyCoefs, idx)
		gskj := new(big.Int).Set(selfSecret)
		for j := 0; j < n-1; j++ {
			sharedSecret := sharedSecretsArray[j]
			gskj.Add(gskj, sharedSecret)
		}
		gskjArray[ell] = gskj
	}

	shareDistributionEnd, err := c.TSHAREDISTRIBUTIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting ShareDistributionEnd")
	}
	AdvanceBlocksUntil(sim, shareDistributionEnd)
	// Current block number is now 47 > 46 == T_SHARE_DISTRIBUTION_END;
	// in Dispute phase
	disputeEnd, err := c.TDISPUTEEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting DisputeEnd")
	}
	AdvanceBlocksUntil(sim, disputeEnd)
	// Current block number is now 72 > 71 == T_DISPUTE_END;
	// in Key Derivation phase

	// Check block number here
	curBlock := sim.Blockchain().CurrentBlock().Number()
	keyShareSubmissionEnd, err := c.TKEYSHARESUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting KeyShareSubmissionEnd")
	}
	validBlockNumber := (disputeEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(keyShareSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in Key Share Submission Phase")
	}

	// Now to submit key shares
	keyShareArrayG1 := make([]*cloudflare.G1, n)
	keyShareArrayG2 := make([]*cloudflare.G2, n)
	keyShareArrayDLEQProof := make([][2]*big.Int, n)

	h1BaseMsg := []byte("MadHive Rocks!")
	g1Base := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	h1Base, err := cloudflare.HashToG1(h1BaseMsg)
	if err != nil {
		t.Fatal("Error when computing HashToG1([]byte(\"MadHive Rock!\"))")
	}
	h2Base := new(cloudflare.G2).ScalarBaseMult(big.NewInt(1))
	orderMinus1, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495616", 10)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	for ell := 0; ell < n; ell++ {
		secretValue := secretValuesArray[ell]
		g1Value := new(cloudflare.G1).ScalarBaseMult(secretValue)
		keyShareG1 := new(cloudflare.G1).ScalarMult(h1Base, secretValue)
		keyShareG2 := new(cloudflare.G2).ScalarMult(h2Base, secretValue)

		// Generate and Verify DLEQ Proof
		keyShareDLEQProof, err := cloudflare.GenerateDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, secretValue, rand.Reader)
		if err != nil {
			t.Fatal("Error when generating DLEQ Proof")
		}
		err = cloudflare.VerifyDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, keyShareDLEQProof)
		if err != nil {
			t.Fatal("Invalid DLEQ h1Value proof")
		}

		// PairingCheck to ensure keyShareG1 and keyShareG2 form valid pair
		validPair := cloudflare.PairingCheck([]*cloudflare.G1{keyShareG1, h1Base}, []*cloudflare.G2{h2Neg, keyShareG2})
		if !validPair {
			t.Fatal("Error in PairingCheck")
		}

		auth := authArray[ell]
		txOpt := &bind.TransactOpts{
			From:     auth.From,
			Nonce:    nil,
			Signer:   auth.Signer,
			Value:    nil,
			GasPrice: nil,
			GasLimit: gasLim,
			Context:  nil,
		}

		// Check Key Shares to ensure not submitted
		keyShareBig0, err := c.KeyShares(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (0)")
		}
		keyShareBig1, err := c.KeyShares(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (1)")
		}
		zeroKeyShare := (keyShareBig0.Cmp(big0) == 0) && (keyShareBig1.Cmp(big0) == 0)
		if !zeroKeyShare {
			t.Fatal("Unexpected error: KeyShare is nonzero and already present")
		}

		// Check Share Distribution Hashes
		authHash, err := c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
		if err != nil {
			t.Fatal("Error when calling ShareDistributionHashes")
		}
		zeroBytes := make([]byte, numBytes)
		validHash := !bytes.Equal(authHash[:], zeroBytes)
		if !validHash {
			t.Fatal("Unexpected error: invalid hash")
		}

		keyShareG1Big := G1ToBigIntArray(keyShareG1)
		keyShareG2Big := G2ToBigIntArray(keyShareG2)

		_, err = c.SubmitKeyShare(txOpt, auth.From, keyShareG1Big, keyShareDLEQProof, keyShareG2Big)
		if err != nil {
			t.Fatal("Unexpected error occurred when submitting key shares")
		}

		keyShareArrayG1[ell] = keyShareG1
		keyShareArrayG2[ell] = keyShareG2
		keyShareArrayDLEQProof[ell] = keyShareDLEQProof
	}
	sim.Commit()

	// Need to check key share submission and confirm validity
	for ell := 0; ell < n; ell++ {
		auth := authArray[ell]
		keyShareG1 := keyShareArrayG1[ell]
		keyShareBig0, err := c.KeyShares(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (0)")
		}
		keyShareBig1, err := c.KeyShares(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (1)")
		}
		keyShareG1Rcvd, err := BigIntArrayToG1([2]*big.Int{keyShareBig0, keyShareBig1})
		if err != nil {
			t.Fatal("Error in BigIntArrayToG1 call")
		}
		if !keyShareG1.IsEqual(keyShareG1Rcvd) {
			t.Fatal("KeyShareG1 mismatch between submission and received!")
		}
	}

	AdvanceBlocksUntil(sim, keyShareSubmissionEnd)
	// Check block number here
	curBlock = sim.Blockchain().CurrentBlock().Number()
	mpkSubmissionEnd, err := c.TMPKSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting MPKSubmissionEnd")
	}
	validBlockNumber = (keyShareSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(mpkSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in MPK Submission Phase")
	}

	// Make Master Public Key (this is not how you would actually do this)
	mpk := new(cloudflare.G2).Add(keyShareArrayG2[0], keyShareArrayG2[1])
	for ell := 2; ell < n; ell++ {
		mpk.Add(mpk, keyShareArrayG2[ell])
	}
	mpkBig := G2ToBigIntArray(mpk)

	// For G1 version
	mpkG1 := new(cloudflare.G1).Add(keyShareArrayG1[0], keyShareArrayG1[1])
	for ell := 2; ell < n; ell++ {
		mpkG1.Add(mpkG1, keyShareArrayG1[ell])
	}

	// Perform PairingCheck on mpk and mpkG1 to ensure valid pair
	validPair := cloudflare.PairingCheck([]*cloudflare.G1{mpkG1, h1Base}, []*cloudflare.G2{h2Neg, mpk})
	if !validPair {
		t.Fatal("Error in PairingCheck for mpk")
	}

	auth := authArray[0]
	txOpt := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	_, err = c.SubmitMasterPublicKey(txOpt, mpkBig)
	if err != nil {
		t.Fatal("Unexpected error occurred when submitting master public key")
	}
	sim.Commit()

	mpkRcvd0, err := c.MasterPublicKey(&bind.CallOpts{}, big0)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (0)")
	}
	mpkRcvd1, err := c.MasterPublicKey(&bind.CallOpts{}, big1)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (1)")
	}
	mpkRcvd2, err := c.MasterPublicKey(&bind.CallOpts{}, big2)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (2)")
	}
	mpkRcvd3, err := c.MasterPublicKey(&bind.CallOpts{}, big3)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (3)")
	}
	mpkSubmittedMatchesRcvd := (mpkBig[0].Cmp(mpkRcvd0) == 0) && (mpkBig[1].Cmp(mpkRcvd1) == 0) && (mpkBig[2].Cmp(mpkRcvd2) == 0) && (mpkBig[3].Cmp(mpkRcvd3) == 0)
	if !mpkSubmittedMatchesRcvd {
		t.Fatal("mpk submitted does not match received!")
	}

	// We now proceed to submit gpkj's; they were created above

	// Check block number here
	AdvanceBlocksUntil(sim, mpkSubmissionEnd)
	gpkjSubmissionEnd, err := c.TGPKJSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting GPKJSubmissionEnd")
	}
	curBlock = sim.Blockchain().CurrentBlock().Number()
	validBlockNumber = (mpkSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(gpkjSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in GPKj Submission Phase")
	}

	// Make and submit gpkj's; signature is invalid and not on curve
	idx := 0
	gskj := gskjArray[idx]
	gpkj := new(cloudflare.G2).ScalarBaseMult(gskj)
	gpkjBig := G2ToBigIntArray(gpkj)
	initialSigBadBig := [2]*big.Int{big.NewInt(1), big.NewInt(3)}

	auth = authArray[idx]
	txOpt = &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	// Ensure no previous submission
	gpkjSubmission0, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission0")
	}
	gpkjSubmission1, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission1")
	}
	gpkjSubmission2, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big2)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission2")
	}
	gpkjSubmission3, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big3)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission3")
	}
	emptyGPKjSub := (gpkjSubmission0.Cmp(big0) == 0) && (gpkjSubmission1.Cmp(big0) == 0) && (gpkjSubmission2.Cmp(big0) == 0) && (gpkjSubmission3.Cmp(big0) == 0)
	if !emptyGPKjSub {
		t.Fatal("Unexpected error; gpkj already submitted")
	}

	// Submit gpkj; should fail due to invalid (off-curve) signature
	_, err = c.SubmitGPKj(txOpt, gpkjBig, initialSigBadBig)
	if err != nil {
		t.Fatal("Error occurred when submitting gpkj")
	}
	sim.Commit()

	// Ensure submission failed
	gpkjRcvd0, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission0")
	}
	gpkjRcvd1, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission1")
	}
	gpkjRcvd2, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big2)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission2")
	}
	gpkjRcvd3, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big3)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission3")
	}
	emptyGPKjRcvd := (gpkjRcvd0.Cmp(big0) == 0) && (gpkjRcvd1.Cmp(big0) == 0) && (gpkjRcvd2.Cmp(big0) == 0) && (gpkjRcvd3.Cmp(big0) == 0)
	if !emptyGPKjRcvd {
		t.Fatal("Unexpected error; gpkj should not have been submitted because invalid signature")
	}
}

// Attempt to submit gpkj and fail because gpkj and submitted signature
// are not a valid pair.
func TestSubmitGPKjFailInvalidSignature(t *testing.T) {
	n := 4
	threshold, _ := thresholdFromUsers(n) // threshold, k are return values
	_ = threshold                         // for linter
	//c, sim, keyArray, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	c, _, sim, _, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	defer sim.Close()
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error in getting RegistrationEnd")
	}
	AdvanceBlocksUntil(sim, registrationEnd)
	// Current block number is now 22 > 21 == T_REGISTRATION_END;
	// in Share Distribution phase

	// These are the standard secrets be used for testing purposes
	secretValuesArray := make([]*big.Int, n)
	secretBase := big.NewInt(100)
	for j := 0; j < n; j++ {
		secretValuesArray[j] = new(big.Int).Add(secretBase, big.NewInt(int64(j)))
	}

	// These are the standard private polynomial coefs for testing purposes
	basePrivatePolynomialCoefs := make([]*big.Int, threshold+1)
	for j := 1; j < threshold+1; j++ {
		basePrivatePolynomialCoefs[j] = big.NewInt(int64(j))
	}

	// Create private polynomials for all users
	privPolyCoefsArray := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		privPolyCoefsArray[ell] = make([]*big.Int, threshold+1)
		privPolyCoefsArray[ell][0] = secretValuesArray[ell]
		for j := 1; j < threshold+1; j++ {
			privPolyCoefsArray[ell][j] = basePrivatePolynomialCoefs[j]
		}
	}

	// Create public coefficients for all users
	pubCoefsArray := make([][]*cloudflare.G1, n)
	for ell := 0; ell < n; ell++ {
		pubCoefsArray[ell] = make([]*cloudflare.G1, threshold+1)
		pubCoefsArray[ell] = cloudflare.GeneratePublicCoefs(privPolyCoefsArray[ell])
	}

	// Create big.Int version of public coefficients
	pubCoefsBigArray := make([][][2]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		pubCoefsBigArray[ell] = make([][2]*big.Int, threshold+1)
		for j := 0; j < threshold+1; j++ {
			pubCoefsBigArray[ell][j] = G1ToBigIntArray(pubCoefsArray[ell][j])
		}
	}

	// Create encrypted shares to submit
	encSharesArray := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		privK := privKArray[ell]
		pubK := pubKArray[ell]
		encSharesArray[ell] = make([]*big.Int, n-1)
		secretsArray, err := cloudflare.GenerateSecretShares(pubK, privPolyCoefsArray[ell], pubKArray)
		if err != nil {
			t.Fatal("Error occurred while generating sharing secrets")
		}
		encSharesArray[ell], err = cloudflare.GenerateEncryptedShares(secretsArray, privK, pubKArray)
		if err != nil {
			t.Fatal("Error occurred while generating commitments")
		}
	}

	// Create arrays to hold submitted information
	// First index is participant receiving (n), then who from (n), then values (n-1);
	// note that this would have to be changed in practice
	rcvdEncShares := make([][][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		rcvdEncShares[ell] = make([][]*big.Int, n)
		for j := 0; j < n; j++ {
			rcvdEncShares[ell][j] = make([]*big.Int, n-1)
		}
	}
	rcvdCommitments := make([][][][2]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		rcvdCommitments[ell] = make([][][2]*big.Int, n)
		for j := 0; j < n; j++ {
			rcvdCommitments[ell][j] = make([][2]*big.Int, threshold+1)
		}
	}

	big0 := big.NewInt(0)
	big1 := big.NewInt(1)
	big2 := big.NewInt(2)
	big3 := big.NewInt(3)

	// Submit encrypted shares and commitments
	for ell := 0; ell < n; ell++ {
		auth := authArray[ell]
		encShares := encSharesArray[ell]
		pubCoefs := pubCoefsBigArray[ell]
		txOpt := &bind.TransactOpts{
			From:     auth.From,
			Nonce:    nil,
			Signer:   auth.Signer,
			Value:    nil,
			GasPrice: nil,
			GasLimit: gasLim,
			Context:  nil,
		}
		// Check public_key to ensure registered
		pubKBigRcvd0, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (0)")
		}
		pubKBigRcvd1, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (1)")
		}
		registeredPubK := (pubKBigRcvd0.Cmp(big0) != 0) || (pubKBigRcvd1.Cmp(big0) != 0)
		if !registeredPubK {
			t.Fatal("Public Key already exists")
		}
		txn, err := c.DistributeShares(txOpt, encShares, pubCoefs)
		if err != nil {
			t.Fatal("Unexpected error arose in DistributeShares submission")
		}
		sim.Commit()
		receipt, err := sim.TransactionReceipt(context.Background(), txn.Hash())
		if err != nil {
			t.Fatal("Unexpected error in TransactionReceipt")
		}
		shareDistEvent, err := c.ETHDKGFilterer.ParseShareDistribution(*receipt.Logs[0])
		if err != nil {
			t.Fatal("Unexpected error in ParseShareDistribution")
		}
		// Save values in arrays for everyone
		for j := 0; j < n; j++ {
			if j == ell {
				continue
			}
			rcvdEncShares[j][ell] = shareDistEvent.EncryptedShares
			rcvdCommitments[j][ell] = shareDistEvent.Commitments
		}
	}
	// Everything above is good but now we want to check stuff like events and logs

	ssArrayAll := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		ssArrayAll[ell] = make([]*big.Int, n-1)
	}

	// HERE WE GO
	for ell := 0; ell < n; ell++ {
		rcvdEncSharesEll := rcvdEncShares[ell]
		pubK := pubKArray[ell]
		privK := privKArray[ell]
		sharedEncryptedArray, err := cloudflare.CondenseCommitments(pubK, rcvdEncSharesEll, pubKArray)
		if err != nil {
			t.Fatal("Unexpected error occurred when condensing commitments")
		}
		sharedSecretsArray, err := cloudflare.GenerateDecryptedShares(privK, sharedEncryptedArray, pubKArray)
		if err != nil {
			t.Fatal("Unexpected error occurred when decrypting secrets")
		}
		ssArrayAll[ell] = sharedSecretsArray
	}

	gskjArray := make([]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		sharedSecretsArray := ssArrayAll[ell]
		privPolyCoefs := privPolyCoefsArray[ell]
		idx := ell + 1
		selfSecret := cloudflare.PrivatePolyEval(privPolyCoefs, idx)
		gskj := new(big.Int).Set(selfSecret)
		for j := 0; j < n-1; j++ {
			sharedSecret := sharedSecretsArray[j]
			gskj.Add(gskj, sharedSecret)
		}
		gskjArray[ell] = gskj
	}

	shareDistributionEnd, err := c.TSHAREDISTRIBUTIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting ShareDistributionEnd")
	}
	AdvanceBlocksUntil(sim, shareDistributionEnd)
	// Current block number is now 47 > 46 == T_SHARE_DISTRIBUTION_END;
	// in Dispute phase
	disputeEnd, err := c.TDISPUTEEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting DisputeEnd")
	}
	AdvanceBlocksUntil(sim, disputeEnd)
	// Current block number is now 72 > 71 == T_DISPUTE_END;
	// in Key Derivation phase

	// Check block number here
	curBlock := sim.Blockchain().CurrentBlock().Number()
	keyShareSubmissionEnd, err := c.TKEYSHARESUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting KeyShareSubmissionEnd")
	}
	validBlockNumber := (disputeEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(keyShareSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in Key Share Submission Phase")
	}

	// Now to submit key shares
	keyShareArrayG1 := make([]*cloudflare.G1, n)
	keyShareArrayG2 := make([]*cloudflare.G2, n)
	keyShareArrayDLEQProof := make([][2]*big.Int, n)

	h1BaseMsg := []byte("MadHive Rocks!")
	g1Base := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	h1Base, err := cloudflare.HashToG1(h1BaseMsg)
	if err != nil {
		t.Fatal("Error when computing HashToG1([]byte(\"MadHive Rock!\"))")
	}
	h2Base := new(cloudflare.G2).ScalarBaseMult(big.NewInt(1))
	orderMinus1, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495616", 10)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	for ell := 0; ell < n; ell++ {
		secretValue := secretValuesArray[ell]
		g1Value := new(cloudflare.G1).ScalarBaseMult(secretValue)
		keyShareG1 := new(cloudflare.G1).ScalarMult(h1Base, secretValue)
		keyShareG2 := new(cloudflare.G2).ScalarMult(h2Base, secretValue)

		// Generate and Verify DLEQ Proof
		keyShareDLEQProof, err := cloudflare.GenerateDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, secretValue, rand.Reader)
		if err != nil {
			t.Fatal("Error when generating DLEQ Proof")
		}
		err = cloudflare.VerifyDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, keyShareDLEQProof)
		if err != nil {
			t.Fatal("Invalid DLEQ h1Value proof")
		}

		// PairingCheck to ensure keyShareG1 and keyShareG2 form valid pair
		validPair := cloudflare.PairingCheck([]*cloudflare.G1{keyShareG1, h1Base}, []*cloudflare.G2{h2Neg, keyShareG2})
		if !validPair {
			t.Fatal("Error in PairingCheck")
		}

		auth := authArray[ell]
		txOpt := &bind.TransactOpts{
			From:     auth.From,
			Nonce:    nil,
			Signer:   auth.Signer,
			Value:    nil,
			GasPrice: nil,
			GasLimit: gasLim,
			Context:  nil,
		}

		// Check Key Shares to ensure not submitted
		keyShareBig0, err := c.KeyShares(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (0)")
		}
		keyShareBig1, err := c.KeyShares(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (1)")
		}
		zeroKeyShare := (keyShareBig0.Cmp(big0) == 0) && (keyShareBig1.Cmp(big0) == 0)
		if !zeroKeyShare {
			t.Fatal("Unexpected error: KeyShare is nonzero and already present")
		}

		// Check Share Distribution Hashes
		authHash, err := c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
		if err != nil {
			t.Fatal("Error when calling ShareDistributionHashes")
		}
		zeroBytes := make([]byte, numBytes)
		validHash := !bytes.Equal(authHash[:], zeroBytes)
		if !validHash {
			t.Fatal("Unexpected error: invalid hash")
		}

		keyShareG1Big := G1ToBigIntArray(keyShareG1)
		keyShareG2Big := G2ToBigIntArray(keyShareG2)

		_, err = c.SubmitKeyShare(txOpt, auth.From, keyShareG1Big, keyShareDLEQProof, keyShareG2Big)
		if err != nil {
			t.Fatal("Unexpected error occurred when submitting key shares")
		}

		keyShareArrayG1[ell] = keyShareG1
		keyShareArrayG2[ell] = keyShareG2
		keyShareArrayDLEQProof[ell] = keyShareDLEQProof
	}
	sim.Commit()

	// Need to check key share submission and confirm validity
	for ell := 0; ell < n; ell++ {
		auth := authArray[ell]
		keyShareG1 := keyShareArrayG1[ell]
		keyShareBig0, err := c.KeyShares(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (0)")
		}
		keyShareBig1, err := c.KeyShares(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Error occurred when calling c.KeyShares (1)")
		}
		keyShareG1Rcvd, err := BigIntArrayToG1([2]*big.Int{keyShareBig0, keyShareBig1})
		if err != nil {
			t.Fatal("Error in BigIntArrayToG1 call")
		}
		if !keyShareG1.IsEqual(keyShareG1Rcvd) {
			t.Fatal("KeyShareG1 mismatch between submission and received!")
		}
	}

	AdvanceBlocksUntil(sim, keyShareSubmissionEnd)
	// Check block number here
	curBlock = sim.Blockchain().CurrentBlock().Number()
	mpkSubmissionEnd, err := c.TMPKSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting MPKSubmissionEnd")
	}
	validBlockNumber = (keyShareSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(mpkSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in MPK Submission Phase")
	}

	// Make Master Public Key (this is not how you would actually do this)
	mpk := new(cloudflare.G2).Add(keyShareArrayG2[0], keyShareArrayG2[1])
	for ell := 2; ell < n; ell++ {
		mpk.Add(mpk, keyShareArrayG2[ell])
	}
	mpkBig := G2ToBigIntArray(mpk)

	// For G1 version
	mpkG1 := new(cloudflare.G1).Add(keyShareArrayG1[0], keyShareArrayG1[1])
	for ell := 2; ell < n; ell++ {
		mpkG1.Add(mpkG1, keyShareArrayG1[ell])
	}

	// Perform PairingCheck on mpk and mpkG1 to ensure valid pair
	validPair := cloudflare.PairingCheck([]*cloudflare.G1{mpkG1, h1Base}, []*cloudflare.G2{h2Neg, mpk})
	if !validPair {
		t.Fatal("Error in PairingCheck for mpk")
	}

	auth := authArray[0]
	txOpt := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	_, err = c.SubmitMasterPublicKey(txOpt, mpkBig)
	if err != nil {
		t.Fatal("Unexpected error occurred when submitting master public key")
	}
	sim.Commit()

	mpkRcvd0, err := c.MasterPublicKey(&bind.CallOpts{}, big0)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (0)")
	}
	mpkRcvd1, err := c.MasterPublicKey(&bind.CallOpts{}, big1)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (1)")
	}
	mpkRcvd2, err := c.MasterPublicKey(&bind.CallOpts{}, big2)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (2)")
	}
	mpkRcvd3, err := c.MasterPublicKey(&bind.CallOpts{}, big3)
	if err != nil {
		t.Fatal("Unexpected error when calling mpk (3)")
	}
	mpkSubmittedMatchesRcvd := (mpkBig[0].Cmp(mpkRcvd0) == 0) && (mpkBig[1].Cmp(mpkRcvd1) == 0) && (mpkBig[2].Cmp(mpkRcvd2) == 0) && (mpkBig[3].Cmp(mpkRcvd3) == 0)
	if !mpkSubmittedMatchesRcvd {
		t.Fatal("mpk submitted does not match received!")
	}

	// We now proceed to submit gpkj's; they were created above

	// Check block number here
	AdvanceBlocksUntil(sim, mpkSubmissionEnd)
	gpkjSubmissionEnd, err := c.TGPKJSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting GPKJSubmissionEnd")
	}
	curBlock = sim.Blockchain().CurrentBlock().Number()
	validBlockNumber = (mpkSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(gpkjSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in GPKj Submission Phase")
	}

	initialMessage, err := c.InitialMessage(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error when getting InitialMessage for gpkj signature")
	}

	// Make and submit gpkj's
	idx := 0
	gskj := gskjArray[idx]
	gpkj := new(cloudflare.G2).ScalarBaseMult(gskj)
	initialSigBad := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	gpkjBig := G2ToBigIntArray(gpkj)
	initialSigBadBig := G1ToBigIntArray(initialSigBad)

	auth = authArray[idx]
	txOpt = &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	// Ensure no previous submission
	gpkjSubmission0, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission0")
	}
	gpkjSubmission1, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission1")
	}
	gpkjSubmission2, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big2)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission2")
	}
	gpkjSubmission3, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big3)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission3")
	}
	emptyGPKjSub := (gpkjSubmission0.Cmp(big0) == 0) && (gpkjSubmission1.Cmp(big0) == 0) && (gpkjSubmission2.Cmp(big0) == 0) && (gpkjSubmission3.Cmp(big0) == 0)
	if !emptyGPKjSub {
		t.Fatal("Unexpected error; gpkj already submitted")
	}

	// Verify signature
	validSig, err := cloudflare.Verify(initialMessage, initialSigBad, gpkj, cloudflare.HashToG1)
	if err != nil {
		t.Fatal("Error when calling cloudflare.Verify for (initialSig, gpkj) verification")
	}
	if validSig {
		t.Fatal("Unexpected error; initialSig should fail cloudflare.Verify")
	}

	// Attempt to submit invalid signature; should fail
	_, err = c.SubmitGPKj(txOpt, gpkjBig, initialSigBadBig)
	if err != nil {
		t.Fatal("Error should have occurred; attempted to submit gpkj and invalid signature")
	}
	sim.Commit()

	// Confirm submission failed
	gpkjRcvd0, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission0")
	}
	gpkjRcvd1, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission1")
	}
	gpkjRcvd2, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big2)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission2")
	}
	gpkjRcvd3, err := c.GpkjSubmissions(&bind.CallOpts{}, auth.From, big3)
	if err != nil {
		t.Fatal("Error when calling GpkjSubmission3")
	}
	emptyGPKjRcvd := (gpkjRcvd0.Cmp(big0) == 0) && (gpkjRcvd1.Cmp(big0) == 0) && (gpkjRcvd2.Cmp(big0) == 0) && (gpkjRcvd3.Cmp(big0) == 0)
	if !emptyGPKjRcvd {
		t.Fatal("Unexpected error; gpkj submission should have failed due to invalid signature")
	}
}
