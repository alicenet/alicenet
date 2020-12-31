package deploy

import (
	"context"
	"math/big"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/config"
	"github.com/MadBase/MadNet/logging"
	"github.com/spf13/cobra"
)

var RequiredWei = big.NewInt(8_000_000_000_000)

// Command is the cobra.Command specifically for running as an edge node, i.e. not a validator or relay
var Command = cobra.Command{
	Use:   "deploy",
	Short: "Deploys required smart contracts to Ethereum",
	Long:  "deploy uses Go bindings to deploy required smart contracts.",
	Run:   deployNode}

func deployNode(cmd *cobra.Command, args []string) {
	logger := logging.GetLogger("deploy")
	logger.Info("Deploying contracts...")

	eth, err := blockchain.NewEthereumEndpoint(
		config.Configuration.Ethereum.Endpoint,
		config.Configuration.Ethereum.Keystore,
		config.Configuration.Ethereum.Passcodes,
		config.Configuration.Ethereum.DefaultAccount,
		config.Configuration.Ethereum.Timeout,
		config.Configuration.Ethereum.RetryCount,
		config.Configuration.Ethereum.RetryDelay,
		config.Configuration.Ethereum.FinalityDelay)
	if err != nil {
		logger.Errorf("Could not connect to Ethereum: %v", err)
	}
	c := eth.Contracts()

	acct := eth.GetDefaultAccount()
	err = eth.UnlockAccount(acct)
	if err != nil {
		logger.Fatal(err)
	}

	bal, err := eth.GetBalance(acct.Address)
	if err != nil {
		logger.Warnf("Could not get balance for %v: %v", acct.Address.Hex(), err)
	}
	logger.Infof("DeployAccount: %v Balance: %v", acct.Address.Hex(), bal.String())

	if bal.Cmp(RequiredWei) < 0 {
		logger.Warnf("Probably insufficient gas, but trying anyway")
	}
	_, _, err = c.DeployContracts(context.Background(), acct)
	if err != nil {
		logger.Errorf("Could not deploy contracts: %v", err)
	}
}
