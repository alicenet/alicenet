// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC721/utils/ERC721Holder.sol";
import "contracts/interfaces/IERC721Transferable.sol";
import "contracts/NativeERCBridgePoolBase.sol";
import "contracts/utils/auth/ImmutableBridgePoolFactory.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

contract NativeERC721BridgePoolV1 is
    Initializable,
    ImmutableBridgeRouter,
    ERC721Holder,
    NativeERCBridgePoolBase,
    ImmutableBridgePoolFactory
{
    address internal _erc721Contract;

    constructor(
        address alicenetFactoryAddress,
        address snapshotsAddress
    ) NativeERCBridgePoolBase(alicenetFactoryAddress, snapshotsAddress) {}

    function initialize(address erc721Contract_) public onlyBridgePoolFactory initializer {
        _erc721Contract = erc721Contract_;
    }

    /// @notice Transfers token from sender to Bridge Pool
    /// @param sender The address of ERC sender
    /// @param depositParameters_ encoded deposit parameters (ERC20:tokenAmount, ERC721:tokenId or ERC1155:tokenAmount+tokenId)
    function deposit(address sender, bytes calldata depositParameters_) public virtual override {
        super.deposit(sender, depositParameters_);
        DepositParameters memory _depositParameters = abi.decode(
            depositParameters_,
            (DepositParameters)
        );
        IERC721Transferable(_erc721Contract).safeTransferFrom(
            sender,
            address(this),
            _depositParameters.tokenId
        );
    }

    /// @notice Tranfers token from Bridge Pool to receiver upon UTXO verification
    /// @param receiver The address of ERC receiver
    /// @param vsPreImage burned UTXO in L2
    /// @param proofs Proofs of inclusion of burned UTXO
    function withdraw(
        address receiver,
        bytes memory vsPreImage,
        bytes[4] memory proofs
    ) public virtual override returns (uint256 tokenId) {
        tokenId = super.withdraw(receiver, vsPreImage, proofs);
        IERC721Transferable(_erc721Contract).safeTransferFrom(address(this), receiver, tokenId);
    }
}
