// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

interface ICentralBridgeRouter {
    struct EventData {
        bytes32[] topics;
        bytes logData;
    }

    struct DepositReturnData {
        EventData[] eventData;
        uint256 fee;
    }

    struct RouterConfig {
        address routerAddress;
        bool notOnline;
    }

    function forwardDeposit(
        address msgSender_,
        uint16 poolVersion_,
        bytes calldata depositData_
    ) external returns (uint256);

    function disableRouter(uint16 routerVersion_) external;

    function getRouterCount() external view returns (uint16);

    function getRouterAddress(uint16 routerVersion_) external view returns (address);
}
