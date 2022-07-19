package tests

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/stretchr/testify/assert"
)

func TestEthereum_SendTransactionOnlyDefaultAccountUnlocked(t *testing.T) {
	unlockAllAccounts := false
	_, eth, logger := setupEthereum(t, 4, unlockAllAccounts)

	accountList := eth.GetKnownAccounts()

	for _, acct := range accountList {
		testTxn, err := ethereum.TransferEther(eth, logger, acct.Address, eth.GetDefaultAccount().Address, big.NewInt(1))
		if bytes.Equal(acct.Address.Bytes(), eth.GetDefaultAccount().Address.Bytes()) {
			assert.Nil(t, err)
			assert.NotNil(t, testTxn)
		} else {
			assert.NotNil(t, err, "unlocked account should not be able to sign a transaction: %v", acct.Address)
			assert.Contains(t, err.Error(), "account either not found or not unlocked")
			assert.Nil(t, testTxn)
		}
	}

}

func TestEthereum_SendTransactionAllAccountUnlocked(t *testing.T) {
	unlockAllAccounts := true
	_, eth, logger := setupEthereum(t, 4, unlockAllAccounts)

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
				if !strings.Contains(err.Error(), "no such file or directory") {
					t.Errorf("Failing test with an unexpected error %v", err)
				}
				return true
			},
		},
		{
			name: "Create new ethereum endpoint failing with specified account not found",
			args: args{"", fixture.KeyStorePath, fixture.PassCodePath, "", false, 0, 500, 0},
			want: false,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if !strings.Contains(err.Error(), "could not find specified account") {
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
				if !strings.Contains(err.Error(), "missing address") {
					t.Errorf("Failing test with an unexpected error %v", err)
				}
				return true
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
