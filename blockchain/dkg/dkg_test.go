package dkg_test

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// pseudo-constants
var initialMessage []byte = []byte("Hello")

func TestCalculateThreshold(t *testing.T) {
	threshold, _ := dkg.ThresholdForUserCount(4)
	assert.Equal(t, 2, threshold)
}

// func TestNChooseK(t *testing.T) {
// 	n := 10
// 	k, _ := dkg.ThresholdForUserCount(n)
// 	// n := 6
// 	// k := 3

// 	c := make([]int, k+3) // We're just going to ignore index 0

// 	// L1
// 	for j := 1; j <= k; j++ {
// 		c[j] = j - 1
// 	}
// 	c[k+1] = n
// 	c[k+2] = 0

// 	// Loop
// 	done := false
// 	for done == false {

// 		// L2
// 		t.Logf("C...%v", Reverse(c[1:k+1]))

// 		// L3
// 		j := 1
// 		for c[j]+1 == c[j+1] {
// 			c[j] = j - 1
// 			j++
// 		}

// 		// L4
// 		if j > k {
// 			t.Logf("done")
// 			done = true
// 		} else {
// 			// L5
// 			c[j] = c[j] + 1
// 		}

// 	}

// }

func TestGenerateKeys(t *testing.T) {
	private, public, err := dkg.GenerateKeys()
	assert.Nil(t, err, "error generating keys")

	assert.NotNil(t, private, "private key is nil")
	assert.NotNil(t, public, "public key is nil")

	assert.NotNil(t, public[0], "public key missing element")
	assert.NotNil(t, public[1], "public key missing element")
}

func TestGenerateShares(t *testing.T) {

	// Number participants in key generation
	n := 4
	threshold, _ := dkg.ThresholdForUserCount(n)
	assert.Equal(t, 2, threshold)

	// Make n participants
	participants := []*dkg.Participant{}
	for idx := 0; idx < n; idx++ {

		address, _, publicKey := generateTestAddress(t)

		participant := &dkg.Participant{
			Address:   address,
			Index:     idx,
			PublicKey: publicKey}

		participants = append(participants, participant)
	}

	// Overwrite the first
	private, public, _ := dkg.GenerateKeys()
	participants[0].PublicKey = public

	// Now actually generate shares and sanity check them
	encryptedShares, privateCoefficients, commitments, err := dkg.GenerateShares(private, public, participants, threshold)
	assert.Nil(t, err, "error generating shares")
	assert.Equal(t, threshold+1, len(encryptedShares))
	assert.Equal(t, threshold+1, len(privateCoefficients))
	assert.Equal(t, threshold+1, len(commitments))

	t.Logf("encryptedShares:%x privateCoefficients:%x commitments:%x", encryptedShares, privateCoefficients, commitments)
}

func TestGenerateKeyShare(t *testing.T) {

	// Number participants in key generation
	n := 4
	threshold, _ := dkg.ThresholdForUserCount(n)

	// Make n participants
	participants := []*dkg.Participant{{Index: 0}}
	for idx := 0; idx < n; idx++ {

		address, _, publicKey := generateTestAddress(t)

		participant := &dkg.Participant{
			Address:   address,
			Index:     idx,
			PublicKey: publicKey}

		participants = append(participants, participant)
	}

	// Overwrite the first
	private, public, _ := dkg.GenerateKeys()
	participants[0].PublicKey = public

	// Generate shares and sanity check them
	_, privateCoefficients, _, err := dkg.GenerateShares(private, public, participants, threshold)

	// Generate key share and sanity check it
	keyShare1, keyShare1Proof, keyShare2, err := dkg.GenerateKeyShare(privateCoefficients[0])
	assert.Nil(t, err, "error generating key share")
	assert.NotNil(t, keyShare1[0], "key share 1 missing element")
	assert.NotNil(t, keyShare1[1], "key share 1 missing element")

	assert.NotNil(t, keyShare1Proof[0], "key share 1 proof missing element")
	assert.NotNil(t, keyShare1Proof[1], "key share 1 proof missing element")

	assert.NotNil(t, keyShare2[0], "key share 2 missing element")
	assert.NotNil(t, keyShare2[1], "key share 2 missing element")
	assert.NotNil(t, keyShare2[0], "key share 2 missing element")
	assert.NotNil(t, keyShare2[1], "key share 2 missing element")

	t.Logf("keyShare1:%x keyShare1Proof:%x keyShare2:%x", keyShare1, keyShare1Proof, keyShare2)
}

