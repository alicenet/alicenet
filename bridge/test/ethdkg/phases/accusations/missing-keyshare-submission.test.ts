import { ethers } from "hardhat";
import {
  getReceiptForFailedTransaction,
  getValidatorEthAccount,
} from "../../../setup";
import { validators4 } from "../../assets/4-validators-successful-case";
import {
  endCurrentAccusationPhase,
  endCurrentPhase,
  expect,
  getInfoForIncorrectPhaseCustomError,
  Phase,
  startAtSubmitKeyShares,
  submitValidatorsKeyShares,
} from "../../setup";

describe("ETHDKG: Accuse participant of not submitting key shares", () => {
  it("allows accusation of all missing validators after Key share phase [ @skip-on-coverage ]", async function () {
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

    const txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .submitMasterPublicKey(validators4[0].mpk);
    const [
      ethDKGPhases,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.MPKSubmission,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
      ]);
  });

  it("allows accusation of some missing validators after Key share phase [ @skip-on-coverage ]", async function () {
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

    const txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .submitMasterPublicKey(validators4[0].mpk);
    const [
      ethDKGPhases,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.MPKSubmission,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
      ]);
  });

  it("do not allow validators to proceed to the next phase if not all validators submitted their key shares [ @skip-on-coverage ]", async function () {
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

    const txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .submitMasterPublicKey(validators4[0].mpk);
    const [
      ethDKGPhases,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.MPKSubmission,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
      ]);
  });

  it("won't let not-distributed shares accusations to take place while ETHDKG Distribute Share Phase is open [ @skip-on-coverage ]", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    // distribute shares only for validators 0 and 1
    await submitValidatorsKeyShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    const txPromise = ethdkg.accuseParticipantDidNotSubmitKeyShares([
      validators4[2].address,
    ]);
    const expectedBlockNumber = (
      await getReceiptForFailedTransaction(txPromise)
    ).blockNumber;
    const expectedCurrentPhase = await ethdkg.getETHDKGPhase();
    const phaseStartBlock = await ethdkg.getPhaseStartBlock();
    const phaseLength = await ethdkg.getPhaseLength();

    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.KeyShareSubmission,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);
  });

  it("should not allow validators who did not submit key shares in time to submit on the accusation phase [ @skip-on-coverage ]", async function () {
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

    let txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[2].address))
      .submitKeyShare(
        validators4[2].keyShareG1,
        validators4[2].keyShareG1CorrectnessProof,
        validators4[2].keyShareG2
      );
    let [
      ethDKGPhases,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.KeyShareSubmission,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
        [
          Phase.DisputeShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    // non-participant user tries to go to the next phase
    txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[3].address))
      .submitMasterPublicKey(validators4[3].mpk);
    [
      ethDKGPhases,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    await expect(txPromise)
      .to.be.revertedWithCustomError(ethDKGPhases, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.MPKSubmission,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
      ]);
  });

  it("should not allow accusation of not submitting key shares of validators submitted their key shares [ @skip-on-coverage ]", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
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
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `AccusedSubmittedSharesInRound`
      )
      .withArgs(ethers.utils.getAddress(validators4[0].address));

    await expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow accusation of not submitting key shares for non-validators [ @skip-on-coverage ]", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
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
    const accusedAddress = "0x23EA3Bad9115d436190851cF4C49C1032fA7579A";
    await expect(
      ethdkg.accuseParticipantDidNotSubmitKeyShares([accusedAddress])
    )
      .to.be.revertedWithCustomError(ETHDKGAccusations, `AccusedNotValidator`)
      .withArgs(ethers.utils.getAddress(accusedAddress));

    await expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow not submitted key shares accusations after accusation window has finished [ @skip-on-coverage ]", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
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

    const txPromise = ethdkg.accuseParticipantDidNotSubmitKeyShares([
      validators4[2].address,
    ]);
    const expectedBlockNumber = (
      await getReceiptForFailedTransaction(txPromise)
    ).blockNumber;
    const expectedCurrentPhase = await ethdkg.getETHDKGPhase();
    const phaseStartBlock = await ethdkg.getPhaseStartBlock();
    const phaseLength = await ethdkg.getPhaseLength();

    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.KeyShareSubmission,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    await expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow accusing a user that submitted the key shares in the middle of the ones that did not [ @skip-on-coverage ]", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
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
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `AccusedSubmittedSharesInRound`
      )
      .withArgs(ethers.utils.getAddress(validators4[0].address));

    await expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow double accusation of a user that did not submit his key shares [ @skip-on-coverage ]", async function () {
    const [ethdkg, validatorPool, expectedNonce] = await startAtSubmitKeyShares(
      validators4
    );

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
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
    )
      .to.be.revertedWithCustomError(ETHDKGAccusations, `AccusedNotValidator`)
      .withArgs(ethers.utils.getAddress(validators4[2].address));

    await expect(await ethdkg.getBadParticipants()).to.equal(1);
  });
});
