package math

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/MadBase/MadNet/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// Distributed Key Generation related errors
var (
	ErrInsufficientGoodSigners = errors.New("Insufficient non-malicious signers to identify malicious signers")
	ErrTooFew                  = errors.New("Building array of size n with less than n required + optional")
	ErrTooMany                 = errors.New("Building array of size n with more than n required")
)

// Evil
var logger *logrus.Logger = logging.GetLogger("dkg")

// Useful pseudo-constants
var (
	empty2Big     [2]*big.Int
	empty4Big     [4]*big.Int
	h1BaseMessage []byte = []byte("MadHive Rocks!")
)

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
func GenerateShares(transportPrivateKey *big.Int, transportPublicKey [2]*big.Int, participants objects.ParticipantList, threshold int) ([]*big.Int, []*big.Int, [][2]*big.Int, error) {

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
		logger.Debugf("participants[%v]: %v", idx, participant)
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
func GenerateGroupKeys(initialMessage []byte, transportPrivateKey *big.Int, transportPublicKey [2]*big.Int, privateCoefficients []*big.Int, encryptedShares [][]*big.Int, index int, participants objects.ParticipantList, threshold int) (*big.Int, [4]*big.Int, [2]*big.Int, error) {

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
func VerifyGroupSigners(initialMessage []byte, masterPublicKey [4]*big.Int, publishedPublicKeys [][4]*big.Int, publishedSignatures [][2]*big.Int, participants objects.ParticipantList, threshold int) (bool, error) {

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

		indices[idx] = participant.Index + 1
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
			logger.Debugf("Signature good for %v", participant.Index)
		}

		logger.Debugf("Participant: 0x%x Idx: %v Index: %v", participant.Address, idx, participant.Index)
	}

	logger.Debugf("Aggregating Signatures: %v", signatures)
	logger.Debugf("Aggregating Indices: %v", indices)
	logger.Debugf("Aggregating Threshold: %v", threshold)
	groupSignature, err := cloudflare.AggregateSignatures(signatures, indices, threshold)
	if err != nil {
		return false, err
	}

	masterPublicKeyG2, err := bn256.BigIntArrayToG2(masterPublicKey)

	validGrpSig, err := cloudflare.Verify(initialMessage, groupSignature, masterPublicKeyG2, cloudflare.HashToG1)
	logger.Debugf("GroupSignature Verify:%v", validGrpSig)
	if err != nil {
		return false, fmt.Errorf("Could not verify group signature: %v", err)
	}

	return validGrpSig, nil
}

func Reverse(a []int) []int {
	n := len(a)
	b := make([]int, n)
	j := n / 2
	for idx := 0; idx <= j; idx++ {
		b[idx], b[n-idx-1] = a[n-idx-1], a[idx]
	}
	return b
}

func NChooseK(n int, k int, visitor func([]int) (bool, error)) (bool, []int, error) {

	c := make([]int, k+3) // We're just going to ignore index 0

	// L1
	for j := 1; j <= k; j++ {
		c[j] = j - 1
	}
	c[k+1] = n
	c[k+2] = 0

	// Loop
	var err error
	success := false
	done := false
	for done == false && success == false {

		// L2
		success, err = visitor(c[1 : k+1])
		if err != nil {
			return false, nil, err
		}
		if !success {

			// L3
			j := 1
			for c[j]+1 == c[j+1] {
				c[j] = j - 1
				j++
			}

			// L4
			if j > k {
				done = true
			} else {
				// L5
				c[j] = c[j] + 1
			}
		}

	}

	return success, c[1 : k+1], nil
}

