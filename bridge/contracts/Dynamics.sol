// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.13;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/interfaces/ISnapshots.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/libraries/dynamics/DoublyLinkedList.sol";
import "contracts/libraries/errors/DynamicsErrors.sol";

struct CanonicalVersion {
    uint64 major;
    uint64 minor;
    uint64 patch;
    bytes32 binaryHash;
}

/// @custom:salt Dynamics
/// @custom:deploy-type deployUpgradeable
contract Dynamics is Initializable, ImmutableSnapshots {
    using DoublyLinkedListLogic for DoublyLinkedList;

    event DeployedStorageContract(address contractAddr);
    event DynamicValueChanged(uint256 epoch, bytes rawDynamicValues);
    event NewNodeVersionAvailable(CanonicalVersion version);

    // enum to keep track of versions of the dynamic struct for the encoding and
    // decoding algorithms
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
        uint64 maxBlockSize;
        uint64 atomicSwapFee;
        // Second slot
        uint64 dataStoreFee;
        uint64 valueStoreFee;
        uint128 minScaledTransactionFee;
    }

    struct Configuration {
        uint128 MinEpochsBetweenUpdates;
        uint128 MaxEpochsBetweenUpdates;
    }

    bytes8 internal constant _UNIVERSAL_DEPLOY_CODE = 0x38585839386009f3;
    Version internal constant _currentVersion = Version.V1;

    DoublyLinkedList internal _dynamicValues;
    Configuration internal _configuration;
    CanonicalVersion internal _nodeCanonicalVersion;

    constructor() ImmutableFactory(msg.sender) ImmutableSnapshots() {}

    function initialize() public onlyFactory initializer {
        DynamicValues memory initialValues = DynamicValues(
            Version.V1,
            4000,
            4000,
            3000,
            3000,
            3000000,
            0,
            0,
            0,
            0
        );
        // minimum 2 epochs,
        uint128 MinEpochsBetweenUpdates = 2;
        // max 336 epochs (approx 1 month considering a snapshot every 2h)
        uint128 MaxEpochsBetweenUpdates = 336;
        _configuration = Configuration(MinEpochsBetweenUpdates, MaxEpochsBetweenUpdates);
        _addNode(1, initialValues);
    }

    function deployStorage(bytes calldata data) public returns (address contractAddr) {
        return _deployStorage(data);
    }

    function changeDynamicValues(uint32 relativeExecutionEpoch, DynamicValues memory newValue)
        public
        onlyFactory
    {
        _changeDynamicValues(relativeExecutionEpoch, newValue);
    }

    function updateHead(uint32 currentEpoch) public onlySnapshots {
        uint32 nextEpoch = _dynamicValues.getNextEpoch(_dynamicValues.getHead());
        if (nextEpoch != 0 && currentEpoch >= nextEpoch) {
            _dynamicValues.setHead(nextEpoch);
        }
    }

    function updateNodeVersion(
        uint32 majorVersion,
        uint32 minorVersion,
        uint32 patch,
        bytes32 binaryHash
    ) public onlyFactory {
        _updateNodeVersion(majorVersion, minorVersion, patch, binaryHash);
    }

    function setConfiguration(Configuration calldata newConfig) public onlyFactory {
        _configuration = newConfig;
    }

    function getConfiguration() public view returns (Configuration memory) {
        return _configuration;
    }

    function decodeDynamicValues(address addr) public view returns (DynamicValues memory) {
        return _decodeDynamicValues(addr);
    }

    function encodeDynamicValues(DynamicValues memory value) public pure returns (bytes memory) {
        return _encodeDynamicValues(value);
    }

    function getEncodingVersion() public pure returns (Version) {
        return _currentVersion;
    }

    function getPreviousDynamicValues(uint256 epoch) public view returns (DynamicValues memory) {
        uint256 head = _dynamicValues.getHead();
        if (head <= epoch) {
            return _decodeDynamicValues(_dynamicValues.getValue(head));
        }
        uint256 previous = _dynamicValues.getPreviousEpoch(head);
        if (previous != 0 && previous <= epoch) {
            return _decodeDynamicValues(_dynamicValues.getValue(previous));
        }
        revert DynamicsErrors.DynamicValueNotFound(epoch);
    }

    function getLatestDynamicValues() public view returns (DynamicValues memory) {
        return _decodeDynamicValues(_dynamicValues.getValue(_dynamicValues.getHead()));
    }

    function _addNode(uint32 executionEpoch, DynamicValues memory value) internal {
        bytes memory encodedData = _encodeDynamicValues(value);
        address dataAddress = _deployStorage(encodedData);
        _dynamicValues.addNode(executionEpoch, dataAddress);
        emit DynamicValueChanged(executionEpoch, encodedData);
    }

    function _deployStorage(bytes memory data) internal returns (address) {
        bytes memory deployCode = abi.encodePacked(_UNIVERSAL_DEPLOY_CODE, data);
        address addr;
        assembly {
            addr := create(0, add(deployCode, 0x20), mload(deployCode))
            if iszero(addr) {
                //if contract creation fails, we want to return any err messages
                returndatacopy(0x00, 0x00, returndatasize())
                //revert and return errors
                revert(0x00, returndatasize())
            }
        }
        emit DeployedStorageContract(addr);
        return addr;
    }

    function _decodeDynamicValues(address addr) internal view returns (DynamicValues memory values) {
        uint256 ptr;
        uint256 retPtr;
        // DynamicValues memory retValue = DynamicValues(Version.V1, 0, 0, 0, 0, 0, 0, 0, 0, 0);
        uint8[10] memory sizes = [8, 24, 32, 32, 32, 64, 64, 64, 64, 128];
        uint256 dynamicValuesTotalSize = 64;
        uint256 extCodeSize;
        assembly {
            ptr := mload(0x40)
            //need to figure out how to get this value dynamically
            retPtr := 0x1c0
            extCodeSize := extcodesize(addr)
            extcodecopy(addr, ptr, 0, extCodeSize)
        }
        if (extCodeSize == 0 || extCodeSize < dynamicValuesTotalSize) {
            revert DynamicsErrors.InvalidExtCodeSize(addr, extCodeSize);
        }
        
        for (uint8 i = 0; i < sizes.length; i++) {
            uint8 size = sizes[i];
            assembly {
                mstore(retPtr, shr(sub(256, size), mload(ptr)))
                ptr := add(ptr, div(size, 8))
                retPtr := add(retPtr, 0x20)
            }
        }
    }

    function _encodeDynamicValues(DynamicValues memory newValue)
        internal
        pure
        returns (bytes memory)
    {
        bytes memory data = abi.encodePacked(
            newValue.encoderVersion,
            newValue.messageTimeout,
            newValue.proposalTimeout,
            newValue.preVoteTimeout,
            newValue.preCommitTimeout,
            newValue.maxBlockSize,
            newValue.atomicSwapFee,
            newValue.dataStoreFee,
            newValue.valueStoreFee,
            newValue.minScaledTransactionFee
        );
        // Remove the first 32 bytes with the bytes length
        // assembly {
        //     data := add(data, 0x20)
        // }
        return data;
    }

    function _updateNodeVersion(
        uint32 majorVersion,
        uint32 minorVersion,
        uint32 patch,
        bytes32 binaryHash
    ) internal {
        CanonicalVersion memory version = CanonicalVersion(
            majorVersion,
            minorVersion,
            patch,
            binaryHash
        );
        _nodeCanonicalVersion = version;
        emit NewNodeVersionAvailable(version);
    }

    function _changeDynamicValues(uint32 relativeExecutionEpoch, DynamicValues memory newValue)
        internal
    {
        Configuration memory config = _configuration;
        if (
            relativeExecutionEpoch < config.MinEpochsBetweenUpdates ||
            relativeExecutionEpoch > config.MaxEpochsBetweenUpdates
        ) {
            revert DynamicsErrors.InvalidScheduledDate(
                relativeExecutionEpoch,
                config.MinEpochsBetweenUpdates,
                config.MaxEpochsBetweenUpdates
            );
        }
        uint32 currentEpoch = uint32(ISnapshots(_snapshotsAddress()).getEpoch());
        uint32 executionEpoch = relativeExecutionEpoch + currentEpoch;
        _addNode(executionEpoch, newValue);
    }
}

