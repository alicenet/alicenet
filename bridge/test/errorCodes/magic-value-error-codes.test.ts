import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing MagicTokenTransfer error codes", async () => {
  let magicValueErrorCodesContract: Contract;

  beforeEach(async function () {
    magicValueErrorCodesContract = await deployLibrary("MagicValueErrorCodes");
  });

  it("MAGICVALUE_BAD_MAGIC returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      magicValueErrorCodesContract.MAGICVALUE_BAD_MAGIC,
      2200
    );
  });
});
