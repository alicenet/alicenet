// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

interface IBridgePoolFactory {
    function deployNewExternalPool(
        uint8 tokenType_,
        address ercContract_,
        uint16 poolVersion_,
        bytes calldata initCallData
    ) external;

    function deployNewNativePool(
        uint8 tokenType_,
        address ercContract_,
        uint16 poolVersion_,
        bytes calldata initCallData
    ) external;

    function deployNewArbitraryPool(
        uint8 poolType_,
        uint8 tokenType_,
        address ercContract_,
        uint16 poolVersion_,
        bytes calldata initCallData
    ) external;

    function deployPoolLogic(
        uint8 poolType_,
        uint8 tokenType_,
        uint256 value_,
        bytes calldata deployCode_
    ) external;

    function togglePublicPoolDeployment() external;

    function lookupBridgePoolAddress(bytes32 bridgePoolSalt_) external;

    function getFactoryAddress() external;

    function getBridgePoolSalt(
        address tokenContractAddr_,
        uint8 tokenType_,
        uint256 chainID_,
        uint16 version_
    ) external;
}
