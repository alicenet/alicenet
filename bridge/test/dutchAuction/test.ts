import { expect } from "chai";

import { ethers, network } from "hardhat";
import { DutchAuction } from "../../typechain-types";

import {
  deployStaticWithFactory,
  factoryCallAnyFixture,
  Fixture,
  getFixture,
} from "../setup";

let fixture: Fixture;
let dutchAuction: DutchAuction;
const validators: String[] = [];
const stakingTokenIds: String[] = [];

describe("Testing Dutch Auction", async () => {
  beforeEach(async function () {
    fixture = await getFixture(true, false, false);
    validators.push("0x0000000000000000000000000000000000000000");
    validators.push("0x0000000000000000000000000000000000000001");
    validators.push("0x0000000000000000000000000000000000000002");
    validators.push("0x0000000000000000000000000000000000000003");
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    dutchAuction = (await deployStaticWithFactory(
      fixture.factory,
      "DutchAuction",
      "DutchAuction",
      undefined
    )) as DutchAuction;
  });

  it("Should obtain bid price", async () => {
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("1961725490196078431372")
    );
  });

  it("Should obtain prices through blocks according with dutch auction curve", async () => {
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("1961725490196078431372")
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("1949737623232709209017")
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("1937895450056969236612")
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("1926196330690826727066")
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("1914637688555347091932")
    );
  });

  it.skip("Simulate a 7 days run (skipped since it may take too long)", async () => {
    for (let d = 1; d <= 15; d++) {
      console.log(
        "Day",
        d,
        "Price",
        ethers.utils.formatEther(await dutchAuction.getPrice()),
        await dutchAuction.getPrice()
      );
      for (let i = 0; i < 5760; i++) {
        await network.provider.send("evm_mine");
      }
    }
  });
});
