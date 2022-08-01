// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

contract BridgeRouterMock {
    function routeDeposit(
        address account,
        uint256 maxTokens,
        bytes calldata data
    ) external returns (uint256) {
        return 1000;
    }
}
