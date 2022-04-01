import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing BToken error codes", async () => {
  let bTokenErrorCodesContract: Contract;

  beforeEach(async function () {
    bTokenErrorCodesContract = await deployLibrary("BTokenErrorCodes");
  });

  it("BTOKEN_INVALID_DEPOSIT_ID returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bTokenErrorCodesContract.BTOKEN_INVALID_DEPOSIT_ID,
      300
    );
  });
  it("BTOKEN_INVALID_BALANCE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bTokenErrorCodesContract.BTOKEN_INVALID_BALANCE,
      301
    );
  });
  it("BTOKEN_INVALID_BURN_AMOUNT returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bTokenErrorCodesContract.BTOKEN_INVALID_BURN_AMOUNT,
      302
    );
  });
  it("BTOKEN_CONTRACTS_DISALLOWED_DEPOSITS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bTokenErrorCodesContract.BTOKEN_CONTRACTS_DISALLOWED_DEPOSITS,
      303
    );
  });
  it("BTOKEN_DEPOSIT_AMOUNT_ZERO returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bTokenErrorCodesContract.BTOKEN_DEPOSIT_AMOUNT_ZERO,
      304
    );
  });
  it("BTOKEN_DEPOSIT_BURN_FAIL returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bTokenErrorCodesContract.BTOKEN_DEPOSIT_BURN_FAIL,
      305
    );
  });
  it("BTOKEN_MARKET_SPREAD_TOO_LOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bTokenErrorCodesContract.BTOKEN_MARKET_SPREAD_TOO_LOW,
      306
    );
  });
  it("BTOKEN_MINT_INSUFFICIENT_ETH returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bTokenErrorCodesContract.BTOKEN_MINT_INSUFFICIENT_ETH,
      307
    );
  });
  it("BTOKEN_MINIMUM_MINT_NOT_MET returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bTokenErrorCodesContract.BTOKEN_MINIMUM_MINT_NOT_MET,
      308
    );
  });
  it("BTOKEN_MINIMUM_BURN_NOT_MET returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bTokenErrorCodesContract.BTOKEN_MINIMUM_BURN_NOT_MET,
      309
    );
  });
  it("BTOKEN_SPLIT_VALUE_SUM_ERROR returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bTokenErrorCodesContract.BTOKEN_SPLIT_VALUE_SUM_ERROR,
      310
    );
  });
  it("BTOKEN_BURN_AMOUNT_EXCEEDS_SUPPLY returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bTokenErrorCodesContract.BTOKEN_BURN_AMOUNT_EXCEEDS_SUPPLY,
      311
    );
  });
});
