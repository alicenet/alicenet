package transport

import "errors"

var (

	// ErrUint32Overflow occurs when attempting to decode a string
	// representation of bytes and the number of bytes > 4, which
	// would overflow if one attempts to convert the string to uint32.
	ErrUint32Overflow = errors.New("invalid uint32 from hex conversion. Number overflows uint32")

	// ErrHandshakeTimeout occurs in (client|server)Mux when the handshake
	// takes too long.
	ErrHandshakeTimeout = errors.New("handshake timeout")

	// ErrListenerClosed occurs if any channel receives a nil value
	// when called by Accept or AcceptFailures.
	ErrListenerClosed = errors.New("listener closed")

	// ErrInvalidPubKeyLength occurs in NewP2PAddrPort when public key hex
	// string has the incorrect length.
	ErrInvalidPubKeyLength = errors.New("invalid public key format: Length")

	// ErrEmptyPrivKey occurs when private key byte slice is empty;
	// this is an invalid private key.
	ErrEmptyPrivKey = errors.New("empty private key hex string")

	// ErrInvalidPrivKey occurs when private key bytes is strictly less than
	// 16 bytes in length; this is an invalid private key.
	ErrInvalidPrivKey = errors.New("invalid private key hex string")
)
