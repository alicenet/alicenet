package state

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/crypto/bn256"
	"github.com/alicenet/alicenet/crypto/bn256/cloudflare"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// EthDKGPhase is used to indicate what phase we are currently in.
type EthDKGPhase uint8

// These are the valid phases of ETHDKG.
const (
	RegistrationOpen EthDKGPhase = iota
	ShareDistribution
	DisputeShareDistribution
	KeyShareSubmission
	MPKSubmission
	GPKJSubmission
	DisputeGPKJSubmission
	Completion
)

func (phase EthDKGPhase) String() string {
	return [...]string{
		"RegistrationOpen",
		"ShareDistribution",
		"DisputeShareDistribution",
		"KeyShareSubmission",
		"MPKSubmission",
		"GPKJSubmission",
		"DisputeGPKJSubmission",
		"Completion",
	}[phase]
}

// DkgState is used to track the state of the ETHDKG.
type DkgState struct {
	IsValidator        bool        `json:"isValidator"`
	Phase              EthDKGPhase `json:"phase"`
	PhaseLength        uint64      `json:"phaseLength"`
	ConfirmationLength uint64      `json:"confirmationLength"`
	PhaseStart         uint64      `json:"phaseStart"`

	// Local validator info
	////////////////////////////////////////////////////////////////////////////
	// Account is the Ethereum account corresponding to the Ethereum Public Key
	// of the local Validator
	Account accounts.Account `json:"account"`
	// Index is the Base-1 index of the local Validator which is used
	// during the Share Distribution phase for verifiable secret sharing.
	// REPEAT: THIS IS BASE-1
	Index int `json:"index"`
	// ValidatorAddresses stores all validator addresses at the beginning of ETHDKG
	ValidatorAddresses []common.Address `json:"validatorAddresses"`
	// NumberOfValidators is equal to len(ValidatorAddresses)
	NumberOfValidators int `json:"numberOfValidators"`
	// ETHDKG nonce
	Nonce uint64 `json:"nonce"`
	// ValidatorThreshold is the threshold number of validators for the system.
	// If n = NumberOfValidators and t = threshold, then
	// 			t+1 > 2*n/3
	ValidatorThreshold int `json:"validatorThreshold"`
	// TransportPrivateKey is the private key corresponding to TransportPublicKey.
	TransportPrivateKey *big.Int `json:"transportPrivateKey"`
	// TransportPublicKey is the public key used in EthDKG.
	// This public key is used for secret communication over the open channel
	// of Ethereum.
	TransportPublicKey [2]*big.Int `json:"transportPublicKey"`
	// SecretValue is the secret value which is to be shared during
	// the verifiable secret sharing.
	// The sum of all the secret values of all the participants
	// is the master secret key, the secret key of the master public key
	// (MasterPublicKey)
	SecretValue *big.Int `json:"secretValue"`
	// PrivateCoefficients is the private polynomial which is used to share
	// the shared secret. This is performed via Shamir Secret Sharing.
	PrivateCoefficients []*big.Int `json:"privateCoefficients"`
	// MasterPublicKey is the public key for the entire group.
	// As mentioned above, the secret key called the master secret key
	// and is the sum of all the shared secrets of all the participants.
	MasterPublicKey [4]*big.Int `json:"masterPublicKey"`
	// GroupPrivateKey is the local Validator's portion of the master secret key.
	// This is also denoted gskj.
	GroupPrivateKey *big.Int `json:"groupPrivateKey"`

	// Remote validator info
	////////////////////////////////////////////////////////////////////////////
	// Participants is the list of Validators
	Participants map[common.Address]*Participant `json:"participants"`

	// Group Public Key (GPKj) Accusation Phase
	//////////////////////////////////////////////////
	// DishonestValidatorsIndices stores the list indices of dishonest
	// validators
	DishonestValidators ParticipantList `json:"dishonestValidators"`
	// HonestValidatorsIndices stores the list indices of honest
	// validators
	HonestValidators ParticipantList `json:"honestValidators"`
	// Inverse stores the multiplicative inverses
	// of elements. This may be used in GPKJGroupAccusation logic.
	Inverse []*big.Int `json:"inverse"`
}

// GetSortedParticipants returns the participant list sorted by Index field.
func (state *DkgState) GetSortedParticipants() ParticipantList {
	list := make(ParticipantList, len(state.Participants))

	for _, p := range state.Participants {
		list[p.Index-1] = p
	}

	return list
}

// asserting that DkgState struct implements interface tasks.TaskState.
var _ tasks.TaskState = &DkgState{}

