package bn256

import (
	"bytes"
	"context"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
)

// In this test all participants correctly submit their key share,
// their portion of the master public key mpk.
func TestSubmitKeyShareSuccess(t *testing.T) {
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
		t.Fatal("Unexpected error; in Key Share Phase")
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
}

// Attempt to submit key share but fail due to incorrect block number.
func TestSubmitKeyShareFailBlockNumber(t *testing.T) {
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
	//AdvanceBlocksUntil(sim, disputeEnd)
	// Current block number is now 72 > 71 == T_DISPUTE_END;
	// in Key Derivation phase

	// Check block number here; will fail
	curBlock := sim.Blockchain().CurrentBlock().Number()
	keyShareSubmissionEnd, err := c.TKEYSHARESUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting KeyShareSubmissionEnd")
	}
	validBlockNumber := (disputeEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(keyShareSubmissionEnd) <= 0)
	if validBlockNumber {
		t.Fatal("Should fail; not in Dispute Phase")
	}

	h1BaseMsg := []byte("MadHive Rocks!")
	g1Base := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	h1Base, err := cloudflare.HashToG1(h1BaseMsg)
	if err != nil {
		t.Fatal("Error when computing HashToG1([]byte(\"MadHive Rock!\"))")
	}
	h2Base := new(cloudflare.G2).ScalarBaseMult(big.NewInt(1))
	orderMinus1, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495616", 10)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	idx := 0
	secretValue := secretValuesArray[idx]
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

	auth := authArray[idx]
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
	sim.Commit()

	keyShareBig0, err = c.KeyShares(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Error occurred when calling c.KeyShares (0)")
	}
	keyShareBig1, err = c.KeyShares(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Error occurred when calling c.KeyShares (1)")
	}
	zeroKeyShare = (keyShareBig0.Cmp(big0) == 0) && (keyShareBig1.Cmp(big0) == 0)
	if !zeroKeyShare {
		t.Fatal("KeyShareG1 should be zero due to wrong block number!")
	}
}

