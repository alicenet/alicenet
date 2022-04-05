import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing Admin error codes", async () => {
  let adminCodesContract: Contract;

  beforeEach(async function () {
    adminCodesContract = await deployLibrary("AdminErrorCodes");
  });

  it("ADMIN_SENDER_MUST_BE_ADMIN returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      adminCodesContract.ADMIN_SENDER_MUST_BE_ADMIN,
      1700
    );
  });
});
