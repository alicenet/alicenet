package bn256

import (
	"bytes"
	"context"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/assert"
)

// In this test we proceed all the way through the DKG protocol through
// the completion phase. Here, all validators correctly submit their gpkj.
func TestSuccessfulCompletion(t *testing.T) {
	n := 4
	threshold, _ := thresholdFromUsers(n) // threshold, k are return values
	_ = threshold                         // for linter
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
	// note that this would have to be changed in practice;
	// in practice, each validating will be receiving all of these shares
	// for himself.
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
		receipt, err := sim.WaitForReceipt(context.Background(), txn)
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

	// We how go through and obtain the shared secret arrays for each validator.
	// This entails condensing the commitments and decrypting the shared
	// secrets. We then check to ensure that the shared secret is valid.
	ssArrayAll := make([][]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		ssArrayAll[ell] = make([]*big.Int, n-1)
	}
	for ell := 0; ell < n; ell++ {
		rcvdEncSharesEll := rcvdEncShares[ell]
		rcvdCommitmentsEll := rcvdCommitments[ell]
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
		// Confirm all encrypted stares are valid;
		// failure means that a secret was incorrectly shared.
		ctr := 0
		idx := ell + 1
		for j := 0; j < n; j++ {
			if j == ell {
				continue
			}
			commitmentsEllJ := rcvdCommitmentsEll[j]
			// Need to convert commitments (public coefs) to cloudflare.G1
			pubCoefsEllJ, err := BigIntArraySliceToG1(commitmentsEllJ)
			if err != nil {
				t.Fatal("Error occurred in big.Int to G1 conversion")
			}
			sharedSecretJ := sharedSecretsArray[ctr]
			err = cloudflare.CompareSharedSecret(sharedSecretJ, idx, pubCoefsEllJ)
			if err != nil {
				t.Fatal("Unexpected error; should have valid secret")
			}
			ctr++
		}
		ssArrayAll[ell] = sharedSecretsArray
	}

	// Using the decrypted shares and because everyone correctly shared his
	// secret, we construct the portions of the group secret key (gskj).
	// These will be use to for the portions fo the group public key (gpkj)
	// later.
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
	curBlock := CurrentBlock(sim)
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

	// Submit and receive key shares; check validity
	rcvdKeySharesG1 := make([][][2]*big.Int, n)
	rcvdKeySharesG2 := make([][][4]*big.Int, n)
	for ell := 0; ell < n; ell++ {
		rcvdKeySharesG1[ell] = make([][2]*big.Int, n)
		rcvdKeySharesG2[ell] = make([][4]*big.Int, n)
	}
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

		txn, err := c.SubmitKeyShare(txOpt, auth.From, keyShareG1Big, keyShareDLEQProof, keyShareG2Big)
		if err != nil {
			t.Fatal("Unexpected error occurred when submitting key shares")
		}
		sim.Commit()
		receipt, err := sim.WaitForReceipt(context.Background(), txn)
		if err != nil {
			t.Fatal("Unexpected error in TransactionReceipt")
		}
		keyShareEvent, err := c.ETHDKGFilterer.ParseKeyShareSubmission(*receipt.Logs[0])
		if err != nil {
			t.Fatal("Unexpected error in ParseKeyShareSubmission")
		}
		// Save values in arrays for everyone
		for j := 0; j < n; j++ {
			rcvdKeySharesG1[j][ell] = keyShareEvent.KeyShareG1
			rcvdKeySharesG2[j][ell] = keyShareEvent.KeyShareG2
		}

		keyShareArrayG1[ell] = keyShareG1
		keyShareArrayG2[ell] = keyShareG2
		keyShareArrayDLEQProof[ell] = keyShareDLEQProof
	}

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

	// At this point, we move toward creating the master public key (mpk).
	AdvanceBlocksUntil(sim, keyShareSubmissionEnd)
	// Check block number here
	curBlock = CurrentBlock(sim)
	mpkSubmissionEnd, err := c.TMPKSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting MPKSubmissionEnd")
	}
	validBlockNumber = (keyShareSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(mpkSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in MPK Submission Phase")
	}

	// Make Master Public Key for both the G2 (real) and G1 (dual) versions;
	// how to *actually* do this
	mpk := new(cloudflare.G2)
	mpkG1 := new(cloudflare.G1)
	{
		// This takes one slice of all the received values;
		// we then proceed to sum them after conversion to the appropriate
		// cloudflare type
		rKeySharesG1 := rcvdKeySharesG1[0]
		rKeySharesG2 := rcvdKeySharesG2[0]
		for ell := 0; ell < n; ell++ {
			ksG1Big := rKeySharesG1[ell]
			ksG2Big := rKeySharesG2[ell]
			ksG1, err := BigIntArrayToG1(ksG1Big)
			if err != nil {
				t.Fatal(err)
			}
			mpkG1.Add(mpkG1, ksG1)
			ksG2, err := BigIntArrayToG2(ksG2Big)
			if err != nil {
				t.Fatal(err)
			}
			mpk.Add(mpk, ksG2)
		}
	}
	mpkBig := G2ToBigIntArray(mpk)

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

	// We check to confirm that the mpk was correctly submitted and matches
	// our submission.
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

	// We now proceed to submit gpkj's; the gskj's were created above

	// Check block number here
	AdvanceBlocksUntil(sim, mpkSubmissionEnd)
	gpkjSubmissionEnd, err := c.TGPKJSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting GPKJSubmissionEnd")
	}
	curBlock = CurrentBlock(sim)
	validBlockNumber = (mpkSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(gpkjSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; should be in GPKj Submission Phase")
	}

	// The initialMessage is required because this forces knowledge of
	// the secret key corresponding to gpkj. This ensures that anyone who
	// submits an invalid gpkj is malicious because he had to create a
	// signature which matches gpkj.
	initialMessage, err := c.InitialMessage(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error when getting InitialMessage for gpkj signature")
	}

	// Make and submit gpkj's;
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

	// Confirm the gpkj submissions match what we submitted.
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

	// Confirm no one is malicious
	for ell := 0; ell < n; ell++ {
		auth := authArray[ell]
		isMalicious, err := c.IsMalicious(&bind.CallOpts{}, auth.From)
		if err != nil {
			t.Fatal("Error when calling c.IsMalicious")
		}
		if isMalicious {
			t.Fatal("Should not be malicious at this point")
		}
	}

	sim.Commit()

	// Validate gpkj's by looking at aggregate signatures

	// Test first batch; this will succeed because all of these validators
	// are honest and correctly submitted their signatures and gpkj's.
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

	// Test second batch; this will fail because the final validator submitted
	// an invalid signature and gpkj, although the pair was created correctly.
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
		//t.Fatal("Second batch should succeed; it contains an invalid gpkj")
		t.Fatal("Error in cloudflare.AggregateSignatures")
	}

	// Proceed to GPKj Accusation Phase; in this phase we will accuse and
	// prove the last particpant is dishonest.
	gpkjDisputeEnd, err := c.TGPKJDISPUTEEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting GPKJDisputeEnd")
	}
	AdvanceBlocksUntil(sim, gpkjSubmissionEnd)
	curBlock = CurrentBlock(sim)
	validBlockNumber = (gpkjSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(gpkjDisputeEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in GPKj Dispute Phase")
	}

	f, err := c.TDKGCOMPLETE(&bind.CallOpts{})
	assert.Nilf(t, err, "Unexpected error in getting GPKJSubmissionEnd %v", err)
	f.Add(f, big.NewInt(1))
	AdvanceBlocksUntil(sim, f)

	txn, err := c.SuccessfulCompletion(txOpt)
	assert.Nilf(t, err, "Successful Completion failed... %v", err)
	assert.NotNilf(t, txn, "Could not retrieve transaction from Successful Completion.")

	sim.Commit()

	receipt, err := sim.WaitForReceipt(context.Background(), txn)
	assert.Nilf(t, err, "TransactionReceipt failed... %v", err)
	assert.NotNilf(t, receipt, "Could not retrieve transaction receipt: %v", receipt)

	gpkjIdx := 0
	for _, log := range receipt.Logs {

		if log.Topics[0].Hex() == "0x81c5303ede18e440988d8363b5e854faa79d3baa68891b893cee03c0ff00064b" {
			validatorSet, err := c.ParseValidatorSet(*log)
			assert.Nilf(t, err, "Failed to parse ValidatorSet %v", err)
			assert.Equalf(t, 0, validatorSet.Epoch.Cmp(big.NewInt(0)), "Epoch should be 0, but is %v", validatorSet.Epoch)
			assert.Equalf(t, uint8(4), validatorSet.ValidatorCount, "Should be 4 validators but there are %d", validatorSet.ValidatorCount)

			assert.Equalf(t, mpkBig[0], validatorSet.GroupKey0, "")
			assert.Equalf(t, mpkBig[1], validatorSet.GroupKey1, "")
			assert.Equalf(t, mpkBig[2], validatorSet.GroupKey2, "")
			assert.Equalf(t, mpkBig[3], validatorSet.GroupKey3, "")
		} else if log.Topics[0].Hex() == "0x113b129fac2dde341b9fbbec2bb79a95b9945b0e80fda711fc8ae5c7b0ea83b0" {
			validator, err := c.ParseValidatorMember(*log)
			assert.Nilf(t, err, "Failed to parse Validator %v", err)
			assert.NotNil(t, validator, "Should have gotten an event back")

			gpkjBig := G2ToBigIntArray(gpkjArray[gpkjIdx])
			gpkjIdx++

			t.Logf("gpjkBig[0]: 0x%x", gpkjBig[0])
			t.Logf("    share0: 0x%x", validator.Share0)

			assert.Equalf(t, gpkjBig[0], validator.Share0, "Share0 is wrong")
			assert.Equalf(t, gpkjBig[1], validator.Share1, "Share1 is wrong")
			assert.Equalf(t, gpkjBig[2], validator.Share2, "Share2 is wrong")
			assert.Equalf(t, gpkjBig[3], validator.Share3, "Share3 is wrong")

		} else if log.Topics[0].Hex() == "0x00913d46aef0f0d115d70ea1c7c23198505f577d1d1916cc60710ca2204ae6ae" {
			t.Log("Received Fine event.")
		} else {
			t.Fatalf("Got some unknown event: %s", log.Topics[0].Hex())
		}
	}

	// Confirm no changes
	for ell := 0; ell < n; ell++ {
		idx := ell
		auth := authArray[idx]
		gpkj := gpkjArray[idx]
		gpkjBig := G2ToBigIntArray(gpkj)

		// Confirm gpkj is the same
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

		// No one is malicious
		isMalicious, err := c.IsMalicious(&bind.CallOpts{}, auth.From)
		if err != nil {
			t.Fatal("Error when calling c.IsMalicious")
		}
		if isMalicious {
			t.Fatal("Should not be malicious")
		}
	}
}
