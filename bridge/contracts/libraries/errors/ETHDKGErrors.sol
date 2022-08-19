// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/libraries/ethdkg/ETHDKGStorage.sol";

library ETHDKGErrors {
    error OnlyValidatorsAllowed(address sender);
    error VariableNotSettableWhileETHDKGRunning();
    error MinimumValidatorsNotMet(uint256 currentValidatorsLength);
    error ETHDKGNotInPostRegistrationAccusationPhase(Phase currentPhase);
    error AccusedNotValidator(address accused);
    error AccusedParticipatingInRound(address accused);
    error NotInPostSharedDistributionPhase(Phase currentPhase);
    error AccusedNotParticipatingInRound(address accused);
    error AccusedDistributedSharesInRound(address accused);
    error AccusedHasCommitments(address accused);
    error ETHDKGNotInDisputePhase(Phase currentPhase);
    error DisputerNotParticipatingInRound(address disputer);
    error AccusedDidNotDistributeSharesInRound(address accused);
    error DisputerDidNotDistributeSharesInRound(address disputer);
    error SharesAndCommitmentsMismatch(bytes32 expected, bytes32 actual);
    error InvalidKeyOrProof();
    error ETHDKGNotInPostKeyshareSubmissionPhase(Phase currentPhase);
    error AccusedSubmittedSharesInRound(address accused);
    error ETHDKGNotInPostGPKJSubmissionPhase(Phase currentPhase);
    error AccusedDidNotParticipateInGPKJSubmission(address accused);
    error AccusedDistributedGPKJ(address accused);
    error AccusedDidNotSubmitGPKJInRound(address accused);
    error DisputerDidNotSubmitGPKJInRound(address disputer);
    error ArgumentsLengthDoesNotEqualNumberOfParticipants(
        uint256 validatorsLength,
        uint256 encryptedSharesHashLength,
        uint256 commitmentsLength,
        uint256 numParticipants
    );
    error InvalidCommitments(uint256 commitmentsLength, uint256 expectedCommitmentsLength);
    error InvalidOrDuplicatedParticipant(address participant);
    error InvalidSharesOrCommitments(bytes32 expectedHash, bytes32 actualHash);
    error ETHDKGNotInRegistrationPhase(Phase currentPhase);
    error PublicKeyZero();
    error PublicKeyNotOnCurve();
    error ParticipantParticipatingInRound(
        address participant,
        uint256 participantNonce,
        uint256 maxExpectedNonce
    );
    error ETHDKGNotInSharedDistributionPhase(Phase currentPhase);
    error InvalidNonce(uint256 participantNonce, uint256 nonce);
    error ParticipantDistributedSharesInRound(address participant);
    error InvalidEncryptedSharesAmount(uint256 sharesLength, uint256 expectedSharesLength);
    error InvalidCommitmentsAmount(uint256 commitmentsLength, uint256 expectedCommitmentsLength);
    error CommitmentNotOnCurve();
    error CommitmentZero();
    error DistributedShareHashZero();
    error ETHDKGNotInKeyshareSubmissionPhase(Phase currentPhase);
    error ParticipantSubmittedKeysharesInRound(address participant);
    error InvalidKeyshareG1();
    error InvalidKeyshareG2();
    error ETHDKGNotInMasterPublicKeySubmissionPhase(Phase currentPhase);
    error MasterPublicKeyPairingCheckFailure();
    error ETHDKGNotInGPKJSubmissionPhase(Phase currentPhase);
    error ParticipantSubmittedGPKJInRound(address participant);
    error GPKJZero();
    error ETHDKGNotInPostGPKJDisputePhase(Phase currentPhase);
    error ETHDKGRequisitesIncomplete();
    error MigrationRequiresZeroNonce(uint256 nonce);
    error MigrationInputDataMismatch(
        uint256 validatorsAccountsLength,
        uint256 validatorIndexesLength,
        uint256 validatorSharesLength
    );
    error ParticipantNotFoundInLastRound(address addr);
}
