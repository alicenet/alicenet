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
    /// @param _bClaims informed bClaims
    /// @param _bClaimsSigGroup bClaims signature
    /// @param _txInPreImage withdraw (burn) tx
    /// @param _proofs inclusion proofs for StateRoot, TxRoot, TxHash and HeaderRoot
    function withdraw(
        bytes memory _bClaims,
        bytes memory _bClaimsSigGroup,
        bytes memory _txInPreImage,
        bytes[4] memory _proofs
    )
        public
        override
    {
        super.withdraw(_bClaims, _bClaimsSigGroup, _txInPreImage,_proofs);
    }
}
