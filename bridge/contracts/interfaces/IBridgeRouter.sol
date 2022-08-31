// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

interface IBridgeRouter {
    function routeDeposit(address account, bytes calldata data) external returns (uint256);

    function getBridgePoolSalt(
        address tokenContractAddr_,
        uint8 tokenType_,
        uint256 chainID_,
        uint16 version_
    ) external pure returns (bytes32);
}
