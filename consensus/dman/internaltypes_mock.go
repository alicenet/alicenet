package dman

import (
	"context"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type ReqBusViewMock struct {
	mock.Mock
}

var _ reqBusView = &ReqBusViewMock{}

func (r *ReqBusViewMock) RequestP2PGetPendingTx(ctx context.Context, txHashes [][]byte, opts ...grpc.CallOption) ([][]byte, error) {
	args := r.Called(ctx, txHashes, opts)
	return args.Get(0).([][]byte), args.Error(1)
}
func (r *ReqBusViewMock) RequestP2PGetMinedTxs(ctx context.Context, txHashes [][]byte, opts ...grpc.CallOption) ([][]byte, error) {
	args := r.Called(ctx, txHashes, opts)
	return args.Get(0).([][]byte), args.Error(1)
}
func (r *ReqBusViewMock) RequestP2PGetBlockHeaders(ctx context.Context, blockNums []uint32, opts ...grpc.CallOption) ([]*objs.BlockHeader, error) {
	args := r.Called(ctx, blockNums, opts)
	return args.Get(0).([]*objs.BlockHeader), args.Error(1)
}
