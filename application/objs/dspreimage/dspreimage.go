package dspreimage

import (
	"bytes"

	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the DSPreImage object.
func Marshal(v mdefs.DSPreImage) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	out := utils.CopySlice(raw)
	return out, nil
}

// Unmarshal will unmarshal the DSPreImage object.
func Unmarshal(data []byte) (mdefs.DSPreImage, error) {
	var err error
	fn := func() (mdefs.DSPreImage, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		dataCopy := utils.CopySlice(data)
		msg := &capnp.Message{Arena: capnp.SingleSegment(dataCopy)}
		obj, tmp := mdefs.ReadRootDSPreImage(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.DSPreImage{}, err
	}
	return obj, nil
}

// Validate will validate the DSPreImage object
func Validate(v mdefs.DSPreImage) error {
	if v.ChainID() < 1 {
		return errorz.ErrInvalid{}.New("dspreimage capn obj is not valid; invalid ChainID")
	}
	if !v.HasIndex() {
		return errorz.ErrInvalid{}.New("dspreimage capn obj does not have Index")
	}
	if len(v.Index()) != constants.HashLen {
		return errorz.ErrInvalid{}.New("dspreimage capn obj is not valid; invalid Index; incorrect byte length")
	}
	zeroBytes := make([]byte, constants.HashLen)
	if bytes.Equal(v.Index(), zeroBytes) {
		return errorz.ErrInvalid{}.New("dspreimage capn obj is not valid; invalid index; all zeros")
	}
	if v.IssuedAt() < 1 {
		return errorz.ErrInvalid{}.New("dspreimage capn obj is not valid; invalid IssuedAt")
	}
	value := &uint256.Uint256{}
	err := value.FromUint32Array([8]uint32{v.Deposit(), v.Deposit1(), v.Deposit2(), v.Deposit3(), v.Deposit4(), v.Deposit5(), v.Deposit6(), v.Deposit7()})
	if err != nil {
		return err
	}
	if value.Lt(uint256.DSPIMinDeposit()) {
		return errorz.ErrInvalid{}.New("dspreimage capn obj is not valid; invalid Deposit: less than minimum possible")
	}
	if len(v.Owner()) == 0 {
		return errorz.ErrInvalid{}.New("dspreimage capn obj is not valid; invalid Owner; zero byte length")
	}
	if !v.HasRawData() {
		return errorz.ErrInvalid{}.New("dspreimage capn obj does not have RawData")
	}
	if len(v.RawData()) == 0 {
		return errorz.ErrInvalid{}.New("dspreimage capn obj is not valid: invalid RawData; zero byte length")
	}
	if int(v.TXOutIdx()) >= constants.MaxTxVectorLength {
		return errorz.ErrInvalid{}.New("dspreimage capn obj is not valid: output index is too large")
	}
	return nil
}
