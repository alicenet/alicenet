// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

interface IBridgePoolRouter {
    struct DepositReturnData {
        bytes32[] topics;
        bytes logData;
        uint256 fee;
    }

    function routeDeposit(address account, bytes calldata data) external returns (bytes memory);
}
