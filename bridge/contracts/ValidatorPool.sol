// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/INFTStake.sol";
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
import "contracts/interfaces/INFTStake.sol";
import "@openzeppelin/contracts/token/ERC721/utils/ERC721Holder.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";

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

    constructor() ValidatorPoolStorage() {

    }

    modifier onlyValidator() {
        require(_isValidator(msg.sender), "ValidatorPool: Only validators allowed!");
        _;
    }

    modifier assertNotConsensusRunning() {
        require(!_isConsensusRunning, "ValidatorPool: Error Madnet Consensus should be halted!");
        _;
    }

    modifier assertNotETHDKGRunning() {
        require(
            !IETHDKG(_ETHDKGAddress()).isETHDKGRunning(),
            "ValidatorPool: There's an ETHDKG round running!"
        );
        _;
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
        require(index_ < _validators.length(), "Index out boundaries!");
        return _validators.at(index_)._address;
    }

    function getValidatorData(uint256 index_) public view returns (ValidatorData memory) {
        require(index_ < _validators.length(), "Index out boundaries!");
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
            return (true, _ValidatorNFTAddress(), _validators.get(account_)._tokenID);
        } else if (_isInExitingQueue(account_)) {
            return (true, _StakeNFTAddress(), _exitingValidatorsData[account_]._tokenID);
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
        IETHDKG(_ETHDKGAddress()).initializeETHDKG();
    }

    function completeETHDKG() public onlyETHDKG {
        _isMaintenanceScheduled = false;
        _isConsensusRunning = true;
    }

    // todo: check async in Madnet
    function pauseConsensus() public onlySnapshots {
        _isConsensusRunning = false;
    }

    function pauseConsensusOnArbitraryHeight(uint256 madnetHeight_) public onlyFactory {
        require(
            block.number >
                ISnapshots(_SnapshotsAddress()).getCommittedHeightFromLatestSnapshot() +
                    _maxIntervalWithoutSnapshot,
            "ValidatorPool: Condition not met to stop consensus!"
        );
        _isConsensusRunning = false;
        IETHDKG(_ETHDKGAddress()).setCustomMadnetHeight(madnetHeight_);
    }

    function registerValidators(address[] calldata validators_, uint256[] calldata stakerTokenIDs_)
        external
        onlyFactory
        assertNotETHDKGRunning
        assertNotConsensusRunning
    {
        require(
            validators_.length + _validators.length() <= _maxNumValidators,
            "ValidatorPool: There are not enough free spots for all new validators!"
        );
        require(
            validators_.length == stakerTokenIDs_.length,
            "ValidatorPool: Both input array should have same length!"
        );

        for (uint256 i = 0; i < validators_.length; i++) {
            require(
                msg.sender == IERC721(_StakeNFTAddress()).ownerOf(stakerTokenIDs_[i]),
                "ValidatorPool: The factory should be the owner of the StakeNFT position!"
            );
            _registerValidator(validators_[i], stakerTokenIDs_[i]);
        }
    }

    function unregisterValidators(address[] calldata validators_)
        external
        onlyFactory
        assertNotETHDKGRunning
        assertNotConsensusRunning
    {
        require(
            validators_.length <= _validators.length(),
            "ValidatorPool: There are not enough validators to be removed!"
        );
        for (uint256 i = 0; i < validators_.length; i++) {
            _unregisterValidator(validators_[i]);
        }
    }

    function unregisterAllValidators()
        external
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
        external
        onlyValidator
        returns (uint256 payoutEth, uint256 payoutToken)
    {
        require(
            _isConsensusRunning,
            "ValidatorPool: Profits can only be claimable when consensus is running!"
        );

        uint256 balanceBeforeToken = IERC20Transferable(_MadTokenAddress()).balanceOf(
            address(this)
        );
        uint256 balanceBeforeEth = address(this).balance;

        uint256 validatorTokenID = _validators.get(msg.sender)._tokenID;
        payoutEth = INFTStake(_ValidatorNFTAddress()).collectEthTo(msg.sender, validatorTokenID);
        payoutToken = INFTStake(_ValidatorNFTAddress()).collectTokenTo(
            msg.sender,
            validatorTokenID
        );

        require(
            balanceBeforeToken == IERC20Transferable(_MadTokenAddress()).balanceOf(address(this)),
            "ValidatorPool: Invalid transaction, token balance of the contract changed!"
        );
        require(
            balanceBeforeEth == address(this).balance,
            "ValidatorPool: Invalid transaction, eth balance of the contract changed!"
        );

        return (payoutEth, payoutToken);
    }

    function claimExitingNFTPosition() public returns (uint256) {
        ExitingValidatorData memory data = _exitingValidatorsData[msg.sender];
        require(data._freeAfter > 0, "ValidatorPool: Address not in the exitingQueue!");
        require(
            ISnapshots(_SnapshotsAddress()).getEpoch() > data._freeAfter,
            "ValidatorPool: The waiting period is not over yet!"
        );

        _removeExitingQueueData(msg.sender);

        INFTStake(_StakeNFTAddress()).lockOwnPosition(data._tokenID, _positionLockPeriod);

        IERC721Transferable(_StakeNFTAddress()).safeTransferFrom(
            address(this),
            msg.sender,
            data._tokenID
        );

        return data._tokenID;
    }

    function majorSlash(address dishonestValidator_, address disputer_) public onlyETHDKG {
        uint256 balanceBeforeToken = IERC20Transferable(_MadTokenAddress()).balanceOf(
            address(this)
        );
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

        IERC20Transferable(_MadTokenAddress()).approve(_ValidatorNFTAddress(), minerShares);
        INFTStake(_ValidatorNFTAddress()).depositToken(_getMagic(), minerShares);
        // transfer to the disputer any profit that the dishonestValidator had when his
        // position was burned + the disputerReward
        _transferEthAndTokens(disputer_, payoutEth, payoutToken);

        require(
            balanceBeforeToken == IERC20Transferable(_MadTokenAddress()).balanceOf(address(this)),
            "ValidatorPool: Invalid transaction, token balance of the contract changed!"
        );
        require(
            balanceBeforeEth == address(this).balance,
            "ValidatorPool: Invalid transaction, eth balance of the contract changed!"
        );

        emit ValidatorMajorSlashed(dishonestValidator_);
    }

    function minorSlash(address dishonestValidator_, address disputer_) public onlyETHDKG {
        uint256 balanceBeforeToken = IERC20Transferable(_MadTokenAddress()).balanceOf(
            address(this)
        );
        uint256 balanceBeforeEth = address(this).balance;

        (uint256 minerShares, uint256 payoutEth, uint256 payoutToken) = _slash(dishonestValidator_);
        uint256 stakeTokenID;
        // In case there's not enough shares to create a new stakeNFT position, state is just
        // cleaned and the rest of the funds is sent to the disputer
        if (minerShares > 0) {
            stakeTokenID = _mintStakeNFTPosition(minerShares);
            _moveToExitingQueue(dishonestValidator_, stakeTokenID);
        } else {
            _removeExitingQueueData(dishonestValidator_);
        }
        _transferEthAndTokens(disputer_, payoutEth, payoutToken);

        require(
            balanceBeforeToken == IERC20Transferable(_MadTokenAddress()).balanceOf(address(this)),
            "ValidatorPool: Invalid transaction, token balance of the contract changed!"
        );
        require(
            balanceBeforeEth == address(this).balance,
            "ValidatorPool: Invalid transaction, eth balance of the contract changed!"
        );

        emit ValidatorMinorSlashed(dishonestValidator_, stakeTokenID);
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

    function _transferEthAndTokens(
        address to_,
        uint256 payoutEth_,
        uint256 payoutToken_
    ) internal {
        _safeTransferERC20(IERC20Transferable(_MadTokenAddress()), to_, payoutToken_);
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
            "ValidatorPool: There are no free spots for new validators!"
        );
        require(
            !_isAccusable(validator_),
            "ValidatorPool: Address is already a validator or it is in the exiting line!"
        );

        uint256 balanceBeforeToken = IERC20Transferable(_MadTokenAddress()).balanceOf(
            address(this)
        );
        uint256 balanceBeforeEth = address(this).balance;
        (validatorTokenID, payoutEth, payoutToken) = _swapStakeNFTForValidatorNFT(
            msg.sender,
            stakerTokenID_
        );

        _validators.add(ValidatorData(validator_, validatorTokenID));
        // transfer back any profit that was available for the stakeNFT position by the time that we
        // burned it
        _transferEthAndTokens(validator_, payoutEth, payoutToken);
        require(
            balanceBeforeToken == IERC20Transferable(_MadTokenAddress()).balanceOf(address(this)),
            "ValidatorPool: Invalid transaction, token balance of the contract changed!"
        );
        require(
            balanceBeforeEth == address(this).balance,
            "ValidatorPool: Invalid transaction, eth balance of the contract changed!"
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
        require(_isValidator(validator_), "ValidatorPool: Address is not a validator_!");

        uint256 balanceBeforeToken = IERC20Transferable(_MadTokenAddress()).balanceOf(
            address(this)
        );
        uint256 balanceBeforeEth = address(this).balance;
        (stakeTokenID, payoutEth, payoutToken) = _swapValidatorNFTForStakeNFT(validator_);

        _moveToExitingQueue(validator_, stakeTokenID);

        // transfer back any profit that was available for the stakeNFT position by the time that we
        // burned it
        _transferEthAndTokens(validator_, payoutEth, payoutToken);
        require(
            balanceBeforeToken == IERC20Transferable(_MadTokenAddress()).balanceOf(address(this)),
            "ValidatorPool: Invalid transaction, token balance of the contract changed!"
        );
        require(
            balanceBeforeEth == address(this).balance,
            "ValidatorPool: Invalid transaction, eth balance of the contract changed!"
        );

        emit ValidatorLeft(validator_, stakeTokenID);
    }

    function _swapStakeNFTForValidatorNFT(address to_, uint256 stakerTokenID_)
        internal
        returns (
            uint256 validatorTokenID,
            uint256 payoutEth,
            uint256 payoutToken
        )
    {
        (uint256 stakeShares, , , , ) = INFTStake(_StakeNFTAddress()).getPosition(stakerTokenID_);
        uint256 stakeAmount = _stakeAmount;
        require(
            stakeShares >= stakeAmount,
            "ValidatorStakeNFT: Error, the Stake position doesn't have enough funds!"
        );
        IERC721Transferable(_StakeNFTAddress()).safeTransferFrom(
            to_,
            address(this),
            stakerTokenID_
        );
        (payoutEth, payoutToken) = INFTStake(_StakeNFTAddress()).burn(stakerTokenID_);

        // Subtracting the shares from StakeNFT profit. The shares will be used to mint the new
        // ValidatorPosition
        //payoutToken should always have the minerShares in it!
        require(
            payoutToken >= stakeShares,
            "ValidatorPool: Miner shares greater then the total payout in tokens!"
        );
        payoutToken -= stakeAmount;

        validatorTokenID = _mintValidatorNFTPosition(stakeAmount);

        return (validatorTokenID, payoutEth, payoutToken);
    }

    function _swapValidatorNFTForStakeNFT(address validator_)
        internal
        returns (
            uint256,
            uint256,
            uint256
        )
    {
        (uint256 minerShares, uint256 payoutEth, uint256 payoutToken) = _burnValidatorNFTPosition(
            validator_
        );
        //payoutToken should always have the minerShares in it!
        require(
            payoutToken >= minerShares,
            "ValidatorPool: Miner shares greater then the total payout in tokens!"
        );
        payoutToken -= minerShares;

        uint256 stakeTokenID = _mintStakeNFTPosition(minerShares);

        return (stakeTokenID, payoutEth, payoutToken);
    }

    function _mintValidatorNFTPosition(uint256 minerShares_)
        internal
        returns (uint256 validatorTokenID)
    {
        // We should approve the ValidatorNFT to transferFrom the tokens of this contract
        IERC20Transferable(_MadTokenAddress()).approve(_ValidatorNFTAddress(), minerShares_);
        validatorTokenID = INFTStake(_ValidatorNFTAddress()).mint(minerShares_);
    }

    function _mintStakeNFTPosition(uint256 minerShares_) internal returns (uint256 stakeTokenID) {
        // We should approve the StakeNFT to transferFrom the tokens of this contract
        IERC20Transferable(_MadTokenAddress()).approve(_StakeNFTAddress(), minerShares_);
        stakeTokenID = INFTStake(_StakeNFTAddress()).mint(minerShares_);
    }

    function _burnValidatorNFTPosition(address validator_)
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
            _ValidatorNFTAddress()
        );
    }

    function _burnExitingStakeNFTPosition(address validator_)
        internal
        returns (
            uint256 minerShares,
            uint256 payoutEth,
            uint256 payoutToken
        )
    {
        uint256 stakerTokenID = _exitingValidatorsData[validator_]._tokenID;
        (minerShares, payoutEth, payoutToken) = _burnNFTPosition(stakerTokenID, _StakeNFTAddress());
    }

    function _burnNFTPosition(uint256 tokenID_, address stakeContractAddress_)
        internal
        returns (
            uint256 minerShares,
            uint256 payoutEth,
            uint256 payoutToken
        )
    {
        INFTStake stakeContract = INFTStake(stakeContractAddress_);
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
        require(_isAccusable(dishonestValidator_), "ValidatorPool: Address is not accusable!");
        // If the user accused is a valid validator, we should burn is ValidatorNFT position,
        // otherwise we burn the user's stakeNFT in the exiting line
        if (_isValidator(dishonestValidator_)) {
            (minerShares, payoutEth, payoutToken) = _burnValidatorNFTPosition(dishonestValidator_);
        } else {
            (minerShares, payoutEth, payoutToken) = _burnExitingStakeNFTPosition(
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
            "ValidatorPool: Miner shares greater then the total payout in tokens!"
        );
        payoutToken -= minerShares;
    }

    function _moveToExitingQueue(address validator_, uint256 stakeTokenID_) internal {
        if (_isValidator(validator_)) {
            _removeValidatorData(validator_);
        }
        _exitingValidatorsData[validator_] = ExitingValidatorData(
            uint128(stakeTokenID_),
            uint128(ISnapshots(_SnapshotsAddress()).getEpoch() + _claimPeriod)
        );
    }

    function _removeValidatorData(address validator_) internal {
        _validators.remove(validator_);
        delete _ipLocations[validator_];
    }

    function _removeExitingQueueData(address validator_) internal {
        delete _exitingValidatorsData[validator_];
    }

    receive() external payable onlyValidatorNFT() {
    }
}
