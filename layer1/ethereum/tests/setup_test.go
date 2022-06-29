package tests

import (
	"os"
	"testing"

	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var HardHat *tests.Hardhat

func setupEthereum(t *testing.T, n int, unlockAllAccounts bool) (*tests.ClientFixture, layer1.Client, *logrus.Entry) {
	logging.GetLogger("ethereum").SetLevel(logrus.InfoLevel)
	logger := logging.GetLogger("test").WithField("test", t.Name())

	fixture := tests.NewClientFixture(HardHat, 0, n, logger, unlockAllAccounts, false, false)
	assert.NotNil(t, fixture)
	eth := fixture.Client
	assert.NotNil(t, eth)

	assert.Equal(t, n, len(eth.GetKnownAccounts()))

	t.Cleanup(func() {
		fixture.Close()
		ethereum.CleanGlobalVariables(t)
	})

	return fixture, eth, logger
}

func TestMain(m *testing.M) {
	hardhat, err := tests.StartHardHatNodeWithDefaultHost()
	if err != nil {
		panic(err)
	}
	HardHat = hardhat
	code := m.Run()
	hardhat.Close()
	os.Exit(code)
}
