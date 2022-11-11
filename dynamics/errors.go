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

	// ErrNodeValueNilPointer is an error which results from a node which has not
	// been initialized.
	ErrNodeValueNilPointer = errors.New("invalid node: nil pointer")

	// ErrZeroEpoch is an error which is raised whenever the epoch is given
	// as zero; there is no zero epoch.
	ErrZeroEpoch = errors.New("invalid epoch: no zero epoch")

	// ErrUnmarshalEmpty is an error which is raised whenever attempting
	// to unmarshal an empty byte slice.
	ErrUnmarshalEmpty = errors.New("invalid: attempting to unmarshal empty byte slice")

	// ErrValueIsEmpty is an error which is raised whenever attempting
	// to copy a not initialized dynamic value.
	ErrValueIsEmpty = errors.New("invalid: value is empty or not initialized")

	// ErrKeyNotPresent is an error which is raised when a key is not present
	// in the database.
	ErrKeyNotPresent = badger.ErrKeyNotFound

	// ErrInvalid is an error which is returned when the struct is invalid.
	ErrInvalid = errors.New("invalid value")

	// ErrInvalidNodeKey is an error which occurs when the NodeKey is invalid
	ErrInvalidNodeKey = errors.New("invalid nodeKey")

	// ErrInvalidNode is an error which occurs when a previous Node is invalid
	ErrInvalidPrevNode = errors.New("invalid previous node")

	// ErrInvalidLinkedList is an error which occurs when a linked list is invalid
	// or corrupted
	ErrInvalidLinkedList = errors.New("invalid linked list")
)

type ErrInvalidDynamicValueStructLen struct {
	data        string
	actualLen   int
	expectedLen int
}

func (e *ErrInvalidDynamicValueStructLen) Error() string {
	return fmt.Sprintf("got data %s with length %d, expected length %d", e.data, e.actualLen, e.expectedLen)
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

type ErrInvalidNode struct {
	node *Node
}

func (e *ErrInvalidNode) Error() string {
	return fmt.Sprintf("invalid node: %+v", e.node)
}
