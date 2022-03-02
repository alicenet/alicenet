import { validators4 } from "../../assets/4-validators-successful-case";
import { ethers } from "hardhat";
import {
  endCurrentPhase,
  endCurrentAccusationPhase,
  distributeValidatorsShares,
  startAtDistributeShares,
  expect,
} from "../../setup";
import { getValidatorEthAccount } from "../../../setup";

describe("ETHDKG: Missing distribute share accusation", () => {
  it("allows accusation of all missing validators after distribute shares Phase", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    // Only validator 0 and 1 distributed shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Distribute Share phase
    await endCurrentPhase(ethdkg);

    // now we can accuse the validator3 who did not participate.
    // keep in mind that when all missing validators are reported,
    // the ethdkg process will restart automatically and emit "RegistrationOpened" event
    expect(await ethdkg.getBadParticipants()).to.equal(0);

    await ethdkg.accuseParticipantDidNotDistributeShares([
      validators4[2].address,
      validators4[3].address,
    ]);

    await expect(await ethdkg.getBadParticipants()).to.equal(2);

    // move to the end of Distribute Share Dispute phase
    await endCurrentPhase(ethdkg);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitKeyShare(
          validators4[0].keyShareG1,
          validators4[0].keyShareG1CorrectnessProof,
          validators4[0].keyShareG2
        )
    ).to.be.revertedWith(
      "ETHDKG: cannot participate on key share submission phase"
    );
  });

  it("allows accusation of some missing validators after distribute shares Phase", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    // Only validator 0 and 1 distributed shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Distribute Share phase
    await endCurrentPhase(ethdkg);

    // now we can accuse the validator2 and 3 who did not participate.
    expect(await ethdkg.getBadParticipants()).to.equal(0);

    await ethdkg.accuseParticipantDidNotDistributeShares([
      validators4[2].address,
    ]);
    expect(await ethdkg.getBadParticipants()).to.equal(1);

    await ethdkg.accuseParticipantDidNotDistributeShares([
      validators4[3].address,
    ]);
    expect(await ethdkg.getBadParticipants()).to.equal(2);

    // move to the end of Distribute Share Dispute phase
    await endCurrentPhase(ethdkg);

    // user tries to go to the next phase
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitKeyShare(
          validators4[0].keyShareG1,
          validators4[0].keyShareG1CorrectnessProof,
          validators4[0].keyShareG2
        )
    ).to.be.revertedWith(
      "ETHDKG: cannot participate on key share submission phase"
    );
  });

  it("do not allow validators to proceed to the next phase if not all validators distributed their shares", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    // Only validator 0 and 1 distributed shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Distribute Share phase
    await endCurrentPhase(ethdkg);

    // move to the end of Distribute Share Dispute phase
    await endCurrentPhase(ethdkg);

    // valid user tries to go to the next phase
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitKeyShare(
          validators4[0].keyShareG1,
          validators4[0].keyShareG1CorrectnessProof,
          validators4[0].keyShareG2
        )
    ).to.be.revertedWith(
      "ETHDKG: cannot participate on key share submission phase"
    );
  });

  // MISSING REGISTRATION ACCUSATION TESTS

  it("won't let not-distributed shares accusations to take place while ETHDKG Distribute Share Phase is open", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    // Only validator 0 and 1 distributed shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    await expect(
      ethdkg.accuseParticipantDidNotDistributeShares([validators4[2].address])
    ).to.be.revertedWith(
      "ETHDKG: should be in post-ShareDistribution accusation phase!"
    );
  });

  it("should not allow validators who did not distributed shares in time to distribute on the accusation phase", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    // Only validator 0 and 1 distributed shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Distribute Share phase
    await endCurrentPhase(ethdkg);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[2].address))
        .distributeShares(
          validators4[2].encryptedShares,
          validators4[2].commitments
        )
    ).to.be.revertedWith("ETHDKG: cannot participate on this phase");
  });

  it("should not allow validators who did not distributed shares in time to submit Key shares", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    // Only validator 0 and 1 distributed shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Distribute Share phase
    await endCurrentPhase(ethdkg);

    // move to the end of Distribute Share Dispute phase
    await endCurrentPhase(ethdkg);

    // valid user tries to go to the next phase
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .submitKeyShare(
          validators4[0].keyShareG1,
          validators4[0].keyShareG1CorrectnessProof,
          validators4[0].keyShareG2
        )
    ).to.be.revertedWith(
      "ETHDKG: cannot participate on key share submission phase"
    );

    // non-participant user tries to go to the next phase
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[3].address))
        .submitKeyShare(
          validators4[0].keyShareG1,
          validators4[0].keyShareG1CorrectnessProof,
          validators4[0].keyShareG2
        )
    ).to.be.revertedWith(
      "ETHDKG: cannot participate on key share submission phase"
    );
  });

  it("should not allow accusation of not distributing shares of validators that distributed shares", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    //Only validator 0 and 1 distributed shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Distribute Share phase
    await endCurrentPhase(ethdkg);

    // now we can accuse the validator2 and 3 who did not participate.
    expect(await ethdkg.getBadParticipants()).to.equal(0);

    await expect(
      ethdkg.accuseParticipantDidNotDistributeShares([validators4[0].address])
    ).to.be.revertedWith(
      "ETHDKG: Dispute failed! Supposed dishonest participant distributed its share in this ETHDKG round!"
    );

    expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow accusation of not distributing shares for non-validators", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    //Only validator 0 and 1 distributed shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Distribute Share phase
    await endCurrentPhase(ethdkg);

    // now we can accuse the validator2 and 3 who did not participate.
    expect(await ethdkg.getBadParticipants()).to.equal(0);

    await expect(
      ethdkg.accuseParticipantDidNotDistributeShares([
        "0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac",
      ])
    ).to.be.revertedWith("ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!");

    expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow not distributed shares accusations after accusation window has finished", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    //Only validator 0 and 1 distributed shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Distribute Share phase
    await endCurrentPhase(ethdkg);

    // now we can accuse the validator2 and 3 who did not participate.
    expect(await ethdkg.getBadParticipants()).to.equal(0);

    await endCurrentAccusationPhase(ethdkg);

    await expect(
      ethdkg.accuseParticipantDidNotDistributeShares([validators4[2].address])
    ).to.be.revertedWith(
      "ETHDKG: should be in post-ShareDistribution accusation phase!"
    );
  });

  it("should not allow accusing a user that distributed the shares in the middle of the ones that did not", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    // Only validator 0 and 1 distributed shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Distribute Share phase
    await endCurrentPhase(ethdkg);

    expect(await ethdkg.getBadParticipants()).to.equal(0);

    await expect(
      ethdkg.accuseParticipantDidNotDistributeShares([
        validators4[2].address,
        validators4[3].address,
        validators4[0].address,
      ])
    ).to.be.revertedWith(
      "ETHDKG: Dispute failed! Supposed dishonest participant distributed its share in this ETHDKG round!"
    );

    expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow double accusation of a user that did not share his shares", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    // Only validator 0 and 1 distributed shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of Distribute Share phase
    await endCurrentPhase(ethdkg);

    expect(await ethdkg.getBadParticipants()).to.equal(0);

    await expect(
      ethdkg.accuseParticipantDidNotDistributeShares([validators4[2].address])
    );

    await expect(
      ethdkg.accuseParticipantDidNotDistributeShares([validators4[2].address])
    ).to.be.revertedWith("ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!");

    expect(await ethdkg.getBadParticipants()).to.equal(1);
  });
});
