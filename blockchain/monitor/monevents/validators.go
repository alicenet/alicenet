package monevents

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// ProcessValidatorSet handles receiving validatorSet changes
func ProcessValidatorSet(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log,
	adminHandler interfaces.AdminHandler) error {

	c := eth.Contracts()

	updatedState := state

	event, err := c.Ethdkg().ParseValidatorSet(log)
	if err != nil {
		return err
	}

	epoch := uint32(event.Epoch.Int64())

	vs := state.ValidatorSets[epoch]
	vs.NotBeforeMadNetHeight = event.MadHeight
	vs.ValidatorCount = event.ValidatorCount
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
			delete(updatedState.Validators, epoch)
		}
	}
	updatedState.ValidatorSets[epoch] = vs

	err = checkValidatorSet(updatedState, epoch, logger, adminHandler)
	if err != nil {
		return err
	}

	return nil
}

// ProcessValidatorMember handles receiving keys for a specific validator
func ProcessValidatorMember(eth interfaces.Ethereum, logger *logrus.Entry, state *objects.MonitorState, log types.Log,
	adminHandler interfaces.AdminHandler) error {

	c := eth.Contracts()

	event, err := c.Ethdkg().ParseValidatorMember(log)
	if err != nil {
		return err
	}

	epoch := uint32(event.Epoch.Int64())

	index := uint8(event.Index.Uint64()) - 1

	v := objects.Validator{
		Account:   event.Account,
		Index:     index,
		SharedKey: [4]*big.Int{event.Share0, event.Share1, event.Share2, event.Share3},
	}
	if len(state.Validators[epoch]) < int(index+1) {
		newValList := make([]objects.Validator, int(index+1))
		copy(newValList, state.Validators[epoch])
		state.Validators[epoch] = newValList
	}
	state.Validators[epoch][index] = v
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
		"GroupShare": groupShareHex}).Infof("Received Validator")

	err = checkValidatorSet(state, epoch, logger, adminHandler)
	if err != nil {
		return err
	}

	return nil
}

func checkValidatorSet(state *objects.MonitorState, epoch uint32, logger *logrus.Entry, adminHandler interfaces.AdminHandler) error {

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

	// See how many validator members we've seen and how many we expect
	receivedCount := len(validators)
	expectedCount := int(validatorSet.ValidatorCount)

	// Log validator set status
	logger.WithFields(logrus.Fields{
		"NotBeforeMadNetHeight": validatorSet.NotBeforeMadNetHeight,
		"ValidatorsReceived":    receivedCount,
		"ValidatorsExpected":    expectedCount,
	}).Infof("Building ValidatorSet...")

	if receivedCount == expectedCount {
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
			vs.Validators[validator.Index] = v
			logger.WithFields(logrus.Fields{
				"Index":      validator.Index,
				"GroupShare": fmt.Sprintf("0x%x", groupShare),
				"Validator":  fmt.Sprintf("0x%x", v.VAddr),
			}).Info("ValidatorMember")
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
