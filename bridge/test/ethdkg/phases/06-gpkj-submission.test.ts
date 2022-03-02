import { validators4 } from "../assets/4-validators-successful-case";
import { ethers } from "hardhat";
import {
  startAtMPKSubmission,
  startAtGPKJ,
  submitValidatorsGPKJ,
  expect,
} from "../setup";
import { getValidatorEthAccount } from "../../setup";

describe("ETHDKG: GPKj submission", () => {
  it("should not allow GPKj submission outside of GPKjSubmission phase", async () => {
    let [ethdkg] = await startAtMPKSubmission(validators4);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitGPKJ(validators4[0].gpkj)
    ).to.be.revertedWith("ETHDKG: Not in GPKJ submission phase");
  });

  it("should not allow non-validators to submit GPKj submission", async () => {
    let [ethdkg] = await startAtGPKJ(validators4);

    const validator11 = "0x23EA3Bad9115d436190851cF4C49C1032fA7579A";

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validator11))
        .submitGPKJ(validators4[0].gpkj)
    ).to.be.revertedWith("ETHDKG: Only validators allowed!");
  });

  it("should not allow submission of GPKj more than once from a validator", async () => {
    let [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(validators4);

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
    ).to.be.revertedWith(
      "ETHDKG: Participant already submitted GPKj this ETHDKG round!"
    );
  });

  it("should not allow submission of empty GPKj", async () => {
    let [ethdkg] = await startAtGPKJ(validators4);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitGPKJ(["0", "0", "0", "0"])
    ).to.be.revertedWith("ETHDKG: GPKj cannot be all zeros!");
  });
});
