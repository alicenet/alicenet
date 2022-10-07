// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "@openzeppelin/contracts/token/ERC721/utils/ERC721Holder.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/utils/EthSafeTransfer.sol";
import "contracts/utils/ERC20SafeTransfer.sol";
import "contracts/utils/MagicEthTransfer.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "contracts/interfaces/IStakingNFT.sol";
import "contracts/libraries/lockup/AccessControlled.sol";
import "contracts/RewardPool.sol";

/**
 * @notice This contract holds all ALCA that is held in escrow for lockup
 * bonuses. All ALCA is hold into a single staked position that is owned
 * locally. deployed by RewardPool
 */
contract BonusPool is
    ImmutableAToken,
    ImmutablePublicStaking,
    ImmutableFoundation,
    ERC20SafeTransfer,
    EthSafeTransfer,
    ERC721Holder,
    AccessControlled,
    MagicEthTransfer
{
    error BonusTokenAlreadyCreated();
    error BonusRateAlreadySet();
    error NotEnoughALCAToStake(uint256 currentBalance, uint256 expectedAmount);

    // one token 10^-18 per token not valid but placeholder
    // todo: needs to be a fraction of _SCALING_FACTOR
    uint256 internal constant _SCALING_FACTOR = 10**18;
    uint256 internal immutable _totalBonusAmount;
    address internal immutable _lockupContract;
    address internal immutable _rewardPool;
    uint256 internal _bonusRate;
    uint256 internal _tokenID;
    // the original amount of alca that we used to compute the _bonusRate (amount
    // that was locked at the end of the enrollment process in the lockup contract)
    uint256 internal _originalSharesLocked;

    constructor(
        address aliceNetFactory_,
        address lockupContract_,
        address rewardPool_,
        uint256 totalBonusAmount_
    )
        ImmutableFactory(aliceNetFactory_)
        ImmutableAToken()
        ImmutablePublicStaking()
        ImmutableFoundation()
    {
        _totalBonusAmount = totalBonusAmount_;
        _lockupContract = lockupContract_;
        _rewardPool = rewardPool_;
    }

    function getLockupContractAddress() public view returns (address) {
        return _getLockupContractAddress();
    }

    function getRewardPoolAddress() public view returns (address) {
        return _getRewardPoolAddress();
    }

    function createBonusStakedPosition() public onlyFactory {
        if (_tokenID != 0) {
            revert BonusTokenAlreadyCreated();
        }
        //get the total balance of ALCA owned by bonus pool as stake amount
        uint256 _stakeAmount = IERC20(_aTokenAddress()).balanceOf(address(this));
        uint256 totalBonusAmount = _totalBonusAmount;
        if (_stakeAmount < totalBonusAmount) {
            revert NotEnoughALCAToStake(_stakeAmount, totalBonusAmount);
        }
        _tokenID = IStakingNFT(_publicStakingAddress()).mint(totalBonusAmount);
    }

    // todo: Hunter double check this.
    function setBonusRate(uint256 totalLocked) public onlyLockup {
        if (_bonusRate != 0) {
            revert BonusRateAlreadySet();
        }
        _bonusRate = (_totalBonusAmount * _SCALING_FACTOR) / totalLocked;
        _originalSharesLocked = totalLocked;
    }

    function estimateBonusAmount(uint256 shares) public view returns (uint256) {
        return (shares * _bonusRate) / _SCALING_FACTOR;
    }

    function estimateBonusAmountWithReward(uint256 currentSharesLocked, uint256 shares)
        public
        view
        returns (uint256 estimatedTokenReward, uint256 estimatedEthReward)
    {
        uint256 estimatedPayoutEth = IStakingNFT(_publicStakingAddress()).estimateEthCollection(
            _tokenID
        );

        uint256 estimatedPayoutToken = IStakingNFT(_publicStakingAddress()).estimateTokenCollection(
            _tokenID
        );

        (, uint256 bonusRewardToken, uint256 bonusRewardEth) = _computeProportions(
            currentSharesLocked,
            estimatedPayoutEth,
            estimatedPayoutToken
        );

        uint256 userProportion = (shares * _SCALING_FACTOR) / currentSharesLocked;
        estimatedTokenReward = (shares * _bonusRate) / _SCALING_FACTOR;
        estimatedTokenReward += (userProportion * bonusRewardToken) / _SCALING_FACTOR;
        estimatedEthReward = (userProportion * bonusRewardEth) / _SCALING_FACTOR;
    }

    function terminate(uint256 finalSharesLocked) public onlyLockup {
        // burn the nft to collect all profits.
        (uint256 payoutEth, uint256 payoutToken) = IStakingNFT(_publicStakingAddress()).burn(
            _tokenID
        );

        // we subtract the shares amount from payoutToken to have the final amount of
        // ALCA yield gained by the bonus position
        payoutToken -= _totalBonusAmount;

        (
            uint256 bonusShares,
            uint256 bonusRewardToken,
            uint256 bonusRewardEth
        ) = _computeProportions(finalSharesLocked, payoutEth, payoutToken);

        _safeTransferERC20(
            IERC20Transferable(_aTokenAddress()),
            _getRewardPoolAddress(),
            bonusShares + bonusRewardToken
        );

        RewardPool(_getRewardPoolAddress()).deposit{value: bonusRewardEth}(
            bonusShares + bonusRewardToken
        );

        // send the left overs of ALCA to the aliceNetFactory contract.
        _safeTransferERC20(
            IERC20Transferable(_aTokenAddress()),
            _factoryAddress(),
            IERC20(_aTokenAddress()).balanceOf(address(this))
        );

        // send the left overs of ether to the foundation contract.
        _safeTransferEthWithMagic(IMagicEthTransfer(_foundationAddress()), address(this).balance);
        // todo: self destruct?
    }

    function _computeProportions(
        uint256 currentSharesLocked,
        uint256 payoutEth,
        uint256 payoutToken
    )
        internal
        view
        returns (
            uint256 bonusShares,
            uint256 bonusRewardEth,
            uint256 bonusRewardToken
        )
    {
        uint256 scaledStakedProportion = (currentSharesLocked * _SCALING_FACTOR) /
            _originalSharesLocked;
        // mathematical equivalent to:  (scaledStakedProportion * finalSharesLocked) / _SCALING_FACTOR
        bonusShares = (_bonusRate * currentSharesLocked) / _SCALING_FACTOR;
        bonusRewardToken = (scaledStakedProportion * payoutToken) / _SCALING_FACTOR;
        bonusRewardEth = (scaledStakedProportion * payoutEth) / _SCALING_FACTOR;
    }

    function _getLockupContractAddress() internal view override returns (address) {
        return _lockupContract;
    }

    function _getBonusPoolAddress() internal view override returns (address) {
        return address(this);
    }

    function _getRewardPoolAddress() internal view override returns (address) {
        return _rewardPool;
    }
}
