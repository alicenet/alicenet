package objs

import (
	"bytes"

	"github.com/MadBase/MadNet/errorz"

	"github.com/MadBase/MadNet/consensus/objs/blockheader"
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	gUtils "github.com/MadBase/MadNet/utils"
	capnp "zombiezen.com/go/capnproto2"
)

// BlockHeader ...
type BlockHeader struct {
	BClaims  *BClaims
	SigGroup []byte
	TxHshLst [][]byte
	// Not Part of actual object below this line
	GroupKey []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// BlockHeader object
func (b *BlockHeader) UnmarshalBinary(data []byte) error {
	bh, err := blockheader.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *BlockHeader) UnmarshalCapn(bh mdefs.BlockHeader) error {
	if err := b.unmarshalCapnUnsafe(bh); err != nil {
		return err
	}
	if len(b.TxHshLst) != int(b.BClaims.TxCount) {
		return errorz.ErrInvalid{}.New("cap bclaims txhsh lst len mismatch")
	}
	return nil
}

// UnmarshalBinaryUnsafe unmarshals a BlockHeader in an unsafe manner;
// that is, we do not look at the TxHshLst. This is necessary during
// fast synchronization, as transactions are not synced.
func (b *BlockHeader) UnmarshalBinaryUnsafe(data []byte) error {
	bh, err := blockheader.Unmarshal(data)
	if err != nil {
		return err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return b.unmarshalCapnUnsafe(bh)
}

func (b *BlockHeader) unmarshalCapnUnsafe(bh mdefs.BlockHeader) error {
	b.BClaims = &BClaims{}
	err := blockheader.Validate(bh)
	if err != nil {
		return err
	}
	txHshLst := bh.TxHshLst()
	lst, err := SplitHashes(txHshLst)
	if err != nil {
		return err
	}
	b.TxHshLst = lst
	err = b.BClaims.UnmarshalCapn(bh.BClaims())
	if err != nil {
		return err
	}
	b.SigGroup = bh.SigGroup()
	return nil
}

// MarshalBinary takes the BlockHeader object and returns the canonical
// byte slice
func (b *BlockHeader) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	defer bh.Struct.Segment().Message().Reset(nil)
	return blockheader.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *BlockHeader) MarshalCapn(seg *capnp.Segment) (mdefs.BlockHeader, error) {
	if b == nil {
		return mdefs.BlockHeader{}, errorz.ErrInvalid{}.New("not initialized")
	}
	var bh mdefs.BlockHeader
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bh, err
		}
		tmp, err := mdefs.NewRootBlockHeader(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewBlockHeader(seg)
		if err != nil {
			return bh, err
		}
		bh = tmp
	}
	bc, err := b.BClaims.MarshalCapn(seg)
	if err != nil {
		return bh, err
	}
	err = bh.SetBClaims(bc)
	if err != nil {
		return mdefs.BlockHeader{}, err
	}
	err = bh.SetSigGroup(b.SigGroup)
	if err != nil {
		return mdefs.BlockHeader{}, err
	}
	err = bh.SetTxHshLst(bytes.Join(b.TxHshLst, []byte("")))
	if err != nil {
		return mdefs.BlockHeader{}, err
	}
	return bh, nil
}

// BlockHash returns the BlockHash of BlockHeader
func (b *BlockHeader) BlockHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.BClaims.BlockHash()
}

// ValidateSignatures validates the TxRoot and group signature
// on the Blockheader
func (b *BlockHeader) ValidateSignatures(bnVal *crypto.BNGroupValidator) error {
	if b == nil || b.BClaims == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	txRoot, err := MakeTxRoot(b.TxHshLst)
	if err != nil {
		return err
	}
	if !bytes.Equal(txRoot, b.BClaims.TxRoot) {
		return errorz.ErrInvalid{}.New("bclaims txroot mismatch")
	}
	bhsh, err := b.BlockHash()
	if err != nil {
		return err
	}
	if b.BClaims.Height > 1 {
		groupKey, err := bnVal.Validate(bhsh, b.SigGroup)
		if err != nil {
			return err
		}
		b.GroupKey = groupKey
	} else {
		b.GroupKey = make([]byte, constants.CurveBN256EthPubkeyLen)
	}
	return nil
}

// GetRCert returns the RCert for BlockHeader
func (b *BlockHeader) GetRCert() (*RCert, error) {
	bhsh, err := b.BlockHash()
	if err != nil {
		return nil, err
	}
	rc := &RCert{
		SigGroup: b.SigGroup,
		RClaims: &RClaims{
			ChainID:   b.BClaims.ChainID,
			Height:    b.BClaims.Height + 1,
			Round:     1,
			PrevBlock: bhsh,
		},
	}
	return rc, nil
}

// MakeDeadRoundProposal makes the proposal for the DeadBlockRound
func (b *BlockHeader) MakeDeadRoundProposal(rcert *RCert, headerRoot []byte) (*Proposal, error) {
	if b == nil || b.BClaims == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	txRoot, err := MakeTxRoot([][]byte{})
	StateRoot := gUtils.CopySlice(b.BClaims.StateRoot)
	prevBlock := gUtils.CopySlice(rcert.RClaims.PrevBlock)
	if err != nil {
		return nil, err
	}
	p := &Proposal{
		PClaims: &PClaims{
			RCert: rcert,
			BClaims: &BClaims{
				TxRoot:     txRoot,
				StateRoot:  StateRoot,
				Height:     b.BClaims.Height + 1,
				ChainID:    b.BClaims.ChainID,
				HeaderRoot: headerRoot,
				TxCount:    0,
				PrevBlock:  prevBlock,
			},
		},
	}
	return p, nil
}
