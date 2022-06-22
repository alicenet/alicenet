package objs

import "github.com/google/uuid"

type Accusation interface {
	SubmitToSmartContracts() error
	GetUUID() uuid.UUID
	SetUUID(uuid uuid.UUID)
	// IsProcessed() bool
	// MarshalBinary() ([]byte, error)
	// UnmarshalBinary([]byte) error
	// hash evidence as ID
}
