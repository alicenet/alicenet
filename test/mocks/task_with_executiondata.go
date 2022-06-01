package mocks

import (
	"context"
	"github.com/MadBase/MadNet/blockchain/tasks/dkg/objects"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/ethereum/go-ethereum/accounts"
	common "github.com/ethereum/go-ethereum/common"
	logrus "github.com/sirupsen/logrus"
)

type MockITaskWithExecutionData struct {
	*MockITask
	Task *objects.Task
}

func NewMockITaskWithExecutionData(start uint64, end uint64) *MockITaskWithExecutionData {
	task := NewMockITask()
	ed := objects.NewTask(objects.NewDkgState(accounts.Account{}), start, end)
	task.GetExecutionDataFunc.SetDefaultReturn(ed)
	task.DoWorkFunc.SetDefaultHook(func(context.Context, *logrus.Entry, interfaces.Ethereum) error {
		ed.TxOpts.TxHashes = append(ed.TxOpts.TxHashes, common.BigToHash(big.NewInt(131231214123871239)))
		if ed.TxOpts.GasFeeCap == nil {
			ed.TxOpts.GasFeeCap = big.NewInt(142356)
		}
		if ed.TxOpts.GasTipCap == nil {
			ed.TxOpts.GasTipCap = big.NewInt(37)
		}
		return nil
	})
	task.DoRetryFunc.SetDefaultHook(task.DoWork)
	return &MockITaskWithExecutionData{MockITask: task, Task: ed}
}
