// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;
struct Deposit {
    uint8 accountType;
    address account;
    uint256 value;
}

interface IUtilityToken {
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
    ) external returns (uint256);

    function distribute()
        external
        returns (
            uint256 minerAmount,
            uint256 stakingAmount,
            uint256 lpStakingAmount,
            uint256 foundationAmount
        );

    function deposit(
        uint8 accountType_,
        address to_,
        uint256 amount_
    ) external returns (uint256);

    function mintDeposit(
        uint8 accountType_,
        address to_,
        uint256 minBTK_
    ) external payable returns (uint256);

    function mint(uint256 minBTK_) external payable returns (uint256 numBTK);

    function mintTo(address to_, uint256 minBTK_) external payable returns (uint256 numBTK);

    function burn(uint256 amount_, uint256 minEth_) external returns (uint256 numEth);

    function burnTo(
        address to_,
        uint256 amount_,
        uint256 minEth_
    ) external returns (uint256 numEth);

    function getPoolBalance() external returns (uint256);

    function getTotalBTokensDeposited() external returns (uint256);

    function getDeposit(uint256 depositID) external returns (Deposit memory);

    function bTokensToEth(
        uint256 poolBalance_,
        uint256 totalSupply_,
        uint256 numBTK_
    ) external returns (uint256 numEth);

    function ethToBTokens(uint256 poolBalance_, uint256 numEth_) external returns (uint256);
}
