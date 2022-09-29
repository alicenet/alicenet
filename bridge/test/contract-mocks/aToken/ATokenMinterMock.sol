// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/interfaces/IValidatorVault.sol";
import "contracts/interfaces/IStakingToken.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IERC20Transferable.sol";

contract ATokenMinterMock is ImmutableAToken, ImmutableValidatorVault {
    constructor() ImmutableFactory(msg.sender) ImmutableAToken() ImmutableValidatorVault() {}

    function mintWithDilution(address to, uint256 amount) public {
        // minting dilution adjustment
        uint256 dilutionAdjustment = amount / 10;
        uint256 adjustmentPrice = IValidatorVault(_validatorVaultAddress()).getAdjustmentPrice(
            dilutionAdjustment
        );

        IStakingToken(_aTokenAddress()).externalMint(to, amount);
        IStakingToken(_aTokenAddress()).externalMint(address(this), adjustmentPrice);

        IValidatorVault(_validatorVaultAddress()).depositDilutionAdjustment(adjustmentPrice);
    }

    function mint(address to, uint256 amount) public {
        IStakingToken(_aTokenAddress()).externalMint(to, amount);
    }

    function depositDilutionAdjustment(uint256 adjustmentPrice) public {
        IERC20Transferable(_aTokenAddress()).approve(_validatorVaultAddress(), adjustmentPrice);
        IValidatorVault(_validatorVaultAddress()).depositDilutionAdjustment(adjustmentPrice);
    }
}
