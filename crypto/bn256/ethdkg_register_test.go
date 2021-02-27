package bn256

import (
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// Here we correctly register participants and then confirm the registrations
// occurred correctly.
func TestRegister(t *testing.T) {
	n := 4
	c, _, v, sim, _, authArray := EthdkgContractSetup(t, n)
	defer sim.Close()

	big0 := big.NewInt(0)
	big1 := big.NewInt(1)

	_, err := c.InitializeState(authArray[0])
	assert.Nil(t, err)
	sim.Commit()

	// Check block number here
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Error in getting RegistrationEnd")
	}
	curBlock := CurrentBlock(sim)
	validBlockNumber := curBlock.Cmp(registrationEnd) <= 0
	if !validBlockNumber {
		t.Fatal("Unexpected error; in Registration Phase")
	}

	privKArray := make([]*big.Int, n)
	pubKArray := make([]*cloudflare.G1, n)
	baseKey, _ := new(big.Int).SetString("1234567890", 10)
	for k, auth := range authArray {
		privK := new(big.Int).Add(baseKey, big.NewInt(int64(k)))
		pubK := new(cloudflare.G1).ScalarBaseMult(privK)
		privKArray[k] = privK
		pubKArray[k] = pubK
		pubKBig := G1ToBigIntArray(pubK)
		txOpt := &bind.TransactOpts{
			From:     auth.From,
			Nonce:    nil,
			Signer:   auth.Signer,
			Value:    nil,
			GasPrice: nil,
			GasLimit: gasLim,
			Context:  nil,
		}

		// Check public_key to ensure not registered
		pubKBigRcvd0, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (0)")
		}
		pubKBigRcvd1, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (1)")
		}
		registeredPubK := (pubKBigRcvd0.Cmp(big0) != 0) || (pubKBigRcvd1.Cmp(big0) != 0)
		if registeredPubK {
			t.Fatal("Public Key already exists")
		}

		_, err = v.AddValidator(txOpt, txOpt.From, pubKBig)
		assert.Nilf(t, err, "Failed to add validator prior to registering: %v", err)

		_, err = c.Register(txOpt, pubKBig)
		assert.Nilf(t, err, "Something went wrong with c.Register: %v", err)
	}
	sim.Commit()

	// Confirm submission
	for k, auth := range authArray {
		pubK := pubKArray[k]
		pubKBig := G1ToBigIntArray(pubK)
		pubKBig0 := pubKBig[0]
		pubKBig1 := pubKBig[1]

		// Check public_key to ensure not registered
		pubKBigRcvd0, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
		assert.Nilf(t, err, "Something went wrong with c.PublicKeys (0)")

		pubKBigRcvd1, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
		assert.Nilf(t, err, "Something went wrong with c.PublicKeys (1)")

		// Compare with submitted correct value
		submittedPubKMatch := (pubKBigRcvd0.Cmp(pubKBig0) == 0) && (pubKBigRcvd1.Cmp(pubKBig1) == 0)
		assert.Truef(t, submittedPubKMatch, "Public Key incorrect!")
	}
}

// We attempt to register after the registration phase and are unable
// to do so.
func TestRegisterFailRegisterLate(t *testing.T) {
	n := 1
	c, _, _, sim, _, authArray := EthdkgContractSetup(t, n)
	defer sim.Close()

	big0 := big.NewInt(0)
	big1 := big.NewInt(1)

	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Error in getting RegistrationEnd")
	}
	AdvanceBlocksUntil(sim, registrationEnd)
	// Current block number is now 22 > 21 == T_REGISTRATION_END;
	// no additional registration allowed

	// Check block number here; we are not in the correct block
	curBlock := CurrentBlock(sim)
	validBlockNumber := curBlock.Cmp(registrationEnd) <= 0
	if validBlockNumber {
		t.Fatal("Unexpected error; not in Registration Phase")
	}

	auth := authArray[0]
	privK, _ := new(big.Int).SetString("1234567890", 10)
	pubK := new(cloudflare.G1).ScalarBaseMult(privK)
	pubKBig := G1ToBigIntArray(pubK)
	txOpt := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	// Check public_key to ensure not registered
	pubKBigRcvd0, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Something went wrong with c.PublicKeys (0)")
	}
	pubKBigRcvd1, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Something went wrong with c.PublicKeys (1)")
	}
	registeredPubK := (pubKBigRcvd0.Cmp(big0) != 0) || (pubKBigRcvd1.Cmp(big0) != 0)
	if registeredPubK {
		t.Fatal("Public Key already exists")
	}

	_, err = c.Register(txOpt, pubKBig)
	if err != nil {
		t.Fatal("Unexpected error in c.Register")
	}
	sim.Commit()

	// Check public_key to ensure not registered
	pubKBigRcvd0, err = c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Something went wrong with c.PublicKeys (0)")
	}
	pubKBigRcvd1, err = c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Something went wrong with c.PublicKeys (1)")
	}
	// Compare with 0 for confirmation
	submissionFailed := (pubKBigRcvd0.Cmp(big0) == 0) && (pubKBigRcvd1.Cmp(big0) == 0)
	if !submissionFailed {
		t.Fatal("Public Key is nonzero!")
	}
}

