// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/interfaces/IBridgePool.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/libraries/errors/LocalERCBridgePoolBaseErrors.sol";
import "contracts/utils/AccusationsLibrary.sol";
import "contracts/libraries/errors/AccusationsErrors.sol";
import "contracts/libraries/parsers/MerkleProofParserLibrary.sol";
import "contracts/libraries/parsers/VSPreImageParserLibrary.sol";
import "contracts/libraries/parsers/PClaimsParserLibrary.sol";
import "contracts/utils/MerkleProofLibrary.sol";
import "contracts/Snapshots.sol";

/// @custom:salt LocalERCBridgePoolBase
/// @custom:deploy-type deployUpgradeable
abstract contract LocalERCBridgePoolBase is IBridgePool, ImmutableSnapshots {
    using MerkleProofParserLibrary for bytes;
    using MerkleProofLibrary for MerkleProofParserLibrary.MerkleProof;

    struct DepositParameters {
        uint256 tokenId;
        uint256 tokenAmount;
    }

    mapping(bytes32 => bool) private _consumedUTXOIDs;

    constructor() ImmutableFactory(msg.sender) {}

    /// @notice Preforms previous required tasks for depositing
    /// @param msgSender The address of ERC sender
    /// @param depositParameters encoded deposit parameters (ERC20:tokenAmount, ERC721:tokenId or ERC1155:tokenAmount+tokenId)
    function deposit(address msgSender, bytes calldata depositParameters) public virtual {}

    /// @notice Performs previous tasks required for withdrawal
    /// @param _vsPreImage burned UTXO's VSPreImage
    /// @param _proofs an array of merkle proof structs in the following order:
    /// proof against StateRoot: Proof of inclusion of UTXO in the stateTrie
    /// proof of inclusion in TXRoot: Proof of inclusion of the transaction included in the txRoot trie.
    /// proof of inclusion in TXHash: Proof of inclusion of txOut in the txHash trie
    /// proof of inclusion in HeaderRoot: Proof of inclusion of the block header in the header trie
    function withdraw(bytes memory _vsPreImage, bytes[4] memory _proofs)
        public
        virtual
        returns (
            bytes32,
            address,
            uint256
        )
    {
        BClaimsParserLibrary.BClaims memory bClaims = Snapshots(_snapshotsAddress())
            .getBlockClaimsFromLatestSnapshot();

        // Validate ProofInclusionTxRoot against BClaims.HeaderRoot.
        MerkleProofParserLibrary.MerkleProof
            memory proofInclusionHeaderRoot = MerkleProofParserLibrary.extractMerkleProof(
                _proofs[3]
            );
        MerkleProofLibrary.verifyInclusion(proofInclusionHeaderRoot, bClaims.headerRoot);

        // Validate ProofInclusionTxRoot against BClaims.TxRoot.
        MerkleProofParserLibrary.MerkleProof memory proofInclusionTxRoot = MerkleProofParserLibrary
            .extractMerkleProof(_proofs[1]);
        MerkleProofLibrary.verifyInclusion(proofInclusionTxRoot, bClaims.txRoot);

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
        MerkleProofLibrary.verifyInclusion(proofAgainstStateRoot, bClaims.stateRoot);
        if (vsPreImage.account != msg.sender) {
            revert LocalERCBridgePoolBaseErrors.UTXOAccountDifferentThanReceiver(
                vsPreImage.account,
                msg.sender
            );
        }
        if (_consumedUTXOIDs[computedUTXOID] == true) {
            revert LocalERCBridgePoolBaseErrors.UTXOAlreadyWithdrawn(computedUTXOID);
        }
        _consumedUTXOIDs[computedUTXOID] = true;
        return (computedUTXOID, vsPreImage.account, vsPreImage.value);
    }
}
