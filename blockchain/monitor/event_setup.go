package monitor

import (
	"github.com/MadBase/MadNet/blockchain/dkg/dkgevents"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/monitor/monevents"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

func SetupEventMap(em *objects.EventMap, consensusDb *db.Database, adminHandler interfaces.AdminHandler, depositHandler interfaces.DepositHandler) error {

	// DKG event processors
	if err := em.RegisterLocked("0x9c6f8368fe7e77e8cb9438744581403bcb3f53298e517f04c1b8475487402e97", "RegistrationOpen",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return dkgevents.ProcessOpenRegistration(eth, logger, state, log, adminHandler)
		}); err != nil {
		return err
	}
	if err := em.RegisterLocked("0xa84d294194d6169652a99150fd2ef10e18b0d2caa10beeea237bbddcc6e22b10", "ShareDistribution", dkgevents.ProcessShareDistribution); err != nil {
		return err
	}
	if err := em.RegisterLocked("0xb0ee36c3780de716eb6c83687f433ae2558a6923e090fd238b657fb6c896badc", "KeyShareSubmission", dkgevents.ProcessKeyShareSubmission); err != nil {
		return err
	}

	// Events to pass through to side chain
	if err := em.RegisterLocked("0x5b063c6569a91e8133fc6cd71d31a4ca5c65c652fd53ae093f46107754f08541", "DepositReceived",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessDepositReceived(eth, logger, state, log, consensusDb, depositHandler)
		}); err != nil {
		return err
	}

	if err := em.RegisterLocked("0x6d438b6b835d16cdae6efdc0259fdfba17e6aa32dae81863a2467866f85f724a", "SnapshotTaken",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessSnapshotTaken(eth, logger, state, log, adminHandler)
		}); err != nil {
		return err
	}

	if err := em.RegisterLocked("0x113b129fac2dde341b9fbbec2bb79a95b9945b0e80fda711fc8ae5c7b0ea83b0", "ValidatorMember",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessValidatorMember(eth, logger, state, log, adminHandler)
		}); err != nil {
		return err
	}
	if err := em.RegisterLocked("0x1c85ff1efe0a905f8feca811e617102cb7ec896aded693eb96366c8ef22bb09f", "ValidatorSet",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessValidatorSet(eth, logger, state, log, adminHandler)
		}); err != nil {
		return err
	}

	if err := em.RegisterLocked("0x36dcd0e03525dedd9d5c21a263ef5f35d030298b5c48f1a713006aefc064ad05", "ValueUpdated",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessValueUpdated(eth, logger, state, log, adminHandler)
		}); err != nil {
		return err
	}

	// Registering just for informational purposes
	if err := em.RegisterLocked("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925", "DSTokenApproval", nil); err != nil {
		return err
	}
	if err := em.RegisterLocked("0xce241d7ca1f669fee44b6fc00b8eba2df3bb514eed0f6f668f8f89096e81ed94", "LogSetOwner", nil); err != nil {
		return err
	}
	if err := em.RegisterLocked("0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885", "Mint", nil); err != nil {
		return err
	}
	if err := em.RegisterLocked("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", "Transfer", nil); err != nil {
		return err
	}
	if err := em.RegisterLocked("0x8c25e214c5693ebaf8008875bacedeb9e0aafd393864a314ed1801b2a4e13dd9", "ValidatorJoined", nil); err != nil {
		return err
	}
	if err := em.RegisterLocked("0x319bbadb03b94aedc69babb34a28675536a9cb30f4bbde343e1d0018c44ebd94", "ValidatorLeft", nil); err != nil {
		return err
	}
	if err := em.RegisterLocked("0x1de2f07b0a1c69916a8b25b889051644192307ea08444a2e11f8654d1db3ab0c", "LockedStake", nil); err != nil {
		return err
	}

	return nil
}
