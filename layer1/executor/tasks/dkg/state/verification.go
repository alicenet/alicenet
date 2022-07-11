package state

import (
	"context"
	"errors"
	"math/big"

	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/crypto/bn256"
	"github.com/alicenet/alicenet/crypto/bn256/cloudflare"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// RegistrationStatus is an enumeration indicating the current status of a registration
type RegistrationStatus int

// The possible registration statuses:
// * Undefined       - unknown if the address is registered
// * Registered      - address is registered with expected public key
// * NoRegistration  - address does not have a public key registered
// * BadRegistration - address is regisered with an unexpected public key
const (
	Undefined RegistrationStatus = iota
	Registered
	NoRegistration
	BadRegistration
)

func (status RegistrationStatus) String() string {
	return [...]string{
		"Undefined",
		"Registered",
		"NoRegistration",
		"BadRegistration",
	}[status]
}

// KeyShareStatus is an enumeration indicated the current status of keyshare
type KeyShareStatus int

// The possible key share statuses:
// * UnknownKeyShare - unknown if the address has shared a key
// * KeyShared       - address has shared the expected key
// * NoKeyShared     - address does not have a key share
// * BadKeyShared    - address has an unexpected key share
const (
	UnknownKeyShare KeyShareStatus = iota
	KeyShared
	NoKeyShared
	BadKeyShared
)

// CheckRegistration checks if given address is registered as expected
func CheckRegistration(ethdkg bindings.IETHDKG,
	logger *logrus.Entry,
	callOpts *bind.CallOpts,
	addr common.Address,
	publicKey [2]*big.Int) (RegistrationStatus, error) {

	var receivedPublicKey [2]*big.Int
	var err error

	participantState, err := ethdkg.GetParticipantInternalState(callOpts, addr)
	if err != nil {
		logger.Debugf("could not check if we're registered: %v", err)
		return Undefined, err
	}
	// Grab both parts of registered public key
	receivedPublicKey = participantState.PublicKey

	// Check if anything is registered
	if receivedPublicKey[0].Cmp(big.NewInt(0)) == 0 && receivedPublicKey[1].Cmp(big.NewInt(0)) == 0 {
		return NoRegistration, nil
	}

	// get ethdkg nonce
	nonce, err := ethdkg.GetNonce(callOpts)
	if err != nil {
		return Undefined, err
	}

	if participantState.Nonce != nonce.Uint64() {
		return NoRegistration, nil
	}

	// Check if expected public key is registered
	if !(receivedPublicKey[0].Cmp(publicKey[0]) == 0 &&
		receivedPublicKey[1].Cmp(publicKey[1]) == 0) {

		logger.Debugf("address (%v) is already registered with another publicKey %x", addr.Hex(), receivedPublicKey)

		return BadRegistration, nil
	}

	return Registered, nil
}

// CheckKeyShare checks if a given address submitted the keyshare expected
func CheckKeyShare(ctx context.Context, ethdkg bindings.IETHDKG,
	logger *logrus.Entry,
	callOpts *bind.CallOpts,
	addr common.Address,
	keyshare [2]*big.Int) (KeyShareStatus, error) {

	var receivedKeyShare [2]*big.Int
	var err error

	participantState, err := ethdkg.GetParticipantInternalState(callOpts, addr)
	if err != nil {
		logger.Debugf("could not check if we're registered: %v", err)
		return NoKeyShared, err
	}
	// Grab both parts of registered public key
	receivedKeyShare = participantState.KeyShares

	// Check if anything is registered
	if receivedKeyShare[0].Cmp(big.NewInt(0)) == 0 && receivedKeyShare[1].Cmp(big.NewInt(0)) == 0 {
		return NoKeyShared, nil
	}

	// Check if expected public key is registered
	if receivedKeyShare[0].Cmp(keyshare[0]) != 0 &&
		receivedKeyShare[1].Cmp(keyshare[1]) != 0 {

		logger.Debugf("address (%v) is already registered with %x", addr.Hex(), receivedKeyShare)

		return BadKeyShared, nil
	}

	return KeyShared, nil
}

// VerifyDistributedShares verifies the distributed shares and returns
//		true/false if the share is valid/invalid;
//		true/false if present/not present;
// 		error if raised
//
// If an error is raised, then something unrecoverable has occurred.
func VerifyDistributedShares(dkgState *DkgState, participant *Participant) (bool, bool, error) {
	if dkgState == nil {
		return false, false, errors.New("invalid dkgState")
	}
	if participant == nil {
		return false, false, errors.New("invalid participant")
	}
	if dkgState.TransportPrivateKey == nil {
		return false, false, errors.New("transport private key not set")
	}

	// Check participant is not self
	if dkgState.Index == participant.Index {
		// We do not verify our own submission
		return true, true, nil
	}

	n := dkgState.NumberOfValidators
	// TODO: this hardcoded value should reference the minimum some place else
	if n < 4 {
		return false, false, errors.New("invalid participants; not enough validators")
	}
	threshold := ThresholdForUserCount(int(n))

	// Get commitments
	commitments := dkgState.Participants[participant.Address].Commitments
	// Get encryptedShares
	encryptedShares := dkgState.Participants[participant.Address].EncryptedShares

	// confirm correct length of commitments
	if len(commitments) != threshold+1 {
		return false, false, errors.New("invalid commitments: incorrect length")
	}

	// confirm correct length of encryptedShares
	if len(encryptedShares) != int(n)-1 {
		return false, false, errors.New("invalid encryptedShares: incorrect length")
	}

	// Perform commitment conversions
	publicCoefficients := make([]*cloudflare.G1, threshold+1)
	for i := 0; i < threshold+1; i++ {
		tmp, err := bn256.BigIntArrayToG1(commitments[i])
		if err != nil {
			return false, false, errors.New("invalid commitment: failed conversion")
		}
		publicCoefficients[i] = new(cloudflare.G1)
		publicCoefficients[i].Set(tmp)
	}

	// Get public key
	publicKeyG1, err := bn256.BigIntArrayToG1(participant.PublicKey)
	if err != nil {
		return false, false, err
	}

	// Decrypt secret
	encShareIdx := dkgState.Index - 1
	if participant.Index < dkgState.Index {
		encShareIdx--
	}
	encryptedSecret := encryptedShares[encShareIdx]
	secret := cloudflare.Decrypt(encryptedSecret, dkgState.TransportPrivateKey, publicKeyG1, dkgState.Index)

	// Compare shared secret
	valid, err := cloudflare.CompareSharedSecret(secret, dkgState.Index, publicCoefficients)
	if err != nil {
		return false, false, err
	}
	if !valid {
		// Invalid shared secret; submit dispute
		return false, true, nil
	}
	// Valid shared secret
	return true, true, nil
}
