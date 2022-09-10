package dynamics

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/alicenet/alicenet/constants"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestDynamicValuesMarshal(t *testing.T) {
	dv := &DynamicValues{}
	_, err := dv.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	s := &Storage{}
	_, err = s.DynamicValues.Marshal()
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestDynamicValuesUnmarshal(t *testing.T) {
	dv := &DynamicValues{}
	v, err := dv.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	dv2 := &DynamicValues{}
	err = dv2.Unmarshal(v)
	if err != nil {
		t.Fatal(err)
	}

	v = []byte{}
	dv3 := &DynamicValues{}
	err = dv3.Unmarshal(v)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	s := &Storage{}
	err = s.DynamicValues.Unmarshal(v)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestDynamicValuesCopy(t *testing.T) {
	// Copy empty DynamicValues
	dv1 := &DynamicValues{}
	dv2, err := dv1.Copy()
	if err != nil {
		t.Fatal(err)
	}
	dv1Bytes, err := dv1.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	dv2Bytes, err := dv2.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dv1Bytes, dv2Bytes) {
		t.Fatal("Should have equal bytes (1)")
	}

	// Copy DynamicValues with parameters
	dv1.standardParameters()
	dv2, err = dv1.Copy()
	if err != nil {
		t.Fatal(err)
	}
	dv1Bytes, err = dv1.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	dv2Bytes, err = dv2.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dv1Bytes, dv2Bytes) {
		t.Fatal("Should have equal bytes (2)")
	}

	// Copy DynamicValues with some parameters set to zero
	dv1.MaxBlockSize = 0
	dv2, err = dv1.Copy()
	if err != nil {
		t.Fatal(err)
	}
	dv1Bytes, err = dv1.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	dv2Bytes, err = dv2.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dv1Bytes, dv2Bytes) {
		t.Fatal("Should have equal bytes (3)")
	}

	s := &Storage{}
	_, err = s.DynamicValues.Copy()
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestDynamicValuesStandardParameters(t *testing.T) {
	dv := &DynamicValues{}
	dv.standardParameters()

	retMaxBytes := dv.GetMaxBlockSize()
	if retMaxBytes != constants.InitialMaxBlockSize {
		t.Fatal("Should be equal (1)")
	}

	retMaxProposalSize := dv.GetMaxProposalSize()
	if retMaxProposalSize != constants.InitialMaxBlockSize {
		t.Fatal("Should be equal (2)")
	}

	retProposalTimeout := dv.GetProposalTimeout()
	if retProposalTimeout != constants.InitialProposalTimeout {
		t.Fatal("Should be equal (5)")
	}

	retPreVoteTimeout := dv.GetPreVoteTimeout()
	if retPreVoteTimeout != constants.InitialPreVoteTimeout {
		t.Fatal("Should be equal (6)")
	}

	retPreCommitTimeout := dv.GetPreCommitTimeout()
	if retPreCommitTimeout != constants.InitialPreCommitTimeout {
		t.Fatal("Should be equal (7)")
	}
	sum := retProposalTimeout + retPreVoteTimeout + retPreCommitTimeout

	retDBRNRTO := dv.GetDeadBlockRoundNextRoundTimeout()
	if retDBRNRTO != (sum*5)/2 {
		t.Fatal("Should be equal (8)")
	}

	retDownloadTO := dv.GetDownloadTimeout()
	if retDownloadTO != sum {
		t.Fatal("Should be equal (9)")
	}
}

func TestDecodeDynamicValuesV1(t *testing.T) {
	maxUint128, ok := new(big.Int).SetString("340282366920938463463374607431768211455", 10)
	assert.True(t, ok)
	randUint128, ok := new(big.Int).SetString("253455453978969546304077282559958826669", 10)
	assert.True(t, ok)
	tests := []struct {
		name     string
		input    string
		expected *DynamicValues
	}{
		{
			name:  "Correctly decode dynamic values",
			input: "00000fa000000bb800000bb8002dc6c00000000000000000000000000000000000000000000000000000000000000000",
			expected: &DynamicValues{
				V1,
				time.Duration(4) * time.Second,
				time.Duration(3) * time.Second,
				time.Duration(3) * time.Second,
				3_000_000,
				new(big.Int).SetUint64(0),
				new(big.Int).SetUint64(0),
				new(big.Int).SetUint64(0),
			},
		},
		{
			name:  "Correctly decode dynamic values 2",
			input: "00babecadeadcafedeadcafebeadbeadbeadbeadbeadbeadbeadbeadbeadbeadbeadbeadbeadbeadbeadbeadbeadbead",
			expected: &DynamicValues{
				V1,
				time.Duration(12_238_538) * time.Millisecond,
				time.Duration(3_735_931_646) * time.Millisecond,
				time.Duration(3_735_931_646) * time.Millisecond,
				3_199_057_581,
				new(big.Int).SetUint64(13_739_847_691_614_928_557),
				new(big.Int).SetUint64(13_739_847_691_614_928_557),
				randUint128,
			},
		},
		{
			name:  "Correctly decode dynamic values with extra bytes",
			input: "00000fa000000bb800000bb8002dc6c00000000000000000000000000000000000000000000000000000000000000000cafecafe",
			expected: &DynamicValues{
				V1,
				time.Duration(4) * time.Second,
				time.Duration(3) * time.Second,
				time.Duration(3) * time.Second,
				3_000_000,
				new(big.Int).SetUint64(0),
				new(big.Int).SetUint64(0),
				new(big.Int).SetUint64(0),
			},
		},
		{
			name:  "Correctly decode dynamic values with all zeros",
			input: "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			expected: &DynamicValues{
				V1,
				time.Duration(0) * time.Second,
				time.Duration(0) * time.Second,
				time.Duration(0) * time.Second,
				0,
				new(big.Int).SetUint64(0),
				new(big.Int).SetUint64(0),
				new(big.Int).SetUint64(0),
			},
		},
		{
			name:  "Correctly decode dynamic values with all max values",
			input: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			expected: &DynamicValues{
				255,
				time.Duration(16_777_215) * time.Millisecond,
				time.Duration(4_294_967_295) * time.Millisecond,
				time.Duration(4_294_967_295) * time.Millisecond,
				4_294_967_295,
				new(big.Int).SetUint64(18_446_744_073_709_551_615),
				new(big.Int).SetUint64(18_446_744_073_709_551_615),
				maxUint128,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := hex.DecodeString(tt.input)
			assert.Nil(t, err)
			dynamicValues, err := DecodeDynamicValues(data)
			assert.Nil(t, err)
			bigIntComparer := func(a *big.Int, b *big.Int) bool {
				return a.Cmp(b) == 0
			}
			if diff := cmp.Diff(dynamicValues, tt.expected, cmp.Comparer(bigIntComparer)); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}

		})
	}
}

