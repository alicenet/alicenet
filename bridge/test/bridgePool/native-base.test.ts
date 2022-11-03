import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { Contract } from "ethers";
import { defaultAbiCoder } from "ethers/lib/utils";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";

import {
  deployUpgradeableWithFactory,
  Fixture,
  getFixture,
  getImpersonatedSigner
} from "../setup";
import {
  getMockBlockClaimsForSnapshot,
  proofs,
  vsPreImage,
  wrongChainIdVSPreImage,
  wrongKeyProofs,
  wrongProofs,
  wrongUTXOIDVSPreImage,
} from "./setup";

let fixture: Fixture;
let user: SignerWithAddress;
let utxoOwnerSigner: SignerWithAddress;
let merkleProofLibraryErrors: Contract;
let nativeERCBridgePool: Contract;
let nativeERCBridgePoolBaseErrors: Contract;
let asBridgeRouter: any;
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
    [, user] = await ethers.getSigners();
    // Take a mock snapshot
    await fixture.snapshots.snapshot(
      Buffer.from("0x0"),
      encodedMockBlockClaims
    );
    nativeERCBridgePoolBaseErrors = await (
      await (
        await ethers.getContractFactory("NativeERCBridgePoolBaseErrors")
      ).deploy()
    ).deployed();
    merkleProofLibraryErrors = await (
      await (
        await ethers.getContractFactory("MerkleProofLibraryErrors")
      ).deploy()
    ).deployed();
    nativeERCBridgePool = await deployUpgradeableWithFactory(
      fixture.factory,
      "NativeERCBridgePoolMock",
      "NativeERCBridgePoolMock",
      undefined,
      undefined,
      undefined
    );
    const bridgeRouter = await deployUpgradeableWithFactory(
      fixture.factory,
      "BridgeRouterMock",
      "BridgeRouter",
      undefined,
      [1000]
    );
    utxoOwnerSigner = await getImpersonatedSigner(
      "0x38e959391dd8598ae80d5d6d114a7822a09d313a"
    );
    asBridgeRouter = await getImpersonatedSigner(bridgeRouter.address);
  }

  beforeEach(async function () {
    await loadFixture(deployFixture);
  });

  it("Should call a deposit", async () => {
    await nativeERCBridgePool
      .connect(asBridgeRouter)
      .deposit(user.address, encodedDepositParameters);
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

  it("Should not call a withdraw if chainId in UTXO does not match chainId in snapshot's claims", async () => {
    await expect(
      nativeERCBridgePool
        .connect(utxoOwnerSigner)
        .withdraw(wrongChainIdVSPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "ChainIdDoesNotMatch"
    );
  });

  it("Should not call a withdraw if UTXOID in UTXO does not match UTXOID in proof", async () => {
    await expect(
      nativeERCBridgePool
        .connect(utxoOwnerSigner)
        .withdraw(wrongUTXOIDVSPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "MerkleProofKeyDoesNotMatchUTXOID"
    );
  });

  it("Should not call a withdraw if state key in proofs does not match txhash key", async () => {
    await expect(
      nativeERCBridgePool
        .connect(utxoOwnerSigner)
        .withdraw(vsPreImage, wrongKeyProofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "UTXODoesnotMatch"
    );
  });
});
