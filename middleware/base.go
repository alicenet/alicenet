package middleware

import (
	"google.golang.org/grpc"
)

type P2PMiddleware interface {
	grpc.CallOption
	isMW()
}

type P2POption struct {
	CanBlock bool
}
