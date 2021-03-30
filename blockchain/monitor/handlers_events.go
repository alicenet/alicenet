package monitor

import (
	"context"
	"math/big"
	"strings"

	aobjs "github.com/MadBase/MadNet/application/objs"
	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/MadBase/MadNet/blockchain/tasks/dkgtasks"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/logging"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// ProcessRegistrationOpen when we see ETHDKG has initialized we need to start
func (svcs *Services) ProcessRegistrationOpen(state *State, log types.Log) error {

	eth := svcs.eth
	c := eth.Contracts()
	logger := svcs.logger

	event, err := c.Ethdkg.ParseRegistrationOpen(log)
	if err != nil {
		return err
	}

	if ETHDKGInProgress(state.ethdkg, log.BlockNumber) {
		logger.Warnf("Received RegistrationOpen event while ETHDKG in progress. Aborting old round.")
		AbortETHDKG(state.ethdkg)
	}

	if event.RegistrationEnds.Uint64() > state.HighestBlockFinalized {

		private, public, err := dkg.GenerateKeys()
		if err != nil {
			return err
		}

		logger.Infof(strings.Repeat("-", 80))
		logger.Infof("Registration Open from block %v to %v", event.DkgStarts, event.RegistrationEnds)
		logger.Infof(strings.Repeat("-", 80))
		logger.Infof("          publicKey: %v", dkgtasks.FormatBigIntSlice(public[:]))
		logger.Infof(strings.Repeat("-", 80))

		// Record the schedule
		schedule := &EthDKGSchedule{}
		schedule.RegistrationStart = log.BlockNumber
		schedule.RegistrationEnd = event.RegistrationEnds.Uint64()

		schedule.ShareDistributionStart = schedule.RegistrationEnd + 1
		schedule.ShareDistributionEnd = event.ShareDistributionEnds.Uint64()

		schedule.DisputeStart = schedule.ShareDistributionEnd + 1
		schedule.DisputeEnd = event.DisputeEnds.Uint64()

		schedule.KeyShareSubmissionStart = schedule.DisputeEnd + 1
		schedule.KeyShareSubmissionEnd = event.KeyShareSubmissionEnds.Uint64()

		schedule.MPKSubmissionStart = schedule.KeyShareSubmissionEnd + 1
		schedule.MPKSubmissionEnd = event.MpkSubmissionEnds.Uint64()

		schedule.GPKJSubmissionStart = schedule.MPKSubmissionEnd + 1
		schedule.GPKJSubmissionEnd = event.GpkjSubmissionEnds.Uint64()

		schedule.GPKJGroupAccusationStart = schedule.GPKJSubmissionEnd + 1
		schedule.GPKJGroupAccusationEnd = event.GpkjDisputeEnds.Uint64()

		schedule.CompleteStart = schedule.GPKJGroupAccusationEnd + 1
		schedule.CompleteEnd = event.DkgComplete.Uint64()

		// TODO associate names with these also to help with debugging/logging
		ib := make(map[uint64]func(*State, uint64) error)
		ib[schedule.ShareDistributionStart] = svcs.DoDistributeShares      // Do ShareDistribution
		ib[schedule.DisputeStart] = svcs.DoSubmitDispute                   // Do Disputes
		ib[schedule.KeyShareSubmissionStart] = svcs.DoSubmitKeyShare       // Do KeyShareSubmission
		ib[schedule.MPKSubmissionStart] = svcs.DoSubmitMasterPublicKey     // Do MPKSubmission
		ib[schedule.GPKJSubmissionStart] = svcs.DoSubmitGPKj               // Do GPKJSubmission
		ib[schedule.GPKJGroupAccusationStart] = svcs.DoGroupAccusationGPKj // Do GPKJDisputes
		ib[schedule.CompleteStart] = svcs.DoSuccessfulCompletion           // Do SuccessfulCompletion

		logger.Infof("Adding block processors for %v", ib)
		state.interestingBlocks = ib

		acct := eth.GetDefaultAccount()

		state.ethdkg = NewEthDKGState()
		state.ethdkg.Address = acct.Address
		state.ethdkg.Schedule = schedule
		state.ethdkg.TransportPrivateKey = private
		state.ethdkg.TransportPublicKey = public

		taskLogger := logging.GetLogger("rt")

		task := dkgtasks.NewRegisterTask(
			state.ethdkg.TransportPublicKey,
			state.ethdkg.Schedule.RegistrationEnd)

		state.ethdkg.RegistrationTH = svcs.taskMan.NewTaskHandler(taskLogger, eth, task)

		state.ethdkg.RegistrationTH.Start()

	} else {
		logger.Infof("Not participating in DKG... registration ends at height %v but height %v is finalized.",
			event.RegistrationEnds, state.HighestBlockFinalized)
		// state.ethdkg = EthDKGState{} // TODO I need to cancel any TaskHandlers too
	}

	return nil
}

// ProcessShareDistribution accumulates everyones shares ETHDKG
func (svcs *Services) ProcessShareDistribution(state *State, log types.Log) error {
	logger := svcs.logger

	logger.Info(strings.Repeat("-", 60))
	logger.Info("ProcessShareDistribution()")
	logger.Info(strings.Repeat("-", 60))

	if !ETHDKGInProgress(state.ethdkg, log.BlockNumber) {
		logger.Warn("Ignoring share distribution since we are not participating this round...")
		return ErrCanNotContinue
	}

	eth := svcs.eth
	c := eth.Contracts()
	ethdkg := state.ethdkg

	event, err := c.Ethdkg.ParseShareDistribution(log)
	if err != nil {
		return err
	}

	ethdkg.Commitments[event.Issuer] = event.Commitments
	ethdkg.EncryptedShares[event.Issuer] = event.EncryptedShares

	return nil
}

// ProcessKeyShareSubmission ETHDKG
func (svcs *Services) ProcessKeyShareSubmission(state *State, log types.Log) error {

	logger := svcs.logger

	logger.Info(strings.Repeat("-", 60))
	logger.Info("ProcessKeyShareSubmission()")
	logger.Info(strings.Repeat("-", 60))

	if !ETHDKGInProgress(state.ethdkg, log.BlockNumber) {
		logger.Warn("Ignoring key share submission since we are not participating this round...")
		return ErrCanNotContinue
	}

	eth := svcs.eth
	c := eth.Contracts()
	event, err := c.Ethdkg.ParseKeyShareSubmission(log)
	if err != nil {
		return err
	}

	addr := event.Issuer
	keyshareG1 := event.KeyShareG1
	keyshareG1Proof := event.KeyShareG1CorrectnessProof
	keyshareG2 := event.KeyShareG2

	logger.Infof("keyshareG1:%v keyshareG2:%v", dkgtasks.FormatBigIntSlice(keyshareG1[:]), dkgtasks.FormatBigIntSlice(keyshareG2[:]))

	state.ethdkg.KeyShareG1s[addr] = keyshareG1
	state.ethdkg.KeyShareG1CorrectnessProofs[addr] = keyshareG1Proof
	state.ethdkg.KeyShareG2s[addr] = keyshareG2

	return nil
}

// ProcessValidatorSet handles receiving validatorSet changes
func (svcs *Services) ProcessValidatorSet(state *State, log types.Log) error {

	eth := svcs.eth
	c := eth.Contracts()

	updatedState := state

	event, err := c.Ethdkg.ParseValidatorSet(log)
	if err != nil {
		return err
	}

	epoch := uint32(event.Epoch.Int64())

	vs := state.ValidatorSets[epoch]
	vs.NotBeforeMadNetHeight = event.MadHeight
	vs.ValidatorCount = event.ValidatorCount
	vs.GroupKey[0] = *event.GroupKey0
	vs.GroupKey[1] = *event.GroupKey1
	vs.GroupKey[2] = *event.GroupKey2
	vs.GroupKey[3] = *event.GroupKey3

	updatedState.ValidatorSets[epoch] = vs

	err = svcs.checkValidatorSet(updatedState, epoch)
	if err != nil {
		return err
	}

	return nil
}

// ProcessValidatorMember handles receiving keys for a specific validator
func (svcs *Services) ProcessValidatorMember(state *State, log types.Log) error {

	eth := svcs.eth
	c := eth.Contracts()

	event, err := c.Ethdkg.ParseValidatorMember(log)
	if err != nil {
		return err
	}

	epoch := uint32(event.Epoch.Int64())

	index := uint8(event.Index.Uint64()) - 1

	v := Validator{
		Account:   event.Account,
		Index:     index,
		SharedKey: [4]big.Int{*event.Share0, *event.Share1, *event.Share2, *event.Share3},
	}
	if len(state.Validators) < int(index+1) {
		newValList := make([]Validator, int(index+1))
		copy(newValList, state.Validators[epoch])
		state.Validators[epoch] = newValList
	}
	state.Validators[epoch][index] = v
	ptrGroupShare := [4]*big.Int{
		&v.SharedKey[0], &v.SharedKey[1],
		&v.SharedKey[2], &v.SharedKey[3]}
	groupShare, err := bn256.MarshalG2Big(ptrGroupShare)
	if err != nil {
		svcs.logger.Errorf("Failed to marshal groupShare: %v", err)
		return err
	}
	svcs.logger.Debugf("Validator member %v %x", v.Index, groupShare)
	err = svcs.checkValidatorSet(state, epoch)
	if err != nil {
		return err
	}

	return nil
}

func (svcs *Services) checkValidatorSet(state *State, epoch uint32) error {
	logger := svcs.logger

	// Make sure we've received a validator set event
	validatorSet, present := state.ValidatorSets[epoch]
	if !present {
		logger.Warnf("No validator set received for epoch %v", epoch)
	}

	// Make sure we've received a validator member event
	validators, present := state.Validators[epoch]
	if !present {
		logger.Warnf("No validators received for epoch %v", epoch)
	}

	// See how many validator members we've seen and how many we expect
	receivedCount := len(validators)
	expectedCount := int(validatorSet.ValidatorCount)

	// Log validator set status
	logLevel := logrus.WarnLevel
	if receivedCount == expectedCount && expectedCount > 0 {
		logLevel = logrus.InfoLevel
	}
	logger.Logf(logLevel, "Epoch: %v NotBeforeMadNetHeight: %v Validators Received: %v of %v", epoch, validatorSet.NotBeforeMadNetHeight, receivedCount, expectedCount)

	if receivedCount == expectedCount {
		// Start by building the ValidatorSet
		ptrGroupKey := [4]*big.Int{&validatorSet.GroupKey[0], &validatorSet.GroupKey[1], &validatorSet.GroupKey[2], &validatorSet.GroupKey[3]}
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
				&validator.SharedKey[0], &validator.SharedKey[1],
				&validator.SharedKey[2], &validator.SharedKey[3]}
			groupShare, err := bn256.MarshalG2Big(ptrGroupShare)
			if err != nil {
				logger.Errorf("Failed to marshal groupShare: %v", err)
				return err
			}
			v := &objs.Validator{
				VAddr:      validator.Account.Bytes(),
				GroupShare: groupShare}
			vs.Validators[validator.Index] = v
			logger.Infof("ValidatorMember[%v]: {GroupShare: 0x%x, VAddr: %x}", validator.Index, groupShare, v.VAddr)
		}
		logger.Infof("ValidatorSet: {GroupKey: 0x%x, NotBefore: %v, Validators: %v }", vs.GroupKey, vs.NotBefore, vs.Validators)
		err = svcs.ah.AddValidatorSet(vs)
		if err != nil {
			logger.Errorf("Unable to add validator set: %v", err) // TODO handle -- MUST retry or consensus shuts down
		}
	}
	return nil
}

