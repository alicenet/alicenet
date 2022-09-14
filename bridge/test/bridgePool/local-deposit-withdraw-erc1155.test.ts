import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { defaultAbiCoder } from "ethers/lib/utils";
import { ethers } from "hardhat";
import { IBridgePool } from "../../typechain-types";
import { expect } from "../chai-setup";
import { Fixture, getFixture } from "../setup";
import {
  getEncodedBurnedUTXO,
  getMockBlockClaimsForStateRoot,
  getSimulatedBridgeRouter,
  merkleProof,
  stateRoot,
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
let encodedBurnedUTXO: any;
const tokenId = 100;
const tokenAmount = 1;
const encodedDepositParameters = defaultAbiCoder.encode(
  ["tuple(uint256 tokenId_, uint256 tokenAmount_)"],
  [
    {
      tokenId_: tokenId,
      tokenAmount_: tokenAmount,
    },
  ]
);

describe("Testing BridgePool Deposit/Withdraw for tokenType ERC1155", async () => {
  beforeEach(async () => {
    [, user, user2] = await ethers.getSigners();
    fixture = await getFixture(true, true, false);
    const encodedMockBlockClaims = getMockBlockClaimsForStateRoot(stateRoot);
    // Take a mock snapshot
    await fixture.snapshots.snapshot(
      Buffer.from("0x0"),
      encodedMockBlockClaims
    );
    encodedBurnedUTXO = getEncodedBurnedUTXO(
      user.address,
      tokenId,
      tokenAmount
    );
    merkleProofLibraryErrors = await (
      await (
        await ethers.getContractFactory("MerkleProofLibraryErrors")
      ).deploy()
    ).deployed();
    // Simulate a bridge router with some gas for transactions
    bridgeRouter = await getSimulatedBridgeRouter(fixture.factory.address);
    bridgePool = fixture.localERC1155BridgePoolV1 as IBridgePool;
    fixture.erc1155Mock.mint(user.address, tokenId, tokenAmount);
    fixture.erc1155Mock
      .connect(user)
      .setApprovalForAll(bridgePool.address, true);
    initialUserBalance = await fixture.erc1155Mock.balanceOf(
      user.address,
      tokenId
    );
    initialBridgePoolBalance = await fixture.erc1155Mock.balanceOf(
      bridgePool.address,
      tokenId
    );
  });

  it("Should make a deposit", async () => {
    await bridgePool
      .connect(bridgeRouter)
      .deposit(user.address, encodedDepositParameters);
    expect(await fixture.erc1155Mock.balanceOf(user.address, tokenId)).to.be.eq(
      initialUserBalance - tokenAmount
    );
    expect(
      await fixture.erc1155Mock.balanceOf(bridgePool.address, tokenId)
    ).to.be.eq(initialBridgePoolBalance + tokenAmount);
  });

  it("Should make a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
    // Make first a deposit to withdraw afterwards
    await bridgePool
      .connect(bridgeRouter)
      .deposit(user.address, encodedDepositParameters);
    await bridgePool.connect(user).withdraw(merkleProof, encodedBurnedUTXO);
    expect(await fixture.erc1155Mock.balanceOf(user.address, tokenId)).to.be.eq(
      initialUserBalance
    );
    expect(
      await fixture.erc1155Mock.balanceOf(bridgePool.address, tokenId)
    ).to.be.eq(initialBridgePoolBalance);
  });

  it("Should not make a withdraw for amount specified on informed burned UTXO with wrong merkle proof", async () => {
    await expect(
      bridgePool.connect(user).withdraw(wrongMerkleProof, encodedBurnedUTXO)
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
      bridgePool.connect(user).withdraw(merkleProof, encodedBurnedUTXO)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });

  it("Should not make a withdraw to an address that is not the owner in informed burned UTXO", async () => {
    await expect(
      bridgePool.connect(user2).withdraw(merkleProof, encodedBurnedUTXO)
    ).to.be.revertedWithCustomError(
      bridgePool,
      "ReceiverIsNotOwnerOnProofOfBurnUTXO"
    );
  });
});
