import { ethers } from "hardhat";
import {
  getReceiptForFailedTransaction,
  getValidatorEthAccount,
} from "../../../setup";
import { validators4 } from "../../assets/4-validators-successful-case";
import {
  distributeValidatorsShares,
  endCurrentAccusationPhase,
  endCurrentPhase,
  expect,
  Phase,
  startAtDistributeShares,
} from "../../setup";

describe("ETHDKG: Missing distribute share accusation", () => {
  it("allows accusation of all missing validators after distribute shares Phase", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);
    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
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
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInKeyshareSubmissionPhase`
      )
      .withArgs(Phase.ShareDistribution);
  });

  it("allows accusation of some missing validators after distribute shares Phase", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);
    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
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
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInKeyshareSubmissionPhase`
      )
      .withArgs(Phase.ShareDistribution);
  });

  it("do not allow validators to proceed to the next phase if not all validators distributed their shares", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);
    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
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
        ),
      `ETHDKGNotInKeyshareSubmissionPhase(1)`
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInKeyshareSubmissionPhase`
      )
      .withArgs(Phase.ShareDistribution);
  });

  // MISSING REGISTRATION ACCUSATION TESTS

  it("won't let not-distributed shares accusations to take place while ETHDKG Distribute Share Phase is open", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    // Only validator 0 and 1 distributed shares
    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    const txPromise = ethdkg.accuseParticipantDidNotDistributeShares([
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
          Phase.ShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);
  });

  it("should not allow validators who did not distributed shares in time to distribute on the accusation phase", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);
    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
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
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInSharedDistributionPhase`
      )
      .withArgs(Phase.ShareDistribution);
  });

  it("should not allow validators who did not distributed shares in time to submit Key shares", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);
    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
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
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInKeyshareSubmissionPhase`
      )
      .withArgs(Phase.ShareDistribution);

    // non-participant user tries to go to the next phase
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[3].address))
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
      .withArgs(Phase.ShareDistribution);
  });

  it("should not allow accusation of not distributing shares of validators that distributed shares", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
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

    await expect(
      ethdkg.accuseParticipantDidNotDistributeShares([validators4[0].address])
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `AccusedDistributedSharesInRound`
      )
      .withArgs(ethers.utils.getAddress(validators4[0].address));

    expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow accusation of not distributing shares for non-validators", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
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

    const accusedAddress = "0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac";
    await expect(
      ethdkg.accuseParticipantDidNotDistributeShares([accusedAddress])
    )
      .to.be.revertedWithCustomError(ETHDKGAccusations, `AccusedNotValidator`)
      .withArgs(ethers.utils.getAddress(accusedAddress));

    expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow not distributed shares accusations after accusation window has finished", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
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

    await endCurrentAccusationPhase(ethdkg);

    const txPromise = ethdkg.accuseParticipantDidNotDistributeShares([
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
          Phase.ShareDistribution,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);
  });

  it("should not allow accusing a user that distributed the shares in the middle of the ones that did not", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
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
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `AccusedDistributedSharesInRound`
      )
      .withArgs(ethers.utils.getAddress(validators4[0].address));

    expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow double accusation of a user that did not share his shares", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
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
    )
      .to.be.revertedWithCustomError(ETHDKGAccusations, `AccusedNotValidator`)
      .withArgs(ethers.utils.getAddress(validators4[2].address));

    expect(await ethdkg.getBadParticipants()).to.equal(1);
  });
});
