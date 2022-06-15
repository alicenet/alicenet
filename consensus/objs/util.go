package objs

import (
	"bytes"
	"encoding/binary"
	"fmt"

	trie "github.com/MadBase/MadNet/badgerTrie"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
)

// ExtractHR extracts the Height and Round from an interface;
// it panics on undefined types.
func ExtractHR(any interface{}) (uint32, uint32) {
	switch v := any.(type) {
	case *RoundState:
		rcert := v.RCert
		return rcert.RClaims.Height, rcert.RClaims.Round
	case *RoundStateHistoricKey:
		return v.Height, v.Round
	case *RClaims:
		return v.Height, v.Round
	case *RCert:
		rc := v.RClaims
		return rc.Height, rc.Round
	case *PClaims:
		rc := v.RCert.RClaims
		return rc.Height, rc.Round
	case *Proposal:
		rc := v.PClaims.RCert.RClaims
		return rc.Height, rc.Round
	case *PreVote:
		rc := v.Proposal.PClaims.RCert.RClaims
		return rc.Height, rc.Round
	case *PreVoteNil:
		rc := v.RCert.RClaims
		return rc.Height, rc.Round
	case *PreCommit:
		rc := v.Proposal.PClaims.RCert.RClaims
		return rc.Height, rc.Round
	case *PreCommitNil:
		rc := v.RCert.RClaims
		return rc.Height, rc.Round
	case *NRClaims:
		rc := v.RCert.RClaims
		return rc.Height, rc.Round
	case *NextRound:
		rc := v.NRClaims.RCert.RClaims
		return rc.Height, rc.Round
	case *NHClaims:
		rc := v.Proposal.PClaims.RCert.RClaims
		return rc.Height, rc.Round
	case *NextHeight:
		rc := v.NHClaims.Proposal.PClaims.RCert.RClaims
		return rc.Height, rc.Round
	case *BlockHeader:
		rc := v.BClaims
		return rc.Height, 1
	case *BClaims:
		rc := v
		return rc.Height, 1
	default:
		panic(fmt.Sprintf("ExtractHR; undefined type %T", v))
	}
}

// ExtractHCID extracts the Height and ChainID from an interface;
// it panics on undefined types.
func ExtractHCID(any interface{}) (uint32, uint32) {
	switch v := any.(type) {
	case *RCert:
		rc := v.RClaims
		return rc.Height, rc.ChainID
	case *Proposal:
		rc := v.PClaims.RCert.RClaims
		return rc.Height, rc.ChainID
	case *PreVote:
		rc := v.Proposal.PClaims.RCert.RClaims
		return rc.Height, rc.ChainID
	case *PreVoteNil:
		rc := v.RCert.RClaims
		return rc.Height, rc.ChainID
	case *PreCommit:
		rc := v.Proposal.PClaims.RCert.RClaims
		return rc.Height, rc.ChainID
	case *PreCommitNil:
		rc := v.RCert.RClaims
		return rc.Height, rc.ChainID
	case *NextRound:
		rc := v.NRClaims.RCert.RClaims
		return rc.Height, rc.ChainID
	case *NextHeight:
		rc := v.NHClaims.Proposal.PClaims.RCert.RClaims
		return rc.Height, rc.ChainID
	case *BlockHeader:
		rc := v.BClaims
		return rc.Height, rc.ChainID
	case *BClaims:
		rc := v
		return rc.Height, rc.ChainID
	default:
		panic(fmt.Sprintf("ExtractHCID; undefined type %T", v))
	}
}

// ExtractRCertAny extracts an RCert from an interface;
// it panics on undefined types.
func ExtractRCertAny(any interface{}) (*RCert, error) {
	switch v := any.(type) {
	case *BlockHeader:
		return v.GetRCert()
	default:
		return ExtractRCert(any), nil
	}
}

// ExtractRCert extracts an RCert from an interface;
// it panics on undefined types.
func ExtractRCert(any interface{}) *RCert {
	switch v := any.(type) {
	case *RoundState:
		rc := v.RCert
		return rc
	case *RCert:
		rc := v
		return rc
	case *PClaims:
		rc := v.RCert
		return rc
	case *Proposal:
		rc := v.PClaims.RCert
		return rc
	case *PreVote:
		rc := v.Proposal.PClaims.RCert
		return rc
	case *PreVoteNil:
		rc := v.RCert
		return rc
	case *PreCommit:
		rc := v.Proposal.PClaims.RCert
		return rc
	case *PreCommitNil:
		rc := v.RCert
		return rc
	case *NRClaims:
		rc := v.RCert
		return rc
	case *NextRound:
		rc := v.NRClaims.RCert
		return rc
	case *NHClaims:
		rc := v.Proposal.PClaims.RCert
		return rc
	case *NextHeight:
		rc := v.NHClaims.Proposal.PClaims.RCert
		return rc
	default:
		panic(fmt.Sprintf("ExtractRCert; undefined type %T", v))
	}
}

