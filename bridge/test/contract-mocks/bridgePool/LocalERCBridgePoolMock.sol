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
    /// @param number The number of tokens to be deposited
    function deposit(address msgSender, uint256 number) public override {
        super.deposit(msgSender, number);
        //IERC20Transferable(_ercContract).transferFrom(msgSender, address(this), number);
    }

    /// @notice Transfer tokens to sender upon a verificable proof of burn in sidechain
    /// @param encodedMerkleProof The merkle proof
    /// @param encodedBurnedUTXO The burned UTXO
    function withdraw(bytes memory encodedMerkleProof, bytes memory encodedBurnedUTXO)
        public
        override
    {
        super.withdraw(encodedMerkleProof, encodedBurnedUTXO);
        // IERC20Transferable(_ercContract).transfer(msg.sender, 1);
    }
}
