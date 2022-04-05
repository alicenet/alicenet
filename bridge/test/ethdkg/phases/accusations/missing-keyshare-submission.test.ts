import { getValidatorEthAccount } from "../../../setup";
import { validators4 } from "../../assets/4-validators-successful-case";
import {
  endCurrentAccusationPhase,
  endCurrentPhase,
  expect,
  startAtSubmitKeyShares,
  submitValidatorsKeyShares,
} from "../../setup";

describe("ETHDKG: Accuse participant of not submitting key shares", () => {
  it("allows accusation of all missing validators after Key share phase", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    // distribute shares only for validators 0 and 1
    await submitValidatorsKeyShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Key Share phase
    await endCurrentPhase(ethdkg);

    await expect(await ethdkg.getBadParticipants()).to.equal(0);
    await ethdkg.accuseParticipantDidNotSubmitKeyShares([
      validators4[2].address,
      validators4[3].address,
    ]);

    await expect(await ethdkg.getBadParticipants()).to.equal(2);
    // move to the end of Key Share Accusation phase
    await endCurrentPhase(ethdkg);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitMasterPublicKey(validators4[0].mpk)
    ).to.be.revertedWith("143");
  });

  it("allows accusation of some missing validators after Key share phase", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    // distribute shares only for validators 0 and 1
    await submitValidatorsKeyShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Key Share phase
    await endCurrentPhase(ethdkg);

    await expect(await ethdkg.getBadParticipants()).to.equal(0);
    await ethdkg.accuseParticipantDidNotSubmitKeyShares([
      validators4[2].address,
    ]);
    await expect(await ethdkg.getBadParticipants()).to.equal(1);
    await ethdkg.accuseParticipantDidNotSubmitKeyShares([
      validators4[3].address,
    ]);
    await expect(await ethdkg.getBadParticipants()).to.equal(2);

    // move to the end of Key Share Accusation phase
    await endCurrentPhase(ethdkg);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitMasterPublicKey(validators4[0].mpk)
    ).to.be.revertedWith("143");
  });

  it("do not allow validators to proceed to the next phase if not all validators submitted their key shares", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    // distribute shares only for validators 0 and 1
    await submitValidatorsKeyShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Key Share phase
    await endCurrentPhase(ethdkg);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitMasterPublicKey(validators4[0].mpk)
    ).to.be.revertedWith("143");
  });

  it("won't let not-distributed shares accusations to take place while ETHDKG Distribute Share Phase is open", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    // distribute shares only for validators 0 and 1
    await submitValidatorsKeyShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    await expect(
      ethdkg.accuseParticipantDidNotSubmitKeyShares([validators4[2].address])
    ).to.be.revertedWith("116");
  });

  it("should not allow validators who did not submit key shares in time to submit on the accusation phase", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    // distribute shares only for validators 0 and 1
    await submitValidatorsKeyShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Key Share Accusation phase
    await endCurrentPhase(ethdkg);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[2].address))
        .submitKeyShare(
          validators4[2].keyShareG1,
          validators4[2].keyShareG1CorrectnessProof,
          validators4[2].keyShareG2
        )
    ).to.revertedWith("140");

    // non-participant user tries to go to the next phase
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[3].address))
        .submitMasterPublicKey(validators4[3].mpk)
    ).to.be.revertedWith("143");
  });

  it("should not allow accusation of not submitting key shares of validators submitted their key shares", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    // distribute shares only for validators 0 and 1
    await submitValidatorsKeyShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Key Share Accusation phase
    await endCurrentPhase(ethdkg);

    await expect(await ethdkg.getBadParticipants()).to.equal(0);

    await expect(
      ethdkg.accuseParticipantDidNotSubmitKeyShares([validators4[0].address])
    ).to.be.revertedWith("117");

    await expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow accusation of not submitting key shares for non-validators", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    // distribute shares only for validators 0 and 1
    await submitValidatorsKeyShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Key Share Accusation phase
    await endCurrentPhase(ethdkg);

    await expect(await ethdkg.getBadParticipants()).to.equal(0);

    // try to accuse a non validator
    await expect(
      ethdkg.accuseParticipantDidNotSubmitKeyShares([
        "0x23EA3Bad9115d436190851cF4C49C1032fA7579A",
      ])
    ).to.be.revertedWith("104");

    await expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow not submitted key shares accusations after accusation window has finished", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    // distribute shares only for validators 0 and 1
    await submitValidatorsKeyShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Key Share phase
    await endCurrentPhase(ethdkg);

    // move to the end of Key Share Accusation phase
    await endCurrentAccusationPhase(ethdkg);

    await expect(
      ethdkg.accuseParticipantDidNotSubmitKeyShares([validators4[2].address])
    ).to.be.revertedWith("116");

    await expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow accusing a user that submitted the key shares in the middle of the ones that did not", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    // distribute shares only for validators 0 and 1
    await submitValidatorsKeyShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Key Share phase
    await endCurrentPhase(ethdkg);

    await expect(
      ethdkg.accuseParticipantDidNotSubmitKeyShares([
        validators4[2].address,
        validators4[0].address,
      ])
    ).to.be.revertedWith("117");

    await expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow double accusation of a user that did not submit his key shares", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    // distribute shares only for validators 0 and 1
    await submitValidatorsKeyShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Key Share phase
    await endCurrentPhase(ethdkg);

    await ethdkg.accuseParticipantDidNotSubmitKeyShares([
      validators4[2].address,
    ]);

    await expect(await ethdkg.getBadParticipants()).to.equal(1);

    await expect(
      ethdkg.accuseParticipantDidNotSubmitKeyShares([validators4[2].address])
    ).to.be.revertedWith("104");

    await expect(await ethdkg.getBadParticipants()).to.equal(1);
  });
});
