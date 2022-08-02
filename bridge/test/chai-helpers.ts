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

export async function assertError(
  p: Promise<any>,
  errorName: string,
  errorArgs: Map<string, any>
): Promise<void> {
  return p.then(
    (value) => {
      expect.fail(`Found value instead of error: ${value}`);
    },
    (reason) => {
      expect(reason.errorName).to.equal(errorName);
      expect(reason.errorArgs.length).to.equal(errorArgs.size);

      for (const [key, value] of errorArgs) {
        expect(reason.errorArgs[key]).to.deep.equal(value);
      }
    }
  );
}
