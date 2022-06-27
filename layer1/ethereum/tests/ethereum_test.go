//go:build integration

package tests

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"net"
	"strings"
	"testing"

	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func setupEthereum(t *testing.T, n int, unlockAllAccounts bool) (*tests.ClientFixture, layer1.Client, *logrus.Entry) {
	logging.GetLogger("ethereum").SetLevel(logrus.InfoLevel)
	logger := logging.GetLogger("test").WithField("test", t.Name())

	fixture := tests.NewClientFixture(HardHat, 0, n, logger, unlockAllAccounts, false, false)
	assert.NotNil(t, fixture)
	eth := fixture.Client
	assert.NotNil(t, eth)

	assert.Equal(t, n, len(eth.GetKnownAccounts()))

	return fixture, eth, logger
}

func TestEthereum_SendTransactionOnlyDefaultAccountUnlocked(t *testing.T) {
	hardhat, err := tests.StartHardHatNodeWithDefaultHost()
	assert.Nil(t, err)
	defer hardhat.Close()
	unlockAllAccounts := false
	fixture, eth, logger := setupEthereum(t, 4, unlockAllAccounts)
	defer fixture.Close()

	accountList := eth.GetKnownAccounts()

	for _, acct := range accountList {
		testTxn, err := ethereum.TransferEther(eth, logger, acct.Address, eth.GetDefaultAccount().Address, big.NewInt(1))
		if bytes.Equal(acct.Address.Bytes(), eth.GetDefaultAccount().Address.Bytes()) {
			assert.Nil(t, err)
			assert.NotNil(t, testTxn)
		} else {
			assert.NotNil(t, err, "unlocked account should not be able to sign a transaction: %v", acct.Address)
			assert.Nil(t, testTxn)
		}
	}

}

func TestEthereum_SendTransactionAllAccountUnlocked(t *testing.T) {
	unlockAllAccounts := true
	fixture, eth, logger := setupEthereum(t, 4, unlockAllAccounts)
	defer fixture.Close()

	accountList := eth.GetKnownAccounts()

	for _, acct := range accountList {
		testTxn, err := ethereum.TransferEther(eth, logger, acct.Address, eth.GetDefaultAccount().Address, big.NewInt(1))
		assert.Nil(t, err)
		assert.NotNil(t, testTxn)
	}

}

func TestEthereum_HardhatNode(t *testing.T) {
	isRunning, err := HardHat.IsHardHatRunning()
	assert.Nil(t, err)
	assert.True(t, isRunning)
}

func TestEthereum_NewEthereumEndpoint(t *testing.T) {
	fixture, eth, _ := setupEthereum(t, 4, false)
	defer fixture.Close()

	type args struct {
		endpoint                 string
		pathKeystore             string
		pathPassCodes            string
		defaultAccount           string
		unlockAdditionalAccounts bool
		finalityDelay            uint64
		txMaxGasFeeAllowedInGwei uint64
		endpointMinimumPeers     uint64
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Create new ethereum endpoint failing with txMaxGasFeeAllowedInGwei too low",
			args: args{"", "", "", "", false, 0, 0, 0},
			want: false,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if !strings.Contains(err.Error(), "txMaxGasFeeAllowedInGwei should be greater than") {
					t.Errorf("Failing test with an unexpected error %v", err)
				}
				return true
			},
		},
		{
			name: "Create new ethereum endpoint failing with passCode file not found",
			args: args{"", "", "", "", false, 0, 500, 0},
			want: false,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				_, ok := err.(*fs.PathError)
				if !ok {
					t.Errorf("Failing test with an unexpected error %v", err)
				}
				return ok
			},
		},
		{
			name: "Create new ethereum endpoint failing with specified account not found",
			args: args{"", fixture.KeyStorePath, fixture.PassCodePath, "", false, 0, 500, 0},
			want: false,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if !errors.Is(err, ethereum.ErrAccountNotFound) {
					t.Errorf("Failing test with an unexpected error %v", err)
				}
				return true
			},
		},
		{
			name: "Create new ethereum endpoint failing on Dial Context",
			args: args{
				"",
				fixture.KeyStorePath,
				fixture.PassCodePath,
				eth.GetDefaultAccount().Address.Hex(),
				false,
				0,
				500,
				0,
			},
			want: false,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				_, ok := err.(*net.OpError)
				if !ok {
					t.Errorf("Failing test with an unexpected error %v", err)
				}
				return ok
			},
		},
		{
			name: "Create new ethereum endpoint returning Client struct",
			args: args{
				eth.GetEndpoint(),
				fixture.KeyStorePath,
				fixture.PassCodePath,
				eth.GetDefaultAccount().Address.Hex(),
				false,
				0,
				500,
				0,
			},
			want: true,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return true
			},
		},
		{
			name: "Create new ethereum endpoint unlocking all accounts",
			args: args{
				eth.GetEndpoint(),
				fixture.KeyStorePath,
				fixture.PassCodePath,
				eth.GetDefaultAccount().Address.Hex(),
				true,
				0,
				500,
				0,
			},
			want: true,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ethereum.NewClient(
				tt.args.endpoint,
				tt.args.pathKeystore,
				tt.args.pathPassCodes,
				tt.args.defaultAccount,
				tt.args.unlockAdditionalAccounts,
				tt.args.finalityDelay,
				tt.args.txMaxGasFeeAllowedInGwei,
				tt.args.endpointMinimumPeers,
			)
			if !tt.wantErr(
				t,
				err,
				fmt.Sprintf(
					"NewEthereumClient(%v, %v, %v, %v, %v, %v, %v, %v)",
					tt.args.endpoint,
					tt.args.pathKeystore,
					tt.args.pathPassCodes,
					tt.args.defaultAccount,
					tt.args.unlockAdditionalAccounts,
					tt.args.finalityDelay,
					tt.args.txMaxGasFeeAllowedInGwei,
					tt.args.endpointMinimumPeers,
				),
			) {
				return
			}
			if tt.want {
				assert.NotNilf(t, got, "Ethereum Details must not be nil")
			}
		})
	}
}
