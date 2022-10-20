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
import "contracts/Lockup.sol";

/**
 * @notice This contract holds all ALCA that is held in escrow for lockup
 * bonuses. All ALCA is hold into a single staked position that is owned
 * locally.
 * @dev deployed by the RewardPool contract
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
    event BonusPositionCreated(uint256 tokenID);
    error BonusTokenNotCreated();
    error BonusTokenAlreadyCreated();
    error BonusRateAlreadySet();
    error InvalidOriginalSharesValue();
    error NotEnoughALCAToStake(uint256 currentBalance, uint256 expectedAmount);
    error AddressNotAllowedToSendEther();

    uint256 public constant SCALING_FACTOR = 10**18;
    uint256 internal immutable _totalBonusAmount;
    address internal immutable _lockupContract;
    address internal immutable _rewardPool;
    // tokenID of the position created to hold the amount that will be redistributed as bonus
    uint256 internal _tokenID;

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

    receive() external payable {
        if (msg.sender != _publicStakingAddress()) {
            revert AddressNotAllowedToSendEther();
        }
    }

    /// @notice function that creates/mint a publicStaking position with an amount that will be
    /// redistributed as bonus at the end of the lockup period. The amount of ALCA has to be
    /// transferred before calling this function.
    /// @dev can be only called by the AliceNet factory
    function createBonusStakedPosition() public onlyFactory {
        if (_tokenID != 0) {
            revert BonusTokenAlreadyCreated();
        }
        IERC20 alca = IERC20(_aTokenAddress());
        //get the total balance of ALCA owned by bonus pool as stake amount
        uint256 _stakeAmount = alca.balanceOf(address(this));
        uint256 totalBonusAmount = _totalBonusAmount;
        if (_stakeAmount < totalBonusAmount) {
            revert NotEnoughALCAToStake(_stakeAmount, totalBonusAmount);
        }
        // approve the staking contract to transfer the ALCA
        alca.approve(_publicStakingAddress(), totalBonusAmount);
        uint256 tokenID = IStakingNFT(_publicStakingAddress()).mint(totalBonusAmount);
        _tokenID = tokenID;
        emit BonusPositionCreated(_tokenID);
    }

    /// @notice gets the lockup contract address
    /// @return the lockup contract address
    function getLockupContractAddress() public view returns (address) {
        return _getLockupContractAddress();
    }

    /// @notice gets the rewardPool contract address
    /// @return the rewardPool contract address
    function getRewardPoolAddress() public view returns (address) {
        return _getRewardPoolAddress();
    }

    /// @notice gets the scaled bonusRate, rate in which the shares will be multiplied to determine
    /// the bonus amount owed by a position
    /// @dev the _bonusRate has a scaling factor built in, since it can be less than 0, to get its
    /// real value with decimal points, divide it by the SCALING_FACTOR
    /// @return the scaled Bonus Rate
    function getScaledBonusRate(uint256 originalSharesLocked_) public view returns (uint256) {
        return _getBonusRate(originalSharesLocked_);
    }

    /// @notice gets the tokenID of the publicStaking position that has the whole bonus amount
    /// @return the tokenID of the publicStaking position that has the whole bonus amount
    function getBonusStakedPosition() public view returns (uint256) {
        return _tokenID;
    }

    /// @notice estimates a user's bonus amount + bonus position profits.
    /// @dev a user profit can be determined by:
    /// (currentLockedShares/expectedLockedShares * userShares/currentLockedShares) * profit
    /// @param currentSharesLocked_ The current number of shares locked in the lockup contract
    /// @param userShares_ The amount of shares that a user locked-up.
    /// @return The estimated amount of ALCA bonus , ether profits and ALCA profits
    function estimateBonusAmountWithReward(
        uint256 currentSharesLocked_,
        uint256 originalSharesLocked_,
        uint256 userShares_
    )
        public
        view
        returns (
            uint256,
            uint256,
            uint256
        )
    {
        if (_tokenID == 0) {
            return (0, 0, 0);
        }
        if (originalSharesLocked_ == 0 || currentSharesLocked_ > originalSharesLocked_) {
            revert InvalidOriginalSharesValue();
        }

        (uint256 estimatedPayoutEth, uint256 estimatedPayoutToken) = IStakingNFT(
            _publicStakingAddress()
        ).estimateAllProfits(_tokenID);

        // computing the proportion of reward that will be sent to the rewardPool based on the
        // number of expected users that locked-up versus the current number of users that still
        // have positions locked.
        uint256 proportion = (currentSharesLocked_ * SCALING_FACTOR) / originalSharesLocked_;
        (, uint256 bonusRewardEth, uint256 bonusRewardToken) = _computeBonusByProportions(
            originalSharesLocked_,
            proportion,
            currentSharesLocked_,
            estimatedPayoutEth,
            estimatedPayoutToken
        );

        // compute what will be amount that a user will receive from the amount that will be sent to
        // the reward contract.
        uint256 userProportion = (userShares_ * SCALING_FACTOR) / currentSharesLocked_;
        return
            _computeBonusByProportions(
                originalSharesLocked_,
                userProportion,
                userShares_,
                bonusRewardEth,
                bonusRewardToken
            );
    }

    /// @notice Burns that bonus staked position, and send the bonus amount of shares + profits to
    /// the rewardPool contract, so users can collect.
    /// @dev The amount sent to the rewardPool contract is determined by the initial amount of
    /// users that locked their positions versus the final amount of users that kept their position
    /// locked until the end.
    /// @param finalSharesLocked_ The final amount of shares locked up in the lockup contract.
    function terminate(uint256 finalSharesLocked_, uint256 originalSharesLocked_)
        public
        onlyLockup
    {
        if (_tokenID == 0) {
            revert BonusTokenNotCreated();
        }
        if (originalSharesLocked_ == 0 || finalSharesLocked_ > originalSharesLocked_) {
            revert InvalidOriginalSharesValue();
        }
        // burn the nft to collect all profits.
        (uint256 payoutEth, uint256 payoutToken) = IStakingNFT(_publicStakingAddress()).burn(
            _tokenID
        );
        // restarting the _tokenID
        _tokenID = 0;

        // we subtract the shares amount from payoutToken to have the final amount of ALCA yield
        // gained by the bonus position
        payoutToken -= _totalBonusAmount;
        uint256 proportion = (finalSharesLocked_ * SCALING_FACTOR) / originalSharesLocked_;
        (
            uint256 bonusShares,
            uint256 bonusRewardEth,
            uint256 bonusRewardToken
        ) = _computeBonusByProportions(
                originalSharesLocked_,
                proportion,
                finalSharesLocked_,
                payoutEth,
                payoutToken
            );

        _safeTransferERC20(
            IERC20Transferable(_aTokenAddress()),
            _getRewardPoolAddress(),
            bonusShares + bonusRewardToken
        );

        RewardPool(_getRewardPoolAddress()).deposit{value: bonusRewardEth}(
            bonusShares + bonusRewardToken
        );

        uint256 tokenBalance = IERC20(_aTokenAddress()).balanceOf(address(this));
        uint256 ethBalance = address(this).balance;
        if (tokenBalance > 0) {
            // send the left overs of ALCA to the aliceNetFactory contract.
            _safeTransferERC20(
                IERC20Transferable(_aTokenAddress()),
                _factoryAddress(),
                IERC20(_aTokenAddress()).balanceOf(address(this))
            );
        }
        if (ethBalance > 0) {
            // send the left overs of ether to the foundation contract.
            _safeTransferEthWithMagic(IMagicEthTransfer(_foundationAddress()), ethBalance);
        }
    }

    function _computeBonusByProportions(
        uint256 originalShares_,
        uint256 proportion_,
        uint256 shares_,
        uint256 payoutEth_,
        uint256 payoutToken_
    )
        internal
        view
        returns (
            uint256 bonusShares,
            uint256 bonusRewardEth,
            uint256 bonusRewardToken
        )
    {
        // mathematical equivalent to:
        // (proportion * shares) / _SCALING_FACTOR
        bonusShares = (_getBonusRate(originalShares_) * shares_) / SCALING_FACTOR;
        bonusRewardEth = (proportion_ * payoutEth_) / SCALING_FACTOR;
        bonusRewardToken = (proportion_ * payoutToken_) / SCALING_FACTOR;
    }

    function _getBonusRate(uint256 originalTotalLocked_) internal view returns (uint256) {
        return (_totalBonusAmount * SCALING_FACTOR) / originalTotalLocked_;
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
