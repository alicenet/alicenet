import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber, Contract } from "ethers";
import { defaultAbiCoder } from "ethers/lib/utils";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  deployUpgradeableWithFactory,
  Fixture,
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
let bridgeRouter: any;
let initialUserBalance: any;
let initialBridgePoolBalance: any;
let merkleProofLibraryErrors: any;
let nativeERC1155BridgePoolV1: Contract;
let nativeERCBridgePoolBaseErrors: Contract;
let erc1155Mock: Contract;
let initialBatchUserBalance: BigNumber[];
const tokenId = 296850137; // to match the value in test UTXO
const tokenAmount = 1;
let batchUserAccounts: string[];
let batchBridgePoolAccounts: string[];
const batchTokenIds = [0, 1, 2, 3];
const batchTokenAmounts = [100, 100, 1, 1];
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

describe("Testing Bridge Pool Deposit/Withdraw for tokenType ERC1155", async () => {
  beforeEach(async () => {
    [, user] = await ethers.getSigners();
    fixture = await getFixture(true, true, false);
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
    erc1155Mock = await (
      await (await ethers.getContractFactory("ERC1155Mock")).deploy()
    ).deployed();
    nativeERC1155BridgePoolV1 = await deployUpgradeableWithFactory(
      fixture.factory,
      "NativeERC1155BridgePoolV1",
      "NativeERC1155BridgePoolV1",
      [erc1155Mock.address],
      undefined,
      undefined
    );
    batchUserAccounts = [
      user.address,
      user.address,
      user.address,
      user.address,
    ];
    utxoOwnerSigner = await getImpersonatedSigner(utxoOwner);
    utxoOwnerSignerAddress = await utxoOwnerSigner.getAddress();
    const bridgeRouterAddress = getMetamorphicAddress(
      fixture.factory.address,
      "BridgeRouter"
    );
    bridgeRouter = await getImpersonatedSigner(bridgeRouterAddress);
    batchBridgePoolAccounts = [
      nativeERC1155BridgePoolV1.address,
      nativeERC1155BridgePoolV1.address,
      nativeERC1155BridgePoolV1.address,
      nativeERC1155BridgePoolV1.address,
    ];
    erc1155Mock.mint(user.address, tokenId, tokenAmount);
    for (let i = 0; i < batchUserAccounts.length; i++) {
      erc1155Mock.mint(
        batchUserAccounts[i],
        batchTokenIds[i],
        batchTokenAmounts[i]
      );
    }
    erc1155Mock
      .connect(user)
      .setApprovalForAll(nativeERC1155BridgePoolV1.address, true);
    initialUserBalance = await erc1155Mock.balanceOf(user.address, tokenId);
    initialBridgePoolBalance = await erc1155Mock.balanceOf(
      nativeERC1155BridgePoolV1.address,
      tokenId
    );
    initialBatchUserBalance = await erc1155Mock.balanceOfBatch(
      batchUserAccounts,
      batchTokenIds
    );
  });

  it("Should make a deposit", async () => {
    await nativeERC1155BridgePoolV1
      .connect(bridgeRouter)
      .deposit(user.address, encodedDepositParameters);
    expect(await erc1155Mock.balanceOf(user.address, tokenId)).to.be.eq(
      initialUserBalance - tokenAmount
    );
    expect(
      await erc1155Mock.balanceOf(nativeERC1155BridgePoolV1.address, tokenId)
    ).to.be.eq(initialBridgePoolBalance + tokenAmount);
  });

  it("Should make a batch deposit", async () => {
    await nativeERC1155BridgePoolV1
      .connect(bridgeRouter)
      .batchDeposit(user.address, batchTokenIds, batchTokenAmounts);
    expect(
      await erc1155Mock.balanceOfBatch(batchBridgePoolAccounts, batchTokenIds)
    ).to.be.deep.eq(initialBatchUserBalance);
  });

  it("Should not make a batch deposit with arrays of different length", async () => {
    await expect(
      nativeERC1155BridgePoolV1
        .connect(bridgeRouter)
        .batchDeposit(user.address, batchTokenIds, batchTokenAmounts.slice(1))
    ).to.be.revertedWith("ERC1155: ids and amounts length mismatch");
  });

  it("Should not make a deposit to address 0", async () => {
    const userAddress = ethers.constants.AddressZero;
    erc1155Mock.mint(userAddress, tokenId, tokenAmount);
    erc1155Mock
      .connect(userAddress)
      .setApprovalForAll(nativeERC1155BridgePoolV1.address, true);
    await expect(
      nativeERC1155BridgePoolV1
        .connect(bridgeRouter)
        .deposit(userAddress, encodedDepositParameters)
    ).to.be.revertedWith("ERC1155: caller is not token owner nor approved");
  });

  it("Should make a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
    // Make first a deposit to withdraw afterwards
    await nativeERC1155BridgePoolV1
      .connect(bridgeRouter)
      .deposit(utxoOwnerSignerAddress, encodedDepositParameters);
    await nativeERC1155BridgePoolV1
      .connect(utxoOwnerSigner)
      .withdraw(vsPreImage, proofs);
    expect(await erc1155Mock.balanceOf(utxoOwnerSignerAddress)).to.be.eq(
      initialUserBalance
    );
    expect(
      await erc1155Mock.balanceOf(nativeERC1155BridgePoolV1.address)
    ).to.be.eq(initialBridgePoolBalance);
  });

  it("Should not make a withdraw on an already withdrawn UTXO upon proofs verification", async () => {
    await nativeERC1155BridgePoolV1
      .connect(bridgeRouter)
      .deposit(utxoOwnerSignerAddress, encodedDepositParameters);
    await nativeERC1155BridgePoolV1
      .connect(utxoOwnerSigner)
      .withdraw(vsPreImage, proofs);
    await expect(
      nativeERC1155BridgePoolV1
        .connect(utxoOwnerSigner)
        .withdraw(vsPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "UTXOAlreadyWithdrawn"
    );
  });

  it("Should not make a withdraw for amount specified on informed burned UTXO with wrong merkle proof", async () => {
    await expect(
      nativeERC1155BridgePoolV1
        .connect(utxoOwnerSigner)
        .withdraw(vsPreImage, wrongProofs)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });

  it("Should not make a withdraw if sender does not match UTXO owner", async () => {
    await expect(
      nativeERC1155BridgePoolV1.connect(user).withdraw(vsPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "UTXOAccountDoesNotMatchReceiver"
    );
  });

  it("Should not make a withdraw if chainId in UTXO does not match chainId in snapshot's claims", async () => {
    await expect(
      nativeERC1155BridgePoolV1
        .connect(utxoOwnerSigner)
        .withdraw(wrongChainIdVSPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "ChainIdDoesNotMatch"
    );
  });

  it("Should not make a withdraw if UTXOID in UTXO does not match UTXOID in proof", async () => {
    await expect(
      nativeERC1155BridgePoolV1
        .connect(utxoOwnerSigner)
        .withdraw(wrongUTXOIDVSPreImage, proofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "MerkleProofKeyDoesNotMatchUTXOID"
    );
  });

  it("Should not call a withdraw if state key does not match txhash key", async () => {
    await expect(
      nativeERC1155BridgePoolV1
        .connect(utxoOwnerSigner)
        .withdraw(vsPreImage, wrongKeyProofs)
    ).to.be.revertedWithCustomError(
      nativeERCBridgePoolBaseErrors,
      "UTXODoesnotMatch"
    );
  });
});