func TestGenerateMasterPublicKey(t *testing.T) {

	// Number participants in key generation
	n := 4
	threshold, _ := dkg.ThresholdForUserCount(n)

	// Make n participants
	privateKeys := make(map[common.Address]*big.Int)
	participants := []*dkg.Participant{{Index: 0}}
	for idx := 0; idx < n; idx++ {

		address, privateKey, publicKey := generateTestAddress(t)

		privateKeys[address] = privateKey
		participant := &dkg.Participant{
			Address:   address,
			Index:     idx,
			PublicKey: publicKey}

		participants = append(participants, participant)
	}

	// Overwrite the first
	private, public, _ := dkg.GenerateKeys()
	participants[0].PublicKey = public
	privateKeys[participants[0].Address] = private

	// Generate encrypted shares on behalf of participants
	encryptedShares := [][]*big.Int{}
	keyShare1s := [][2]*big.Int{}
	keyShare2s := [][4]*big.Int{}
	for _, participant := range participants {
		publicKey := participant.PublicKey
		privateKey := privateKeys[participant.Address]

		participantEncryptedShares, participantPrivateCoefficients, _, err := dkg.GenerateShares(privateKey, publicKey, participants, threshold)
		assert.Nil(t, err)

		keyShare1, _, keyShare2, err := dkg.GenerateKeyShare(participantPrivateCoefficients[0])
		assert.Nil(t, err)

		encryptedShares = append(encryptedShares, participantEncryptedShares)
		keyShare1s = append(keyShare1s, keyShare1)
		keyShare2s = append(keyShare2s, keyShare2)
	}

	// Generate the master public key and sanity check it
	masterPublicKey, err := dkg.GenerateMasterPublicKey(keyShare1s, keyShare2s)
	assert.Nil(t, err)

	assert.NotNil(t, masterPublicKey[0], "missing element of master public key")
	assert.NotNil(t, masterPublicKey[1], "missing element of master public key")
	assert.NotNil(t, masterPublicKey[2], "missing element of master public key")
	assert.NotNil(t, masterPublicKey[3], "missing element of master public key")
}

func TestGenerateGroupKeys(t *testing.T) {

	// Number participants in key generation
	n := 4
	threshold, _ := dkg.ThresholdForUserCount(n)

	// Make n participants
	privateKeys := make(map[common.Address]*big.Int)
	participants := []*dkg.Participant{{Index: 0}}
	for idx := 0; idx < n; idx++ {

		address, privateKey, publicKey := generateTestAddress(t)

		privateKeys[address] = privateKey
		participant := &dkg.Participant{
			Address:   address,
			Index:     idx,
			PublicKey: publicKey}

		participants = append(participants, participant)
	}

	// Overwrite the first
	private, public, _ := dkg.GenerateKeys()
	participants[0].PublicKey = public
	privateKeys[participants[0].Address] = private

	// Generate shares
	_, privateCoefficients, _, err := dkg.GenerateShares(private, public, participants, threshold)
	// keyShare1, keyShare1Proof, keyShare2, err := dkg.GenerateKeyShare(privateCoefficients)

	encryptedShares := [][]*big.Int{}
	// Generate encrypted shares on behalf of participants
	for _, participant := range participants {
		publicKey := participant.PublicKey
		privateKey := privateKeys[participant.Address]

		participantEncryptedShares, _, _, _ := dkg.GenerateShares(privateKey, publicKey, participants, threshold)
		encryptedShares = append(encryptedShares, participantEncryptedShares)
	}

	// Generate the Group Keys and sanity check them
	groupPrivate, groupPublic, groupSignature, err := dkg.GenerateGroupKeys(initialMessage, private, public, privateCoefficients, encryptedShares, 0, participants, threshold)
	assert.Nil(t, err, "error generating key share")
	assert.NotNil(t, groupPrivate, "group private key is missing")
	assert.NotNil(t, groupPublic[0], "group public key element is missing")
	assert.NotNil(t, groupPublic[1], "group public key element is missing")
	assert.NotNil(t, groupPublic[2], "group public key element is missing")
	assert.NotNil(t, groupPublic[3], "group public key element is missing")
	assert.NotNil(t, groupSignature[0], "group signature element is missing")
	assert.NotNil(t, groupSignature[1], "group signature element is missing")

	t.Logf("groupPrivate:%x groupPublic:%x groupSignature:%x", groupPrivate, groupPublic, groupSignature)
}

