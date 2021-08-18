package middleware

import (
	"errors"

	"google.golang.org/grpc"
)

var ErrWouldBlock = errors.New("Call would block")

type backpressureCallOption struct {
	*grpc.EmptyCallOption
}

func (pm *backpressureCallOption) pushback() {}

func WithNoBlocking() grpc.CallOption {
	return &backpressureCallOption{&grpc.EmptyCallOption{}}
}

func CanBlock(opts ...grpc.CallOption) bool {
	for i := 0; i < len(opts); i++ {
		_, ok := opts[i].(*backpressureCallOption)
		if ok {
			return false
		}
	}
	return true
}
