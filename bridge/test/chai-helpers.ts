import { expect } from "chai";
import { ContractTransaction } from "ethers";

export async function assertErrorMessage(
  p: Promise<ContractTransaction>,
  message: string
): Promise<void> {
  return p.then(
    (value) => {
      expect.fail(`Found value instead of error: ${value}`);
    },
    (reason) => {
      expect(reason.message).to.contain(message);
    }
  );
}
