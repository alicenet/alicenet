import { assert, expect } from "chai";
import { MultipleProposalAccusation } from "../../typechain-types";
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

  let accusation: MultipleProposalAccusation;

  beforeEach(async function () {
    fixture = await getFixture(true, true);

    accusation = fixture.multipleProposalAccusation;
  });

  describe("Accusation", async () => {
    const expectedAddress = "0xe17aB3698B3e93bCE3303Fd5b2E5dB8E376E1cb9";
    it("deploys to expected address", async function () {
      const address = accusation.address;

      assert.equal(address, expectedAddress);
    });
  });

  describe("AccuseMultipleProposal:", async () => {
    it("returns signer when valid", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } = generateSigAndPClaims1();

      const signerAccount0 = "0x38e959391dD8598aE80d5d6D114a7822A09d313A";

      await addValidators(fixture.validatorPool, [signerAccount0]);

      const signer = await accusation.AccuseMultipleProposal(
        sig0,
        pClaims0,
        sig1,
        pClaims1
      );

      assert.equal(signer, signerAccount0);
    });

    it("reverts when signer is not valid", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } = generateSigAndPClaims1();

      await expect(
        accusation.AccuseMultipleProposal(sig0, pClaims0, sig1, pClaims1)
      ).to.be.revertedWith(
        "Accusations: the signer of these proposals is not a valid validator!"
      );
    });

    it("reverts when duplicate data for pClaims0 and pClaims1", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();

      await expect(
        accusation.AccuseMultipleProposal(sig0, pClaims0, sig0, pClaims0)
      ).to.be.revertedWith("Accusations: the PClaims are equal!");
    });

    it("reverts when proposals have different signature", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1 } = generateSigAndPClaims1();

      await expect(
        accusation.AccuseMultipleProposal(sig0, pClaims0, sig1, pClaims0)
      ).to.be.revertedWith(
        "Accusations: the signers of the proposals should be the same"
      );
    });

    it("reverts when proposals have different block height", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } =
        generateSigAndPClaimsDifferentHeight();

      await expect(
        accusation.AccuseMultipleProposal(sig0, pClaims0, sig1, pClaims1)
      ).to.be.revertedWith(
        "Accusations: the block heights between the proposals are different!"
      );
    });

    it("reverts when proposals have different round", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } =
        generateSigAndPClaimsDifferentRound();

      await expect(
        accusation.AccuseMultipleProposal(sig0, pClaims0, sig1, pClaims1)
      ).to.be.revertedWith(
        "Accusations: the round between the proposals are different!"
      );
    });

    it("reverts when proposals have different chain id", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } =
        generateSigAndPClaimsDifferentChainId();

      await expect(
        accusation.AccuseMultipleProposal(sig0, pClaims0, sig1, pClaims1)
      ).to.be.revertedWith(
        "Accusations: the chainId between the proposals are different!"
      );
    });
  });
});
