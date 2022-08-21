// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IStakingNFT.sol";
import "contracts/interfaces/IERC20Transferable.sol";
import "contracts/interfaces/IERC721Transferable.sol";
import "contracts/interfaces/ISnapshots.sol";
import "contracts/interfaces/IETHDKG.sol";
import "contracts/interfaces/IValidatorPool.sol";
import "contracts/utils/EthSafeTransfer.sol";
import "contracts/utils/ERC20SafeTransfer.sol";
import "contracts/utils/MagicValue.sol";
import "contracts/utils/CustomEnumerableMaps.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/libraries/validatorPool/ValidatorPoolStorage.sol";
import "contracts/interfaces/IERC20Transferable.sol";
import "contracts/interfaces/IStakingNFT.sol";
import "@openzeppelin/contracts/token/ERC721/utils/ERC721Holder.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "contracts/libraries/errors/ValidatorPoolErrors.sol";

/// @custom:salt ValidatorPool
/// @custom:deploy-type deployUpgradeable
contract ValidatorPool is
    Initializable,
    ValidatorPoolStorage,
    IValidatorPool,
    MagicValue,
    EthSafeTransfer,
    ERC20SafeTransfer,
    ERC721Holder
{
    using CustomEnumerableMaps for ValidatorDataMap;

    modifier onlyValidator() {
        if (!_isValidator(msg.sender)) {
            revert ValidatorPoolErrors.CallerNotValidator(msg.sender);
        }
        _;
    }

    modifier assertNotConsensusRunning() {
        if (_isConsensusRunning) {
            revert ValidatorPoolErrors.ConsensusRunning();
        }
        _;
    }

    modifier assertNotETHDKGRunning() {
        if (IETHDKG(_ethdkgAddress()).isETHDKGRunning()) {
            revert ValidatorPoolErrors.ETHDKGRoundRunning();
        }
        _;
    }

    modifier balanceShouldNotChange() {
        uint256 balanceBeforeToken = IERC20Transferable(_aTokenAddress()).balanceOf(address(this));
        uint256 balanceBeforeEth = address(this).balance;
        _;
        if (balanceBeforeToken != IERC20Transferable(_aTokenAddress()).balanceOf(address(this))) {
            revert ValidatorPoolErrors.TokenBalanceChangedDuringOperation();
        }
        if (balanceBeforeEth != address(this).balance) {
            revert ValidatorPoolErrors.EthBalanceChangedDuringOperation();
        }
    }

    constructor() ValidatorPoolStorage() {}

    receive() external payable {
        if (msg.sender != _validatorStakingAddress() && msg.sender != _publicStakingAddress()) {
            revert ValidatorPoolErrors.OnlyStakingContractsAllowed();
        }
    }

    function initialize(
        uint256 stakeAmount_,
        uint256 maxNumValidators_,
        uint256 disputerReward_,
        uint256 maxIntervalWithoutSnapshots
    ) public onlyFactory initializer {
        _stakeAmount = stakeAmount_;
        _maxNumValidators = maxNumValidators_;
        _disputerReward = disputerReward_;
        _maxIntervalWithoutSnapshots = maxIntervalWithoutSnapshots;
    }

    function setStakeAmount(uint256 stakeAmount_) public onlyFactory {
        _stakeAmount = stakeAmount_;
    }

    function setMaxIntervalWithoutSnapshots(uint256 maxIntervalWithoutSnapshots)
        public
        onlyFactory
    {
        if (maxIntervalWithoutSnapshots == 0) {
            revert ValidatorPoolErrors.MaxIntervalWithoutSnapshotsMustBeNonZero();
        }
        _maxIntervalWithoutSnapshots = maxIntervalWithoutSnapshots;
    }

    function setMaxNumValidators(uint256 maxNumValidators_) public onlyFactory {
        if (maxNumValidators_ < _validators.length()) {
            revert ValidatorPoolErrors.MaxNumValidatorsIsTooLow(
                maxNumValidators_,
                _validators.length()
            );
        }
        _maxNumValidators = maxNumValidators_;
    }

    function setDisputerReward(uint256 disputerReward_) public onlyFactory {
        _disputerReward = disputerReward_;
    }

    function setLocation(string calldata ip_) public onlyValidator {
        _ipLocations[msg.sender] = ip_;
    }

    function scheduleMaintenance() public onlyFactory {
        _isMaintenanceScheduled = true;
        emit MaintenanceScheduled();
    }

    function initializeETHDKG()
        public
        onlyFactory
        assertNotETHDKGRunning
        assertNotConsensusRunning
    {
        IETHDKG(_ethdkgAddress()).initializeETHDKG();
    }

    function completeETHDKG() public onlyETHDKG {
        _isMaintenanceScheduled = false;
        _isConsensusRunning = true;
    }

    // todo: check async in AliceNet
    function pauseConsensus() public onlySnapshots {
        _isConsensusRunning = false;
    }

    function pauseConsensusOnArbitraryHeight(uint256 aliceNetHeight_) public onlyFactory {
        uint256 targetBlockNumber = ISnapshots(_snapshotsAddress())
            .getCommittedHeightFromLatestSnapshot() + _maxIntervalWithoutSnapshots;
        if (block.number <= targetBlockNumber) {
            revert ValidatorPoolErrors.MinimumBlockIntervalNotMet(block.number, targetBlockNumber);
        }
        _isConsensusRunning = false;
        IETHDKG(_ethdkgAddress()).setCustomAliceNetHeight(aliceNetHeight_);
    }

    function registerValidators(address[] memory validators_, uint256[] memory stakerTokenIDs_)
        public
        onlyFactory
        assertNotETHDKGRunning
        assertNotConsensusRunning
    {
        if (validators_.length + _validators.length() > _maxNumValidators) {
            revert ValidatorPoolErrors.NotEnoughValidatorSlotsAvailable(
                validators_.length,
                _maxNumValidators - _validators.length()
            );
        }
        if (validators_.length != stakerTokenIDs_.length) {
            revert ValidatorPoolErrors.RegistrationParameterLengthMismatch(
                validators_.length,
                stakerTokenIDs_.length
            );
        }

        for (uint256 i = 0; i < validators_.length; i++) {
            if (msg.sender != IERC721(_publicStakingAddress()).ownerOf(stakerTokenIDs_[i])) {
                revert ValidatorPoolErrors.SenderShouldOwnPosition(stakerTokenIDs_[i]);
            }
            _registerValidator(validators_[i], stakerTokenIDs_[i]);
        }
    }

    function unregisterValidators(address[] memory validators_)
        public
        onlyFactory
        assertNotETHDKGRunning
        assertNotConsensusRunning
    {
        if (validators_.length > _validators.length()) {
            revert ValidatorPoolErrors.LengthGreaterThanAvailableValidators(
                validators_.length,
                _validators.length()
            );
        }
        for (uint256 i = 0; i < validators_.length; i++) {
            _unregisterValidator(validators_[i]);
        }
    }

    function unregisterAllValidators()
        public
        onlyFactory
        assertNotETHDKGRunning
        assertNotConsensusRunning
    {
        while (_validators.length() > 0) {
            address validator = _validators.at(_validators.length() - 1)._address;
            _unregisterValidator(validator);
        }
    }

    function collectProfits()
        public
        onlyValidator
        balanceShouldNotChange
        returns (uint256 payoutEth, uint256 payoutToken)
    {
        if (!_isConsensusRunning) {
            revert ValidatorPoolErrors.ProfitsOnlyClaimableWhileConsensusRunning();
        }

        uint256 validatorTokenID = _validators.get(msg.sender)._tokenID;
        payoutEth = IStakingNFT(_validatorStakingAddress()).collectEthTo(
            msg.sender,
            validatorTokenID
        );
        payoutToken = IStakingNFT(_validatorStakingAddress()).collectTokenTo(
            msg.sender,
            validatorTokenID
        );

        return (payoutEth, payoutToken);
    }

    function claimExitingNFTPosition() public returns (uint256) {
        ExitingValidatorData memory data = _exitingValidatorsData[msg.sender];
        if (data._freeAfter == 0) {
            revert ValidatorPoolErrors.SenderNotInExitingQueue(msg.sender);
        }
        if (ISnapshots(_snapshotsAddress()).getEpoch() <= data._freeAfter) {
            revert ValidatorPoolErrors.WaitingPeriodNotMet();
        }

        _removeExitingQueueData(msg.sender);

        IStakingNFT(_publicStakingAddress()).lockOwnPosition(data._tokenID, POSITION_LOCK_PERIOD);

        IERC721Transferable(_publicStakingAddress()).safeTransferFrom(
            address(this),
            msg.sender,
            data._tokenID
        );

        return data._tokenID;
    }

    function majorSlash(address dishonestValidator_, address disputer_)
        public
        onlyETHDKG
        balanceShouldNotChange
    {
        if (!_isAccusable(dishonestValidator_)) {
            revert ValidatorPoolErrors.AddressNotAccusable(dishonestValidator_);
        }

        (uint256 minerShares, uint256 payoutEth, uint256 payoutToken) = _slash(dishonestValidator_);
        // deciding which state to clean based if the accusable person was a active validator or was
        // in the exiting line
        if (isValidator(dishonestValidator_)) {
            _removeValidatorData(dishonestValidator_);
        } else {
            _removeExitingQueueData(dishonestValidator_);
        }
        // redistribute the dishonest staking equally with the other validators

        IERC20Transferable(_aTokenAddress()).approve(_validatorStakingAddress(), minerShares);
        IStakingNFT(_validatorStakingAddress()).depositToken(_getMagic(), minerShares);
        // transfer to the disputer any profit that the dishonestValidator had when his
        // position was burned + the disputerReward
        _transferEthAndTokens(disputer_, payoutEth, payoutToken);

        emit ValidatorMajorSlashed(dishonestValidator_);
    }

    function minorSlash(address dishonestValidator_, address disputer_)
        public
        onlyETHDKG
        balanceShouldNotChange
    {
        if (!_isAccusable(dishonestValidator_)) {
            revert ValidatorPoolErrors.AddressNotAccusable(dishonestValidator_);
        }
        (uint256 minerShares, uint256 payoutEth, uint256 payoutToken) = _slash(dishonestValidator_);
        uint256 stakeTokenID;
        // In case there's not enough shares to create a new PublicStaking position, state is just
        // cleaned and the rest of the funds is sent to the disputer
        if (minerShares > 0) {
            stakeTokenID = _mintPublicStakingPosition(minerShares);
            _moveToExitingQueue(dishonestValidator_, stakeTokenID);
        } else {
            if (isValidator(dishonestValidator_)) {
                _removeValidatorData(dishonestValidator_);
            } else {
                _removeExitingQueueData(dishonestValidator_);
            }
        }
        _transferEthAndTokens(disputer_, payoutEth, payoutToken);
        emit ValidatorMinorSlashed(dishonestValidator_, stakeTokenID);
    }

    /// skimExcessEth will allow the Admin role to refund any Eth sent to this contract in error by a
    /// user. This function should only be necessary if a user somehow manages to accidentally
    /// selfDestruct a contract with this contract as the recipient or use the PublicStaking burnTo with the
    /// address of this contract.
    function skimExcessEth(address to_) public onlyFactory returns (uint256 excess) {
        // This contract shouldn't held any eth balance.
        // todo: revisit this when we have the dutch auction
        excess = address(this).balance;
        _safeTransferEth(to_, excess);
        return excess;
    }

    /// skimExcessToken will allow the Admin role to refund any AToken sent to this contract in error
    /// by a user.
    function skimExcessToken(address to_) public onlyFactory returns (uint256 excess) {
        // This contract shouldn't held any token balance.
        IERC20Transferable aToken = IERC20Transferable(_aTokenAddress());
        excess = aToken.balanceOf(address(this));
        _safeTransferERC20(aToken, to_, excess);
        return excess;
    }

    function getStakeAmount() public view returns (uint256) {
        return _stakeAmount;
    }

    function getMaxIntervalWithoutSnapshots()
        public
        view
        returns (uint256 maxIntervalWithoutSnapshots)
    {
        return _maxIntervalWithoutSnapshots;
    }

    function getMaxNumValidators() public view returns (uint256) {
        return _maxNumValidators;
    }

    function getDisputerReward() public view returns (uint256) {
        return _disputerReward;
    }

    function getValidatorsCount() public view returns (uint256) {
        return _validators.length();
    }

    function getValidatorsAddresses() public view returns (address[] memory) {
        return _validators.addressValues();
    }

    function getValidator(uint256 index_) public view returns (address) {
        if (index_ >= _validators.length()) {
            revert ValidatorPoolErrors.InvalidIndex(index_);
        }
        return _validators.at(index_)._address;
    }

    function getValidatorData(uint256 index_) public view returns (ValidatorData memory) {
        if (index_ >= _validators.length()) {
            revert ValidatorPoolErrors.InvalidIndex(index_);
        }
        return _validators.at(index_);
    }

    function getLocation(address validator_) public view returns (string memory) {
        return _ipLocations[validator_];
    }

    function getLocations(address[] calldata validators_) public view returns (string[] memory) {
        string[] memory ret = new string[](validators_.length);
        for (uint256 i = 0; i < validators_.length; i++) {
            ret[i] = _ipLocations[validators_[i]];
        }
        return ret;
    }

    /// @notice Try to get the NFT tokenID for an account.
    /// @param account_ address of the account to try to retrieve the tokenID
    /// @return tuple (bool, address, uint256). Return true if the value was found, false if not.
    /// Returns the address of the NFT contract and the tokenID. In case the value was not found, tokenID
    /// and address are 0.
    function tryGetTokenID(address account_)
        public
        view
        returns (
            bool,
            address,
            uint256
        )
    {
        if (_isValidator(account_)) {
            return (true, _validatorStakingAddress(), _validators.get(account_)._tokenID);
        } else if (_isInExitingQueue(account_)) {
            return (true, _publicStakingAddress(), _exitingValidatorsData[account_]._tokenID);
        } else {
            return (false, address(0), 0);
        }
    }

    function isValidator(address account_) public view returns (bool) {
        return _isValidator(account_);
    }

    function isInExitingQueue(address account_) public view returns (bool) {
        return _isInExitingQueue(account_);
    }

    function isAccusable(address account_) public view returns (bool) {
        return _isAccusable(account_);
    }

    function isMaintenanceScheduled() public view returns (bool) {
        return _isMaintenanceScheduled;
    }

    function isConsensusRunning() public view returns (bool) {
        return _isConsensusRunning;
    }

    function _transferEthAndTokens(
        address to_,
        uint256 payoutEth_,
        uint256 payoutToken_
    ) internal {
        _safeTransferERC20(IERC20Transferable(_aTokenAddress()), to_, payoutToken_);
        _safeTransferEth(to_, payoutEth_);
    }

    function _registerValidator(address validator_, uint256 stakerTokenID_)
        internal
        balanceShouldNotChange
        returns (
            uint256 validatorTokenID,
            uint256 payoutEth,
            uint256 payoutToken
        )
    {
        if (_validators.length() >= _maxNumValidators) {
            revert ValidatorPoolErrors.NotEnoughValidatorSlotsAvailable(1, 0);
        }
        if (_isAccusable(validator_)) {
            revert ValidatorPoolErrors.AddressAlreadyValidator(validator_);
        }

        (validatorTokenID, payoutEth, payoutToken) = _swapPublicStakingForValidatorStaking(
            msg.sender,
            stakerTokenID_
        );

        _validators.add(ValidatorData(validator_, validatorTokenID));
        // transfer back any profit that was available for the PublicStaking position by the time that we
        // burned it
        _transferEthAndTokens(validator_, payoutEth, payoutToken);
        emit ValidatorJoined(validator_, validatorTokenID);
    }

    function _unregisterValidator(address validator_)
        internal
        balanceShouldNotChange
        returns (
            uint256 stakeTokenID,
            uint256 payoutEth,
            uint256 payoutToken
        )
    {
        if (!_isValidator(validator_)) {
            revert ValidatorPoolErrors.AddressNotValidator(validator_);
        }
        (stakeTokenID, payoutEth, payoutToken) = _swapValidatorStakingForPublicStaking(validator_);

        _moveToExitingQueue(validator_, stakeTokenID);

        // transfer back any profit that was available for the PublicStaking position by the time that we
        // burned it
        _transferEthAndTokens(validator_, payoutEth, payoutToken);
        emit ValidatorLeft(validator_, stakeTokenID);
    }

    function _swapPublicStakingForValidatorStaking(address to_, uint256 stakerTokenID_)
        internal
        returns (
            uint256 validatorTokenID,
            uint256 payoutEth,
            uint256 payoutToken
        )
    {
        (uint256 stakeShares, , , , ) = IStakingNFT(_publicStakingAddress()).getPosition(
            stakerTokenID_
        );
        uint256 stakeAmount = _stakeAmount;
        if (stakeShares < stakeAmount) {
            revert ValidatorPoolErrors.InsufficientFundsInStakePosition(stakeShares, stakeAmount);
        }
        IERC721Transferable(_publicStakingAddress()).safeTransferFrom(
            to_,
            address(this),
            stakerTokenID_
        );
        (payoutEth, payoutToken) = IStakingNFT(_publicStakingAddress()).burn(stakerTokenID_);

        // Subtracting the shares from PublicStaking profit. The shares will be used to mint the new
        // ValidatorPosition
        //payoutToken should always have the minerShares in it!
        if (payoutToken < stakeShares) {
            revert ValidatorPoolErrors.PayoutTooLow();
        }
        payoutToken -= stakeAmount;

        validatorTokenID = _mintValidatorStakingPosition(stakeAmount);

        return (validatorTokenID, payoutEth, payoutToken);
    }

    function _swapValidatorStakingForPublicStaking(address validator_)
        internal
        returns (
            uint256,
            uint256,
            uint256
        )
    {
        (
            uint256 minerShares,
            uint256 payoutEth,
            uint256 payoutToken
        ) = _burnValidatorStakingPosition(validator_);
        //payoutToken should always have the minerShares in it!
        if (payoutToken < minerShares) {
            revert ValidatorPoolErrors.PayoutTooLow();
        }
        payoutToken -= minerShares;

        uint256 stakeTokenID = _mintPublicStakingPosition(minerShares);

        return (stakeTokenID, payoutEth, payoutToken);
    }

    function _mintValidatorStakingPosition(uint256 minerShares_)
        internal
        returns (uint256 validatorTokenID)
    {
        // We should approve the validatorStaking to transferFrom the tokens of this contract
        IERC20Transferable(_aTokenAddress()).approve(_validatorStakingAddress(), minerShares_);
        validatorTokenID = IStakingNFT(_validatorStakingAddress()).mint(minerShares_);
    }

    function _mintPublicStakingPosition(uint256 minerShares_)
        internal
        returns (uint256 stakeTokenID)
    {
        // We should approve the PublicStaking to transferFrom the tokens of this contract
        IERC20Transferable(_aTokenAddress()).approve(_publicStakingAddress(), minerShares_);
        stakeTokenID = IStakingNFT(_publicStakingAddress()).mint(minerShares_);
    }

    function _burnValidatorStakingPosition(address validator_)
        internal
        returns (
            uint256 minerShares,
            uint256 payoutEth,
            uint256 payoutToken
        )
    {
        uint256 validatorTokenID = _validators.get(validator_)._tokenID;
        (minerShares, payoutEth, payoutToken) = _burnNFTPosition(
            validatorTokenID,
            _validatorStakingAddress()
        );
    }

    function _burnExitingPublicStakingPosition(address validator_)
        internal
        returns (
            uint256 minerShares,
            uint256 payoutEth,
            uint256 payoutToken
        )
    {
        uint256 stakerTokenID = _exitingValidatorsData[validator_]._tokenID;
        (minerShares, payoutEth, payoutToken) = _burnNFTPosition(
            stakerTokenID,
            _publicStakingAddress()
        );
    }

    function _burnNFTPosition(uint256 tokenID_, address stakeContractAddress_)
        internal
        returns (
            uint256 minerShares,
            uint256 payoutEth,
            uint256 payoutToken
        )
    {
        IStakingNFT stakeContract = IStakingNFT(stakeContractAddress_);
        (minerShares, , , , ) = stakeContract.getPosition(tokenID_);
        (payoutEth, payoutToken) = stakeContract.burn(tokenID_);
    }

    function _slash(address dishonestValidator_)
        internal
        returns (
            uint256 minerShares,
            uint256 payoutEth,
            uint256 payoutToken
        )
    {
        if (!_isAccusable(dishonestValidator_)) {
            revert ValidatorPoolErrors.AddressNotAccusable(dishonestValidator_);
        }
        // If the user accused is a valid validator, we should burn is validatorStaking position,
        // otherwise we burn the user's PublicStaking in the exiting line
        if (_isValidator(dishonestValidator_)) {
            (minerShares, payoutEth, payoutToken) = _burnValidatorStakingPosition(
                dishonestValidator_
            );
        } else {
            (minerShares, payoutEth, payoutToken) = _burnExitingPublicStakingPosition(
                dishonestValidator_
            );
        }
        uint256 disputerReward = _disputerReward;
        if (minerShares >= disputerReward) {
            minerShares -= disputerReward;
        } else {
            // In case there's not enough shares to cover the _disputerReward, minerShares is set to
            // 0 and the rest of the payout Token is sent to disputer
            minerShares = 0;
        }
        //payoutToken should always have the minerShares in it!
        if (payoutToken < minerShares) {
            revert ValidatorPoolErrors.PayoutTooLow();
        }
        payoutToken -= minerShares;
    }

    function _moveToExitingQueue(address validator_, uint256 stakeTokenID_) internal {
        if (_isValidator(validator_)) {
            _removeValidatorData(validator_);
        }
        _exitingValidatorsData[validator_] = ExitingValidatorData(
            uint128(stakeTokenID_),
            uint128(ISnapshots(_snapshotsAddress()).getEpoch() + CLAIM_PERIOD)
        );
    }

    function _removeValidatorData(address validator_) internal {
        _validators.remove(validator_);
        delete _ipLocations[validator_];
    }

    function _removeExitingQueueData(address validator_) internal {
        delete _exitingValidatorsData[validator_];
    }

    function _isValidator(address account_) internal view returns (bool) {
        return _validators.contains(account_);
    }

    function _isInExitingQueue(address account_) internal view returns (bool) {
        return _exitingValidatorsData[account_]._freeAfter > 0;
    }

    function _isAccusable(address account_) internal view returns (bool) {
        return _isValidator(account_) || _isInExitingQueue(account_);
    }
}
