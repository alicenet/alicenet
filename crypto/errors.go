package crypto

import "errors"

var (
	// ErrPubkeyGroupNotSet occurs when the group public key
	// (master public key [mpk]) has not been set
	ErrPubkeyGroupNotSet = errors.New("groupPubk not set")

	// ErrPrivkNotSet occurs when the private key has not been set
	ErrPrivkNotSet = errors.New("privk not set")

	// ErrInvalidSignature occurs when signature validation fails
	ErrInvalidSignature = errors.New("signature validation failed")

	// ErrInvalid occurs when signer is not valid;
	// this occurs when signer is not initialized.
	ErrInvalid = errors.New("invalid signer")

	// ErrInvalidPubkeyShares occurs when multiple copies of the same public
	// key are contained when attempting to set GroupShares.
	ErrInvalidPubkeyShares = errors.New("groupShares contains repeated public keys")
)
