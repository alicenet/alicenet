package interfaces

// KeyResolver allows a service to request a key from an object conforming to
// the interface. This is used to decrypt objects stored in the Database
type KeyResolver interface {
	GetKey(kid []byte) ([]byte, error)
}
