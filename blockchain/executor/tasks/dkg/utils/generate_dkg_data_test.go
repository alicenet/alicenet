package utils

import (
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/objects"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestMath_CalculateThreshold(t *testing.T) {
	threshold := ThresholdForUserCount(4)
	assert.Equal(t, 2, threshold)
	threshold = ThresholdForUserCount(5)
	assert.Equal(t, 3, threshold)
	threshold = ThresholdForUserCount(6)
	assert.Equal(t, 4, threshold)
	threshold = ThresholdForUserCount(7)
	assert.Equal(t, 4, threshold)
	threshold = ThresholdForUserCount(8)
	assert.Equal(t, 5, threshold)
	threshold = ThresholdForUserCount(9)
	assert.Equal(t, 6, threshold)
}

func TestMath_InverseArrayForUserCount(t *testing.T) {
	n := 3
	_, err := InverseArrayForUserCount(n)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	n = 10
	invArray, err := InverseArrayForUserCount(n)
	if err != nil {
		t.Fatal(err)
	}
	if len(invArray) != n-1 {
		t.Fatal("Incorrect array length")
	}
	big1 := big.NewInt(1)
	for idx := 0; idx < n-1; idx++ {
		k := idx + 1
		kBig := big.NewInt(int64(k))
		kInv := invArray[idx]
		res := new(big.Int).Mul(kBig, kInv)
		res.Mod(res, cloudflare.Order)
		if res.Cmp(big1) != 0 {
			t.Fatal("invalid inverse")
		}
	}
}

func TestMath_GenerateKeys(t *testing.T) {
	private, public, err := GenerateKeys()
	assert.Nil(t, err, "error generating keys")

	assert.NotNil(t, private, "private key is nil")
	assert.NotNil(t, public, "public key is nil")

	assert.NotNil(t, public[0], "public key missing element")
	assert.NotNil(t, public[1], "public key missing element")
}

func TestMath_GenerateShares(t *testing.T) {
	// Number participants in key generation
	n := 4
	threshold := ThresholdForUserCount(n)
	assert.Equal(t, 2, threshold)

	// Make n participants
	participants := []*objects.Participant{}
	for idx := 0; idx < n; idx++ {

		address, _, publicKey := generateTestAddress(t)

		participant := &objects.Participant{
			Address:   address,
			Index:     idx + 1,
			PublicKey: publicKey}

		participants = append(participants, participant)
	}

	// Overwrite the first
	private, public, _ := GenerateKeys()
	participants[0].PublicKey = public

	// Now actually generate shares and sanity check them
	encryptedShares, privateCoefficients, commitments, err := GenerateShares(private, participants)
	assert.Nil(t, err, "error generating shares")
	assert.Equal(t, threshold+1, len(encryptedShares))
	assert.Equal(t, threshold+1, len(privateCoefficients))
	assert.Equal(t, threshold+1, len(commitments))

	t.Logf("encryptedShares:%x privateCoefficients:%x commitments:%x", encryptedShares, privateCoefficients, commitments)
}

