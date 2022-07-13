import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing Governance error codes", async () => {
  let governanceCodesContract: Contract;

  beforeEach(async function () {
    governanceCodesContract = await deployLibrary("GovernanceErrorCodes");
  });

  it("GOVERNANCE_ONLY_FACTORY_ALLOWED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      governanceCodesContract.GOVERNANCE_ONLY_FACTORY_ALLOWED,
      200
    );
  });
});
