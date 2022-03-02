import { Fixture, getFixture, getValidatorEthAccount } from "../setup";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { completeETHDKGRound } from "../ethdkg/setup";
import { BigNumber } from "ethers";
import {
  validatorsSnapshots,
  validSnapshot1024,
  invalidSnapshot500,
  invalidSnapshotChainID2,
  invalidSnapshotIncorrectSig,
  validSnapshot2048,
} from "./assets/4-validators-snapshots-1";
import { Snapshots } from "../../typechain-types";

describe("Snapshots: With successful snapshot completed", () => {
  let fixture: Fixture;
  let snapshots: Snapshots;
  let snapshotNumber: BigNumber;

  beforeEach(async function () {
    fixture = await getFixture(true, false);

    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });

    snapshots = fixture.snapshots as Snapshots;
    await snapshots
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .snapshot(validSnapshot1024.GroupSignature, validSnapshot1024.BClaims);
    snapshotNumber = BigNumber.from(1);
  });

  it("Does not allow snapshot with data from previous snapshot", async function () {
    let validValidator = await getValidatorEthAccount(validatorsSnapshots[0]);
    await expect(
      fixture.snapshots
        .connect(validValidator)
        .snapshot(validSnapshot1024.GroupSignature, validSnapshot1024.BClaims)
    ).to.be.revertedWith(`Snapshots: Incorrect Madnet height for snapshot!`);
  });

  it("Does not allow snapshot if ETHDKG round is Running", async function () {
    await fixture.validatorPool.scheduleMaintenance();
    await snapshots
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .snapshot(validSnapshot2048.GroupSignature, validSnapshot2048.BClaims);
    await fixture.validatorPool.initializeETHDKG();
    let junkData =
      "0x0000000000000000000000000000000000000000000000000000006d6168616d";
    let validValidator = await getValidatorEthAccount(validatorsSnapshots[0]);
    await expect(
      fixture.snapshots.connect(validValidator).snapshot(junkData, junkData)
    ).to.be.revertedWith(`Snapshots: Consensus is not running!`);
  });

  it("getLatestSnapshot returns correct snapshot data", async function () {
    const expectedSignature = [
      BigNumber.from(
        "1255022359938341263552008964652785372053438514616831677297275448520908946987"
      ),
      BigNumber.from(
        "14701588978138831040868532458058035157389630420138682442198805011661026372629"
      ),
    ];
    const expectedChainId = BigNumber.from(1);
    const expectedHeight = BigNumber.from(1024);
    const expectedTxCount = BigNumber.from(0);
    const expectedPrevBlock = BigNumber.from(
      "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    );
    const expectedTxRoot = BigNumber.from(
      "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    );
    const expectedStateRoot = BigNumber.from(
      "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    );
    const expectedHeaderRoot = BigNumber.from(
      "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    );

    const snapshotData = await snapshots
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .getLatestSnapshot();

    const blockClaims = snapshotData.blockClaims;
    await expect(blockClaims.chainId).to.be.equal(expectedChainId);
    await expect(blockClaims.height).to.be.equal(expectedHeight);
    await expect(blockClaims.txCount).to.be.equal(expectedTxCount);
    await expect(blockClaims.prevBlock).to.be.equal(expectedPrevBlock);
    await expect(blockClaims.txRoot).to.be.equal(expectedTxRoot);
    await expect(blockClaims.stateRoot).to.be.equal(expectedStateRoot);
    await expect(blockClaims.headerRoot).to.be.equal(expectedHeaderRoot);
  });

  it("getBlockClaimsFromSnapshot returns correct data", async function () {
    const expectedChainId = BigNumber.from(1);
    const expectedHeight = BigNumber.from(1024);
    const expectedTxCount = BigNumber.from(0);
    const expectedPrevBlock = BigNumber.from(
      "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    );
    const expectedTxRoot = BigNumber.from(
      "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    );
    const expectedStateRoot = BigNumber.from(
      "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    );
    const expectedHeaderRoot = BigNumber.from(
      "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    );

    const blockClaims = await snapshots
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .getBlockClaimsFromSnapshot(snapshotNumber);

    await expect(blockClaims.chainId).to.be.equal(expectedChainId);
    await expect(blockClaims.height).to.be.equal(expectedHeight);
    await expect(blockClaims.txCount).to.be.equal(expectedTxCount);
    await expect(blockClaims.prevBlock).to.be.equal(expectedPrevBlock);
    await expect(blockClaims.txRoot).to.be.equal(expectedTxRoot);
    await expect(blockClaims.stateRoot).to.be.equal(expectedStateRoot);
    await expect(blockClaims.headerRoot).to.be.equal(expectedHeaderRoot);
  });

  it("getBlockClaimsFromLatestSnapshot returns correct data", async function () {
    const expectedChainId = BigNumber.from(1);
    const expectedHeight = BigNumber.from(1024);
    const expectedTxCount = BigNumber.from(0);
    const expectedPrevBlock = BigNumber.from(
      "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    );
    const expectedTxRoot = BigNumber.from(
      "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    );
    const expectedStateRoot = BigNumber.from(
      "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    );
    const expectedHeaderRoot = BigNumber.from(
      "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    );

    const blockClaims = await snapshots
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .getBlockClaimsFromLatestSnapshot();

    await expect(blockClaims.chainId).to.be.equal(expectedChainId);
    await expect(blockClaims.height).to.be.equal(expectedHeight);
    await expect(blockClaims.txCount).to.be.equal(expectedTxCount);
    await expect(blockClaims.prevBlock).to.be.equal(expectedPrevBlock);
    await expect(blockClaims.txRoot).to.be.equal(expectedTxRoot);
    await expect(blockClaims.stateRoot).to.be.equal(expectedStateRoot);
    await expect(blockClaims.headerRoot).to.be.equal(expectedHeaderRoot);
  });

  it("getMadnetHeightFromSnapshot returns correct data", async function () {
    const expectedHeight = BigNumber.from(1024);

    const height = await snapshots
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .getMadnetHeightFromSnapshot(snapshotNumber);

    await expect(height).to.be.equal(expectedHeight);
  });

  it("getMadnetHeightFromLatestSnapshot returns correct data", async function () {
    const expectedHeight = BigNumber.from(1024);

    const height = await snapshots
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .getMadnetHeightFromLatestSnapshot();

    await expect(height).to.be.equal(expectedHeight);
  });

  it("getChainIdFromSnapshot returns correct chain id", async function () {
    const expectedChainId = 1;
    const chainId = await snapshots
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .getChainIdFromSnapshot(snapshotNumber);

    await expect(chainId).to.be.equal(expectedChainId);
  });

  it("getChainIdFromLatestSnapshot returns correct chain id", async function () {
    const expectedChainId = 1;
    const chainId = await snapshots
      .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
      .getChainIdFromLatestSnapshot();

    await expect(chainId).to.be.equal(expectedChainId);
  });
});
