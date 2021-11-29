package blockchain_test

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/stretchr/testify/assert"
)

const SnapshotTakenSelector string = "0x6d438b6b835d16cdae6efdc0259fdfba17e6aa32dae81863a2467866f85f724a"

func TestSnapshot(t *testing.T) {
	rawBlockHeaderString := "" +
		"000000000000030008000000010004005900000002060000b500000002000000" +
		"2a000000004000000d0000000201000019000000020100002500000002010000" +
		"31000000020100007e06a605256de00205be97e3db46a7179d10baa270991a68" +
		"82adff2b3ca99d37c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b" +
		"7bfad8045d85a470000000000000000000000000000000000000000000000000" +
		"00000000000000007682aa2f2a0cacceb6abbb88b081b76481dd2704ceb42194" +
		"bb4d7aa8e41759110a1673b5fb0848a5fea6fb60aa3d013df90d1797f8b5511c" +
		"242f1c4060cbf32512443fa842e474f906eb7aedbff7a2a20818b277ef9e9fed" +
		"bae4d4012cdd476021b1d4a7f125e9199e945f602942928ccebfe5f76822bce2" +
		"c25b05da413cf9431097b5fc8ed39f381362375f1de1680cdd0525c59a76959b" +
		"b91deac7590ecdd12686f605b19f284323f20d30a2b1aa5333f7471acc3787a1" +
		"c9b24fed41717ba612f6f612c92fdee07fd6636ed067a0262971ace406b1242a" +
		"7c41397d34b642ed"

	// Just make sure it unmarshals as expected
	rawBlockHeader, err := hex.DecodeString(rawBlockHeaderString)
	assert.Nil(t, err)
	assert.Equal(t, 392, len(rawBlockHeader))

	t.Logf("rawBlockHeader: %x", rawBlockHeader)

	blockHeader := &objs.BlockHeader{}
	err = blockHeader.UnmarshalBinary(rawBlockHeader)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, uint32(42), blockHeader.BClaims.ChainID)
	assert.Equal(t, uint32(16384), blockHeader.BClaims.Height)
	assert.Equal(t, uint32(0), blockHeader.BClaims.TxCount)

	// pull out the block claims
	bclaims := blockHeader.BClaims
	rawBclaims, err := bclaims.MarshalBinary()
	assert.Nil(t, err)
	t.Logf("rawBclaims: %x", rawBclaims)

	// pull out the sig
	rawSigGroup := blockHeader.SigGroup
	assert.Equal(t, rawSigGroup, rawSigGroup)

	publicKeyG2, signatureG1, err := cloudflare.UnmarshalSignature(rawSigGroup)
	assert.Nil(t, err)

	publicKey, err := bn256.G2ToBigIntArray(publicKeyG2)
	assert.Nil(t, err)

	for idx := 0; idx < 4; idx++ {
		t.Logf("publicKey[%d]: %x", idx, publicKey[idx])
	}

	signature, err := bn256.G1ToBigIntArray(signatureG1)
	assert.Nil(t, err)

	for idx := 0; idx < 2; idx++ {
		t.Logf("signature[%d]: %x", idx, signature[idx])
	}

	fmt.Printf("rawBclaims: %x\n", rawBclaims)
	bhsh := crypto.Hasher(rawBclaims)
	// fmt.Printf("blockHash: %x", )
	assert.Nil(t, err)

	ok, err := cloudflare.Verify(bhsh, signatureG1, publicKeyG2, cloudflare.HashToG1)
	assert.Nil(t, err)
	assert.True(t, ok)

	// Check validity with Crypto
	eth, err := setupEthereum(t, 4)
	assert.Nil(t, err)

	c := eth.Contracts()
	ctx := context.TODO()
	acct := eth.GetDefaultAccount()
	callOpts := eth.GetCallOpts(ctx, acct)
	txnOpts, err := eth.GetTransactionOpts(ctx, acct)
	assert.Nil(t, err)

	good, err := c.Crypto().Verify(callOpts, bhsh, signature, publicKey)
	assert.Nil(t, err)
	assert.True(t, good)

	txn, err := c.Validators().Snapshot(txnOpts, rawSigGroup, rawBclaims)
	assert.Nil(t, err)
	assert.NotNil(t, txn)
	eth.Commit()

	rcpt, err := eth.Queue().QueueAndWait(context.Background(), txn)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), rcpt.Status)

	// Look for the snapshot taken event
	foundIt := false
	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == SnapshotTakenSelector {
			snapshotTaken, err := c.Validators().ParseSnapshotTaken(*log)
			assert.Nil(t, err)
			assert.Equal(t, uint64(1), snapshotTaken.Epoch.Uint64())
			foundIt = true

			// Now see if I can reconstruct the header from what we have
			rawEventBclaims, err := c.Validators().GetRawBlockClaimsSnapshot(callOpts, snapshotTaken.Epoch)
			assert.Nil(t, err)

			rawEventSigGroup, err := c.Validators().GetRawSignatureSnapshot(callOpts, snapshotTaken.Epoch)
			assert.Nil(t, err)

			assert.Equal(t, rawBclaims, rawEventBclaims)
			assert.Equal(t, rawSigGroup, rawEventSigGroup)

			bclaims := &objs.BClaims{}
			err = bclaims.UnmarshalBinary(rawEventBclaims)
			if err != nil {
				t.Fatal(err)
			}
			header := &objs.BlockHeader{}
			header.BClaims = bclaims
			header.SigGroup = rawEventSigGroup

			assert.Equal(t, uint32(42), header.BClaims.ChainID)
			assert.Equal(t, uint32(16384), header.BClaims.Height)
			assert.Equal(t, uint32(0), header.BClaims.TxCount)
		}
	}
	assert.True(t, foundIt, "Should have received SnapshotTaken event")
}

