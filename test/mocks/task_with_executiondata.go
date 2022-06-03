package mocks

import (
	"context"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	exObjects "github.com/MadBase/MadNet/blockchain/executor/objects"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/state"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

type MockITaskWithExecutionData struct {
	*MockITask
	Task *exObjects.Task
}

func NewMockITaskWithExecutionData(name string, start uint64, end uint64) *MockITaskWithExecutionData {
	task := NewMockITask()
	ed := exObjects.NewTask(state.NewDkgState(accounts.Account{}), name, start, end)
	task.GetExecutionDataFunc.SetDefaultReturn(ed)
	task.DoWorkFunc.SetDefaultHook(func(context.Context, *logrus.Entry, ethereum.Network) error {
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
