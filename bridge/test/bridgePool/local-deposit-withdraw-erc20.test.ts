import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { IBridgePool } from "../../typechain-types";
import { expect } from "../chai-setup";

import { Fixture, getFixture } from "../setup";
import {
  encodedBurnedUTXOERC20,
  getMockBlockClaimsForStateRoot,
  getSimulatedBridgeRouter,
  merkleProof,
  stateRoot,
  tokenAmount,
  wrongMerkleProof,
} from "./setup";

let fixture: Fixture;
let user: SignerWithAddress;
let user2: SignerWithAddress;
let bridgePool: IBridgePool;
let bridgeRouter: any;
let initialUserBalance: any;
let initialBridgePoolBalance: any;
let merkleProofLibraryErrors: any;
const erc20data = {
  it: "ERC20",
  options: {
    ercContractName: "erc20Mock",
    tokenType: 1,
    bridgeImpl: "localERC20BridgePoolV1",
    quantity: tokenAmount,
    errorReason: "ERC20: insufficient allowance",
  },
};

describe("Testing BridgePool Deposit/Withdraw for tokenType ERC20", async () => {
  beforeEach(async () => {
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
    // Simulate a bridge router with some gas for transactions
    bridgeRouter = await getSimulatedBridgeRouter(fixture.factory.address);
    bridgePool = fixture[erc20data.options.bridgeImpl] as IBridgePool;
    fixture[erc20data.options.ercContractName].mint(user.address, tokenAmount);
    fixture[erc20data.options.ercContractName]
      .connect(user)
      .approve(bridgePool.address, tokenAmount);
    initialUserBalance = await fixture.erc20Mock.balanceOf(user.address);
    initialBridgePoolBalance = await fixture.erc20Mock.balanceOf(
      bridgePool.address
    );
  });

  it("Should make a deposit", async () => {
    await bridgePool.connect(bridgeRouter).deposit(user.address, tokenAmount);
    expect(await fixture.erc20Mock.balanceOf(user.address)).to.be.eq(
      initialUserBalance - tokenAmount
    );
    expect(await fixture.erc20Mock.balanceOf(bridgePool.address)).to.be.eq(
      initialBridgePoolBalance + tokenAmount
    );
  });

  it("Should make a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
    // Make first a deposit to withdraw afterwards
    await bridgePool.connect(bridgeRouter).deposit(user.address, tokenAmount);
    await bridgePool
      .connect(user)
      .withdraw(merkleProof, encodedBurnedUTXOERC20);
    expect(await fixture.erc20Mock.balanceOf(user.address)).to.be.eq(
      initialUserBalance
    );
    expect(await fixture.erc20Mock.balanceOf(bridgePool.address)).to.be.eq(
      initialBridgePoolBalance
    );
  });

  it("Should not make a withdraw for amount specified on informed burned UTXO with wrong merkle proof", async () => {
    await expect(
      bridgePool
        .connect(user)
        .withdraw(wrongMerkleProof, encodedBurnedUTXOERC20)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });

  it("Should not make a withdraw for amount specified on informed burned UTXO with wrong root", async () => {
    const wrongStateRoot =
      "0x0000000000000000000000000000000000000000000000000000000000000000";
    const encodedMockBlockClaims =
      getMockBlockClaimsForStateRoot(wrongStateRoot);
    await fixture.snapshots.snapshot(
      Buffer.from("0x0"),
      encodedMockBlockClaims
    );
    await expect(
      bridgePool.connect(user).withdraw(merkleProof, encodedBurnedUTXOERC20)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });

  it("Should not make a withdraw to an address that is not the owner in informed burned UTXO", async () => {
    await expect(
      bridgePool.connect(user2).withdraw(merkleProof, encodedBurnedUTXOERC20)
    ).to.be.revertedWithCustomError(
      bridgePool,
      "ReceiverIsNotOwnerOnProofOfBurnUTXO"
    );
  });
});
