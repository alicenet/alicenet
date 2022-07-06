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
import "contracts/BridgePoolFactory.sol";
import "hardhat/console.sol";

/// @custom:salt BridgePoolV1
/// @custom:deploy-type deployStatic
contract BridgePoolV1 is
    IBridgePool,
    Initializable,
    ImmutableSnapshots,
    ERC20SafeTransfer,
    ImmutableBridgePool,
    ImmutableBridgePoolDepositNotifier,
    ImmutableBridgePoolFactory,
    ImmutableBToken,
    MagicEthTransfer
{
    using MerkleProofParserLibrary for bytes;
    using MerkleProofLibrary for MerkleProofParserLibrary.MerkleProof;

    address internal _ercTokenContract;
    address internal _bTokenContract;

    struct UTXO {
        uint32 chainID;
        address owner;
        uint256 value;
        uint256 fee;
        bytes32 txHash;
    }

    constructor() ImmutableFactory(msg.sender) {}

    function initialize(address erc20TokenContract_, address bTokenContract_)
        public
        // onlyFactory
        initializer
    {
        _ercTokenContract = erc20TokenContract_;
        _bTokenContract = bTokenContract_;
    }

    /// @notice Transfer tokens from sender and emit a "Deposited" event for minting correspondent tokens in sidechain
    /// @param accountType_ The type of account
    /// @param aliceNetAddress_ The address on the sidechain where to mint the tokens
    /// @param ercAmount_ The amount of ERC tokens to deposit
    /// @param bTokenAmount_ The fee for deposit in bTokens
    function deposit(
        uint8 accountType_,
        address aliceNetAddress_,
        uint256 ercAmount_,
        uint256 bTokenAmount_
    ) public {
        require(
            ERC20(_ercTokenContract).transferFrom(msg.sender, address(this), ercAmount_),
            string(
                abi.encodePacked(
                    BridgePoolErrorCodes.BRIDGEPOOL_UNABLE_TO_TRANSFER_DEPOSIT_AMOUNT_FROM_SENDER
                )
            )
        );
        require(
            ERC20(_bTokenContract).transferFrom(msg.sender, address(this), bTokenAmount_),
            string(
                abi.encodePacked(
                    BridgePoolErrorCodes.BRIDGEPOOL_UNABLE_TO_TRANSFER_DEPOSIT_FEE_FROM_SENDER
                )
            )
        );
        uint256 value = BToken(_bTokenContract).burnTo(address(this), bTokenAmount_, 1);
        require(
            value > 0,
            string(abi.encodePacked(BridgePoolErrorCodes.BRIDGEPOOL_UNABLE_TO_BURN_DEPOSIT_FEE))
        );
        // for erc20 tokenID field is 0
        // To transfer eth (obtained from burning bToken fee) to BToken contract. Ssince it is not payable, we use this code that creates a contract that transfer value to bToken and then selfdestructs
        address btoken = _bTokenContract;
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
        /*         console.log(
            _bridgePoolFactoryAddress(),
            _bridgePoolDepositNotifierAddress(),
            _ercTokenContract
        ); */
        bytes32 salt = BridgePoolFactory(_bridgePoolFactoryAddress()).getLocalBridgePoolSalt(
            _ercTokenContract
        );
        BridgePoolDepositNotifier(_bridgePoolDepositNotifierAddress()).doEmit(
            salt,
            _ercTokenContract,
            ercAmount_,
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
        _safeTransferERC20(IERC20Transferable(_ercTokenContract), msg.sender, burnedUTXO.value);
    }

    receive() external payable {}
}
