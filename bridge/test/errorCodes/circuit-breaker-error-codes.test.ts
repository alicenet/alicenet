import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing CircuitBreaker error codes", async () => {
  let circuitBreakerErrorCodesContract: Contract;

  beforeEach(async function () {
    circuitBreakerErrorCodesContract = await deployLibrary(
      "CircuitBreakerErrorCodes"
    );
  });

  it("CIRCUIT_BREAKER_OPENED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      circuitBreakerErrorCodesContract.CIRCUIT_BREAKER_OPENED,
      500
    );
  });
  it("CIRCUIT_BREAKER_CLOSED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      circuitBreakerErrorCodesContract.CIRCUIT_BREAKER_CLOSED,
      501
    );
  });
});
