import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers, network } from "hardhat";
import { DutchAuction } from "../../typechain-types";
import { PromiseOrValue } from "../../typechain-types/common";

import {
  deployUpgradeableWithFactory,
  factoryCallAny,
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  mineBlocks,
} from "../setup";

let fixture: Fixture;
let dutchAuction: DutchAuction;
const VALIDATORS: PromiseOrValue<string>[] = [];
const EXPECTED_PRICE_BLOCK_ZERO = ethers.utils.parseEther(
  "10000.950400000000000000"
);
const EXPECTED_PRICE_BLOCK_ONE = ethers.utils.parseEther(
  "9984.975974440894568690"
);
const EXPECTED_PRICE_BLOCK_TWO = ethers.utils.parseEther(
  "9969.052503987240829346"
);
const EXPECTED_PRICE_BLOCK_THREE = ethers.utils.parseEther(
  "9953.179745222929936305"
);
const EXPECTED_PRICE_BLOCK_FOUR = ethers.utils.parseEther(
  "9937.357456279809220985"
);
const dailyExpectedPriceFirstMonth = [
  "10000950400000000000000",
  "979815755677368833202",
  "515574573898723754631",
  "350024172018989109187",
  "265062852313543207268",
  "213364214103653355989",
  "178592343328122779593",
  "153603643912565636829",
  "134778520501017021732",
  "120086922710378347469",
  "108302103907256333190",
  "98639049777291552707",
  "90572072550003584486",
  "83735895636050592675",
  "77868804528394757890",
  "72778374030451019821",
  "68319961200625101040",
  "64382740879801106093",
  "60880368151095345381",
  "57744572752464452823",
  "54920664796028491258",
  "52364317966854463955",
  "50039225725391652597",
  "47915366064385259757",
  "45967698124077341302",
  "44175170267934312878",
  "42519956112644213186",
  "40986859649684588043",
  "39562847348753898891",
  "38236676706527897891",
];

async function deployFixture() {
  const fixture = await getFixture(true, false, false);
  // Simulate 4 registered validators
  VALIDATORS.push("0x0000000000000000000000000000000000000000");
  VALIDATORS.push("0x0000000000000000000000000000000000000001");
  VALIDATORS.push("0x0000000000000000000000000000000000000002");
  VALIDATORS.push("0x0000000000000000000000000000000000000003");
  await factoryCallAnyFixture(fixture, "validatorPool", "registerValidators", [
    VALIDATORS,
    [],
  ]);
  dutchAuction = (await deployUpgradeableWithFactory(
    fixture.factory,
    "DutchAuction",
    "DutchAuction"
  )) as DutchAuction;
  return { fixture, dutchAuction };
}

describe("Testing Dutch Auction", async () => {
  beforeEach(async function () {
    ({ fixture, dutchAuction } = await loadFixture(deployFixture));
  });

  it("Should obtain bid price at first auction block", async () => {
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
  });

  it("Should obtain prices through five blocks according to dutch auction curve", async () => {
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
    await network.provider.send("evm_mine");
    expect(await dutchAuction.getPrice()).to.be.equal(EXPECTED_PRICE_BLOCK_ONE);
    await network.provider.send("evm_mine");
    expect(await dutchAuction.getPrice()).to.be.equal(EXPECTED_PRICE_BLOCK_TWO);
    await network.provider.send("evm_mine");
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_THREE
    );
    await network.provider.send("evm_mine");
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_FOUR
    );
  });

  it("Should restart the auction", async () => {
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
    await network.provider.send("evm_mine");
    expect(await dutchAuction.getPrice()).to.be.equal(EXPECTED_PRICE_BLOCK_ONE);
    await factoryCallAny(fixture.factory, dutchAuction, "resetAuction", []);
    expect(await dutchAuction.getPrice()).to.be.equal(
      EXPECTED_PRICE_BLOCK_ZERO
    );
  });

  it("should get right prices for the first 30 days", async () => {
    for (let day = 1; day <= 30; day++) {
      expect(await dutchAuction.getPrice()).to.be.equal(
        dailyExpectedPriceFirstMonth[day - 1]
      );
      await mineBlocks(5760n);
    }
  });
});