contract test is Dynamics {
    function testEnconding() public view returns(bytes memory) {
        DynamicValues memory initialValues = DynamicValues(
            Version.V1,
            4000,
            4000,
            3000,
            3000,
            3000000,
            0,
            0,
            0,
            0
        );
        return Dynamics._encodeDynamicValues(initialValues);
    }


}

// function getValueAtIndex(address storageAddr, uint8 index) public view returns (uint256) {
//     uint8[10] memory sizes = [8, 24, 32, 32, 32, 64, 64, 64, 64, 128];
//     uint32 offset = 0;
//     uint256 ptr;
//     uint256 retPtr;
//     uint8 size = sizes[index];
//     assembly {
//         ptr := mload(0x40)
//         let csize := extcodesize(storageAddr)
//         extcodecopy(storageAddr, ptr, 0, csize)
//         retPtr := add(ptr, size)
//     }
//     for (uint8 i = 0; i < index; i++) {
//         offset += sizes[i];
//     }
//     offset = offset / 8;
//     assembly {
//         mstore(retPtr, shr(sub(256, size), mload(add(ptr, offset))))
//         return(retPtr, 0x20)
//     }
// }

//  function deployStorageDeterministic(uint256 blockHeight) public {
//     address addr;
//     bytes32 salt_ = bytes32(blockHeight);
//     assembly {
//         let ptr := mload(0x40)
//         mstore(ptr, shl(48, 0x5863e8c0cf5a60e01b81528081602083335AFA3d82833e3d82f3))
//         addr := create2(0, ptr, add(ptr, 0x1a), salt_)
//         //if the returndatasize is not 0 revert with the error message
//         if iszero(iszero(returndatasize())) {
//             returndatacopy(0x00, 0x00, returndatasize())
//             revert(0, returndatasize())
//         }
//         //if contractAddr or code size at contractAddr is 0 revert with deploy fail message
//         if or(iszero(addr), iszero(extcodesize(addr))) {
//             mstore(0, "storage deployment failed")
//             revert(0, 0x20)
//         }
//     }
//     emit DeployedStorageContract(addr);
// }

