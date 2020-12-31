package objs

// Signer is the interface for a signature generation struct.
type Signer interface {
	Sign(msg []byte) (sig []byte, err error)
	Pubkey() ([]byte, error)
}
