//go:build tools

package MadNet

// Package tools records tool dependencies. It cannot actually be compiled.

import (
	_ "github.com/MadBase/go-capnproto2/v2/capnpc-go"
	_ "github.com/elazarl/go-bindata-assetfs"
	_ "github.com/go-bindata/go-bindata/v3"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
