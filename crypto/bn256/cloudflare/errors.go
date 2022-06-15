package cloudflare

import "errors"

var (
	// ErrDangerousPoint occurs when HashToG1 attempts to return a point p
	// in G1 whose value is dangerous when they arise in signature generation;
	// in particular, Infinity (the identity element) or the curve generator
	// or its negation.
	// These points introduce security concerns should they be
	// used as valid hash G1 points for a signature. Therefore, we will
	// raise an error should they ever occur.
	ErrDangerousPoint = errors.New("hashToG1: p == curveGen or p == Inf or p == negCurveGen; dangerous hash point for signatures")

	// ErrInvalidPoint occurs when HashToG1 or HashToG2 produces a point p
	// which is not actually on the corresponding elliptic curve (curve or
	// twist). This should never happen.
	ErrInvalidPoint = errors.New("hashToG1/HashToG2: hash point not on curve")

	// ErrInvalid occurs when something generic is invalid.
	ErrInvalid = errors.New("invalid state")

	// ErrInvalidSharedSecret means that the secret shared by this participant
	// is invalid; this implies this participant performed a malicious action.
	ErrInvalidSharedSecret = errors.New("compareSharedSecret: Failed shared secret")

	// ErrInvalidThreshold occurs when threshold is less than 2, which
	// is not possible.
	ErrInvalidThreshold = errors.New("threshold < 2; not possible")

	// ErrInvalidSignatureLength occurs when marshalled signature is not
	// 6*32 bytes in length.
	ErrInvalidSignatureLength = errors.New("marshalled Signature: incorrect byte length")

	// ErrMissingIndex occurs when a public key is missing from the array
	// of elements meant to store all of the public keys for computing
	// public indices.
	ErrMissingIndex = errors.New("unable to find index for missing public key")

	// ErrBelowThreshold occurs when there are not enough signatures present
	// to sign a message.
	ErrBelowThreshold = errors.New("signatures below required threshold")

	// ErrLIArrayMismatch occurs in LagrangeInterpolationG1 when pointsG1
	// and indices arrays do not have the same length.
	ErrLIArrayMismatch = errors.New("lagrangeInterpolation: Mismatch between interpolation points and indices")

	// ErrInsufficientData occurs when attempting to Unmarshal bytes to form
	// G1, G2, or GT element, which requires 32, 2*32, or 12*32 bytes,
	// respectively.
	ErrInsufficientData = errors.New("cloudflare: Insufficient state to construct point")

	// ErrMalformedPoint occurs when submitted byte slice does not correspond
	// to valid curve point in G1, G2, or GT.
	ErrMalformedPoint = errors.New("cloudflare: Byte slice yielded invalid curve point")

	// ErrDLEQInvalidProof occurs when the submitted DLEQ proof is invalid.
	ErrDLEQInvalidProof = errors.New("invalid DLEQ Proof")

	// ErrInvalidCoordinate occurs when submitted gfP coordinate equals
	// or exceeds modulus P.
	ErrInvalidCoordinate = errors.New("cloudflare: coordinate equals or exceeds modulus")

	// ErrMismatchedSlices occurs when signature and index slices have
	// different lengths.
	ErrMismatchedSlices = errors.New("sig and indices slices have different lengths")

	// ErrArrayMismatch occurs when attempting to condense the encrypted
	// big.Ints from other participants into one array but the lengths
	// do not match.
	ErrArrayMismatch = errors.New("arrays have incompatible lengths")
)
