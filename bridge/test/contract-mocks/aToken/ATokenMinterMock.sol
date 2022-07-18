// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IAToken.sol";
import "contracts/interfaces/IValidatorVault.sol";
import "contracts/utils/ImmutableAuth.sol";

contract ATokenMinterMock is ImmutableAToken, ImmutableValidatorVault {
    constructor() ImmutableFactory(msg.sender) ImmutableAToken() ImmutableValidatorVault() {}

    function mint(address to, uint256 amount) public {
        // minting dilution adjustment
        uint256 dilutionAdjustment = amount / 10;
        uint256 adjustmentPrice = IValidatorVault(_validatorVaultAddress()).getAdjustmentPrice(
            dilutionAdjustment
        );

        IAToken(_aTokenAddress()).externalMint(to, amount + adjustmentPrice);

        IValidatorVault(_validatorVaultAddress()).depositDilutionAdjustment(adjustmentPrice);
    }
}
