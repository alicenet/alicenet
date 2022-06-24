package localrpc

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	consensusObjs "github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
)

func TestClient_GetBlockHeader(t *testing.T) {
	type args struct {
		ctx    context.Context
		height uint32
	}
	tests := []struct {
		name    string
		args    args
		want    *consensusObjs.BlockHeader
		wantErr bool
	}{
		{args: args{
			ctx:    context.Background(),
			height: 1,
		},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lrpc.GetBlockHeader(tt.args.ctx, tt.args.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlockHeader() error = %v, wantErr %v \n", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.BClaims.Height, tt.args.height) {
				t.Errorf("GetBlockHeader() got = %v, want %v \n", got, tt.want)
			}
		})
	}
}

/* func TestClient_GetBlockHeightForTx(t *testing.T) {
	type fields struct {
		Mutex       sync.Mutex
		closeChan   chan struct{}
		closeOnce   sync.Once
		Address     string
		TimeOut     time.Duration
		conn        *grpc.ClientConn
		client      proto.LocalStateClient
		wg          sync.WaitGroup
		isConnected bool
	}
	type args struct {
		ctx    context.Context
		txHash []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint32
		wantErr bool
	}{
		{name: constants.LoggerApp,
			fields: fields{
				Mutex:       sync.Mutex{},
				closeChan:   nil,
				closeOnce:   sync.Once{},
				Address:     address,
				TimeOut:     timeout,
				conn:        nil,
				client:      nil,
				wg:          sync.WaitGroup{},
				isConnected: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lrpc := &Client{
				Mutex:       tt.fields.Mutex,
				closeChan:   tt.fields.closeChan,
				closeOnce:   tt.fields.closeOnce,
				Address:     tt.fields.Address,
				TimeOut:     tt.fields.TimeOut,
				conn:        tt.fields.conn,
				client:      tt.fields.client,
				wg:          tt.fields.wg,
				isConnected: tt.fields.isConnected,
			}
			if err := lrpc.Connect(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Printf("Connect() error = %v, wantErr %v %v \n", err, tt.wantErr)
			}
			got, err := lrpc.GetBlockHeightForTx(tt.args.ctx, tt.args.txHash)
			if (err != nil) != tt.wantErr {
				t.Printf("GetBlockHeightForTx() error = %v, wantErr %v %v \n", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Printf("GetBlockHeightForTx() got = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
func TestClient_GetBlockNumber(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    uint32
		wantErr bool
	}{
		{name: constants.LoggerApp,
			args: args{
				ctx: context.Background(),
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lrpc.GetBlockNumber(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlockNumber() error = %v, wantErr %v \n", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetBlockNumber() got = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
func TestClient_GetData(t *testing.T) {
    type fields struct {
        Mutex       sync.Mutex
        closeChan   chan struct{}
        closeOnce   sync.Once
        Address     string
        TimeOut     time.Duration
        conn        *grpc.ClientConn
        client      proto.LocalStateClient
        wg          sync.WaitGroup
        isConnected bool
    }
    type args struct {
        ctx       context.Context
        curveSpec constants.CurveSpec
        account   []byte
        index     []byte
    }
    tests := []struct {
        name    string
        fields  fields
        args    args
        want    []byte
        wantErr bool
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            lrpc := &Client{
                Mutex:       tt.fields.Mutex,
                closeChan:   tt.fields.closeChan,
                closeOnce:   tt.fields.closeOnce,
                Address:     tt.fields.Address,
                TimeOut:     tt.fields.TimeOut,
                conn:        tt.fields.conn,
                client:      tt.fields.client,
                wg:          tt.fields.wg,
                isConnected: tt.fields.isConnected,
            }
            got, err := lrpc.GetData(tt.args.ctx, tt.args.curveSpec, tt.args.account, tt.args.index)
            if (err != nil) != tt.wantErr {
                t.Printf("GetData() error = %v, wantErr %v %v \n", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Printf("GetData() got = %v, want %v", got, tt.want)
            }
        })
    }
}

*/

