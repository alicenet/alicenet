package state

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/ethereum/go-ethereum/common"
)

// Distributed Key Generation related errors
var (
	ErrInsufficientGoodSigners = errors.New("insufficient non-malicious signers to identify malicious signers")
	ErrTooFew                  = errors.New("building array of size n with less than n required + optional")
	ErrTooMany                 = errors.New("building array of size n with more than n required")
)

// Useful pseudo-constants
var (
	empty2Big     [2]*big.Int
	empty4Big     [4]*big.Int
	h1BaseMessage []byte = []byte("MadHive Rocks!")
)

// ThresholdForUserCount returns the threshold user count;
// see crypto for full documentation and discussion.
func ThresholdForUserCount(n int) int {
	return crypto.CalcThreshold(n)
}

// InverseArrayForUserCount pre-calculates an inverse array for use by ethereum contracts
func InverseArrayForUserCount(n int) ([]*big.Int, error) {
	if n < 4 {
		return nil, errors.New("invalid user count")
	}
	bigNeg2 := big.NewInt(-2)
	orderMinus2 := new(big.Int).Add(cloudflare.Order, bigNeg2)

	// Get inverse array; this array is required to help keep gas costs down
	// in the smart contract. Modular multiplication is much cheaper than
	// modular inversion (exponentiation).
	invArrayBig := make([]*big.Int, n-1)
	for idx := 0; idx < n-1; idx++ {
		m := big.NewInt(int64(idx + 1))
		mInv := new(big.Int).Exp(m, orderMinus2, cloudflare.Order)
		// Confirm
		res := new(big.Int).Mul(m, mInv)
		res.Mod(res, cloudflare.Order)
		if res.Cmp(common.Big1) != 0 {
			return nil, errors.New("error when computing inverseArray")
		}
		invArrayBig[idx] = mInv
	}
	return invArrayBig, nil
}

// GenerateKeys returns a private key and public key
func GenerateKeys() (*big.Int, [2]*big.Int, error) {
	privateKey, publicKeyG1, err := cloudflare.RandomG1(rand.Reader)
	if err != nil {
		return nil, empty2Big, err
	}
	publicKey, err := bn256.G1ToBigIntArray(publicKeyG1)
	if err != nil {
		return nil, empty2Big, err
	}
	return privateKey, publicKey, nil
}

// GenerateShares returns encrypted shares, private coefficients, and commitments
func GenerateShares(transportPrivateKey *big.Int, participants ParticipantList) ([]*big.Int, []*big.Int, [][2]*big.Int, error) {
	if transportPrivateKey == nil {
		return nil, nil, nil, errors.New("invalid transportPrivateKey")
	}

	numParticipants := len(participants)
	threshold := ThresholdForUserCount(numParticipants)

	// create coefficients (private/public)
	privateCoefficients, err := cloudflare.ConstructPrivatePolyCoefs(rand.Reader, threshold)
	if err != nil {
		return nil, nil, nil, err
	}
	publicCoefficients := cloudflare.GeneratePublicCoefs(privateCoefficients)

	// create commitments
	commitments := make([][2]*big.Int, len(publicCoefficients))
	for idx, publicCoefficient := range publicCoefficients {
		com, err := bn256.G1ToBigIntArray(publicCoefficient)
		if err != nil {
			return nil, nil, nil, err
		}
		commitments[idx] = com
	}

	// secret shares
	transportPublicKeyG1 := new(cloudflare.G1).ScalarBaseMult(transportPrivateKey)

	// convert public keys into G1 structs
	publicKeyG1s := []*cloudflare.G1{}
	for idx := 0; idx < numParticipants; idx++ {
		participant := participants[idx]
		if participant != nil && participant.PublicKey[0] != nil && participant.PublicKey[1] != nil {
			publicKeyG1, err := bn256.BigIntArrayToG1(participant.PublicKey)
			if err != nil {
				return nil, nil, nil, err
			}
			publicKeyG1s = append(publicKeyG1s, publicKeyG1)
		}
	}

	// check for missing data
	if len(publicKeyG1s) != numParticipants {
		return nil, nil, nil, fmt.Errorf("only have %v of %v public keys", len(publicKeyG1s), numParticipants)
	}

	if len(privateCoefficients) != threshold+1 {
		return nil, nil, nil, fmt.Errorf("only have %v of %v private coefficients", len(privateCoefficients), threshold+1)
	}

	//
	secretsArray, err := cloudflare.GenerateSecretShares(transportPublicKeyG1, privateCoefficients, publicKeyG1s)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate secret shares: %v", err)
	}

	// final encrypted shares
	encryptedShares, err := cloudflare.GenerateEncryptedShares(secretsArray, transportPrivateKey, publicKeyG1s)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate encrypted shares: %v", err)
	}

	return encryptedShares, privateCoefficients, commitments, nil
}

