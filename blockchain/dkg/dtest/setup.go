package dtest

import (
	"context"
	"crypto/ecdsa"
	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/stretchr/testify/assert"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"

	dkgMath "github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/etest"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
)

const SETUP_GROUP int = 13

func InitializeNewDetDkgStateInfo(n int) ([]*objects.DkgState, []*ecdsa.PrivateKey) {
	return InitializeNewDkgStateInfo(n, true)
}

func InitializeNewNonDetDkgStateInfo(n int) ([]*objects.DkgState, []*ecdsa.PrivateKey) {
	return InitializeNewDkgStateInfo(n, false)
}

func InitializeNewDkgStateInfo(n int, deterministicShares bool) ([]*objects.DkgState, []*ecdsa.PrivateKey) {
	// Get private keys for validators
	privKeys := etest.SetupPrivateKeys(n)
	accountsArray := etest.SetupAccounts(privKeys)
	dkgStates := []*objects.DkgState{}
	threshold := crypto.CalcThreshold(n)

	// Make base for secret key
	baseSecretBytes := make([]byte, 32)
	baseSecretBytes[0] = 101
	baseSecretBytes[31] = 101
	baseSecretValue := new(big.Int).SetBytes(baseSecretBytes)

	// Make base for transport key
	baseTransportBytes := make([]byte, 32)
	baseTransportBytes[0] = 1
	baseTransportBytes[1] = 1
	baseTransportValue := new(big.Int).SetBytes(baseTransportBytes)

	// Beginning dkgState initialization
	for k := 0; k < n; k++ {
		bigK := big.NewInt(int64(k))
		// Get base DkgState
		dkgState := objects.NewDkgState(accountsArray[k])
		// Set Index
		dkgState.Index = k + 1
		// Set Number of Validators
		dkgState.NumberOfValidators = n
		dkgState.ValidatorThreshold = threshold

		// Setup TransportKey
		transportPrivateKey := new(big.Int).Add(baseTransportValue, bigK)
		dkgState.TransportPrivateKey = transportPrivateKey
		transportPublicKeyG1 := new(cloudflare.G1).ScalarBaseMult(dkgState.TransportPrivateKey)
		transportPublicKey, err := bn256.G1ToBigIntArray(transportPublicKeyG1)
		if err != nil {
			panic(err)
		}
		dkgState.TransportPublicKey = transportPublicKey

		// Append to state array
		dkgStates = append(dkgStates, dkgState)
	}

	// Generate Participants
	for k := 0; k < n; k++ {
		participantList := GenerateParticipantList(dkgStates)
		for _, p := range participantList {
			dkgStates[k].Participants[p.Address] = p
		}
	}

	// Prepare secret shares
	for k := 0; k < n; k++ {
		bigK := big.NewInt(int64(k))
		// Set SecretValue and PrivateCoefficients
		dkgState := dkgStates[k]
		if deterministicShares {
			// Deterministic shares
			secretValue := new(big.Int).Add(baseSecretValue, bigK)
			privCoefs := GenerateDeterministicPrivateCoefficients(n)
			privCoefs[0].Set(secretValue) // Overwrite constant term
			dkgState.SecretValue = secretValue
			dkgState.PrivateCoefficients = privCoefs
		} else {
			// Random shares
			_, privCoefs, _, err := dkgMath.GenerateShares(dkgState.TransportPrivateKey, dkgState.GetSortedParticipants())
			if err != nil {
				panic(err)
			}
			dkgState.SecretValue = new(big.Int)
			dkgState.SecretValue.Set(privCoefs[0])
			dkgState.PrivateCoefficients = privCoefs
		}
	}

	return dkgStates, privKeys
}

func GenerateParticipantList(dkgStates []*objects.DkgState) objects.ParticipantList {
	n := len(dkgStates)
	participants := make(objects.ParticipantList, int(n))
	for idx := 0; idx < n; idx++ {
		addr := dkgStates[idx].Account.Address
		publicKey := [2]*big.Int{}
		publicKey[0] = new(big.Int)
		publicKey[1] = new(big.Int)
		publicKey[0].Set(dkgStates[idx].TransportPublicKey[0])
		publicKey[1].Set(dkgStates[idx].TransportPublicKey[1])
		participant := &objects.Participant{}
		participant.Address = addr
		participant.PublicKey = publicKey
		participant.Index = dkgStates[idx].Index
		participants[idx] = participant
	}
	return participants
}

