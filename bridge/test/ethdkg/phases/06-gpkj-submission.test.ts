import { assertErrorMessage } from "../../chai-helpers";
import { getValidatorEthAccount } from "../../setup";
import { validators4 } from "../assets/4-validators-successful-case";
import {
  expect,
  startAtGPKJ,
  startAtMPKSubmission,
  submitValidatorsGPKJ,
} from "../setup";

describe("ETHDKG: GPKj submission", () => {
  it("should not allow GPKj submission outside of GPKjSubmission phase", async () => {
    const [ethdkg] = await startAtMPKSubmission(validators4);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitGPKJ(validators4[0].gpkj)
    ).to.be.revertedWith("145");
  });

  it("should not allow non-validators to submit GPKj submission", async () => {
    const [ethdkg] = await startAtGPKJ(validators4);

    const validator11 = "0x23EA3Bad9115d436190851cF4C49C1032fA7579A";
    await assertErrorMessage(
      ethdkg
        .connect(await getValidatorEthAccount(validator11))
        .submitGPKJ(validators4[0].gpkj),
      `OnlyValidatorsAllowed("${validator11}")`
    );
  });

  it("should not allow submission of GPKj more than once from a validator", async () => {
    const [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(
      validators4
    );

    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce,
      0
    );

    await expect(
      submitValidatorsGPKJ(
        ethdkg,
        validatorPool,
        validators4.slice(0, 1),
        expectedNonce,
        0
      )
    ).to.be.revertedWith("146");
  });

  it("should not allow submission of empty GPKj", async () => {
    const [ethdkg] = await startAtGPKJ(validators4);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitGPKJ(["0", "0", "0", "0"])
    ).to.be.revertedWith("147");
  });
});
