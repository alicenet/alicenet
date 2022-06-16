package peering

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/MadBase/MadNet/interfaces"
	pb "github.com/MadBase/MadNet/proto"
	"github.com/MadBase/MadNet/transport"
	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"
)

type wrappedMock struct {
	*MockP2PClient
	closeChan chan struct{}
}

func (wm *wrappedMock) CloseChan() <-chan struct{} {
	wm.MockP2PClient.CloseChan()
	return wm.closeChan
}

func (wm *wrappedMock) Close() error {
	close(wm.closeChan)
	wm.MockP2PClient.Close()
	return nil
}

func (wm *wrappedMock) GetSnapShotHdrNode(context.Context, *pb.GetSnapShotHdrNodeRequest, ...grpc.CallOption) (*pb.GetSnapShotHdrNodeResponse, error) {
	return nil, nil
}

func TestActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	P2PClientOne := NewMockP2PClient(ctrl)
	P2PClientTwo := NewMockP2PClient(ctrl)

	P2PClientOneChannel := make(chan struct{})
	P2PClientTwoChannel := make(chan struct{})

	clientOne := &wrappedMock{P2PClientOne, P2PClientOneChannel}
	clientTwo := &wrappedMock{P2PClientTwo, P2PClientTwoChannel}

	activePeerStoreObj := activePeerStore{
		canClose:  true,
		store:     make(map[string]interfaces.P2PClient),
		pid:       make(map[string]uint64),
		closeChan: make(chan struct{}),
		closeOnce: sync.Once{},
	}
	randomNodeAddr, err := transport.RandomNodeAddr()
	if err != nil {
		t.Fatal(err)
	}
	clientOne.EXPECT().NodeAddr().Return(randomNodeAddr)
	clientOne.EXPECT().NodeAddr().Return(randomNodeAddr)
	clientOne.EXPECT().NodeAddr().Return(randomNodeAddr)
	clientOne.EXPECT().CloseChan()
	activePeerStoreObj.add(clientOne)
	time.Sleep(3 * time.Second)

	clientTwo.EXPECT().NodeAddr().Return(randomNodeAddr)
	clientOne.EXPECT().CloseChan()
	clientTwo.EXPECT().Close()
	activePeerStoreObj.add(clientTwo)
	if len(activePeerStoreObj.store) != 1 {
		t.Fatal("not one")
	}
	if len(activePeerStoreObj.pid) != 1 {
		t.Fatal("not one")
	}
	time.Sleep(3 * time.Second)

	clientOne.EXPECT().NodeAddr().Return(randomNodeAddr)
	clientOne.EXPECT().NodeAddr().Return(randomNodeAddr)
	clientOne.EXPECT().NodeAddr().Return(randomNodeAddr)
	close(P2PClientOneChannel)
	time.Sleep(3 * time.Second)

	activePeerStoreObj.RLock()
	if len(activePeerStoreObj.store) != 0 {
		t.Fatal("not zero")
	}
	if len(activePeerStoreObj.pid) != 0 {
		t.Fatal("not zero")
	}
	activePeerStoreObj.RUnlock()

	// reset the close channel
	clientTwo.closeChan = make(chan struct{})

	clientTwo.EXPECT().NodeAddr().Return(randomNodeAddr)
	clientTwo.EXPECT().NodeAddr().Return(randomNodeAddr)
	clientTwo.EXPECT().NodeAddr().Return(randomNodeAddr)
	clientTwo.EXPECT().CloseChan()
	clientTwo.EXPECT().NodeAddr().Return(randomNodeAddr)
	activePeerStoreObj.add(clientTwo)
	time.Sleep(3 * time.Second)

	clientTwo.EXPECT().Close()
	activePeerStoreObj.del(randomNodeAddr)
	time.Sleep(3 * time.Second)

	activePeerStoreObj.RLock()
	if len(activePeerStoreObj.store) != 0 {
		t.Fatal("not zero")
	}
	if len(activePeerStoreObj.pid) != 0 {
		t.Fatal("not zero")
	}
	activePeerStoreObj.RUnlock()

	// reset the close channel
	clientTwo.closeChan = make(chan struct{})

	clientTwo.EXPECT().NodeAddr().Return(randomNodeAddr)
	clientTwo.EXPECT().NodeAddr().Return(randomNodeAddr)
	clientTwo.EXPECT().NodeAddr().Return(randomNodeAddr)
	clientTwo.EXPECT().CloseChan()
	activePeerStoreObj.add(clientTwo)
	time.Sleep(3 * time.Second)

	clientTwo.EXPECT().Close()
	activePeerStoreObj.close()
	time.Sleep(3 * time.Second)
}

