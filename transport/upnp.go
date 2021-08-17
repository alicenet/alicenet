package transport

import (
	"sync"

	"github.com/jcuga/go-upnp"
	"github.com/sirupsen/logrus"
)

type UPnPProtocol int

const (
	TCP UPnPProtocol = iota
	UDP
)

type UPnPMapper struct {
	// This is the logger for the transport
	logger *logrus.Logger
	// Port to map
	port uint16
	// Protocol to map
	protoStr string
	// Gateway
	gateway upnp.IGD
	// Start / stop mutex
	mux sync.Mutex
}

func NewUPnPMapper(logger *logrus.Logger, port int, proto UPnPProtocol) (*UPnPMapper, error) {
	manager := &UPnPMapper{
		logger:   logrus.New(),
		port:     uint16(port),
		protoStr: upnpProtocolToString(proto),
	}
	return manager, nil
}

func (u *UPnPMapper) Start() {
	u.mux.Lock()
	defer u.mux.Unlock()

	if u.gateway != nil {
		return
	}

	d, err := upnp.Discover()
	if err != nil {
		u.logger.WithError(err).Error("UPnP discovery failed")
		return
	}
	u.gateway = d

	err = u.gateway.Forward(u.port, "Requested by MadNet", u.protoStr)
	if err != nil {
		u.logger.WithError(err).Error("UPnP forward failed")
		return
	}
}

func (u *UPnPMapper) Close() {
	u.mux.Lock()
	defer u.mux.Unlock()

	if u.gateway == nil {
		return
	}

	err := u.gateway.Clear(u.port, u.protoStr)
	if err != nil {
		u.logger.WithError(err).Error("UPnP close forward failed")
		return
	}
	u.gateway = nil
}

func upnpProtocolToString(p UPnPProtocol) string {
	switch p {
	case TCP:
		return "TCP"
	case UDP:
		return "UDP"
	default:
		return "Unknown"
	}
}
