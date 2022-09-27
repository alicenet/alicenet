package dynamics

import (
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v2"
)

var (
	// ErrDynamicValueNilPointer is an error which results from a
	// DynamicValues struct which has not been initialized.
	ErrDynamicValueNilPointer = errors.New("invalid DynamicValues: nil pointer")

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
	// for updating dynamicValues is invalid.
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

type ErrInvalidDynamicValueStructLen struct {
	data        string
	actualLen   int
	expectedLen int
}

func (e *ErrInvalidDynamicValueStructLen) Error() string {
	return fmt.Sprintf("Got data %s with length %d, expected length %d", e.data, e.actualLen, e.expectedLen)
}

type ErrInvalidDynamicValue struct {
	err  string
	name string
}

func (e *ErrInvalidDynamicValue) Error() string {
	return fmt.Sprintf("failed to decode %s in dynamic value: %v", e.name, e.err)
}

type ErrInvalidSize struct {
	message string
}

func (e *ErrInvalidSize) Error() string {
	return e.message
}
