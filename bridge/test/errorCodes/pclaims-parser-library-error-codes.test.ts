import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing PClaimsParserLibrary error codes", async () => {
  let errorCodesContract: Contract;

  beforeEach(async function () {
    errorCodesContract = await deployLibrary("PClaimsParserLibraryErrorCodes");
  });

  it("PCLAIMSPARSERLIB_DATA_OFFSET_OVERFLOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.PCLAIMSPARSERLIB_DATA_OFFSET_OVERFLOW,
      1300
    );
  });
  it("PCLAIMSPARSERLIB_INSUFFICIENT_BYTES returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.PCLAIMSPARSERLIB_INSUFFICIENT_BYTES,
      1301
    );
  });
});