func TestSubmitKeyShareFailInvalidHash(t *testing.T) {
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

	// Begin messing with correct values here
	//
	// issuer is getting accused; disputer is accusing issuer of malicious behavior
	// Participant 1 (index base 1)
	disputerListIdx := 0
	disputerListIdxBig := big.NewInt(int64(disputerListIdx))
	// Participant 2 (index base 1)
	issuerListIdx := 1
	issuerListIdxBig := big.NewInt(int64(issuerListIdx))
	if issuerListIdx == disputerListIdx {
		t.Fatal("Cannot have disputer accuse himself")
	}

	// Make bad values now
	badSecretValueIssToDisp := big.NewInt(0) // We replace true with this incorrect value

	// Test encrypted shares array to confirm we compute the correct value
	trueEncValueIssToDisp := new(big.Int)
	trueEncValueIssToDisp.SetInt64(0)

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
		if ell == issuerListIdx {
			// Change the secret value from what it should be
			if disputerListIdx < issuerListIdx {
				trueEncValueIssToDisp = secretsArray[disputerListIdx]
				secretsArray[disputerListIdx] = badSecretValueIssToDisp
			} else {
				trueEncValueIssToDisp = secretsArray[disputerListIdx-1]
				secretsArray[disputerListIdx-1] = badSecretValueIssToDisp
			}
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

	shareDistributionEnd, err := c.TSHAREDISTRIBUTIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting ShareDistributionEnd")
	}
	disputeEnd, err := c.TDISPUTEEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting DisputeEnd")
	}
	AdvanceBlocksUntil(sim, shareDistributionEnd)
	curBlock := sim.Blockchain().CurrentBlock().Number()
	// Current block number is now 47 > 46 == T_SHARE_DISTRIBUTION_END;
	// in Dispute phase

	// Confirm valid shares
	for ell := 0; ell < n; ell++ {
		if ell == disputerListIdx {
			// We will handle the disputed participant separately,
			// as there should be no other invalid entries.
			continue
		}
		// Now to loop through and confirm valid secrets
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
	}

	// Now to confirm all shares of disputer except invalid one
	{
		rcvdEncSharesDisp := rcvdEncShares[disputerListIdx]
		rcvdCommitmentsDisp := rcvdCommitments[disputerListIdx]
		pubK := pubKArray[disputerListIdx]
		privK := privKArray[disputerListIdx]
		sharedEncryptedArray, err := cloudflare.CondenseCommitments(pubK, rcvdEncSharesDisp, pubKArray)
		if err != nil {
			t.Fatal("Unexpected error occurred when condensing commitments")
		}
		sharedSecretsArray, err := cloudflare.GenerateDecryptedShares(privK, sharedEncryptedArray, pubKArray)
		if err != nil {
			t.Fatal("Unexpected error occurred when decrypting secrets")
		}
		idx := disputerListIdx + 1
		ctr := 0
		for j := 0; j < n; j++ {
			if j == disputerListIdx {
				continue
			}
			if j != issuerListIdx {
				commitmentsDispJ := rcvdCommitmentsDisp[j]
				// Need to convert commitments (public coefs) to cloudflare.G1
				pubCoefsDispJ, err := BigIntArraySliceToG1(commitmentsDispJ)
				if err != nil {
					t.Fatal("Error occurred in big.Int to G1 conversion")
				}
				sharedSecretJ := sharedSecretsArray[ctr]
				err = cloudflare.CompareSharedSecret(sharedSecretJ, idx, pubCoefsDispJ)
				if err != nil {
					t.Fatal("Unexpected error; should have valid secret")
				}
			} else {
				commitmentsDispIss := rcvdCommitmentsDisp[issuerListIdx]
				// Need to convert commitments (public coefs) to cloudflare.G1
				pubCoefsDispIss, err := BigIntArraySliceToG1(commitmentsDispIss)
				if err != nil {
					t.Fatal("Error occurred in big.Int to G1 conversion")
				}
				sharedSecretIss := sharedSecretsArray[ctr]
				err = cloudflare.CompareSharedSecret(sharedSecretIss, idx, pubCoefsDispIss)
				if err == nil {
					t.Fatal("Error should have been raised; invalid secret share")
				}
			}
			ctr++
		}
	}
	// At this point, we have shown that Issuer sent Disputer an invalid secret

	// Generate DLEQ Proof and confirm it is valid; confirmation occurs later
	g1Base := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	pubKIssuer := pubKArray[issuerListIdx]
	privKDisputer := privKArray[disputerListIdx]
	pubKDisputer := pubKArray[disputerListIdx]
	sharedSecret := cloudflare.GenerateSharedSecretG1(privKDisputer, pubKIssuer)
	sharedSecretBig := G1ToBigIntArray(sharedSecret)
	sharedSecretProof, err := cloudflare.GenerateDLEQProofG1(g1Base, pubKDisputer, pubKIssuer, sharedSecret, privKDisputer, rand.Reader)
	if err != nil {
		t.Fatal("Error arose in DLEQ G1 proof generation")
	}

	// Confirm invalid submission:
	encSubFromIssToDisp := new(big.Int)
	encSubFromIssToDisp.SetInt64(0)
	if disputerListIdx < issuerListIdx {
		encSubFromIssToDisp = rcvdEncShares[disputerListIdx][issuerListIdx][disputerListIdx]
	} else {
		encSubFromIssToDisp = rcvdEncShares[disputerListIdx][issuerListIdx][disputerListIdx-1]
	}
	decryptDispIdx := disputerListIdx + 1
	secretValueIssToDisp := cloudflare.Decrypt(encSubFromIssToDisp, privKDisputer, pubKIssuer, decryptDispIdx)
	if secretValueIssToDisp.Cmp(badSecretValueIssToDisp) != 0 {
		t.Fatal("Issued secret value (decrypted) should match!")
	}

	encSharesDispFromIss := rcvdEncShares[disputerListIdx][issuerListIdx]
	commitmentsDispFromIss := rcvdCommitments[disputerListIdx][issuerListIdx]

	auth := authArray[disputerListIdx]
	txOpts := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}
	issuer := authArray[issuerListIdx].From

	// Check block number here
	validBlockNumber := (shareDistributionEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(disputeEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Should succeed; in Dispute Phase")
	}

	// Check issuer and disputer
	issuerAddrCheck, err := c.Addresses(&bind.CallOpts{}, issuerListIdxBig)
	if err != nil {
		t.Fatal("Error when calling issuer address")
	}
	disputerAddrCheck, err := c.Addresses(&bind.CallOpts{}, disputerListIdxBig)
	if err != nil {
		t.Fatal("Error when calling disputer address")
	}
	validAddresses := (issuerAddrCheck == issuer) && (disputerAddrCheck == auth.From)
	if !validAddresses {
		t.Fatal("Issuer and disputer addresses not correct")
	}

	// Check distribution hash
	shareBytes, err := MarshalBigIntSlice(encSharesDispFromIss)
	if err != nil {
		t.Fatal("Something when wrong with marshalling encrypted shares")
	}
	commitmentBytes, err := MarshalG1BigSlice(commitmentsDispFromIss)
	if err != nil {
		t.Fatal("Something when wrong with marshalling commitments")
	}
	bytesLen := len(shareBytes) + len(commitmentBytes)
	combinedBytes := make([]byte, bytesLen)
	for k := 0; k < len(shareBytes); k++ {
		combinedBytes[k] = shareBytes[k]
	}
	for k := 0; k < len(commitmentBytes); k++ {
		combinedBytes[len(shareBytes)+k] = commitmentBytes[k]
	}
	shareCommitmentHash := crypto.Keccak256(combinedBytes)
	issHash, err := c.ShareDistributionHashes(&bind.CallOpts{}, issuer)
	if err != nil {
		t.Fatal("Error raised when obtaining share distribution hash")
	}
	validHash := bytes.Equal(shareCommitmentHash, issHash[:])
	if !validHash {
		t.Fatal("Our computed commitment hash does not match submitted one")
	}

	// Verify DLEQ equality
	err = cloudflare.VerifyDLEQProofG1(g1Base, pubKDisputer, pubKIssuer, sharedSecret, sharedSecretProof)
	if err != nil {
		t.Fatal("Invalid DLEQ G1 Proof for accusation")
	}

	// In above code, already confirmed that shared secret is invalid

	_, err = c.SubmitDispute(txOpts, issuer, issuerListIdxBig, disputerListIdxBig, encSharesDispFromIss, commitmentsDispFromIss, sharedSecretBig, sharedSecretProof)
	if err != nil {
		t.Fatal("c.SubmitDispute should not have raised an error for submitting valid dispute")
	}

	// Now move on to Submit Key Share phase and attempt to have Issuer (who
	// submitted bad shares) submit a key share; this should fail
	AdvanceBlocksUntil(sim, disputeEnd)
	// Current block number is now 72 > 71 == T_DISPUTE_END;
	// in Key Derivation phase

	// Check block number here; will fail
	curBlock = sim.Blockchain().CurrentBlock().Number()
	keyShareSubmissionEnd, err := c.TKEYSHARESUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting KeyShareSubmissionEnd")
	}
	validBlockNumber = (disputeEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(keyShareSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in Dispute Phase")
	}

	h1BaseMsg := []byte("MadHive Rocks!")
	//g1Base := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	h1Base, err := cloudflare.HashToG1(h1BaseMsg)
	if err != nil {
		t.Fatal("Error when computing HashToG1([]byte(\"MadHive Rock!\"))")
	}
	h2Base := new(cloudflare.G2).ScalarBaseMult(big.NewInt(1))
	orderMinus1, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495616", 10)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	// issuerListIdx has been removed for invalid shares
	idx := issuerListIdx
	secretValue := secretValuesArray[idx]
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

	auth = authArray[idx]
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

	// Check Share Distribution Hashes; this will fail
	authHash, err := c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
	if err != nil {
		t.Fatal("Error when calling ShareDistributionHashes")
	}
	zeroBytes := make([]byte, numBytes)
	validHash = !bytes.Equal(authHash[:], zeroBytes)
	if validHash {
		t.Fatal("Unexpected error: hash should be invalid because participant removed")
	}

	keyShareG1Big := G1ToBigIntArray(keyShareG1)
	keyShareG2Big := G2ToBigIntArray(keyShareG2)

	_, err = c.SubmitKeyShare(txOpt, auth.From, keyShareG1Big, keyShareDLEQProof, keyShareG2Big)
	if err != nil {
		t.Fatal("Unexpected error occurred when submitting key shares")
	}
	sim.Commit()

	keyShareBig0, err = c.KeyShares(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Error occurred when calling c.KeyShares (0)")
	}
	keyShareBig1, err = c.KeyShares(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Error occurred when calling c.KeyShares (1)")
	}
	zeroKeyShare = (keyShareBig0.Cmp(big0) == 0) && (keyShareBig1.Cmp(big0) == 0)
	if !zeroKeyShare {
		t.Fatal("KeyShareG1 should be zero due to bad shares!")
	}
}

