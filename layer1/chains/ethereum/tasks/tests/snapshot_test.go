//go:build integration

package tests

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/crypto/bn256"
	dkgState "github.com/alicenet/alicenet/layer1/chains/ethereum/tasks/dkg/state"
	"github.com/alicenet/alicenet/layer1/chains/ethereum/tasks/snapshots"
	"github.com/alicenet/alicenet/layer1/chains/ethereum/tasks/snapshots/state"
	"github.com/alicenet/alicenet/layer1/tests"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	chainID = 1337
)

func Test_SnapshotTask(t *testing.T) {
	valCount := 5
	fixture, suite := CompleteEthDkgCeremony(t, valCount)
	eth := fixture.Client
	ctx := context.Background()

	bnSigners := make([]*crypto.BNGroupSigner, 0)
	for idx := 0; idx < valCount; idx++ {
		dkgState, err := dkgState.GetDkgState(suite.DKGStatesDbs[idx])
		require.Nil(t, err)
		signer := &crypto.BNGroupSigner{}
		err = signer.SetPrivk(dkgState.GroupPrivateKey.Bytes())
		require.Nil(t, err)
		bnSigners = append(bnSigners, signer)
		groupKey, err := bn256.MarshalBigIntSlice(dkgState.MasterPublicKey[:])
		if err != nil {
			t.Fatal(err)
		}
		err = signer.SetGroupPubk(groupKey)
		require.Nil(t, err)
	}

	// Valid at 1024
	dkgState, err := dkgState.GetDkgState(suite.DKGStatesDbs[0])
	require.Nil(t, err)
	grpSig1024, bClaimsBin1024, err := GenerateSnapshotData(
		t,
		chainID,
		constants.EpochLength,
		bnSigners,
		valCount,
		dkgState.MasterPublicKey[:],
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	bclaims := &objs.BClaims{}
	err = bclaims.UnmarshalBinary(bClaimsBin1024)
	require.Nil(t, err)

	bh := &objs.BlockHeader{
		BClaims:  bclaims,
		TxHshLst: [][]byte{},
		SigGroup: grpSig1024,
	}

	currentHeight, err := eth.GetCurrentHeight(ctx)
	assert.Nil(t, err)

	for idx := 0; idx < valCount; idx++ {
		snapshotState := &state.SnapshotState{
			Account:     eth.GetDefaultAccount(),
			BlockHeader: bh,
		}

		err := state.SaveSnapshotState(suite.DKGStatesDbs[idx], snapshotState)
		require.Nil(t, err)
	}

	for idx := 0; idx < valCount; idx++ {
		snapshotTask := snapshots.NewSnapshotTask(currentHeight, valCount, idx)

		err := snapshotTask.Initialize(
			suite.DKGStatesDbs[idx],
			fixture.Logger,
			suite.Eth,
			fixture.Contracts,
			"snapshotTask",
			"task-id",
			0,
			0,
			false,
			nil,
			nil,
		)
		assert.Nil(t, err)
		err = snapshotTask.Prepare(ctx)
		assert.Nil(t, err)

		snapshotState, err := state.GetSnapshotState(snapshotTask.GetDB())
		assert.Nil(t, err)

		shouldExecute, err := snapshotTask.ShouldExecute(ctx)
		assert.Nil(t, err)
		if shouldExecute {
			txn, taskError := snapshotTask.Execute(ctx)
			amILeading, err := utils.AmILeading(
				suite.Eth,
				ctx,
				fixture.Logger,
				snapshotState.LastSnapshotHeight,
				snapshotState.RandomSeedHash,
				snapshotTask.NumOfValidators,
				snapshotTask.ValidatorIndex,
				snapshotState.DesperationFactor,
				snapshotState.DesperationDelay,
			)
			assert.Nil(t, err)
			if amILeading {
				assert.Nil(t, taskError)
				rcptResponse, err := fixture.Watcher.Subscribe(ctx, txn, nil)
				assert.Nil(t, err)
				tests.WaitGroupReceipts(t, suite.Eth, []transaction.ReceiptResponse{rcptResponse})
			} else {
				assert.Nil(t, txn)
				require.NotNil(t, taskError)
				assert.True(t, taskError.IsRecoverable())
			}
		}
	}
}

func Test_LeaderElectionWithRandomData(t *testing.T) {
	t.Parallel()
	valMin := 4
	valMax := 50
	fixture := setupEthereum(t, 5)
	eth := fixture.Client
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		valNum := getRandomNumberInRange(i, valMin, valMax)
		valIndex := getRandomNumberInRange(i, 0, valNum)
		hash := []byte(uuid.New().String())
		keccakedHash := ethCrypto.Keccak256(hash)
		var hashSlice32 [32]byte
		copy(hashSlice32[:], keccakedHash)

		currentHeight, err := eth.GetCurrentHeight(ctx)
		assert.Nil(t, err)

		blocksToMine := getRandomNumberInRange(i, 1, 1000)
		advanceToBlock := currentHeight + uint64(blocksToMine)
		tests.AdvanceTo(eth, advanceToBlock)

		accounts := eth.GetKnownAccounts()
		owner := accounts[0]
		callOpts, err := eth.GetCallOpts(ctx, owner)
		assert.Nil(t, err)

		start := getRandomNumberInRange(i, int(advanceToBlock)-128, int(advanceToBlock))
		desperationFactor, err := fixture.Contracts.EthereumContracts().
			Snapshots().
			GetSnapshotDesperationFactor(callOpts)
		assert.Nil(t, err)

		desperationDelay, err := fixture.Contracts.EthereumContracts().
			Snapshots().
			GetSnapshotDesperationDelay(callOpts)
		assert.Nil(t, err)

		blocksSinceDesperation := utils.GetBlocksSinceDesperation(
			int(advanceToBlock),
			start,
			int(desperationDelay.Int64()),
		)
		goResult, err := utils.AmILeading(
			eth,
			ctx,
			fixture.Logger,
			start,
			keccakedHash,
			valNum,
			valIndex,
			int(desperationFactor.Int64()),
			int(desperationDelay.Int64()),
		)
		assert.Nil(t, err)

		solidityResult, err := fixture.Contracts.EthereumContracts().
			Snapshots().
			MayValidatorSnapshot(callOpts, big.NewInt(int64(valNum)), big.NewInt(int64(valIndex)), big.NewInt(int64(blocksSinceDesperation)), hashSlice32, desperationFactor)
		assert.Nil(t, err)

		assert.Equal(t, solidityResult, goResult)
	}
}

