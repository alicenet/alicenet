// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "./utils/ImmutableAuth.sol";
import "contracts/utils/ERC20SafeTransfer.sol";
import "contracts/utils/EthSafeTransfer.sol";
import "contracts/interfaces/IValidatorPool.sol";

/// @custom:salt ValidatorVault
/// @custom:deploy-type deployUpgradeable
contract ValidatorVault is
    ImmutableAToken,
    ImmutableATokenMinter,
    ImmutableValidatorPool,
    ERC20SafeTransfer,
    EthSafeTransfer
{
    struct Vault {
        uint256 _amount;
        uint256 _accumulator;
    }

    uint256 public globalAccumulator;
    uint256 public totalReserve;
    mapping(uint256 => Vault) internal _vaults;

    constructor()
        ImmutableFactory(msg.sender)
        ImmutableAToken()
        ImmutableValidatorPool()
        ImmutableATokenMinter()
    {}

    function depositDilutionAdjustment(uint256 adjustmentAmount_) public onlyATokenMinter {
        (uint256 userAdjustmentAmount, uint256 numValidators) = _getAdjustmentPerUser(
            adjustmentAmount_
        );
        uint256 totalAdjustmentAmount = userAdjustmentAmount * numValidators;

        _safeTransferFromERC20(IERC20Transferable(_aTokenAddress()), msg.sender, adjustmentAmount_);

        // we transfer the whole amount that was sent to this contract, but only keep track of the amount
        // that is perfectly divisible by the number of validators staked at the moment. The excess can be
        // retrieved via skimExcess function
        totalReserve += totalAdjustmentAmount;
        globalAccumulator += userAdjustmentAmount;
        // update minimum amount to become a validator to take into account the AToken dilution
        uint256 stakeAmount = IValidatorPool(_validatorPoolAddress()).getStakeAmount();
        stakeAmount += userAdjustmentAmount;
        IValidatorPool(_validatorPoolAddress()).setStakeAmount(stakeAmount);
    }

    function depositStake(uint256 stakePosition_, uint256 amount_) public onlyValidatorPool {
        _safeTransferFromERC20(IERC20Transferable(_aTokenAddress()), msg.sender, amount_);
        totalReserve += amount_;
        _vaults[stakePosition_] = Vault(amount_, globalAccumulator);
    }

    function withdrawStake(uint256 stakePosition_) public onlyValidatorPool returns (uint256) {
        Vault memory userVault = _updateVaultWithDilution(stakePosition_);
        _safeTransferERC20(IERC20Transferable(_aTokenAddress()), msg.sender, userVault._amount);
        totalReserve -= userVault._amount;
        delete _vaults[stakePosition_];
        return userVault._amount;
    }

    /// skimExcessEth will send to the address passed as to_ any amount of Eth held by this contract that
    /// is not tracked. This function allows the Admin role to refund any Eth sent to this contract in
    /// error by a user. This method can not return any funds sent to the contract via the depositEth
    /// method. This function should only be necessary if a user somehow manages to accidentally
    /// selfDestruct a contract with this contract as the recipient.
    function skimExcessEth(address to_) public onlyFactory returns (uint256 excess) {
        excess = address(this).balance;
        _safeTransferEth(to_, excess);
        return excess;
    }

    /// skimExcessToken will send to the address passed as to_ any amount of AToken held by this contract
    /// that is not tracked. This function allows the Admin role to refund any AToken sent to this
    /// contract in error by a user.
    function skimExcessToken(address to_) public onlyFactory returns (uint256 excess) {
        IERC20Transferable aToken = IERC20Transferable(_aTokenAddress());
        uint256 balance = aToken.balanceOf(address(this));
        require(
            balance >= totalReserve,
            "The balance of the contract is less then the tracked reserve!"
        );
        excess = balance - totalReserve;
        _safeTransferERC20(aToken, to_, excess);
        return excess;
    }

    function estimateStakedAmount(uint256 stakePosition_) public view returns (uint256) {
        Vault memory userVault = _updateVaultWithDilution(stakePosition_);
        return userVault._amount;
    }

    function getAdjustmentPrice(uint256 adjustmentAmount_) public view returns (uint256) {
        return _getAdjustmentPrice(adjustmentAmount_);
    }

    function getAdjustmentPerUser(uint256 adjustmentAmount_)
        public
        view
        returns (uint256 userAdjustmentAmount, uint256 numValidators)
    {
        return _getAdjustmentPerUser(adjustmentAmount_);
    }

    function _getAdjustmentPrice(uint256 adjustmentAmount_) internal view returns (uint256) {
        (uint256 userAdjustmentAmount, uint256 numValidators) = _getAdjustmentPerUser(
            adjustmentAmount_
        );
        uint256 totalAdjustmentAmount = userAdjustmentAmount * numValidators;
        return totalAdjustmentAmount;
    }

    function _getAdjustmentPerUser(uint256 adjustmentAmount_)
        internal
        view
        returns (uint256 userAdjustmentAmount, uint256 numValidators)
    {
        numValidators = IValidatorPool(_validatorPoolAddress()).getValidatorsCount();
        // if there's no validators there's no need to adjust the dilution
        if (numValidators == 0) {
            return (0, 0);
        }
        // the adjustmentAmount should be equally divisible by the number of validators staked at the moment
        userAdjustmentAmount = (adjustmentAmount_ / numValidators);
    }

    function _updateVaultWithDilution(uint256 stakePosition_)
        internal
        view
        returns (Vault memory userVault)
    {
        userVault = _vaults[stakePosition_];
        uint256 _globalAccumulator = globalAccumulator;
        uint256 deltaAccumulator;
        // handling a 'possible' overflow in the global accumulator
        if (userVault._accumulator > _globalAccumulator) {
            deltaAccumulator = type(uint256).max - userVault._accumulator;
            deltaAccumulator += _globalAccumulator;
        } else {
            deltaAccumulator = _globalAccumulator - userVault._accumulator;
        }
        userVault._amount += deltaAccumulator;
        userVault._accumulator = _globalAccumulator;
    }
}
