package objs

import "testing"

func TestValidator(t *testing.T) {
	_, bnSigners, _, secpSigners, _ := makeSigners2(t)

	for i, ss := range secpSigners {
		groupKey, _ := bnSigners[i].PubkeyShare()
		secpKey, _ := ss.Pubkey()
		val := &Validator{
			VAddr:      secpKey, // change
			GroupShare: groupKey,
		}

		bt, err := val.MarshalBinary()
		if err != nil {
			t.Fatalf("Error MarshalBinary: %v", err)
		}
		newVal := &Validator{}
		err = newVal.UnmarshalBinary(bt)

		if err != nil {
			t.Fatalf("Error UnmarshalBinary: %v", err)
		}
	}
}
