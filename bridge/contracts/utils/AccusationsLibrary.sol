// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library AccusationsLibrary {
    /// @notice Recovers the signer of a message
    /// @param signature The ECDSA signature
    /// @param prefix The prefix of the message
    /// @param message The message
    /// @return the address of the signer
    function recoverSigner(
        bytes memory signature,
        bytes memory prefix,
        bytes memory message
    ) internal pure returns (address) {
        require(signature.length == 65, "Accusations: Signature should be 65 bytes");

        bytes32 hashedMessage = keccak256(abi.encodePacked(prefix, message));

        bytes32 r;
        bytes32 s;
        uint8 v;

        assembly {
            r := mload(add(signature, 32))
            s := mload(add(signature, 64))
            v := byte(0, mload(add(signature, 96)))
        }

        v = (v < 27) ? (v + 27) : v;

        require(v == 27 || v == 28, "Accusations: Signature uses invalid version");

        return ecrecover(hashedMessage, v, r, s);
    }

    /// @notice Recovers the signer of a message in MadNet
    /// @param signature The ECDSA signature
    /// @param message The message
    /// @return the address of the signer
    function recoverMadNetSigner(bytes memory signature, bytes memory message)
        internal
        pure
        returns (address)
    {
        return recoverSigner(signature, "Proposal", message);
    }

    /// @notice Computes the UTXOID
    /// @param txHash the transaction hash
    /// @param txIdx the transaction index
    /// @return the UTXOID
    function computeUTXOID(bytes32 txHash, uint32 txIdx) internal pure returns (bytes32) {
        return keccak256(abi.encodePacked(txHash, txIdx));
    }
}