// OnRegistrationOpened processes data from RegistrationOpened event.
func (state *DkgState) OnRegistrationOpened(startBlock, phaseLength, confirmationLength, nonce uint64) {
	state.Phase = RegistrationOpen
	state.PhaseStart = startBlock
	state.PhaseLength = phaseLength
	state.ConfirmationLength = confirmationLength
	state.Nonce = nonce
}

// OnAddressRegistered processes data from AddressRegistered event.
func (state *DkgState) OnAddressRegistered(account common.Address, index int, nonce uint64, publicKey [2]*big.Int) {
	state.Participants[account] = &Participant{
		Address:   account,
		Index:     index,
		PublicKey: publicKey,
		Phase:     RegistrationOpen,
		Nonce:     nonce,
	}

	// update state.Index with my index, if this event was mine
	if account.String() == state.Account.Address.String() {
		state.Index = index
	}
}

// OnRegistrationComplete processes data from RegistrationComplete event.
func (state *DkgState) OnRegistrationComplete(shareDistributionStartBlockNumber uint64) {
	state.Phase = ShareDistribution
	state.PhaseStart = shareDistributionStartBlockNumber + state.ConfirmationLength
}

// OnSharesDistributed processes data from SharesDistributed event.
func (state *DkgState) OnSharesDistributed(logger *logrus.Entry, account common.Address, encryptedShares []*big.Int, commitments [][2]*big.Int) error {
	// compute distributed shares hash
	distributedSharesHash, _, _, err := ComputeDistributedSharesHash(encryptedShares, commitments)
	if err != nil {
		logger.Errorf("ProcessShareDistribution: error calculating distributed shares hash: %v", err)
		return err
	}

	state.Participants[account].Phase = ShareDistribution
	state.Participants[account].DistributedSharesHash = distributedSharesHash
	state.Participants[account].Commitments = commitments
	state.Participants[account].EncryptedShares = encryptedShares

	return nil
}

// OnShareDistributionComplete processes data from ShareDistributionComplete event.
func (state *DkgState) OnShareDistributionComplete(disputeShareDistributionStartBlock uint64) {
	state.Phase = DisputeShareDistribution

	// schedule DisputeShareDistributionTask
	dispShareStartBlock := disputeShareDistributionStartBlock + state.ConfirmationLength
	state.PhaseStart = dispShareStartBlock
}

// OnKeyShareSubmissionComplete processes data from KeyShareSubmissionComplete event.
func (state *DkgState) OnKeyShareSubmissionComplete(mpkSubmissionStartBlock uint64) {
	state.Phase = MPKSubmission
	state.PhaseStart = mpkSubmissionStartBlock + state.ConfirmationLength
}

// OnMPKSet processes data from MPKSet event.
func (state *DkgState) OnMPKSet(gpkjSubmissionStartBlock uint64) {
	state.Phase = GPKJSubmission
	state.PhaseStart = gpkjSubmissionStartBlock
}

// OnGPKJSubmissionComplete processes data from GPKJSubmissionComplete event.
func (state *DkgState) OnGPKJSubmissionComplete(disputeGPKjStartBlock uint64) {
	state.Phase = DisputeGPKJSubmission
	state.PhaseStart = disputeGPKjStartBlock + state.ConfirmationLength
}

// OnKeyShareSubmitted processes data from KeyShareSubmitted event.
func (state *DkgState) OnKeyShareSubmitted(account common.Address, keyShareG1 [2]*big.Int, keyShareG1CorrectnessProof [2]*big.Int, keyShareG2 [4]*big.Int) {
	state.Phase = KeyShareSubmission

	state.Participants[account].Phase = KeyShareSubmission
	state.Participants[account].KeyShareG1s = keyShareG1
	state.Participants[account].KeyShareG1CorrectnessProofs = keyShareG1CorrectnessProof
	state.Participants[account].KeyShareG2s = keyShareG2
}

// OnGPKjSubmitted processes data from GPKjSubmitted event.
func (state *DkgState) OnGPKjSubmitted(account common.Address, gpkj [4]*big.Int) {
	state.Participants[account].GPKj = gpkj
	state.Participants[account].Phase = GPKJSubmission
}

// OnCompletion processes data from ValidatorSetCompleted event.
func (state *DkgState) OnCompletion() {
	state.Phase = Completion
}

// NewDkgState makes a new DkgState object.
func NewDkgState(account accounts.Account) *DkgState {
	return &DkgState{
		Account:      account,
		Participants: make(map[common.Address]*Participant),
	}
}

