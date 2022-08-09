// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

interface IStakingToken {
    function migrate(uint256 amount) external;

    function externalMint(address to, uint256 amount) external;

    function externalBurn(address from, uint256 amount) external;

    function allowMigration() external;

    function getLegacyTokenAddress() external view returns (address);
}

interface IStakingTokenMinter {
    function mint(address to, uint256 amount) external;
}

interface IStakingTokenBurner {
    function burn(address to, uint256 amount) external;
}
