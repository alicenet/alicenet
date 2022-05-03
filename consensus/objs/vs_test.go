package objs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidatorsSet(t *testing.T) {
	groupk, bnSigners, bnShares, secpSigners, secpPubks := makeSigners2(t)
	_ = secpPubks
	_ = bnShares
	vs := makeValidatorSet(1, groupk, secpSigners, bnSigners)

	bt, err := vs.MarshalBinary()
	if err != nil {
		t.Fatalf("Error MarshalBinary: %v", err)
	}
	newVs := &ValidatorSet{}
	err = newVs.UnmarshalBinary(bt)

	if err != nil {
		t.Fatalf("Error UnmarshalBinary: %v", err)
	}

	isGroupShareValidator, groupShareIdx := newVs.GroupShareIdx(vs.Validators[1].GroupShare)
	assert.True(t, isGroupShareValidator)
	assert.Equal(t, 1, groupShareIdx)

	isValidator, valIdx := newVs.VAddrIdx(vs.Validators[2].VAddr)
	assert.True(t, isValidator)
	assert.Equal(t, 2, valIdx)

	isValidTuple := newVs.IsValidTuple(vs.Validators[0].VAddr, vs.GroupKey)
	assert.True(t, isValidTuple)

	isValidTriplet := newVs.IsValidTriplet(vs.Validators[0].VAddr, vs.Validators[0].GroupShare, vs.GroupKey)
	assert.True(t, isValidTriplet)
}
