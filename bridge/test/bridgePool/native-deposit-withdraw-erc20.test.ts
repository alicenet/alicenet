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
  getMetamorphicAddress,
} from "../setup";
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
let utxoOwnerSigner: SignerWithAddress;
let utxoOwnerSignerAddress: string;
let bridgeRouter: any;
let initialUserBalance: any;
let initialBridgePoolBalance: any;
let merkleProofLibraryErrors: any;
let nativeERC20BridgePoolV1: Contract;
let nativeERCBridgePoolBaseErrors: Contract;
let erc20Mock: Contract;
const tokenId = 0;
const tokenAmount = 296850137;
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

describe("Testing BridgePool Deposit/Withdraw for tokenType ERC20", async () => {
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
    erc20Mock = await (
      await (await ethers.getContractFactory("ERC20Mock")).deploy()
    ).deployed();

    nativeERC20BridgePoolV1 = await deployUpgradeableWithFactory(
      fixture.factory,
      "NativeERC20BridgePoolV1",
      "NativeERC20BridgePoolV1",
      [erc20Mock.address],
      undefined,
      undefined
    );
    utxoOwnerSigner = await getImpersonatedSigner(
      "0x38e959391dd8598ae80d5d6d114a7822a09d313a"
    );
    utxoOwnerSignerAddress = await utxoOwnerSigner.getAddress();

    // Simulate a bridge router with some gas for transactions
    const bridgeRouterAddress = getMetamorphicAddress(
      fixture.factory.address,
      "BridgeRouter"
    );
    bridgeRouter = await getImpersonatedSigner(bridgeRouterAddress);
    erc20Mock.mint(utxoOwnerSignerAddress, tokenAmount);
    erc20Mock
      .connect(utxoOwnerSigner)
      .approve(nativeERC20BridgePoolV1.address, tokenAmount);
    initialUserBalance = await erc20Mock.balanceOf(utxoOwnerSignerAddress);
    initialBridgePoolBalance = await erc20Mock.balanceOf(
      nativeERC20BridgePoolV1.address
    );
  }

  beforeEach(async function () {
    await loadFixture(deployFixture);
  });

  it("Should make a deposit", async () => {
    await nativeERC20BridgePoolV1
      .connect(bridgeRouter)
      .deposit(utxoOwnerSignerAddress, encodedDepositParameters);
    expect(await erc20Mock.balanceOf(utxoOwnerSignerAddress)).to.be.eq(
      initialUserBalance - tokenAmount
    );
    expect(await erc20Mock.balanceOf(nativeERC20BridgePoolV1.address)).to.be.eq(
      initialBridgePoolBalance + tokenAmount
    );
  });

  it("Should make a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
    // Make first a deposit to withdraw afterwards
    await nativeERC20BridgePoolV1
      .connect(bridgeRouter)
      .deposit(utxoOwnerSignerAddress, encodedDepositParameters);
    await nativeERC20BridgePoolV1
      .connect(utxoOwnerSigner)
      .withdraw(vsPreImage, proofs);
    expect(await erc20Mock.balanceOf(utxoOwnerSignerAddress)).to.be.eq(
      initialUserBalance
    );
    expect(await erc20Mock.balanceOf(nativeERC20BridgePoolV1.address)).to.be.eq(
      initialBridgePoolBalance
    );
  });

  it("Should not make a withdraw on an already withdrawn UTXO upon proofs verification", async () => {
    await nativeERC20BridgePoolV1
      .connect(bridgeRouter)
      .deposit(utxoOwnerSignerAddress, encodedDepositParameters);
    await nativeERC20BridgePoolV1
      .connect(utxoOwnerSigner)
      .withdraw(vsPreImage, proofs);
    await expect(
      nativeERC20BridgePoolV1
        .connect(utxoOwnerSigner)
        .withdraw(vsPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "UTXOAlreadyWithdrawn"
    );
  });

  it("Should not make a withdraw for amount specified on informed burned UTXO with wrong merkle proof", async () => {
    await expect(
      nativeERC20BridgePoolV1
        .connect(utxoOwnerSigner)
        .withdraw(vsPreImage, wrongProofs)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });

  it("Should not make a withdraw if sender does not match UTXO owner", async () => {
    await expect(
      nativeERC20BridgePoolV1.connect(user).withdraw(vsPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "UTXOAccountDoesNotMatchReceiver"
    );
  });

  it("Should not make a withdraw if chainId in UTXO does not match chainId in snapshot's claims", async () => {
    await expect(
      nativeERC20BridgePoolV1
        .connect(utxoOwnerSigner)
        .withdraw(wrongChainIdVSPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "ChainIdDoesNotMatch"
    );
  });

  it("Should not make a withdraw if UTXOID in UTXO does not match UTXOID in proof", async () => {
    await expect(
      nativeERC20BridgePoolV1
        .connect(utxoOwnerSigner)
        .withdraw(wrongUTXOIDVSPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "MerkleProofKeyDoesNotMatchUTXOID"
    );
  });
});