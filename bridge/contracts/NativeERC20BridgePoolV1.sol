// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "contracts/NativeERCBridgePoolBase.sol";
import "contracts/interfaces/IERC20Transferable.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/interfaces/IBridgePool.sol";
import "contracts/utils/ERC20SafeTransfer.sol";

/// @custom:salt LocalERC20BridgePoolV1
/// @custom:deploy-type deployUpgradeable
contract NativeERC20BridgePoolV1 is
    NativeERCBridgePoolBase,
    Initializable,
    ImmutableBridgeRouter,
    ERC20SafeTransfer
{
    address internal _erc20Contract;

    function initialize(address erc20Contract_) public onlyFactory initializer {
        _erc20Contract = erc20Contract_;
    }

    /// @notice Transfer tokens from sender and emit a "Deposited" event for minting correspondent tokens in sidechain
    /// @param msgSender The address of ERC sender
    /// @param depositParameters_ encoded deposit parameters (ERC20:tokenAmount, ERC721:tokenId or ERC1155:tokenAmount+tokenId)
    function deposit(address msgSender, bytes calldata depositParameters_) public onlyBridgeRouter {
        DepositParameters memory _depositParameters = abi.decode(
            depositParameters_,
            (DepositParameters)
        );
        _safeTransferFromERC20(
            IERC20Transferable(_erc20Contract),
            msgSender,
            _depositParameters.tokenAmount
        );
    }

    function withdraw(bytes memory vsPreImage, bytes[4] memory proofs) public {
        MerkleProofParserLibrary.MerkleProof memory proofAgainstStateRoot = super.verifyProofs(
            proofs
        );
        (, address account, uint256 value) = super.getValidatedTransferData(
            vsPreImage,
            proofAgainstStateRoot
        );
        _safeTransferERC20(IERC20Transferable(_erc20Contract), account, value);
    }
}