func Test_LeaderElectionWithFixedData(t *testing.T) {
	t.Parallel()
	start := 1
	nValidator := 10
	desperationDelay := 10
	desperationFactor := 40
	// this groupSignatureHash happens to coincide with a starting index of 7 in the case of 10 validators
	groupSignatureHash := common.HexToHash(
		"0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563",
	)
	eth := mocks.NewMockClient()
	ctx := context.Background()
	logger := logging.GetLogger("test").WithField("", "")

	//blocksSinceDesperation = 0
	iValidator := 7
	eth.GetCurrentHeightFunc.SetDefaultReturn(uint64(11), nil)
	result, err := utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 8
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.False(t, result)

	//blocksSinceDesperation = 1
	iValidator = 7
	eth.GetCurrentHeightFunc.SetDefaultReturn(uint64(12), nil)
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 8
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 9
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.False(t, result)

	//blocksSinceDesperation = 40
	iValidator = 7
	eth.GetCurrentHeightFunc.SetDefaultReturn(uint64(51), nil)
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 8
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 9
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.False(t, result)

	//blocksSinceDesperation = 41
	iValidator = 7
	eth.GetCurrentHeightFunc.SetDefaultReturn(uint64(52), nil)
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 8
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 9
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 0
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.False(t, result)

	//blocksSinceDesperation = desperationFactor + math.floor(desperationFactor / 2)
	eth.GetCurrentHeightFunc.SetDefaultReturn(uint64(71), nil)
	iValidator = 7
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 8
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 9
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 0
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.False(t, result)

	//blocksSinceDesperation = desperationFactor + math.floor(desperationFactor / 2) + 1
	eth.GetCurrentHeightFunc.SetDefaultReturn(uint64(72), nil)
	iValidator = 7
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 8
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 9
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 0
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 1
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.False(t, result)

	//blocksSinceDesperation = 100000
	eth.GetCurrentHeightFunc.SetDefaultReturn(uint64(100011), nil)
	iValidator = 7
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 8
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 9
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 0
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 1
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 2
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 3
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 4
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 5
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 6
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	//blocksSinceDesperation = 1
	eth.GetCurrentHeightFunc.SetDefaultReturn(uint64(12), nil)
	groupSignatureHash = common.HexToHash(
		"0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
	)

	iValidator = 1
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.False(t, result)

	iValidator = 2
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 3
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.True(t, result)

	iValidator = 4
	result, err = utils.AmILeading(
		eth,
		ctx,
		logger,
		start,
		groupSignatureHash.Bytes(),
		nValidator,
		iValidator,
		desperationFactor,
		desperationDelay,
	)
	assert.Nil(t, err)
	assert.False(t, result)
}

