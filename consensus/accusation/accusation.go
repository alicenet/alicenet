package accusation

import "github.com/google/uuid"

type Accusation interface {
	SubmitToSmartContracts() error
	GetUUID() uuid.UUID
	SetUUID(uuid uuid.UUID)
}
