// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/LocalERCBridgePoolBase.sol";

contract LocalERCBridgePoolMock is Initializable, LocalERCBridgePoolBase {
    address internal _ercContract;

    function initialize(address ercContract_) public onlyFactory initializer {
        _ercContract = ercContract_;
    }

    function deposit(address msgSender, bytes calldata depositParameters_) public override {
        super.deposit(msgSender, depositParameters_);
    }

    function withdraw(bytes memory vsPreImage_, bytes[4] memory proofs_)
        public
        override
        returns (
            bytes32,
            address,
            uint256
        )
    {
        (bytes32 computedUTXOID, address account, uint256 value) = super.withdraw(
            vsPreImage_,
            proofs_
        );
        return (computedUTXOID, account, value);
    }
}
