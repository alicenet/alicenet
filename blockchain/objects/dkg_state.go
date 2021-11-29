package objects

import (
	"encoding/json"
	"errors"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
)

// ErrCanNotContinue standard error if we must drop out of ETHDKG
var (
	ErrCanNotContinue = errors.New("can not continue distributed key generation")
)

// DkgState is used to track the state of the ETHDKG
type DkgState struct {
	sync.RWMutex

	// Local validator info
	////////////////////////////////////////////////////////////////////////////
	// Account is the Ethereum account corresponding to the Ethereum Public Key
	// of the local Validator
	Account accounts.Account
	// Index is the Base-1 index of the local Validator which is used
	// during the Share Distribution phase for verifiable secret sharing.
	// REPEAT: THIS IS BASE-1
	Index int
	// NumberOfValidators is the total number of validators
	NumberOfValidators int
	// ValidatorThreshold is the threshold number of validators for the system.
	// If n = NumberOfValidators and t = threshold, then
	// 			t+1 > 2*n/3
	ValidatorThreshold int
	// TransportPrivateKey is the private key corresponding to TransportPublicKey.
	TransportPrivateKey *big.Int
	// TransportPublicKey is the public key used in EthDKG.
	// This public key is used for secret communication over the open channel
	// of Ethereum.
	TransportPublicKey [2]*big.Int
	// SecretValue is the secret value which is to be shared during
	// the verifiable secret sharing.
	// The sum of all the secret values of all the participants
	// is the master secret key, the secret key of the master public key
	// (MasterPublicKey)
	SecretValue *big.Int
	// PrivateCoefficients is the private polynomial which is used to share
	// the shared secret. This is performed via Shamir Secret Sharing.
	PrivateCoefficients []*big.Int
	// MasterPublicKey is the public key for the entire group.
	// As mentioned above, the secret key called the master secret key
	// and is the sum of all the shared secrets of all the participants.
	MasterPublicKey [4]*big.Int
	// InitialMessage is a message which is signed to help ensure
	// valid group public key submission.
	InitialMessage []byte
	// GroupPrivateKey is the local Validator's portion of the master secret key.
	// This is also denoted gskj.
	GroupPrivateKey *big.Int
	// GroupPublicKey is the local Validator's portion of the master public key.
	// This is also denoted gpkj.
	GroupPublicKey [4]*big.Int
	// GroupSignature is the signature of InitialMessage corresponding to
	// GroupPublicKey. The smart contract logic verifies that GroupSignature
	// is a valid signature of GroupPublicKey.
	// This may be used in the GroupSignature validation logic.
	GroupSignature [2]*big.Int

	// Remote validator info
	////////////////////////////////////////////////////////////////////////////
	// Participants is the list of Validators
	Participants ParticipantList // Index, Address & PublicKey

	// Share Distribution Phase
	//////////////////////////////////////////////////
	// Commitments stores the Public Coefficients
	// corresponding to public polynomial
	// in Shamir Secret Sharing protocol.
	// The first coefficient (constant term) is the public commitment
	// corresponding to the secret share (SecretValue).
	Commitments map[common.Address][][2]*big.Int
	// EncryptedShares are the encrypted secret shares
	// in the Shamir Secret Sharing protocol.
	EncryptedShares map[common.Address][]*big.Int

	// Share Dispute Phase
	//////////////////////////////////////////////////
	// These are the participants with bad shares
	BadShares map[common.Address]*Participant

	// Key Share Submission Phase
	//////////////////////////////////////////////////
	// KeyShareG1s stores the key shares of G1 element
	// for each participant
	KeyShareG1s map[common.Address][2]*big.Int
	// KeyShareG1CorrectnessProofs stores the proofs of each
	// G1 element for each participant.
	KeyShareG1CorrectnessProofs map[common.Address][2]*big.Int
	// KeyShareG2s stores the key shares of G2 element
	// for each participant.
	// Adding all the G2 shares together produces the
	// master public key (MasterPublicKey).
	KeyShareG2s map[common.Address][4]*big.Int

	// Group Public Key (GPKj) Submission Phase
	//////////////////////////////////////////////////
	// GroupPublicKeys stores the group public keys (gpkj)
	// for each participant.
	GroupPublicKeys map[common.Address][4]*big.Int // Retrieved to validate group keys
	// GroupSignatures stores the group signatures
	// for each participant.
	GroupSignatures map[common.Address][2]*big.Int // "

	// Group Public Key (GPKj) Accusation Phase
	//////////////////////////////////////////////////
	// DishonestValidatorsIndices stores the list indices of dishonest
	// validators
	DishonestValidators ParticipantList // Calculated for group accusation
	// HonestValidatorsIndices stores the list indices of honest
	// validators
	HonestValidators ParticipantList // "
	// Inverse stores the multiplicative inverses
	// of elements. This may be used in GPKJGroupAccusation logic.
	Inverse []*big.Int // "

	// Flags indicating phase success
	Registration        bool
	ShareDistribution   bool
	Dispute             bool
	KeyShareSubmission  bool
	MPKSubmission       bool
	GPKJSubmission      bool
	GPKJGroupAccusation bool
	Complete            bool

	// Phase schedule
	RegistrationStart        uint64
	RegistrationEnd          uint64
	ShareDistributionStart   uint64
	ShareDistributionEnd     uint64
	DisputeStart             uint64
	DisputeEnd               uint64
	KeyShareSubmissionStart  uint64
	KeyShareSubmissionEnd    uint64
	MPKSubmissionStart       uint64
	MPKSubmissionEnd         uint64
	GPKJSubmissionStart      uint64
	GPKJSubmissionEnd        uint64
	GPKJGroupAccusationStart uint64
	GPKJGroupAccusationEnd   uint64
	CompleteStart            uint64
	CompleteEnd              uint64
}

