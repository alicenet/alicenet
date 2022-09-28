import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { Contract } from "ethers";
import { defaultAbiCoder } from "ethers/lib/utils";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";

import { deployUpgradeableWithFactory, Fixture, getFixture } from "../setup";
import {
  getImpersonatedSigner,
  getMockBlockClaimsForSnapshot,
  proofs,
  txInPreImage,
  wrongProofs,
} from "./setup";

let fixture: Fixture;
let user: SignerWithAddress;
let user2: SignerWithAddress;
let utxoOwnerSigner: SignerWithAddress;
let merkleProofLibraryErrors: Contract;
let localERCBridgePool: Contract;
let localERCBridgePoolBaseErrors: Contract;
const tokenId = 0;
const tokenAmount = 0;
const encodedDepositParameters = defaultAbiCoder.encode(
  ["tuple(uint256 tokenId_, uint256 tokenAmount_)"],
  [
    {
      tokenId_: tokenId,
      tokenAmount_: tokenAmount,
    },
  ]
);
const encodedMockBlockClaims = getMockBlockClaimsForSnapshot();

async function deployFixture() {
  fixture = await getFixture(true, true, false);
  [, user, user2] = await ethers.getSigners();
  // Take a mock snapshot
  await fixture.snapshots.snapshot(Buffer.from("0x0"), encodedMockBlockClaims);
  merkleProofLibraryErrors = await (
    await (await ethers.getContractFactory("MerkleProofLibraryErrors")).deploy()
  ).deployed();
  localERCBridgePoolBaseErrors = await (
    await (
      await ethers.getContractFactory("LocalERCBridgePoolBaseErrors")
    ).deploy()
  ).deployed();
  localERCBridgePool = await deployUpgradeableWithFactory(
    fixture.factory,
    "LocalERCBridgePoolMock",
    "LocalERCBridgePoolMock",
    undefined,
    undefined,
    undefined,
    false
  );
  utxoOwnerSigner = await getImpersonatedSigner(
    "0x38e959391dd8598ae80d5d6d114a7822a09d313a"
  );

  return {
    fixture,
    user,
    user2,
    merkleProofLibraryErrors,
    localERCBridgePool,
    localERCBridgePoolBaseErrors,
  };
}

beforeEach(async function () {
  ({
    fixture,
    user,
    user2,
    merkleProofLibraryErrors,
    localERCBridgePool,
    localERCBridgePoolBaseErrors,
  } = await loadFixture(deployFixture));
});

describe("Testing Base BridgePool Deposit/Withdraw", async () => {
  it("Should call a deposit", async () => {
    await localERCBridgePool.deposit(user.address, encodedDepositParameters);
  });

  it("Should call a withdraw upon proofs verification", async () => {
    await localERCBridgePool
      .connect(utxoOwnerSigner)
      .withdraw(txInPreImage, proofs);
  });

  it("Should not call a withdraw on an already withdrawn UTXO upon proofs verification", async () => {
    await localERCBridgePool
      .connect(utxoOwnerSigner)
      .withdraw(txInPreImage, proofs);
    await expect(
      localERCBridgePool.connect(utxoOwnerSigner).withdraw(txInPreImage, proofs)
    ).to.be.revertedWithCustomError(
      localERCBridgePoolBaseErrors,
      "UTXOAlreadyWithdrawn"
    );
  });

  it("Should not call a withdraw with wrong proof", async () => {
    await expect(
      localERCBridgePool.connect(user).withdraw(txInPreImage, wrongProofs)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });
});
