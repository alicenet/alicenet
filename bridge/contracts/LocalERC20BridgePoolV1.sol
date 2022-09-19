// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "contracts/LocalERCBridgePoolBase.sol";
import "contracts/interfaces/IERC20Transferable.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/interfaces/IBridgePool.sol";
import "contracts/utils/ERC20SafeTransfer.sol";

/// @custom:salt LocalERC20BridgePoolV1
/// @custom:deploy-type deployStatic
contract LocalERC20BridgePoolV1 is
    LocalERCBridgePoolBase,
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
    function deposit(address msgSender, bytes calldata depositParameters_)
        public
        override
        onlyBridgeRouter
    {
        super.deposit(msgSender, depositParameters_);
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

    /// @notice Transfer tokens to sender upon a verificable proof of burn in sidechain
    /// @param encodedBurnedUTXO encoded UTXO burned in sidechain
    /// @param encodedMerkleProof merkle proof of burn
    function withdraw(bytes memory encodedBurnedUTXO, bytes memory encodedMerkleProof)
        public
        override
    {
        super.withdraw(encodedBurnedUTXO, encodedMerkleProof);
        UTXO memory burnedUTXO = abi.decode(encodedBurnedUTXO, (UTXO));
        _safeTransferERC20(IERC20Transferable(_erc20Contract), msg.sender, burnedUTXO.tokenAmount);
    }
}