func GenerateEncryptedSharesAndCommitments(dkgStates []*objects.DkgState) {
	n := len(dkgStates)
	for k := 0; k < n; k++ {
		dkgState := dkgStates[k]
		publicCoefs := GeneratePublicCoefficients(dkgState.PrivateCoefficients)
		encryptedShares := GenerateEncryptedShares(dkgStates, k)
		// Loop through entire list and save in map
		for ell := 0; ell < n; ell++ {
			dkgStates[ell].Participants[dkgState.Account.Address].Commitments = publicCoefs
			dkgStates[ell].Participants[dkgState.Account.Address].EncryptedShares = encryptedShares
		}
	}
}

func GenerateDeterministicPrivateCoefficients(n int) []*big.Int {
	threshold := crypto.CalcThreshold(n)
	privCoefs := []*big.Int{}
	privCoefs = append(privCoefs, big.NewInt(0))
	for k := 1; k <= threshold; k++ {
		privCoef := big.NewInt(1)
		privCoefs = append(privCoefs, privCoef)
	}
	return privCoefs
}

func GeneratePublicCoefficients(privCoefs []*big.Int) [][2]*big.Int {
	publicCoefsG1 := cloudflare.GeneratePublicCoefs(privCoefs)
	publicCoefs := [][2]*big.Int{}
	for k := 0; k < len(publicCoefsG1); k++ {
		coefG1 := publicCoefsG1[k]
		coef, err := bn256.G1ToBigIntArray(coefG1)
		if err != nil {
			panic(err)
		}
		publicCoefs = append(publicCoefs, coef)
	}
	return publicCoefs
}

func GenerateEncryptedShares(dkgStates []*objects.DkgState, idx int) []*big.Int {
	dkgState := dkgStates[idx]
	// Get array of public keys and convert to cloudflare.G1
	publicKeysBig := [][2]*big.Int{}
	for k := 0; k < len(dkgStates); k++ {
		publicKeysBig = append(publicKeysBig, dkgStates[k].TransportPublicKey)
	}
	publicKeysG1, err := bn256.BigIntArraySliceToG1(publicKeysBig)
	if err != nil {
		panic(err)
	}

	// Get public key for caller
	publicKeyBig := dkgState.TransportPublicKey
	publicKey, err := bn256.BigIntArrayToG1(publicKeyBig)
	if err != nil {
		panic(err)
	}
	privCoefs := dkgState.PrivateCoefficients
	secretShares, err := cloudflare.GenerateSecretShares(publicKey, privCoefs, publicKeysG1)
	if err != nil {
		panic(err)
	}
	encryptedShares, err := cloudflare.GenerateEncryptedShares(secretShares, dkgState.TransportPrivateKey, publicKeysG1)
	if err != nil {
		panic(err)
	}
	return encryptedShares
}

func GenerateKeyShares(dkgStates []*objects.DkgState) {
	n := len(dkgStates)
	for k := 0; k < n; k++ {
		dkgState := dkgStates[k]
		g1KeyShare, g1Proof, g2KeyShare, err := dkgMath.GenerateKeyShare(dkgState.SecretValue)
		if err != nil {
			panic(err)
		}
		addr := dkgState.Account.Address
		// Loop through entire list and save in map
		for ell := 0; ell < n; ell++ {
			dkgStates[ell].Participants[addr].KeyShareG1s = g1KeyShare
			dkgStates[ell].Participants[addr].KeyShareG1CorrectnessProofs = g1Proof
			dkgStates[ell].Participants[addr].KeyShareG2s = g2KeyShare
		}
	}
}

// GenerateMasterPublicKey computes the mpk for the protocol.
// This computes this by using all of the secret values from dkgStates.
func GenerateMasterPublicKey(dkgStates []*objects.DkgState) []*objects.DkgState {
	n := len(dkgStates)
	msk := new(big.Int)
	for k := 0; k < n; k++ {
		msk.Add(msk, dkgStates[k].SecretValue)
	}
	msk.Mod(msk, cloudflare.Order)
	for k := 0; k < n; k++ {
		mpkG2 := new(cloudflare.G2).ScalarBaseMult(msk)
		mpk, err := bn256.G2ToBigIntArray(mpkG2)
		if err != nil {
			panic(err)
		}
		dkgStates[k].MasterPublicKey = mpk
	}
	return dkgStates
}

