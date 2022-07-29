// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library ValidatorPoolErrors {
    error CallerNotValidator(address caller);
    error ConsensusRunning();
    error ETHDKGRoundRunning();
    error OnlyNFTContractsAllowed();
    error MinimumBlockIntervalNotMet();
    error NotEnoughValidatorSlotsAvailable(uint256 requiredSlots, uint256 availableSlots);
    error RegistrationParameterLengthMismatch(
        uint256 validatorsLength,
        uint256 stakerTokenIDsLength
    );
    error FactoryShouldOwnPosition(uint256 positionId);
    error LengthGreaterThanAvailableValidators(uint256 length, uint256 availableValidators);
    error ProfitsNotClaimableWhileConsensusNotRunning();
    error TokenBalanceChangedDuringOperation();
    error EthBalanceChangedDuringOperation();
    error SenderNotInExitingQueue(address sender);
    error WaitingPeriodNotMet();
    error DishonestValidatorNotAccusable(address validator);
    error InvalidIndex(uint256 index);
    error AddressAlreadyValidator(address addr);
    error AddressNotValidator(address addr);
    error PayoutTooLow();
    error InsufficientFundsInStakePosition(uint256 stakeAmount, uint256 minimumRequiredAmount);
}
