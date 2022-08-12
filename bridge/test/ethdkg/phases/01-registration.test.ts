import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { getFixture, getValidatorEthAccount } from "../../setup";
import { validators4 } from "../assets/4-validators-successful-case";
import { addValidators, expect, initializeETHDKG, Phase } from "../setup";

describe("ETHDKG: Registration Open", () => {
  it("does not let registrations before ETHDKG Registration is open", async function () {
    const { ethdkg, validatorPool } = await getFixture(true);

    // add validators
    await addValidators(validatorPool, validators4);

    // for this test, ETHDKG is not started
    // register validator0
    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .register(validators4[0].aliceNetPublicKey)
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ETHDKGNotInRegistrationPhase`
      )
      .withArgs(Phase.RegistrationOpen);
  });

  it("does not let validators to register more than once", async function () {
    const { ethdkg, validatorPool } = await getFixture(true);

    // add validators
    await addValidators(validatorPool, validators4);
    await initializeETHDKG(ethdkg, validatorPool);

    // register one validator
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .register(validators4[0].aliceNetPublicKey)
    ).to.emit(ethdkg, "AddressRegistered");

    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );
    const currentNonce = await ethdkg.getNonce();
    const participantData = await ethdkg.getParticipantInternalState(
      validators4[0].address
    );
    // register that same validator again
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .register(validators4[0].aliceNetPublicKey)
    )
      .to.be.revertedWithCustomError(
        ethDKGPhases,
        `ParticipantParticipatingInRound`
      )
      .withArgs(
        ethers.utils.getAddress(validators4[0].address),
        participantData.nonce,
        currentNonce.sub(BigNumber.from(1))
      );
  });

  it("does not let validators to register with an incorrect key", async function () {
    const { ethdkg, validatorPool } = await getFixture(true);

    // add validators
    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // register validator0 with invalid pubkey
    const signer0 = await getValidatorEthAccount(validators4[0].address);

    const ethDKGPhases = await ethers.getContractAt(
      "ETHDKGPhases",
      ethdkg.address
    );

    await expect(
      ethdkg
        .connect(signer0)
        .register([BigNumber.from("0"), BigNumber.from("1")])
    ).to.be.revertedWithCustomError(ethDKGPhases, `PublicKeyZero`);

    await expect(
      ethdkg
        .connect(signer0)
        .register([BigNumber.from("1"), BigNumber.from("0")])
    ).to.be.revertedWithCustomError(ethDKGPhases, `PublicKeyZero`);

    await expect(
      ethdkg
        .connect(signer0)
        .register([BigNumber.from("1"), BigNumber.from("1")])
    ).to.be.revertedWithCustomError(ethDKGPhases, `PublicKeyNotOnCurve`);
  });

  it("does not let non-validators to register", async function () {
    const { ethdkg, validatorPool } = await getFixture(true);

    // add validators
    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    const validatorAddress = "0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac";
    // try to register with a non validator address
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validatorAddress))
        .register([BigNumber.from("0"), BigNumber.from("0")])
    )
      .to.be.revertedWithCustomError(ethdkg, "OnlyValidatorsAllowed")
      .withArgs(validatorAddress);
  });
});