// GenerateKeyShare returns G1 key share, G1 proof, and G2 key share
func GenerateKeyShare(secretValue *big.Int) ([2]*big.Int, [2]*big.Int, [4]*big.Int, error) {
	if secretValue == nil {
		return empty2Big, empty2Big, empty4Big, errors.New("missing secret value, aka private coefficient[0]")
	}

	h1Base, err := cloudflare.HashToG1(h1BaseMessage)
	if err != nil {
		return empty2Big, empty2Big, empty4Big, err
	}
	orderMinus1 := new(big.Int).Sub(cloudflare.Order, common.Big1)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	keyShareG1 := new(cloudflare.G1).ScalarMult(h1Base, secretValue)
	keyShareG1Big, err := bn256.G1ToBigIntArray(keyShareG1)
	if err != nil {
		return empty2Big, empty2Big, empty4Big, err
	}

	// KeyShare G2
	h2Base := new(cloudflare.G2).ScalarBaseMult(common.Big1)
	keyShareG2 := new(cloudflare.G2).ScalarMult(h2Base, secretValue)
	keyShareG2Big, err := bn256.G2ToBigIntArray(keyShareG2)
	if err != nil {
		return empty2Big, empty2Big, empty4Big, err
	}

	// PairingCheck to ensure keyShareG1 and keyShareG2 form valid pair
	validPair := cloudflare.PairingCheck([]*cloudflare.G1{keyShareG1, h1Base}, []*cloudflare.G2{h2Neg, keyShareG2})
	if !validPair {
		return empty2Big, empty2Big, empty4Big, errors.New("key shares not a valid pair")
	}

	// DLEQ Proof
	g1Base := new(cloudflare.G1).ScalarBaseMult(common.Big1)
	g1Value := new(cloudflare.G1).ScalarBaseMult(secretValue)
	keyShareDLEQProof, err := cloudflare.GenerateDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, secretValue, rand.Reader)
	if err != nil {
		return empty2Big, empty2Big, empty4Big, err
	}

	// Verify DLEQ before sending
	err = cloudflare.VerifyDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, keyShareDLEQProof)
	if err != nil {
		return empty2Big, empty2Big, empty4Big, err
	}

	return keyShareG1Big, keyShareDLEQProof, keyShareG2Big, nil
}

// GenerateMasterPublicKey returns the master public key
func GenerateMasterPublicKey(keyShare1s [][2]*big.Int, keyShare2s [][4]*big.Int) ([4]*big.Int, error) {
	if len(keyShare1s) != len(keyShare2s) {
		return empty4Big, errors.New("len(keyShare1s) != len(keyshare2s)")
	}

	// Some predefined stuff to setup
	h1Base, err := cloudflare.HashToG1(h1BaseMessage)
	if err != nil {
		return empty4Big, err
	}
	orderMinus1 := new(big.Int).Sub(cloudflare.Order, common.Big1)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	// Generate master public key
	masterPublicKeyG1 := new(cloudflare.G1)
	masterPublicKeyG2 := new(cloudflare.G2)

	n := len(keyShare1s)

	for idx := 0; idx < n; idx++ {
		keySharedG1, err := bn256.BigIntArrayToG1(keyShare1s[idx])
		if err != nil {
			return empty4Big, err
		}
		masterPublicKeyG1.Add(masterPublicKeyG1, keySharedG1)

		keySharedG2, err := bn256.BigIntArrayToG2(keyShare2s[idx])
		if err != nil {
			return empty4Big, err
		}
		masterPublicKeyG2.Add(masterPublicKeyG2, keySharedG2)
	}

	masterPublicKey, err := bn256.G2ToBigIntArray(masterPublicKeyG2)
	if err != nil {
		return empty4Big, err
	}

	validPair := cloudflare.PairingCheck([]*cloudflare.G1{masterPublicKeyG1, h1Base}, []*cloudflare.G2{h2Neg, masterPublicKeyG2})
	if !validPair {
		return empty4Big, errors.New("invalid pairing for master public key")
	}
	return masterPublicKey, nil
}