// ProcessDepositReceived handles logic around receiving a deposit event
func (svcs *Services) ProcessDepositReceived(state *State, log types.Log) error {

	eth := svcs.eth
	c := eth.Contracts()
	logger := svcs.logger

	event, err := c.Deposit.ParseDepositReceived(log)
	if err != nil {
		return err
	}

	logger.Infof("deposit depositID:%x ethereum:0x%x amount:%d",
		event.DepositID, event.Depositor, event.Amount)

	return svcs.consensusDb.Update(func(txn *badger.Txn) error {
		depositNonce := event.DepositID.Bytes()
		account := event.Depositor.Bytes()
		owner := &aobjs.Owner{}
		err := owner.New(account, constants.CurveSecp256k1)
		if err != nil {
			logger.Debugf("Error in Services.ProcessDepositReceived at owner.New: %v", err)
			return err
		}
		return svcs.dph.Add(txn, svcs.chainID, depositNonce, event.Amount, owner)
	})
}

// ProcessSnapshotTaken handles receiving snapshots
func (svcs *Services) ProcessSnapshotTaken(state *State, log types.Log) error {

	eth := svcs.eth
	c := eth.Contracts()
	logger := svcs.logger
	callOpts := eth.GetCallOpts(context.TODO(), eth.GetDefaultAccount())

	event, err := c.Validators.ParseSnapshotTaken(log)
	if err != nil {
		return err // TODO consensus on side will stop
	}

	epoch := event.Epoch
	ethDkgStarted := event.StartingETHDKG

	logger.Infof("SnapshotTaken -> ChainID:%v Epoch:%v Height:%v Validator:%v StartingETHDKG:%v", event.ChainId, epoch, event.Height, event.Validator.Hex(), event.StartingETHDKG)

	rawBClaims, err := c.Validators.GetRawBlockClaimsSnapshot(callOpts, epoch)
	if err != nil {
		return err // TODO consensus on side will stop
	}

	rawSignature, err := c.Validators.GetRawSignatureSnapshot(callOpts, epoch)
	if err != nil {
		return err // TODO consensus on side will stop
	}

	// put it back together
	bclaims := &objs.BClaims{}
	err = bclaims.UnmarshalBinary(rawBClaims)
	if err != nil {
		return err
	}
	header := &objs.BlockHeader{}
	header.BClaims = bclaims
	header.SigGroup = rawSignature

	// send the reconstituted header to a handler
	err = svcs.ah.AddSnapshot(header, ethDkgStarted) // TODO must happen or things stuff
	if err != nil {
		return err
	}

	return nil
}
