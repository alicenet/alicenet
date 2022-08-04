import { expect } from "chai";

import { ethers, network } from "hardhat";
import { DutchAuction } from "../../typechain-types";
import { assertErrorMessage } from "../chai-helpers";

import {
  deployStaticWithFactory,
  factoryCallAny,
  Fixture,
  getFixture,
} from "../setup";

let fixture: Fixture;
let dutchAuction: DutchAuction;

const startPrice = ethers.utils.parseEther("10.0");
const finalPrice = ethers.utils.parseEther("1.0");
const bidders = 20; // potential validators
const durationInBlocks = 4;
const initialDelta = 20;

describe("Testing Dutch Auction", async () => {
  beforeEach(async function () {
    fixture = await getFixture(false, false, false);
    dutchAuction = (await deployStaticWithFactory(
      fixture.factory,
      "DutchAuction",
      "DutchAuction",
      [initialDelta]
    )) as DutchAuction;
  });

  it("Should obtain price is auction is started", async () => {
    await factoryCallAny(fixture.factory, dutchAuction, "startAuction", [
      startPrice,
      finalPrice,
      bidders,
      durationInBlocks,
    ]);
    await dutchAuction.getPrice();
  });

  it("Should not obtain price is auction is not started", async () => {
    await assertErrorMessage(dutchAuction.getPrice(), `AuctionClosed()`);
  });

  it("Should obtain lower bid prices through blocks", async () => {
    await factoryCallAny(fixture.factory, dutchAuction, "startAuction", [
      startPrice,
      finalPrice,
      bidders,
      durationInBlocks,
    ]);
    let previousPrice = await dutchAuction.getPrice();
    for (let i = 0; i < durationInBlocks; i++) {
      await network.provider.send("evm_mine");
      expect(Number(previousPrice)).to.be.greaterThan(
        Number(await dutchAuction.getPrice())
      );
      previousPrice = await dutchAuction.getPrice();
    }
  });

  it("Should not obtain price if auction is closed", async () => {
    await factoryCallAny(fixture.factory, dutchAuction, "startAuction", [
      startPrice,
      finalPrice,
      bidders,
      durationInBlocks,
    ]);
    for (let i = 0; i < durationInBlocks; i++) {
      await network.provider.send("evm_mine"); // Mine one block
    }
    await network.provider.send("evm_mine"); // Mine one more block
    await assertErrorMessage(dutchAuction.getPrice(), `AuctionClosed()`);
  });

  it("Should not start auction with start price lower than final price", async () => {
    const startPrice = 0;
    await assertErrorMessage(
      factoryCallAny(fixture.factory, dutchAuction, "startAuction", [
        startPrice,
        finalPrice,
        bidders,
        durationInBlocks,
      ]),
      `IcorrectInitialPrices()`
    );
  });
});
