package main

import (
	"github.com/alicenet/alicenet/cmd/bootnode"
	"github.com/alicenet/alicenet/cmd/ethkey"
	"github.com/alicenet/alicenet/cmd/firewalld"
	"github.com/alicenet/alicenet/cmd/initialization"
	"github.com/alicenet/alicenet/cmd/node"
	"github.com/alicenet/alicenet/cmd/root"
	"github.com/alicenet/alicenet/cmd/utils"
	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Version from git tag.
	version               = "dev"
	defaultConfigLocation = "/.alicenet/scripts/base-files/bootnode.toml"
)

func main() {
	logger := logging.GetLogger("main")
	logger.SetLevel(logrus.InfoLevel)

	config.Configuration.Version = version

	// Root for all commands
	rootCmd := root.Cmd
	utilsCmd := utils.Command
	utilsSendWeiCmd := utils.SendWeiCommand
	nodeCmd := node.Command
	bootNodeCmd := bootnode.Command
	ethkeyCmd := ethkey.Generate
	firewallCmd := firewalld.Command
	initializationCmd := initialization.Command
	utilsCmd.AddCommand(utilsSendWeiCmd)
	rootCmd.AddCommand(
		utilsCmd,
		nodeCmd,
		bootNodeCmd,
		ethkeyCmd,
		firewallCmd,
		initializationCmd,
	)

	hierarchy := map[*cobra.Command]*cobra.Command{
		firewallCmd:       rootCmd,
		bootNodeCmd:       rootCmd,
		utilsCmd:          rootCmd,
		nodeCmd:           rootCmd,
		ethkeyCmd:         rootCmd,
		initializationCmd: rootCmd,
		utilsSendWeiCmd:   utilsCmd,
	}

	options := []*cobra.Command{rootCmd, utilsCmd, nodeCmd, bootNodeCmd, ethkeyCmd}
	// If none command and option are present, the `node` command with the default --config option will be executed.
	config.LoadConfig(options, logger, nodeCmd, hierarchy)

	// Really start application here
	err := rootCmd.Execute()
	if err != nil {
		logger.Fatalf("Execute() failed:%q", err)
	}
	logger.Debugf("main() -- Configuration:%v", config.Configuration.Ethereum)
}
