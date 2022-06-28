package tests

import (
	"strings"
	"testing"

	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/monitor/events"
	"github.com/alicenet/alicenet/layer1/monitor/interfaces"
	"github.com/alicenet/alicenet/layer1/monitor/objects"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/assert"
)

func TestRegisteringETHDKGEvents(t *testing.T) {

	var em *objects.EventMap = objects.NewEventMap()
	db := mocks.NewTestDB()
	var adminHandler interfaces.AdminHandler = mocks.NewMockAdminHandler()

	events.RegisterETHDKGEvents(em, db, adminHandler, make(chan tasks.TaskRequest, 100))

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