func TestSubmitKeyShareFailInvalidDLEQProof(t *testing.T) {
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

	// Check block number here; will fail
	curBlock := sim.Blockchain().CurrentBlock().Number()
	keyShareSubmissionEnd, err := c.TKEYSHARESUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting KeyShareSubmissionEnd")
	}
	validBlockNumber := (disputeEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(keyShareSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in Dispute Phase")
	}

	h1BaseMsg := []byte("MadHive Rocks!")
	g1Base := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	h1Base, err := cloudflare.HashToG1(h1BaseMsg)
	if err != nil {
		t.Fatal("Error when computing HashToG1([]byte(\"MadHive Rock!\"))")
	}
	h2Base := new(cloudflare.G2).ScalarBaseMult(big.NewInt(1))
	orderMinus1, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495616", 10)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	idx := 0
	secretValue := secretValuesArray[idx]
	g1Value := new(cloudflare.G1).ScalarBaseMult(secretValue)
	keyShareG1 := new(cloudflare.G1).ScalarMult(h1Base, secretValue)
	keyShareG2 := new(cloudflare.G2).ScalarMult(h2Base, secretValue)

	// Generate and Verify DLEQ Proof; have invalid proof
	keyShareDLEQProofBad := [2]*big.Int{big.NewInt(1), big.NewInt(1)}
	err = cloudflare.VerifyDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, keyShareDLEQProofBad)
	if err == nil {
		t.Fatal("Unexpected error; should have invalid DLEQ h1Value proof")
	}

	// PairingCheck to ensure keyShareG1 and keyShareG2 form valid pair
	validPair := cloudflare.PairingCheck([]*cloudflare.G1{keyShareG1, h1Base}, []*cloudflare.G2{h2Neg, keyShareG2})
	if !validPair {
		t.Fatal("Error in PairingCheck")
	}

	auth := authArray[idx]
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

	// Check Share Distribution Hashes to ensure able to submit
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

	_, err = c.SubmitKeyShare(txOpt, auth.From, keyShareG1Big, keyShareDLEQProofBad, keyShareG2Big)
	if err != nil {
		t.Fatal("Unexpected error occurred when submitting key shares")
	}
	sim.Commit()

	keyShareBig0, err = c.KeyShares(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Error occurred when calling c.KeyShares (0)")
	}
	keyShareBig1, err = c.KeyShares(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Error occurred when calling c.KeyShares (1)")
	}
	zeroKeyShare = (keyShareBig0.Cmp(big0) == 0) && (keyShareBig1.Cmp(big0) == 0)
	if !zeroKeyShare {
		t.Fatal("KeyShareG1 should be zero due to invalid DLEQ proof!")
	}
}

