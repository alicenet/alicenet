import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";
import {
  ETHDKG,
  ValidatorPool,
  ValidatorPoolMock,
} from "../../typechain-types";
import { validators10 } from "./assets/10-validators-successful-case";
import { validators4 } from "./assets/4-validators-successful-case";
import {
  completeETHDKGRound,
  expect,
  getInfoForIncorrectPhaseCustomError,
  Phase,
  registerValidators,
} from "./setup";

function deployFixture() {
  return completeETHDKGRound(validators10);
}

describe("ETHDKG: Complete an ETHDKG Round and change validators", () => {
  let ethdkg: ETHDKG;
  let validatorPool: ValidatorPool | ValidatorPoolMock;
  let expectedNonce: number;

  beforeEach(async function () {
    [ethdkg, validatorPool, expectedNonce] = await loadFixture(deployFixture);
  });

  it("completes ETHDKG with 10 validators then change to 4 validators", async function () {
    expect(expectedNonce).eq(1);
    expect(await ethdkg.getMasterPublicKey()).to.be.deep.equal(
      validators10[0].mpk
    );
    const expectedHash1 = ethers.utils.solidityKeccak256(
      ["uint256", "uint256", "uint256", "uint256"],
      [...validators10[0].mpk]
    );
    const expectedHash2 = ethers.utils.solidityKeccak256(
      ["uint256", "uint256", "uint256", "uint256"],
      [...validators4[0].mpk]
    );
    const randomHash = ethers.utils.solidityKeccak256(
      ["uint256"],
      [validators4[0].mpk[0]]
    );
    expect(await ethdkg.getMasterPublicKeyHash()).to.be.equal(expectedHash1);
    expect(await ethdkg.isValidMasterPublicKey(expectedHash1)).to.be.equal(
      true
    );
    // the second mpk should be false since the we didn't run ethdkg with the new validators
    expect(await ethdkg.isValidMasterPublicKey(expectedHash2)).to.be.equal(
      false
    );
    // random mpk should be false
    expect(await ethdkg.isValidMasterPublicKey(randomHash)).to.be.equal(false);
    await validatorPool.unregisterAllValidators();
    [, , expectedNonce] = await completeETHDKGRound(validators4, {
      ethdkg,
      validatorPool,
    });
    expect(expectedNonce).eq(2);
    expect(await ethdkg.isValidMasterPublicKey(expectedHash1)).to.be.equal(
      true
    );
    // the second mpk should be true now
    expect(await ethdkg.isValidMasterPublicKey(expectedHash2)).to.be.equal(
      true
    );
    // random mpk should be false
    expect(await ethdkg.isValidMasterPublicKey(randomHash)).to.be.equal(false);
  });

  it("completes ETHDKG with 10 validators then a validator try to register without registration open", async function () {
    const txPromise = registerValidators(
      ethdkg,
      validatorPool,
      validators10,
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
});