func GetDkgState(monDB *db.Database) (*DkgState, error) {
	dkgState := &DkgState{}
	err := monDB.View(func(txn *badger.Txn) error {
		return dkgState.LoadState(txn)
	})
	if err != nil {
		return nil, err
	}
	return dkgState, nil
}

func SaveDkgState(monDB *db.Database, dkgState *DkgState) error {
	err := monDB.Update(func(txn *badger.Txn) error {
		return dkgState.PersistState(txn)
	})
	if err != nil {
		return err
	}
	if err = monDB.Sync(); err != nil {
		return fmt.Errorf("Failed to set sync of dkgState: %v", err)
	}
	return nil
}

func (state *DkgState) PersistState(txn *badger.Txn) error {
	logger := logging.GetLogger("staterecover").WithField("State", "dkgState")
	rawData, err := json.Marshal(state)
	if err != nil {
		return err
	}

	key := dbprefix.PrefixEthereumDKGState()
	logger.WithField("Key", string(key)).Debug("Saving state in the database")
	if err = utils.SetValue(txn, key, rawData); err != nil {
		return err
	}

	return nil
}

func (state *DkgState) LoadState(txn *badger.Txn) error {
	logger := logging.GetLogger("staterecover").WithField("State", "dkgState")
	key := dbprefix.PrefixEthereumDKGState()
	logger.WithField("Key", string(key)).Debug("Loading state from database")
	rawData, err := utils.GetValue(txn, key)
	if err != nil {
		return err
	}

	err = json.Unmarshal(rawData, state)
	if err != nil {
		return err
	}

	return nil
}

// Participant contains what we know about other participants, i.e. public information.
type Participant struct {
	// Address is the Ethereum address corresponding to the Ethereum Public Key
	// for the Participant.
	Address common.Address `json:"address"`
	// Index is the Base-1 index of the participant.
	// This is used during the Share Distribution phase to perform
	// verifyiable secret sharing.
	// REPEAT: THIS IS BASE-1
	Index int `json:"index"`
	// PublicKey is the TransportPublicKey of Participant.
	PublicKey [2]*big.Int `json:"publicKey"`
	Nonce     uint64      `json:"nonce"`
	Phase     EthDKGPhase `json:"phase"`

	// Share Distribution Phase
	//////////////////////////////////////////////////
	// Commitments stores the Public Coefficients
	// corresponding to public polynomial
	// in Shamir Secret Sharing protocol.
	// The first coefficient (constant term) is the public commitment
	// corresponding to the secret share (SecretValue).
	Commitments [][2]*big.Int `json:"commitments"`
	// EncryptedShares are the encrypted secret shares
	// in the Shamir Secret Sharing protocol.
	EncryptedShares       []*big.Int `json:"encryptedShares"`
	DistributedSharesHash [32]byte   `json:"distributedSharesHash"`

	CommitmentsFirstCoefficient [2]*big.Int `json:"commitmentsFirstCoefficient"`

	// Key Share Submission Phase
	//////////////////////////////////////////////////
	// KeyShareG1s stores the key shares of G1 element
	// for each participant
	KeyShareG1s [2]*big.Int `json:"keyShareG1s"`

	// KeyShareG1CorrectnessProofs stores the proofs of each
	// G1 element for each participant.
	KeyShareG1CorrectnessProofs [2]*big.Int `json:"keyShareG1CorrectnessProofs"`

	// KeyShareG2s stores the key shares of G2 element
	// for each participant.
	// Adding all the G2 shares together produces the
	// master public key (MasterPublicKey).
	KeyShareG2s [4]*big.Int `json:"keyShareG2s"`

	// GPKj is the local Validator's portion of the master public key.
	// This is also denoted GroupPublicKey.
	GPKj [4]*big.Int `json:"gpkj"`
}

// ParticipantList is a required type alias since the Sort interface is awful.
type ParticipantList []*Participant

// Simplify logging.
func (p *Participant) String() string {
	out, err := json.Marshal(p)
	if err != nil {
		return err.Error()
	}

	return string(out)
}

// Copy makes returns a copy of Participant.
func (p *Participant) Copy() *Participant {
	c := &Participant{}
	c.Index = p.Index
	c.PublicKey = [2]*big.Int{new(big.Int).Set(p.PublicKey[0]), new(big.Int).Set(p.PublicKey[1])}
	addrBytes := p.Address.Bytes()
	c.Address.SetBytes(addrBytes)
	return c
}

// ExtractIndices returns the list of indices of ParticipantList.
func (pl ParticipantList) ExtractIndices() []int {
	indices := []int{}
	for k := 0; k < len(pl); k++ {
		indices = append(indices, pl[k].Index)
	}
	return indices
}

