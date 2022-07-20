package transport

import (
	"net"
)

var _ net.Addr = (*addr)(nil)

// This struct is used to created objects that conform to the
// net.Addr interface.
type addr struct {
	network string
	address string
}

// Network ... See net.Addr docs.
func (ad *addr) Network() string {
	return ad.network
}

// String ... See net.Addr docs.
func (ad *addr) String() string {
	return ad.address
}
