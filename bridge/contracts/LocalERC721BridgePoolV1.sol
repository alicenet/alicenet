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
import "contracts/BridgePoolFactory.sol";
import "hardhat/console.sol";

/// @custom:salt LocalERC721BridgePoolV1
/// @custom:deploy-type deployStatic

contract LocalERC721BridgePoolV1 is
    ERC721Holder,
    IBridgePool,
    Initializable,
    ImmutableSnapshots,
    ImmutableBridgePool,
    ImmutableBToken,
    ImmutableBridgePoolDepositNotifier,
    ImmutableBridgePoolFactory
{
    using MerkleProofParserLibrary for bytes;
    using MerkleProofLibrary for MerkleProofParserLibrary.MerkleProof;

    address internal _erc721SourceContract;

    struct UTXO {
        uint32 chainID;
        address owner;
        uint256 value;
        uint256 fee;
        bytes32 txHash;
    }

    constructor() ImmutableFactory(msg.sender) {}

    function initialize(address erc721SourceContract_) public onlyBridgePoolFactory initializer {
        _erc721SourceContract = erc721SourceContract_;
    }

    /// @notice Transfer tokens from sender and emit a "Deposited" event for minting correspondent tokens in sidechain
    /// @param accountType_ The type of account
    /// @param aliceNetAddress_ The address on the sidechain where to mint the tokens
    /// @param tokenId_ The token ID of the NFT to be deposited
    /// @param bTokenAmount_ The fee for deposit in bTokens
    function deposit(
        uint8 accountType_,
        address aliceNetAddress_,
        uint256 tokenId_,
        uint256 bTokenAmount_
    ) public {
        IERC721Transferable(_erc721SourceContract).safeTransferFrom(
            msg.sender,
            address(this),
            tokenId_
        );
        require(
            ERC20(_bTokenAddress()).transferFrom(msg.sender, address(this), bTokenAmount_),
            string(
                abi.encodePacked(
                    BridgePoolErrorCodes.BRIDGEPOOL_UNABLE_TO_TRANSFER_DEPOSIT_FEE_FROM_SENDER
                )
            )
        );
        uint256 value = BToken(_bTokenAddress()).burnTo(address(this), bTokenAmount_, 1);
        require(
            value > 0,
            string(abi.encodePacked(BridgePoolErrorCodes.BRIDGEPOOL_UNABLE_TO_BURN_DEPOSIT_FEE))
        );
        address btoken = _bTokenAddress();
        assembly {
            mstore8(0x00, 0x73)
            mstore(0x01, shl(96, btoken))
            mstore8(0x15, 0xff)
            let addr := create(value, 0x00, 0x16)
            if iszero(addr) {
                returndatacopy(0x00, 0x00, returndatasize())
                revert(0x00, returndatasize())
            }
        }
        bytes32 salt = BridgePoolFactory(_bridgePoolFactoryAddress()).getLocalBridgePoolSalt(
            _erc721SourceContract
        );
        BridgePoolDepositNotifier(_bridgePoolDepositNotifierAddress()).doEmit(
            salt,
            _erc721SourceContract,
            tokenId_,
            msg.sender
        );
    }

    /// @notice Transfer funds to sender upon a verificable proof of burn in sidechain
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
        IERC721Transferable(_erc721SourceContract).safeTransferFrom(
            address(this),
            msg.sender,
            burnedUTXO.value // tokenId
        );
    }

    receive() external payable {}
}
