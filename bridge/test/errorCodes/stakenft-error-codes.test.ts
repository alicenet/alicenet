import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing StakingNFT error codes", async () => {
  let errorCodesContract: Contract;

  beforeEach(async function () {
    errorCodesContract = await deployLibrary("StakingNFTErrorCodes");
  });

  it("STAKENFT_CALLER_NOT_TOKEN_OWNER returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.STAKENFT_CALLER_NOT_TOKEN_OWNER,
      600
    );
  });
  it("STAKENFT_LOCK_DURATION_GREATER_THAN_GOVERNANCE_LOCK returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.STAKENFT_LOCK_DURATION_GREATER_THAN_GOVERNANCE_LOCK,
      601
    );
  });
  it("STAKENFT_LOCK_DURATION_GREATER_THAN_MINT_LOCK returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.STAKENFT_LOCK_DURATION_GREATER_THAN_MINT_LOCK,
      602
    );
  });
  it("STAKENFT_LOCK_DURATION_WITHDRAW_TIME_NOT_REACHED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.STAKENFT_LOCK_DURATION_WITHDRAW_TIME_NOT_REACHED,
      603
    );
  });
  it("STAKENFT_INVALID_TOKEN_ID returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.STAKENFT_INVALID_TOKEN_ID,
      604
    );
  });
  it("STAKENFT_MINT_AMOUNT_EXCEEDS_MAX_SUPPLY returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.STAKENFT_MINT_AMOUNT_EXCEEDS_MAX_SUPPLY,
      605
    );
  });
  it("STAKENFT_FREE_AFTER_TIME_NOT_REACHED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.STAKENFT_FREE_AFTER_TIME_NOT_REACHED,
      606
    );
  });
  it("STAKENFT_BALANCE_LESS_THAN_RESERVE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.STAKENFT_BALANCE_LESS_THAN_RESERVE,
      607
    );
  });
  it("STAKENFT_SLUSH_TOO_LARGE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.STAKENFT_SLUSH_TOO_LARGE,
      608
    );
  });
});
