package tests

import (
	"os"
	"testing"

	"github.com/alicenet/alicenet/layer1/tests"
)

var HardHat *tests.Hardhat

func TestMain(m *testing.M) {
	// hardhat, err := tests.StartHardHatNodeWithDefaultHost()
	// if err != nil {
	// 	panic(err)
	// }
	// HardHat = hardhat
	code := m.Run()
	// hardhat.Close()
	os.Exit(code)
}
