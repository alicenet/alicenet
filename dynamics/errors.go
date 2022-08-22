package dynamics

import (
	"errors"

	"github.com/dgraph-io/badger/v2"
)

var (
	// ErrRawStorageNilPointer is an error which results from a
	// RawStorage struct which has not been initialized.
	ErrRawStorageNilPointer = errors.New("invalid RawStorage: nil pointer")

	// ErrZeroEpoch is an error which is raised whenever the epoch is given
	// as zero; there is no zero epoch.
	ErrZeroEpoch = errors.New("invalid epoch: no zero epoch")

	// ErrUnmarshalEmpty is an error which is raised whenever attempting
	// to unmarshal an empty byte slice.
	ErrUnmarshalEmpty = errors.New("invalid: attempting to unmarshal empty byte slice")

	// ErrKeyNotPresent is an error which is raised when a key is not present
	// in the database.
	ErrKeyNotPresent = badger.ErrKeyNotFound

	// ErrInvalidUpdateValue is an error which is returned when the state
	// for updating rawStorage is invalid.
	ErrInvalidUpdateValue = errors.New("invalid update value for storage")

	// ErrInvalidValue is an error which is returned when the value is invalid.
	ErrInvalidValue = errors.New("invalid value")

	// ErrInvalid is an error which is returned when the struct is invalid.
	ErrInvalid = errors.New("invalid value")

	// ErrInvalidNodeKey is an error which occurs when the NodeKey is invalid
	ErrInvalidNodeKey = errors.New("invalid NodeKey")

	// ErrInvalidNode is an error which occurs when a Node is invalid
	ErrInvalidNode = errors.New("invalid Node")
)
