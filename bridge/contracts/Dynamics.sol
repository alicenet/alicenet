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
    function deployStorage(bytes calldata data)public returns(address contractAddr){
        bytes memory deployCode = abi.encodePacked(
            _UNIVERSAL_DEPLOY_CODE,
            data
        );
        assembly{
            contractAddr := create(0,add(deployCode, 0x20), deployCode)
        }
    }
    function decodeBlob(address addr)public view returns (ValueTypes memory){
        uint8[12] memory sizes = [32,16,16,16,16,16,16,128,64,64,64,64];
        uint256 ptr;
        uint256 retPtr;
        assembly{
            ptr := mload(0x40)
            retPtr := 0x80
            let size := extcodesize(addr)
            extcodecopy(addr, ptr, 0, size)
        }
        for(uint8 i = 0; i < sizes.length; i++){
            uint8 size = sizes[i];
            assembly{
                mstore(retPtr, shr(sub(256,size),mload(ptr)))
                ptr := add(ptr, div(size,8))
                retPtr := add(retPtr, 0x20)
            }
        }
    }

    function decodeDeterministicBlob(address addr) public view returns(ConstantValues memory){
        uint8[1] memory sizes = [32];
        uint256 ptr;
        uint256 retPtr;
        assembly{
            ptr := mload(0x40)
            retPtr := 0x80
            let offset := 0x13
            let size := sub(extcodesize(addr), offset)
            extcodecopy(addr, ptr, 0x1d, size)
        }
        for(uint8 i = 0; i < sizes.length; i++){
            uint8 size = sizes[i];
            assembly{
                mstore(retPtr, shr(sub(256,size),mload(ptr)))
                ptr := add(ptr, div(size,8))
                retPtr := add(retPtr, 0x20)
            }
        }
    }
    function getValueAtIndex(address storageAddr, uint8 index) public view returns (uint256){
        uint8[12] memory sizes = [32,16,16,16,16,16,16,128,64,64,64,64];
        uint32 offset = 0;
        uint256 ptr;
        uint8 size = sizes[index];
        assembly{
            ptr := mload(0x40)
            let csize := extcodesize(storageAddr)
            extcodecopy(storageAddr, ptr, 0, csize)
        }
        for(uint8 i = 0; i < index; i++){
            offset = offset + sizes[i];
        }
        offset = offset/8;
        assembly{
            mstore(0x80, shr(sub(256,size), mload(add(ptr,offset))))
            return(0x80, 0x20)
        }

    }

    function getValueAtIndexDeterministic(address storageAddr, uint8 index) public view returns (uint256){
        uint8[1] memory sizes = [32];
        uint32 offset = 0;
        uint256 ptr;
        uint8 size = sizes[index];
        assembly{
            ptr := mload(0x40)
            let extOffset := 0x1d
            extcodecopy(storageAddr, ptr, extOffset, sub(extcodesize(storageAddr), extOffset))
        }
        for(uint8 i = 0; i < index; i++){
            offset = offset + sizes[i];
        }
        offset = offset/8;
        assembly{
            mstore(0x80, shr(sub(256,size), mload(add(ptr,offset))))
            return(0x80, 0x20)
        }
    }

    function deployStorageDeterministic(uint256 blockheight) public{
        address addr;
        //FeeStorage byte string
        bytes32 salt_ = bytes32(blockheight);
        assembly {
            let ptr := mload(0x40)
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
                mstore(0, "storage deployment failed")
                revert(0, 0x20)
            }
        }
        emit DeployedStorage(addr);
    }

     function getStorageCode() external view returns(bytes memory){
        ConstantValues memory values = constantFees;
        address self = factoryAddress;
        bytes memory data = abi.encodePacked(hex"73", self, hex"3303601c5733ff5b", values.tokenDepositFee);
        assembly{
            return(add(data,0x20), mload(data))
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
