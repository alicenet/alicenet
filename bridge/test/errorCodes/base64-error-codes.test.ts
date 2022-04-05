import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing Base64 error codes", async () => {
  let base64ErrorCodesContract: Contract;

  beforeEach(async function () {
    base64ErrorCodesContract = await deployLibrary("Base64ErrorCodes");
  });

  it("BASE64_INVALID_INPUT returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      base64ErrorCodesContract.BASE64_INVALID_INPUT,
      1800
    );
  });
});
