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
  vsPreImage,
  wrongChainIdVSPreImage,
  wrongProofs,
  wrongUTXOIDVSPreImage,
} from "./setup";

let fixture: Fixture;
let user: SignerWithAddress;
let user2: SignerWithAddress;
let utxoOwnerSigner: SignerWithAddress;
let merkleProofLibraryErrors: Contract;
let accusationsErrors: Contract;
let nativeERCBridgePool: Contract;
let nativeERCBridgePoolBaseErrors: Contract;
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

describe("Testing Base BridgePool Deposit/Withdraw", async () => {
  async function deployFixture() {
    fixture = await getFixture(true, true, false);
    [, user, user2] = await ethers.getSigners();
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
    accusationsErrors = await (
      await (await ethers.getContractFactory("AccusationsErrors")).deploy()
    ).deployed();
    nativeERCBridgePoolBaseErrors = await (
      await (
        await ethers.getContractFactory("NativeERCBridgePoolBaseErrors")
      ).deploy()
    ).deployed();
    nativeERCBridgePool = await deployUpgradeableWithFactory(
      fixture.factory,
      "NativeERCBridgePoolMock",
      "NativeERCBridgePoolMock",
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
      nativeERCBridgePool,
      nativeERCBridgePoolBaseErrors,
    };
  }

  beforeEach(async function () {
    ({
      fixture,
      user,
      user2,
      merkleProofLibraryErrors,
      nativeERCBridgePool,
      nativeERCBridgePoolBaseErrors,
    } = await loadFixture(deployFixture));
  });

  it("Should call a deposit", async () => {
    await nativeERCBridgePool.deposit(user.address, encodedDepositParameters);
  });

  it("Should call a withdraw upon proofs verification", async () => {
    await nativeERCBridgePool
      .connect(utxoOwnerSigner)
      .withdraw(vsPreImage, proofs);
  });

  it("Should not call a withdraw on an already withdrawn UTXO upon proofs verification", async () => {
    await nativeERCBridgePool
      .connect(utxoOwnerSigner)
      .withdraw(vsPreImage, proofs);
    await expect(
      nativeERCBridgePool.connect(utxoOwnerSigner).withdraw(vsPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "UTXOAlreadyWithdrawn"
    );
  });

  it("Should not call a withdraw with wrong proof", async () => {
    await expect(
      nativeERCBridgePool
        .connect(utxoOwnerSigner)
        .withdraw(vsPreImage, wrongProofs)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });

  it("Should not call a withdraw if sender does not match UTXO owner", async () => {
    await expect(
      nativeERCBridgePool.connect(user).withdraw(vsPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "UTXOAccountDoesNotMatchReceiver"
    );
  });

  it("Should not call a withdraw if chainId in UTXO does not match chainId in snapsht's claims", async () => {
    await expect(
      nativeERCBridgePool
        .connect(utxoOwnerSigner)
        .withdraw(wrongChainIdVSPreImage, proofs)
    ).to.be.revertedWithCustomError(accusationsErrors, "ChainIdDoesNotMatch");
  });

  it("Should not call a withdraw with wrong proof", async () => {
    await expect(
      nativeERCBridgePool
        .connect(utxoOwnerSigner)
        .withdraw(wrongUTXOIDVSPreImage, proofs)
    ).to.be.revertedWithCustomError(
      accusationsErrors,
      "MerkleProofKeyDoesNotMatchUTXOIDBeingSpent"
    );
  });
});
