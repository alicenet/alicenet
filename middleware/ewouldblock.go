package middleware

import (
	"errors"

	"google.golang.org/grpc"
)

// ErrWouldBlock is an error that indicates a WithNoBlocking request failed
// due to blocking
var ErrWouldBlock = errors.New("would block")

type backpressureCallOption struct {
	*grpc.EmptyCallOption
}

//nolint:unused
func (pm *backpressureCallOption) pushback() {}

// WithNoBlocking generates a grpc.CallOption that is honored by the methods
// of the P2P system such that any request that contains this CallOption will
// not block while trying to be passed to the work buffers. If the call would
// block then an error should be raised.
func WithNoBlocking() grpc.CallOption {
	return &backpressureCallOption{&grpc.EmptyCallOption{}}
}

// CanBlock is a function that allows the P2P system to check for the existence
// of the WithNoBlocking CallOption to determine request specific blocking
// policy.
func CanBlock(opts ...grpc.CallOption) bool {
	for i := 0; i < len(opts); i++ {
		_, ok := opts[i].(*backpressureCallOption)
		if ok {
			return false
		}
	}
	return true
}
