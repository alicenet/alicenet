import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing CustomEnumerableMaps error codes", async () => {
  let customEnumerableMapsCodesContract: Contract;

  beforeEach(async function () {
    customEnumerableMapsCodesContract = await deployLibrary(
      "CustomEnumerableMapsErrorCodes"
    );
  });

  it("CUSTOMENUMMAP_KEY_NOT_IN_MAP returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      customEnumerableMapsCodesContract.CUSTOMENUMMAP_KEY_NOT_IN_MAP,
      1900
    );
  });
});
