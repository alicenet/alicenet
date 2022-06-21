//go:build integration

package monitor_test

import (
	"context"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/blockchain"
	"github.com/alicenet/alicenet/blockchain/dkg/dtest"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationOpenEvent(t *testing.T) {

	privateKeys, _ := dtest.InitializePrivateKeysAndAccounts(4)
	eth, err := blockchain.NewEthereumSimulator(
		privateKeys,
		6,
		10*time.Second,
		30*time.Second,
		0,
		big.NewInt(math.MaxInt64),
		50,
		math.MaxInt64,
		5*time.Second,
		30*time.Second)
	defer eth.Close()

	assert.Nil(t, err, "Error creating Ethereum simulator")
	c := eth.Contracts()
	assert.NotNil(t, c, "Need a *Contracts")

	err = dtest.StartHardHatNode(eth)
	if err != nil {
		t.Fatalf("error starting hardhat node: %v", err)
	}

	t.Logf("waiting on hardhat node to start...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = dtest.WaitForHardHatNode(ctx)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	height, err := eth.GetCurrentHeight(context.TODO())
	assert.Nil(t, err, "could not get height")
	assert.Equal(t, uint64(0), height, "Height should be 0")

	eth.Commit()

	height, err = eth.GetCurrentHeight(context.TODO())
	assert.Nil(t, err, "could not get height")
	assert.Equal(t, uint64(1), height, "Height should be 1")
}
