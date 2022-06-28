package state_test

import (
	"math/big"
	"testing"

	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
	dkgTestUtils "github.com/alicenet/alicenet/layer1/executor/tasks/dkg/tests/utils"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMath_VerifyDistributedSharesGood1(t *testing.T) {
	// Number participants in key generation
	n := 4
	// Test with deterministic private coefficients
	deterministicShares := true
	dkgStates, _ := dkgTestUtils.InitializeNewDkgStateInfo(t.TempDir(), n, deterministicShares)
	dkgTestUtils.GenerateEncryptedSharesAndCommitments(dkgStates)
	for idx := 0; idx < n; idx++ {
		dkgState := dkgStates[idx]
		for partIdx := 0; partIdx < n; partIdx++ {
			participant := dkgState.GetSortedParticipants()[partIdx]
			valid, present, err := state.VerifyDistributedShares(dkgState, participant)
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

func TestMath_VerifyDistributedSharesGood2(t *testing.T) {
	// Number participants in key generation
	n := 5
	// Test with random private coefficients
	deterministicShares := false
	dkgStates, _ := dkgTestUtils.InitializeNewDkgStateInfo(t.TempDir(), n, deterministicShares)
	dkgTestUtils.GenerateEncryptedSharesAndCommitments(dkgStates)
	for idx := 0; idx < n; idx++ {
		dkgState := dkgStates[idx]
		for partIdx := 0; partIdx < n; partIdx++ {
			participant := dkgState.GetSortedParticipants()[partIdx]
			valid, present, err := state.VerifyDistributedShares(dkgState, participant)
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

func TestMath_VerifyDistributedSharesGood3(t *testing.T) {
	// Number participants in key generation
	n := 7
	// Test with deterministic private coefficients
	deterministicShares := false
	dkgStates, _ := dkgTestUtils.InitializeNewDkgStateInfo(t.TempDir(), n, deterministicShares)
	dkgTestUtils.GenerateEncryptedSharesAndCommitments(dkgStates)

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
		valid, present, err := state.VerifyDistributedShares(dkgState, badParticipant)
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

func TestMath_VerifyDistributedSharesGood4(t *testing.T) {
	// Number participants in key generation
	n := 4
	// Test with deterministic private coefficients
	deterministicShares := true
	dkgStates, _ := dkgTestUtils.InitializeNewDkgStateInfo(t.TempDir(), n, deterministicShares)
	dkgTestUtils.GenerateEncryptedSharesAndCommitments(dkgStates)

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
		valid, present, err := state.VerifyDistributedShares(dkgState, badParticipant)
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

func TestMath_VerifyDistributedSharesBad1(t *testing.T) {
	// Test for raised error for nil arguments
	_, _, err := state.VerifyDistributedShares(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dkgState := &state.DkgState{}
	_, _, err = state.VerifyDistributedShares(dkgState, nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestMath_VerifyDistributedSharesBad2(t *testing.T) {
	// Test for error upon invalid number of participants
	dkgState := &state.DkgState{}
	dkgState.Index = 1
	participant := &state.Participant{}
	participant.Index = 2
	_, _, err := state.VerifyDistributedShares(dkgState, participant)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestMath_VerifyDistributedSharesBad3(t *testing.T) {
	// Test for error with invalid commitments and encrypted shares
	n := 4
	threshold := state.ThresholdForUserCount(n)

	// Setup keys
	_, _, accountsArray := tests.CreateAccounts(t.TempDir(), n)

	// Validator Setup
	dkgIdx := 0
	dkgState := state.NewDkgState(accountsArray[dkgIdx])
	dkgState.Index = dkgIdx + 1
	dkgState.NumberOfValidators = n
	dkgState.ValidatorThreshold = threshold

	// state.Participant Setup
	partIdx := 1
	participantState := state.NewDkgState(accountsArray[partIdx])
	participant := &state.Participant{}
	participant.Index = partIdx + 1
	participant.Address = participantState.Account.Address
	dkgState.Participants[participant.Address] = participant

	//Test after initial setup; nothing present
	valid, present, err := state.VerifyDistributedShares(dkgState, participant)
	assert.NotNil(t, err)
	assert.False(t, present)
	assert.False(t, valid)

	// no commitment present but (invalid) shares
	encryptedSharesBad := make([]*big.Int, 0)
	dkgState.Participants[participant.Address].EncryptedShares = encryptedSharesBad
	_, _, err = state.VerifyDistributedShares(dkgState, participant)
	assert.NotNil(t, err)

	// Remove shares from map
	dkgState.Participants[participant.Address].EncryptedShares = nil

	// Make empty commitment list of big ints; raise error from incorrect length
	commitmentsBad0 := make([][2]*big.Int, 0)
	dkgState.Participants[participant.Address].Commitments = commitmentsBad0
	_, _, err = state.VerifyDistributedShares(dkgState, participant)
	assert.NotNil(t, err)

	dkgState.Participants[participant.Address].Commitments = nil

	// Raise error from invalid commitment length
	commitmentsBad1 := make([][2]*big.Int, threshold)
	dkgState.Participants[participant.Address].Commitments = commitmentsBad1
	dkgState.Participants[participant.Address].EncryptedShares = encryptedSharesBad
	_, _, err = state.VerifyDistributedShares(dkgState, participant)
	assert.NotNil(t, err)

	dkgState.Participants[participant.Address].Commitments = nil
	dkgState.Participants[participant.Address].EncryptedShares = nil

	// Raise error from invalid encryptedSharesList length
	commitmentsBad2 := make([][2]*big.Int, threshold+1)
	dkgState.Participants[participant.Address].Commitments = commitmentsBad2
	dkgState.Participants[participant.Address].EncryptedShares = encryptedSharesBad
	_, _, err = state.VerifyDistributedShares(dkgState, participant)
	assert.NotNil(t, err)

	dkgState.Participants[participant.Address].Commitments = nil
	dkgState.Participants[participant.Address].EncryptedShares = nil

	// Make empty encrypted share list; raise error invalid unmarshalling
	encryptedSharesEmpty := make([]*big.Int, n-1)
	dkgState.Participants[participant.Address].Commitments = commitmentsBad2
	dkgState.Participants[participant.Address].EncryptedShares = encryptedSharesEmpty
	_, _, err = state.VerifyDistributedShares(dkgState, participant)
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
	_, _, err = state.VerifyDistributedShares(dkgState, participant)
	assert.NotNil(t, err)
}

func TestMath_CategorizeGroupSigners(t *testing.T) {
	n := 10
	_, publishedPublicKeys, participants, commitmentArray := setupGroupSigners(t, n)

	honest, dishonest, missing, err := state.CategorizeGroupSigners(publishedPublicKeys, participants, commitmentArray)
	assert.Nil(t, err, "failed to categorize group signers")
	assert.Equal(t, len(participants), len(honest), "all participants should be honest")
	assert.Equal(t, 0, len(dishonest), "no participants should be dishonest")
	assert.Equal(t, 0, len(missing), "no participants should be missing")
}

func TestMath_CategorizeGroupSigners1Negative(t *testing.T) {
	n := 30

	logger := logging.GetLogger("dkg")
	logger.SetLevel(logrus.DebugLevel)

	_, publishedPublicKeys, participants, commitmentArray := setupGroupSigners(t, n)

	participants[0].Index = n + 100

	honest, dishonest, missing, err := state.CategorizeGroupSigners(publishedPublicKeys, participants, commitmentArray)
	assert.Nil(t, err, "failed to categorize group signers")
	assert.Equal(t, len(participants)-1, len(honest), "all but 1 participant are honest")
	assert.Equal(t, 1, len(dishonest), "1 participant is dishonest")
	assert.Equal(t, 0, len(missing), "0 participants are missing")
}

func TestMath_CategorizeGroupSigners2Negative(t *testing.T) {
	n := 10
	threshold := state.ThresholdForUserCount(n)

	_, publishedPublicKeys, participants, commitmentArray := setupGroupSigners(t, n)

	participants[n-1].Index = n + 100
	participants[n-2].Index = n + 101

	honest, dishonest, missing, err := state.CategorizeGroupSigners(publishedPublicKeys, participants, commitmentArray)
	assert.Nil(t, err, "failed to categorize group signers")

	t.Logf("n:%v threshold:%v", n, threshold)

	t.Logf("%v participant are honest", len(participants)-2)
	assert.Equal(t, len(participants)-2, len(honest))

	t.Logf("%v participant are dishonest", len(dishonest))
	assert.Equal(t, 2, len(dishonest))

	t.Logf("%v participants are missing", len(missing))
	assert.Equal(t, 0, len(missing))
}

func TestMath_CategorizeGroupSignersBad(t *testing.T) {
	n := 4
	_, publishedPublicKeys, participants, commitmentArray := setupGroupSigners(t, n)
	threshold := state.ThresholdForUserCount(n)

	// Raise error for bad number of commitments
	commitmentBad := commitmentArray[:n-1]
	_, _, _, err := state.CategorizeGroupSigners(publishedPublicKeys, participants, commitmentBad)
	if err == nil {
		t.Fatal("Should have raised error (0)")
	}

	// Raise error for bad number of public keys
	publishedPublicKeysBad := publishedPublicKeys[:n-1]
	_, _, _, err = state.CategorizeGroupSigners(publishedPublicKeysBad, participants, commitmentArray)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	// Raise error for incorrect commitment lengths
	commitmentBad2 := [][][2]*big.Int{}
	for k := 0; k < n; k++ {
		commitmentBad2 = append(commitmentBad2, [][2]*big.Int{})
	}
	_, _, _, err = state.CategorizeGroupSigners(publishedPublicKeys, participants, commitmentBad2)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	// Raise error for nil public keys;
	// raises error when converting to G2.
	publishedPublicKeysBad2 := [][4]*big.Int{}
	for k := 0; k < n; k++ {
		publishedPublicKeysBad2 = append(publishedPublicKeysBad2, [4]*big.Int{big.NewInt(1), big.NewInt(1), big.NewInt(1), big.NewInt(1)})
	}
	_, _, _, err = state.CategorizeGroupSigners(publishedPublicKeysBad2, participants, commitmentArray)
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
	_, _, _, err = state.CategorizeGroupSigners(publishedPublicKeys, participants, commitmentBad3)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
}

func TestMath_CategorizeGroupSignersBad2(t *testing.T) {
	n := 4
	_, publishedPublicKeys, participants, commitmentArray := setupGroupSigners(t, n)
	publishedPublicKeysBad := [][4]*big.Int{}
	for k := 0; k < len(publishedPublicKeys); k++ {
		zeroPubKey := [4]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
		publishedPublicKeysBad = append(publishedPublicKeysBad, zeroPubKey)
	}
	honest, dishonest, missing, err := state.CategorizeGroupSigners(publishedPublicKeysBad, participants, commitmentArray)
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

func setupGroupSigners(t *testing.T, n int) ([4]*big.Int, [][4]*big.Int, []*state.Participant, [][][2]*big.Int) {
	// Make n participants
	privateKeys := make(map[common.Address]*big.Int)
	participants := []*state.Participant{}

	for idx := 0; idx < n; idx++ {

		address, privateKey, publicKey := dkgTestUtils.GenerateTestAddress(t)

		privateKeys[address] = privateKey
		participant := &state.Participant{
			Address:   address,
			Index:     idx + 1,
			PublicKey: publicKey}

		participants = append(participants, participant)
	}

	// Overwrite the first
	private, public, _ := state.GenerateKeys()
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

		participantEncryptedShares, participantPrivateCoefficients, commitments, err := state.GenerateShares(privateKey, participants)
		assert.Nil(t, err)

		keyShare1, _, keyShare2, err := state.GenerateKeyShare(participantPrivateCoefficients[0])
		assert.Nil(t, err)

		encryptedShares = append(encryptedShares, participantEncryptedShares)
		privateCoefficients = append(privateCoefficients, participantPrivateCoefficients)
		keyShare1s = append(keyShare1s, keyShare1)
		keyShare2s = append(keyShare2s, keyShare2)
		commitmentArray = append(commitmentArray, commitments)
	}

	// Generate the master public key and sanity check it
	masterPublicKey, err := state.GenerateMasterPublicKey(keyShare1s, keyShare2s)
	assert.Nil(t, err, "failed to generate master public key")

	publishedPublicKeys := [][4]*big.Int{}
	//publishedSignatures := [][2]*big.Int{}
	for idx, participant := range participants {
		privateKey := privateKeys[participant.Address]

		_, groupPublicKey, err := state.GenerateGroupKeys(privateKey, privateCoefficients[idx], encryptedShares, participant.Index, participants)
		assert.Nil(t, err, "failed to generate group keys")

		publishedPublicKeys = append(publishedPublicKeys, groupPublicKey)
		//publishedSignatures = append(publishedSignatures, groupSignature)
	}

	return masterPublicKey, publishedPublicKeys, participants, commitmentArray
}
