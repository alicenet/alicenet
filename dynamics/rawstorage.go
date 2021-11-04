package dynamics

import (
	"encoding/json"
	"math/big"
	"time"
)

// TODO: RACE ON READ OF MsgTimeout FROM SYNCHRONIZER.SETUPLOOPS AND FROM RAWSTORAGE.START

// RawStorage is the struct which stores dynamic values;
// these values may change from epoch to epoch
type RawStorage struct {
	MaxBytes                       uint32        `json:"maxBytes,omitempty"`
	MaxProposalSize                uint32        `json:"maxProposalSize,omitempty"`
	ProposalStepTimeout            time.Duration `json:"proposalStepTimeout,omitempty"`
	PreVoteStepTimeout             time.Duration `json:"preVoteStepTimeout,omitempty"`
	PreCommitStepTimeout           time.Duration `json:"preCommitStepTimeout,omitempty"`
	DeadBlockRoundNextRoundTimeout time.Duration `json:"deadBlockRoundNextRoundTimeout,omitempty"`
	DownloadTimeout                time.Duration `json:"downloadTimeout,omitempty"`
	SrvrMsgTimeout                 time.Duration `json:"srvrMsgTimeout,omitempty"`
	MsgTimeout                     time.Duration `json:"msgTimeout,omitempty"`

	MinTxFee       *big.Int `json:"minTxFee,omitempty"`
	TxValidVersion uint32   `json:"txValidVersion,omitempty"`

	ValueStoreFee          *big.Int `json:"valueStoreFee,omitempty"`
	ValueStoreValidVersion uint32   `json:"valueStoreValidVersion,omitempty"`

	AtomicSwapFee            *big.Int `json:"atomicSwapFee,omitempty"`
	AtomicSwapValidStopEpoch uint32   `json:"atomicSwapValidStopEpoch,omitempty"`

	DataStoreEpochFee     *big.Int `json:"dataStoreEpochFee,omitempty"`
	DataStoreValidVersion uint32   `json:"dataStoreValidVersion,omitempty"`
}

// Marshal performs json.Marshal on the RawStorage struct.
func (rs *RawStorage) Marshal() ([]byte, error) {
	if rs == nil {
		return nil, ErrRawStorageNilPointer
	}
	return json.Marshal(rs)
}

// Unmarshal performs json.Unmarshal on the RawStorage struct.
func (rs *RawStorage) Unmarshal(v []byte) error {
	if rs == nil {
		return ErrRawStorageNilPointer
	}
	if len(v) == 0 {
		return ErrUnmarshalEmpty
	}
	return json.Unmarshal(v, rs)
}