func TestMath_GenerateSharesBad(t *testing.T) {
	_, _, _, err := GenerateShares(nil, objects.ParticipantList{})
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	privateKey := big.NewInt(1)
	_, _, _, err = GenerateShares(privateKey, objects.ParticipantList{})
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	participants := objects.ParticipantList{nil, nil, nil, nil}
	_, _, _, err = GenerateShares(privateKey, participants)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestMath_GenerateKeyShare(t *testing.T) {
	// Number participants in key generation
	n := 4

	// Make n participants
	participants := []*objects.Participant{{Index: 0}}
	for idx := 0; idx < n; idx++ {

		address, _, publicKey := generateTestAddress(t)

		participant := &objects.Participant{
			Address:   address,
			Index:     idx + 1,
			PublicKey: publicKey}

		participants = append(participants, participant)
	}

	// Overwrite the first
	private, public, _ := GenerateKeys()
	participants[0].PublicKey = public

	// Generate shares and sanity check them
	_, privateCoefficients, _, err := GenerateShares(private, participants)
	if err != nil {
		t.Fatal(err)
	}

	// Generate key share and sanity check it
	keyShare1, keyShare1Proof, keyShare2, err := GenerateKeyShare(privateCoefficients[0])
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

func TestMath_GenerateKeyShareBad(t *testing.T) {
	_, _, _, err := GenerateKeyShare(nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestMath_GenerateMasterPublicKey(t *testing.T) {
	// Number participants in key generation
	n := 4

	// Make n participants
	privateKeys := make(map[common.Address]*big.Int)
	participants := []*objects.Participant{{Index: 0}}
	for idx := 0; idx < n; idx++ {

		address, privateKey, publicKey := generateTestAddress(t)

		privateKeys[address] = privateKey
		participant := &objects.Participant{
			Address:   address,
			Index:     idx + 1,
			PublicKey: publicKey}

		participants = append(participants, participant)
	}

	// Overwrite the first
	private, public, _ := GenerateKeys()
	participants[0].PublicKey = public
	privateKeys[participants[0].Address] = private

	// Generate encrypted shares on behalf of participants
	keyShare1s := [][2]*big.Int{}
	keyShare2s := [][4]*big.Int{}
	for _, participant := range participants {
		privateKey := privateKeys[participant.Address]

		_, participantPrivateCoefficients, _, err := GenerateShares(privateKey, participants)
		assert.Nil(t, err)

		keyShare1, _, keyShare2, err := GenerateKeyShare(participantPrivateCoefficients[0])
		assert.Nil(t, err)

		keyShare1s = append(keyShare1s, keyShare1)
		keyShare2s = append(keyShare2s, keyShare2)
	}

	// Generate the master public key and sanity check it
	masterPublicKey, err := GenerateMasterPublicKey(keyShare1s, keyShare2s)
	assert.Nil(t, err)

	assert.NotNil(t, masterPublicKey[0], "missing element of master public key")
	assert.NotNil(t, masterPublicKey[1], "missing element of master public key")
	assert.NotNil(t, masterPublicKey[2], "missing element of master public key")
	assert.NotNil(t, masterPublicKey[3], "missing element of master public key")
}

func TestMath_GenerateMasterPublicKeyBad(t *testing.T) {
	keyShare1s := [][2]*big.Int{[2]*big.Int{nil, nil}}
	keyShare2s := [][4]*big.Int{}
	_, err := GenerateMasterPublicKey(keyShare1s, keyShare2s)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	keyShare1s = [][2]*big.Int{[2]*big.Int{nil, nil}}
	keyShare2s = [][4]*big.Int{[4]*big.Int{nil, nil, nil, nil}}
	_, err = GenerateMasterPublicKey(keyShare1s, keyShare2s)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	keyShare1s = [][2]*big.Int{[2]*big.Int{common.Big1, common.Big2}}
	_, err = GenerateMasterPublicKey(keyShare1s, keyShare2s)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func TestMath_GenerateGroupKeys(t *testing.T) {
	// Number participants in key generation
	n := 4

	// Make n participants
	privateKeys := make(map[common.Address]*big.Int)
	participants := []*objects.Participant{{Index: 1}}
	for idx := 0; idx < n; idx++ {
		address, privateKey, publicKey := generateTestAddress(t)
		privateKeys[address] = privateKey
		participant := &objects.Participant{
			Address:   address,
			Index:     idx + 1,
			PublicKey: publicKey,
		}
		participants = append(participants, participant)
	}

	// Overwrite the first
	private, public, err := GenerateKeys()
	if err != nil {
		t.Fatal(err)
	}
	participants[0].PublicKey = public
	privateKeys[participants[0].Address] = private

	// Generate shares
	_, privateCoefficients, _, err := GenerateShares(private, participants)
	if err != nil {
		t.Fatal(err)
	}

	encryptedShares := [][]*big.Int{}
	// Generate encrypted shares on behalf of participants
	for _, participant := range participants {
		privateKey := privateKeys[participant.Address]

		participantEncryptedShares, _, _, _ := GenerateShares(privateKey, participants)
		encryptedShares = append(encryptedShares, participantEncryptedShares)
	}

	// Generate the Group Keys and sanity check them
	index := 1
	groupPrivate, groupPublic, err := GenerateGroupKeys(private, privateCoefficients, encryptedShares, index, participants)
	assert.Nil(t, err, "error generating key share")
	assert.NotNil(t, groupPrivate, "group private key is missing")
	assert.NotNil(t, groupPublic[0], "group public key element is missing")
	assert.NotNil(t, groupPublic[1], "group public key element is missing")
	assert.NotNil(t, groupPublic[2], "group public key element is missing")
	assert.NotNil(t, groupPublic[3], "group public key element is missing")
	// assert.NotNil(t, groupSignature[0], "group signature element is missing")
	// assert.NotNil(t, groupSignature[1], "group signature element is missing")

	//t.Logf("groupPrivate:%x groupPublic:%x groupSignature:%x", groupPrivate, groupPublic, groupSignature)
}

func TestMath_GenerateGroupKeysBad1(t *testing.T) {
	// Initial Setup
	n := 4
	deterministicShares := true
	dkgStates, _ := InitializeNewDkgStateInfo(n, deterministicShares)
	participants := GenerateParticipantList(dkgStates)

	// Start raising errors
	// Raise error for nil transportPrivateKey
	index := 0
	_, _, err := GenerateGroupKeys(nil, nil, nil, index, participants)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	// Raise error for zero index
	transportPrivateKey := big.NewInt(123456789)
	_, _, err = GenerateGroupKeys(transportPrivateKey, nil, nil, index, participants)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	// Raise error for invalid private coefficients
	index = 1
	_, _, err = GenerateGroupKeys(transportPrivateKey, nil, nil, index, participants)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	// Raise an error for invalid encrypted shares
	threshold := ThresholdForUserCount(n)
	privCoefs := make([]*big.Int, threshold+1)
	_, _, err = GenerateGroupKeys(transportPrivateKey, privCoefs, nil, index, participants)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func TestMath_GenerateGroupKeysBad2(t *testing.T) {
	// Initial Setup
	n := 4
	deterministicShares := true
	dkgStates, _ := InitializeNewDkgStateInfo(n, deterministicShares)
	participants := GenerateParticipantList(dkgStates)

	transportPrivateKey := big.NewInt(123456789)
	index := 1
	threshold := ThresholdForUserCount(n)
	privCoefs := make([]*big.Int, threshold+1)
	encryptedShares := make([][]*big.Int, n)

	// Mess up public key
	participants[0].PublicKey = [2]*big.Int{}
	_, _, err := GenerateGroupKeys(transportPrivateKey, privCoefs, encryptedShares, index, participants)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	// Reset participant list
	participants = GenerateParticipantList(dkgStates)
	// Raise an error for condensing commitments
	_, _, err = GenerateGroupKeys(transportPrivateKey, privCoefs, encryptedShares, index, participants)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

// ---------------------------------------------------------------------------
func generateTestAddress(t *testing.T) (common.Address, *big.Int, [2]*big.Int) {

	// Generating a valid ethereum address
	key, _ := crypto.GenerateKey()
	chainId := big.NewInt(1337)
	transactor, err := bind.NewKeyedTransactorWithChainID(key, chainId)
	assert.Nil(t, err, "failed to create transactor")

	// Generate a public key
	privateKey, publicKey, err := GenerateKeys()
	assert.Nilf(t, err, "failed to generate keys")

	return transactor.From, privateKey, publicKey
}

// ---------------------------------------------------------------------------
func setupGroupSigners(t *testing.T, n int) ([4]*big.Int, [][4]*big.Int, []*objects.Participant, [][][2]*big.Int) {
	// Make n participants
	privateKeys := make(map[common.Address]*big.Int)
	participants := []*objects.Participant{}

	for idx := 0; idx < n; idx++ {

		address, privateKey, publicKey := generateTestAddress(t)

		privateKeys[address] = privateKey
		participant := &objects.Participant{
			Address:   address,
			Index:     idx + 1,
			PublicKey: publicKey}

		participants = append(participants, participant)
	}

	// Overwrite the first
	private, public, _ := GenerateKeys()
	participants[0].PublicKey = public
	privateKeys[participants[0].Address] = private

	// Generate encrypted shares on behalf of participants
	encryptedShares := [][]*big.Int{}
	keyShare1s := [][2]*big.Int{}
	keyShare2s := [][4]*big.Int{}
	privateCoefficients := [][]*big.Int{}
	commitmentArray := [][][2]*big.Int{}

	for _, participant := range participants {
		privateKey := privateKeys[participant.Address]

		participantEncryptedShares, participantPrivateCoefficients, commitments, err := GenerateShares(privateKey, participants)
		assert.Nil(t, err)

		keyShare1, _, keyShare2, err := GenerateKeyShare(participantPrivateCoefficients[0])
		assert.Nil(t, err)

		encryptedShares = append(encryptedShares, participantEncryptedShares)
		privateCoefficients = append(privateCoefficients, participantPrivateCoefficients)
		keyShare1s = append(keyShare1s, keyShare1)
		keyShare2s = append(keyShare2s, keyShare2)
		commitmentArray = append(commitmentArray, commitments)
	}

	// Generate the master public key and sanity check it
	masterPublicKey, err := GenerateMasterPublicKey(keyShare1s, keyShare2s)
	assert.Nil(t, err, "failed to generate master public key")

	publishedPublicKeys := [][4]*big.Int{}
	//publishedSignatures := [][2]*big.Int{}
	for idx, participant := range participants {
		privateKey := privateKeys[participant.Address]

		_, groupPublicKey, err := GenerateGroupKeys(privateKey, privateCoefficients[idx], encryptedShares, participant.Index, participants)
		assert.Nil(t, err, "failed to generate group keys")

		publishedPublicKeys = append(publishedPublicKeys, groupPublicKey)
		//publishedSignatures = append(publishedSignatures, groupSignature)
	}

	return masterPublicKey, publishedPublicKeys, participants, commitmentArray
}
