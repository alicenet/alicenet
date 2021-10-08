package lib

import "github.com/sirupsen/logrus"

type Implementation interface {
	GetAllowedAddresses() (AddressSet, error)
	UpdateAllowedAddresses(toAdd AddressSet, toDelete AddressSet) error
}

type ImplementationConstructor func(*logrus.Logger) (Implementation, error)

type CmdRunner func(c ...string) ([]byte, error)

var _ CmdRunner = RunCmd
