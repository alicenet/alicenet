// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/libraries/ethdkg/ETHDKGStorage.sol";

library ETHDKGErrors {
    // ETHDKG errors
    error OnlyValidatorsAllowed(address sender); //"ETHDKG: Only validators allowed!"
    error VariableNotSettableWhileETHDKGRunning(); //"ETHDKG: This variable cannot be set if an ETHDKG round is running!"
    error MinimumValidatorsNotMet(uint256 currentValidatorsLength); //"ETHDKG: Minimum number of validators staked not met!"
    error ETHDKGNotInPostRegistrationAccusationPhase(Phase currentPhase); //"ETHDKG: should be in post-registration accusation phase!"
    error AccusedNotValidator(address accused); //"ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!"
    error AccusedParticipatingInRound(address accused); //"ETHDKG: Dispute failed! dishonestParticipant is participating in this ETHDKG round!"
    error NotInPostSharedDistributionPhase(Phase currentPhase); // "ETHDKG: should be in post-ShareDistribution accusation phase!"
    error AccusedNotParticipatingInRound(address accused); //"ETHDKG: Dispute failed! Dishonest Participant is not participating in this ETHDKG round!"
    error AccusedDistributedSharesInRound(address accused); //"ETHDKG: Dispute failed! Supposed dishonest participant distributed its share in this ETHDKG round!"
    error AccusedHasCommitments(address accused); //"ETHDKG: Dispute failed! It looks like the supposed dishonest participant had commitments! "
    error ETHDKGNotInDisputePhase(Phase currentPhase); // "ETHDKG: Dispute failed! Contract is not in dispute phase!"
    error DisputerNotParticipatingInRound(address disputer); // "ETHDKG: Dispute failed! Disputer is not participating in this ETHDKG round!"
    error AccusedDidNotDistributeSharesInRound(address accused); // "ETHDKG: Dispute failed! dishonestParticipant did not distribute shares!"
    error DisputerDidNotDistributeSharesInRound(address disputer); // "ETHDKG: Dispute failed! Disputer did not distribute shares!"
    error SharesAndCommitmentsMismatch(bytes32 expected, bytes32 actual); //  "ETHDKG: Dispute failed! Submitted commitments and encrypted shares don't match!"
    error InvalidKeyOrProof(); // "ETHDKG: Dispute failed! Invalid shared key or proof!"
    error ETHDKGNotInPostKeyshareSubmissionPhase(Phase currentPhase); // "ETHDKG: Dispute failed! Should be in post-KeyShareSubmission phase!"
    error AccusedSubmittedSharesInRound(address accused); // "ETHDKG: Dispute failed! Dishonest participant submitted its key shares in this ETHDKG round!"
    error ETHDKGNotInPostGPKJSubmissionPhase(Phase currentPhase); // "ETHDKG: Dispute Failed! Should be in post-GPKJSubmission phase!"
    error AccusedDidNotParticipateInGPKJSubmission(address accused); // "ETHDKG: Dispute failed! Dishonest participant did participate in this GPKj submission!"
    error AccusedDistributedGPKJ(address accused); //  "ETHDKG: Dispute failed! It looks like the dishonestParticipant distributed its GPKJ!"
    error AccusedDidNotSubmitGPKJInRound(address accused); // "ETHDKG: Dispute Failed! Dishonest participant didn't submit his GPKJ for this round!"
    error DisputerDidNotSubmitGPKJInRound(address disputer); // "ETHDKG: Dispute Failed! Disputer didn't submit his GPKJ for this round!"
    error ArgumentsLengthDoesNotEqualNumberOfParticipants(
        uint256 validatorsLength,
        uint256 encryptedSharesHashLength,
        uint256 commitmentsLength,
        uint256 numParticipants
    ); // "ETHDKG: Dispute Failed! Invalid submission of arguments!"

    error InvalidCommitments(uint256 commitmentsLength, uint256 expectedCommitmentsLength); // "ETHDKG: Dispute Failed! Invalid number of commitments provided!"
    error InvalidOrDuplicatedParticipant(address participant); // "ETHDKG: Dispute Failed! Invalid or duplicated participant address!"
    error InvalidSharesOrCommitments(bytes32 expectedHash, bytes32 actualHash); // "ETHDKG: Dispute Failed! Invalid shares or commitments!"
    error ETHDKGNotInRegistrationPhase(Phase currentPhase); //  "ETHDKG: Cannot register at the moment"
    error PublicKeyZero(); // "ETHDKG: Registration failed - pubKey should be different from 0!"
    error PublicKeyNotOnCurve(); // "ETHDKG: Registration failed - public key not on elliptic curve!"
    error ParticipantParticipatingInRound(address participant); // "ETHDKG: Participant is already participating in this ETHDKG round!"
    error ETHDKGNotInSharedDistributionPhase(Phase currentPhase); // "ETHDKG: cannot participate on this phase"
    error InvalidNonce(uint256 participantNonce, uint256 nonce); // "ETHDKG: Share distribution failed, participant with invalid nonce!"
    error ParticipantDistributedSharesInRound(address participant); // "ETHDKG: Participant already distributed shares this ETHDKG round!"
    error InvalidEncryptedSharesAmount(uint256 sharesLength, uint256 expectedSharesLength); // "ETHDKG: Share distribution failed - invalid number of encrypted shares provided!"
    error InvalidCommitmentsAmount(uint256 commitmentsLength, uint256 expectedCommitmentsLength); // "ETHDKG: Key sharing failed - invalid number of commitments provided!"
    error CommitmentNotOnCurve(); // "ETHDKG: Key sharing failed - commitment not on elliptic curve!"
    error CommitmentZero(); // "ETHDKG: Key sharing failed - commitment not on elliptic curve!"
    error DistributedShareHashZero(); // "ETHDKG: The hash of encryptedShares and commitments should be different from 0!"
    error ETHDKGNotInKeyshareSubmissionPhase(Phase currentPhase); // "ETHDKG: cannot participate on key share submission phase"
    error ParticipantSubmittedKeysharesInRound(address participant); // "ETHDKG: Participant already submitted key shares this ETHDKG round!"
    error InvalidKeyshareG1(); //"ETHDKG: Key share submission failed - invalid key share G1!"
    error InvalidKeyshareG2(); //"ETHDKG: Key share submission failed - invalid key share G1!"

    bytes32 public constant ETHDKG_NOT_IN_MASTER_PUBLIC_KEY_SUBMISSION_PHASE = "143"; // "ETHDKG: cannot participate on master public key submission phase"
    bytes32 public constant ETHDKG_MASTER_PUBLIC_KEY_PAIRING_CHECK_FAILURE = "144"; // "ETHDKG: Master key submission pairing check failed!"
    bytes32 public constant ETHDKG_NOT_IN_GPKJ_SUBMISSION_PHASE = "145"; // "ETHDKG: Not in GPKJ submission phase"
    bytes32 public constant ETHDKG_PARTICIPANT_SUBMITTED_GPKJ_IN_ROUND = "146"; // "ETHDKG: Participant already submitted GPKj this ETHDKG round!"
    bytes32 public constant ETHDKG_GPKJ_ZERO = "147"; // "ETHDKG: GPKj cannot be all zeros!"
    bytes32 public constant ETHDKG_NOT_IN_POST_GPKJ_DISPUTE_PHASE = "148"; // "ETHDKG: should be in post-GPKJDispute phase!"
    bytes32 public constant ETHDKG_REQUISITES_INCOMPLETE = "149"; //  "ETHDKG: Not all requisites to complete this ETHDKG round were completed!"
    bytes32 public constant ETHDKG_KEYSHARE_PHASE_INVALID_NONCE = "150"; // "ETHDKG: Key share submission failed, participant with invalid nonce!"
    bytes32 public constant ETHDKG_MIGRATION_INVALID_NONCE = "151"; // "ETHDKG: Only can execute this with nonce 0!"
    bytes32 public constant ETHDKG_MIGRATION_INPUT_DATA_MISMATCH = "152"; // "ETHDKG: All input data length should match!"
}
