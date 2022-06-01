package events

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/MadBase/MadNet/blockchain/tasks/dkg/utils"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/dgraph-io/badger/v2"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/monitor/objects"
	dkg_obj "github.com/MadBase/MadNet/blockchain/tasks/dkg/objects"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// ProcessValidatorSetCompleted handles receiving validatorSet changes
func ProcessValidatorSetCompleted(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log, cdb *db.Database,
	adminHandler interfaces.AdminHandler) error {

	c := eth.Contracts()

	state.Lock()
	defer state.Unlock()

	updatedState := state

	event, err := c.Ethdkg().ParseValidatorSetCompleted(log)
	if err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"ValidatorCount": event.ValidatorCount,
		"Nonce":          event.Nonce,
		"Epoch":          event.Epoch,
		"EthHeight":      event.EthHeight,
		"AliceNetHeight": event.AliceNetHeight,
		"GroupKey0":      event.GroupKey0,
		"GroupKey1":      event.GroupKey1,
		"GroupKey2":      event.GroupKey2,
		"GroupKey3":      event.GroupKey3,
	}).Infof("ProcessValidatorSetCompleted()")

	epoch := uint32(event.Epoch.Int64())

	vs := state.ValidatorSets[epoch]
	vs.NotBeforeMadNetHeight = uint32(event.AliceNetHeight.Uint64())
	vs.ValidatorCount = uint8(event.ValidatorCount.Uint64())
	vs.GroupKey[0] = event.GroupKey0
	vs.GroupKey[1] = event.GroupKey1
	vs.GroupKey[2] = event.GroupKey2
	vs.GroupKey[3] = event.GroupKey3

	validatorSet, present := updatedState.ValidatorSets[epoch]
	if present {
		vs0b := validatorSet.GroupKey[0].Bytes()
		vs1b := vs.GroupKey[0].Bytes()
		if !bytes.Equal(vs0b, vs1b) {
			delete(updatedState.ValidatorSets, epoch)
		}
	}
	updatedState.ValidatorSets[epoch] = vs

	err = checkValidatorSet(updatedState, epoch, logger, cdb, adminHandler)
	if err != nil {
		return err
	}

	//TODO: remove all the EthDKG tasks
	//logger.WithFields(logrus.Fields{
	//	"Phase": state.EthDKG.Phase,
	//}).Infof("Purging schedule")
	//state.Schedule.Purge()

	err = cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := dkg_obj.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		dkgState.OnCompletion()

		err = dkg_obj.PersistEthDkgState(txn, logger, dkgState)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	if err = cdb.Sync(); err != nil {
		return err
	}

	return nil
}

