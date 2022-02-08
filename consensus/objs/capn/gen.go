package capn

import (
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Import check to ensure capnp is installed.
var _ = capnp.Tag

//go:generate capnp compile -I $GOPATH/pkg/mod/github.com/!mad!base/go-capnproto2/v2@v2.18.0-custom-schema.1/std -ogo consensus.capnp
