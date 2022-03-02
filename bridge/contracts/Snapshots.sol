// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/ISnapshots.sol";
import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/IETHDKG.sol";
import "contracts/libraries/parsers/RCertParserLibrary.sol";
import "contracts/libraries/parsers/BClaimsParserLibrary.sol";
import "contracts/libraries/math/CryptoLibrary.sol";
import "contracts/libraries/snapshots/SnapshotsStorage.sol";
import "contracts/utils/DeterministicAddress.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

/// @custom:salt Snapshots
/// @custom:deploy-type deployUpgradeable
contract Snapshots is Initializable, SnapshotsStorage, ISnapshots {
    constructor(uint256 chainID_, uint256 epochLength_) SnapshotsStorage(chainID_, epochLength_) {}

    function initialize(uint32 desperationDelay_, uint32 desperationFactor_)
        public
        onlyFactory
        initializer
    {
        _snapshotDesperationDelay = desperationDelay_;
        _snapshotDesperationFactor = desperationFactor_;
    }

    function setSnapshotDesperationDelay(uint32 desperationDelay_) public onlyFactory {
        _snapshotDesperationDelay = desperationDelay_;
    }

    function getSnapshotDesperationDelay() public view returns (uint256) {
        return _snapshotDesperationDelay;
    }

    function setSnapshotDesperationFactor(uint32 desperationFactor_) public onlyFactory {
        _snapshotDesperationFactor = desperationFactor_;
    }

    function getSnapshotDesperationFactor() public view returns (uint256) {
        return _snapshotDesperationFactor;
    }

    function getChainId() public view returns (uint256) {
        return _chainId;
    }

    function getEpoch() public view returns (uint256) {
        return _epoch;
    }

    function getEpochLength() public view returns (uint256) {
        return _epochLength;
    }

    function getChainIdFromSnapshot(uint256 epoch_) public view returns (uint256) {
        return _snapshots[epoch_].blockClaims.chainId;
    }

    function getChainIdFromLatestSnapshot() public view returns (uint256) {
        return _snapshots[_epoch].blockClaims.chainId;
    }

    function getBlockClaimsFromSnapshot(uint256 epoch_)
        public
        view
        returns (BClaimsParserLibrary.BClaims memory)
    {
        return _snapshots[epoch_].blockClaims;
    }

    function getBlockClaimsFromLatestSnapshot()
        public
        view
        returns (BClaimsParserLibrary.BClaims memory)
    {
        return _snapshots[_epoch].blockClaims;
    }

    function getCommittedHeightFromSnapshot(uint256 epoch_) public view returns (uint256) {
        return _snapshots[epoch_].committedAt;
    }

    function getCommittedHeightFromLatestSnapshot() public view returns (uint256) {
        return _snapshots[_epoch].committedAt;
    }

    function getMadnetHeightFromSnapshot(uint256 epoch_) public view returns (uint256) {
        return _snapshots[epoch_].blockClaims.height;
    }

    function getMadnetHeightFromLatestSnapshot() public view returns (uint256) {
        return _snapshots[_epoch].blockClaims.height;
    }

    function getSnapshot(uint256 epoch_) public view returns (Snapshot memory) {
        return _snapshots[epoch_];
    }

    function getLatestSnapshot() public view returns (Snapshot memory) {
        return _snapshots[_epoch];
    }

    /// @notice Saves next snapshot
    /// @param groupSignature_ The group signature used to sign the snapshots' block claims
    /// @param bClaims_ The claims being made about given block
    /// @return Flag whether we should kick off another round of key generation
    function snapshot(bytes calldata groupSignature_, bytes calldata bClaims_)
        public
        returns (bool)
    {
        require(
            IValidatorPool(_ValidatorPoolAddress()).isValidator(msg.sender),
            "Snapshots: Only validators allowed!"
        );
        require(
            IValidatorPool(_ValidatorPoolAddress()).isConsensusRunning(),
            "Snapshots: Consensus is not running!"
        );

        (bool success, uint256 validatorIndex) = IETHDKG(_ETHDKGAddress()).tryGetParticipantIndex(
            msg.sender
        );
        require(success, "Snapshots: Caller didn't participate in the last ethdkg round!");
        // todo: critical! add eth min blocks between snapshots

        //todo: are we going to snapshot on epoch 0?
        uint32 epoch = _epoch + 1;
        // todo: explicitly verify min eth boundary
        // uint256 ethBlocksSinceLastSnapshot = block.number - _snapshots[epoch - 1].committedAt;

        // TODO: BRING BACK AFTER GOLANG LOGIC IS DEBUGED AND MERGED
        /*
        uint256 blocksSinceDesperation = ethBlocksSinceLastSnapshot >= _snapshotDesperationDelay
            ? ethBlocksSinceLastSnapshot - _snapshotDesperationDelay
            : 0;
        */

        // Check if sender is the elected validator allowed to make the snapshot
        // TODO: BRING BACK AFTER GOLANG LOGIC IS DEBUGED AND MERGED
        /*
        require(
            _mayValidatorSnapshot(
                IValidatorPool(_ValidatorPoolAddress()).getValidatorsCount(),
                validatorIndex - 1,
                blocksSinceDesperation,
                keccak256(bClaims_),
                uint256(_snapshotDesperationFactor)
            ),
            "Snapshots: Validator not elected to do snapshot!"
        );
        */

        {
            (uint256[4] memory masterPublicKey, uint256[2] memory signature) = RCertParserLibrary
                .extractSigGroup(groupSignature_, 0);

            require(
                keccak256(abi.encodePacked(masterPublicKey)) ==
                    keccak256(abi.encodePacked(IETHDKG(_ETHDKGAddress()).getMasterPublicKey())),
                "Snapshots: Wrong master public key!"
            );

            require(
                CryptoLibrary.Verify(abi.encodePacked(keccak256(bClaims_)), signature, masterPublicKey),
                "Snapshots: Signature verification failed!"
            );
        }

        BClaimsParserLibrary.BClaims memory blockClaims = BClaimsParserLibrary.extractBClaims(
            bClaims_
        );

        require(
            epoch * _epochLength == blockClaims.height,
            "Snapshots: Incorrect Madnet height for snapshot!"
        );

        require(blockClaims.chainId == _chainId, "Snapshots: Incorrect chainID for snapshot!");

        bool isSafeToProceedConsensus = true;
        if (IValidatorPool(_ValidatorPoolAddress()).isMaintenanceScheduled()) {
            isSafeToProceedConsensus = false;
            IValidatorPool(_ValidatorPoolAddress()).pauseConsensus();
        }

        _snapshots[epoch] = Snapshot(block.number, blockClaims);
        _epoch = epoch;

        emit SnapshotTaken(
            _chainId,
            epoch,
            blockClaims.height,
            msg.sender,
            isSafeToProceedConsensus,
            groupSignature_
        );
        return isSafeToProceedConsensus;
    }

    function mayValidatorSnapshot(
        uint256 numValidators,
        uint256 myIdx,
        uint256 blocksSinceDesperation,
        bytes32 blsig,
        uint256 desperationFactor
    ) public pure returns (bool) {
        return
            _mayValidatorSnapshot(
                numValidators,
                myIdx,
                blocksSinceDesperation,
                blsig,
                desperationFactor
            );
    }

    function _mayValidatorSnapshot(
        uint256 numValidators,
        uint256 myIdx,
        uint256 blocksSinceDesperation,
        bytes32 blsig,
        uint256 desperationFactor
    ) internal pure returns (bool) {
        uint256 numValidatorsAllowed = 1;

        uint256 desperation = 0;
        while (desperation < blocksSinceDesperation && numValidatorsAllowed <= numValidators / 3) {
            desperation += desperationFactor / numValidatorsAllowed;
            numValidatorsAllowed++;
        }

        uint256 rand = uint256(blsig);
        uint256 start = (rand % numValidators);
        uint256 end = (start + numValidatorsAllowed) % numValidators;

        if (end > start) {
            return myIdx >= start && myIdx < end;
        } else {
            return myIdx >= start || myIdx < end;
        }
    }
}
