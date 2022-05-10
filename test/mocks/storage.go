package mocks

import "github.com/MadBase/MadNet/dynamics"

func NewMockStorage() *dynamics.Storage {
	storage := &dynamics.Storage{}
	storage.Init(NewMockDb(), NewMockLogger())
	return storage
}
