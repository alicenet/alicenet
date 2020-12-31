package dkg

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// Evil
var logger *logrus.Logger = logging.GetLogger("dkg")

// Useful pseudo-constants
var (
	empty2Big     [2]*big.Int
	empty4Big     [4]*big.Int
	h1BaseMessage []byte = []byte("MadHive Rocks!")
)

// Participant contains what we know about other participants, i.e. public information
type Participant struct {
	Address   common.Address
	Index     int
	PublicKey [2]*big.Int
}

// ParticipantList is a required type alias since the Sort interface is awful
type ParticipantList []*Participant

// Simplify logging
func (p *Participant) String() string {
	out, err := json.Marshal(p)
	if err != nil {
		return err.Error()
	}

	return string(out)
}

// Len returns the len of the collection
func (pl ParticipantList) Len() int {
	return len(pl)
}

// Less decides if element i is 'Less' than element j -- less ~= before
func (pl ParticipantList) Less(i, j int) bool {
	return pl[i].Index < pl[j].Index
}

// Swap swaps elements i and j within the collection
func (pl ParticipantList) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}

// ThresholdForUserCount returns the threshold user count and k for successful key generation
func ThresholdForUserCount(n int) (int, int) {
	k := n / 3
	threshold := 2 * k
	if (n - 3*k) == 2 {
		threshold = threshold + 1
	}
	return int(threshold), int(k)
}

// InverseArrayForUserCount pre-calculates an inverse array for use by ethereum contracts
func InverseArrayForUserCount(n int) ([]*big.Int, error) {
	bigNeg2 := big.NewInt(-2)
	orderMinus2 := new(big.Int).Add(cloudflare.Order, bigNeg2)

	// Get inverse array; this array is required to help keep gas costs down
	// in the smart contract. Modular multiplication is much cheaper than
	// modular inversion (expopnentiation).
	invArrayBig := make([]*big.Int, n-1)
	for idx := 0; idx < n-1; idx++ {
		m := big.NewInt(int64(idx + 1))
		mInv := new(big.Int).Exp(m, orderMinus2, cloudflare.Order)
		// Confirm
		res := new(big.Int).Mul(m, mInv)
		res.Mod(res, cloudflare.Order)
		if res.Cmp(common.Big1) != 0 {
			return nil, errors.New("Error when computing inverseArray")
		}
		invArrayBig[idx] = mInv
	}
	return invArrayBig, nil
}

// GenerateKeys returns a private key, a public key and potentially an error
func GenerateKeys() (*big.Int, [2]*big.Int, error) {
	privateKey, publicKeyG1, err := cloudflare.RandomG1(rand.Reader)
	publicKey := bn256.G1ToBigIntArray(publicKeyG1)

	return privateKey, publicKey, err
}

