import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { BigNumberish } from "ethers";
import { ethers, expect } from "hardhat";
import { getValidatorEthAccount } from "../../setup";
import { validators4 } from "../assets/4-validators-successful-case";
import {
  assertETHDKGPhase,
  assertEventMPKSet,
  getInfoForIncorrectPhaseCustomError,
  Phase,
  startAtMPKSubmission,
  startAtSubmitKeyShares,
  submitValidatorsKeyShares,
  waitNextPhaseStartDelay,
} from "../setup";

describe("ETHDKG: Submit Master Public Key", () => {
  describe("when not in MasterPublicKeySubmission phase", () => {
    it("should not allow submission of master public key when not in MPKSubmission phase [ @skip-on-coverage ]", async () => {
      const [ethdkg, validatorPool, expectedNonce] =
        await startAtSubmitKeyShares(validators4);
      // distribute shares for all but 1 validators
      await submitValidatorsKeyShares(
        ethdkg,
        validatorPool,
        validators4.slice(0, 3),
        expectedNonce
      );

      await assertETHDKGPhase(ethdkg, Phase.KeyShareSubmission);
      await waitNextPhaseStartDelay(ethdkg);
      await assertETHDKGPhase(ethdkg, Phase.KeyShareSubmission);

      const txPromise = ethdkg
        .connect(await getValidatorEthAccount(validators4[3].address))
        .submitMasterPublicKey(validators4[3].mpk);
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
  });

  describe("when in MasterPublicKeySubmission phase", () => {
    async function deployFixture() {
      return startAtMPKSubmission(validators4);
    }
    it("should allow submission of master public key by a non-validator [ @skip-on-coverage ]", async () => {
      const [ethdkg, , expectedNonce] = await loadFixture(deployFixture);

      // non-validator user tries to submit the Master Public key
      const validator11 = "0x23EA3Bad9115d436190851cF4C49C1032fA7579A";
      const val11MPK: [BigNumberish, BigNumberish, BigNumberish, BigNumberish] =
        validators4[0].mpk;

      const tx = await ethdkg
        .connect(await getValidatorEthAccount(validator11))
        .submitMasterPublicKey(val11MPK);

      await assertEventMPKSet(tx, expectedNonce, val11MPK);
    });

    it("should not allow submission of master public key more than once [ @skip-on-coverage ]", async () => {
      const [ethdkg, , expectedNonce] = await loadFixture(deployFixture);

      // non-validator user tries to submit the Master Public key
      const validator11 = "0x23EA3Bad9115d436190851cF4C49C1032fA7579A";
      const val11MPK: [BigNumberish, BigNumberish, BigNumberish, BigNumberish] =
        validators4[0].mpk;

      const tx = await ethdkg
        .connect(await getValidatorEthAccount(validator11))
        .submitMasterPublicKey(val11MPK);

      await assertEventMPKSet(tx, expectedNonce, val11MPK);

      const txPromise = ethdkg
        .connect(await getValidatorEthAccount(validator11))
        .submitMasterPublicKey(val11MPK);
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

    it("should not allow submission of empty master public key [ @skip-on-coverage ]", async () => {
      const [ethdkg, ,] = await loadFixture(deployFixture);

      // empty MPK
      const mpk: [BigNumberish, BigNumberish, BigNumberish, BigNumberish] = [
        "0",
        "0",
        "0",
        "0",
      ];

      const ethDKGPhases = await ethers.getContractAt(
        "ETHDKGPhases",
        ethdkg.address
      );
      await expect(
        ethdkg
          .connect(await getValidatorEthAccount(validators4[0].address))
          .submitMasterPublicKey(mpk)
      ).to.be.revertedWithCustomError(
        ethDKGPhases,
        `MasterPublicKeyPairingCheckFailure`
      );
    });
  });
});
