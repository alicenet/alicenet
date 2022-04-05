import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing TXInPreImageParserLibrary error codes", async () => {
  let errorCodesContract: Contract;

  beforeEach(async function () {
    errorCodesContract = await deployLibrary(
      "TXInPreImageParserLibraryErrorCodes"
    );
  });

  it("TXINPREIMAGEPARSERLIB_DATA_OFFSET_OVERFLOW returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.TXINPREIMAGEPARSERLIB_DATA_OFFSET_OVERFLOW,
      1600
    );
  });
  it("TXINPREIMAGEPARSERLIB_INSUFFICIENT_BYTES returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.TXINPREIMAGEPARSERLIB_INSUFFICIENT_BYTES,
      1601
    );
  });
  it("TXINPREIMAGEPARSERLIB_CHAINID_ZERO returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      errorCodesContract.TXINPREIMAGEPARSERLIB_CHAINID_ZERO,
      1602
    );
  });
});
