// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

interface IBToken {
    function setSplits(
        uint256 validatorStakingSplit_,
        uint256 publicStakingSplit_,
        uint256 liquidityProviderStakingSplit_,
        uint256 protocolFee_
    ) external;

    function virtualMintDeposit(
        uint8 accountType_,
        address to_,
        uint256 amount_
    ) external;
}
