package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/constants"
)

func rsckEqual(t *testing.T, rsck, rsck2 *RoundStateCurrentKey) {
	if !bytes.Equal(rsck.Prefix, rsck2.Prefix) {
		t.Fatal("fail")
	}
	if !bytes.Equal(rsck.GroupKey, rsck2.GroupKey) {
		t.Fatal("fail")
	}
	if !bytes.Equal(rsck.VAddr, rsck2.VAddr) {
		t.Fatal("fail")
	}
}

func TestRoundStateCurrentKey(t *testing.T) {
	prefix := []byte("Prefix")
	GroupKey := make([]byte, constants.CurveBN256EthPubkeyLen)
	vaddr := make([]byte, constants.OwnerLen)
	rsck := &RoundStateCurrentKey{
		Prefix:   prefix,
		GroupKey: GroupKey,
		VAddr:    vaddr,
	}
	data, err := rsck.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	rsck2 := &RoundStateCurrentKey{}
	err = rsck2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	rsckEqual(t, rsck, rsck2)
}

func TestRoundStateCurrentKeyBad(t *testing.T) {
	dataBad0 := []byte("00000000|00000000000000")
	rsck := &RoundStateCurrentKey{}
	err := rsck.UnmarshalBinary(dataBad0)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}
