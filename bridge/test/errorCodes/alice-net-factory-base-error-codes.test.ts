import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing AliceNetFactoryBase error codes", async () => {
  let aliceNetFactoryBaseCodesContract: Contract;

  beforeEach(async function () {
    aliceNetFactoryBaseCodesContract = await deployLibrary(
      "AliceNetFactoryBaseErrorCodes"
    );
  });

  it("ALICENETFACTORYBASE_UNAUTHORIZED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      aliceNetFactoryBaseCodesContract.ALICENETFACTORYBASE_UNAUTHORIZED,
      900
    );
  });
  it("ALICENETFACTORYBASE_CODE_SIZE_ZERO returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      aliceNetFactoryBaseCodesContract.ALICENETFACTORYBASE_CODE_SIZE_ZERO,
      901
    );
  });
});
