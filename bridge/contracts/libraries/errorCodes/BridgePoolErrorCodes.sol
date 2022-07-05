// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library BridgePoolErrorCodes {
    // BridgePool error codes
    bytes32 public constant BRIDGEPOOL_RECEIVER_IS_NOT_OWNER_ON_PROOF_OF_BURN_UTXO = "2400"; //"BridgePool: Deposit can only be requested for the owner in burned UTXO"
    bytes32 public constant BRIDGEPOOL_COULD_NOT_VERIFY_PROOF_OF_BURN = "2401"; //"BridgePool: Proof of burn in aliceNet could not be verified"
    bytes32
        public constant BRIDGEPOOL_COULD_NOT_APPROVE_ALLOWANCE_TO_TRANSFER_WITHDRAW_AMOUNT_TO_RECEIVER =
        "2402";
    bytes32 public constant BRIDGEPOOL_COULD_NOT_TRANSFER_WITHDRAW_AMOUNT_TO_RECEIVER = "2403";
    bytes32 public constant BRIDGEPOOL_UNABLE_TO_TRANSFER_DEPOSIT_AMOUNT_FROM_SENDER = "2404";
    bytes32 public constant BRIDGEPOOL_UNABLE_TO_TRANSFER_DEPOSIT_FEE_FROM_SENDER = "2405";
    bytes32 public constant BRIDGEPOOL_UNABLE_TO_BURN_DEPOSIT_FEE = "2406";
}
