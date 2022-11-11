// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

library NativeERCBridgePoolBaseErrors {
    error OnlyBridgePool();
    error ChainIdDoesNotMatch(uint256 bClaimsChainId, uint256 snapshotsChainId);
    error UTXODoesnotMatch(bytes32 proofOfInclusionStateRootKey, bytes32 proofOfInclusionTxHashKey);
    error UTXOAlreadyWithdrawn(bytes32 computedUTXOID);
    error UTXOAccountDoesNotMatchReceiver(address utxoAccount, address msgSender);
    error MerkleProofKeyDoesNotMatchUTXOID(
        bytes32 proofOfInclusionTxHashKey,
        bytes32 proofOfInclusionStateRootKey
    );
}
