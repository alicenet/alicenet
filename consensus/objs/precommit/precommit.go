package precommit

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the PreCommit object.
func Marshal(v mdefs.PreCommit) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal PreCommit.
func Unmarshal(data []byte) (mdefs.PreCommit, error) {
	var err error
	fn := func() (mdefs.PreCommit, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootPreCommit(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.PreCommit{}, err
	}
	return obj, nil
}

// Validate will validate the PreCommit object
func Validate(p mdefs.PreCommit) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("precommit capn obj is not valid")
	}
	if !p.HasProposal() {
		return errorz.ErrInvalid{}.New("precommit capn obj does not have Proposal")
	}
	if !p.HasPreVotes() {
		return errorz.ErrInvalid{}.New("precommit capn obj does not have PreVotes")
	}
	if !p.HasSignature() {
		return errorz.ErrInvalid{}.New("precommit capn obj does not have Signature")
	}
	return nil
}
