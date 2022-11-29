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
import "contracts/utils/auth/ImmutableFactory.sol";
import "contracts/utils/auth/ImmutableSnapshots.sol";
import "contracts/utils/auth/ImmutableETHDKG.sol";
import "contracts/utils/auth/ImmutableValidatorPool.sol";
import "contracts/utils/AccusationsLibrary.sol";
import "contracts/libraries/errors/AccusationsErrors.sol";

/// @custom:salt-type Accusation
/// @custom:salt InvalidTxConsumptionAccusation
/// @custom:deploy-type deployUpgradeable
contract InvalidTxConsumptionAccusation is
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
    ) public view returns (address) {
        // Require that the previous block is signed by correct group key for validator set.
        _verifySignatureGroup(_bClaims, _bClaimsSigGroup);

        // Require that height delta is 1.
        BClaimsParserLibrary.BClaims memory bClaims = BClaimsParserLibrary.extractBClaims(_bClaims);
        PClaimsParserLibrary.PClaims memory pClaims = PClaimsParserLibrary.extractPClaims(_pClaims);

        if (pClaims.bClaims.txCount == 0) {
            revert AccusationsErrors.NoTransactionInAccusedProposal();
        }

        // todo: check if the height is at least 2 epochs behind
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
        address signerAccount = AccusationsLibrary.recoverMadNetSigner(_pClaimsSig, _pClaims);

        if (!IValidatorPool(_validatorPoolAddress()).isAccusable(signerAccount)) {
            revert AccusationsErrors.SignerNotValidValidator(signerAccount);
        }

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
        if (proofAgainstStateRoot.key != proofOfInclusionTxHash.key) {
            revert AccusationsErrors.UTXODoesnotMatch(
                proofAgainstStateRoot.key,
                proofOfInclusionTxHash.key
            );
        }

        TXInPreImageParserLibrary.TXInPreImage memory txInPreImage = TXInPreImageParserLibrary
            .extractTXInPreImage(_txInPreImage);

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

        //todo burn the validator's tokens
        return signerAccount;
    }

    /// @notice This function verifies the signature group of a BClaims.
    /// @param _bClaims the BClaims of the accusation
    /// @param _bClaimsSigGroup the signature group of Pclaims
    function _verifySignatureGroup(
        bytes memory _bClaims,
        bytes memory _bClaimsSigGroup
    ) internal view {
        uint256[4] memory publicKey;
        uint256[2] memory signature;
        (publicKey, signature) = RCertParserLibrary.extractSigGroup(_bClaimsSigGroup, 0);

        // todo: check if the signature is equals to any of the previous master public key?

        if (
            !CryptoLibrary.verifySignatureASM(
                abi.encodePacked(keccak256(_bClaims)),
                signature,
                publicKey
            )
        ) {
            revert AccusationsErrors.SignatureVerificationFailed();
        }
    }
}