// GenerateGroupKeys returns the group private key, group public key, and a signature
func GenerateGroupKeys(transportPrivateKey *big.Int, privateCoefficients []*big.Int, encryptedShares [][]*big.Int, index int, participants ParticipantList) (*big.Int, [4]*big.Int, error) {
	// setup
	n := len(participants)
	threshold := ThresholdForUserCount(n)
	if transportPrivateKey == nil {
		return nil, empty4Big, errors.New("missing transportPrivateKey")
	}
	if index <= 0 {
		return nil, empty4Big, fmt.Errorf("invalid index: require index > 0; index = %v", index)
	}
	if len(privateCoefficients) != threshold+1 {
		return nil, empty4Big, fmt.Errorf("invalid privateCoefficients array: require length == threshold+1; length == %v", len(privateCoefficients))
	}
	if len(encryptedShares) != n {
		return nil, empty4Big, fmt.Errorf("invalid encryptedShares array: require length == len(Participants); length == %v", len(encryptedShares))
	}

	// build portions of group secret key
	publicKeyG1s := make([]*cloudflare.G1, n)

	for idx := 0; idx < n; idx++ {
		publicKeyG1, err := bn256.BigIntArrayToG1(participants[idx].PublicKey)
		if err != nil {
			return nil, empty4Big, fmt.Errorf("error converting public key to g1: %v", err)
		}
		publicKeyG1s[idx] = publicKeyG1
	}

	transportPublicKeyG1 := new(cloudflare.G1).ScalarBaseMult(transportPrivateKey)
	sharedEncrypted, err := cloudflare.CondenseCommitments(transportPublicKeyG1, encryptedShares, publicKeyG1s)
	if err != nil {
		return nil, empty4Big, fmt.Errorf("error condensing commitments: %v", err)
	}

	sharedSecrets, err := cloudflare.GenerateDecryptedShares(transportPrivateKey, sharedEncrypted, publicKeyG1s)
	if err != nil {
		return nil, empty4Big, fmt.Errorf("error generating decrypted shares: %v", err)
	}

	// here's the final group secret
	gskj := cloudflare.PrivatePolyEval(privateCoefficients, index)
	for idx := 0; idx < len(sharedSecrets); idx++ {
		gskj.Add(gskj, sharedSecrets[idx])
	}
	gskj.Mod(gskj, cloudflare.Order)

	// here's the group public
	gpkj := new(cloudflare.G2).ScalarBaseMult(gskj)
	gpkjBig, err := bn256.G2ToBigIntArray(gpkj)
	if err != nil {
		return nil, empty4Big, err
	}

	return gskj, gpkjBig, nil
}

// ComputeDistributedSharesHash computes the distributed shares hash, encrypted shares hash and commitments hash
func ComputeDistributedSharesHash(encryptedShares []*big.Int, commitments [][2]*big.Int) ([32]byte, [32]byte, [32]byte, error) {
	var emptyBytes32 [32]byte

	// encrypted shares hash
	encryptedSharesBin, err := bn256.MarshalBigIntSlice(encryptedShares)
	if err != nil {
		return emptyBytes32, emptyBytes32, emptyBytes32, fmt.Errorf("ComputeDistributedSharesHash encryptedSharesBin failed: %v", err)
	}
	hashSlice := crypto.Hasher(encryptedSharesBin)
	var encryptedSharesHash [32]byte
	copy(encryptedSharesHash[:], hashSlice)

	// commitments hash
	commitmentsBin, err := bn256.MarshalG1BigSlice(commitments)
	if err != nil {
		return emptyBytes32, emptyBytes32, emptyBytes32, fmt.Errorf("ComputeDistributedSharesHash commitmentsBin failed: %v", err)
	}
	hashSlice = crypto.Hasher(commitmentsBin)
	var commitmentsHash [32]byte
	copy(commitmentsHash[:], hashSlice)

	// distributed shares hash
	var distributedSharesBin = append(encryptedSharesHash[:], commitmentsHash[:]...)
	hashSlice = crypto.Hasher(distributedSharesBin)
	var distributedSharesHash [32]byte
	copy(distributedSharesHash[:], hashSlice)

	return distributedSharesHash, encryptedSharesHash, commitmentsHash, nil
}
