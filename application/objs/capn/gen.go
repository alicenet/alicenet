package capn

import (
	capnp "zombiezen.com/go/capnproto2"
)

// Import check to ensure capnp is installed.
var _ = capnp.Tag

//go:generate capnp compile -I /home/et3p/go/src/zombiezen.com/go/capnproto2/std -ogo application.capnp
