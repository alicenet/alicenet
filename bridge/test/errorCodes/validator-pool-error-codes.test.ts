import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing ValidatorPool error codes", async () => {
  let errorCodesContract: Contract;

  beforeEach(async function () {
    errorCodesContract = await deployLibrary("ValidatorPoolErrorCodes");
  });

  it("VALIDATORPOOL_CALLER_NOT_VALIDATOR returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_CALLER_NOT_VALIDATOR,
      800
    );
  });
  it("VALIDATORPOOL_CONSENSUS_RUNNING returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_CONSENSUS_RUNNING,
      801
    );
  });
  it("VALIDATORPOOL_ETHDKG_ROUND_RUNNING returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_ETHDKG_ROUND_RUNNING,
      802
    );
  });
  it("VALIDATORPOOL_ONLY_CONTRACTS_ALLOWED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_ONLY_CONTRACTS_ALLOWED,
      803
    );
  });
  it("VALIDATORPOOL_MIN_BLOCK_INTERVAL_NOT_MET returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_MIN_BLOCK_INTERVAL_NOT_MET,
      804
    );
  });
  it("VALIDATORPOOL_MAX_VALIDATORS_MET returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_MAX_VALIDATORS_MET,
      805
    );
  });
  it("VALIDATORPOOL_REGISTRATION_PARAMETER_LENGTH_MISMATCH returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_REGISTRATION_PARAMETER_LENGTH_MISMATCH,
      806
    );
  });
  it("VALIDATORPOOL_FACTORY_SHOULD_OWN_POSITION returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_FACTORY_SHOULD_OWN_POSITION,
      807
    );
  });
  it("VALIDATORPOOL_VALIDATORS_GREATER_THAN_AVAILABLE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_VALIDATORS_GREATER_THAN_AVAILABLE,
      808
    );
  });
  it("VALIDATORPOOL_PROFITS_ONLY_CLAIMABLE_DURING_CONSENSUS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_PROFITS_ONLY_CLAIMABLE_DURING_CONSENSUS,
      809
    );
  });
  it("VALIDATORPOOL_TOKEN_BALANCE_CHANGED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_TOKEN_BALANCE_CHANGED,
      810
    );
  });
  it("VALIDATORPOOL_ETH_BALANCE_CHANGED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_ETH_BALANCE_CHANGED,
      811
    );
  });
  it("VALIDATORPOOL_SENDER_NOT_IN_EXITING_QUEUE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_SENDER_NOT_IN_EXITING_QUEUE,
      812
    );
  });
  it("VALIDATORPOOL_WAITING_PERIOD_NOT_MET returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_WAITING_PERIOD_NOT_MET,
      813
    );
  });
  it("VALIDATORPOOL_DISHONEST_VALIDATOR_NOT_ACCUSABLE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_DISHONEST_VALIDATOR_NOT_ACCUSABLE,
      814
    );
  });
  it("VALIDATORPOOL_INVALID_INDEX returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_INVALID_INDEX,
      815
    );
  });
  it("VALIDATORPOOL_ADDRESS_ALREADY_VALIDATOR returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_ADDRESS_ALREADY_VALIDATOR,
      816
    );
  });
  it("VALIDATORPOOL_ADDRESS_NOT_VALIDATOR returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_ADDRESS_NOT_VALIDATOR,
      817
    );
  });
  it("VALIDATORPOOL_MINIMUM_STAKE_NOT_MET returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_MINIMUM_STAKE_NOT_MET,
      818
    );
  });
  it("VALIDATORPOOL_PAYOUT_TOO_LOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_PAYOUT_TOO_LOW,
      819
    );
  });
  it("VALIDATORPOOL_ADDRESS_NOT_ACCUSABLE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_ADDRESS_NOT_ACCUSABLE,
      820
    );
  });
  it("VALIDATORPOOL_INSUFFICIENT_FUNDS_IN_STAKE_POSITION returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.VALIDATORPOOL_INSUFFICIENT_FUNDS_IN_STAKE_POSITION,
      821
    );
  });
});
