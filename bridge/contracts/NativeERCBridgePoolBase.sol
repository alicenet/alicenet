// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/interfaces/IBridgePool.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/libraries/errors/NativeERCBridgePoolBaseErrors.sol";
import "contracts/utils/AccusationsLibrary.sol";
import "contracts/libraries/errors/AccusationsErrors.sol";
import "contracts/libraries/parsers/MerkleProofParserLibrary.sol";
import "contracts/libraries/parsers/VSPreImageParserLibrary.sol";
import "contracts/libraries/parsers/PClaimsParserLibrary.sol";
import "contracts/utils/MerkleProofLibrary.sol";
import "contracts/Snapshots.sol";

/// @custom:salt NativeERCBridgePoolBase
/// @custom:deploy-type deployUpgradeable
abstract contract NativeERCBridgePoolBase is ImmutableSnapshots {
    using MerkleProofParserLibrary for bytes;
    using MerkleProofLibrary for MerkleProofParserLibrary.MerkleProof;

    struct DepositParameters {
        uint256 tokenId;
        uint256 tokenAmount;
    }

    mapping(bytes32 => bool) private _consumedUTXOIDs;

    constructor() ImmutableFactory(msg.sender) {}

    /// @notice Verify withdraw proofs
    /// @param _proofs an array of merkle proof structs in the following order:
    /// proof of inclusion in StateRoot: Proof of inclusion of UTXO in the stateTrie
    /// proof of inclusion in TXRoot: Proof of inclusion of the transaction included in the txRoot trie.
    /// proof of inclusion in TXHash: Proof of inclusion of txOut in the txHash trie
    /// proof of inclusion in HeaderRoot: Proof of inclusion of the block header in the header trie
    function verifyProofs(bytes[4] memory _proofs)
        public
        returns (MerkleProofParserLibrary.MerkleProof memory)
    {
        BClaimsParserLibrary.BClaims memory bClaims = Snapshots(_snapshotsAddress())
            .getBlockClaimsFromLatestSnapshot();

        // Validate proofInclusionHeaderRoot against bClaims.headerRoot.
        MerkleProofParserLibrary.MerkleProof
            memory proofInclusionHeaderRoot = MerkleProofParserLibrary.extractMerkleProof(
                _proofs[3]
            );
        MerkleProofLibrary.verifyInclusion(proofInclusionHeaderRoot, bClaims.headerRoot);

        // Validate proofInclusionTxRoot against bClaims.txRoot.
        MerkleProofParserLibrary.MerkleProof memory proofInclusionTxRoot = MerkleProofParserLibrary
            .extractMerkleProof(_proofs[1]);
        MerkleProofLibrary.verifyInclusion(proofInclusionTxRoot, bClaims.txRoot);

        // Validate proofOfInclusionTxHash against the target hash from proofInclusionTxRoot.
        MerkleProofParserLibrary.MerkleProof
            memory proofOfInclusionTxHash = MerkleProofParserLibrary.extractMerkleProof(_proofs[2]);
        MerkleProofLibrary.verifyInclusion(proofOfInclusionTxHash, proofInclusionTxRoot.key);

        // Validate proofOfInclusionTxHash against bClaims.stateRoot.
        MerkleProofParserLibrary.MerkleProof memory proofAgainstStateRoot = MerkleProofParserLibrary
            .extractMerkleProof(_proofs[0]);
        MerkleProofLibrary.verifyInclusion(proofAgainstStateRoot, bClaims.stateRoot);

        if (proofAgainstStateRoot.key != proofOfInclusionTxHash.key) {
            revert AccusationsErrors.UTXODoesnotMatch(
                proofAgainstStateRoot.key,
                proofOfInclusionTxHash.key
            );
        }
        return proofAgainstStateRoot;
    }

    /// @notice Obtains trasfer data upon UTXO verification
    /// @param _vsPreImage burned UTXO
    /// @param proofAgainstStateRoot Proof of inclusion of UTXO in the stateTrie
    function getValidatedTransferData(
        bytes memory _vsPreImage,
        MerkleProofParserLibrary.MerkleProof memory proofAgainstStateRoot
    )
        public
        returns (
            bytes32,
            address,
            uint256
        )
    {
        VSPreImageParserLibrary.VSPreImage memory vsPreImage = VSPreImageParserLibrary
            .extractVSPreImage(_vsPreImage);

        if (vsPreImage.chainId != ISnapshots(_snapshotsAddress()).getChainId()) {
            revert AccusationsErrors.ChainIdDoesNotMatch(
                vsPreImage.chainId,
                vsPreImage.chainId,
                ISnapshots(_snapshotsAddress()).getChainId()
            );
        }

        bytes32 computedUTXOID = AccusationsLibrary.computeUTXOID(
            vsPreImage.txHash,
            vsPreImage.txOutIdx
        );

        if (computedUTXOID != proofAgainstStateRoot.key) {
            revert AccusationsErrors.MerkleProofKeyDoesNotMatchUTXOIDBeingSpent(
                computedUTXOID,
                proofAgainstStateRoot.key
            );
        }
        if (vsPreImage.account != msg.sender) {
            revert NativeERCBridgePoolBaseErrors.UTXOAccountDoesNotMatchReceiver(
                vsPreImage.account,
                msg.sender
            );
        }
        if (_consumedUTXOIDs[computedUTXOID] == true) {
            revert NativeERCBridgePoolBaseErrors.UTXOAlreadyWithdrawn(computedUTXOID);
        }
        _consumedUTXOIDs[computedUTXOID] = true;
        return (computedUTXOID, vsPreImage.account, vsPreImage.value);
    }
}