func TestClient_GetEpochNumber(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    uint32
		wantErr bool
	}{
		{name: constants.LoggerApp,
			args: args{
				ctx: context.Background(),
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lrpc.GetEpochNumber(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEpochNumber() error = %v, wantErr %v \n", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetEpochNumber() got = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
func TestClient_GetMinedTransaction(t *testing.T) {
    type fields struct {
        Mutex       sync.Mutex
        closeChan   chan struct{}
        closeOnce   sync.Once
        Address     string
        TimeOut     time.Duration
        conn        *grpc.ClientConn
        client      proto.LocalStateClient
        wg          sync.WaitGroup
        isConnected bool
    }
    type args struct {
        ctx    context.Context
        txHash []byte
    }
    tests := []struct {
        name    string
        fields  fields
        args    args
        want    *aobjs.Tx
        wantErr bool
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            lrpc := &Client{
                Mutex:       tt.fields.Mutex,
                closeChan:   tt.fields.closeChan,
                closeOnce:   tt.fields.closeOnce,
                Address:     tt.fields.Address,
                TimeOut:     tt.fields.TimeOut,
                conn:        tt.fields.conn,
                client:      tt.fields.client,
                wg:          tt.fields.wg,
                isConnected: tt.fields.isConnected,
            }
            got, err := lrpc.GetMinedTransaction(tt.args.ctx, tt.args.txHash)
            if (err != nil) != tt.wantErr {
                t.Printf("GetMinedTransaction() error = %v, wantErr %v %v \n", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Printf("GetMinedTransaction() got = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestClient_GetPendingTransaction(t *testing.T) {
    type fields struct {
        Mutex       sync.Mutex
        closeChan   chan struct{}
        closeOnce   sync.Once
        Address     string
        TimeOut     time.Duration
        conn        *grpc.ClientConn
        client      proto.LocalStateClient
        wg          sync.WaitGroup
        isConnected bool
    }
    type args struct {
        ctx    context.Context
        txHash []byte
    }
    tests := []struct {
        name    string
        fields  fields
        args    args
        want    *aobjs.Tx
        wantErr bool
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            lrpc := &Client{
                Mutex:       tt.fields.Mutex,
                closeChan:   tt.fields.closeChan,
                closeOnce:   tt.fields.closeOnce,
                Address:     tt.fields.Address,
                TimeOut:     tt.fields.TimeOut,
                conn:        tt.fields.conn,
                client:      tt.fields.client,
                wg:          tt.fields.wg,
                isConnected: tt.fields.isConnected,
            }
            got, err := lrpc.GetPendingTransaction(tt.args.ctx, tt.args.txHash)
            if (err != nil) != tt.wantErr {
                t.Printf("GetPendingTransaction() error = %v, wantErr %v %v \n", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Printf("GetPendingTransaction() got = %v, want %v", got, tt.want)
            }
        })
    }
}
*/
func TestClient_GetTxFees(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{name: constants.LoggerApp,
			args: args{
				ctx: context.Background(),
			},
			want: []string{
				fmt.Sprintf("%064d", 4),
				fmt.Sprintf("%064d", 1),
				fmt.Sprintf("%064d", 3),
				fmt.Sprintf("%064d", 2),
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lrpc.GetTxFees(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTxFees() error = %v, wantErr %v \n", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTxFees() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetUTXO(t *testing.T) {
	type args struct {
		ctx     context.Context
		utxoIDs [][]byte
	}
	tests := []struct {
		name    string
		args    args
		want    [][]byte
		wantErr bool
	}{
		{name: constants.LoggerApp,
			args: args{
				ctx:     context.Background(),
				utxoIDs: utxoTx1IDs,
			},
			want: utxoTx1IDs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lrpc.GetUTXO(tt.args.ctx, tt.args.utxoIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUTXO() error = %v, wantErr %v \n", err, tt.wantErr)
				return
			}
			gotUtxoId, err1 := got.UTXOID()
			if err1 != nil {
				t.Errorf("GetUTXO() Could not get UtxoIds %v \n", err)
				return
			}
			if !reflect.DeepEqual(gotUtxoId, tt.want) {
				t.Errorf("GetUTXO() got = %x, want %x \n", gotUtxoId, tt.want)
			}
		})
	}
}

/*
func TestClient_GetValueForOwner(t *testing.T) {
	type args struct {
		ctx       context.Context
		curveSpec constants.CurveSpec
		account   []byte
		minValue  *uint256.Uint256
	}
	tests := []struct {
		name    string
		args    args
		want    [][]byte
		want1   *uint256.Uint256
		wantErr bool
	}{
		{name: constants.LoggerApp,
			args: args{
				ctx:       context.Background(),
				curveSpec: 1,
				account:   account,
				minValue:  uint256.Zero(),
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := lrpc.GetValueForOwner(tt.args.ctx, tt.args.curveSpec, tt.args.account, tt.args.minValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValueForOwner() error = %v, wantErr %v \n", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetValueForOwner() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetValueForOwner() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}


func TestClient_PaginateDataStoreUTXOByOwner(t *testing.T) {
    type fields struct {
        Mutex       sync.Mutex
        closeChan   chan struct{}
        closeOnce   sync.Once
        Address     string
        TimeOut     time.Duration
        conn        *grpc.ClientConn
        client      proto.LocalStateClient
        wg          sync.WaitGroup
        isConnected bool
    }
    type args struct {
        ctx        context.Context
        curveSpec  constants.CurveSpec
        account    []byte
        num        uint8
        startIndex []byte
    }
    tests := []struct {
        name    string
        fields  fields
        args    args
        want    []*aobjs.PaginationResponse
        wantErr bool
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            lrpc := &Client{
                Mutex:       tt.fields.Mutex,
                closeChan:   tt.fields.closeChan,
                closeOnce:   tt.fields.closeOnce,
                Address:     tt.fields.Address,
                TimeOut:     tt.fields.TimeOut,
                conn:        tt.fields.conn,
                client:      tt.fields.client,
                wg:          tt.fields.wg,
                isConnected: tt.fields.isConnected,
            }
            got, err := lrpc.PaginateDataStoreUTXOByOwner(tt.args.ctx, tt.args.curveSpec, tt.args.account, tt.args.num, tt.args.startIndex)
            if (err != nil) != tt.wantErr {
                t.Printf("PaginateDataStoreUTXOByOwner() error = %v, wantErr %v %v \n", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Printf("PaginateDataStoreUTXOByOwner() got = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestClient_SendTransaction(t *testing.T) {
    type fields struct {
        Mutex       sync.Mutex
        closeChan   chan struct{}
        closeOnce   sync.Once
        Address     string
        TimeOut     time.Duration
        conn        *grpc.ClientConn
        client      proto.LocalStateClient
        wg          sync.WaitGroup
        isConnected bool
    }
    type args struct {
        ctx context.Context
        tx  *aobjs.Tx
    }
    tests := []struct {
        name    string
        fields  fields
        args    args
        want    []byte
        wantErr bool
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            lrpc := &Client{
                Mutex:       tt.fields.Mutex,
                closeChan:   tt.fields.closeChan,
                closeOnce:   tt.fields.closeOnce,
                Address:     tt.fields.Address,
                TimeOut:     tt.fields.TimeOut,
                conn:        tt.fields.conn,
                client:      tt.fields.client,
                wg:          tt.fields.wg,
                isConnected: tt.fields.isConnected,
            }
            got, err := lrpc.SendTransaction(tt.args.ctx, tt.args.tx)
            if (err != nil) != tt.wantErr {
                t.Printf("SendTransaction() error = %v, wantErr %v %v \n", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Printf("SendTransaction() got = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestClient_contextGuard(t *testing.T) {
    type fields struct {
        Mutex       sync.Mutex
        closeChan   chan struct{}
        closeOnce   sync.Once
        Address     string
        TimeOut     time.Duration
        conn        *grpc.ClientConn
        client      proto.LocalStateClient
        wg          sync.WaitGroup
        isConnected bool
    }
    type args struct {
        ctx context.Context
    }
    tests := []struct {
        name   string
        fields fields
        args   args
        want   context.Context
        want1  func()
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            lrpc := &Client{
                Mutex:       tt.fields.Mutex,
                closeChan:   tt.fields.closeChan,
                closeOnce:   tt.fields.closeOnce,
                Address:     tt.fields.Address,
                TimeOut:     tt.fields.TimeOut,
                conn:        tt.fields.conn,
                client:      tt.fields.client,
                wg:          tt.fields.wg,
                isConnected: tt.fields.isConnected,
            }
            got, got1 := lrpc.contextGuard(tt.args.ctx)
            if !reflect.DeepEqual(got, tt.want) {
                t.Printf("contextGuard() got = %v, want %v", got, tt.want)
            }
            if !reflect.DeepEqual(got1, tt.want1) {
                t.Printf("contextGuard() got1 = %v, want %v", got1, tt.want1)
            }
        })
    }
}

func TestClient_entrancyGuard(t *testing.T) {
    type fields struct {
        Mutex       sync.Mutex
        closeChan   chan struct{}
        closeOnce   sync.Once
        Address     string
        TimeOut     time.Duration
        conn        *grpc.ClientConn
        client      proto.LocalStateClient
        wg          sync.WaitGroup
        isConnected bool
    }
    tests := []struct {
        name    string
        fields  fields
        wantErr bool
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            lrpc := &Client{
                Mutex:       tt.fields.Mutex,
                closeChan:   tt.fields.closeChan,
                closeOnce:   tt.fields.closeOnce,
                Address:     tt.fields.Address,
                TimeOut:     tt.fields.TimeOut,
                conn:        tt.fields.conn,
                client:      tt.fields.client,
                wg:          tt.fields.wg,
                isConnected: tt.fields.isConnected,
            }
            if err := lrpc.entrancyGuard(); (err != nil) != tt.wantErr {
                t.Printf("entrancyGuard() error = %v, wantErr %v %v \n", err, tt.wantErr)
            }
        })
    }
}
*/
