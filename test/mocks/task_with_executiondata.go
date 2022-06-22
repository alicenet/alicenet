package mocks

import (
	"context"
	"math/big"

	"github.com/alicenet/alicenet/blockchain/dkg/dkgtasks"
	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/ethereum/go-ethereum/accounts"
	common "github.com/ethereum/go-ethereum/common"
	logrus "github.com/sirupsen/logrus"
)

type MockTaskWithExecutionData struct {
	*MockTask
	ExecutionData *dkgtasks.ExecutionData
}

func NewMockTaskWithExecutionData(start uint64, end uint64) *MockTaskWithExecutionData {
	task := NewMockTask()
	ed := dkgtasks.NewExecutionData(objects.NewDkgState(accounts.Account{}), start, end)
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
	return &MockTaskWithExecutionData{MockTask: task, ExecutionData: ed}
}
