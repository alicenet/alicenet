package dtest

import (
	"crypto/ecdsa"
	"math/big"

	dkgMath "github.com/MadBase/MadNet/blockchain/dkg/math"
	"github.com/MadBase/MadNet/blockchain/etest"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/MadBase/MadNet/utils"
)

func InitializeNewDetDkgStateInfo(n int) ([]*objects.DkgState, []*ecdsa.PrivateKey) {
	return InitializeNewDkgStateInfo(n, true)
}

func InitializeNewNonDetDkgStateInfo(n int) ([]*objects.DkgState, []*ecdsa.PrivateKey) {
	return InitializeNewDkgStateInfo(n, false)
}

func InitializeNewDkgStateInfo(n int, deterministicShares bool) ([]*objects.DkgState, []*ecdsa.PrivateKey) {
	initialMessage := []byte("MadHive Rocks!")

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
		// Set initial message
		dkgState.InitialMessage = utils.CopySlice(initialMessage)

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
		dkgStates[k].Participants = participantList
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
			_, privCoefs, _, err := dkgMath.GenerateShares(dkgState.TransportPrivateKey, dkgState.Participants)
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
			dkgStates[ell].Commitments[dkgState.Account.Address] = publicCoefs
			dkgStates[ell].EncryptedShares[dkgState.Account.Address] = encryptedShares
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
			dkgStates[ell].KeyShareG1s[addr] = g1KeyShare
			dkgStates[ell].KeyShareG1CorrectnessProofs[addr] = g1Proof
			dkgStates[ell].KeyShareG2s[addr] = g2KeyShare
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
		for idx, participant := range dkgState.Participants {
			pes, present := dkgState.EncryptedShares[participant.Address]
			if present && idx >= 0 && idx < n {
				encryptedShares[idx] = pes
			} else {
				panic("Encrypted share state broken")
			}
		}

		groupPrivateKey, groupPublicKey, groupSignature, err := dkgMath.GenerateGroupKeys(dkgState.InitialMessage,
			dkgState.TransportPrivateKey, dkgState.PrivateCoefficients,
			encryptedShares, dkgState.Index, dkgState.Participants)
		if err != nil {
			panic("Could not generate group keys")
		}

		dkgState.GroupPrivateKey = groupPrivateKey
		dkgState.GroupPublicKey = groupPublicKey
		dkgState.GroupSignature = groupSignature

		// Loop through entire list and save in map
		for ell := 0; ell < n; ell++ {
			dkgStates[ell].GroupPublicKeys[dkgState.Account.Address] = groupPublicKey
			dkgStates[ell].GroupSignatures[dkgState.Account.Address] = groupSignature
		}
	}
}

func PopulateEncryptedSharesAndCommitments(dkgStates []*objects.DkgState) {
	n := len(dkgStates)
	for k := 0; k < n; k++ {
		dkgState := dkgStates[k]
		publicCoefs := dkgState.Commitments[dkgState.Account.Address]
		encryptedShares := dkgState.EncryptedShares[dkgState.Account.Address]
		// Loop through entire list and save in map
		for ell := 0; ell < n; ell++ {
			if ell == k {
				continue
			}
			dkgStates[ell].Commitments[dkgState.Account.Address] = publicCoefs
			dkgStates[ell].EncryptedShares[dkgState.Account.Address] = encryptedShares
		}
	}
}
