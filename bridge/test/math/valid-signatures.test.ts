import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";
import { CryptoLibraryWrapper } from "../../typechain-types";
import { expect } from "../chai-setup";
import { signedData } from "./assets/4-validators-1000-snapshots";

describe("CryptoLibrary: Validate Signature", () => {
  let crypto: CryptoLibraryWrapper;
  const amountRuns = 100;

  async function deployFixture() {
    const crypto = await (
      await ethers.getContractFactory("CryptoLibraryWrapper")
    ).deploy();
    await crypto.deployed();
    return { crypto };
  }

  before(async () => {
    ({ crypto } = await loadFixture(deployFixture));
  });

  it("Validate Signature", async function () {
    const amountRuns = 10;
    for (let i = 0; i < amountRuns; i++) {
      const [success] = await crypto.validateSignature(
        signedData[i].GroupSignature,
        signedData[i].BClaims
      );
      expect(success).to.be.equal(true);
    }
  });

  it("Validate SignatureASM [ @skip-on-coverage ]", async function () {
    for (let i = 0; i < amountRuns; i++) {
      const [success] = await crypto.validateSignatureASM(
        signedData[i].GroupSignature,
        signedData[i].BClaims
      );
      expect(success).to.be.equal(true);
    }
  });

  it("ValidateSignatureASM should not validate bad signatures [ @skip-on-coverage ]", async function () {
    for (let i = 0; i < amountRuns; i++) {
      const [success] = await crypto.validateBadSignatureASM(
        signedData[i].GroupSignature,
        signedData[i].BClaims,
        i
      );
      expect(success).to.be.equal(false);
    }
  });
});