// NewDkgState makes a new DkgState object
func NewDkgState(account accounts.Account) *DkgState {
	return &DkgState{
		Account:                     account,
		BadShares:                   make(map[common.Address]*Participant),
		Commitments:                 make(map[common.Address][][2]*big.Int),
		EncryptedShares:             make(map[common.Address][]*big.Int),
		GroupPublicKeys:             make(map[common.Address][4]*big.Int),
		GroupSignatures:             make(map[common.Address][2]*big.Int),
		KeyShareG1s:                 make(map[common.Address][2]*big.Int),
		KeyShareG1CorrectnessProofs: make(map[common.Address][2]*big.Int),
		KeyShareG2s:                 make(map[common.Address][4]*big.Int),
	}
}

// Participant contains what we know about other participants, i.e. public information
type Participant struct {
	// Address is the Ethereum address corresponding to the Ethereum Public Key
	// for the Participant.
	Address common.Address
	// Index is the Base-1 index of the participant.
	// This is used during the Share Distribution phase to perform
	// verifyiable secret sharing.
	// REPEAT: THIS IS BASE-1
	Index int
	// PublicKey is the TransportPublicKey of Participant.
	PublicKey [2]*big.Int
}

// ParticipantList is a required type alias since the Sort interface is awful
type ParticipantList []*Participant

// Simplify logging
func (p *Participant) String() string {
	out, err := json.Marshal(p)
	if err != nil {
		return err.Error()
	}

	return string(out)
}

// Copy makes returns a copy of Participant
func (p *Participant) Copy() *Participant {
	c := &Participant{}
	c.Index = p.Index
	c.PublicKey = [2]*big.Int{new(big.Int).Set(p.PublicKey[0]), new(big.Int).Set(p.PublicKey[1])}
	addrBytes := p.Address.Bytes()
	c.Address.SetBytes(addrBytes)
	return c
}

// ExtractIndices returns the list of indices of ParticipantList
func (pl ParticipantList) ExtractIndices() []int {
	indices := []int{}
	for k := 0; k < len(pl); k++ {
		indices = append(indices, pl[k].Index)
	}
	return indices
}

// Len returns the len of the collection
func (pl ParticipantList) Len() int {
	return len(pl)
}

// Less decides if element i is 'Less' than element j -- less ~= before
func (pl ParticipantList) Less(i, j int) bool {
	return pl[i].Index < pl[j].Index
}

// Swap swaps elements i and j within the collection
func (pl ParticipantList) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}
