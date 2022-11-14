// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/interfaces/IBridgePool.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/libraries/errors/NativeERCBridgePoolBaseErrors.sol";
import "contracts/utils/AccusationsLibrary.sol";
import "contracts/libraries/parsers/MerkleProofParserLibrary.sol";
import "contracts/libraries/parsers/VSPreImageParserLibrary.sol";
import "contracts/libraries/parsers/PClaimsParserLibrary.sol";
import "contracts/utils/MerkleProofLibrary.sol";
import "contracts/Snapshots.sol";

abstract contract NativeERCBridgePoolBase is ImmutableFactory, ImmutableBridgeRouter, IBridgePool {
    using MerkleProofParserLibrary for bytes;
    using MerkleProofLibrary for MerkleProofParserLibrary.MerkleProof;

    struct DepositParameters {
        uint256 tokenId;
        uint256 tokenAmount;
    }

    address private immutable _snapshotsAddress;

    mapping(bytes32 => bool) private _consumedUTXOIDs;

    constructor(address alicenetFactoryAddress, address snapshotsAddress_)
        ImmutableFactory(alicenetFactoryAddress)
    {
        _snapshotsAddress = snapshotsAddress_;
    }

    /// @notice Transfer tokens from sender
    /// @param msgSender The address of ERC sender
    /// @param depositParameters_ encoded deposit parameters (ERC20:tokenAmount, ERC721:tokenId or ERC1155:tokenAmount+tokenId)
    function deposit(address msgSender, bytes calldata depositParameters_)
        public
        virtual
        override
        onlyBridgeRouter
    {}

    /// @notice Transfer tokens to sender upon proofs verification
    /// @param msgReceiver The address of ERC receiver
    /// @param vsPreImage burned UTXO in chain
    /// @param proofs Proofs of inclusion of burned UTXO
    function withdraw(
        address msgReceiver,
        bytes memory vsPreImage,
        bytes[4] memory proofs
    ) public virtual override onlyBridgeRouter returns (address account, uint256 value) {
        MerkleProofParserLibrary.MerkleProof memory proofInclusionStateRoot = _verifyProofs(proofs);
        (account, value) = _getValidatedTransferData(
            msgReceiver,
            vsPreImage,
            proofInclusionStateRoot
        );
    }

    /// @notice Obtains trasfer data upon UTXO verification
    /// @param msgReceiver The address of ERC receiver
    /// @param _vsPreImage burned UTXO
    /// @param proofInclusionStateRoot Proof of inclusion of UTXO in the stateTrie
    function _getValidatedTransferData(
        address msgReceiver,
        bytes memory _vsPreImage,
        MerkleProofParserLibrary.MerkleProof memory proofInclusionStateRoot
    ) internal returns (address, uint256) {
        VSPreImageParserLibrary.VSPreImage memory vsPreImage = VSPreImageParserLibrary
            .extractVSPreImage(_vsPreImage);
        if (vsPreImage.chainId != ISnapshots(_snapshotsAddress).getChainId()) {
            revert NativeERCBridgePoolBaseErrors.ChainIdDoesNotMatch(
                vsPreImage.chainId,
                ISnapshots(_snapshotsAddress).getChainId()
            );
        }
        bytes32 computedUTXOID = AccusationsLibrary.computeUTXOID(
            vsPreImage.txHash,
            vsPreImage.txOutIdx
        );
        if (computedUTXOID != proofInclusionStateRoot.key) {
            revert NativeERCBridgePoolBaseErrors.MerkleProofKeyDoesNotMatchUTXOID(
                computedUTXOID,
                proofInclusionStateRoot.key
            );
        }
        if (vsPreImage.account != msgReceiver) {
            revert NativeERCBridgePoolBaseErrors.UTXOAccountDoesNotMatchReceiver(
                vsPreImage.account,
                msgReceiver
            );
        }
        if (_consumedUTXOIDs[computedUTXOID] == true) {
            revert NativeERCBridgePoolBaseErrors.UTXOAlreadyWithdrawn(computedUTXOID);
        }
        _consumedUTXOIDs[computedUTXOID] = true;
        return (vsPreImage.account, vsPreImage.value);
    }

    /// @notice Verify withdraw proofs
    /// @param _proofs an array of merkle proof structs in the following order:
    /// proof of inclusion in StateRoot: Proof of inclusion of UTXO in the stateTrie
    /// proof of inclusion in TXRoot: Proof of inclusion of the transaction included in the txRoot trie.
    /// proof of inclusion in TXHash: Proof of inclusion of txOut in the txHash trie
    /// proof of inclusion in HeaderRoot: Proof of inclusion of the block header in the header trie
    function _verifyProofs(bytes[4] memory _proofs)
        internal
        view
        returns (MerkleProofParserLibrary.MerkleProof memory)
    {
        BClaimsParserLibrary.BClaims memory bClaims = Snapshots(_snapshotsAddress)
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
        MerkleProofParserLibrary.MerkleProof
            memory proofInclusionStateRoot = MerkleProofParserLibrary.extractMerkleProof(
                _proofs[0]
            );
        if (proofInclusionStateRoot.key != proofOfInclusionTxHash.key) {
            revert NativeERCBridgePoolBaseErrors.UTXODoesnotMatch(
                proofInclusionStateRoot.key,
                proofOfInclusionTxHash.key
            );
        }
        MerkleProofLibrary.verifyInclusion(proofInclusionStateRoot, bClaims.stateRoot);
        return proofInclusionStateRoot;
    }
}
