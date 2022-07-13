import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing BClaimsParserLibrary error codes", async () => {
  let bClaimsParserLibraryErrorCodesContract: Contract;

  beforeEach(async function () {
    bClaimsParserLibraryErrorCodesContract = await deployLibrary(
      "BClaimsParserLibraryErrorCodes"
    );
  });

  it("BCLAIMSPARSERLIB_SIZE_THRESHOLD_EXCEEDED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bClaimsParserLibraryErrorCodesContract.BCLAIMSPARSERLIB_SIZE_THRESHOLD_EXCEEDED,
      1100
    );
  });
  it("BCLAIMSPARSERLIB_DATA_OFFSET_OVERFLOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bClaimsParserLibraryErrorCodesContract.BCLAIMSPARSERLIB_DATA_OFFSET_OVERFLOW,
      1101
    );
  });
  it("BCLAIMSPARSERLIB_NOT_ENOUGH_BYTES returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bClaimsParserLibraryErrorCodesContract.BCLAIMSPARSERLIB_NOT_ENOUGH_BYTES,
      1102
    );
  });
  it("BCLAIMSPARSERLIB_CHAINID_ZERO returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bClaimsParserLibraryErrorCodesContract.BCLAIMSPARSERLIB_CHAINID_ZERO,
      1103
    );
  });
  it("BCLAIMSPARSERLIB_HEIGHT_ZERO returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      bClaimsParserLibraryErrorCodesContract.BCLAIMSPARSERLIB_HEIGHT_ZERO,
      1104
    );
  });
});
