import { expect } from "chai";
import { BigNumber } from "ethers";

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
let durationInBlocks = 4;

describe("Testing Dutch Auction", async () => {
  beforeEach(async function () {
    fixture = await getFixture(false, false, false);
    dutchAuction = (await deployStaticWithFactory(
      fixture.factory,
      "DutchAuction",
      "DutchAuction", undefined)) as DutchAuction;
  });

  it("Should obtain price is auction is started", async () => {
    await factoryCallAny(fixture.factory, dutchAuction, "startAuction");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("200000000000000384000000")
    );
  });

  it("Should not obtain price is auction is not started", async () => {
    await assertErrorMessage(dutchAuction.getPrice(), `AuctionNotStarted()`);
  });

  it("Should obtain prices according with dutch auction curve", async () => {
    await factoryCallAny(fixture.factory, dutchAuction, "startAuction", [
    ]);
    let previousPrice = await dutchAuction.getPrice();
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("200000000000000384000000") // 8200 ETH
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("81967213114754539016393") // 5100 ETH
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("51546391752577774845360") // 3700 ETH
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("37593984962406476992481") // 3000 ETH
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("29585798816568513136094") // 3000 ETH
    );
  });

  it.skip("Simulate a 15 days run", async () => {
    await factoryCallAny(fixture.factory, dutchAuction, "startAuction", [
    ]);
    let previousPrice = await dutchAuction.getPrice();
    for (let d = 0; d < 15; d++) {
      for (let i = 0; i < 5760; i++) {
        await network.provider.send("evm_mine");
      }
      console.log("Day", d, "Price", ethers.utils.formatEther(await dutchAuction.getPrice()));
    }
  });

});
