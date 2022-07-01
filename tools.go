//go:build tools

package alicenet

// This file tracks tool dependencies.
// This is so that "go mod tidy" doesnt remove deps that we actually use for generate commands.
// It will not actually be compiled due to the build tag used above.

//go:generate go run github.com/bufbuild/buf/cmd/buf generate

import (
	_ "github.com/MadBase/go-capnproto2/v2/capnpc-go"
	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/derision-test/go-mockgen/cmd/go-mockgen"
	_ "github.com/ethereum/go-ethereum/cmd/abigen"
	_ "github.com/ethereum/go-ethereum/cmd/ethkey"
	_ "github.com/ethereum/go-ethereum/cmd/geth"
	_ "github.com/vburenin/ifacemaker"
	_ "golang.org/x/tools/cmd/goimports"
)
