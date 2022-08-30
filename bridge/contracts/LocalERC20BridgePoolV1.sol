// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "contracts/interfaces/IERC20Transferable.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/libraries/errors/BridgePoolErrors.sol";
import "contracts/libraries/parsers/MerkleProofParserLibrary.sol";
import "contracts/utils/MerkleProofLibrary.sol";
import "contracts/interfaces/IBridgePool.sol";
import "contracts/Snapshots.sol";
import "contracts/libraries/parsers/BClaimsParserLibrary.sol";
import "contracts/utils/ERC20SafeTransfer.sol";
import "contracts/BridgePoolDepositNotifier.sol";

/// @custom:salt LocalERC20BridgePoolV1
/// @custom:deploy-type deployStatic

contract LocalERC20BridgePoolV1 is
    IBridgePool,
    Initializable,
    ImmutableSnapshots,
    ImmutableBridgePoolFactory,
    ImmutableBridgePoolDepositNotifier
{
    using MerkleProofParserLibrary for bytes;
    using MerkleProofLibrary for MerkleProofParserLibrary.MerkleProof;

    address internal _erc20Contract;

    struct UTXO {
        uint32 chainID;
        address owner;
        uint256 value;
        uint256 fee;
        bytes32 txHash;
    }

    constructor() ImmutableFactory(msg.sender) {}

    function initialize(address erc20Contract_) public onlyBridgePoolFactory initializer {
        _erc20Contract = erc20Contract_;
    }

    /// @notice Transfer tokens from sender and emit a "Deposited" event for minting correspondent tokens in sidechain
    /// @param msgSender The address of ERC sender
    /// @param number The number of tokens to be deposited
    function deposit(address msgSender, uint256 number) public {
        IERC20Transferable(_erc20Contract).transferFrom(msgSender, address(this), number);
        uint8 bridgeType = 1;
        uint256 chainId = 1337;
        uint16 bridgeImplVersion = 1;
        bytes32 salt = getBridgePoolSalt(_erc20Contract, bridgeType, chainId, bridgeImplVersion);
        BridgePoolDepositNotifier(_bridgePoolDepositNotifierAddress()).doEmit(
            salt,
            _erc20Contract,
            msgSender,
            bridgeType,
            number
        );
    }

    /// @notice Transfer tokens to sender upon a verificable proof of burn in sidechain
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

        IERC20Transferable(_erc20Contract).transfer(msg.sender, burnedUTXO.value);
    }

    // @notice calculates salt for a BridgePool contract based on ERC contract's address, tokenType, chainID and version_
    // @param tokenContractAddr_ address of ERC contract of BridgePool
    // @param tokenType_ type of token (1=ERC20, 2=ERC721)
    // @param version_ version of the implementation
    // @param chainID_ chain ID
    // @return calculated calculated salt
    function getBridgePoolSalt(
        address tokenContractAddr_,
        uint8 tokenType_,
        uint256 chainID_,
        uint16 version_
    ) public pure returns (bytes32) {
        return
            keccak256(
                bytes.concat(
                    keccak256(abi.encodePacked(tokenContractAddr_)),
                    keccak256(abi.encodePacked(tokenType_)),
                    keccak256(abi.encodePacked(chainID_)),
                    keccak256(abi.encodePacked(version_))
                )
            );
    }
}
