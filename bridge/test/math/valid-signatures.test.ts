import { ethers } from "hardhat";
import { CryptoLibraryWrapper } from "../../typechain-types";
import { expect } from "../chai-setup";
import { signedData } from "./assets/4-validators-1000-snapshots";

describe("CryptoLibrary: Validate Signature", () => {
  let crypto: CryptoLibraryWrapper;
  const amountRuns = 100;

  before(async () => {
    crypto = await (
      await ethers.getContractFactory("CryptoLibraryWrapper")
    ).deploy();
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

  it("Validate SignatureASM", async function () {
    for (let i = 0; i < amountRuns; i++) {
      const [success] = await crypto.validateSignatureASM(
        signedData[i].GroupSignature,
        signedData[i].BClaims
      );
      expect(success).to.be.equal(true);
    }
  });

  it("ValidateSignatureASM should not validate bad signatures", async function () {
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
