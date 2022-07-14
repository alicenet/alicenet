// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

interface IBridgePool {
    function initialize(address ercContract_) external;

    function deposit(
        uint8 accountType_,
        address owner,
        uint256 number
    ) external;

    function withdraw(bytes memory encodedMerkleProof, bytes memory encodedBurnedUTXO) external;
}
