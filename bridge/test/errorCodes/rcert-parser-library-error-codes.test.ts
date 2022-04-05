import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing RCertParserLibrary error codes", async () => {
  let errorCodesContract: Contract;

  beforeEach(async function () {
    errorCodesContract = await deployLibrary("RCertParserLibraryErrorCodes");
  });

  it("RCERTPARSERLIB_DATA_OFFSET_OVERFLOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.RCERTPARSERLIB_DATA_OFFSET_OVERFLOW,
      1400
    );
  });
  it("RCERTPARSERLIB_INSUFFICIENT_BYTES returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.RCERTPARSERLIB_INSUFFICIENT_BYTES,
      1401
    );
  });
  it("RCERTPARSERLIB_INSUFFICIENT_BYTES_TO_EXTRACT returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.RCERTPARSERLIB_INSUFFICIENT_BYTES_TO_EXTRACT,
      1402
    );
  });
});
