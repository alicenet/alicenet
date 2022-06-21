package monitor_test

import (
	"strings"
	"testing"

	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/blockchain/monitor"
	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/assert"
)

func TestRegisteringETHDKGEvents(t *testing.T) {

	var em *objects.EventMap = objects.NewEventMap()
	var adminHandler interfaces.AdminHandler = mocks.NewMockAdminHandler()

	monitor.RegisterETHDKGEvents(em, adminHandler)

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
