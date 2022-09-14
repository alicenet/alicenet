// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/interfaces/IBridgePool.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/libraries/errors/LocalERCBridgePoolBaseErrors.sol";
import "contracts/libraries/parsers/MerkleProofParserLibrary.sol";
import "contracts/utils/MerkleProofLibrary.sol";
import "contracts/Snapshots.sol";
import "contracts/libraries/parsers/BClaimsParserLibrary.sol";

/// @custom:salt LocalERCBridgePoolBase
/// @custom:deploy-type deployStatic
abstract contract LocalERCBridgePoolBase is IBridgePool, ImmutableSnapshots {
    using MerkleProofParserLibrary for bytes;
    using MerkleProofLibrary for MerkleProofParserLibrary.MerkleProof;

    struct UTXO {
        uint32 chainID;
        address owner;
        uint256 tokenId;
        uint256 tokenAmount;
        uint256 fee;
        bytes32 txHash;
    }

    struct DepositParameters {
        uint256 tokenId;
        uint256 tokenAmount;
    }

    constructor() ImmutableFactory(msg.sender) {}

    /// @notice Transfer ERC20 or ERC721 tokens from sender for minting them on sidechain
    /// @param msgSender The address of ERC sender
    /// @param depositParameters encoded deposit parameters (ERC20:tokenAmount, ERC721:tokenId or ERC1155:tokenAmount+tokenId)
    function deposit(address msgSender, bytes calldata depositParameters) public virtual {}

    /// @notice Transfer tokens to sender upon a verificable proof of burn in sidechain
    /// @param encodedMerkleProof The merkle proof
    /// @param encodedBurnedUTXO encoded burned UTXO
    function withdraw(bytes memory encodedMerkleProof, bytes memory encodedBurnedUTXO)
        public
        virtual
    {
        BClaimsParserLibrary.BClaims memory bClaims = Snapshots(_snapshotsAddress())
            .getBlockClaimsFromLatestSnapshot();
        MerkleProofParserLibrary.MerkleProof memory merkleProof = encodedMerkleProof
            .extractMerkleProof();
        merkleProof.verifyInclusion(bClaims.stateRoot);
        UTXO memory burnedUTXO = abi.decode(encodedBurnedUTXO, (UTXO));
        if (burnedUTXO.owner != msg.sender) {
            revert LocalERCBridgePoolBaseErrors.ReceiverIsNotOwnerOnProofOfBurnUTXO();
        }
    }
}
