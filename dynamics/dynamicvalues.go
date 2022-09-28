package dynamics

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/alicenet/alicenet/utils"
)

// Version is an enumeration indicating which dynamic values versions we can get
type Version int

// Possible versions that we can have during the dynamic value event change
const (
	V1 Version = iota
)

func (version Version) String() string {
	return [...]string{
		"V1",
	}[version]
}

const DynamicsValueV1Size int = 48

type DynamicValues struct {
	EncoderVersion          Version
	ProposalTimeout         time.Duration
	PreVoteTimeout          time.Duration
	PreCommitTimeout        time.Duration
	MaxBlockSize            uint32
	DataStoreFee            *big.Int
	ValueStoreFee           *big.Int
	MinScaledTransactionFee *big.Int
}

func DecodeDynamicValues(data []byte) (*DynamicValues, error) {
	if len(data) < DynamicsValueV1Size {
		return nil, &ErrInvalidDynamicValueStructLen{fmt.Sprintf("%x", data), len(data), DynamicsValueV1Size}
	}
	encoderVersion, err := decodeUInt32WithArbitraryLength(data[0:1])
	if err != nil {
		return nil, &ErrInvalidDynamicValue{"version", err.Error()}
	}

	proposalTimeout, err := decodeTimeDurationInMilliSeconds(data[1:4])
	if err != nil {
		return nil, &ErrInvalidDynamicValue{"proposalTimeout", err.Error()}
	}

	preVoteTimeout, err := decodeTimeDurationInMilliSeconds(data[4:8])
	if err != nil {
		return nil, &ErrInvalidDynamicValue{"preVoteTimeout", err.Error()}
	}

	preCommitTimeout, err := decodeTimeDurationInMilliSeconds(data[8:12])
	if err != nil {
		return nil, &ErrInvalidDynamicValue{"preCommitTimeout", err.Error()}
	}

	// maxBlockSize can be converted to uint32, since the smart contract will
	// enforce max value lest than < max(uint32)
	maxBlockSize, err := decodeUInt64WithArbitraryLength(data[12:16])
	if err != nil {
		return nil, &ErrInvalidDynamicValue{"maxBlockSize", err.Error()}
	}

	dataStoreFee := new(big.Int).SetBytes(data[16:24])
	valueStoreFee := new(big.Int).SetBytes(data[24:32])
	minScaledTransactionFee := new(big.Int).SetBytes(data[32:48])

	dynamicValuesV1 := &DynamicValues{
		Version(encoderVersion),
		proposalTimeout,
		preVoteTimeout,
		preCommitTimeout,
		uint32(maxBlockSize),
		dataStoreFee,
		valueStoreFee,
		minScaledTransactionFee,
	}

	return dynamicValuesV1, nil
}

// Marshal performs json.Marshal on the DynamicValuesV1 struct.
func (dv *DynamicValues) Marshal() ([]byte, error) {
	if dv == nil {
		return nil, ErrDynamicValueNilPointer
	}
	return json.Marshal(dv)
}

// Unmarshal performs json.Unmarshal on the DynamicValuesV1 struct.
func (dv *DynamicValues) Unmarshal(v []byte) error {
	if dv == nil {
		return ErrDynamicValueNilPointer
	}
	if len(v) == 0 {
		return ErrUnmarshalEmpty
	}
	return json.Unmarshal(v, dv)
}

// Copy makes a complete copy of DynamicValuesV1 struct.
func (dv *DynamicValues) Copy() (*DynamicValues, error) {
	dvBytes, err := dv.Marshal()
	if err != nil {
		return nil, err
	}
	c := &DynamicValues{}
	err = c.Unmarshal(dvBytes)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// IsValid returns true if we can successfully make a copy
func (dv *DynamicValues) IsValid() bool {
	_, err := dv.Copy()
	return err == nil
}

// GetMaxBlockSize returns the maximum allowed bytes in a block
func (dv *DynamicValues) GetMaxBlockSize() uint32 {
	return dv.MaxBlockSize
}

// GetMaxProposalSize returns the maximum size of bytes allowed in a proposal
func (dv *DynamicValues) GetMaxProposalSize() uint32 {
	// proposal size as it stands today is equal to the block size
	return dv.MaxBlockSize
}

// GetProposalTimeout returns the proposal step timeout
func (dv *DynamicValues) GetProposalTimeout() time.Duration {
	return dv.ProposalTimeout
}

// GetPreVoteTimeout returns the prevote step timeout
func (dv *DynamicValues) GetPreVoteTimeout() time.Duration {
	return dv.PreVoteTimeout
}

// GetPreCommitTimeout returns the precommit step timeout
func (dv *DynamicValues) GetPreCommitTimeout() time.Duration {
	return dv.PreCommitTimeout
}

// GetDeadBlockRoundNextRoundTimeout returns the timeout required before
// moving into the DeadBlockRound
func (dv *DynamicValues) GetDeadBlockRoundNextRoundTimeout() time.Duration {
	sum := dv.ProposalTimeout + dv.PreVoteTimeout + dv.PreCommitTimeout
	deadBlockRoundNextRoundTimeout := (5 * sum) / 2
	return deadBlockRoundNextRoundTimeout
}

// GetDownloadTimeout returns the timeout for downloads
func (dv *DynamicValues) GetDownloadTimeout() time.Duration {
	return dv.ProposalTimeout + dv.PreVoteTimeout + dv.PreCommitTimeout
}

// GetMinScaledTransactionFee returns the minimum tx burned fee
func (dv *DynamicValues) GetMinScaledTransactionFee() *big.Int {
	if dv.MinScaledTransactionFee == nil {
		dv.MinScaledTransactionFee = new(big.Int)
	}
	return dv.MinScaledTransactionFee
}

// GetDataStoreFee returns the minimum ValueStore burned fee
func (dv *DynamicValues) GetDataStoreFee() *big.Int {
	if dv.DataStoreFee == nil {
		dv.DataStoreFee = new(big.Int)
	}
	return dv.DataStoreFee
}

// GetValueStoreFee returns the minimum ValueStore burned fee
func (dv *DynamicValues) GetValueStoreFee() *big.Int {
	if dv.ValueStoreFee == nil {
		dv.ValueStoreFee = new(big.Int)
	}
	return dv.ValueStoreFee
}

func decodeUInt32WithArbitraryLength(data []byte) (uint32, error) {
	size := 4
	if len(data) == 0 || len(data) > size {
		return 0, &ErrInvalidSize{
			fmt.Sprintf("invalid number of bytes to convert to a uint32: %d", len(data)),
		}
	}
	value, err := utils.UnmarshalUint32(utils.ForceSliceToLength(data, size))
	if err != nil {
		return 0, err
	}
	return value, nil
}

func decodeTimeDurationInMilliSeconds(data []byte) (time.Duration, error) {
	timeDurationInt, err := decodeUInt32WithArbitraryLength(data)
	if err != nil {
		return 0, err
	}
	timeDuration := time.Duration(timeDurationInt) * time.Millisecond
	return timeDuration, nil
}

func decodeUInt64WithArbitraryLength(data []byte) (uint64, error) {
	size := 8
	if len(data) == 0 || len(data) > size {
		return 0, &ErrInvalidSize{
			fmt.Sprintf("invalid number of bytes to convert to a uint64: %d", len(data)),
		}
	}
	value, err := utils.UnmarshalUint64(utils.ForceSliceToLength(data, size))
	if err != nil {
		return 0, err
	}
	return value, nil
}
