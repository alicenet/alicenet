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

	Phase              EthDKGPhase
	PhaseLength        uint64
	ConfirmationLength uint64
	PhaseStart         uint64

	// Local validator info
	////////////////////////////////////////////////////////////////////////////
	// Account is the Ethereum account corresponding to the Ethereum Public Key
	// of the local Validator
	Account accounts.Account
	// Index is the Base-1 index of the local Validator which is used
	// during the Share Distribution phase for verifiable secret sharing.
	// REPEAT: THIS IS BASE-1
	Index int
	// ValidatorAddresses stores all validator addresses at the beginning of ETHDKG
	ValidatorAddresses []common.Address
	// NumberOfValidators is equal to len(ValidatorAddresses)
	NumberOfValidators int
	// ETHDKG nonce
	Nonce uint64
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
	// GroupPrivateKey is the local Validator's portion of the master secret key.
	// This is also denoted gskj.
	GroupPrivateKey *big.Int

	// Remote validator info
	////////////////////////////////////////////////////////////////////////////
	// Participants is the list of Validators
	Participants map[common.Address]*Participant // Index, Address & PublicKey

	// Share Dispute Phase
	//////////////////////////////////////////////////
	// These are the participants with bad shares
	BadShares map[common.Address]*Participant

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
}

// GetSortedParticipants returns the participant list sorted by Index field
func (dkg *DkgState) GetSortedParticipants() ParticipantList {
	var list = make(ParticipantList, len(dkg.Participants))

	for _, p := range dkg.Participants {
		list[p.Index-1] = p
	}

	return list
}

// GetSortedParticipants returns the participant list sorted by Index field
func (dkg *DkgState) OnGPKjSubmitted(account common.Address, gpkj [4]*big.Int) {
	dkg.Participants[account].GPKj = gpkj
	dkg.Participants[account].Phase = uint8(GPKJSubmission)
}

// NewDkgState makes a new DkgState object
func NewDkgState(account accounts.Account) *DkgState {
	return &DkgState{
		Account:      account,
		BadShares:    make(map[common.Address]*Participant),
		Participants: make(map[common.Address]*Participant),
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
	Nonce     uint64
	Phase     uint8

	// Share Distribution Phase
	//////////////////////////////////////////////////
	// Commitments stores the Public Coefficients
	// corresponding to public polynomial
	// in Shamir Secret Sharing protocol.
	// The first coefficient (constant term) is the public commitment
	// corresponding to the secret share (SecretValue).
	Commitments [][2]*big.Int
	// EncryptedShares are the encrypted secret shares
	// in the Shamir Secret Sharing protocol.
	EncryptedShares       []*big.Int
	DistributedSharesHash [32]byte

	CommitmentsFirstCoefficient [2]*big.Int

	// todo: delete this
	//KeyShares [2]*big.Int

	// Key Share Submission Phase
	//////////////////////////////////////////////////
	// KeyShareG1s stores the key shares of G1 element
	// for each participant
	KeyShareG1s [2]*big.Int

	// KeyShareG1CorrectnessProofs stores the proofs of each
	// G1 element for each participant.
	KeyShareG1CorrectnessProofs [2]*big.Int

	// KeyShareG2s stores the key shares of G2 element
	// for each participant.
	// Adding all the G2 shares together produces the
	// master public key (MasterPublicKey).
	KeyShareG2s [4]*big.Int

	// GPKj is the local Validator's portion of the master public key.
	// This is also denoted GroupPublicKey.
	GPKj [4]*big.Int
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
