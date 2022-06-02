package events

import (
	"github.com/MadBase/MadNet/blockchain/executor/interfaces"
	interfaces2 "github.com/MadBase/MadNet/blockchain/monitor/interfaces"
	"github.com/MadBase/MadNet/blockchain/monitor/objects"
	"strings"
	"testing"

	"github.com/MadBase/MadNet/bridge/bindings"
	"github.com/MadBase/MadNet/test/mocks"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/assert"
)

func TestRegisteringETHDKGEvents(t *testing.T) {

	var em *objects.EventMap = objects.NewEventMap()
	db := mocks.NewTestDB()
	var adminHandler interfaces2.IAdminHandler = mocks.NewMockIAdminHandler()

	RegisterETHDKGEvents(em, db, adminHandler, make(chan interfaces.ITask), make(chan string))

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