func TestVerifyGroupSigners(t *testing.T) {

	n := 4
	masterPublicKey, publishedPublicKeys, publishedSignatures, participants, _ := setupGroupSigners(t, n)
	threshold := 3 // Adjusting threshold so verify will look at all signatures

	good, err := dkg.VerifyGroupSigners(initialMessage, masterPublicKey, publishedPublicKeys, publishedSignatures, participants, threshold)
	assert.Nil(t, err, "failed verifying group signers")
	assert.True(t, good, "group signers are all good")
}

func TestVerifyGroupSignersFail(t *testing.T) {

	n := 4
	masterPublicKey, publishedPublicKeys, publishedSignatures, participants, _ := setupGroupSigners(t, n)
	threshold := 3 // Adjusting threshold so verify will look at all signatures

	// Corrupt last signature
	lastSignature := publishedSignatures[n-1]
	lastSignature[0].Add(lastSignature[0], common.Big1) // Not a valid point on the curve so we will fail

	good, err := dkg.VerifyGroupSigners(initialMessage, masterPublicKey, publishedPublicKeys, publishedSignatures, participants, threshold)
	assert.NotNil(t, err, "should have failed verification")
	assert.False(t, good, "a signer is bad")
}

func TestVerifyGroupSignersNegative(t *testing.T) {

	n := 4
	masterPublicKey, publishedPublicKeys, publishedSignatures, participants, _ := setupGroupSigners(t, n)
	threshold := 3 // Adjusting threshold so verify will look at all signatures

	// Replace last signature with a random G1
	_, randomG1, err := cloudflare.RandomG1(rand.Reader)
	badSignature := bn256.G1ToBigIntArray(randomG1) // This will be a valid point but not a valid signature
	publishedSignatures[3][0] = badSignature[0]
	publishedSignatures[3][1] = badSignature[1]

	good, err := dkg.VerifyGroupSigners(initialMessage, masterPublicKey, publishedPublicKeys, publishedSignatures, participants, threshold)
	assert.Nilf(t, err, "failed verifying group signers: %v", err)
	assert.False(t, good, "a signer is bad")
}

func TestCategorizeGroupSigners(t *testing.T) {

	masterPublicKey, publishedPublicKeys, publishedSignatures, participants, threshold := setupGroupSigners(t, 10)

	honest, dishonest, err := dkg.CategorizeGroupSigners(initialMessage, masterPublicKey, publishedPublicKeys, publishedSignatures, participants, threshold)
	assert.Nil(t, err, "failed to categorize group signers")
	assert.Equal(t, len(participants), len(honest), "all participants are honest")
	assert.Equal(t, 0, len(dishonest), "no participants are dishonest")
}

func TestCategorizeGroupSigners1Negative(t *testing.T) {

	n := 30

	logger := logging.GetLogger("dkg")
	logger.SetLevel(logrus.DebugLevel)

	masterPublicKey, publishedPublicKeys, publishedSignatures, participants, threshold := setupGroupSigners(t, n)

	// participants[n-1].Index = n + 1
	participants[0].Index = n + 1

	honest, dishonest, err := dkg.CategorizeGroupSigners(initialMessage, masterPublicKey, publishedPublicKeys, publishedSignatures, participants, threshold)
	assert.Nil(t, err, "failed to categorize group signers")
	assert.Equal(t, len(participants)-1, len(honest), "all but 1 participant are honest")
	assert.Equal(t, 1, len(dishonest), "1 participant is dishonest")
}

func TestCategorizeGroupSigners2Negative(t *testing.T) {

	n := 10

	masterPublicKey, publishedPublicKeys, publishedSignatures, participants, threshold := setupGroupSigners(t, n)

	participants[n-1].Index = n + 1
	participants[n-2].Index = n + 2

	honest, dishonest, err := dkg.CategorizeGroupSigners(initialMessage, masterPublicKey, publishedPublicKeys, publishedSignatures, participants, threshold)
	assert.Nil(t, err, "failed to categorize group signers")
	assert.Equal(t, len(participants)-2, len(honest), "all but 2 participant are honest")
	assert.Equal(t, 2, len(dishonest), "2 participant are dishonest")
}

