import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing Mutex error codes", async () => {
  let mutexErrorCodesContract: Contract;

  beforeEach(async function () {
    mutexErrorCodesContract = await deployLibrary("MutexErrorCodes");
  });

  it("MUTEX_LOCKED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      mutexErrorCodesContract.MUTEX_LOCKED,
      2300
    );
  });
});
