// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/libraries/math/CryptoLibrary.sol";
import "contracts/interfaces/IETHDKG.sol";
import "contracts/interfaces/IETHDKGEvents.sol";
import "contracts/libraries/ethdkg/ETHDKGStorage.sol";
import "contracts/utils/ETHDKGUtils.sol";

///@custom:salt ETHDKGPhases
///@custom:deploy-type deployUpgradeable
contract ETHDKGPhases is ETHDKGStorage, IETHDKGEvents, ETHDKGUtils {
    constructor() ETHDKGStorage(){
    }
    function getMyAddress() public view returns(address) {
        return address(this);
    }

    function register(uint256[2] memory publicKey) external {
        require(
            _ethdkgPhase == Phase.RegistrationOpen &&
                block.number >= _phaseStartBlock &&
                block.number < _phaseStartBlock + _phaseLength,
            "ETHDKG: Cannot register at the moment"
        );
        require(
            publicKey[0] != 0 && publicKey[1] != 0,
            "ETHDKG: Registration failed - pubKey should be different from 0!"
        );

        require(
            CryptoLibrary.bn128_is_on_curve(publicKey),
            "ETHDKG: Registration failed - public key not on elliptic curve!"
        );
        require(
            _participants[msg.sender].nonce < _nonce,
            "ETHDKG: Participant is already participating in this ETHDKG round!"
        );
        uint32 numRegistered = uint32(_numParticipants);
        numRegistered++;
        _participants[msg.sender] = Participant({
            publicKey: publicKey,
            index: numRegistered,
            nonce: _nonce,
            phase: _ethdkgPhase,
            distributedSharesHash: 0x0,
            commitmentsFirstCoefficient: [uint256(0), uint256(0)],
            keyShares: [uint256(0), uint256(0)],
            gpkj: [uint256(0), uint256(0), uint256(0), uint256(0)]
        });

        emit AddressRegistered(msg.sender, numRegistered, _nonce, publicKey);
        if (
            _moveToNextPhase(
                Phase.ShareDistribution,
                IValidatorPool(_ValidatorPoolAddress()).getValidatorsCount(),
                numRegistered
            )
        ) {
            emit RegistrationComplete(block.number);
        }
    }

    function distributeShares(uint256[] memory encryptedShares, uint256[2][] memory commitments)
        external
    {
        require(
            _ethdkgPhase == Phase.ShareDistribution &&
                block.number >= _phaseStartBlock &&
                block.number < _phaseStartBlock + _phaseLength,
            "ETHDKG: cannot participate on this phase"
        );
        Participant memory participant = _participants[msg.sender];
        require(
            participant.nonce == _nonce,
            "ETHDKG: Share distribution failed, participant with invalid nonce!"
        );
        require(
            participant.phase == Phase.RegistrationOpen,
            "ETHDKG: Participant already distributed shares this ETHDKG round!"
        );

        uint256 numValidators = IValidatorPool(_ValidatorPoolAddress()).getValidatorsCount();
        uint256 threshold = _getThreshold(numValidators);
        require(
            encryptedShares.length == numValidators - 1,
            "ETHDKG: Share distribution failed - invalid number of encrypted shares provided!"
        );
        require(
            commitments.length == threshold + 1,
            "ETHDKG: Key sharing failed - invalid number of commitments provided!"
        );
        for (uint256 k = 0; k <= threshold; k++) {
            require(
                CryptoLibrary.bn128_is_on_curve(commitments[k]),
                "ETHDKG: Key sharing failed - commitment not on elliptic curve!"
            );
            require(commitments[k][0] != 0, "ETHDKG: Commitments shouldn't be 0!");
        }

        bytes32 encryptedSharesHash = keccak256(abi.encodePacked(encryptedShares));
        bytes32 commitmentsHash = keccak256(abi.encodePacked(commitments));
        participant.distributedSharesHash = keccak256(
            abi.encodePacked(encryptedSharesHash, commitmentsHash)
        );
        require(
            participant.distributedSharesHash != 0x0,
            "ETHDKG: The hash of encryptedShares and commitments should be different from 0!"
        );
        participant.commitmentsFirstCoefficient = commitments[0];
        participant.phase = Phase.ShareDistribution;

        _participants[msg.sender] = participant;
        uint256 numParticipants = _numParticipants + 1;

        emit SharesDistributed(
            msg.sender,
            participant.index,
            participant.nonce,
            encryptedShares,
            commitments
        );

        if (_moveToNextPhase(Phase.DisputeShareDistribution, numValidators, numParticipants)) {
            emit ShareDistributionComplete(block.number);
        }
    }

    function submitKeyShare(
        uint256[2] memory keyShareG1,
        uint256[2] memory keyShareG1CorrectnessProof,
        uint256[4] memory keyShareG2
    ) external {
        // Only progress if all participants distributed their shares
        // and no bad participant was found
        require(
            (_ethdkgPhase == Phase.KeyShareSubmission &&
                block.number >= _phaseStartBlock &&
                block.number < _phaseStartBlock + _phaseLength) ||
                (_ethdkgPhase == Phase.DisputeShareDistribution &&
                    block.number >= _phaseStartBlock + _phaseLength &&
                    block.number < _phaseStartBlock + 2 * _phaseLength &&
                    _badParticipants == 0),
            "ETHDKG: cannot participate on key share submission phase"
        );

        // Since we had a dispute stage prior this state we need to set global state in here
        if (_ethdkgPhase != Phase.KeyShareSubmission) {
            _setPhase(Phase.KeyShareSubmission);
        }
        Participant memory participant = _participants[msg.sender];
        require(
            participant.nonce == _nonce,
            "ETHDKG: Key share submission failed, participant with invalid nonce!"
        );
        require(
            participant.phase == Phase.ShareDistribution,
            "ETHDKG: Participant already submitted key shares this ETHDKG round!"
        );

        require(
            CryptoLibrary.dleq_verify(
                [CryptoLibrary.H1x, CryptoLibrary.H1y],
                keyShareG1,
                [CryptoLibrary.G1x, CryptoLibrary.G1y],
                participant.commitmentsFirstCoefficient,
                keyShareG1CorrectnessProof
            ),
            "ETHDKG: Key share submission failed - invalid key share G1!"
        );
        require(
            CryptoLibrary.bn128_check_pairing(
                [
                    keyShareG1[0],
                    keyShareG1[1],
                    CryptoLibrary.H2xi,
                    CryptoLibrary.H2x,
                    CryptoLibrary.H2yi,
                    CryptoLibrary.H2y,
                    CryptoLibrary.H1x,
                    CryptoLibrary.H1y,
                    keyShareG2[0],
                    keyShareG2[1],
                    keyShareG2[2],
                    keyShareG2[3]
                ]
            ),
            "ETHDKG: Key share submission failed - invalid key share G2!"
        );

        participant.keyShares = keyShareG1;
        participant.phase = Phase.KeyShareSubmission;
        _participants[msg.sender] = participant;

        uint256[2] memory mpkG1 = _mpkG1;
        _mpkG1 = CryptoLibrary.bn128_add(
            [mpkG1[0], mpkG1[1], participant.keyShares[0], participant.keyShares[1]]
        );

        uint256 numParticipants = _numParticipants + 1;
        emit KeyShareSubmitted(
            msg.sender,
            participant.index,
            participant.nonce,
            keyShareG1,
            keyShareG1CorrectnessProof,
            keyShareG2
        );

        if (
            _moveToNextPhase(
                Phase.MPKSubmission,
                IValidatorPool(_ValidatorPoolAddress()).getValidatorsCount(),
                numParticipants
            )
        ) {
            emit KeyShareSubmissionComplete(block.number);
        }
    }

    function submitMasterPublicKey(uint256[4] memory masterPublicKey_) external {
        require(
            _ethdkgPhase == Phase.MPKSubmission &&
                block.number >= _phaseStartBlock &&
                block.number < _phaseStartBlock + _phaseLength,
            "ETHDKG: cannot participate on master public key submission phase"
        );
        uint256[2] memory mpkG1 = _mpkG1;
        require(
            CryptoLibrary.bn128_check_pairing(
                [
                    mpkG1[0],
                    mpkG1[1],
                    CryptoLibrary.H2xi,
                    CryptoLibrary.H2x,
                    CryptoLibrary.H2yi,
                    CryptoLibrary.H2y,
                    CryptoLibrary.H1x,
                    CryptoLibrary.H1y,
                    masterPublicKey_[0],
                    masterPublicKey_[1],
                    masterPublicKey_[2],
                    masterPublicKey_[3]
                ]
            ),
            "ETHDKG: Master key submission pairing check failed!"
        );

        _masterPublicKey = masterPublicKey_;

        _setPhase(Phase.GPKJSubmission);
        emit MPKSet(block.number, _nonce, masterPublicKey_);
    }

    function submitGPKJ(uint256[4] memory gpkj) external {
        //todo: should we evict all validators if no one sent the master public key in time?
        require(
            _ethdkgPhase == Phase.GPKJSubmission &&
                block.number >= _phaseStartBlock &&
                block.number < _phaseStartBlock + _phaseLength,
            "ETHDKG: Not in GPKJ submission phase"
        );

        Participant memory participant = _participants[msg.sender];

        require(
            participant.nonce == _nonce,
            "ETHDKG: Key share submission failed, participant with invalid nonce!"
        );
        require(
            participant.phase == Phase.KeyShareSubmission,
            "ETHDKG: Participant already submitted GPKj this ETHDKG round!"
        );

        require(
            gpkj[0] != 0 || gpkj[1] != 0 || gpkj[2] != 0 || gpkj[3] != 0,
            "ETHDKG: GPKj cannot be all zeros!"
        );

        participant.gpkj = gpkj;
        participant.phase = Phase.GPKJSubmission;
        _participants[msg.sender] = participant;

        emit ValidatorMemberAdded(
            msg.sender,
            participant.index,
            participant.nonce,
            ISnapshots(_SnapshotsAddress()).getEpoch(),
            participant.gpkj[0],
            participant.gpkj[1],
            participant.gpkj[2],
            participant.gpkj[3]
        );

        uint256 numParticipants = _numParticipants + 1;
        if (
            _moveToNextPhase(
                Phase.DisputeGPKJSubmission,
                IValidatorPool(_ValidatorPoolAddress()).getValidatorsCount(),
                numParticipants
            )
        ) {
            emit GPKJSubmissionComplete(block.number);
        }
    }

    function complete() external {
        //todo: should we reward ppl here?
        require(
            (_ethdkgPhase == Phase.DisputeGPKJSubmission &&
                block.number >= _phaseStartBlock + _phaseLength) &&
                block.number < _phaseStartBlock + 2 * _phaseLength,
            "ETHDKG: should be in post-GPKJDispute phase!"
        );
        require(
            _badParticipants == 0,
            "ETHDKG: Not all requisites to complete this ETHDKG round were completed!"
        );

        // Since we had a dispute stage prior this state we need to set global state in here
        _setPhase(Phase.Completion);

        IValidatorPool(_ValidatorPoolAddress()).completeETHDKG();

        uint256 epoch = ISnapshots(_SnapshotsAddress()).getEpoch();
        uint256 ethHeight = ISnapshots(_SnapshotsAddress()).getCommittedHeightFromLatestSnapshot();
        uint256 madHeight;
        if (_customMadnetHeight == 0) {
            madHeight = ISnapshots(_SnapshotsAddress()).getMadnetHeightFromLatestSnapshot();
        } else {
            madHeight = _customMadnetHeight;
            _customMadnetHeight = 0;
        }
        emit ValidatorSetCompleted(
            uint8(IValidatorPool(_ValidatorPoolAddress()).getValidatorsCount()),
            _nonce,
            epoch,
            ethHeight,
            madHeight,
            _masterPublicKey[0],
            _masterPublicKey[1],
            _masterPublicKey[2],
            _masterPublicKey[3]
        );
    }

    function _setPhase(Phase phase_) internal {
        _ethdkgPhase = phase_;
        _phaseStartBlock = uint64(block.number);
        _numParticipants = 0;
    }

    function _moveToNextPhase(
        Phase phase_,
        uint256 numValidators_,
        uint256 numParticipants_
    ) internal returns (bool) {
        // if all validators have registered, we can proceed to the next phase
        if (numParticipants_ == numValidators_) {
            _setPhase(phase_);
            _phaseStartBlock += _confirmationLength;
            return true;
        } else {
            _numParticipants = uint32(numParticipants_);
            return false;
        }
    }

    function _isMasterPublicKeySet() internal view returns (bool) {
        return ((_masterPublicKey[0] != 0) ||
            (_masterPublicKey[1] != 0) ||
            (_masterPublicKey[2] != 0) ||
            (_masterPublicKey[3] != 0));
    }
}