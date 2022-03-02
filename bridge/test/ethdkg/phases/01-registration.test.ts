import { validators4 } from "../assets/4-validators-successful-case";
import { ethers } from "hardhat";
import { BigNumber } from "ethers";
import {
  addValidators,
  initializeETHDKG,
  expect,
} from "../setup";
import { getFixture, getValidatorEthAccount } from "../../setup";

describe("ETHDKG: Registration Open", () => {
  it("does not let registrations before ETHDKG Registration is open", async function () {
    const { ethdkg, validatorPool } = await getFixture(true);

    // add validators
    await addValidators(validatorPool, validators4);

    // for this test, ETHDKG is not started
    // register validator0
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .register(validators4[0].madNetPublicKey)
    ).to.be.revertedWith("ETHDKG: Cannot register at the moment");
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
        .register(validators4[0].madNetPublicKey)
    ).to.emit(ethdkg, "AddressRegistered");

    // register that same validator again
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .register(validators4[0].madNetPublicKey)
    ).to.be.revertedWith(
      "ETHDKG: Participant is already participating in this ETHDKG round!"
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
    await expect(
      ethdkg
        .connect(signer0)
        .register([BigNumber.from("0"), BigNumber.from("1")])
    ).to.be.revertedWith("ETHDKG: Registration failed - pubKey should be different from 0!");

    await expect(
      ethdkg
        .connect(signer0)
        .register([BigNumber.from("1"), BigNumber.from("0")])
    ).to.be.revertedWith("ETHDKG: Registration failed - pubKey should be different from 0!");

    await expect(
      ethdkg
        .connect(signer0)
        .register([BigNumber.from("1"), BigNumber.from("1")])
    ).to.be.revertedWith(
      "ETHDKG: Registration failed - public key not on elliptic curve!"
    );
  });

  it("does not let non-validators to register", async function () {
    const { ethdkg, validatorPool } = await getFixture(true);

    // add validators
    await addValidators(validatorPool, validators4);

    // start ETHDKG
    await initializeETHDKG(ethdkg, validatorPool);

    // try to register with a non validator address
    await expect(
      ethdkg
        .connect(
          await getValidatorEthAccount("0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac")
        )
        .register([BigNumber.from("0"), BigNumber.from("0")])
    ).to.be.revertedWith("ETHDKG: Only validators allowed!");
  });
});