func TestCategorizeGroupSignersNegative(t *testing.T) {

	n := 15

	// logger := logging.GetLogger("dkg")
	// logger.SetLevel(logrus.WarnLevel)
	masterPublicKey, publishedPublicKeys, publishedSignatures, participants, threshold := setupGroupSigners(t, n)

	for idx := 0; idx < n-threshold-1; idx++ {
		participants[idx].Index = idx + 1 + n
	}

	honest, dishonest, err := dkg.CategorizeGroupSigners(initialMessage, masterPublicKey, publishedPublicKeys, publishedSignatures, participants, threshold)
	assert.Nil(t, err, "failed to categorize group signers")
	assert.Equalf(t, threshold+1, len(honest), "%v participant are honest", threshold+1)
	assert.Equalf(t, n-threshold-1, len(dishonest), "%v participant are dishonest", n-threshold-1)
}

// ---------------------------------------------------------------------------
func generateTestAddress(t *testing.T) (common.Address, *big.Int, [2]*big.Int) {

	// Generating a valid ethereum address
	key, _ := crypto.GenerateKey()
	transactor := bind.NewKeyedTransactor(key)

	// Generate a public key
	privateKey, publicKey, err := dkg.GenerateKeys()
	assert.Nilf(t, err, "failed to generate keys")

	return transactor.From, privateKey, publicKey
}

// ---------------------------------------------------------------------------
func setupGroupSigners(t *testing.T, n int) ([4]*big.Int, [][4]*big.Int, [][2]*big.Int, []*dkg.Participant, int) {

	// Number participants in key generation
	threshold, _ := dkg.ThresholdForUserCount(n)

	// Make n participants
	privateKeys := make(map[common.Address]*big.Int)
	participants := []*dkg.Participant{}

	for idx := 0; idx < n; idx++ {

		address, privateKey, publicKey := generateTestAddress(t)

		privateKeys[address] = privateKey
		participant := &dkg.Participant{
			Address:   address,
			Index:     idx,
			PublicKey: publicKey}

		participants = append(participants, participant)
	}

	// Overwrite the first
	private, public, _ := dkg.GenerateKeys()
	participants[0].PublicKey = public
	privateKeys[participants[0].Address] = private

	// Generate encrypted shares on behalf of participants
	encryptedShares := [][]*big.Int{}
	keyShare1s := [][2]*big.Int{}
	keyShare2s := [][4]*big.Int{}
	privateCoefficients := [][]*big.Int{}

	for _, participant := range participants {
		publicKey := participant.PublicKey
		privateKey := privateKeys[participant.Address]

		participantEncryptedShares, participantPrivateCoefficients, _, err := dkg.GenerateShares(privateKey, publicKey, participants, threshold)
		assert.Nil(t, err)

		keyShare1, _, keyShare2, err := dkg.GenerateKeyShare(participantPrivateCoefficients[0])
		assert.Nil(t, err)

		encryptedShares = append(encryptedShares, participantEncryptedShares)
		privateCoefficients = append(privateCoefficients, participantPrivateCoefficients)
		keyShare1s = append(keyShare1s, keyShare1)
		keyShare2s = append(keyShare2s, keyShare2)
	}

	// Generate the master public key and sanity check it
	masterPublicKey, err := dkg.GenerateMasterPublicKey(keyShare1s, keyShare2s)
	assert.Nil(t, err, "failed to generate master public key")

	publishedPublicKeys := [][4]*big.Int{}
	publishedSignatures := [][2]*big.Int{}
	for idx, participant := range participants {

		publicKey := participant.PublicKey
		privateKey := privateKeys[participant.Address]

		_, groupPublicKey, groupSignature, err := dkg.GenerateGroupKeys(initialMessage, privateKey, publicKey, privateCoefficients[idx], encryptedShares, participant.Index, participants, threshold)
		assert.Nil(t, err, "failed to generate group keys")

		publishedPublicKeys = append(publishedPublicKeys, groupPublicKey)
		publishedSignatures = append(publishedSignatures, groupSignature)
	}

	return masterPublicKey, publishedPublicKeys, publishedSignatures, participants, threshold
}
