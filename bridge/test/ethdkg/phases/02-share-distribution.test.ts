import { validators4 } from "../assets/4-validators-successful-case";
import { ethers } from "hardhat";
import { BigNumber } from "ethers";
import {
  addValidators,
  initializeETHDKG,
  registerValidators,
  assertEventSharesDistributed,
  distributeValidatorsShares,
  startAtDistributeShares,
  expect,
} from "../setup";
import { getFixture, getValidatorEthAccount } from "../../setup";

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

    // distribute shares before the time
    await expect(
      distributeValidatorsShares(
        ethdkg,
        validatorPool,
        validators4.slice(0, 1),
        expectedNonce
      )
    ).to.be.rejectedWith("ETHDKG: cannot participate on this phase");
  });

  it("does not let non-validators to distribute shares", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    // try to distribute shares with a non validator address
    await expect(
      ethdkg
        .connect(
          await getValidatorEthAccount("0x26D3D8Ab74D62C26f1ACc220dA1646411c9880Ac")
        )
        .distributeShares(
          [BigNumber.from("0")],
          [[BigNumber.from("0"), BigNumber.from("0")]]
        )
    ).to.be.revertedWith("ETHDKG: Only validators allowed!");
  });

  it("does not let validator to distribute shares more than once", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    await distributeValidatorsShares(
      ethdkg,
      validatorPool,
      validators4.slice(0, 1),
      expectedNonce
    );

    // distribute shares before the time
    await expect(
      distributeValidatorsShares(
        ethdkg,
        validatorPool,
        validators4.slice(0, 1),
        expectedNonce
      )
    ).to.be.rejectedWith(
      "ETHDKG: Participant already distributed shares this ETHDKG round!"
    );
  });

  it("does not let validator send empty commitments or encrypted shares", async function () {
    let [ethdkg, validatorPool, expectedNonce] = await startAtDistributeShares(
      validators4
    );

    // distribute shares with empty data
    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .distributeShares([BigNumber.from("0")], validators4[0].commitments)
    ).to.be.rejectedWith(
      "ETHDKG: Share distribution failed - invalid number of encrypted shares provided!"
    );

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .distributeShares(validators4[0].encryptedShares, [
          [BigNumber.from("0"), BigNumber.from("0")],
        ])
    ).to.be.rejectedWith(
      "ETHDKG: Key sharing failed - invalid number of commitments provided!"
    );

    await expect(
      ethdkg
        .connect(await getValidatorEthAccount(validators4[0].address))
        .distributeShares(validators4[0].encryptedShares, [
          [BigNumber.from("0"), BigNumber.from("0")],
          [BigNumber.from("0"), BigNumber.from("0")],
          [BigNumber.from("0"), BigNumber.from("0")],
        ])
    ).to.be.rejectedWith("ETHDKG: Key sharing failed - commitment not on elliptic curve!");

    // the user can send empty encrypted shares on this phase, the accusation window will be
    // handling this!
    let tx = await ethdkg
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
