// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
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
import "hardhat/console.sol";

/// @custom:salt LocalERC20BridgePoolV1
/// @custom:deploy-type deployStatic
contract LocalERC20BridgePoolV1 is
    IBridgePool,
    Initializable,
    ERC20SafeTransfer,
    ImmutableSnapshots,
    ImmutableBridgePoolDepositNotifier,
    ImmutableBridgeRouter,
    ImmutableBToken
{
    using MerkleProofParserLibrary for bytes;
    using MerkleProofLibrary for MerkleProofParserLibrary.MerkleProof;

    address internal _ercTokenContract;

    struct UTXO {
        uint32 chainID;
        address owner;
        uint256 value;
        uint256 fee;
        bytes32 txHash;
    }

    constructor() ImmutableFactory(msg.sender) {}

    function initialize(address erc20TokenContract_) public onlyBridgeRouter initializer {
        _ercTokenContract = erc20TokenContract_;
    }

    /// @notice Transfer tokens from sender and emit a "Deposited" event for minting correspondent tokens in sidechain
    /// @param accountType_ The type of account
    /// @param aliceNetAddress_ The address on the sidechain where to mint the tokens
    /// @param ercValue ercAmount if ERC20 or tokenId if ERC721 or ERC1155
    /// @param bTokenAmount_ The fee for deposit in bTokens
    function deposit(
        uint8 accountType_,
        address aliceNetAddress_,
        uint256 ercValue,
        uint256 bTokenAmount_
    ) public {
        require(
            ERC20(_ercTokenContract).transferFrom(msg.sender, address(this), ercValue),
            string(
                abi.encodePacked(
                    BridgePoolErrorCodes.BRIDGEPOOL_UNABLE_TO_TRANSFER_DEPOSIT_AMOUNT_FROM_SENDER
                )
            )
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
        uint16 chainId = 1337;
        uint8 bridgeType = 1;
        uint8 bridgeImplVersion = 1;
        bytes32 salt = BridgeRouter(_bridgeRouterAddress()).getBridgePoolSalt(
            _ercTokenContract,
            bridgeType,
            chainId,
            bridgeImplVersion
        );
        BridgePoolDepositNotifier(_bridgePoolDepositNotifierAddress()).doEmit(
            salt,
            _ercTokenContract,
            msg.sender,
            bridgeType,
            ercValue
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
        _safeTransferERC20(IERC20Transferable(_ercTokenContract), msg.sender, burnedUTXO.value);
    }

    receive() external payable {}
}