func Test_activePeerStore_add(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	P2PClientOne := NewMockP2PClient(ctrl)
	P2PClientOneChannel := make(chan struct{})
	clientOne := &wrappedMock{
		P2PClientOne,
		P2PClientOneChannel,
	}
	randomNodeAddr, err := transport.RandomNodeAddr()
	if err != nil {
		t.Fatal(err)
	}
	clientOne.EXPECT().NodeAddr().Return(randomNodeAddr).Times(3)
	clientOne.EXPECT().CloseChan()

	type args struct {
		c interfaces.P2PClient
	}
	var tests = []struct {
		name string
		args args
	}{
		{
			name: "Adding client to active peer store",
			args: args{
				c: clientOne,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &activePeerStore{
				canClose:  true,
				store:     make(map[string]interfaces.P2PClient),
				pid:       make(map[string]uint64),
				closeChan: make(chan struct{}),
			}
			ps.add(tt.args.c)
			assert.Equal(t, 1, ps.len())
		})
	}
	time.Sleep(3 * time.Second)
}

func Test_activePeerStore_close(t *testing.T) {
	closeChan := make(chan struct{})
	ps := &activePeerStore{
		canClose:  true,
		store:     make(map[string]interfaces.P2PClient),
		pid:       make(map[string]uint64),
		closeChan: closeChan,
	}
	ps.close()
	_, isOpen := <-closeChan
	assert.Equal(t, false, isOpen)
}

func Test_activePeerStore_contains(t *testing.T) {
	randomNodeAddr, err := transport.RandomNodeAddr()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		c interfaces.NodeAddr
	}
	tests := []struct {
		name        string
		args        args
		prePopulate bool
		want        bool
	}{
		{
			name:        "Test active peer store contains identity",
			args:        struct{ c interfaces.NodeAddr }{c: randomNodeAddr},
			prePopulate: true,
			want:        true,
		},
		{
			name:        "Test active peer store contains identity with empty active peer store",
			args:        struct{ c interfaces.NodeAddr }{c: randomNodeAddr},
			prePopulate: false,
			want:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &activePeerStore{
				canClose:  true,
				store:     make(map[string]interfaces.P2PClient),
				pid:       make(map[string]uint64),
				closeChan: make(chan struct{}),
			}
			if tt.prePopulate {
				ps.store[tt.args.c.Identity()] = nil
			}
			if got := ps.contains(tt.args.c); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_activePeerStore_del(t *testing.T) {
	randomNodeAddr, err := transport.RandomNodeAddr()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		c interfaces.NodeAddr
	}

	tests := []struct {
		name        string
		args        args
		prePopulate bool
	}{
		{
			name:        "Test delete peer from store",
			args:        struct{ c interfaces.NodeAddr }{c: randomNodeAddr},
			prePopulate: true,
		},
		{
			name:        "Test delete peer from store with empty store",
			args:        struct{ c interfaces.NodeAddr }{c: randomNodeAddr},
			prePopulate: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &activePeerStore{
				canClose:  false,
				store:     make(map[string]interfaces.P2PClient),
				pid:       make(map[string]uint64),
				closeChan: make(chan struct{}),
				closeOnce: sync.Once{},
			}
			if tt.prePopulate {
				ps.store[tt.args.c.Identity()] = nil
				assert.Equal(t, 1, len(ps.store))
			} else {
				assert.Equal(t, 0, len(ps.store))
			}
			ps.del(tt.args.c)
			_, ok := ps.store[randomNodeAddr.Identity()]
			assert.False(t, ok)
			assert.Equal(t, 0, len(ps.store))
		})
	}
}

