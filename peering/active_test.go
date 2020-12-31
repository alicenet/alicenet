package peering

import (
	"context"
	"sync"
	"testing"
	"time"

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
	c1Raw := NewMockP2PClient(ctrl)
	c1cc := make(chan struct{})
	c1 := &wrappedMock{c1Raw, c1cc}
	c2Raw := NewMockP2PClient(ctrl)
	c2cc := make(chan struct{})
	c2 := &wrappedMock{c2Raw, c2cc}
	obj := activePeerStore{
		canClose:  true,
		store:     make(map[string]interfaces.P2PClient),
		pid:       make(map[string]uint64),
		closeChan: make(chan struct{}),
		closeOnce: sync.Once{},
	}
	c1na, err := transport.RandomNodeAddr()
	if err != nil {
		t.Fatal(err)
	}
	c1.EXPECT().NodeAddr().Return(c1na)
	c1.EXPECT().NodeAddr().Return(c1na)
	c1.EXPECT().NodeAddr().Return(c1na)
	c1.EXPECT().CloseChan()
	obj.add(c1)
	time.Sleep(3 * time.Second)

	c2.EXPECT().NodeAddr().Return(c1na)
	c1.EXPECT().CloseChan()
	c2.EXPECT().Close()
	obj.add(c2)
	if len(obj.store) != 1 {
		t.Fatal("not one")
	}
	if len(obj.pid) != 1 {
		t.Fatal("not one")
	}
	time.Sleep(3 * time.Second)

	c1.EXPECT().NodeAddr().Return(c1na)
	c1.EXPECT().NodeAddr().Return(c1na)
	c1.EXPECT().NodeAddr().Return(c1na)
	close(c1cc)
	time.Sleep(3 * time.Second)

	if len(obj.store) != 0 {
		t.Fatal("not zero")
	}
	if len(obj.pid) != 0 {
		t.Fatal("not zero")
	}

	// reset the close channel
	c2.closeChan = make(chan struct{})

	c2.EXPECT().NodeAddr().Return(c1na)
	c2.EXPECT().NodeAddr().Return(c1na)
	c2.EXPECT().NodeAddr().Return(c1na)
	c2.EXPECT().CloseChan()
	c2.EXPECT().NodeAddr().Return(c1na)
	obj.add(c2)
	time.Sleep(3 * time.Second)

	c2.EXPECT().Close()
	obj.del(c1na)
	time.Sleep(3 * time.Second)

	if len(obj.store) != 0 {
		t.Fatal("not zero")
	}
	if len(obj.pid) != 0 {
		t.Fatal("not zero")
	}

	// reset the close channel
	c2.closeChan = make(chan struct{})

	c2.EXPECT().NodeAddr().Return(c1na)
	c2.EXPECT().NodeAddr().Return(c1na)
	c2.EXPECT().NodeAddr().Return(c1na)
	c2.EXPECT().CloseChan()
	obj.add(c2)
	time.Sleep(3 * time.Second)

	c2.EXPECT().Close()
	obj.close()
	time.Sleep(3 * time.Second)
}
