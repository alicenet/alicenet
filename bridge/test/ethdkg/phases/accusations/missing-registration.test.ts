import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";
import {
  getFixture,
  getReceiptForFailedTransaction,
  getValidatorEthAccount,
} from "../../../setup";
import { validators10 } from "../../assets/10-validators-successful-case";
import { validators4 } from "../../assets/4-validators-successful-case";
import {
  addValidators,
  assertETHDKGPhase,
  endCurrentAccusationPhase,
  endCurrentPhase,
  expect,
  getInfoForIncorrectPhaseCustomError,
  initializeETHDKG,
  Phase,
  registerValidators,
} from "../../setup";

function deployFixture() {
  return getFixture(true);
}

describe("ETHDKG: Missing registration Accusation", () => {
  it("allows accusation of all missing validators after ETHDKG registration [ @skip-on-coverage ]", async function () {
    const { ethdkg, validatorPool } = await loadFixture(deployFixture);

    const expectedNonce = 1;

    // add validators

    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register validators 0 to 2. validator3 won't register
    await registerValidators(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of RegistrationOpen phase
    await endCurrentPhase(ethdkg);

    // now we can accuse the validator3 who did not participate.
    // keep in mind that when all missing validators are accused,
    // the ethdkg process will restart automatically
    // if there are enough validators registered (>=4 _minValidators)
    expect(await ethdkg.getBadParticipants()).to.equal(0);

    await expect(
      ethdkg.accuseParticipantNotRegistered([
        validators4[2].address,
        validators4[3].address,
      ])
    );

    expect(await ethdkg.getBadParticipants()).to.equal(2);
    await assertETHDKGPhase(ethdkg, Phase.RegistrationOpen);
  });

  it("allows accusation of some missing validators after ETHDKG registration [ @skip-on-coverage ]", async function () {
    const { ethdkg, validatorPool } = await loadFixture(deployFixture);

    const expectedNonce = 1;

    // add validators

    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register validators 0 to 1. validator2 and 3 won't register
    await registerValidators(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of RegistrationOpen phase
    await endCurrentPhase(ethdkg);

    // now we can accuse the validator2 and 3 who did not participate.
    // keep in mind that when all missing validators are reported,
    // the ethdkg process will restart automatically and emit "RegistrationOpened" event if #validators >= 4
    expect(await ethdkg.getBadParticipants()).to.equal(0);

    await ethdkg.accuseParticipantNotRegistered([validators4[2].address]);
    expect(await ethdkg.getBadParticipants()).to.equal(1);

    await ethdkg.accuseParticipantNotRegistered([validators4[3].address]);

    expect(await ethdkg.getBadParticipants()).to.equal(2);
    await assertETHDKGPhase(ethdkg, Phase.RegistrationOpen);
  });

  // MISSING REGISTRATION ACCUSATION TESTS

  it("won't let non-registration accusations to take place while ETHDKG registration is open [ @skip-on-coverage ]", async function () {
    const { ethdkg, validatorPool } = await loadFixture(deployFixture);

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    const expectedNonce = 1;

    // add validators

    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register validators 0 to 1. validator2 and 3 won't register
    await registerValidators(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    const txPromise = ethdkg.accuseParticipantNotRegistered([
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
          Phase.RegistrationOpen,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);
    expect(await ethdkg.getBadParticipants()).to.equal(0);
    await assertETHDKGPhase(ethdkg, Phase.RegistrationOpen);
  });

  it("should not allow validators to proceed to next phase if 2 out of 4 did not register and the phase has finished [ @skip-on-coverage ]", async function () {
    const { ethdkg, validatorPool } = await loadFixture(deployFixture);

    const expectedNonce = 1;

    // add validators

    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register validators 0 to 1. validator2 and 3 won't register
    await registerValidators(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of RegistrationOpen phase
    await endCurrentPhase(ethdkg);

    // validator0 should not be able to distribute shares
    const signer0 = await getValidatorEthAccount(validators4[0].address);
    const txPromise = ethdkg
      .connect(signer0)
      .distributeShares(
        validators4[0].encryptedShares,
        validators4[0].commitments
      );
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
          Phase.ShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
      ]);
  });

  it("should not allow validators who did not register in time to register on the accusation phase [ @skip-on-coverage ]", async function () {
    const { ethdkg, validatorPool } = await loadFixture(deployFixture);

    const expectedNonce = 1;

    // add validators

    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register validators 0 to 1. validator2 and 3 won't register
    await registerValidators(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of RegistrationOpen phase
    await endCurrentPhase(ethdkg);

    const signer2 = await getValidatorEthAccount(validators4[2].address);

    const txPromise = ethdkg
      .connect(signer2)
      .register(validators4[2].aliceNetPublicKey);
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
          Phase.RegistrationOpen,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
      ]);
  });

  it("should not allow validators who did not register in time to distribute shares [ @skip-on-coverage ]", async function () {
    const { ethdkg, validatorPool } = await loadFixture(deployFixture);

    const expectedNonce = 1;

    // add validators

    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register validators 0 to 1. validator2 and 3 won't register
    await registerValidators(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of RegistrationOpen phase
    await endCurrentPhase(ethdkg);

    // validator2 should not be able to distribute shares
    const signer2 = await getValidatorEthAccount(validators4[2].address);

    const txPromise = ethdkg
      .connect(signer2)
      .distributeShares(
        validators4[0].encryptedShares,
        validators4[0].commitments
      );
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
          Phase.ShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
      ]);
  });

  it("should not allow accusation of validators that registered in ETHDKG [ @skip-on-coverage ]", async function () {
    const { ethdkg, validatorPool } = await loadFixture(deployFixture);

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    const expectedNonce = 1;

    // add validators

    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register validators 0 to 1. validator2 and 3 won't register
    await registerValidators(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of RegistrationOpen phase
    await endCurrentPhase(ethdkg);

    // accuse a participant validator
    await expect(
      ethdkg.accuseParticipantNotRegistered([validators4[0].address])
    )
      .to.be.revertedWithCustomError(
        ETHDKGAccusations,
        `AccusedParticipatingInRound`
      )
      .withArgs(ethers.utils.getAddress(validators4[0].address));

    expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow accusation of non-existent validators in ETHDKG [ @skip-on-coverage ]", async function () {
    const { ethdkg, validatorPool } = await loadFixture(deployFixture);

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    const expectedNonce = 1;

    // add validators

    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register validators 0 to 1. validator2 and 3 won't register
    await registerValidators(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of RegistrationOpen phase
    await endCurrentPhase(ethdkg);

    // accuse a non-existent validator
    const accusedAddress = "0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac";
    await expect(ethdkg.accuseParticipantNotRegistered([accusedAddress]))
      .to.be.revertedWithCustomError(ETHDKGAccusations, `AccusedNotValidator`)
      .withArgs(ethers.utils.getAddress(accusedAddress));

    expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow accusations after the accusation window [ @skip-on-coverage ]", async function () {
    const { ethdkg, validatorPool } = await loadFixture(deployFixture);

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    const expectedNonce = 1;

    // add validators

    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register validators 0 to 1. validator2 and 3 won't register
    await registerValidators(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of RegistrationOpen phase
    await endCurrentPhase(ethdkg);

    // move to the end of RegistrationAccusation phase
    await endCurrentAccusationPhase(ethdkg);

    const txPromise = ethdkg.accuseParticipantNotRegistered([
      validators4[2].address,
    ]);
    const expectedBlockNumber = (
      await getReceiptForFailedTransaction(txPromise)
    ).blockNumber;
    const expectedCurrentPhase = await ethdkg.getETHDKGPhase();
    const phaseStartBlock = await ethdkg.getPhaseStartBlock();
    const phaseLength = await ethdkg.getPhaseLength();

    // accuse a non-participant validator
    await expect(txPromise)
      .to.be.revertedWithCustomError(ETHDKGAccusations, `IncorrectPhase`)
      .withArgs(expectedCurrentPhase, expectedBlockNumber, [
        [
          Phase.RegistrationOpen,
          phaseStartBlock.add(phaseLength),
          phaseStartBlock.add(phaseLength.mul(2)),
        ],
      ]);

    expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not allow accusations of non-existent validators along with existent [ @skip-on-coverage ]", async function () {
    const { ethdkg, validatorPool } = await loadFixture(deployFixture);

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    const expectedNonce = 1;

    // add validators

    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register validators 0 to 1. validator2 and 3 won't register
    await registerValidators(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of RegistrationOpen phase
    await endCurrentPhase(ethdkg);

    // accuse a participant validator
    const accusedAddress = "0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac";
    await expect(
      ethdkg.accuseParticipantNotRegistered([
        validators4[2].address,
        validators4[3].address,
        accusedAddress,
      ])
    )
      .to.be.revertedWithCustomError(ETHDKGAccusations, `AccusedNotValidator`)
      .withArgs(ethers.utils.getAddress(accusedAddress));

    expect(await ethdkg.getBadParticipants()).to.equal(0);
  });

  it("should not move to ShareDistribution phase when only 2 out of 4 validators have participated [ @skip-on-coverage ]", async function () {
    // Accuse 1 participant that didn't participate and wait the window to expire and try to go to the next phase after accusation

    const { ethdkg, validatorPool } = await loadFixture(deployFixture);

    const expectedNonce = 1;

    // add validators

    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register validators 0 to 1. validator 2 and 3 won't register
    await registerValidators(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of RegistrationOpen phase
    await endCurrentPhase(ethdkg);

    // accuse a non-registered validator
    await ethdkg.accuseParticipantNotRegistered([validators4[2].address]);

    expect(await ethdkg.getBadParticipants()).to.equal(1);

    // move to the end of RegistrationAccusation phase
    await endCurrentAccusationPhase(ethdkg);

    // try to move into Distribute Shares phase
    const txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .distributeShares(
        validators4[0].encryptedShares,
        validators4[0].commitments
      );
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
          Phase.ShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
      ]);

    await assertETHDKGPhase(ethdkg, Phase.RegistrationOpen);
  });

  it("should move to ShareDistribution phase when all non-participant validators have been accused and #validators >= _minValidators [ @skip-on-coverage ]", async function () {
    const { ethdkg, validatorPool } = await loadFixture(deployFixture);

    const expectedNonce = 1;

    // validator 11
    const validator11 = "0x23EA3Bad9115d436190851cF4C49C1032fA7579A";

    // add validators

    await addValidators(validatorPool, validators10);

    // add validator 11
    await validatorPool.registerValidators([validator11], []);
    expect(await validatorPool.isValidator(validator11)).to.equal(true);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register all 10 validators
    await registerValidators(
      ethdkg,
      validatorPool,
      validators10,
      expectedNonce
    );

    // move to the end of RegistrationOpen phase
    await endCurrentPhase(ethdkg);

    expect(await ethdkg.getBadParticipants()).to.equal(0);
    // accuse non-participant validator 11
    await ethdkg.accuseParticipantNotRegistered([validator11]);

    expect(await ethdkg.getBadParticipants()).to.equal(1);

    // move to the end of MissingRegistrationAccusation phase
    await endCurrentPhase(ethdkg);
    await assertETHDKGPhase(ethdkg, Phase.RegistrationOpen);

    // try distributing shares
    const txPromise = ethdkg
      .connect(await getValidatorEthAccount(validators10[0].address))
      .distributeShares(
        validators10[0].encryptedShares,
        validators10[0].commitments
      );
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
          Phase.ShareDistribution,
          phaseStartBlock,
          phaseStartBlock.add(phaseLength),
        ],
      ]);

    await assertETHDKGPhase(ethdkg, Phase.RegistrationOpen);
  });

  it("should not allow double accusation for missing registration [ @skip-on-coverage ]", async function () {
    const { ethdkg, validatorPool } = await loadFixture(deployFixture);

    const ETHDKGAccusations = await ethers.getContractAt(
      "ETHDKGAccusations",
      ethdkg.address
    );
    const expectedNonce = 1;

    // add validators

    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register validators 0 to 1. validator2 and 3 won't register
    await registerValidators(
      ethdkg,
      validatorPool,
      validators4.slice(0, 2),
      expectedNonce
    );

    // move to the end of RegistrationOpen phase
    await endCurrentPhase(ethdkg);

    // accuse non-participant validator 2, twice
    await expect(
      ethdkg.accuseParticipantNotRegistered([
        validators4[2].address,
        validators4[2].address,
      ])
    )
      .to.be.revertedWithCustomError(ETHDKGAccusations, `AccusedNotValidator`)
      .withArgs(ethers.utils.getAddress(validators4[2].address));

    await assertETHDKGPhase(ethdkg, Phase.RegistrationOpen);
  });
});
