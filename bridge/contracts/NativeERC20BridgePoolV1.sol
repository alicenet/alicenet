// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "contracts/NativeERCBridgePoolBase.sol";
import "contracts/interfaces/IERC20Transferable.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/interfaces/IBridgePool.sol";
import "contracts/utils/ERC20SafeTransfer.sol";

/// @custom:salt NativeERC20BridgePoolV1
/// @custom:deploy-type deployStatic
contract NativeERC20BridgePoolV1 is
    NativeERCBridgePoolBase,
    Initializable,
    ERC20SafeTransfer,
    ImmutableBridgePoolFactory
{
    address internal _erc20Contract;

    constructor(address alicenetFactoryAddress) NativeERCBridgePoolBase(alicenetFactoryAddress) {}

    function initialize(address erc20Contract_) public onlyBridgePoolFactory initializer {
        _erc20Contract = erc20Contract_;
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
        _safeTransferFromERC20(
            IERC20Transferable(_erc20Contract),
            msgSender,
            _depositParameters.tokenAmount
        );
    }

    function withdraw(bytes memory vsPreImage, bytes[4] memory proofs)
        public
        virtual
        override
        returns (address account, uint256 value)
    {
        (account, value) = super.withdraw(vsPreImage, proofs);
        _safeTransferERC20(IERC20Transferable(_erc20Contract), account, value);
    }
}
