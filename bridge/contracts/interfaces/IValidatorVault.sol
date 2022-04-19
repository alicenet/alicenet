// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

interface IValidatorVault {
    function depositDilutionAdjustment(uint256 adjustmentAmount_) external;

    function depositStake(uint256 stakePosition_, uint256 amount_) external;

    function withdrawStake(uint256 stakePosition_) external returns (uint256);

    function skimExcessEth(address to_) external returns (uint256 excess);

    function skimExcessToken(address to_) external returns (uint256 excess);

    function estimateStakedAmount(uint256 stakePosition_) external view returns (uint256);
}
