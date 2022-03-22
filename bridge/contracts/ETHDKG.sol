// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/libraries/math/CryptoLibrary.sol";
import "contracts/utils/AtomicCounter.sol";
import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/IETHDKG.sol";
import "contracts/interfaces/IETHDKGEvents.sol";
import "contracts/libraries/ethdkg/ETHDKGStorage.sol";
import "contracts/utils/ETHDKGUtils.sol";
import "contracts/utils/ImmutableAuth.sol";

import "contracts/interfaces/IProxy.sol";

/// @custom:salt ETHDKG
/// @custom:deploy-type deployUpgradeable
/// @custom:deploy-group ethdkg
/// @custom:deploy-group-index 2
contract ETHDKG is
    ETHDKGStorage,
    IETHDKG,
    IETHDKGEvents,
    ETHDKGUtils,
    ImmutableETHDKGAccusations,
    ImmutableETHDKGPhases
{
    address internal immutable _ethdkgAccusations;
    address internal immutable _ethdkgPhases;

    modifier onlyValidator() {
        require(
            IValidatorPool(_validatorPoolAddress()).isValidator(msg.sender),
            "ETHDKG: Only validators allowed!"
        );
        _;
    }

    constructor() ETHDKGStorage() ImmutableETHDKGAccusations() ImmutableETHDKGPhases() {
        // bytes32("ETHDKGPhases") = 0x455448444b475068617365730000000000000000000000000000000000000000;
        address ethdkgPhases = IProxy(_ethdkgPhasesAddress()).getImplementationAddress();
        assembly {
            if iszero(extcodesize(ethdkgPhases)) {
                mstore(0x00, "ethdkgPhases size 0")
                revert(0x00, 0x20)
            }
        }
        _ethdkgPhases = ethdkgPhases;
        // bytes32("ETHDKGAccusations") = 0x455448444b4741636375736174696f6e73000000000000000000000000000000;
        address ethdkgAccusations = IProxy(_ethdkgAccusationsAddress()).getImplementationAddress();
        assembly {
            if iszero(extcodesize(ethdkgAccusations)) {
                mstore(0x00, "ethdkgAccusations size 0")
                revert(0x00, 0x20)
            }
        }
        _ethdkgAccusations = ethdkgAccusations;
    }

    function initialize(uint256 phaseLength_, uint256 confirmationLength_)
        public
        initializer
        onlyFactory
    {
        _phaseLength = uint16(phaseLength_);
        _confirmationLength = uint16(confirmationLength_);
    }

    function setPhaseLength(uint16 phaseLength_) public onlyFactory {
        require(
            !_isETHDKGRunning(),
            "ETHDKG: This variable cannot be set if an ETHDKG round is running!"
        );
        _phaseLength = phaseLength_;
    }

    function setConfirmationLength(uint16 confirmationLength_) public onlyFactory {
        require(
            !_isETHDKGRunning(),
            "ETHDKG: This variable cannot be set if an ETHDKG round is running!"
        );
        _confirmationLength = confirmationLength_;
    }

    function setCustomMadnetHeight(uint256 madnetHeight) public onlyValidatorPool {
        _customMadnetHeight = madnetHeight;
        emit ValidatorSetCompleted(
            0,
            _nonce,
            ISnapshots(_snapshotsAddress()).getEpoch(),
            ISnapshots(_snapshotsAddress()).getCommittedHeightFromLatestSnapshot(),
            madnetHeight,
            0x0,
            0x0,
            0x0,
            0x0
        );
    }

    function initializeETHDKG() public onlyValidatorPool {
        _initializeETHDKG();
    }

    function register(uint256[2] memory publicKey) public onlyValidator {
        _callPhaseContract(abi.encodeWithSignature("register(uint256[2])", publicKey));
    }

    function accuseParticipantNotRegistered(address[] memory dishonestAddresses) public {
        _callAccusationContract(
            abi.encodeWithSignature("accuseParticipantNotRegistered(address[])", dishonestAddresses)
        );
    }

    function distributeShares(uint256[] memory encryptedShares, uint256[2][] memory commitments)
        public
        onlyValidator
    {
        _callPhaseContract(
            abi.encodeWithSignature(
                "distributeShares(uint256[],uint256[2][])",
                encryptedShares,
                commitments
            )
        );
    }

    ///
    function accuseParticipantDidNotDistributeShares(address[] memory dishonestAddresses) public {
        _callAccusationContract(
            abi.encodeWithSignature(
                "accuseParticipantDidNotDistributeShares(address[])",
                dishonestAddresses
            )
        );
    }

    // Someone sent bad shares
    function accuseParticipantDistributedBadShares(
        address dishonestAddress,
        uint256[] memory encryptedShares,
        uint256[2][] memory commitments,
        uint256[2] memory sharedKey,
        uint256[2] memory sharedKeyCorrectnessProof
    ) public onlyValidator {
        _callAccusationContract(
            abi.encodeWithSignature(
                "accuseParticipantDistributedBadShares(address,uint256[],uint256[2][],uint256[2],uint256[2])",
                dishonestAddress,
                encryptedShares,
                commitments,
                sharedKey,
                sharedKeyCorrectnessProof
            )
        );
    }

    function submitKeyShare(
        uint256[2] memory keyShareG1,
        uint256[2] memory keyShareG1CorrectnessProof,
        uint256[4] memory keyShareG2
    ) public onlyValidator {
        _callPhaseContract(
            abi.encodeWithSignature(
                "submitKeyShare(uint256[2],uint256[2],uint256[4])",
                keyShareG1,
                keyShareG1CorrectnessProof,
                keyShareG2
            )
        );
    }

    function accuseParticipantDidNotSubmitKeyShares(address[] memory dishonestAddresses) public {
        _callAccusationContract(
            abi.encodeWithSignature(
                "accuseParticipantDidNotSubmitKeyShares(address[])",
                dishonestAddresses
            )
        );
    }

    function submitMasterPublicKey(uint256[4] memory masterPublicKey_) public {
        _callPhaseContract(
            abi.encodeWithSignature("submitMasterPublicKey(uint256[4])", masterPublicKey_)
        );
    }

    function submitGPKJ(uint256[4] memory gpkj) public onlyValidator {
        _callPhaseContract(abi.encodeWithSignature("submitGPKJ(uint256[4])", gpkj));
    }

    function accuseParticipantDidNotSubmitGPKJ(address[] memory dishonestAddresses) public {
        _callAccusationContract(
            abi.encodeWithSignature(
                "accuseParticipantDidNotSubmitGPKJ(address[])",
                dishonestAddresses
            )
        );
    }

    function accuseParticipantSubmittedBadGPKJ(
        address[] memory validators,
        bytes32[] memory encryptedSharesHash,
        uint256[2][][] memory commitments,
        address dishonestAddress
    ) public onlyValidator {
        _callAccusationContract(
            abi.encodeWithSignature(
                "accuseParticipantSubmittedBadGPKJ(address[],bytes32[],uint256[2][][],address)",
                validators,
                encryptedSharesHash,
                commitments,
                dishonestAddress
            )
        );
    }

    // Successful_Completion should be called at the completion of the DKG algorithm.
    function complete() public onlyValidator {
        _callPhaseContract(abi.encodeWithSignature("complete()"));
    }

    function isETHDKGRunning() public view returns (bool) {
        return _isETHDKGRunning();
    }

    function isETHDKGCompleted() public view returns (bool) {
        return _isETHDKGCompleted();
    }

    function isETHDKGHalted() public view returns (bool) {
        return _isETHDKGHalted();
    }

    function isMasterPublicKeySet() public view returns (bool) {
        return ((_masterPublicKey[0] != 0) ||
            (_masterPublicKey[1] != 0) ||
            (_masterPublicKey[2] != 0) ||
            (_masterPublicKey[3] != 0));
    }

    function getNonce() public view returns (uint256) {
        return _nonce;
    }

    function getPhaseStartBlock() public view returns (uint256) {
        return _phaseStartBlock;
    }

    function getPhaseLength() public view returns (uint256) {
        return _phaseLength;
    }

    function getConfirmationLength() public view returns (uint256) {
        return _confirmationLength;
    }

    function getETHDKGPhase() public view returns (Phase) {
        return _ethdkgPhase;
    }

    function getNumParticipants() public view returns (uint256) {
        return _numParticipants;
    }

    function getBadParticipants() public view returns (uint256) {
        return _badParticipants;
    }

    function getParticipantInternalState(address participant)
        public
        view
        returns (Participant memory)
    {
        return _participants[participant];
    }

    function getParticipantsInternalState(address[] calldata participantAddresses)
        public
        view
        returns (Participant[] memory)
    {
        Participant[] memory participants = new Participant[](participantAddresses.length);

        for (uint256 i = 0; i < participantAddresses.length; i++) {
            participants[i] = _participants[participantAddresses[i]];
        }

        return participants;
    }

    function tryGetParticipantIndex(address participant) public view returns (bool, uint256) {
        Participant memory participantData = _participants[participant];
        if (participantData.nonce == _nonce && _nonce != 0) {
            return (true, _participants[participant].index);
        }
        return (false, 0);
    }

    function getMasterPublicKey() public view returns (uint256[4] memory) {
        return _masterPublicKey;
    }

    function getMinValidators() public pure returns (uint256) {
        return _MIN_VALIDATORS;
    }

    function _callAccusationContract(bytes memory callData) internal returns (bytes memory) {
        (bool success, bytes memory returnData) = _ethdkgAccusations.delegatecall(callData);
        if (!success) {
            // solhint-disable no-inline-assembly
            assembly {
                let ptr := mload(0x40)
                let size := returndatasize()
                returndatacopy(ptr, 0, size)
                revert(ptr, size)
            }
        }
        return returnData;
    }

    function _callPhaseContract(bytes memory callData) internal returns (bytes memory) {
        (bool success, bytes memory returnData) = _ethdkgPhases.delegatecall(callData);
        if (!success) {
            // solhint-disable no-inline-assembly
            assembly {
                let ptr := mload(0x40)
                let size := returndatasize()
                returndatacopy(ptr, 0, size)
                revert(ptr, size)
            }
        }
        return returnData;
    }

    function _initializeETHDKG() internal {
        //todo: should we reward ppl here?
        uint256 numberValidators = IValidatorPool(_validatorPoolAddress()).getValidatorsCount();
        require(
            numberValidators >= _MIN_VALIDATORS,
            "ETHDKG: Minimum number of validators staked not met!"
        );

        _phaseStartBlock = uint64(block.number);
        _nonce++;
        _numParticipants = 0;
        _badParticipants = 0;
        _mpkG1 = [uint256(0), uint256(0)];
        _ethdkgPhase = Phase.RegistrationOpen;

        delete _masterPublicKey;

        emit RegistrationOpened(
            block.number,
            numberValidators,
            _nonce,
            _phaseLength,
            _confirmationLength
        );
    }

    function _isETHDKGCompleted() internal view returns (bool) {
        return _ethdkgPhase == Phase.Completion;
    }

    function _isETHDKGRunning() internal view returns (bool) {
        // Handling initial case
        if (_phaseStartBlock == 0) {
            return false;
        }
        return !_isETHDKGCompleted() && !_isETHDKGHalted();
    }

    // todo: generate truth table
    function _isETHDKGHalted() internal view returns (bool) {
        bool ethdkgFailedInDisputePhase = (_ethdkgPhase == Phase.DisputeShareDistribution ||
            _ethdkgPhase == Phase.DisputeGPKJSubmission) &&
            block.number >= _phaseStartBlock + _phaseLength &&
            _badParticipants != 0;
        bool ethdkgFailedInNormalPhase = block.number >= _phaseStartBlock + 2 * _phaseLength;
        return ethdkgFailedInNormalPhase || ethdkgFailedInDisputePhase;
    }
}
