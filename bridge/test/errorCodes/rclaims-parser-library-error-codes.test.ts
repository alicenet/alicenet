import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing RClaimsParserLibrary error codes", async () => {
  let errorCodesContract: Contract;

  beforeEach(async function () {
    errorCodesContract = await deployLibrary("RClaimsParserLibraryErrorCodes");
  });

  it("RCLAIMSPARSERLIB_DATA_OFFSET_OVERFLOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.RCLAIMSPARSERLIB_DATA_OFFSET_OVERFLOW,
      1500
    );
  });
  it("RCLAIMSPARSERLIB_INSUFFICIENT_BYTES returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.RCLAIMSPARSERLIB_INSUFFICIENT_BYTES,
      1501
    );
  });
  it("RCLAIMSPARSERLIB_CHAINID_ZERO returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.RCLAIMSPARSERLIB_CHAINID_ZERO,
      1502
    );
  });
  it("RCLAIMSPARSERLIB_HEIGHT_ZERO returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.RCLAIMSPARSERLIB_HEIGHT_ZERO,
      1503
    );
  });
  it("RCLAIMSPARSERLIB_ROUND_ZERO returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.RCLAIMSPARSERLIB_ROUND_ZERO,
      1504
    );
  });
});
