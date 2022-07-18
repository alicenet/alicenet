// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC721/utils/ERC721Holder.sol";
import "contracts/interfaces/IERC721Transferable.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/BToken.sol";
import {BridgePoolErrorCodes} from "contracts/libraries/errorCodes/BridgePoolErrorCodes.sol";
import "contracts/libraries/parsers/MerkleProofParserLibrary.sol";
import "contracts/libraries/MerkleProofLibrary.sol";
import "contracts/interfaces/IBridgePool.sol";
import "contracts/Snapshots.sol";
import "contracts/libraries/parsers/BClaimsParserLibrary.sol";
import "contracts/utils/ERC20SafeTransfer.sol";
import "contracts/BridgePoolDepositNotifier.sol";
import "contracts/BridgeRouter.sol";

/// @custom:salt LocalERC721BridgePoolV1
/// @custom:deploy-type deployStatic

contract LocalERC721BridgePoolV1 is
    ERC721Holder,
    IBridgePool,
    Initializable,
    ImmutableSnapshots,
    ImmutableBToken,
    ImmutableBridgePoolDepositNotifier,
    ImmutableBridgeRouter
{
    using MerkleProofParserLibrary for bytes;
    using MerkleProofLibrary for MerkleProofParserLibrary.MerkleProof;

    address internal _erc721Contract;

    struct UTXO {
        uint32 chainID;
        address owner;
        uint256 value;
        uint256 fee;
        bytes32 txHash;
    }

    constructor() ImmutableFactory(msg.sender) {}

    function initialize(address erc721Contract_) public onlyBridgeRouter initializer {
        _erc721Contract = erc721Contract_;
    }

    /// @notice Transfer tokens from sender and emit a "Deposited" event for minting correspondent tokens in sidechain
    /// @param msgSender The address of ERC sender
    /// @param number The token ID of the NFT to be deposited
    function deposit(address msgSender, uint256 number) public onlyBridgeRouter {
        IERC721Transferable(_erc721Contract).safeTransferFrom(msgSender, address(this), number);
        uint8 bridgeType = 2;
        uint256 chainId = 1337;
        uint16 bridgeImplVersion = 1;
        bytes32 salt = BridgeRouter(_bridgeRouterAddress()).getBridgePoolSalt(
            _erc721Contract,
            bridgeType,
            chainId,
            bridgeImplVersion
        );
        BridgePoolDepositNotifier(_bridgePoolDepositNotifierAddress()).doEmit(
            salt,
            _erc721Contract,
            msgSender,
            bridgeType,
            number
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
        require(
            burnedUTXO.owner == msg.sender,
            string(
                abi.encodePacked(
                    BridgePoolErrorCodes.BRIDGEPOOL_RECEIVER_IS_NOT_OWNER_ON_PROOF_OF_BURN_UTXO
                )
            )
        );
        require(
            merkleProof.checkProof(bClaims.stateRoot, merkleProof.computeLeafHash()),
            string(abi.encodePacked(BridgePoolErrorCodes.BRIDGEPOOL_COULD_NOT_VERIFY_PROOF_OF_BURN))
        );
        IERC721Transferable(_erc721Contract).safeTransferFrom(
            address(this),
            msg.sender,
            burnedUTXO.value // tokenId
        );
    }
}
