package dkgtasks

import (
	"context"
	"math/big"

	"github.com/MadBase/bridge/bindings"
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
func CheckRegistration(ethdkg *bindings.ETHDKG,
	logger *logrus.Entry,
	callOpts *bind.CallOpts,
	addr common.Address,
	publicKey [2]*big.Int) (RegistrationStatus, error) {

	var receivedPublicKey [2]*big.Int
	var err error

	participantState, err := ethdkg.GetParticipantInternalState(callOpts, addr)
	if err != nil {
		logger.Warnf("could not check if we're registered: %v", err)
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
		return Undefined, nil
	}

	if participantState.Nonce != nonce.Uint64() {
		return NoRegistration, nil
	}

	// Check if expected public key is registered
	if !(receivedPublicKey[0].Cmp(publicKey[0]) == 0 &&
		receivedPublicKey[1].Cmp(publicKey[1]) == 0) {

		logger.Warnf("address (%v) is already registered with another publicKey %x", addr.Hex(), receivedPublicKey)

		return BadRegistration, nil
	}

	return Registered, nil
}

// CheckKeyShare checks if a given address submitted the keyshare expected
func CheckKeyShare(ctx context.Context, ethdkg *bindings.ETHDKG,
	logger *logrus.Entry,
	callOpts *bind.CallOpts,
	addr common.Address,
	keyshare [2]*big.Int) (KeyShareStatus, error) {

	var receivedKeyShare [2]*big.Int
	var err error

	participantState, err := ethdkg.GetParticipantInternalState(callOpts, addr)
	if err != nil {
		logger.Warnf("could not check if we're registered: %v", err)
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

		logger.Warnf("address (%v) is already registered with %x", addr.Hex(), receivedKeyShare)

		return BadKeyShared, nil
	}

	return KeyShared, nil
}