// CategorizeGroupSigners returns 0 based indicies of honest participants, 0 based indicies of dishonest participants or an error
func CategorizeGroupSigners(initialMessage []byte, masterPublicKey [4]*big.Int, publishedPublicKeys [][4]*big.Int, publishedSignatures [][2]*big.Int, participants objects.ParticipantList, threshold int) ([]int, []int, error) {

	// Setup + sanity checks before starting
	n := len(participants)
	k := threshold + 1
	if n < k {
		return []int{}, []int{}, fmt.Errorf("not enough signers (%v) to meet threshold + 1 (%v)", n, k)
	}

	// len(publishedPublicKeys) must equal len(publishedSignatures) must equal len(participants)
	if n != len(publishedPublicKeys) || n != len(publishedSignatures) {
		return []int{}, []int{}, fmt.Errorf(
			"mismatched public keys (%v), signatures (%v) and participants (%v)", len(publishedPublicKeys), len(publishedSignatures), n)
	}

	// This function is used  when visiting a combination to determine if signer group is valid
	visitor := func(indices []int) (bool, error) {

		// Build the public keys, signatures and participants
		groupPublicKeys := make([][4]*big.Int, k)
		groupSignatures := make([][2]*big.Int, k)
		groupParticipants := make([]*objects.Participant, k)

		for idx := 0; idx < k; idx++ {
			groupPublicKeys[idx] = publishedPublicKeys[indices[idx]]
			groupSignatures[idx] = publishedSignatures[indices[idx]]
			groupParticipants[idx] = participants[indices[idx]]
		}

		good, err := VerifyGroupSigners(initialMessage, masterPublicKey, groupPublicKeys, groupSignatures, groupParticipants, threshold)
		if err != nil {
			return false, err
		}

		return good, err
	}

	// Not visiting all the combinations, just looking for the first with good signers
	success, good, err := NChooseK(n, k, visitor)
	if err != nil {
		return []int{}, []int{}, err
	}
	if !success {
		return []int{}, []int{}, ErrInsufficientGoodSigners
	}

	// We have a good set which we know is > half so we check the rest and hope they are also good
	logger.Debugf("CAT good indices:%v", good)

	unknown := make([]int, n-k)
	oic := 0
	for idx := 0; idx < n; idx++ {
		if !contains(good, idx) {
			unknown[oic] = idx
			oic++
		}
	}

	logger.Debugf("CAT unknown indices:%v", unknown)
	ts, err := PadCollection(k, unknown, good)
	if err != nil {
		return nil, nil, err
	}

	logger.Debugf("CAT testing indices:%v", ts)

	mb, err := visitor(ts)
	if err != nil {
		return nil, nil, err
	}

	if mb {
		good = CombineCollections(good, unknown)
		logger.Debugf("CAT good indices:%v", good)
		return good, []int{}, nil
	}

	// Ok, here we have a good batch and at least 1 bad; so we loop over the unknowns
	bad := make([]int, 0)

	for idx := 0; idx < len(unknown); idx++ {
		scratch := make([]int, len(good))
		copy(scratch, good)
		scratch[0] = unknown[idx]

		stillGood, err := visitor(scratch)
		if err != nil {
			return nil, nil, err
		}

		if stillGood {
			good = append(good, unknown[idx])
		} else {
			bad = append(bad, unknown[idx])
		}
	}

	return good, bad, nil
}

func contains(collection []int, number int) bool {
	for idx := range collection {
		if collection[idx] == number {
			return true
		}
	}

	return false
}

func PadCollection(sz int, definite []int, possible []int) ([]int, error) {

	// Sanity checks
	dn := len(definite)
	if dn > sz {
		return nil, ErrTooMany
	}

	pn := len(possible)
	if sz > dn+pn {
		return nil, ErrTooFew
	}

	// Setup
	n := make([]int, sz)

	//
	didx := 0
	pidx := 0
	for idx := 0; idx < sz; idx++ {
		if idx < dn {
			n[idx] = definite[didx]
			didx++
		} else if idx < pn {
			n[idx] = possible[pidx]
			pidx++
		}
	}

	return n, nil
}

func CombineCollections(a []int, b []int) []int {
	la := len(a)
	lb := len(b)
	c := make([]int, la)

	copy(c, a)

	// cidx := la

	logger.Debugf("a:%v", a)
	logger.Debugf("b:%v", b)
	logger.Debugf("c:%v", c)

	for idx := 0; idx < lb; idx++ {
		if !contains(c, b[idx]) {
			c = append(c, b[idx])
		}
	}

	return c
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
