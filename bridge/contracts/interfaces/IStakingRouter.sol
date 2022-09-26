// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

interface IStakingRouter {
    function migrate() external; 
    function migrateStake() external;
    function stakeRestake() external;
    function migrateStakeRestake() external;
    function unrestake() external;
    function unstake() external;
    function unstakeUnrestake() external;
}
