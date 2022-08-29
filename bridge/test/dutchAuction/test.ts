import { expect } from "chai";
import { ethers, network } from "hardhat";
import { DutchAuction } from "../../typechain-types";

import {
  deployStaticWithFactory,
  factoryCallAny,
  factoryCallAnyFixture,
  Fixture,
  getFixture,
} from "../setup";

let fixture: Fixture;
let dutchAuction: DutchAuction;
const VALIDATORS: String[] = [];
const EXPECTED_PRICE_BLOCK_ZERO: Number = Number("10000950400000000000000");
const EXPECTED_PRICE_BLOCK_ONE: Number = Number("9984975974440894568690");
const EXPECTED_PRICE_BLOCK_TWO: Number = Number("9969052503987240829346");
const EXPECTED_PRICE_BLOCK_THREE: Number = Number("9953179745222929936305");
const EXPECTED_PRICE_BLOCK_FOUR: Number = Number("9937357456279809220985");

describe("Testing Dutch Auction", async () => {
  beforeEach(async function () {
    fixture = await getFixture(true, false, false);
    //Simulate 4 registered validators
    VALIDATORS.push("0x0000000000000000000000000000000000000000");
    VALIDATORS.push("0x0000000000000000000000000000000000000001");
    VALIDATORS.push("0x0000000000000000000000000000000000000002");
    VALIDATORS.push("0x0000000000000000000000000000000000000003");
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [VALIDATORS, []]
    );
    dutchAuction = (await deployStaticWithFactory(
      fixture.factory,
      "DutchAuction",
      "DutchAuction",
      undefined
    )) as DutchAuction;
  });

  it("Should obtain bid price at first auction block", async () => {
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
  });

  it("Should obtain prices through five blocks according to dutch auction curve", async () => {
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      EXPECTED_PRICE_BLOCK_ONE
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      EXPECTED_PRICE_BLOCK_TWO
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      EXPECTED_PRICE_BLOCK_THREE
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      EXPECTED_PRICE_BLOCK_FOUR
    );
  });

  it("Should restart the auction", async () => {
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      EXPECTED_PRICE_BLOCK_ONE
    );
    await factoryCallAny(fixture.factory, dutchAuction, "resetAuction", []);
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
  });

  it.skip("Simulates a 30 days run (skipped since it takes several minutes)", async () => {
    for (let d = 1; d <= 30; d++) {
      console.log(
        "Day",
        d,
        "Price",
        ethers.utils.formatEther(await dutchAuction.getPrice()),
        await dutchAuction.getPrice(),
        "Block",
        await ethers.provider.getBlockNumber()
      );
      for (let i = 0; i < 5760; i++) {
        await network.provider.send("evm_mine");
      }
    }
  });
});
