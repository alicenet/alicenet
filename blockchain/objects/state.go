package objects

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/ethereum/go-ethereum/common"
)

// MonitorState contains info required to monitor Ethereum
type MonitorState struct {
	sync.RWMutex           `json:"-"`
	Version                uint8                   `json:"version"`
	CommunicationFailures  uint32                  `json:"communicationFailtures"`
	EthereumInSync         bool                    `json:"ethereumInSync"`
	HighestBlockProcessed  uint64                  `json:"highestBlockProcessed"`
	HighestBlockFinalized  uint64                  `json:"highestBlockFinalized"`
	HighestEpochProcessed  uint32                  `json:"highestEpochProcessed"`
	HighestEpochSeen       uint32                  `json:"highestEpochSeen"`
	InSync                 bool                    `json:"inSync"`
	LatestDepositProcessed uint32                  `json:"latestDepositProcessed"`
	LatestDepositSeen      uint32                  `json:"latestDepositSeen"`
	PeerCount              uint32                  `json:"peerCount"`
	ValidatorSets          map[uint32]ValidatorSet `json:"validatorSets"`
	Validators             map[uint32][]Validator  `json:"validators"`
	Schedule               interfaces.Schedule     `json:"schedule"`
	EthDKG                 *DkgState               `json:"ethDKG"`
}

// EthDKGPhase is used to indicate what phase we are currently in
type EthDKGPhase int

// These are the valid phases of ETHDKG
const (
	Registration EthDKGPhase = iota
	ShareDistribution
	Dispute
	KeyShareSubmission
	MPKSubmission
	GPKJSubmission
	GPKJGroupAccusation
)

// EthDKGSchedule RegistrationOpen event publishes phase schedule, so we record that here
// type EthDKGSchedule struct {
// 	RegistrationStart        uint64
// 	RegistrationEnd          uint64
// 	ShareDistributionStart   uint64
// 	ShareDistributionEnd     uint64
// 	DisputeStart             uint64
// 	DisputeEnd               uint64
// 	KeyShareSubmissionStart  uint64
// 	KeyShareSubmissionEnd    uint64
// 	MPKSubmissionStart       uint64
// 	MPKSubmissionEnd         uint64
// 	GPKJSubmissionStart      uint64
// 	GPKJSubmissionEnd        uint64
// 	GPKJGroupAccusationStart uint64
// 	GPKJGroupAccusationEnd   uint64
// 	CompleteStart            uint64
// 	CompleteEnd              uint64
// }

// EthDKGState is used to track the state of the ETHDKG
// type EthDKGState struct {

// 	// Local validator info
// 	Address             common.Address
// 	Index               int
// 	GroupPrivateKey     *big.Int
// 	GroupPublicKey      [4]*big.Int
// 	MasterPublicKey     [4]*big.Int
// 	NumberOfValidators  int
// 	PrivateCoefficients []*big.Int
// 	Schedule            *EthDKGSchedule
// 	SecretValue         *big.Int
// 	ValidatorThreshold  int
// 	Tasks               *SequentialSchedule
// 	TransportPrivateKey *big.Int
// 	TransportPublicKey  [2]*big.Int

// 	// Remote validator info
// 	Commitments                 map[common.Address][][2]*big.Int // ShareDistribution Event
// 	EncryptedShares             map[common.Address][]*big.Int    // "
// 	KeyShareG1s                 map[common.Address][2]*big.Int   // KeyShare Event
// 	KeyShareG1CorrectnessProofs map[common.Address][2]*big.Int   // "
// 	KeyShareG2s                 map[common.Address][4]*big.Int   // "
// 	Participants                dkg.ParticipantList              // Index, Address & PublicKey

// 	// Handlers
// 	RegistrationTH        tasks.TaskHandler
// 	ShareDistributionTH   tasks.TaskHandler
// 	DisputeTH             tasks.TaskHandler
// 	KeyShareSubmissionTH  tasks.TaskHandler
// 	MPKSubmissionTH       tasks.TaskHandler
// 	GPKJSubmissionTH      tasks.TaskHandler
// 	GPKJGroupAccusationTH tasks.TaskHandler
// 	CompleteTH            tasks.TaskHandler
// }

