import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { IBridgePool } from "../../typechain-types";
import { expect } from "../chai-setup";
import { Fixture, getFixture } from "../setup";
import {
  encodedBurnedUTXOERC721,
  getMockBlockClaimsForStateRoot,
  getSimulatedBridgeRouter,
  merkleProof,
  stateRoot,
  tokenAmount,
  tokenId,
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
const erc721data = {
  it: "ERC721",
  options: {
    ercContractName: "erc721Mock",
    tokenType: 1,
    bridgeImpl: "localERC721BridgePoolV1",
    quantity: tokenAmount,
    errorReason: "ERC721: insufficient allowance",
  },
};

describe("Testing BridgePool Deposit/Withdraw for tokenType ERC721", async () => {
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
    bridgePool = fixture[erc721data.options.bridgeImpl] as IBridgePool;
    fixture[erc721data.options.ercContractName].mint(user.address, tokenId);
    fixture[erc721data.options.ercContractName]
      .connect(user)
      .approve(bridgePool.address, tokenId);
    initialUserBalance = await fixture.erc721Mock.balanceOf(user.address);
    initialBridgePoolBalance = await fixture.erc721Mock.balanceOf(
      bridgePool.address
    );
  });

  it("Should make a deposit", async () => {
    await bridgePool.connect(bridgeRouter).deposit(user.address, tokenId);
    expect(await fixture.erc721Mock.balanceOf(user.address)).to.be.eq(
      initialUserBalance - tokenAmount
    );
    expect(await fixture.erc721Mock.balanceOf(bridgePool.address)).to.be.eq(
      initialBridgePoolBalance + tokenAmount
    );
  });

  it("Should make a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
    // Make first a deposit to withdraw afterwards
    await bridgePool.connect(bridgeRouter).deposit(user.address, tokenId);
    await bridgePool
      .connect(user)
      .withdraw(merkleProof, encodedBurnedUTXOERC721);
    expect(await fixture.erc721Mock.balanceOf(user.address)).to.be.eq(
      initialUserBalance
    );
    expect(await fixture.erc721Mock.balanceOf(bridgePool.address)).to.be.eq(
      initialBridgePoolBalance
    );
  });

  it("Should not make a withdraw for amount specified on informed burned UTXO with wrong merkle proof", async () => {
    await expect(
      bridgePool
        .connect(user)
        .withdraw(wrongMerkleProof, encodedBurnedUTXOERC721)
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
      bridgePool.connect(user).withdraw(merkleProof, encodedBurnedUTXOERC721)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });

  it("Should not make a withdraw to an address that is not the owner in informed burned UTXO", async () => {
    await expect(
      bridgePool.connect(user2).withdraw(merkleProof, encodedBurnedUTXOERC721)
    ).to.be.revertedWithCustomError(
      bridgePool,
      "ReceiverIsNotOwnerOnProofOfBurnUTXO"
    );
  });
});
