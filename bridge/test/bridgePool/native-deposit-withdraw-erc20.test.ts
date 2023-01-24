import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { Contract, ContractFactory } from "ethers";
import { BytesLike, defaultAbiCoder } from "ethers/lib/utils";
import { ethers } from "hardhat";
import { BridgePoolFactory } from "../../typechain-types";
import { expect } from "../chai-setup";

import {
  deployUpgradeableWithFactory,
  Fixture,
  getContractAddressFromBridgePoolCreatedEvent,
  getFixture,
  getImpersonatedSigner,
  getMetamorphicAddress,
} from "../setup";
import {
  getMockBlockClaimsForSnapshot,
  proofs,
  utxoOwner,
  vsPreImage,
  wrongChainIdVSPreImage,
  wrongKeyProofs,
  wrongProofs,
  wrongUTXOIDVSPreImage,
} from "./setup";

let fixture: Fixture;
let user: SignerWithAddress;
let utxoOwnerSigner: SignerWithAddress;
let utxoOwnerSignerAddress: string;
let factorySigner: SignerWithAddress;
let asBridgeRouter: any;
let initialUserBalance: any;
let initialBridgePoolBalance: any;
let merkleProofLibraryErrors: any;
let bridgePool: Contract;
let nativeERCBridgePoolBaseErrors: Contract;
let erc20Contract: Contract;
let bridgePoolImplFactory: ContractFactory;
const bridgePoolTokenType = 0; // ERC20
const bridgePoolType = 0; // Native
const bridgePoolNativeChainId = 1337;
const bridgePoolVersion = 1;
const bridgePoolValue = 0;