// Copy makes a complete copy of RawStorage struct.
func (rs *RawStorage) Copy() (*RawStorage, error) {
	rsBytes, err := rs.Marshal()
	if err != nil {
		return nil, err
	}
	c := &RawStorage{}
	err = c.Unmarshal(rsBytes)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// IsValid returns true if we can successfully make a copy
func (rs *RawStorage) IsValid() bool {
	_, err := rs.Copy()
	if err != nil {
		return false
	}
	return true
}

// UpdateValue updates the field with the appropriate value.
func (rs *RawStorage) UpdateValue(update Updater) error {
	value := update.Value()
	switch update.Type() {
	case MaxBytesType:
		// uint32
		v, err := stringToUint32(value)
		if err != nil {
			return err
		}
		rs.SetMaxBytes(v)
	case ProposalStepTimeoutType:
		// time.Duration
		v, err := stringToTimeDuration(value)
		if err != nil {
			return err
		}
		rs.SetProposalStepTimeout(v)
	case PreVoteStepTimeoutType:
		// time.Duration
		v, err := stringToTimeDuration(value)
		if err != nil {
			return err
		}
		rs.SetPreVoteStepTimeout(v)
	case PreCommitStepTimeoutType:
		// time.Duration
		v, err := stringToTimeDuration(value)
		if err != nil {
			return err
		}
		rs.SetPreCommitStepTimeout(v)
	case MsgTimeoutType:
		// time.Duration
		v, err := stringToTimeDuration(value)
		if err != nil {
			return err
		}
		rs.SetMsgTimeout(v)
	case MinTxFeeType:
		// *big.Int
		v, err := stringToBigInt(value)
		if err != nil {
			return err
		}
		err = rs.SetMinTxFee(v)
		if err != nil {
			return err
		}
	case TxValidVersionType:
		// uint32
		v, err := stringToUint32(value)
		if err != nil {
			return err
		}
		rs.SetTxValidVersion(v)
	case ValueStoreFeeType:
		// *big.Int
		v, err := stringToBigInt(value)
		if err != nil {
			return err
		}
		err = rs.SetValueStoreFee(v)
		if err != nil {
			return err
		}
	case ValueStoreValidVersionType:
		// uint32
		v, err := stringToUint32(value)
		if err != nil {
			return err
		}
		rs.SetValueStoreValidVersion(v)
	case AtomicSwapFeeType:
		// *big.Int
		v, err := stringToBigInt(value)
		if err != nil {
			return err
		}
		err = rs.SetAtomicSwapFee(v)
		if err != nil {
			return err
		}
	case AtomicSwapValidStopEpochType:
		// uint32
		v, err := stringToUint32(value)
		if err != nil {
			return err
		}
		rs.SetAtomicSwapValidStopEpoch(v)
	case DataStoreEpochFeeType:
		// *big.Int
		v, err := stringToBigInt(value)
		if err != nil {
			return err
		}
		err = rs.SetDataStoreEpochFee(v)
		if err != nil {
			return err
		}
	case DataStoreValidVersionType:
		// uint32
		v, err := stringToUint32(value)
		if err != nil {
			return err
		}
		rs.SetDataStoreValidVersion(v)
	default:
		return ErrInvalidUpdateValue
	}
	return nil
}

// standardParameters initializes RawStorage with the standard (original)
// parameters for the system.
func (rs *RawStorage) standardParameters() {
	rs.MaxBytes = maxBytes
	rs.MaxProposalSize = maxProposalSize
	rs.ProposalStepTimeout = proposalStepTO
	rs.PreVoteStepTimeout = preVoteStepTO
	rs.PreCommitStepTimeout = preCommitStepTO
	rs.DeadBlockRoundNextRoundTimeout = dBRNRTO
	rs.DownloadTimeout = downloadTO
	rs.SrvrMsgTimeout = srvrMsgTimeout
	rs.MsgTimeout = msgTimeout
}

// GetMaxBytes returns the maximum allowed bytes
func (rs *RawStorage) GetMaxBytes() uint32 {
	return rs.MaxBytes
}

// SetMaxBytes sets the maximum allowed bytes
func (rs *RawStorage) SetMaxBytes(value uint32) {
	rs.MaxBytes = value
	rs.MaxProposalSize = value
}

// GetMaxProposalSize returns the maximum size of bytes allowed in a proposal
func (rs *RawStorage) GetMaxProposalSize() uint32 {
	return rs.MaxProposalSize
}

// GetSrvrMsgTimeout returns the time before timeout of server message
func (rs *RawStorage) GetSrvrMsgTimeout() time.Duration {
	return rs.SrvrMsgTimeout
}

// GetMsgTimeout returns the timeout to receive a message
func (rs *RawStorage) GetMsgTimeout() time.Duration {
	return rs.MsgTimeout
}

// SetMsgTimeout sets the timeout to receive a message
func (rs *RawStorage) SetMsgTimeout(value time.Duration) {
	rs.MsgTimeout = value
	rs.SrvrMsgTimeout = (3 * value) / 4
}

// GetProposalStepTimeout returns the proposal step timeout
func (rs *RawStorage) GetProposalStepTimeout() time.Duration {
	return rs.ProposalStepTimeout
}

// SetProposalStepTimeout sets the proposal step timeout
func (rs *RawStorage) SetProposalStepTimeout(value time.Duration) {
	rs.ProposalStepTimeout = value
	sum := rs.ProposalStepTimeout + rs.PreVoteStepTimeout + rs.PreCommitStepTimeout
	rs.DownloadTimeout = sum
	rs.DeadBlockRoundNextRoundTimeout = (5 * sum) / 2
}

// GetPreVoteStepTimeout returns the prevote step timeout
func (rs *RawStorage) GetPreVoteStepTimeout() time.Duration {
	return rs.PreVoteStepTimeout
}

// SetPreVoteStepTimeout sets the prevote step timeout
func (rs *RawStorage) SetPreVoteStepTimeout(value time.Duration) {
	rs.PreVoteStepTimeout = value
	sum := rs.ProposalStepTimeout + rs.PreVoteStepTimeout + rs.PreCommitStepTimeout
	rs.DownloadTimeout = sum
	rs.DeadBlockRoundNextRoundTimeout = (5 * sum) / 2
}

// GetPreCommitStepTimeout returns the precommit step timeout
func (rs *RawStorage) GetPreCommitStepTimeout() time.Duration {
	return rs.PreCommitStepTimeout
}

// SetPreCommitStepTimeout sets the precommit step timeout
func (rs *RawStorage) SetPreCommitStepTimeout(value time.Duration) {
	rs.PreCommitStepTimeout = value
	sum := rs.ProposalStepTimeout + rs.PreVoteStepTimeout + rs.PreCommitStepTimeout
	rs.DownloadTimeout = sum
	rs.DeadBlockRoundNextRoundTimeout = (5 * sum) / 2
}

// GetDeadBlockRoundNextRoundTimeout returns the timeout required before
// moving into the DeadBlockRound
func (rs *RawStorage) GetDeadBlockRoundNextRoundTimeout() time.Duration {
	return rs.DeadBlockRoundNextRoundTimeout
}

// GetDownloadTimeout returns the timeout for downloads
func (rs *RawStorage) GetDownloadTimeout() time.Duration {
	return rs.DownloadTimeout
}

// GetMinTxFee returns the minimun tx burned fee
func (rs *RawStorage) GetMinTxFee() *big.Int {
	if rs.MinTxFee == nil {
		rs.MinTxFee = new(big.Int)
	}
	return rs.MinTxFee
}

// SetMinTxFee sets the minimun tx burned fee
func (rs *RawStorage) SetMinTxFee(value *big.Int) error {
	if value == nil {
		return ErrInvalidValue
	}
	if rs.MinTxFee == nil {
		rs.MinTxFee = new(big.Int)
	}
	if value.Sign() < 0 {
		return ErrInvalidValue
	}
	rs.MinTxFee.Set(value)
	return nil
}

// GetTxValidVersion returns the valid version of tx
func (rs *RawStorage) GetTxValidVersion() uint32 {
	return rs.TxValidVersion
}

// SetTxValidVersion sets the minimun tx burned fee
func (rs *RawStorage) SetTxValidVersion(value uint32) {
	rs.TxValidVersion = value
}

// GetDataStoreEpochFee returns the minimun ValueStore burned fee
func (rs *RawStorage) GetDataStoreEpochFee() *big.Int {
	if rs.DataStoreEpochFee == nil {
		rs.DataStoreEpochFee = new(big.Int)
	}
	return rs.DataStoreEpochFee
}

// SetDataStoreEpochFee sets the minimun ValueStore burned fee
func (rs *RawStorage) SetDataStoreEpochFee(value *big.Int) error {
	if value == nil {
		return ErrInvalidValue
	}
	if rs.DataStoreEpochFee == nil {
		rs.DataStoreEpochFee = new(big.Int)
	}
	if value.Sign() < 0 {
		return ErrInvalidValue
	}
	rs.DataStoreEpochFee.Set(value)
	return nil
}

// GetValueStoreFee returns the minimun ValueStore burned fee
func (rs *RawStorage) GetValueStoreFee() *big.Int {
	if rs.ValueStoreFee == nil {
		rs.ValueStoreFee = new(big.Int)
	}
	return rs.ValueStoreFee
}

// SetValueStoreFee sets the minimun ValueStore burned fee
func (rs *RawStorage) SetValueStoreFee(value *big.Int) error {
	if value == nil {
		return ErrInvalidValue
	}
	if rs.ValueStoreFee == nil {
		rs.ValueStoreFee = new(big.Int)
	}
	if value.Sign() < 0 {
		return ErrInvalidValue
	}
	rs.ValueStoreFee.Set(value)
	return nil
}

// GetValueStoreValidVersion returns the valid version of ValueStore
func (rs *RawStorage) GetValueStoreValidVersion() uint32 {
	return rs.ValueStoreValidVersion
}

// SetValueStoreValidVersion sets the valid version of ValueStore
func (rs *RawStorage) SetValueStoreValidVersion(value uint32) {
	rs.ValueStoreValidVersion = value
}

// GetAtomicSwapFee returns the minimun AtomicSwap burned fee
func (rs *RawStorage) GetAtomicSwapFee() *big.Int {
	if rs.AtomicSwapFee == nil {
		rs.AtomicSwapFee = new(big.Int)
	}
	return rs.AtomicSwapFee
}

// SetAtomicSwapFee sets the minimun AtomicSwap burned fee
func (rs *RawStorage) SetAtomicSwapFee(value *big.Int) error {
	if value == nil {
		return ErrInvalidValue
	}
	if rs.AtomicSwapFee == nil {
		rs.AtomicSwapFee = new(big.Int)
	}
	if value.Sign() < 0 {
		return ErrInvalidValue
	}
	rs.AtomicSwapFee.Set(value)
	return nil
}

// GetAtomicSwapValidStopEpoch returns the valid version of AtomicSwap
func (rs *RawStorage) GetAtomicSwapValidStopEpoch() uint32 {
	return rs.AtomicSwapValidStopEpoch
}

// SetAtomicSwapValidStopEpoch sets the valid version of AtomicSwap
func (rs *RawStorage) SetAtomicSwapValidStopEpoch(value uint32) {
	rs.AtomicSwapValidStopEpoch = value
}

// GetDataStoreValidVersion returns the valid version of DataStore
func (rs *RawStorage) GetDataStoreValidVersion() uint32 {
	return rs.DataStoreValidVersion
}

// SetDataStoreValidVersion sets the valid version of AtomicSwap
func (rs *RawStorage) SetDataStoreValidVersion(value uint32) {
	rs.DataStoreValidVersion = value
}
