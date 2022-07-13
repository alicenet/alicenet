import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing BaseParserLibrary error codes", async () => {
  let baseParserLibraryErrorCodesContract: Contract;

  beforeEach(async function () {
    baseParserLibraryErrorCodesContract = await deployLibrary(
      "BaseParserLibraryErrorCodes"
    );
  });

  it("BASEPARSERLIB_OFFSET_PARAMETER_OVERFLOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_OFFSET_PARAMETER_OVERFLOW,
      1000
    );
  });

  it("BASEPARSERLIB_OFFSET_OUT_OF_BOUNDS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_OFFSET_OUT_OF_BOUNDS,
      1001
    );
  });

  it("BASEPARSERLIB_LE_UINT16_OFFSET_PARAMETER_OVERFLOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_LE_UINT16_OFFSET_PARAMETER_OVERFLOW,
      1002
    );
  });

  it("BASEPARSERLIB_LE_UINT16_OFFSET_OUT_OF_BOUNDS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_LE_UINT16_OFFSET_OUT_OF_BOUNDS,
      1003
    );
  });
  it("BASEPARSERLIB_BE_UINT16_OFFSET_PARAMETER_OVERFLOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_BE_UINT16_OFFSET_PARAMETER_OVERFLOW,
      1004
    );
  });
  it("BASEPARSERLIB_BE_UINT16_OFFSET_OUT_OF_BOUNDS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_BE_UINT16_OFFSET_OUT_OF_BOUNDS,
      1005
    );
  });
  it("BASEPARSERLIB_BOOL_OFFSET_PARAMETER_OVERFLOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_BOOL_OFFSET_PARAMETER_OVERFLOW,
      1006
    );
  });
  it("BASEPARSERLIB_BOOL_OFFSET_OUT_OF_BOUNDS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_BOOL_OFFSET_OUT_OF_BOUNDS,
      1007
    );
  });
  it("BASEPARSERLIB_LE_UINT256_OFFSET_PARAMETER_OVERFLOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_LE_UINT256_OFFSET_PARAMETER_OVERFLOW,
      1008
    );
  });
  it("BASEPARSERLIB_LE_UINT256_OFFSET_OUT_OF_BOUNDS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_LE_UINT256_OFFSET_OUT_OF_BOUNDS,
      1009
    );
  });
  it("BASEPARSERLIB_BE_UINT256_OFFSET_PARAMETER_OVERFLOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_BE_UINT256_OFFSET_PARAMETER_OVERFLOW,
      1010
    );
  });
  it("BASEPARSERLIB_BE_UINT256_OFFSET_OUT_OF_BOUNDS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_BE_UINT256_OFFSET_OUT_OF_BOUNDS,
      1011
    );
  });
  it("BASEPARSERLIB_BYTES_OFFSET_PARAMETER_OVERFLOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_BYTES_OFFSET_PARAMETER_OVERFLOW,
      1012
    );
  });
  it("BASEPARSERLIB_BYTES_OFFSET_OUT_OF_BOUNDS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_BYTES_OFFSET_OUT_OF_BOUNDS,
      1013
    );
  });
  it("BASEPARSERLIB_BYTES32_OFFSET_PARAMETER_OVERFLOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_BYTES32_OFFSET_PARAMETER_OVERFLOW,
      1014
    );
  });
  it("BASEPARSERLIB_BYTES32_OFFSET_OUT_OF_BOUNDS returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      baseParserLibraryErrorCodesContract.BASEPARSERLIB_BYTES32_OFFSET_OUT_OF_BOUNDS,
      1015
    );
  });
});