func TestSubmitKeyShareFailInvalidPairingCheck(t *testing.T) {
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
	keyShareSubmissionEnd, err := c.TKEYSHARESUBMISSIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting KeyShareSubmissionEnd")
	}
	curBlock := sim.Blockchain().CurrentBlock().Number()
	validBlockNumber := (disputeEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(keyShareSubmissionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Unexpected error; in Dispute Phase")
	}

	h1BaseMsg := []byte("MadHive Rocks!")
	g1Base := new(cloudflare.G1).ScalarBaseMult(big.NewInt(1))
	h1Base, err := cloudflare.HashToG1(h1BaseMsg)
	if err != nil {
		t.Fatal("Error when computing HashToG1([]byte(\"MadHive Rock!\"))")
	}
	h2Base := new(cloudflare.G2).ScalarBaseMult(big.NewInt(1))
	orderMinus1, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495616", 10)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	idx := 0
	bigOne := big.NewInt(1)
	secretValue := secretValuesArray[idx]
	secretValueBad := new(big.Int).Add(secretValue, bigOne)
	g1Value := new(cloudflare.G1).ScalarBaseMult(secretValue)
	keyShareG1 := new(cloudflare.G1).ScalarMult(h1Base, secretValue)
	keyShareG2Bad := new(cloudflare.G2).ScalarMult(h2Base, secretValueBad)

	// Generate and Verify DLEQ Proof
	keyShareDLEQProof, err := cloudflare.GenerateDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, secretValue, rand.Reader)
	if err != nil {
		t.Fatal("Error when generating DLEQ Proof")
	}
	err = cloudflare.VerifyDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, keyShareDLEQProof)
	if err != nil {
		t.Fatal("Unexpected error; should have valid DLEQ h1Value proof")
	}

	// PairingCheck to ensure keyShareG1 and keyShareG2 form valid pair; will fail
	validPair := cloudflare.PairingCheck([]*cloudflare.G1{keyShareG1, h1Base}, []*cloudflare.G2{h2Neg, keyShareG2Bad})
	if validPair {
		t.Fatal("Unexpected error; PairingCheck should fail")
	}

	auth := authArray[idx]
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

	// Check Share Distribution Hashes to ensure able to submit
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
	keyShareG2BadBig := G2ToBigIntArray(keyShareG2Bad)

	_, err = c.SubmitKeyShare(txOpt, auth.From, keyShareG1Big, keyShareDLEQProof, keyShareG2BadBig)
	if err != nil {
		t.Fatal("Unexpected error occurred when submitting key shares")
	}
	sim.Commit()

	keyShareBig0, err = c.KeyShares(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Error occurred when calling c.KeyShares (0)")
	}
	keyShareBig1, err = c.KeyShares(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Error occurred when calling c.KeyShares (1)")
	}
	zeroKeyShare = (keyShareBig0.Cmp(big0) == 0) && (keyShareBig1.Cmp(big0) == 0)
	if !zeroKeyShare {
		t.Fatal("KeyShareG1 should be zero due to invalid PairingCheck!")
	}
}
