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

  it("Should obtain bid price at first auction block", async () => {
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("10000950400000000000000")
    );
  });

  it("Should obtain prices through five blocks according to dutch auction curve", async () => {
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("10000950400000000000000")
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("9984975974440894568690")
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("9969052503987240829346")
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("9953179745222929936305")
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("9937357456279809220985")
    );
  });

  it("Should restart the auction", async () => {
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("10000950400000000000000")
    );
    await network.provider.send("evm_mine");
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("9984975974440894568690")
    );
    await factoryCallAny(fixture.factory, dutchAuction, "resetAuction", []);
    expect(Number(await dutchAuction.getPrice())).to.be.equal(
      Number("10000950400000000000000")
    );
  });

  it.skip("Simulates a 30 days run (skipped since it might take too long)", async () => {
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
