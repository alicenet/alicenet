import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { Contract } from "ethers"; 
import { ethers } from "hardhat";
import { expect } from "../chai-setup";

import { deployStaticWithFactory, Fixture, getFixture } from "../setup";
import {
  getEncodedBurnedUTXO,
  getMockBlockClaimsForStateRoot,
  merkleProof,
  stateRoot,
  valueOrId,
  wrongMerkleProof,
} from "./setup";

let fixture: Fixture;
let user: SignerWithAddress;
let user2: SignerWithAddress;
let merkleProofLibraryErrors: Contract;
let localERCBridgePoolBase: Contract;
let encodedBurnedUTXO: string;

describe("Testing BridgePool Router Deposit/Withdraw for tokenType ", async () => {
  before(async () => {
    [, user, user2] = await ethers.getSigners();
    fixture = await getFixture(true, true, false);
    const encodedMockBlockClaims = getMockBlockClaimsForStateRoot(stateRoot);
    // Take a mock snapshot
    await fixture.snapshots.snapshot(
      Buffer.from("0x0"),
      encodedMockBlockClaims
    );
    merkleProofLibraryErrors = await (
      await (
        await ethers.getContractFactory("MerkleProofLibraryErrors")
      ).deploy()
    ).deployed();

    encodedBurnedUTXO = getEncodedBurnedUTXO(user.address, valueOrId);
  });

  it("Should call a deposit", async () => {
    await localERCBridgePoolBase.deposit(user.address, valueOrId);
  });

  it("Should call a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
    await localERCBridgePoolBase
      .connect(user)
      .withdraw(merkleProof, encodedBurnedUTXO);
  });

  it("Should not call a withdraw for amount specified on informed burned UTXO with wrong merkle proof", async () => {
    await expect(
      localERCBridgePoolBase
        .connect(user)
        .withdraw(wrongMerkleProof, encodedBurnedUTXO)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });

  it("Should not call a withdraw for amount specified on informed burned UTXO with wrong root", async () => {
    const wrongStateRoot =
      "0x0000000000000000000000000000000000000000000000000000000000000000";
    const encodedMockBlockClaims =
      getMockBlockClaimsForStateRoot(wrongStateRoot);
    await fixture.snapshots.snapshot(
      Buffer.from("0x0"),
      encodedMockBlockClaims
    );
    await expect(
      localERCBridgePoolBase
        .connect(user)
        .withdraw(merkleProof, encodedBurnedUTXO)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });

  it("Should not call a withdraw to an address that is not the owner in informed burned UTXO", async () => {
    await expect(
      localERCBridgePoolBase
        .connect(user2)
        .withdraw(merkleProof, encodedBurnedUTXO)
    ).to.be.revertedWithCustomError(
      localERCBridgePoolBase,
      "ReceiverIsNotOwnerOnProofOfBurnUTXO"
    );
  });
});
