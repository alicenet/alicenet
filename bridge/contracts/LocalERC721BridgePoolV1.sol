// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC721/utils/ERC721Holder.sol";
import "contracts/interfaces/IERC721Transferable.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/libraries/errors/BridgePoolErrors.sol";
import "contracts/libraries/parsers/MerkleProofParserLibrary.sol";
import "contracts/utils/MerkleProofLibrary.sol";
import "contracts/interfaces/IBridgePool.sol";
import "contracts/Snapshots.sol";
import "contracts/libraries/parsers/BClaimsParserLibrary.sol";
import "contracts/utils/ERC20SafeTransfer.sol";

/// @custom:salt LocalERC721BridgePoolV1
/// @custom:deploy-type deployStatic

contract LocalERC721BridgePoolV1 is
    ERC721Holder,
    IBridgePool,
    Initializable,
    ImmutableSnapshots,
    ImmutableBridgePoolFactory,
    ImmutableBridgeRouter
{
    using MerkleProofParserLibrary for bytes;
    using MerkleProofLibrary for MerkleProofParserLibrary.MerkleProof;

    struct UTXO {
        uint32 chainID;
        address owner;
        uint256 value;
        uint256 fee;
        bytes32 txHash;
    }

    address internal _erc721Contract;

    constructor() ImmutableFactory(msg.sender) {}

    function initialize(address erc721Contract_) public onlyBridgePoolFactory initializer {
        _erc721Contract = erc721Contract_;
    }

    /// @notice Transfer tokens from sender and emit a "Deposited" event for minting correspondent tokens in sidechain
    /// @param msgSender The address of ERC sender
    /// @param number The token ID of the NFT to be deposited
    function deposit(address msgSender, uint256 number) public onlyBridgeRouter {
        IERC721Transferable(_erc721Contract).safeTransferFrom(msgSender, address(this), number);
    }

    /// @notice Transfer token to sender upon a verificable proof of burn in sidechain
    /// @param encodedMerkleProof The merkle proof
    /// @param encodedBurnedUTXO The burned UTXO
    function withdraw(bytes memory encodedMerkleProof, bytes memory encodedBurnedUTXO) public {
        BClaimsParserLibrary.BClaims memory bClaims = Snapshots(_snapshotsAddress())
            .getBlockClaimsFromLatestSnapshot();
        MerkleProofParserLibrary.MerkleProof memory merkleProof = encodedMerkleProof
            .extractMerkleProof();
        UTXO memory burnedUTXO = abi.decode(encodedBurnedUTXO, (UTXO));
        if (burnedUTXO.owner != msg.sender) {
            revert BridgePoolErrors.ReceiverIsNotOwnerOnProofOfBurnUTXO();
        }
        merkleProof.verifyInclusion(bClaims.stateRoot);
        IERC721Transferable(_erc721Contract).safeTransferFrom(
            address(this),
            msg.sender,
            burnedUTXO.value // tokenId
        );
    }
}
