// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC1155/ERC1155.sol";
import "@openzeppelin/contracts/token/ERC1155/IERC1155.sol";
import "@openzeppelin/contracts/token/ERC1155/utils/ERC1155Holder.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/libraries/errors/BridgePoolErrors.sol";
import "contracts/libraries/parsers/MerkleProofParserLibrary.sol";
import "contracts/utils/MerkleProofLibrary.sol";
import "contracts/interfaces/IBridgePool.sol";
import "contracts/Snapshots.sol";
import "contracts/libraries/parsers/BClaimsParserLibrary.sol";
import "contracts/utils/ERC20SafeTransfer.sol";
import "hardhat/console.sol";

/// @custom:salt LocalERC1155BridgePoolV1
/// @custom:deploy-type deployStatic

contract LocalERC1155BridgePoolV1 is
    ERC1155Holder,
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
        uint256 number;
        uint256 fee;
        bytes32 txHash;
    }

    bytes private constant _ERC1155_EMPTY_DATA = abi.encodePacked("");
    address internal _erc1155Contract;

    constructor() ImmutableFactory(msg.sender) {}

    function initialize(address erc1155Contract_) public onlyFactory initializer {
        _erc1155Contract = erc1155Contract_;
    }

    /// @notice Transfer tokens from sender and emit a "Deposited" event for minting correspondent tokens in sidechain
    /// @param msgSender The address of ERC sender
    /// @param number The amount or token Id
    function deposit(
        address msgSender,
        uint256 number,
        uint256 tokenId
    ) public onlyBridgeRouter {
        console.log(msgSender, number, tokenId);
        IERC1155(_erc1155Contract).safeTransferFrom(
            msgSender,
            address(this),
            tokenId,
            number,
            _ERC1155_EMPTY_DATA
        );
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
        IERC1155(_erc1155Contract).safeTransferFrom(
            address(this),
            msg.sender,
            1,
            burnedUTXO.number,
            abi.encodePacked("")
        );
    }
}
