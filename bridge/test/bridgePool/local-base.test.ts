import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { Contract } from "ethers";
import { defaultAbiCoder } from "ethers/lib/utils";
import { ethers } from "hardhat";
import { IBridgePool } from "../../typechain-types";
import { addValidators, getValidAccusationDataForNonExistentUTXO } from "../accusations/accusations-test-helpers";
import { expect } from "../chai-setup";

import { deployUpgradeableWithFactory, Fixture, getFixture } from "../setup";
import {
  getEncodedBurnedUTXO,
  getMockBlockClaimsForStateRoot,
  getWithdrawParameters,
  getWithdrawParameters2,
  merkleProof,
  stateRoot,
  wrongMerkleProof,
} from "./setup";

let fixture: Fixture;
let user: SignerWithAddress;
let user2: SignerWithAddress;
let merkleProofLibraryErrors: Contract;
let localERCBridgePoolBase: Contract;
let encodedBurnedUTXO: string;
let bridgePool: any;
const withdrawParameters = getWithdrawParameters2();
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

describe("Testing Base BridgePool Deposit/Withdraw", async () => {
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
    localERCBridgePoolBase = await deployUpgradeableWithFactory(
      fixture.factory,
      "LocalERCBridgePoolMock",
      "LocalERCBridgePoolMock",
      undefined,
      undefined,
      undefined,
      false
    );
    encodedBurnedUTXO = getEncodedBurnedUTXO(
      user.address,
      tokenId,
      tokenAmount
    );
    bridgePool = localERCBridgePoolBase as IBridgePool;
  });

  it("Should call a deposit", async () => {
    await localERCBridgePoolBase.deposit(
      user.address,
      encodedDepositParameters
    );
  });

  it("Should call a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
     await localERCBridgePoolBase
      .connect(user)
      .withdraw(withdrawParameters.bClaims, withdrawParameters.bClaimsSigGroup, withdrawParameters.txInPreImage, withdrawParameters.proofs);
  });

  it.only("Should not call a withdraw for amount specified on informed burned UTXO upon proof verification", async () => {
    await localERCBridgePoolBase
     .connect(user)
     .withdraw(withdrawParameters.bClaims, withdrawParameters.bClaimsSigGroup, withdrawParameters.txInPreImage, withdrawParameters.proofs);
     await localERCBridgePoolBase
     .connect(user)
     .withdraw(withdrawParameters.bClaims, withdrawParameters.bClaimsSigGroup, withdrawParameters.txInPreImage, withdrawParameters.proofs);
 });

  it("Should not call a withdraw for amount specified on informed burned UTXO with wrong merkle proof", async () => {
    await expect(
      localERCBridgePoolBase
        .connect(user)
        .withdraw(encodedBurnedUTXO, wrongMerkleProof)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });

  it("Should not call a withdraw for amount specified on informed burned UTXO with wrong root", async () => {
    const wrongStateRoot =
      "0x0000000000000000000000000000000000000000000000000000000000000000";
    const encodedMockBlockClaims =
      getMockBlockClaimsForStateRoot(wrongStateRoot);
    await fixture.snapshots.snapshot(
      Buffer.from("0x0"),
      encodedMockBlockClaims
    );
    await expect(
      localERCBridgePoolBase
        .connect(user)
        .withdraw(encodedBurnedUTXO, merkleProof)
    ).to.be.revertedWithCustomError(
      merkleProofLibraryErrors,
      "ProofDoesNotMatchTrieRoot"
    );
  });

  it("Should not call a withdraw to an address that is not the owner in informed burned UTXO", async () => {
    await expect(
      localERCBridgePoolBase
        .connect(user2)
        .withdraw(encodedBurnedUTXO, merkleProof)
    ).to.be.revertedWithCustomError(
      bridgePool,
      "ReceiverIsNotOwnerOnProofOfBurnUTXO"
    );
  });
});
