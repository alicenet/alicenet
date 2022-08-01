// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.13;

import "contracts/interfaces/ISnapshots.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/libraries/dynamics/DoublyLinkedList.sol";

struct CanonicalVersion {
    uint64 major;
    uint64 minor;
    uint64 patch;
}

// the minimum amount of ethereum without a snapshot that we kick the validators

/// @custom:salt Dynamics
/// @custom:deploy-type deployUpgradeable
contract Dynamics is ImmutableSnapshots {
    event DeployedStorage(address contractAddr);
    event ValueChanged(uint256 epoch, ValueType valueType, int256 newValue);

    // enum ValueType {
    //     DepositFee,
    //     AtomicSwapFee,
    //     DataStoreFee,
    //     ValueStoreFee,
    //     MinTxFee,
    //     MaxProposalSize,
    //     MessageTimeout,
    //     ProposalStepTimeout,
    //     PreVoteStepTimeout,
    //     PreCommitStepTimeout,
    //     MaxAmountOfBlocksWithoutSnapshots,
    //     MinEpochsBetweenDynamicValuesUpdates
    // }

    enum Version {
        V1
    }

    struct DynamicValues {
        // first slot
        Version encoderVersion;
        uint24 messageTimeout;
        uint32 proposalTimeout;
        uint32 preVoteTimeout;
        uint32 preCommitTimeout;
        uint128 minScaledTransactionFee;
        // Second slot
        uint64 atomicSwapFee;
        uint64 dataStoreFee;
        uint64 valueStoreFee;
    }

    uint64 depositFee;
    uint16 MaxAmountOfBlocksWithoutSnapshots;
    uint16 MinEpochsBetweenDynamicValuesUpdates;

    constructor() ImmutableFactory(msg.sender) ImmutableSnapshots() {}

    function deployStorage() public {
        address addr;
        //FeeStorage byte string
        bytes32 salt_ = "0x46656553746f72616765";
        assembly {
            let ptr := mload(0x40)
            //metamorphic hash 0xae390f39abdbe0aaec7e24a97c6f672ef87c9ea259ab2d3eccf5dc2541f28c56
            mstore(ptr, shl(48, 0x5863e8c0cf5a60e01b81528081602083335AFA3d82833e3d82f3))
            //mstore(ptr, shl(136, 0x5880818283335afa3d82833e3d91f3))
            addr := create2(0, ptr, add(ptr, 0x1a), salt_)
            //if the returndatasize is not 0 revert with the error message
            if iszero(iszero(returndatasize())) {
                returndatacopy(0x00, 0x00, returndatasize())
                revert(0, returndatasize())
            }
            //if contractAddr or code size at contractAddr is 0 revert with deploy fail message
            if or(iszero(addr), iszero(extcodesize(addr))) {
                mstore(0, "static pool deploy failed")
                revert(0, 0x20)
            }
        }
        emit DeployedStorage(addr);
    }

    function getStorageCode() external view returns (bytes memory) {
        uint256 basePtr;
        address self = factoryAddress;
        Fee[] memory input = fees;
        assembly {
            let length := mload(0x80)
            let lengthPrefixOffset := sub(mload(0xa0), 0x20)
            mstore(lengthPrefixOffset, length)
            let ptr := sub(lengthPrefixOffset, 0x20)
            mstore(ptr, 0x20)
            ptr := sub(ptr, 0x20)
            basePtr := ptr
            mstore(ptr, hex"73")
            ptr := add(ptr, 0x01)
            mstore(ptr, shl(96, self))
            ptr := add(ptr, 0x14)
            mstore(ptr, hex"3303601c5733ff5b")
            return(basePtr, sub(msize(), basePtr))
        }
    }

    function setValue(
        uint256 epoch,
        uint256 newValue,
        ValueType valueType,
        Value storage value_
    ) internal {
        uint256 currentEpoch = ISnapshots(_addressSnapshots()).getEpoch();
        require(
            epoch >= currentEpoch + minEpochsBetweenDynamicValuesUpdates,
            "invalid minimum epoch between update"
        );
        // todo: add ring buffer
        value_.futureValues.push(ScheduledValue(epoch, newValue));
        emit ValueChanged(epoch, valueType, newValue);
    }

    function setAtomicSwapFee(uint256 epoch, uint256 newAtomicSwapFee) public {
        setValue(epoch, AtomicSwapFee, newAtomicSwapFee, atomicSwapFee);
    }

    function setMaxProposalSize(uint256 epoch, uint256 newMaxProposalSize) public {
        setValue(epoch, MaxProposalSize, newMaxProposalSize);
    }

    function setDataStoreEpochFee() public onlyFactory {
        setValue(epoch, MaxProposalSize, newMaxProposalSize);
    }
    // function setValueStoreFee()public onlyFactory {
    //     setValue(epoch, valueStoreFee, );
    // }
    // function setMinTxFee()public onlyFactory {
    //     setValue(epoch, MinTxFee, );
    // }

    // function setMessageTimeout()public onlyFactory {
    //     setValue(epoch, MessageTimeout, );
    // }
    // function setProposalStepTimeout()public onlyFactory {

    // }
    // function setPreVoteStepTimeout()public onlyFactory {

    // }
    // function setPreCommitStepTimeout()public onlyFactory {

    // }
}