// GenerateShares returns encrypted shares, private coefficients, commitments and potentially an error
func GenerateShares(transportPrivateKey *big.Int, transportPublicKey [2]*big.Int, participants ParticipantList, threshold int) ([]*big.Int, []*big.Int, [][2]*big.Int, error) {

	// create coefficients (private/public)
	privateCoefficients, err := cloudflare.ConstructPrivatePolyCoefs(rand.Reader, threshold)
	if err != nil {
		return nil, nil, nil, err
	}
	publicCoefficients := cloudflare.GeneratePublicCoefs(privateCoefficients)

	// create commitments
	commitments := make([][2]*big.Int, len(publicCoefficients))
	for idx, publicCoefficient := range publicCoefficients {
		commitments[idx] = bn256.G1ToBigIntArray(publicCoefficient)
	}

	// secret shares
	transportPublicKeyG1, err := bn256.BigIntArrayToG1(transportPublicKey)
	if err != nil {
		return nil, nil, nil, err
	}

	// convert public keys into G1 structs
	publicKeyG1s := []*cloudflare.G1{}
	for idx := 0; idx < len(participants); idx++ {
		participant := participants[idx]
		logger.Infof("participants[%v]: %v", idx, participant)
		if participant != nil && participant.PublicKey[0] != nil && participant.PublicKey[1] != nil {
			publicKeyG1, err := bn256.BigIntArrayToG1(participant.PublicKey)
			if err != nil {
				return nil, nil, nil, err
			}
			publicKeyG1s = append(publicKeyG1s, publicKeyG1)
		}
	}

	// check for missing data
	if len(publicKeyG1s) != len(participants) {
		return nil, nil, nil, fmt.Errorf("only have %v of %v public keys", len(publicKeyG1s), len(participants))
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

// GenerateKeyShare returns G1 key share, G1 proof, G2 key share and potentially an error
func GenerateKeyShare(firstPrivateCoefficients *big.Int) ([2]*big.Int, [2]*big.Int, [4]*big.Int, error) {

	h1Base, err := cloudflare.HashToG1(h1BaseMessage)
	if err != nil {
		return empty2Big, empty2Big, empty4Big, err
	}
	orderMinus1, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495616", 10)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	if firstPrivateCoefficients == nil {
		return empty2Big, empty2Big, empty4Big, errors.New("Missing secret value, aka private coefficient[0]")
	}

	keyShareG1 := new(cloudflare.G1).ScalarMult(h1Base, firstPrivateCoefficients)
	keyShareG1Big := bn256.G1ToBigIntArray(keyShareG1)

	// KeyShare G2
	h2Base := new(cloudflare.G2).ScalarBaseMult(common.Big1)
	keyShareG2 := new(cloudflare.G2).ScalarMult(h2Base, firstPrivateCoefficients)
	keyShareG2Big := bn256.G2ToBigIntArray(keyShareG2)

	// PairingCheck to ensure keyShareG1 and keyShareG2 form valid pair
	validPair := cloudflare.PairingCheck([]*cloudflare.G1{keyShareG1, h1Base}, []*cloudflare.G2{h2Neg, keyShareG2})
	if !validPair {
		return empty2Big, empty2Big, empty4Big, errors.New("key shares not a valid pair")
	}

	// DLEQ Prooof
	g1Base := new(cloudflare.G1).ScalarBaseMult(common.Big1)
	g1Value := new(cloudflare.G1).ScalarBaseMult(firstPrivateCoefficients)
	keyShareDLEQProof, err := cloudflare.GenerateDLEQProofG1(h1Base, keyShareG1, g1Base, g1Value, firstPrivateCoefficients, rand.Reader)
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
	orderMinus1, _ := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495616", 10)
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

	masterPublicKey := bn256.G2ToBigIntArray(masterPublicKeyG2)

	validPair := cloudflare.PairingCheck([]*cloudflare.G1{masterPublicKeyG1, h1Base}, []*cloudflare.G2{h2Neg, masterPublicKeyG2})
	if !validPair {
		return empty4Big, errors.New("invalid pairing for master public key")
	}

	return masterPublicKey, nil
}

// GenerateGroupKeys returns the group private key, group public key, a signature and potentially an error
func GenerateGroupKeys(initialMessage []byte, transportPrivateKey *big.Int, transportPublicKey [2]*big.Int, privateCoefficients []*big.Int, encryptedShares [][]*big.Int, index int, participants ParticipantList, threshold int) (*big.Int, [4]*big.Int, [2]*big.Int, error) {

	// setup
	n := len(participants)

	// build portions of group secret key
	publicKeyG1s := make([]*cloudflare.G1, n)

	for idx := 0; idx < n; idx++ {
		publicKeyG1, err := bn256.BigIntArrayToG1(participants[idx].PublicKey)
		if err != nil {
			return nil, empty4Big, empty2Big, fmt.Errorf("error converting public key to g1: %v", err)
		}
		publicKeyG1s[idx] = publicKeyG1
	}

	transportPublicKeyG1, err := bn256.BigIntArrayToG1(transportPublicKey)
	if err != nil {
		return nil, empty4Big, empty2Big, fmt.Errorf("error converting transport public key to g1: %v", err)
	}

	sharedEncrypted, err := cloudflare.CondenseCommitments(transportPublicKeyG1, encryptedShares, publicKeyG1s)
	if err != nil {
		return nil, empty4Big, empty2Big, fmt.Errorf("error condensing commitments: %v", err)
	}

	sharedSecrets, err := cloudflare.GenerateDecryptedShares(transportPrivateKey, sharedEncrypted, publicKeyG1s)
	if err != nil {
		return nil, empty4Big, empty2Big, fmt.Errorf("error generating decrypted shares: %v", err)
	}

	// here's the final group secret
	gskj := cloudflare.PrivatePolyEval(privateCoefficients, 1+index)
	for idx := 0; idx < len(sharedSecrets); idx++ {
		gskj.Add(gskj, sharedSecrets[idx])
	}
	gskj.Mod(gskj, cloudflare.Order)

	// here's the group public
	gpkj := new(cloudflare.G2).ScalarBaseMult(gskj)
	gpkjBig := bn256.G2ToBigIntArray(gpkj)

	// create sig
	sig, err := cloudflare.Sign(initialMessage, gskj, cloudflare.HashToG1)
	if err != nil {
		return nil, empty4Big, empty2Big, fmt.Errorf("error signing message: %v", err)
	}
	sigBig := bn256.G1ToBigIntArray(sig)

	// verify signature
	validSig, err := cloudflare.Verify(initialMessage, sig, gpkj, cloudflare.HashToG1)
	if err != nil {
		return nil, empty4Big, empty2Big, fmt.Errorf("error verifying signature: %v", err)
	}

	if !validSig {
		return nil, empty4Big, empty2Big, errors.New("not a valid group signature")
	}

	return gskj, gpkjBig, sigBig, nil
}

// VerifyGroupSigners returns whether the participants are valid or potentially an error
func VerifyGroupSigners(initialMessage []byte, masterPublicKey [4]*big.Int, publishedPublicKeys [][4]*big.Int, publishedSignatures [][2]*big.Int, participants ParticipantList, threshold int) (bool, error) {

	// setup
	n := len(participants)

	signers := threshold + 1
	if signers != n {
		return false, fmt.Errorf("Number of signers (%v) != threshold + 1 (%v)", n, threshold+1)
	}

	// publishedSignatures, indicies and particiapnts must all be the same length
	if !(len(publishedSignatures) == n) {
		return false, fmt.Errorf("len() -> participants:%v publishedSignatures:%v", n, len(publishedSignatures))
	}

	var err error
	indices := make([]int, n)
	publicKeys := make([]*cloudflare.G2, n)
	signatures := make([]*cloudflare.G1, n)
	for idx := 0; idx < n; idx++ {
		participant := participants[idx]

		publicKeys[idx], err = bn256.BigIntArrayToG2(publishedPublicKeys[idx])
		if err != nil {
			return false, fmt.Errorf("failed to convert group public key for %v: %v", idx, err)
		}

		signatures[idx], err = bn256.BigIntArrayToG1(publishedSignatures[idx])
		if err != nil {
			return false, fmt.Errorf("failed to convert signature for %v: %v", idx, err)
		}

		signatureValid, err := cloudflare.Verify(initialMessage, signatures[idx], publicKeys[idx], cloudflare.HashToG1)
		if err != nil {
			return false, fmt.Errorf("failed to verify signature for %v", idx)
		}

		if !signatureValid {
			logger.Warnf("Signature not valid for %v", participant.Index)
		} else {
			logger.Infof("Signature good for %v", participant.Index)
		}

		indices[idx] = participant.Index + 1

		logger.Infof("Participant: 0x%x Idx: %v Index: %v", participant.Address, idx, participant.Index)
	}

	groupSignature, err := cloudflare.AggregateSignatures(signatures, indices, threshold)
	if err != nil {
		return false, err
	}

	masterPublicKeyG2, err := bn256.BigIntArrayToG2(masterPublicKey)

	validGrpSig, err := cloudflare.Verify(initialMessage, groupSignature, masterPublicKeyG2, cloudflare.HashToG1)
	if err != nil {
		return false, fmt.Errorf("Could not verify group signature: %v", err)
	}

	return validGrpSig, nil
}

// CategorizeGroupSigners returns 0 based indicies of honest participants, 0 based indicies of dishonest participants or an error
func CategorizeGroupSigners(initialMessage []byte, masterPublicKey [4]*big.Int, publishedPublicKeys [][4]*big.Int, publishedSignatures [][2]*big.Int, participants ParticipantList, threshold int) ([]int, []int, error) {

	// useful bit of info
	n := len(participants)
	chunkSize := threshold + 1

	// if we can't meet threshold we can't do much
	if n < chunkSize {
		return []int{}, []int{}, fmt.Errorf("not enough signers (%v) to meet threshold + 1 (%v)", n, chunkSize)
	}

	// len(publishedPublicKeys) must equal len(publishedSignatures) must equal len(participants)
	if n != len(publishedPublicKeys) || n != len(publishedSignatures) {
		return []int{}, []int{}, fmt.Errorf(
			"mismatched public keys (%v), signatures (%v) and participants (%v)", len(publishedPublicKeys), len(publishedSignatures), n)
	}

	// Now we we chunk arrays and verify chunks seperately
	knownGood := make([]bool, n)
	for begin := 0; begin < n; begin += chunkSize {
		end := begin + chunkSize
		if end > n {
			begin -= (end - n)
			end = n
		}

		logger.Infof("Verifying %v >= index > %v", begin, end)

		groupPublicKeys := publishedPublicKeys[begin:end]
		groupSignatures := publishedSignatures[begin:end]
		groupParticipants := participants[begin:end]

		good, err := VerifyGroupSigners(initialMessage, masterPublicKey, groupPublicKeys, groupSignatures, groupParticipants, threshold)
		if err != nil {
			return []int{}, []int{}, fmt.Errorf("failed verifying group signers between %v and %v: %v", begin, end, err)
		}

		// if the chunk verified then we mark each element as good
		if good {
			for idx := begin; idx < end; idx++ {
				knownGood[idx] = true // TODO this should be the participant index not idx
			}
		}
		logger.Infof("VerifyGroupSigners([%v:%v]): %v -> %v", begin, end, knownGood, good)
	}

	// Hopefully everything is good
	allGood := all(knownGood)
	logger.Infof("VerifyGroupSigners(...): %v", allGood)

	indices := make([]int, n)
	for idx, participant := range participants {
		indices[idx] = participant.Index
	}

	if allGood {
		return indices, []int{}, nil
	}

	return []int{}, indices, nil
}

// ------------------------------------
func all(m []bool) bool {
	for _, v := range m {
		if !v {
			return false
		}
	}
	return true
}
