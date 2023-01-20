// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "contracts/NativeERCBridgePoolBase.sol";
import "contracts/interfaces/IERC20Transferable.sol";
import "contracts/utils/ImmutableBridgePoolFactAuth.sol";
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

    constructor(address alicenetFactoryAddress, address snapshotsAddress)
        NativeERCBridgePoolBase(alicenetFactoryAddress, snapshotsAddress)
    {}

    function initialize(address erc20Contract_) public onlyBridgePoolFactory initializer {
        _erc20Contract = erc20Contract_;
    }

    /// @notice Transfer tokens from sender to Bridge Pool
    /// @param sender The address of ERC sender
    /// @param depositParameters_ encoded deposit parameters (ERC20:tokenAmount, ERC721:tokenId or ERC1155:tokenAmount+tokenId)
    function deposit(address sender, bytes calldata depositParameters_) public virtual override {
        super.deposit(sender, depositParameters_);
        DepositParameters memory _depositParameters = abi.decode(
            depositParameters_,
            (DepositParameters)
        );
        _safeTransferFromERC20(
            IERC20Transferable(_erc20Contract),
            sender,
            _depositParameters.tokenAmount
        );
    }

    /// @notice Transfer tokens from Bridge Pool to receiver upon proofs verification
    /// @param receiver The address of ERC receiver
    /// @param encodedVsPreImage burned UTXO in chain
    /// @param proofs Proofs of inclusion of burned UTXO
    function withdraw(
        address receiver,
        bytes memory encodedVsPreImage,
        bytes[4] memory proofs
    ) public virtual override returns (uint256 amount) {
       amount = super.withdraw(receiver, encodedVsPreImage, proofs);
        _safeTransferERC20(IERC20Transferable(_erc20Contract), receiver, amount);
    }
}