func Test_activePeerStore_getPeers(t *testing.T) {
	randomNodeAddrFirst, _ := transport.RandomNodeAddr()
	randomNodeAddrSecond, _ := transport.RandomNodeAddr()

	type args struct {
		addresses []string
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 bool
	}{
		{
			name: "Testing get Peers with 2 peers",
			args: struct{ addresses []string }{addresses: []string{
				randomNodeAddrFirst.Identity(),
				randomNodeAddrSecond.Identity(),
			}},
			want:  2,
			want1: true,
		},
		{
			name:  "Testing get Peers with no peers",
			args:  struct{ addresses []string }{addresses: []string{}},
			want:  0,
			want1: false,
		},
	}
	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		P2PClient := NewMockP2PClient(ctrl)
		P2PClientChannel := make(chan struct{})

		t.Run(tt.name, func(t *testing.T) {
			ps := &activePeerStore{
				canClose:  false,
				store:     make(map[string]interfaces.P2PClient),
				pid:       make(map[string]uint64),
				closeChan: make(chan struct{}),
			}
			for _, identity := range tt.args.addresses {
				client := &wrappedMock{P2PClient, P2PClientChannel}
				ps.store[identity] = client
			}
			p2pClients, booleanVal := ps.getPeers()
			if !reflect.DeepEqual(len(p2pClients), tt.want) {
				t.Errorf("getPeers() got = %v, want %v", p2pClients, tt.want)
			}
			if booleanVal != tt.want1 {
				t.Errorf("getPeers() got1 = %v, want %v", booleanVal, tt.want1)
			}
		})
	}
}

func Test_activePeerStore_random(t *testing.T) {
	randomNodeAddrFirst, _ := transport.RandomNodeAddr()

	type args struct {
		addresses []string
	}

	tests := []struct {
		name      string
		args      args
		populated bool
		want      string
		want1     bool
	}{
		{
			name:  "Testing randomness with no peers",
			args:  struct{ addresses []string }{addresses: []string{}},
			want:  "",
			want1: false,
		},
		{
			name: "Testing randomness with peers",
			args: struct{ addresses []string }{addresses: []string{
				randomNodeAddrFirst.Identity(),
			}},
			populated: true,
			want:      "",
			want1:     true,
		},
	}
	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		P2PClient := NewMockP2PClient(ctrl)
		P2PClientChannel := make(chan struct{})

		t.Run(tt.name, func(t *testing.T) {
			ps := &activePeerStore{
				canClose:  false,
				store:     make(map[string]interfaces.P2PClient),
				pid:       make(map[string]uint64),
				closeChan: make(chan struct{}),
			}

			if tt.populated {
				client := &wrappedMock{P2PClient, P2PClientChannel}
				ps.store[tt.args.addresses[0]] = client
				gomock.InOrder(
					client.EXPECT().NodeAddr().Return(randomNodeAddrFirst),
				)
			}
			nodeAddress, booleanVal := ps.random()
			if nodeAddress != tt.want && !tt.populated {
				t.Errorf("random() got = %v, want %v", nodeAddress, tt.want)
			}
			if tt.populated && nodeAddress == "" {
				t.Errorf("random() got = %v, want %v", nodeAddress, tt.want)
			}
			if booleanVal != tt.want1 {
				t.Errorf("random() got1 = %v, want %v", booleanVal, tt.want1)
			}
		})
	}
}

func Test_activePeerStore_randomClient(t *testing.T) {
	randomNodeAddrFirst, _ := transport.RandomNodeAddr()
	randomNodeAddrSecond, _ := transport.RandomNodeAddr()

	type args struct {
		addresses []string
	}
	tests := []struct {
		name  string
		args  args
		want  interfaces.P2PClient
		want1 bool
	}{
		{
			name:  "Testing random client with no nodes",
			want:  nil,
			want1: false,
		},
		{
			name: "Testing random client with nodes",
			args: struct{ addresses []string }{addresses: []string{
				randomNodeAddrFirst.Identity(),
				randomNodeAddrSecond.Identity(),
			}},
			want:  nil,
			want1: true,
		},
	}
	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		P2PClient := NewMockP2PClient(ctrl)
		P2PClientChannel := make(chan struct{})

		t.Run(tt.name, func(t *testing.T) {
			ps := &activePeerStore{
				// RWMutex:   tt.fields.RWMutex,
				canClose:  false,
				store:     make(map[string]interfaces.P2PClient),
				pid:       make(map[string]uint64),
				closeChan: make(chan struct{}),
				closeOnce: sync.Once{},
			}
			for _, identity := range tt.args.addresses {
				client := &wrappedMock{P2PClient, P2PClientChannel}
				ps.store[identity] = client
			}
			got, got1 := ps.randomClient()
			if len(tt.args.addresses) > 0 {
				assert.NotNilf(t, got, "Should return a client")
				assert.Truef(t, got1, "Should return true")
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("randomClient() got = %v, want %v", got, tt.want)
				}
				if got1 != tt.want1 {
					t.Errorf("randomClient() got1 = %v, want %v", got1, tt.want1)
				}
			}
		})
	}
}
