package mocks

import "github.com/MadBase/MadNet/dynamics"

func NewMockStorage() *dynamics.Storage {
	storage := &dynamics.Storage{}
	err := storage.Init(NewTestDB(), NewMockLogger())
	if err != nil {
		panic(err)
	}
	return storage
}
