import { ethers } from "hardhat";
import { getValidatorEthAccount } from "../../setup";
import { validators4 } from "../assets/4-validators-successful-case";
import {
  assertETHDKGPhase,
  assertEventValidatorSetCompleted,
  endCurrentPhase,
  expect,
  Phase,
  PLACEHOLDER_ADDRESS,
  startAtGPKJ,
  submitValidatorsGPKJ,
} from "../setup";

describe("ETHDKG: ETHDKG Completion", () => {
  it("should not allow completion until after the DisputeGPKj phase", async () => {
    const [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(
      validators4
    );

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce,
      0
    );

    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .complete()
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInPostGPKJDisputePhase`
      )
      .withArgs(Phase.DisputeGPKJSubmission);

    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);
    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);

    const tx = await ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .complete();

    await assertEventValidatorSetCompleted(
      tx,
      validators4.length,
      expectedNonce,
      0,
      0,
      0,
      validators4[0].mpk
    );
  });

  it("should not allow non-validators to complete ETHDKG", async () => {
    const [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(
      validators4
    );
    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce,
      0
    );

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .complete()
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInPostGPKJDisputePhase`
      )
      .withArgs(Phase.DisputeGPKJSubmission);

    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);
    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);

    const validatorAddress = "0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac";
    // non-validator tries to complete ethdkg
    await expect(
      ethdkg.connect(await getValidatorEthAccount(validatorAddress)).complete()
    )
      .to.be.revertedWithCustomError(ethdkg, `OnlyValidatorsAllowed`)
      .withArgs(validatorAddress);
  });

  it("should not allow double completion of ETHDKG", async () => {
    const [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(
      validators4
    );

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce,
      0
    );

    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .complete()
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInPostGPKJDisputePhase`
      )
      .withArgs(Phase.DisputeGPKJSubmission);

    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);
    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);

    const tx = await ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .complete();

    await assertEventValidatorSetCompleted(
      tx,
      validators4.length,
      expectedNonce,
      0,
      0,
      0,
      validators4[0].mpk
    );

    // try completing again
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .complete()
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInPostGPKJDisputePhase`
      )
      .withArgs(Phase.Completion);
  });

  it("should not allow validators to participate in previous phases", async () => {
    const [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(
      validators4
    );
    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );
    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4,
      expectedNonce,
      0
    );

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .complete()
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInPostGPKJDisputePhase`
      )
      .withArgs(Phase.DisputeGPKJSubmission);

    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);
    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.DisputeGPKJSubmission);

    const tx = await ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .complete();

    await assertEventValidatorSetCompleted(
      tx,
      validators4.length,
      expectedNonce,
      0,
      0,
      0,
      validators4[0].mpk
    );

    // try participating in previous phases
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .register(validators4[0].aliceNetPublicKey)
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInRegistrationPhase`
      )
      .withArgs(Phase.Completion);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .accuseParticipantNotRegistered([])
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `ETHDKGNotInPostRegistrationAccusationPhase`
      )
      .withArgs(Phase.Completion);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .distributeShares(
          validators4[0].encryptedShares,
          validators4[0].commitments
        )
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInSharedDistributionPhase`
      )
      .withArgs(Phase.Completion);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .accuseParticipantDidNotDistributeShares([])
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `NotInPostSharedDistributionPhase`
      )
      .withArgs(Phase.Completion);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .accuseParticipantDistributedBadShares(
          PLACEHOLDER_ADDRESS,
          [],
          [
            [0, 0],
            [0, 0],
          ],
          [0, 0],
          [0, 0]
        )
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `ETHDKGNotInDisputePhase`
      )
      .withArgs(Phase.Completion);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitKeyShare(
          validators4[0].keyShareG1,
          validators4[0].keyShareG1CorrectnessProof,
          validators4[0].keyShareG2
        )
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInKeyshareSubmissionPhase`
      )
      .withArgs(Phase.Completion);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .accuseParticipantDidNotSubmitKeyShares([])
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `ETHDKGNotInPostKeyshareSubmissionPhase`
      )
      .withArgs(Phase.Completion);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitMasterPublicKey([0, 0, 0, 0])
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInMasterPublicKeySubmissionPhase`
      )
      .withArgs(Phase.Completion);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitGPKJ([0, 0, 0, 0])
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInGPKJSubmissionPhase`
      )
      .withArgs(Phase.Completion);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .accuseParticipantDidNotSubmitGPKJ([])
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `ETHDKGNotInPostGPKJSubmissionPhase`
      )
      .withArgs(Phase.Completion);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .accuseParticipantSubmittedBadGPKJ(
          [],
          [],
          [[[0, 0]]],
          PLACEHOLDER_ADDRESS
        )
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `ETHDKGNotInPostGPKJSubmissionPhase`
      )
      .withArgs(Phase.Completion);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .complete()
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInPostGPKJDisputePhase`
      )
      .withArgs(Phase.Completion);
  });
});
