import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { assert } from "chai";
import { ethers } from "hardhat";
import { AccusationsLibraryMock } from "../../typechain-types";
import {
  deployLibrary,
  generateSigAndPClaims0,
  generateSigAndPClaims1,
} from "./accusations-test-helpers";
describe("AccusationsLibrary: Tests AccusationsLibrary methods", async () => {
  let accusation: AccusationsLibraryMock;

  beforeEach(async function () {
    accusation = await loadFixture(deployLibrary);
  });

  describe("recoverSigner:", async () => {
    it("returns signer when valid", async function () {
      const sig =
        "0x" +
        "cba766e2ba024aad86db556635cec9f104e76644b235f77759ff80bfefc990c5774d2d5ff3069a5099e4f9fadc9b08ab20472e2ef432fba94498d93c10cc584b00";
      const prefix = ethers.utils.toUtf8Bytes("");
      const message =
        "0x" +
        "54686520717569636b2062726f776e20666f782064696420736f6d657468696e67";
      const expectedAddress = "0x38e959391dD8598aE80d5d6D114a7822A09d313A";

      const who = await accusation.recoverSigner(sig, prefix, message);

      assert.equal(expectedAddress, who);
    });

    it("returns signer with pclaims data", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const prefix = ethers.utils.toUtf8Bytes("Proposal");
      const expectedAddress = "0x38e959391dD8598aE80d5d6D114a7822A09d313A";

      const who = await accusation.recoverSigner(sig0, prefix, pClaims0);

      assert.equal(expectedAddress, who);
    });
  });
  describe("recoverMadNetSigner:", async () => {
    it("returns signer when valid", async function () {
      const { sig: sig0, pClaims: pClaims0 } = generateSigAndPClaims0();
      const { sig: sig1, pClaims: pClaims1 } = generateSigAndPClaims1();
      const expectedAddress = "0x38e959391dD8598aE80d5d6D114a7822A09d313A";

      const signerAccount0 = await accusation.recoverMadNetSigner(
        sig0,
        pClaims0
      );
      const signerAccount1 = await accusation.recoverMadNetSigner(
        sig1,
        pClaims1
      );

      assert.equal(expectedAddress, signerAccount0);
      assert.equal(expectedAddress, signerAccount1);
    });
  });

  describe("computeUTXOID:", async () => {
    it("returns correct tx id when valid", async function () {
      const txHash =
        "0xf172873c63909462ac4de545471fd3ad3e9eeadeec4608b92d16ce6b500704cc";
      const txHash2 =
        "0xb4aec67f3220a8bcdee78d4aaec6ea419171e3db9c27c65d70cc85d60e07a3f7";

      const txIdx = 0;
      const txIdx2 = 1;

      const expected =
        "0xda3dc36dc016d513fbac07ed6605c6157088d8c673df3b5bb09682b7937d5250";
      const expected2 =
        "0x4f6b55978f29b3eae295b96d213a58c4d69ef65f20b3c4463ff682aeb0407625";

      const actual = await accusation.computeUTXOID(txHash, txIdx);
      const actual2 = await accusation.computeUTXOID(txHash2, txIdx2);

      assert.equal(actual, expected);
      assert.equal(actual2, expected2);
    });
  });
});
