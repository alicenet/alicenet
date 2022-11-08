// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "@openzeppelin/contracts/token/ERC1155/utils/ERC1155Holder.sol";
import "contracts/interfaces/IERC1155Transferable.sol";
import "contracts/NativeERCBridgePoolBase.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "hardhat/console.sol";

/// @custom:salt NativeERC1155BridgePoolV1
/// @custom:deploy-type deployUpgreadable
contract NativeERC1155BridgePoolV1 is
    Initializable,
    ImmutableBridgeRouter,
    ERC1155Holder,
    NativeERCBridgePoolBase
{
    address internal _erc1155Contract;

    function initialize(address erc1155Contract_) public onlyFactory initializer {
        _erc1155Contract = erc1155Contract_;
    }

    /// @notice Transfer tokens from sender and emit a "Deposited" event for minting correspondent tokens in sidechain
    /// @param msgSender The address of ERC sender
    /// @param depositParameters_ encoded deposit parameters (ERC20:tokenAmount, ERC721:tokenId or ERC1155:tokenAmount+tokenId)
    function deposit(address msgSender, bytes calldata depositParameters_) public virtual override {
        super.deposit(msgSender, depositParameters_);
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

    /// @notice Transfer multiple tokens from sender in one operation and emit a "Deposited" event for minting correspondent tokens in sidechain
    /// @param msgSender The address of ERC sender
    /// @param tokenIds array of tokenIds to transfer
    /// @param tokenAmounts array of tokenAmounts to transfer
    function batchDeposit(
        address msgSender,
        uint256[] calldata tokenIds,
        uint256[] calldata tokenAmounts
    ) public onlyBridgeRouter {
        IERC1155Transferable(_erc1155Contract).safeBatchTransferFrom(
            msgSender,
            address(this),
            tokenIds,
            tokenAmounts,
            abi.encodePacked("")
        );
    }

    /// @notice Transfer tokens to sender upon a verificable proof of burn in sidechain
    /// @param vsPreImage UTXO burned in L2
    /// @param proofs merkle proofs of burn
    function withdraw(bytes memory vsPreImage, bytes[4] memory proofs)
        public
        virtual
        override
        returns (address account, uint256 value)
    {
        (account, value) = super.withdraw(vsPreImage, proofs);
        IERC1155Transferable(_erc1155Contract).safeTransferFrom(
            address(this),
            msg.sender,
            value,
            value,
            abi.encodePacked("")
        );
    }
}
