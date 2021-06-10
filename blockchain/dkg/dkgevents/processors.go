package dkgevents

import (
	"strings"

	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func ProcessShareDistribution(logger *logrus.Logger, state *objects.MonitorState, log types.Log) error {

	logger.Info(strings.Repeat("-", 60))
	logger.Info("ProcessShareDistribution()")
	logger.Info(strings.Repeat("-", 60))

	// if !ETHDKGInProgress(state.EthDKG, log.BlockNumber) {
	// 	logger.Warn("Ignoring share distribution since we are not participating this round...")
	// 	return ErrCanNotContinue
	// }

	// eth := svcs.eth
	// c := eth.Contracts()
	// ethdkg := state.EthDKG

	// event, err := c.Ethdkg.ParseShareDistribution(log)
	// if err != nil {
	// 	return err
	// }

	// ethdkg.Commitments[event.Issuer] = event.Commitments
	// ethdkg.EncryptedShares[event.Issuer] = event.EncryptedShares

	return nil
}
