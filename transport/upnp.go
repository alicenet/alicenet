package transport

import (
	"sync"

	"github.com/sirupsen/logrus"
	"gitlab.com/NebulousLabs/go-upnp"
)

type UPnPProtocol int

type UPnPMapper struct {
	// This is the logger for the transport
	logger *logrus.Logger
	// Port to map
	port uint16
	// Gateway
	gateway *upnp.IGD
	// Start / stop mutex
	mux sync.Mutex
}

func NewUPnPMapper(logger *logrus.Logger, port int) (*UPnPMapper, error) {
	manager := &UPnPMapper{
		logger: logger,
		port:   uint16(port),
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

	err = u.gateway.Forward(u.port, "Requested by AliceNet")
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

	err := u.gateway.Clear(u.port)
	if err != nil {
		u.logger.WithError(err).Error("UPnP close forward failed")
		return
	}
	u.gateway = nil
}
