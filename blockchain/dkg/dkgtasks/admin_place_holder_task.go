package dkgtasks

import (
	"context"
	"fmt"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/consensus/admin"
	"github.com/sirupsen/logrus"
)

type AdminPlaceHolder struct {
	state *objects.DkgState
}

func NewAdminPlaceHolder(state *objects.DkgState) *AdminPlaceHolder {
	return &AdminPlaceHolder{state: state}
}

func (ph *AdminPlaceHolder) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	logger.Infof("ph dowork")
	return nil
}
func (ph *AdminPlaceHolder) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	logger.Infof("ph dowork")
	return nil
}

func (ph *AdminPlaceHolder) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	logger.Infof("ph doretry")
	return nil
}

func (ph *AdminPlaceHolder) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	logger.Infof("ph shouldretry")
	return false
}

func (ph *AdminPlaceHolder) DoDone(logger *logrus.Entry) {
	logger.Infof("ph done")
}

func (ph *AdminPlaceHolder) SetAdminHandler(adminHandler *admin.Handlers) {
	fmt.Printf("setting admin handler: %p\n", adminHandler)
}
