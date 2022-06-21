package dkgtasks

import (
	"context"

	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/sirupsen/logrus"
)

type PlaceHolder struct {
	state *objects.DkgState
}

func NewPlaceHolder(state *objects.DkgState) *PlaceHolder {
	return &PlaceHolder{state: state}
}

func (ph *PlaceHolder) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	logger.Infof("ph dowork")
	return nil
}
func (ph *PlaceHolder) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	logger.Infof("ph dowork")
	return nil
}

func (ph *PlaceHolder) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	logger.Infof("ph doretry")
	return nil
}

func (ph *PlaceHolder) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	logger.Infof("ph shouldretry")
	return false
}

func (ph *PlaceHolder) DoDone(logger *logrus.Entry) {
	logger.Infof("ph done")
}

func (ph *PlaceHolder) GetExecutionData() interface{} {
	return nil
}
