import { expect } from "chai";

export async function assertErrorMessage(
  p: Promise<any>,
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
