package mock

import (
	"encoding/json"
	"sync"

	"github.com/alicenet/alicenet/cmd/firewalld/lib"
)

type Msg struct {
	Bytes json.RawMessage
	Err   error
}
type UpdateCall struct {
	ToAdd    lib.AddressSet
	ToDelete lib.AddressSet
}
type GetRet struct {
	Ret lib.AddressSet
	Err error
}
type Calls struct {
	Update []UpdateCall
	Get    int
}
type MockImplementation struct {
	GetRet    GetRet
	UpdateRet error
	calls     Calls
	mu        sync.Mutex
}

func NewImplementation() *MockImplementation {
	return &MockImplementation{calls: Calls{Update: make([]UpdateCall, 0)}}
}
func (mi *MockImplementation) GetAllowedAddresses() (lib.AddressSet, error) {
	mi.mu.Lock()
	mi.calls.Get++
	mi.mu.Unlock()
	return mi.GetRet.Ret, mi.GetRet.Err
}
func (mi *MockImplementation) UpdateAllowedAddresses(toAdd lib.AddressSet, toDelete lib.AddressSet) error {
	mi.mu.Lock()
	mi.calls.Update = append(mi.calls.Update, UpdateCall{toAdd, toDelete})
	mi.mu.Unlock()
	return mi.UpdateRet
}
func (mi *MockImplementation) Calls() Calls {
	mi.mu.Lock()
	update := make([]UpdateCall, len(mi.calls.Update))
	copy(update, mi.calls.Update)
	mi.mu.Unlock()
	return Calls{Update: update, Get: mi.calls.Get}
}
