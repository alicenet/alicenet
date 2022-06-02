package interfaces

import (
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
)

//
// Mock implementation of interfaces.AdminHandler
//
type MockAdminHandler struct {
}

func (ah *MockAdminHandler) AddPrivateKey([]byte, constants.CurveSpec) error {
	return nil
}

func (ah *MockAdminHandler) AddSnapshot(*objs.BlockHeader, bool) error {
	return nil
}
func (ah *MockAdminHandler) AddValidatorSet(*objs.ValidatorSet) error {
	return nil
}

func (ah *MockAdminHandler) RegisterSnapshotCallback(func(*objs.BlockHeader) error) {

}

func (ah *MockAdminHandler) SetSynchronized(v bool) {

}
