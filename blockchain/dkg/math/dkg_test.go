package math_test

import (
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/objects"
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
	threshold := math.ThresholdForUserCount(4)
	assert.Equal(t, 2, threshold)
	threshold = math.ThresholdForUserCount(5)
	assert.Equal(t, 3, threshold)
	threshold = math.ThresholdForUserCount(6)
	assert.Equal(t, 4, threshold)
	threshold = math.ThresholdForUserCount(7)
	assert.Equal(t, 4, threshold)
	threshold = math.ThresholdForUserCount(8)
	assert.Equal(t, 5, threshold)
	threshold = math.ThresholdForUserCount(9)
	assert.Equal(t, 6, threshold)
}

func TestInverseArrayForUserCount(t *testing.T) {
	n := 3
	_, err := math.InverseArrayForUserCount(n)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	n = 10
	invArray, err := math.InverseArrayForUserCount(n)
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

func TestGenerateKeys(t *testing.T) {
	private, public, err := math.GenerateKeys()
	assert.Nil(t, err, "error generating keys")

	assert.NotNil(t, private, "private key is nil")
	assert.NotNil(t, public, "public key is nil")

	assert.NotNil(t, public[0], "public key missing element")
	assert.NotNil(t, public[1], "public key missing element")
}

func TestGenerateShares(t *testing.T) {
	// Number participants in key generation
	n := 4
	threshold := math.ThresholdForUserCount(n)
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
	private, public, _ := math.GenerateKeys()
	participants[0].PublicKey = public

	// Now actually generate shares and sanity check them
	encryptedShares, privateCoefficients, commitments, err := math.GenerateShares(private, participants)
	assert.Nil(t, err, "error generating shares")
	assert.Equal(t, threshold+1, len(encryptedShares))
	assert.Equal(t, threshold+1, len(privateCoefficients))
	assert.Equal(t, threshold+1, len(commitments))

	t.Logf("encryptedShares:%x privateCoefficients:%x commitments:%x", encryptedShares, privateCoefficients, commitments)
}

func TestGenerateSharesBad(t *testing.T) {
	_, _, _, err := math.GenerateShares(nil, objects.ParticipantList{})
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	privateKey := big.NewInt(1)
	_, _, _, err = math.GenerateShares(privateKey, objects.ParticipantList{})
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	participants := objects.ParticipantList{nil, nil, nil, nil}
	_, _, _, err = math.GenerateShares(privateKey, participants)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestVerifyDistributedSharesGood1(t *testing.T) {
	// Number participants in key generation
	n := 4
	// Test with deterministic private coefficients
	deterministicShares := true
	dkgStates, _ := InitializeNewDkgStateInfo(n, deterministicShares)
	GenerateEncryptedSharesAndCommitments(dkgStates)
	for idx := 0; idx < n; idx++ {
		dkgState := dkgStates[idx]
		for partIdx := 0; partIdx < n; partIdx++ {
			participant := dkgState.GetSortedParticipants()[partIdx]
			valid, present, err := math.VerifyDistributedShares(dkgState, participant)
			if err != nil {
				t.Fatalf("Error raised in VerifyDistributedShares: s_i->j; i: %v; j: %v\nerr:= %v\n", participant.Index, dkgState.Index, err)
			}
			if !present {
				t.Fatalf("Invalid share in VerifyDistributedShares: s_i->j; i: %v; j: %v\nnot present\n", participant.Index, dkgState.Index)
			}
			if !valid {
				t.Fatalf("Invalid share in VerifyDistributedShares: s_i->j; i: %v; j: %v\n", participant.Index, dkgState.Index)
			}
		}
	}
}

func TestVerifyDistributedSharesGood2(t *testing.T) {
	// Number participants in key generation
	n := 5
	// Test with random private coefficients
	deterministicShares := false
	dkgStates, _ := InitializeNewDkgStateInfo(n, deterministicShares)
	GenerateEncryptedSharesAndCommitments(dkgStates)
	for idx := 0; idx < n; idx++ {
		dkgState := dkgStates[idx]
		for partIdx := 0; partIdx < n; partIdx++ {
			participant := dkgState.GetSortedParticipants()[partIdx]
			valid, present, err := math.VerifyDistributedShares(dkgState, participant)
			if err != nil {
				t.Fatalf("Error raised in VerifyDistributedShares: s_i->j; i: %v; j: %v\nerr: %v", participant.Index, dkgState.Index, err)
			}
			if !present {
				t.Fatalf("Invalid share in VerifyDistributedShares: s_i->j; i: %v; j: %v\nnot present\n", participant.Index, dkgState.Index)
			}
			if !valid {
				t.Fatalf("Invalid share in VerifyDistributedShares: s_i->j; i: %v; j: %v\n", participant.Index, dkgState.Index)
			}
		}
	}
}

func TestVerifyDistributedSharesGood3(t *testing.T) {
	// Number participants in key generation
	n := 7
	// Test with deterministic private coefficients
	deterministicShares := false
	dkgStates, _ := InitializeNewDkgStateInfo(n, deterministicShares)
	GenerateEncryptedSharesAndCommitments(dkgStates)

	// We now mess up the scheme, ensuring that we have an invalid share.
	badIdx := 0
	badParticipant := dkgStates[0].GetSortedParticipants()[badIdx]
	badEncryptedShares := make([]*big.Int, n-1)
	for k := 0; k < len(badEncryptedShares); k++ {
		badEncryptedShares[k] = new(big.Int)
	}
	for idx := 0; idx < n; idx++ {
		dkgStates[idx].Participants[badParticipant.Address].EncryptedShares = badEncryptedShares
	}

	// Loop through all participants and ensure that they all evaluate
	// to invalid shares (outside of self).
	for idx := 0; idx < n; idx++ {
		if idx == badIdx {
			continue
		}
		dkgState := dkgStates[idx]
		valid, present, err := math.VerifyDistributedShares(dkgState, badParticipant)
		if err != nil {
			t.Fatalf("Error raised in VerifyDistributedShares: s_i->j; i: %v; j: %v\nerr: %v\n", badParticipant.Index, dkgState.Index, err)
		}
		if !present {
			t.Fatalf("Invalid share in VerifyDistributedShares: s_i->j; i: %v; j: %v\nnot present\n", badParticipant.Index, dkgState.Index)
		}
		if valid {
			t.Fatalf("Valid share in VerifyDistributedShares: s_i->j; i: %v; j: %v\n", badParticipant.Index, dkgState.Index)
		}
	}
}

func TestVerifyDistributedSharesGood4(t *testing.T) {
	// Number participants in key generation
	n := 4
	// Test with deterministic private coefficients
	deterministicShares := true
	dkgStates, _ := InitializeNewDkgStateInfo(n, deterministicShares)
	GenerateEncryptedSharesAndCommitments(dkgStates)

	// We now mess up the scheme, ensuring that we have an invalid share.
	badIdx := 0
	badParticipant := dkgStates[0].GetSortedParticipants()[badIdx]
	badEncryptedShares := make([]*big.Int, n-1)
	for k := 0; k < len(badEncryptedShares); k++ {
		badEncryptedShares[k] = new(big.Int)
	}
	for idx := 0; idx < n; idx++ {
		dkgStates[idx].Participants[badParticipant.Address].EncryptedShares = badEncryptedShares
		//dkgStates[idx].Participants[badParticipant.Address].Commitments = nil
	}

	// Loop through all participants and ensure that they all evaluate
	// to invalid shares (outside of self).
	for idx := 0; idx < n; idx++ {
		if idx == badIdx {
			continue
		}
		dkgState := dkgStates[idx]
		valid, present, err := math.VerifyDistributedShares(dkgState, badParticipant)
		if err != nil {
			t.Fatalf("Error raised in VerifyDistributedShares: s_i->j; i: %v; j: %v\nerr: %v\n", badParticipant.Index, dkgState.Index, err)
		}
		if !present {
			t.Fatalf("Invalid share in VerifyDistributedShares: s_i->j; i: %v; j: %v\nnot present\n", badParticipant.Index, dkgState.Index)
		}
		if valid {
			t.Fatalf("Valid share in VerifyDistributedShares: s_i->j; i: %v; j: %v\n", badParticipant.Index, dkgState.Index)
		}
	}
}

func TestVerifyDistributedSharesBad1(t *testing.T) {
	// Test for raised error for nil arguments
	_, _, err := math.VerifyDistributedShares(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dkgState := &objects.DkgState{}
	_, _, err = math.VerifyDistributedShares(dkgState, nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestVerifyDistributedSharesBad2(t *testing.T) {
	// Test for error upon invalid number of participants
	dkgState := &objects.DkgState{}
	dkgState.Index = 1
	participant := &objects.Participant{}
	participant.Index = 2
	_, _, err := math.VerifyDistributedShares(dkgState, participant)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestVerifyDistributedSharesBad3(t *testing.T) {
	// Test for error with invalid commitments and encrypted shares
	n := 4
	threshold := math.ThresholdForUserCount(n)

	// Setup keys
	ecdsaPrivKeys := SetupPrivateKeys(n)
	accountsArray := SetupAccounts(ecdsaPrivKeys)

	// Validator Setup
	dkgIdx := 0
	dkgState := objects.NewDkgState(accountsArray[dkgIdx])
	dkgState.Index = dkgIdx + 1
	dkgState.NumberOfValidators = n
	dkgState.ValidatorThreshold = threshold

	// Participant Setup
	partIdx := 1
	participantState := objects.NewDkgState(accountsArray[partIdx])
	participant := &objects.Participant{}
	participant.Index = partIdx + 1
	participant.Address = participantState.Account.Address
	dkgState.Participants[participant.Address] = participant

	//Test after initial setup; nothing present
	valid, present, err := math.VerifyDistributedShares(dkgState, participant)
	assert.NotNil(t, err)
	assert.False(t, present)
	assert.False(t, valid)

	// no commitment present but (invalid) shares
	encryptedSharesBad := make([]*big.Int, 0)
	dkgState.Participants[participant.Address].EncryptedShares = encryptedSharesBad
	_, _, err = math.VerifyDistributedShares(dkgState, participant)
	assert.NotNil(t, err)

	// Remove shares from map
	dkgState.Participants[participant.Address].EncryptedShares = nil

	// Make empty commitment list of big ints; raise error from incorrect length
	commitmentsBad0 := make([][2]*big.Int, 0)
	dkgState.Participants[participant.Address].Commitments = commitmentsBad0
	_, _, err = math.VerifyDistributedShares(dkgState, participant)
	assert.NotNil(t, err)

	dkgState.Participants[participant.Address].Commitments = nil

	// Raise error from invalid commitment length
	commitmentsBad1 := make([][2]*big.Int, threshold)
	dkgState.Participants[participant.Address].Commitments = commitmentsBad1
	dkgState.Participants[participant.Address].EncryptedShares = encryptedSharesBad
	_, _, err = math.VerifyDistributedShares(dkgState, participant)
	assert.NotNil(t, err)

	dkgState.Participants[participant.Address].Commitments = nil
	dkgState.Participants[participant.Address].EncryptedShares = nil

	// Raise error from invalid encryptedSharesList length
	commitmentsBad2 := make([][2]*big.Int, threshold+1)
	dkgState.Participants[participant.Address].Commitments = commitmentsBad2
	dkgState.Participants[participant.Address].EncryptedShares = encryptedSharesBad
	_, _, err = math.VerifyDistributedShares(dkgState, participant)
	assert.NotNil(t, err)

	dkgState.Participants[participant.Address].Commitments = nil
	dkgState.Participants[participant.Address].EncryptedShares = nil

	// Make empty encrypted share list; raise error invalid unmarshalling
	encryptedSharesEmpty := make([]*big.Int, n-1)
	dkgState.Participants[participant.Address].Commitments = commitmentsBad2
	dkgState.Participants[participant.Address].EncryptedShares = encryptedSharesEmpty
	_, _, err = math.VerifyDistributedShares(dkgState, participant)
	assert.NotNil(t, err)

	dkgState.Participants[participant.Address].Commitments = nil
	dkgState.Participants[participant.Address].EncryptedShares = nil

	//Make commitment list of correct length and valid;
	//raise an error for invalid public key
	commitments := make([][2]*big.Int, threshold+1)
	for k := 0; k < len(commitments); k++ {
		commitments[k] = [2]*big.Int{common.Big1, common.Big2}
	}
	dkgState.Participants[participant.Address].Commitments = commitments
	dkgState.Participants[participant.Address].EncryptedShares = encryptedSharesEmpty
	_, _, err = math.VerifyDistributedShares(dkgState, participant)
	assert.NotNil(t, err)
}

func TestGenerateKeyShare(t *testing.T) {
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
	private, public, _ := math.GenerateKeys()
	participants[0].PublicKey = public

	// Generate shares and sanity check them
	_, privateCoefficients, _, err := math.GenerateShares(private, participants)
	if err != nil {
		t.Fatal(err)
	}

	// Generate key share and sanity check it
	keyShare1, keyShare1Proof, keyShare2, err := math.GenerateKeyShare(privateCoefficients[0])
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

func TestGenerateKeyShareBad(t *testing.T) {
	_, _, _, err := math.GenerateKeyShare(nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestGenerateMasterPublicKey(t *testing.T) {
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
	private, public, _ := math.GenerateKeys()
	participants[0].PublicKey = public
	privateKeys[participants[0].Address] = private

	// Generate encrypted shares on behalf of participants
	encryptedShares := [][]*big.Int{}
	keyShare1s := [][2]*big.Int{}
	keyShare2s := [][4]*big.Int{}
	for _, participant := range participants {
		privateKey := privateKeys[participant.Address]

		participantEncryptedShares, participantPrivateCoefficients, _, err := math.GenerateShares(privateKey, participants)
		assert.Nil(t, err)

		keyShare1, _, keyShare2, err := math.GenerateKeyShare(participantPrivateCoefficients[0])
		assert.Nil(t, err)

		encryptedShares = append(encryptedShares, participantEncryptedShares)
		keyShare1s = append(keyShare1s, keyShare1)
		keyShare2s = append(keyShare2s, keyShare2)
	}

	// Generate the master public key and sanity check it
	masterPublicKey, err := math.GenerateMasterPublicKey(keyShare1s, keyShare2s)
	assert.Nil(t, err)

	assert.NotNil(t, masterPublicKey[0], "missing element of master public key")
	assert.NotNil(t, masterPublicKey[1], "missing element of master public key")
	assert.NotNil(t, masterPublicKey[2], "missing element of master public key")
	assert.NotNil(t, masterPublicKey[3], "missing element of master public key")
}

func TestGenerateMasterPublicKeyBad(t *testing.T) {
	keyShare1s := [][2]*big.Int{[2]*big.Int{nil, nil}}
	keyShare2s := [][4]*big.Int{}
	_, err := math.GenerateMasterPublicKey(keyShare1s, keyShare2s)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	keyShare1s = [][2]*big.Int{[2]*big.Int{nil, nil}}
	keyShare2s = [][4]*big.Int{[4]*big.Int{nil, nil, nil, nil}}
	_, err = math.GenerateMasterPublicKey(keyShare1s, keyShare2s)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	keyShare1s = [][2]*big.Int{[2]*big.Int{common.Big1, common.Big2}}
	_, err = math.GenerateMasterPublicKey(keyShare1s, keyShare2s)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func TestGenerateGroupKeys(t *testing.T) {
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
	private, public, err := math.GenerateKeys()
	if err != nil {
		t.Fatal(err)
	}
	participants[0].PublicKey = public
	privateKeys[participants[0].Address] = private

	// Generate shares
	_, privateCoefficients, _, err := math.GenerateShares(private, participants)
	if err != nil {
		t.Fatal(err)
	}

	encryptedShares := [][]*big.Int{}
	// Generate encrypted shares on behalf of participants
	for _, participant := range participants {
		privateKey := privateKeys[participant.Address]

		participantEncryptedShares, _, _, _ := math.GenerateShares(privateKey, participants)
		encryptedShares = append(encryptedShares, participantEncryptedShares)
	}

	// Generate the Group Keys and sanity check them
	index := 1
	groupPrivate, groupPublic, err := math.GenerateGroupKeys(private, privateCoefficients, encryptedShares, index, participants)
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

func TestGenerateGroupKeysBad1(t *testing.T) {
	// Initial Setup
	n := 4
	deterministicShares := true
	dkgStates, _ := InitializeNewDkgStateInfo(n, deterministicShares)
	participants := GenerateParticipantList(dkgStates)

	// Start raising errors
	// Raise error for nil transportPrivateKey
	index := 0
	_, _, err := math.GenerateGroupKeys(nil, nil, nil, index, participants)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	// Raise error for zero index
	transportPrivateKey := big.NewInt(123456789)
	_, _, err = math.GenerateGroupKeys(transportPrivateKey, nil, nil, index, participants)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	// Raise error for invalid private coefficients
	index = 1
	_, _, err = math.GenerateGroupKeys(transportPrivateKey, nil, nil, index, participants)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	// Raise an error for invalid encrypted shares
	threshold := math.ThresholdForUserCount(n)
	privCoefs := make([]*big.Int, threshold+1)
	_, _, err = math.GenerateGroupKeys(transportPrivateKey, privCoefs, nil, index, participants)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func TestGenerateGroupKeysBad2(t *testing.T) {
	// Initial Setup
	n := 4
	deterministicShares := true
	dkgStates, _ := InitializeNewDkgStateInfo(n, deterministicShares)
	participants := GenerateParticipantList(dkgStates)

	transportPrivateKey := big.NewInt(123456789)
	index := 1
	threshold := math.ThresholdForUserCount(n)
	privCoefs := make([]*big.Int, threshold+1)
	encryptedShares := make([][]*big.Int, n)

	// Mess up public key
	participants[0].PublicKey = [2]*big.Int{}
	_, _, err := math.GenerateGroupKeys(transportPrivateKey, privCoefs, encryptedShares, index, participants)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	// Reset participant list
	participants = GenerateParticipantList(dkgStates)
	// Raise an error for condensing commitments
	_, _, err = math.GenerateGroupKeys(transportPrivateKey, privCoefs, encryptedShares, index, participants)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestCategorizeGroupSigners(t *testing.T) {
	n := 10
	_, publishedPublicKeys, participants, commitmentArray := setupGroupSigners(t, n)

	honest, dishonest, missing, err := math.CategorizeGroupSigners(publishedPublicKeys, participants, commitmentArray)
	assert.Nil(t, err, "failed to categorize group signers")
	assert.Equal(t, len(participants), len(honest), "all participants should be honest")
	assert.Equal(t, 0, len(dishonest), "no participants should be dishonest")
	assert.Equal(t, 0, len(missing), "no participants should be missing")
}

func TestCategorizeGroupSigners1Negative(t *testing.T) {
	n := 30

	logger := logging.GetLogger("dkg")
	logger.SetLevel(logrus.DebugLevel)

	_, publishedPublicKeys, participants, commitmentArray := setupGroupSigners(t, n)

	participants[0].Index = n + 100

	honest, dishonest, missing, err := math.CategorizeGroupSigners(publishedPublicKeys, participants, commitmentArray)
	assert.Nil(t, err, "failed to categorize group signers")
	assert.Equal(t, len(participants)-1, len(honest), "all but 1 participant are honest")
	assert.Equal(t, 1, len(dishonest), "1 participant is dishonest")
	assert.Equal(t, 0, len(missing), "0 participants are missing")
}

func TestCategorizeGroupSigners2Negative(t *testing.T) {
	n := 10
	threshold := math.ThresholdForUserCount(n)

	_, publishedPublicKeys, participants, commitmentArray := setupGroupSigners(t, n)

	participants[n-1].Index = n + 100
	participants[n-2].Index = n + 101

	honest, dishonest, missing, err := math.CategorizeGroupSigners(publishedPublicKeys, participants, commitmentArray)
	assert.Nil(t, err, "failed to categorize group signers")

	t.Logf("n:%v threshold:%v", n, threshold)

	t.Logf("%v participant are honest", len(participants)-2)
	assert.Equal(t, len(participants)-2, len(honest))

	t.Logf("%v participant are dishonest", len(dishonest))
	assert.Equal(t, 2, len(dishonest))

	t.Logf("%v participants are missing", len(missing))
	assert.Equal(t, 0, len(missing))
}

func TestCategorizeGroupSignersBad(t *testing.T) {
	n := 4
	_, publishedPublicKeys, participants, commitmentArray := setupGroupSigners(t, n)
	threshold := math.ThresholdForUserCount(n)

	// Raise error for bad number of commitments
	commitmentBad := commitmentArray[:n-1]
	_, _, _, err := math.CategorizeGroupSigners(publishedPublicKeys, participants, commitmentBad)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	// Raise error for bad number of public keys
	publishedPublicKeysBad := publishedPublicKeys[:n-1]
	_, _, _, err = math.CategorizeGroupSigners(publishedPublicKeysBad, participants, commitmentArray)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	// Raise error for incorrect commitment lengths
	commitmentBad2 := [][][2]*big.Int{}
	for k := 0; k < n; k++ {
		commitmentBad2 = append(commitmentBad2, [][2]*big.Int{})
	}
	_, _, _, err = math.CategorizeGroupSigners(publishedPublicKeys, participants, commitmentBad2)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	// Raise error for nil public keys;
	// raises error when converting to G2.
	publishedPublicKeysBad2 := [][4]*big.Int{}
	for k := 0; k < n; k++ {
		publishedPublicKeysBad2 = append(publishedPublicKeysBad2, [4]*big.Int{big.NewInt(1), big.NewInt(1), big.NewInt(1), big.NewInt(1)})
	}
	_, _, _, err = math.CategorizeGroupSigners(publishedPublicKeysBad2, participants, commitmentArray)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	// Raise error for nil commitments;
	// raises error when converting to G1.
	commitmentBad3 := [][][2]*big.Int{}
	for k := 0; k < n; k++ {
		com := make([][2]*big.Int, threshold+1)
		commitmentBad3 = append(commitmentBad3, com)
	}
	_, _, _, err = math.CategorizeGroupSigners(publishedPublicKeys, participants, commitmentBad3)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
}

func TestCategorizeGroupSignersBad2(t *testing.T) {
	n := 4
	_, publishedPublicKeys, participants, commitmentArray := setupGroupSigners(t, n)
	publishedPublicKeysBad := [][4]*big.Int{}
	for k := 0; k < len(publishedPublicKeys); k++ {
		zeroPubKey := [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
		publishedPublicKeysBad = append(publishedPublicKeysBad, zeroPubKey)
	}
	honest, dishonest, missing, err := math.CategorizeGroupSigners(publishedPublicKeysBad, participants, commitmentArray)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v participants are honest", len(honest))
	assert.Equal(t, 0, len(honest))
	t.Logf("%v participants are dishonest", len(dishonest))
	assert.Equal(t, 0, len(dishonest))
	t.Logf("%v participants are missing", len(missing))
	assert.Equal(t, n, len(missing))
}

// ---------------------------------------------------------------------------
func generateTestAddress(t *testing.T) (common.Address, *big.Int, [2]*big.Int) {

	// Generating a valid ethereum address
	key, _ := crypto.GenerateKey()
	transactor := bind.NewKeyedTransactor(key)

	// Generate a public key
	privateKey, publicKey, err := math.GenerateKeys()
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
	private, public, _ := math.GenerateKeys()
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

		participantEncryptedShares, participantPrivateCoefficients, commitments, err := math.GenerateShares(privateKey, participants)
		assert.Nil(t, err)

		keyShare1, _, keyShare2, err := math.GenerateKeyShare(participantPrivateCoefficients[0])
		assert.Nil(t, err)

		encryptedShares = append(encryptedShares, participantEncryptedShares)
		privateCoefficients = append(privateCoefficients, participantPrivateCoefficients)
		keyShare1s = append(keyShare1s, keyShare1)
		keyShare2s = append(keyShare2s, keyShare2)
		commitmentArray = append(commitmentArray, commitments)
	}

	// Generate the master public key and sanity check it
	masterPublicKey, err := math.GenerateMasterPublicKey(keyShare1s, keyShare2s)
	assert.Nil(t, err, "failed to generate master public key")

	publishedPublicKeys := [][4]*big.Int{}
	//publishedSignatures := [][2]*big.Int{}
	for idx, participant := range participants {
		privateKey := privateKeys[participant.Address]

		_, groupPublicKey, err := math.GenerateGroupKeys(privateKey, privateCoefficients[idx], encryptedShares, participant.Index, participants)
		assert.Nil(t, err, "failed to generate group keys")

		publishedPublicKeys = append(publishedPublicKeys, groupPublicKey)
		//publishedSignatures = append(publishedSignatures, groupSignature)
	}

	return masterPublicKey, publishedPublicKeys, participants, commitmentArray
}