func TestBlockHeaderParsing(t *testing.T) {
	rawBlockHeaderString := "" +
		"000000000000030008000000010004005900000002060000b500000002000000" +
		"2a000000004000000d0000000201000019000000020100002500000002010000" +
		"31000000020100007e06a605256de00205be97e3db46a7179d10baa270991a68" +
		"82adff2b3ca99d37c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b" +
		"7bfad8045d85a470000000000000000000000000000000000000000000000000" +
		"00000000000000007682aa2f2a0cacceb6abbb88b081b76481dd2704ceb42194" +
		"bb4d7aa8e41759110a1673b5fb0848a5fea6fb60aa3d013df90d1797f8b5511c" +
		"242f1c4060cbf32512443fa842e474f906eb7aedbff7a2a20818b277ef9e9fed" +
		"bae4d4012cdd476021b1d4a7f125e9199e945f602942928ccebfe5f76822bce2" +
		"c25b05da413cf9431097b5fc8ed39f381362375f1de1680cdd0525c59a76959b" +
		"b91deac7590ecdd12686f605b19f284323f20d30a2b1aa5333f7471acc3787a1" +
		"c9b24fed41717ba612f6f612c92fdee07fd6636ed067a0262971ace406b1242a" +
		"7c41397d34b642ed"

	// Convert the string to binary and make a copy for comparison later
	rawBlockHeader, err := hex.DecodeString(rawBlockHeaderString)
	assert.Nil(t, err)

	clonedRawBlockHeader := make([]byte, len(rawBlockHeader))
	copy(clonedRawBlockHeader, rawBlockHeader)

	// Just make sure it unmarshals as expected
	blockHeader := &objs.BlockHeader{}
	err = blockHeader.UnmarshalBinary(rawBlockHeader)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, uint32(42), blockHeader.BClaims.ChainID)
	assert.Equal(t, uint32(16384), blockHeader.BClaims.Height)
	assert.Equal(t, uint32(0), blockHeader.BClaims.TxCount)

	// Make sure unmarshal->marshal is identical
	bh, err := blockHeader.MarshalBinary()
	assert.Nil(t, err)
	for idx := 0; idx < 392; idx++ {
		assert.Equal(t, rawBlockHeader[idx], bh[idx])
	}

	// see what changes
	blockHeader.BClaims.ChainID = 42
	// blockHeader.BClaims.Height = 0x12345678
	// blockHeader.BClaims.TxCount = 0

	bh, err = blockHeader.MarshalBinary()
	assert.Nil(t, err)

	// what changed?
	differences := make(map[int]string)

	for idx := 0; idx < 392; idx++ {
		a := clonedRawBlockHeader[idx]
		b := bh[idx]
		if a != b {
			differences[idx] = fmt.Sprintf("{0x%02x -> 0x%02x}", a, b)
		}
	}
	t.Logf("Change count: %v", len(differences))
	t.Logf("     Changes: %v", differences)
}

func TestBulkProcessBlockHeaders(t *testing.T) {
	file, err := os.Open("../assets/test/blockheaders.txt")
	assert.Nil(t, err)

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		hexText := scanner.Text()
		rawBlockHeader, err := hex.DecodeString(hexText)
		assert.Nil(t, err)

		processBlockHeader(t, rawBlockHeader)
	}
}

