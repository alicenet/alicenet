// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/LocalERCBridgePoolBase.sol";

/// @custom:salt LocalERCBridgePoolMock
/// @custom:deploy-type deployStatic
contract LocalERCBridgePoolMock is Initializable, LocalERCBridgePoolBase {
    address internal _ercContract;

    function initialize(address ercContract_) public onlyFactory initializer {
        _ercContract = ercContract_;
    }

    /// @notice Transfer tokens from sender for minting them on sidechain
    /// @param msgSender The address of ERC sender
    /// @param depositParameters_ encoded deposit parameters (ERC20:tokenAmount, ERC721:tokenId or ERC1155:tokenAmount+tokenId)
    function deposit(address msgSender, bytes calldata depositParameters_) public override {
        super.deposit(msgSender, depositParameters_);
    }

    /// @notice Transfer tokens to sender upon a verificable proof of burn in sidechain
    /// @param encodedBurnedUTXO encoded UTXO burned in sidechain
    /// @param encodedMerkleProof merkle proof of burn
    function withdraw(bytes memory encodedBurnedUTXO, bytes memory encodedMerkleProof)
        public
        override
    {
        super.withdraw(encodedBurnedUTXO, encodedMerkleProof);
    }
}
