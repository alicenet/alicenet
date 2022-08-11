// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/utils/MerkleProofLibrary.sol";
import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/ISnapshots.sol";
import "contracts/libraries/parsers/PClaimsParserLibrary.sol";
import "contracts/libraries/parsers/RCertParserLibrary.sol";
import "contracts/libraries/parsers/MerkleProofParserLibrary.sol";
import "contracts/libraries/parsers/TXInPreImageParserLibrary.sol";
import "contracts/libraries/math/CryptoLibrary.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/utils/AccusationsLibrary.sol";


/// @custom:deploy-type deployUpgradeable
/// @custom:role Accusation
contract AccusationInvalidTxConsumption is
    ImmutableFactory,
    ImmutableSnapshots,
    ImmutableETHDKG,
    ImmutableValidatorPool
{
    // this is the keccak256 of "AccusationInvalidTxConsumption"
    bytes32 constant public PRE_SALT = 0xf40095839ea6635a5869735bd0c363085cb0ebd561e0f361f826103b958c27e5;
    mapping(bytes32 => bool) internal _accusations;

    constructor()
        ImmutableFactory(msg.sender)
        ImmutableSnapshots()
        ImmutableETHDKG()
        ImmutableValidatorPool()
    {}

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
    function accuseInvalidTransactionConsumption(
        bytes memory _pClaims,
        bytes memory _pClaimsSig,
        bytes memory _bClaims,
        bytes memory _bClaimsSigGroup,
        bytes memory _txInPreImage,
        bytes[3] memory _proofs
    ) public returns (address) {
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
        address signerAccount = AccusationsLibrary.recoverMadNetSigner(_pClaimsSig, _pClaims);

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
                AccusationsLibrary.computeUTXOID(
                    txInPreImage.consumedTxHash,
                    txInPreImage.consumedTxIdx
                ) == proofAgainstStateRoot.key,
                "Accusations: The key of Merkle Proof should be equal to the UTXOID being spent!"
            );
            MerkleProofLibrary.verifyNonInclusion(proofAgainstStateRoot, bClaims.stateRoot);
        }

        IValidatorPool(_validatorPoolAddress()).majorSlash(signerAccount, msg.sender, PRE_SALT);

        return signerAccount;
    }

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
}