func processBlockHeader(t *testing.T, rawBlockHeader []byte) {

	blockHeader := &objs.BlockHeader{}
	err := blockHeader.UnmarshalBinary(rawBlockHeader)
	assert.Nil(t, err)

	// pull out the block claims
	bclaims := blockHeader.BClaims
	rawBclaims, err := bclaims.MarshalBinary()
	assert.Nil(t, err)
	bclaimsHash := crypto.Hasher(rawBclaims)

	// pull out the sig
	rawSigGroup := blockHeader.SigGroup

	publicKeyG2, signatureG1, err := cloudflare.UnmarshalSignature(rawSigGroup)
	assert.Nil(t, err)

	ok, err := cloudflare.Verify(bclaimsHash, signatureG1, publicKeyG2, cloudflare.HashToG1)
	assert.Nil(t, err)
	assert.True(t, ok, "verify should return true")

	// Check validity with Crypto
	assert.Nil(t, err)

	eth, err := setupEthereum(t, 5)
	assert.Nil(t, err)
	c := eth.Contracts()
	ctx := context.TODO()
	acct := eth.GetDefaultAccount()
	callOpts := eth.GetCallOpts(ctx, acct)
	txnOpts, err := eth.GetTransactionOpts(ctx, acct)
	assert.Nil(t, err)

	// Convert from G1/G2 into []*big.Int's
	publicKey, err := bn256.G2ToBigIntArray(publicKeyG2)
	assert.Nil(t, err)

	signature, err := bn256.G1ToBigIntArray(signatureG1)
	assert.Nil(t, err)

	good, err := c.Crypto().Verify(callOpts, bclaimsHash, signature, publicKey)
	assert.Nil(t, err)
	assert.True(t, good)

	t.Logf("rawBclaims: 0x%x", rawBclaims)

	txn, err := c.Validators().Snapshot(txnOpts, rawSigGroup, rawBclaims)
	assert.Nil(t, err)
	assert.NotNil(t, txn)
	eth.Commit()

	rcpt, err := eth.Queue().QueueAndWait(context.Background(), txn)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), rcpt.Status)

	// Look for the snapshot taken event
	foundIt := false
	for _, log := range rcpt.Logs {
		if log.Topics[0].String() == SnapshotTakenSelector {
			snapshotTaken, err := c.Validators().ParseSnapshotTaken(*log)
			assert.Nil(t, err)
			assert.Equal(t, uint64(1), snapshotTaken.Epoch.Uint64())
			foundIt = true

			// Now see if I can reconstruct the header from what we have
			rawEventBclaims, err := c.Validators().GetRawBlockClaimsSnapshot(callOpts, snapshotTaken.Epoch)
			assert.Nil(t, err)

			rawEventSigGroup, err := c.Validators().GetRawSignatureSnapshot(callOpts, snapshotTaken.Epoch)
			assert.Nil(t, err)

			chainId, err := c.Validators().GetChainIdFromSnapshot(callOpts, snapshotTaken.Epoch)
			assert.Nil(t, err)

			height, err := c.Validators().GetMadHeightFromSnapshot(callOpts, snapshotTaken.Epoch)
			assert.Nil(t, err)

			assert.Equal(t, rawBclaims, rawEventBclaims)
			assert.Equal(t, rawSigGroup, rawEventSigGroup)

			bclaims := &objs.BClaims{}
			err = bclaims.UnmarshalBinary(rawEventBclaims)
			assert.Nil(t, err)

			header := &objs.BlockHeader{}
			header.BClaims = bclaims
			header.SigGroup = rawEventSigGroup

			t.Logf("ChainID:%v Height:%v TxCount:%v", bclaims.ChainID, bclaims.Height, bclaims.TxCount)

			assert.Equal(t, blockHeader.BClaims.ChainID, chainId, "ChainID isn't as expected")
			assert.Equal(t, blockHeader.BClaims.Height, height, "Height isn't as expected")
			assert.Equal(t, blockHeader.BClaims.ChainID, header.BClaims.ChainID)
			assert.Equal(t, blockHeader.BClaims.Height, header.BClaims.Height)
			assert.Equal(t, blockHeader.BClaims.TxCount, header.BClaims.TxCount)
		}
	}
	assert.True(t, foundIt, "Should have received SnapshotTaken event")
}