func GenerateGPKJ(dkgStates []*objects.DkgState) {
	n := len(dkgStates)
	for k := 0; k < n; k++ {
		dkgState := dkgStates[k]

		encryptedShares := make([][]*big.Int, n)
		for idx, participant := range dkgState.GetSortedParticipants() {
			p, present := dkgState.Participants[participant.Address]
			if present && idx >= 0 && idx < n {
				encryptedShares[idx] = p.EncryptedShares
			} else {
				panic("Encrypted share state broken")
			}
		}

		groupPrivateKey, groupPublicKey, err := dkgMath.GenerateGroupKeys(dkgState.TransportPrivateKey, dkgState.PrivateCoefficients,
			encryptedShares, dkgState.Index, dkgState.GetSortedParticipants())
		if err != nil {
			panic("Could not generate group keys")
		}

		dkgState.GroupPrivateKey = groupPrivateKey

		// Loop through entire list and save in map
		for ell := 0; ell < n; ell++ {
			dkgStates[ell].Participants[dkgState.Account.Address].GPKj = groupPublicKey
		}
	}
}

func InitializePrivateKeysAndAccounts(n int) ([]*ecdsa.PrivateKey, []accounts.Account) {
	ecdsaPrivateKeys := etest.SetupPrivateKeys(n)
	accounts := etest.SetupAccounts(ecdsaPrivateKeys)

	return ecdsaPrivateKeys, accounts
}

func ConnectSimulatorEndpoint(t *testing.T, privateKeys []*ecdsa.PrivateKey, blockInterval time.Duration) interfaces.Ethereum {
	eth, err := blockchain.NewEthereumSimulator(
		privateKeys,
		6,
		1*time.Second,
		5*time.Second,
		0,
		big.NewInt(math.MaxInt64),
		50,
		math.MaxInt64,
		5*time.Second,
		30*time.Second)

	assert.Nil(t, err, "Failed to build Ethereum endpoint...")
	assert.True(t, eth.IsEthereumAccessible(), "Web3 endpoint is not available.")

	// Mine a block once a second
	if blockInterval > 1*time.Millisecond {
		go func() {
			for {
				time.Sleep(blockInterval)
				eth.Commit()
			}
		}()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Unlock the default account and use it to deploy contracts
	deployAccount := eth.GetDefaultAccount()
	err = eth.UnlockAccount(deployAccount)
	assert.Nil(t, err, "Failed to unlock default account")

	// Deploy all the contracts
	c := eth.Contracts()
	_, _, err = c.DeployContracts(ctx, deployAccount)
	assert.Nil(t, err, "Failed to deploy contracts...")

	// For each address passed set them up as a validator
	txnOpts, err := eth.GetTransactionOpts(ctx, deployAccount)
	assert.Nil(t, err, "Failed to create txn opts")

	accountList := eth.GetKnownAccounts()
	for idx, acct := range accountList {
		//	for idx := 1; idx < len(accountAddresses); idx++ {
		// acct, err := eth.GetAccount(common.HexToAddress(accountAddresses[idx]))
		//assert.Nil(t, err)
		err = eth.UnlockAccount(acct)
		assert.Nil(t, err)

		eth.Commit()
		t.Logf("# unlocked %v of %v", idx+1, len(accountList))

		// 1. Give 'acct' tokens
		txn, err := c.StakingToken().Transfer(txnOpts, acct.Address, big.NewInt(10_000_000))
		assert.Nilf(t, err, "Failed on transfer %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)

		o, err := eth.GetTransactionOpts(ctx, acct)
		assert.Nil(t, err)

		// 2. Allow system to take tokens from 'acct' for staking
		txn, err = c.StakingToken().Approve(o, c.ValidatorsAddress(), big.NewInt(10_000_000))
		assert.Nilf(t, err, "Failed on approval %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)

		// 3. Tell system to take tokens from 'acct' for staking
		txn, err = c.Staking().LockStake(o, big.NewInt(1_000_000))
		assert.Nilf(t, err, "Failed on lock %v", idx)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)

		txn, err = c.ValidatorPool().AddValidator(o, acct.Address)
		assert.Nilf(t, err, "Failed on register %v", idx)
		assert.NotNil(t, txn)
		eth.Queue().QueueGroupTransaction(ctx, SETUP_GROUP, txn)
		t.Logf("Finished loop %v of %v", idx+1, len(accountList))
		eth.Commit()
	}

	// Wait for all transactions for all accounts to complete
	rcpts, err := eth.Queue().WaitGroupTransactions(ctx, SETUP_GROUP)
	assert.Nil(t, err)

	// Make sure all transactions were successful
	t.Logf("# rcpts: %v", len(rcpts))
	for _, rcpt := range rcpts {
		assert.Equal(t, uint64(1), rcpt.Status)
	}

	return eth
}
