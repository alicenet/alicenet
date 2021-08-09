package bn256

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/MadBase/bridge/bindings"
	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// Use whenever submitting transaction to ensure transaction will go through;
// do not use gas estimate!
const gasLim uint64 = 10000000

// Perform the initial setup of the Ethdkg contract for n users and return the
// necessary objects.
func EthdkgContractSetup(t *testing.T, n int) (*bindings.ETHDKG, *bindings.Crypto, *bindings.Validators, interfaces.Ethereum, []*ecdsa.PrivateKey, []*bind.TransactOpts) {
	if n < 1 || n > 6 {
		t.Fatal("Must have at between 1 and 6 user(s) for contract setup")
	}

	wei, ok := new(big.Int).SetString("9000000000000000000000", 10)
	assert.True(t, ok)

	addressPool := []string{
		"546f99f244b7b58b855330ae0e2bc1b30b41302f",
		"9ac1c9afbaec85278679ff75ef109217f26b1417",
		"26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac",
		"615695C4a4D6a60830e5fca4901FbA099DF26271",
		"63a6627b79813A7A43829490C4cE409254f64177",
		"16564cF3e880d9F5d09909F51b922941EbBbC24d"}

	// Build
	eth, err := blockchain.NewEthereumSimulator(
		"../../assets/test/keys",
		"../../assets/test/passcodes.txt",
		1,
		time.Second*1,
		time.Second*5,
		0,
		wei,
		addressPool...)
	assert.Nil(t, err)

	keyArray := make([]*ecdsa.PrivateKey, n)
	authArray := make([]*bind.TransactOpts, n)
	addresses := make([]string, n)
	for k := 0; k < n; k++ {
		addr := common.HexToAddress(addressPool[k])
		assert.Nil(t, err)

		acct, err := eth.GetAccount(addr)
		assert.Nil(t, err)

		assert.Nil(t, eth.UnlockAccount(acct))

		key, err := eth.GetAccountKeys(addr)
		assert.Nil(t, err)

		// 	auth := bind.NewKeyedTransactor(key)
		keyArray[k] = key.PrivateKey
		authArray[k], err = eth.GetTransactionOpts(context.TODO(), acct)
		assert.Nil(t, err)

		addresses[k] = addressPool[k]
	}

	acct := eth.GetDefaultAccount()
	assert.Nil(t, eth.UnlockAccount(acct))

	c := eth.Contracts()
	_, _, err = c.DeployContracts(context.TODO(), acct)
	assert.Nil(t, err)

	stakingToken := c.StakingToken()
	stakingAddr := c.ValidatorsAddress()
	staking := c.Staking()

	ethdkg := c.Ethdkg()
	cryptoContract := c.Crypto()
	validators := c.Validators()

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
	eth.Commit()

	return ethdkg, cryptoContract, validators, eth, keyArray, authArray
}

// Allow for the res
func InitialTestSetup(t *testing.T, n int) (*bindings.ETHDKG, *bindings.Crypto, interfaces.Ethereum, []*ecdsa.PrivateKey, []*bind.TransactOpts, []*big.Int, []*cloudflare.G1) {
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
	curBlock := CurrentBlock(sim)
	assert.Truef(t, curBlock.Cmp(registrationEnd) <= 0, "Registration ended. Current:%v End:%v", curBlock, registrationEnd)

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

func AdvanceBlocksUntil(eth interfaces.Ethereum, m *big.Int) {
	ctx := context.TODO()

	height, _ := eth.GetCurrentHeight(ctx)
	current := new(big.Int).SetUint64(height)

	for current.Cmp(m) <= 0 {
		eth.Commit()

		height, _ = eth.GetCurrentHeight(ctx)
		current = new(big.Int).SetUint64(height)
	}
}

func CurrentBlock(eth interfaces.Ethereum) *big.Int {
	height, _ := eth.GetCurrentHeight(context.TODO())
	return new(big.Int).SetUint64(height)
}

func thresholdFromUsers(n int) (int, int) {
	k := n / 3
	threshold := 2 * k
	if (n - 3*k) == 2 {
		threshold = threshold + 1
	}
	return threshold, k
}
