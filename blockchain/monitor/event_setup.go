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

func SetupEventMap(em *objects.EventMap, cdb *db.Database, adminHandler interfaces.AdminHandler, depositHandler interfaces.DepositHandler) error {

	// DKG event processors
	if err := em.RegisterLocked("0xbda431b9b63510f1398bf33d700e013315bcba905507078a1780f13ea5b354b9", "RegistrationOpened", dkgevents.ProcessRegistrationOpened); err != nil {
		return err
	}

	if err := em.RegisterLocked("0x7f1304057ec61140fbf2f5f236790f34fcafe123d3eb0d298d92317c97da500d", "AddressRegistered", dkgevents.ProcessAddressRegistered); err != nil {
		return err
	}

	if err := em.RegisterLocked("0x833013b96b786b4eca83baac286920e5e53956c21ff3894f1d9f02e97d6ed764", "RegistrationComplete", dkgevents.ProcessRegistrationComplete); err != nil {
		return err
	}

	if err := em.RegisterLocked("0xf0c8b0ef2867c2b4639b404a0296b6bbf0bf97e20856af42144a5a6035c0d0d2", "SharesDistributed", dkgevents.ProcessShareDistribution); err != nil {
		return err
	}

	if err := em.RegisterLocked("0xbfe94ffef5ddde4d25ac7b652f3f67686ea63f9badbfe1f25451e26fc262d11c", "ShareDistributionComplete", dkgevents.ProcessShareDistributionComplete); err != nil {
		return err
	}

	if err := em.RegisterLocked("0x6162e2d11398e4063e4c8565dafc4fb6755bbead93747ea836a5ef73a594aaf7", "KeyShareSubmitted", dkgevents.ProcessKeyShareSubmitted); err != nil {
		return err
	}

	if err := em.RegisterLocked("0x522cec98f6caa194456c44afa9e8cef9ac63eecb0be60e20d180ce19cfb0ef59", "KeyShareSubmissionComplete", dkgevents.ProcessKeyShareSubmissionComplete); err != nil {
		return err
	}

	if err := em.RegisterLocked("0x71b1ebd27be320895a22125d6458e3363aefa6944a312ede4bf275867e6d5a71", "MPKSet", func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
		return dkgevents.ProcessMPKSet(eth, logger, state, log, adminHandler)
	}); err != nil {
		return err
	}

	if err := em.RegisterLocked("0x09b90b08bbc3dbe22e9d2a0bc9c2c7614c7511cd0ad72177727a1e762115bf06", "ValidatorMemberAdded",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessValidatorMemberAdded(eth, logger, state, log, adminHandler)
		}); err != nil {
		return err
	}

	if err := em.RegisterLocked("0x87bfe600b78cad9f7cf68c99eb582c1748f636b3269842b37d5873b0e069f628", "GPKJSubmissionComplete", dkgevents.ProcessGPKJSubmissionComplete); err != nil {
		return err
	}

	if err := em.RegisterLocked("0xd7237b781669fa700ecf77be6cd8fa0f4b98b1a24ac584a9b6b44c509216718a", "ValidatorSetCompleted",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessValidatorSetCompleted(eth, logger, state, log, adminHandler)
		}); err != nil {
		return err
	}

	//
	// done

	// Events to pass through to side chain
	if err := em.RegisterLocked("0x5b063c6569a91e8133fc6cd71d31a4ca5c65c652fd53ae093f46107754f08541", "DepositReceived",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessDepositReceived(eth, logger, state, log, cdb, depositHandler)
		}); err != nil {
		return err
	}

	if err := em.RegisterLocked("0x6d438b6b835d16cdae6efdc0259fdfba17e6aa32dae81863a2467866f85f724a", "SnapshotTaken",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessSnapshotTaken(eth, logger, state, log, adminHandler)
		}); err != nil {
		return err
	}

	// todo: delete this bc deprecated
	if err := em.RegisterLocked("0x1c85ff1efe0a905f8feca811e617102cb7ec896aded693eb96366c8ef22bb09f", "ValidatorSet",
		func(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log) error {
			return monevents.ProcessValidatorSetCompleted(eth, logger, state, log, adminHandler)
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
