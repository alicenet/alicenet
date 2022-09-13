// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/token/ERC721/utils/ERC721Holder.sol";
import "contracts/interfaces/IERC721Transferable.sol";
import "contracts/LocalERCBridgePoolBase.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

/// @custom:salt LocalERC721BridgePoolV1
/// @custom:deploy-type deployStatic
contract LocalERC721BridgePoolV1 is
    ERC721Holder,
    LocalERCBridgePoolBase,
    Initializable,
    ImmutableBridgeRouter
{
    address internal _erc721Contract;

    function initialize(address erc721Contract_) public onlyFactory initializer {
        _erc721Contract = erc721Contract_;
    }

    /// @notice Transfer tokens from sender and emit a "Deposited" event for minting correspondent tokens in sidechain
    /// @param msgSender The address of ERC sender
    /// @param number The number of tokens to be deposited
    function deposit(address msgSender, uint256 number) public override onlyBridgeRouter {
        super.deposit(msgSender, number);
        IERC721Transferable(_erc721Contract).safeTransferFrom(msgSender, address(this), number);
    }

    /// @notice Transfer tokens to sender upon a verificable proof of burn in sidechain
    /// @param encodedMerkleProof The merkle proof
    /// @param encodedBurnedUTXO The burned UTXO in sidechain
    function withdraw(bytes memory encodedMerkleProof, bytes memory encodedBurnedUTXO)
        public
        override
    {
        super.withdraw(encodedMerkleProof, encodedBurnedUTXO);
        UTXO memory burnedUTXO = abi.decode(encodedBurnedUTXO, (UTXO));
        IERC721Transferable(_erc721Contract).safeTransferFrom(
            address(this),
            msg.sender,
            burnedUTXO.value
        );
    }
}