// function decodeBlobDeterministic(address addr) public view returns (uint32) {
//     uint8[1] memory sizes = [32];
//     uint256 ptr;
//     uint256 retPtr;
//     uint256 extCodeSize;
//     assembly {
//         ptr := mload(0x40)
//         let offset := 0x1d
//         let size := sub(extcodesize(addr), offset)
//         extcodecopy(addr, ptr, 0x1d, size)
//         retPtr := add(ptr, size)
//     }
//     if (extCodeSize == 0) {
//         revert DynamicsErrors.ExtCodeSizeZero(addr);
//     }
//     for (uint8 i = 0; i < sizes.length; i++) {
//         uint8 size = sizes[i];
//         assembly {
//             mstore(retPtr, shr(sub(256, size), mload(ptr)))
//             ptr := add(ptr, div(size, 8))
//             retPtr := add(retPtr, 0x20)
//         }
//     }
// }

// function getValueAtIndexDeterministic(address storageAddr, uint8 index)
//     public
//     view
//     returns (uint256)
// {
//     uint8[1] memory sizes = [32];
//     uint32 offset = 0;
//     uint256 ptr;
//     uint256 retPtr;
//     uint8 size = sizes[index];
//     assembly {
//         ptr := mload(0x40)
//         let extOffset := 0x1d
//         extcodecopy(storageAddr, ptr, extOffset, sub(extcodesize(storageAddr), extOffset))
//         retPtr := add(ptr, size)
//     }
//     for (uint8 i = 0; i < index; i++) {
//         offset += sizes[i];
//     }
//     offset = offset / 8;
//     assembly {
//         mstore(retPtr, shr(sub(256, size), mload(add(ptr, offset))))
//         return(retPtr, 0x20)
//     }
// }

// function getStorageCode() external view returns (bytes memory) {
//     ConstantValues memory values = constantFees;
//     address self = factoryAddress;
//     bytes memory data = abi.encodePacked(
//         hex"73",
//         self,
//         hex"3303601c5733ff5b",
//         values.tokenDepositFee
//     );
//     assembly {
//         data := add(data, 0x20)
//     }
// }
