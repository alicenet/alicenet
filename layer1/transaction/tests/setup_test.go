//go:build integration

package tests

import (
	"os"
	"testing"

	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/logging"
	"github.com/stretchr/testify/assert"
)

var HardHat *tests.Hardhat

func TestMain(m *testing.M) {
	hardhat, err := tests.StartHardHatNodeWithDefaultHost()
	if err != nil {
		panic(err)
	}
	HardHat = hardhat
	code := 1
	func() {
		defer hardhat.Close()
		code = m.Run()
	}()

	os.Exit(code)
}

func setupEthereum(t *testing.T, n int) *tests.ClientFixture {
	logger := logging.GetLogger("test").WithField("test", t.Name())
	fixture := tests.NewClientFixture(HardHat, 0, n, logger, true, true, true)
	assert.NotNil(t, fixture)

	eth := fixture.Client
	assert.NotNil(t, eth)
	assert.Equal(t, n, len(eth.GetKnownAccounts()))

	t.Cleanup(func() {
		fixture.Close()
	})

	return fixture
}
