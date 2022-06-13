//go:build tools

package MadNet

// This file tracks tool dependencies.
// This is so that "go mod tidy" doesnt remove deps that we actually use for generate commands.
// It will not actually be compiled due to the build tag used above.

//go:generate buf generate

import (
	_ "github.com/MadBase/go-capnproto2/v2/capnpc-go"
	_ "github.com/derision-test/go-mockgen/cmd/go-mockgen"
	_ "github.com/vburenin/ifacemaker"
)
