package bn256

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// Here we successfully submit shares and check that we have a nonzero
// hash value.
func TestDistributeShares(t *testing.T) {
	n := 4
	threshold, _ := thresholdFromUsers(n) // threshold, k are return values
	_ = threshold                         // for linter
	//c, sim, keyArray, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	c, _, sim, _, authArray, privKArray, pubKArray := InitialTestSetup(t, n)
	defer sim.Close()

	big0 := big.NewInt(0)
	big1 := big.NewInt(1)

	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error in getting RegistrationEnd")
	}
	shareDistributionEnd, err := c.TSHAREDISTRIBUTIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting ShareDistributionEnd")
	}
	AdvanceBlocksUntil(sim, registrationEnd)
	curBlock := CurrentBlock(sim)
	// Current block number is now 22 > 21 == T_REGISTRATION_END;
	// in Share Distribution phase

	// These are the standard secrets be used for testing purposes
	secretsValuesArray := make([]*big.Int, n)
	secretBase := big.NewInt(100)
	for j := 0; j < n; j++ {
		secretsValuesArray[j] = new(big.Int).Add(secretBase, big.NewInt(int64(j)))
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
		privPolyCoefsArray[ell][0] = secretsValuesArray[ell]
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

	// Check block number here
	validBlockNumber := (registrationEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(shareDistributionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Should succeed; in Dispute Phase")
	}

	// Submit encrypted shares and commitments
	for ell := 0; ell < n; ell++ {
		privK := privKArray[ell]
		auth := authArray[ell]
		encShares, err := cloudflare.GenerateEncryptedShares(privPolyCoefsArray[ell], privK, pubKArray)
		if err != nil {
			t.Fatal("Error occurred while generating commitments")
		}
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

		_, err = c.DistributeShares(txOpt, encShares, pubCoefsBigArray[ell])
		if err != nil {
			t.Fatal("Unexpected error arose in DistributeShares submission")
		}
	}
	sim.Commit()

	// Confirm shares are submitted by looking at hash output
	for ell := 0; ell < n; ell++ {
		auth := authArray[ell]
		// Check distribution hash
		authHash, err := c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
		if err != nil {
			t.Fatal("Error raised when obtaining share distribution hash")
		}
		zeroBytes := make([]byte, 32)
		emptyHash := bytes.Equal(zeroBytes, authHash[:])
		if emptyHash {
			t.Fatal("Failed to submit shares resulting in a valid hash!")
		}
	}
}

// We attempt to distribute shares but fail because we are not in the
// correct phase.
func TestDistributeSharesFailBlockNumber(t *testing.T) {
	n := 1
	c, _, sim, _, authArray, _, _ := InitialTestSetup(t, n)
	defer sim.Close()
	// Current block number is now 1, so cannot submit shares

	// Check block number here
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error in getting RegistrationEnd")
	}
	shareDistributionEnd, err := c.TSHAREDISTRIBUTIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting ShareDistributionEnd")
	}
	curBlock := CurrentBlock(sim)
	validBlockNumber := (registrationEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(shareDistributionEnd) <= 0)
	if validBlockNumber {
		t.Fatal("Should fail; in Registration Phase")
	}

	auth := authArray[0]
	encShares := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2)}
	com0 := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	com1 := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	com2 := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	commitments := [][2]*big.Int{com0, com1, com2}
	txOpt := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	// Check distribution hash prior to submission; should be zero
	authHash, err := c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
	if err != nil {
		t.Fatal("Error raised when obtaining share distribution hash")
	}
	zeroBytes := make([]byte, 32)
	emptyHash := bytes.Equal(zeroBytes, authHash[:])
	if !emptyHash {
		t.Fatal("Should have empty hash!")
	}

	_, err = c.DistributeShares(txOpt, encShares, commitments)
	if err != nil {
		t.Fatal("Unexpected error arose in DistributeShares submission")
	}
	sim.Commit()

	// Check distribution hash
	authHash, err = c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
	if err != nil {
		t.Fatal("Error raised when obtaining share distribution hash")
	}
	emptyHash = bytes.Equal(zeroBytes, authHash[:])
	if !emptyHash {
		t.Fatal("Submitted shares resulting in a valid hash; should fail from wrong block number!")
	}
}

