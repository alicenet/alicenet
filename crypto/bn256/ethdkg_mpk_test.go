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

// In this test we proceed to correctly submit the master public key.
func TestSubmitMasterPublicKeySuccess(t *testing.T) {
	n := 4
	threshold, _ := thresholdFromUsers(n) // threshold, k are return values
	_ = threshold                         // for linter
	//c, sim, keyArray, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	c, _, sim, _, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	defer sim.Close()
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Error in getting RegistrationEnd")
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
			t.Error("Error occurred while generating sharing secrets")
		}
		encSharesArray[ell], err = cloudflare.GenerateEncryptedShares(secretsArray, privK, pubKArray)
		if err != nil {
			t.Error("Error occurred while generating commitments")
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
		receipt, err := sim.WaitForReceipt(context.Background(), txn)
		if err != nil {
			t.Error("Unexpected error in TransactionReceipt")
		}
		shareDistEvent, err := c.ETHDKGFilterer.ParseShareDistribution(*receipt.Logs[0])
		if err != nil {
			t.Error("Unexpected error in ParseShareDistribution")
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

	shareDistributionEnd, err := c.TSHAREDISTRIBUTIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Unexpected error in getting ShareDistributionEnd")
	}
	AdvanceBlocksUntil(sim, shareDistributionEnd)
	// Current block number is now 47 > 46 == T_SHARE_DISTRIBUTION_END;
	// in Dispute phase
	disputeEnd, err := c.TDISPUTEEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Unexpected error in getting DisputeEnd")
	}
	AdvanceBlocksUntil(sim, disputeEnd)
	// Current block number is now 72 > 71 == T_DISPUTE_END;
	// in Key Derivation phase

	// Check block number here
	curBlock := CurrentBlock(sim)
	keyShareSubmissionEnd, err := c.TKEYSHARESUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Unexpected error in getting KeyShareSubmissionEnd")
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
	curBlock = CurrentBlock(sim)
	mpkSubmissionEnd, err := c.TMPKSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Unexpected error in getting MPKSubmissionEnd")
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
}

// In this test we attempt to correctly submit the master public key;
// it fails because we are not in the mpk submission phase.
func TestSubmitMasterPublicKeyFailWrongBlock(t *testing.T) {
	n := 4
	threshold, _ := thresholdFromUsers(n) // threshold, k are return values
	_ = threshold                         // for linter
	//c, sim, keyArray, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	c, _, sim, _, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	defer sim.Close()
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Error in getting RegistrationEnd")
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
			t.Error("Error occurred while generating sharing secrets")
		}
		encSharesArray[ell], err = cloudflare.GenerateEncryptedShares(secretsArray, privK, pubKArray)
		if err != nil {
			t.Error("Error occurred while generating commitments")
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
		receipt, err := sim.WaitForReceipt(context.Background(), txn)
		if err != nil {
			t.Error("Unexpected error in TransactionReceipt")
		}
		shareDistEvent, err := c.ETHDKGFilterer.ParseShareDistribution(*receipt.Logs[0])
		if err != nil {
			t.Error("Unexpected error in ParseShareDistribution")
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

	shareDistributionEnd, err := c.TSHAREDISTRIBUTIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Unexpected error in getting ShareDistributionEnd")
	}
	AdvanceBlocksUntil(sim, shareDistributionEnd)
	// Current block number is now 47 > 46 == T_SHARE_DISTRIBUTION_END;
	// in Dispute phase
	disputeEnd, err := c.TDISPUTEEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Unexpected error in getting DisputeEnd")
	}
	AdvanceBlocksUntil(sim, disputeEnd)
	// Current block number is now 72 > 71 == T_DISPUTE_END;
	// in Key Derivation phase

	// Check block number here
	curBlock := CurrentBlock(sim)
	keyShareSubmissionEnd, err := c.TKEYSHARESUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Unexpected error in getting KeyShareSubmissionEnd")
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

	// Check block number here; will fail
	curBlock = CurrentBlock(sim)
	mpkSubmissionEnd, err := c.TMPKSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Unexpected error in getting MPKSubmissionEnd")
	}
	AdvanceBlocksUntil(sim, mpkSubmissionEnd)
	validBlockNumber = (keyShareSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(mpkSubmissionEnd) <= 0)
	if validBlockNumber {
		t.Fatal("Unexpected error; not in MPK Submission Phase")
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

	zeroMPK := (big0.Cmp(mpkRcvd0) == 0) && (big0.Cmp(mpkRcvd1) == 0) && (big0.Cmp(mpkRcvd2) == 0) && (big0.Cmp(mpkRcvd3) == 0)
	if !zeroMPK {
		t.Fatal("mpk should be zero because submitted during wrong block!")
	}
}