// ProcessValidatorMemberAdded handles receiving keys for a specific validator
func ProcessValidatorMemberAdded(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log, cdb *db.Database) error {

	state.Lock()
	defer state.Unlock()

	c := eth.Contracts()

	event, err := c.Ethdkg().ParseValidatorMemberAdded(log)
	if err != nil {
		return err
	}

	epoch := uint32(event.Epoch.Int64())

	participantIndex := uint32(event.Index.Uint64())
	arrayIndex := participantIndex - 1

	v := objects.Validator{
		Account:   event.Account,
		Index:     uint8(participantIndex),
		SharedKey: [4]*big.Int{event.Share0, event.Share1, event.Share2, event.Share3},
	}

	err = cdb.Update(func(txn *badger.Txn) error {
		dkgState, err := dkg_obj.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		// sanity check
		if v.Account == dkgState.Account.Address &&
			dkgState.Participants[event.Account].GPKj[0] != nil &&
			dkgState.Participants[event.Account].GPKj[1] != nil &&
			dkgState.Participants[event.Account].GPKj[2] != nil &&
			dkgState.Participants[event.Account].GPKj[3] != nil &&
			(dkgState.Participants[event.Account].GPKj[0].Cmp(v.SharedKey[0]) != 0 ||
				dkgState.Participants[event.Account].GPKj[1].Cmp(v.SharedKey[1]) != 0 ||
				dkgState.Participants[event.Account].GPKj[2].Cmp(v.SharedKey[2]) != 0 ||
				dkgState.Participants[event.Account].GPKj[3].Cmp(v.SharedKey[3]) != 0) {

			return utils.LogReturnErrorf(logger, "my own GPKj doesn't match event! mine: %v | event: %v", dkgState.Participants[event.Account].GPKj, v.SharedKey)
		}

		// state update
		dkgState.OnGPKjSubmitted(event.Account, v.SharedKey)
		err = dkg_obj.PersistEthDkgState(txn, logger, dkgState)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return utils.LogReturnErrorf(logger, "Failed to save dkgState on ProcessValidatorMemberAdded: %v", err)
	}

	if err = cdb.Sync(); err != nil {
		return utils.LogReturnErrorf(logger, "Failed to set sync on ProcessValidatorMemberAdded: %v", err)
	}

	if len(state.Validators[epoch]) < int(participantIndex) {
		newValList := make([]objects.Validator, int(participantIndex))
		copy(newValList, state.Validators[epoch])
		state.Validators[epoch] = newValList
	}
	state.Validators[epoch][arrayIndex] = v
	ptrGroupShare := [4]*big.Int{
		v.SharedKey[0], v.SharedKey[1],
		v.SharedKey[2], v.SharedKey[3]}
	groupShare, err := bn256.MarshalG2Big(ptrGroupShare)
	if err != nil {
		logger.Errorf("Failed to marshal groupShare: %v", err)
		return err
	}

	groupShareHex := fmt.Sprintf("%x", groupShare)
	logger.WithFields(logrus.Fields{
		"Index":      v.Index,
		"GroupShare": groupShareHex,
	}).Infof("Received Validator")

	return nil
}

// ProcessValidatorMajorSlashed handles the Major Slash event
func ProcessValidatorMajorSlashed(eth interfaces.Ethereum, logger *logrus.Entry, log types.Log) error {

	logger.Info("ProcessValidatorMajorSlashed() ...")

	event, err := eth.Contracts().ValidatorPool().ParseValidatorMajorSlashed(log)
	if err != nil {
		return err
	}

	logger = logger.WithFields(logrus.Fields{
		"Account": event.Account.String(),
	})

	logger.Infof("ValidatorMajorSlashed")

	return nil
}

// ProcessValidatorMinorSlashed handles the Minor Slash event
func ProcessValidatorMinorSlashed(eth interfaces.Ethereum, logger *logrus.Entry, log types.Log) error {

	logger.Info("ProcessValidatorMinorSlashed() ...")

	event, err := eth.Contracts().ValidatorPool().ParseValidatorMinorSlashed(log)
	if err != nil {
		return err
	}

	logger = logger.WithFields(logrus.Fields{
		"Account":               event.Account.String(),
		"PublicStaking.TokenID": event.PublicStakingTokenID.Uint64(),
	})

	logger.Infof("ValidatorMinorSlashed")

	return nil
}

func checkValidatorSet(state *objects.MonitorState, epoch uint32, logger *logrus.Entry, cdb *db.Database, adminHandler interfaces.AdminHandler) error {

	logger = logger.WithField("Epoch", epoch)

	// Make sure we've received a validator set event
	validatorSet, present := state.ValidatorSets[epoch]
	if !present {
		logger.Warnf("No ValidatorSet received for epoch")
	}

	// Make sure we've received a validator member event
	validators, present := state.Validators[epoch]
	if !present {
		logger.Warnf("No ValidatorMember received for epoch")
	}

	dkgState := &dkg_obj.DkgState{}
	var err error
	err = cdb.View(func(txn *badger.Txn) error {
		dkgState, err = dkg_obj.LoadEthDkgState(txn, logger)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return utils.LogReturnErrorf(logger, "Failed to save dkgState on checkValidatorSet: %v", err)
	}

	// See how many validator members we've seen and how many we expect
	receivedCount := len(validators)
	expectedCount := dkgState.NumberOfValidators

	// Log validator set status
	logger.WithFields(logrus.Fields{
		"NotBeforeMadNetHeight": validatorSet.NotBeforeMadNetHeight,
		"ValidatorsReceived":    receivedCount,
		"ValidatorsExpected":    expectedCount,
	}).Infof("Building ValidatorSet...")

	if receivedCount == expectedCount || receivedCount == 0 {
		// Start by building the ValidatorSet
		ptrGroupKey := [4]*big.Int{validatorSet.GroupKey[0], validatorSet.GroupKey[1], validatorSet.GroupKey[2], validatorSet.GroupKey[3]}
		groupKey, err := bn256.MarshalG2Big(ptrGroupKey)
		if err != nil {
			logger.Errorf("Failed to marshal groupKey: %v", err)
			return err
		}
		vs := &objs.ValidatorSet{
			GroupKey:   groupKey,
			Validators: make([]*objs.Validator, validatorSet.ValidatorCount),
			NotBefore:  validatorSet.NotBeforeMadNetHeight}
		// Loop over the Validators
		if receivedCount != 0 {
			for _, validator := range validators {
				ptrGroupShare := [4]*big.Int{
					validator.SharedKey[0], validator.SharedKey[1],
					validator.SharedKey[2], validator.SharedKey[3]}
				groupShare, err := bn256.MarshalG2Big(ptrGroupShare)
				if err != nil {
					logger.Errorf("Failed to marshal groupShare: %v", err)
					return err
				}
				v := &objs.Validator{
					VAddr:      validator.Account.Bytes(),
					GroupShare: groupShare}
				vs.Validators[validator.Index-1] = v
				logger.WithFields(logrus.Fields{
					"Index":      validator.Index,
					"GroupShare": fmt.Sprintf("0x%x", groupShare),
					"Validator":  fmt.Sprintf("0x%x", v.VAddr),
				}).Info("ValidatorMember")
			}
		}

		validatorStrings := make([]string, len(vs.Validators))
		for idx := range vs.Validators {
			validatorStrings[idx] = fmt.Sprintf("0x%x", vs.Validators[idx].VAddr)
		}

		groupKeyStr := fmt.Sprintf("0x%x", vs.GroupKey)
		logger.WithFields(logrus.Fields{
			"GroupKey":   groupKeyStr,
			"NotBefore":  vs.NotBefore,
			"Validators": strings.Join(validatorStrings, ","),
		}).Infof("Complete ValidatorSet...")

		err = adminHandler.AddValidatorSet(vs)
		if err != nil {
			logger.Errorf("Unable to add validator set: %v", err) // TODO handle -- MUST retry or consensus shuts down
		}
	}
	return nil
}