// RelateHR relates Height and Round between objects.
// Simply:
//		if a is before b, return -1
//		if a is after b, return 1
//		if a equals b, return 0
//
// More explicitly, we extract the height and round from objects a and b.
//
// If (aHeight == bHeight) && (aRound == bRound)
//		return 0
//
// If (aHeight < bHeight) || ((aHeight == bHeight) && (aRound < bRound))
//		return -1
//
// If (aHeight > bHeight) || ((aHeight == bHeight) && (aRound > bRound))
//		return 1
func RelateHR(a, b interface{}) int {
	ah, ar := ExtractHR(a)
	bh, br := ExtractHR(b)
	abuf := make([]byte, 8)
	binary.BigEndian.PutUint32(abuf[0:4], ah)
	binary.BigEndian.PutUint32(abuf[4:], ar)
	ahr := binary.BigEndian.Uint64(abuf)
	bbuf := make([]byte, 8)
	binary.BigEndian.PutUint32(bbuf[0:4], bh)
	binary.BigEndian.PutUint32(bbuf[4:], br)
	bhr := binary.BigEndian.Uint64(bbuf)
	if ahr < bhr {
		return -1
	}
	if ahr == bhr {
		return 0
	}
	return 1
}

// RelateH relates Height between objects
//
// If aHeight == bHeight
//		return 0
//
// If aHeight < bHeight
//		return -1
//
// If aHeight > bHeight
//		return 1
func RelateH(a, b interface{}) int {
	if a == nil && b != nil {
		return -1
	}
	if a != nil && b == nil {
		return 1
	}
	ah, _ := ExtractHR(a)
	bh, _ := ExtractHR(b)
	if ah < bh {
		return -1
	}
	if ah == bh {
		return 0
	}
	return 1
}

// BClaimsEqual determines if two objects have equal BClaims objects
func BClaimsEqual(a, b interface{}) (bool, error) {
	ab := ExtractBClaims(a)
	bb := ExtractBClaims(b)
	ahsh, err := ab.BlockHash()
	if err != nil {
		return false, err
	}
	bhsh, err := bb.BlockHash()
	if err != nil {
		return false, err
	}
	if !bytes.Equal(ahsh, bhsh) {
		return false, nil
	}
	return true, nil
}

// ExtractBClaims extracts BClaims from an interface;
// it panics on undefined types.
func ExtractBClaims(any interface{}) *BClaims {
	switch v := any.(type) {
	case *BlockHeader:
		return v.BClaims
	case *Proposal:
		return v.PClaims.BClaims
	case *PreVote:
		return v.Proposal.PClaims.BClaims
	case *PreCommit:
		return v.Proposal.PClaims.BClaims
	case *NextHeight:
		return v.NHClaims.Proposal.PClaims.BClaims
	default:
		panic(fmt.Sprintf("ExtractBClaims; undefined type %T", v))
	}
}

// PrevBlockEqual determines if objects agree on the previous block
func PrevBlockEqual(a, b interface{}) bool {
	ab := ExtractRCert(a)
	bb := ExtractRCert(b)
	return bytes.Equal(ab.RClaims.PrevBlock, bb.RClaims.PrevBlock)
}

// IsDeadBlockRound determines if an object is for the DeadBlockRound
func IsDeadBlockRound(any interface{}) bool {
	_, r := ExtractHR(any)
	return r == constants.DEADBLOCKROUND
}

// MakeTxRoot creates a txRootHsh from a list of transaction hashes
func MakeTxRoot(txHashes [][]byte) ([]byte, error) {
	if len(txHashes) == 0 {
		return crypto.Hasher([]byte{}), nil
	}
	values := [][]byte{}
	for i := 0; i < len(txHashes); i++ {
		txHash := txHashes[i]
		if len(txHash) != constants.HashLen {
			return nil, errorz.ErrInvalid{}.New("MakeTxRoot; incorrect txHash length")
		}
		values = append(values, crypto.Hasher(txHash))
	}
	// new in memory smt
	smt := trie.NewMemoryTrie()
	// smt update
	txHashesSorted, valuesSorted, err := utils.SortKVs(txHashes, values)
	if err != nil {
		return nil, err
	}
	rootHash, err := smt.Update(txHashesSorted, valuesSorted)
	if err != nil {
		return nil, err
	}
	return rootHash, nil
}

// GetProposerIdx will return the index of the proposer of this round
// from the list of validators
func GetProposerIdx(numv int, height uint32, round uint32) uint8 {
	return uint8(int(height+round-1) % numv)
}

// SplitBlob separates a blob of fixed size state types into a slice of slices
func SplitBlob(s []byte, blen int) ([][]byte, error) {
	if blen <= 0 {
		return [][]byte{}, errorz.ErrInvalid{}.New("SplitBlob; blen <= 0")
	}
	if len(s)%blen != 0 {
		return [][]byte{}, errorz.ErrInvalid{}.New("SplitBlob; invalid length")
	}
	buf := [][]byte{}
	for i := 0; i < len(s)/blen; i++ {
		b := append([]byte{}, s[i*blen:i*blen+blen]...)
		buf = append(buf, b)
	}
	return buf, nil
}

// SplitSignatures splits signatures by chopping up a blob of state into byte slices by length.
// return an error if the length is not correct. IE the total length
// is not a multiple of expected length for a single element of type.
func SplitSignatures(s []byte) ([][]byte, error) {
	return SplitBlob(s, constants.CurveSecp256k1SigLen)
}

// SplitHashes splits hashes by chopping up a blob of state into byte slices by length.
// return an error if the length is not correct. IE the total length
// is not a multiple of expected length for a single element of type.
func SplitHashes(s []byte) ([][]byte, error) {
	return SplitBlob(s, constants.HashLen)
}
