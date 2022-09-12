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
    ImmutableBridgeRouter
{

    address internal _erc20Contract;

    function initialize(address erc20Contract_) public onlyFactory initializer {
        _erc20Contract = erc20Contract_;
    }

    /// @notice Transfer tokens from sender and emit a "Deposited" event for minting correspondent tokens in sidechain
    /// @param msgSender The address of ERC sender
    /// @param number The number of tokens to be deposited
    function deposit(address msgSender, uint256 number) public onlyBridgeRouter override {
        super.deposit(msgSender,number);
        IERC20Transferable(_erc20Contract).transferFrom(msgSender, address(this), number);
    }

    /// @notice Transfer tokens to sender upon a verificable proof of burn in sidechain
    /// @param encodedMerkleProof The merkle proof
    /// @param encodedBurnedUTXO The burned UTXO
    function withdraw(bytes memory encodedMerkleProof, bytes memory encodedBurnedUTXO) public override {
        super.withdraw(encodedMerkleProof, encodedBurnedUTXO);
        UTXO memory burnedUTXO = abi.decode(encodedBurnedUTXO, (UTXO));
        IERC20Transferable(_erc20Contract).transfer(msg.sender, burnedUTXO.value);
    }
}
