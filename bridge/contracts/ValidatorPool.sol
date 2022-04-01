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
import {ValidatorPoolErrorCodes} from "contracts/libraries/errorCodes/ValidatorPoolErrorCodes.sol";
import "@openzeppelin/contracts/utils/Strings.sol";

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
    using Strings for uint16;

    modifier onlyValidator() {
        require(
            _isValidator(msg.sender),
            ValidatorPoolErrorCodes.VALIDATORPOOL_CALLER_NOT_VALIDATOR.toString()
        );
        _;
    }

    modifier assertNotConsensusRunning() {
        require(
            !_isConsensusRunning,
            ValidatorPoolErrorCodes.VALIDATORPOOL_CONSENSUS_RUNNING.toString()
        );
        _;
    }

    modifier assertNotETHDKGRunning() {
        require(
            !IETHDKG(_ethdkgAddress()).isETHDKGRunning(),
            ValidatorPoolErrorCodes.VALIDATORPOOL_ETHDKG_ROUND_RUNNING.toString()
        );
        _;
    }

    constructor() ValidatorPoolStorage() {}

    receive() external payable {
        require(
            msg.sender == _validatorStakingAddress() || msg.sender == _publicStakingAddress(),
            ValidatorPoolErrorCodes.VALIDATORPOOL_ONLY_CONTRACTS_ALLOWED.toString()
        );
    }

    function initialize(
        uint256 stakeAmount_,
        uint256 maxNumValidators_,
        uint256 disputerReward_
    ) public onlyFactory initializer {
        _stakeAmount = stakeAmount_;
        _maxNumValidators = maxNumValidators_;
        _disputerReward = disputerReward_;
    }

    function setStakeAmount(uint256 stakeAmount_) public onlyFactory {
        _stakeAmount = stakeAmount_;
    }

    function setMaxNumValidators(uint256 maxNumValidators_) public onlyFactory {
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
        require(
            block.number >
                ISnapshots(_snapshotsAddress()).getCommittedHeightFromLatestSnapshot() +
                    MAX_INTERVAL_WITHOUT_SNAPSHOTS,
            ValidatorPoolErrorCodes.VALIDATORPOOL_MIN_BLOCK_INTERVAL_NOT_MET.toString()
        );
        _isConsensusRunning = false;
        IETHDKG(_ethdkgAddress()).setCustomAliceNetHeight(aliceNetHeight_);
    }

    function registerValidators(address[] memory validators_, uint256[] memory stakerTokenIDs_)
        public
        onlyFactory
        assertNotETHDKGRunning
        assertNotConsensusRunning
    {
        require(
            validators_.length + _validators.length() <= _maxNumValidators,
            ValidatorPoolErrorCodes.VALIDATORPOOL_MAX_VALIDATORS_MET.toString()
        );
        require(
            validators_.length == stakerTokenIDs_.length,
            ValidatorPoolErrorCodes.VALIDATORPOOL_REGISTRATION_PARAMETER_LENGTH_MISMATCH.toString()
        );

        for (uint256 i = 0; i < validators_.length; i++) {
            require(
                msg.sender == IERC721(_publicStakingAddress()).ownerOf(stakerTokenIDs_[i]),
                ValidatorPoolErrorCodes.VALIDATORPOOL_FACTORY_SHOULD_OWN_POSITION.toString()
            );
            _registerValidator(validators_[i], stakerTokenIDs_[i]);
        }
    }

    function unregisterValidators(address[] memory validators_)
        public
        onlyFactory
        assertNotETHDKGRunning
        assertNotConsensusRunning
    {
        require(
            validators_.length <= _validators.length(),
            ValidatorPoolErrorCodes.VALIDATORPOOL_VALIDATORS_GREATER_THAN_AVAILABLE.toString()
        );
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
        returns (uint256 payoutEth, uint256 payoutToken)
    {
        require(
            _isConsensusRunning,
            ValidatorPoolErrorCodes.VALIDATORPOOL_PROFITS_ONLY_CLAIMABLE_DURING_CONSENSUS.toString()
        );

        uint256 balanceBeforeToken = IERC20Transferable(_aTokenAddress()).balanceOf(address(this));
        uint256 balanceBeforeEth = address(this).balance;

        uint256 validatorTokenID = _validators.get(msg.sender)._tokenID;
        payoutEth = IStakingNFT(_validatorStakingAddress()).collectEthTo(
            msg.sender,
            validatorTokenID
        );
        payoutToken = IStakingNFT(_validatorStakingAddress()).collectTokenTo(
            msg.sender,
            validatorTokenID
        );

        require(
            balanceBeforeToken == IERC20Transferable(_aTokenAddress()).balanceOf(address(this)),
            ValidatorPoolErrorCodes.VALIDATORPOOL_TOKEN_BALANCE_CHANGED.toString()
        );
        require(
            balanceBeforeEth == address(this).balance,
            ValidatorPoolErrorCodes.VALIDATORPOOL_ETH_BALANCE_CHANGED.toString()
        );

        return (payoutEth, payoutToken);
    }

    function claimExitingNFTPosition() public returns (uint256) {
        ExitingValidatorData memory data = _exitingValidatorsData[msg.sender];
        require(
            data._freeAfter > 0,
            ValidatorPoolErrorCodes.VALIDATORPOOL_SENDER_NOT_IN_EXITING_QUEUE.toString()
        );
        require(
            ISnapshots(_snapshotsAddress()).getEpoch() > data._freeAfter,
            ValidatorPoolErrorCodes.VALIDATORPOOL_WAITING_PERIOD_NOT_MET.toString()
        );

        _removeExitingQueueData(msg.sender);

        IStakingNFT(_publicStakingAddress()).lockOwnPosition(data._tokenID, POSITION_LOCK_PERIOD);

        IERC721Transferable(_publicStakingAddress()).safeTransferFrom(
            address(this),
            msg.sender,
            data._tokenID
        );

        return data._tokenID;
    }

    function majorSlash(address dishonestValidator_, address disputer_) public onlyETHDKG {
        require(
            _isAccusable(dishonestValidator_),
            ValidatorPoolErrorCodes.VALIDATORPOOL_DISHONEST_VALIDATOR_NOT_ACCUSABLE.toString()
        );
        uint256 balanceBeforeToken = IERC20Transferable(_aTokenAddress()).balanceOf(address(this));
        uint256 balanceBeforeEth = address(this).balance;

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

        require(
            balanceBeforeToken == IERC20Transferable(_aTokenAddress()).balanceOf(address(this)),
            ValidatorPoolErrorCodes.VALIDATORPOOL_TOKEN_BALANCE_CHANGED.toString()
        );
        require(
            balanceBeforeEth == address(this).balance,
            ValidatorPoolErrorCodes.VALIDATORPOOL_ETH_BALANCE_CHANGED.toString()
        );

        emit ValidatorMajorSlashed(dishonestValidator_);
    }

    function minorSlash(address dishonestValidator_, address disputer_) public onlyETHDKG {
        require(
            _isAccusable(dishonestValidator_),
            ValidatorPoolErrorCodes.VALIDATORPOOL_DISHONEST_VALIDATOR_NOT_ACCUSABLE.toString()
        );
        uint256 balanceBeforeToken = IERC20Transferable(_aTokenAddress()).balanceOf(address(this));
        uint256 balanceBeforeEth = address(this).balance;

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

        require(
            balanceBeforeToken == IERC20Transferable(_aTokenAddress()).balanceOf(address(this)),
            ValidatorPoolErrorCodes.VALIDATORPOOL_TOKEN_BALANCE_CHANGED.toString()
        );
        require(
            balanceBeforeEth == address(this).balance,
            ValidatorPoolErrorCodes.VALIDATORPOOL_ETH_BALANCE_CHANGED.toString()
        );

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
        require(
            index_ < _validators.length(),
            ValidatorPoolErrorCodes.VALIDATORPOOL_INVALID_INDEX.toString()
        );
        return _validators.at(index_)._address;
    }

    function getValidatorData(uint256 index_) public view returns (ValidatorData memory) {
        require(
            index_ < _validators.length(),
            ValidatorPoolErrorCodes.VALIDATORPOOL_INVALID_INDEX.toString()
        );
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
        returns (
            uint256 validatorTokenID,
            uint256 payoutEth,
            uint256 payoutToken
        )
    {
        require(
            _validators.length() <= _maxNumValidators,
            ValidatorPoolErrorCodes.VALIDATORPOOL_MAX_VALIDATORS_MET.toString()
        );
        require(
            !_isAccusable(validator_),
            ValidatorPoolErrorCodes.VALIDATORPOOL_ADDRESS_ALREADY_VALIDATOR.toString()
        );

        uint256 balanceBeforeToken = IERC20Transferable(_aTokenAddress()).balanceOf(address(this));
        uint256 balanceBeforeEth = address(this).balance;
        (validatorTokenID, payoutEth, payoutToken) = _swapPublicStakingForValidatorStaking(
            msg.sender,
            stakerTokenID_
        );

        _validators.add(ValidatorData(validator_, validatorTokenID));
        // transfer back any profit that was available for the PublicStaking position by the time that we
        // burned it
        _transferEthAndTokens(validator_, payoutEth, payoutToken);
        require(
            balanceBeforeToken == IERC20Transferable(_aTokenAddress()).balanceOf(address(this)),
            ValidatorPoolErrorCodes.VALIDATORPOOL_TOKEN_BALANCE_CHANGED.toString()
        );
        require(
            balanceBeforeEth == address(this).balance,
            ValidatorPoolErrorCodes.VALIDATORPOOL_ETH_BALANCE_CHANGED.toString()
        );

        emit ValidatorJoined(validator_, validatorTokenID);
    }

    function _unregisterValidator(address validator_)
        internal
        returns (
            uint256 stakeTokenID,
            uint256 payoutEth,
            uint256 payoutToken
        )
    {
        require(
            _isValidator(validator_),
            ValidatorPoolErrorCodes.VALIDATORPOOL_ADDRESS_NOT_VALIDATOR.toString()
        );

        uint256 balanceBeforeToken = IERC20Transferable(_aTokenAddress()).balanceOf(address(this));
        uint256 balanceBeforeEth = address(this).balance;
        (stakeTokenID, payoutEth, payoutToken) = _swapValidatorStakingForPublicStaking(validator_);

        _moveToExitingQueue(validator_, stakeTokenID);

        // transfer back any profit that was available for the PublicStaking position by the time that we
        // burned it
        _transferEthAndTokens(validator_, payoutEth, payoutToken);
        require(
            balanceBeforeToken == IERC20Transferable(_aTokenAddress()).balanceOf(address(this)),
            ValidatorPoolErrorCodes.VALIDATORPOOL_TOKEN_BALANCE_CHANGED.toString()
        );
        require(
            balanceBeforeEth == address(this).balance,
            ValidatorPoolErrorCodes.VALIDATORPOOL_ETH_BALANCE_CHANGED.toString()
        );

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
        require(
            stakeShares >= stakeAmount,
            ValidatorPoolErrorCodes.VALIDATORPOOL_INSUFFICIENT_FUNDS_IN_STAKE_POSITION.toString()
        );
        IERC721Transferable(_publicStakingAddress()).safeTransferFrom(
            to_,
            address(this),
            stakerTokenID_
        );
        (payoutEth, payoutToken) = IStakingNFT(_publicStakingAddress()).burn(stakerTokenID_);

        // Subtracting the shares from PublicStaking profit. The shares will be used to mint the new
        // ValidatorPosition
        //payoutToken should always have the minerShares in it!
        require(
            payoutToken >= stakeShares,
            ValidatorPoolErrorCodes.VALIDATORPOOL_PAYOUT_TOO_LOW.toString()
        );
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
        require(
            payoutToken >= minerShares,
            ValidatorPoolErrorCodes.VALIDATORPOOL_PAYOUT_TOO_LOW.toString()
        );
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
        require(
            _isAccusable(dishonestValidator_),
            ValidatorPoolErrorCodes.VALIDATORPOOL_ADDRESS_NOT_ACCUSABLE.toString()
        );
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
        require(
            payoutToken >= minerShares,
            ValidatorPoolErrorCodes.VALIDATORPOOL_PAYOUT_TOO_LOW.toString()
        );
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
