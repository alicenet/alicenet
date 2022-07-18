// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library BridgePoolErrorCodes {
    // BridgePool error codes
    bytes32 public constant BRIDGEPOOL_RECEIVER_IS_NOT_OWNER_ON_PROOF_OF_BURN_UTXO = "2400";
    bytes32 public constant BRIDGEPOOL_COULD_NOT_VERIFY_PROOF_OF_BURN = "2401";
}
