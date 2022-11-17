// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

interface IBridgePool {
    function initialize(address ercContract_) external;

    function deposit(address sender, bytes calldata depositParameters) external;

    function withdraw(
        address receiver,
        bytes memory vsPreImage,
        bytes[4] memory proofs
    ) external returns (uint256 value);
}
