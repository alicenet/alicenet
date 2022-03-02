// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IETHDKG.sol";
import "contracts/interfaces/ISnapshots.sol";
import "contracts/interfaces/IERC20Transferable.sol";
import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/INFTStake.sol";
import "contracts/utils/CustomEnumerableMaps.sol";
import "contracts/utils/DeterministicAddress.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

contract ValidatorPoolMock is
    Initializable,
    IValidatorPool,
    ImmutableFactory,
    ImmutableSnapshots,
    ImmutableETHDKG,
    ImmutableValidatorNFT,
    ImmutableMadToken
{
    using CustomEnumerableMaps for ValidatorDataMap;

    uint256 public constant _positionLockPeriod = 3; // Actual is 172800

    uint256 public constant _maxIntervalWithoutSnapshot = 0; // Actual is 8192

    uint256 internal _tokenIDCounter;
    //IETHDKG immutable internal _ethdkg;
    //ISnapshots immutable internal _snapshots;

    ValidatorDataMap internal _validators;

    address internal _admin;

    bool internal _isMaintenanceScheduled;
    bool internal _isConsensusRunning;

    uint256 internal _stakeAmount;

    // solhint-disable no-empty-blocks
    constructor()
        ImmutableFactory(msg.sender)
        ImmutableValidatorNFT()
        ImmutableSnapshots()
        ImmutableETHDKG()
        ImmutableMadToken()
    {}

    function initialize() public onlyFactory initializer {
        //20000*10**18 MadWei = 20k MadTokens
        _stakeAmount = 20000 * 10**18;
    }

    function initializeETHDKG() external {
        IETHDKG(_ETHDKGAddress()).initializeETHDKG();
    }

    modifier onlyAdmin() {
        require(msg.sender == _admin, "Validators: requires admin privileges");
        _;
    }

    function mintValidatorNFT() public returns (uint256 stakeID_) {
        IERC20Transferable(_MadTokenAddress()).transferFrom(
            msg.sender,
            address(this),
            _stakeAmount
        );
        IERC20Transferable(_MadTokenAddress()).approve(_ValidatorNFTAddress(), _stakeAmount);
        stakeID_ = INFTStake(_ValidatorNFTAddress()).mint(_stakeAmount);
    }

    function burnValidatorNFT(uint256 tokenID_)
        public
        returns (uint256 payoutEth, uint256 payoutMadToken)
    {
        return INFTStake(_ValidatorNFTAddress()).burn(tokenID_);
    }

    function mintToValidatorNFT(address to_) public returns (uint256 stakeID_) {
        IERC20Transferable(_MadTokenAddress()).transferFrom(
            msg.sender,
            address(this),
            _stakeAmount
        );
        IERC20Transferable(_MadTokenAddress()).approve(_ValidatorNFTAddress(), _stakeAmount);
        stakeID_ = INFTStake(_ValidatorNFTAddress()).mintTo(to_, _stakeAmount, 1);
    }

    function burnToValidatorNFT(uint256 tokenID_, address to_)
        public
        returns (uint256 payoutEth, uint256 payoutMadToken)
    {
        return INFTStake(_ValidatorNFTAddress()).burnTo(to_, tokenID_);
    }

    function isValidator(address participant) public view returns (bool) {
        return _isValidator(participant);
    }

    function _isValidator(address participant) internal view returns (bool) {
        return _validators.contains(participant);
    }

    function getStakeAmount() public view returns (uint256) {
        return _stakeAmount;
    }

    function getMaxNumValidators() public pure returns (uint256) {
        return 5;
    }

    function getDisputerReward() public pure returns (uint256) {
        return 1;
    }

    function registerValidators(address[] memory v) internal {
        for (uint256 idx; idx < v.length; idx++) {
            uint256 tokenID = _tokenIDCounter + 1;
            _validators.add(ValidatorData(v[idx], tokenID));
            _tokenIDCounter = tokenID;
        }
    }

    function getValidatorsCount() public view returns (uint256) {
        return _validators.length();
    }

    function getValidator(uint256 index_) public view returns (address) {
        require(index_ < _validators.length(), "Index out boundaries!");
        return _validators.at(index_)._address;
    }

    function getValidatorsAddresses() external view returns (address[] memory addresses) {
        return _validators.addressValues();
    }

    function minorSlash(address validator, address disputer) public {
        disputer; //no-op to suppress warning of not using disputer address
        _removeValidator(validator);
    }

    function majorSlash(address validator, address disputer) public {
        disputer; //no-op to suppress warning of not using disputer address
        _removeValidator(validator);
    }

    function unregisterValidators(address[] memory validators) public {
        for (uint256 idx; idx < validators.length; idx++) {
            _removeValidator(validators[idx]);
        }
    }

    function unregisterAllValidators() public {
        while (_validators.length() > 0) {
            _removeValidator(_validators.at(_validators.length() - 1)._address);
        }
    }

    function _removeValidator(address validator_) internal {
        _validators.remove(validator_);
    }

    function completeETHDKG() public {
        _isMaintenanceScheduled = false;
        _isConsensusRunning = true;
    }

    function pauseConsensus() public onlySnapshots {
        _isConsensusRunning = false;
    }

    function scheduleMaintenance() public {
        _isMaintenanceScheduled = true;
    }

    function isMaintenanceScheduled() public view returns (bool) {
        return _isMaintenanceScheduled;
    }

    function isConsensusRunning() public view returns (bool) {
        return _isConsensusRunning;
    }

    function setConsensusRunning(bool isRunning) public {
        _isConsensusRunning = isRunning;
    }

    function claimExitingNFTPosition() public pure returns (uint256) {
        return 0;
    }

    function tryGetTokenID(address account_)
        public
        pure
        returns (
            bool,
            address,
            uint256
        )
    {
        account_; //no-op to suppress compiling warnings
        return (false, address(0), 0);
    }

    function collectProfits() public pure returns (uint256 payoutEth, uint256 payoutToken) {
        return (0, 0);
    }

    function getLocations(address[] calldata validators_) external pure returns (string[] memory) {
        validators_;
        return new string[](1);
    }

    function getValidatorData(uint256 index) external view returns (ValidatorData memory) {
        return _validators.at(index);
    }

    function setStakeAmount(uint256 stakeAmount_) external {}

    function setSnapshot(address _address) external {}

    function setMaxNumValidators(uint256 maxNumValidators_) public {}

    function setLocation(string calldata ip) external {}

    function registerValidators(address[] calldata validators, uint256[] calldata stakerTokenIDs)
        public
    {
        stakerTokenIDs;
        registerValidators(validators);
    }

    function isInExitingQueue(address participant) external pure returns (bool) {
        participant;
        return false;
    }

    function isAccusable(address participant) external pure returns (bool) {
        participant;
        return false;
    }

    function getLocation(address validator) external pure returns (string memory) {
        validator;
        return "";
    }

    function setDisputerReward(uint256 disputerReward_) public {}

    function isMock() public pure returns (bool) {
        return true;
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
}
