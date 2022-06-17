package accusation

import "github.com/google/uuid"

type Accusation interface {
	SubmitToSmartContracts() error
	GetUUID() uuid.UUID
	SetUUID(uuid uuid.UUID)
	// IsProcessed() bool
	// MarshallBinary() ([]byte, error)
	// UnmarshalBinary([]byte) error
	// hash evidence as ID
}