func getRandomNumberInRange(seed, min, max int) int {
	rand.Seed(int64(seed))
	return rand.Intn(max-min) + min
}

func GenerateSnapshotData(
	t *testing.T,
	chainID uint32,
	height uint32,
	bnSigners []*crypto.BNGroupSigner,
	n int,
	mpkI []*big.Int,
	fakeSig bool,
) ([]byte, []byte, error) {
	t.Helper()
	bclaims := &objs.BClaims{
		ChainID:    chainID,
		Height:     height,
		TxCount:    0,
		PrevBlock:  crypto.Hasher([]byte("")),
		TxRoot:     crypto.Hasher([]byte("")),
		StateRoot:  crypto.Hasher([]byte("")),
		HeaderRoot: crypto.Hasher([]byte("")),
	}

	blockHash, err := bclaims.BlockHash()
	require.Nil(t, err)

	bClaimsBin, err := bclaims.MarshalBinary()
	require.Nil(t, err)

	var grpsig []byte
	if fakeSig {
		grpsig, err = bnSigners[0].Sign(blockHash)
		require.Nil(t, err)
	} else {
		grpsig, err = GenerateBlockSignature(t, bnSigners, n, blockHash, mpkI)
		require.Nil(t, err)
	}

	bnVal := &crypto.BNGroupValidator{}
	_, err = bnVal.Validate(blockHash, grpsig)
	require.Nil(t, err)

	return grpsig, bClaimsBin, nil
}

func GenerateBlockSignature(
	t *testing.T,
	bnSigners []*crypto.BNGroupSigner,
	n int,
	blockHash []byte,
	mpkI []*big.Int,
) ([]byte, error) {
	t.Helper()
	sigs := [][]byte{}
	groupShares := [][]byte{}
	for idx := 0; idx < n; idx++ {
		sig, err := bnSigners[idx].Sign(blockHash)
		require.Nil(t, err)

		fmt.Printf("Sig: %x\n", sig)
		sigs = append(sigs, sig)
		pkShare, err := bnSigners[idx].PubkeyShare()
		require.Nil(t, err)

		groupShares = append(groupShares, pkShare)
		fmt.Printf("Pkshare: %x\n", pkShare)
	}
	s := new(crypto.BNGroupSigner)
	mpk, err := bn256.MarshalBigIntSlice(mpkI)
	assert.Nil(t, err)
	err = s.SetGroupPubk(mpk)
	require.Nil(t, err)

	// Finally submit signature
	grpsig, err := s.Aggregate(sigs, groupShares)
	require.Nil(t, err)
	return grpsig, nil

}