// In this test we attempt to submit the master public key;
// it fails because the mpk we submit is invalid.
func TestSubmitMasterPublicKeyFailInvalidMPK(t *testing.T) {
	n := 4
	threshold, _ := thresholdFromUsers(n) // threshold, k are return values
	_ = threshold                         // for linter
	//c, sim, keyArray, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	c, _, sim, _, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	defer sim.Close()
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Error in getting RegistrationEnd")
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
			t.Error("Error occurred while generating sharing secrets")
		}
		encSharesArray[ell], err = cloudflare.GenerateEncryptedShares(secretsArray, privK, pubKArray)
		if err != nil {
			t.Error("Error occurred while generating commitments")
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
		receipt, err := sim.WaitForReceipt(context.Background(), txn)
		if err != nil {
			t.Error("Unexpected error in TransactionReceipt")
		}
		shareDistEvent, err := c.ETHDKGFilterer.ParseShareDistribution(*receipt.Logs[0])
		if err != nil {
			t.Error("Unexpected error in ParseShareDistribution")
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

	shareDistributionEnd, err := c.TSHAREDISTRIBUTIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Unexpected error in getting ShareDistributionEnd")
	}
	AdvanceBlocksUntil(sim, shareDistributionEnd)
	// Current block number is now 47 > 46 == T_SHARE_DISTRIBUTION_END;
	// in Dispute phase
	disputeEnd, err := c.TDISPUTEEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Unexpected error in getting DisputeEnd")
	}
	AdvanceBlocksUntil(sim, disputeEnd)
	// Current block number is now 72 > 71 == T_DISPUTE_END;
	// in Key Derivation phase

	// Check block number here
	curBlock := CurrentBlock(sim)
	keyShareSubmissionEnd, err := c.TKEYSHARESUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Unexpected error in getting KeyShareSubmissionEnd")
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
	curBlock = CurrentBlock(sim)
	mpkSubmissionEnd, err := c.TMPKSUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Unexpected error in getting MPKSubmissionEnd")
	}
	validBlockNumber = (keyShareSubmissionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(mpkSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in MPK Submission Phase")
	}

	// Make Master Public Key; this will fail
	mpkBad := new(cloudflare.G2).ScalarBaseMult(big.NewInt(1))
	mpkBadBig := G2ToBigIntArray(mpkBad)

	// For G1 version
	mpkG1 := new(cloudflare.G1).Add(keyShareArrayG1[0], keyShareArrayG1[1])
	for ell := 2; ell < n; ell++ {
		mpkG1.Add(mpkG1, keyShareArrayG1[ell])
	}

	// Perform PairingCheck on mpk and mpkG1 to ensure valid pair; will fail
	validPair := cloudflare.PairingCheck([]*cloudflare.G1{mpkG1, h1Base}, []*cloudflare.G2{h2Neg, mpkBad})
	if validPair {
		t.Fatal("Unexpected Error; this should fail")
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

	_, err = c.SubmitMasterPublicKey(txOpt, mpkBadBig)
	if err != nil {
		t.Fatal("Unexpected error occurred when submitting master public key")
	}

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

	zeroMPK := (big0.Cmp(mpkRcvd0) == 0) && (big0.Cmp(mpkRcvd1) == 0) && (big0.Cmp(mpkRcvd2) == 0) && (big0.Cmp(mpkRcvd3) == 0)
	if !zeroMPK {
		t.Fatal("mpk should be zero because invalid submission!")
	}
}
