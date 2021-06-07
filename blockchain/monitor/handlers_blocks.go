package monitor

import (
	"errors"
	"math/big"
	"strings"

	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/MadBase/bridge/bindings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// Useful pseudo-constants
var big0 = big.NewInt(0)
var big1 = big.NewInt(1)
var big2 = big.NewInt(2)
var big3 = big.NewInt(3)

// DoDistributeShares this should happen when it's time to distribute shares
func (svcs *Services) DoDistributeShares(state *State, block uint64) error {

	// eth := svcs.eth
	logger := svcs.logger

	logger.Infof(strings.Repeat("-", 60))
	logger.Infof("DoDistributeShares()")
	logger.Infof(strings.Repeat("-", 60))

	// // If registration wasn't successful then quit now
	// if state.EthDKG.RegistrationTH == nil || !state.EthDKG.RegistrationTH.Successful() {
	// 	AbortETHDKG(state.EthDKG)
	// 	return fmt.Errorf("Registration didn't complete succesful, exiting ETHDKG")
	// }

	// // Basic setup
	// ctx, cancel := context.WithTimeout(context.Background(), eth.Timeout())
	// defer cancel()

	// callOpts := eth.GetCallOpts(ctx, eth.GetDefaultAccount())

	// // Retrieve validators
	// participants, myIndex, err := RetrieveParticipants(eth, callOpts)
	// if err != nil {
	// 	return ErrCanNotContinue
	// }

	// if myIndex == math.MaxInt32 {
	// 	logger.Errorf("Can't determine my (%v) index ", state.EthDKG.Address)
	// 	return ErrCanNotContinue
	// }

	// // Save state
	// ethdkg := state.EthDKG

	// ethdkg.Index = myIndex
	// ethdkg.NumberOfValidators = len(participants)
	// ethdkg.Participants = participants
	// ethdkg.ValidatorThreshold, _ = thresholdFromUsers(state.EthDKG.NumberOfValidators)

	// // Do the math
	// encryptedShares, privateCoefficients, commitments, err := dkg.GenerateShares(
	// 	ethdkg.TransportPrivateKey, ethdkg.TransportPublicKey,
	// 	ethdkg.Participants, ethdkg.ValidatorThreshold)
	// if err != nil {
	// 	return fmt.Errorf("Can't GenerateShares: %v", err)
	// }

	// // Store everything we'll need later
	// ethdkg.PrivateCoefficients = privateCoefficients
	// ethdkg.SecretValue = privateCoefficients[0]

	// // Do the mechanics of calling
	// taskLogger := logging.GetLogger("sdt")
	// taskLogger.Infof("%v %v", encryptedShares, commitments)

	// // task := dkgtasks.NewShareDistributionTask(eth.GetDefaultAccount(),
	// // 	ethdkg.TransportPublicKey, encryptedShares, commitments,
	// // 	ethdkg.Schedule.RegistrationEnd, ethdkg.Schedule.ShareDistributionEnd)

	// // ethdkg.ShareDistributionTH = svcs.taskMan.NewTaskHandler(taskLogger, eth, task)
	// // ethdkg.ShareDistributionTH.Start()

	return nil
}

// DoSubmitDispute submits a dispute if any of the shares we've seen are bad
func (svcs *Services) DoSubmitDispute(state *State, block uint64) error {
	svcs.logger.Infof(strings.Repeat("-", 60))
	svcs.logger.Infof("=== DoSubmitDispute                                     ===")
	svcs.logger.Infof(strings.Repeat("-", 60))

	// First confirm we distributed shares
	// if state.EthDKG.ShareDistributionTH == nil || !state.EthDKG.ShareDistributionTH.Successful() {
	// 	AbortETHDKG(state.EthDKG)
	// 	return fmt.Errorf("Share distribution didn't complete succesful, exiting ETHDKG")
	// }

	// //

	// // Setup and start task
	// taskLogger := logging.GetLogger("dispute")
	// eth := svcs.eth
	// ethdkg := state.EthDKG

	// task := dkgtasks.NewDisputeTask(
	// 	eth.GetDefaultAccount(),
	// 	ethdkg.TransportPublicKey,
	// 	ethdkg.Schedule.RegistrationEnd, ethdkg.Schedule.DisputeEnd)

	// ethdkg.DisputeTH = svcs.taskMan.NewTaskHandler(taskLogger, eth, task)
	// ethdkg.DisputeTH.Start()

	return nil
}

// DoSubmitKeyShare does something
func (svcs *Services) DoSubmitKeyShare(state *State, block uint64) error {

	logger := svcs.logger

	logger.Infof(strings.Repeat("-", 60))
	logger.Infof("DoSubmitKeyShare()")
	logger.Infof(strings.Repeat("-", 60))

	// If dispute wasn't successful then quit now
	// if state.EthDKG.DisputeTH == nil || !state.EthDKG.DisputeTH.Successful() {
	// 	AbortETHDKG(state.EthDKG)
	// 	return fmt.Errorf("Share dispute didn't complete succesful, exiting ETHDKG")
	// }

	// // Generate the key shares
	// g1KeyShare, g1Proof, g2KeyShare, err := dkg.GenerateKeyShare(state.EthDKG.SecretValue)
	// if err != nil {
	// 	return fmt.Errorf("Can't GenerateKeyShare: %v", err)
	// }

	// eth := svcs.eth
	// ethdkg := state.EthDKG

	// taskLogger := logging.GetLogger("kst")

	// task := dkgtasks.NewKeyshareSubmissionTask(
	// 	eth.GetDefaultAccount(),
	// 	ethdkg.TransportPublicKey, g1KeyShare, g1Proof, g2KeyShare,
	// 	ethdkg.Schedule.RegistrationEnd, ethdkg.Schedule.KeyShareSubmissionEnd)

	// ethdkg.KeyShareSubmissionTH = svcs.taskMan.NewTaskHandler(taskLogger, eth, task)
	// ethdkg.KeyShareSubmissionTH.Start()

	return nil
}

// DoSubmitMasterPublicKey does something
func (svcs *Services) DoSubmitMasterPublicKey(state *State, block uint64) error {

	logger := svcs.logger

	logger.Infof(strings.Repeat("-", 60))
	logger.Infof("DoSubmitMasterPublicKey()")
	logger.Infof(strings.Repeat("-", 60))

	// First confirm we submitted key shares
	// if state.EthDKG.KeyShareSubmissionTH == nil || !state.EthDKG.KeyShareSubmissionTH.Successful() {
	// 	AbortETHDKG(state.EthDKG)
	// 	return fmt.Errorf("Key share submission didn't complete succesful, exiting ETHDKG")
	// }
	// ethdkg := state.EthDKG

	// keyShareG1s := make([][2]*big.Int, ethdkg.NumberOfValidators)
	// keyShareG2s := make([][4]*big.Int, ethdkg.NumberOfValidators)

	// for _, participant := range ethdkg.Participants {
	// 	logger.Infof("Participant: %v", participant.Address.Hex())

	// 	pg1 := ethdkg.KeyShareG1s[participant.Address]
	// 	pg2 := ethdkg.KeyShareG2s[participant.Address]

	// 	logger.Infof("pg1: %v", pg1)
	// 	logger.Infof("pg2: %v", pg2)

	// 	keyShareG1s[participant.Index] = pg1
	// 	keyShareG2s[participant.Index] = pg2
	// }

	// // TODO Guard against missing keyshares, panic can happen
	// mpk, err := dkg.GenerateMasterPublicKey(keyShareG1s, keyShareG2s)
	// if err != nil {
	// 	return fmt.Errorf("Can't GenerateMasterPublicKey: %v", err)
	// }

	// state.EthDKG.MasterPublicKey = mpk

	// logger.Infof("MasterPublicKey: %v", dkgtasks.FormatBigIntSlice(mpk[:]))

	// // Task setup
	// eth := svcs.eth

	// task := dkgtasks.NewMPKSubmissionTask(
	// 	eth.GetDefaultAccount(),
	// 	ethdkg.TransportPublicKey, ethdkg.MasterPublicKey,
	// 	ethdkg.Schedule.RegistrationEnd, ethdkg.Schedule.MPKSubmissionEnd)

	// state.EthDKG.MPKSubmissionTH = svcs.taskMan.NewTaskHandler(logger, eth, task)
	// state.EthDKG.MPKSubmissionTH.Start()

	return nil
}

// DoSubmitGPKj does something
func (svcs *Services) DoSubmitGPKj(state *State, block uint64) error {

	logger := svcs.logger

	logger.Infof(strings.Repeat("-", 60))
	logger.Infof("DoSubmitGPKj()")
	logger.Infof(strings.Repeat("-", 60))

	// If dispute wasn't successful then quit now
	// if state.EthDKG.MPKSubmissionTH == nil || !state.EthDKG.MPKSubmissionTH.Successful() {
	// 	AbortETHDKG(state.EthDKG)
	// 	return fmt.Errorf("Share dispute didn't complete succesful, exiting ETHDKG")
	// }

	// // setup
	// eth := svcs.eth
	// acct := eth.GetDefaultAccount()
	// c := eth.Contracts()
	// callOpts := eth.GetCallOpts(context.Background(), acct)

	// initialMessage, err := c.Ethdkg.InitialMessage(callOpts)
	// if err != nil {
	// 	logger.Errorf("Can't get initial message: %v", err)
	// 	return ErrCanNotContinue
	// }
	// logger.Infof("InitialMessage: [%v]", string(initialMessage))

	// ethdkg := state.EthDKG
	// encryptedShares := make([][]*big.Int, ethdkg.NumberOfValidators)
	// for _, participant := range ethdkg.Participants {
	// 	pes, present := ethdkg.EncryptedShares[participant.Address]
	// 	idx := participant.Index
	// 	if present && idx >= 0 && idx < ethdkg.NumberOfValidators {
	// 		encryptedShares[participant.Index] = pes
	// 	}
	// }

	// groupPrivateKey, groupPublicKey, groupSignature, err := dkg.GenerateGroupKeys(initialMessage,
	// 	ethdkg.TransportPrivateKey, ethdkg.TransportPublicKey, ethdkg.PrivateCoefficients,
	// 	encryptedShares, ethdkg.Index, ethdkg.Participants, ethdkg.ValidatorThreshold)

	// ethdkg.GroupPrivateKey = groupPrivateKey
	// ethdkg.GroupPublicKey = groupPublicKey

	// //
	// err = svcs.SetBN256PrivateKey(groupPrivateKey.Bytes())
	// if err != nil {
	// 	logger.Errorf("Error in svcs.SetBN256PrivateKey(%x): %v", groupPrivateKey.Bytes(), err)

	// 	return ErrCanNotContinue
	// }

	// task := dkgtasks.NewGPKSubmissionTask(eth.GetDefaultAccount(), ethdkg.TransportPublicKey, groupPublicKey, groupSignature,
	// 	ethdkg.Schedule.RegistrationEnd, ethdkg.Schedule.GPKJSubmissionEnd)

	// state.EthDKG.GPKJSubmissionTH = svcs.taskMan.NewTaskHandler(logger, eth, task)
	// state.EthDKG.GPKJSubmissionTH.Start()

	return nil
}

// DoGroupAccusationGPKj does something
func (svcs *Services) DoGroupAccusationGPKj(state *State, block uint64) error {
	logger := svcs.logger

	logger.Infof(strings.Repeat("-", 60))
	logger.Infof("DoGroupAccusationGPKj()")
	logger.Infof(strings.Repeat("-", 60))

	// Setup
	// eth := svcs.eth
	// acct := eth.GetDefaultAccount()
	// c := eth.Contracts()
	// callOpts := eth.GetCallOpts(context.Background(), acct)

	// //
	// initialMessage, err := c.Ethdkg.InitialMessage(callOpts)
	// if err != nil {
	// 	logger.Errorf("Can't get initial message: %v", err)
	// 	return ErrCanNotContinue
	// }
	// logger.Infof("InitialMessage: [%v]", string(initialMessage))

	// //
	// for idx := 0; idx < state.EthDKG.Participants.Len(); idx++ {

	// }

	// dkg.CategorizeGroupSigners(initialMessage, state.EthDKG.MasterPublicKey, publishedPublicKeys [][4]*big.Int, publishedSignatures [][2]*big.Int, participants dkg.ParticipantList, threshold int)

	return nil
}

// DoSuccessfulCompletion does something
func (svcs *Services) DoSuccessfulCompletion(state *State, block uint64) error {
	logger := svcs.logger

	logger.Infof(strings.Repeat("-", 60))
	logger.Infof("DoSuccessfulCompletion()")
	logger.Infof(strings.Repeat("-", 60))

	// eth := svcs.eth
	// ethdkg := state.EthDKG

	// task := dkgtasks.NewCompletionTask(eth.GetDefaultAccount(), ethdkg.TransportPublicKey,
	// 	ethdkg.Schedule.RegistrationEnd, ethdkg.Schedule.CompleteEnd)

	// ethdkg.CompleteTH = svcs.taskMan.NewTaskHandler(logger, eth, task)
	// ethdkg.CompleteTH.Start()

	return nil
}

func calculateInverseArray(n uint8) ([]*big.Int, error) {
	bigNeg2 := big.NewInt(-2)
	orderMinus2 := new(big.Int).Add(cloudflare.Order, bigNeg2)

	// Get inverse array; this array is required to help keep gas costs down
	// in the smart contract. Modular multiplication is much cheaper than
	// modular inversion (exponentiation).
	invArrayBig := make([]*big.Int, n-1)
	for idx := uint8(0); idx < n-1; idx++ {
		m := big.NewInt(int64(idx + 1))
		mInv := new(big.Int).Exp(m, orderMinus2, cloudflare.Order)
		// Confirm
		res := new(big.Int).Mul(m, mInv)
		res.Mod(res, cloudflare.Order)
		if res.Cmp(big1) != 0 {
			return nil, errors.New("error when computing inverseArray")
		}
		invArrayBig[idx] = mInv
	}
	return invArrayBig, nil
}

func loadSignature(callOpts *bind.CallOpts, ethdkg *bindings.ETHDKG, addr common.Address) (*cloudflare.G1, error) {
	var err error
	var sigBig [2]*big.Int

	sigBig[0], err = ethdkg.InitialSignatures(callOpts, addr, big0)
	if err != nil {
		return nil, err
	}

	sigBig[1], err = ethdkg.InitialSignatures(callOpts, addr, big1)
	if err != nil {
		return nil, err
	}

	sig, err := bn256.BigIntArrayToG1(sigBig)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func loadGroupPublicKey(callOpts *bind.CallOpts, ethdkg *bindings.ETHDKG, addr common.Address) (*cloudflare.G2, error) {
	var err error
	var gpkjBig [4]*big.Int

	gpkjBig[0], err = ethdkg.GpkjSubmissions(callOpts, addr, big0)
	if err != nil {
		return nil, err
	}

	gpkjBig[1], err = ethdkg.GpkjSubmissions(callOpts, addr, big1)
	if err != nil {
		return nil, err
	}

	gpkjBig[2], err = ethdkg.GpkjSubmissions(callOpts, addr, big2)
	if err != nil {
		return nil, err
	}

	gpkjBig[3], err = ethdkg.GpkjSubmissions(callOpts, addr, big3)
	if err != nil {
		return nil, err
	}

	gpkj, err := bn256.BigIntArrayToG2(gpkjBig)
	if err != nil {
		return nil, err
	}

	return gpkj, nil
}
