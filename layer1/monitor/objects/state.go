package objects

import (
	"encoding/json"
	"fmt"
	"github.com/alicenet/alicenet/bridge/bindings"
	"math/big"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
)

// MonitorState contains info required to monitor Ethereum.
type MonitorState struct {
	sync.RWMutex           `json:"-"`
	Version                uint8                                 `json:"version"`
	CommunicationFailures  uint32                                `json:"communicationFailures"`
	EthereumInSync         bool                                  `json:"-"`
	HighestBlockProcessed  uint64                                `json:"highestBlockProcessed"`
	HighestBlockFinalized  uint64                                `json:"highestBlockFinalized"`
	HighestEpochProcessed  uint32                                `json:"highestEpochProcessed"`
	HighestEpochSeen       uint32                                `json:"highestEpochSeen"`
	EndpointInSync         bool                                  `json:"-"`
	LatestDepositProcessed uint32                                `json:"latestDepositProcessed"`
	LatestDepositSeen      uint32                                `json:"latestDepositSeen"`
	PeerCount              uint32                                `json:"peerCount"`
	IsInitialized          bool                                  `json:"-"`
	ValidatorSets          map[uint32]ValidatorSet               `json:"validatorSets"`
	Validators             map[uint32][]Validator                `json:"validators"`
	PotentialValidators    map[common.Address]PotentialValidator `json:"potentialValidators"`
	CanonicalVersion       bindings.CanonicalVersion             `json:"canonicalVersion"`
}

// ValidatorSet is summary information about a ValidatorSet that participated on ETHDKG.
type ValidatorSet struct {
	ValidatorCount          uint8       `json:"validator_count"`
	GroupKey                [4]*big.Int `json:"group_key"`
	NotBeforeAliceNetHeight uint32      `json:"not_before_mad_net_height"`
}

// Validator contains information about a Validator that participated on ETHDKG.
type Validator struct {
	Account   common.Address `json:"account"`
	Index     uint8          `json:"index"`
	SharedKey [4]*big.Int    `json:"shared_key"`
}

// Potential Validator contains information about a validators that entered the
// pool, but might not participated on ETHDKG yet.
type PotentialValidator struct {
	Account common.Address `json:"account"`
	TokenID uint64         `json:"tokenID"`
}

func NewMonitorState() *MonitorState {
	return &MonitorState{
		ValidatorSets:       make(map[uint32]ValidatorSet),
		Validators:          make(map[uint32][]Validator),
		PotentialValidators: make(map[common.Address]PotentialValidator),
	}
}

// Get a copy of the monitor state that is saved on disk.
func GetMonitorState(db *db.Database) (*MonitorState, error) {
	monState := NewMonitorState()
	err := monState.LoadState(db)
	if err != nil {
		return nil, err
	}
	return monState, nil
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

// Clone builds a deep copy of a small portion of state
// TODO Make this create a complete clone of state.
func (s *MonitorState) Clone() *MonitorState {
	ns := NewMonitorState()

	ns.CommunicationFailures = s.CommunicationFailures
	ns.EthereumInSync = s.EthereumInSync
	ns.HighestBlockFinalized = s.HighestBlockFinalized
	ns.HighestBlockProcessed = s.HighestBlockProcessed
	ns.HighestEpochProcessed = s.HighestEpochProcessed
	ns.HighestEpochSeen = s.HighestEpochSeen
	ns.EndpointInSync = s.EndpointInSync
	ns.LatestDepositProcessed = s.LatestDepositProcessed
	ns.LatestDepositSeen = s.LatestDepositSeen
	ns.PeerCount = s.PeerCount

	return ns
}

func (s *MonitorState) LoadState(db *db.Database) error {
	logger := logging.GetLogger("staterecover").WithField("State", "monitorState")

	s.Lock()
	defer s.Unlock()

	if err := db.View(func(txn *badger.Txn) error {
		key := dbprefix.PrefixMonitorState()
		logger.WithField("Key", string(key)).Debug("Loading state from database")
		rawData, err := utils.GetValue(txn, key)
		if err != nil {
			return err
		}

		err = json.Unmarshal(rawData, s)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (mon *MonitorState) PersistState(db *db.Database) error {
	logger := logging.GetLogger("staterecover").WithField("State", "monitorState")

	mon.Lock()
	defer mon.Unlock()

	rawData, err := json.Marshal(mon)
	if err != nil {
		return err
	}

	err = db.Update(func(txn *badger.Txn) error {
		key := dbprefix.PrefixMonitorState()
		logger.WithField("Key", string(key)).Debug("Saving state in the database")
		if err := utils.SetValue(txn, key, rawData); err != nil {
			logger.Error("Failed to set Value")
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := db.Sync(); err != nil {
		logger.Error("Failed to set sync")
		return err
	}

	return nil
}

// Diff builds a textual description between states.
func (s *MonitorState) Diff(o *MonitorState) (string, bool) {
	s.RLock()
	defer s.RUnlock()

	o.RLock()
	defer o.RUnlock()

	d := []string{}
	shouldWrite := false
	if s.CommunicationFailures != o.CommunicationFailures {
		d = append(d, fmt.Sprintf("CommunicationFailures: %v -> %v", s.CommunicationFailures, o.CommunicationFailures))
	}

	if s.EthereumInSync != o.EthereumInSync {
		d = append(d, fmt.Sprintf("EthereumInSync: %v -> %v", s.EthereumInSync, o.EthereumInSync))
	}

	if s.HighestBlockFinalized != o.HighestBlockFinalized {
		d = append(d, fmt.Sprintf("HighestBlockFinalized: %v -> %v", s.HighestBlockFinalized, o.HighestBlockFinalized))
	}

	if s.HighestBlockProcessed != o.HighestBlockProcessed {
		// if we are syncing Ethereum blocks,
		// only write intermittent state diffs on blocks
		// with no other changes of concern
		shouldWrite = true
		d = append(d, fmt.Sprintf("HighestBlockProcessed: %v -> %v", s.HighestBlockProcessed, o.HighestBlockProcessed))
	}

	if s.HighestEpochProcessed != o.HighestEpochProcessed {
		shouldWrite = true
		d = append(d, fmt.Sprintf("HighestEpochProcessed: %v -> %v", s.HighestEpochProcessed, o.HighestEpochProcessed))
	}

	if s.HighestEpochSeen != o.HighestEpochSeen {
		shouldWrite = true
		d = append(d, fmt.Sprintf("HighestEpochSeen: %v -> %v", s.HighestEpochSeen, o.HighestEpochSeen))
	}

	if s.EndpointInSync != o.EndpointInSync {
		d = append(d, fmt.Sprintf("EndpointInSync: %v -> %v", s.EndpointInSync, o.EndpointInSync))
	}

	if s.LatestDepositProcessed != o.LatestDepositProcessed {
		shouldWrite = true
		d = append(d, fmt.Sprintf("LatestDepositProcessed: %v -> %v", s.LatestDepositProcessed, o.LatestDepositProcessed))
	}

	if s.LatestDepositSeen != o.LatestDepositSeen {
		shouldWrite = true
		d = append(d, fmt.Sprintf("LatestDepositSeen: %v -> %v", s.LatestDepositSeen, o.LatestDepositSeen))
	}

	return strings.Join(d, ", "), shouldWrite
}
