// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import "./utils/ImmutableAuth.sol";
import "contracts/interfaces/IDistribution.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/utils/MagicEthTransfer.sol";
import "contracts/utils/EthSafeTransfer.sol";
import "contracts/libraries/errors/BTokenErrors.sol";
import "contracts/utils/Mutex.sol";

/// @custom:salt Distribution
/// @custom:deploy-type deployUpgradeable
contract Distribution is
    IDistribution,
    Initializable,
    Mutex,
    MagicEthTransfer,
    EthSafeTransfer,
    ImmutableFactory,
    ImmutablePublicStaking,
    ImmutableValidatorStaking,
    ImmutableLiquidityProviderStaking,
    ImmutableFoundation
{
    // Scaling factor to get the staking percentages
    uint256 internal constant _PERCENTAGE_SCALE = 1000;

    // Value of the percentages that will send to each staking contract. Divide
    // this value by _PERCENTAGE_SCALE = 1000 to get the corresponding percentages.
    // These values must sum to 1000.
    Splits internal _splits;

    constructor()
        ImmutableFactory(msg.sender)
        ImmutablePublicStaking()
        ImmutableValidatorStaking()
        ImmutableLiquidityProviderStaking()
        ImmutableFoundation()
    {}

    function initialize() public onlyFactory initializer {
        _setSplitsInternal(332, 332, 332, 4);
    }

    /// @dev sets the percentage that will be divided between all the staking
    /// contracts, must only be called by _admin
    function setSplits(
        uint256 validatorStakingSplit_,
        uint256 publicStakingSplit_,
        uint256 liquidityProviderStakingSplit_,
        uint256 protocolFee_
    ) public onlyFactory {
        _setSplitsInternal(
            validatorStakingSplit_,
            publicStakingSplit_,
            liquidityProviderStakingSplit_,
            protocolFee_
        );
    }

    /// Gets the value of the percentages that will send to each staking contract.
    /// Divide this value by _PERCENTAGE_SCALE = 1000 to get the corresponding
    /// percentages.
    function getSplits() public view returns (Splits memory) {
        return _splits;
    }

    /// Distributes the yields of the BToken sale to all stakeholders
    function distribute() public returns (bool) {
        return _distribute();
    }

    function _setSplitsInternal(
        uint256 validatorStakingSplit_,
        uint256 publicStakingSplit_,
        uint256 liquidityProviderStakingSplit_,
        uint256 protocolFee_
    ) internal {
        if (
            validatorStakingSplit_ +
                publicStakingSplit_ +
                liquidityProviderStakingSplit_ +
                protocolFee_ !=
            _PERCENTAGE_SCALE
        ) {
            revert BTokenErrors.SplitValueSumError();
        }
        _splits = Splits(
            uint32(validatorStakingSplit_),
            uint32(publicStakingSplit_),
            uint32(liquidityProviderStakingSplit_),
            uint32(protocolFee_)
        );
    }

    /// Distributes the yields from the BToken minting to all stake holders.
    function _distribute() internal returns (bool) {
        uint256 excess = address(this).balance;
        Splits memory splits = _splits;
        // take out protocolFee from excess and decrement excess
        uint256 foundationAmount = (excess * splits.protocolFee) / _PERCENTAGE_SCALE;
        // split remaining between miners, stakers and lp stakers
        uint256 stakingAmount = (excess * splits.publicStaking) / _PERCENTAGE_SCALE;
        uint256 lpStakingAmount = (excess * splits.liquidityProviderStaking) / _PERCENTAGE_SCALE;
        // then give miners the difference of the original and the sum of the
        // stakingAmount
        uint256 minerAmount = excess - (stakingAmount + lpStakingAmount + foundationAmount);

        if (foundationAmount != 0) {
            _safeTransferEthWithMagic(IMagicEthTransfer(_foundationAddress()), foundationAmount);
        }
        if (minerAmount != 0) {
            _safeTransferEthWithMagic(IMagicEthTransfer(_validatorStakingAddress()), minerAmount);
        }
        if (stakingAmount != 0) {
            _safeTransferEthWithMagic(IMagicEthTransfer(_publicStakingAddress()), stakingAmount);
        }
        if (lpStakingAmount != 0) {
            _safeTransferEthWithMagic(
                IMagicEthTransfer(_liquidityProviderStakingAddress()),
                lpStakingAmount
            );
        }
        // invariants hold
        return true;
    }
}