// We attempt to submit shares but do so without being registered.
func TestDistributeSharesFailNotRegistered(t *testing.T) {
	n := 1
	c, _, sim, _, authArray, _, _ := InitialTestSetup(t, n)
	defer sim.Close()
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error in getting RegistrationEnd")
	}
	shareDistributionEnd, err := c.TSHAREDISTRIBUTIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting ShareDistributionEnd")
	}
	AdvanceBlocksUntil(sim, registrationEnd)
	// Current block number is now 22 > 21 == T_REGISTRATION_END;
	// no additional registration allowed, so cannot submit shares

	// Check block number here
	curBlock := CurrentBlock(sim)
	validBlockNumber := (registrationEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(shareDistributionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Should succeed; in Distribution Phase")
	}

	auth := authArray[0]
	encShares := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2)}
	com0 := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	com1 := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	com2 := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	commitments := [][2]*big.Int{com0, com1, com2}
	txOpt := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	// Check distribution hash prior to submission; should be zero
	authHash, err := c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
	if err != nil {
		t.Fatal("Error raised when obtaining share distribution hash")
	}
	zeroBytes := make([]byte, 32)
	emptyHash := bytes.Equal(zeroBytes, authHash[:])
	if !emptyHash {
		t.Fatal("Should have empty hash!")
	}

	_, err = c.DistributeShares(txOpt, encShares, commitments)
	if err != nil {
		t.Fatal("Unexpected error arose in DistributeShares submission")
	}
	sim.Commit()

	// Check distribution hash
	authHash, err = c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
	if err != nil {
		t.Fatal("Error raised when obtaining share distribution hash")
	}
	emptyHash = bytes.Equal(zeroBytes, authHash[:])
	if !emptyHash {
		t.Fatal("Submitted shares resulting in a valid hash; should fail from no registration!")
	}
}

// We attempt to submit shares but do not have the correct number of shares.
func TestDistributeSharesFailIncorrectNumberShares(t *testing.T) {
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
	shareDistributionEnd, err := c.TSHAREDISTRIBUTIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Unexpected error in getting ShareDistributionEnd")
	}
	AdvanceBlocksUntil(sim, registrationEnd)
	// Current block number is now 22 > 21 == T_REGISTRATION_END;
	// in Share Distribution phase

	// Check block number here
	curBlock := CurrentBlock(sim)
	validBlockNumber := (registrationEnd.Cmp(curBlock) < 0) && (curBlock.Cmp(shareDistributionEnd) <= 0)
	if !validBlockNumber {
		t.Fatal("Should succeed; in Distribution Phase")
	}

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

	privPolyCoefs := make([]*big.Int, threshold+1)
	privPolyCoefs[0] = secretValuesArray[0]
	for j := 1; j < len(privPolyCoefs); j++ {
		privPolyCoefs[j] = basePrivatePolynomialCoefs[j]
	}
	pubCoefs := cloudflare.GeneratePublicCoefs(privPolyCoefs)
	pubCoefsBig := make([][2]*big.Int, len(pubCoefs))
	for j := 0; j < len(pubCoefs); j++ {
		pubCoefsBig[j] = G1ToBigIntArray(pubCoefs[j])
	}

	privK := privKArray[0]
	auth := authArray[0]
	encShares, err := cloudflare.GenerateEncryptedShares(privPolyCoefs, privK, pubKArray)
	encSharesFalse := encShares[:n-2] // Remove last element, making shares invalid
	if err != nil {
		t.Fatal("Error occurred while generating commitments")
	}
	txOpt := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	// Check distribution hash prior to submission; should be zero
	authHash, err := c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
	if err != nil {
		t.Fatal("Error raised when obtaining share distribution hash")
	}
	zeroBytes := make([]byte, 32)
	emptyHash := bytes.Equal(zeroBytes, authHash[:])
	if !emptyHash {
		t.Fatal("Should have empty hash!")
	}

	_, err = c.DistributeShares(txOpt, encSharesFalse, pubCoefsBig)
	if err != nil {
		t.Fatal("Unexpected error arose in DistributeShares submission")
	}
	sim.Commit()

	// Check distribution hash
	authHash, err = c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
	if err != nil {
		t.Fatal("Error raised when obtaining share distribution hash")
	}
	emptyHash = bytes.Equal(zeroBytes, authHash[:])
	if !emptyHash {
		t.Fatal("Submitted shares resulting in a valid hash; should fail from incorrect number of shares!")
	}
}

