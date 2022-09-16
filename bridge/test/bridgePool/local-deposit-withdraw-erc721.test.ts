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
const tokenId = 1;
const tokenAmount = 1;

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
    bridgePool = fixture.localERC721BridgePoolV1 as IBridgePool;
    fixture.erc721Mock.mint(user.address, tokenId);
    fixture.erc721Mock.connect(user).approve(bridgePool.address, tokenId);
    initialUserBalance = await fixture.erc721Mock.balanceOf(user.address);
    initialBridgePoolBalance = await fixture.erc721Mock.balanceOf(
      bridgePool.address
    );
  });

  it("Should make a deposit", async () => {
    await bridgePool
      .connect(bridgeRouter)
      .deposit(user.address, encodedDepositParameters);
    expect(await fixture.erc721Mock.balanceOf(user.address)).to.be.eq(
      initialUserBalance - tokenAmount
    );
    expect(await fixture.erc721Mock.balanceOf(bridgePool.address)).to.be.eq(
      initialBridgePoolBalance + tokenAmount
    );
  });

  it("Should make a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
    // Make first a deposit to withdraw afterwards
    await bridgePool
      .connect(bridgeRouter)
      .deposit(user.address, encodedDepositParameters);
    await bridgePool.connect(user).withdraw(encodedBurnedUTXO, merkleProof);
    expect(await fixture.erc721Mock.balanceOf(user.address)).to.be.eq(
      initialUserBalance
    );
    expect(await fixture.erc721Mock.balanceOf(bridgePool.address)).to.be.eq(
      initialBridgePoolBalance
    );
  });

  it("Should not make a withdraw for amount specified on informed burned UTXO with wrong merkle proof", async () => {
    await expect(
      bridgePool.connect(user).withdraw(encodedBurnedUTXO, wrongMerkleProof)
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
      bridgePool.connect(user).withdraw(encodedBurnedUTXO, merkleProof)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });

  it("Should not make a withdraw to an address that is not the owner in informed burned UTXO", async () => {
    await expect(
      bridgePool.connect(user2).withdraw(encodedBurnedUTXO, merkleProof)
    ).to.be.revertedWithCustomError(
      bridgePool,
      "ReceiverIsNotOwnerOnProofOfBurnUTXO"
    );
  });
});

const encodedDepositParameters = defaultAbiCoder.encode(
  ["tuple(uint256 tokenId_, uint256 tokenAmount_)"],
  [
    {
      tokenId_: tokenId,
      tokenAmount_: tokenAmount,
    },
  ]
);
