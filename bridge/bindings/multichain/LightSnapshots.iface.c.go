// Generated by ifacemaker. DO NOT EDIT.

package multichain

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// ILightSnapshotsCaller ...
type ILightSnapshotsCaller interface {
	// CheckBClaimsSignature is a free data retrieval call binding the contract method 0x0204dffd.
	//
	// Solidity: function checkBClaimsSignature(bytes groupSignature_, bytes bClaims_) pure returns(bool)
	CheckBClaimsSignature(opts *bind.CallOpts, groupSignature_ []byte, bClaims_ []byte) (bool, error)
	// GetAliceNetHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0xff07fc0e.
	//
	// Solidity: function getAliceNetHeightFromLatestSnapshot() view returns(uint256)
	GetAliceNetHeightFromLatestSnapshot(opts *bind.CallOpts) (*big.Int, error)
	// GetAliceNetHeightFromSnapshot is a free data retrieval call binding the contract method 0xc5e8fde1.
	//
	// Solidity: function getAliceNetHeightFromSnapshot(uint256 epoch_) view returns(uint256)
	GetAliceNetHeightFromSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (*big.Int, error)
	// GetBlockClaimsFromLatestSnapshot is a free data retrieval call binding the contract method 0xc2ea6603.
	//
	// Solidity: function getBlockClaimsFromLatestSnapshot() view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
	GetBlockClaimsFromLatestSnapshot(opts *bind.CallOpts) (BClaimsParserLibraryBClaims, error)
	// GetBlockClaimsFromSnapshot is a free data retrieval call binding the contract method 0x45dfc599.
	//
	// Solidity: function getBlockClaimsFromSnapshot(uint256 epoch_) view returns((uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32))
	GetBlockClaimsFromSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (BClaimsParserLibraryBClaims, error)
	// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
	//
	// Solidity: function getChainId() view returns(uint256)
	GetChainId(opts *bind.CallOpts) (*big.Int, error)
	// GetChainIdFromLatestSnapshot is a free data retrieval call binding the contract method 0xd9c11657.
	//
	// Solidity: function getChainIdFromLatestSnapshot() view returns(uint256)
	GetChainIdFromLatestSnapshot(opts *bind.CallOpts) (*big.Int, error)
	// GetChainIdFromSnapshot is a free data retrieval call binding the contract method 0x19f74669.
	//
	// Solidity: function getChainIdFromSnapshot(uint256 epoch_) view returns(uint256)
	GetChainIdFromSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (*big.Int, error)
	// GetCommittedHeightFromLatestSnapshot is a free data retrieval call binding the contract method 0x026c2b7e.
	//
	// Solidity: function getCommittedHeightFromLatestSnapshot() view returns(uint256)
	GetCommittedHeightFromLatestSnapshot(opts *bind.CallOpts) (*big.Int, error)
	// GetCommittedHeightFromSnapshot is a free data retrieval call binding the contract method 0xe18c697a.
	//
	// Solidity: function getCommittedHeightFromSnapshot(uint256 epoch_) view returns(uint256)
	GetCommittedHeightFromSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (*big.Int, error)
	// GetEpoch is a free data retrieval call binding the contract method 0x757991a8.
	//
	// Solidity: function getEpoch() view returns(uint256)
	GetEpoch(opts *bind.CallOpts) (*big.Int, error)
	// GetEpochFromHeight is a free data retrieval call binding the contract method 0x2eee30ce.
	//
	// Solidity: function getEpochFromHeight(uint256 height) view returns(uint256)
	GetEpochFromHeight(opts *bind.CallOpts, height *big.Int) (*big.Int, error)
	// GetEpochLength is a free data retrieval call binding the contract method 0xcfe8a73b.
	//
	// Solidity: function getEpochLength() view returns(uint256)
	GetEpochLength(opts *bind.CallOpts) (*big.Int, error)
	// GetLatestSnapshot is a free data retrieval call binding the contract method 0xd518f243.
	//
	// Solidity: function getLatestSnapshot() view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
	GetLatestSnapshot(opts *bind.CallOpts) (Snapshot, error)
	// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
	//
	// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
	GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error)
	// GetMinimumIntervalBetweenSnapshots is a free data retrieval call binding the contract method 0x42438d7b.
	//
	// Solidity: function getMinimumIntervalBetweenSnapshots() view returns(uint256)
	GetMinimumIntervalBetweenSnapshots(opts *bind.CallOpts) (*big.Int, error)
	// GetSnapshot is a free data retrieval call binding the contract method 0x76f10ad0.
	//
	// Solidity: function getSnapshot(uint256 epoch_) view returns((uint256,(uint32,uint32,uint32,bytes32,bytes32,bytes32,bytes32)))
	GetSnapshot(opts *bind.CallOpts, epoch_ *big.Int) (Snapshot, error)
	// GetSnapshotDesperationDelay is a free data retrieval call binding the contract method 0xd17fcc56.
	//
	// Solidity: function getSnapshotDesperationDelay() view returns(uint256)
	GetSnapshotDesperationDelay(opts *bind.CallOpts) (*big.Int, error)
	// GetSnapshotDesperationFactor is a free data retrieval call binding the contract method 0x7cc4cce6.
	//
	// Solidity: function getSnapshotDesperationFactor() view returns(uint256)
	GetSnapshotDesperationFactor(opts *bind.CallOpts) (*big.Int, error)
	// IsMock is a free data retrieval call binding the contract method 0x28ccaa29.
	//
	// Solidity: function isMock() pure returns(bool)
	IsMock(opts *bind.CallOpts) (bool, error)
	// IsValidatorElectedToPerformSnapshot is a free data retrieval call binding the contract method 0xc0e83e81.
	//
	// Solidity: function isValidatorElectedToPerformSnapshot(address validator, uint256 lastSnapshotCommittedAt, bytes32 groupSignatureHash) pure returns(bool)
	IsValidatorElectedToPerformSnapshot(opts *bind.CallOpts, validator common.Address, lastSnapshotCommittedAt *big.Int, groupSignatureHash [32]byte) (bool, error)
	// MayValidatorSnapshot is a free data retrieval call binding the contract method 0xf45fa246.
	//
	// Solidity: function mayValidatorSnapshot(uint256 numValidators, uint256 myIdx, uint256 blocksSinceDesperation, bytes32 blsig, uint256 desperationFactor) pure returns(bool)
	MayValidatorSnapshot(opts *bind.CallOpts, numValidators *big.Int, myIdx *big.Int, blocksSinceDesperation *big.Int, blsig [32]byte, desperationFactor *big.Int) (bool, error)
	// MigrateSnapshots is a free data retrieval call binding the contract method 0xae2728ea.
	//
	// Solidity: function migrateSnapshots(bytes[] groupSignature_, bytes[] bClaims_) pure returns(bool)
	MigrateSnapshots(opts *bind.CallOpts, groupSignature_ [][]byte, bClaims_ [][]byte) (bool, error)
}
