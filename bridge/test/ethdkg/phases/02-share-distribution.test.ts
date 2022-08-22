import { BigNumber } from "ethers";
import { ethers, expect } from "hardhat";
import { getFixture, getValidatorEthAccount } from "../../setup";
import { validators4 } from "../assets/4-validators-successful-case";
import {
  addValidators,
  assertEventSharesDistributed,
  distributeValidatorsShares,
  getInfoForIncorrectPhaseCustomError,
  initializeETHDKG,
  Phase,
  registerValidators,
  startAtDistributeShares,
} from "../setup";

describe("ETHDKG: Distribute Shares", () => {
  it("does not let distribute shares before Distribute Share Phase is open", async function () {
    const { ethdkg, validatorPool } = await getFixture(true);

    const expectedNonce = 1;

    // add validators
    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register only validator 0, so the registration phase hasn't finished yet
    await registerValidators(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce
    );
    const txPromise = distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce
    );
    const [
      ethDKGPhases,
      ,
      expectedBlockNumber,
      expectedCurrentPhase,
      phaseStartBlock,
      phaseLength,
    ] = await getInfoForIncorrectPhaseCustomError(txPromise, ethdkg);
    // distribute shares before the time
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

  it("does not let non-validators to distribute shares", async function () {
    const [ethdkg, ,] = await startAtDistributeShares(validators4);
    const validatorAddress = "0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac";
    // try to distribute shares with a non validator address
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validatorAddress))
        .distributeShares(
          [BigNumber.from("0")],
          [[BigNumber.from("0"), BigNumber.from("0")]]
        )
    )
      .to.be.revertedWithCustomError(ethdkg, "OnlyValidatorsAllowed")
      .withArgs(validatorAddress);
  });

  it("does not let validator to distribute shares more than once", async function () {
    const [ethdkg, validatorPool, expectedNonce] =
      await startAtDistributeShares(validators4);

    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce
    );
    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );
    // distribute shares before the time
    await expect(
      distributeValidatorsShares(
        ethdkg,
        validatorPool,
        validators4.slice(0, 1),
        expectedNonce
      )
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ParticipantDistributedSharesInRound`
      )
      .withArgs(ethers.utils.getAddress(validators4[0].address));
  });

  it("does not let validator send empty commitments or encrypted shares", async function () {
    const [ethdkg, , expectedNonce] = await startAtDistributeShares(
      validators4
    );

    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );
    // distribute shares with empty state
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .distributeShares([BigNumber.from("0")], validators4[0].commitments)
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `InvalidEncryptedSharesAmount`
      )
      .withArgs(1, 3);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .distributeShares(validators4[0].encryptedShares, [
          [BigNumber.from("0"), BigNumber.from("0")],
        ])
    )
      .to.be.revertedWithCustomError(ethDKGPhases, `InvalidCommitmentsAmount`)
      .withArgs(1, 3);

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .distributeShares(validators4[0].encryptedShares, [
          [BigNumber.from("0"), BigNumber.from("0")],
          [BigNumber.from("0"), BigNumber.from("0")],
          [BigNumber.from("0"), BigNumber.from("0")],
        ])
    ).to.be.revertedWithCustomError(ethDKGPhases, `CommitmentNotOnCurve`);

    // the user can send empty encrypted shares on this phase, the accusation window will be
    // handling this!
    const tx = await ethdkg
      .connect(await getValidatorEthAccount(validators4[0].address))
      .distributeShares(
        [BigNumber.from("0"), BigNumber.from("0"), BigNumber.from("0")],
        validators4[0].commitments
      );
    await assertEventSharesDistributed(
      tx,
      validators4[0].address,
      1,
      expectedNonce,
      [BigNumber.from("0"), BigNumber.from("0"), BigNumber.from("0")],
      validators4[0].commitments
    );
  });
});
