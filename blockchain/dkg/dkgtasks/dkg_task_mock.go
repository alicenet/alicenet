package dkgtasks

import (
	"context"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/blockchain/tasks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

type DkgTaskMock struct {
	*tasks.Task
	mock.Mock
}

var _ interfaces.ITask = &DkgTaskMock{}

func NewDkgTaskMock(state *objects.DkgState, start uint64, end uint64) *DkgTaskMock {
	dkgTaskMock := &DkgTaskMock{}
	dkgTaskMock.Task = &tasks.Task{
		Start:   start,
		End:     end,
		State:   state,
		Success: false,
		TxOpts:  &tasks.TxOpts{},
	}

	return dkgTaskMock
}

func (d *DkgTaskMock) DoDone(logger *logrus.Entry) {
}

func (d *DkgTaskMock) DoRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	d.TxOpts.TxHashes = append(d.TxOpts.TxHashes, common.BigToHash(big.NewInt(131231214123871239)))
	if d.TxOpts.GasFeeCap == nil {
		d.TxOpts.GasFeeCap = big.NewInt(142356)
	}
	if d.TxOpts.GasTipCap == nil {
		d.TxOpts.GasTipCap = big.NewInt(37)
	}

	args := d.Called(ctx, logger, eth)
	return args.Error(0)
}

func (d *DkgTaskMock) DoWork(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	d.TxOpts.TxHashes = append(d.TxOpts.TxHashes, common.BigToHash(big.NewInt(131231214123871239)))
	if d.TxOpts.GasFeeCap == nil {
		d.TxOpts.GasFeeCap = big.NewInt(142356)
	}
	if d.TxOpts.GasTipCap == nil {
		d.TxOpts.GasTipCap = big.NewInt(37)
	}

	args := d.Called(ctx, logger, eth)
	return args.Error(0)
}

func (d *DkgTaskMock) Initialize(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	args := d.Called(ctx, logger, eth)
	return args.Error(0)
}

func (d *DkgTaskMock) ShouldRetry(ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) bool {
	args := d.Called(ctx, logger, eth)
	return args.Bool(0)
}

func (d *DkgTaskMock) GetExecutionData() interfaces.ITaskExecutionData {
	return d.Task
}
