import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing ETHDKG error codes", async () => {
  let ethdkgErrorCodesContract: Contract;

  beforeEach(async function () {
    ethdkgErrorCodesContract = await deployLibrary("ETHDKGErrorCodes");
  });

  it("ETHDKG_ONLY_VALIDATORS_ALLOWED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_ONLY_VALIDATORS_ALLOWED,
      100
    );
  });
  it("ETHDKG_VARIABLE_CANNOT_BE_SET_WHILE_RUNNING returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_VARIABLE_CANNOT_BE_SET_WHILE_RUNNING,
      101
    );
  });
  it("ETHDKG_MIN_VALIDATORS_NOT_MET returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_MIN_VALIDATORS_NOT_MET,
      102
    );
  });
  it("ETHDKG_NOT_IN_POST_REGISTRATION_PHASE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_NOT_IN_POST_REGISTRATION_PHASE,
      103
    );
  });
  it("ETHDKG_ACCUSED_NOT_VALIDATOR returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_ACCUSED_NOT_VALIDATOR,
      104
    );
  });
  it("ETHDKG_ACCUSED_PARTICIPATING_IN_ROUND returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_ACCUSED_PARTICIPATING_IN_ROUND,
      105
    );
  });
  it("ETHDKG_NOT_IN_POST_SHARED_DISTRIBUTION_PHASE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_NOT_IN_POST_SHARED_DISTRIBUTION_PHASE,
      106
    );
  });
  it("ETHDKG_ACCUSED_NOT_PARTICIPATING_IN_ROUND returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_ACCUSED_NOT_PARTICIPATING_IN_ROUND,
      107
    );
  });
  it("ETHDKG_ACCUSED_DISTRIBUTED_SHARES_IN_ROUND returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_ACCUSED_DISTRIBUTED_SHARES_IN_ROUND,
      108
    );
  });
  it("ETHDKG_ACCUSED_HAS_COMMITMENTS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_ACCUSED_HAS_COMMITMENTS,
      109
    );
  });
  it("ETHDKG_NOT_IN_DISPUTE_PHASE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_NOT_IN_DISPUTE_PHASE,
      110
    );
  });
  it("ETHDKG_DISPUTER_NOT_PARTICIPATING_IN_ROUND returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_DISPUTER_NOT_PARTICIPATING_IN_ROUND,
      111
    );
  });
  it("ETHDKG_ACCUSED_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_ACCUSED_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND,
      112
    );
  });
  it("ETHDKG_DISPUTER_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_DISPUTER_DID_NOT_DISTRIBUTE_SHARES_IN_ROUND,
      113
    );
  });
  it("ETHDKG_SHARES_AND_COMMITMENTS_MISMATCH returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_SHARES_AND_COMMITMENTS_MISMATCH,
      114
    );
  });
  it("ETHDKG_INVALID_KEY_OR_PROOF returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_INVALID_KEY_OR_PROOF,
      115
    );
  });
  it("ETHDKG_NOT_IN_POST_KEYSHARE_SUBMISSION_PHASE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_NOT_IN_POST_KEYSHARE_SUBMISSION_PHASE,
      116
    );
  });
  it("ETHDKG_ACCUSED_SUBMITTED_SHARES_IN_ROUND returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_ACCUSED_SUBMITTED_SHARES_IN_ROUND,
      117
    );
  });
  it("ETHDKG_NOT_IN_POST_GPKJ_SUBMISSION_PHASE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_NOT_IN_POST_GPKJ_SUBMISSION_PHASE,
      118
    );
  });
  it("ETHDKG_ACCUSED_DID_NOT_PARTICIPATE_IN_GPKJ_SUBMISSION returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_ACCUSED_DID_NOT_PARTICIPATE_IN_GPKJ_SUBMISSION,
      120
    );
  });
  it("ETHDKG_ACCUSED_DISTRIBUTED_GPKJ returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_ACCUSED_DISTRIBUTED_GPKJ,
      121
    );
  });
  it("ETHDKG_ACCUSED_DID_NOT_SUBMIT_GPKJ_IN_ROUND returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_ACCUSED_DID_NOT_SUBMIT_GPKJ_IN_ROUND,
      122
    );
  });
  it("ETHDKG_DISPUTER_DID_NOT_SUBMIT_GPKJ_IN_ROUND returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_DISPUTER_DID_NOT_SUBMIT_GPKJ_IN_ROUND,
      123
    );
  });
  it("ETHDKG_INVALID_ARGS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_INVALID_ARGS,
      124
    );
  });
  it("ETHDKG_INVALID_COMMITMENTS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_INVALID_COMMITMENTS,
      125
    );
  });
  it("ETHDKG_INVALID_OR_DUPLICATED_PARTICIPANT returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_INVALID_OR_DUPLICATED_PARTICIPANT,
      126
    );
  });
  it("ETHDKG_INVALID_SHARES_OR_COMMITMENTS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_INVALID_SHARES_OR_COMMITMENTS,
      127
    );
  });
  it("ETHDKG_NOT_IN_REGISTRATION_PHASE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_NOT_IN_REGISTRATION_PHASE,
      128
    );
  });
  it("ETHDKG_PUBLIC_KEY_ZERO returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_PUBLIC_KEY_ZERO,
      130
    );
  });
  it("ETHDKG_PUBLIC_KEY_NOT_ON_CURVE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_PUBLIC_KEY_NOT_ON_CURVE,
      131
    );
  });
  it("ETHDKG_PARTICIPANT_PARTICIPATING_IN_ROUND returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_PARTICIPANT_PARTICIPATING_IN_ROUND,
      132
    );
  });
  it("ETHDKG_NOT_IN_SHARED_DISTRIBUTION_PHASE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_NOT_IN_SHARED_DISTRIBUTION_PHASE,
      133
    );
  });
  it("ETHDKG_INVALID_NONCE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_INVALID_NONCE,
      134
    );
  });
  it("ETHDKG_PARTICIPANT_DISTRIBUTED_SHARES_IN_ROUND returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_PARTICIPANT_DISTRIBUTED_SHARES_IN_ROUND,
      135
    );
  });
  it("ETHDKG_INVALID_NUM_ENCRYPTED_SHARES returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_INVALID_NUM_ENCRYPTED_SHARES,
      136
    );
  });
  it("ETHDKG_INVALID_NUM_COMMITMENTS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_INVALID_NUM_COMMITMENTS,
      137
    );
  });
  it("ETHDKG_COMMITMENT_NOT_ON_CURVE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_COMMITMENT_NOT_ON_CURVE,
      138
    );
  });
  it("ETHDKG_COMMITMENT_ZERO returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_COMMITMENT_ZERO,
      138
    );
  });
  it("ETHDKG_DISTRIBUTED_SHARE_HASH_ZERO returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_DISTRIBUTED_SHARE_HASH_ZERO,
      139
    );
  });
  it("ETHDKG_NOT_IN_KEYSHARE_SUBMISSION_PHASE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_NOT_IN_KEYSHARE_SUBMISSION_PHASE,
      140
    );
  });
  it("ETHDKG_PARTICIPANT_SUBMITTED_KEYSHARES_IN_ROUND returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_PARTICIPANT_SUBMITTED_KEYSHARES_IN_ROUND,
      141
    );
  });
  it("ETHDKG_INVALID_KEYSHARE_G1 returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_INVALID_KEYSHARE_G1,
      141
    );
  });
  it("ETHDKG_INVALID_KEYSHARE_G2 returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_INVALID_KEYSHARE_G2,
      142
    );
  });
  it("ETHDKG_NOT_IN_MASTER_PUBLIC_KEY_SUBMISSION_PHASE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_NOT_IN_MASTER_PUBLIC_KEY_SUBMISSION_PHASE,
      143
    );
  });
  it("ETHDKG_MASTER_PUBLIC_KEY_PAIRING_CHECK_FAILURE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_MASTER_PUBLIC_KEY_PAIRING_CHECK_FAILURE,
      144
    );
  });
  it("ETHDKG_NOT_IN_GPKJ_SUBMISSION_PHASE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_NOT_IN_GPKJ_SUBMISSION_PHASE,
      145
    );
  });
  it("ETHDKG_PARTICIPANT_SUBMITTED_GPKJ_IN_ROUND returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_PARTICIPANT_SUBMITTED_GPKJ_IN_ROUND,
      146
    );
  });
  it("ETHDKG_GPKJ_ZERO returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_GPKJ_ZERO,
      147
    );
  });
  it("ETHDKG_NOT_IN_POST_GPKJ_DISPUTE_PHASE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_NOT_IN_POST_GPKJ_DISPUTE_PHASE,
      148
    );
  });
  it("ETHDKG_REQUISITES_INCOMPLETE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_REQUISITES_INCOMPLETE,
      149
    );
  });
  it("ETHDKG_KEYSHARE_PHASE_INVALID_NONCE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      ethdkgErrorCodesContract.ETHDKG_KEYSHARE_PHASE_INVALID_NONCE,
      150
    );
  });
});
