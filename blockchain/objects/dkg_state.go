package objects

import (
	"encoding/json"
	"errors"
	"math/big"
	"sync"

	"github.com/MadBase/bridge/bindings"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
)

// ErrCanNotContinue standard error if we must drop out of ETHDKG
var (
	ErrCanNotContinue = errors.New("can not continue distributed key generation")
)

// EthDKGState is used to track the state of the ETHDKG
type DkgState struct {
	sync.RWMutex // TODO Make sure this is always used

	// Local validator info
	Account             accounts.Account
	Index               int
	InitialMessage      []byte
	GroupPrivateKey     *big.Int
	GroupPublicKey      [4]*big.Int
	GroupSignature      [2]*big.Int
	MasterPublicKey     [4]*big.Int
	NumberOfValidators  int
	PrivateCoefficients []*big.Int
	SecretValue         *big.Int
	ValidatorThreshold  int
	TransportPrivateKey *big.Int
	TransportPublicKey  [2]*big.Int

	// Remote validator info
	Commitments                 map[common.Address][][2]*big.Int // ShareDistribution Event
	EncryptedShares             map[common.Address][]*big.Int    // "
	DishonestValidatorsIndicies []*big.Int                       //  Calculated for group accusation
	HonestValidatorsIndicies    []*big.Int                       // "
	Inverse                     []*big.Int                       // "
	KeyShareG1s                 map[common.Address][2]*big.Int   // KeyShare Event
	KeyShareG1CorrectnessProofs map[common.Address][2]*big.Int   // "
	KeyShareG2s                 map[common.Address][4]*big.Int   // "
	Participants                ParticipantList                  // Index, Address & PublicKey
	GroupPublicKeys             map[common.Address][4]*big.Int   // Retrieved to validate group keys
	GroupSignatures             map[common.Address][2]*big.Int   // ""

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

func NewDkgState(account accounts.Account) *DkgState {
	return &DkgState{
		Account:                     account,
		Commitments:                 make(map[common.Address][][2]*big.Int),
		EncryptedShares:             make(map[common.Address][]*big.Int),
		GroupPublicKeys:             make(map[common.Address][4]*big.Int),
		GroupSignatures:             make(map[common.Address][2]*big.Int),
		KeyShareG1s:                 make(map[common.Address][2]*big.Int),
		KeyShareG1CorrectnessProofs: make(map[common.Address][2]*big.Int),
		KeyShareG2s:                 make(map[common.Address][4]*big.Int),
	}
}

func (state *DkgState) PopulateSchedule(event *bindings.ETHDKGRegistrationOpen) {

	state.RegistrationStart = event.DkgStarts.Uint64()
	state.RegistrationEnd = event.RegistrationEnds.Uint64()

	state.ShareDistributionStart = state.RegistrationEnd + 1
	state.ShareDistributionEnd = event.ShareDistributionEnds.Uint64()

	state.DisputeStart = state.ShareDistributionEnd + 1
	state.DisputeEnd = event.DisputeEnds.Uint64()

	state.KeyShareSubmissionStart = state.DisputeEnd + 1
	state.KeyShareSubmissionEnd = event.KeyShareSubmissionEnds.Uint64()

	state.MPKSubmissionStart = state.KeyShareSubmissionEnd + 1
	state.MPKSubmissionEnd = event.MpkSubmissionEnds.Uint64()

	state.GPKJSubmissionStart = state.MPKSubmissionEnd + 1
	state.GPKJSubmissionEnd = event.GpkjSubmissionEnds.Uint64()

	state.GPKJGroupAccusationStart = state.GPKJSubmissionEnd + 1
	state.GPKJGroupAccusationEnd = event.GpkjDisputeEnds.Uint64()

	state.CompleteStart = state.GPKJGroupAccusationEnd + 1
	state.CompleteEnd = event.DkgComplete.Uint64()
}

// Participant contains what we know about other participants, i.e. public information
type Participant struct {
	Address   common.Address
	Index     int
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
