// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

contract BridgeRouterMock {
    function routeDeposit(
        address account,
        bytes calldata data
    ) external returns (uint256) {
        account = account;
        maxTokens = maxTokens;
        data = data;
        return 1000;
    }
}
