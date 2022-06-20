package events

import (
	"strings"
	"testing"

	"github.com/MadBase/MadNet/bridge/bindings"
	"github.com/MadBase/MadNet/layer1/executor/interfaces"
	interfaces2 "github.com/MadBase/MadNet/layer1/monitor/interfaces"
	"github.com/MadBase/MadNet/layer1/monitor/objects"
	"github.com/MadBase/MadNet/test/mocks"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/assert"
)

func TestRegisteringETHDKGEvents(t *testing.T) {

	var em *objects.EventMap = objects.NewEventMap()
	db := mocks.NewTestDB()
	var adminHandler interfaces2.AdminHandler = mocks.NewMockIAdminHandler()

	RegisterETHDKGEvents(em, db, adminHandler, make(chan interfaces.Task), make(chan string))

	ethDkgABI, err := abi.JSON(strings.NewReader(bindings.ETHDKGMetaData.ABI))
	if err != nil {
		t.Fatal(err)
	}

	for name, event := range ethDkgABI.Events {
		eventInfo, ok := em.Lookup(event.ID.String())
		assert.True(t, ok)
		assert.Equal(t, name, eventInfo.Name)
	}
}
