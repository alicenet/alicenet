// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

interface IBridgePool {
    function initialize(address ercContract_) external;

    function deposit(address msgSender, bytes calldata depositParameters) external;

    function withdraw(bytes memory encodedBurnedUTXO, bytes memory encodedMerkleProof) external;
}
