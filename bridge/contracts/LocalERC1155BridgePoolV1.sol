// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "@openzeppelin/contracts/token/ERC1155/utils/ERC1155Holder.sol";
import "contracts/interfaces/IERC1155Transferable.sol";
import "contracts/LocalERCBridgePoolBase.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

/// @custom:salt LocalERC1155BridgePoolV1
/// @custom:deploy-type deployStatic
contract LocalERC1155BridgePoolV1 is
    ERC1155Holder,
    LocalERCBridgePoolBase,
    Initializable,
    ImmutableBridgeRouter
{
    address internal _erc1155Contract;

    function initialize(address erc1155Contract_) public onlyFactory initializer {
        _erc1155Contract = erc1155Contract_;
    }

    /// @notice Transfer tokens from sender and emit a "Deposited" event for minting correspondent tokens in sidechain
    /// @param msgSender The address of ERC sender
    /// @param depositParameters_ encoded deposit parameters (ERC1155:tokenId+tokenAmount)
    function deposit(address msgSender, bytes calldata depositParameters_)
        public
        override
        onlyBridgeRouter
    {
        DepositParameters memory _depositParameters = abi.decode(
            depositParameters_,
            (DepositParameters)
        );
        IERC1155Transferable(_erc1155Contract).safeTransferFrom(
            msgSender,
            address(this),
            _depositParameters.tokenId,
            _depositParameters.tokenAmount,
            abi.encodePacked("")
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
        IERC1155Transferable(_erc1155Contract).safeTransferFrom(
            address(this),
            msg.sender,
            burnedUTXO.tokenId,
            burnedUTXO.tokenAmount,
            abi.encodePacked("")
        );
    }
}
