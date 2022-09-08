// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

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
import "contracts/libraries/errors/AccusationsErrors.sol";

/// @custom:salt InvalidTxConsumptionAccusation
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
    /// @param pClaims_ the PClaims of the accusation
    /// @param pClaimsSig_ the signature of PClaims
    /// @param bClaims_ the BClaims of the accusation
    /// @param bClaimsSigGroup_ the signature group of BClaims
    /// @param txInPreImage_ the TXInPreImage consuming the invalid transaction
    /// @param proofs_ an array of merkle proof structs in the following order:
    /// proof against StateRoot: Proof of inclusion or exclusion of the deposit or UTXO in the stateTrie
    /// proof of inclusion in TXRoot: Proof of inclusion of the transaction that included the invalid input in the txRoot trie.
    /// proof of inclusion in TXHash: Proof of inclusion of the invalid input (txIn) in the txHash trie (transaction tested against the TxRoot).
    /// @return the address of the signer
    function accuseInvalidTransactionConsumption(
        bytes memory pClaims_,
        bytes memory pClaimsSig_,
        bytes memory bClaims_,
        bytes memory bClaimsSigGroup_,
        bytes memory txInPreImage_,
        bytes[3] memory proofs_
    ) public returns (address) {
        // Require that the previous block is signed by correct group key for validator set.
        _verifySignatureGroup(bClaims_, bClaimsSigGroup_);

        // Require that height delta is 1.
        BClaimsParserLibrary.BClaims memory bClaims = BClaimsParserLibrary.extractBClaims(bClaims_);
        PClaimsParserLibrary.PClaims memory pClaims = PClaimsParserLibrary.extractPClaims(pClaims_);

        if (pClaims.bClaims.txCount == 0) {
            revert AccusationsErrors.NoTransactionInAccusedProposal();
        }
        if (bClaims.height + 1 != pClaims.bClaims.height) {
            revert AccusationsErrors.HeightDeltaShouldBeOne(bClaims.height, pClaims.bClaims.height);
        }

        // Require that chainID is equal.
        if (
            bClaims.chainId != pClaims.bClaims.chainId ||
            bClaims.chainId != ISnapshots(_snapshotsAddress()).getChainId()
        ) {
            revert AccusationsErrors.ChainIdDoesNotMatch(
                bClaims.chainId,
                pClaims.bClaims.chainId,
                ISnapshots(_snapshotsAddress()).getChainId()
            );
        }

        // Require that Proposal was signed by active validator.
        address signerAccount = AccusationsLibrary.recoverMadNetSigner(pClaimsSig_, pClaims_);

        if (!IValidatorPool(_validatorPoolAddress()).isAccusable(signerAccount)) {
            revert AccusationsErrors.SignerNotValidValidator(signerAccount);
        }

        // Validate ProofInclusionTxRoot against PClaims.BClaims.TxRoot.
        MerkleProofParserLibrary.MerkleProof memory proofInclusionTxRoot = MerkleProofParserLibrary
            .extractMerkleProof(proofs_[1]);
        MerkleProofLibrary.verifyInclusion(proofInclusionTxRoot, pClaims.bClaims.txRoot);

        // Validate ProofOfInclusionTxHash against the target hash from ProofInclusionTxRoot.
        MerkleProofParserLibrary.MerkleProof
            memory proofOfInclusionTxHash = MerkleProofParserLibrary.extractMerkleProof(proofs_[2]);
        MerkleProofLibrary.verifyInclusion(proofOfInclusionTxHash, proofInclusionTxRoot.key);

        MerkleProofParserLibrary.MerkleProof memory proofAgainstStateRoot = MerkleProofParserLibrary
            .extractMerkleProof(proofs_[0]);
        if (proofAgainstStateRoot.key != proofOfInclusionTxHash.key) {
            revert AccusationsErrors.UTXODoesnotMatch(
                proofAgainstStateRoot.key,
                proofOfInclusionTxHash.key
            );
        }

        TXInPreImageParserLibrary.TXInPreImage memory txInPreImage = TXInPreImageParserLibrary
            .extractTXInPreImage(txInPreImage_);

        // checking if we are consuming a deposit or an UTXO
        if (txInPreImage.consumedTxIdx == 0xFFFFFFFF) {
            // Double spending problem, i.e, consuming a deposit that was already consumed
            if (txInPreImage.consumedTxHash != proofAgainstStateRoot.key) {
                revert AccusationsErrors.MerkleProofKeyDoesNotMatchConsumedDepositKey(
                    txInPreImage.consumedTxHash,
                    proofAgainstStateRoot.key
                );
            }
            MerkleProofLibrary.verifyInclusion(proofAgainstStateRoot, bClaims.stateRoot);
            // todo: deposit that doesn't exist in the chain. Maybe split this in separate functions?
        } else {
            // Consuming a non existing UTXO
            {
                bytes32 computedUTXOID = AccusationsLibrary.computeUTXOID(
                    txInPreImage.consumedTxHash,
                    txInPreImage.consumedTxIdx
                );
                if (computedUTXOID != proofAgainstStateRoot.key) {
                    revert AccusationsErrors.MerkleProofKeyDoesNotMatchUTXOIDBeingSpent(
                        computedUTXOID,
                        proofAgainstStateRoot.key
                    );
                }
            }
            MerkleProofLibrary.verifyNonInclusion(proofAgainstStateRoot, bClaims.stateRoot);
        }

        // burn the validator's tokens
        IValidatorPool(_validatorPoolAddress()).majorSlash(signerAccount, msg.sender, PRE_SALT);

        return signerAccount;
    }

    /// @notice This function verifies the signature group of a BClaims.
    /// @param bClaims_ the BClaims of the accusation
    /// @param bClaimsSigGroup_ the signature group of Pclaims
    function _verifySignatureGroup(bytes memory bClaims_, bytes memory bClaimsSigGroup_)
        internal
        view
    {
        uint256[4] memory publicKey;
        uint256[2] memory signature;
        (publicKey, signature) = RCertParserLibrary.extractSigGroup(bClaimsSigGroup_, 0);

        // todo: check if the signature is equals to any of the previous master public key?

        if (
            !CryptoLibrary.verifySignatureASM(
                abi.encodePacked(keccak256(bClaims_)),
                signature,
                publicKey
            )
        ) {
            revert AccusationsErrors.SignatureVerificationFailed();
        }
    }
}
