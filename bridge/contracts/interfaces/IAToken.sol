// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

interface IAToken {
    function migrate(uint256 amount) external;

    function externalMint(address to, uint256 amount) external;

    function externalBurn(address from, uint256 amount) external;

    function allowMigration() external;

    function getLegacyTokenAddress() external view returns (address);
}

interface IATokenMinter {
    function mint(address to, uint256 amount) external;
}

interface IATokenBurner {
    function burn(address to, uint256 amount) external;
}