// We attempt to submit the incorrect number commitments (coefficients of
// public polynomial).
func TestDistributeSharesFailIncorrectNumberCommitments(t *testing.T) {
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

	privPolyCoefs := make([]*big.Int, threshold+1)
	privPolyCoefs[0] = secretValuesArray[0]
	for j := 1; j < len(privPolyCoefs); j++ {
		privPolyCoefs[j] = basePrivatePolynomialCoefs[j]
	}
	pubCoefs := cloudflare.GeneratePublicCoefs(privPolyCoefs)
	pubCoefsBig := make([][2]*big.Int, len(pubCoefs))
	for j := 0; j < len(pubCoefs); j++ {
		pubCoefsBig[j] = G1ToBigIntArray(pubCoefs[j])
	}
	pubCoefsBigFalse := pubCoefsBig[:threshold] // Remove last element, making commitments invalid

	privK := privKArray[0]
	auth := authArray[0]
	encShares, err := cloudflare.GenerateEncryptedShares(privPolyCoefs, privK, pubKArray)
	if err != nil {
		t.Fatal("Error occurred while generating commitments")
	}
	txOpt := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	// Check distribution hash prior to submission; should be zero
	authHash, err := c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
	if err != nil {
		t.Fatal("Error raised when obtaining share distribution hash")
	}
	zeroBytes := make([]byte, 32)
	emptyHash := bytes.Equal(zeroBytes, authHash[:])
	if !emptyHash {
		t.Fatal("Should have empty hash!")
	}

	_, err = c.DistributeShares(txOpt, encShares, pubCoefsBigFalse)
	if err != nil {
		t.Fatal("Unexpected error arose in DistributeShares submission")
	}
	sim.Commit()

	// Check distribution hash
	authHash, err = c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
	if err != nil {
		t.Fatal("Error raised when obtaining share distribution hash")
	}
	emptyHash = bytes.Equal(zeroBytes, authHash[:])
	if !emptyHash {
		t.Fatal("Submitted shares resulting in a valid hash; should fail from incorrect number of commitments!")
	}
}

// We attempt to submit invalid commitments (public polynomial coefficients).
func TestDistributeSharesFailInvalidCommitments(t *testing.T) {
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

	privPolyCoefs := make([]*big.Int, threshold+1)
	privPolyCoefs[0] = secretValuesArray[0]
	for j := 1; j < len(privPolyCoefs); j++ {
		privPolyCoefs[j] = basePrivatePolynomialCoefs[j]
	}
	pubCoefs := cloudflare.GeneratePublicCoefs(privPolyCoefs)
	pubCoefsBig := make([][2]*big.Int, len(pubCoefs))
	for j := 0; j < len(pubCoefs); j++ {
		pubCoefsBig[j] = G1ToBigIntArray(pubCoefs[j])
	}
	pubCoefsBigFalse := pubCoefsBig
	pubCoefsBigFalse[0] = [2]*big.Int{big.NewInt(1), big.NewInt(3)} // (1,3) is not valid curve point

	privK := privKArray[0]
	auth := authArray[0]
	encShares, err := cloudflare.GenerateEncryptedShares(privPolyCoefs, privK, pubKArray)
	if err != nil {
		t.Fatal("Error occurred while generating commitments")
	}
	txOpt := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	// Check distribution hash prior to submission; should be zero
	authHash, err := c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
	if err != nil {
		t.Fatal("Error raised when obtaining share distribution hash")
	}
	zeroBytes := make([]byte, 32)
	emptyHash := bytes.Equal(zeroBytes, authHash[:])
	if !emptyHash {
		t.Fatal("Should have empty hash!")
	}

	_, err = c.DistributeShares(txOpt, encShares, pubCoefsBigFalse)
	if err != nil {
		t.Fatal("Unexpected error arose in DistributeShares submission")
	}

	// Check distribution hash
	authHash, err = c.ShareDistributionHashes(&bind.CallOpts{}, auth.From)
	if err != nil {
		t.Fatal("Error raised when obtaining share distribution hash")
	}
	emptyHash = bytes.Equal(zeroBytes, authHash[:])
	if !emptyHash {
		t.Fatal("Submitted shares resulting in a valid hash; should fail from invalid commitments!")
	}
}
