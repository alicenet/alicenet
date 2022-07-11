//go:build integration

package tests

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func setupEthereum(t *testing.T, n int) layer1.Client {
	logging.GetLogger("ethereum").SetLevel(logrus.InfoLevel)

	ecdsaPrivateKeys, _ := tests.InitializePrivateKeysAndAccounts(n)
	eth := tests.ConnectSimulatorEndpoint(t, ecdsaPrivateKeys, 1000*time.Millisecond)
	assert.NotNil(t, eth)

	acct := eth.GetDefaultAccount()
	assert.Nil(t, eth.UnlockAccount(acct))

	return eth
}

func TestEthereum_AccountsFound(t *testing.T) {
	eth := setupEthereum(t, 4)
	defer eth.Close()

	accountList := eth.GetKnownAccounts()

	for _, acct := range accountList {

		err := eth.UnlockAccount(acct)
		assert.Nilf(t, err, "Not able to unlock account: %v", acct.Address)

		_, err = eth.GetAccountKeys(acct.Address)
		assert.Nilf(t, err, "Not able to get keys for account: %v", acct.Address)
	}

}

func TestEthereum_HardhatNode(t *testing.T) {
	privateKeys, _ := tests.InitializePrivateKeysAndAccounts(4)

	eth, err := ethereum.NewSimulator(
		privateKeys,
		6,
		10*time.Second,
		30*time.Second,
		0,
		big.NewInt(math.MaxInt64),
		50,
		math.MaxInt64)
	defer func() {
		err := eth.Close()
		if err != nil {
			t.Fatalf("error closing eth: %v", err)
		}
	}()

	assert.Nil(t, err, "Failed to build Ethereum endpoint...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Unlock the default account and use it to deploy contracts
	deployAccount := eth.GetDefaultAccount()
	err = eth.UnlockAccount(deployAccount)
	assert.Nil(t, err, "Failed to unlock default account")

	t.Logf("deploy account: %v", deployAccount.Address.String())

	err = tests.StartHardHatNode(eth)
	if err != nil {
		t.Fatalf("error starting hardhat node: %v", err)
	}

	t.Logf("waiting on hardhat node to start...")

	err = tests.WaitForHardHatNode(ctx)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	t.Logf("done testing")
}

func TestEthereum_NewEthereumEndpoint(t *testing.T) {

	eth := setupEthereum(t, 4)
	defer eth.Close()

	type args struct {
		endpoint                    string
		pathKeystore                string
		pathPasscodes               string
		defaultAccount              string
		timeout                     time.Duration
		retryCount                  int
		retryDelay                  time.Duration
		finalityDelay               int
		txFeePercentageToIncrease   int
		getTxMaxGasFeeAllowedInGwei uint64
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{

		{
			name: "Create new ethereum endpoint failing with passcode file not found",
			args: args{"", "", "", "", 0, 0, 0, 0, 0, 0},
			want: false,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				_, ok := err.(*fs.PathError)
				if !ok {
					t.Errorf("Failing test with an unexpected error")
				}
				return ok
			},
		},
		{
			name: "Create new ethereum endpoint failing with specified account not found",
			args: args{"", "", "../assets/test/passcodes.txt", "", 0, 0, 0, 0, 0, 0},
			want: false,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				if !errors.Is(err, ethereum.ErrAccountNotFound) {
					t.Errorf("Failing test with an unexpected error")
				}
				return true
			},
		},
		{
			name: "Create new ethereum endpoint failing on Dial Context",
			args: args{
				eth.GetEndpoint(),
				"../assets/test/keys",
				"../assets/test/passcodes.txt",
				eth.GetDefaultAccount().Address.String(),
				eth.Timeout(),
				eth.RetryCount(),
				eth.RetryDelay(),
				int(eth.GetFinalityDelay()),
				eth.GetTxFeePercentageToIncrease(),
				eth.GetTxMaxGasFeeAllowedInGwei(),
			},
			want: false,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				_, ok := err.(*net.OpError)
				if !ok {
					t.Errorf("Failing test with an unexpected error")
				}
				return ok
			},
		},
		{
			name: "Create new ethereum endpoint returning EthereumDetails",
			args: args{
				"http://localhost:8545",
				"../assets/test/keys",
				"../assets/test/passcodes.txt",
				eth.GetDefaultAccount().Address.String(),
				eth.Timeout(),
				eth.RetryCount(),
				eth.RetryDelay(),
				int(eth.GetFinalityDelay()),
				eth.GetTxFeePercentageToIncrease(),
				eth.GetTxMaxGasFeeAllowedInGwei(),
			},
			want: true,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ethereum.NewClient(tt.args.endpoint, tt.args.pathKeystore, tt.args.pathPasscodes, tt.args.defaultAccount, tt.args.timeout, tt.args.retryCount, tt.args.retryDelay, tt.args.txFeePercentageToIncrease, tt.args.getTxMaxGasFeeAllowedInGwei)
			if !tt.wantErr(t, err, fmt.Sprintf("NewEthereumEndpoint(%v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v)", tt.args.endpoint, tt.args.pathKeystore, tt.args.pathPasscodes, tt.args.defaultAccount, tt.args.timeout, tt.args.retryCount, tt.args.retryDelay, tt.args.txFeePercentageToIncrease, tt.args.getTxMaxGasFeeAllowedInGwei)) {
				return
			}
			if tt.want {
				assert.NotNilf(t, got, "Ethereum Details must not be nil")
			}
		})
	}
}