// NewEthDKGState creates a new EthDKGState with maps initialized
// func NewEthDKGState() *EthDKGState {
// 	return &EthDKGState{
// 		Commitments:                 make(map[common.Address][][2]*big.Int),
// 		EncryptedShares:             make(map[common.Address][]*big.Int),
// 		KeyShareG1s:                 make(map[common.Address][2]*big.Int),
// 		KeyShareG1CorrectnessProofs: make(map[common.Address][2]*big.Int),
// 		KeyShareG2s:                 make(map[common.Address][4]*big.Int),
// 		Schedule:                    &EthDKGSchedule{},
// 	}
// }

// ValidatorSet is summary information about a ValidatorSet
type ValidatorSet struct {
	ValidatorCount        uint8
	GroupKey              [4]big.Int
	NotBeforeMadNetHeight uint32
}

// Validator contains information about a Validator
type Validator struct {
	Account   common.Address
	Index     uint8
	SharedKey [4]big.Int
}

// Share is temporary storage of shares coming from validators
type Share struct {
	Issuer          common.Address
	Commitments     [][2]*big.Int
	EncryptedShares []*big.Int
}

func NewMonitorState(dkgState *DkgState, schedule interfaces.Schedule) *MonitorState {
	return &MonitorState{
		EthDKG:        dkgState,
		Schedule:      schedule,
		ValidatorSets: make(map[uint32]ValidatorSet),
		Validators:    make(map[uint32][]Validator),
	}
}

func (s *MonitorState) String() string {
	s.RLock()
	defer s.RUnlock()

	str, err := json.Marshal(s)
	if err != nil {
		return fmt.Sprintf("%#v", s)
	}

	return string(str)
}

// Clone builds a deep copy of a state
func (s *MonitorState) Clone() *MonitorState {
	ns := NewMonitorState(s.EthDKG, s.Schedule)

	ns.HighestBlockFinalized = s.HighestBlockFinalized
	ns.HighestBlockProcessed = s.HighestBlockProcessed
	ns.HighestEpochProcessed = s.HighestEpochProcessed
	ns.HighestEpochSeen = s.HighestEpochSeen
	ns.InSync = s.InSync
	ns.EthereumInSync = s.EthereumInSync

	return ns
}

// Diff builds a textual description between states
func (s *MonitorState) Diff(o *MonitorState) string {
	s.RLock()
	defer s.RUnlock()

	o.RLock()
	defer o.RUnlock()

	d := []string{}

	if s.InSync != o.InSync {
		d = append(d, fmt.Sprintf("InSync: %v -> %v", s.InSync, o.InSync))
	}

	if s.HighestBlockProcessed != o.HighestBlockProcessed {
		d = append(d, fmt.Sprintf("HighestBlockProcessed: %v -> %v", s.HighestBlockProcessed, o.HighestBlockProcessed))
	}

	if s.HighestBlockFinalized != o.HighestBlockFinalized {
		d = append(d, fmt.Sprintf("HighestBlockFinalized: %v -> %v", s.HighestBlockFinalized, o.HighestBlockFinalized))
	}

	if s.HighestEpochProcessed != o.HighestEpochProcessed {
		d = append(d, fmt.Sprintf("HighestEpochProcessed: %v -> %v", s.HighestEpochProcessed, o.HighestEpochProcessed))
	}

	if s.HighestEpochSeen != o.HighestEpochSeen {
		d = append(d, fmt.Sprintf("HighestEpochSeen: %v -> %v", s.HighestEpochSeen, o.HighestEpochSeen))
	}

	if s.LatestDepositProcessed != o.LatestDepositProcessed {
		d = append(d, fmt.Sprintf("LatestDepositProcessed: %v -> %v", s.LatestDepositProcessed, o.LatestDepositProcessed))
	}

	if s.LatestDepositSeen != o.LatestDepositSeen {
		d = append(d, fmt.Sprintf("LatestDepositSeen: %v -> %v", s.LatestDepositSeen, o.LatestDepositSeen))
	}

	if s.EthereumInSync != o.EthereumInSync {
		d = append(d, fmt.Sprintf("EthereumInSync: %v -> %v", s.EthereumInSync, o.EthereumInSync))
	}

	if s.CommunicationFailures != o.CommunicationFailures {
		d = append(d, fmt.Sprintf("CommunicationFailures: %v -> %v", s.CommunicationFailures, o.CommunicationFailures))
	}

	return strings.Join(d, ", ")
}