const tokenId = 0;
const tokenAmount = 296850137; // To match SMT proof test output
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
    utxoOwnerSigner = await getImpersonatedSigner(utxoOwner);
    factorySigner = await getImpersonatedSigner(fixture.factory.address);
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
    erc20Contract = await (
      await (await ethers.getContractFactory("ERC20Mock")).deploy()
    ).deployed();
    const bridgePoolFactory = (await deployUpgradeableWithFactory(
      fixture.factory,
      "BridgePoolFactory",
      "BridgePoolFactory"
    )) as BridgePoolFactory;
    bridgePoolImplFactory = await ethers.getContractFactory(
      "NativeERC20BridgePoolV1"
    );
    const bridgePoolImplBytecode = bridgePoolImplFactory.getDeployTransaction(
      fixture.factory.address,
      fixture.snapshots.address
    ).data as BytesLike;
    await bridgePoolFactory
      .connect(factorySigner)
      .deployPoolLogic(
        bridgePoolType,
        bridgePoolTokenType,
        bridgePoolVersion,
        bridgePoolImplBytecode,
        bridgePoolValue
      );
    const initializeEncodedParams =
      bridgePoolImplFactory.interface.encodeFunctionData("initialize", [
        erc20Contract.address,
      ]);
    const tx = await bridgePoolFactory
      .connect(factorySigner)
      .deployNewNativePool(
        bridgePoolTokenType,
        erc20Contract.address,
        bridgePoolVersion,
        bridgePoolNativeChainId,
        initializeEncodedParams
      );
    bridgePool = await ethers.getContractAt(
      "NativeERC20BridgePoolV1",
      await getContractAddressFromBridgePoolCreatedEvent(tx)
    );
    utxoOwnerSigner = await getImpersonatedSigner(utxoOwner);
    utxoOwnerSignerAddress = await utxoOwnerSigner.getAddress();
    // Simulate a bridge router with some gas for transactions
    const bridgeRouterAddress = getMetamorphicAddress(
      fixture.factory.address,
      "BridgeRouter"
    );
    asBridgeRouter = await getImpersonatedSigner(bridgeRouterAddress);
    // Mint and approve deposit amount
    erc20Contract.mint(utxoOwnerSignerAddress, tokenAmount);
    erc20Contract
      .connect(utxoOwnerSigner)
      .approve(bridgePool.address, tokenAmount);
    // Register initial balances
    initialUserBalance = await erc20Contract.balanceOf(utxoOwnerSignerAddress);
    initialBridgePoolBalance = await erc20Contract.balanceOf(
      bridgePool.address
    );
  }

  beforeEach(async function () {
    await loadFixture(deployFixture);
  });

  it("Should deposit if called from Bridge Router", async () => {
    await bridgePool
      .connect(asBridgeRouter)
      .deposit(utxoOwnerSignerAddress, encodedDepositParameters);
    expect(await erc20Contract.balanceOf(utxoOwnerSignerAddress)).to.be.eq(
      initialUserBalance - tokenAmount
    );
    expect(await erc20Contract.balanceOf(bridgePool.address)).to.be.eq(
      initialBridgePoolBalance + tokenAmount
    );
  });

  it("Should fail to deposit if not called from Bridge Router", async () => {
    await expect(
      bridgePool.deposit(utxoOwnerSignerAddress, encodedDepositParameters)
    ).to.be.revertedWithCustomError(bridgePoolImplFactory, "OnlyBridgeRouter");
  });

  it("Should make a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
    // Make first a deposit to withdraw afterwards
    await bridgePool
      .connect(asBridgeRouter)
      .deposit(utxoOwnerSignerAddress, encodedDepositParameters);
    await bridgePool
      .connect(asBridgeRouter)
      .withdraw(utxoOwnerSigner.address, vsPreImage, proofs);
    expect(await erc20Contract.balanceOf(utxoOwnerSignerAddress)).to.be.eq(
      initialUserBalance
    );
    expect(await erc20Contract.balanceOf(bridgePool.address)).to.be.eq(
      initialBridgePoolBalance
    );
  });

  it("Should fail to make a withdraw if not called from Bridge Router", async () => {
    await expect(
      bridgePool.withdraw(utxoOwnerSigner.address, vsPreImage, proofs)
    ).to.be.revertedWithCustomError(bridgePoolImplFactory, "OnlyBridgeRouter");
  });

  it("Should not make a withdraw on an already withdrawn UTXO upon proofs verification", async () => {
    await bridgePool
      .connect(asBridgeRouter)
      .deposit(utxoOwnerSignerAddress, encodedDepositParameters);
    await bridgePool
      .connect(asBridgeRouter)
      .withdraw(utxoOwnerSigner.address, vsPreImage, proofs);
    await expect(
      bridgePool
        .connect(asBridgeRouter)
        .withdraw(utxoOwnerSigner.address, vsPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "UTXOAlreadyWithdrawn"
    );
  });

  it("Should not make a withdraw for amount specified on informed burned UTXO with wrong merkle proof", async () => {
    await expect(
      bridgePool
        .connect(asBridgeRouter)
        .withdraw(utxoOwnerSigner.address, vsPreImage, wrongProofs)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });

  it("Should not make a withdraw if sender does not match UTXO owner", async () => {
    await expect(
      bridgePool
        .connect(asBridgeRouter)
        .withdraw(user.address, vsPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "UTXOAccountDoesNotMatchReceiver"
    );
  });

  it("Should not make a withdraw if chainId in UTXO does not match chainId in snapshot's claims", async () => {
    await expect(
      bridgePool
        .connect(asBridgeRouter)
        .withdraw(utxoOwnerSigner.address, wrongChainIdVSPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "ChainIdDoesNotMatch"
    );
  });

  it("Should not make a withdraw if UTXOID in UTXO does not match UTXOID in proof", async () => {
    await expect(
      bridgePool
        .connect(asBridgeRouter)
        .withdraw(utxoOwnerSigner.address, wrongUTXOIDVSPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "MerkleProofKeyDoesNotMatchUTXOID"
    );
  });

  it("Should not call a withdraw if state key does not match txhash key", async () => {
    await expect(
      bridgePool
        .connect(asBridgeRouter)
        .withdraw(utxoOwnerSigner.address, vsPreImage, wrongKeyProofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "UTXODoesnotMatch"
    );
  });
});