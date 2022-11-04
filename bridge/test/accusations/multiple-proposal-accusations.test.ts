import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { assert, expect } from "chai";
import { AccusationMultipleProposal } from "../../typechain-types";
import { Fixture, getFixture } from "../setup";
import {
  addValidators,
  generateAccusationID,
  generateSigAndPClaims0,
  generateSigAndPClaims1,
  generateSigAndPClaimsDifferentChainId,
  generateSigAndPClaimsDifferentHeight,
  generateSigAndPClaimsDifferentRound,
} from "./accusations-test-helpers";

describe("AccusationMultipleProposal: Tests AccusationMultipleProposal methods", async () => {
  let fixture: Fixture;

  let accusation: AccusationMultipleProposal;

  function deployFixture() {
    return getFixture(true, true);
  }

  beforeEach(async function () {
    fixture = await loadFixture(deployFixture);

    accusation = fixture.accusationMultipleProposal;
  });

  describe("accuseMultipleProposal:", async () => {
    const signerAccount0 = "0x38e959391dD8598aE80d5d6D114a7822A09d313A";
    const signerAccount1 = "0x727604F7ad3AF6A4dd1479D2bAf75B97F0592cFc";
    it("returns signer when valid", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } = generateSigAndPClaims1();

      await addValidators(fixture.validatorPool, [signerAccount0]);

      await (
        await accusation.accuseMultipleProposal(sig0, pClaims0, sig1, pClaims1)
      ).wait();

      const isValidator = await fixture.validatorPool.isValidator(
        signerAccount0
      );
      assert.equal(isValidator, false);

      const id = generateAccusationID(
        signerAccount0,
        1,
        2,
        1,
        "0x17287210c71008320429d4cce2075373f0b2c5217b507513fe4904fead741aad"
      );
      const isAccused = await accusation.isAccused(id);
      assert.equal(isAccused, true);
    });

    it("reverts when signer is not valid", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } = generateSigAndPClaims1();

      await expect(
        accusation.accuseMultipleProposal(sig0, pClaims0, sig1, pClaims1)
      )
        .to.be.revertedWithCustomError(
          fixture.validatorPool,
          "AddressNotAccusable"
        )
        .withArgs(signerAccount0);
    });

    it("reverts when duplicate data for pClaims0 and pClaims1", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();

      await expect(
        accusation.accuseMultipleProposal(sig0, pClaims0, sig0, pClaims0)
      ).to.be.revertedWithCustomError(accusation, "PClaimsAreEqual");
    });

    it("reverts when proposals have different signature", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1 } = generateSigAndPClaims1();

      await expect(
        accusation.accuseMultipleProposal(sig0, pClaims0, sig1, pClaims0)
      )
        .to.be.revertedWithCustomError(accusation, "SignersDoNotMatch")
        .withArgs(signerAccount0, signerAccount1);
    });

    it("reverts when proposals have different block height", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } =
        generateSigAndPClaimsDifferentHeight();
      const pClaims0Height = 2;
      const pClaims1Height = 3;
      await expect(
        accusation.accuseMultipleProposal(sig0, pClaims0, sig1, pClaims1)
      )
        .to.be.revertedWithCustomError(accusation, "PClaimsHeightsDoNotMatch")
        .withArgs(pClaims0Height, pClaims1Height);
    });

    it("reverts when proposals have different round", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } =
        generateSigAndPClaimsDifferentRound();

      const pClaims0Height = 1;
      const pClaims1Height = 2;
      await expect(
        accusation.accuseMultipleProposal(sig0, pClaims0, sig1, pClaims1)
      )
        .to.be.revertedWithCustomError(accusation, "PClaimsRoundsDoNotMatch")
        .withArgs(pClaims0Height, pClaims1Height);
    });

    it("reverts when proposals have different chain id", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } =
        generateSigAndPClaimsDifferentChainId();

      const pClaims0ChainId = 1;
      const pClaims1ChainId = 11;

      await expect(
        accusation.accuseMultipleProposal(sig0, pClaims0, sig1, pClaims1)
      )
        .to.be.revertedWithCustomError(accusation, "PClaimsChainIdsDoNotMatch")
        .withArgs(pClaims0ChainId, pClaims1ChainId);
    });
  });
});
