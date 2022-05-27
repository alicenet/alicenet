// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/utils/MerkleProofLibrary.sol";
import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/ISnapshots.sol";
import "contracts/interfaces/IETHDKG.sol";
import "contracts/libraries/parsers/PClaimsParserLibrary.sol";
import "contracts/libraries/parsers/RCertParserLibrary.sol";
import "contracts/libraries/parsers/MerkleProofParserLibrary.sol";
import "contracts/libraries/parsers/TXInPreImageParserLibrary.sol";
import "contracts/libraries/math/CryptoLibrary.sol";
import "contracts/utils/ImmutableAuth.sol";

contract Accusations is
    ImmutableFactory,
    ImmutableSnapshots,
    ImmutableETHDKG,
    ImmutableValidatorPool
{
    mapping(bytes32 => bool) internal _accusations;

    constructor()
        ImmutableFactory(msg.sender)
        ImmutableSnapshots()
        ImmutableETHDKG()
        ImmutableValidatorPool()
    {}

    /// @notice This function verifies the signature group of a BClaims.
    /// @param _bClaims the BClaims of the accusation
    /// @param _bClaimsSigGroup the signature group of Pclaims
    function _verifySignatureGroup(bytes memory _bClaims, bytes memory _bClaimsSigGroup)
        internal
        view
    {
        uint256[4] memory publicKey;
        uint256[2] memory signature;
        (publicKey, signature) = RCertParserLibrary.extractSigGroup(_bClaimsSigGroup, 0);

        // todo: check if the signature is equals to any of the previous master public key?

        require(
            CryptoLibrary.verifySignatureASM(
                abi.encodePacked(keccak256(_bClaims)),
                signature,
                publicKey
            ),
            "Accusations: Signature verification failed"
        );
    }

    /// @notice This function validates an accusation of non-existent utxo consumption, as well as invalid deposit consumption.
    /// @param _pClaims the PClaims of the accusation
    /// @param _pClaimsSig the signature of PClaims
    /// @param _bClaims the BClaims of the accusation
    /// @param _bClaimsSigGroup the signature group of BClaims
    /// @param _txInPreImage the TXInPreImage consuming the invalid transaction
    /// @param _proofs an array of merkle proof structs in the following order:
    /// proof against StateRoot: Proof of inclusion or exclusion of the deposit or UTXO in the stateTrie
    /// proof of inclusion in TXRoot: Proof of inclusion of the transaction that included the invalid input in the txRoot trie.
    /// proof of inclusion in TXHash: Proof of inclusion of the invalid input (txIn) in the txHash trie (transaction tested against the TxRoot).
    /// @return the address of the signer
    function AccuseInvalidTransactionConsumption(
        bytes memory _pClaims,
        bytes memory _pClaimsSig,
        bytes memory _bClaims,
        bytes memory _bClaimsSigGroup,
        bytes memory _txInPreImage,
        bytes[3] memory _proofs
    ) public view returns (address) {
        // Require that the previous block is signed by correct group key for validator set.
        _verifySignatureGroup(_bClaims, _bClaimsSigGroup);

        // Require that height delta is 1.
        BClaimsParserLibrary.BClaims memory bClaims = BClaimsParserLibrary.extractBClaims(_bClaims);
        PClaimsParserLibrary.PClaims memory pClaims = PClaimsParserLibrary.extractPClaims(_pClaims);

        require(
            pClaims.bClaims.txCount > 0,
            "Accusations: The accused proposal doesn't have any transaction!"
        );
        require(
            bClaims.height + 1 == pClaims.bClaims.height,
            "Accusations: Height delta should be 1"
        );

        // Require that chainID is equal.
        require(
            bClaims.chainId == pClaims.bClaims.chainId &&
                bClaims.chainId == ISnapshots(_snapshotsAddress()).getChainId(),
            "Accusations: ChainId should be the same"
        );

        // Require that Proposal was signed by active validator.
        address signerAccount = recoverMadNetSigner(_pClaimsSig, _pClaims);

        require(
            IValidatorPool(_validatorPoolAddress()).isAccusable(signerAccount),
            "Accusations: the signer of these proposal is not a valid validator!"
        );

        // Validate ProofInclusionTxRoot against PClaims.BClaims.TxRoot.
        MerkleProofParserLibrary.MerkleProof memory proofInclusionTxRoot = MerkleProofParserLibrary
            .extractMerkleProof(_proofs[1]);
        MerkleProofLibrary.verifyInclusion(proofInclusionTxRoot, pClaims.bClaims.txRoot);

        // Validate ProofOfInclusionTxHash against the target hash from ProofInclusionTxRoot.
        MerkleProofParserLibrary.MerkleProof
            memory proofOfInclusionTxHash = MerkleProofParserLibrary.extractMerkleProof(_proofs[2]);
        MerkleProofLibrary.verifyInclusion(proofOfInclusionTxHash, proofInclusionTxRoot.key);

        MerkleProofParserLibrary.MerkleProof memory proofAgainstStateRoot = MerkleProofParserLibrary
            .extractMerkleProof(_proofs[0]);
        require(
            proofAgainstStateRoot.key == proofOfInclusionTxHash.key,
            "Accusations: The UTXO should match!"
        );

        TXInPreImageParserLibrary.TXInPreImage memory txInPreImage = TXInPreImageParserLibrary
            .extractTXInPreImage(_txInPreImage);

        // checking if we are consuming a deposit or an UTXO
        if (txInPreImage.consumedTxIdx == 0xFFFFFFFF) {
            // Double spending problem, i.e, consuming a deposit that was already consumed
            require(
                txInPreImage.consumedTxHash == proofAgainstStateRoot.key,
                "Accusations: The key of Merkle Proof should be equal to the consumed deposit key!"
            );
            MerkleProofLibrary.verifyInclusion(proofAgainstStateRoot, bClaims.stateRoot);
            // todo: deposit that doesn't exist in the chain. Maybe split this in separate functions?
        } else {
            // Consuming a non existing UTXO
            require(
                computeUTXOID(txInPreImage.consumedTxHash, txInPreImage.consumedTxIdx) ==
                    proofAgainstStateRoot.key,
                "Accusations: The key of Merkle Proof should be equal to the UTXOID being spent!"
            );
            MerkleProofLibrary.verifyNonInclusion(proofAgainstStateRoot, bClaims.stateRoot);
        }

        //todo burn the validator's tokens
        return signerAccount;
    }

    /// @notice This function validates an accusation of multiple proposals.
    /// @param _signature0 The signature of pclaims0
    /// @param _pClaims0 The PClaims of the accusation
    /// @param _signature1 The signature of pclaims1
    /// @param _pClaims1 The PClaims of the accusation
    /// @return the address of the signer
    function AccuseMultipleProposal(
        bytes calldata _signature0,
        bytes calldata _pClaims0,
        bytes calldata _signature1,
        bytes calldata _pClaims1
    ) internal view returns (address) {
        // ecrecover sig0/1 and ensure both are valid and accounts are equal
        address signerAccount0 = recoverMadNetSigner(_signature0, _pClaims0);
        address signerAccount1 = recoverMadNetSigner(_signature1, _pClaims1);

        require(
            signerAccount0 == signerAccount1,
            "Accusations: the signers of the proposals should be the same"
        );

        // ensure the hashes of blob0/1 are different
        require(
            keccak256(_pClaims0) != keccak256(_pClaims1),
            "Accusations: the PClaims are equal!"
        );

        PClaimsParserLibrary.PClaims memory pClaims0 = PClaimsParserLibrary.extractPClaims(
            _pClaims0
        );
        PClaimsParserLibrary.PClaims memory pClaims1 = PClaimsParserLibrary.extractPClaims(
            _pClaims1
        );

        // ensure the height of blob0/1 are equal using RCert sub object of PClaims
        require(
            pClaims0.rCert.rClaims.height == pClaims1.rCert.rClaims.height,
            "Accusations: the block heights between the proposals are different!"
        );

        // ensure the round of blob0/1 are equal using RCert sub object of PClaims
        require(
            pClaims0.rCert.rClaims.round == pClaims1.rCert.rClaims.round,
            "Accusations: the round between the proposals are different!"
        );

        // ensure the chainid of blob0/1 are equal using RCert sub object of PClaims
        require(
            pClaims0.rCert.rClaims.chainId == pClaims1.rCert.rClaims.chainId,
            "Accusations: the chainId between the proposals are different!"
        );

        // ensure the chainid of blob0 is correct for this chain using RCert sub object of PClaims
        uint256 chainId = ISnapshots(_snapshotsAddress()).getChainId();
        require(
            pClaims0.rCert.rClaims.chainId == chainId,
            "Accusations: the chainId is invalid for this chain!"
        );

        // ensure both accounts are applicable to a currently locked validator - Note<may be done in different layer?>
        require(
            IValidatorPool(_validatorPoolAddress()).isAccusable(signerAccount0),
            "Accusations: the signer of these proposals is not a valid validator!"
        );

        return signerAccount0;
    }

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
