import { assert, expect } from "chai";
import { AccusationMultipleProposal } from "../../typechain-types";
import { Fixture, getFixture } from "../setup";
import {
  addValidators,
  generateSigAndPClaims0,
  generateSigAndPClaims1,
  generateSigAndPClaimsDifferentChainId,
  generateSigAndPClaimsDifferentHeight,
  generateSigAndPClaimsDifferentRound,
} from "./accusations-test-helpers";

describe("MultipleProposalAccusation: Tests MultipleProposalAccusation methods", async () => {
  let fixture: Fixture;

  let accusation: AccusationMultipleProposal;

  beforeEach(async function () {
    fixture = await getFixture(true, true);

    accusation = fixture.accusationMultipleProposal;
  });

  describe("accuseMultipleProposal:", async () => {
    it("returns signer when valid", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } = generateSigAndPClaims1();

      const signerAccount0 = "0x38e959391dD8598aE80d5d6D114a7822A09d313A";

      await addValidators(fixture.validatorPool, [signerAccount0]);

      let isValidator = await fixture.validatorPool.isValidator(signerAccount0);
      assert.equal(isValidator, true);

      await (await accusation.accuseMultipleProposal(
        sig0,
        pClaims0,
        sig1,
        pClaims1
      )).wait();

      isValidator = await fixture.validatorPool.isValidator(signerAccount0);
      assert.equal(isValidator, false);
    });

    it("reverts when signer is not valid", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } = generateSigAndPClaims1();

      await expect(
        accusation.accuseMultipleProposal(sig0, pClaims0, sig1, pClaims1)
      ).to.be.revertedWith("814");
    });

    it("reverts when duplicate data for pClaims0 and pClaims1", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();

      await expect(
        accusation.accuseMultipleProposal(sig0, pClaims0, sig0, pClaims0)
      ).to.be.revertedWith("Accusations: the PClaims are equal!");
    });

    it("reverts when proposals have different signature", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1 } = generateSigAndPClaims1();

      await expect(
        accusation.accuseMultipleProposal(sig0, pClaims0, sig1, pClaims0)
      ).to.be.revertedWith(
        "Accusations: the signers of the proposals should be the same"
      );
    });

    it("reverts when proposals have different block height", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } =
        generateSigAndPClaimsDifferentHeight();

      await expect(
        accusation.accuseMultipleProposal(sig0, pClaims0, sig1, pClaims1)
      ).to.be.revertedWith(
        "Accusations: the block heights between the proposals are different!"
      );
    });

    it("reverts when proposals have different round", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } =
        generateSigAndPClaimsDifferentRound();

      await expect(
        accusation.accuseMultipleProposal(sig0, pClaims0, sig1, pClaims1)
      ).to.be.revertedWith(
        "Accusations: the round between the proposals are different!"
      );
    });

    it("reverts when proposals have different chain id", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } =
        generateSigAndPClaimsDifferentChainId();

      await expect(
        accusation.accuseMultipleProposal(sig0, pClaims0, sig1, pClaims1)
      ).to.be.revertedWith(
        "Accusations: the chainId between the proposals are different!"
      );
    });
  });
});