// Here we attempt to register twice (with different public keys)
// in order to show that the second attempt will fail.
func TestRegisterFailRegisterTwice(t *testing.T) {
	n := 1
	c, _, v, sim, _, authArray := EthdkgContractSetup(t, n)
	defer sim.Close()

	big0 := big.NewInt(0)
	big1 := big.NewInt(1)

	// Start round
	_, err := c.InitializeState(authArray[0])
	assert.Nil(t, err)
	sim.Commit()

	// Check block number here
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Fatal("Error in getting RegistrationEnd")
	}
	curBlock := CurrentBlock(sim)
	validBlockNumber := curBlock.Cmp(registrationEnd) <= 0
	if !validBlockNumber {
		t.Fatal("Unexpected error; in Registration Phase")
	}

	auth := authArray[0]
	privK, _ := new(big.Int).SetString("1234567890", 10)
	pubK := new(cloudflare.G1).ScalarBaseMult(privK)
	pubKBig := G1ToBigIntArray(pubK)
	txOpt := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	// Check public_key to ensure not registered
	pubKBigRcvd0, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Something went wrong with c.PublicKeys (0)")
	}
	pubKBigRcvd1, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Something went wrong with c.PublicKeys (1)")
	}
	registeredPubK := (pubKBigRcvd0.Cmp(big0) != 0) || (pubKBigRcvd1.Cmp(big0) != 0)
	if registeredPubK {
		t.Fatal("Public Key already exists")
	}

	_, err = v.AddValidator(txOpt, txOpt.From, pubKBig)
	assert.Nilf(t, err, "Failed to add validator prior to registering: %v", err)

	_, err = c.Register(txOpt, pubKBig)
	if err != nil {
		t.Fatal("Something went wrong with c.Register")
	}
	sim.Commit()

	// Create new pubK and attempt to submit
	privKNew := new(big.Int).Add(privK, big1)
	pubKNew := new(cloudflare.G1).ScalarBaseMult(privKNew)
	pubKNewBig := G1ToBigIntArray(pubKNew)

	// Attempt to re-submit
	_, err = c.Register(txOpt, pubKNewBig)
	if err != nil {
		t.Fatal("Something went wrong with c.Register")
	}
	sim.Commit()

	// Check public_key to ensure not registered; will fail this time
	pubKNewBigRcvd0, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Something went wrong with c.PublicKeys")
	}
	pubKNewBigRcvd1, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Something went wrong with c.PublicKeys")
	}
	changedPubK := (pubKNewBigRcvd0.Cmp(pubKBig[0]) != 0) || (pubKNewBigRcvd1.Cmp(pubKBig[1]) != 0)
	if changedPubK {
		t.Fatal("Unexpected error; Public Key should not have changed")
	}
}

// We attempt to register with an invalid (off-curve) public key.
func TestRegisterFailInvalidG1PubKey(t *testing.T) {
	n := 1
	c, _, _, sim, _, authArray := EthdkgContractSetup(t, n)
	defer sim.Close()

	big0 := big.NewInt(0)
	big1 := big.NewInt(1)

	// Start round
	_, err := c.InitializeState(authArray[0])
	assert.Nil(t, err)
	sim.Commit()

	// Check block number here
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Error in getting RegistrationEnd")
	}
	curBlock := CurrentBlock(sim)
	validBlockNumber := curBlock.Cmp(registrationEnd) <= 0
	if !validBlockNumber {
		t.Fatal("Unexpected error; in Registration Phase")
	}

	auth := authArray[0]
	pubKBig := [2]*big.Int{big.NewInt(1), big.NewInt(3)} // Invalid G1 point
	txOpt := &bind.TransactOpts{
		From:     auth.From,
		Nonce:    nil,
		Signer:   auth.Signer,
		Value:    nil,
		GasPrice: nil,
		GasLimit: gasLim,
		Context:  nil,
	}

	// Check public key to ensure not registered
	pubKBigRcvd0, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Something went wrong with c.PublicKeys (0)")
	}
	pubKBigRcvd1, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Something went wrong with c.PublicKeys (1)")
	}
	registeredPubK := (pubKBigRcvd0.Cmp(big0) != 0) || (pubKBigRcvd1.Cmp(big0) != 0)
	if registeredPubK {
		t.Fatal("Public Key already exists")
	}

	_, err = c.Register(txOpt, pubKBig)
	if err != nil {
		t.Error("Error in c.Register call")
	}
	sim.Commit()

	// Confirm public was not registered
	pubKBigRcvd0, err = c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
	if err != nil {
		t.Fatal("Something went wrong with c.PublicKeys (0)")
	}
	pubKBigRcvd1, err = c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
	if err != nil {
		t.Fatal("Something went wrong with c.PublicKeys (1)")
	}
	validPubK := (pubKBigRcvd0.Cmp(big0) != 0) || (pubKBigRcvd1.Cmp(big0) != 0)
	if validPubK {
		t.Fatal("Public Key should not be valid")
	}
}
