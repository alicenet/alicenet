package dynamics

import (
	"math/big"
	"strconv"
	"time"
)

// Ensuring interface check
var _ Updater = (*Update)(nil)

// UpdateType specifies which field will be updated
type UpdateType int

const (
	// MaxBytesType is the UpdateType for updating MaxBytes and MaxProposalSize
	MaxBytesType UpdateType = iota + 1

	// ProposalStepTimeoutType is the UpdateType for updating ProposalStepTimeout;
	// this also parameterizes DownloadTimeout and DeadBlockRoundNextRoundTimeout
	ProposalStepTimeoutType

	// PreVoteStepTimeoutType is the UpdateType for updating PreVoteStepTimeout;
	// this also parameterizes DownloadTimeout and DeadBlockRoundNextRoundTimeout
	PreVoteStepTimeoutType

	// PreCommitStepTimeoutType is the UpdateType for updating PreCommitStepTimeout;
	// this also parameterizes DownloadTimeout and DeadBlockRoundNextRoundTimeout
	PreCommitStepTimeoutType

	// MsgTimeoutType is the UpdateType for updating MsgTimeout;
	// this also parameterizes SrvrMsgTimeout
	MsgTimeoutType

	// MinTxFeeCostRatioType is the UpdateType for updating MinTxFee
	MinTxFeeCostRatioType

	// TxValidVersionType is the UpdateType for updating TxValidVersion
	TxValidVersionType

	// ValueStoreFeeType is the UpdateType for updating ValueStoreFee
	ValueStoreFeeType

	// ValueStoreValidVersionType is the UpdateType for updating ValueStoreValidVersion
	ValueStoreValidVersionType

	// AtomicSwapFeeType is the UpdateType for updating AtomicSwapFee
	AtomicSwapFeeType

	// AtomicSwapValidStopEpochType is the UpdateType for updating AtomicSwapValidStopEpoch
	AtomicSwapValidStopEpochType

	// DataStoreEpochFeeType is the UpdateType for updating DataStoreEpochFee
	DataStoreEpochFeeType

	// DataStoreValidVersionType is the UpdateType for updating DataStoreValidVersion
	DataStoreValidVersionType
)

// Updater specifies the interface we use for updating Storage
type Updater interface {
	Name() string
	Type() UpdateType
	Value() string
	Epoch() uint32
}

// Update is an implementation of Updater interface
type Update struct {
	name  string
	key   UpdateType
	value string
	epoch uint32
}

// Name returns the name of Update
func (u *Update) Name() string {
	return u.name
}

// Type returns the type of Update
func (u *Update) Type() UpdateType {
	return u.key
}

// Value returns the value of Update
func (u *Update) Value() string {
	return u.value
}

// Epoch returns the epoch of Update
func (u *Update) Epoch() uint32 {
	return u.epoch
}

// NewUpdate makes a valid valid Update struct which is then used
func NewUpdate(field, value string, epoch uint32) (*Update, error) {
	keyType, err := convertFieldToType(field)
	if err != nil {
		return nil, err
	}
	u := &Update{
		name:  field,
		key:   keyType,
		value: value,
		epoch: epoch,
	}
	return u, nil
}

// stringToInt32 converts a string into an int32
func stringToInt32(value string) (int32, error) {
	v64, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}
	v := int32(v64)
	return v, nil
}

// stringToUint32 converts a string into a uint32
func stringToUint32(value string) (uint32, error) {
	v64, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	v := uint32(v64)
	return v, nil
}

// stringToTimeDuration converts a string into time.Duration.
// This conversion is particular for our situation, which is why
// we do not allow negative time.Duration values.
func stringToTimeDuration(value string) (time.Duration, error) {
	v64, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	if v64 < 0 {
		return 0, ErrInvalid
	}
	v := time.Duration(v64)
	return v, nil
}

// stringToBigInt converts a string into *big.Int
// This conversion is particular for our situation, which is why
// we do not allow negative big.Int values.
func stringToBigInt(value string) (*big.Int, error) {
	v, valid := new(big.Int).SetString(value, 10)
	if !valid {
		return nil, ErrInvalid
	}
	if v.Sign() < 0 {
		return nil, ErrInvalid
	}
	return v, nil
}

// convertFieldToType returns the type associated types from the field
func convertFieldToType(field string) (UpdateType, error) {
	switch field {
	case "maxBytes":
		return MaxBytesType, nil
	case "proposalStepTimeout":
		return ProposalStepTimeoutType, nil
	case "preVoteStepTimeout":
		return PreVoteStepTimeoutType, nil
	case "preCommitStepTimeout":
		return PreCommitStepTimeoutType, nil
	case "msgTimeout":
		return MsgTimeoutType, nil
	case "minTxFeeCostRatio":
		return MinTxFeeCostRatioType, nil
	case "txValidVersion":
		return TxValidVersionType, nil
	case "valueStoreFee":
		return ValueStoreFeeType, nil
	case "valueStoreValidVersion":
		return ValueStoreValidVersionType, nil
	case "atomicSwapFee":
		return AtomicSwapFeeType, nil
	case "atomicSwapValidStopEpoch":
		return AtomicSwapValidStopEpochType, nil
	case "dataStoreEpochFee":
		return DataStoreEpochFeeType, nil
	case "dataStoreValidVersion":
		return DataStoreValidVersionType, nil
	default:
		return UpdateType(0), ErrInvalid
	}
}
