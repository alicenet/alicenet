//go:build tools

package MadNet

// This file tracks tool dependencies.
// This is so that "go mod tidy" doesnt remove deps that we actually use for generate commands.
// It will not actually be compiled due to the build tag used above.

import (
	_ "github.com/MadBase/go-capnproto2/v2/capnpc-go"
	_ "github.com/elazarl/go-bindata-assetfs"
	_ "github.com/go-bindata/go-bindata/v3"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
