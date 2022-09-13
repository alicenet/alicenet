import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { IBridgePool } from "../../typechain-types";
import { expect } from "../chai-setup";

import { assert } from "console";
import { Fixture, getFixture } from "../setup";
import {
  encodedBurnedUTXO,
  getMockBlockClaimsForStateRoot,
  getSimulatedBridgeRouter,
  merkleProof,
  stateRoot,
  valueOrId,
  wrongMerkleProof,
} from "./setup";

let fixture: Fixture;
let admin: SignerWithAddress;
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
    quantity: valueOrId,
    errorReason: "ERC721: insufficient allowance",
  },
};

describe("Testing BridgePool Deposit/Withdraw for tokenType ERC721", async () => {
  beforeEach(async () => {
    [admin, user, user2] = await ethers.getSigners();
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
    //Simulate a bridge router with some gas for transactions
    bridgeRouter = await getSimulatedBridgeRouter(fixture.factory.address);
    bridgePool = fixture[erc721data.options.bridgeImpl] as IBridgePool;
    fixture[erc721data.options.ercContractName].mint(user.address, valueOrId);
    fixture[erc721data.options.ercContractName]
      .connect(user)
      .approve(bridgePool.address, valueOrId);
    initialUserBalance = await fixture.erc721Mock.balanceOf(user.address);
    initialBridgePoolBalance = await fixture.erc721Mock.balanceOf(
      bridgePool.address
    );
  });

  it("Should make a deposit", async () => {
    await bridgePool.connect(bridgeRouter).deposit(user.address, valueOrId);
    assert(
      await fixture.erc721Mock.balanceOf(user.address),
      initialUserBalance - 1
    );
    assert(
      await fixture.erc721Mock.balanceOf(bridgePool.address),
      initialBridgePoolBalance + 1
    );
  });

  it("Should make a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
    // Make first a deposit to withdraw afterwards
    await bridgePool.connect(bridgeRouter).deposit(user.address, valueOrId);
    await bridgePool.connect(user).withdraw(merkleProof, encodedBurnedUTXO);
    assert(
      await fixture.erc721Mock.balanceOf(user.address),
      initialUserBalance
    );
    assert(
      await fixture.erc721Mock.balanceOf(bridgePool.address),
      initialBridgePoolBalance
    );
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