// Len returns the len of the collection.
func (pl ParticipantList) Len() int {
	return len(pl)
}

// Less decides if element i is 'Less' than element j -- less ~= before.
func (pl ParticipantList) Less(i, j int) bool {
	return pl[i].Index < pl[j].Index
}

// Swap swaps elements i and j within the collection.
func (pl ParticipantList) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}

// CategorizeGroupSigners returns 0 based indices of honest participants, 0 based indices of dishonest participants.
func CategorizeGroupSigners(publishedPublicKeys [][4]*big.Int, participants ParticipantList, commitments [][][2]*big.Int) (ParticipantList, ParticipantList, ParticipantList, error) {
	// Setup + sanity checks before starting
	n := len(participants)
	threshold := ThresholdForUserCount(n)

	good := ParticipantList{}
	bad := ParticipantList{}
	missing := ParticipantList{}

	// len(publishedPublicKeys) must equal len(publishedSignatures) must equal len(participants)
	if n != len(publishedPublicKeys) || n != len(commitments) {
		return ParticipantList{}, ParticipantList{}, ParticipantList{}, fmt.Errorf(
			"mismatched public keys (%v), participants (%v), commitments (%v)", len(publishedPublicKeys), n, len(commitments))
	}

	// Require each commitment has length threshold+1
	for k := 0; k < n; k++ {
		if len(commitments[k]) != threshold+1 {
			return ParticipantList{}, ParticipantList{}, ParticipantList{}, fmt.Errorf(
				"invalid commitments: required (%v); actual (%v)", threshold+1, len(commitments[k]))
		}
	}

	// We need commitments.
	// 		For each participant, loop through and form gpkj* term.
	//		Perform a PairingCheck to ensure valid gpkj.
	//		If invalid, add to bad list.

	g1Base := new(cloudflare.G1).ScalarBaseMult(common.Big1)
	orderMinus1 := new(big.Int).Sub(cloudflare.Order, common.Big1)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	// commitments:
	//		First dimension is participant index;
	//		Second dimension is commitment number
	for idx := 0; idx < n; idx++ {
		// Loop through all participants to confirm each is valid
		participant := participants[idx]

		// If public key is all zeros, then no public key was submitted;
		// add to missing.
		big0 := big.NewInt(0)
		if (publishedPublicKeys[idx][0] == nil ||
			publishedPublicKeys[idx][1] == nil ||
			publishedPublicKeys[idx][2] == nil ||
			publishedPublicKeys[idx][3] == nil) || (publishedPublicKeys[idx][0].Cmp(big0) == 0 &&
			publishedPublicKeys[idx][1].Cmp(big0) == 0 &&
			publishedPublicKeys[idx][2].Cmp(big0) == 0 &&
			publishedPublicKeys[idx][3].Cmp(big0) == 0) {
			missing = append(missing, participant.Copy())
			continue
		}

		j := participant.Index // participant index
		jBig := big.NewInt(int64(j))

		tmp0 := new(cloudflare.G1)
		gpkj, err := bn256.BigIntArrayToG2(publishedPublicKeys[idx])
		if err != nil {
			return ParticipantList{}, ParticipantList{}, ParticipantList{}, fmt.Errorf("error converting BigIntArray to G2: %v", err)
		}

		// Outer loop determines what needs to be exponentiated
		for polyDegreeIdx := 0; polyDegreeIdx <= threshold; polyDegreeIdx++ {
			tmp1 := new(cloudflare.G1)
			// Inner loop loops through participants
			for participantIdx := 0; participantIdx < n; participantIdx++ {
				tmp2Big := commitments[participantIdx][polyDegreeIdx]
				tmp2, err := bn256.BigIntArrayToG1(tmp2Big)
				if err != nil {
					return ParticipantList{}, ParticipantList{}, ParticipantList{}, fmt.Errorf("error converting BigIntArray to G1: %v", err)
				}
				tmp1.Add(tmp1, tmp2)
			}
			polyDegreeIdxBig := big.NewInt(int64(polyDegreeIdx))
			exponent := new(big.Int).Exp(jBig, polyDegreeIdxBig, cloudflare.Order)
			tmp1.ScalarMult(tmp1, exponent)

			tmp0.Add(tmp0, tmp1)
		}

		gpkjStar := new(cloudflare.G1).Set(tmp0)
		validPair := cloudflare.PairingCheck([]*cloudflare.G1{gpkjStar, g1Base}, []*cloudflare.G2{h2Neg, gpkj})
		if validPair {
			good = append(good, participant.Copy())
		} else {
			bad = append(bad, participant.Copy())
		}
	}

	return good, bad, missing, nil
}
