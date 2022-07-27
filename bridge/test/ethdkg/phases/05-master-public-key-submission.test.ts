import { BigNumberish } from "ethers";
import { ethers, expect } from "hardhat";
import { getValidatorEthAccount } from "../../setup";
import { validators4 } from "../assets/4-validators-successful-case";
import {
  assertETHDKGPhase,
  assertEventMPKSet,
  Phase,
  startAtMPKSubmission,
  startAtSubmitKeyShares,
  submitValidatorsKeyShares,
  waitNextPhaseStartDelay,
} from "../setup";

describe("ETHDKG: Submit Master Public Key", () => {
  it("should not allow submission of master public key when not in MPKSubmission phase", async () => {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );
    // distribute shares for all but 1 validators
    await submitValidatorsKeyShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 3),
      expectedNonce
    );

    await assertETHDKGPhase(ethdkg, Phase.KeyShareSubmission);
    await waitNextPhaseStartDelay(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.KeyShareSubmission);

    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[3].address))
        .submitMasterPublicKey(validators4[3].mpk)
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInMasterPublicKeySubmissionPhase`
      )
      .withArgs(Phase.KeyShareSubmission);
  });

  it("should allow submission of master public key by a non-validator", async () => {
    const [ethdkg, , expectedNonce] = await startAtMPKSubmission(validators4);

    // non-validator user tries to submit the Master Public key
    const validator11 = "0x23EA3Bad9115d436190851cF4C49C1032fA7579A";
    const val11MPK: [BigNumberish, BigNumberish, BigNumberish, BigNumberish] =
      validators4[0].mpk;

    const tx = await ethdkg
      .connect(await getValidatorEthAccount(validator11))
      .submitMasterPublicKey(val11MPK);

    await assertEventMPKSet(tx, expectedNonce, val11MPK);
  });

  it("should not allow submission of master public key more than once", async () => {
    const [ethdkg, , expectedNonce] = await startAtMPKSubmission(validators4);

    // non-validator user tries to submit the Master Public key
    const validator11 = "0x23EA3Bad9115d436190851cF4C49C1032fA7579A";
    const val11MPK: [BigNumberish, BigNumberish, BigNumberish, BigNumberish] =
      validators4[0].mpk;

    const tx = await ethdkg
      .connect(await getValidatorEthAccount(validator11))
      .submitMasterPublicKey(val11MPK);

    await assertEventMPKSet(tx, expectedNonce, val11MPK);

    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validator11))
        .submitMasterPublicKey(val11MPK)
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInMasterPublicKeySubmissionPhase`
      )
      .withArgs(Phase.GPKJSubmission);
  });

  it("should not allow submission of empty master public key", async () => {
    const [ethdkg, ,] = await startAtMPKSubmission(validators4);

    // empty MPK
    const mpk: [BigNumberish, BigNumberish, BigNumberish, BigNumberish] = [
      "0",
      "0",
      "0",
      "0",
    ];

    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitMasterPublicKey(mpk)
    ).to.be.revertedWithCustomError(
      ethDKGPhases,
      `MasterPublicKeyPairingCheckFailure`
    );
  });
});
