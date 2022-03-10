package dkgtasks

import (
	"context"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

type DkgTaskMock struct {
	*DkgTask
	mock.Mock
}

func NewDkgTaskMock(state *objects.DkgState, start uint64, end uint64) *DkgTaskMock {
	dkgTaskMock := &DkgTaskMock{}
	dkgTaskMock.DkgTask = &DkgTask{
		Start:   start,
		End:     end,
		State:   state,
		Success: false,
	}

	return dkgTaskMock
}

func (d *DkgTaskMock) DoDone(logger *logrus.Entry) {
}

func (d *DkgTaskMock) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	args := d.Called(ctx, logger, eth)
	return args.Error(0)
}

func (d *DkgTaskMock) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	args := d.Called(ctx, logger, eth)
	return args.Error(0)
}

func (d *DkgTaskMock) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum, state interface{}) error {
	args := d.Called(ctx, logger, eth, state)
	return args.Error(0)
}

func (d *DkgTaskMock) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	args := d.Called(ctx, logger, eth)
	return args.Bool(0)
}

func (d *DkgTaskMock) GetDkgTask() *DkgTask {
	return d.DkgTask
}

func (d *DkgTaskMock) SetDkgTask(dkgTask *DkgTask) {
	d.DkgTask = dkgTask
}
