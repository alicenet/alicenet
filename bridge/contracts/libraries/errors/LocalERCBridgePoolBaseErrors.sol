// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

library LocalERCBridgePoolBaseErrors {
    error OnlyBridgePool();
    error ReceiverIsNotOwnerOnProofOfBurnUTXO();
    error ChainIdDoesNotMatch(uint256 bClaimsChainId, uint256 snapshotsChainId);
    error UTXODoesnotMatch(bytes32 proofAgainstStateRootKey, bytes32 proofOfInclusionTxHashKey);
    error SignatureVerificationFailed();
    error UTXOAlreadyWithdrawn(bytes32 computedUTXOID);
    error UTXOAccountDifferentThanReceiver(address utxoAccount, address msgSender);
}