func TestShouldNotDecodeIncorrectSizeDynamicValuesV1(t *testing.T) {
	t.Parallel()
	data, err := hex.DecodeString("00000fa000000bb800000bb8002dc6c000000000000000000000000000000000000000")
	assert.Nil(t, err)
	_, err = DecodeDynamicValues(data)
	targetError := &ErrInvalidDynamicValueStructLen{}
	if !errors.As(err, &targetError) {
		t.Fatalf("expected function to fail with error type: &ErrInvalidDynamicValueStructLen{} got %v", err)
	}
}

func TestShouldNotDecodeIncorrectSize2DynamicValuesV1(t *testing.T) {
	t.Parallel()
	data, err := hex.DecodeString("00000fa000000bb800000bb8002dc6c000000000000000000000000000000000000000000000000000000000000000")
	assert.Nil(t, err)
	_, err = DecodeDynamicValues(data)
	targetError := &ErrInvalidDynamicValueStructLen{}
	if !errors.As(err, &targetError) {
		t.Fatalf("expected function to fail with error type: &ErrInvalidDynamicValueStructLen{} got %v", err)
	}
}

func TestShouldDecodeUint32(t *testing.T) {
	t.Parallel()
	data, err := hex.DecodeString("babebabe")
	assert.Nil(t, err)
	uint32Data, err := decodeUInt32WithArbitraryLength(data)
	assert.Nil(t, err)
	assert.Equal(t, uint32Data, uint32(0xbabebabe))
}

func TestShouldDecode2Uint32(t *testing.T) {
	t.Parallel()
	data, err := hex.DecodeString("babe")
	assert.Nil(t, err)
	uint32Data, err := decodeUInt32WithArbitraryLength(data)
	assert.Nil(t, err)
	assert.Equal(t, uint32Data, uint32(0xbabe))
}

func TestShouldNotDecodeInvalidUint32(t *testing.T) {
	t.Parallel()
	data, err := hex.DecodeString("babebabeca")
	assert.Nil(t, err)
	_, err = decodeUInt32WithArbitraryLength(data)
	targetError := &ErrInvalidSize{}
	if !errors.As(err, &targetError) {
		t.Fatalf("expected function to fail with error type: &ErrInvalidSize{} got %v", err)
	}
}

func TestShouldNotDecodeEmptyUint32(t *testing.T) {
	t.Parallel()
	data, err := hex.DecodeString("")
	assert.Nil(t, err)
	_, err = decodeUInt32WithArbitraryLength(data)
	targetError := &ErrInvalidSize{}
	if !errors.As(err, &targetError) {
		t.Fatalf("expected function to fail with error type: &ErrInvalidSize{} got %v", err)
	}
}

func TestShouldDecodeTimeDuration(t *testing.T) {
	t.Parallel()
	data, err := hex.DecodeString("babebabe")
	assert.Nil(t, err)
	uint32Data, err := decodeTimeDurationInMilliSeconds(data)
	assert.Nil(t, err)
	assert.Equal(t, uint32Data, time.Duration(0xbabebabe)*time.Millisecond)
}

func TestShouldDecodeUint64(t *testing.T) {
	t.Parallel()
	data, err := hex.DecodeString("babebabecafecafe")
	assert.Nil(t, err)
	uint64Data, err := decodeUInt64WithArbitraryLength(data)
	assert.Nil(t, err)
	assert.Equal(t, uint64Data, uint64(0xbabebabecafecafe))
}

func TestShouldDecode2Uint64(t *testing.T) {
	t.Parallel()
	data, err := hex.DecodeString("babe")
	assert.Nil(t, err)
	uint64Data, err := decodeUInt64WithArbitraryLength(data)
	assert.Nil(t, err)
	assert.Equal(t, uint64Data, uint64(0xbabe))
}

func TestShouldNotDecodeInvalidUint64(t *testing.T) {
	t.Parallel()
	data, err := hex.DecodeString("babebabecafecafeba")
	assert.Nil(t, err)
	_, err = decodeUInt64WithArbitraryLength(data)
	targetError := &ErrInvalidSize{}
	if !errors.As(err, &targetError) {
		t.Fatalf("expected function to fail with error type: &ErrInvalidSize{} got %v", err)
	}
}

func TestShouldNotDecodeEmptyUint64(t *testing.T) {
	t.Parallel()
	data, err := hex.DecodeString("")
	assert.Nil(t, err)
	_, err = decodeUInt64WithArbitraryLength(data)
	targetError := &ErrInvalidSize{}
	if !errors.As(err, &targetError) {
		t.Fatalf("expected function to fail with error type: &ErrInvalidSize{} got %v", err)
	}
}
