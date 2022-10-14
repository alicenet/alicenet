// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "contracts/interfaces/IERC721Transferable.sol";
import "contracts/interfaces/IStakingNFT.sol";
import "contracts/interfaces/IStakingToken.sol";
import "contracts/utils/EthSafeTransfer.sol";
import "contracts/utils/ERC20SafeTransfer.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/Lockup.sol";

contract StakingRouterV1 is
    ImmutablePublicStaking,
    ImmutableAToken,
    ERC20SafeTransfer,
    EthSafeTransfer
{
    error InvalidStakingAmount(uint256 stakingAmount, uint256 migratedAmount);

    address internal immutable _legacyToken;
    address internal immutable _lockupContract;

    constructor(address lockupContract_)
        ImmutableFactory(msg.sender)
        ImmutablePublicStaking()
        ImmutableAToken()
    {
        _legacyToken = IStakingToken(_aTokenAddress()).getLegacyTokenAddress();
        _lockupContract = lockupContract_;
    }

    /// @notice Migrates an amount of legacy token (MADToken) to ALCA tokens and stake them in the
    /// PublicStaking contract. User calling this function must have approved this contract to
    /// transfer the `migrationAmount_` MADTokens beforehand.
    /// @param to_ the address that will own the position
    /// @param migrationAmount_ the amount of legacy token to migrate
    /// @param stakingAmount_ the amount of ALCA that will staked and locked
    function migrateAndStake(
        address to_,
        uint256 migrationAmount_,
        uint256 stakingAmount_
    ) public {
        uint256 migratedAmount = _migrate(address(this), migrationAmount_);
        _verifyAndSendAnyRemainder(to_, migratedAmount, stakingAmount_);
        _stake(to_, migratedAmount);
    }

    /// @notice Migrates an amount of legacy token (MADToken) to ALCA tokens, stake them in the
    /// PublicStaking contract and in sequence lock the position. User calling this function must have
    /// approved this contract to transfer the `migrationAmount_` MADTokens beforehand.
    /// @param to_ the address that will own the locked position
    /// @param migrationAmount_ the amount of legacy token to migrate
    /// @param stakingAmount_ the amount of ALCA that will staked and locked
    function migrateStakeAndLock(
        address to_,
        uint256 migrationAmount_,
        uint256 stakingAmount_
    ) public {
        uint256 migratedAmount = _migrate(address(this), migrationAmount_);
        _verifyAndSendAnyRemainder(to_, migratedAmount, stakingAmount_);
        // mint the position directly to the lockup contract
        uint256 tokenID = _stake(_lockupContract, stakingAmount_);
        // right in sequence claim the minted position
        Lockup(payable(_lockupContract)).lockFromTransfer(tokenID, to_);
    }

    /// @notice Stake an amount of ALCA in the PublicStaking contract and lock the position in
    /// sequence. User calling this function must have approved this contract to
    /// transfer the `stakingAmount_` ALCA beforehand.
    /// @param to_ the address that will own the locked position
    /// @param stakingAmount_ the amount of ALCA that will staked
    function stakeAndLock(address to_, uint256 stakingAmount_) public {
        _safeTransferFromERC20(IERC20Transferable(_aTokenAddress()), address(this), stakingAmount_);
        // mint the position directly to the lockup contract
        uint256 tokenID = _stake(_lockupContract, stakingAmount_);
        // right in sequence claim the minted position
        Lockup(payable(_lockupContract)).lockFromTransfer(tokenID, to_);
    }

    /// @notice Get the address of the legacy token.
    /// @return the address of the legacy token (MADToken).
    function getLegacyTokenAddress() public view returns (address) {
        return _legacyToken;
    }

    function _migrate(address to_, uint256 amount_) internal returns (uint256 migratedAmount_) {
        _safeTransferFromERC20(IERC20Transferable(_legacyToken), address(this), amount_);
        IERC20(_legacyToken).approve(_aTokenAddress(), amount_);
        migratedAmount_ = IStakingToken(_aTokenAddress()).migrateTo(to_, amount_);
    }

    function _stake(address to_, uint256 stakingAmount_) internal returns (uint256 tokenID_) {
        IERC20(_aTokenAddress()).approve(_publicStakingAddress(), stakingAmount_);
        tokenID_ = IStakingNFT(_publicStakingAddress()).mintTo(to_, stakingAmount_, 0);
    }

    function _verifyAndSendAnyRemainder(
        address to_,
        uint256 migratedAmount_,
        uint256 stakingAmount_
    ) internal {
        if (stakingAmount_ > migratedAmount_) {
            revert InvalidStakingAmount(stakingAmount_, migratedAmount_);
        }
        uint256 remainder = migratedAmount_ - stakingAmount_;
        if (remainder > 0) {
            _safeTransferERC20(IERC20Transferable(_aTokenAddress()), to_, remainder);
        }
    }
}
