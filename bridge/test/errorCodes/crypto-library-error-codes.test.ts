import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing CryptoLibrary error codes", async () => {
  let cryptoLibraryCodesContract: Contract;

  beforeEach(async function () {
    cryptoLibraryCodesContract = await deployLibrary("CryptoLibraryErrorCodes");
  });

  it("CRYPTOLIB_ELLIPTIC_CURVE_ADDITION returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      cryptoLibraryCodesContract.CRYPTOLIB_ELLIPTIC_CURVE_ADDITION,
      700
    );
  });
  it("CRYPTOLIB_ELLIPTIC_CURVE_MULTIPLICATION returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      cryptoLibraryCodesContract.CRYPTOLIB_ELLIPTIC_CURVE_MULTIPLICATION,
      701
    );
  });
  it("CRYPTOLIB_ELLIPTIC_CURVE_PAIRING returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      cryptoLibraryCodesContract.CRYPTOLIB_ELLIPTIC_CURVE_PAIRING,
      702
    );
  });
  it("CRYPTOLIB_MODULAR_EXPONENTIATION returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      cryptoLibraryCodesContract.CRYPTOLIB_MODULAR_EXPONENTIATION,
      703
    );
  });
  it("CRYPTOLIB_HASH_POINT_NOT_ON_CURVE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      cryptoLibraryCodesContract.CRYPTOLIB_HASH_POINT_NOT_ON_CURVE,
      704
    );
  });
  it("CRYPTOLIB_HASH_POINT_UNSAFE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      cryptoLibraryCodesContract.CRYPTOLIB_HASH_POINT_UNSAFE,
      705
    );
  });
  it("CRYPTOLIB_POINT_NOT_ON_CURVE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      cryptoLibraryCodesContract.CRYPTOLIB_POINT_NOT_ON_CURVE,
      706
    );
  });
  it("CRYPTOLIB_SIGNATURES_INDICES_LENGTH_MISMATCH returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      cryptoLibraryCodesContract.CRYPTOLIB_SIGNATURES_INDICES_LENGTH_MISMATCH,
      707
    );
  });
  it("CRYPTOLIB_SIGNATURES_LENGTH_THRESHOLD_NOT_MET returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      cryptoLibraryCodesContract.CRYPTOLIB_SIGNATURES_LENGTH_THRESHOLD_NOT_MET,
      708
    );
  });
  it("CRYPTOLIB_INVERSE_ARRAY_INCORRECT returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      cryptoLibraryCodesContract.CRYPTOLIB_INVERSE_ARRAY_INCORRECT,
      709
    );
  });
  it("CRYPTOLIB_INVERSE_ARRAY_THRESHOLD_EXCEEDED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      cryptoLibraryCodesContract.CRYPTOLIB_INVERSE_ARRAY_THRESHOLD_EXCEEDED,
      710
    );
  });
  it("CRYPTOLIB_POINTSG1_INDICES_LENGTH_MISMATCH returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      cryptoLibraryCodesContract.CRYPTOLIB_POINTSG1_INDICES_LENGTH_MISMATCH,
      711
    );
  });
  it("CRYPTOLIB_K_EQUAL_TO_J returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      cryptoLibraryCodesContract.CRYPTOLIB_K_EQUAL_TO_J,
      712
    );
  });
});
