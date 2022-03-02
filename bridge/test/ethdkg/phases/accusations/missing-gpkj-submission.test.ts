import { validators4 } from "../../assets/4-validators-successful-case";
import { ethers } from "hardhat";
import {
  endCurrentPhase,
  assertETHDKGPhase,
  Phase,
  waitNextPhaseStartDelay,
  startAtGPKJ,
  submitValidatorsGPKJ,
  expect,
} from "../../setup";
import { getValidatorEthAccount } from "../../../setup";

describe("ETHDKG: Accuse participant of not submitting GPKj", () => {
  it("allows accusation of all missing validators at once", async () => {
    let [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(validators4);

    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce,
      0
    );

    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    // accuse all at once
    let missingValidators = validators4.slice(1).map((v) => v.address);
    await ethdkg.accuseParticipantDidNotSubmitGPKJ(missingValidators);

    expect(await ethdkg.getBadParticipants()).to.equal(3);
    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    // should not allow to finish ethdkg
    await endCurrentPhase(ethdkg);
    await expect(
      ethdkg.connect(await getValidatorEthAccount(validators4[0].address)).complete()
    ).to.be.revertedWith("ETHDKG: should be in post-GPKJDispute phase!");
  });

  it("allows accusation of missing validators one at a time", async () => {
    let [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(validators4);

    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce,
      0
    );

    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    // accuse one at a time
    let i = 1;
    for (; i < validators4.length; i++) {
      await ethdkg.accuseParticipantDidNotSubmitGPKJ([validators4[i].address]);
    }

    expect(await ethdkg.getBadParticipants()).to.equal(3);
    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
  });

  it("should not move to next phase if there are accusations related to missing GPKj submissions", async () => {
    let [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(validators4);

    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce,
      0
    );

    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    // accuse one validator only
    ethdkg.accuseParticipantDidNotSubmitGPKJ([validators4[1].address]);

    expect(await ethdkg.getBadParticipants()).to.equal(1);
    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    // wait for next phase
    await endCurrentPhase(ethdkg);

    // should not allow to finish ethdkg
    await expect(
      ethdkg.connect(await getValidatorEthAccount(validators4[0].address)).complete()
    ).to.be.revertedWith("ETHDKG: should be in post-GPKJDispute phase!");

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
  });

  it("should not allow accusations related to missing GPKj submissions while on GPKj submission phase", async () => {
    let [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(validators4);

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    // accuse one validator only
    await expect(
      ethdkg.accuseParticipantDidNotSubmitGPKJ([validators4[1].address])
    ).to.be.revertedWith("ETHDKG: Dispute Failed! Should be in post-GPKJSubmission phase!");
  });

  it("should not allow GPKj submission after the GPKj submission phase", async () => {
    let [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(validators4);

    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce,
      0
    );

    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    await expect(
      submitValidatorsGPKJ(
        ethdkg,
        validatorPool,
        validators4.slice(1, 2),
        expectedNonce,
        0
      )
    ).to.be.revertedWith("ETHDKG: Not in GPKJ submission phase");

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
  });

  it("should not allow missing validators to complete ETHDKG after the GPKj submission phase", async () => {
    let [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(validators4);

    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce,
      0
    );

    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    // should not allow finishing ethdkg
    await expect(
      ethdkg.connect(await getValidatorEthAccount(validators4[1].address)).complete()
    ).to.be.revertedWith("ETHDKG: should be in post-GPKJDispute phase!");

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
  });

  it("should not allow accusation of missing GPKj to a validator that actually submitted it", async () => {
    let [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(validators4);

    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce,
      0
    );

    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    // accuse
    await expect(
      ethdkg.accuseParticipantDidNotSubmitGPKJ([validators4[0].address])
    ).to.be.revertedWith(
      "ETHDKG: Dispute failed! Dishonest participant did participate in this GPKj submission!"
    );

    expect(await ethdkg.getBadParticipants()).to.equal(0);

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
  });

  it("should not allow accusation of missing GPKj to a non-existent validator", async () => {
    let [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(validators4);

    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce,
      0
    );

    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    // accuse
    await expect(
      ethdkg.accuseParticipantDidNotSubmitGPKJ([
        "0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac",
      ])
    ).to.be.revertedWith("ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!");

    expect(await ethdkg.getBadParticipants()).to.equal(0);

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
  });

  it("should not allow accusation of missing GPKj after the accusation window is over", async () => {
    let [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(validators4);

    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce,
      0
    );

    await waitNextPhaseStartDelay(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    // accuse
    await expect(
      ethdkg.accuseParticipantDidNotSubmitGPKJ([validators4[1].address])
    ).to.be.revertedWith("ETHDKG: Dispute Failed! Should be in post-GPKJSubmission phase!");

    expect(await ethdkg.getBadParticipants()).to.equal(0);

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
  });

  it("should not allow accusation of missing GPKj against a list of non-participants, non-validators, and legit participants", async () => {
    let [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(validators4);

    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce,
      0
    );

    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);

    // accuse
    let addresses = validators4.map((v) => v.address);
    addresses.push("0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac");

    await expect(
      ethdkg.accuseParticipantDidNotSubmitGPKJ(addresses)
    ).to.be.revertedWith(
      "ETHDKG: Dispute failed! Dishonest participant did participate in this GPKj submission!"
    );

    expect(await ethdkg.getBadParticipants()).to.equal(0);

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
  });

  it("should not allow double accusation of missing GPKj a non-participating validator", async () => {
    let [ethdkg, validatorPool, expectedNonce] = await startAtGPKJ(validators4);

    await submitValidatorsGPKJ(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce,
      0
    );

    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
    expect(await ethdkg.getBadParticipants()).to.equal(0);

    // accuse
    await ethdkg.accuseParticipantDidNotSubmitGPKJ([validators4[1].address]);
    expect(await ethdkg.getBadParticipants()).to.equal(1);

    await expect(
      ethdkg.accuseParticipantDidNotSubmitGPKJ([validators4[1].address])
    ).to.be.revertedWith("ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!");

    expect(await ethdkg.getBadParticipants()).to.equal(1);

    await assertETHDKGPhase(ethdkg, Phase.GPKJSubmission);
  });
});
