package bn256

import (
	"crypto/ecdsa"
	"log"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/MadBase/bridge/bindings"
	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

// Use whenever submitting transaction to ensure transaction will go through;
// do not use gas estimate!
const gasLim uint64 = 10000000

// Perform the initial setup of the Ethdkg contract for n users and return the
// necessary objects.
func EthdkgContractSetup(t *testing.T, n int) (*bindings.ETHDKG, *bindings.Crypto, *bindings.Validators, *backends.SimulatedBackend, []*ecdsa.PrivateKey, []*bind.TransactOpts) {
	if n < 1 {
		t.Fatal("Must have at least 1 user for contract setup")
	}
	// Generate a new random account and a funded simulator
	gasLimit := uint64(10000000000000000)
	genAlloc := make(core.GenesisAlloc)
	keyArray := make([]*ecdsa.PrivateKey, n)
	authArray := make([]*bind.TransactOpts, n)
	for k := 0; k < n; k++ {
		key, _ := crypto.GenerateKey()
		auth := bind.NewKeyedTransactor(key)
		genAlloc[auth.From] = core.GenesisAccount{Balance: big.NewInt(9223372036854775807)}
		keyArray[k] = key
		authArray[k] = auth
	}

	sim := backends.NewSimulatedBackend(genAlloc, gasLimit) // Deploy a token contract on the simulated blockchain

	registryAddress, _, registry, err := bindings.DeployRegistry(authArray[0], sim)
	if err != nil {
		log.Fatalf("Failed to deploy new contract 1: %v", err)
	}

	stakingTokenAddr, _, stakingToken, err := bindings.DeployToken(authArray[0], sim,
		blockchain.StringToBytes32("STK"), blockchain.StringToBytes32("MadNet Staking"))
	if err != nil {
		log.Fatalf("Failed to deploy new contract 2: %v", err)
	}

	utilityTokenAddr, _, _, err := bindings.DeployToken(authArray[0], sim,
		blockchain.StringToBytes32("UTL"), blockchain.StringToBytes32("MadNet Utility"))
	if err != nil {
		log.Fatalf("Failed to deploy new contract 2: %v", err)
	}

	cryptoAddr, _, cryptoContract, err := bindings.DeployCrypto(authArray[0], sim)
	if err != nil {
		log.Fatalf("Failed to deploy new contract 3: %v", err)
	}

	depositAddr, _, _, err := bindings.DeployDeposit(authArray[0], sim, registryAddress)
	if err != nil {
		log.Fatalf("Failed to deploy new contract 4: %v", err)
	}

	stakingAddr, _, staking, err := bindings.DeployStaking(authArray[0], sim, registryAddress)
	if err != nil {
		log.Fatalf("Failed to deploy new contract 5: %v", err)
	}

	validatorsAddr, _, validators, err := bindings.DeployValidators(authArray[0], sim, 10, registryAddress)
	if err != nil {
		log.Fatalf("Failed to deploy new contract 6: %v", err)
	}

	validatorsSnapshotAddr, _, _, err := bindings.DeployValidatorsSnapshot(authArray[0], sim)
	if err != nil {
		log.Fatalf("Failed to deploy new contract 6: %v", err)
	}

	ethdkgAddr, _, ethdkg, err := bindings.DeployETHDKG(authArray[0], sim, registryAddress) // Contract transaction submitted to chain
	if err != nil {
		log.Fatalf("Failed to deploy new contract 7: %v", err)
	}

	ethdkgCompletionAddress, _, _, err := bindings.DeployETHDKGCompletion(authArray[0], sim)
	if err != nil {
		log.Fatalf("Failed to deploy new contract 8: %v", err)
	}

	ethdkgGroupAccusationAddress, _, _, err := bindings.DeployETHDKGGroupAccusation(authArray[0], sim)
	if err != nil {
		log.Fatalf("Failed to deploy new contract 9: %v", err)
	}

	ethdkgSubmitMPKAddress, _, _, err := bindings.DeployETHDKGSubmitMPK(authArray[0], sim)
	if err != nil {
		log.Fatalf("Failed to deploy new contract 9: %v", err)
	}

	sim.Commit() // Simulated blockchain moves forward and advances block number by 1

	_, err = registry.Register(authArray[0], "crypto/v1", cryptoAddr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Register(authArray[0], "deposit/v1", depositAddr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Register(authArray[0], "ethdkg/v1", ethdkgAddr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Register(authArray[0], "ethdkgCompletion/v1", ethdkgCompletionAddress)
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Register(authArray[0], "ethdkgGroupAccusation/v1", ethdkgGroupAccusationAddress)
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Register(authArray[0], "ethdkgSubmitMPK/v1", ethdkgSubmitMPKAddress)
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Register(authArray[0], "staking/v1", stakingAddr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Register(authArray[0], "stakingToken/v1", stakingTokenAddr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Register(authArray[0], "utilityToken/v1", utilityTokenAddr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Register(authArray[0], "validators/v1", validatorsAddr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Register(authArray[0], "validatorsSnapshot/v1", validatorsSnapshotAddr)
	if err != nil {
		t.Fatal(err)
	}
	sim.Commit()

	_, err = ethdkg.ReloadRegistry(authArray[0])
	if err != nil {
		t.Fatal(err)
	}
	_, err = staking.ReloadRegistry(authArray[0])
	if err != nil {
		t.Fatal(err)
	}
	_, err = validators.ReloadRegistry(authArray[0])
	if err != nil {
		t.Fatal(err)
	}

	//
	initialTokenBalance := big.NewInt(9000000000000000000)
	initialTokenApproval := big.NewInt(4000000000000000000)
	initialTokenStake := big.NewInt(1000000000000000000)

	// Setup token and staking
	for idx := 0; idx < len(authArray); idx++ {

		if idx > 0 {
			_, err := stakingToken.Transfer(authArray[0], authArray[idx].From, initialTokenBalance)
			assert.Nilf(t, err, "Token transfer failed: %v", err)
		}

		_, err = stakingToken.Approve(authArray[idx], stakingAddr, initialTokenApproval)
		assert.Nil(t, err, "Token staking approval failed")

		// balance, err := token.BalanceOf(&bind.CallOpts{}, authArray[idx].From)
		// assert.Nil(t, err, "balance failed")

		// allowance, err := token.Allowance(&bind.CallOpts{}, authArray[idx].From, stakingAddr)
		// assert.Nil(t, err, "allowance failed")

		_, err = staking.LockStake(authArray[idx], initialTokenStake)
		assert.Nil(t, err, "Locking stake failed")
	}
	sim.Commit()

	return ethdkg, cryptoContract, validators, sim, keyArray, authArray
}

// Allow for the res
func InitialTestSetup(t *testing.T, n int) (*bindings.ETHDKG, *bindings.Crypto, *backends.SimulatedBackend, []*ecdsa.PrivateKey, []*bind.TransactOpts, []*big.Int, []*cloudflare.G1) {
	c, cc, v, sim, keyArray, authArray := EthdkgContractSetup(t, n)

	big0 := big.NewInt(0)
	big1 := big.NewInt(1)

	_, err := c.InitializeState(authArray[0])
	assert.Nil(t, err, "Failed to initialize ETHDKG")

	sim.Commit()

	// Check block number here
	registrationEnd, err := c.TREGISTRATIONEND(&bind.CallOpts{})
	if err != nil {
		t.Error("Error in getting RegistrationEnd")
	}
	curBlock := sim.Blockchain().CurrentBlock().Number()
	assert.Truef(t, curBlock.Cmp(registrationEnd) <= 0, "Registration ended. Current:%v End:%v", curBlock, registrationEnd)
	// validBlockNumber := (curBlock.Cmp(registrationEnd) <= 0)
	// if !validBlockNumber {
	// 	t.Fatal("Unexpected error; in Registration Phase")
	// }

	privKArray := make([]*big.Int, n)
	pubKArray := make([]*cloudflare.G1, n)
	baseKey, _ := new(big.Int).SetString("10", 10)
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
			GasLimit: 0,
			Context:  nil,
		}

		// Check public_key to ensure not registered
		pubKBigInt0, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys")
		}
		registeredPubK := pubKBigInt0.Cmp(big0) != 0
		if registeredPubK {
			t.Fatal("Public Key already exists")
		}
		_, err = v.AddValidator(txOpt, txOpt.From, pubKBig)
		assert.Nilf(t, err, "Something went wrong with c.AddValidator")

		_, err = c.Register(txOpt, pubKBig)
		assert.Nilf(t, err, "Something went wrong with c.Register")
	}
	sim.Commit()

	// Confirm submissions are valid
	for k, auth := range authArray {
		pubK := pubKArray[k]
		pubKBig := G1ToBigIntArray(pubK)
		pubKBig0 := pubKBig[0]
		pubKBig1 := pubKBig[1]

		// Check public_key to ensure not registered
		pubKBigInt0, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big0)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (0)")
		}
		pubKBigInt1, err := c.PublicKeys(&bind.CallOpts{}, auth.From, big1)
		if err != nil {
			t.Fatal("Something went wrong with c.PublicKeys (1)")
		}
		// Compare with submitted correct value
		submittedPubKMatch := (pubKBigInt0.Cmp(pubKBig0) == 0) && (pubKBigInt1.Cmp(pubKBig1) == 0)
		if !submittedPubKMatch {
			t.Fatal("Public Key incorrect!")
		}
	}
	return c, cc, sim, keyArray, authArray, privKArray, pubKArray
}

func AdvanceBlocksUntil(sim *backends.SimulatedBackend, m *big.Int) {
	for sim.Blockchain().CurrentBlock().Number().Cmp(m) <= 0 {
		sim.Commit()
	}
}

func thresholdFromUsers(n int) (int, int) {
	k := n / 3
	threshold := 2 * k
	if (n - 3*k) == 2 {
		threshold = threshold + 1
	}
	return threshold, k
}
